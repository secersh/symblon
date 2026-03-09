package event

import (
	"encoding/json"
	"fmt"
	"time"
)

// Provider identifies the webhook source system.
type Provider string

const (
	ProviderGitHub Provider = "github"
)

// WebhookEvent is the provider-agnostic representation of any incoming webhook.
// This mirrors undercover/pkg/event.WebhookEvent.
// TODO: extract to a shared types module (e.g. symblon/types) when a Go
// workspace is introduced across the monorepo.
type WebhookEvent struct {
	ID         string          `json:"id"`
	Provider   Provider        `json:"provider"`
	EventType  string          `json:"event_type"`
	Action     string          `json:"action,omitempty"`
	ActorLogin string          `json:"actor_login,omitempty"`
	ActorID    int64           `json:"actor_id,omitempty"`
	RepoOwner  string          `json:"repo_owner,omitempty"`
	RepoName   string          `json:"repo_name,omitempty"`
	Ref        string          `json:"ref,omitempty"`
	ReceivedAt time.Time       `json:"received_at"`
	Payload    json.RawMessage `json:"payload"`
}

// PartitionPath returns the Hive-style path used by paquetier when storing
// this event.  Overseer uses it to construct partition-pruned DuckDB queries.
func (e *WebhookEvent) PartitionPath() string {
	return fmt.Sprintf(
		"provider=%s/year=%d/month=%02d/day=%02d/event_type=%s",
		e.Provider,
		e.ReceivedAt.UTC().Year(),
		int(e.ReceivedAt.UTC().Month()),
		e.ReceivedAt.UTC().Day(),
		e.EventType,
	)
}
