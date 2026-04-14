-- name: InsertUserSettings :execrows
INSERT INTO user_settings(user_id, notifications_enabled, theme, "language", timezone)
    VALUES (sqlc.arg('UserID'), sqlc.narg('NotificationsEnabled'), sqlc.narg('Theme'), sqlc.narg('Language'), sqlc.narg('Timezone'));

-- name: UpdateUserSettings :execrows
UPDATE
    user_settings
SET
    notifications_enabled = COALESCE(sqlc.narg('NotificationsEnabled'), notifications_enabled),
    theme = COALESCE(sqlc.narg('Theme'), theme),
    "language" = COALESCE(sqlc.narg('Language'), "language"),
    timezone = COALESCE(sqlc.narg('Timezone'), timezone)
WHERE
    user_id = sqlc.arg('UserID');

-- name: GetUserSettingsByUserID :one
SELECT
    us.user_id,
    us.notifications_enabled,
    us.theme,
    us."language",
    us.timezone
FROM
    user_settings us
WHERE
    us.user_id = sqlc.arg('UserID');

