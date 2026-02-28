-- Migration 012: Multi-tenancy support
-- Adds tenant_id to all tenant-scoped tables. Hosts remain shared infrastructure.

BEGIN;

-- 1. agents
ALTER TABLE agents ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_agents_tenant ON agents (tenant_id);
ALTER TABLE agents ALTER COLUMN tenant_id DROP DEFAULT;

-- 2. scoped_credentials
ALTER TABLE scoped_credentials ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_scoped_credentials_tenant ON scoped_credentials (tenant_id);
ALTER TABLE scoped_credentials ALTER COLUMN tenant_id DROP DEFAULT;

-- 3. tasks
ALTER TABLE tasks ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_tasks_tenant ON tasks (tenant_id);
ALTER TABLE tasks ALTER COLUMN tenant_id DROP DEFAULT;

-- 4. workspaces
ALTER TABLE workspaces ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_workspaces_tenant ON workspaces (tenant_id);
ALTER TABLE workspaces ALTER COLUMN tenant_id DROP DEFAULT;

-- 5. workspace_snapshots
ALTER TABLE workspace_snapshots ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_workspace_snapshots_tenant ON workspace_snapshots (tenant_id);
ALTER TABLE workspace_snapshots ALTER COLUMN tenant_id DROP DEFAULT;

-- 6. guardrail_rules
ALTER TABLE guardrail_rules ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_guardrail_rules_tenant ON guardrail_rules (tenant_id);
ALTER TABLE guardrail_rules ALTER COLUMN tenant_id DROP DEFAULT;

-- 7. guardrail_sets: name was globally unique, now unique per tenant
ALTER TABLE guardrail_sets ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
DROP INDEX IF EXISTS idx_guardrail_sets_name;
ALTER TABLE guardrail_sets DROP CONSTRAINT IF EXISTS guardrail_sets_name_key;
ALTER TABLE guardrail_sets ADD CONSTRAINT guardrail_sets_tenant_name_key UNIQUE (tenant_id, name);
CREATE INDEX idx_guardrail_sets_tenant ON guardrail_sets (tenant_id);
ALTER TABLE guardrail_sets ALTER COLUMN tenant_id DROP DEFAULT;

-- 8. human_requests
ALTER TABLE human_requests ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_human_requests_tenant ON human_requests (tenant_id);
ALTER TABLE human_requests ALTER COLUMN tenant_id DROP DEFAULT;

-- 9. delivery_channels: unique per (tenant, user, channel_type)
ALTER TABLE delivery_channels ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
ALTER TABLE delivery_channels DROP CONSTRAINT IF EXISTS delivery_channels_user_id_channel_type_key;
ALTER TABLE delivery_channels ADD CONSTRAINT delivery_channels_tenant_user_channel_key UNIQUE (tenant_id, user_id, channel_type);
CREATE INDEX idx_delivery_channels_tenant ON delivery_channels (tenant_id);
ALTER TABLE delivery_channels ALTER COLUMN tenant_id DROP DEFAULT;

-- 10. timeout_policies: unique per (tenant, scope, scope_id)
ALTER TABLE timeout_policies ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
ALTER TABLE timeout_policies DROP CONSTRAINT IF EXISTS timeout_policies_scope_scope_id_key;
ALTER TABLE timeout_policies ADD CONSTRAINT timeout_policies_tenant_scope_key UNIQUE (tenant_id, scope, scope_id);
CREATE INDEX idx_timeout_policies_tenant ON timeout_policies (tenant_id);
ALTER TABLE timeout_policies ALTER COLUMN tenant_id DROP DEFAULT;

-- 11. action_records: DDL (ALTER TABLE ADD COLUMN) is not blocked by immutability triggers
ALTER TABLE action_records ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_action_records_tenant ON action_records (tenant_id, recorded_at);
ALTER TABLE action_records ALTER COLUMN tenant_id DROP DEFAULT;

-- 12. budgets: unique per (tenant, agent)
ALTER TABLE budgets ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS budgets_agent_id_key;
ALTER TABLE budgets ADD CONSTRAINT budgets_tenant_agent_key UNIQUE (tenant_id, agent_id);
CREATE INDEX idx_budgets_tenant ON budgets (tenant_id);
ALTER TABLE budgets ALTER COLUMN tenant_id DROP DEFAULT;

-- 13. usage_records
ALTER TABLE usage_records ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default-tenant';
CREATE INDEX idx_usage_records_tenant ON usage_records (tenant_id);
ALTER TABLE usage_records ALTER COLUMN tenant_id DROP DEFAULT;

COMMIT;
