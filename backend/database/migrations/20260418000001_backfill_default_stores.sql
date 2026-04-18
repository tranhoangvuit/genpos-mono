-- ============================================================================
-- 010_backfill_default_stores.sql
-- Ensure every organization has at least one active store. Required by
-- inventory features (purchase orders, stock takes) which reference store_id.
-- ============================================================================

INSERT INTO stores (org_id, name, status)
SELECT o.id, 'Main Store', 'active'
FROM organizations o
WHERE NOT EXISTS (
    SELECT 1 FROM stores s WHERE s.org_id = o.id AND s.deleted_at IS NULL
);
