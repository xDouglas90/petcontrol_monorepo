-- name: InsertEmployeeBenefits :execrows
INSERT INTO employee_benefits(company_employee_id, meal_ticket, meal_ticket_value, transport_voucher, transport_voucher_qty, transport_voucher_value, valid_from, valid_until)
    VALUES (sqlc.arg('CompanyEmployeeID'), sqlc.narg('MealTicket'), sqlc.narg('MealTicketValue'), sqlc.narg('TransportVoucher'), sqlc.narg('TransportVoucherQty'), sqlc.narg('TransportVoucherValue'), sqlc.narg('ValidFrom'), sqlc.narg('ValidUntil'));

-- name: GetEmployeeBenefitsByCompanyEmployeeID :one
SELECT
    eb.id,
    eb.company_employee_id,
    eb.meal_ticket,
    eb.meal_ticket_value,
    eb.transport_voucher,
    eb.transport_voucher_qty,
    eb.transport_voucher_value,
    eb.valid_from,
    eb.valid_until,
    eb.created_at,
    eb.updated_at
FROM
    employee_benefits eb
WHERE
    eb.company_employee_id = sqlc.arg('CompanyEmployeeID')
LIMIT 1;

-- name: UpdateEmployeeBenefits :execrows
UPDATE
    employee_benefits
SET
    meal_ticket = coalesce(sqlc.narg('MealTicket'), meal_ticket),
    meal_ticket_value = coalesce(sqlc.narg('MealTicketValue'), meal_ticket_value),
    transport_voucher = coalesce(sqlc.narg('TransportVoucher'), transport_voucher),
    transport_voucher_qty = coalesce(sqlc.narg('TransportVoucherQty'), transport_voucher_qty),
    transport_voucher_value = coalesce(sqlc.narg('TransportVoucherValue'), transport_voucher_value),
    valid_from = coalesce(sqlc.narg('ValidFrom'), valid_from),
    valid_until = coalesce(sqlc.narg('ValidUntil'), valid_until),
    updated_at = now()
WHERE
    company_employee_id = sqlc.arg('CompanyEmployeeID');

-- name: DeleteEmployeeBenefits :execrows
DELETE FROM employee_benefits
WHERE company_employee_id = sqlc.arg('CompanyEmployeeID');

-- name: GetFullEmployeeBenefitsByCompanyEmployeeID :one
SELECT
    pi.full_name AS employee_full_name,
    pi.short_name AS employee_short_name,
    pi.gender_identity AS employee_gender_identity,
    pi.marital_status AS employee_marital_status,
    pi.image_url AS employee_image_url,
    pi.birth_date AS employee_birth_date,
    pi.cpf AS employee_cpf,
    pi.created_at AS employee_identifications_created_at,
    pi.updated_at AS employee_identifications_updated_at,
    eb.id,
    eb.company_employee_id,
    eb.meal_ticket,
    eb.meal_ticket_value,
    eb.transport_voucher,
    eb.transport_voucher_qty,
    eb.transport_voucher_value,
    eb.valid_from,
    eb.valid_until,
    eb.created_at AS benefits_created_at,
    eb.updated_at AS benefits_updated_at,
    ce.created_at AS employee_created_at,
    ce.deleted_at AS employee_deleted_at
FROM
    employee_benefits eb
    JOIN company_employees ce ON eb.company_employee_id = ce.id
    JOIN people_identifications pi ON ce.person_id = pi.person_id
WHERE
    eb.company_employee_id = sqlc.arg('CompanyEmployeeID')
LIMIT 1;

-- name: GetCompanyEmployeesBenefits :many
SELECT
    eb.id,
    eb.company_employee_id,
    eb.meal_ticket,
    eb.meal_ticket_value,
    eb.transport_voucher,
    eb.transport_voucher_qty,
    eb.transport_voucher_value,
    eb.valid_from,
    eb.valid_until,
    eb.created_at,
    eb.updated_at
FROM
    employee_benefits eb
    JOIN company_employees ce ON eb.company_employee_id = ce.id
WHERE
    ce.company_id = sqlc.arg('CompanyID')
ORDER BY
    ce.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: GetCompanyFullEmployeeBenefits :many
SELECT
    pi.full_name AS employee_full_name,
    pi.short_name AS employee_short_name,
    pi.gender_identity AS employee_gender_identity,
    pi.marital_status AS employee_marital_status,
    pi.image_url AS employee_image_url,
    pi.birth_date AS employee_birth_date,
    pi.cpf AS employee_cpf,
    pi.created_at AS employee_identifications_created_at,
    pi.updated_at AS employee_identifications_updated_at,
    eb.id,
    eb.company_employee_id,
    eb.meal_ticket,
    eb.meal_ticket_value,
    eb.transport_voucher,
    eb.transport_voucher_qty,
    eb.transport_voucher_value,
    eb.valid_from,
    eb.valid_until,
    eb.created_at AS benefits_created_at,
    eb.updated_at AS benefits_updated_at,
    ce.created_at AS employee_created_at,
    ce.deleted_at AS employee_deleted_at
FROM
    employee_benefits eb
    JOIN company_employees ce ON eb.company_employee_id = ce.id
    JOIN people_identifications pi ON ce.person_id = pi.person_id
WHERE
    ce.company_id = sqlc.arg('CompanyID')
ORDER BY
    ce.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

