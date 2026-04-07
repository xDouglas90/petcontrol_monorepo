-- name: GetUserAuthByUserID :one
SELECT
    ua.user_id,
    ua.password_hash,
    ua.password_changed_at,
    ua.must_change_password,
    ua.login_attempts,
    ua.locked_until,
    ua.last_login_at,
    ua.created_at,
    ua.updated_at
FROM
    user_auth ua
WHERE
    ua.user_id = sqlc.arg('UserID')
LIMIT 1;

-- name: IncrementUserAuthLoginAttempts :exec
UPDATE
    user_auth
SET
    login_attempts = login_attempts + 1,
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

-- name: ResetUserAuthLoginAttempts :exec
UPDATE
    user_auth
SET
    login_attempts = 0,
    locked_until = NULL,
    last_login_at = now(),
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

-- name: SetUserAuthLockedUntil :exec
UPDATE
    user_auth
SET
    locked_until = sqlc.arg('LockedUntil'),
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

-- name: InsertLoginHistory :exec
INSERT INTO login_history(user_id, ip_address, user_agent, result, failure_detail)
    VALUES (sqlc.narg('UserID'), sqlc.arg('IPAddress'), sqlc.arg('UserAgent'), sqlc.arg('Result'), sqlc.narg('FailureDetail'));

-- name: HasActiveCompanyModuleByCode :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            company_modules cm
            INNER JOIN modules m ON m.id = cm.module_id
        WHERE
            cm.company_id = sqlc.arg('CompanyID')
            AND cm.is_active = TRUE
            AND (cm.expires_at IS NULL
                OR cm.expires_at > now())
            AND m.code = sqlc.arg('Code')
            AND m.is_active = TRUE
            AND m.deleted_at IS NULL) AS has_access;

