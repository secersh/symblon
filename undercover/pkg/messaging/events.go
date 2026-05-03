package messaging

import "time"

// Routing key constants for all events on the symblon exchange.
const (
	// RoutingKeyActivityGitHub carries normalised GitHub webhook events
	// published by the ingestor.
	RoutingKeyActivityGitHub = "activity.github"

	// RoutingKeyAgentInstalled is emitted by the registrar when a user
	// installs an agent. Overseer consumes this to populate its local
	// rule index.
	RoutingKeyAgentInstalled = "agent.installed"

	// RoutingKeyAgentUninstalled is emitted by the registrar when a user
	// uninstalls an agent. Overseer consumes this to remove the rules from
	// its local index.
	RoutingKeyAgentUninstalled = "agent.uninstalled"

	// RoutingKeyAgentResolved is emitted by the overseer when a symbol is granted.
	RoutingKeyAgentResolved = "agent.resolved.#"

	// RoutingKeyParquetFlushed is emitted by paquetier after each Parquet flush.
	// Overseer uses this as the trigger to evaluate rules — guaranteeing the
	// flushed events are already readable from S3 before any SQL runs.
	RoutingKeyParquetFlushed = "parquet.flushed"
)

// AgentInstalledEvent is published to RoutingKeyAgentInstalled when a user
// installs an agent via the registrar.
type AgentInstalledEvent struct {
	// UserID is the Supabase UUID of the installer.
	UserID string `json:"user_id"`

	// ActorLogin is the GitHub username — used by the overseer to match events.
	ActorLogin string `json:"actor_login"`

	// AgentID is the fully-qualified agent reference: publisher/handle/version.
	AgentID string `json:"agent_id"`

	// Symbols lists the individual rules overseer must index for this install.
	Symbols []InstalledSymbol `json:"symbols"`

	InstalledAt time.Time `json:"installed_at"`
}

// InstalledSymbol carries the evaluation details overseer needs for one symbol
// within an installed agent.
type InstalledSymbol struct {
	// SymbolID is the stable identifier declared in the agent manifest.
	SymbolID string `json:"symbol_id"`

	// SQLPath is the S3 path to the rule file: agents/<publisher>/<handle>/<version>/rules/<id>.sql
	SQLPath string `json:"sql_path"`

	// Type is realtime or temporal.
	Type string `json:"type"`

	// WindowHours is only set for temporal symbols.
	WindowHours int `json:"window_hours,omitempty"`
}

// ParquetFlushedEvent is published by paquetier after each successful flush.
// Actors lists the distinct actor_logins whose events were written in this batch.
type ParquetFlushedEvent struct {
	FlushedAt time.Time `json:"flushed_at"`
	Actors    []string  `json:"actors"`
}

// AgentUninstalledEvent is published to RoutingKeyAgentUninstalled when a user
// uninstalls an agent. Overseer removes all associated rules from its index.
type AgentUninstalledEvent struct {
	// UserID is the Symblon user identifier.
	UserID string `json:"user_id"`

	// AgentID is the fully-qualified agent reference: publisher/handle/version.
	AgentID string `json:"agent_id"`

	UninstalledAt time.Time `json:"uninstalled_at"`
}
