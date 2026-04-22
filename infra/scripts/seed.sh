#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

db_url="${DATABASE_URL}"
network_arg=""

if [ -n "${DOCKER_NETWORK:-}" ]; then
  network_arg="--network ${DOCKER_NETWORK}"
else
  case "$(uname -s)" in
    Linux*)
      network_arg="--network host"
      ;;
    *)
      db_url=$(printf '%s' "${db_url}" | sed 's/@localhost:/@host.docker.internal:/g; s/@127\.0\.0\.1:/@host.docker.internal:/g')
      ;;
  esac
fi

docker run --rm \
  ${network_arg} \
  -i \
  postgres:18-alpine \
  psql "${db_url}" -v ON_ERROR_STOP=1 <<'SQL'
-- Modules
WITH module_seed(code, name, description, min_package) AS (
  VALUES
    ('CFG', 'Configurações', 'Módulo de Configurações', 'starter'::module_package),
    ('UCR', 'Usuários', 'Módulo de Usuários', 'starter'::module_package),
    ('SCH', 'Agendamentos', 'Módulo de Agendamentos', 'starter'::module_package),
    ('SVC', 'Serviços', 'Módulo de Serviços', 'starter'::module_package),
    ('PPL', 'Pessoas', 'Módulo de Pessoas', 'starter'::module_package),
    ('SPM', 'Planos de Serviços', 'Módulo de Planos de Serviços', 'basic'::module_package),
    ('PET', 'Pets', 'Módulo de Pets', 'basic'::module_package),
    ('TNT', 'Empresas', 'Módulo de Empresas', 'internal'::module_package),
    ('DHB', 'Dashboard', 'Módulo de Dashboard/Estatísticas', 'basic'::module_package),
    ('CLI', 'Clientes', 'Módulo de Clientes', 'starter'::module_package),
    ('RPT', 'Relatórios', 'Módulo de Relatórios', 'basic'::module_package),
    ('CRP', 'Relatórios Personalizados', 'Módulo de Relatórios Personalizados', 'premium'::module_package),
    ('PRD', 'Produtos', 'Módulo de Produtos', 'essential'::module_package),
    ('GSM', 'Agendamentos por Profissionais', 'Módulo de Agendamentos por Profissionais', 'essential'::module_package),
    ('DLV', 'Tele-busca/Entrega de Pets', 'Módulo de Tele-busca/Entrega de Pets', 'essential'::module_package),
    ('PDC', 'Creche de Pets', 'Módulo de Creche de Pets', 'premium'::module_package),
    ('PHO', 'Hotel Pet', 'Módulo de Hotel', 'premium'::module_package),
    ('CHT', 'Chat', 'Módulo de Chat', 'premium'::module_package),
    ('NTF', 'Notificações', 'Módulo de Notificações', 'premium'::module_package),
    ('FIN', 'Finanças', 'Módulo de Finanças', 'premium'::module_package),
    ('INV', 'Estoque', 'Módulo de Estoque', 'essential'::module_package),
    ('SUP', 'Fornecedores', 'Módulo de Fornecedores', 'premium'::module_package),
    ('EUA', 'Acesso de Usuários Externos', 'Módulo de Acesso de Usuários Externos', 'premium'::module_package),
    ('AUD', 'Logs de Auditoria', 'Módulo de Logs', 'internal'::module_package),
    ('ATL', 'Logs de Autenticação', 'Módulo de Logs de Autenticação', 'internal'::module_package),
    -- Transitional legacy code still used by current app routes/middleware.
    ('CRM', 'Gestão de Clientes', 'Módulo de Gestão de Clientes legado', 'starter'::module_package)
)
INSERT INTO modules (code, name, description, min_package)
SELECT
  ms.code,
  ms.name,
  ms.description,
  ms.min_package
FROM module_seed ms
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  min_package = EXCLUDED.min_package,
  updated_at = NOW();

-- Permissions catalog based on docs/conventions/permissions.md
WITH permission_seed(code, description, default_roles) AS (
  VALUES
    ('company_settings:edit', 'Editar configurações gerais', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('plan_settings:edit', 'Editar configurações de plano', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('payment_settings:edit', 'Editar configurações de pagamento', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('notification_settings:edit', 'Editar configurações de notificações', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('integration_settings:edit', 'Editar configurações de integração', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('security_settings:edit', 'Editar configurações de segurança', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:create', 'Criar usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:view', 'Visualizar usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:update', 'Atualizar usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:delete', 'Deletar usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:restore', 'Restaurar usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:block', 'Bloquear usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('users:unblock', 'Desbloquear usuário', ARRAY['root'::user_role_type, 'internal'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('people:create', 'Criar pessoa', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('people:view', 'Visualizar pessoa', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('people:update', 'Atualizar pessoa', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:create', 'Criar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:view', 'Visualizar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:update', 'Atualizar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:delete', 'Deletar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:restore', 'Restaurar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:deactivate', 'Desativar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('clients:reactivate', 'Reativar cliente', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pets:create', 'Criar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('pets:view', 'Visualizar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('pets:update', 'Atualizar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('pets:delete', 'Deletar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('pets:deactivate', 'Desativar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('pets:reactivate', 'Reativar pet', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:create', 'Criar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:view', 'Visualizar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:update', 'Atualizar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:delete', 'Deletar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:deactivate', 'Desativar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('schedules:reactivate', 'Reativar agendamento', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('products:create', 'Criar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('products:view', 'Visualizar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('products:update', 'Atualizar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('products:delete', 'Deletar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('products:deactivate', 'Desativar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('products:reactivate', 'Reativar produto', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('services:create', 'Criar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('services:view', 'Visualizar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type, 'common'::user_role_type]::user_role_type[]),
    ('services:update', 'Atualizar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('services:delete', 'Deletar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('services:deactivate', 'Desativar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('services:reactivate', 'Reativar serviço', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('logs:view', 'Visualizar logs', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('plans:create', 'Criar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:view', 'Visualizar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:update', 'Atualizar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:delete', 'Deletar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:restore', 'Restaurar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:deactivate', 'Desativar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('plans:reactivate', 'Reativar plano', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:create', 'Criar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:view', 'Visualizar relatórios', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('reports:update', 'Atualizar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:delete', 'Deletar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:restore', 'Restaurar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:deactivate', 'Desativar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('reports:reactivate', 'Reativar relatório', ARRAY['root'::user_role_type]::user_role_type[]),
    ('finances:create', 'Criar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:view', 'Visualizar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:update', 'Atualizar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:delete', 'Deletar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:restore', 'Restaurar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:deactivate', 'Desativar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('finances:reactivate', 'Reativar transação', ARRAY['root'::user_role_type, 'admin'::user_role_type]::user_role_type[]),
    ('suppliers:create', 'Criar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:view', 'Visualizar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:update', 'Atualizar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:delete', 'Deletar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:restore', 'Restaurar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:deactivate', 'Desativar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('suppliers:reactivate', 'Reativar fornecedor', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:create', 'Criar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:view', 'Visualizar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:update', 'Atualizar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:delete', 'Deletar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:deactivate', 'Desativar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('external_access:reactivate', 'Reativar acesso', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:create', 'Criar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:view', 'Visualizar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:update', 'Atualizar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:delete', 'Deletar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:deactivate', 'Desativar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('stock:reactivate', 'Reativar estoque', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:create', 'Criar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:view', 'Visualizar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:update', 'Atualizar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:delete', 'Deletar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:restore', 'Restaurar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:deactivate', 'Desativar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('daycare:reactivate', 'Reativar creche', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:create', 'Criar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:view', 'Visualizar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:update', 'Atualizar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:delete', 'Deletar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:restore', 'Restaurar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:deactivate', 'Desativar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('hotel:reactivate', 'Reativar hotel', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:create', 'Criar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:view', 'Visualizar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:update', 'Atualizar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:delete', 'Deletar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:restore', 'Restaurar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:deactivate', 'Desativar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('chat:reactivate', 'Reativar chat', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:create', 'Criar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:view', 'Visualizar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:update', 'Atualizar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:delete', 'Deletar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:restore', 'Restaurar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:deactivate', 'Desativar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('notifications:reactivate', 'Reativar notificação', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:create', 'Criar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:view', 'Visualizar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:update', 'Atualizar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:delete', 'Deletar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:restore', 'Restaurar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:deactivate', 'Desativar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[]),
    ('pickup_delivery:reactivate', 'Reativar serviço de tele-busca', ARRAY['root'::user_role_type, 'admin'::user_role_type, 'system'::user_role_type]::user_role_type[])
)
INSERT INTO permissions (code, description, default_roles)
SELECT
  ps.code,
  ps.description,
  ps.default_roles
FROM permission_seed ps
ON CONFLICT (code) DO UPDATE SET
  description = EXCLUDED.description,
  default_roles = EXCLUDED.default_roles,
  updated_at = NOW();

-- Module permissions required by tenant settings and module-driven access
WITH module_permission_seed(module_code, permission_code) AS (
  VALUES
    ('CFG', 'company_settings:edit'),
    ('CFG', 'plan_settings:edit'),
    ('CFG', 'payment_settings:edit'),
    ('CFG', 'notification_settings:edit'),
    ('CFG', 'integration_settings:edit'),
    ('CFG', 'security_settings:edit'),
    ('UCR', 'users:create'),
    ('UCR', 'users:view'),
    ('UCR', 'users:update'),
    ('UCR', 'users:delete'),
    ('UCR', 'users:restore'),
    ('UCR', 'users:block'),
    ('UCR', 'users:unblock'),
    ('PPL', 'people:create'),
    ('PPL', 'people:view'),
    ('PPL', 'people:update'),
    ('CLI', 'clients:create'),
    ('CLI', 'clients:view'),
    ('CLI', 'clients:update'),
    ('CLI', 'clients:delete'),
    ('CLI', 'clients:restore'),
    ('CLI', 'clients:deactivate'),
    ('CLI', 'clients:reactivate'),
    ('PET', 'pets:create'),
    ('PET', 'pets:view'),
    ('PET', 'pets:update'),
    ('PET', 'pets:delete'),
    ('PET', 'pets:deactivate'),
    ('PET', 'pets:reactivate'),
    ('SCH', 'schedules:create'),
    ('SCH', 'schedules:view'),
    ('SCH', 'schedules:update'),
    ('SCH', 'schedules:delete'),
    ('SCH', 'schedules:deactivate'),
    ('SCH', 'schedules:reactivate'),
    ('SVC', 'services:create'),
    ('SVC', 'services:view'),
    ('SVC', 'services:update'),
    ('SVC', 'services:delete'),
    ('SVC', 'services:deactivate'),
    ('SVC', 'services:reactivate'),
    ('RPT', 'reports:create'),
    ('RPT', 'reports:view'),
    ('RPT', 'reports:update'),
    ('RPT', 'reports:delete'),
    ('RPT', 'reports:restore'),
    ('RPT', 'reports:deactivate'),
    ('RPT', 'reports:reactivate'),
    ('PRD', 'products:create'),
    ('PRD', 'products:view'),
    ('PRD', 'products:update'),
    ('PRD', 'products:delete'),
    ('PRD', 'products:deactivate'),
    ('PRD', 'products:reactivate'),
    ('DLV', 'pickup_delivery:create'),
    ('DLV', 'pickup_delivery:view'),
    ('DLV', 'pickup_delivery:update'),
    ('DLV', 'pickup_delivery:delete'),
    ('DLV', 'pickup_delivery:restore'),
    ('DLV', 'pickup_delivery:deactivate'),
    ('DLV', 'pickup_delivery:reactivate'),
    ('INV', 'stock:create'),
    ('INV', 'stock:view'),
    ('INV', 'stock:update'),
    ('INV', 'stock:delete'),
    ('INV', 'stock:deactivate'),
    ('INV', 'stock:reactivate'),
    ('PDC', 'daycare:create'),
    ('PDC', 'daycare:view'),
    ('PDC', 'daycare:update'),
    ('PDC', 'daycare:delete'),
    ('PDC', 'daycare:restore'),
    ('PDC', 'daycare:deactivate'),
    ('PDC', 'daycare:reactivate'),
    ('PHO', 'hotel:create'),
    ('PHO', 'hotel:view'),
    ('PHO', 'hotel:update'),
    ('PHO', 'hotel:delete'),
    ('PHO', 'hotel:restore'),
    ('PHO', 'hotel:deactivate'),
    ('PHO', 'hotel:reactivate'),
    ('CHT', 'chat:create'),
    ('CHT', 'chat:view'),
    ('CHT', 'chat:update'),
    ('CHT', 'chat:delete'),
    ('CHT', 'chat:restore'),
    ('CHT', 'chat:deactivate'),
    ('CHT', 'chat:reactivate'),
    ('NTF', 'notifications:create'),
    ('NTF', 'notifications:view'),
    ('NTF', 'notifications:update'),
    ('NTF', 'notifications:delete'),
    ('NTF', 'notifications:restore'),
    ('NTF', 'notifications:deactivate'),
    ('NTF', 'notifications:reactivate'),
    ('FIN', 'finances:create'),
    ('FIN', 'finances:view'),
    ('FIN', 'finances:update'),
    ('FIN', 'finances:delete'),
    ('FIN', 'finances:restore'),
    ('FIN', 'finances:deactivate'),
    ('FIN', 'finances:reactivate'),
    ('SUP', 'suppliers:create'),
    ('SUP', 'suppliers:view'),
    ('SUP', 'suppliers:update'),
    ('SUP', 'suppliers:delete'),
    ('SUP', 'suppliers:restore'),
    ('SUP', 'suppliers:deactivate'),
    ('SUP', 'suppliers:reactivate'),
    ('EUA', 'external_access:create'),
    ('EUA', 'external_access:view'),
    ('EUA', 'external_access:update'),
    ('EUA', 'external_access:delete'),
    ('EUA', 'external_access:deactivate'),
    ('EUA', 'external_access:reactivate'),
    ('AUD', 'logs:view')
), resolved_module_permissions AS (
  SELECT
    m.id AS module_id,
    p.id AS permission_id
  FROM module_permission_seed mps
  INNER JOIN modules m ON m.code = mps.module_code AND m.deleted_at IS NULL
  INNER JOIN permissions p ON p.code = mps.permission_code
)
INSERT INTO module_permissions (module_id, permission_id)
SELECT
  rmp.module_id,
  rmp.permission_id
FROM resolved_module_permissions rmp
ON CONFLICT (module_id, permission_id) DO NOTHING;

-- Plan types
INSERT INTO plan_types (name, description)
SELECT 'Monthly', 'Default monthly billing cycle'
WHERE NOT EXISTS (SELECT 1 FROM plan_types WHERE name = 'Monthly' AND deleted_at IS NULL);

INSERT INTO plan_types (name, description)
SELECT 'Annual', 'Default annual billing cycle'
WHERE NOT EXISTS (SELECT 1 FROM plan_types WHERE name = 'Annual' AND deleted_at IS NULL);

-- Plans
WITH monthly_type AS (
  SELECT id FROM plan_types WHERE name = 'Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
)
INSERT INTO plans (plan_type_id, name, description, package, price, billing_cycle_days, max_users, is_active)
SELECT mt.id, 'Starter Monthly', 'Initial starter monthly plan', 'starter', 99.90, 30, 5, TRUE
FROM monthly_type mt
WHERE NOT EXISTS (
  SELECT 1 FROM plans p WHERE p.name = 'Starter Monthly' AND p.deleted_at IS NULL
);

-- Plan modules
WITH starter_plan AS (
  SELECT id FROM plans WHERE name = 'Starter Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
), starter_modules AS (
  SELECT id FROM modules WHERE code IN ('CFG', 'UCR', 'SCH', 'SVC', 'PPL', 'CLI', 'CRM')
)
INSERT INTO plan_modules (plan_id, module_id, is_active)
SELECT sp.id, sm.id, TRUE
FROM starter_plan sp
CROSS JOIN starter_modules sm
ON CONFLICT (plan_id, module_id) DO NOTHING;

-- Responsible person (required by companies.responsible_id)
INSERT INTO people (kind, is_active, has_system_user)
SELECT 'responsible', TRUE, FALSE
WHERE NOT EXISTS (
  SELECT
    1
  FROM
    companies
  WHERE
    slug = 'petcontrol-dev'
);

-- Dev company
WITH responsible AS (
  SELECT
    id
  FROM
    people
  WHERE
    kind = 'responsible'
  ORDER BY
    created_at DESC
  LIMIT 1
)
INSERT INTO companies (slug, name, fantasy_name, cnpj, responsible_id, active_package, is_active)
SELECT
  'petcontrol-dev',
  'PetControl Desenvolvimento LTDA',
  'PetControl Dev',
  '12345678000195',
  r.id,
  'starter',
  TRUE
FROM
  responsible r
WHERE NOT EXISTS (
  SELECT
    1
  FROM
    companies c
  WHERE
    c.slug = 'petcontrol-dev'
);

-- Root user (dev bootstrap)
INSERT INTO users (email, email_verified, email_verified_at, role, is_active)
SELECT 'root@petcontrol.local', TRUE, NOW(), 'root', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'root@petcontrol.local');

-- Root auth profile (password: password123, requires password change)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HAtO6l.iXD27nYmeaFjSEeeiYPo0TVPANJzxhUG/DvC5xzdAp2QrG', TRUE
FROM users u
WHERE u.email = 'root@petcontrol.local'
ON CONFLICT (user_id) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  must_change_password = EXCLUDED.must_change_password,
  updated_at = NOW();

-- Admin user compatible with web default credentials
INSERT INTO users (email, email_verified, email_verified_at, role, is_active)
SELECT 'admin@petcontrol.local', TRUE, NOW(), 'admin', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@petcontrol.local');

-- Admin auth profile (password: password123)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HAtO6l.iXD27nYmeaFjSEeeiYPo0TVPANJzxhUG/DvC5xzdAp2QrG', FALSE
FROM users u
WHERE u.email = 'admin@petcontrol.local'
ON CONFLICT (user_id) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  must_change_password = EXCLUDED.must_change_password,
  updated_at = NOW();

-- System user for tenant support flows
INSERT INTO users (email, email_verified, email_verified_at, role, is_active)
SELECT 'system@petcontrol.local', TRUE, NOW(), 'system', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'system@petcontrol.local');

-- System auth profile (password: password123)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HAtO6l.iXD27nYmeaFjSEeeiYPo0TVPANJzxhUG/DvC5xzdAp2QrG', FALSE
FROM users u
WHERE u.email = 'system@petcontrol.local'
ON CONFLICT (user_id) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  must_change_password = EXCLUDED.must_change_password,
  updated_at = NOW();

-- Active memberships for seeded users
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_users AS (
  SELECT id, email FROM users WHERE email IN ('admin@petcontrol.local', 'system@petcontrol.local')
)
INSERT INTO company_users (company_id, user_id, kind, is_owner, is_active)
SELECT
  dc.id,
  su.id,
  CASE WHEN su.email = 'admin@petcontrol.local' THEN 'owner'::user_kind ELSE 'employee'::user_kind END,
  CASE WHEN su.email = 'admin@petcontrol.local' THEN TRUE ELSE FALSE END,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_users su
WHERE NOT EXISTS (
  SELECT 1
  FROM company_users cu
  WHERE cu.company_id = dc.id AND cu.user_id = su.id
);

-- Default permissions for every user according to permissions.default_roles
INSERT INTO user_permissions (user_id, permission_id, granted_by)
SELECT
  u.id,
  p.id,
  NULL
FROM users u
INNER JOIN permissions p ON u.role = ANY(p.default_roles)
WHERE NOT EXISTS (
  SELECT 1
  FROM user_permissions up
  WHERE up.user_id = u.id
    AND up.permission_id = p.id
);

-- System people for seeded users
INSERT INTO people (kind, is_active, has_system_user)
SELECT 'employee', TRUE, TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'root@petcontrol.local'
);

INSERT INTO people (kind, is_active, has_system_user)
SELECT 'employee', TRUE, TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'admin@petcontrol.local'
);

INSERT INTO people (kind, is_active, has_system_user)
SELECT 'employee', TRUE, TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'system@petcontrol.local'
);

WITH root_user AS (
  SELECT id
  FROM users
  WHERE email = 'root@petcontrol.local'
  LIMIT 1
), root_person AS (
  SELECT p.id
  FROM people p
  LEFT JOIN user_profiles up ON up.person_id = p.id
  WHERE p.kind = 'employee'
    AND p.has_system_user = TRUE
    AND up.user_id IS NULL
  ORDER BY p.created_at ASC
  LIMIT 1
)
INSERT INTO user_profiles (user_id, person_id)
SELECT ru.id, rp.id
FROM root_user ru
CROSS JOIN root_person rp
WHERE NOT EXISTS (
  SELECT 1 FROM user_profiles up WHERE up.user_id = ru.id
);

WITH admin_user AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
), admin_person AS (
  SELECT p.id
  FROM people p
  LEFT JOIN user_profiles up ON up.person_id = p.id
  WHERE p.kind = 'employee'
    AND p.has_system_user = TRUE
    AND up.user_id IS NULL
  ORDER BY p.created_at DESC
  LIMIT 1
)
INSERT INTO user_profiles (user_id, person_id)
SELECT au.id, ap.id
FROM admin_user au
CROSS JOIN admin_person ap
WHERE NOT EXISTS (
  SELECT 1 FROM user_profiles up WHERE up.user_id = au.id
);

INSERT INTO user_profiles (user_id, person_id)
SELECT
  su.id,
  sp.id
FROM users su
CROSS JOIN LATERAL (
  SELECT p.id
  FROM people p
  LEFT JOIN user_profiles up ON up.person_id = p.id
  WHERE p.kind = 'employee'
    AND p.has_system_user = TRUE
    AND up.user_id IS NULL
  ORDER BY p.created_at DESC
  LIMIT 1
) sp
WHERE su.email = 'system@petcontrol.local'
  AND NOT EXISTS (
  SELECT 1 FROM user_profiles up WHERE up.user_id = su.id
);

WITH root_profile AS (
  SELECT up.person_id
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'root@petcontrol.local'
  LIMIT 1
)
INSERT INTO people_identifications (
  person_id,
  full_name,
  short_name,
  gender_identity,
  marital_status,
  birth_date,
  cpf
)
SELECT
  rp.person_id,
  'Root PetControl',
  'Root',
  'not_to_expose',
  'single',
  DATE '1990-01-01',
  '00000000001'
FROM root_profile rp
WHERE NOT EXISTS (
  SELECT 1 FROM people_identifications pi WHERE pi.person_id = rp.person_id
);

WITH admin_profile AS (
  SELECT up.person_id
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'admin@petcontrol.local'
  LIMIT 1
)
INSERT INTO people_identifications (
  person_id,
  full_name,
  short_name,
  gender_identity,
  marital_status,
  birth_date,
  cpf
)
SELECT
  ap.person_id,
  'Admin PetControl',
  'Admin',
  'not_to_expose',
  'single',
  DATE '1991-01-01',
  '00000000002'
FROM admin_profile ap
WHERE NOT EXISTS (
  SELECT 1 FROM people_identifications pi WHERE pi.person_id = ap.person_id
);

WITH system_profile AS (
  SELECT up.person_id
  FROM user_profiles up
  INNER JOIN users u ON u.id = up.user_id
  WHERE u.email = 'system@petcontrol.local'
  LIMIT 1
)
INSERT INTO people_identifications (
  person_id,
  full_name,
  short_name,
  gender_identity,
  marital_status,
  birth_date,
  cpf
)
SELECT
  sp.person_id,
  'System PetControl',
  'System',
  'not_to_expose',
  'single',
  DATE '1992-01-01',
  '00000000003'
FROM system_profile sp
WHERE NOT EXISTS (
  SELECT 1 FROM people_identifications pi WHERE pi.person_id = sp.person_id
);

-- Link seeded user people to the tenant-wide people registry
WITH dev_company AS (
  SELECT id
  FROM companies
  WHERE slug = 'petcontrol-dev'
  LIMIT 1
), seeded_profiles AS (
  SELECT up.person_id
  FROM user_profiles up
  INNER JOIN company_users cu ON cu.user_id = up.user_id
  INNER JOIN users u ON u.id = up.user_id
  INNER JOIN dev_company dc ON dc.id = cu.company_id
  WHERE u.email IN ('admin@petcontrol.local', 'system@petcontrol.local')
)
INSERT INTO company_people (company_id, person_id)
SELECT dc.id, sp.person_id
FROM dev_company dc
CROSS JOIN seeded_profiles sp
WHERE NOT EXISTS (
  SELECT 1
  FROM company_people cp
  WHERE cp.company_id = dc.id
    AND cp.person_id = sp.person_id
);

-- Persisted admin-system conversation for dashboard chat
WITH dev_company AS (
  SELECT id
  FROM companies
  WHERE slug = 'petcontrol-dev'
  LIMIT 1
), admin_user AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
), support_user AS (
  SELECT id
  FROM users
  WHERE email = 'system@petcontrol.local'
  LIMIT 1
)
INSERT INTO admin_system_conversations (company_id, admin_user_id, system_user_id, updated_at)
SELECT dc.id, au.id, su.id, NOW()
FROM dev_company dc
CROSS JOIN admin_user au
CROSS JOIN support_user su
ON CONFLICT (company_id, admin_user_id, system_user_id) DO UPDATE SET
  updated_at = EXCLUDED.updated_at;

WITH conversation_seed AS (
  SELECT ascv.id AS conversation_id, ascv.company_id, ascv.admin_user_id, ascv.system_user_id
  FROM admin_system_conversations ascv
  INNER JOIN companies c ON c.id = ascv.company_id
  INNER JOIN users au ON au.id = ascv.admin_user_id
  INNER JOIN users su ON su.id = ascv.system_user_id
  WHERE c.slug = 'petcontrol-dev'
    AND au.email = 'admin@petcontrol.local'
    AND su.email = 'system@petcontrol.local'
  LIMIT 1
)
INSERT INTO admin_system_messages (conversation_id, company_id, sender_user_id, body, created_at)
SELECT cs.conversation_id, cs.company_id, seed.sender_user_id, raw_messages.body, raw_messages.created_at
FROM conversation_seed cs
CROSS JOIN (
  VALUES
    ('admin'::text, 'Bom dia, preciso acompanhar a operação desta semana.', NOW() - INTERVAL '1 day'),
    ('system'::text, 'Tudo certo. Já deixei o suporte operacional monitorando os agendamentos.', NOW() - INTERVAL '23 hours'),
    ('admin'::text, 'Perfeito. Quero validar também a eficiência mensal no novo dashboard.', NOW() - INTERVAL '22 hours'),
    ('system'::text, 'O histórico persistido deste chat já está ativo para o tenant de desenvolvimento.', NOW() - INTERVAL '21 hours')
) AS raw_messages(sender_kind, body, created_at)
CROSS JOIN LATERAL (
  SELECT CASE
    WHEN raw_messages.sender_kind = 'admin' THEN cs.admin_user_id
    ELSE cs.system_user_id
  END AS sender_user_id
) AS seed
WHERE NOT EXISTS (
  SELECT 1
  FROM admin_system_messages asm
  WHERE asm.conversation_id = cs.conversation_id
);

-- Active subscription for current seeded plan
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), starter_plan AS (
  SELECT id, price, billing_cycle_days FROM plans WHERE name = 'Starter Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
)
INSERT INTO company_subscriptions (company_id, plan_id, started_at, expires_at, is_active, price_paid, notes)
SELECT
  dc.id,
  sp.id,
  NOW() - INTERVAL '1 day',
  NOW() + make_interval(days => sp.billing_cycle_days),
  TRUE,
  sp.price,
  'Seeded development subscription'
FROM dev_company dc
CROSS JOIN starter_plan sp
WHERE NOT EXISTS (
  SELECT 1
  FROM company_subscriptions cs
  WHERE cs.company_id = dc.id
    AND cs.plan_id = sp.id
    AND cs.is_active = TRUE
    AND cs.canceled_at IS NULL
    AND cs.expires_at > NOW()
);

-- Active company modules for the seeded tenant
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), starter_modules AS (
  SELECT id FROM modules WHERE code IN ('CFG', 'UCR', 'SCH', 'SVC', 'PPL', 'CLI', 'CRM')
)
INSERT INTO company_modules (company_id, module_id, is_active)
SELECT dc.id, sm.id, TRUE
FROM dev_company dc
CROSS JOIN starter_modules sm
ON CONFLICT (company_id, module_id) DO UPDATE SET
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- System configuration required by admin dashboard
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
)
INSERT INTO company_system_configs (
  company_id,
  schedule_init_time,
  schedule_pause_init_time,
  schedule_pause_end_time,
  schedule_end_time,
  min_schedules_per_day,
  max_schedules_per_day,
  schedule_days,
  dynamic_cages,
  total_small_cages,
  total_medium_cages,
  total_large_cages,
  total_giant_cages,
  whatsapp_notifications,
  whatsapp_conversation,
  whatsapp_business_phone
)
SELECT
  dc.id,
  TIME '08:00',
  TIME '12:00',
  TIME '13:00',
  TIME '18:00',
  4,
  18,
  ARRAY[
    'monday'::week_day,
    'tuesday'::week_day,
    'wednesday'::week_day,
    'thursday'::week_day,
    'friday'::week_day,
    'saturday'::week_day
  ],
  FALSE,
  8,
  6,
  4,
  2,
  TRUE,
  TRUE,
  '+5511999990001'
FROM dev_company dc
ON CONFLICT (company_id) DO UPDATE SET
  schedule_init_time = EXCLUDED.schedule_init_time,
  schedule_pause_init_time = EXCLUDED.schedule_pause_init_time,
  schedule_pause_end_time = EXCLUDED.schedule_pause_end_time,
  schedule_end_time = EXCLUDED.schedule_end_time,
  min_schedules_per_day = EXCLUDED.min_schedules_per_day,
  max_schedules_per_day = EXCLUDED.max_schedules_per_day,
  schedule_days = EXCLUDED.schedule_days,
  dynamic_cages = EXCLUDED.dynamic_cages,
  total_small_cages = EXCLUDED.total_small_cages,
  total_medium_cages = EXCLUDED.total_medium_cages,
  total_large_cages = EXCLUDED.total_large_cages,
  total_giant_cages = EXCLUDED.total_giant_cages,
  whatsapp_notifications = EXCLUDED.whatsapp_notifications,
  whatsapp_conversation = EXCLUDED.whatsapp_conversation,
  whatsapp_business_phone = EXCLUDED.whatsapp_business_phone,
  updated_at = NOW();

-- Seeded client person for operational flows
INSERT INTO people (kind, is_active, has_system_user)
SELECT 'client', TRUE, FALSE
WHERE NOT EXISTS (
  SELECT 1
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
);

WITH seeded_person AS (
  SELECT p.id
  FROM people p
  LEFT JOIN people_identifications pi ON pi.person_id = p.id
  WHERE pi.cpf = '12345678901'
     OR (
       pi.person_id IS NULL
       AND p.kind = 'client'
       AND p.has_system_user = FALSE
     )
  ORDER BY p.created_at DESC
  LIMIT 1
)
INSERT INTO people_identifications (
  person_id,
  full_name,
  short_name,
  gender_identity,
  marital_status,
  birth_date,
  cpf
)
SELECT
  sp.id,
  'Maria Silva',
  'Maria',
  'woman_cisgender',
  'single',
  DATE '1992-06-15',
  '12345678901'
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1 FROM people_identifications WHERE cpf = '12345678901'
);

WITH seeded_person AS (
  SELECT pi.person_id AS id
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
  LIMIT 1
)
INSERT INTO people_contacts (
  person_id,
  email,
  phone,
  cellphone,
  has_whatsapp,
  is_primary
)
SELECT
  sp.id,
  'maria.silva@petcontrol.local',
  '+551130000000',
  '+5511999990001',
  TRUE,
  TRUE
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1
  FROM people_contacts pc
  WHERE pc.person_id = sp.id
    AND pc.email = 'maria.silva@petcontrol.local'
);

WITH seeded_person AS (
  SELECT pi.person_id AS id
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
  LIMIT 1
)
INSERT INTO clients (
  person_id,
  client_since,
  notes
)
SELECT
  sp.id,
  CURRENT_DATE - INTERVAL '45 days',
  'Cliente seedado para fluxos operacionais locais'
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1
  FROM clients c
  WHERE c.person_id = sp.id
    AND c.deleted_at IS NULL
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
)
INSERT INTO company_clients (
  company_id,
  client_id,
  is_active
)
SELECT
  dc.id,
  sc.id,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_client sc
WHERE NOT EXISTS (
  SELECT 1
  FROM company_clients cc
  WHERE cc.company_id = dc.id
    AND cc.client_id = sc.id
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_person AS (
  SELECT pi.person_id
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
  LIMIT 1
)
INSERT INTO company_people (company_id, person_id)
SELECT dc.id, sp.person_id
FROM dev_company dc
CROSS JOIN seeded_person sp
WHERE NOT EXISTS (
  SELECT 1
  FROM company_people cp
  WHERE cp.company_id = dc.id
    AND cp.person_id = sp.person_id
);

WITH seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
)
INSERT INTO pets (
  name,
  size,
  kind,
  temperament,
  birth_date,
  owner_id,
  is_active,
  notes
)
SELECT
  'Thor',
  'medium',
  'dog',
  'playful',
  DATE '2021-08-20',
  sc.id,
  TRUE,
  'Pet seedado para validar fluxos de agendamento'
FROM seeded_client sc
WHERE NOT EXISTS (
  SELECT 1
  FROM pets p
  WHERE p.owner_id = sc.id
    AND p.name = 'Thor'
    AND p.deleted_at IS NULL
);

INSERT INTO service_types (name, description)
SELECT 'Banho', 'Serviços de banho e higienização'
WHERE NOT EXISTS (
  SELECT 1
  FROM service_types st
  WHERE st.name = 'Banho'
    AND st.deleted_at IS NULL
);

WITH banho_type AS (
  SELECT id
  FROM service_types
  WHERE name = 'Banho'
    AND deleted_at IS NULL
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO services (
  type_id,
  title,
  description,
  notes,
  price,
  discount_rate,
  is_active
)
SELECT
  bt.id,
  'Banho completo',
  'Banho com secagem, perfume e escovação',
  'Serviço seedado para o catálogo local',
  89.90,
  0.00,
  TRUE
FROM banho_type bt
WHERE NOT EXISTS (
  SELECT 1
  FROM services s
  WHERE s.title = 'Banho completo'
    AND s.deleted_at IS NULL
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_service AS (
  SELECT id
  FROM services
  WHERE title = 'Banho completo'
    AND deleted_at IS NULL
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO company_services (
  company_id,
  service_id,
  is_active
)
SELECT
  dc.id,
  ss.id,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_service ss
WHERE NOT EXISTS (
  SELECT 1
  FROM company_services cs
  WHERE cs.company_id = dc.id
    AND cs.service_id = ss.id
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
), seeded_pet AS (
  SELECT p.id
  FROM pets p
  INNER JOIN seeded_client sc ON sc.id = p.owner_id
  WHERE p.name = 'Thor'
    AND p.deleted_at IS NULL
  LIMIT 1
), seeded_admin AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
), dashboard_schedule_scenarios AS (
  SELECT *
  FROM (
    VALUES
      (
        'dashboard-yesterday-confirmed',
        ((CURRENT_DATE - 1) + TIME '10:00')::timestamp,
        ((CURRENT_DATE - 1) + TIME '11:00')::timestamp,
        'Dashboard seed: comparativo diário (ontem)',
        'confirmed'
      ),
      (
        'dashboard-today-morning-waiting',
        (CURRENT_DATE + TIME '09:00')::timestamp,
        (CURRENT_DATE + TIME '10:00')::timestamp,
        'Dashboard seed: turno manhã aguardando',
        'waiting'
      ),
      (
        'dashboard-today-morning-in-progress',
        (CURRENT_DATE + TIME '10:30')::timestamp,
        (CURRENT_DATE + TIME '11:45')::timestamp,
        'Dashboard seed: turno manhã em andamento',
        'in_progress'
      ),
      (
        'dashboard-today-afternoon-confirmed',
        (CURRENT_DATE + TIME '14:00')::timestamp,
        (CURRENT_DATE + TIME '15:00')::timestamp,
        'Dashboard seed: turno tarde confirmado',
        'confirmed'
      ),
      (
        'dashboard-today-afternoon-finished',
        (CURRENT_DATE + TIME '15:15')::timestamp,
        (CURRENT_DATE + TIME '16:15')::timestamp,
        'Dashboard seed: turno tarde finalizado',
        'finished'
      ),
      (
        'dashboard-current-month-week-1',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:30')::timestamp,
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '10:30')::timestamp,
        'Dashboard seed: desempenho mês atual semana 1',
        'delivered'
      ),
      (
        'dashboard-current-month-week-2',
        (((date_trunc('month', CURRENT_DATE)::date) + 7) + TIME '13:30')::timestamp,
        (((date_trunc('month', CURRENT_DATE)::date) + 7) + TIME '14:30')::timestamp,
        'Dashboard seed: desempenho mês atual semana 2',
        'confirmed'
      ),
      (
        'dashboard-current-month-week-3',
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '16:00')::timestamp,
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '17:00')::timestamp,
        'Dashboard seed: desempenho mês atual semana 3',
        'finished'
      ),
      (
        'dashboard-previous-month-week-1',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '09:00')::timestamp,
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '10:00')::timestamp,
        'Dashboard seed: desempenho mês anterior semana 1',
        'delivered'
      ),
      (
        'dashboard-previous-month-week-2',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 6) + TIME '14:00')::timestamp,
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 6) + TIME '15:00')::timestamp,
        'Dashboard seed: desempenho mês anterior semana 2',
        'confirmed'
      )
  ) AS scenarios(label, scheduled_local, estimated_end_local, notes, final_status)
)
INSERT INTO schedules (
  company_id,
  client_id,
  pet_id,
  scheduled_at,
  estimated_end,
  notes,
  created_by
)
SELECT
  dc.id,
  sc.id,
  sp.id,
  ds.scheduled_local AT TIME ZONE 'America/Sao_Paulo',
  ds.estimated_end_local AT TIME ZONE 'America/Sao_Paulo',
  ds.notes,
  sa.id
FROM dev_company dc
CROSS JOIN seeded_client sc
CROSS JOIN seeded_pet sp
CROSS JOIN seeded_admin sa
CROSS JOIN dashboard_schedule_scenarios ds
WHERE NOT EXISTS (
  SELECT 1
  FROM schedules s
  WHERE s.company_id = dc.id
    AND s.client_id = sc.id
    AND s.pet_id = sp.id
    AND s.scheduled_at = ds.scheduled_local AT TIME ZONE 'America/Sao_Paulo'
    AND s.deleted_at IS NULL
);

WITH dashboard_schedule_scenarios AS (
  SELECT *
  FROM (
    VALUES
      (
        'dashboard-yesterday-confirmed',
        ((CURRENT_DATE - 1) + TIME '10:00')::timestamp,
        'confirmed',
        ((CURRENT_DATE - 1) + TIME '09:30')::timestamp,
        'Aguardando confirmação automática do seed'
      ),
      (
        'dashboard-yesterday-confirmed',
        ((CURRENT_DATE - 1) + TIME '10:00')::timestamp,
        'confirmed',
        ((CURRENT_DATE - 1) + TIME '09:50')::timestamp,
        'Confirmado para alimentar comparativo diário'
      ),
      (
        'dashboard-today-morning-waiting',
        (CURRENT_DATE + TIME '09:00')::timestamp,
        'waiting',
        (CURRENT_DATE + TIME '08:45')::timestamp,
        'Aguardando início do atendimento'
      ),
      (
        'dashboard-today-morning-in-progress',
        (CURRENT_DATE + TIME '10:30')::timestamp,
        'waiting',
        (CURRENT_DATE + TIME '10:15')::timestamp,
        'Entrada na fila do turno da manhã'
      ),
      (
        'dashboard-today-morning-in-progress',
        (CURRENT_DATE + TIME '10:30')::timestamp,
        'confirmed',
        (CURRENT_DATE + TIME '10:25')::timestamp,
        'Confirmado no check-in local'
      ),
      (
        'dashboard-today-morning-in-progress',
        (CURRENT_DATE + TIME '10:30')::timestamp,
        'in_progress',
        (CURRENT_DATE + TIME '10:40')::timestamp,
        'Atendimento iniciado para o dashboard'
      ),
      (
        'dashboard-today-afternoon-confirmed',
        (CURRENT_DATE + TIME '14:00')::timestamp,
        'waiting',
        (CURRENT_DATE + TIME '13:30')::timestamp,
        'Aguardando atendimento do turno da tarde'
      ),
      (
        'dashboard-today-afternoon-confirmed',
        (CURRENT_DATE + TIME '14:00')::timestamp,
        'confirmed',
        (CURRENT_DATE + TIME '13:50')::timestamp,
        'Confirmado para o turno da tarde'
      ),
      (
        'dashboard-today-afternoon-finished',
        (CURRENT_DATE + TIME '15:15')::timestamp,
        'waiting',
        (CURRENT_DATE + TIME '14:55')::timestamp,
        'Entrada na fila do banho da tarde'
      ),
      (
        'dashboard-today-afternoon-finished',
        (CURRENT_DATE + TIME '15:15')::timestamp,
        'confirmed',
        (CURRENT_DATE + TIME '15:05')::timestamp,
        'Confirmado no balcão'
      ),
      (
        'dashboard-today-afternoon-finished',
        (CURRENT_DATE + TIME '15:15')::timestamp,
        'in_progress',
        (CURRENT_DATE + TIME '15:20')::timestamp,
        'Execução iniciada'
      ),
      (
        'dashboard-today-afternoon-finished',
        (CURRENT_DATE + TIME '15:15')::timestamp,
        'finished',
        (CURRENT_DATE + TIME '16:10')::timestamp,
        'Concluído para alimentar duração final'
      ),
      (
        'dashboard-current-month-week-1',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:30')::timestamp,
        'waiting',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:00')::timestamp,
        'Aguardando atendimento da primeira semana'
      ),
      (
        'dashboard-current-month-week-1',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:30')::timestamp,
        'confirmed',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:20')::timestamp,
        'Confirmado na primeira semana'
      ),
      (
        'dashboard-current-month-week-1',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:30')::timestamp,
        'finished',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '10:20')::timestamp,
        'Finalizado na primeira semana'
      ),
      (
        'dashboard-current-month-week-1',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '09:30')::timestamp,
        'delivered',
        (((date_trunc('month', CURRENT_DATE)::date) + 2) + TIME '10:35')::timestamp,
        'Entregue na primeira semana'
      ),
      (
        'dashboard-current-month-week-2',
        (((date_trunc('month', CURRENT_DATE)::date) + 7) + TIME '13:30')::timestamp,
        'confirmed',
        (((date_trunc('month', CURRENT_DATE)::date) + 7) + TIME '13:10')::timestamp,
        'Confirmado na segunda semana'
      ),
      (
        'dashboard-current-month-week-3',
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '16:00')::timestamp,
        'confirmed',
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '15:40')::timestamp,
        'Confirmado na terceira semana'
      ),
      (
        'dashboard-current-month-week-3',
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '16:00')::timestamp,
        'finished',
        (((date_trunc('month', CURRENT_DATE)::date) + 11) + TIME '16:50')::timestamp,
        'Finalizado na terceira semana'
      ),
      (
        'dashboard-previous-month-week-1',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '09:00')::timestamp,
        'confirmed',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '08:45')::timestamp,
        'Confirmado no mês anterior semana 1'
      ),
      (
        'dashboard-previous-month-week-1',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '09:00')::timestamp,
        'finished',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '09:50')::timestamp,
        'Finalizado no mês anterior semana 1'
      ),
      (
        'dashboard-previous-month-week-1',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '09:00')::timestamp,
        'delivered',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 2) + TIME '10:05')::timestamp,
        'Entregue no mês anterior semana 1'
      ),
      (
        'dashboard-previous-month-week-2',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 6) + TIME '14:00')::timestamp,
        'confirmed',
        (((date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date) + 6) + TIME '13:45')::timestamp,
        'Confirmado no mês anterior semana 2'
      )
  ) AS history(label, scheduled_local, status, changed_at_local, notes)
), seeded_schedules AS (
  SELECT s.id, dss.status, dss.changed_at_local, dss.notes
  FROM schedules s
  INNER JOIN companies c ON c.id = s.company_id
  INNER JOIN clients cl ON cl.id = s.client_id
  INNER JOIN people_identifications pi ON pi.person_id = cl.person_id
  INNER JOIN pets p ON p.id = s.pet_id
  INNER JOIN dashboard_schedule_scenarios dss
    ON dss.scheduled_local AT TIME ZONE 'America/Sao_Paulo' = s.scheduled_at
  WHERE c.slug = 'petcontrol-dev'
    AND pi.cpf = '12345678901'
    AND p.name = 'Thor'
    AND s.deleted_at IS NULL
), seeded_admin AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
)
INSERT INTO schedule_status_history (
  schedule_id,
  status,
  changed_at,
  changed_by,
  notes
)
SELECT
  ss.id,
  ss.status::schedule_status,
  ss.changed_at_local AT TIME ZONE 'America/Sao_Paulo',
  sa.id,
  ss.notes
FROM seeded_schedules ss
CROSS JOIN seeded_admin sa
WHERE NOT EXISTS (
  SELECT 1
  FROM schedule_status_history ssh
  WHERE ssh.schedule_id = ss.id
    AND ssh.status = ss.status::schedule_status
    AND ssh.changed_at = ss.changed_at_local AT TIME ZONE 'America/Sao_Paulo'
);
SQL
