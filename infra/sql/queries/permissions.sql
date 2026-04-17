-- name: InsertPermission :execrows
INSERT INTO permissions(code, description, default_roles)
    VALUES (sqlc.arg('Code'), sqlc.arg('Description'), sqlc.arg('DefaultRoles'));

-- name: GetPermissionByCode :one
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    permissions p
WHERE
    p.code = sqlc.arg('Code')::varchar;

-- name: ListPermissions :many
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    permissions p
ORDER BY
    p.code ASC
LIMIT sqlc.arg('Limit')::int OFFSET sqlc.arg('Offset')::int;

-- name: ListPermissionsByRole :many
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    permissions p
WHERE
    sqlc.arg('Role')::user_role_type = ANY (p.default_roles)
ORDER BY
    p.code ASC
LIMIT sqlc.arg('Limit')::int OFFSET sqlc.arg('Offset')::int;

-- name: ListPermissionsByModule :many
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    permissions p
    JOIN module_permissions mp ON p.id = mp.permission_id
WHERE
    mp.module_id = sqlc.arg('ModuleID')::uuid
ORDER BY
    p.code ASC
LIMIT sqlc.arg('Limit')::int OFFSET sqlc.arg('Offset')::int;

-- name: UpdatePermission :execrows
UPDATE
    permissions
SET
    code = coalesce(sqlc.narg('Code')::varchar, code),
    description = coalesce(sqlc.narg('Description')::varchar, description),
    default_roles = coalesce(sqlc.narg('DefaultRoles')::user_role_type[], default_roles),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')::uuid;

-- name: DeletePermission :execrows
DELETE FROM permissions
WHERE id = sqlc.arg('ID')::uuid;

