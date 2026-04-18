-- name: ListSuppliers :many
SELECT id, org_id, name, contact_name, email, phone, address, notes, created_at, updated_at
FROM suppliers
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: GetSupplierByID :one
SELECT id, org_id, name, contact_name, email, phone, address, notes, created_at, updated_at
FROM suppliers
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: CreateSupplier :one
INSERT INTO suppliers (org_id, name, contact_name, email, phone, address, notes)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'), sqlc.narg('contact_name'),
        sqlc.narg('email'), sqlc.narg('phone'), sqlc.narg('address'), sqlc.narg('notes'))
RETURNING id, org_id, name, contact_name, email, phone, address, notes, created_at, updated_at;

-- name: UpdateSupplier :one
UPDATE suppliers
SET name         = sqlc.arg('name'),
    contact_name = sqlc.narg('contact_name'),
    email        = sqlc.narg('email'),
    phone        = sqlc.narg('phone'),
    address      = sqlc.narg('address'),
    notes        = sqlc.narg('notes'),
    updated_at   = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, contact_name, email, phone, address, notes, created_at, updated_at;

-- name: SoftDeleteSupplier :exec
UPDATE suppliers
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
