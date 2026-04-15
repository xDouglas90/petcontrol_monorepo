-- name: InsertCompanyClient :execrows
INSERT INTO company_clients(company_id, client_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ClientID'), sqlc.narg('IsActive'));

-- name: GetCompanyClient :one
SELECT
    cc.id,
    cc.company_id,
    cc.client_id,
    cc.is_active,
    cc.joined_at,
    cc.left_at
FROM
    company_clients cc
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND cc.client_id = sqlc.arg('ClientID')
LIMIT 1;

-- name: UpdateCompanyClient :execrows
UPDATE
    company_clients
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    left_at = coalesce(sqlc.narg('LeftAt'), left_at)
WHERE
    company_id = sqlc.arg('CompanyID')
    AND client_id = sqlc.arg('ClientID');

-- name: DeactivateCompanyClient :execrows
UPDATE
    company_clients
SET
    is_active = FALSE,
    left_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND client_id = sqlc.arg('ClientID')
    AND is_active = TRUE;

-- name: DeleteCompanyClient :execrows
DELETE FROM company_clients
WHERE company_id = sqlc.arg('CompanyID')
    AND client_id = sqlc.arg('ClientID');

-- name: ListCompanyClients :many
SELECT
    cc.id,
    cc.company_id,
    cc.client_id,
    cc.is_active,
    cc.joined_at,
    cc.left_at,
    pi.full_name AS client_name,
    pi.short_name AS client_short_name,
    pi.cpf AS client_cpf,
    pi.image_url AS client_image_url
FROM
    company_clients cc
    JOIN clients c ON cc.client_id = c.id
    JOIN people_identifications pi ON c.person_id = pi.person_id
WHERE
    cc.company_id = sqlc.arg('CompanyID')
ORDER BY
    cc.joined_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListActiveCompanyClients :many
SELECT
    cc.id,
    cc.company_id,
    cc.client_id,
    cc.is_active,
    cc.joined_at,
    cc.left_at,
    pi.full_name AS client_name,
    pi.short_name AS client_short_name,
    pi.cpf AS client_cpf,
    pi.image_url AS client_image_url
FROM
    company_clients cc
    JOIN clients c ON cc.client_id = c.id
    JOIN people_identifications pi ON c.person_id = pi.person_id
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
ORDER BY
    cc.joined_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

