-- name: GetOrgBySlug :one
SELECT id, slug, name, created_at, updated_at
FROM organizations
WHERE slug = sqlc.arg('slug');

-- name: GetOrgByID :one
SELECT id, slug, name, created_at, updated_at
FROM organizations
WHERE id = sqlc.arg('id');

-- name: CreateOrg :one
INSERT INTO organizations (slug, name)
VALUES (sqlc.arg('slug'), sqlc.arg('name'))
RETURNING id, slug, name, created_at, updated_at;
