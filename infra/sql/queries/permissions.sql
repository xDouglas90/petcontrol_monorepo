-- name: InsertPermission :execrows
INSERT INTO permissions(code, "name", description)
    VALUES (sqlc.arg('Code'), sqlc.arg('Name'), sqlc.narg('Description'));

-- name: GetPermissionByCode :one
SELECT
    p.id,
    p.code,
    p."name",
    p.description,
    p.created_at,
    p.updated_at
FROM
    permissions p
WHERE
    p.code = sqlc.arg('Code');

-- name: ListPermissions :many
SELECT
    p.id,
    p.code,
    p."name",
    p.description,
    p.created_at,
    p.updated_at
FROM
    permissions p
WHERE
    p.deleted_at IS NULL
ORDER BY
    p.code ASC;

-- name: UpdatePermission :execrows
UPDATE
    permissions
SET
    code = coalesce(sqlc.narg('Code'), code),
    "name" = coalesce(sqlc.narg('Name'), "name"),
    description = coalesce(sqlc.narg('Description'), description),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: DeletePermission :execrows
DELETE FROM permissions
WHERE id = sqlc.arg('ID');

