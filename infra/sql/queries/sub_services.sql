-- name: InsertSubService :one
INSERT INTO sub_services(service_id, type_id, title, description, notes, price, discount_rate, image_url, is_active)
    VALUES (sqlc.arg('ServiceID'), sqlc.narg('TypeID'), sqlc.arg('Title'), sqlc.arg('Description'), sqlc.narg('Notes'), sqlc.arg('Price'), sqlc.narg('DiscountRate'), sqlc.narg('ImageURL'), sqlc.narg('IsActive'))
RETURNING
    *;

-- name: GetSubServiceByID :one
SELECT
    ss.id,
    ss.service_id,
    ss.type_id,
    ss.title,
    ss.description,
    ss.notes,
    ss.price,
    ss.discount_rate,
    ss.image_url,
    ss.is_active,
    ss.created_at,
    ss.updated_at,
    ss.deleted_at
FROM
    sub_services ss
WHERE
    ss.id = sqlc.arg('ID')
    AND ss.deleted_at IS NULL
LIMIT 1;

-- name: UpdateSubService :one
UPDATE
    sub_services
SET
    type_id = coalesce(sqlc.narg('TypeID'), type_id),
    title = coalesce(sqlc.narg('Title'), title),
    description = coalesce(sqlc.narg('Description'), description),
    notes = coalesce(sqlc.narg('Notes'), notes),
    price = coalesce(sqlc.narg('Price'), price),
    discount_rate = coalesce(sqlc.narg('DiscountRate'), discount_rate),
    image_url = coalesce(sqlc.narg('ImageURL'), image_url),
    is_active = coalesce(sqlc.narg('IsActive'), is_active),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteSubService :one
UPDATE
    sub_services
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteSubServicesByServiceID :execrows
UPDATE
    sub_services
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    service_id = sqlc.arg('ServiceID')
    AND deleted_at IS NULL;

-- name: ListSubServicesByServiceID :many
SELECT
    ss.id,
    ss.service_id,
    ss.type_id,
    ss.title,
    ss.description,
    ss.notes,
    ss.price,
    ss.discount_rate,
    ss.image_url,
    ss.is_active,
    ss.created_at,
    ss.updated_at,
    ss.deleted_at
FROM
    sub_services ss
WHERE
    ss.service_id = sqlc.arg('ServiceID')
    AND ss.deleted_at IS NULL
ORDER BY
    ss.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');
