package evaluation

import "time"

// Status is the outcome of a single agent evaluation run.
type Status string

const (
	// StatusPending means the evaluation has been queued but not yet run.
	StatusPending Status = "pending"
	// StatusResolved means the rule conditions were satisfied; the symbol is granted.
	StatusResolved Status = "resolved"
	// StatusSkipped means the rule ran successfully but conditions were not met.
	StatusSkipped Status = "skipped"
	// StatusFailed means the rule could not be evaluated due to an error.
	StatusFailed Status = "failed"
)

// Evaluation is an immutable record of one agent run triggered by a WebhookEvent.
//
// In production this is appended to an event store (EventStoreDB stream or
// Postgres append-only table).  The in-memory store is used during development.
// Naming note: "Evaluation" reflects the overseer's role — it evaluates whether
// an agent's conditions have been met — rather than the more generic "execution".
type Evaluation struct {
	// ID is the unique identifier for this evaluation.
	ID string `json:"id"`

	// AgentID is the agent that was evaluated.
	AgentID string `json:"agent_id"`

	// TriggerEventID is the WebhookEvent.ID that caused this evaluation.
	TriggerEventID string `json:"trigger_event_id"`

	// ActorLogin is the GitHub/provider login of the user being evaluated.
	ActorLogin string `json:"actor_login"`

	// OrgLogin is set when the trigger event belongs to an org context.
	OrgLogin string `json:"org_login,omitempty"`

	// TriggeredAt is when the evaluation was initiated.
	TriggeredAt time.Time `json:"triggered_at"`

	// CompletedAt is when the rule finished running.  Nil while pending.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Status is the outcome of the evaluation.
	Status Status `json:"status"`

	// Detail holds the error message on failure or a summary on resolution.
	Detail string `json:"detail,omitempty"`
}
