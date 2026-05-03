package installconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"undercover/internal/overseer/store"
	"undercover/pkg/messaging"
)

const (
	queueInstalled   = "overseer.agent.installed"
	queueUninstalled = "overseer.agent.uninstalled"
)

// Consumer subscribes to agent.installed and agent.uninstalled events from
// the registrar and keeps the overseer's installed_rules table in sync.
type Consumer struct {
	rules store.InstalledRuleStore
	mq    messaging.MessagingService
}

// New returns a Consumer wired to the given store and messaging service.
func New(rules store.InstalledRuleStore, mq messaging.MessagingService) *Consumer {
	return &Consumer{rules: rules, mq: mq}
}

// Start binds the queues and begins consuming. Returns on bind/subscribe error.
func (c *Consumer) Start() error {
	if err := c.mq.BindQueue(queueInstalled, messaging.RoutingKeyAgentInstalled); err != nil {
		return fmt.Errorf("installconsumer: bind installed queue: %w", err)
	}
	if err := c.mq.BindQueue(queueUninstalled, messaging.RoutingKeyAgentUninstalled); err != nil {
		return fmt.Errorf("installconsumer: bind uninstalled queue: %w", err)
	}

	if err := c.mq.Subscribe(queueInstalled, c.handleInstalled); err != nil {
		return fmt.Errorf("installconsumer: subscribe installed: %w", err)
	}
	if err := c.mq.Subscribe(queueUninstalled, c.handleUninstalled); err != nil {
		return fmt.Errorf("installconsumer: subscribe uninstalled: %w", err)
	}

	return nil
}

func (c *Consumer) handleInstalled(msg string) {
	var evt messaging.AgentInstalledEvent
	if err := json.Unmarshal([]byte(msg), &evt); err != nil {
		log.Printf("[installconsumer] unmarshal installed event: %v", err)
		return
	}

	rules := make([]store.InstalledRule, 0, len(evt.Symbols))
	for _, s := range evt.Symbols {
		rules = append(rules, store.InstalledRule{
			UserID:      evt.UserID,
			ActorLogin:  evt.ActorLogin,
			AgentID:     evt.AgentID,
			SymbolID:    s.SymbolID,
			SQLPath:     s.SQLPath,
			Type:        s.Type,
			WindowHours: s.WindowHours,
			InstalledAt: evt.InstalledAt,
		})
	}

	if err := c.rules.Save(context.Background(), rules); err != nil {
		log.Printf("[installconsumer] save rules for user=%s agent=%s: %v", evt.UserID, evt.AgentID, err)
		return
	}

	log.Printf("[installconsumer] indexed %d rules for user=%s agent=%s", len(rules), evt.UserID, evt.AgentID)
}

func (c *Consumer) handleUninstalled(msg string) {
	var evt messaging.AgentUninstalledEvent
	if err := json.Unmarshal([]byte(msg), &evt); err != nil {
		log.Printf("[installconsumer] unmarshal uninstalled event: %v", err)
		return
	}

	if err := c.rules.Delete(context.Background(), evt.UserID, evt.AgentID); err != nil {
		log.Printf("[installconsumer] delete rules for user=%s agent=%s: %v", evt.UserID, evt.AgentID, err)
		return
	}

	log.Printf("[installconsumer] removed rules for user=%s agent=%s", evt.UserID, evt.AgentID)
}
