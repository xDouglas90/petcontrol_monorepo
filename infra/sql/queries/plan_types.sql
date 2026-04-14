-- name: GetPlanTypeByID :one
SELECT
    pt.id,
    pt.name,
    pt.description,
    pt.created_at,
    pt.updated_at,
    pt.deleted_at
FROM
    plan_types pt
WHERE
    pt.id = sqlc.arg('ID')
    AND pt.deleted_at IS NULL
LIMIT 1;

-- name: UpdatePlanType :execrows
UPDATE
    plan_types
SET
    name = coalesce(sqlc.narg('Name'), name),
    description = coalesce(sqlc.narg('Description'), description),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL;

-- name: DeletePlanType :execrows
UPDATE
    plan_types
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL;

