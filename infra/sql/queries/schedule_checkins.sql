-- name: InsertScheduleCheckin :one
INSERT INTO schedule_checkins(schedule_id, checked_in_at, check_in_person_id, check_in_notes, check_in_photo_url)
    VALUES (sqlc.arg('ScheduleID'), sqlc.narg('CheckedInAt'), sqlc.narg('CheckInPersonID'), sqlc.narg('CheckInNotes'), sqlc.narg('CheckInPhotoURL'))
RETURNING
    *;

-- name: GetScheduleCheckinByScheduleID :one
SELECT
    sc.id,
    sc.schedule_id,
    sc.checked_in_at,
    sc.check_in_person_id,
    sc.check_in_notes,
    sc.check_in_photo_url,
    sc.checked_out_at,
    sc.check_out_person_id,
    sc.check_out_notes,
    sc.check_out_photo_url,
    sc.service_executor_person_id,
    sc.created_at,
    sc.updated_at
FROM
    schedule_checkins sc
WHERE
    sc.schedule_id = sqlc.arg('ScheduleID')
LIMIT 1;

-- name: UpdateScheduleCheckin :execrows
UPDATE
    schedule_checkins
SET
    checked_in_at = coalesce(sqlc.narg('CheckedInAt'), checked_in_at),
    check_in_person_id = coalesce(sqlc.narg('CheckInPersonID'), check_in_person_id),
    check_in_notes = coalesce(sqlc.narg('CheckInNotes'), check_in_notes),
    check_in_photo_url = coalesce(sqlc.narg('CheckInPhotoURL'), check_in_photo_url),
    updated_at = now()
WHERE
    schedule_id = sqlc.arg('ScheduleID');

-- name: UpdateScheduleCheckout :execrows
UPDATE
    schedule_checkins
SET
    checked_out_at = coalesce(sqlc.narg('CheckedOutAt'), checked_out_at),
    check_out_person_id = coalesce(sqlc.narg('CheckOutPersonID'), check_out_person_id),
    check_out_notes = coalesce(sqlc.narg('CheckOutNotes'), check_out_notes),
    check_out_photo_url = coalesce(sqlc.narg('CheckOutPhotoURL'), check_out_photo_url),
    service_executor_person_id = coalesce(sqlc.narg('ServiceExecutorPersonID'), service_executor_person_id),
    updated_at = now()
WHERE
    schedule_id = sqlc.arg('ScheduleID');

-- name: DeleteScheduleCheckin :execrows
DELETE FROM schedule_checkins
WHERE schedule_id = sqlc.arg('ScheduleID');

