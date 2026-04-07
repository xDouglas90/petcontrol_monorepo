-- name: CreateCompanyUser :one
INSERT INTO company_users(company_id, user_id, is_owner, is_active)
    VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetCompanyUserByID :one
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.is_owner,
    cu.is_active,
    cu.joined_at,
    cu.left_at
FROM
    company_users cu
WHERE
    cu.id = sqlc.arg('ID')
LIMIT 1;

-- name: GetCompanyUser :one
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.is_owner,
    cu.is_active,
    cu.joined_at,
    cu.left_at
FROM
    company_users cu
WHERE
    cu.company_id = sqlc.arg('CompanyID')
    AND cu.user_id = sqlc.arg('UserID')
LIMIT 1;

-- name: ListCompanyUsersByCompanyID :many
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.is_owner,
    cu.is_active,
    cu.joined_at,
    cu.left_at
FROM
    company_users cu
WHERE
    cu.company_id = sqlc.arg('CompanyID')
    AND cu.is_active = TRUE
ORDER BY
    cu.joined_at DESC;

-- name: DeactivateCompanyUser :exec
UPDATE
    company_users
SET
    is_active = FALSE,
    left_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND user_id = sqlc.arg('UserID');

