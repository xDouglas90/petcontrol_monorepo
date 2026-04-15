-- name: InsertCompanyServicePlan :one
INSERT INTO company_service_plans(company_id, service_plan_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ServicePlanID'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetCompanyServicePlan :one
SELECT
    csp.id,
    csp.company_id,
    csp.service_plan_id,
    csp.is_active,
    csp.created_at,
    csp.updated_at
FROM
    company_service_plans csp
WHERE
    csp.company_id = sqlc.arg('CompanyID')
    AND csp.service_plan_id = sqlc.arg('ServicePlanID')
LIMIT 1;

-- name: UpdateCompanyServicePlan :one
UPDATE
    company_service_plans
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND service_plan_id = sqlc.arg('ServicePlanID')
RETURNING
    *;

-- name: DeleteCompanyServicePlan :one
DELETE FROM company_service_plans
WHERE company_id = sqlc.arg('CompanyID')
    AND service_plan_id = sqlc.arg('ServicePlanID')
RETURNING
    *;

-- name: ListCompanyServicePlans :many
SELECT
    csp.id,
    csp.company_id,
    csp.service_plan_id,
    csp.is_active,
    csp.created_at,
    csp.updated_at,
    sp.title AS service_plan_title,
    sp.description AS service_plan_description,
    sp.price AS service_plan_price,
    sp.discount_rate AS service_plan_discount_rate,
    sp.image_url AS service_plan_image_url,
    sp.deleted_at AS service_plan_deleted_at
FROM
    company_service_plans csp
    JOIN service_plans sp ON csp.service_plan_id = sp.id
WHERE
    csp.company_id = sqlc.arg('CompanyID')
ORDER BY
    csp.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListActiveCompanyServicePlans :many
SELECT
    csp.id,
    csp.company_id,
    csp.service_plan_id,
    csp.is_active,
    csp.created_at,
    csp.updated_at,
    sp.title AS service_plan_title,
    sp.description AS service_plan_description,
    sp.price AS service_plan_price,
    sp.discount_rate AS service_plan_discount_rate,
    sp.image_url AS service_plan_image_url
FROM
    company_service_plans csp
    JOIN service_plans sp ON csp.service_plan_id = sp.id
WHERE
    csp.company_id = sqlc.arg('CompanyID')
    AND csp.is_active = TRUE
    AND sp.is_active = TRUE
    AND sp.deleted_at IS NULL
ORDER BY
    sp.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

