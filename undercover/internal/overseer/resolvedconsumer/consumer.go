package resolvedconsumer

import (
	"context"
	"encoding/json"
	"log"

	"undercover/internal/overseer/store"
	"undercover/internal/overseer/trigger"
	"undercover/pkg/messaging"
)

const queue = "overseer.agent.resolved"

// Consumer subscribes to agent.resolved.# events and writes to issued_symbols.
type Consumer struct {
	rules store.InstalledRuleStore
	mq    messaging.MessagingService
}

// New returns a Consumer wired to the given store and messaging service.
func New(rules store.InstalledRuleStore, mq messaging.MessagingService) *Consumer {
	return &Consumer{rules: rules, mq: mq}
}

// Start binds the queue and begins consuming resolved events.
func (c *Consumer) Start() error {
	if err := c.mq.BindQueue(queue, messaging.RoutingKeyAgentResolved); err != nil {
		return err
	}
	return c.mq.Subscribe(queue, func(msg string) {
		var evt trigger.ResolvedEvent
		if err := json.Unmarshal([]byte(msg), &evt); err != nil {
			log.Printf("[resolvedconsumer] unmarshal: %v", err)
			return
		}

		ctx := context.Background()

		userID, err := c.rules.LookupUserID(ctx, evt.ActorLogin)
		if err != nil {
			log.Printf("[resolvedconsumer] lookup user for actor=%s: %v", evt.ActorLogin, err)
			return
		}

		if err := c.rules.RecordIssued(ctx, userID, evt.AgentID, evt.SymbolID); err != nil {
			log.Printf("[resolvedconsumer] record issued user=%s agent=%s symbol=%s: %v",
				userID, evt.AgentID, evt.SymbolID, err)
			return
		}

		log.Printf("[resolvedconsumer] issued symbol=%s agent=%s user=%s", evt.SymbolID, evt.AgentID, userID)
	})
}
