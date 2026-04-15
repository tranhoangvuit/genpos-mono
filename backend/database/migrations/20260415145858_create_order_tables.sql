-- ============================================================================
-- 005_orders.sql
-- Orders, order line items, payments, refunds, refund line items
-- ============================================================================

-- ============================================================================
-- Orders
-- ============================================================================
CREATE TABLE orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    register_id         UUID REFERENCES registers(id),
    shift_id            UUID,                   -- FK added after shifts table exists
    customer_id         UUID REFERENCES customers(id),
    user_id             UUID REFERENCES users(id),
    order_number        TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'open'
                            CHECK (status IN ('open', 'completed', 'voided', 'refunded', 'partially_refunded')),
    subtotal            NUMERIC(12,4) NOT NULL DEFAULT 0,
    tax_total           NUMERIC(12,4) NOT NULL DEFAULT 0,
    discount_total      NUMERIC(12,4) NOT NULL DEFAULT 0,
    total               NUMERIC(12,4) NOT NULL DEFAULT 0,
    notes               TEXT,
    completed_at        TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_orders_org_number ON orders(org_id, order_number);
CREATE INDEX idx_orders_org_store ON orders(org_id, store_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_org_status ON orders(org_id, status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_org_customer ON orders(org_id, customer_id) WHERE customer_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_orders_shift ON orders(org_id, shift_id) WHERE shift_id IS NOT NULL;

-- Enable row security level
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY orders_org_isolation ON orders
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Order line items — snapshots product data at time of sale
-- ============================================================================
CREATE TABLE order_line_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id),
    order_id                UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    variant_id              UUID REFERENCES product_variants(id),
    product_name            TEXT NOT NULL,
    variant_name            TEXT NOT NULL DEFAULT 'Default',
    sku                     TEXT,
    quantity                NUMERIC(10,4) NOT NULL,
    unit_price              NUMERIC(12,4) NOT NULL,
    cost_price_snapshot     NUMERIC(12,4) NOT NULL DEFAULT 0,
    tax_rate                NUMERIC(6,4) NOT NULL DEFAULT 0,
    tax_amount              NUMERIC(12,4) NOT NULL DEFAULT 0,
    discount_amount         NUMERIC(12,4) NOT NULL DEFAULT 0,
    line_total              NUMERIC(12,4) NOT NULL,
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_line_items_org_order ON order_line_items(org_id, order_id);
CREATE INDEX idx_line_items_org_variant ON order_line_items(org_id, variant_id)
    WHERE variant_id IS NOT NULL;

-- Enable row security level
ALTER TABLE order_line_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY order_line_items_org_isolation ON order_line_items
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Payments — split payments per order
-- ============================================================================
CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    payment_method_id   UUID NOT NULL REFERENCES payment_methods(id),
    amount              NUMERIC(12,4) NOT NULL,
    tendered            NUMERIC(12,4),          -- for cash: amount given
    change_amount       NUMERIC(12,4),          -- for cash: change returned
    reference           TEXT,                   -- card auth code, mobile tx id, etc.
    status              TEXT NOT NULL DEFAULT 'completed'
                            CHECK (status IN ('completed', 'voided', 'refunded')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_org_order ON payments(org_id, order_id);

-- Enable row security level
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
CREATE POLICY payments_org_isolation ON payments
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Refunds
-- ============================================================================
CREATE TABLE refunds (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    register_id         UUID REFERENCES registers(id),
    order_id            UUID NOT NULL REFERENCES orders(id),
    user_id             UUID REFERENCES users(id),
    refund_number       TEXT NOT NULL,
    amount              NUMERIC(12,4) NOT NULL,
    reason              TEXT,
    status              TEXT NOT NULL DEFAULT 'completed'
                            CHECK (status IN ('completed', 'voided')),
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_refunds_org_number ON refunds(org_id, refund_number);
CREATE INDEX idx_refunds_org_order ON refunds(org_id, order_id);

-- Enable row security level
ALTER TABLE refunds ENABLE ROW LEVEL SECURITY;
CREATE POLICY refunds_org_isolation ON refunds
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Refund line items
-- ============================================================================
CREATE TABLE refund_line_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID NOT NULL REFERENCES organizations(id),
    refund_id               UUID NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    order_line_item_id      UUID NOT NULL REFERENCES order_line_items(id),
    variant_id              UUID REFERENCES product_variants(id),
    quantity                NUMERIC(10,4) NOT NULL,
    amount                  NUMERIC(12,4) NOT NULL,
    restock                 BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refund_items_org_refund ON refund_line_items(org_id, refund_id);

-- Enable row security level
ALTER TABLE refund_line_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY refund_line_items_org_isolation ON refund_line_items
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
