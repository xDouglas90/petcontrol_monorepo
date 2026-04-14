-- name: InsertPersonFinance :execrows
INSERT INTO people_finances(person_id, finance_id, is_primary)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('FinanceID'), sqlc.narg('IsPrimary'));

-- name: UpdatePersonFinance :execrows
UPDATE
    people_finances
SET
    is_primary = COALESCE(sqlc.narg('IsPrimary'), is_primary),
    updated_at = now()
WHERE
    person_id = sqlc.arg('ID')
    AND finance_id = sqlc.arg('FinanceID');

-- name: GetPersonFinance :one
SELECT
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
    f.updated_at AS finance_updated_at,
    pf.id AS person_finance_id,
    pf.person_id,
    pf.is_primary,
    pf.created_at AS person_finance_created_at,
    pf.updated_at AS person_finance_updated_at
FROM
    people_finances pf
    JOIN finances f ON pf.finance_id = f.id
WHERE
    pf.person_id = sqlc.arg('PersonID')
    AND pf.finance_id = sqlc.arg('FinanceID');

-- name: ListPersonFinances :many
SELECT
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
    f.updated_at AS finance_updated_at,
    pf.id AS person_finance_id,
    pf.person_id,
    pf.is_primary,
    pf.created_at AS person_finance_created_at,
    pf.updated_at AS person_finance_updated_at
FROM
    people_finances pf
    JOIN finances f ON pf.finance_id = f.id
WHERE
    pf.person_id = sqlc.arg('PersonID')
ORDER BY
    pf.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

