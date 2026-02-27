BEGIN;

ALTER TABLE workspaces ADD COLUMN host_address TEXT NOT NULL DEFAULT '';
ALTER TABLE workspaces ADD COLUMN sandbox_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_workspaces_sandbox ON workspaces (sandbox_id) WHERE sandbox_id != '';

COMMIT;
