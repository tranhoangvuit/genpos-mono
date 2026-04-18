-- name: ListPurchaseOrderSummaries :many
SELECT po.id,
       po.po_number,
       po.supplier_name,
       po.status,
       s.name AS store_name,
       po.expected_at,
       COUNT(poi.id)::INTEGER AS item_count,
       COALESCE(SUM(poi.quantity_ordered * poi.cost_price)::TEXT, '0') AS total,
       po.created_at
FROM purchase_orders po
LEFT JOIN stores s ON s.id = po.store_id AND s.deleted_at IS NULL
LEFT JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
WHERE po.deleted_at IS NULL
GROUP BY po.id, po.po_number, po.supplier_name, po.status, s.name,
         po.expected_at, po.created_at
ORDER BY po.created_at DESC;

-- name: GetPurchaseOrderByID :one
SELECT id, org_id, store_id, user_id, po_number, supplier_name, status, notes,
       expected_at, received_at, created_at, updated_at
FROM purchase_orders
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: CountPurchaseOrdersForPrefix :one
SELECT COUNT(*)::int AS n
FROM purchase_orders
WHERE org_id = sqlc.arg('org_id')
  AND po_number LIKE sqlc.arg('prefix');

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (org_id, store_id, user_id, po_number, supplier_name, status, notes, expected_at)
VALUES (sqlc.arg('org_id'), sqlc.arg('store_id'), sqlc.narg('user_id'),
        sqlc.arg('po_number'), sqlc.narg('supplier_name'),
        'draft', sqlc.narg('notes'), sqlc.narg('expected_at'))
RETURNING id, org_id, store_id, user_id, po_number, supplier_name, status, notes,
          expected_at, received_at, created_at, updated_at;

-- name: UpdatePurchaseOrder :one
UPDATE purchase_orders
SET store_id      = sqlc.arg('store_id'),
    supplier_name = sqlc.narg('supplier_name'),
    notes         = sqlc.narg('notes'),
    expected_at   = sqlc.narg('expected_at'),
    updated_at    = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, store_id, user_id, po_number, supplier_name, status, notes,
          expected_at, received_at, created_at, updated_at;

-- name: UpdatePurchaseOrderStatus :exec
UPDATE purchase_orders
SET status      = sqlc.arg('status'),
    received_at = sqlc.narg('received_at'),
    updated_at  = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SoftDeletePurchaseOrder :exec
UPDATE purchase_orders
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListPurchaseOrderItems :many
SELECT poi.id, poi.org_id, poi.purchase_order_id, poi.variant_id,
       poi.quantity_ordered, poi.quantity_received, poi.cost_price,
       poi.created_at, poi.updated_at,
       pv.name AS variant_name, p.name AS product_name
FROM purchase_order_items poi
JOIN product_variants pv ON pv.id = poi.variant_id
JOIN products p ON p.id = pv.product_id
WHERE poi.purchase_order_id = sqlc.arg('purchase_order_id')
ORDER BY poi.created_at ASC;

-- name: GetPurchaseOrderItemByID :one
SELECT id, org_id, purchase_order_id, variant_id, quantity_ordered,
       quantity_received, cost_price, created_at, updated_at
FROM purchase_order_items
WHERE id = sqlc.arg('id');

-- name: InsertPurchaseOrderItem :one
INSERT INTO purchase_order_items (org_id, purchase_order_id, variant_id, quantity_ordered, cost_price)
VALUES (sqlc.arg('org_id'), sqlc.arg('purchase_order_id'), sqlc.arg('variant_id'),
        sqlc.arg('quantity_ordered'), sqlc.arg('cost_price'))
RETURNING id, org_id, purchase_order_id, variant_id, quantity_ordered,
          quantity_received, cost_price, created_at, updated_at;

-- name: DeletePurchaseOrderItemsByPO :exec
DELETE FROM purchase_order_items
WHERE purchase_order_id = sqlc.arg('purchase_order_id');

-- name: AddPurchaseOrderItemReceived :exec
UPDATE purchase_order_items
SET quantity_received = quantity_received + sqlc.arg('delta'),
    updated_at        = now()
WHERE id = sqlc.arg('id');
