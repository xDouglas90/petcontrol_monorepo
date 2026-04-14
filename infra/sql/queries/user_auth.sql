-- name: InsertUserAuth :exec
INSERT INTO user_auth(user_id, password_hash, must_change_password)
    VALUES (sqlc.arg('UserID'), sqlc.arg('PasswordHash'), sqlc.narg('MustChangePassword'));

-- name: UpdateUserAuthPassword :exec
UPDATE
    user_auth
SET
    password_hash = sqlc.arg('PasswordHash'),
    password_changed_at = now(),
    must_change_password = FALSE,
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

-- name: UpdateUserAuthMustChangePassword :exec
UPDATE
    user_auth
SET
    must_change_password = sqlc.arg('MustChangePassword'),
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

-- name: DeleteUserAuth :execrows
DELETE FROM user_auth
WHERE user_id = sqlc.arg('UserID');

