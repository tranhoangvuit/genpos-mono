-- name: ListPaymentMethods :many
SELECT id, org_id, name, type, is_active, sort_order, created_at, updated_at
FROM payment_methods
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, name ASC;

-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (org_id, name, type, is_active, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'), sqlc.arg('type'),
        sqlc.arg('is_active'), sqlc.arg('sort_order'))
RETURNING id, org_id, name, type, is_active, sort_order, created_at, updated_at;

-- name: UpdatePaymentMethod :one
UPDATE payment_methods
SET name       = sqlc.arg('name'),
    type       = sqlc.arg('type'),
    is_active  = sqlc.arg('is_active'),
    sort_order = sqlc.arg('sort_order'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, type, is_active, sort_order, created_at, updated_at;

-- name: SoftDeletePaymentMethod :execrows
UPDATE payment_methods
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
