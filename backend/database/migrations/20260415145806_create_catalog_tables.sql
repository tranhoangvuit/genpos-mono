-- ============================================================================
-- 003_catalog.sql
-- Categories, products, product variants, tax rates, discounts
-- ============================================================================

-- ============================================================================
-- Categories
-- ============================================================================
CREATE TABLE categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    parent_id       UUID REFERENCES categories(id),
    name            TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    color           TEXT,
    image_url       TEXT,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_org ON categories(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_parent ON categories(org_id, parent_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE categories ENABLE ROW LEVEL SECURITY;
CREATE POLICY categories_org_isolation ON categories
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Products
-- ============================================================================
CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    category_id     UUID REFERENCES categories(id),
    name            TEXT NOT NULL,
    description     TEXT,
    image_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_org ON products(org_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_org_category ON products(org_id, category_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
CREATE POLICY products_org_isolation ON products
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Product variants
-- ============================================================================
CREATE TABLE product_variants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT 'Default',
    sku             TEXT,
    barcode         TEXT,
    price           NUMERIC(12,4) NOT NULL DEFAULT 0,
    cost_price      NUMERIC(12,4) NOT NULL DEFAULT 0,
    track_stock     BOOLEAN NOT NULL DEFAULT TRUE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_variants_org_product ON product_variants(org_id, product_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_variants_org_sku ON product_variants(org_id, sku)
    WHERE sku IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_variants_org_barcode ON product_variants(org_id, barcode)
    WHERE barcode IS NOT NULL AND deleted_at IS NULL;

-- Enable row security level
ALTER TABLE product_variants ENABLE ROW LEVEL SECURITY;
CREATE POLICY variants_org_isolation ON product_variants
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Tax rates
-- ============================================================================
CREATE TABLE tax_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    rate            NUMERIC(6,4) NOT NULL,
    is_inclusive     BOOLEAN NOT NULL DEFAULT FALSE,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tax_rates_org ON tax_rates(org_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE tax_rates ENABLE ROW LEVEL SECURITY;
CREATE POLICY tax_rates_org_isolation ON tax_rates
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- Discounts
-- ============================================================================
CREATE TABLE discounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('percentage', 'fixed')),
    value           NUMERIC(12,4) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_discounts_org ON discounts(org_id) WHERE deleted_at IS NULL;

-- Enable row security level
ALTER TABLE discounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY discounts_org_isolation ON discounts
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);
