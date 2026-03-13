# ADR-002: Multi-provider integration architecture

**Date:** 2026-03-12
**Status:** Accepted

---

## Context

Symblon's core value is recording developer activity and issuing symbols based on
that activity. Initially the only supported activity source is GitHub (via GitHub
App). However, other code hosting platforms (GitLab, Bitbucket, etc.) represent
potential future data sources with similar event models.

---

## Decision

Design the integration layer as a **provider abstraction** from day one, even
though only GitHub will be implemented initially.

- The Settings UI groups all integrations under a single "Integrations" section,
  showing a card per provider (GitHub live, others as "coming soon")
- Each provider card has a consistent shape: logo, name, status, connection state,
  and action (install / manage / coming soon)
- Internally, `github_installation_id` in user metadata is the GitHub-specific
  field; future providers will add their own equivalent fields
- The event pipeline (`undercover/ingestor`) already normalizes events into
  `WebhookEvent` — the normalizer layer is where per-provider logic lives

**Provider roadmap (planning only, not scheduled):**
1. GitHub — implemented
2. GitLab — future
3. Bitbucket — future

---

## Consequences

- Settings page is structured as a provider grid, not a single-provider page
- Adding a new provider requires: a new normalizer in `undercover`, a new card in
  the Integrations settings section, and a new connection field in user metadata
- No shared OAuth/App callback infrastructure is needed now — each provider has
  its own callback route under `/settings/{provider}/callback`

---

## Alternatives considered

### Build GitHub-only settings, refactor later
Rejected. The refactor cost is low now and the provider-card UI pattern is
cleaner and more scalable than a single flat settings panel.
