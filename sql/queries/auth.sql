-- name: SaveRefreshToken :execresult
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, NOW(), NOW(), $2, $3, NULL);


-- name: GetUserIdFromRefreshToken :one
SELECT user_id, revoked_at 
FROM refresh_tokens
WHERE token = $1;


-- name: RevokeToken :execresult
UPDATE refresh_tokens 
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1;
