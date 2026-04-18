-- ============================================================================
-- 009_catalog_extensions.sql
-- Product options/values/variant-value-joins, product images,
-- suppliers (pre-land for POs), customer_groups + members (pre-land)
-- ============================================================================

-- ============================================================================
-- Product options (e.g. "Size", "Color")
-- ============================================================================
CREATE TABLE product_options (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_options_product ON product_options(org_id, product_id);

ALTER TABLE product_options ENABLE ROW LEVEL SECURITY;
CREATE POLICY product_options_org_isolation ON product_options
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Product option values (e.g. "S", "M", "L")
-- ============================================================================
CREATE TABLE product_option_values (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    option_id       UUID NOT NULL REFERENCES product_options(id) ON DELETE CASCADE,
    value           TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_option_values_option ON product_option_values(org_id, option_id);

ALTER TABLE product_option_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY product_option_values_org_isolation ON product_option_values
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Variant <-> option value join
-- ============================================================================
CREATE TABLE product_variant_option_values (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    variant_id      UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    option_value_id UUID NOT NULL REFERENCES product_option_values(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_variant_option_values_unique
    ON product_variant_option_values(variant_id, option_value_id);

ALTER TABLE product_variant_option_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY product_variant_option_values_org_isolation ON product_variant_option_values
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Product images
-- ============================================================================
CREATE TABLE product_images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id      UUID REFERENCES product_variants(id) ON DELETE SET NULL,
    url             TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_images_product ON product_images(org_id, product_id);

ALTER TABLE product_images ENABLE ROW LEVEL SECURITY;
CREATE POLICY product_images_org_isolation ON product_images
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Suppliers (pre-land for purchase orders)
-- ============================================================================
CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    contact_name    TEXT,
    email           TEXT,
    phone           TEXT,
    address         TEXT,
    notes           TEXT,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_suppliers_org ON suppliers(org_id) WHERE deleted_at IS NULL;

ALTER TABLE suppliers ENABLE ROW LEVEL SECURITY;
CREATE POLICY suppliers_org_isolation ON suppliers
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Customer groups (pre-land for customers feature)
-- ============================================================================
CREATE TABLE customer_groups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    description     TEXT,
    discount_type   TEXT CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value  NUMERIC(12,4),
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_customer_groups_org ON customer_groups(org_id) WHERE deleted_at IS NULL;

ALTER TABLE customer_groups ENABLE ROW LEVEL SECURITY;
CREATE POLICY customer_groups_org_isolation ON customer_groups
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Customer group members
-- ============================================================================
CREATE TABLE customer_group_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    group_id        UUID NOT NULL REFERENCES customer_groups(id) ON DELETE CASCADE,
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_customer_group_members_unique
    ON customer_group_members(group_id, customer_id);

ALTER TABLE customer_group_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY customer_group_members_org_isolation ON customer_group_members
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
