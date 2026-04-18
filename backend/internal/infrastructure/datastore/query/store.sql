-- name: CreateStore :one
INSERT INTO stores (org_id, name, address, phone, email, timezone, status)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'),
        sqlc.narg('address'), sqlc.narg('phone'), sqlc.narg('email'),
        sqlc.narg('timezone'), sqlc.arg('status'))
RETURNING id, org_id, name, address, phone, email, timezone, status, created_at, updated_at;

-- name: GetFirstStoreForOrg :one
SELECT id, org_id, name, address, phone, email, timezone, status, created_at, updated_at
FROM stores
WHERE org_id = sqlc.arg('org_id') AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 1;

-- name: ListStores :many
SELECT id, org_id, name, address, phone, email, timezone, status, created_at, updated_at
FROM stores
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateStore :one
UPDATE stores
SET name       = sqlc.arg('name'),
    address    = sqlc.narg('address'),
    phone      = sqlc.narg('phone'),
    email      = sqlc.narg('email'),
    timezone   = sqlc.narg('timezone'),
    status     = sqlc.arg('status'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, address, phone, email, timezone, status, created_at, updated_at;

-- name: SoftDeleteStore :execrows
UPDATE stores
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListStoreRefs :many
SELECT id, name
FROM stores
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: ListVariantPickerItems :many
SELECT v.id,
       p.name AS product_name,
       v.name AS variant_name,
       COALESCE(v.sku, '') AS sku,
       v.price::TEXT AS price,
       v.cost_price::TEXT AS cost_price
FROM product_variants v
JOIN products p ON p.id = v.product_id
WHERE v.is_active = TRUE
  AND v.deleted_at IS NULL
  AND p.is_active = TRUE
  AND p.deleted_at IS NULL
ORDER BY p.name ASC, v.sort_order ASC;
