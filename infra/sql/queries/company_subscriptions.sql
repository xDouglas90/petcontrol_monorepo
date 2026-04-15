-- name: InsertCOmpanySubscription :execrows
INSERT INTO company_subscriptions(company_id, plan_id, started_at, expires_at, canceled_at, price_paid, notes)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('PlanID'), sqlc.arg('StartedAt'), sqlc.arg('ExpiresAt'), sqlc.narg('CanceledAt'), sqlc.arg('PricePaid'), sqlc.narg('Notes'));

-- name: GetActiveCompanySubscriptionByCompanyID :one
SELECT
    cs.id,
    cs.company_id,
    cs.plan_id,
    cs.started_at,
    cs.expires_at,
    cs.canceled_at,
    cs.is_active,
    cs.price_paid,
    cs.notes,
    cs.created_at,
    cs.updated_at
FROM
    company_subscriptions cs
WHERE
    cs.company_id = sqlc.arg('CompanyID')
    AND cs.is_active = TRUE
ORDER BY
    cs.started_at DESC
LIMIT 1;

-- name: ListCompanySubscriptionsByCompanyID :many
SELECT
    cs.id,
    cs.company_id,
    cs.plan_id,
    cs.started_at,
    cs.expires_at,
    cs.canceled_at,
    cs.is_active,
    cs.price_paid,
    cs.notes,
    cs.created_at,
    cs.updated_at
FROM
    company_subscriptions cs
WHERE
    cs.company_id = sqlc.arg('CompanyID')
ORDER BY
    cs.started_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: CancelCompanySubscription :execrows
UPDATE
    company_subscriptions
SET
    canceled_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND plan_id = sqlc.arg('SubscriptionID')
    AND is_active = TRUE;

-- name: UpdateCompanySubscription :execrows
UPDATE
    company_subscriptions
SET
    expires_at = coalesce(sqlc.narg('ExpiresAt'), expires_at),
    price_paid = coalesce(sqlc.narg('PricePaid'), price_paid),
    notes = coalesce(sqlc.narg('Notes'), notes),
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND plan_id = sqlc.arg('SubscriptionID')
    AND is_active = TRUE;

-- name: DeleteCompanySubscription :execrows
DELETE FROM company_subscriptions
WHERE company_id = sqlc.arg('CompanyID')
    AND plan_id = sqlc.arg('SubscriptionID');

