-- name: ListProducts :many
SELECT id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at
FROM products
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListProductSummaries :many
SELECT p.id,
       p.name,
       p.category_id,
       c.name  AS category_name,
       COALESCE(MIN(v.price)::TEXT, '0') AS price,
       COUNT(v.id)::INTEGER AS variant_count,
       p.is_active
FROM products p
LEFT JOIN categories c
       ON c.id = p.category_id AND c.deleted_at IS NULL
LEFT JOIN product_variants v
       ON v.product_id = p.id AND v.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.category_id, c.name, p.is_active, p.sort_order, p.created_at
ORDER BY p.sort_order ASC, p.created_at DESC;
