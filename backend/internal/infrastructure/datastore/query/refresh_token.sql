-- name: GetRefreshTokenByHash :one
SELECT id, user_id, org_id, token_hash, expires_at, revoked_at, user_agent, created_at
FROM refresh_tokens
WHERE token_hash = sqlc.arg('token_hash');

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, org_id, token_hash, expires_at, user_agent)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('org_id'),
    sqlc.arg('token_hash'),
    sqlc.arg('expires_at'),
    sqlc.arg('user_agent')
)
RETURNING id, user_id, org_id, token_hash, expires_at, revoked_at, user_agent, created_at;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = sqlc.arg('revoked_at')
WHERE id = sqlc.arg('id');
