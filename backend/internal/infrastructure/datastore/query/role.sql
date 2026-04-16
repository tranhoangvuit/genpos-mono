-- name: CreateRole :one
INSERT INTO roles (org_id, name, permissions, is_system)
VALUES (
    sqlc.arg('org_id'),
    sqlc.arg('name'),
    sqlc.arg('permissions'),
    sqlc.arg('is_system')
)
RETURNING id, org_id, name, permissions, is_system, created_at, updated_at;

-- name: GetRoleByOrgAndName :one
SELECT id, org_id, name, permissions, is_system, created_at, updated_at
FROM roles
WHERE org_id = sqlc.arg('org_id') AND name = sqlc.arg('name') AND deleted_at IS NULL;

-- name: ListRolesByOrg :many
SELECT id, org_id, name, permissions, is_system, created_at, updated_at
FROM roles
WHERE org_id = sqlc.arg('org_id') AND deleted_at IS NULL
ORDER BY is_system DESC, name ASC;
