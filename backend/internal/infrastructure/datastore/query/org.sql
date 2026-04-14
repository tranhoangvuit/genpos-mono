-- name: GetOrgBySlug :one
SELECT id, slug, name, created_at, updated_at
FROM orgs
WHERE slug = sqlc.arg('slug');

-- name: GetOrgByID :one
SELECT id, slug, name, created_at, updated_at
FROM orgs
WHERE id = sqlc.arg('id');

-- name: CreateOrg :one
INSERT INTO orgs (slug, name)
VALUES (sqlc.arg('slug'), sqlc.arg('name'))
RETURNING id, slug, name, created_at, updated_at;
