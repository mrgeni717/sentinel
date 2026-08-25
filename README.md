# sentinel

A lightweight systems monitoring tool — uptime checks, host metrics via a
small agent, and threshold-based alerting with webhook notifications. A
portfolio project written in Go, deliberately using the same language as
the tooling it's inspired by (Prometheus, Docker, Kubernetes, Terraform
are all Go).

## Status

🚧 In progress. Core backend (uptime checker, metric ingestion, alert
engine, dashboard) built; not yet deployed to AWS.

## Architecture

Two binaries share one Go module:

- **`cmd/server`** — the collector: runs the uptime-check scheduler (one
  goroutine per target, each on its own configured interval), exposes a
  REST API, evaluates alert rules, and serves the web dashboard as static
  files. Storage is SQLite via a pure-Go driver (`modernc.org/sqlite`) —
  no CGO, no C compiler needed, and no separate database process to run.
- **`cmd/agent`** — a small standalone binary you run on any host you want
  to monitor. Collects CPU/memory/disk/load stats (via `gopsutil`) and
  pushes them to the server on an interval. Being a static Go binary, it
  has no runtime dependencies — copy the compiled binary to any machine
  and run it.

```
cmd/
  server/    the collector + API + dashboard
  agent/     the metrics-pushing agent
internal/
  store/     SQLite schema + all queries
  checker/   uptime-check scheduler
  alertengine/  threshold evaluation + webhook firing
  api/       HTTP handlers
  model/     shared structs
web/static/  the dashboard (plain HTML/CSS/JS, no build step)
```

## Design notes

- **No external HTTP router** — Go 1.22+'s standard library `http.ServeMux`
  supports method+path patterns (`"GET /api/targets/{id}"`) directly, so
  there's no dependency needed just for routing.
- **Alert rule evaluation is point-in-time**, not duration-aware in v1: a
  rule fires as soon as a single reading breaches its threshold, rather
  than requiring the breach to hold for N seconds first. `DurationSeconds`
  exists on the model as a placeholder for that but isn't implemented yet
  — noted here rather than silently glossed over.
- **The ingest endpoint** (`POST /api/ingest/metrics`, what the agent
  calls) is protected by a shared-secret header (`X-Ingest-Key`), not full
  user auth — it's a machine-to-machine endpoint, not something a browser
  user calls directly.

## Running locally

Requires: Go 1.23+ (no Docker, no external database needed).

```bash
go mod tidy          # resolves and downloads dependencies
go run ./cmd/server   # starts the collector on :8090
```

Open `http://localhost:8090` for the dashboard.

To run the agent against it (in a second terminal, from the same
machine or a different one):

```bash
set SENTINEL_SERVER_URL=http://localhost:8090   # Windows PowerShell: $env:SENTINEL_SERVER_URL="http://localhost:8090"
go run ./cmd/agent
```

The agent pushes real metrics from whatever machine it's run on. Within
a few seconds it should appear under **Hosts** on the dashboard.

## Environment variables

**Server:**
- `SENTINEL_DB_PATH` — SQLite file path (default `sentinel.db`)
- `SENTINEL_PORT` — HTTP port (default `8090`)
- `SENTINEL_STATIC_DIR` — dashboard static files path (default `web/static`)
- `SENTINEL_INGEST_KEY` — shared secret for the agent ingest endpoint (optional; unset = unauthenticated, fine for local dev)

**Agent:**
- `SENTINEL_SERVER_URL` — where to push metrics (default `http://localhost:8090`)
- `SENTINEL_HOST_NAME` — override the reported hostname (default: OS hostname)
- `SENTINEL_INGEST_KEY` — must match the server's, if the server has one set
- `SENTINEL_PUSH_INTERVAL_SECONDS` — how often to push (default `30`)

## Alerting via Slack

Any alert rule can have a `webhookUrl`. Slack's [Incoming
Webhooks](https://api.slack.com/messaging/webhooks) accept a plain
`POST` with a JSON body, so pointing a rule's webhook at a Slack webhook
URL works directly — Slack will render the payload sentinel sends.

## Deployment

Single AWS EC2 instance (Terraform), Docker Compose — no separate database
container needed since SQLite is embedded, so this is a one-container
deployment (unlike `bank-demo`'s app+MySQL two-container setup).

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars   # then edit github_repo_url
terraform init
terraform apply
```

Wait a few minutes for the instance to install Docker, clone the repo, and
build the image, then check instance health before testing:

```bash
aws ec2 describe-instance-status --instance-ids <ID> --output table
```

Once `InstanceStatus` shows `ok`, open the `app_url` from the Terraform
outputs.

**Instance size:** defaults to `t3.small` rather than `t3.micro`. Even
though Go itself compiles fast, `modernc.org/sqlite` (the pure-Go SQLite
driver, chosen to avoid needing a C toolchain) is a large transpiled
package that's genuinely heavy to compile — building it inside Docker on
a 1GB-RAM `t3.micro` risks the same out-of-memory instability hit during
the `bank-demo` deployment.

Tear down when done testing/demoing:

```bash
cd infra
terraform destroy
```
