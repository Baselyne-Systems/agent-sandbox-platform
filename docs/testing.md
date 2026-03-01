# Testing Guide

This guide covers the Bulkhead end-to-end test suite. The tests serve two purposes: they validate cross-service behavior, and they demonstrate every major platform workflow — making them a practical reference for operators learning the system.

> **See also:** [Operator Guide](getting-started/operator-guide.md) | [Architecture](architecture.md) | [API Reference](api-reference.md) | [Deployment Guide](deployment.md)

---

## Running Tests

| Command | What it runs | Requirements | Time |
|---------|-------------|-------------|------|
| `make test` | Unit tests (Go + Rust) | None | ~10s |
| `make test-integration` | Per-service DB tests | Docker | ~30s |
| `make test-e2e` | Control-plane E2E tests | Docker | ~8s |
| `make test-e2e-full` | Full-stack E2E tests (real Rust runtime) | Docker + Rust toolchain | ~15s |
| `make test-e2e-all` | All E2E tests (control-plane + full-stack) | Docker + Rust toolchain | ~21s |

### Prerequisites

- **Docker** — TestContainers starts PostgreSQL 16 automatically; no manual database setup.
- **Rust toolchain** (full-stack tests only) — `make test-e2e-full` builds the `host-agent` binary in release mode before running tests.

### Running a single test

```bash
cd control-plane
go test -count=1 -v -run TestBudgetFullCycle ./e2e/...
```

---

## Test Architecture

The E2E tests live in `control-plane/e2e/` and wire **all 9 real Go services** against a real PostgreSQL instance. Only two boundaries are mocked:

```
┌──────────────────────────────────────────────────────────────────┐
│                         TestMain                                 │
│                                                                  │
│  PostgreSQL 16 (TestContainers)                                  │
│       │                                                          │
│  ┌────┴────────────────────────────────────────────────────┐     │
│  │  Real Repositories (8 PostgresRepository instances)     │     │
│  └────┬────────────────────────────────────────────────────┘     │
│       │                                                          │
│  ┌────┴────────────────────────────────────────────────────┐     │
│  │  Real Services                                          │     │
│  │  Identity · Task · Workspace · Compute · Guardrails     │     │
│  │  Economics · Human · Activity · Governance               │     │
│  └────┬────────────────────────────────────────────────────┘     │
│       │                                                          │
│  ┌────┴───────────────────┐  ┌─────────────────────────────┐    │
│  │  fakeHostAgentClient   │  │  fakeSnapshotStore          │    │
│  │  (mock gRPC boundary)  │  │  (mock object store)        │    │
│  └────────────────────────┘  └─────────────────────────────┘    │
│                                                                  │
│  Full-stack tests replace the mock with the real Rust binary:    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  host-agent binary (subprocess)                         │    │
│  │  ENABLE_DOCKER=false (noop container runtime)           │    │
│  │  ← connects to real Go gRPC servers on dynamic ports →  │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

Each test calls `clean(t)` to truncate all tables and reset mock state, ensuring full isolation between tests.

---

## Test Catalog

### Agent Lifecycle (`agent_lifecycle_test.go`)

These tests cover the identity service: agent registration, credential management, trust levels, and suspension.

| Test | What it demonstrates |
|------|---------------------|
| `TestAgentFullLifecycle` | Register → Get → MintCredential → UpdateTrustLevel → Suspend (credential minting blocked) → Reactivate (credential minting restored) → Deactivate |
| `TestAgentRegistrationValidation` | Empty name, owner ID, or tenant ID returns an error |
| `TestAgentTrustLevelProgression` | Trust level transitions: `new` → `established` → `trusted`; invalid level rejected |

**Operator reference:** This is the agent lifecycle you manage with `bkctl agent register`, `bkctl agent update-trust`, and `bkctl agent suspend`. A suspended agent cannot mint credentials, which prevents it from authenticating to any sandbox.

---

### Task and Workspace Provisioning (`task_workspace_test.go`)

These tests cover the end-to-end flow from creating a task to workspace provisioning, including failure handling and state transitions.

| Test | What it demonstrates |
|------|---------------------|
| `TestTaskCreationToCompletion` | Create task (pending) → Start (provisions workspace, sandbox created) → Complete (sandbox destroyed, workspace terminated) |
| `TestTaskCancellation` | Create + start → cancel → workspace terminated |
| `TestTaskFailureOnProvisioningError` | No compute host available → start task → task transitions to `failed` |
| `TestTaskStatusTransitions` | All valid state transitions succeed; invalid transitions (`completed→running`, `failed→running`) are rejected |
| `TestTaskWaitingOnHuman` | `running` → `waiting_on_human` → `running` → `completed` |

**Operator reference:** When you create a task via `bkctl task create`, the platform automatically: (1) finds a host with sufficient resources, (2) compiles guardrail rules into a policy bundle, (3) mints short-lived credentials, and (4) calls CreateSandbox on the assigned Host Agent. If no host has capacity, the task moves to `failed` — you can resolve this by adding hosts to the fleet.

---

### Multi-Tenant Isolation (`multi_tenant_test.go`)

These tests verify that tenant data is completely isolated at the database level.

| Test | What it demonstrates |
|------|---------------------|
| `TestCrossTenantIsolation` | Data created in tenant-A is invisible to tenant-B: agents, tasks, guardrail rules, budgets, and list operations are all scoped |
| `TestSharedComputeHosts` | A single host serves workspaces for multiple tenants — compute is shared infrastructure, tenant isolation is at the sandbox level |
| `TestTenantImplicitCreation` | Registering an agent with a new tenant ID works without any CreateTenant API |

**Operator reference:** There is no tenant provisioning step. A tenant comes into existence when its first agent is registered. All data queries are scoped by `tenant_id` — there is no way for one tenant's agents to see another's guardrails, budgets, or activity records. Compute hosts, however, are global shared infrastructure.

---

### Guardrails Pipeline (`guardrails_pipeline_test.go`)

These tests cover the full guardrails workflow: rule creation, policy compilation, and simulation.

| Test | What it demonstrates |
|------|---------------------|
| `TestGuardrailPolicyPipeline` | Create deny rule (shell) + allow rule (web_search) → group into a rule set → compile to policy JSON |
| `TestPolicySimulation` | Deny rule for "shell" → simulate("shell") returns DENY; simulate("web_search") returns ALLOW (default-allow when no rules match) |
| `TestRulePriorityOrdering` | Rule A: priority 10, allow shell; Rule B: priority 1, deny shell → lowest priority number wins → DENY |
| `TestParameterCheckRule` | Parameter-check rule with condition `env=production` → DENY when params match; ALLOW when they don't |
| `TestGuardrailSetResolution` | Rule sets resolve to their member rule IDs; comma-separated IDs resolve directly |

**Operator reference:** This is the workflow you use with `bkctl guardrail create-rule`, `bkctl guardrail create-set`, and `bkctl guardrail simulate`. Priority ordering matters — lower numbers win. Use `simulate` to test policies before deploying them to production workspaces. Policies are compiled into a binary format that the Rust evaluator can process in <50ms.

---

### Budget Enforcement (`budget_enforcement_test.go`)

These tests cover per-agent spending limits, usage tracking, and enforcement modes.

| Test | What it demonstrates |
|------|---------------------|
| `TestBudgetFullCycle` | Set budget ($100) → record usage ($30) → check ($10 estimated) allowed → record more ($50) → check ($30 estimated) denied with `halt` enforcement |
| `TestBudgetWarningThreshold` | Budget $100 with 20% warning threshold → $85 used → check returns allowed but with `warning=true` |
| `TestBudgetOnExceededModes` | Three enforcement modes when budget exceeded: `halt` (denied), `warn` (allowed + warning), `request_increase` (denied + enforcement action) |
| `TestBudgetPreservesUsedOnUpdate` | Updating a budget limit preserves the existing usage amount |
| `TestNoBudgetAllowsAll` | No budget set → any estimated cost is allowed |

**Operator reference:** Set budgets with `bkctl budget set`. Choose `on_exceeded` behavior: `halt` stops the agent immediately (safest), `warn` lets it continue but flags the overage, `request_increase` pauses the agent and signals that a human should approve more budget. Budgets are checked on every tool call in the Rust runtime — there is no way for an agent to bypass the check.

---

### Human Interaction (`human_interaction_test.go`)

These tests cover the human-in-the-loop system: approval requests, expiration, delivery channels, and timeout policies.

| Test | What it demonstrates |
|------|---------------------|
| `TestHumanRequestLifecycle` | Create request (pending) → get → respond ("yes") → verify response and responder ID |
| `TestHumanRequestExpiration` | Create request with 1s timeout → wait 2s → status is `expired` |
| `TestDeliveryChannelConfiguration` | Configure a Slack webhook delivery channel → retrieve and verify |
| `TestTimeoutPolicies` | Set global timeout policy (300s, escalate) and agent-scoped policy (60s, halt) → verify independent retrieval |

**Operator reference:** When an agent calls `RequestHumanInput`, the request appears in the Human Interaction Service and can be delivered via configured webhook channels (Slack, email, etc.). Timeout policies control what happens when nobody responds: `escalate` sends to a secondary responder list; `halt` stops the agent. Agent-scoped policies override global policies.

---

### Compute Fleet (`compute_fleet_test.go`)

These tests cover host registration, workspace placement, capacity management, warm pools, and liveness detection.

| Test | What it demonstrates |
|------|---------------------|
| `TestHostRegistrationAndHeartbeat` | Register host (16GB RAM) → verify ready → heartbeat with updated resources → verify available resources updated |
| `TestWorkspacePlacement` | Register host (4096MB) → place workspace (512MB) → resources decremented |
| `TestPlacementExhaustsCapacity` | Register host (1024MB) → place 512MB twice → third placement fails with no capacity |
| `TestTierAwarePlacement` | Host A: [standard], Host B: [standard, hardened] → place hardened workspace → assigned to Host B; place isolated → no capacity |
| `TestWarmPoolWorkflow` | Configure warm pool (3 slots) → replenish → verify pre-allocated slots → place workspace → claims warm slot |
| `TestHostLivenessDetection` | Register host → set stale heartbeat → liveness sweep → host marked offline |

**Operator reference:** Hosts self-register and heartbeat every 30s. Use `bkctl compute list-hosts` to monitor fleet status. If a host stops heartbeating for 3 minutes, the liveness worker marks it offline. Warm pools (`bkctl compute configure-warm-pool`) pre-allocate sandbox slots for instant placement — configure these for latency-sensitive workloads. Placement respects isolation tiers: a workspace requiring `hardened` will only land on a host that supports it.

---

### Activity and Audit (`activity_audit_test.go`)

These tests cover the append-only audit trail: recording actions, querying, and exporting.

| Test | What it demonstrates |
|------|---------------------|
| `TestActionRecordingAndQuery` | Record 5 actions (different tools, outcomes) → query by workspace (all 5), by tool name (subset), by outcome (subset) |
| `TestActionExportJSON` | Record 3 actions → export as NDJSON → parse and verify each record |
| `TestActionExportCSV` | Record 3 actions → export as CSV → verify header row + 3 data rows |
| `TestAlertConfigurationWithoutRepo` | Alert configuration returns `ErrAlertsNotEnabled` when no alert repository is wired |

**Operator reference:** Every tool call — allowed, denied, or escalated — is recorded with tool name, parameters, verdict, matched rule, and latency. Use `bkctl activity query` to search by workspace, tool, outcome, or time range. Use `bkctl activity export --format json` or `--format csv` for compliance reporting. Records are append-only and cannot be modified or deleted.

---

### Workspace Orchestration (`workspace_orchestration_test.go`)

These tests cover the workspace lifecycle: creation, snapshot/restore, termination, credential injection, and failure handling.

| Test | What it demonstrates |
|------|---------------------|
| `TestWorkspaceFullOrchestration` | Register agent + host → create workspace → verify `running` status, host_id, sandbox_id, and exactly 1 CreateSandbox call |
| `TestWorkspaceSnapshotRestore` | Create workspace (running) → snapshot → verify `paused` + snapshot record → restore → `running` again |
| `TestWorkspaceTermination` | Create (running) → terminate with reason → `terminated` status, 1 DestroySandbox call |
| `TestWorkspaceCredentialInjection` | Create workspace with `AllowedTools` and `MaxDurationSecs` → verify `BULKHEAD_AGENT_TOKEN` and `BULKHEAD_AGENT_ID` injected as env vars |
| `TestWorkspaceFailsGracefully` | Host Agent returns error → workspace transitions to `failed` without panic |

**Operator reference:** Workspace creation is fully automated: the platform selects a host, compiles guardrails, mints a scoped credential, and passes everything to the Host Agent. The credential is injected as `BULKHEAD_AGENT_TOKEN` and `BULKHEAD_AGENT_ID` environment variables — agents use these to authenticate. Snapshot/restore enables pausing a workspace and resuming later on a different host.

---

### Isolation Tier Auto-Selection (`tier_selection_test.go`)

This test validates all 12 combinations of agent trust level and data classification against the expected isolation tier.

| Trust Level \ Data | Public | Internal | Confidential | Restricted |
|-------------------|--------|----------|--------------|------------|
| **Trusted** | standard | standard | standard | isolated |
| **Established** | standard | standard | hardened | isolated |
| **New** | hardened | hardened | isolated | isolated |

**Operator reference:** You don't need to set isolation tiers manually. When you specify a `data_classification` on the workspace spec, the platform cross-references it with the agent's trust level and selects the appropriate tier. Use `bkctl agent update-trust` to promote agents as they demonstrate trustworthy behavior.

---

### Full-Stack Tests (`full_stack_test.go`)

These tests start the **real Rust Host Agent binary** alongside real Go gRPC servers. The runtime runs with `ENABLE_DOCKER=false` (noop container runtime) so no Docker-in-Docker is needed. These tests exercise the actual network path: Go gRPC servers ↔ Rust runtime ↔ Agent API.

| Test | What it demonstrates |
|------|---------------------|
| `TestFullStackRuntimeRegistration` | Start runtime → verifies it self-registers with the Compute Plane and reports 8192 MB memory |
| `TestFullStackWorkspaceCreation` | Register agent → create workspace → runtime creates sandbox (noop) → workspace is `running` |
| `TestFullStackExecuteToolAllow` | Create sandbox (no guardrails) → call ExecuteTool("web_search") → verdict ALLOW |
| `TestFullStackExecuteToolDeny` | Create deny rule for "shell" → compile policy → create sandbox with policy → ExecuteTool("shell") → verdict DENY with reason |
| `TestFullStackBudgetEnforcement` | Set budget ($10) → exhaust ($15 usage) → ExecuteTool → DENY from budget check |
| `TestFullStackHumanInteraction` | Create sandbox → RequestHumanInput("approve deployment?") → respond via HIS → CheckHumanRequest returns "yes" |

**Operator reference:** These tests demonstrate the complete enforcement pipeline as it runs in production. When an agent calls `ExecuteTool`, the request flows through: (1) budget check via Economics Service, (2) guardrails evaluation in the Rust evaluator, (3) DLP inspection via Governance Service, (4) activity recording. A DENY at any layer stops the action. The full-stack tests prove this works across the real Go ↔ Rust boundary, not just in mocked unit tests.

---

## Test Infrastructure Details

### TestMain setup

`TestMain` in `e2e_test.go` handles all wiring:

1. **PostgreSQL** — Started via TestContainers. Migrations run from `control-plane/migrations/`.
2. **Repositories** — 8 real PostgresRepository instances (one per service with a data store).
3. **Services** — All 9 services wired with real dependencies. Cross-service adapters bridge interface boundaries (e.g., `workspaceProvisionerAdapter` wraps the workspace service to implement `task.WorkspaceProvisioner`).
4. **Mocks** — Only 2: `fakeHostAgentClient` (simulates the Rust runtime for control-plane tests) and `fakeSnapshotStore` (in-memory snapshot storage).

### Full-stack test setup

Full-stack tests in `full_stack_test.go` add:

1. **gRPC servers** — All 7 Go gRPC handlers started on dynamic ports with a `tenantResolverInterceptor` that resolves `tenant_id` from request fields via DB lookup (simulating auth middleware).
2. **Rust runtime** — The `host-agent` binary started as a subprocess with `ENABLE_DOCKER=false`. It self-registers with the Compute Plane and communicates with all Go services over gRPC.
3. **Sandbox creation** — Tests create sandboxes directly on the runtime via `HostAgentService.CreateSandbox`, then pass the sandbox ID as `x-sandbox-id` gRPC metadata on Agent API calls.

### Writing new tests

```go
func TestMyNewFeature(t *testing.T) {
    clean(t)                                    // Truncate all tables + reset mocks
    ctx := context.Background()
    tenant := uniqueTenant()                    // Unique tenant for isolation

    agent := registerAgent(t, ctx, tenant, "my-agent")
    registerHost(t, ctx, "host.local:9090", 4096, 4000, 10240, []string{"standard"})

    // Use the real services directly:
    ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
    if err != nil {
        t.Fatalf("create workspace: %v", err)
    }
    // Assert on the result...
}
```

Helpers available: `clean(t)`, `uniqueTenant()`, `registerAgent()`, `registerAgentWithTrust()`, `registerHost()`.

For full-stack tests, add: `startGRPCServers()`, `startRuntime()`, `waitForRuntime()`, `dialAgentAPI()`, `dialHostAgentService()`, `createSandboxOnRuntime()`, `withSandboxID()`, `makeParams()`.
