-- name: InsertCompanyAddress :execrows
INSERT INTO company_addresses(company_id, address_id, is_main)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('AddressID'), sqlc.narg('IsPrimary'));

-- name: GetCompanyAddress :one
SELECT
    ca.id,
    ca.company_id,
    ca.address_id,
    ca.is_main,
    ca.created_at
FROM
    company_addresses ca
WHERE
    ca.id = sqlc.arg('ID');

-- name: UpdateCompanyAddress :execrows
UPDATE
    company_addresses
SET
    address_id = coalesce(sqlc.narg('AddressID'), address_id),
    is_main = coalesce(sqlc.narg('IsPrimary'), is_main),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: DeleteCompanyAddress :execrows
DELETE FROM company_addresses
WHERE id = sqlc.arg('ID');

-- name: ListCompanyAddresses :many
SELECT
    ca.id,
    ca.company_id,
    ca.address_id,
    ca.is_main,
    ca.created_at,
    a.street,
    a.number,
    a.complement,
    a.district,
    a.city,
    a.state,
    a.zip_code,
    a.created_at AS address_created_at,
    a.updated_at AS address_updated_at
FROM
    company_addresses ca
    JOIN addresses a ON ca.address_id = a.id
WHERE
    ca.company_id = sqlc.arg('CompanyID')
ORDER BY
    ca.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

