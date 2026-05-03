package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements AgentStore and InstallStore backed by Postgres.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore opens a pgx connection pool to the given DSN.
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("registrar: connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("registrar: ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Close releases the connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}

// SaveAgent persists a newly published agent and its symbols in a transaction.
func (s *PostgresStore) SaveAgent(ctx context.Context, agent *Agent, symbols []Symbol) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("SaveAgent: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO agents (publisher, publisher_name, handle, version, name, description, visibility, pricing_model, price_usd)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (publisher, handle, version) DO NOTHING
		RETURNING id, published_at
	`, agent.Publisher, agent.PublisherName, agent.Handle, agent.Version, agent.Name,
		agent.Description, agent.Visibility, agent.PricingModel, agent.PriceUSD,
	).Scan(&agent.ID, &agent.PublishedAt)
	if err != nil {
		return fmt.Errorf("SaveAgent: insert agent: %w", err)
	}

	for i := range symbols {
		symbols[i].AgentID = agent.ID
		err = tx.QueryRow(ctx, `
			INSERT INTO symbols (agent_id, symbol_id, name, description, type, window_hours, sql_path, image_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`, symbols[i].AgentID, symbols[i].SymbolID, symbols[i].Name,
			symbols[i].Description, symbols[i].Type, symbols[i].WindowHours, symbols[i].SQLPath, symbols[i].ImageURL,
		).Scan(&symbols[i].ID)
		if err != nil {
			return fmt.Errorf("SaveAgent: insert symbol %q: %w", symbols[i].SymbolID, err)
		}
	}

	return tx.Commit(ctx)
}

// GetAgent returns the agent for the given publisher/handle/version.
func (s *PostgresStore) GetAgent(ctx context.Context, publisher, handle, version string) (*Agent, error) {
	a := &Agent{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, publisher, publisher_name, handle, version, name, description, visibility, pricing_model, price_usd, published_at
		FROM agents
		WHERE publisher = $1 AND handle = $2 AND version = $3
	`, publisher, handle, version).Scan(
		&a.ID, &a.Publisher, &a.PublisherName, &a.Handle, &a.Version, &a.Name,
		&a.Description, &a.Visibility, &a.PricingModel, &a.PriceUSD, &a.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetAgent: %w", err)
	}
	return a, nil
}

// ListAgents returns the latest version of each public agent, optionally filtered by publisher.
func (s *PostgresStore) ListAgents(ctx context.Context, publisher string) ([]*Agent, error) {
	query := `
		SELECT DISTINCT ON (publisher, handle)
			id, publisher, publisher_name, handle, version, name, description, visibility, pricing_model, price_usd, published_at
		FROM agents
		WHERE visibility = 'public'
	`
	args := []any{}
	if publisher != "" {
		query += " AND publisher = $1"
		args = append(args, publisher)
	}
	query += " ORDER BY publisher, handle, published_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListAgents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		a := &Agent{}
		if err := rows.Scan(
			&a.ID, &a.Publisher, &a.PublisherName, &a.Handle, &a.Version, &a.Name,
			&a.Description, &a.Visibility, &a.PricingModel, &a.PriceUSD, &a.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("ListAgents: scan: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// SymbolsByAgentIDs returns symbols for multiple agents keyed by agent ID.
func (s *PostgresStore) SymbolsByAgentIDs(ctx context.Context, agentIDs []string) (map[string][]Symbol, error) {
	if len(agentIDs) == 0 {
		return map[string][]Symbol{}, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, agent_id, symbol_id, name, description, type, window_hours, sql_path, image_url
		FROM symbols
		WHERE agent_id = ANY($1)
	`, agentIDs)
	if err != nil {
		return nil, fmt.Errorf("SymbolsByAgentIDs: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]Symbol)
	for rows.Next() {
		var sym Symbol
		if err := rows.Scan(
			&sym.ID, &sym.AgentID, &sym.SymbolID, &sym.Name,
			&sym.Description, &sym.Type, &sym.WindowHours, &sym.SQLPath, &sym.ImageURL,
		); err != nil {
			return nil, fmt.Errorf("SymbolsByAgentIDs: scan: %w", err)
		}
		result[sym.AgentID] = append(result[sym.AgentID], sym)
	}
	return result, rows.Err()
}

// SymbolsByAgent returns all symbols for the given agent ID.
func (s *PostgresStore) SymbolsByAgent(ctx context.Context, agentID string) ([]Symbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, agent_id, symbol_id, name, description, type, window_hours, sql_path, image_url
		FROM symbols
		WHERE agent_id = $1
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("SymbolsByAgent: %w", err)
	}
	defer rows.Close()

	var symbols []Symbol
	for rows.Next() {
		var sym Symbol
		if err := rows.Scan(
			&sym.ID, &sym.AgentID, &sym.SymbolID, &sym.Name,
			&sym.Description, &sym.Type, &sym.WindowHours, &sym.SQLPath, &sym.ImageURL,
		); err != nil {
			return nil, fmt.Errorf("SymbolsByAgent: scan: %w", err)
		}
		symbols = append(symbols, sym)
	}
	return symbols, rows.Err()
}

// Install records a user installing an agent. Idempotent; clears removed_at on reinstall.
func (s *PostgresStore) Install(ctx context.Context, userID, agentID string) (*Install, error) {
	inst := &Install{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO installs (user_id, agent_id, installed_at, removed_at)
		VALUES ($1, $2, now(), NULL)
		ON CONFLICT (user_id, agent_id) DO UPDATE
			SET installed_at = now(), removed_at = NULL
		RETURNING id, user_id, agent_id, installed_at
	`, userID, agentID).Scan(&inst.ID, &inst.UserID, &inst.AgentID, &inst.InstalledAt)
	if err != nil {
		return nil, fmt.Errorf("Install: %w", err)
	}
	return inst, nil
}

// Uninstall soft-deletes an install record by setting removed_at.
func (s *PostgresStore) Uninstall(ctx context.Context, userID, agentID string) error {
	_, err := s.pool.Exec(ctx,
		"UPDATE installs SET removed_at = now() WHERE user_id = $1 AND agent_id = $2",
		userID, agentID,
	)
	return err
}

// ListInstalledAgents returns active (non-removed) agent records for the given user.
func (s *PostgresStore) ListInstalledAgents(ctx context.Context, userID string) ([]*Agent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.publisher, a.publisher_name, a.handle, a.version, a.name, a.description,
		       a.visibility, a.pricing_model, a.price_usd, a.published_at
		FROM installs i
		JOIN agents a ON a.id = i.agent_id
		WHERE i.user_id = $1 AND i.removed_at IS NULL
		ORDER BY i.installed_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListInstalledAgents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		a := &Agent{}
		if err := rows.Scan(
			&a.ID, &a.Publisher, &a.PublisherName, &a.Handle, &a.Version, &a.Name,
			&a.Description, &a.Visibility, &a.PricingModel, &a.PriceUSD, &a.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("ListInstalledAgents: scan: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

// ListOwnedAgentKeys returns "publisher/handle" for every agent the user has ever installed.
func (s *PostgresStore) ListOwnedAgentKeys(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT a.publisher || '/' || a.handle
		FROM installs i
		JOIN agents a ON a.id = i.agent_id
		WHERE i.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedAgentKeys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("ListOwnedAgentKeys scan: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetInstall returns the install record for a user/agent pair.
func (s *PostgresStore) GetInstall(ctx context.Context, userID, agentID string) (*Install, error) {
	inst := &Install{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, agent_id, installed_at
		FROM installs
		WHERE user_id = $1 AND agent_id = $2
	`, userID, agentID).Scan(&inst.ID, &inst.UserID, &inst.AgentID, &inst.InstalledAt)
	if err != nil {
		return nil, fmt.Errorf("GetInstall: %w", err)
	}
	return inst, nil
}

// ListIssuedSymbols returns all symbols issued to the given user, joined with symbol metadata.
func (s *PostgresStore) ListIssuedSymbols(ctx context.Context, userID string) ([]IssuedSymbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.agent_id, i.symbol_id, i.issued_at,
		       s.name, s.description, s.image_url
		FROM issued_symbols i
		JOIN symbols s ON s.symbol_id = i.symbol_id AND s.agent_id = (
			SELECT a.id FROM agents a
			WHERE a.publisher || '/' || a.handle || '/' || a.version = i.agent_id
			LIMIT 1
		)
		WHERE i.user_id = $1
		ORDER BY i.issued_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListIssuedSymbols: %w", err)
	}
	defer rows.Close()

	var result []IssuedSymbol
	for rows.Next() {
		var s IssuedSymbol
		if err := rows.Scan(&s.ID, &s.AgentID, &s.SymbolID, &s.IssuedAt, &s.Name, &s.Description, &s.ImageURL); err != nil {
			return nil, fmt.Errorf("ListIssuedSymbols scan: %w", err)
		}
		result = append(result, s)
	}
	if result == nil {
		result = []IssuedSymbol{}
	}
	return result, rows.Err()
}

// Migrate runs the schema migrations. Safe to call on every startup.
func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agents (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			publisher      TEXT NOT NULL,
			publisher_name TEXT NOT NULL DEFAULT '',
			handle         TEXT NOT NULL,
			version        TEXT NOT NULL,
			name           TEXT NOT NULL,
			description    TEXT NOT NULL,
			visibility     TEXT NOT NULL CHECK (visibility IN ('public', 'org')),
			pricing_model  TEXT NOT NULL CHECK (pricing_model IN ('free', 'paid')),
			price_usd      NUMERIC(10,2),
			published_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (publisher, handle, version)
		);

		ALTER TABLE agents ADD COLUMN IF NOT EXISTS publisher_name TEXT NOT NULL DEFAULT '';

		CREATE TABLE IF NOT EXISTS symbols (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id     UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			symbol_id    TEXT NOT NULL,
			name         TEXT NOT NULL,
			description  TEXT NOT NULL,
			type         TEXT NOT NULL CHECK (type IN ('realtime', 'temporal')),
			window_hours INT NOT NULL DEFAULT 0,
			sql_path     TEXT NOT NULL,
			image_url    TEXT NOT NULL DEFAULT '',
			UNIQUE (agent_id, symbol_id)
		);

		ALTER TABLE symbols ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

		CREATE TABLE IF NOT EXISTS installs (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id      TEXT NOT NULL,
			agent_id     UUID NOT NULL REFERENCES agents(id),
			installed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			removed_at   TIMESTAMPTZ,
			UNIQUE (user_id, agent_id)
		);
		ALTER TABLE installs ADD COLUMN IF NOT EXISTS removed_at TIMESTAMPTZ;
	`)
	if err != nil {
		return fmt.Errorf("registrar: migrate: %w", err)
	}

	return nil
}
