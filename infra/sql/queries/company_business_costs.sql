-- name: InsertCompanyBusinessCost :one
INSERT INTO company_business_costs(company_id, invoice_number, invoice_url, description, total_cost, reference_month, comments)
    VALUES (sqlc.arg('CompanyID'), sqlc.narg('InvoiceNumber'), sqlc.narg('InvoiceURL'), sqlc.arg('Description'), sqlc.arg('TotalCost'), sqlc.arg('ReferenceMonth'), sqlc.narg('Comments'))
RETURNING
    *;

-- name: GetCompanyBusinessCostByID :one
SELECT
    cbc.id,
    cbc.company_id,
    cbc.invoice_number,
    cbc.invoice_url,
    cbc.description,
    cbc.total_cost,
    cbc.reference_month,
    cbc.comments,
    cbc.created_at,
    cbc.updated_at,
    cbc.deleted_at
FROM
    company_business_costs cbc
WHERE
    cbc.id = sqlc.arg('ID')
    AND cbc.company_id = sqlc.arg('CompanyID')
    AND cbc.deleted_at IS NULL
LIMIT 1;

-- name: UpdateCompanyBusinessCost :execrows
UPDATE
    company_business_costs
SET
    invoice_number = coalesce(sqlc.narg('InvoiceNumber'), invoice_number),
    invoice_url = coalesce(sqlc.narg('InvoiceURL'), invoice_url),
    description = coalesce(sqlc.narg('Description'), description),
    total_cost = coalesce(sqlc.narg('TotalCost'), total_cost),
    reference_month = coalesce(sqlc.narg('ReferenceMonth'), reference_month),
    comments = coalesce(sqlc.narg('Comments'), comments),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: DeleteCompanyBusinessCost :execrows
UPDATE
    company_business_costs
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: ListCompanyBusinessCosts :many
SELECT
    cbc.id,
    cbc.company_id,
    cbc.invoice_number,
    cbc.invoice_url,
    cbc.description,
    cbc.total_cost,
    cbc.reference_month,
    cbc.comments,
    cbc.created_at,
    cbc.updated_at,
    cbc.deleted_at
FROM
    company_business_costs cbc
WHERE
    cbc.company_id = sqlc.arg('CompanyID')
    AND cbc.deleted_at IS NULL
ORDER BY
    cbc.reference_month DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListCompanyBusinessCostsByMonth :many
SELECT
    cbc.id,
    cbc.company_id,
    cbc.invoice_number,
    cbc.invoice_url,
    cbc.description,
    cbc.total_cost,
    cbc.reference_month,
    cbc.comments,
    cbc.created_at,
    cbc.updated_at,
    cbc.deleted_at
FROM
    company_business_costs cbc
WHERE
    cbc.company_id = sqlc.arg('CompanyID')
    AND cbc.reference_month = sqlc.arg('ReferenceMonth')
    AND cbc.deleted_at IS NULL
ORDER BY
    cbc.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

