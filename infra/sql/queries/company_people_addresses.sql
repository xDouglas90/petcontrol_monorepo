-- name: InsertCompanyPeopleAddresses :execrows
INSERT INTO company_people_addresses(company_id, person_id, address_id, is_main)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('PersonID'), sqlc.arg('AddressID'), sqlc.narg('IsMain'));

-- name: GetCompanyPeopleAddresses :one
SELECT
    cpa.id,
    cpa.company_id,
    cpa.person_id,
    cpa.address_id,
    cpa.is_main,
    cpa.created_at
FROM
    company_people_addresses cpa
WHERE
    cpa.company_id = sqlc.arg('CompanyID')
    AND cpa.person_id = sqlc.arg('PersonID')
LIMIT 1;

-- name: UpdateCompanyPeopleAddresses :execrows
UPDATE
    company_people_addresses
SET
    address_id = coalesce(sqlc.narg('AddressID'), address_id),
    is_main = coalesce(sqlc.narg('IsMain'), is_main),
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND person_id = sqlc.arg('PersonID');

-- name: DeleteCompanyPeopleAddresses :execrows
DELETE FROM company_people_addresses
WHERE company_id = sqlc.arg('CompanyID')
    AND person_id = sqlc.arg('PersonID');

-- name: ListCompanyPeopleAddresses :many
SELECT
    cpa.id,
    cpa.company_id,
    cpa.person_id,
    cpa.address_id,
    cpa.is_main,
    cpa.created_at,
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
    company_people_addresses cpa
    JOIN addresses a ON cpa.address_id = a.id
WHERE
    cpa.company_id = sqlc.arg('CompanyID')
ORDER BY
    cpa.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListCompanyPersonAddresses :many
SELECT
    cpa.id,
    cpa.company_id,
    cpa.person_id,
    cpa.address_id,
    cpa.is_main,
    cpa.created_at,
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
    company_people_addresses cpa
    JOIN addresses a ON cpa.address_id = a.id
WHERE
    cpa.company_id = sqlc.arg('CompanyID')
    AND cpa.person_id = sqlc.arg('PersonID')
ORDER BY
    cpa.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

