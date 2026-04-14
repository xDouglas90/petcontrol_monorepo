-- name: InsertUserProfile :execrows
INSERT INTO user_profiles(user_id, person_id)
    VALUES (sqlc.arg('UserID'), sqlc.arg('PersonID'));

-- name: GetUserProfile :one
SELECT
    up.user_id,
    up.person_id,
    up.created_at
FROM
    user_profiles up
WHERE
    up.user_id = sqlc.arg('UserID');

-- name: UpdateUserProfile :execrows
UPDATE
    user_profiles
SET
    person_id = COALESCE(sqlc.arg('PersonID'), person_id),
    updated_at = now()
WHERE
    user_id = sqlc.arg('UserID');

