# symblon

> Developer achievements rooted in real activity.

Symblon watches your GitHub activity and issues **symbols** — achievements with full context: which org, which repo, which moment. Collect them, share them, build a profile that actually means something.

**[symblon.cc](https://symblon.cc)** — early development, not yet open for signups.

---

## What is a symbol?

A symbol is an achievement issued when specific activity criteria are met. It carries metadata — the org, the repo, the agent that issued it — so it tells a real story rather than just "you did a thing."

Symbols have themes:

| Theme | Look |
|---|---|
| Scouts | Badges |
| Wizards | Spells |
| Hackers | Hacks |

They can have a **multiplier** (e.g. ×2 streak bonus) and a **limit** (max times earnable).

---

## How it works

1. Install the GitHub App and connect your repos
2. Symblon ingests your activity via webhooks
3. Symbol agents evaluate criteria in real-time or over time windows
4. Symbols land on your profile with full context attached

### Symbol agents

Agents are the programs that evaluate criteria and issue symbols. They can be:

- **Real-time** — triggered immediately on activity (e.g. PR merged)
- **Temporal** — evaluate a condition over a time window with a quantifier (e.g. close 5 bug issues within 48h)

Anyone can register agents. Agents can be public or private.

```yaml
agent:
  type: temporal
  name: "Bug Squasher"
  rules:
    - description: "Close 5 bug-labelled issues within 48h."
      symbol: "/bug-squasher/main"
      window_hours: 48
      quantifier: 5
```

---

## Usage models

**Personal** — connect GitHub, configure repos, earn symbols from your activity everywhere.

**Organization** — create a Symblon org, link GitHub orgs as data sources, issue symbols to contributors. Org pays; members inherit access.

---

## Architecture

```
undercover/          # backend monorepo (Go)
  cmd/ingestor/      # receives GitHub webhooks, publishes to RabbitMQ
  cmd/paquetier/     # consumes events, writes Hive-partitioned Parquet to S3
  cmd/overseer/      # symbol agent registry + evaluation engine
  pkg/event/         # canonical WebhookEvent type

frontcover/          # web app (SvelteKit)
pagelander/          # landing page (Astro)
```

Events flow: `GitHub webhook → ingestor → RabbitMQ → paquetier → S3 (Parquet)`  
Evaluation: `overseer reads Parquet via DuckDB, evaluates SQL-based agent rules`

### Running locally

```sh
cd undercover
cp .env.example .env   # fill in S3 + GitHub App credentials
docker compose up
```

Requires Docker. S3-compatible storage needed for paquetier (Supabase Storage works).

---

## Status

Early development. Core pipeline is working. Web app and agent evaluation in progress.

Discussions for symbol ideas and feedback: [github.com/secersh/symblon/discussions](https://github.com/secersh/symblon/discussions)
