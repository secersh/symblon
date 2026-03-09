package agent

import "time"

// Scope controls who can use an agent and which actor set is evaluated.
type Scope string

const (
	// ScopeGlobal agents are built-in and available to everyone.  OwnerID is empty.
	ScopeGlobal Scope = "global"
	// ScopeOrg agents belong to a specific organisation.  OwnerID is the org login.
	ScopeOrg Scope = "org"
	// ScopeUser agents are published by an individual.  OwnerID is the user login.
	ScopeUser Scope = "user"
)

// Agent is the definition of an achievement rule.
//
// The Rule.Query is a DuckDB SQL statement executed by overseer's evaluation
// engine.  Named parameters (:actor_login, :repo_owner, :repo_name,
// :event_type, :ref, :action) are bound from the triggering WebhookEvent at
// evaluation time.  The condition is considered satisfied when the query
// returns at least one row.
//
// Example rule query:
//
//	SELECT COUNT(*) AS cnt
//	FROM read_parquet('/data/provider=github/**/*.parquet', hive_partitioning=true)
//	WHERE actor_login = :actor_login
//	  AND event_type  = 'issues'
//	  AND action      = 'closed'
//	  AND received_at >= NOW() - INTERVAL 48 HOURS
//	HAVING cnt >= 5
type Agent struct {
	// ID is the unique identifier for this agent.
	ID string `json:"id"`

	// Name is a human-readable label shown in the UI.
	Name string `json:"name"`

	// Description explains what the agent detects.
	Description string `json:"description"`

	// Scope controls visibility and who the rule is evaluated against.
	Scope Scope `json:"scope"`

	// OwnerID is the org or user login for non-global agents.  Empty for global.
	OwnerID string `json:"owner_id,omitempty"`

	// SymbolPath is the path of the symbol granted when the rule resolves.
	SymbolPath string `json:"symbol_path"`

	// Trigger describes which exchange messages wake up this agent.
	Trigger Trigger `json:"trigger"`

	// Rule is the declarative SQL evaluation logic.
	Rule Rule `json:"rule"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Trigger describes which RabbitMQ routing key pattern wakes up this agent.
type Trigger struct {
	// RoutingKey is an AMQP topic pattern (e.g. "activity.github").
	// Overseer subscribes to "activity.#" and routes each received event to
	// every agent whose RoutingKey matches the event's actual routing key.
	RoutingKey string `json:"routing_key"`
}

// Rule holds the declarative evaluation logic for an agent.
type Rule struct {
	// Query is a DuckDB SQL statement with named parameters bound from the
	// triggering WebhookEvent.  An empty result set means "not yet satisfied".
	Query string `json:"query"`
}
