-- Extend customers with profile fields needed for real-world imports.
ALTER TABLE customers
    ADD COLUMN code          TEXT,
    ADD COLUMN address       TEXT,
    ADD COLUMN company       TEXT,
    ADD COLUMN tax_code      TEXT,
    ADD COLUMN date_of_birth DATE,
    ADD COLUMN gender        TEXT,
    ADD COLUMN facebook      TEXT,
    ADD COLUMN is_active     BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX idx_customers_org_code ON customers(org_id, code)
    WHERE code IS NOT NULL AND deleted_at IS NULL;
