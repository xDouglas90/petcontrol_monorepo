-- name: InsertClientServicePlan :one
INSERT INTO client_service_plans(client_id, service_plan_id, started_at, expires_at, price_paid, is_active)
    VALUES (sqlc.arg('ClientID'), sqlc.arg('ServicePlanID'), sqlc.narg('StartedAt'), sqlc.arg('ExpiresAt'), sqlc.arg('PricePaid'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetClientServicePlan :one
SELECT
    csp.id,
    csp.client_id,
    csp.service_plan_id,
    csp.started_at,
    csp.expires_at,
    csp.price_paid,
    csp.is_active,
    csp.created_at,
    csp.updated_at
FROM
    client_service_plans csp
WHERE
    csp.client_id = sqlc.arg('ClientID')
    AND csp.service_plan_id = sqlc.arg('ServicePlanID')
LIMIT 1;

-- name: UpdateClientServicePlan :execrows
UPDATE
    client_service_plans
SET
    expires_at = coalesce(sqlc.narg('ExpiresAt'), expires_at),
    price_paid = coalesce(sqlc.narg('PricePaid'), price_paid),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    client_id = sqlc.arg('ClientID')
    AND service_plan_id = sqlc.arg('ServicePlanID');

-- name: DeactivateClientServicePlan :execrows
UPDATE
    client_service_plans
SET
    is_active = FALSE,
    updated_at = now()
WHERE
    client_id = sqlc.arg('ClientID')
    AND service_plan_id = sqlc.arg('ServicePlanID')
    AND is_active = TRUE;

-- name: DeleteClientServicePlan :execrows
DELETE FROM client_service_plans
WHERE client_id = sqlc.arg('ClientID')
    AND service_plan_id = sqlc.arg('ServicePlanID');

-- name: ListClientServicePlans :many
SELECT
    csp.id,
    csp.client_id,
    csp.service_plan_id,
    csp.started_at,
    csp.expires_at,
    csp.price_paid,
    csp.is_active,
    csp.created_at,
    csp.updated_at,
    sp.title AS service_plan_title,
    sp.description AS service_plan_description,
    sp.price AS service_plan_price,
    sp.discount_rate AS service_plan_discount_rate,
    sp.image_url AS service_plan_image_url
FROM
    client_service_plans csp
    JOIN service_plans sp ON csp.service_plan_id = sp.id
WHERE
    csp.client_id = sqlc.arg('ClientID')
ORDER BY
    csp.started_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

