-- name: InsertWarning :one
INSERT INTO warnings(company_id, title, content, image_url, sender_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('Title'), sqlc.arg('Content'), sqlc.narg('ImageURL'), sqlc.arg('SenderID'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetWarningByID :one
SELECT
    w.id,
    w.company_id,
    w.title,
    w.content,
    w.image_url,
    w.sender_id,
    w.is_active,
    w.created_at,
    w.updated_at,
    w.deleted_at
FROM
    warnings w
WHERE
    w.id = sqlc.arg('ID')
    AND w.company_id = sqlc.arg('CompanyID')
    AND w.deleted_at IS NULL
LIMIT 1;

-- name: UpdateWarning :execrows
UPDATE
    warnings
SET
    title = coalesce(sqlc.narg('Title'), title),
    content = coalesce(sqlc.narg('Content'), content),
    image_url = coalesce(sqlc.narg('ImageURL'), image_url),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: DeleteWarning :execrows
UPDATE
    warnings
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL;

-- name: ListWarningsByCompanyID :many
SELECT
    w.id,
    w.company_id,
    w.title,
    w.content,
    w.image_url,
    w.sender_id,
    w.is_active,
    w.created_at,
    w.updated_at,
    w.deleted_at
FROM
    warnings w
WHERE
    w.company_id = sqlc.arg('CompanyID')
    AND w.deleted_at IS NULL
ORDER BY
    w.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: ListActiveWarningsByCompanyID :many
SELECT
    w.id,
    w.company_id,
    w.title,
    w.content,
    w.image_url,
    w.sender_id,
    w.is_active,
    w.created_at,
    w.updated_at,
    w.deleted_at
FROM
    warnings w
WHERE
    w.company_id = sqlc.arg('CompanyID')
    AND w.is_active = TRUE
    AND w.deleted_at IS NULL
ORDER BY
    w.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

