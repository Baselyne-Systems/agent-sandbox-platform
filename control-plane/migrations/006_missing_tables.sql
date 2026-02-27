BEGIN;

-- Add snapshot_id column to workspaces (referenced by workspace/repository.go SetSnapshotID)
ALTER TABLE workspaces ADD COLUMN snapshot_id TEXT NOT NULL DEFAULT '';

-- Workspace snapshots table (referenced by workspace/repository.go CreateSnapshot/GetSnapshot)
CREATE TABLE workspace_snapshots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    agent_id    TEXT NOT NULL,
    task_id     TEXT NOT NULL DEFAULT '',
    size_bytes  BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_workspace_snapshots_workspace ON workspace_snapshots (workspace_id);

-- Delivery channels table (referenced by human/repository.go UpsertDeliveryChannel/GetDeliveryChannel)
CREATE TABLE delivery_channels (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      TEXT NOT NULL,
    channel_type TEXT NOT NULL CHECK (channel_type IN ('slack', 'email', 'teams')),
    endpoint     TEXT NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, channel_type)
);

CREATE INDEX idx_delivery_channels_user ON delivery_channels (user_id);

-- Timeout policies table (referenced by human/repository.go UpsertTimeoutPolicy/GetTimeoutPolicy)
CREATE TABLE timeout_policies (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope              TEXT NOT NULL CHECK (scope IN ('global', 'agent', 'workspace')),
    scope_id           TEXT NOT NULL DEFAULT '',
    timeout_secs       BIGINT NOT NULL,
    action             TEXT NOT NULL CHECK (action IN ('escalate', 'continue', 'halt')),
    escalation_targets JSONB NOT NULL DEFAULT '[]',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (scope, scope_id)
);

CREATE INDEX idx_timeout_policies_scope ON timeout_policies (scope, scope_id);

COMMIT;
