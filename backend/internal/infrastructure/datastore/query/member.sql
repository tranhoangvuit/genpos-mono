-- name: ListMembers :many
SELECT u.id, u.org_id, u.name, u.email, u.phone, u.role_id, r.name AS role_name,
       u.status, u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
ORDER BY u.name ASC;

-- name: GetMemberByID :one
SELECT u.id, u.org_id, u.name, u.email, u.phone, u.role_id, r.name AS role_name,
       u.status, u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = sqlc.arg('id') AND u.deleted_at IS NULL;

-- name: CreateMember :one
INSERT INTO users (org_id, role_id, name, email, phone, password_hash, status)
VALUES (sqlc.arg('org_id'), sqlc.arg('role_id'), sqlc.arg('name'),
        sqlc.narg('email'), sqlc.narg('phone'),
        sqlc.narg('password_hash'), 'active')
RETURNING id;

-- name: UpdateMember :execrows
UPDATE users
SET name       = sqlc.arg('name'),
    phone      = sqlc.narg('phone'),
    role_id    = sqlc.arg('role_id'),
    status     = sqlc.arg('status'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SoftDeleteMember :execrows
UPDATE users
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
