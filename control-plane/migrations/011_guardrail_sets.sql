-- 011: GuardrailSet — named, reusable collections of guardrail rules.

CREATE TABLE IF NOT EXISTS guardrail_sets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    rule_ids    JSONB NOT NULL DEFAULT '[]',
    labels      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_guardrail_sets_name ON guardrail_sets (name);
