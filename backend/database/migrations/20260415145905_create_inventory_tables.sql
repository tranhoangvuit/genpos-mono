-- ============================================================================
-- 006_inventory.sql
-- Stock movements, stock cost prices, purchase orders, stock takes
-- ============================================================================

-- ============================================================================
-- Stock movements — per-store inventory ledger
-- ============================================================================
CREATE TABLE stock_movements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    register_id         UUID REFERENCES registers(id),
    variant_id          UUID NOT NULL REFERENCES product_variants(id),
    direction           TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    quantity            NUMERIC(10,4) NOT NULL CHECK (quantity > 0),
    movement_type       TEXT NOT NULL CHECK (movement_type IN (
        'purchase',
        'stock_in',
        'sale',
        'refund',
        'adjustment',
        'stock_take',
        'transfer_in',
        'transfer_out'
    )),
    reference_type      TEXT,
    reference_id        UUID,
    transfer_store_id   UUID REFERENCES stores(id),
    user_id             UUID REFERENCES users(id),
    notes               TEXT,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_movements_org_store_variant
    ON stock_movements(org_id, store_id, variant_id, created_at);
CREATE INDEX idx_movements_org_direction
    ON stock_movements(org_id, store_id, direction);
CREATE INDEX idx_movements_org_type
    ON stock_movements(org_id, movement_type, created_at);
CREATE INDEX idx_movements_org_reference
    ON stock_movements(org_id, reference_type, reference_id);
CREATE INDEX idx_movements_stock_calc
    ON stock_movements(org_id, store_id, variant_id, direction)
    WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE stock_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY stock_movements_org_isolation ON stock_movements
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Stock cost prices — per-store costing
-- ============================================================================
CREATE TABLE stock_cost_prices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    store_id        UUID NOT NULL REFERENCES stores(id),
    variant_id      UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    method          TEXT NOT NULL DEFAULT 'avg'
                        CHECK (method IN ('avg', 'fifo', 'lifo')),
    avg_cost        NUMERIC(12,4) NOT NULL DEFAULT 0,
    last_cost       NUMERIC(12,4) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_cost_prices_org_store_variant
    ON stock_cost_prices(org_id, store_id, variant_id);

-- Enable row security level
ALTER TABLE stock_cost_prices ENABLE ROW LEVEL SECURITY;
CREATE POLICY stock_cost_prices_org_isolation ON stock_cost_prices
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Stock cost price tracks — cost audit trail per movement
-- ============================================================================
CREATE TABLE stock_cost_price_tracks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    store_id            UUID NOT NULL REFERENCES stores(id),
    variant_id          UUID NOT NULL REFERENCES product_variants(id),
    movement_id         UUID NOT NULL REFERENCES stock_movements(id),
    direction           TEXT NOT NULL CHECK (direction IN ('in', 'out')),
    movement_type       TEXT NOT NULL,
    quantity            NUMERIC(10,4) NOT NULL,
    unit_cost           NUMERIC(12,4) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(12,4) NOT NULL DEFAULT 0,
    avg_cost_before     NUMERIC(12,4) NOT NULL DEFAULT 0,
    avg_cost_after      NUMERIC(12,4) NOT NULL DEFAULT 0,
    stock_on_hand_before NUMERIC(10,4) NOT NULL DEFAULT 0,
    stock_on_hand_after  NUMERIC(10,4) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_cost_tracks_movement
    ON stock_cost_price_tracks(org_id, movement_id);
CREATE INDEX idx_cost_tracks_variant_timeline
    ON stock_cost_price_tracks(org_id, store_id, variant_id, created_at);

-- Enable row security level
ALTER TABLE stock_cost_price_tracks ENABLE ROW LEVEL SECURITY;
CREATE POLICY stock_cost_price_tracks_org_isolation ON stock_cost_prices
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Purchase orders
-- ============================================================================
CREATE TABLE purchase_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    store_id        UUID NOT NULL REFERENCES stores(id),
    user_id         UUID REFERENCES users(id),
    po_number       TEXT NOT NULL,
    supplier_name   TEXT,
    status          TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft', 'submitted', 'partial', 'received', 'cancelled')),
    notes           TEXT,
    expected_at     TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_po_org_number ON purchase_orders(org_id, po_number);
CREATE INDEX idx_po_org_store ON purchase_orders(org_id, store_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_po_org_status ON purchase_orders(org_id, status) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY purchase_orders_org_isolation ON purchase_orders
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Purchase order items
-- ============================================================================
CREATE TABLE purchase_order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    quantity_ordered NUMERIC(10,4) NOT NULL,
    quantity_received NUMERIC(10,4) NOT NULL DEFAULT 0,
    cost_price      NUMERIC(12,4) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_po_items_org_po ON purchase_order_items(org_id, purchase_order_id);

-- Enable row security level
ALTER TABLE purchase_order_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY po_items_org_isolation ON purchase_order_items
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Stock takes
-- ============================================================================
CREATE TABLE stock_takes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    store_id        UUID NOT NULL REFERENCES stores(id),
    user_id         UUID REFERENCES users(id),
    status          TEXT NOT NULL DEFAULT 'in_progress'
                        CHECK (status IN ('in_progress', 'completed', 'cancelled')),
    notes           TEXT,
    completed_at    TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_takes_org_store ON stock_takes(org_id, store_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE stock_takes ENABLE ROW LEVEL SECURITY;
CREATE POLICY stock_takes_org_isolation ON stock_takes
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Stock take items
-- ============================================================================
CREATE TABLE stock_take_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    stock_take_id   UUID NOT NULL REFERENCES stock_takes(id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES product_variants(id),
    expected_qty    NUMERIC(10,4) NOT NULL DEFAULT 0,
    counted_qty     NUMERIC(10,4) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_stock_take_items_org_take ON stock_take_items(org_id, stock_take_id);

-- Enable row security level
ALTER TABLE stock_take_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY stock_take_items_org_isolation ON stock_take_items
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
