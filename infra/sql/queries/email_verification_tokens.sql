-- name: InsertEmailVerificationToken :execrows
INSERT INTO email_verification_tokens(user_id, token_hash, email, expires_at, used_at, revoked_at, request_ip, request_user_agent, consumed_ip, consumed_user_agent)
    VALUES (sqlc.arg('UserID'), sqlc.arg('TokenHash'), sqlc.arg('Email'), sqlc.arg('ExpiresAt'), sqlc.narg('UsedAt'), sqlc.narg('RevokedAt'), sqlc.narg('RequestIP'), sqlc.narg('RequestUserAgent'), sqlc.narg('ConsumedIP'), sqlc.narg('ConsumedUserAgent'));

-- name: GetEmailVerificationTokenByHash :one
SELECT
    evt.id,
    evt.user_id,
    evt.token_hash,
    evt.email,
    evt.expires_at,
    evt.used_at,
    evt.revoked_at,
    evt.request_ip,
    evt.request_user_agent,
    evt.consumed_ip,
    evt.consumed_user_agent
FROM
    email_verification_tokens evt
WHERE
    evt.token_hash = sqlc.arg('TokenHash');

-- name: MarkEmailVerificationTokenAsUsed :execrows
UPDATE
    email_verification_tokens
SET
    used_at = sqlc.arg('UsedAt'),
    consumed_ip = sqlc.arg('ConsumedIP'),
    consumed_user_agent = sqlc.arg('ConsumedUserAgent')
WHERE
    token_hash = sqlc.arg('TokenHash')
    AND used_at IS NULL
    AND revoked_at IS NULL;

-- name: RevokeEmailVerificationToken :execrows
UPDATE
    email_verification_tokens
SET
    revoked_at = sqlc.arg('RevokedAt')
WHERE
    token_hash = sqlc.arg('TokenHash')
    AND used_at IS NULL
    AND revoked_at IS NULL;

-- name: ListActiveEmailVerificationTokensByUserID :many
SELECT
    evt.id,
    evt.user_id,
    evt.token_hash,
    evt.email,
    evt.expires_at,
    evt.used_at,
    evt.revoked_at,
    evt.request_ip,
    evt.request_user_agent,
    evt.consumed_ip,
    evt.consumed_user_agent
FROM
    email_verification_tokens evt
WHERE
    evt.user_id = sqlc.arg('UserID')
    AND evt.used_at IS NULL
    AND evt.revoked_at IS NULL
    AND evt.expires_at > NOW()
ORDER BY
    evt.expires_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

