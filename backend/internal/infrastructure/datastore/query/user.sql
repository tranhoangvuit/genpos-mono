-- name: GetUserByEmail :one
SELECT id, org_id, email, password_hash, name, role, created_at, updated_at
FROM users
WHERE lower(email) = lower(sqlc.arg('email'));

-- name: GetUserByID :one
SELECT id, org_id, email, password_hash, name, role, created_at, updated_at
FROM users
WHERE id = sqlc.arg('id');

-- name: CreateUser :one
INSERT INTO users (org_id, email, password_hash, name, role)
VALUES (
    sqlc.arg('org_id'),
    sqlc.arg('email'),
    sqlc.arg('password_hash'),
    sqlc.arg('name'),
    sqlc.arg('role')
)
RETURNING id, org_id, email, password_hash, name, role, created_at, updated_at;
