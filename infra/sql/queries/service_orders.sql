-- name: InsertServiceOrder :one
INSERT INTO service_orders(schedule_id, printed_by, gross_total, amount_paid, amount_to_pay, has_perfume, has_ornament, notes)
    VALUES (sqlc.arg('ScheduleID'), sqlc.narg('PrintedBy'), sqlc.arg('GrossTotal'), sqlc.arg('AmountPaid'), sqlc.arg('AmountToPay'), sqlc.narg('HasPerfume'), sqlc.narg('HasOrnament'), sqlc.narg('Notes'))
RETURNING
    *;

-- name: GetServiceOrderByScheduleID :one
SELECT
    so.id,
    so.schedule_id,
    so.printed_by,
    so.printed_at,
    so.gross_total,
    so.amount_paid,
    so.amount_to_pay,
    so.has_perfume,
    so.has_ornament,
    so.notes,
    so.created_at,
    so.updated_at
FROM
    service_orders so
WHERE
    so.schedule_id = sqlc.arg('ScheduleID')
LIMIT 1;

-- name: GetServiceOrderByID :one
SELECT
    so.id,
    so.schedule_id,
    so.printed_by,
    so.printed_at,
    so.gross_total,
    so.amount_paid,
    so.amount_to_pay,
    so.has_perfume,
    so.has_ornament,
    so.notes,
    so.created_at,
    so.updated_at
FROM
    service_orders so
WHERE
    so.id = sqlc.arg('ID')
LIMIT 1;

-- name: UpdateServiceOrder :execrows
UPDATE
    service_orders
SET
    gross_total = coalesce(sqlc.narg('GrossTotal'), gross_total),
    amount_paid = coalesce(sqlc.narg('AmountPaid'), amount_paid),
    amount_to_pay = coalesce(sqlc.narg('AmountToPay'), amount_to_pay),
    has_perfume = coalesce(sqlc.narg('HasPerfume'), has_perfume),
    has_ornament = coalesce(sqlc.narg('HasOrnament'), has_ornament),
    notes = coalesce(sqlc.narg('Notes'), notes),
    updated_at = now()
WHERE
    schedule_id = sqlc.arg('ScheduleID');

-- name: ListServiceOrdersByCompanyID :many
SELECT
    so.id,
    so.schedule_id,
    so.printed_by,
    so.printed_at,
    so.gross_total,
    so.amount_paid,
    so.amount_to_pay,
    so.has_perfume,
    so.has_ornament,
    so.notes,
    so.created_at,
    so.updated_at,
    s.scheduled_at,
    s.client_id,
    s.pet_id
FROM
    service_orders so
    JOIN schedules s ON so.schedule_id = s.id
WHERE
    s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
ORDER BY
    so.printed_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

