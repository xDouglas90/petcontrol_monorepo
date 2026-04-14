-- name: InsertEployment :execrows
INSERT INTO employments(company_employee_id, "role", admission_date, resignation_date, salary)
    VALUES (sqlc.arg('CompanyEmployeeID'), sqlc.arg('Role'), sqlc.arg('AdmissionDate'), sqlc.arg('ResignationDate'), sqlc.arg('Salary'));

-- name: UpdateEmployment :execrows
UPDATE
    employments
SET
    "role" = COALESCE(sqlc.narg('Role'), "role"),
    admission_date = COALESCE(sqlc.narg('AdmissionDate'), admission_date),
    resignation_date = COALESCE(sqlc.narg('ResignationDate'), resignation_date),
    salary = COALESCE(sqlc.narg('Salary'), salary),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetEmployment :one
SELECT
    p.id AS person_id,
    p.kind AS person_kind,
    p.is_active AS person_is_active,
    p.has_system_user AS person_has_system_user,
    p.created_at AS person_created_at,
    p.updated_at AS person_updated_at,
    pi.full_name AS identifications_full_name,
    PI.short_name AS identifications_short_name,
    pi.gender_identity AS identifications_gender_identity,
    pi.marital_status AS identifications_marital_status,
    pi.image_url AS identifications_image_url,
    pi.birth_date AS identifications_birth_date,
    pi.cpf AS identifications_cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at,
    e.id AS employment_id,
    e.company_employee_id AS employee_id,
    e."role",
    e.admission_date,
    e.resignation_date,
    e.salary,
    e.created_at,
    e.updated_at
FROM
    employments e
    JOIN company_employees ce ON e.company_employee_id = ce.id
    JOIN people p ON ce.person_id = p.id
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
WHERE
    e.id = sqlc.arg('ID');

