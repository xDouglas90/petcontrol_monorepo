-- name: CreateSchedule :one
INSERT INTO schedules (
    company_id,
    client_id,
    pet_id,
    scheduled_at,
    estimated_end,
    notes,
    created_by
)
VALUES (
    sqlc.arg('CompanyID'),
    sqlc.arg('ClientID'),
    sqlc.arg('PetID'),
    sqlc.arg('ScheduledAt'),
    sqlc.narg('EstimatedEnd'),
    sqlc.narg('Notes'),
    sqlc.narg('CreatedBy')
)
RETURNING *;

-- name: GetScheduleByIDAndCompanyID :one
SELECT
    s.id,
    s.company_id,
    s.client_id,
    s.pet_id,
    s.scheduled_at,
    s.estimated_end,
    s.notes,
    s.created_by,
    s.created_at,
    s.updated_at,
    s.deleted_at,
    COALESCE(ssh_current.status, 'waiting'::schedule_status) AS current_status
FROM
    schedules s
    LEFT JOIN LATERAL (
        SELECT
            ssh.status
        FROM
            schedule_status_history ssh
        WHERE
            ssh.schedule_id = s.id
        ORDER BY
            ssh.changed_at DESC
        LIMIT 1
    ) ssh_current ON TRUE
WHERE
    s.id = sqlc.arg('ID')
    AND s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
LIMIT 1;

-- name: ListSchedulesByCompanyID :many
SELECT
    s.id,
    s.company_id,
    s.client_id,
    s.pet_id,
    s.scheduled_at,
    s.estimated_end,
    s.notes,
    s.created_by,
    s.created_at,
    s.updated_at,
    s.deleted_at,
    COALESCE(ssh_current.status, 'waiting'::schedule_status) AS current_status
FROM
    schedules s
    LEFT JOIN LATERAL (
        SELECT
            ssh.status
        FROM
            schedule_status_history ssh
        WHERE
            ssh.schedule_id = s.id
        ORDER BY
            ssh.changed_at DESC
        LIMIT 1
    ) ssh_current ON TRUE
WHERE
    s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
ORDER BY
    s.scheduled_at ASC;

-- name: UpdateSchedule :execrows
UPDATE
    schedules
SET
    client_id = COALESCE(sqlc.narg('ClientID'), client_id),
    pet_id = COALESCE(sqlc.narg('PetID'), pet_id),
    scheduled_at = COALESCE(sqlc.narg('ScheduledAt'), scheduled_at),
    estimated_end = COALESCE(sqlc.narg('EstimatedEnd'), estimated_end),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: DeleteSchedule :execrows
UPDATE
    schedules
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: InsertScheduleStatusHistory :one
INSERT INTO schedule_status_history (
    schedule_id,
    status,
    changed_by,
    notes
)
VALUES (
    sqlc.arg('ScheduleID'),
    sqlc.arg('Status'),
    sqlc.narg('ChangedBy'),
    sqlc.narg('Notes')
)
RETURNING *;

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

-- name: ValidateScheduleOwnership :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            company_clients cc
            INNER JOIN clients c ON c.id = cc.client_id
            INNER JOIN pets p ON p.id = sqlc.arg('PetID')
        WHERE
            cc.company_id = sqlc.arg('CompanyID')
            AND cc.client_id = sqlc.arg('ClientID')
            AND cc.client_id = p.owner_id
            AND cc.is_active = TRUE
            AND c.deleted_at IS NULL
            AND p.deleted_at IS NULL
    ) AS is_valid;
