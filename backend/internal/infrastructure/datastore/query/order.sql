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
       COALESCE(c.name, '')   AS customer_name,
       o.created_at
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
       COALESCE(c.name, '')   AS customer_name,
       o.created_at,
       o.completed_at
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
