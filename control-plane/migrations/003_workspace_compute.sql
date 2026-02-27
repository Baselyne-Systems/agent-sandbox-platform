BEGIN;

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    task_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('pending', 'creating', 'running', 'paused', 'terminating', 'terminated', 'failed')),
    memory_mb BIGINT NOT NULL DEFAULT 512,
    cpu_millicores INTEGER NOT NULL DEFAULT 500,
    disk_mb BIGINT NOT NULL DEFAULT 1024,
    max_duration_secs BIGINT NOT NULL DEFAULT 3600,
    allowed_tools JSONB NOT NULL DEFAULT '[]',
    guardrail_policy_id TEXT NOT NULL DEFAULT '',
    env_vars JSONB NOT NULL DEFAULT '{}',
    host_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_workspaces_agent ON workspaces (agent_id);
CREATE INDEX idx_workspaces_status ON workspaces (status);
CREATE INDEX idx_workspaces_host ON workspaces (host_id);

CREATE TABLE hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ready', 'draining', 'offline')),
    total_memory_mb BIGINT NOT NULL,
    total_cpu_millicores INTEGER NOT NULL,
    total_disk_mb BIGINT NOT NULL,
    available_memory_mb BIGINT NOT NULL,
    available_cpu_millicores INTEGER NOT NULL,
    available_disk_mb BIGINT NOT NULL,
    active_sandboxes INTEGER NOT NULL DEFAULT 0,
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hosts_placement ON hosts (status, available_memory_mb, available_cpu_millicores, available_disk_mb)
    WHERE status = 'ready';

COMMIT;
