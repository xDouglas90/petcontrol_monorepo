-- name: UpsertAdminSystemConversation :one
INSERT INTO admin_system_conversations(company_id, admin_user_id, system_user_id, updated_at)
VALUES (
  sqlc.arg('CompanyID'),
  sqlc.arg('AdminUserID'),
  sqlc.arg('SystemUserID'),
  NOW()
)
ON CONFLICT (company_id, admin_user_id, system_user_id)
DO UPDATE SET
  updated_at = NOW()
RETURNING *;

-- name: ListAdminSystemMessages :many
SELECT
  asm.id,
  asm.conversation_id,
  asm.company_id,
  asm.sender_user_id,
  asm.body,
  asm.created_at,
  asm.updated_at,
  asm.deleted_at,
  COALESCE(pi.short_name, pi.full_name, sender.email) AS sender_name,
  pi.image_url AS sender_image_url,
  sender.role AS sender_role
FROM
  admin_system_messages asm
  INNER JOIN admin_system_conversations ascv ON ascv.id = asm.conversation_id
  INNER JOIN users sender ON sender.id = asm.sender_user_id
  LEFT JOIN user_profiles up ON up.user_id = sender.id
  LEFT JOIN people_identifications pi ON pi.person_id = up.person_id
WHERE
  ascv.company_id = sqlc.arg('CompanyID')
  AND ascv.admin_user_id = sqlc.arg('AdminUserID')
  AND ascv.system_user_id = sqlc.arg('SystemUserID')
  AND asm.deleted_at IS NULL
ORDER BY
  asm.created_at ASC;

-- name: InsertAdminSystemMessage :one
INSERT INTO admin_system_messages(conversation_id, company_id, sender_user_id, body)
VALUES (
  sqlc.arg('ConversationID'),
  sqlc.arg('CompanyID'),
  sqlc.arg('SenderUserID'),
  sqlc.arg('Body')
)
RETURNING *;
