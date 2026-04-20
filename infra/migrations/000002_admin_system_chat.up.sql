CREATE TABLE admin_system_conversations(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  admin_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  system_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  CONSTRAINT chk_admin_system_conversations_users_different CHECK (admin_user_id <> system_user_id),
  UNIQUE (company_id, admin_user_id, system_user_id)
);

CREATE INDEX idx_admin_system_conversations_company ON admin_system_conversations(company_id);
CREATE INDEX idx_admin_system_conversations_admin_user ON admin_system_conversations(admin_user_id);
CREATE INDEX idx_admin_system_conversations_system_user ON admin_system_conversations(system_user_id);

CREATE TABLE admin_system_messages(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  conversation_id uuid NOT NULL REFERENCES admin_system_conversations(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  sender_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_admin_system_messages_conversation ON admin_system_messages(conversation_id, created_at ASC);
CREATE INDEX idx_admin_system_messages_company ON admin_system_messages(company_id);
CREATE INDEX idx_admin_system_messages_sender ON admin_system_messages(sender_user_id);
