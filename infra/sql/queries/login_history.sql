-- name: GetLoginHistoryByID :one
SELECT
    lh.id,
    lh.user_id,
    lh.ip_address,
    lh.user_agent,
    lh.result,
    lh.failure_detail,
    lh.attempted_at
FROM
    login_history lh
WHERE
    lh.id = sqlc.arg('ID')
LIMIT 1;

-- name: ListLoginHistoryByUserID :many
SELECT
    lh.id,
    lh.user_id,
    lh.ip_address,
    lh.user_agent,
    lh.result,
    lh.failure_detail,
    lh.attempted_at
FROM
    login_history lh
WHERE
    lh.user_id = sqlc.arg('UserID')
ORDER BY
    lh.attempted_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListLoginHistoryByResult :many
SELECT
    lh.id,
    lh.user_id,
    lh.ip_address,
    lh.user_agent,
    lh.result,
    lh.failure_detail,
    lh.attempted_at
FROM
    login_history lh
WHERE
    lh.result = sqlc.arg('Result')
ORDER BY
    lh.attempted_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

