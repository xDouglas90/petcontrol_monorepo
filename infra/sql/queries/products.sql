-- name: InsertProduct :one
INSERT INTO products(name, batch_number, description, image_url, expiration_date, quantity)
    VALUES (sqlc.arg('Name'), sqlc.narg('BatchNumber'), sqlc.narg('Description'), sqlc.narg('ImageURL'), sqlc.narg('ExpirationDate'), sqlc.narg('Quantity'))
RETURNING
    *;

-- name: GetProductByID :one
SELECT
    p.id,
    p.name,
    p.batch_number,
    p.description,
    p.image_url,
    p.expiration_date,
    p.quantity,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM
    products p
WHERE
    p.id = sqlc.arg('ID')
    AND p.deleted_at IS NULL
LIMIT 1;

-- name: UpdateProduct :one
UPDATE
    products
SET
    name = coalesce(sqlc.narg('Name'), name),
    batch_number = coalesce(sqlc.narg('BatchNumber'), batch_number),
    description = coalesce(sqlc.narg('Description'), description),
    image_url = coalesce(sqlc.narg('ImageURL'), image_url),
    expiration_date = coalesce(sqlc.narg('ExpirationDate'), expiration_date),
    quantity = coalesce(sqlc.narg('Quantity'), quantity),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: UpdateProductQuantity :one
UPDATE
    products
SET
    quantity = sqlc.arg('Quantity'),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteProduct :one
UPDATE
    products
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: ListProducts :many
SELECT
    p.id,
    p.name,
    p.batch_number,
    p.description,
    p.image_url,
    p.expiration_date,
    p.quantity,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM
    products p
WHERE
    p.deleted_at IS NULL
ORDER BY
    p.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

