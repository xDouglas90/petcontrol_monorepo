-- name: InsertCompanyFinance :execrows
INSERT INTO company_finances(company_id, finance_id, is_primary)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('FinanceID'), sqlc.narg('IsPrimary'));

-- name: GetCompanyFinanceByCompanyID :one
SELECT
    cf.id,
    cf.company_id,
    cf.finance_id,
    cf.is_primary,
    cf.created_at
FROM
    company_finances cf
WHERE
    cf.company_id = sqlc.arg('CompanyID')
    AND cf.finance_id = sqlc.arg('FinanceID')
LIMIT 1;

-- name: UpdateCompanyFinance :execrows
UPDATE
    company_finances
SET
    is_primary = coalesce(sqlc.narg('IsPrimary'), is_primary),
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND finance_id = sqlc.arg('FinanceID');

-- name: DeleteCompanyFinance :execrows
DELETE FROM company_finances
WHERE company_id = sqlc.arg('CompanyID')
    AND finance_id = sqlc.arg('FinanceID');

-- name: ListCompanyFinances :many
SELECT
    cf.id,
    cf.company_id,
    cf.finance_id,
    cf.is_primary,
    cf.created_at,
    f.id AS finance_id,
    f.bank_name,
    f.bank_code,
    f.bank_branch,
    f.bank_account,
    f.bank_account_digit,
    f.bank_account_type,
    f.has_pix,
    f.pix_key,
    f.pix_key_type,
    f.created_at AS finance_created_at,
    f.updated_at AS finance_updated_at
FROM
    company_finances cf
    JOIN finances f ON cf.finance_id = f.id
WHERE
    cf.company_id = sqlc.arg('CompanyID')
ORDER BY
    cf.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

