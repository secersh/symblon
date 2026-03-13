package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"undercover/internal/overseer/agent"
	"undercover/internal/overseer/evaluation"
	"undercover/internal/overseer/store"
	"undercover/pkg/event"
	"undercover/pkg/messaging"
)

const (
	queueName  = "overseer"
	routingKey = "activity.#"
)

// Evaluator runs an agent's SQL rule against the Parquet data store and
// reports whether the condition is satisfied.
//
// The stub implementation (StubEvaluator) is used until DuckDB/Parquet
// persistence is wired up.  Replace it with a DuckDBEvaluator that executes
// agent.Rule.Query with named parameters bound from the WebhookEvent.
type Evaluator interface {
	Evaluate(ctx context.Context, a *agent.Agent, evt *event.WebhookEvent) (bool, error)
}

// StubEvaluator is a placeholder that logs the evaluation and always returns
// false (condition not yet met).  It will be replaced by a real DuckDB-backed
// implementation once the Parquet store is accessible from overseer.
//
// TODO(overseer): replace with DuckDBEvaluator that:
//  1. Opens a DuckDB in-process connection
//  2. Binds :actor_login, :repo_owner, :repo_name, :event_type, :ref, :action
//     from the triggering WebhookEvent
//  3. Executes agent.Rule.Query via read_parquet with hive_partitioning=true
//  4. Returns true when the result set is non-empty
type StubEvaluator struct{}

func (StubEvaluator) Evaluate(_ context.Context, a *agent.Agent, evt *event.WebhookEvent) (bool, error) {
	log.Printf("[trigger] stub evaluate — agent=%s actor=%s rule=%q", a.ID, evt.ActorLogin, a.Rule.Query)
	return false, nil
}

// ResolvedEvent is the message published to RabbitMQ when an agent resolves.
// Downstream services (symbol issuance, frontcover) subscribe to
// "agent.resolved.#" to act on this.
type ResolvedEvent struct {
	AgentID      string    `json:"agent_id"`
	AgentName    string    `json:"agent_name"`
	SymbolPath   string    `json:"symbol_path"`
	ActorLogin   string    `json:"actor_login"`
	OrgLogin     string    `json:"org_login,omitempty"`
	EvaluationID string    `json:"evaluation_id"`
	GrantedAt    time.Time `json:"granted_at"`
}

// Trigger is the RabbitMQ consumer that drives agent evaluation.
//
// For every WebhookEvent received on "activity.#" it:
//  1. Looks up all agents whose trigger routing key matches the event's routing key
//  2. Runs the Evaluator for each matching agent
//  3. Appends an Evaluation record to the store
//  4. If resolved, publishes a ResolvedEvent to "agent.resolved.<agent_id>"
type Trigger struct {
	msgService messaging.MessagingService
	agents     store.AgentStore
	evals      store.EvaluationStore
	evaluator  Evaluator
}

// New creates a Trigger.  Call Start to begin consuming.
func New(
	msgService messaging.MessagingService,
	agents store.AgentStore,
	evals store.EvaluationStore,
	evaluator Evaluator,
) *Trigger {
	return &Trigger{
		msgService: msgService,
		agents:     agents,
		evals:      evals,
		evaluator:  evaluator,
	}
}

// Start binds the overseer queue to "activity.#" and begins consuming messages.
// It returns an error if the queue binding or subscription fails.
func (t *Trigger) Start() error {
	if err := t.msgService.BindQueue(queueName, routingKey); err != nil {
		return fmt.Errorf("bind queue %q: %w", queueName, err)
	}

	return t.msgService.Subscribe(queueName, func(msg string) {
		var evt event.WebhookEvent
		if err := json.Unmarshal([]byte(msg), &evt); err != nil {
			log.Printf("[trigger] unmarshal event: %v", err)
			return
		}
		t.dispatch(context.Background(), &evt)
	})
}

// dispatch fans the event out to every matching agent and evaluates each one.
func (t *Trigger) dispatch(ctx context.Context, evt *event.WebhookEvent) {
	// Derive the routing key from the event (matches the key used by ingestor).
	rk := fmt.Sprintf("activity.%s", evt.Provider)

	agents, err := t.agents.ListByRoutingKey(ctx, rk)
	if err != nil {
		log.Printf("[trigger] list agents for key %q: %v", rk, err)
		return
	}

	for _, a := range agents {
		t.evaluate(ctx, a, evt)
	}
}

// evaluate runs a single agent against the triggering event and records the outcome.
func (t *Trigger) evaluate(ctx context.Context, a *agent.Agent, evt *event.WebhookEvent) {
	now := time.Now().UTC()

	eval := &evaluation.Evaluation{
		ID:             uuid.NewString(),
		AgentID:        a.ID,
		TriggerEventID: evt.ID,
		ActorLogin:     evt.ActorLogin,
		OrgLogin:       evt.RepoOwner,
		TriggeredAt:    now,
	}

	resolved, err := t.evaluator.Evaluate(ctx, a, evt)

	completedAt := time.Now().UTC()
	eval.CompletedAt = &completedAt

	switch {
	case err != nil:
		eval.Status = evaluation.StatusFailed
		eval.Detail = err.Error()
		log.Printf("[trigger] agent=%s failed: %v", a.ID, err)

	case resolved:
		eval.Status = evaluation.StatusResolved
		eval.Detail = fmt.Sprintf("symbol %s granted to %s", a.SymbolPath, evt.ActorLogin)
		t.publishResolved(a, evt, eval.ID, now)

	default:
		eval.Status = evaluation.StatusSkipped
	}

	if err := t.evals.Append(ctx, eval); err != nil {
		log.Printf("[trigger] append evaluation for agent=%s: %v", a.ID, err)
	}
}

// publishResolved emits a ResolvedEvent to "agent.resolved.<agent_id>".
func (t *Trigger) publishResolved(a *agent.Agent, evt *event.WebhookEvent, evalID string, grantedAt time.Time) {
	payload, err := json.Marshal(ResolvedEvent{
		AgentID:      a.ID,
		AgentName:    a.Name,
		SymbolPath:   a.SymbolPath,
		ActorLogin:   evt.ActorLogin,
		OrgLogin:     evt.RepoOwner,
		EvaluationID: evalID,
		GrantedAt:    grantedAt,
	})
	if err != nil {
		log.Printf("[trigger] marshal resolved event for agent=%s: %v", a.ID, err)
		return
	}

	rk := fmt.Sprintf("agent.resolved.%s", a.ID)
	if err := t.msgService.PublishTo(rk, string(payload)); err != nil {
		log.Printf("[trigger] publish resolved event for agent=%s: %v", a.ID, err)
	} else {
		log.Printf("[trigger] resolved — agent=%s actor=%s symbol=%s", a.ID, evt.ActorLogin, a.SymbolPath)
	}
}
