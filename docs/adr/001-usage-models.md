# ADR-001: Personal vs Organization usage models

**Date:** 2026-03-12
**Status:** Accepted

---

## Context

Symblon needs to support two distinct usage modes from day one. The distinction
affects data scoping, billing, agent evaluation, and the UI flows users go through.
This ADR records the decisions made about how each mode works and where they differ.

---

## Decision

### Personal usage model

- A personal account is auto-created when a user signs in via GitHub OAuth.
- Symbols are earned from **any GitHub org or repo** where the Symblon GitHub App
  is installed and the user has recorded activity. Personal usage is not limited
  to personal/private repos.
- Personal agents evaluate activity scoped to the user's configured GitHub access
  (repos the GitHub App installation covers).
- Symbols carry `org` and `repo` metadata. On the user's profile, symbols can be
  viewed flat (default), grouped by org, or grouped by org + repo.
- Symblon organization membership count is shown on the profile only when > 0.
- New users who have not installed the GitHub App or configured repo access are
  shown an integration guide instead of an empty dashboard.

### Organization usage model

- Organizations are created **explicitly** by a user — no auto-creation.
- A Symblon organization can optionally link one or more GitHub organizations as
  data sources. This uses the same GitHub App and is fully opt-in.
- Every Symblon organization must define a **membership badge**. This badge is
  automatically issued to a user when they join the organization.
- Org agents evaluate activity of all members across all linked GitHub orgs/repos.
- Billing is at the org level; members inherit access.
- Role hierarchy: owner → admin → member.

### Key distinction

Personal answers: *"What have I achieved across everything I contribute to?"*
Organization answers: *"What has our org chosen to recognize its contributors for?"*

---

## Consequences

- The GitHub App is the central integration point for both modes. Without it,
  neither real-time nor temporal symbols can be issued.
- Symbols must store `{ github_org, github_repo }` metadata at issuance time to
  support profile grouping.
- Temporal evaluation is always scoped to `(symblon_org_or_user, github_user)` —
  a user must be evaluated within a context, even in personal mode (context = their
  personal account).
- The "membership badge" requirement means org creation has a mandatory step before
  the org is considered active.
- Multiple GitHub orgs mapping to one Symblon org means the event pipeline
  (`undercover/ingestor`) must tag events with the Symblon org ID, not just the
  GitHub org name.

---

## Alternatives considered

### Auto-create Symblon orgs from GitHub org membership
Rejected. Creates noise — users would end up in Symblon orgs they never intended
to use. Explicit creation keeps intent clear and billing intentional.

### Hard-couple Symblon orgs 1:1 with GitHub orgs
Rejected. A company may have multiple GitHub orgs (e.g. main org + open-source org)
and want a single Symblon org to span them. Many-to-one is more flexible.

### Separate GitHub Apps for personal vs org
Rejected. One App simplifies installation UX and permission management.
