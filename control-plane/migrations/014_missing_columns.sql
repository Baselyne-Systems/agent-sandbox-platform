-- 014_missing_columns.sql
-- Add columns referenced by service code but missing from previous migrations.

-- guardrail_rules: add scope column for rule scoping (agent, workspace, tier).
ALTER TABLE guardrail_rules ADD COLUMN IF NOT EXISTS scope JSONB NOT NULL DEFAULT '{}';

-- budgets: add on_exceeded action and warning_threshold columns.
ALTER TABLE budgets ADD COLUMN IF NOT EXISTS on_exceeded TEXT NOT NULL DEFAULT 'halt';
ALTER TABLE budgets ADD COLUMN IF NOT EXISTS warning_threshold DOUBLE PRECISION NOT NULL DEFAULT 0;
