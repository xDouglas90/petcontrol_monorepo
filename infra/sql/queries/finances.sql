-- name: InsertFinance :execrows
INSERT INTO finances(bank_name, bank_code, bank_branch, bank_account, bank_account_digit, bank_account_type, has_pix, pix_key, pix_key_type)
    VALUES (sqlc.arg('BankName'), sqlc.arg('BankCode'), sqlc.arg('BankBranch'), sqlc.arg('BankAccount'), sqlc.arg('BankAccountDigit'), sqlc.arg('BankAccountType'), sqlc.narg('HasPix'), sqlc.arg('PixKey'), sqlc.arg('PixKeyType'));

-- name: UpdateFinance :execrows
UPDATE
    finances
SET
    bank_name = COALESCE(sqlc.arg('BankName'), bank_name),
    bank_code = COALESCE(sqlc.arg('BankCode'), bank_code),
    bank_branch = COALESCE(sqlc.arg('BankBranch'), bank_branch),
    bank_account = COALESCE(sqlc.arg('BankAccount'), bank_account),
    bank_account_digit = COALESCE(sqlc.arg('BankAccountDigit'), bank_account_digit),
    bank_account_type = COALESCE(sqlc.arg('BankAccountType'), bank_account_type),
    has_pix = COALESCE(sqlc.narg('HasPix'), has_pix),
    pix_key = COALESCE(sqlc.arg('PixKey'), pix_key),
    pix_key_type = COALESCE(sqlc.arg('PixKeyType'), pix_key_type),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetFinance :one
SELECT
    f.id,
    f.bank_name,
    f.bank_code,
    f.bank_branch,
    f.bank_account,
    f.bank_account_digit,
    f.bank_account_type,
    f.has_pix,
    f.pix_key,
    f.pix_key_type,
    f.created_at,
    f.updated_at
FROM
    finances f
WHERE
    f.id = sqlc.arg('ID');

-- name: DeleteFinance :execrows
DELETE FROM finances
WHERE id = sqlc.arg('ID');

