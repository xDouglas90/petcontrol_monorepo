-- name: InsertNotification :one
INSERT INTO notifications(company_id, title, summary, content, send_to_whatsapp, created_by)
    VALUES (sqlc.narg('CompanyID'), sqlc.arg('Title'), sqlc.arg('Summary'), sqlc.arg('Content'), sqlc.narg('SendToWhatsapp'), sqlc.narg('CreatedBy'))
RETURNING
    *;

-- name: GetNotificationByID :one
SELECT
    n.id,
    n.company_id,
    n.title,
    n.summary,
    n.content,
    n.send_to_whatsapp,
    n.created_by,
    n.created_at,
    n.updated_at,
    n.deleted_at
FROM
    notifications n
WHERE
    n.id = sqlc.arg('ID')
    AND n.deleted_at IS NULL
LIMIT 1;

-- name: UpdateNotification :execrows
UPDATE
    notifications
SET
    title = coalesce(sqlc.narg('Title'), title),
    summary = coalesce(sqlc.narg('Summary'), summary),
    content = coalesce(sqlc.narg('Content'), content),
    send_to_whatsapp = coalesce(sqlc.narg('SendToWhatsapp'), send_to_whatsapp),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL;

-- name: DeleteNotification :execrows
UPDATE
    notifications
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL;

-- name: ListNotificationsByCompanyID :many
SELECT
    n.id,
    n.company_id,
    n.title,
    n.summary,
    n.content,
    n.send_to_whatsapp,
    n.created_by,
    n.created_at,
    n.updated_at,
    n.deleted_at
FROM
    notifications n
WHERE (n.company_id = sqlc.arg('CompanyID')
    OR n.company_id IS NULL)
AND n.deleted_at IS NULL
ORDER BY
    n.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: InsertNotificationReceiver :execrows
INSERT INTO notification_receivers(notification_id, user_id)
    VALUES (sqlc.arg('NotificationID'), sqlc.arg('UserID'));

-- name: MarkNotificationAsRead :execrows
UPDATE
    notification_receivers
SET
    read_at = now()
WHERE
    notification_id = sqlc.arg('NotificationID')
    AND user_id = sqlc.arg('UserID')
    AND read_at IS NULL;

-- name: DeleteNotificationReceiver :execrows
DELETE FROM notification_receivers
WHERE notification_id = sqlc.arg('NotificationID')
    AND user_id = sqlc.arg('UserID');

-- name: ListNotificationReceiversByUserID :many
SELECT
    nr.id,
    nr.notification_id,
    nr.user_id,
    nr.read_at,
    nr.created_at,
    n.title AS notification_title,
    n.summary AS notification_summary,
    n.company_id AS notification_company_id
FROM
    notification_receivers nr
    JOIN notifications n ON nr.notification_id = n.id
WHERE
    nr.user_id = sqlc.arg('UserID')
    AND n.deleted_at IS NULL
ORDER BY
    nr.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: CountUnreadNotificationsByUserID :one
SELECT
    COUNT(*) AS unread_count
FROM
    notification_receivers nr
    JOIN notifications n ON nr.notification_id = n.id
WHERE
    nr.user_id = sqlc.arg('UserID')
    AND nr.read_at IS NULL
    AND n.deleted_at IS NULL;

