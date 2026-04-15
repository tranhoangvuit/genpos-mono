-- ============================================================================
-- 007_operations.sql
-- Shifts, store config, deferred foreign keys
-- ============================================================================

-- ============================================================================
-- Shifts — register-scoped cash management sessions
-- ============================================================================
CREATE TABLE shifts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    register_id         UUID REFERENCES registers(id),
    user_id             UUID NOT NULL REFERENCES users(id),
    status              TEXT NOT NULL DEFAULT 'open'
                            CHECK (status IN ('open', 'closed')),
    opening_amount      NUMERIC(12,4) NOT NULL DEFAULT 0,
    closing_amount      NUMERIC(12,4),
    expected_amount     NUMERIC(12,4),
    difference          NUMERIC(12,4),
    notes               TEXT,
    opened_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at           TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shifts_org_store ON shifts(org_id, store_id, opened_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_shifts_org_register ON shifts(org_id, register_id)
    WHERE register_id IS NOT NULL AND deleted_at IS NULL;
-- Fast lookup: which shift is currently open on this register?
CREATE UNIQUE INDEX idx_shifts_open_register ON shifts(org_id, register_id)
    WHERE status = 'open' AND deleted_at IS NULL;

-- ============================================================================
-- Add deferred FK: orders.shift_id -> shifts.id
-- ============================================================================
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_shift
    FOREIGN KEY (shift_id) REFERENCES shifts(id);

-- Enable row security level
ALTER TABLE shifts ENABLE ROW LEVEL SECURITY;
CREATE POLICY shifts_org_isolation ON shifts
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Store config — per-store key/value settings
-- ============================================================================
CREATE TABLE store_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    store_id        UUID NOT NULL REFERENCES stores(id),
    key             TEXT NOT NULL,
    value           TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_store_config_org_store_key ON store_config(org_id, store_id, key);

-- Enable row securit level
ALTER TABLE store_config ENABLE ROW LEVEL SECURITY;
CREATE POLICY store_config_org_isolation ON store_config
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
