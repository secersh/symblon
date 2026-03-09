package store

import (
	"context"

	"overseer/internal/overseer/agent"
	"overseer/internal/overseer/evaluation"
)

// AgentStore is the read/write interface for agent definitions.
//
// Implementations:
//   - InMemoryAgentStore  — development / tests
//   - (future) EventStoreDB stream per agent ("agents-<id>")
//   - (future) Postgres agents table with JSONB rule column
type AgentStore interface {
	// Save creates or updates an agent definition.
	Save(ctx context.Context, a *agent.Agent) error

	// Get returns the agent with the given ID.
	Get(ctx context.Context, id string) (*agent.Agent, error)

	// List returns all registered agents.
	List(ctx context.Context) ([]*agent.Agent, error)

	// Delete removes an agent definition.
	Delete(ctx context.Context, id string) error

	// ListByRoutingKey returns all agents whose trigger routing key matches the
	// given key.  Used by the trigger consumer to fan-out incoming events.
	ListByRoutingKey(ctx context.Context, routingKey string) ([]*agent.Agent, error)
}

// EvaluationStore is an append-only log of agent evaluation outcomes.
//
// Implementations:
//   - InMemoryEvaluationStore — development / tests
//   - (future) EventStoreDB persistent subscription / stream per agent
//   - (future) Postgres agent_evaluations append-only table
type EvaluationStore interface {
	// Append records a new evaluation.  Existing records must never be mutated.
	Append(ctx context.Context, e *evaluation.Evaluation) error

	// ListByAgent returns all evaluations for the given agent, newest first.
	ListByAgent(ctx context.Context, agentID string) ([]*evaluation.Evaluation, error)

	// ListByActor returns all evaluations for the given actor login, newest first.
	ListByActor(ctx context.Context, actorLogin string) ([]*evaluation.Evaluation, error)
}
