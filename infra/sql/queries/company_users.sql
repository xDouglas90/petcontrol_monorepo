-- name: CreateCompanyUser :one
INSERT INTO company_users(company_id, user_id, kind, is_owner, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('UserID'), sqlc.arg('Kind'), sqlc.arg('IsOwner'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetCompanyUserByID :one
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.kind,
    cu.is_owner,
    cu.is_active,
    cu.created_at,
    cu.updated_at,
    cu.deleted_at
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
    cu.kind,
    cu.is_owner,
    cu.is_active,
    cu.created_at,
    cu.updated_at,
    cu.deleted_at
FROM
    company_users cu
WHERE
    cu.company_id = sqlc.arg('CompanyID')
    AND cu.user_id = sqlc.arg('UserID')
LIMIT 1;

-- name: GetActiveCompanyUserByUserID :one
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.kind,
    cu.is_owner,
    cu.is_active,
    cu.created_at,
    cu.updated_at,
    cu.deleted_at
FROM
    company_users cu
WHERE
    cu.user_id = sqlc.arg('UserID')
    AND cu.is_active = TRUE
ORDER BY
    cu.kind DESC,
    cu.created_at ASC
LIMIT 1;

-- name: ListCompanyUsersByCompanyID :many
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.kind,
    cu.is_owner,
    cu.is_active,
    cu.created_at,
    cu.updated_at,
    cu.deleted_at
FROM
    company_users cu
WHERE
    cu.company_id = sqlc.arg('CompanyID')
    AND cu.is_active = TRUE
ORDER BY
    cu.created_at DESC;

-- name: ListCompanyUsersByKind :many
SELECT
    cu.id,
    cu.company_id,
    cu.user_id,
    cu.kind,
    cu.is_owner,
    cu.is_active,
    cu.created_at,
    cu.updated_at,
    cu.deleted_at
FROM
    company_users cu
WHERE
    cu.company_id = sqlc.arg('CompanyID')
    AND cu.kind = sqlc.arg('Kind')
    AND cu.is_active = TRUE
ORDER BY
    cu.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeactivateCompanyUser :exec
UPDATE
    company_users
SET
    is_active = FALSE,
    deleted_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND user_id = sqlc.arg('UserID');

