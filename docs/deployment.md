# Deployment Guide

> **Getting started?** See the [Operator Guide](getting-started/operator-guide.md) for a step-by-step walkthrough of deploying and using the platform.

## Local Development

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.24+ | Control plane services |
| Rust | 1.83+ | Host Agent |
| Python | 3.10+ | SDK for building agents |
| Docker | 24+ | PostgreSQL, full stack, sandbox containers |
| Docker Compose | v2+ | Service orchestration |
| buf | latest | Proto generation (Go) |
| protoc | 3.x | Protocol buffer compiler |
| grpcurl | latest | Manual API testing (optional — `bkctl` covers most workflows) |

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

For development, you can run binaries individually against a local PostgreSQL:

```bash
# Start PostgreSQL only
docker compose -f deploy/docker-compose.yml up -d postgres

# Run the control-plane binary (Identity + Task + Workspace + Compute)
cd control-plane && DATABASE_URL="postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable" go run ./cmd/control-plane

# Run the policy binary (Guardrails + Governance)
cd control-plane && DATABASE_URL="postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable" go run ./cmd/policy

# Run the observability binary (Activity + Economics + Human)
cd control-plane && DATABASE_URL="postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable" go run ./cmd/observability

# Run the Rust Host Agent
cd runtime && RUST_LOG=debug cargo run
```

---

## Docker Compose (Full Stack)

The full stack runs 5 containers: 3 Go binaries (each hosting multiple gRPC services), 1 Rust Host Agent, and PostgreSQL (plus optional Jaeger for tracing).

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

        cp["control-plane :50060<br/>(Identity, Task, Workspace, Compute)"]
        policy["policy :50062<br/>(Guardrails, Governance)"]
        obs["observability :50065<br/>(Activity, Economics, Human)"]
        runtime["host-agent :50052"]
        jaeger["jaeger :16686 / :4317"]

        PG --- cp & policy & obs

        cp -->|guardrails| policy
        runtime --> obs
        runtime -->|compute| cp
        runtime -->|governance| policy
    end
```

### Port Mappings

| Binary | Services | Internal Port | External Port | Protocol |
|--------|----------|--------------|---------------|----------|
| PostgreSQL | — | 5432 | 5432 | PostgreSQL |
| control-plane | Identity, Task, Workspace, Compute | 50051 | 50060 | gRPC |
| policy | Guardrails, Data Governance | 50051 | 50062 | gRPC |
| observability | Activity, Economics, Human | 50051 | 50065 | gRPC |
| host-agent | Host Agent | 50052 | 50052 | gRPC |
| Jaeger UI | — | 16686 | 16686 | HTTP |
| Jaeger OTLP | — | 4317 | 4317 | gRPC |

### Service Dependencies

Services start in dependency order via Docker Compose `depends_on` with health checks:

1. **PostgreSQL** starts first (healthcheck: `pg_isready`)
2. **policy** and **observability** start next (depend only on PostgreSQL)
3. **control-plane** starts after policy is healthy (Workspace needs Guardrails cross-binary)
4. **Host Agent** starts after control-plane, policy, and observability are healthy

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

## Kubernetes (Helm)

For production deployments, Bulkhead includes a Helm chart that deploys the full stack to any Kubernetes cluster.

### Prerequisites

- Kubernetes 1.28+
- Helm 3+
- Container registry with Bulkhead images (build with `make docker-build`)

### Install

```bash
# Install with defaults
helm install bulkhead deploy/helm/bulkhead

# Install with custom values
helm install bulkhead deploy/helm/bulkhead -f custom-values.yaml

# Upgrade
helm upgrade bulkhead deploy/helm/bulkhead

# Uninstall
helm uninstall bulkhead
```

### Architecture

The Helm chart creates:

| Resource Type | Count | Components |
|--------------|-------|------------|
| **Deployment** | 3 | control-plane (Identity+Task+Workspace+Compute), policy (Guardrails+Governance), observability (Activity+Economics+Human) |
| **DaemonSet** | 1 | Host Agent — runs on every node for local sandbox management |
| **StatefulSet** | 1 | PostgreSQL with persistent volume |
| **Job** | 1 | Database migration (runs before services start, backoffLimit: 3) |
| **Service** | 5 | ClusterIP services for inter-binary gRPC communication |
| **NetworkPolicy** | 1 | Optional — restricts traffic to declared dependencies |
| **Secret** | 1 | Database credentials |

### Key Configuration Values

| Value | Default | Description |
|-------|---------|-------------|
| `postgres.storage` | `10Gi` | Persistent volume size for PostgreSQL |
| `postgres.resources.limits.memory` | `1Gi` | PostgreSQL memory limit |
| `controlPlane.replicaCount` | `1` | Replicas per Go service |
| `controlPlane.image.repository` | `bulkhead/control-plane` | Control plane image base |
| `controlPlane.resources.limits.memory` | `512Mi` | Memory limit per Go service |
| `controlPlane.resources.limits.cpu` | `500m` | CPU limit per Go service |
| `controlPlane.logLevel` | `info` | Log level for Go services |
| `hostAgent.enabled` | `true` | Deploy Host Agent DaemonSet |
| `hostAgent.enableDocker` | `true` | Enable Docker container lifecycle |
| `hostAgent.resources.limits.memory` | `1Gi` | Host Agent memory limit |
| `networkPolicy.enabled` | `true` | Create NetworkPolicy resources |
| `observability.otelEndpoint` | `""` | OTLP endpoint (e.g., `jaeger-collector:4317`) |
| `serviceType` | `ClusterIP` | Kubernetes Service type |

### Host Agent DaemonSet

The Host Agent runs as a DaemonSet so every node in the cluster can host sandboxes:

- Mounts the Docker socket (`/var/run/docker.sock`) for container lifecycle management
- Tolerates `NoSchedule` taints so it runs on all nodes including tainted ones
- Advertise address is set from the pod IP (`status.podIP`) so the control plane can route sandbox creation to the correct node
- Service endpoints are derived from Helm release name (e.g., `{{ .Release.Name }}-observability:50051`)

### Migration Job

The migration Job runs all SQL files from `control-plane/migrations/` against PostgreSQL before any service starts. It has `backoffLimit: 3` — if migration fails after 3 attempts, the job fails and services will not start. Check migration logs with:

```bash
kubectl logs job/bulkhead-migration
```

---

## Configuration Reference

### Shared Configuration (All Go Services)

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/sandbox?sslmode=disable` | PostgreSQL connection string |
| `GRPC_PORT` | `50051` | Port for the gRPC server |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `""` (disabled) | OTLP gRPC endpoint for distributed tracing (e.g., `jaeger:4317`). If empty, tracing is disabled. |

### Binary-Specific Configuration

**control-plane binary** (Identity, Task, Workspace, Compute):

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `GUARDRAILS_ENDPOINT` | — | gRPC endpoint for the policy binary (e.g., `policy:50051`). If unset, empty policies are used. |
| `HEARTBEAT_TIMEOUT_SECS` | `180` | Seconds without a heartbeat before a host is marked offline. The liveness worker checks every 60 seconds. |

Task→Workspace and Workspace→Compute calls are in-process (same binary) and require no endpoint configuration.

**Host Agent (Rust):**

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CONTROL_PLANE` | — | Control plane IP or hostname. Derives all service endpoints using well-known ports (see table below). Individual endpoint vars override. |
| `GRPC_PORT` | `50052` | Port for HostAgentService and HostAgentAPIService |
| `ADVERTISE_ADDRESS` | auto-detected | Hostname/IP returned to agents as the API endpoint. Auto-detected from OS routing table if not set. |
| `HIS_ENDPOINT` | from `CONTROL_PLANE` :50065 | gRPC endpoint for Human Interaction (observability binary, e.g., `http://observability:50051`). If neither this nor `CONTROL_PLANE` is set, `RequestHumanInput` returns `UNAVAILABLE`. |
| `ACTIVITY_ENDPOINT` | from `CONTROL_PLANE` :50065 | gRPC endpoint for Activity Store (observability binary, e.g., `http://observability:50051`). If unset, action records are not persisted. |
| `ECONOMICS_ENDPOINT` | from `CONTROL_PLANE` :50065 | gRPC endpoint for Economics (observability binary, e.g., `http://observability:50051`). If unset, budget enforcement is disabled. |
| `GOVERNANCE_ENDPOINT` | from `CONTROL_PLANE` :50062 | gRPC endpoint for Data Governance (policy binary, e.g., `http://policy:50051`). If unset, DLP egress inspection is disabled. |
| `COMPUTE_ENDPOINT` | from `CONTROL_PLANE` :50060 | gRPC endpoint for Compute Plane (control-plane binary, e.g., `http://control-plane:50051`). Enables host self-registration and periodic heartbeats. |
| `TOTAL_MEMORY_MB` | auto-detected | Total memory (MB) advertised to Compute Plane. Auto-detected from host system if not set. |
| `TOTAL_CPU_MILLICORES` | auto-detected | Total CPU (millicores) advertised to Compute Plane. Auto-detected (cores × 1000) if not set. |
| `TOTAL_DISK_MB` | auto-detected | Total disk (MB) advertised to Compute Plane. Auto-detected from root mount if not set. |
| `ENABLE_DOCKER` | `false` | Set to `true` to enable Docker container management via bollard. Requires Docker socket access. |
| `SUPPORTED_TIERS` | `standard,hardened` | Comma-separated isolation tiers this host supports. Options: `standard`, `hardened`, `isolated`. |
| `ISOLATED_RUNTIME` | (not set) | Docker runtime for the `isolated` tier (e.g., `runsc` for gVisor, `kata` for Kata Containers). Setting this automatically adds `isolated` to supported tiers. |
| `RUST_LOG` | `info` | Tracing filter (e.g., `debug`, `host_agent=trace,tower=warn`) |

#### CONTROL_PLANE Well-Known Ports

When `CONTROL_PLANE` is set, the Host Agent derives service endpoints by combining the address with these ports (matching the Docker Compose external port mappings):

| Binary | Port | Services | Env Var Override |
|--------|------|----------|-----------------|
| control-plane | 50060 | Identity, Task, Workspace, Compute | `COMPUTE_ENDPOINT` |
| policy | 50062 | Guardrails, Governance | `GOVERNANCE_ENDPOINT` |
| observability | 50065 | Activity, Economics, Human | `HIS_ENDPOINT`, `ACTIVITY_ENDPOINT`, `ECONOMICS_ENDPOINT` |

### Docker Compose Environment

In Docker Compose, all Go services connect to PostgreSQL via the internal network:

```yaml
DATABASE_URL: postgres://postgres:postgres@postgres:5432/sandbox?sslmode=disable
```

Cross-binary endpoints use Docker Compose service names:

```yaml
# control-plane binary
GUARDRAILS_ENDPOINT: policy:50051

# Host Agent
HIS_ENDPOINT: http://observability:50051
ACTIVITY_ENDPOINT: http://observability:50051
ECONOMICS_ENDPOINT: http://observability:50051
GOVERNANCE_ENDPOINT: http://policy:50051
COMPUTE_ENDPOINT: http://control-plane:50051
TOTAL_MEMORY_MB: "16384"
TOTAL_CPU_MILLICORES: "8000"
TOTAL_DISK_MB: "102400"
ENABLE_DOCKER: "true"
SUPPORTED_TIERS: "standard,hardened"
# ISOLATED_RUNTIME: "runsc"  # Uncomment if gVisor is installed on the host
```

The Host Agent container needs Docker socket access for managing agent containers:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
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
| `007_container_image.sql` | Adds `container_image` column to workspaces table |
| `008_egress_allowlist.sql` | Adds `egress_allowlist` JSONB column to workspaces table |
| `009_isolation_tiers.sql` | Adds `supported_tiers` TEXT[] column to hosts table and `isolation_tier` TEXT column to workspaces table |
| `010_multi_tenant.sql` | Multi-tenancy: adds `tenant_id` column across all tables with composite indexes |
| `011_guardrail_sets.sql` | Guardrail sets: `guardrail_sets` table for reusable rule collections |
| `012_alert_system.sql` | Alert system: `alert_configs` and `alerts` tables for Activity Store alerting |
| `013_warm_pool.sql` | Warm pool: `warm_pool_configs` and `warm_pool_slots` tables for pre-reserved sandbox slots |

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
| `warm_pool_configs` | Compute | Per-tier warm pool targets and default resources |
| `warm_pool_slots` | Compute | Pre-reserved sandbox slots on hosts |

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

All 3 Go binaries share this Dockerfile with different `SERVICE` args (`control-plane`, `policy`, `observability`).

### Host Agent (`deploy/docker/Dockerfile.host-agent`)

Multi-stage build for the Rust Host Agent:

```dockerfile
FROM rust:1.83-alpine AS builder
# Copies proto/ for build.rs proto compilation
# Builds release binary

FROM alpine:3.20
# Single binary: /host-agent
```

---

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make proto` | Generate Go protobuf code via `buf generate` |
| `make build` | Build Go + Rust |
| `make build-go` | Build Go control plane only |
| `make build-bkctl` | Build `bkctl` operator CLI with version/commit info |
| `make build-rust` | Build Rust Host Agent only |
| `make test` | Run Go + Rust unit tests |
| `make test-go` | Run Go unit tests only |
| `make test-rust` | Run Rust Host Agent unit tests only |
| `make test-integration` | Run Go integration tests (requires Docker) |
| `make test-integration-<svc>` | Run integration tests for a specific service (e.g., `make test-integration-identity`) |
| `make test-e2e` | Run E2E tests — control-plane services with real PostgreSQL (requires Docker) |
| `make test-e2e-full` | Run full-stack E2E tests with real Rust runtime binary (requires Docker + Rust) |
| `make test-e2e-all` | Run all E2E tests (control-plane + full-stack) |
| `make dev` | Start Docker Compose stack |
| `make dev-down` | Stop Docker Compose stack |
| `make fmt` | Format Go + Rust code |
| `make lint` | Lint protos (buf), Go (vet), Rust (clippy) |
| `make clean` | Remove build artifacts |
