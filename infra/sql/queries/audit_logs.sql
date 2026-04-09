-- name: InsertAuditLog :exec
INSERT INTO audit_logs (
    action,
    entity_table,
    entity_id,
    company_id,
    old_data,
    new_data,
    changed_by,
    ip_address,
    user_agent
)
VALUES (
    sqlc.arg('Action'),
    sqlc.arg('EntityTable'),
    sqlc.arg('EntityID'),
    sqlc.narg('CompanyID'),
    sqlc.narg('OldData'),
    sqlc.arg('NewData'),
    sqlc.narg('ChangedBy'),
    sqlc.narg('IPAddress'),
    sqlc.narg('UserAgent')
);
