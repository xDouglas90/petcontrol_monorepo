-- name: InsertSchedulePayment :one
INSERT INTO schedule_payments(schedule_id, payment_method_one, payment_method_two, payment_method_three, payment_date, gross_value, discount_value, net_value, amount_paid, notes)
    VALUES (sqlc.arg('ScheduleID'), sqlc.arg('PaymentMethodOne'), sqlc.narg('PaymentMethodTwo'), sqlc.narg('PaymentMethodThree'), sqlc.arg('PaymentDate'), sqlc.arg('GrossValue'), sqlc.narg('DiscountValue'), sqlc.arg('NetValue'), sqlc.narg('AmountPaid'), sqlc.narg('Notes'))
RETURNING
    *;

-- name: GetSchedulePaymentByScheduleID :one
SELECT
    sp.id,
    sp.schedule_id,
    sp.payment_method_one,
    sp.payment_method_two,
    sp.payment_method_three,
    sp.payment_date,
    sp.gross_value,
    sp.discount_value,
    sp.net_value,
    sp.amount_paid,
    sp.amount_remaining,
    sp.notes,
    sp.created_at,
    sp.updated_at
FROM
    schedule_payments sp
WHERE
    sp.schedule_id = sqlc.arg('ScheduleID')
LIMIT 1;

-- name: UpdateSchedulePayment :execrows
UPDATE
    schedule_payments
SET
    payment_method_one = coalesce(sqlc.narg('PaymentMethodOne'), payment_method_one),
    payment_method_two = coalesce(sqlc.narg('PaymentMethodTwo'), payment_method_two),
    payment_method_three = coalesce(sqlc.narg('PaymentMethodThree'), payment_method_three),
    payment_date = coalesce(sqlc.narg('PaymentDate'), payment_date),
    gross_value = coalesce(sqlc.narg('GrossValue'), gross_value),
    discount_value = coalesce(sqlc.narg('DiscountValue'), discount_value),
    net_value = coalesce(sqlc.narg('NetValue'), net_value),
    amount_paid = coalesce(sqlc.narg('AmountPaid'), amount_paid),
    notes = coalesce(sqlc.narg('Notes'), notes),
    updated_at = now()
WHERE
    schedule_id = sqlc.arg('ScheduleID');

-- name: DeleteSchedulePayment :execrows
DELETE FROM schedule_payments
WHERE schedule_id = sqlc.arg('ScheduleID');

-- name: ListSchedulePaymentsByCompanyID :many
SELECT
    sp.id,
    sp.schedule_id,
    sp.payment_method_one,
    sp.payment_method_two,
    sp.payment_method_three,
    sp.payment_date,
    sp.gross_value,
    sp.discount_value,
    sp.net_value,
    sp.amount_paid,
    sp.amount_remaining,
    sp.notes,
    sp.created_at,
    sp.updated_at,
    s.scheduled_at,
    s.client_id,
    s.pet_id
FROM
    schedule_payments sp
    JOIN schedules s ON sp.schedule_id = s.id
WHERE
    s.company_id = sqlc.arg('CompanyID')
    AND s.deleted_at IS NULL
ORDER BY
    sp.payment_date DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

