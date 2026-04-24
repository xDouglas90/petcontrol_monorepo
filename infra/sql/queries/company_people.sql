-- name: InsertCompanyPerson :execrows
INSERT INTO company_people(company_id, person_id)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('PersonID'));

-- name: GetCompanyPerson :one
SELECT
    cp.id AS company_person_id,
    cp.company_id,
    cp.person_id,
    cp.created_at AS company_person_created_at,
    p.id AS person_id,
    p.kind AS person_kind,
    p.is_active AS person_is_active,
    p.has_system_user AS person_has_system_user,
    p.created_at AS person_created_at,
    p.updated_at AS person_updated_at,
    pi.full_name AS identifications_full_name,
    PI.short_name AS identifications_short_name,
    pc.email AS contacts_email,
    pi.gender_identity AS identifications_gender_identity,
    pi.marital_status AS identifications_marital_status,
    pi.image_url AS identifications_image_url,
    pi.birth_date AS identifications_birth_date,
    pi.cpf AS identifications_cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at
FROM
    company_people cp
    JOIN people p ON cp.person_id = p.id
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
    LEFT JOIN people_contacts pc ON p.id = pc.person_id AND pc.is_primary = TRUE
WHERE
    cp.company_id = sqlc.arg('CompanyID')
    AND cp.person_id = sqlc.arg('PersonID');

-- name: ListCompanyPeople :many
SELECT
    cp.id AS company_person_id,
    cp.company_id,
    cp.person_id,
    cp.created_at AS company_person_created_at,
    p.id AS person_id,
    p.kind AS person_kind,
    p.is_active AS person_is_active,
    p.has_system_user AS person_has_system_user,
    p.created_at AS person_created_at,
    p.updated_at AS person_updated_at,
    pi.full_name AS identifications_full_name,
    PI.short_name AS identifications_short_name,
    pc.email AS contacts_email,
    pi.gender_identity AS identifications_gender_identity,
    pi.marital_status AS identifications_marital_status,
    pi.image_url AS identifications_image_url,
    pi.birth_date AS identifications_birth_date,
    pi.cpf AS identifications_cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at
FROM
    company_people cp
    JOIN people p ON cp.person_id = p.id
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
    LEFT JOIN people_contacts pc ON p.id = pc.person_id AND pc.is_primary = TRUE
WHERE
    cp.company_id = sqlc.arg('CompanyID')
ORDER BY
    cp.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeleteCompanyPerson :exec
DELETE FROM company_people
WHERE company_id = sqlc.arg('CompanyID')
    AND person_id = sqlc.arg('PersonID');
