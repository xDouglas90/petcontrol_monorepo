-- name: InsertServicePlan :one
INSERT INTO service_plans(plan_type_id, title, description, notes, price, discount_rate, image_url, is_active)
    VALUES (sqlc.arg('PlanTypeID'), sqlc.arg('Title'), sqlc.arg('Description'), sqlc.narg('Notes'), sqlc.arg('Price'), sqlc.narg('DiscountRate'), sqlc.narg('ImageURL'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetServicePlanByID :one
SELECT
    sp.id,
    sp.plan_type_id,
    sp.title,
    sp.description,
    sp.notes,
    sp.price,
    sp.discount_rate,
    sp.image_url,
    sp.is_active,
    sp.created_at,
    sp.updated_at,
    sp.deleted_at
FROM
    service_plans sp
WHERE
    sp.id = sqlc.arg('ID')
    AND sp.deleted_at IS NULL
LIMIT 1;

-- name: UpdateServicePlan :one
UPDATE
    service_plans
SET
    plan_type_id = coalesce(sqlc.narg('PlanTypeID'), plan_type_id),
    title = coalesce(sqlc.narg('Title'), title),
    description = coalesce(sqlc.narg('Description'), description),
    notes = coalesce(sqlc.narg('Notes'), notes),
    price = coalesce(sqlc.narg('Price'), price),
    discount_rate = coalesce(sqlc.narg('DiscountRate'), discount_rate),
    image_url = coalesce(sqlc.narg('ImageURL'), image_url),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteServicePlan :one
UPDATE
    service_plans
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: ListServicePlans :many
SELECT
    sp.id,
    sp.plan_type_id,
    sp.title,
    sp.description,
    sp.notes,
    sp.price,
    sp.discount_rate,
    sp.image_url,
    sp.is_active,
    sp.created_at,
    sp.updated_at,
    sp.deleted_at
FROM
    service_plans sp
WHERE
    sp.deleted_at IS NULL
ORDER BY
    sp.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListActiveServicePlans :many
SELECT
    sp.id,
    sp.plan_type_id,
    sp.title,
    sp.description,
    sp.notes,
    sp.price,
    sp.discount_rate,
    sp.image_url,
    sp.is_active,
    sp.created_at,
    sp.updated_at,
    sp.deleted_at
FROM
    service_plans sp
WHERE
    sp.is_active = TRUE
    AND sp.deleted_at IS NULL
ORDER BY
    sp.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

