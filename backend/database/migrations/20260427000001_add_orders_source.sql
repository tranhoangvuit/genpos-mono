-- Track the channel that produced each order. POS orders are rung up by a
-- cashier and require user_id; orders pulled from external channels (Shopify,
-- WooCommerce, online store) have no operator and must be allowed to leave
-- user_id NULL. external_id + source make imports idempotent.

ALTER TABLE orders
    ADD COLUMN source             TEXT NOT NULL DEFAULT 'pos',
    ADD COLUMN external_id        TEXT,
    ADD COLUMN external_source_id TEXT;

ALTER TABLE orders
    ADD CONSTRAINT orders_source_check
        CHECK (source IN ('pos','online_store','shopify','woocommerce','manual','import'));

ALTER TABLE orders
    ADD CONSTRAINT orders_user_id_required_for_pos
        CHECK (source <> 'pos' OR user_id IS NOT NULL);

CREATE INDEX idx_orders_org_source
    ON orders(org_id, source, created_at DESC) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_orders_org_source_external
    ON orders(org_id, source, external_id) WHERE external_id IS NOT NULL;
