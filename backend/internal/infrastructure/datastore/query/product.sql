-- name: ListProducts :many
SELECT id, org_id, name, sku, price_cents, active, created_at, updated_at
FROM products
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
