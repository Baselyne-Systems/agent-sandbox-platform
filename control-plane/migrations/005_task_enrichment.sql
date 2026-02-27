BEGIN;

-- Enrich agents table with spec fields
ALTER TABLE agents ADD COLUMN purpose TEXT NOT NULL DEFAULT '';
ALTER TABLE agents ADD COLUMN trust_level TEXT NOT NULL DEFAULT 'new' CHECK (trust_level IN ('new', 'established', 'trusted'));
ALTER TABLE agents ADD COLUMN capabilities JSONB NOT NULL DEFAULT '[]';

-- Enrich tasks table to match spec (replace minimal scaffold)
ALTER TABLE tasks ADD COLUMN goal TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN isolation_tier TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN persistent BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE tasks ADD COLUMN guardrail_policy_id TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN workspace_config JSONB NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN human_interaction_config JSONB NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN budget_config JSONB NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN max_duration_without_checkin_secs BIGINT NOT NULL DEFAULT 0;
ALTER TABLE tasks ADD COLUMN input JSONB NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN labels JSONB NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN completed_at TIMESTAMPTZ;

-- Update tasks status check to include waiting_on_human
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check
  CHECK (status IN ('pending', 'running', 'waiting_on_human', 'completed', 'failed', 'cancelled'));

-- Enrich human_requests with type, urgency, task_id
ALTER TABLE human_requests ADD COLUMN type TEXT NOT NULL DEFAULT 'question' CHECK (type IN ('approval', 'question', 'escalation'));
ALTER TABLE human_requests ADD COLUMN urgency TEXT NOT NULL DEFAULT 'normal' CHECK (urgency IN ('low', 'normal', 'high', 'critical'));
ALTER TABLE human_requests ADD COLUMN task_id UUID REFERENCES tasks(id);

CREATE INDEX idx_human_requests_task ON human_requests (task_id) WHERE task_id IS NOT NULL;
CREATE INDEX idx_tasks_workspace ON tasks (workspace_id) WHERE workspace_id IS NOT NULL;

COMMIT;
