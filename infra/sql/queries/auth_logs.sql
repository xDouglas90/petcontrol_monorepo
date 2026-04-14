-- name: InsertAuthLog :exec
INSERT INTO auth_logs(user_id, action, company_id, session_id, ip_address, user_agent, result, detail)
    VALUES (sqlc.narg('UserID'), sqlc.arg('Action'), sqlc.narg('CompanyID'), sqlc.narg('SessionID'), sqlc.arg('IPAddress'), sqlc.arg('UserAgent'), sqlc.narg('Result'), sqlc.narg('Detail'));

-- name: GetAuthLogByID :one
SELECT
    al.id,
    al.user_id,
    al.action,
    al.company_id,
    al.session_id,
    al.ip_address,
    al.user_agent,
    al.result,
    al.detail,
    al.occurred_at
FROM
    auth_logs al
WHERE
    al.id = sqlc.arg('ID')
LIMIT 1;

-- name: ListAuthLogsByUserID :many
SELECT
    al.id,
    al.user_id,
    al.action,
    al.company_id,
    al.session_id,
    al.ip_address,
    al.user_agent,
    al.result,
    al.detail,
    al.occurred_at
FROM
    auth_logs al
WHERE
    al.user_id = sqlc.arg('UserID')
ORDER BY
    al.occurred_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListAuthLogsByCompanyID :many
SELECT
    al.id,
    al.user_id,
    al.action,
    al.company_id,
    al.session_id,
    al.ip_address,
    al.user_agent,
    al.result,
    al.detail,
    al.occurred_at
FROM
    auth_logs al
WHERE
    al.company_id = sqlc.arg('CompanyID')
ORDER BY
    al.occurred_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

