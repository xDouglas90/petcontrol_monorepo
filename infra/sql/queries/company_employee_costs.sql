-- name: InsertCompanyEmployeeCosts :execrows
INSERT INTO company_employee_costs(company_employee_id, costs, reference_month, comments)
    VALUES (sqlc.arg('CompanyEmployeeID'), sqlc.arg('Costs'), sqlc.arg('ReferenceMonth'), sqlc.narg('Comments'));

-- name: GetCompanyEmployeeCostsByCompanyEmployeeID :one
SELECT
    cec.id,
    cec.company_employee_id,
    cec.costs,
    cec.reference_month,
    cec.comments,
    cec.created_at,
    cec.updated_at
FROM
    company_employee_costs cec
WHERE
    cec.company_employee_id = sqlc.arg('CompanyEmployeeID')
LIMIT 1;

-- name: UpdateCompanyEmployeeCosts :execrows
UPDATE
    company_employee_costs
SET
    COSTS = coalesce(sqlc.narg('Costs'), costs),
    reference_month = coalesce(sqlc.narg('ReferenceMonth'), reference_month),
    comments = coalesce(sqlc.narg('Comments'), comments),
    updated_at = now()
WHERE
    company_employee_id = sqlc.arg('CompanyEmployeeID');

-- name: DeleteCompanyEmployeeCosts :execrows
DELETE FROM company_employee_costs
WHERE company_employee_id = sqlc.arg('CompanyEmployeeID');

