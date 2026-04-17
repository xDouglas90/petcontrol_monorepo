-- name: InsertPlanPermission :execrows
INSERT INTO plan_permissions(plan_id, permission_id)
    VALUES (sqlc.arg('PlanID'), sqlc.arg('PermissionID'));

-- name: BulkInsertPlanPermissions :execrows
INSERT INTO plan_permissions(plan_id, permission_id)
SELECT
    sqlc.arg('PlanID'),
    unnest(sqlc.arg('PermissionIDs')::uuid[]);

-- name: ListPermissionsByPlanID :many
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    plan_permissions pp
    JOIN permissions p ON pp.permission_id = p.id
WHERE
    pp.plan_id = sqlc.arg('PlanID')
    AND p.deleted_at IS NULL
ORDER BY
    p.code ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListPlansByPermissionID :many
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
    p.updated_at
FROM
    plan_permissions pp
    JOIN plans p ON pp.plan_id = p.id
WHERE
    pp.permission_id = sqlc.arg('PermissionID')
    AND p.deleted_at IS NULL
ORDER BY
    p.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

