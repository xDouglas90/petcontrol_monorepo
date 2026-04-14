-- name: InsertCompanyModule :execrows
INSERT INTO company_modules(company_id, module_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ModuleID'), sqlc.narg('IsActive'));

-- name: BulkInsertCompanyModules :execrows
INSERT INTO company_modules(company_id, module_id, is_active)
SELECT
    cm.company_id,
    unnest(sqlc.arg('ModuleIDs')::uuid[]),
    sqlc.narg('IsActive');

-- name: DeleteCompanyModule :execrows
DELETE FROM company_modules
WHERE company_id = sqlc.arg('CompanyID')
    AND module_id = sqlc.arg('ModuleID');

-- name: ListModulesByCompanyID :many
SELECT
    m.id,
    m.code,
    m."name",
    m.description,
    m.created_at,
    m.updated_at,
    m.deleted_at
FROM
    company_modules cm
    JOIN modules m ON cm.module_id = m.id
WHERE
    cm.company_id = sqlc.arg('CompanyID')
    AND m.deleted_at IS NULL
ORDER BY
    m.code ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

