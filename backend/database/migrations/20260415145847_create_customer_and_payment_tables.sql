-- ============================================================================
-- 004_customers_payments.sql
-- Customers, payment methods
-- ============================================================================

-- ============================================================================
-- Customers
-- ============================================================================
CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    notes           TEXT,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_customers_org ON customers(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_org_email ON customers(org_id, email)
    WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_customers_org_phone ON customers(org_id, phone)
    WHERE phone IS NOT NULL AND deleted_at IS NULL;

-- Enable row security level
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
CREATE POLICY customers_org_isolation ON customers
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Payment methods — org-configurable payment types
-- ============================================================================
CREATE TABLE payment_methods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('cash', 'card', 'mobile', 'bank_transfer', 'voucher', 'other')),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_methods_org ON payment_methods(org_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE payment_methods ENABLE ROW LEVEL SECURITY;
CREATE POLICY payment_methods_org_isolation ON payment_methods
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
