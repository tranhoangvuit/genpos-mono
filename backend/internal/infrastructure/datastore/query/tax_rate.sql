-- name: ListTaxRates :many
SELECT id, org_id, name, rate, is_inclusive, is_default, created_at, updated_at
FROM tax_rates
WHERE deleted_at IS NULL
ORDER BY is_default DESC, name ASC;

-- name: CreateTaxRate :one
INSERT INTO tax_rates (org_id, name, rate, is_inclusive, is_default)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'), sqlc.arg('rate'),
        sqlc.arg('is_inclusive'), sqlc.arg('is_default'))
RETURNING id, org_id, name, rate, is_inclusive, is_default, created_at, updated_at;

-- name: UpdateTaxRate :one
UPDATE tax_rates
SET name         = sqlc.arg('name'),
    rate         = sqlc.arg('rate'),
    is_inclusive = sqlc.arg('is_inclusive'),
    is_default   = sqlc.arg('is_default'),
    updated_at   = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, rate, is_inclusive, is_default, created_at, updated_at;

-- name: SoftDeleteTaxRate :execrows
UPDATE tax_rates
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ClearDefaultTaxRates :exec
UPDATE tax_rates
SET is_default = FALSE, updated_at = now()
WHERE is_default = TRUE AND deleted_at IS NULL;
