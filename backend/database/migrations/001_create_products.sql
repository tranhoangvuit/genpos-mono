CREATE TABLE products (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    sku         TEXT        NOT NULL,
    price_cents BIGINT      NOT NULL DEFAULT 0,
    active      BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_org_id ON products (org_id);
CREATE UNIQUE INDEX idx_products_org_sku ON products (org_id, sku);

-- Enable RLS for multi-tenancy
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE products FORCE ROW LEVEL SECURITY;

CREATE POLICY products_tenant_isolation ON products
    USING (org_id = current_setting('app.current_org_id', true));
