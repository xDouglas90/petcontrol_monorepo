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

-- name: ListPermissionsByCodes :many
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
    p.code = ANY (sqlc.slice('Codes')::varchar[])
ORDER BY
    p.code ASC;

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

-- name: ListTenantSettingsPermissionsByCompanyID :many
SELECT
    m.id AS module_id,
    m.code AS module_code,
    m.name AS module_name,
    m.description AS module_description,
    m.min_package AS module_min_package,
    p.id,
    p.code,
    p.description,
    p.default_roles,
    p.created_at,
    p.updated_at
FROM
    companies c
    INNER JOIN company_modules cm ON cm.company_id = c.id
    INNER JOIN modules m ON m.id = cm.module_id
    INNER JOIN module_permissions mp ON mp.module_id = m.id
    INNER JOIN permissions p ON p.id = mp.permission_id
WHERE
    c.id = sqlc.arg('CompanyID')
    AND cm.is_active = TRUE
    AND m.is_active = TRUE
    AND m.deleted_at IS NULL
    AND m.min_package != 'internal'
    AND (
        CASE c.active_package
            WHEN 'trial' THEN 1
            WHEN 'starter' THEN 1
            WHEN 'basic' THEN 2
            WHEN 'essential' THEN 3
            WHEN 'premium' THEN 4
            WHEN 'internal' THEN 5
            ELSE 0
        END
    ) >= (
        CASE m.min_package
            WHEN 'trial' THEN 1
            WHEN 'starter' THEN 1
            WHEN 'basic' THEN 2
            WHEN 'essential' THEN 3
            WHEN 'premium' THEN 4
            WHEN 'internal' THEN 5
            ELSE 0
        END
    )
ORDER BY
    m.min_package ASC,
    m.code ASC,
    p.code ASC;

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
