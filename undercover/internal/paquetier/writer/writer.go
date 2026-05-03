// Package writer persists [event.WebhookEvent] records to Parquet files using
// an in-process DuckDB instance.  Events are buffered in an in-memory DuckDB
// table and flushed to Hive-partitioned Parquet files on demand (e.g. at
// shutdown or on a periodic timer).
//
// Two storage backends are supported:
//
//   - Local disk (default): files are written under dataDir.
//   - S3-compatible object storage (e.g. Supabase Storage): files are written
//     to s3://<bucket>/<prefix>/... using DuckDB's built-in httpfs extension.
//     Set S3Config to enable.
//
// Hive partition layout (same for both backends):
//
//	provider=github/year=2026/month=03/day=09/event_type=push/
//	  events_<nanos>_0.parquet
package writer

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"
	"undercover/pkg/event"

	_ "github.com/marcboeker/go-duckdb"
)

// S3Config holds credentials for an S3-compatible object storage backend.
// Supabase Storage values:
//
//	Endpoint  — <project-ref>.supabase.co/storage/v1/s3
//	Region    — auto  (or the region shown in Supabase Storage settings)
//	Bucket    — the bucket name you created in Supabase Storage
//	AccessKey — from Supabase Dashboard → Storage → S3 credentials
//	SecretKey — from Supabase Dashboard → Storage → S3 credentials
type S3Config struct {
	Endpoint  string // host[:port], no scheme
	Region    string
	Bucket    string
	Prefix    string // optional key prefix, e.g. "events"
	AccessKey string
	SecretKey string
}

// Writer buffers WebhookEvents in an in-memory DuckDB table and flushes them
// to Hive-partitioned Parquet files.
type Writer struct {
	db      *sql.DB
	destDir string // local path OR s3://bucket/prefix
	mu      sync.Mutex
	actors  map[string]struct{} // actor_logins buffered since last flush
}

// NewWriter opens an in-memory DuckDB instance and returns a ready-to-use
// Writer.
//
// If s3 is non-nil the writer uses Supabase / S3-compatible storage; dataDir
// is ignored in that case.  If s3 is nil the writer writes to dataDir on the
// local filesystem (dataDir is created if it does not exist).
func NewWriter(dataDir string, s3 *S3Config) (*Writer, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("writer: open duckdb: %w", err)
	}

	var destDir string

	if s3 != nil {
		if err := configureS3(db, s3); err != nil {
			db.Close()
			return nil, err
		}
		prefix := s3.Prefix
		if prefix != "" {
			prefix = "/" + prefix
		}
		destDir = fmt.Sprintf("s3://%s%s", s3.Bucket, prefix)
	} else {
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			db.Close()
			return nil, fmt.Errorf("writer: create data dir: %w", err)
		}
		destDir = dataDir
	}

	if _, err = db.Exec(`
		CREATE TABLE webhook_events (
			id          VARCHAR,
			provider    VARCHAR,
			event_type  VARCHAR,
			action      VARCHAR,
			actor_login VARCHAR,
			actor_id    BIGINT,
			repo_owner  VARCHAR,
			repo_name   VARCHAR,
			ref         VARCHAR,
			received_at TIMESTAMP,
			payload     JSON,
			year        INTEGER,
			month       INTEGER,
			day         INTEGER
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("writer: create table: %w", err)
	}

	return &Writer{db: db, destDir: destDir, actors: make(map[string]struct{})}, nil
}

// configureS3 installs and configures DuckDB's httpfs extension for the given
// S3-compatible endpoint.
func configureS3(db *sql.DB, cfg *S3Config) error {
	stmts := []string{
		"INSTALL httpfs",
		"LOAD httpfs",
		fmt.Sprintf("SET s3_endpoint='%s'", cfg.Endpoint),
		fmt.Sprintf("SET s3_region='%s'", cfg.Region),
		fmt.Sprintf("SET s3_access_key_id='%s'", cfg.AccessKey),
		fmt.Sprintf("SET s3_secret_access_key='%s'", cfg.SecretKey),
		"SET s3_use_ssl=true",
		"SET s3_url_style='path'",
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("writer: s3 setup (%q): %w", s, err)
		}
	}
	return nil
}

// Insert appends a single WebhookEvent to the in-memory buffer.
func (w *Writer) Insert(e *event.WebhookEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	ts := e.ReceivedAt.UTC()
	_, err := w.db.Exec(`
		INSERT INTO webhook_events VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		e.ID,
		string(e.Provider),
		e.EventType,
		e.Action,
		e.ActorLogin,
		e.ActorID,
		e.RepoOwner,
		e.RepoName,
		e.Ref,
		ts,
		string(e.Payload),
		ts.Year(),
		int(ts.Month()),
		ts.Day(),
	)
	if err == nil && e.ActorLogin != "" {
		w.actors[e.ActorLogin] = struct{}{}
	}
	return err
}

// Flush writes all buffered events to Hive-partitioned Parquet files, clears
// the in-memory buffer, and returns the distinct actor_logins that were flushed.
// Returns nil actors and no error when the buffer is empty.
func (w *Writer) Flush() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	row := w.db.QueryRow("SELECT COUNT(*) FROM webhook_events")
	var n int64
	if err := row.Scan(&n); err != nil {
		return nil, fmt.Errorf("writer: count rows: %w", err)
	}
	if n == 0 {
		return nil, nil
	}

	pattern := fmt.Sprintf("events_%d", time.Now().UnixNano())
	query := fmt.Sprintf(
		"COPY (SELECT * FROM webhook_events) TO '%s' "+
			"(FORMAT PARQUET, PARTITION_BY (provider, year, month, day, event_type), "+
			"FILENAME_PATTERN '%s', OVERWRITE_OR_IGNORE TRUE)",
		w.destDir, pattern,
	)
	if _, err := w.db.Exec(query); err != nil {
		return nil, fmt.Errorf("writer: flush parquet: %w", err)
	}

	if _, err := w.db.Exec("DELETE FROM webhook_events"); err != nil {
		return nil, fmt.Errorf("writer: clear buffer: %w", err)
	}

	actors := make([]string, 0, len(w.actors))
	for a := range w.actors {
		actors = append(actors, a)
	}
	w.actors = make(map[string]struct{})

	return actors, nil
}

// Close flushes any remaining events and closes the DuckDB connection.
func (w *Writer) Close() error {
	if _, err := w.Flush(); err != nil {
		return err
	}
	return w.db.Close()
}
