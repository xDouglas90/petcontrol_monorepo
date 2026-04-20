-- =============================================================================
-- PETCONTROL - SCHEMA PostgreSQL v2
-- Aplicação Multi-Tenant SaaS com controle de planos e módulos
-- =============================================================================
-- Extensão necessária para UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- SEÇÃO 1: ENUMS
-- =============================================================================
-- Tipos de usuário no sistema(papel técnico/sistêmico)
CREATE TYPE user_role_type AS ENUM(
  'root', -- Acesso total ao sistema
  'internal', -- Acesso total ao sistema(Desenvolvedores, criados por um root)
  'admin', -- Gerencia tenants, usuários, configurações, permissões(Criado por um root)
  'system', -- Usuário do sistema(Criado por um admin)
  'common', -- Usuário comum(Criados por clientes dos tenants)
  'free' -- Conta gratuita / trial(Criado por um root)
);

-- Tipos de usuário no contexto de negócio(quem é na empresa)
CREATE TYPE user_kind AS ENUM(
  'owner', -- Proprietário/Sócio da empresa
  'employee', -- Funcionário da empresa cliente
  'client', -- Cliente final atendido pela empresa
  'supplier', -- Fornecedor da empresa cliente
  'outsourced_employee' -- Funcionário terceirizado(ex: free lancers, prestadores de serviço)
);

-- Pacotes de módulos disponíveis nos planos
CREATE TYPE module_package AS ENUM(
  'internal',
  'starter',
  'basic',
  'essential',
  'premium',
  'trial'
);

-- Ações possíveis registradas em logs de auditoria
CREATE TYPE log_action AS ENUM(
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
CREATE TYPE gender_identity AS ENUM(
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
CREATE TYPE marital_status AS ENUM(
  'single',
  'married',
  'divorced',
  'widowed',
  'separated'
);

-- Porte dos pets
CREATE TYPE pet_size AS ENUM(
  'small',
  'medium',
  'large',
  'giant'
);

-- Temperamento dos pets
CREATE TYPE pet_temperament AS ENUM(
  'calm',
  'nervous',
  'aggressive',
  'playful',
  'loving'
);

-- Tipos de pets
CREATE TYPE pet_kind AS ENUM(
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
CREATE TYPE employee_kind AS ENUM(
  'internal',
  'outsourced'
);

-- Tipos de chave Pix
CREATE TYPE pix_key_kind AS ENUM(
  'cpf',
  'cnpj',
  'email',
  'phone',
  'random'
);

-- Dias da semana
CREATE TYPE week_day AS ENUM(
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday'
);

-- Níveis de escolaridade
CREATE TYPE graduation_level AS ENUM(
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
CREATE TYPE person_kind AS ENUM(
  'client',
  'employee',
  'outsourced_employee',
  'supplier',
  'guardian',
  'responsible'
);

-- Tipos de conta bancária
CREATE TYPE bank_account_kind AS ENUM(
  'checking',
  'savings',
  'salary'
);

-- Métodos de pagamento
CREATE TYPE payment_method AS ENUM(
  'credit_card',
  'debit_card',
  'pix',
  'cash',
  'check',
  'bank_slip',
  'transfer'
);

-- Status dos agendamentos
CREATE TYPE schedule_status AS ENUM(
  'waiting',
  'confirmed',
  'canceled',
  'in_progress',
  'finished',
  'delivered'
);

-- Tipos de produto
CREATE TYPE product_kind AS ENUM(
  'service',
  'customer',
  'internal_usage'
);

-- Resultado de tentativa de login
CREATE TYPE login_result AS ENUM(
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
CREATE TYPE logout_reason AS ENUM(
  'user_initiated',
  'session_expired',
  'forced_by_admin',
  'password_changed',
  'account_deactivated',
  'suspicious_activity'
);

CREATE TYPE notification_level AS ENUM(
  'info',
  'success',
  'warning',
  'error',
  'alert'
);

-- =============================================================================
-- SEÇÃO 2: AUTENTICAÇÃO E SESSÕES
-- =============================================================================
-- Usuários do sistema(somente dados de acesso)
CREATE TABLE users(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email varchar(150) UNIQUE NOT NULL,
  email_verified boolean NOT NULL DEFAULT FALSE,
  email_verified_at timestamptz DEFAULT NULL,
  "role" user_role_type NOT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz -- soft delete; nunca remover fisicamente
);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_users_active ON users(is_active)
WHERE
  deleted_at IS NULL;

CREATE INDEX idx_users_role ON users(ROLE);

-- Credenciais separadas dos dados do usuário
CREATE TABLE user_auth(
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash text NOT NULL,
  password_changed_at timestamptz DEFAULT NULL,
  must_change_password boolean NOT NULL DEFAULT FALSE,
  login_attempts smallint NOT NULL DEFAULT 0,
  locked_until timestamptz DEFAULT NULL,
  last_login_at timestamptz DEFAULT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Histórico de cada tentativa de login(imutável — sem ON DELETE CASCADE)
CREATE TABLE login_history(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid REFERENCES users(id) ON DELETE SET NULL, -- preserva histórico
  ip_address inet NOT NULL, -- usar INET ao invés de VARCHAR para IPs
  user_agent text NOT NULL,
  result login_result NOT NULL,
  failure_detail varchar(200) DEFAULT NULL,
  attempted_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_login_history_user_id ON login_history(user_id);

CREATE INDEX idx_login_history_attempted ON login_history(attempted_at DESC);

CREATE INDEX idx_login_history_result ON login_history(result);

-- Sessões ativas e históricas(separado do histórico de login)
CREATE TABLE user_sessions(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  login_history_id uuid REFERENCES login_history(id) ON DELETE SET NULL,
  session_token text UNIQUE NOT NULL,
  ip_address inet NOT NULL,
  user_agent text NOT NULL,
  last_activity_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamptz NOT NULL,
  logged_out_at timestamptz DEFAULT NULL,
  logout_reason logout_reason DEFAULT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);

CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);

CREATE INDEX idx_user_sessions_active ON user_sessions(expires_at)
WHERE
  logged_out_at IS NULL;

-- Tokens de recuperação de senha
CREATE TABLE password_recovery_tokens(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  requested_email text NOT NULL,
  expires_at timestamptz NOT NULL,
  used_at timestamptz DEFAULT NULL,
  revoked_at timestamptz DEFAULT NULL,
  request_ip inet DEFAULT NULL,
  request_user_agent text DEFAULT NULL,
  consumed_ip inet DEFAULT NULL,
  consumed_user_agent text DEFAULT NULL,
  triggered_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL, -- admin reset
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pwd_recovery_user_id ON password_recovery_tokens(user_id);

CREATE INDEX idx_pwd_recovery_expires ON password_recovery_tokens(expires_at);

CREATE INDEX idx_pwd_recovery_active ON password_recovery_tokens(user_id)
WHERE
  used_at IS NULL AND revoked_at IS NULL;

-- Tokens de verificação de e-mail
CREATE TABLE email_verification_tokens(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  email text NOT NULL,
  expires_at timestamptz NOT NULL,
  used_at timestamptz DEFAULT NULL,
  revoked_at timestamptz DEFAULT NULL,
  request_ip inet DEFAULT NULL,
  request_user_agent text DEFAULT NULL,
  consumed_ip inet DEFAULT NULL,
  consumed_user_agent text DEFAULT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_verif_user_id ON email_verification_tokens(user_id);

CREATE INDEX idx_email_verif_expires ON email_verification_tokens(expires_at);

CREATE INDEX idx_email_verif_active ON email_verification_tokens(user_id)
WHERE
  used_at IS NULL AND revoked_at IS NULL;

-- Preferências de UI do usuário
CREATE TABLE user_settings(
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  notifications_enabled boolean NOT NULL DEFAULT TRUE,
  theme varchar(20) NOT NULL DEFAULT 'light',
  language varchar
(10) NOT NULL DEFAULT 'pt-BR',
  timezone varchar(50) NOT NULL DEFAULT 'America/Sao_Paulo',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- =============================================================================
-- SEÇÃO 3: PERMISSÕES E CONTROLE DE ACESSO
-- =============================================================================
-- Permissões granulares do sistema(ex: schedule.create, client.delete)
CREATE TABLE permissions(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code varchar(100) UNIQUE NOT NULL, -- ex: 'schedule.create'
  description varchar(255),
  default_roles user_role_type[] NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  CONSTRAINT default_roles_not_empty CHECK (array_length(default_roles, 1) > 0)
);

CREATE INDEX idx_permissions_code ON permissions(code);

-- Permissões atribuídas a cada usuário(escopadas por empresa em companies_users)
CREATE TABLE user_permissions(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  granted_by uuid REFERENCES users(id) ON DELETE SET NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  granted_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_by uuid REFERENCES users(id) ON DELETE SET NULL,
  revoked_at timestamptz DEFAULT NULL,
  UNIQUE (user_id, permission_id)
);

CREATE INDEX idx_users_permissions_user_id ON user_permissions(user_id);

-- =============================================================================
-- SEÇÃO 4: MÓDULOS E PLANOS(CONTROLE DE FEATURES)
-- =============================================================================
-- Módulos/funcionalidades da plataforma
CREATE TABLE modules(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code varchar(10) UNIQUE NOT NULL, -- ex: 'SCH', 'FIN', 'CRM'
  name varchar(100) NOT NULL,
  description varchar(255) NOT NULL,
  min_package module_package NOT NULL, -- pacote mínimo para ter acesso
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_modules_code ON modules(code);

CREATE TABLE module_permissions(
  module_id uuid NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (module_id, permission_id)
);

CREATE INDEX idx_module_permissions_module_id ON module_permissions(module_id);

CREATE INDEX idx_module_permissions_permission_id ON module_permissions(permission_id);

-- Tipos de planos(ex: Mensal, Anual, Trial)
CREATE TABLE plan_types(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name varchar(50) NOT NULL,
  description varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Planos de assinatura
CREATE TABLE plans(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_type_id uuid NOT NULL REFERENCES plan_types(id),
  name varchar(100) NOT NULL,
  description varchar(255) NOT NULL,
  package module_package NOT NULL,
  price numeric(12, 2) NOT NULL DEFAULT 0.00, -- NUNCA usar REAL para dinheiro
  billing_cycle_days integer NOT NULL DEFAULT 30,
  max_users integer DEFAULT NULL, -- NULL = ilimitado
  is_active boolean NOT NULL DEFAULT TRUE,
  image_url varchar(500),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Módulos incluídos em cada plano
CREATE TABLE plan_modules(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_id uuid NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  module_id uuid NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (plan_id, module_id)
);

-- Permissões incluídas em cada plano
CREATE TABLE plan_permissions(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_id uuid NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (plan_id, permission_id)
);

-- =============================================================================
-- SEÇÃO 5: EMPRESAS(MULTI-TENANT CORE)
-- =============================================================================
-- Idiomas suportados
CREATE TABLE languages(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code varchar(5) UNIQUE NOT NULL, -- ex: 'pt-BR', 'en-US'
  name varchar(50) NOT NULL,
  native_name varchar(50) NOT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Empresas clientes da plataforma
CREATE TABLE companies(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug varchar(100) UNIQUE NOT NULL, -- identificador URL-friendly
  name varchar(255) NOT NULL,
  fantasy_name varchar(255) NOT NULL,
  cnpj varchar(14) UNIQUE NOT NULL,
  foundation_date date,
  logo_url varchar(500),
  responsible_id uuid NOT NULL, -- FK adicionada após criação de people
  active_package module_package NOT NULL DEFAULT 'starter',
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_companies_slug ON companies(slug);

CREATE INDEX idx_companies_cnpj ON companies(cnpj);

CREATE INDEX idx_companies_active ON companies(is_active)
WHERE
  deleted_at IS NULL;

-- Assinaturas de planos por empresa
CREATE TABLE company_subscriptions(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  plan_id uuid NOT NULL REFERENCES plans(id),
  started_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamptz NOT NULL,
  canceled_at timestamptz DEFAULT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  price_paid numeric(12, 2) NOT NULL, -- preço no momento da contratação
  notes varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

CREATE INDEX idx_company_subscriptions_company ON company_subscriptions(company_id);

CREATE INDEX idx_company_subscriptions_active ON company_subscriptions(company_id)
WHERE
  is_active = TRUE AND canceled_at IS NULL;

-- Módulos ativos por empresa(derivado do plano, mas pode ter exceções manuais)
CREATE TABLE company_modules(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  module_id uuid NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (company_id, module_id)
);

CREATE INDEX idx_company_modules_company ON company_modules(company_id);

-- Vínculo entre usuários e empresas
CREATE TABLE company_users(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind user_kind NOT NULL,
  is_owner boolean NOT NULL DEFAULT FALSE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz,
  UNIQUE (company_id, user_id)
);

CREATE INDEX idx_company_users_company ON company_users(company_id);

CREATE INDEX idx_company_users_user ON company_users(user_id);

-- Configurações de sistema por empresa
CREATE TABLE company_system_configs(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid UNIQUE NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  schedule_init_time time NOT NULL DEFAULT '08:00',
  schedule_pause_init_time time DEFAULT NULL,
  schedule_pause_end_time time DEFAULT NULL,
  schedule_end_time time NOT NULL DEFAULT '18:00',
  min_schedules_per_day smallint NOT NULL DEFAULT 4,
  max_schedules_per_day smallint NOT NULL DEFAULT 6,
  schedule_days week_day[] NOT NULL DEFAULT ARRAY['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday']::week_day[],
  dynamic_cages boolean NOT NULL DEFAULT TRUE,
  total_small_cages smallint NOT NULL DEFAULT 0,
  total_medium_cages smallint NOT NULL DEFAULT 0,
  total_large_cages smallint NOT NULL DEFAULT 0,
  total_giant_cages smallint NOT NULL DEFAULT 0,
  whatsapp_notifications boolean NOT NULL DEFAULT FALSE,
  whatsapp_conversation boolean NOT NULL DEFAULT FALSE,
  whatsapp_business_phone varchar(25) DEFAULT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- =============================================================================
-- SEÇÃO 6: PESSOAS(DADOS PESSOAIS)
-- =============================================================================
-- Entidade base de pessoa(polimórfica: clientes, funcionários, etc.)
CREATE TABLE people(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  kind person_kind NOT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  has_system_user boolean NOT NULL DEFAULT FALSE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

CREATE INDEX idx_people_kind ON people(kind);

CREATE INDEX idx_people_active ON people(is_active);

-- Identificação pessoal
CREATE TABLE people_identifications(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  full_name varchar(150) NOT NULL,
  short_name varchar(75) NOT NULL,
  gender_identity gender_identity NOT NULL,
  marital_status marital_status NOT NULL,
  image_url varchar(500),
  birth_date date NOT NULL, -- DATE ao invés de TIMESTAMPTZ para nascimento
  cpf varchar(11) UNIQUE NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

CREATE INDEX idx_people_ident_cpf ON people_identifications(cpf);

-- Documentos específicos de funcionários CLT/contratados
CREATE TABLE employee_documents(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  rg varchar(15) UNIQUE NOT NULL,
  issuing_body varchar(20) NOT NULL,
  issuing_date date NOT NULL,
  ctps varchar(15) UNIQUE NOT NULL,
  ctps_series varchar(10) NOT NULL,
  ctps_state char(2) NOT NULL,
  pis varchar(11) UNIQUE NOT NULL,
  voter_registration varchar(12) UNIQUE,
  vote_zone varchar(5),
  vote_section varchar(5),
  military_certificate varchar(12) UNIQUE,
  military_series varchar(6),
  military_category char(2),
  has_special_needs boolean DEFAULT FALSE,
  has_children boolean NOT NULL DEFAULT FALSE,
  children_qty smallint DEFAULT 0,
  has_children_under_18 boolean DEFAULT FALSE,
  has_family_special_needs boolean DEFAULT FALSE,
  graduation graduation_level NOT NULL,
  has_cnh boolean DEFAULT FALSE,
  cnh_type varchar(3),
  cnh_number varchar(15),
  cnh_expiration_date date,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Contatos de pessoas
CREATE TABLE people_contacts(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  email varchar(150) NOT NULL,
  phone varchar(25),
  cellphone varchar(25) NOT NULL,
  has_whatsapp boolean NOT NULL DEFAULT FALSE,
  instagram_user varchar(100),
  emergency_contact varchar(150),
  emergency_phone varchar(25),
  is_primary boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

CREATE INDEX idx_people_contacts_person ON people_contacts(person_id);

-- Endereços
CREATE TABLE addresses(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  zip_code varchar(9) NOT NULL,
  street varchar(255) NOT NULL,
  number varchar(15) NOT NULL,
  complement varchar(100),
  district varchar(100) NOT NULL,
  city varchar(100) NOT NULL,
  state char(2) NOT NULL, -- sigla UF
  country varchar(100) NOT NULL DEFAULT 'Brasil',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Endereços de pessoas(pessoa pode ter múltiplos)
CREATE TABLE people_addresses(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  address_id uuid NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main boolean NOT NULL DEFAULT FALSE,
  label varchar(50) DEFAULT NULL, -- ex: 'Casa', 'Trabalho'
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (person_id, address_id)
);

CREATE INDEX idx_people_addresses_person ON people_addresses(person_id);

-- Dados financeiros/bancários de pessoas
CREATE TABLE finances(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  bank_name varchar(100) NOT NULL,
  bank_code varchar(10), -- código do banco(ex: '001' = BB)
  bank_branch varchar(10) NOT NULL,
  bank_account varchar(15) NOT NULL,
  bank_account_digit varchar(2) NOT NULL,
  bank_account_type bank_account_kind NOT NULL,
  has_pix boolean NOT NULL DEFAULT FALSE,
  pix_key varchar(255),
  pix_key_type pix_key_kind,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Vínculo financeiro entre pessoa e seus dados bancários
CREATE TABLE people_finances(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  finance_id uuid NOT NULL REFERENCES finances(id) ON DELETE CASCADE,
  is_primary boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (person_id, finance_id)
);

-- Vínculo entre users e people(usuário do sistema ↔ pessoa real)
CREATE TABLE user_profiles(
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  person_id uuid UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- SEÇÃO 7: EMPRESAS ↔ PESSOAS
-- =============================================================================
-- Pessoas vinculadas a empresas(funcionários, responsáveis, etc.)
CREATE TABLE company_people(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (company_id, person_id)
);

CREATE INDEX idx_company_people_company ON company_people(company_id);

-- Funcionários de empresa
CREATE TABLE company_employees(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamptz,
  UNIQUE (company_id, person_id)
);

CREATE INDEX idx_company_employees_company ON company_employees(company_id);

-- Vínculo empregatício do funcionário
CREATE TABLE employments(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id uuid NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  "role" varchar(100) NOT NULL,
  admission_date date NOT NULL,
  resignation_date date,
  salary numeric(12, 2) NOT NULL, -- NUMERIC para dinheiro
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_employments_employee ON employments(company_employee_id);

-- Benefícios dos funcionários(com histórico via valid_from/valid_until)
CREATE TABLE employee_benefits(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id uuid NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  meal_ticket boolean NOT NULL DEFAULT FALSE,
  meal_ticket_value numeric(10, 2) NOT NULL DEFAULT 0.00,
  transport_voucher boolean NOT NULL DEFAULT FALSE,
  transport_voucher_qty smallint NOT NULL DEFAULT 0,
  transport_voucher_value numeric(10, 2) NOT NULL DEFAULT 0.00,
  valid_from date NOT NULL DEFAULT CURRENT_DATE,
  valid_until date,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

CREATE INDEX idx_employee_benefits_employee ON employee_benefits(company_employee_id);

-- Custo total de funcionários para a empresa
CREATE TABLE company_employee_costs(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_employee_id uuid NOT NULL REFERENCES company_employees(id) ON DELETE CASCADE,
  costs numeric(12, 2) NOT NULL DEFAULT 0.00,
  reference_month date NOT NULL, -- primeiro dia do mês de referência
  comments varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Endereços de pessoas no contexto de uma empresa
CREATE TABLE company_people_addresses(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  person_id uuid NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  address_id uuid NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main boolean NOT NULL DEFAULT FALSE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (company_id, person_id, address_id)
);

-- Dados bancários de empresas
CREATE TABLE company_finances(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  finance_id uuid NOT NULL REFERENCES finances(id) ON DELETE CASCADE,
  is_primary boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (company_id, finance_id)
);

-- Endereços das empresas
CREATE TABLE company_addresses(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  address_id uuid NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
  is_main boolean NOT NULL DEFAULT FALSE,
  label varchar(50),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (company_id, address_id)
);

-- FK de responsible_id adicionada depois de people existir
ALTER TABLE companies
  ADD CONSTRAINT fk_companies_responsible FOREIGN KEY (responsible_id) REFERENCES people(id) ON DELETE RESTRICT;

-- =============================================================================
-- SEÇÃO 8: CLIENTES E PETS
-- =============================================================================
-- Clientes
CREATE TABLE clients(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  person_id uuid UNIQUE NOT NULL REFERENCES people(id) ON DELETE CASCADE,
  client_since date,
  recommended_by uuid REFERENCES clients(id) ON DELETE SET NULL,
  notes varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Vínculo entre empresa e seus clientes
CREATE TABLE company_clients(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  client_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  joined_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at timestamptz DEFAULT NULL,
  UNIQUE (company_id, client_id)
);

CREATE INDEX idx_company_clients_company ON company_clients(company_id);

-- Pets
CREATE TABLE pets(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name varchar(100) NOT NULL,
  size pet_size NOT NULL,
  kind pet_kind NOT NULL,
  temperament pet_temperament NOT NULL,
  image_url varchar(500),
  birth_date date,
  owner_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  guardian_id uuid REFERENCES people(id) ON DELETE SET NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  notes varchar(500),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_pets_owner ON pets(owner_id);

CREATE INDEX idx_pets_active ON pets(is_active)
WHERE
  deleted_at IS NULL;

-- Planos contratados por clientes
CREATE TABLE client_plans(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  client_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  plan_id uuid NOT NULL REFERENCES plans(id),
  started_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamptz NOT NULL,
  price_paid numeric(12, 2) NOT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (client_id, plan_id)
);

CREATE INDEX idx_client_plans_client ON client_plans(client_id);

-- =============================================================================
-- SEÇÃO 9: PRODUTOS E SERVIÇOS
-- =============================================================================
-- Tipos de serviço
CREATE TABLE service_types(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name varchar(100) NOT NULL,
  description varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Serviços oferecidos
CREATE TABLE services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  type_id uuid NOT NULL REFERENCES service_types(id),
  title varchar(100) NOT NULL,
  description varchar(500) NOT NULL,
  notes varchar(255),
  price numeric(12, 2) NOT NULL,
  discount_rate numeric(5, 2) NOT NULL DEFAULT 0.00, -- percentual
  image_url varchar(500),
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Sub-serviços(variações de um serviço, ex: banho pequeno, banho grande)
CREATE TABLE sub_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  type_id uuid REFERENCES service_types(id),
  title varchar(100) NOT NULL,
  description varchar(500) NOT NULL,
  notes varchar(255),
  price numeric(12, 2) NOT NULL,
  discount_rate numeric(5, 2) NOT NULL DEFAULT 0.00,
  image_url varchar(500),
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_sub_services_service ON sub_services(service_id);

-- Serviços disponíveis por empresa
CREATE TABLE company_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (company_id, service_id)
);

CREATE INDEX idx_company_services_company ON company_services(company_id);

-- Produtos(estoque/venda)
CREATE TABLE products(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name varchar(100) NOT NULL,
  batch_number varchar(50),
  description varchar(500),
  image_url varchar(500),
  expiration_date date,
  quantity integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Produtos por empresa
CREATE TABLE company_products(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  product_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  kind product_kind NOT NULL,
  has_stock boolean NOT NULL DEFAULT TRUE,
  for_sale boolean NOT NULL DEFAULT FALSE,
  cost_per_unit numeric(12, 2) NOT NULL, -- custo de aquisição
  profit_margin numeric(5, 2) NOT NULL DEFAULT 0.00,
  sale_price numeric(12, 2) NOT NULL, -- preço final calculado ou manual
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz,
  UNIQUE (company_id, product_id)
);

CREATE INDEX idx_company_products_company ON company_products(company_id);

-- Custos operacionais da empresa(contas, notas fiscais, etc.)
CREATE TABLE company_business_costs(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  invoice_number varchar(50),
  invoice_url varchar(500),
  description varchar(255) NOT NULL,
  total_cost numeric(12, 2) NOT NULL,
  reference_month date NOT NULL,
  comments varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_company_business_costs_company ON company_business_costs(company_id);

-- =============================================================================
-- SEÇÃO 10: PLANOS DE SERVIÇO PARA CLIENTES
-- =============================================================================
-- Pacotes de serviços oferecidos como planos para clientes
CREATE TABLE service_plans(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_type_id uuid NOT NULL REFERENCES plan_types(id),
  title varchar(100) NOT NULL,
  description varchar(500) NOT NULL,
  notes varchar(255),
  price numeric(12, 2) NOT NULL,
  discount_rate numeric(5, 2) NOT NULL DEFAULT 0.00,
  image_url varchar(500),
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

-- Serviços incluídos em cada plano de serviço
CREATE TABLE service_plan_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id uuid NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (service_plan_id, service_id)
);

-- Sub-serviços incluídos em cada plano de serviço
CREATE TABLE service_plan_sub_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id uuid NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  sub_service_id uuid NOT NULL REFERENCES sub_services(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (service_plan_id, sub_service_id)
);

-- Serviços bônus em planos(benefícios extras)
CREATE TABLE service_plan_bonuses(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_plan_id uuid NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  service_id uuid REFERENCES services(id) ON DELETE CASCADE,
  sub_service_id uuid REFERENCES sub_services(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_bonus_has_one CHECK ((service_id IS NOT NULL OR sub_service_id IS NOT NULL))
);

-- Planos de serviço por empresa
CREATE TABLE company_service_plans(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  service_plan_id uuid NOT NULL REFERENCES service_plans(id) ON DELETE CASCADE,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (company_id, service_plan_id)
);

-- Planos de serviço contratados por clientes
CREATE TABLE client_service_plans(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  client_id uuid NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  service_plan_id uuid NOT NULL REFERENCES service_plans(id),
  started_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamptz NOT NULL,
  price_paid numeric(12, 2) NOT NULL,
  is_active boolean NOT NULL DEFAULT TRUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (client_id, service_plan_id)
);

-- =============================================================================
-- SEÇÃO 11: AGENDAMENTOS
-- =============================================================================
-- Agendamentos
-- Unificando date + hour em scheduled_at(TIMESTAMPTZ)
CREATE TABLE schedules(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  client_id uuid NOT NULL REFERENCES clients(id),
  pet_id uuid NOT NULL REFERENCES pets(id),
  scheduled_at timestamptz NOT NULL, -- data e hora unificados
  estimated_end timestamptz,
  notes varchar(500),
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz,
  UNIQUE (company_id, client_id, pet_id, scheduled_at)
);

CREATE INDEX idx_schedules_company ON schedules(company_id);

CREATE INDEX idx_schedules_client ON schedules(client_id);

CREATE INDEX idx_schedules_date ON schedules(scheduled_at);

-- Histórico de status dos agendamentos
CREATE TABLE schedule_status_history(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  status schedule_status NOT NULL DEFAULT 'waiting',
  changed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  changed_by uuid REFERENCES users(id) ON DELETE SET NULL,
  notes varchar(255)
);

CREATE INDEX idx_schedule_status_history_schedule ON schedule_status_history(schedule_id);

-- Serviços do agendamento
CREATE TABLE schedule_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (schedule_id, service_id)
);

-- Sub-serviços do agendamento
CREATE TABLE schedule_sub_services(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  sub_service_id uuid NOT NULL REFERENCES sub_services(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (schedule_id, sub_service_id)
);

-- Check-in/out de agendamentos
CREATE TABLE schedule_checkins(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid UNIQUE NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  checked_in_at timestamptz,
  check_in_person_id uuid REFERENCES people(id) ON DELETE SET NULL,
  check_in_notes varchar(500),
  check_in_photo_url varchar(500),
  checked_out_at timestamptz,
  check_out_person_id uuid REFERENCES people(id) ON DELETE SET NULL,
  check_out_notes varchar(500),
  check_out_photo_url varchar(500),
  service_executor_person_id uuid REFERENCES people(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- Pagamentos de agendamentos
CREATE TABLE schedule_payments(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid UNIQUE NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
  payment_method_one payment_method NOT NULL,
  payment_method_two payment_method,
  payment_method_three payment_method,
  payment_date timestamptz NOT NULL,
  gross_value numeric(12, 2) NOT NULL,
  discount_value numeric(12, 2) NOT NULL DEFAULT 0.00,
  net_value numeric(12, 2) NOT NULL,
  amount_paid numeric(12, 2) NOT NULL DEFAULT 0.00,
  amount_remaining numeric(12, 2) GENERATED ALWAYS AS (net_value - amount_paid) STORED,
  notes varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  CONSTRAINT chk_payment_values CHECK (discount_value >= 0 AND net_value >= 0 AND amount_paid >= 0)
);

-- Ordens de serviço(impressão/comprovante)
CREATE TABLE service_orders(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  schedule_id uuid UNIQUE NOT NULL REFERENCES schedules(id),
  printed_by uuid REFERENCES users(id) ON DELETE SET NULL,
  printed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  gross_total numeric(12, 2) NOT NULL,
  amount_paid numeric(12, 2) NOT NULL,
  amount_to_pay numeric(12, 2) NOT NULL,
  has_perfume boolean NOT NULL DEFAULT FALSE,
  has_ornament boolean NOT NULL DEFAULT FALSE,
  notes varchar(255),
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz
);

-- =============================================================================
-- SEÇÃO 12: NOTIFICAÇÕES E AVISOS
-- =============================================================================
-- Notificações do sistema
CREATE TABLE notifications(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  company_id uuid REFERENCES companies(id) ON DELETE CASCADE, -- NULL = global
  title varchar(100) NOT NULL,
  summary varchar(200) NOT NULL,
  content text NOT NULL,
  image_url varchar(500),
  "level" notification_level NOT NULL,
  send_to_whatsapp boolean NOT NULL DEFAULT FALSE,
  send_to_telegram boolean NOT NULL DEFAULT FALSE,
  send_to_email boolean NOT NULL DEFAULT FALSE,
  send_to_sms boolean NOT NULL DEFAULT FALSE,
  send_to_push boolean NOT NULL DEFAULT FALSE,
  send_to_in_app_notification boolean NOT NULL DEFAULT TRUE,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  deleted_at timestamptz
);

CREATE INDEX idx_notifications_company ON notifications(company_id);

-- Destinatários de notificações
CREATE TABLE notification_receivers(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  notification_id uuid NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  read_at timestamptz DEFAULT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (notification_id, user_id)
);

CREATE INDEX idx_notification_receivers_user ON notification_receivers(user_id);

-- =============================================================================
-- SEÇÃO 12.1: CHAT INTERNO ENTRE ADMIN E SYSTEM
-- =============================================================================
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

-- =============================================================================
-- SEÇÃO 13: AUDITORIA(IMUTÁVEL)
-- =============================================================================
-- Tabelas de auditoria NUNCA devem ter ON DELETE CASCADE em referências a users.
-- Registros de auditoria são imutáveis — não se deletam, não se atualizam.
-- Log geral de ações no sistema
CREATE TABLE audit_logs(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  action log_action NOT NULL,
  entity_table varchar(100) NOT NULL,
  entity_id uuid NOT NULL,
  company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  old_data jsonb, -- JSONB permite queries nos dados históricos
  new_data jsonb NOT NULL,
  changed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  changed_by uuid REFERENCES users(id) ON DELETE SET NULL, -- SET NULL, nunca CASCADE
  ip_address inet,
  user_agent text
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_table, entity_id);

CREATE INDEX idx_audit_logs_company ON audit_logs(company_id);

CREATE INDEX idx_audit_logs_changed_by ON audit_logs(changed_by);

CREATE INDEX idx_audit_logs_changed_at ON audit_logs(changed_at DESC);

CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- Log específico de autenticação(separado por volume e sensibilidade)
CREATE TABLE auth_logs(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid REFERENCES users(id) ON DELETE SET NULL, -- SET NULL, nunca CASCADE
  action log_action NOT NULL, -- login, logout, password_changed, etc.
  company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  session_id uuid REFERENCES user_sessions(id) ON DELETE SET NULL,
  ip_address inet NOT NULL,
  user_agent text NOT NULL,
  result login_result,
  detail varchar(255),
  occurred_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_logs_user_id ON auth_logs(user_id);

CREATE INDEX idx_auth_logs_company_id ON auth_logs(company_id);

CREATE INDEX idx_auth_logs_occurred_at ON auth_logs(occurred_at DESC);

-- =============================================================================
-- SEÇÃO 14: INTERNACIONALIZAÇÃO
-- =============================================================================
-- Traduções de entidades do sistema
CREATE TABLE translations(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  language_code varchar(5) NOT NULL REFERENCES languages(code) ON DELETE CASCADE,
  entity_table varchar(100) NOT NULL,
  entity_id uuid NOT NULL,
  field varchar(100) NOT NULL,
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  UNIQUE (language_code, entity_table, entity_id, field)
);

CREATE INDEX idx_translations_entity ON translations(entity_table, entity_id);
