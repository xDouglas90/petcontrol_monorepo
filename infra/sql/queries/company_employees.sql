-- name: InsertCompanyEmployee :one
INSERT INTO company_employees(company_id, person_id)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('PersonID'))
RETURNING
    id, company_id, person_id, created_at, deleted_at;

-- name: GetCompanyEmployee :one
SELECT
    ce.id AS company_employee_id,
    ce.company_id,
    ce.person_id,
    ce.created_at AS company_employee_created_at,
    ce.deleted_at AS company_employee_deleted_at,
    p.id AS person_id,
    p.kind AS person_kind,
    p.is_active AS person_is_active,
    p.has_system_user AS person_has_system_user,
    p.created_at AS person_created_at,
    p.updated_at AS person_updated_at,
    pi.full_name AS identifications_full_name,
    pi.short_name AS identifications_short_name,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at,
    e.id AS employment_id,
    e.company_employee_id AS employee_id,
    e."role",
    e.admission_date,
    e.resignation_date,
    e.salary,
    e.created_at AS employment_created_at,
    e.updated_at AS employment_updated_at
FROM
    company_employees ce
    JOIN people p ON ce.person_id = p.id
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
    LEFT JOIN employments e ON ce.id = e.company_employee_id
WHERE
    ce.company_id = sqlc.arg('CompanyID')
    AND ce.person_id = sqlc.arg('PersonID')
LIMIT 1;

-- name: ListCompanyEmployees :many
SELECT
    ce.id AS company_employee_id,
    ce.company_id,
    ce.person_id,
    ce.created_at AS company_employee_created_at,
    ce.deleted_at AS company_employee_deleted_at,
    p.id AS person_id,
    p.kind AS person_kind,
    p.is_active AS person_is_active,
    p.has_system_user AS person_has_system_user,
    p.created_at AS person_created_at,
    p.updated_at AS person_updated_at,
    pi.full_name AS identifications_full_name,
    pi.short_name AS identifications_short_name,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at,
    e.id AS employment_id,
    e.company_employee_id AS employee_id,
    e."role",
    e.admission_date,
    e.resignation_date,
    e.salary,
    e.created_at AS employment_created_at,
    e.updated_at AS employment_updated_at
FROM
    company_employees ce
    JOIN people p ON ce.person_id = p.id
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
    LEFT JOIN employments e ON ce.id = e.company_employee_id
WHERE
    ce.company_id = sqlc.arg('CompanyID')
ORDER BY
    ce.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeleteCompanyEmployee :exec
UPDATE
    company_employees
SET
    deleted_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND person_id = sqlc.arg('PersonID')
    AND deleted_at IS NULL;

