-- name: InsertServicePlanService :execrows
INSERT INTO service_plan_services(service_plan_id, service_id, is_active)
    VALUES (sqlc.arg('ServicePlanID'), sqlc.arg('ServiceID'), sqlc.narg('IsActive'));

-- name: GetServicePlanService :one
SELECT
    sps.id,
    sps.service_plan_id,
    sps.service_id,
    sps.is_active,
    sps.created_at
FROM
    service_plan_services sps
WHERE
    sps.service_plan_id = sqlc.arg('ServicePlanID')
    AND sps.service_id = sqlc.arg('ServiceID')
LIMIT 1;

-- name: UpdateServicePlanService :execrows
UPDATE
    service_plan_services
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active)
WHERE
    service_plan_id = sqlc.arg('ServicePlanID')
    AND service_id = sqlc.arg('ServiceID');

-- name: DeleteServicePlanService :execrows
DELETE FROM service_plan_services
WHERE service_plan_id = sqlc.arg('ServicePlanID')
    AND service_id = sqlc.arg('ServiceID');

-- name: ListServicePlanServices :many
SELECT
    sps.id,
    sps.service_plan_id,
    sps.service_id,
    sps.is_active,
    sps.created_at,
    s.title AS service_title,
    s.description AS service_description,
    s.price AS service_price,
    s.discount_rate AS service_discount_rate,
    s.image_url AS service_image_url
FROM
    service_plan_services sps
    JOIN services s ON sps.service_id = s.id
WHERE
    sps.service_plan_id = sqlc.arg('ServicePlanID')
    AND s.deleted_at IS NULL
ORDER BY
    s.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

