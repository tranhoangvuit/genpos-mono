-- name: ListTaxClasses :many
SELECT id, org_id, name, description, is_default, sort_order, created_at, updated_at
FROM tax_classes
WHERE deleted_at IS NULL
ORDER BY is_default DESC, sort_order ASC, name ASC;

-- name: GetTaxClass :one
SELECT id, org_id, name, description, is_default, sort_order, created_at, updated_at
FROM tax_classes
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListTaxClassRatesByClassIDs :many
SELECT id, org_id, tax_class_id, tax_rate_id, sequence, is_compound, created_at, updated_at
FROM tax_class_rates
WHERE tax_class_id = ANY(sqlc.arg('class_ids')::UUID[]) AND deleted_at IS NULL
ORDER BY tax_class_id, sequence;

-- name: CreateTaxClass :one
INSERT INTO tax_classes (org_id, name, description, is_default, sort_order)
VALUES (sqlc.arg('org_id'), sqlc.arg('name'), sqlc.arg('description'),
        sqlc.arg('is_default'), sqlc.arg('sort_order'))
RETURNING id, org_id, name, description, is_default, sort_order, created_at, updated_at;

-- name: UpdateTaxClass :one
UPDATE tax_classes
SET name        = sqlc.arg('name'),
    description = sqlc.arg('description'),
    is_default  = sqlc.arg('is_default'),
    sort_order  = sqlc.arg('sort_order'),
    updated_at  = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, name, description, is_default, sort_order, created_at, updated_at;

-- name: SoftDeleteTaxClass :execrows
UPDATE tax_classes
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ClearDefaultTaxClasses :exec
UPDATE tax_classes
SET is_default = FALSE, updated_at = now()
WHERE is_default = TRUE AND deleted_at IS NULL;

-- name: InsertTaxClassRate :one
INSERT INTO tax_class_rates (org_id, tax_class_id, tax_rate_id, sequence, is_compound)
VALUES (sqlc.arg('org_id'), sqlc.arg('tax_class_id'), sqlc.arg('tax_rate_id'),
        sqlc.arg('sequence'), sqlc.arg('is_compound'))
RETURNING id, org_id, tax_class_id, tax_rate_id, sequence, is_compound, created_at, updated_at;

-- name: SoftDeleteTaxClassRatesByClassID :exec
UPDATE tax_class_rates
SET deleted_at = now(), updated_at = now()
WHERE tax_class_id = sqlc.arg('tax_class_id') AND deleted_at IS NULL;
