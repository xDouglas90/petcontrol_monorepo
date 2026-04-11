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
    pi.full_name AS client_name,
    p.name AS pet_name,
    COALESCE(service_context.service_ids, '{}'::text[]) AS service_ids,
    COALESCE(service_context.service_titles, '{}'::text[]) AS service_titles,
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
    INNER JOIN clients c ON c.id = s.client_id
    INNER JOIN people_identifications pi ON pi.person_id = c.person_id
    INNER JOIN pets p ON p.id = s.pet_id
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
    LEFT JOIN LATERAL (
        SELECT
            array_agg(ss.service_id::text ORDER BY svc.title) AS service_ids,
            array_agg(svc.title ORDER BY svc.title) AS service_titles
        FROM
            schedule_services ss
            INNER JOIN services svc ON svc.id = ss.service_id
        WHERE
            ss.schedule_id = s.id
            AND svc.deleted_at IS NULL
    ) service_context ON TRUE
WHERE
    s.id = sqlc.arg('ID')
    AND s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
LIMIT 1;

-- name: ListSchedulesByCompanyID :many
SELECT
    COUNT(*) OVER() AS total_count,
    s.id,
    s.company_id,
    s.client_id,
    s.pet_id,
    pi.full_name AS client_name,
    p.name AS pet_name,
    COALESCE(service_context.service_ids, '{}'::text[]) AS service_ids,
    COALESCE(service_context.service_titles, '{}'::text[]) AS service_titles,
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
    INNER JOIN clients c ON c.id = s.client_id
    INNER JOIN people_identifications pi ON pi.person_id = c.person_id
    INNER JOIN pets p ON p.id = s.pet_id
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
    LEFT JOIN LATERAL (
        SELECT
            array_agg(ss.service_id::text ORDER BY svc.title) AS service_ids,
            array_agg(svc.title ORDER BY svc.title) AS service_titles
        FROM
            schedule_services ss
            INNER JOIN services svc ON svc.id = ss.service_id
        WHERE
            ss.schedule_id = s.id
            AND svc.deleted_at IS NULL
    ) service_context ON TRUE
WHERE
    s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
    AND (
        sqlc.arg('Search')::text = ''
        OR pi.full_name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR p.name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR s.notes ILIKE '%' || sqlc.arg('Search')::text || '%'
    )
ORDER BY
    s.scheduled_at ASC
LIMIT sqlc.arg('Limit') OFFSET sqlc.arg('Offset');

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

-- name: InsertScheduleService :one
INSERT INTO schedule_services (
    schedule_id,
    service_id
)
VALUES (
    sqlc.arg('ScheduleID'),
    sqlc.arg('ServiceID')
)
RETURNING *;

-- name: DeleteScheduleServicesByScheduleID :execrows
DELETE FROM schedule_services
WHERE
    schedule_id = sqlc.arg('ScheduleID');

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
            AND p.is_active = TRUE
    ) AS is_valid;
