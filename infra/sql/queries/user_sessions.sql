-- name: InsertUserSession :execrows
INSERT INTO user_sessions(user_id, login_history_id, session_token, ip_address, user_agent, last_activity_at, expires_at, logged_out_at, logout_reason)
    VALUES (sqlc.arg('UserID'), sqlc.arg('LoginHistoryID'), sqlc.arg('SessionToken'), sqlc.arg('IPAddress'), sqlc.arg('UserAgent'), sqlc.narg('LastActivityAt'), sqlc.arg('ExpiresAt'), sqlc.narg('LoggedOutAt'), sqlc.narg('LogoutReason'));

-- name: UpdateUserSession :execrows
UPDATE
    user_sessions
SET
    last_activity_at = COALESCE(sqlc.narg('LastActivityAt'), last_activity_at),
    ip_address = COALESCE(sqlc.narg('IPAddress'), ip_address),
    user_agent = COALESCE(sqlc.narg('UserAgent'), user_agent),
    expires_at = COALESCE(sqlc.narg('ExpiresAt'), expires_at),
    logged_out_at = COALESCE(sqlc.narg('LoggedOutAt'), logged_out_at),
    logout_reason = COALESCE(sqlc.narg('LogoutReason'), logout_reason)
WHERE
    session_token = sqlc.arg('SessionToken');

-- name: GetUserSessionByToken :one
SELECT
    us.id,
    us.user_id,
    us.login_history_id,
    us.session_token,
    us.ip_address,
    us.user_agent,
    us.last_activity_at,
    us.expires_at,
    us.logged_out_at,
    us.logout_reason
FROM
    user_sessions us
WHERE
    us.session_token = sqlc.arg('SessionToken');

-- name: GetActiveUserSessionsByUserID :many
SELECT
    us.id,
    us.user_id,
    us.login_history_id,
    us.session_token,
    us.ip_address,
    us.user_agent,
    us.last_activity_at,
    us.expires_at,
    us.logged_out_at,
    us.logout_reason
FROM
    user_sessions us
WHERE
    us.user_id = sqlc.arg('UserID')
    AND us.logged_out_at IS NULL
    AND us.expires_at > NOW();

