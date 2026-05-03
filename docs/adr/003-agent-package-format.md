# ADR-003: Agent package format

**Date:** 2026-05-03
**Status:** Accepted

---

## Context

Symblon agents are programs that evaluate criteria against GitHub activity and issue
symbols to users. An agent is authored externally (by anyone), registered through a
CLI tool, and distributed via the Symblon marketplace. This ADR defines the canonical
on-disk format for an agent package and the rules for publishing it.

---

## Decision

### Agent package structure

An agent package is a directory named after the agent's handle. It contains a
manifest, SQL rule files, and themed symbol assets.

```
<agent-handle>/
├── agent.yaml
├── rules/
│   └── <symbol-id>.sql       # one file per symbol, filename = symbol id
├── themes/
│   ├── default/              # required; theme-agnostic fallback
│   │   └── <symbol-id>.svg   # filename = symbol id
│   ├── scouts/               # optional named themes
│   │   ├── theme.yaml
│   │   └── <custom-name>.svg
│   ├── wizards/
│   │   ├── theme.yaml
│   │   └── <custom-name>.svg
│   └── hackers/
│       ├── theme.yaml
│       └── <custom-name>.svg
└── README.md                 # repo landing page only, not ingested by platform
```

### Manifest (`agent.yaml`)

```yaml
name: <display name>
handle: <url-safe slug>        # becomes /<publisher>/<handle> in the registry
version: <semver>
description: <short description shown in marketplace>
visibility: public | org       # org-scoped agents only visible to org members
pricing:
  model: free | paid
  price_usd: <amount>          # required when model is paid

symbols:
  - id: <symbol-id>            # used to link rule file and default theme asset
    name: <default display name>
    description: <shown on symbol detail>
    type: realtime | temporal
    window_hours: <int>        # required when type is temporal
```

`publisher` is not declared in the manifest — it is inferred from the authenticated
CLI session at publish time. The publisher must have completed Stripe Connect
onboarding before publishing paid agents.

### Rules

Each symbol has exactly one SQL rule file at `rules/<symbol-id>.sql`. Rules are
DuckDB-compatible SQL queries evaluated against the Parquet event store. The query
must return `actor_login` for the user(s) who satisfy the criteria.

Available columns in the events table:

| Column | Type | Description |
|---|---|---|
| `id` | VARCHAR | Unique delivery ID |
| `provider` | VARCHAR | Event source (github) |
| `event_type` | VARCHAR | Normalized event type (issues, pull_request, push, …) |
| `action` | VARCHAR | Sub-action (opened, closed, merged, …) |
| `actor_login` | VARCHAR | GitHub username who triggered the event |
| `actor_id` | BIGINT | GitHub user ID |
| `repo_owner` | VARCHAR | Org or user owning the repo |
| `repo_name` | VARCHAR | Repository name |
| `received_at` | TIMESTAMP | UTC time the ingestor received the event |
| `payload` | JSON | Full raw provider payload for extended field access |

### Themes

Themes affect the display name and visual asset of a symbol. The `default` theme is
required and serves as the fallback for users with no theme preference.

Named theme folders (`scouts`, `wizards`, `hackers`) are optional but must cover
all symbols declared in `agent.yaml` if present.

Each named theme includes a `theme.yaml` that maps symbol IDs to theme-specific
display names and asset filenames:

```yaml
symbols:
  <symbol-id>:
    name: <theme-specific display name>
    asset: <filename>.svg
```

SVG filenames in named themes can be anything — they are resolved via `theme.yaml`.
SVG filenames in `themes/default/` must match the symbol ID exactly.

### Publishing

Publishing is done via the `symblon` CLI:

```sh
symblon publish ./bug-squasher
```

The CLI:
1. Validates the manifest against the package structure (every symbol ID has a
   matching `.sql` and `themes/default/<id>.svg`; named themes are complete)
2. Uploads all assets to Symblon storage
3. Registers the agent in the registry under `/<publisher>/<handle>/<version>`
4. Returns a stable reference for use in evaluation configuration

### Marketplace and pricing

- **Free agents** are immediately available to any user after install
- **Paid agents** use a pay-per-install model. Symblon charges the installer at
  install time and routes a cut to the publisher's connected Stripe account
- **Org-scoped agents** (`visibility: org`) are only discoverable and installable
  by members of the publishing organization

---

## Consequences

- The CLI is required for publishing — there is no web-based drag-and-drop upload
- The `default` theme is a hard requirement; the CLI rejects packages without it
- Symbol IDs are the stable cross-cutting key connecting manifests, rules, and assets
- Publisher Stripe Connect onboarding is a prerequisite for paid agent publishing
- Agent versioning follows semver; the registry stores all versions and the latest
  is resolved by default

---

## Alternatives considered

### Bundle file format (zip with custom extension)
Rejected. Agents are registered once and referenced by ID — portability of a
self-contained file offers no benefit and adds unpack/validation complexity. A
bundle format would also undermine access control for paid agents.

### Stripe account ID in the manifest
Rejected. Embedding financial identifiers in a public file is a security and
maintenance concern. Publisher payment details are stored at the account level,
not the package level.

### README as platform input
Rejected. The manifest already provides `name`, `description`, and per-symbol
descriptions. Extended documentation belongs on the marketplace listing page.
The README is retained solely as a GitHub repo landing page.
