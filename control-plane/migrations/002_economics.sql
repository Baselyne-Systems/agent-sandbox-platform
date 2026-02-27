BEGIN;

CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    workspace_id TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL,
    unit TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_usage_records_agent ON usage_records (agent_id);
CREATE INDEX idx_usage_records_workspace ON usage_records (workspace_id);

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL UNIQUE,
    currency TEXT NOT NULL DEFAULT 'USD',
    "limit" DOUBLE PRECISION NOT NULL,
    used DOUBLE PRECISION NOT NULL DEFAULT 0,
    period_start TIMESTAMPTZ NOT NULL DEFAULT now(),
    period_end TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days'
);

CREATE INDEX idx_budgets_agent ON budgets (agent_id);

COMMIT;
