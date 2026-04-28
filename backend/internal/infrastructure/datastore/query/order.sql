-- name: ListOrdersByDateRange :many
SELECT o.id,
       o.order_number,
       o.status,
       o.subtotal::TEXT       AS subtotal,
       o.tax_total::TEXT      AS tax_total,
       o.discount_total::TEXT AS discount_total,
       o.total::TEXT          AS total,
       o.store_id,
       COALESCE(s.name, '')   AS store_name,
       o.register_id,
       o.user_id,
       COALESCE(u.name, '')   AS user_name,
       o.customer_id,
       COALESCE(c.name, '')         AS customer_name,
       o.created_at,
       o.source,
       COALESCE(o.external_id, '')  AS external_id
FROM orders o
LEFT JOIN stores    s ON s.id = o.store_id    AND s.deleted_at IS NULL
LEFT JOIN users     u ON u.id = o.user_id     AND u.deleted_at IS NULL
LEFT JOIN customers c ON c.id = o.customer_id AND c.deleted_at IS NULL
WHERE o.deleted_at IS NULL
  AND o.created_at >= sqlc.arg('date_from')
  AND o.created_at <  sqlc.arg('date_to')
  AND (sqlc.narg('store_id')::UUID IS NULL OR o.store_id = sqlc.narg('store_id'))
ORDER BY o.created_at DESC;

-- name: GetOrderByID :one
SELECT o.id,
       o.order_number,
       o.status,
       o.subtotal::TEXT       AS subtotal,
       o.tax_total::TEXT      AS tax_total,
       o.discount_total::TEXT AS discount_total,
       o.total::TEXT          AS total,
       COALESCE(o.notes, '')  AS notes,
       o.store_id,
       COALESCE(s.name, '')   AS store_name,
       o.register_id,
       o.user_id,
       COALESCE(u.name, '')   AS user_name,
       o.customer_id,
       COALESCE(c.name, '')         AS customer_name,
       o.created_at,
       o.completed_at,
       o.source,
       COALESCE(o.external_id, '')  AS external_id
FROM orders o
LEFT JOIN stores    s ON s.id = o.store_id    AND s.deleted_at IS NULL
LEFT JOIN users     u ON u.id = o.user_id     AND u.deleted_at IS NULL
LEFT JOIN customers c ON c.id = o.customer_id AND c.deleted_at IS NULL
WHERE o.id = sqlc.arg('id') AND o.deleted_at IS NULL;

-- name: ListOrderLineItems :many
SELECT li.id,
       li.variant_id,
       li.product_name,
       li.variant_name,
       COALESCE(li.sku, '')        AS sku,
       li.quantity::TEXT           AS quantity,
       li.unit_price::TEXT         AS unit_price,
       li.tax_rate::TEXT           AS tax_rate,
       li.tax_amount::TEXT         AS tax_amount,
       li.discount_amount::TEXT    AS discount_amount,
       li.line_total::TEXT         AS line_total,
       COALESCE(li.notes, '')      AS notes
FROM order_line_items li
WHERE li.order_id = sqlc.arg('order_id')
ORDER BY li.created_at ASC;

-- name: ListOrderPayments :many
SELECT p.id,
       p.payment_method_id,
       COALESCE(pm.name, '')            AS payment_method_name,
       p.amount::TEXT                   AS amount,
       COALESCE(p.tendered::TEXT, '')::TEXT      AS tendered,
       COALESCE(p.change_amount::TEXT, '')::TEXT AS change_amount,
       COALESCE(p.reference, '')        AS reference,
       p.status,
       p.created_at
FROM payments p
LEFT JOIN payment_methods pm ON pm.id = p.payment_method_id AND pm.deleted_at IS NULL
WHERE p.order_id = sqlc.arg('order_id')
ORDER BY p.created_at ASC;

-- name: GetOrderByExternalID :one
SELECT o.id,
       o.order_number,
       o.status,
       o.subtotal::TEXT       AS subtotal,
       o.tax_total::TEXT      AS tax_total,
       o.discount_total::TEXT AS discount_total,
       o.total::TEXT          AS total,
       COALESCE(o.notes, '')  AS notes,
       o.store_id,
       COALESCE(s.name, '')   AS store_name,
       o.register_id,
       o.user_id,
       COALESCE(u.name, '')   AS user_name,
       o.customer_id,
       COALESCE(c.name, '')         AS customer_name,
       o.created_at,
       o.completed_at,
       o.source,
       COALESCE(o.external_id, '')  AS external_id
FROM orders o
LEFT JOIN stores    s ON s.id = o.store_id    AND s.deleted_at IS NULL
LEFT JOIN users     u ON u.id = o.user_id     AND u.deleted_at IS NULL
LEFT JOIN customers c ON c.id = o.customer_id AND c.deleted_at IS NULL
WHERE o.source = sqlc.arg('source')
  AND o.external_id = sqlc.arg('external_id')
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: InsertOrder :one
INSERT INTO orders (
    org_id, store_id, register_id, customer_id, user_id,
    order_number, status, subtotal, tax_total, discount_total, total,
    notes, completed_at, source, external_id, external_source_id
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('store_id'), sqlc.narg('register_id'),
    sqlc.narg('customer_id'), sqlc.narg('user_id'),
    sqlc.arg('order_number'), sqlc.arg('status'),
    sqlc.arg('subtotal')::NUMERIC, sqlc.arg('tax_total')::NUMERIC,
    sqlc.arg('discount_total')::NUMERIC, sqlc.arg('total')::NUMERIC,
    sqlc.narg('notes'), sqlc.narg('completed_at'),
    sqlc.arg('source'), sqlc.narg('external_id'), sqlc.narg('external_source_id')
)
RETURNING id, org_id, store_id, register_id, customer_id, user_id,
          order_number, status,
          subtotal::TEXT       AS subtotal,
          tax_total::TEXT      AS tax_total,
          discount_total::TEXT AS discount_total,
          total::TEXT          AS total,
          COALESCE(notes, '')  AS notes,
          completed_at, source,
          COALESCE(external_id, '') AS external_id,
          created_at, updated_at;

-- name: InsertOrderLineItem :one
INSERT INTO order_line_items (
    org_id, order_id, variant_id, product_name, variant_name, sku,
    quantity, unit_price, tax_rate, tax_amount, discount_amount, line_total, notes
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('order_id'), sqlc.narg('variant_id'),
    sqlc.arg('product_name'), sqlc.arg('variant_name'), sqlc.narg('sku'),
    sqlc.arg('quantity')::NUMERIC, sqlc.arg('unit_price')::NUMERIC,
    sqlc.arg('tax_rate')::NUMERIC, sqlc.arg('tax_amount')::NUMERIC,
    sqlc.arg('discount_amount')::NUMERIC, sqlc.arg('line_total')::NUMERIC,
    sqlc.narg('notes')
)
RETURNING id;

-- name: InsertOrderLineTax :exec
INSERT INTO order_line_taxes (
    org_id, line_item_id, sequence, tax_rate_id,
    name_snapshot, rate_snapshot, is_inclusive, is_compound,
    taxable_base, amount
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('line_item_id'), sqlc.arg('sequence'),
    sqlc.narg('tax_rate_id'),
    sqlc.arg('name_snapshot'), sqlc.arg('rate_snapshot')::NUMERIC,
    sqlc.arg('is_inclusive'), sqlc.arg('is_compound'),
    sqlc.arg('taxable_base')::NUMERIC, sqlc.arg('amount')::NUMERIC
);

-- name: InsertOrderLineAdjustment :exec
INSERT INTO order_line_adjustments (
    org_id, line_item_id, sequence, kind, source_type, source_id,
    source_code_snapshot, name_snapshot, reason,
    calculation_type, calculation_value, amount,
    applies_before_tax, applied_by, approved_by
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('line_item_id'), sqlc.arg('sequence'),
    sqlc.arg('kind'), sqlc.arg('source_type'), sqlc.narg('source_id'),
    sqlc.narg('source_code_snapshot'), sqlc.arg('name_snapshot'), sqlc.narg('reason'),
    sqlc.arg('calculation_type'), sqlc.arg('calculation_value')::NUMERIC,
    sqlc.arg('amount')::NUMERIC,
    sqlc.arg('applies_before_tax'), sqlc.narg('applied_by'), sqlc.narg('approved_by')
);

-- name: InsertOrderAdjustment :exec
INSERT INTO order_adjustments (
    org_id, order_id, sequence, kind, source_type, source_id,
    source_code_snapshot, name_snapshot, reason,
    calculation_type, calculation_value, amount,
    applies_before_tax, prorate_strategy, applied_by, approved_by
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('order_id'), sqlc.arg('sequence'),
    sqlc.arg('kind'), sqlc.arg('source_type'), sqlc.narg('source_id'),
    sqlc.narg('source_code_snapshot'), sqlc.arg('name_snapshot'), sqlc.narg('reason'),
    sqlc.arg('calculation_type'), sqlc.arg('calculation_value')::NUMERIC,
    sqlc.arg('amount')::NUMERIC,
    sqlc.arg('applies_before_tax'), sqlc.arg('prorate_strategy'),
    sqlc.narg('applied_by'), sqlc.narg('approved_by')
);

-- name: ListOrderLineTaxesByOrderID :many
SELECT olt.id, olt.line_item_id, olt.sequence, olt.tax_rate_id,
       olt.name_snapshot, olt.rate_snapshot::TEXT AS rate_snapshot,
       olt.is_inclusive, olt.is_compound,
       olt.taxable_base::TEXT AS taxable_base,
       olt.amount::TEXT       AS amount
FROM order_line_taxes olt
JOIN order_line_items li ON li.id = olt.line_item_id
WHERE li.order_id = sqlc.arg('order_id')
ORDER BY olt.line_item_id, olt.sequence;

-- name: ListOrderLineAdjustmentsByOrderID :many
SELECT a.id, a.line_item_id, a.sequence, a.kind, a.source_type, a.source_id,
       COALESCE(a.source_code_snapshot, '') AS source_code_snapshot,
       a.name_snapshot,
       COALESCE(a.reason, '')               AS reason,
       a.calculation_type,
       a.calculation_value::TEXT AS calculation_value,
       a.amount::TEXT            AS amount,
       a.applies_before_tax,
       a.applied_by, a.applied_at, a.approved_by
FROM order_line_adjustments a
JOIN order_line_items li ON li.id = a.line_item_id
WHERE li.order_id = sqlc.arg('order_id')
ORDER BY a.line_item_id, a.sequence;

-- name: ListOrderAdjustmentsByOrderID :many
SELECT a.id, a.sequence, a.kind, a.source_type, a.source_id,
       COALESCE(a.source_code_snapshot, '') AS source_code_snapshot,
       a.name_snapshot,
       COALESCE(a.reason, '')               AS reason,
       a.calculation_type,
       a.calculation_value::TEXT AS calculation_value,
       a.amount::TEXT            AS amount,
       a.applies_before_tax, a.prorate_strategy,
       a.applied_by, a.applied_at, a.approved_by
FROM order_adjustments a
WHERE a.order_id = sqlc.arg('order_id')
ORDER BY a.sequence;

-- name: InsertOrderPayment :exec
INSERT INTO payments (
    org_id, order_id, payment_method_id, amount, tendered, change_amount,
    reference, status
) VALUES (
    sqlc.arg('org_id'), sqlc.arg('order_id'), sqlc.arg('payment_method_id'),
    sqlc.arg('amount')::NUMERIC, sqlc.narg('tendered')::NUMERIC,
    sqlc.narg('change_amount')::NUMERIC, sqlc.narg('reference'),
    sqlc.arg('status')
);

-- name: GetFirstStoreIDForOrg :one
SELECT id
FROM stores
WHERE org_id = sqlc.arg('org_id') AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 1;
