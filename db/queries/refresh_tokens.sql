-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
    id,
    user_id,
    token_hash,
    expires_at
)
VALUES (?, ?, ?, ?);

-- name: GetRefreshTokenByHash :one
SELECT
    id,
    user_id,
    token_hash,
    expires_at,
    created_at,
    revoked_at
FROM refresh_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = ?
    AND revoked_at IS NULL;

-- name: RevokeRefreshTokenIfActive :execresult
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND revoked_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP;