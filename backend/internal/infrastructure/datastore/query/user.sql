-- name: GetUserByEmail :one
SELECT u.id, u.org_id, u.email, u.password_hash, u.name,
       u.role_id, r.name AS role_name, r.permissions AS role_permissions,
       u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE lower(u.email) = lower(sqlc.arg('email'));

-- name: GetUserByID :one
SELECT u.id, u.org_id, u.email, u.password_hash, u.name,
       u.role_id, r.name AS role_name, r.permissions AS role_permissions,
       u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = sqlc.arg('id');

-- name: CreateUser :one
INSERT INTO users (org_id, role_id, email, password_hash, name)
VALUES (
    sqlc.arg('org_id'),
    sqlc.arg('role_id'),
    sqlc.arg('email'),
    sqlc.arg('password_hash'),
    sqlc.arg('name')
)
RETURNING id, org_id, role_id, email, password_hash, name, created_at, updated_at;
