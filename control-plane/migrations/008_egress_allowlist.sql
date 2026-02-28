-- 008: Add egress_allowlist column to workspaces table.
-- Stores approved destination hosts/CIDRs as a JSON array.
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS egress_allowlist JSONB NOT NULL DEFAULT '[]';
