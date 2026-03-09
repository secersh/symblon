// Package normalize provides functions that map provider-specific webhook
// payloads into the provider-agnostic [event.WebhookEvent] type.
package normalize

import (
	"encoding/json"
	"time"
	"undercover/pkg/event"
)

// githubPayload captures the common fields present in most GitHub webhook
// payloads.  Fields are optional because different event types carry different
// subsets of these.
type githubPayload struct {
	Action string `json:"action"`
	Sender struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"sender"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	// Ref is present on push events.
	Ref string `json:"ref"`
}

// GitHub maps a raw GitHub webhook payload into a WebhookEvent.
//
//   - eventType  comes from the X-GitHub-Event header.
//   - deliveryID comes from the X-GitHub-Delivery header.
//   - body       is the raw JSON request body.
func GitHub(eventType, deliveryID string, body []byte) (*event.WebhookEvent, error) {
	var p githubPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}

	return &event.WebhookEvent{
		ID:         deliveryID,
		Provider:   event.ProviderGitHub,
		EventType:  eventType,
		Action:     p.Action,
		ActorLogin: p.Sender.Login,
		ActorID:    p.Sender.ID,
		RepoOwner:  p.Repository.Owner.Login,
		RepoName:   p.Repository.Name,
		Ref:        p.Ref,
		ReceivedAt: time.Now().UTC(),
		Payload:    json.RawMessage(body),
	}, nil
}
