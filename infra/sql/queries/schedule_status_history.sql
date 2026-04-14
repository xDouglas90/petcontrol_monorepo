-- name: InsertScheduleStatusHistory :one
INSERT INTO schedule_status_history(schedule_id, status, changed_by, notes)
    VALUES (sqlc.arg('ScheduleID'), sqlc.arg('Status'), sqlc.narg('ChangedBy'), sqlc.narg('Notes'))
RETURNING
    *;

-- name: GetLatestScheduleStatus :one
SELECT
    ssh.id,
    ssh.schedule_id,
    ssh.status,
    ssh.changed_at,
    ssh.changed_by,
    ssh.notes
FROM
    schedule_status_history ssh
WHERE
    ssh.schedule_id = sqlc.arg('ScheduleID')
ORDER BY
    ssh.changed_at DESC
LIMIT 1;

-- name: ListScheduleStatusHistoryByScheduleID :many
SELECT
    ssh.id,
    ssh.schedule_id,
    ssh.status,
    ssh.changed_at,
    ssh.changed_by,
    ssh.notes
FROM
    schedule_status_history ssh
    INNER JOIN schedules s ON s.id = ssh.schedule_id
WHERE
    ssh.schedule_id = sqlc.arg('ScheduleID')
    AND s.company_id = sqlc.arg('CompanyID')
ORDER BY
    ssh.changed_at DESC;

