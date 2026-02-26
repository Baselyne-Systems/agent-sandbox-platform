-- 001_initial.sql
-- Core tables for the Agent Sandbox Platform control plane.

BEGIN;

-- Agent registry
CREATE TABLE agents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    owner_id    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    labels      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agents_owner ON agents (owner_id);
CREATE INDEX idx_agents_status ON agents (status);

-- Task tracking
CREATE TABLE tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id    UUID NOT NULL REFERENCES agents(id),
    workspace_id UUID,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_agent ON tasks (agent_id);
CREATE INDEX idx_tasks_status ON tasks (status);

-- Activity store (append-only action records)
CREATE TABLE action_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id        UUID NOT NULL,
    agent_id            UUID NOT NULL REFERENCES agents(id),
    task_id             UUID REFERENCES tasks(id),
    tool_name           TEXT NOT NULL,
    parameters          JSONB NOT NULL DEFAULT '{}',
    result              JSONB,
    outcome             TEXT NOT NULL CHECK (outcome IN ('allowed', 'denied', 'escalated', 'error')),
    guardrail_rule_id   UUID,
    denial_reason       TEXT,
    evaluation_latency_us BIGINT,
    execution_latency_us  BIGINT,
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Append-only: no UPDATE or DELETE expected. Optimize for time-range queries.
CREATE INDEX idx_action_records_workspace ON action_records (workspace_id, recorded_at);
CREATE INDEX idx_action_records_agent ON action_records (agent_id, recorded_at);
CREATE INDEX idx_action_records_task ON action_records (task_id, recorded_at);
CREATE INDEX idx_action_records_tool ON action_records (tool_name, recorded_at);

-- Guardrail rules
CREATE TABLE guardrail_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL CHECK (type IN ('tool_filter', 'parameter_check', 'rate_limit', 'budget_limit')),
    condition   TEXT NOT NULL,
    action      TEXT NOT NULL CHECK (action IN ('allow', 'deny', 'escalate', 'log')),
    priority    INT NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    labels      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_guardrail_rules_type ON guardrail_rules (type);
CREATE INDEX idx_guardrail_rules_enabled ON guardrail_rules (enabled) WHERE enabled = true;

-- Human interaction requests
CREATE TABLE human_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL,
    agent_id        UUID NOT NULL REFERENCES agents(id),
    question        TEXT NOT NULL,
    options         JSONB NOT NULL DEFAULT '[]',
    context         TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'responded', 'expired', 'cancelled')),
    response        TEXT,
    responder_id    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ
);

CREATE INDEX idx_human_requests_status ON human_requests (status);
CREATE INDEX idx_human_requests_workspace ON human_requests (workspace_id);

-- Scoped credentials
CREATE TABLE scoped_credentials (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id    UUID NOT NULL REFERENCES agents(id),
    scopes      JSONB NOT NULL DEFAULT '[]',
    token_hash  TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_scoped_credentials_agent ON scoped_credentials (agent_id);
CREATE INDEX idx_scoped_credentials_hash ON scoped_credentials (token_hash) WHERE revoked = false;

COMMIT;
