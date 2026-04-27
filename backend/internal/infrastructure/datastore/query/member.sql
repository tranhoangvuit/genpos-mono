-- name: ListMembers :many
SELECT u.id, u.org_id, u.name, u.email, u.phone, u.role_id, r.name AS role_name,
       u.status, u.all_stores, u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
ORDER BY u.name ASC;

-- name: GetMemberByID :one
SELECT u.id, u.org_id, u.name, u.email, u.phone, u.role_id, r.name AS role_name,
       u.status, u.all_stores, u.created_at, u.updated_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = sqlc.arg('id') AND u.deleted_at IS NULL;

-- name: CreateMember :one
INSERT INTO users (org_id, role_id, name, email, phone, password_hash, all_stores, status)
VALUES (sqlc.arg('org_id'), sqlc.arg('role_id'), sqlc.arg('name'),
        sqlc.narg('email'), sqlc.narg('phone'),
        sqlc.narg('password_hash'), sqlc.arg('all_stores'), 'active')
RETURNING id;

-- name: UpdateMember :execrows
UPDATE users
SET name       = sqlc.arg('name'),
    phone      = sqlc.narg('phone'),
    role_id    = sqlc.arg('role_id'),
    status     = sqlc.arg('status'),
    all_stores = sqlc.arg('all_stores'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: SoftDeleteMember :execrows
UPDATE users
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListMemberStoreIDs :many
SELECT store_id
FROM user_stores
WHERE user_id = sqlc.arg('user_id')
ORDER BY store_id;

-- name: DeleteMemberStores :exec
DELETE FROM user_stores WHERE user_id = sqlc.arg('user_id');

-- name: InsertMemberStore :exec
INSERT INTO user_stores (org_id, user_id, store_id)
VALUES (sqlc.arg('org_id'), sqlc.arg('user_id'), sqlc.arg('store_id'))
ON CONFLICT (org_id, user_id, store_id) DO NOTHING;

-- name: HasStoreAccess :one
SELECT EXISTS (
    SELECT 1 FROM users u
    WHERE u.id = sqlc.arg('user_id')
      AND u.deleted_at IS NULL
      AND u.status = 'active'
      AND (
        u.all_stores
        OR EXISTS (
          SELECT 1 FROM user_stores us
          WHERE us.user_id = u.id AND us.store_id = sqlc.arg('store_id')
        )
      )
) AS has_access;
