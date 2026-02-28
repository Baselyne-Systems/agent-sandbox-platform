# API Reference

> **Getting started?** See the [Operator Guide](getting-started/operator-guide.md), [Agent Developer Guide](getting-started/agent-guide.md), or [LangChain Integration Guide](getting-started/langchain-guide.md).

All services communicate over gRPC using Protocol Buffers. Proto definitions are in `proto/platform/*/v1/*.proto`.

## Authentication

Requests are authenticated via Bearer tokens in gRPC metadata:

```
authorization: Bearer <raw_token>
```

Tokens are issued by the Identity Service via `MintCredential`. Each credential has:
- **Scopes** — a list of permission strings (e.g., `workspace:create`, `tool:execute`)
- **TTL** — expiration time (max 24 hours)
- **Agent binding** — each credential is tied to a specific agent

The token is hashed with SHA-256 for storage and lookup. The raw token is returned exactly once at minting time.

**Exempt endpoints:** gRPC health checks and reflection are not authenticated.

---

## Identity Service

**Proto:** `proto/platform/identity/v1/identity.proto`
**Docker Compose port:** 50060

Manages agent registration, credentials, trust levels, and lifecycle.

### RPCs

| RPC | Description |
|-----|-------------|
| `RegisterAgent` | Register a new agent with name, owner, purpose, trust level, and capabilities. Returns the created agent with a generated UUID. |
| `GetAgent` | Retrieve an agent by ID. |
| `ListAgents` | List agents with optional filters (owner_id, status). Supports cursor-based pagination. |
| `DeactivateAgent` | Permanently deactivate an agent. Atomically revokes all active credentials. |
| `MintCredential` | Issue a scoped, time-limited credential. Returns the raw token (one-time) and credential metadata. TTL must be 1s–86400s (24h). |
| `RevokeCredential` | Revoke a specific credential by ID. |
| `UpdateTrustLevel` | Change an agent's trust level (new/established/trusted) with a justification string. Agent must be active. |
| `SuspendAgent` | Temporarily suspend an agent. Valid from active or already-suspended states. |
| `ReactivateAgent` | Restore a suspended or inactive agent to active status. |

### Key Messages

**Agent:**
```protobuf
message Agent {
  string agent_id = 1;
  string name = 2;
  string description = 3;
  string owner_id = 4;
  AgentStatus status = 5;        // ACTIVE, INACTIVE, SUSPENDED
  map<string, string> labels = 6;
  string purpose = 7;
  AgentTrustLevel trust_level = 8; // NEW, ESTABLISHED, TRUSTED
  repeated string capabilities = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}
```

**ScopedCredential:**
```protobuf
message ScopedCredential {
  string credential_id = 1;
  string agent_id = 2;
  repeated string scopes = 3;
  google.protobuf.Timestamp expires_at = 4;
  google.protobuf.Timestamp created_at = 5;
}
```

---

## Workspace Service

**Proto:** `proto/platform/workspace/v1/workspace.proto`
**Docker Compose port:** 50061

Orchestrates workspace lifecycle — creation, provisioning, termination, and snapshots.

### RPCs

| RPC | Description |
|-----|-------------|
| `CreateWorkspace` | Create a workspace for an agent with resource specs. If orchestration is enabled, triggers placement + guardrails compilation + sandbox creation. |
| `GetWorkspace` | Retrieve a workspace by ID. |
| `ListWorkspaces` | List workspaces with optional filters (agent_id, status). Cursor-based pagination. |
| `TerminateWorkspace` | Terminate a workspace. Destroys the sandbox on the runtime host if provisioned. |
| `SnapshotWorkspace` | Capture a point-in-time snapshot of a running workspace. |
| `RestoreWorkspace` | Restore a workspace from a snapshot ID. Creates a new workspace with the same configuration. |

### Key Messages

**Workspace:**
```protobuf
message Workspace {
  string workspace_id = 1;
  string agent_id = 2;
  string task_id = 3;
  WorkspaceStatus status = 4;    // PENDING, CREATING, RUNNING, PAUSED, TERMINATING, TERMINATED, FAILED
  WorkspaceSpec spec = 5;
  string host_id = 6;
  string snapshot_id = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}
```

**WorkspaceSpec:**
```protobuf
message WorkspaceSpec {
  int64 memory_mb = 1;
  int32 cpu_millicores = 2;
  int64 disk_mb = 3;
  int64 max_duration_secs = 4;
  repeated string allowed_tools = 5;
  string guardrail_policy_id = 6;   // Comma-separated rule IDs
  map<string, string> env_vars = 7;
  string container_image = 8;       // Docker image for sandbox container
  repeated string egress_allowlist = 9;  // Approved destination hosts/CIDRs
  IsolationTier isolation_tier = 10;     // Explicit tier override (auto-selected if unset)
  string data_classification = 11;      // Data sensitivity: public, internal, confidential, restricted (used for tier auto-selection)
}
```

**IsolationTier:**
```protobuf
enum IsolationTier {
  ISOLATION_TIER_UNSPECIFIED = 0;  // Auto-select based on agent trust level + data classification
  ISOLATION_TIER_STANDARD = 1;    // Docker container (cgroups + namespaces)
  ISOLATION_TIER_HARDENED = 2;    // Docker + seccomp + read-only rootfs + dropped caps
  ISOLATION_TIER_ISOLATED = 3;    // Docker + gVisor/Kata runtime
}
```

When `ISOLATION_TIER_UNSPECIFIED`, the Workspace Service auto-selects based on agent trust level and data classification:

| Trust \ Classification | public | internal | confidential | restricted |
|----------------------|--------|----------|--------------|------------|
| **trusted** | standard | standard | standard | isolated |
| **established** | standard | standard | hardened | isolated |
| **new** | hardened | hardened | isolated | isolated |

---

## Task Service

**Proto:** `proto/platform/task/v1/task.proto`
**Docker Compose port:** 50068

Manages task lifecycle. Creating and starting a task triggers workspace provisioning.

### RPCs

| RPC | Description |
|-----|-------------|
| `CreateTask` | Create a task with goal, workspace config, guardrails, budget, and HIS settings. |
| `GetTask` | Retrieve a task by ID. |
| `ListTasks` | List tasks with optional filters (agent_id, status). Cursor-based pagination. |
| `UpdateTaskStatus` | Transition task status. Moving to RUNNING triggers workspace provisioning. Terminal states trigger workspace termination. |
| `CancelTask` | Cancel a task. Allowed from pending, running, or waiting_on_human states. |

### Key Messages

**Task:**
```protobuf
message Task {
  string task_id = 1;
  string agent_id = 2;
  string goal = 3;
  TaskStatus status = 4;          // PENDING, RUNNING, WAITING_ON_HUMAN, COMPLETED, FAILED, CANCELLED
  string workspace_id = 5;
  TaskWorkspaceConfig workspace_config = 6;
  string guardrail_policy_id = 7;
  HumanInteractionConfig human_interaction_config = 8;
  BudgetConfig budget_config = 9;
  int64 max_duration_without_checkin_secs = 10;
  map<string, string> input = 11;
  map<string, string> labels = 12;
  google.protobuf.Timestamp created_at = 13;
  google.protobuf.Timestamp updated_at = 14;
  google.protobuf.Timestamp completed_at = 15;
}
```

---

## Compute Plane Service

**Proto:** `proto/platform/compute/v1/compute.proto`
**Docker Compose port:** 50067

Manages runtime hosts and handles workspace placement.

### RPCs

| RPC | Description |
|-----|-------------|
| `RegisterHost` | Register a new runtime host with address, total resource capacity, and supported isolation tiers. |
| `DeregisterHost` | Set a host to offline status. |
| `ListHosts` | List hosts with optional status filter. |
| `PlaceWorkspace` | Select a host for a workspace. Uses best-fit algorithm with atomic resource reservation and isolation tier filtering. Returns host_id and address. |
| `Heartbeat` | Host reports current resource availability and active sandbox count. Returns the host's current status (so control plane can signal drain). |

### Key Messages

**Host:**
```protobuf
message Host {
  string host_id = 1;
  string address = 2;
  HostStatus status = 3;          // READY, DRAINING, OFFLINE
  HostResources total_resources = 4;
  HostResources available_resources = 5;
  int32 active_sandboxes = 6;
  google.protobuf.Timestamp last_heartbeat = 7;
  repeated string supported_tiers = 8;  // e.g., ["standard", "hardened", "isolated"]
}
```

---

## Guardrails Service

**Proto:** `proto/platform/guardrails/v1/guardrails.proto`
**Docker Compose port:** 50062

Rule management and policy compilation.

### RPCs

| RPC | Description |
|-----|-------------|
| `CreateRule` | Create a guardrail rule with name, type, condition, action, and priority. |
| `GetRule` | Retrieve a rule by ID. |
| `ListRules` | List rules with optional type and enabled filters. Cursor-based pagination. |
| `UpdateRule` | Update an existing rule. |
| `DeleteRule` | Delete a rule by ID. |
| `CompilePolicy` | Compile a set of rules (by ID) into binary bytes for the Rust evaluator. Returns compiled bytes and rule count. |
| `SimulatePolicy` | Dry-run a policy against a sample tool call. Returns the verdict and matched rule without executing anything. |
| `GetBehaviorReport` | Get a behavior analysis report for an agent over a time window. Returns action count, denial rate, error rate, anomaly flags, and recommendation. |

### Key Messages

**GuardrailRule:**
```protobuf
message GuardrailRule {
  string rule_id = 1;
  string name = 2;
  string description = 3;
  RuleType type = 4;              // TOOL_FILTER, PARAMETER_CHECK, RATE_LIMIT, BUDGET_LIMIT
  string condition = 5;           // Tool filter: "exec,shell" | Param check: "path=/etc/shadow"
  RuleAction action = 6;          // ALLOW, DENY, ESCALATE, LOG
  int32 priority = 7;             // Lower number = higher priority
  bool enabled = 8;
  map<string, string> labels = 9;
  RuleScope scope = 12;           // Optional — restricts when this rule is evaluated
}
```

**RuleScope:**
```protobuf
message RuleScope {
  repeated string agent_ids = 1;            // Apply only to these agents (empty = all)
  repeated string tool_names = 2;           // Apply only to these tools (empty = all)
  repeated string trust_levels = 3;         // "new", "established", "trusted" (empty = all)
  repeated string data_classifications = 4; // "public", "internal", "confidential", "restricted"
}
```

**BehaviorReport:**
```protobuf
message BehaviorReport {
  string agent_id = 1;
  google.protobuf.Timestamp window_start = 2;
  google.protobuf.Timestamp window_end = 3;
  int64 action_count = 4;
  double denial_rate = 5;        // 0.0–1.0
  double error_rate = 6;         // 0.0–1.0
  repeated string flags = 7;     // e.g., "high_denial_rate:70%", "stuck_agent:api_call"
  string recommendation = 8;     // Human-readable recommendation
}
```

---

## Human Interaction Service

**Proto:** `proto/platform/human/v1/human.proto`
**Docker Compose port:** 50063

Manages human-in-the-loop requests, delivery channels, and timeout policies.

### RPCs

| RPC | Description |
|-----|-------------|
| `CreateRequest` | Create an approval, question, or escalation request for a human. |
| `GetRequest` | Retrieve a request by ID. Includes in-service expiry check. |
| `RespondToRequest` | Submit a human response with responder ID. |
| `ListRequests` | List requests with optional workspace and status filters. Cursor-based pagination. |
| `ConfigureDeliveryChannel` | Set a notification channel (Slack/email/Teams) for a user. |
| `GetDeliveryChannel` | Retrieve a user's delivery channel configuration. |
| `SetTimeoutPolicy` | Configure timeout behavior (global, per-agent, or per-workspace scope). |
| `GetTimeoutPolicy` | Retrieve a timeout policy. |

### Key Messages

**HumanRequest:**
```protobuf
message HumanRequest {
  string request_id = 1;
  string workspace_id = 2;
  string agent_id = 3;
  string task_id = 4;
  string question = 5;
  repeated string options = 6;
  string context = 7;
  HumanRequestStatus status = 8;  // PENDING, RESPONDED, EXPIRED, CANCELLED
  HumanRequestType type = 9;      // APPROVAL, QUESTION, ESCALATION
  HumanRequestUrgency urgency = 10; // LOW, NORMAL, HIGH, CRITICAL
  string response = 11;
  string responder_id = 12;
  google.protobuf.Timestamp created_at = 13;
  google.protobuf.Timestamp responded_at = 14;
  google.protobuf.Timestamp expires_at = 15;
  int64 timeout_seconds = 16;
}
```

---

## Activity Store

**Proto:** `proto/platform/activity/v1/activity.proto`
**Docker Compose port:** 50065

Append-only action records with query and streaming support.

### RPCs

| RPC | Description |
|-----|-------------|
| `RecordAction` | Append an action record. Returns the generated record ID. |
| `GetAction` | Retrieve a single action record by ID. |
| `QueryActions` | Query records with filters (workspace, agent, task, tool, outcome, time range). Cursor-based pagination. |
| `StreamActions` | Server-streaming RPC. Subscribe to real-time action records with workspace/agent filtering. |
| `ConfigureAlert` | Create or update an alert configuration. Conditions: denial_rate, error_rate, action_velocity, budget_breach, stuck_agent. |
| `ListAlerts` | List triggered alerts with optional agent filter and active-only toggle. Cursor-based pagination. |
| `GetAlert` | Retrieve a specific alert by ID. |
| `ResolveAlert` | Mark an alert as resolved. |

### Key Messages

**ActionRecord:**
```protobuf
message ActionRecord {
  string record_id = 1;
  string workspace_id = 2;
  string agent_id = 3;
  string task_id = 4;
  string tool_name = 5;
  google.protobuf.Struct parameters = 6;
  google.protobuf.Struct result = 7;
  ActionOutcome outcome = 8;      // ALLOWED, DENIED, ESCALATED, ERROR
  string guardrail_rule_id = 9;
  string denial_reason = 10;
  int64 evaluation_latency_us = 11;
  int64 execution_latency_us = 12;
  google.protobuf.Timestamp recorded_at = 13;
}
```

**AlertConfig:**
```protobuf
message AlertConfig {
  string config_id = 1;
  string name = 2;
  AlertConditionType condition_type = 3;  // DENIAL_RATE, ERROR_RATE, ACTION_VELOCITY, BUDGET_BREACH, STUCK_AGENT
  double threshold = 4;
  string agent_id = 5;            // Optional scope (empty = all agents)
  bool enabled = 6;
  string webhook_url = 7;
  google.protobuf.Timestamp created_at = 8;
}
```

**Alert:**
```protobuf
message Alert {
  string alert_id = 1;
  string config_id = 2;
  string agent_id = 3;
  AlertConditionType condition_type = 4;
  string message = 5;
  google.protobuf.Timestamp triggered_at = 6;
  bool resolved = 7;
}
```

---

## Economics Service

**Proto:** `proto/platform/economics/v1/economics.proto`
**Docker Compose port:** 50066

Usage metering, budget management, and cost reporting.

### RPCs

| RPC | Description |
|-----|-------------|
| `RecordUsage` | Record a usage event (resource type, quantity, cost). Also increments the agent's budget used amount. |
| `GetBudget` | Retrieve an agent's budget (limit, used, remaining). |
| `SetBudget` | Set or update an agent's budget limit with optional `on_exceeded` action and `warning_threshold`. Creates a 30-day budget period. Preserves existing used amount on update. |
| `CheckBudget` | Check if an agent can proceed given an estimated cost. Returns allowed (bool), remaining balance, enforcement_action (halt/request_increase), and warning flag. Called in the runtime hot path. |
| `GetCostReport` | Generate a cost report aggregated by resource type. Supports optional agent filter and time range. |

### Key Messages

**Budget:**
```protobuf
message Budget {
  string budget_id = 1;
  string agent_id = 2;
  double limit = 3;
  double used = 4;
  string currency = 5;
  google.protobuf.Timestamp period_start = 6;
  google.protobuf.Timestamp period_end = 7;
  OnExceededAction on_exceeded = 8;    // HALT, REQUEST_INCREASE, WARN
  double warning_threshold = 9;        // 0.0–1.0 fraction of limit
}
```

**CheckBudgetResponse:**
```protobuf
message CheckBudgetResponse {
  bool allowed = 1;
  double remaining = 2;
  string enforcement_action = 3;  // "halt", "request_increase", or "" (budget OK)
  bool warning = 4;               // true if remaining < warning_threshold * limit
}
```

---

## Data Governance Service

**Proto:** `proto/platform/governance/v1/governance.proto`
**Docker Compose port:** 50064

Stateless content classification and DLP. No database required.

### RPCs

| RPC | Description |
|-----|-------------|
| `ClassifyData` | Classify content into a sensitivity level (Public/Internal/Confidential/Restricted). Returns the classification and detected patterns. |
| `CheckPolicy` | Check if an agent can send data at a given classification to a destination. Restricted data is always denied. Confidential data is only allowed to approved destinations. |
| `InspectEgress` | Combined classify + check in a single call. Designed for the hot path. |

### Approved Destinations

- `internal-api`
- `secure-storage`
- `audit-log`

All other destinations are denied for Confidential and Restricted data.

---

## Host Agent — HostAgentService (Control API)

**Proto:** `proto/platform/host_agent/v1/host_agent.proto` — `HostAgentService`
**Docker Compose port:** 50052

Called by the control-plane Workspace Service to manage sandboxes on a host.

### RPCs

| RPC | Description |
|-----|-------------|
| `CreateSandbox` | Provision a sandbox with workspace/agent IDs, compiled guardrails, allowed tools, env vars, `container_image`, `egress_allowlist`, and `isolation_tier`. Starts a Docker container with the tier-specific security profile and applies iptables egress rules. Returns sandbox_id and agent API endpoint. |
| `DestroySandbox` | Tear down a sandbox. Sends lifecycle stopped event. |
| `GetSandboxStatus` | Get sandbox state, resource usage, and action count. |
| `StreamEvents` | Server-streaming RPC. Subscribe to sandbox events (action verdicts, lifecycle changes). |
| `UpdateSandboxGuardrails` | Hot-reload guardrails policy on a running sandbox without restart. |

---

## Host Agent — HostAgentAPIService (Agent-Facing)

**Proto:** `proto/platform/host_agent/v1/host_agent.proto` — `HostAgentAPIService`
**Docker Compose port:** 50052

Called by agents running inside sandboxes. Requires `x-sandbox-id` metadata header.

The Agent API is **policy-only**: `ExecuteTool` evaluates guardrails and budget but does NOT execute tools. The agent executes tools locally inside its container, then calls `ReportActionResult` to record the outcome for the audit trail. See the [Agent Developer Guide](getting-started/agent-guide.md) for the full SDK tutorial.

### RPCs

| RPC | Description |
|-----|-------------|
| `ExecuteTool` | Evaluate guardrails and budget for a tool call. Returns verdict (ALLOW/DENY/ESCALATE) and an `action_id`. Does NOT execute the tool — the agent is responsible for execution. |
| `ReportActionResult` | Record the outcome of an agent-executed tool call. Links to the `action_id` from `ExecuteTool` for the audit trail. |
| `RequestHumanInput` | Submit a question/approval/escalation to a human. Returns immediately with a request_id (non-blocking). |
| `CheckHumanRequest` | Poll for the status of a human request. Returns status (pending/responded/expired), response, and responder_id. |
| `ReportProgress` | Report task progress (message + percent complete). Emits a progress event on the sandbox channel. |

### Python SDK

The Python SDK (`bulkhead-sdk`) handles the evaluate-execute-report cycle transparently via the `@tool` decorator. See the [Agent Developer Guide](getting-started/agent-guide.md) for a full tutorial.

---

## Error Codes

All services map domain errors to standard gRPC status codes:

| gRPC Code | Meaning | Example |
|-----------|---------|---------|
| `NOT_FOUND` | Resource does not exist | Agent, workspace, credential, host, or human request not found |
| `INVALID_ARGUMENT` | Bad input | Empty required fields, invalid TTL, zero resource requests |
| `FAILED_PRECONDITION` | State conflict | Minting credential for inactive agent, invalid status transition |
| `RESOURCE_EXHAUSTED` | No capacity | No host with sufficient resources for placement |
| `UNAUTHENTICATED` | Missing/invalid token | Missing `x-sandbox-id` header, expired credential |
| `UNAVAILABLE` | Service not configured | HIS endpoint not set on runtime |
| `INTERNAL` | Unexpected error | Database errors, lock poisoning, serialization failures |
