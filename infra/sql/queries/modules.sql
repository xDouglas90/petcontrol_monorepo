-- name: CreateModule :one
INSERT INTO modules(code, "name", description, min_package, is_active)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetModuleByCode :one
SELECT
    m.id,
    m.code,
    m."name",
    m.description,
    m.min_package,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.deleted_at
FROM
    modules m
WHERE
    m.code = sqlc.arg('Code')
    AND m.deleted_at IS NULL
LIMIT 1;

-- name: ListModules :many
SELECT
    m.id,
    m.code,
    m."name",
    m.description,
    m.min_package,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.deleted_at
FROM
    modules m
WHERE
    m.deleted_at IS NULL
ORDER BY
    m.code ASC;

-- name: ListActiveModulesByCompanyID :many
SELECT
    m.id,
    m.code,
    m."name",
    m.description,
    m.min_package,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.deleted_at
FROM
    company_modules cm
    INNER JOIN modules m ON m.id = cm.module_id
WHERE
    cm.company_id = sqlc.arg('CompanyID')
    AND cm.is_active = TRUE
    AND m.is_active = TRUE
    AND m.deleted_at IS NULL
ORDER BY
    m.code ASC;

-- name: UpdateModule :one
UPDATE
    modules
SET
    code = coalesce(sqlc.narg('Code'), code),
    "name" = coalesce(sqlc.narg('Name'), "name"),
    description = coalesce(sqlc.narg('Description'), description),
    min_package = coalesce(sqlc.narg('MinPackage'), min_package),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
RETURNING
    *;

-- name: DeleteModule :execrows
UPDATE
    modules
SET
    deleted_at = now(),
    is_active = FALSE
WHERE
    id = sqlc.arg('ID');

