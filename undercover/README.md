# undercover

**undercover** is the webhook ingestion and storage pipeline for the Symblon platform. It receives raw webhooks from Git hosting providers (GitHub today; GitLab, Bitbucket, etc. in the future), normalises them into a single provider-agnostic event format, streams the events through RabbitMQ, and stores them as Hive-partitioned Parquet files — giving downstream achievement-detection services a fast, queryable data source.

---

## Architecture

```
GitHub ──► POST /api/v1/github
                 │
         ┌───────▼────────┐
         │   ingestor     │  HTTP server (Gin)
         │                │  • normalises raw payload → WebhookEvent
         │                │  • publishes JSON to RabbitMQ topic exchange
         └───────┬────────┘
                 │ exchange: symblon
                 │ routing key: activity.github
         ┌───────▼────────┐
         │   RabbitMQ     │  topic exchange "symblon"
         │   queue:       │  queue "paquetier" bound to "activity.#"
         │   paquetier    │
         └───────┬────────┘
                 │
         ┌───────▼────────┐
         │   paquetier    │  queue consumer
         │                │  • buffers events in DuckDB (in-memory)
         │                │  • flushes Hive-partitioned Parquet every 30 s
         └───────┬────────┘
                 │
         /data/provider=github/year=…/month=…/day=…/event_type=…/
                 │
         ┌───────▼────────┐
         │  Parquet files │  queryable by DuckDB, Spark, Athena, etc.
         └────────────────┘
```

Adding a new provider (e.g. GitLab) requires only:
1. A new normalizer in `internal/ingestor/normalize/`
2. A new `Provider` constant in `pkg/event/event.go`
3. A new HTTP route in `internal/ingestor/router/`

The rest of the pipeline is unchanged.

---

## Project layout

```
undercover/
├── cmd/
│   ├── ingestor/        # binary: HTTP server → RabbitMQ producer
│   └── paquetier/       # binary: queue consumer → Parquet writer
├── internal/
│   ├── ingestor/
│   │   ├── handler/     # per-provider Gin handlers
│   │   ├── normalize/   # raw payload → WebhookEvent mappers
│   │   └── router/      # Gin route registration
│   └── paquetier/
│       └── writer/      # DuckDB-backed Parquet writer
├── pkg/
│   ├── event/           # WebhookEvent type (shared schema)
│   └── messaging/       # RabbitMQ client (shared by both binaries)
├── Dockerfile            # multi-stage: ingestor + paquetier targets
├── docker-compose.yml
├── .dockerignore
└── .env.example          # template for optional env vars
```

---

## Components

### ingestor (`cmd/ingestor`)
An HTTP server built on [Gin](https://github.com/gin-gonic/gin).

| Responsibility | Detail |
|---|---|
| Receive webhooks | `POST /api/v1/github` |
| Normalise payload | `normalize.GitHub()` → `WebhookEvent` |
| Publish to queue | JSON body → RabbitMQ exchange `symblon`, routing key `activity.github` |
| Graceful shutdown | `SIGINT`/`SIGTERM` with a 5 s drain timeout |

**Environment variables**

| Variable | Default | Description |
|---|---|---|
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |

### paquetier (`cmd/paquetier`)
A long-running queue consumer that writes events to Parquet.

| Responsibility | Detail |
|---|---|
| Subscribe | Queue `paquetier`, routing key `activity.#` (all providers) |
| Buffer | In-memory DuckDB table |
| Flush | Hive-partitioned Parquet every 30 s and on shutdown |
| Graceful shutdown | `SIGINT`/`SIGTERM` triggers a final flush before exit |

**Environment variables**

| Variable | Default | Description |
|---|---|---|
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |
| `DATA_DIR` | `/data` | Root directory for Parquet output |

**On-disk layout**

```
/data/
  provider=github/
    year=2026/month=03/day=09/
      event_type=push/
        events_<nanos>_0.parquet
      event_type=pull_request/
        events_<nanos>_0.parquet
```

### pkg/event — shared event model
`WebhookEvent` is the single canonical struct used throughout the pipeline.

| Field | Type | Description |
|---|---|---|
| `id` | string | Provider-assigned delivery ID |
| `provider` | string | Source system (`github`, …) |
| `event_type` | string | Normalised event category (`push`, `pull_request`, …) |
| `action` | string | Sub-action (`opened`, `merged`, …); empty for push |
| `actor_login` / `actor_id` | string / int64 | User who triggered the event |
| `repo_owner` / `repo_name` | string | Repository coordinates |
| `ref` | string | Git ref (push events) |
| `received_at` | timestamp | UTC time the ingestor received the event |
| `payload` | JSON | Original unmodified provider payload |

### pkg/messaging — RabbitMQ client
Declares the durable `symblon` topic exchange on startup (idempotent). Key methods:

| Method | Description |
|---|---|
| `Publish(message)` | Publish a JSON message with routing key `activity.github` |
| `BindQueue(queue, key)` | Declare a durable queue and bind it with an AMQP wildcard pattern |
| `Subscribe(queue, fn)` | Consume messages from a queue in a background goroutine |
| `Close()` | Drain and close the channel + connection |

---

## Running locally

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) ≥ 24 with the Compose plugin

### Start the full stack

```bash
cd undercover
docker compose up --build
```

This starts:

| Service | Port(s) | Notes |
|---|---|---|
| `rabbitmq` | `5672` (AMQP), `15672` (management UI) | `guest`/`guest` credentials |
| `ingestor` | `8080` | Waits for RabbitMQ healthcheck |
| `paquetier` | — | Parquet files written to a named Docker volume (`parquet_data`) |

### Send a test webhook

```bash
curl -X POST http://localhost:8080/api/v1/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: test-delivery-1" \
  -d '{"sender":{"login":"alice","id":1},"repository":{"name":"symblon","owner":{"login":"secersh"}},"ref":"refs/heads/main"}'
```

### Expose the ingestor via Cloudflare Tunnel (optional)

Useful for receiving real GitHub webhooks on a local stack without port-forwarding.

**Quick tunnel** (no credentials — Cloudflare assigns a temporary `*.trycloudflare.com` URL):

```bash
docker compose --profile tunnel up --build
```

Look for the public URL in the `cloudflared` container logs:

```
docker compose logs -f cloudflared
```

**Named tunnel** (persistent URL — requires a [Cloudflare account](https://dash.cloudflare.com)):

1. Create a tunnel in [Zero Trust → Networks → Tunnels](https://one.dash.cloudflare.com)
2. Copy the token from the "Install connector" step
3. Create a `.env` file (see `.env.example`):
   ```
   CLOUDFLARE_TUNNEL_TOKEN=your_token_here
   ```
4. `docker compose --profile tunnel up --build`

---

## Querying Parquet output

Once paquetier has flushed at least one batch, you can query the data directly with DuckDB:

```sql
-- Total events per provider and event type
SELECT provider, event_type, COUNT(*) AS n
FROM read_parquet('/data/**/*.parquet', hive_partitioning = true)
GROUP BY 1, 2
ORDER BY 3 DESC;

-- All push events for a specific day
SELECT actor_login, repo_owner, repo_name, received_at
FROM read_parquet('/data/**/*.parquet', hive_partitioning = true)
WHERE provider = 'github'
  AND event_type = 'push'
  AND year = 2026 AND month = 3 AND day = 9;
```

---

## Adding a new provider

1. Add a `Provider` constant in `pkg/event/event.go`:
   ```go
   const (
       ProviderGitHub Provider = "github"
       ProviderGitLab Provider = "gitlab"   // new
   )
   ```
2. Create a normalizer `internal/ingestor/normalize/gitlab.go` that maps the raw payload to `event.WebhookEvent`
3. Add a handler `internal/ingestor/handler/gitlab.go` (copy `github.go`, swap the normalizer call)
4. Register a route in `internal/ingestor/router/router.go`

Events published with any routing key matching `activity.#` are automatically picked up by paquetier — no changes needed there.

---

## Development without Docker

```bash
# Start RabbitMQ only
docker compose up rabbitmq

# Run ingestor
cd undercover
AMQP_URL=amqp://guest:guest@localhost:5672/ go run ./cmd/ingestor

# Run paquetier (separate terminal)
AMQP_URL=amqp://guest:guest@localhost:5672/ DATA_DIR=/tmp/parquet go run ./cmd/paquetier
```
