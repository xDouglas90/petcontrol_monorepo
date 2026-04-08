-- name: InsertCompany :one
INSERT INTO companies(slug, "name", fantasy_name, cnpj, foundation_date, logo_url, responsible_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetCompanyByID :one
SELECT
    c.id,
    c.slug,
    c."name",
    c.fantasy_name,
    c.cnpj,
    c.foundation_date,
    c.logo_url,
    c.responsible_id,
    c.active_package,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM
    companies c
WHERE
    c.id = sqlc.arg('ID')
    AND c.deleted_at IS NULL
LIMIT 1;

-- name: GetCompanyBySlug :one
SELECT
    c.id,
    c.slug,
    c."name",
    c.fantasy_name,
    c.cnpj,
    c.foundation_date,
    c.logo_url,
    c.responsible_id,
    c.active_package,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM
    companies c
WHERE
    c.slug = sqlc.arg('Slug')
    AND c.deleted_at IS NULL
LIMIT 1;

-- name: ListCompanies :many
SELECT
    c.id,
    c.slug,
    c."name",
    c.fantasy_name,
    c.cnpj,
    c.foundation_date,
    c.logo_url,
    c.responsible_id,
    c.active_package,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM
    companies c
WHERE
    c.deleted_at IS NULL
ORDER BY
    c.created_at DESC;

-- name: ListCompaniesByPackage :many
SELECT
    c.id,
    c.slug,
    c."name",
    c.fantasy_name,
    c.cnpj,
    c.foundation_date,
    c.logo_url,
    c.responsible_id,
    c.active_package,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM
    companies c
WHERE
    c.deleted_at IS NULL
    AND c.active_package = sqlc.arg('ActivePackage')
ORDER BY
    c.created_at DESC;

-- name: UpdateCompany :execrows
UPDATE
    companies
SET
    slug = coalesce(sqlc.narg('Slug'), slug),
    "name" = coalesce(sqlc.narg('Name'), "name"),
    fantasy_name = coalesce(sqlc.narg('FantasyName'), fantasy_name),
    cnpj = coalesce(sqlc.narg('CNPJ'), cnpj),
    foundation_date = coalesce(sqlc.narg('FoundationDate'), foundation_date),
    logo_url = coalesce(sqlc.narg('LogoURL'), logo_url),
    responsible_id = coalesce(sqlc.narg('ResponsibleID'), responsible_id),
    active_package = coalesce(sqlc.narg('ActivePackage'), active_package),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: DeleteCompany :execrows
UPDATE
    companies
SET
    deleted_at = now(),
    is_active = FALSE
WHERE
    id = sqlc.arg('ID');

