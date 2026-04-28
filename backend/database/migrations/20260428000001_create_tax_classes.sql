-- Tax classes: a named bucket of one or more tax rates that can be assigned
-- to a product variant. Lets a tenant change the rate(s) for a whole class
-- (e.g. Vietnam VAT 10% -> 8% holiday) by editing the class membership rather
-- than every variant.
--
-- Each variant may belong to at most one class. When a class has multiple
-- rates, sequence + is_compound on tax_class_rates determine how taxes stack:
-- compound rates see the running base + previously applied taxes; non-compound
-- rates always see the original taxable base.

CREATE TABLE tax_classes (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES organizations(id),
    name        TEXT        NOT NULL,
    description TEXT,
    is_default  BOOLEAN     NOT NULL DEFAULT FALSE,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tax_classes_org ON tax_classes(org_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_tax_classes_org_default ON tax_classes(org_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

ALTER TABLE tax_classes ENABLE ROW LEVEL SECURITY;
ALTER TABLE tax_classes FORCE ROW LEVEL SECURITY;
CREATE POLICY tax_classes_org_isolation ON tax_classes
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- Tax class membership. is_compound + sequence encode stacking: a compound
-- rate is computed on (taxable_base + sum of all earlier-sequence rates that
-- have already been applied). Snapshotted onto order_line_taxes at sale time.
CREATE TABLE tax_class_rates (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID        NOT NULL REFERENCES organizations(id),
    tax_class_id UUID        NOT NULL REFERENCES tax_classes(id) ON DELETE CASCADE,
    tax_rate_id  UUID        NOT NULL REFERENCES tax_rates(id),
    sequence     INTEGER     NOT NULL DEFAULT 0,
    is_compound  BOOLEAN     NOT NULL DEFAULT FALSE,
    -- Mirrors tax_classes.deleted_at so soft-deleting a class can propagate
    -- to its rate links without dropping replication for clients that still
    -- need to render historical orders. Engine writes both timestamps in one
    -- transaction.
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_tax_class_rates_unique ON tax_class_rates(tax_class_id, tax_rate_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_tax_class_rates_org ON tax_class_rates(org_id) WHERE deleted_at IS NULL;

ALTER TABLE tax_class_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE tax_class_rates FORCE ROW LEVEL SECURITY;
CREATE POLICY tax_class_rates_org_isolation ON tax_class_rates
    USING (org_id = current_setting('app.current_org_id')::UUID)
    WITH CHECK (org_id = current_setting('app.current_org_id')::UUID);

-- Each variant has 0 or 1 default tax class. NULL means "no automatic tax";
-- the cashier may still attach taxes manually at the line level. The cart
-- engine resolves a non-null variant.tax_class_id into per-line tax rows at
-- sale time and snapshots them onto order_line_taxes.
-- ON DELETE SET NULL: if a tax class is ever hard-deleted (e.g. purge of
-- soft-deleted records), variants drop back to "no automatic tax" rather
-- than blocking the operation.
ALTER TABLE product_variants
    ADD COLUMN tax_class_id UUID REFERENCES tax_classes(id) ON DELETE SET NULL;

CREATE INDEX idx_variants_org_tax_class ON product_variants(org_id, tax_class_id)
    WHERE tax_class_id IS NOT NULL AND deleted_at IS NULL;
