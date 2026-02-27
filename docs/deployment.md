# Deployment Guide

## Local Development

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Control plane services |
| Rust | 1.83+ | Data plane runtime |
| Docker | 24+ | PostgreSQL, full stack |
| Docker Compose | v2+ | Service orchestration |
| buf | latest | Proto generation (Go) |
| protoc | 3.x | Protocol buffer compiler |
| grpcurl | latest | Manual API testing |

### Quick Start

```bash
# Clone the repository
git clone <repo-url> bulkhead && cd bulkhead

# Build everything
make build

# Run unit tests
make test

# Start the full stack via Docker Compose
make dev
```

### Running Individual Services

For development, you can run services individually against a local PostgreSQL:

```bash
# Start PostgreSQL only
docker compose -f deploy/docker-compose.yml up -d postgres

# Run the Identity Service
cd control-plane && DATABASE_URL="postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable" go run ./cmd/identity

# Run the Rust Runtime
cd runtime && RUST_LOG=debug cargo run
```

---

## Docker Compose (Full Stack)

The full stack runs 11 containers: 9 Go microservices, 1 Rust runtime, and PostgreSQL.

### Starting the Stack

```bash
# Build and start all services
docker compose -f deploy/docker-compose.yml up --build

# Start in detached mode
docker compose -f deploy/docker-compose.yml up --build -d

# Check service health
docker compose -f deploy/docker-compose.yml ps

# View logs for a specific service
docker compose -f deploy/docker-compose.yml logs -f workspace

# Stop everything
docker compose -f deploy/docker-compose.yml down

# Stop and remove volumes (clean database)
docker compose -f deploy/docker-compose.yml down -v
```

### Service Topology

```mermaid
graph TB
    subgraph net["bulkhead network"]
        PG[("PostgreSQL :5432")]

        identity["identity :50060"]
        workspace["workspace :50061"]
        guardrails["guardrails :50062"]
        human["human :50063"]
        governance["governance :50064"]
        activity["activity :50065"]
        economics["economics :50066"]
        compute["compute :50067"]
        task["task :50068"]
        runtime["runtime :50052"]

        PG --- identity & workspace & guardrails & human
        PG --- governance & activity & economics & compute & task

        workspace --> compute
        workspace --> guardrails
        workspace --> task
        runtime --> human
        runtime --> activity
        runtime --> economics
    end
```

### Port Mappings

| Service | Internal Port | External Port | Protocol |
|---------|--------------|---------------|----------|
| PostgreSQL | 5432 | 5432 | PostgreSQL |
| Identity | 50051 | 50060 | gRPC |
| Workspace | 50051 | 50061 | gRPC |
| Guardrails | 50051 | 50062 | gRPC |
| Human Interaction | 50051 | 50063 | gRPC |
| Data Governance | 50051 | 50064 | gRPC |
| Activity Store | 50051 | 50065 | gRPC |
| Economics | 50051 | 50066 | gRPC |
| Compute Plane | 50051 | 50067 | gRPC |
| Task | 50051 | 50068 | gRPC |
| Runtime | 50052 | 50052 | gRPC |

### Service Dependencies

Services start in dependency order via Docker Compose `depends_on` with health checks:

1. **PostgreSQL** starts first (healthcheck: `pg_isready`)
2. **Independent services** start next: Identity, Guardrails, Human, Governance, Activity, Economics, Compute (depend only on PostgreSQL)
3. **Workspace** starts after Identity, Compute, and Guardrails are healthy
4. **Task** starts after Workspace is healthy
5. **Runtime** starts after Human, Activity, and Economics are healthy

### Health Checks

All Go services include a gRPC health check server. Docker health checks use `grpc_health_probe`:

```yaml
healthcheck:
  test: ["CMD", "grpc_health_probe", "-addr=:50051"]
  interval: 10s
  timeout: 5s
  retries: 5
```

---

## Configuration Reference

### Shared Configuration (All Go Services)

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable` | PostgreSQL connection string |
| `GRPC_PORT` | `50051` | Port for the gRPC server |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |

### Service-Specific Configuration

**Workspace Service:**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `COMPUTE_ENDPOINT` | — | gRPC endpoint for Compute Plane (e.g., `compute:50051`). If unset, orchestration is disabled. |
| `GUARDRAILS_ENDPOINT` | — | gRPC endpoint for Guardrails Service (e.g., `guardrails:50051`). If unset, empty policies are used. |

**Task Service:**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `WORKSPACE_ENDPOINT` | — | gRPC endpoint for Workspace Service (e.g., `workspace:50051`). Required for workspace provisioning on task start. |

**Runtime (Rust):**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `GRPC_PORT` | `50052` | Port for RuntimeService and AgentAPIService |
| `ADVERTISE_ADDRESS` | `localhost` | Hostname/IP returned to agents as the API endpoint |
| `HIS_ENDPOINT` | — | gRPC endpoint for Human Interaction Service (e.g., `http://human:50051`). If unset, `RequestHumanInput` returns `UNAVAILABLE`. |
| `ACTIVITY_ENDPOINT` | — | gRPC endpoint for Activity Store (e.g., `http://activity:50051`). If unset, action records are not persisted. |
| `ECONOMICS_ENDPOINT` | — | gRPC endpoint for Economics Service (e.g., `http://economics:50051`). If unset, budget enforcement is disabled. |
| `RUST_LOG` | `info` | Tracing filter (e.g., `debug`, `runtime=trace,tower=warn`) |

### Docker Compose Environment

In Docker Compose, all Go services connect to PostgreSQL via the internal network:

```yaml
DATABASE_URL: postgres://postgres:postgres@postgres:5432/sandbox?sslmode=disable
```

Cross-service endpoints use Docker Compose service names:

```yaml
# Workspace service
IDENTITY_ENDPOINT: identity:50051
COMPUTE_ENDPOINT: compute:50051
GUARDRAILS_ENDPOINT: guardrails:50051

# Task service
WORKSPACE_ENDPOINT: workspace:50051

# Runtime
HIS_ENDPOINT: http://human:50051
ACTIVITY_ENDPOINT: http://activity:50051
ECONOMICS_ENDPOINT: http://economics:50051
```

---

## Database

### PostgreSQL Setup

The Docker Compose stack uses PostgreSQL 16 Alpine with automatic schema initialization:

```yaml
postgres:
  image: postgres:16-alpine
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: sandbox
  volumes:
    - pgdata:/var/lib/postgresql/data
    - ../control-plane/migrations:/docker-entrypoint-initdb.d
```

Migration files in `control-plane/migrations/` are automatically executed on first start via PostgreSQL's `docker-entrypoint-initdb.d` mechanism.

### Schema Migrations

| File | Description |
|------|-------------|
| `001_initial.sql` | Core tables: agents, tasks, action_records, guardrail_rules, human_requests, scoped_credentials |
| `002_economics.sql` | Usage metering: usage_records, budgets |
| `003_workspace_compute.sql` | Workspace orchestration: workspaces, hosts |
| `004_workspace_sandbox.sql` | Snapshots: workspace_snapshots table, snapshot_id column on workspaces |
| `005_task_enrichment.sql` | Extended task spec: goal, workspace_config, budget_config, HIS config, labels. Agent enhancements: purpose, trust_level, capabilities |
| `006_missing_tables.sql` | HIS extensions: delivery_channels, timeout_policies. Workspace snapshot_id column |

### Migration Idempotency

The database module tracks applied migrations in a `schema_migrations` table. Migrations are applied automatically on service startup and are safe to re-run:

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT now()
);
```

### Key Tables

| Table | Service | Records |
|-------|---------|---------|
| `agents` | Identity | Agent registry |
| `scoped_credentials` | Identity | Token hashes, scopes, expiry |
| `tasks` | Task | Task lifecycle and configuration |
| `workspaces` | Workspace | Workspace state and placement info |
| `workspace_snapshots` | Workspace | Point-in-time snapshots |
| `hosts` | Compute | Runtime host fleet |
| `guardrail_rules` | Guardrails | Rule definitions |
| `human_requests` | HIS | Approval/question/escalation requests |
| `delivery_channels` | HIS | Per-user notification config |
| `timeout_policies` | HIS | Timeout behavior config |
| `action_records` | Activity | Append-only audit trail |
| `usage_records` | Economics | Usage events |
| `budgets` | Economics | Per-agent spending limits |

---

## Dockerfiles

### Control Plane (`deploy/docker/Dockerfile.control-plane`)

Multi-stage build for any Go service. The `SERVICE` build argument selects which binary to build:

```dockerfile
FROM golang:1.24-alpine AS builder
# ... build ./cmd/${SERVICE}

FROM alpine:3.20
# Includes grpc_health_probe for health checks
```

All 9 Go services share this Dockerfile with different `SERVICE` args.

### Runtime (`deploy/docker/Dockerfile.runtime`)

Multi-stage build for the Rust sandbox runtime:

```dockerfile
FROM rust:1.83-alpine AS builder
# Copies proto/ for build.rs proto compilation
# Builds release binary

FROM alpine:3.20
# Single binary: /runtime
```

---

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make proto` | Generate Go protobuf code via `buf generate` |
| `make build` | Build Go + Rust |
| `make build-go` | Build Go control plane only |
| `make build-rust` | Build Rust runtime only |
| `make test` | Run Go + Rust unit tests |
| `make test-go` | Run Go unit tests only |
| `make test-rust` | Run Rust unit tests only |
| `make test-integration` | Run Go integration tests (requires Docker) |
| `make test-integration-<svc>` | Run integration tests for a specific service (e.g., `make test-integration-identity`) |
| `make dev` | Start Docker Compose stack |
| `make dev-down` | Stop Docker Compose stack |
| `make fmt` | Format Go + Rust code |
| `make lint` | Lint protos (buf), Go (vet), Rust (clippy) |
| `make clean` | Remove build artifacts |
