-- Warm pool configuration: how many pre-warmed slots to maintain per isolation tier.
CREATE TABLE warm_pool_configs (
    isolation_tier TEXT PRIMARY KEY,
    target_count   INTEGER NOT NULL DEFAULT 0,
    memory_mb      BIGINT  NOT NULL DEFAULT 512,
    cpu_millicores INTEGER NOT NULL DEFAULT 1000,
    disk_mb        BIGINT  NOT NULL DEFAULT 10240,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Individual warm pool slots: pre-reserved resources on a host, ready for instant claim.
CREATE TABLE warm_pool_slots (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id        UUID NOT NULL REFERENCES hosts(id),
    isolation_tier TEXT NOT NULL,
    memory_mb      BIGINT  NOT NULL,
    cpu_millicores INTEGER NOT NULL,
    disk_mb        BIGINT  NOT NULL,
    status         TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready', 'claimed', 'expired')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_at     TIMESTAMPTZ
);

CREATE INDEX idx_warm_pool_slots_claim ON warm_pool_slots (isolation_tier, status)
    WHERE status = 'ready';
