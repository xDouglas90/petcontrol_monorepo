-- name: InsertUserPermission :execrows
INSERT INTO user_permissions(user_id, permission_id, granted_by)
    VALUES (sqlc.arg('UserID'), sqlc.arg('PermissionID'), sqlc.arg('GrantedBy'));

-- name: BulkInsertUserPermissions :execrows
INSERT INTO user_permissions(user_id, permission_id, granted_by)
SELECT
    sqlc.arg('UserID'),
    unnest(sqlc.arg('PermissionIDs')::uuid[]),
    sqlc.arg('GrantedBy');

-- name: DeleteUserPermission :execrows
UPDATE
    user_permissions
SET
    is_active = FALSE,
    revoked_at = now(),
    revoked_by = sqlc.arg('RevokedBy')
WHERE
    user_id = sqlc.arg('UserID')
    AND permission_id = sqlc.arg('PermissionID')
    AND revoked_at IS NULL;

-- name: ReactivateUserPermission :execrows
UPDATE
    user_permissions
SET
    is_active = TRUE,
    granted_by = sqlc.arg('GrantedBy'),
    granted_at = now(),
    revoked_by = NULL,
    revoked_at = NULL
WHERE
    user_id = sqlc.arg('UserID')
    AND permission_id = sqlc.arg('PermissionID')
    AND revoked_at IS NOT NULL;

-- name: ListPermissionsByUserID :many
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    up.granted_by,
    up.granted_at,
    up.revoked_by,
    up.revoked_at
FROM
    user_permissions up
    JOIN permissions p ON up.permission_id = p.id
WHERE
    up.user_id = sqlc.arg('UserID')
    AND up.revoked_at IS NULL
ORDER BY
    p.code ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: GetUserPermission :one
SELECT
    p.id,
    p.code,
    p.description,
    p.default_roles,
    up.granted_by,
    up.granted_at,
    up.revoked_by,
    up.revoked_at
FROM
    user_permissions up
    JOIN permissions p ON up.permission_id = p.id
WHERE
    up.user_id = sqlc.arg('UserID')
    AND up.permission_id = sqlc.arg('PermissionID')
    AND up.revoked_at IS NULL
;

-- name: ListUsersByPermissionID :many
SELECT
    u.id,
    u.email,
    u.is_active,
    u.created_at,
    u.updated_at,
    up.granted_by,
    up.granted_at,
    up.revoked_by,
    up.revoked_at
FROM
    user_permissions up
    JOIN users u ON up.user_id = u.id
WHERE
    up.permission_id = sqlc.arg('PermissionID')
    AND up.revoked_at IS NULL
    AND u.deleted_at IS NULL
ORDER BY
    u.email ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');
