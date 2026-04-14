-- name: InsertClientPlan :one
INSERT INTO client_plans(client_id, plan_id, started_at, expires_at, price_paid, is_active)
    VALUES (sqlc.arg('ClientID'), sqlc.arg('PlanID'), sqlc.narg('StartedAt'), sqlc.arg('ExpiresAt'), sqlc.arg('PricePaid'), sqlc.narg('IsActive'))
RETURNING *;

-- name: UpdateClientPlan :execrows
UPDATE
    client_plans
SET
    plan_id = COALESCE(sqlc.narg('PlanID'), plan_id),
    started_at = COALESCE(sqlc.narg('StartedAt'), started_at),
    expires_at = COALESCE(sqlc.narg('ExpiresAt'), expires_at),
    price_paid = COALESCE(sqlc.narg('PricePaid'), price_paid),
    is_active = COALESCE(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetClientPlanByID :one
SELECT
    cp.id,
    cp.client_id,
    cp.plan_id,
    cp.started_at,
    cp.expires_at,
    cp.price_paid,
    cp.is_active,
    cp.created_at,
    cp.updated_at,
    p.name AS plan_name,
    p.description AS plan_description,
    p.price AS plan_price,
    p.billing_cycle_days AS plan_billing_cycle_days,
    p.max_users AS plan_max_users,
    p.is_active AS plan_is_active,
    p.image_url AS plan_image_url,
    p.created_at AS plan_created_at,
    p.updated_at AS plan_updated_at,
    pi.full_name AS identifications_full_name,
    PI.short_name AS identifications_short_name,
    pi.gender_identity AS identifications_gender_identity,
    pi.marital_status AS identifications_marital_status,
    pi.image_url AS identifications_image_url,
    pi.birth_date AS identifications_birth_date,
    pi.cpf AS identifications_cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at,
    c.client_since,
    c.recommended_by,
    c.notes AS client_notes
FROM
    client_plans cp
    JOIN plans p ON cp.plan_id = p.id
    JOIN clients c ON cp.client_id = c.id
    JOIN people_identifications pi ON c.person_id = pi.person_id
WHERE
    cp.id = sqlc.arg('ID')
LIMIT 1;

-- name: ListCompanyClientPlans :many
SELECT
    cp.id,
    cp.client_id,
    cp.plan_id,
    cp.started_at,
    cp.expires_at,
    cp.price_paid,
    cp.is_active,
    cp.created_at,
    cp.updated_at,
    p.name AS plan_name,
    p.description AS plan_description,
    p.price AS plan_price,
    p.billing_cycle_days AS plan_billing_cycle_days,
    p.max_users AS plan_max_users,
    p.is_active AS plan_is_active,
    p.image_url AS plan_image_url,
    p.created_at AS plan_created_at,
    p.updated_at AS plan_updated_at,
    pi.full_name AS identifications_full_name,
    PI.short_name AS identifications_short_name,
    pi.gender_identity AS identifications_gender_identity,
    pi.marital_status AS identifications_marital_status,
    pi.image_url AS identifications_image_url,
    pi.birth_date AS identifications_birth_date,
    pi.cpf AS identifications_cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at,
    c.client_since,
    c.recommended_by,
    c.notes AS client_notes
FROM
    client_plans cp
    JOIN plans p ON cp.plan_id = p.id
    JOIN clients c ON cp.client_id = c.id
    JOIN company_clients cc ON c.id = cc.client_id
    JOIN people_identifications pi ON c.person_id = pi.person_id
WHERE
    cc.company_id = sqlc.arg('CompanyID')
ORDER BY
    cp.created_at DESC,
    cp.id DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

