-- ============================================================================
-- 001_foundation.sql
-- Extensions, helper functions, organizations, stores, registers
-- ============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- gen_random_uuid()

-- ============================================================================
-- Organizations — the RLS isolation boundary
-- ============================================================================
CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    currency        TEXT NOT NULL DEFAULT 'USD',
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    status          TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'suspended', 'churned')),
    settings        JSONB DEFAULT '{}',
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================================
-- Stores — physical locations within an org
-- ============================================================================
CREATE TABLE stores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    address         TEXT,
    phone           TEXT,
    email           TEXT,
    timezone        TEXT,
    status          TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'inactive', 'closed')),
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stores_org ON stores(org_id) WHERE deleted_at IS NULL;

-- Enable row security levels
ALTER TABLE stores ENABLE ROW LEVEL SECURITY;
CREATE POLICY stores_org_isolation ON stores
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Registers — devices within a store
-- ============================================================================
CREATE TABLE registers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    name                TEXT NOT NULL,
    device_fingerprint  TEXT,
    status              TEXT NOT NULL DEFAULT 'inactive'
                            CHECK (status IN ('active', 'inactive', 'deactivated')),
    last_seen_at        TIMESTAMPTZ,
    app_version         TEXT,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_registers_org_store ON registers(org_id, store_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_registers_fingerprint ON registers(org_id, device_fingerprint)
    WHERE device_fingerprint IS NOT NULL AND deleted_at IS NULL;

-- Enable row security level
ALTER TABLE registers ENABLE ROW LEVEL SECURITY;
CREATE POLICY registers_org_isolation ON registers
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
