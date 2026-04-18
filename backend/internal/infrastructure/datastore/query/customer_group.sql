-- name: ListCustomerGroups :many
SELECT id, org_id, name, description, discount_type, discount_value, created_at, updated_at
FROM customer_groups
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: GetCustomerGroupByID :one
SELECT id, org_id, name, description, discount_type, discount_value, created_at, updated_at
FROM customer_groups
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetCustomerGroupByName :one
SELECT id, org_id, name, description, discount_type, discount_value, created_at, updated_at
FROM customer_groups
WHERE name = sqlc.arg('name') AND deleted_at IS NULL
LIMIT 1;

-- name: CreateCustomerGroup :one
INSERT INTO customer_groups (org_id, name, description, discount_type, discount_value)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'), sqlc.narg('description'),
        sqlc.narg('discount_type'), sqlc.narg('discount_value'))
RETURNING id, org_id, name, description, discount_type, discount_value, created_at, updated_at;

-- name: UpdateCustomerGroup :one
UPDATE customer_groups
SET name           = sqlc.arg('name'),
    description    = sqlc.narg('description'),
    discount_type  = sqlc.narg('discount_type'),
    discount_value = sqlc.narg('discount_value'),
    updated_at     = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, description, discount_type, discount_value, created_at, updated_at;

-- name: SoftDeleteCustomerGroup :exec
UPDATE customer_groups
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
