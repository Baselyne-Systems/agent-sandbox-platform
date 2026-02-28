-- 009: Add isolation tier support.
-- supported_tiers on hosts: which isolation tiers this host can run.
-- isolation_tier on workspaces: the resolved tier for this workspace.
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS supported_tiers TEXT[] NOT NULL DEFAULT '{standard}';
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS isolation_tier TEXT NOT NULL DEFAULT '';
