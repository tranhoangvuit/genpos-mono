-- name: ListCategories :many
SELECT id, org_id, parent_id, name, sort_order, color, image_url, created_at, updated_at
FROM categories
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, name ASC;

-- name: GetCategoryByID :one
SELECT id, org_id, parent_id, name, sort_order, color, image_url, created_at, updated_at
FROM categories
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetCategoryByName :one
SELECT id, org_id, parent_id, name, sort_order, color, image_url, created_at, updated_at
FROM categories
WHERE name = sqlc.arg('name') AND deleted_at IS NULL
LIMIT 1;

-- name: CreateCategory :one
INSERT INTO categories (org_id, parent_id, name, sort_order, color)
VALUES (sqlc.arg('org_id'), sqlc.narg('parent_id'), sqlc.arg('name'),
        sqlc.arg('sort_order'), sqlc.narg('color'))
RETURNING id, org_id, parent_id, name, sort_order, color, image_url, created_at, updated_at;

-- name: UpdateCategory :one
UPDATE categories
SET name       = sqlc.arg('name'),
    parent_id  = sqlc.narg('parent_id'),
    color      = sqlc.narg('color'),
    sort_order = sqlc.arg('sort_order'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, org_id, parent_id, name, sort_order, color, image_url, created_at, updated_at;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
