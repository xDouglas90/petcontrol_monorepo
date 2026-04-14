-- name: InsertCompanyService :one
INSERT INTO company_services(company_id, service_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ServiceID'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetCompanyService :one
SELECT
    cs.id,
    cs.company_id,
    cs.service_id,
    cs.is_active,
    cs.created_at,
    cs.updated_at
FROM
    company_services cs
WHERE
    cs.company_id = sqlc.arg('CompanyID')
    AND cs.service_id = sqlc.arg('ServiceID')
LIMIT 1;

-- name: UpdateCompanyService :one
UPDATE
    company_services
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND service_id = sqlc.arg('ServiceID')
RETURNING
    *;

-- name: DeleteCompanyService :one
DELETE FROM company_services
WHERE company_id = sqlc.arg('CompanyID')
    AND service_id = sqlc.arg('ServiceID')
RETURNING
    *;

-- name: ListCompanyServices :many
SELECT
    cs.id,
    cs.company_id,
    cs.service_id,
    cs.is_active,
    cs.created_at,
    cs.updated_at,
    s.title AS service_title,
    s.description AS service_description,
    s.price AS service_price,
    s.discount_rate AS service_discount_rate,
    s.image_url AS service_image_url,
    s.deleted_at AS service_deleted_at
FROM
    company_services cs
    JOIN services s ON cs.service_id = s.id
WHERE
    cs.company_id = sqlc.arg('CompanyID')
ORDER BY
    cs.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListActiveCompanyServices :many
SELECT
    cs.id,
    cs.company_id,
    cs.service_id,
    cs.is_active,
    cs.created_at,
    cs.updated_at,
    s.title AS service_title,
    s.description AS service_description,
    s.price AS service_price,
    s.discount_rate AS service_discount_rate,
    s.image_url AS service_image_url
FROM
    company_services cs
    JOIN services s ON cs.service_id = s.id
WHERE
    cs.company_id = sqlc.arg('CompanyID')
    AND cs.is_active = TRUE
    AND s.is_active = TRUE
    AND s.deleted_at IS NULL
ORDER BY
    s.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

