package event

import (
	"encoding/json"
	"fmt"
	"time"
)

// Provider identifies the webhook source system.
// New providers (e.g. GitLab, Bitbucket) should add a constant here.
type Provider string

const (
	ProviderGitHub Provider = "github"
)

// WebhookEvent is a provider-agnostic representation of any incoming webhook.
// Each provider's normalizer maps its raw payload to this structure so that
// downstream consumers work with a single, stable schema.
type WebhookEvent struct {
	// ID is the unique delivery identifier assigned by the provider.
	ID string `json:"id"`

	// Provider is the source system (github, gitlab, …).
	Provider Provider `json:"provider"`

	// EventType is the normalized event category (push, pull_request, issue, …).
	EventType string `json:"event_type"`

	// Action is the sub-action within an event (opened, closed, merged, …).
	// Empty for events that have no sub-action (e.g. push).
	Action string `json:"action,omitempty"`

	// Actor identifies the user who triggered the event.
	ActorLogin string `json:"actor_login,omitempty"`
	ActorID    int64  `json:"actor_id,omitempty"`

	// Repo identifies the repository the event belongs to.
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`

	// Ref is the git reference affected (push events).
	Ref string `json:"ref,omitempty"`

	// ReceivedAt is the UTC timestamp at which the ingestor received the event.
	ReceivedAt time.Time `json:"received_at"`

	// Payload is the original, unmodified provider payload preserved for
	// audit / re-processing purposes.
	Payload json.RawMessage `json:"payload"`
}

// PartitionPath returns a Hive-style directory path suitable for use as the
// top-level key when writing Parquet files.  Example:
//
//	provider=github/year=2026/month=03/day=09/event_type=push
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
