-- =============================================================================
-- PETCONTROL - ROLLBACK SCHEMA PostgreSQL v2
-- Ordem: Tabelas (Filhas -> Pais) -> Enums -> Extensões
-- =============================================================================
-- SEÇÃO 14: INTERNACIONALIZAÇÃO
DROP TABLE IF EXISTS translations;

DROP TABLE IF EXISTS languages;

-- SEÇÃO 13: AUDITORIA
DROP TABLE IF EXISTS auth_logs;

DROP TABLE IF EXISTS audit_logs;

-- SEÇÃO 12: NOTIFICAÇÕES E AVISOS
DROP TABLE IF EXISTS notification_receivers;

DROP TABLE IF EXISTS notifications;

-- SEÇÃO 11: AGENDAMENTOS E FINANCEIRO DE SERVIÇO
DROP TABLE IF EXISTS service_orders;

DROP TABLE IF EXISTS schedule_payments;

DROP TABLE IF EXISTS schedule_checkins;

DROP TABLE IF EXISTS schedule_sub_services;

DROP TABLE IF EXISTS schedule_services;

DROP TABLE IF EXISTS schedule_status_history;

DROP TABLE IF EXISTS schedules;

-- SEÇÃO 10: PLANOS DE SERVIÇO PARA CLIENTES
DROP TABLE IF EXISTS client_service_plans;

DROP TABLE IF EXISTS company_service_plans;

DROP TABLE IF EXISTS service_plan_bonuses;

DROP TABLE IF EXISTS service_plan_sub_services;

DROP TABLE IF EXISTS service_plan_services;

DROP TABLE IF EXISTS service_plans;

-- SEÇÃO 9: PRODUTOS E SERVIÇOS
DROP TABLE IF EXISTS company_business_costs;

DROP TABLE IF EXISTS company_products;

DROP TABLE IF EXISTS products;

DROP TABLE IF EXISTS company_services;

DROP TABLE IF EXISTS sub_services;

DROP TABLE IF EXISTS services;

DROP TABLE IF EXISTS service_types;

-- SEÇÃO 8: CLIENTES E PETS
DROP TABLE IF EXISTS client_plans;

DROP TABLE IF EXISTS pet_guardians;

DROP TABLE IF EXISTS pets;

DROP TABLE IF EXISTS company_clients;

DROP TABLE IF EXISTS clients;

-- SEÇÃO 7 & 6: EMPRESAS ↔ PESSOAS / DADOS PESSOAIS
-- Removendo primeiro a restrição circular de responsible_id em companies
ALTER TABLE IF EXISTS companies
    DROP CONSTRAINT IF EXISTS fk_companies_responsible;

DROP TABLE IF EXISTS company_addresses;

DROP TABLE IF EXISTS company_finances;

DROP TABLE IF EXISTS company_employee_costs;

DROP TABLE IF EXISTS employee_benefits;

DROP TABLE IF EXISTS employments;

DROP TABLE IF EXISTS company_employees;

DROP TABLE IF EXISTS company_people;

DROP TABLE IF EXISTS user_profiles;

DROP TABLE IF EXISTS people_finances;

DROP TABLE IF EXISTS finances;

DROP TABLE IF EXISTS people_addresses;

DROP TABLE IF EXISTS addresses;

DROP TABLE IF EXISTS people_contacts;

DROP TABLE IF EXISTS employee_documents;

DROP TABLE IF EXISTS people_identifications;

DROP TABLE IF EXISTS people;

-- SEÇÃO 5: EMPRESAS (MULTI-TENANT CORE)
DROP TABLE IF EXISTS company_system_configs;

DROP TRIGGER IF EXISTS trg_company_users_no_root_internal ON company_users;

DROP TABLE IF EXISTS company_users;

DROP FUNCTION IF EXISTS enforce_company_user_role_policy();

DROP TABLE IF EXISTS company_modules;

DROP TABLE IF EXISTS company_subscriptions;

DROP TABLE IF EXISTS companies;

-- SEÇÃO 4: MÓDULOS E PLANOS
DROP TABLE IF EXISTS plan_permissions;

DROP TABLE IF EXISTS plan_modules;

DROP TABLE IF EXISTS plans;

DROP TABLE IF EXISTS plan_types;

DROP TABLE IF EXISTS module_permissions;

DROP TABLE IF EXISTS modules;

-- SEÇÃO 3: PERMISSÕES
DROP TABLE IF EXISTS user_permissions;

DROP TABLE IF EXISTS permissions;

-- SEÇÃO 2: AUTENTICAÇÃO E SESSÕES
DROP TABLE IF EXISTS user_settings;

DROP TABLE IF EXISTS email_verification_tokens;

DROP TABLE IF EXISTS password_recovery_tokens;

DROP TABLE IF EXISTS user_sessions;

DROP TABLE IF EXISTS login_history;

DROP TABLE IF EXISTS user_auth;

DROP TABLE IF EXISTS users;

-- =============================================================================
-- REMOÇÃO DE TIPOS (ENUMS)
-- =============================================================================
DROP TYPE IF EXISTS logout_reason;

DROP TYPE IF EXISTS login_result;

DROP TYPE IF EXISTS product_kind;

DROP TYPE IF EXISTS schedule_status;

DROP TYPE IF EXISTS payment_method;

DROP TYPE IF EXISTS bank_account_kind;

DROP TYPE IF EXISTS person_kind;

DROP TYPE IF EXISTS graduation_level;

DROP TYPE IF EXISTS week_day;

DROP TYPE IF EXISTS pix_key_kind;

DROP TYPE IF EXISTS employee_kind;

DROP TYPE IF EXISTS pet_kind;

DROP TYPE IF EXISTS pet_temperament;

DROP TYPE IF EXISTS pet_size;

DROP TYPE IF EXISTS marital_status;

DROP TYPE IF EXISTS gender_identity;

DROP TYPE IF EXISTS log_action;

DROP TYPE IF EXISTS module_package;

DROP TYPE IF EXISTS user_kind;

DROP TYPE IF EXISTS user_role_type;

DROP TYPE IF EXISTS notification_level;

-- FINALIZAÇÃO
DROP EXTENSION IF EXISTS "uuid-ossp";
