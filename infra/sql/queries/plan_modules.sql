-- name: InsertPlanModule :execrows
INSERT INTO plan_modules(plan_id, module_id)
    VALUES (sqlc.arg('PlanID'), sqlc.arg('ModuleID'));

-- name: BulkInsertPlanModules :execrows
INSERT INTO plan_modules(plan_id, module_id)
SELECT
    sqlc.arg('PlanID'),
    unnest(sqlc.arg('ModuleIDs')::uuid[]);

-- name: DeletePlanModule :execrows
DELETE FROM plan_modules
WHERE plan_id = sqlc.arg('PlanID')
    AND module_id = sqlc.arg('ModuleID');

-- name: ListModulesByPlanID :many
SELECT
    m.id,
    m.code,
    m."name",
    m.description,
    m.created_at,
    m.updated_at,
    m.deleted_at
FROM
    plan_modules pm
    JOIN modules m ON pm.module_id = m.id
WHERE
    pm.plan_id = sqlc.arg('PlanID')
    AND m.deleted_at IS NULL
ORDER BY
    m.code ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListPlansByModuleID :many
SELECT
    p.id,
    p.plan_type_id,
    p.name,
    p.description,
    p.package,
    p.price,
    p.billing_cycle_days,
    p.max_users,
    p.is_active,
    p.image_url,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM
    plan_modules pm
    JOIN plans p ON pm.plan_id = p.id
WHERE
    pm.module_id = sqlc.arg('ModuleID')
    AND p.deleted_at IS NULL
ORDER BY
    p.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

