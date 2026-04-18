-- name: ListStockTakeSummaries :many
SELECT st.id,
       s.name AS store_name,
       st.status,
       COUNT(sti.id)::INTEGER AS item_count,
       COUNT(sti.id) FILTER (WHERE sti.counted_qty != sti.expected_qty)::INTEGER AS variance_lines,
       st.created_at,
       st.completed_at
FROM stock_takes st
LEFT JOIN stores s ON s.id = st.store_id AND s.deleted_at IS NULL
LEFT JOIN stock_take_items sti ON sti.stock_take_id = st.id
WHERE st.deleted_at IS NULL
GROUP BY st.id, s.name, st.status, st.created_at, st.completed_at
ORDER BY st.created_at DESC;

-- name: GetStockTakeByID :one
SELECT id, org_id, store_id, user_id, status, notes, completed_at,
       created_at, updated_at
FROM stock_takes
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: CreateStockTake :one
INSERT INTO stock_takes (org_id, store_id, user_id, status, notes)
VALUES (sqlc.arg('org_id'), sqlc.arg('store_id'), sqlc.narg('user_id'),
        'in_progress', sqlc.narg('notes'))
RETURNING id, org_id, store_id, user_id, status, notes, completed_at,
          created_at, updated_at;

-- name: UpdateStockTakeNotes :exec
UPDATE stock_takes
SET notes = sqlc.narg('notes'), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateStockTakeStatus :exec
UPDATE stock_takes
SET status = sqlc.arg('status'),
    completed_at = sqlc.narg('completed_at'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SoftDeleteStockTake :exec
UPDATE stock_takes
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListStockTakeItems :many
SELECT sti.id, sti.org_id, sti.stock_take_id, sti.variant_id,
       sti.expected_qty, sti.counted_qty, sti.created_at, sti.updated_at,
       pv.name AS variant_name, p.name AS product_name
FROM stock_take_items sti
JOIN product_variants pv ON pv.id = sti.variant_id
JOIN products p ON p.id = pv.product_id
WHERE sti.stock_take_id = sqlc.arg('stock_take_id')
ORDER BY p.name ASC, pv.sort_order ASC;

-- name: InsertStockTakeItem :one
INSERT INTO stock_take_items (org_id, stock_take_id, variant_id, expected_qty, counted_qty)
VALUES (sqlc.arg('org_id'), sqlc.arg('stock_take_id'), sqlc.arg('variant_id'),
        sqlc.arg('expected_qty'), sqlc.arg('counted_qty'))
RETURNING id, org_id, stock_take_id, variant_id, expected_qty, counted_qty,
          created_at, updated_at;

-- name: UpdateStockTakeItemCount :exec
UPDATE stock_take_items
SET counted_qty = sqlc.arg('counted_qty'), updated_at = now()
WHERE id = sqlc.arg('id');

-- name: SeedStockTakeItemsFromOnHand :exec
-- Seeds stock_take_items from current on-hand for all active variants in the
-- given store. Expected qty comes from SUM of stock_movements (in - out).
INSERT INTO stock_take_items (org_id, stock_take_id, variant_id, expected_qty, counted_qty)
SELECT sqlc.arg('org_id'), sqlc.arg('stock_take_id'), v.id,
       COALESCE(SUM(CASE WHEN sm.direction='in' THEN sm.quantity ELSE -sm.quantity END), 0),
       0
FROM product_variants v
LEFT JOIN stock_movements sm
    ON sm.variant_id = v.id
   AND sm.store_id = sqlc.arg('store_id')
   AND sm.deleted_at IS NULL
WHERE v.is_active = TRUE
GROUP BY v.id;
