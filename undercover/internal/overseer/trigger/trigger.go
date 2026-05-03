package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"undercover/internal/overseer/evaluation"
	"undercover/internal/overseer/store"
	"undercover/pkg/event"
	"undercover/pkg/messaging"
)

const (
	queueName  = "overseer.parquet"
	routingKey = "parquet.flushed"
)

// Evaluator runs an installed rule's SQL against the Parquet data store and
// reports whether the condition is satisfied for the triggering actor.
//
// StubEvaluator is used until the DuckDB/Parquet integration is wired up.
// Replace with a DuckDBEvaluator that fetches the SQL from S3, executes it
// via read_parquet with hive_partitioning=true, and returns true when the
// result set is non-empty.
type Evaluator interface {
	Evaluate(ctx context.Context, rule *store.InstalledRule, evt *event.WebhookEvent) (bool, error)
}

// StubEvaluator logs the evaluation attempt and always returns false.
// TODO: replace with DuckDBEvaluator once S3/Parquet integration is ready.
type StubEvaluator struct{}

func (StubEvaluator) Evaluate(_ context.Context, rule *store.InstalledRule, evt *event.WebhookEvent) (bool, error) {
	log.Printf("[trigger] stub evaluate — agent=%s symbol=%s actor=%s sql=%s",
		rule.AgentID, rule.SymbolID, evt.ActorLogin, rule.SQLPath)
	return false, nil
}

// ResolvedEvent is published to "agent.resolved.<agent_id>" when a symbol is granted.
type ResolvedEvent struct {
	AgentID      string    `json:"agent_id"`
	SymbolID     string    `json:"symbol_id"`
	ActorLogin   string    `json:"actor_login"`
	OrgLogin     string    `json:"org_login,omitempty"`
	EvaluationID string    `json:"evaluation_id"`
	GrantedAt    time.Time `json:"granted_at"`
}

// Trigger is the RabbitMQ consumer that drives agent evaluation.
//
// For every WebhookEvent received on "activity.#" it:
//  1. Looks up all installed rules for the event's actor
//  2. Runs the Evaluator for each rule
//  3. Appends an Evaluation record to the store
//  4. If resolved, publishes a ResolvedEvent to "agent.resolved.<agent_id>"
type Trigger struct {
	msgService messaging.MessagingService
	rules      store.InstalledRuleStore
	evals      store.EvaluationStore
	evaluator  Evaluator
}

// New creates a Trigger. Call Start to begin consuming.
func New(
	msgService messaging.MessagingService,
	rules store.InstalledRuleStore,
	evals store.EvaluationStore,
	evaluator Evaluator,
) *Trigger {
	return &Trigger{
		msgService: msgService,
		rules:      rules,
		evals:      evals,
		evaluator:  evaluator,
	}
}

// Start binds the overseer queue to "parquet.flushed" and begins consuming messages.
func (t *Trigger) Start() error {
	if err := t.msgService.BindQueue(queueName, routingKey); err != nil {
		return fmt.Errorf("bind queue %q: %w", queueName, err)
	}

	return t.msgService.Subscribe(queueName, func(msg string) {
		var flushed messaging.ParquetFlushedEvent
		if err := json.Unmarshal([]byte(msg), &flushed); err != nil {
			log.Printf("[trigger] unmarshal parquet.flushed: %v", err)
			return
		}
		ctx := context.Background()
		for _, actor := range flushed.Actors {
			t.dispatch(ctx, actor)
		}
	})
}

// dispatch fans out to every installed rule for the given actor.
func (t *Trigger) dispatch(ctx context.Context, actorLogin string) {
	rules, err := t.rules.ListByActorLogin(ctx, actorLogin)
	if err != nil {
		log.Printf("[trigger] list rules for actor=%s: %v", actorLogin, err)
		return
	}

	for i := range rules {
		t.evaluate(ctx, &rules[i], actorLogin)
	}
}

// evaluate runs a single installed rule for the given actor.
func (t *Trigger) evaluate(ctx context.Context, rule *store.InstalledRule, actorLogin string) {
	now := time.Now().UTC()

	eval := &evaluation.Evaluation{
		ID:          uuid.NewString(),
		AgentID:     rule.AgentID,
		ActorLogin:  actorLogin,
		TriggeredAt: now,
	}

	evt := &event.WebhookEvent{ActorLogin: actorLogin}
	resolved, err := t.evaluator.Evaluate(ctx, rule, evt)

	completedAt := time.Now().UTC()
	eval.CompletedAt = &completedAt

	switch {
	case err != nil:
		eval.Status = evaluation.StatusFailed
		eval.Detail = err.Error()
		log.Printf("[trigger] agent=%s symbol=%s failed: %v", rule.AgentID, rule.SymbolID, err)

	case resolved:
		eval.Status = evaluation.StatusResolved
		eval.Detail = fmt.Sprintf("symbol %s/%s granted to %s", rule.AgentID, rule.SymbolID, actorLogin)
		t.publishResolved(rule, actorLogin, eval.ID, now)

	default:
		eval.Status = evaluation.StatusSkipped
	}

	if err := t.evals.Append(ctx, eval); err != nil {
		log.Printf("[trigger] append evaluation for agent=%s: %v", rule.AgentID, err)
	}
}

// publishResolved emits a ResolvedEvent to "agent.resolved.<agent_id>".
func (t *Trigger) publishResolved(rule *store.InstalledRule, actorLogin string, evalID string, grantedAt time.Time) {
	payload, err := json.Marshal(ResolvedEvent{
		AgentID:      rule.AgentID,
		SymbolID:     rule.SymbolID,
		ActorLogin:   actorLogin,
		EvaluationID: evalID,
		GrantedAt:    grantedAt,
	})
	if err != nil {
		log.Printf("[trigger] marshal resolved event for agent=%s: %v", rule.AgentID, err)
		return
	}

	rk := fmt.Sprintf("agent.resolved.%s", rule.AgentID)
	if err := t.msgService.PublishTo(rk, string(payload)); err != nil {
		log.Printf("[trigger] publish resolved for agent=%s: %v", rule.AgentID, err)
	} else {
		log.Printf("[trigger] resolved — agent=%s symbol=%s actor=%s", rule.AgentID, rule.SymbolID, actorLogin)
	}
}
