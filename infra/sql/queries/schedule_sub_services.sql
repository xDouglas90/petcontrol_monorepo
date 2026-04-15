-- name: InsertScheduleSubService :one
INSERT INTO schedule_sub_services (schedule_id, service_id, sub_service_id)
    VALUES (sqlc.arg('ScheduleID'), sqlc.arg('ServiceID'), sqlc.arg('SubServiceID'))
RETURNING *;

-- name: GetScheduleSubService :one
SELECT
    sss.id,
    sss.schedule_id,
    sss.service_id,
    sss.sub_service_id,
    sss.created_at
FROM
    schedule_sub_services sss
WHERE
    sss.schedule_id = sqlc.arg('ScheduleID')
    AND sss.sub_service_id = sqlc.arg('SubServiceID')
LIMIT 1;

-- name: DeleteScheduleSubService :execrows
DELETE FROM schedule_sub_services
WHERE schedule_id = sqlc.arg('ScheduleID')
    AND sub_service_id = sqlc.arg('SubServiceID');

-- name: DeleteScheduleSubServicesByScheduleID :execrows
DELETE FROM schedule_sub_services
WHERE schedule_id = sqlc.arg('ScheduleID');

-- name: ListScheduleSubServices :many
SELECT
    sss.id,
    sss.schedule_id,
    sss.service_id,
    sss.sub_service_id,
    sss.created_at,
    ss.title AS sub_service_title,
    ss.description AS sub_service_description,
    ss.price AS sub_service_price,
    ss.discount_rate AS sub_service_discount_rate,
    ss.image_url AS sub_service_image_url
FROM
    schedule_sub_services sss
    JOIN sub_services ss ON sss.sub_service_id = ss.id
WHERE
    sss.schedule_id = sqlc.arg('ScheduleID')
    AND ss.deleted_at IS NULL
ORDER BY
    ss.title ASC;

