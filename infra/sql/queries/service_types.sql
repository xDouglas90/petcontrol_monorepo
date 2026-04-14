-- name: InsertServiceType :one
INSERT INTO service_types(name, description)
    VALUES (sqlc.arg('Name'), sqlc.narg('Description'))
RETURNING
    *;

-- name: GetServiceTypeByID :one
SELECT
    st.id,
    st.name,
    st.description,
    st.created_at,
    st.updated_at,
    st.deleted_at
FROM
    service_types st
WHERE
    st.id = sqlc.arg('ID')
    AND st.deleted_at IS NULL
LIMIT 1;

-- name: UpdateServiceType :one
UPDATE
    service_types
SET
    name = coalesce(sqlc.narg('Name'), name),
    description = coalesce(sqlc.narg('Description'), description),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteServiceType :one
UPDATE
    service_types
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: ListServiceTypes :many
SELECT
    st.id,
    st.name,
    st.description,
    st.created_at,
    st.updated_at,
    st.deleted_at
FROM
    service_types st
WHERE
    st.deleted_at IS NULL
ORDER BY
    st.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

