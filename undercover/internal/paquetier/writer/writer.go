// Package writer persists [event.WebhookEvent] records to Parquet files using
// an in-process DuckDB instance.  Events are buffered in an in-memory DuckDB
// table and flushed to disk in Hive-partitioned Parquet files on demand (e.g.
// at shutdown or on a periodic timer).
//
// On-disk layout after a flush:
//
//	<dataDir>/
//	  provider=github/
//	    year=2026/
//	      month=03/
//	        day=09/
//	          event_type=push/
//	            events_<nanos>_0.parquet
//	          event_type=pull_request/
//	            events_<nanos>_0.parquet
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

// Writer buffers WebhookEvents in an in-memory DuckDB table and flushes them
// to Hive-partitioned Parquet files under dataDir.
type Writer struct {
	db      *sql.DB
	dataDir string
	mu      sync.Mutex
}

// NewWriter opens an in-memory DuckDB instance, creates the events table, and
// returns a ready-to-use Writer.  dataDir is created if it does not exist.
func NewWriter(dataDir string) (*Writer, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("writer: create data dir: %w", err)
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("writer: open duckdb: %w", err)
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

	return &Writer{db: db, dataDir: dataDir}, nil
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
	return err
}

// Flush writes all buffered events to Hive-partitioned Parquet files and then
// clears the in-memory buffer.  Calling Flush on an empty buffer is a no-op.
// Each flush uses a nanosecond timestamp in the filename so that successive
// flushes within the same partition directory produce distinct files.
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	row := w.db.QueryRow("SELECT COUNT(*) FROM webhook_events")
	var n int64
	if err := row.Scan(&n); err != nil {
		return fmt.Errorf("writer: count rows: %w", err)
	}
	if n == 0 {
		return nil
	}

	pattern := fmt.Sprintf("events_%d", time.Now().UnixNano())
	query := fmt.Sprintf(
		"COPY (SELECT * FROM webhook_events) TO '%s' "+
			"(FORMAT PARQUET, PARTITION_BY (provider, year, month, day, event_type), "+
			"FILENAME_PATTERN '%s', OVERWRITE_OR_IGNORE TRUE)",
		w.dataDir, pattern,
	)
	if _, err := w.db.Exec(query); err != nil {
		return fmt.Errorf("writer: flush parquet: %w", err)
	}

	if _, err := w.db.Exec("DELETE FROM webhook_events"); err != nil {
		return fmt.Errorf("writer: clear buffer: %w", err)
	}

	return nil
}

// Close flushes any remaining events and closes the DuckDB connection.
func (w *Writer) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	return w.db.Close()
}
