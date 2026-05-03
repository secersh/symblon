package store

import (
	"context"
	"time"
)

// Agent is a registered agent version in the registry.
type Agent struct {
	ID            string    `json:"id"`
	Publisher     string    `json:"publisher"`      // UUID — internal identity
	PublisherName string    `json:"publisher_name"` // display name, e.g. GitHub username
	Handle        string    `json:"handle"`
	Version       string    `json:"version"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Visibility    string    `json:"visibility"`
	PricingModel  string    `json:"pricing_model"`
	PriceUSD      *float64  `json:"price_usd,omitempty"`
	PublishedAt   time.Time `json:"published_at"`
}

// Symbol is one symbol belonging to a registered agent.
type Symbol struct {
	ID          string  `json:"id"`
	AgentID     string  `json:"agent_id"`
	SymbolID    string  `json:"symbol_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	WindowHours int     `json:"window_hours,omitempty"`
	SQLPath     string  `json:"sql_path"`
	ImageURL    string  `json:"image_url,omitempty"`
}

// Install records that a user has installed an agent.
type Install struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	AgentID     string    `json:"agent_id"`
	InstalledAt time.Time `json:"installed_at"`
}

// AgentStore is the read/write interface for the agent registry.
type AgentStore interface {
	// SaveAgent persists a newly published agent and its symbols.
	SaveAgent(ctx context.Context, agent *Agent, symbols []Symbol) error

	// GetAgent returns the agent for the given publisher/handle/version.
	GetAgent(ctx context.Context, publisher, handle, version string) (*Agent, error)

	// ListAgents returns all public agents, optionally filtered by publisher.
	ListAgents(ctx context.Context, publisher string) ([]*Agent, error)

	// SymbolsByAgent returns all symbols for the given agent ID.
	SymbolsByAgent(ctx context.Context, agentID string) ([]Symbol, error)

	// SymbolsByAgentIDs returns symbols for multiple agents keyed by agent ID.
	SymbolsByAgentIDs(ctx context.Context, agentIDs []string) (map[string][]Symbol, error)
}

// IssuedSymbol is a symbol that has been granted to a user.
type IssuedSymbol struct {
	ID       string    `json:"id"`
	AgentID  string    `json:"agent_id"`
	SymbolID string    `json:"symbol_id"`
	IssuedAt time.Time `json:"issued_at"`
	// Joined from symbols table
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url,omitempty"`
}

// SymbolStore is the read interface for issued symbols.
type SymbolStore interface {
	ListIssuedSymbols(ctx context.Context, userID string) ([]IssuedSymbol, error)
}

// InstallStore is the read/write interface for user agent installs.
type InstallStore interface {
	// Install records a user installing an agent. Idempotent.
	Install(ctx context.Context, userID, agentID string) (*Install, error)

	// Uninstall removes an install record.
	Uninstall(ctx context.Context, userID, agentID string) error

	// GetInstall returns the install record for a user/agent pair.
	GetInstall(ctx context.Context, userID, agentID string) (*Install, error)

	// ListInstalledAgents returns the full agent records installed by the given user.
	ListInstalledAgents(ctx context.Context, userID string) ([]*Agent, error)
}
