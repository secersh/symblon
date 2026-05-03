package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InstalledRule is one symbol rule installed by a user, as recorded by the
// overseer after consuming an agent.installed event from the registrar.
type InstalledRule struct {
	UserID      string // Supabase UUID
	ActorLogin  string // GitHub username — used to match parquet.flushed actors
	AgentID     string // fully-qualified: publisher/handle/version
	SymbolID    string
	SQLPath     string // S3 path to the rule file
	Type        string // realtime | temporal
	WindowHours int
	InstalledAt time.Time
}

// InstalledRuleStore is the read/write interface for the overseer's local
// install index. It is populated by consuming agent.installed events from
// the registrar and is the source of truth for evaluation fan-out.
type InstalledRuleStore interface {
	// Save upserts a batch of rules for a user+agent install.
	Save(ctx context.Context, rules []InstalledRule) error

	// Delete removes all rules for the given user+agent pair on uninstall.
	Delete(ctx context.Context, userID, agentID string) error

	// ListByActorLogin returns all installed rules for the given GitHub username.
	ListByActorLogin(ctx context.Context, actorLogin string) ([]InstalledRule, error)

	// LookupUserID returns the Supabase UUID for a given GitHub actor_login.
	LookupUserID(ctx context.Context, actorLogin string) (string, error)

	// RecordIssued inserts an issued symbol record.
	RecordIssued(ctx context.Context, userID, agentID, symbolID string) error
}

// PostgresInstalledRuleStore implements InstalledRuleStore backed by Postgres.
type PostgresInstalledRuleStore struct {
	pool *pgxpool.Pool
}

// NewPostgresInstalledRuleStore opens a pgx connection pool to the given DSN.
func NewPostgresInstalledRuleStore(ctx context.Context, dsn string) (*PostgresInstalledRuleStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("overseer: connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("overseer: ping postgres: %w", err)
	}
	return &PostgresInstalledRuleStore{pool: pool}, nil
}

// Close releases the connection pool.
func (s *PostgresInstalledRuleStore) Close() {
	s.pool.Close()
}

// Migrate creates the installed_rules and issued_symbols tables if they do not exist.
func (s *PostgresInstalledRuleStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS installed_rules (
			user_id      TEXT NOT NULL,
			actor_login  TEXT NOT NULL DEFAULT '',
			agent_id     TEXT NOT NULL,
			symbol_id    TEXT NOT NULL,
			sql_path     TEXT NOT NULL,
			type         TEXT NOT NULL,
			window_hours INT NOT NULL DEFAULT 0,
			installed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (user_id, agent_id, symbol_id)
		);
		ALTER TABLE installed_rules ADD COLUMN IF NOT EXISTS actor_login TEXT NOT NULL DEFAULT '';

		CREATE TABLE IF NOT EXISTS issued_symbols (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     TEXT NOT NULL,
			agent_id    TEXT NOT NULL,
			symbol_id   TEXT NOT NULL,
			issued_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("overseer: migrate: %w", err)
	}
	return nil
}

// Save upserts a batch of rules for a user+agent install.
func (s *PostgresInstalledRuleStore) Save(ctx context.Context, rules []InstalledRule) error {
	for _, r := range rules {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO installed_rules (user_id, actor_login, agent_id, symbol_id, sql_path, type, window_hours, installed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id, agent_id, symbol_id) DO UPDATE
				SET actor_login  = EXCLUDED.actor_login,
				    sql_path     = EXCLUDED.sql_path,
				    type         = EXCLUDED.type,
				    window_hours = EXCLUDED.window_hours,
				    installed_at = EXCLUDED.installed_at
		`, r.UserID, r.ActorLogin, r.AgentID, r.SymbolID, r.SQLPath, r.Type, r.WindowHours, r.InstalledAt)
		if err != nil {
			return fmt.Errorf("Save installed rule %s/%s: %w", r.AgentID, r.SymbolID, err)
		}
	}
	return nil
}

// Delete removes all installed rules for the given user+agent pair.
func (s *PostgresInstalledRuleStore) Delete(ctx context.Context, userID, agentID string) error {
	_, err := s.pool.Exec(ctx,
		"DELETE FROM installed_rules WHERE user_id = $1 AND agent_id = $2",
		userID, agentID,
	)
	return err
}

// ListByActorLogin returns all installed rules for the given GitHub username.
func (s *PostgresInstalledRuleStore) ListByActorLogin(ctx context.Context, actorLogin string) ([]InstalledRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT user_id, actor_login, agent_id, symbol_id, sql_path, type, window_hours, installed_at
		FROM installed_rules
		WHERE actor_login = $1
	`, actorLogin)
	if err != nil {
		return nil, fmt.Errorf("ListByActorLogin: %w", err)
	}
	defer rows.Close()

	var rules []InstalledRule
	for rows.Next() {
		var r InstalledRule
		if err := rows.Scan(&r.UserID, &r.ActorLogin, &r.AgentID, &r.SymbolID, &r.SQLPath, &r.Type, &r.WindowHours, &r.InstalledAt); err != nil {
			return nil, fmt.Errorf("ListByActorLogin scan: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// RecordIssued inserts an issued symbol record.
func (s *PostgresInstalledRuleStore) RecordIssued(ctx context.Context, userID, agentID, symbolID string) error {
	_, err := s.pool.Exec(ctx,
		"INSERT INTO issued_symbols (user_id, agent_id, symbol_id) VALUES ($1, $2, $3)",
		userID, agentID, symbolID,
	)
	return err
}

// LookupUserID returns the Supabase UUID for a given GitHub actor_login.
func (s *PostgresInstalledRuleStore) LookupUserID(ctx context.Context, actorLogin string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx,
		"SELECT user_id FROM installed_rules WHERE actor_login = $1 LIMIT 1",
		actorLogin,
	).Scan(&userID)
	return userID, err
}
