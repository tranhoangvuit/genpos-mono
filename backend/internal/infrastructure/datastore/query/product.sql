-- name: ListProducts :many
SELECT id, org_id, category_id, name, description, image_url, is_active, sort_order, created_at, updated_at
FROM products
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
