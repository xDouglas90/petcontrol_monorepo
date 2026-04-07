-- name: InsertUser :one
INSERT INTO users(email, email_verified, email_verified_at, "role", kind, is_active)
  VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
  *;

-- name: GetUserByID :one
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
  u.kind,
  u.is_active,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM
  users u
WHERE
  u.id = sqlc.arg('ID')
LIMIT 1;

-- name: GetUserByEmail :one
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
  u.kind,
  u.is_active,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM
  users u
WHERE
  u.email = sqlc.arg('Email')
LIMIT 1;

-- name: UpdateUser :execrows
UPDATE
  users
SET
  email = coalesce(sqlc.narg('Email'), email),
  email_verified = coalesce(sqlc.narg('EmailVerified'), email_verified),
  email_verified_at = coalesce(sqlc.narg('EmailVerifiedAt'), email_verified_at),
  "role" = coalesce(sqlc.narg('Role'), "role"),
  kind = coalesce(sqlc.narg('Kind'), kind),
  is_active = coalesce(sqlc.narg('IsActive'), is_active),
  updated_at = now()
WHERE
  id = sqlc.arg('ID');

-- name: DeleteUser :execrows
UPDATE
  users
SET
  deleted_at = now(),
  is_active = FALSE
WHERE
  id = sqlc.arg('ID');

-- name: ListUsers :many
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
  u.kind,
  u.is_active,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM
  users u
WHERE
  u.deleted_at IS NULL
  AND (u.email ILIKE '%' || sqlc.arg('Email') || '%'
    OR sqlc.arg('Email') IS NULL)
  AND (u."role" = sqlc.arg('Role')
    OR sqlc.arg('Role') IS NULL)
  AND (u.kind = sqlc.arg('Kind')
    OR sqlc.arg('Kind') IS NULL)
  AND (u.is_active = sqlc.arg('IsActive')
    OR sqlc.arg('IsActive') IS NULL)
  AND (u.created_at >= sqlc.arg('CreatedAfter')
    OR sqlc.arg('CreatedAfter') IS NULL)
  AND (u.email_verified = sqlc.arg('EmailVerified')
    OR sqlc.arg('EmailVerified') IS NULL)
  AND (u.email_verified_at >= sqlc.arg('EmailVerifiedAfter')
    OR sqlc.arg('EmailVerifiedAfter') IS NULL)
ORDER BY
  u.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListUsersBasic :many
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
  u.kind,
  u.is_active,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM
  users u
WHERE
  u.deleted_at IS NULL
ORDER BY
  u.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

