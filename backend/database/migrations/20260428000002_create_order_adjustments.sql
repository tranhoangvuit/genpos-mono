-- Per-line and per-order adjustment + tax breakdown tables. These hold the
-- expanded children of the aggregate snapshot fields already present on
-- order_line_items and orders (tax_amount, tax_rate, discount_amount,
-- tax_total, discount_total). At write time the engine populates both: the
-- child rows describe the calculation; the aggregates exist for fast list
-- and report queries. Aggregates are not enforced by the DB -- the engine
-- writes both inside one transaction.

-- ============================================================================
-- order_line_taxes -- per-tax breakdown snapshot per line item
-- ============================================================================
-- Snapshots the taxes resolved for this line at sale time. A line may have
-- 0..N taxes (most products have 1; compound stacks have 2+). Snapshot fields
-- (name_snapshot, rate_snapshot, is_inclusive, is_compound) are frozen --
-- editing the source tax_rate or tax_class later does not change historical
-- orders.
CREATE TABLE order_line_taxes (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID          NOT NULL REFERENCES organizations(id),
    line_item_id  UUID          NOT NULL REFERENCES order_line_items(id) ON DELETE CASCADE,
    sequence      INTEGER       NOT NULL DEFAULT 0,
    -- ON DELETE SET NULL preserves historical orders when a source rate is
    -- ever hard-deleted; the snapshot fields below carry the audit trail.
    tax_rate_id   UUID          REFERENCES tax_rates(id) ON DELETE SET NULL,
    name_snapshot TEXT          NOT NULL,
    rate_snapshot NUMERIC(6,4)  NOT NULL,
    is_inclusive  BOOLEAN       NOT NULL,
    is_compound   BOOLEAN       NOT NULL DEFAULT FALSE,
    taxable_base  NUMERIC(12,4) NOT NULL,
    amount        NUMERIC(12,4) NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_line_taxes_org_line ON order_line_taxes(org_id, line_item_id);

ALTER TABLE order_line_taxes ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_line_taxes FORCE ROW LEVEL SECURITY;
CREATE POLICY order_line_taxes_org_isolation ON order_line_taxes
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- order_line_adjustments -- discounts / promotions / fees per line item
-- ============================================================================
-- One row per applied adjustment. amount is signed: negative for discount
-- and comp, positive for fee and service_charge. applies_before_tax controls
-- whether this row reduces the taxable base (true = pre-tax discount,
-- false = e.g. post-tax tip). source_id is FK-less by design -- the source
-- promotion or coupon may be deleted while historical orders must remain
-- valid; source_code_snapshot and name_snapshot preserve identification.
CREATE TABLE order_line_adjustments (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               UUID          NOT NULL REFERENCES organizations(id),
    line_item_id         UUID          NOT NULL REFERENCES order_line_items(id) ON DELETE CASCADE,
    sequence             INTEGER       NOT NULL DEFAULT 0,
    kind                 TEXT          NOT NULL
        CHECK (kind IN ('discount','promotion','fee','service_charge','comp')),
    source_type          TEXT          NOT NULL
        CHECK (source_type IN ('manual','promotion_rule','coupon','loyalty','customer_group','auto')),
    source_id            UUID,
    source_code_snapshot TEXT,
    name_snapshot        TEXT          NOT NULL,
    reason               TEXT,
    calculation_type     TEXT          NOT NULL
        CHECK (calculation_type IN ('percentage','fixed_amount','fixed_price')),
    calculation_value    NUMERIC(12,4) NOT NULL,
    amount               NUMERIC(12,4) NOT NULL,
    applies_before_tax   BOOLEAN       NOT NULL DEFAULT TRUE,
    applied_by           UUID          REFERENCES users(id),
    applied_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    approved_by          UUID          REFERENCES users(id),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_line_adjustments_org_line ON order_line_adjustments(org_id, line_item_id);
CREATE INDEX idx_line_adjustments_source ON order_line_adjustments(org_id, source_type, source_id)
    WHERE source_id IS NOT NULL;

ALTER TABLE order_line_adjustments ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_line_adjustments FORCE ROW LEVEL SECURITY;
CREATE POLICY order_line_adjustments_org_isolation ON order_line_adjustments
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- order_adjustments -- order-level discounts / fees / tips / delivery / rounding
-- ============================================================================
-- Order-scoped adjustments. prorate_strategy controls how the engine
-- distributes this adjustment across line items when computing each line's
-- taxable base: pro_rata_taxable_base spreads by each line's current taxable
-- base (the typical order discount), pro_rata_qty spreads by quantity (e.g.
-- packaging fee per item), no_prorate keeps the adjustment at the order
-- level only without touching line tax (e.g. tip).
CREATE TABLE order_adjustments (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               UUID          NOT NULL REFERENCES organizations(id),
    order_id             UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sequence             INTEGER       NOT NULL DEFAULT 0,
    kind                 TEXT          NOT NULL
        CHECK (kind IN ('discount','promotion','fee','service_charge','tip','delivery','rounding','comp')),
    source_type          TEXT          NOT NULL
        CHECK (source_type IN ('manual','promotion_rule','coupon','loyalty','customer_group','auto','system')),
    source_id            UUID,
    source_code_snapshot TEXT,
    name_snapshot        TEXT          NOT NULL,
    reason               TEXT,
    calculation_type     TEXT          NOT NULL
        CHECK (calculation_type IN ('percentage','fixed_amount','fixed_price')),
    calculation_value    NUMERIC(12,4) NOT NULL,
    amount               NUMERIC(12,4) NOT NULL,
    applies_before_tax   BOOLEAN       NOT NULL DEFAULT TRUE,
    prorate_strategy     TEXT          NOT NULL DEFAULT 'pro_rata_taxable_base'
        CHECK (prorate_strategy IN ('pro_rata_taxable_base','pro_rata_qty','no_prorate')),
    applied_by           UUID          REFERENCES users(id),
    applied_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    approved_by          UUID          REFERENCES users(id),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT now(),
    -- Tips, rounding and delivery sit on top of taxes, not inside the tax
    -- base. These constraints make the column defaults safe: an engine that
    -- forgets to override applies_before_tax / prorate_strategy for these
    -- kinds is rejected at insert time rather than silently mis-taxing the
    -- order.
    CONSTRAINT chk_post_tax_kinds
        CHECK (kind NOT IN ('tip','rounding','delivery') OR applies_before_tax = FALSE),
    CONSTRAINT chk_no_prorate_kinds
        CHECK (kind NOT IN ('tip','rounding') OR prorate_strategy = 'no_prorate')
);

CREATE INDEX idx_order_adjustments_org_order ON order_adjustments(org_id, order_id);

ALTER TABLE order_adjustments ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_adjustments FORCE ROW LEVEL SECURITY;
CREATE POLICY order_adjustments_org_isolation ON order_adjustments
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
