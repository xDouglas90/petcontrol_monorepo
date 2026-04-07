-- =============================================================================
-- PETCONTROL - SCHEMA PostgreSQL v2
-- Aplicação Multi-Tenant SaaS com controle de planos e módulos
-- =============================================================================

-- Extensão necessária para UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- SEÇÃO 1: ENUMS
-- =============================================================================

-- Tipos de usuário no sistema (papel técnico/sistêmico)
CREATE TYPE user_role_type AS ENUM (
  'root',        -- Acesso total ao sistema (Anthropic/dev)
  'admin',       -- Administrador da empresa
  'manager',     -- Gerente com permissões ampliadas
  'employee',    -- Funcionário padrão
  'aux',         -- Auxiliar com acesso limitado
  'general'      -- Usuário genérico
);

-- Tipos de usuário no contexto de negócio (quem é na empresa)
CREATE TYPE user_kind AS ENUM (
  'internal',    -- Equipe interna da plataforma
  'owner',       -- Dono da empresa cliente
  'staff',       -- Funcionário da empresa cliente
  'free'         -- Conta gratuita / trial
);

-- Pacotes de módulos disponíveis nos planos
CREATE TYPE module_package AS ENUM (
  'internal',
  'starter',
  'basic',
  'essential',
  'premium'
);

-- Ações possíveis registradas em logs de auditoria
CREATE TYPE log_action AS ENUM (
  'create',
  'update',
  'delete',
  'restore',
  'login',
  'logout',
  'view',
  'export',
  'deactivate',
  'reactivate'
);

-- Identidades de gênero
CREATE TYPE gender_identity AS ENUM (
  'man_cisgender',
  'woman_cisgender',
  'transgender',
  'non_binary',
  'gender_fluid',
  'gender_queer',
  'agender',
  'gender_non_conforming',
  'not_to_expose'
);

-- Estado civil
CREATE TYPE marital_status AS ENUM (
  'single',
  'married',
  'divorced',
  'widowed',
  'separated'
);

-- Porte dos pets
CREATE TYPE pet_size AS ENUM (
  'small',
  'medium',
  'large',
  'giant'
);

-- Temperamento dos pets
CREATE TYPE pet_temperament AS ENUM (
  'calm',
  'nervous',
  'aggressive',
  'playful',
  'loving'
);

-- Tipos de pets
CREATE TYPE pet_kind AS ENUM (
  'dog',
  'cat',
  'bird',
  'fish',
  'reptile',
  'rodent',
  'rabbit',
  'other'
);

-- Tipos de funcionário
CREATE TYPE employee_kind AS ENUM (
  'internal',
  'outsourced'
);

-- Tipos de chave Pix
CREATE TYPE pix_key_kind AS ENUM (
  'cpf',
  'cnpj',
  'email',
  'phone',
  'random'
);

-- Dias da semana
CREATE TYPE week_day AS ENUM (
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday'
);

-- Níveis de escolaridade
CREATE TYPE graduation_level AS ENUM (
  'elementary_incomplete',
  'elementary_complete',
  'middle_incomplete',
  'middle_complete',
  'high_incomplete',
  'high_complete',
  'college_incomplete',
  'college_complete',
  'postgraduate_incomplete',
  'postgraduate_complete',
  'master_incomplete',
  'master_complete',
  'doctorate_incomplete',
  'doctorate_complete'
);

-- Tipos de pessoa no sistema
CREATE TYPE person_kind AS ENUM (
  'client',
  'employee',
  'outsourced_employee',
  'supplier',
  'guardian',
  'responsible'
);

-- Tipos de conta bancária
CREATE TYPE bank_account_kind AS ENUM (
  'checking',
  'savings',
  'salary'
);

-- Métodos de pagamento
CREATE TYPE payment_method AS ENUM (
  'credit_card',
  'debit_card',
  'pix',
  'cash',
  'check',
  'bank_slip',
  'transfer'
);

-- Status dos agendamentos
CREATE TYPE schedule_status AS ENUM (
  'waiting',
  'confirmed',
  'canceled',
  'in_progress',
  'finished',
  'delivered'
);

-- Tipos de produto
CREATE TYPE product_kind AS ENUM (
  'service',
  'customer',
  'internal_usage'
);

-- Resultado de tentativa de login
CREATE TYPE login_result AS ENUM (
  'success',
  'invalid_credentials',
  'account_locked',
  'account_inactive',
  'email_unverified',
  'token_expired',
  'mfa_required',
  'mfa_failed'
);

-- Razão de logout
CREATE TYPE logout_reason AS ENUM (
  'user_initiated',
  'session_expired',
  'forced_by_admin',
  'password_changed',
  'account_deactivated',
  'suspicious_activity'
);


-- =============================================================================
-- SEÇÃO 2: AUTENTICAÇÃO E SESSÕES
-- =============================================================================

-- Usuários do sistema (somente dados de acesso)
CREATE TABLE users (
  id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  email               VARCHAR(150) UNIQUE NOT NULL,
  email_verified      BOOLEAN     NOT NULL DEFAULT FALSE,
  email_verified_at   TIMESTAMPTZ DEFAULT NULL,
  role                user_role_type NOT NULL DEFAULT 'general',
  kind                user_kind   NOT NULL DEFAULT 'staff',
  is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at          TIMESTAMPTZ,
  deleted_at          TIMESTAMPTZ  -- soft delete; nunca remover fisicamente
);

CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_active   ON users(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role     ON users(role);

-- Credenciais separadas dos dados do usuário
CREATE TABLE user_auth (
  user_id               UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash         TEXT        NOT NULL,
  password_changed_at   TIMESTAMPTZ DEFAULT NULL,
  must_change_password  BOOLEAN     NOT NULL DEFAULT FALSE,
  login_attempts        SMALLINT    NOT NULL DEFAULT 0,
  locked_until          TIMESTAMPTZ DEFAULT NULL,
  last_login_at         TIMESTAMPTZ DEFAULT NULL,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at            TIMESTAMPTZ
);

-- Histórico de cada tentativa de login (imutável — sem ON DELETE CASCADE)
CREATE TABLE login_history (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id         UUID        REFERENCES users(id) ON DELETE SET NULL, -- preserva histórico
  ip_address      INET        NOT NULL,  -- usar INET ao invés de VARCHAR para IPs
  user_agent      TEXT        NOT NULL,
  result          login_result NOT NULL,
  failure_detail  VARCHAR(200) DEFAULT NULL,
  attempted_at    TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_login_history_user_id     ON login_history(user_id);
CREATE INDEX idx_login_history_attempted   ON login_history(attempted_at DESC);
CREATE INDEX idx_login_history_result      ON login_history(result);

-- Sessões ativas e históricas (separado do histórico de login)
CREATE TABLE user_sessions (
  id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  login_history_id  UUID        REFERENCES login_history(id) ON DELETE SET NULL,
  session_token     TEXT        UNIQUE NOT NULL,
  ip_address        INET        NOT NULL,
  user_agent        TEXT        NOT NULL,
  last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  expires_at        TIMESTAMPTZ NOT NULL,
  logged_out_at     TIMESTAMPTZ DEFAULT NULL,
  logout_reason     logout_reason DEFAULT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_user_sessions_user_id       ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token         ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_active        ON user_sessions(expires_at)
  WHERE logged_out_at IS NULL;

-- Tokens de recuperação de senha
CREATE TABLE password_recovery_tokens (
  id                    UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash            TEXT        NOT NULL UNIQUE,
  requested_email       TEXT        NOT NULL,
  expires_at            TIMESTAMPTZ NOT NULL,
  used_at               TIMESTAMPTZ DEFAULT NULL,
  revoked_at            TIMESTAMPTZ DEFAULT NULL,
  request_ip            INET        DEFAULT NULL,
  request_user_agent    TEXT        DEFAULT NULL,
  consumed_ip           INET        DEFAULT NULL,
  consumed_user_agent   TEXT        DEFAULT NULL,
  triggered_by_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL, -- admin reset
  created_at            TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_pwd_recovery_user_id   ON password_recovery_tokens(user_id);
CREATE INDEX idx_pwd_recovery_expires   ON password_recovery_tokens(expires_at);
CREATE INDEX idx_pwd_recovery_active    ON password_recovery_tokens(user_id)
  WHERE used_at IS NULL AND revoked_at IS NULL;

-- Tokens de verificação de e-mail
CREATE TABLE email_verification_tokens (
  id                    UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash            TEXT        NOT NULL UNIQUE,
  email                 TEXT        NOT NULL,
  expires_at            TIMESTAMPTZ NOT NULL,
  used_at               TIMESTAMPTZ DEFAULT NULL,
  revoked_at            TIMESTAMPTZ DEFAULT NULL,
  request_ip            INET        DEFAULT NULL,
  request_user_agent    TEXT        DEFAULT NULL,
  consumed_ip           INET        DEFAULT NULL,
  consumed_user_agent   TEXT        DEFAULT NULL,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_email_verif_user_id  ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verif_expires  ON email_verification_tokens(expires_at);
CREATE INDEX idx_email_verif_active   ON email_verification_tokens(user_id)
  WHERE used_at IS NULL AND revoked_at IS NULL;

-- Preferências de UI do usuário
CREATE TABLE user_settings (
  user_id                 UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  notifications_enabled   BOOLEAN     NOT NULL DEFAULT TRUE,
  theme                   VARCHAR(20) NOT NULL DEFAULT 'light',
  language                VARCHAR(10) NOT NULL DEFAULT 'pt-BR',
  timezone                VARCHAR(50) NOT NULL DEFAULT 'America/Sao_Paulo',
  created_at              TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at              TIMESTAMPTZ
);


-- =============================================================================
-- SEÇÃO 3: PERMISSÕES E CONTROLE DE ACESSO
-- =============================================================================

-- Permissões granulares do sistema (ex: schedule.create, client.delete)
CREATE TABLE permissions (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  code        VARCHAR(50) UNIQUE NOT NULL,  -- ex: 'schedule.create'
  name        VARCHAR(100) NOT NULL,
  description VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_permissions_code ON permissions(code);

-- Permissões atribuídas a cada usuário (escopadas por empresa em companies_users)
CREATE TABLE users_permissions (
  id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission_id UUID        NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  granted_by    UUID        REFERENCES users(id) ON DELETE SET NULL,
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  granted_at    TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  revoked_at    TIMESTAMPTZ DEFAULT NULL,
  UNIQUE (user_id, permission_id)
);

CREATE INDEX idx_users_permissions_user_id ON users_permissions(user_id);


-- =============================================================================
-- SEÇÃO 4: MÓDULOS E PLANOS (CONTROLE DE FEATURES)
-- =============================================================================

-- Módulos/funcionalidades da plataforma
CREATE TABLE modules (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  code            VARCHAR(10)   UNIQUE NOT NULL,  -- ex: 'SCH', 'FIN', 'CRM'
  name            VARCHAR(100)  NOT NULL,
  description     VARCHAR(255)  NOT NULL,
  min_package     module_package NOT NULL,  -- pacote mínimo para ter acesso
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_modules_code ON modules(code);

-- Tipos de planos (ex: Mensal, Anual, Trial)
CREATE TABLE plan_types (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        VARCHAR(50) NOT NULL,
  description VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  deleted_at  TIMESTAMPTZ
);

-- Planos de assinatura
CREATE TABLE plans (
  id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_type_id    UUID            NOT NULL REFERENCES plan_types(id),
  name            VARCHAR(100)    NOT NULL,
  description     VARCHAR(255)    NOT NULL,
  package         module_package  NOT NULL,
  price           NUMERIC(12,2)   NOT NULL DEFAULT 0.00,  -- NUNCA usar REAL para dinheiro
  billing_cycle_days  INTEGER     NOT NULL DEFAULT 30,
  max_users       INTEGER         DEFAULT NULL,  -- NULL = ilimitado
  is_active       BOOLEAN         NOT NULL DEFAULT TRUE,
  image_url       VARCHAR(500),
  created_at      TIMESTAMPTZ     NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Módulos incluídos em cada plano
CREATE TABLE plan_modules (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_id     UUID        NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  module_id   UUID        NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (plan_id, module_id)
);

-- Permissões incluídas em cada plano
CREATE TABLE plan_permissions (
  id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_id       UUID        NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  permission_id UUID        NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (plan_id, permission_id)
);


-- =============================================================================
-- SEÇÃO 5: EMPRESAS (MULTI-TENANT CORE)
-- =============================================================================

-- Idiomas suportados
CREATE TABLE languages (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  code        VARCHAR(5)  UNIQUE NOT NULL,  -- ex: 'pt-BR', 'en-US'
  name        VARCHAR(50) NOT NULL,
  native_name VARCHAR(50) NOT NULL,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

-- Empresas clientes da plataforma
CREATE TABLE companies (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug            VARCHAR(100)  UNIQUE NOT NULL,  -- identificador URL-friendly
  name            VARCHAR(255)  NOT NULL,
  fantasy_name    VARCHAR(255)  NOT NULL,
  cnpj            VARCHAR(14)   UNIQUE NOT NULL,
  foundation_date DATE,
  logo_url        VARCHAR(500),
  responsible_id  UUID          NOT NULL,  -- FK adicionada após criação de people
  active_package  module_package NOT NULL DEFAULT 'starter',
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_companies_slug    ON companies(slug);
CREATE INDEX idx_companies_cnpj    ON companies(cnpj);
CREATE INDEX idx_companies_active  ON companies(is_active) WHERE deleted_at IS NULL;

-- Assinaturas de planos por empresa
CREATE TABLE company_subscriptions (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  plan_id         UUID          NOT NULL REFERENCES plans(id),
  started_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  expires_at      TIMESTAMPTZ   NOT NULL,
  canceled_at     TIMESTAMPTZ   DEFAULT NULL,
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  price_paid      NUMERIC(12,2) NOT NULL,  -- preço no momento da contratação
  notes           VARCHAR(255),
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_company_subscriptions_company  ON company_subscriptions(company_id);
CREATE INDEX idx_company_subscriptions_active   ON company_subscriptions(company_id)
  WHERE is_active = TRUE AND canceled_at IS NULL;

-- Módulos ativos por empresa (derivado do plano, mas pode ter exceções manuais)
CREATE TABLE company_modules (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  module_id       UUID          NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  subscription_id UUID          REFERENCES company_subscriptions(id) ON DELETE SET NULL,
  granted_manually BOOLEAN      NOT NULL DEFAULT FALSE,  -- true = exceção manual
  starts_at       TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  expires_at      TIMESTAMPTZ,
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  UNIQUE (company_id, module_id)
);

CREATE INDEX idx_company_modules_company ON company_modules(company_id);

-- Vínculo entre usuários e empresas
CREATE TABLE company_users (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_owner    BOOLEAN     NOT NULL DEFAULT FALSE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  joined_at   TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  left_at     TIMESTAMPTZ DEFAULT NULL,
  UNIQUE (company_id, user_id)
);

CREATE INDEX idx_company_users_company  ON company_users(company_id);
CREATE INDEX idx_company_users_user     ON company_users(user_id);

-- Configurações de sistema por empresa
CREATE TABLE company_system_configs (
  id                          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id                  UUID        UNIQUE NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  schedule_init_time          TIME        NOT NULL DEFAULT '08:00',
  schedule_pause_init_time    TIME        DEFAULT NULL,
  schedule_pause_end_time     TIME        DEFAULT NULL,
  schedule_end_time           TIME        NOT NULL DEFAULT '18:00',
  min_schedules_per_day       SMALLINT    NOT NULL DEFAULT 4,
  max_schedules_per_day       SMALLINT    NOT NULL DEFAULT 6,
  schedule_days               week_day[]  NOT NULL DEFAULT ARRAY['monday','tuesday','wednesday','thursday','friday','saturday']::week_day[],
  dynamic_cages               BOOLEAN     NOT NULL DEFAULT TRUE,
  total_small_cages           SMALLINT    NOT NULL DEFAULT 0,
  total_medium_cages          SMALLINT    NOT NULL DEFAULT 0,
  total_large_cages           SMALLINT    NOT NULL DEFAULT 0,
  total_giant_cages           SMALLINT    NOT NULL DEFAULT 0,
  whatsapp_notifications      BOOLEAN     NOT NULL DEFAULT FALSE,
  whatsapp_conversation       BOOLEAN     NOT NULL DEFAULT FALSE,
  whatsapp_business_phone     VARCHAR(25) DEFAULT NULL,
  created_at                  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at                  TIMESTAMPTZ
);


-- =============================================================================
-- SEÇÃO 6: PESSOAS (DADOS PESSOAIS)
-- =============================================================================

-- Entidade base de pessoa (polimórfica: clientes, funcionários, etc.)
CREATE TABLE people (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  kind            person_kind NOT NULL,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  has_system_user BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_people_kind     ON people(kind);
CREATE INDEX idx_people_active   ON people(is_active);

-- Identificação pessoal
CREATE TABLE people_identifications (
  id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id       UUID            UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  full_name       VARCHAR(150)    NOT NULL,
  short_name      VARCHAR(75)     NOT NULL,
  gender_identity gender_identity NOT NULL,
  marital_status  marital_status  NOT NULL,
  image_url       VARCHAR(500),
  birth_date      DATE            NOT NULL,  -- DATE ao invés de TIMESTAMPTZ para nascimento
  cpf             VARCHAR(11)     UNIQUE NOT NULL,
  created_at      TIMESTAMPTZ     NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_people_ident_cpf ON people_identifications(cpf);

-- Documentos específicos de funcionários CLT/contratados
CREATE TABLE employee_documents (
  id                      UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id               UUID              UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  rg                      VARCHAR(15)       UNIQUE NOT NULL,
  issuing_body            VARCHAR(20)       NOT NULL,
  issuing_date            DATE              NOT NULL,
  ctps                    VARCHAR(15)       UNIQUE NOT NULL,
  ctps_series             VARCHAR(10)       NOT NULL,
  ctps_state              CHAR(2)           NOT NULL,
  pis                     VARCHAR(11)       UNIQUE NOT NULL,
  voter_registration      VARCHAR(12)       UNIQUE,
  vote_zone               VARCHAR(5),
  vote_section            VARCHAR(5),
  military_certificate    VARCHAR(12)       UNIQUE,
  military_series         VARCHAR(6),
  military_category       CHAR(2),
  has_special_needs       BOOLEAN           DEFAULT FALSE,
  has_children            BOOLEAN           NOT NULL DEFAULT FALSE,
  children_qty            SMALLINT          DEFAULT 0,
  has_children_under_18   BOOLEAN           DEFAULT FALSE,
  has_family_special_needs BOOLEAN          DEFAULT FALSE,
  graduation              graduation_level  NOT NULL,
  has_cnh                 BOOLEAN           DEFAULT FALSE,
  cnh_type                VARCHAR(3),
  cnh_number              VARCHAR(15),
  cnh_expiration_date     DATE,
  created_at              TIMESTAMPTZ       NOT NULL DEFAULT current_timestamp,
  updated_at              TIMESTAMPTZ
);

-- Contatos de pessoas
CREATE TABLE people_contacts (
  id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id         UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  email             VARCHAR(150) NOT NULL,
  phone             VARCHAR(25),
  cellphone         VARCHAR(25) NOT NULL,
  has_whatsapp      BOOLEAN     NOT NULL DEFAULT FALSE,
  instagram_user    VARCHAR(100),
  emergency_contact VARCHAR(150),
  emergency_phone   VARCHAR(25),
  is_primary        BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at        TIMESTAMPTZ
);

CREATE INDEX idx_people_contacts_person ON people_contacts(person_id);

-- Endereços
CREATE TABLE addresses (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  zip_code    VARCHAR(9)  NOT NULL,
  street      VARCHAR(255) NOT NULL,
  number      VARCHAR(15) NOT NULL,
  complement  VARCHAR(100),
  district    VARCHAR(100) NOT NULL,
  city        VARCHAR(100) NOT NULL,
  state       CHAR(2)     NOT NULL,  -- sigla UF
  country     VARCHAR(100) NOT NULL DEFAULT 'Brasil',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ
);

-- Endereços de pessoas (pessoa pode ter múltiplos)
CREATE TABLE people_addresses (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id   UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  address_id  UUID        NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main     BOOLEAN     NOT NULL DEFAULT FALSE,
  label       VARCHAR(50) DEFAULT NULL, -- ex: 'Casa', 'Trabalho'
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (person_id, address_id)
);

CREATE INDEX idx_people_addresses_person ON people_addresses(person_id);

-- Dados financeiros/bancários de pessoas
CREATE TABLE finances (
  id                    UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
  bank_name             VARCHAR(100)      NOT NULL,
  bank_code             VARCHAR(10),      -- código do banco (ex: '001' = BB)
  bank_branch           VARCHAR(10)       NOT NULL,
  bank_account          VARCHAR(15)       NOT NULL,
  bank_account_digit    VARCHAR(2)        NOT NULL,
  bank_account_type     bank_account_kind NOT NULL,
  has_pix               BOOLEAN           NOT NULL DEFAULT FALSE,
  pix_key               VARCHAR(255),
  pix_key_type          pix_key_kind,
  created_at            TIMESTAMPTZ       NOT NULL DEFAULT current_timestamp,
  updated_at            TIMESTAMPTZ
);

-- Vínculo financeiro entre pessoa e seus dados bancários
CREATE TABLE people_finances (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id   UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  finance_id  UUID        NOT NULL REFERENCES finances(id) ON DELETE CASCADE,
  is_primary  BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (person_id, finance_id)
);

-- Vínculo entre users e people (usuário do sistema ↔ pessoa real)
CREATE TABLE user_profiles (
  user_id     UUID  PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  person_id   UUID  UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);


-- =============================================================================
-- SEÇÃO 7: EMPRESAS ↔ PESSOAS
-- =============================================================================

-- Pessoas vinculadas a empresas (funcionários, responsáveis, etc.)
CREATE TABLE company_people (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id   UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  UNIQUE (company_id, person_id)
);

CREATE INDEX idx_company_people_company ON company_people(company_id);

-- Funcionários de empresa
CREATE TABLE company_employees (
  id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id   UUID          NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  kind        employee_kind NOT NULL DEFAULT 'internal',
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  deleted_at  TIMESTAMPTZ,
  UNIQUE (company_id, person_id)
);

CREATE INDEX idx_company_employees_company ON company_employees(company_id);

-- Vínculo empregatício do funcionário
CREATE TABLE employments (
  id                  UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id UUID          NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  role                VARCHAR(100)  NOT NULL,
  admission_date      DATE          NOT NULL,
  resignation_date    DATE,
  salary              NUMERIC(12,2) NOT NULL,  -- NUMERIC para dinheiro
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at          TIMESTAMPTZ,
  deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_employments_employee ON employments(company_employee_id);

-- Benefícios dos funcionários (com histórico via valid_from/valid_until)
CREATE TABLE employee_benefits (
  id                        UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id       UUID          NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  meal_ticket               BOOLEAN       NOT NULL DEFAULT FALSE,
  meal_ticket_value         NUMERIC(10,2) NOT NULL DEFAULT 0.00,
  transport_voucher         BOOLEAN       NOT NULL DEFAULT FALSE,
  transport_voucher_qty     SMALLINT      NOT NULL DEFAULT 0,
  transport_voucher_value   NUMERIC(10,2) NOT NULL DEFAULT 0.00,
  valid_from                DATE          NOT NULL DEFAULT CURRENT_DATE,
  valid_until               DATE,
  created_at                TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at                TIMESTAMPTZ
);

CREATE INDEX idx_employee_benefits_employee ON employee_benefits(company_employee_id);

-- Custo total de funcionários para a empresa
CREATE TABLE company_employee_costs (
  id                  UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id UUID          NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  costs               NUMERIC(12,2) NOT NULL DEFAULT 0.00,
  reference_month     DATE          NOT NULL,  -- primeiro dia do mês de referência
  comments            VARCHAR(255),
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at          TIMESTAMPTZ
);

-- Endereços de pessoas no contexto de uma empresa
CREATE TABLE company_people_addresses (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id   UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  address_id  UUID        NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main     BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (company_id, person_id, address_id)
);

-- Dados bancários de empresas
CREATE TABLE company_finances (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  finance_id  UUID        NOT NULL REFERENCES finances(id) ON DELETE CASCADE,
  is_primary  BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (company_id, finance_id)
);

-- Endereços das empresas
CREATE TABLE company_addresses (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  address_id  UUID        NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main     BOOLEAN     NOT NULL DEFAULT FALSE,
  label       VARCHAR(50),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (company_id, address_id)
);

-- FK de responsible_id adicionada depois de people existir
ALTER TABLE companies
  ADD CONSTRAINT fk_companies_responsible
  FOREIGN KEY (responsible_id) REFERENCES people(id) ON DELETE RESTRICT;


-- =============================================================================
-- SEÇÃO 8: CLIENTES E PETS
-- =============================================================================

-- Clientes
CREATE TABLE clients (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id       UUID        UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  client_since    DATE,
  recommended_by  UUID        REFERENCES clients(id) ON DELETE SET NULL,
  notes           VARCHAR(255),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Vínculo entre empresa e seus clientes
CREATE TABLE company_clients (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  client_id   UUID        NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  joined_at   TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  left_at     TIMESTAMPTZ DEFAULT NULL,
  UNIQUE (company_id, client_id)
);

CREATE INDEX idx_company_clients_company ON company_clients(company_id);

-- Pets
CREATE TABLE pets (
  id            UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
  name          VARCHAR(100)    NOT NULL,
  size          pet_size        NOT NULL,
  kind          pet_kind        NOT NULL,
  temperament   pet_temperament NOT NULL,
  image_url     VARCHAR(500),
  birth_date    DATE,
  owner_id      UUID            NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  guardian_id   UUID            REFERENCES people(id) ON DELETE SET NULL,
  is_active     BOOLEAN         NOT NULL DEFAULT TRUE,
  notes         VARCHAR(500),
  created_at    TIMESTAMPTZ     NOT NULL DEFAULT current_timestamp,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_pets_owner   ON pets(owner_id);
CREATE INDEX idx_pets_active  ON pets(is_active) WHERE deleted_at IS NULL;

-- Planos contratados por clientes
CREATE TABLE client_plans (
  id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  client_id   UUID          NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  plan_id     UUID          NOT NULL REFERENCES plans(id),
  started_at  TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  expires_at  TIMESTAMPTZ   NOT NULL,
  price_paid  NUMERIC(12,2) NOT NULL,
  is_active   BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  UNIQUE (client_id, plan_id)
);

CREATE INDEX idx_client_plans_client ON client_plans(client_id);


-- =============================================================================
-- SEÇÃO 9: PRODUTOS E SERVIÇOS
-- =============================================================================

-- Tipos de serviço
CREATE TABLE service_types (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        VARCHAR(100) NOT NULL,
  description VARCHAR(255),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  deleted_at  TIMESTAMPTZ
);

-- Serviços oferecidos
CREATE TABLE services (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  type_id         UUID          NOT NULL REFERENCES service_types(id),
  title           VARCHAR(100)  NOT NULL,
  description     VARCHAR(500)  NOT NULL,
  notes           VARCHAR(255),
  price           NUMERIC(12,2) NOT NULL,
  discount_rate   NUMERIC(5,2)  NOT NULL DEFAULT 0.00,  -- percentual
  image_url       VARCHAR(500),
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Sub-serviços (variações de um serviço, ex: banho pequeno, banho grande)
CREATE TABLE sub_services (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_id      UUID          NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  type_id         UUID          REFERENCES service_types(id),
  title           VARCHAR(100)  NOT NULL,
  description     VARCHAR(500)  NOT NULL,
  notes           VARCHAR(255),
  price           NUMERIC(12,2) NOT NULL,
  discount_rate   NUMERIC(5,2)  NOT NULL DEFAULT 0.00,
  image_url       VARCHAR(500),
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_sub_services_service ON sub_services(service_id);

-- Serviços disponíveis por empresa
CREATE TABLE company_services (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  service_id  UUID        NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  UNIQUE (company_id, service_id)
);

CREATE INDEX idx_company_services_company ON company_services(company_id);

-- Produtos (estoque/venda)
CREATE TABLE products (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            VARCHAR(100)  NOT NULL,
  batch_number    VARCHAR(50),
  description     VARCHAR(500),
  image_url       VARCHAR(500),
  expiration_date DATE,
  quantity        INTEGER       NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Produtos por empresa
CREATE TABLE company_products (
  id                    UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id            UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  product_id            UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  kind                  product_kind  NOT NULL,
  has_stock             BOOLEAN       NOT NULL DEFAULT TRUE,
  for_sale              BOOLEAN       NOT NULL DEFAULT FALSE,
  cost_per_unit         NUMERIC(12,2) NOT NULL,  -- custo de aquisição
  profit_margin         NUMERIC(5,2)  NOT NULL DEFAULT 0.00,
  sale_price            NUMERIC(12,2) NOT NULL,  -- preço final calculado ou manual
  created_at            TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at            TIMESTAMPTZ,
  deleted_at            TIMESTAMPTZ,
  UNIQUE (company_id, product_id)
);

CREATE INDEX idx_company_products_company ON company_products(company_id);

-- Custos operacionais da empresa (contas, notas fiscais, etc.)
CREATE TABLE company_business_costs (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  invoice_number  VARCHAR(50),
  invoice_url     VARCHAR(500),
  description     VARCHAR(255)  NOT NULL,
  total_cost      NUMERIC(12,2) NOT NULL,
  reference_month DATE          NOT NULL,
  comments        VARCHAR(255),
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_company_business_costs_company ON company_business_costs(company_id);


-- =============================================================================
-- SEÇÃO 10: PLANOS DE SERVIÇO PARA CLIENTES
-- =============================================================================

-- Pacotes de serviços oferecidos como planos para clientes
CREATE TABLE service_plans (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_type_id    UUID          NOT NULL REFERENCES plan_types(id),
  title           VARCHAR(100)  NOT NULL,
  description     VARCHAR(500)  NOT NULL,
  notes           VARCHAR(255),
  price           NUMERIC(12,2) NOT NULL,
  discount_rate   NUMERIC(5,2)  NOT NULL DEFAULT 0.00,
  image_url       VARCHAR(500),
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Serviços incluídos em cada plano de serviço
CREATE TABLE service_plan_services (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id UUID        NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  service_id      UUID        NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (service_plan_id, service_id)
);

-- Sub-serviços incluídos em cada plano de serviço
CREATE TABLE service_plan_sub_services (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id UUID        NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  sub_service_id  UUID        NOT NULL REFERENCES sub_services(id) ON DELETE CASCADE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (service_plan_id, sub_service_id)
);

-- Serviços bônus em planos (benefícios extras)
CREATE TABLE service_plan_bonuses (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id UUID        NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  service_id      UUID        REFERENCES services(id) ON DELETE CASCADE,
  sub_service_id  UUID        REFERENCES sub_services(id) ON DELETE CASCADE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  CONSTRAINT chk_bonus_has_one CHECK (
    (service_id IS NOT NULL OR sub_service_id IS NOT NULL)
  )
);

-- Planos de serviço por empresa
CREATE TABLE company_service_plans (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  service_plan_id UUID        NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  UNIQUE (company_id, service_plan_id)
);

-- Planos de serviço contratados por clientes
CREATE TABLE client_service_plans (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  client_id       UUID          NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  service_plan_id UUID          NOT NULL REFERENCES service_plans(id),
  started_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  expires_at      TIMESTAMPTZ   NOT NULL,
  price_paid      NUMERIC(12,2) NOT NULL,
  is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ,
  UNIQUE (client_id, service_plan_id)
);


-- =============================================================================
-- SEÇÃO 11: AGENDAMENTOS
-- =============================================================================

-- Agendamentos
-- Unificando date + hour em scheduled_at (TIMESTAMPTZ)
CREATE TABLE schedules (
  id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id    UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  client_id     UUID          NOT NULL REFERENCES clients(id),
  pet_id        UUID          NOT NULL REFERENCES pets(id),
  scheduled_at  TIMESTAMPTZ   NOT NULL,  -- data e hora unificados
  estimated_end TIMESTAMPTZ,
  notes         VARCHAR(500),
  created_by    UUID          REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ,
  UNIQUE (company_id, client_id, pet_id, scheduled_at)
);

CREATE INDEX idx_schedules_company     ON schedules(company_id);
CREATE INDEX idx_schedules_client      ON schedules(client_id);
CREATE INDEX idx_schedules_date        ON schedules(scheduled_at);

-- Histórico de status dos agendamentos
CREATE TABLE schedule_status_history (
  id            UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id   UUID            NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  status        schedule_status NOT NULL DEFAULT 'waiting',
  changed_at    TIMESTAMPTZ     NOT NULL DEFAULT current_timestamp,
  changed_by    UUID            REFERENCES users(id) ON DELETE SET NULL,
  notes         VARCHAR(255)
);

CREATE INDEX idx_schedule_status_history_schedule ON schedule_status_history(schedule_id);

-- Serviços do agendamento
CREATE TABLE schedule_services (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id UUID        NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  service_id  UUID        NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (schedule_id, service_id)
);

-- Sub-serviços do agendamento
CREATE TABLE schedule_sub_services (
  id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id         UUID        NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  service_id          UUID        NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  sub_service_id      UUID        NOT NULL REFERENCES sub_services(id) ON DELETE CASCADE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (schedule_id, sub_service_id)
);

-- Check-in/out de agendamentos
CREATE TABLE schedule_checkins (
  id                        UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id               UUID        UNIQUE NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  checked_in_at             TIMESTAMPTZ,
  check_in_person_id        UUID        REFERENCES people(id) ON DELETE SET NULL,
  check_in_notes            VARCHAR(500),
  check_in_photo_url        VARCHAR(500),
  checked_out_at            TIMESTAMPTZ,
  check_out_person_id       UUID        REFERENCES people(id) ON DELETE SET NULL,
  check_out_notes           VARCHAR(500),
  check_out_photo_url       VARCHAR(500),
  service_executor_person_id UUID       REFERENCES people(id) ON DELETE SET NULL,
  created_at                TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at                TIMESTAMPTZ
);

-- Pagamentos de agendamentos
CREATE TABLE schedule_payments (
  id                    UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id           UUID            UNIQUE NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  payment_method_one    payment_method  NOT NULL,
  payment_method_two    payment_method,
  payment_method_three  payment_method,
  payment_date          TIMESTAMPTZ     NOT NULL,
  gross_value           NUMERIC(12,2)   NOT NULL,
  discount_value        NUMERIC(12,2)   NOT NULL DEFAULT 0.00,
  net_value             NUMERIC(12,2)   NOT NULL,
  amount_paid           NUMERIC(12,2)   NOT NULL DEFAULT 0.00,
  amount_remaining      NUMERIC(12,2)   GENERATED ALWAYS AS (net_value - amount_paid) STORED,
  notes                 VARCHAR(255),
  created_at            TIMESTAMPTZ     NOT NULL DEFAULT current_timestamp,
  updated_at            TIMESTAMPTZ,
  CONSTRAINT chk_payment_values CHECK (discount_value >= 0 AND net_value >= 0 AND amount_paid >= 0)
);

-- Ordens de serviço (impressão/comprovante)
CREATE TABLE service_orders (
  id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id     UUID          UNIQUE NOT NULL REFERENCES schedules(id),
  printed_by      UUID          REFERENCES users(id) ON DELETE SET NULL,
  printed_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  gross_total     NUMERIC(12,2) NOT NULL,
  amount_paid     NUMERIC(12,2) NOT NULL,
  amount_to_pay   NUMERIC(12,2) NOT NULL,
  has_perfume     BOOLEAN       NOT NULL DEFAULT FALSE,
  has_ornament    BOOLEAN       NOT NULL DEFAULT FALSE,
  notes           VARCHAR(255),
  created_at      TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp,
  updated_at      TIMESTAMPTZ
);


-- =============================================================================
-- SEÇÃO 12: NOTIFICAÇÕES E AVISOS
-- =============================================================================

-- Notificações do sistema
CREATE TABLE notifications (
  id                UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id        UUID        REFERENCES companies(id) ON DELETE CASCADE,  -- NULL = global
  title             VARCHAR(100) NOT NULL,
  summary           VARCHAR(200) NOT NULL,
  content           TEXT        NOT NULL,
  send_to_whatsapp  BOOLEAN     NOT NULL DEFAULT FALSE,
  created_by        UUID        REFERENCES users(id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at        TIMESTAMPTZ,
  deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_notifications_company ON notifications(company_id);

-- Destinatários de notificações
CREATE TABLE notification_receivers (
  id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  notification_id UUID        NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  read_at         TIMESTAMPTZ DEFAULT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  UNIQUE (notification_id, user_id)
);

CREATE INDEX idx_notification_receivers_user ON notification_receivers(user_id);

-- Avisos internos da empresa
CREATE TABLE warnings (
  id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  title       VARCHAR(100) NOT NULL,
  content     TEXT        NOT NULL,
  image_url   VARCHAR(500),
  sender_id   UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at  TIMESTAMPTZ,
  deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_warnings_company ON warnings(company_id);


-- =============================================================================
-- SEÇÃO 13: AUDITORIA (IMUTÁVEL)
-- =============================================================================
-- Tabelas de auditoria NUNCA devem ter ON DELETE CASCADE em referências a users.
-- Registros de auditoria são imutáveis — não se deletam, não se atualizam.

-- Log geral de ações no sistema
CREATE TABLE audit_logs (
  id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  action        log_action  NOT NULL,
  entity_table  VARCHAR(100) NOT NULL,
  entity_id     UUID        NOT NULL,
  company_id    UUID        REFERENCES companies(id) ON DELETE SET NULL,
  old_data      JSONB,                        -- JSONB permite queries nos dados históricos
  new_data      JSONB       NOT NULL,
  changed_at    TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  changed_by    UUID        REFERENCES users(id) ON DELETE SET NULL,  -- SET NULL, nunca CASCADE
  ip_address    INET,
  user_agent    TEXT
);

CREATE INDEX idx_audit_logs_entity        ON audit_logs(entity_table, entity_id);
CREATE INDEX idx_audit_logs_company       ON audit_logs(company_id);
CREATE INDEX idx_audit_logs_changed_by    ON audit_logs(changed_by);
CREATE INDEX idx_audit_logs_changed_at    ON audit_logs(changed_at DESC);
CREATE INDEX idx_audit_logs_action        ON audit_logs(action);

-- Log específico de autenticação (separado por volume e sensibilidade)
CREATE TABLE auth_logs (
  id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID          REFERENCES users(id) ON DELETE SET NULL,  -- SET NULL, nunca CASCADE
  action        log_action    NOT NULL,  -- login, logout, password_changed, etc.
  company_id    UUID          REFERENCES companies(id) ON DELETE SET NULL,
  session_id    UUID          REFERENCES user_sessions(id) ON DELETE SET NULL,
  ip_address    INET          NOT NULL,
  user_agent    TEXT          NOT NULL,
  result        login_result,
  detail        VARCHAR(255),
  occurred_at   TIMESTAMPTZ   NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_auth_logs_user_id     ON auth_logs(user_id);
CREATE INDEX idx_auth_logs_company_id  ON auth_logs(company_id);
CREATE INDEX idx_auth_logs_occurred_at ON auth_logs(occurred_at DESC);


-- =============================================================================
-- SEÇÃO 14: INTERNACIONALIZAÇÃO
-- =============================================================================

-- Traduções de entidades do sistema
CREATE TABLE translations (
  id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
  language_code VARCHAR(5)  NOT NULL REFERENCES languages(code) ON DELETE CASCADE,
  entity_table  VARCHAR(100) NOT NULL,
  entity_id     UUID        NOT NULL,
  field         VARCHAR(100) NOT NULL,
  content       TEXT        NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  updated_at    TIMESTAMPTZ,
  UNIQUE (language_code, entity_table, entity_id, field)
);

CREATE INDEX idx_translations_entity ON translations(entity_table, entity_id);
