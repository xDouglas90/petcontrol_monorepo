-- name: InsertUser :one
INSERT INTO users(email, email_verified, email_verified_at, "role", is_active)
  VALUES ($1, $2, $3, $4, $5)
RETURNING
  *;

-- name: GetUserByID :one
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
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

-- name: RestoreUser :execrows
UPDATE
  users
SET
  deleted_at = NULL,
  is_active = TRUE
WHERE
  id = sqlc.arg('ID');

-- name: ListUsers :many
SELECT
  u.id,
  u.email,
  u.email_verified,
  u.email_verified_at,
  u."role",
  u.is_active,
  u.created_at,
  u.updated_at,
  u.deleted_at
FROM
  users u
WHERE
  u.deleted_at IS NULL
  AND (u.email ILIKE '%' || sqlc.narg('Email') || '%'
    OR sqlc.narg('Email') IS NULL)
  AND (u."role" = sqlc.narg('Role')
    OR sqlc.narg('Role') IS NULL)
  AND (u.is_active = sqlc.narg('IsActive')
    OR sqlc.narg('IsActive') IS NULL)
  AND (u.created_at >= sqlc.narg('CreatedAfter')
    OR sqlc.narg('CreatedAfter') IS NULL)
  AND (u.email_verified = sqlc.narg('EmailVerified')
    OR sqlc.narg('EmailVerified') IS NULL)
  AND (u.email_verified_at >= sqlc.narg('EmailVerifiedAfter')
    OR sqlc.narg('EmailVerifiedAfter') IS NULL)
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

