-- name: InsertServicePlanBonus :execrows
INSERT INTO service_plan_bonuses(service_plan_id, service_id, sub_service_id, is_active)
    VALUES (sqlc.arg('ServicePlanID'), sqlc.narg('ServiceID'), sqlc.narg('SubServiceID'), sqlc.narg('IsActive'));

-- name: GetServicePlanBonusByID :one
SELECT
    spb.id,
    spb.service_plan_id,
    spb.service_id,
    spb.sub_service_id,
    spb.is_active,
    spb.created_at
FROM
    service_plan_bonuses spb
WHERE
    spb.id = sqlc.arg('ID')
LIMIT 1;

-- name: UpdateServicePlanBonus :execrows
UPDATE
    service_plan_bonuses
SET
    is_active = coalesce(sqlc.narg('IsActive'), is_active)
WHERE
    id = sqlc.arg('ID');

-- name: DeleteServicePlanBonus :execrows
DELETE FROM service_plan_bonuses
WHERE id = sqlc.arg('ID');

-- name: ListServicePlanBonuses :many
SELECT
    spb.id,
    spb.service_plan_id,
    spb.service_id,
    spb.sub_service_id,
    spb.is_active,
    spb.created_at,
    s.title AS service_title,
    s.price AS service_price,
    ss.title AS sub_service_title,
    ss.price AS sub_service_price
FROM
    service_plan_bonuses spb
    LEFT JOIN services s ON spb.service_id = s.id
    LEFT JOIN sub_services ss ON spb.sub_service_id = ss.id
WHERE
    spb.service_plan_id = sqlc.arg('ServicePlanID')
ORDER BY
    spb.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

