-- name: InsertServicePlanSubService :execrows
INSERT INTO service_plan_sub_services(service_plan_id, sub_service_id, is_active)
    VALUES (sqlc.arg('ServicePlanID'), sqlc.arg('SubServiceID'), sqlc.narg('IsActive'));

-- name: GetServicePlanSubService :one
SELECT
    spss.id,
    spss.service_plan_id,
    spss.sub_service_id,
    spss.is_active,
    spss.created_at
FROM
    service_plan_sub_services spss
WHERE
    spss.service_plan_id = sqlc.arg('ServicePlanID')
    AND spss.sub_service_id = sqlc.arg('SubServiceID')
LIMIT 1;

-- name: UpdateServicePlanSubService :execrows
UPDATE
    service_plan_sub_services
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active)
WHERE
    service_plan_id = sqlc.arg('ServicePlanID')
    AND sub_service_id = sqlc.arg('SubServiceID');

-- name: DeleteServicePlanSubService :execrows
DELETE FROM service_plan_sub_services
WHERE service_plan_id = sqlc.arg('ServicePlanID')
    AND sub_service_id = sqlc.arg('SubServiceID');

-- name: ListServicePlanSubServices :many
SELECT
    spss.id,
    spss.service_plan_id,
    spss.sub_service_id,
    spss.is_active,
    spss.created_at,
    ss.title AS sub_service_title,
    ss.description AS sub_service_description,
    ss.price AS sub_service_price,
    ss.discount_rate AS sub_service_discount_rate,
    ss.image_url AS sub_service_image_url
FROM
    service_plan_sub_services spss
    JOIN sub_services ss ON spss.sub_service_id = ss.id
WHERE
    spss.service_plan_id = sqlc.arg('ServicePlanID')
    AND ss.deleted_at IS NULL
ORDER BY
    ss.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

