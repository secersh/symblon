package evaluator

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/marcboeker/go-duckdb"

	"undercover/internal/overseer/store"
	"undercover/pkg/event"
)

// S3Config holds credentials for reading Parquet events and fetching rule SQL.
type S3Config struct {
	Endpoint     string // e.g. <project>.supabase.co/storage/v1/s3
	Region       string
	Bucket       string
	EventsPrefix string // prefix where Parquet events are stored, e.g. "events"
	AgentsPrefix string // prefix where rule SQL files are stored, e.g. "agents"
	AccessKey    string
	SecretKey    string
}

// DuckDBEvaluator implements trigger.Evaluator using an in-process DuckDB
// instance that reads Hive-partitioned Parquet files from S3.
//
// Each evaluation:
//  1. Fetches the rule SQL from S3 (cached after first fetch)
//  2. Substitutes temporal window params if the rule is temporal
//  3. Executes the SQL against the events view
//  4. Returns true if the triggering actor_login appears in the result set
type DuckDBEvaluator struct {
	db       *sql.DB
	s3cfg    S3Config
	s3client *awss3.Client
	sqlCache sync.Map // key: sql_path → value: string
}

// New opens an in-process DuckDB instance, configures S3 access via httpfs,
// and creates the events and issued_symbols views.
func New(cfg S3Config) (*DuckDBEvaluator, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("evaluator: open duckdb: %w", err)
	}

	// DuckDB httpfs expects host only — strip scheme if present.
	cfg.Endpoint = strings.TrimPrefix(cfg.Endpoint, "https://")
	cfg.Endpoint = strings.TrimPrefix(cfg.Endpoint, "http://")

	if err := configureS3(db, cfg); err != nil {
		db.Close()
		return nil, err
	}

	if err := createViews(db, cfg); err != nil {
		db.Close()
		return nil, err
	}

	s3client := awss3.New(awss3.Options{
		BaseEndpoint: aws.String("https://" + cfg.Endpoint),
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: true,
		HTTPClient:   &http.Client{},
	})

	return &DuckDBEvaluator{db: db, s3cfg: cfg, s3client: s3client}, nil
}

// Close releases the DuckDB connection.
func (e *DuckDBEvaluator) Close() error {
	return e.db.Close()
}

// Evaluate fetches the rule SQL, substitutes window params for temporal rules,
// executes it, and returns true if the actor_login is in the result set.
func (e *DuckDBEvaluator) Evaluate(ctx context.Context, rule *store.InstalledRule, evt *event.WebhookEvent) (bool, error) {
	query, err := e.fetchSQL(ctx, rule.SQLPath)
	if err != nil {
		return false, fmt.Errorf("fetch SQL for %s/%s: %w", rule.AgentID, rule.SymbolID, err)
	}

	if rule.Type == "temporal" {
		now := time.Now().UTC()
		windowStart := now.Add(-time.Duration(rule.WindowHours) * time.Hour)
		query = strings.ReplaceAll(query, ":window_start", "'"+windowStart.Format(time.RFC3339)+"'")
		query = strings.ReplaceAll(query, ":window_end", "'"+now.Format(time.RFC3339)+"'")
	}

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("evaluate %s/%s: %w", rule.AgentID, rule.SymbolID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var actorLogin string
		if err := rows.Scan(&actorLogin); err != nil {
			return false, fmt.Errorf("scan result: %w", err)
		}
		if actorLogin == evt.ActorLogin {
			log.Printf("[evaluator] resolved — agent=%s symbol=%s actor=%s", rule.AgentID, rule.SymbolID, evt.ActorLogin)
			return true, nil
		}
	}

	return false, rows.Err()
}

// fetchSQL returns the SQL for the given S3 path, fetching and caching on first call.
func (e *DuckDBEvaluator) fetchSQL(ctx context.Context, sqlPath string) (string, error) {
	if cached, ok := e.sqlCache.Load(sqlPath); ok {
		return cached.(string), nil
	}

	key := e.keyFromPath(sqlPath)
	resp, err := e.s3client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(e.s3cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", sqlPath, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", sqlPath, err)
	}

	query := strings.TrimSpace(string(body))
	e.sqlCache.Store(sqlPath, query)
	return query, nil
}

// keyFromPath strips the s3://bucket/ prefix and returns the object key.
func (e *DuckDBEvaluator) keyFromPath(sqlPath string) string {
	prefix := "s3://" + e.s3cfg.Bucket + "/"
	return strings.TrimPrefix(sqlPath, prefix)
}

func configureS3(db *sql.DB, cfg S3Config) error {
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
			return fmt.Errorf("evaluator: s3 setup (%q): %w", s, err)
		}
	}
	return nil
}

func createViews(db *sql.DB, cfg S3Config) error {
	eventsPath := fmt.Sprintf("s3://%s/%s/**/*.parquet", cfg.Bucket, cfg.EventsPrefix)
	_, err := db.Exec(fmt.Sprintf(`
		CREATE OR REPLACE VIEW events AS
		SELECT * FROM read_parquet('%s', hive_partitioning=true)
	`, eventsPath))
	if err != nil {
		return fmt.Errorf("evaluator: create events view: %w", err)
	}

	// Stub until symbol issuance is implemented — prevents rule SQL from failing
	// on NOT EXISTS (SELECT 1 FROM issued_symbols ...) clauses.
	_, err = db.Exec(`
		CREATE OR REPLACE VIEW issued_symbols AS
		SELECT ''::VARCHAR AS actor_login, ''::VARCHAR AS symbol_id
		WHERE false
	`)
	if err != nil {
		return fmt.Errorf("evaluator: create issued_symbols view: %w", err)
	}

	return nil
}
