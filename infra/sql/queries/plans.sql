-- name: InsertPlanType :one
INSERT INTO plan_types(name, description)
    VALUES ($1, $2)
RETURNING
    *;

-- name: ListPlanTypes :many
SELECT
    pt.id,
    pt.name,
    pt.description,
    pt.created_at,
    pt.updated_at,
    pt.deleted_at
FROM
    plan_types pt
WHERE
    pt.deleted_at IS NULL
ORDER BY
    pt.created_at ASC;

-- name: InsertPlan :one
INSERT INTO plans(plan_type_id, name, description, package, price, billing_cycle_days, max_users, is_active, image_url)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    *;

-- name: GetPlanByID :one
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
    plans p
WHERE
    p.id = sqlc.arg('ID')
    AND p.deleted_at IS NULL
LIMIT 1;

-- name: ListPlans :many
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
    plans p
WHERE
    p.deleted_at IS NULL
ORDER BY
    p.created_at DESC;

-- name: UpdatePlan :execrows
UPDATE
    plans
SET
    plan_type_id = coalesce(sqlc.narg('PlanTypeID'), plan_type_id),
    name = coalesce(sqlc.narg('Name'), name),
    description = coalesce(sqlc.narg('Description'), description),
    package = coalesce(sqlc.narg('Package'), package),
    price = coalesce(sqlc.narg('Price'), price),
    billing_cycle_days = coalesce(sqlc.narg('BillingCycleDays'), billing_cycle_days),
    max_users = coalesce(sqlc.narg('MaxUsers'), max_users),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    image_url = coalesce(sqlc.narg('ImageUrl'), image_url),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL;

