-- name: InsertPasswordRecoveryToken :execrows
INSERT INTO password_recovery_tokens(user_id, token_hash, requested_email, expires_at, used_at, revoked_at, request_ip, request_user_agent, triggered_by_user_id)
    VALUES (sqlc.arg('UserID'), sqlc.arg('TokenHash'), sqlc.arg('RequestedEmail'), sqlc.arg('ExpiresAt'), sqlc.narg('UsedAt'), sqlc.narg('RevokedAt'), sqlc.narg('RequestIP'), sqlc.narg('RequestUserAgent'), sqlc.narg('TriggeredByUserID'));

-- name: GetPasswordRecoveryTokenByHash :one
SELECT
    prt.id,
    prt.user_id,
    prt.token_hash,
    prt.requested_email,
    prt.expires_at,
    prt.used_at,
    prt.revoked_at,
    prt.request_ip,
    prt.request_user_agent,
    prt.triggered_by_user_id
FROM
    password_recovery_tokens prt
WHERE
    prt.token_hash = sqlc.arg('TokenHash');

-- name: MarkPasswordRecoveryTokenAsUsed :execrows
UPDATE
    password_recovery_tokens
SET
    used_at = sqlc.arg('UsedAt')
WHERE
    token_hash = sqlc.arg('TokenHash')
    AND used_at IS NULL
    AND revoked_at IS NULL;

-- name: RevokePasswordRecoveryToken :execrows
UPDATE
    password_recovery_tokens
SET
    revoked_at = sqlc.arg('RevokedAt')
WHERE
    token_hash = sqlc.arg('TokenHash')
    AND used_at IS NULL
    AND revoked_at IS NULL;

-- name: ListActivePasswordRecoveryTokensByUserID :many
SELECT
    prt.id,
    prt.user_id,
    prt.token_hash,
    prt.requested_email,
    prt.expires_at,
    prt.used_at,
    prt.revoked_at,
    prt.request_ip,
    prt.request_user_agent,
    prt.triggered_by_user_id
FROM
    password_recovery_tokens prt
WHERE
    prt.user_id = sqlc.arg('UserID')
    AND prt.used_at IS NULL
    AND prt.revoked_at IS NULL
    AND prt.expires_at > NOW()
ORDER BY
    prt.expires_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

