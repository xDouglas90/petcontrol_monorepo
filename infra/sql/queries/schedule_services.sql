-- name: InsertScheduleService :one
INSERT INTO schedule_services (schedule_id, service_id)
    VALUES (sqlc.arg('ScheduleID'), sqlc.arg('ServiceID'))
RETURNING *;

-- name: GetScheduleService :one
SELECT
    ss.id,
    ss.schedule_id,
    ss.service_id,
    ss.created_at
FROM
    schedule_services ss
WHERE
    ss.schedule_id = sqlc.arg('ScheduleID')
    AND ss.service_id = sqlc.arg('ServiceID')
LIMIT 1;

-- name: DeleteScheduleService :execrows
DELETE FROM schedule_services
WHERE schedule_id = sqlc.arg('ScheduleID')
    AND service_id = sqlc.arg('ServiceID');

-- name: DeleteScheduleServicesByScheduleID :execrows
DELETE FROM schedule_services
WHERE schedule_id = sqlc.arg('ScheduleID');

-- name: ListScheduleServices :many
SELECT
    ss.id,
    ss.schedule_id,
    ss.service_id,
    ss.created_at,
    s.title AS service_title,
    s.description AS service_description,
    s.price AS service_price,
    s.discount_rate AS service_discount_rate,
    s.image_url AS service_image_url
FROM
    schedule_services ss
    JOIN services s ON ss.service_id = s.id
WHERE
    ss.schedule_id = sqlc.arg('ScheduleID')
    AND s.deleted_at IS NULL
ORDER BY
    s.title ASC;

