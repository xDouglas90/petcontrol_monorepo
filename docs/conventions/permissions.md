> *(Em Construção - 17/04/2026)*

# Permissões

## Índice

- [Permissões](#permissões)
  - [Índice](#índice)
  - [Visão geral](#visão-geral)
  - [Fluxo](#fluxo)
    - [Permissions](#permissions)
    - [Users](#users)
    - [Plans](#plans)
  - [Permissões por módulo](#permissões-por-módulo)
    - [Módulo de Configurações da Empresa](#módulo-de-configurações-da-empresa)
    - [Módulo de Pessoas](#módulo-de-pessoas)
    - [Módulo de Usuários](#módulo-de-usuários)
    - [Módulo de Clientes](#módulo-de-clientes)
    - [Módulo de Pets](#módulo-de-pets)
    - [Módulo de Agendamentos](#módulo-de-agendamentos)
    - [Módulo de Produtos](#módulo-de-produtos)
    - [Módulo de Serviços](#módulo-de-serviços)
    - [Módulo de Logs](#módulo-de-logs)
    - [Módulo de Planos](#módulo-de-planos)
    - [Módulo de Relatórios](#módulo-de-relatórios)
    - [Módulo de Financeiro](#módulo-de-financeiro)
    - [Módulo de Fornecedores](#módulo-de-fornecedores)
    - [Módulo de Acesso de Usuários Externos](#módulo-de-acesso-de-usuários-externos)
    - [Módulo de Estoque](#módulo-de-estoque)
    - [Módulo de Creche de Pets](#módulo-de-creche-de-pets)
    - [Módulo de Hotel de Pets](#módulo-de-hotel-de-pets)
    - [Módulo de Chat](#módulo-de-chat)
    - [Módulo de Notificações](#módulo-de-notificações)
    - [Módulo de Tele-busca/Entrega de Pets](#módulo-de-tele-buscaentrega-de-pets)
    - [Módulo de Logs de Auditoria](#módulo-de-logs-de-auditoria)
    - [Módulo de Logs de Autenticação](#módulo-de-logs-de-autenticação)

## Visão geral

As permissões são divididas em módulos e ações. Cada módulo representa uma área do sistema e cada ação representa uma operação que pode ser realizada dentro do módulo.

## Fluxo

### Permissions

- Ao criar uma permissão, deve-se adicionar a permissão na tabela `permissions`.
- Ao atualizar uma permissão, se o item `default_role` for alterado, os usuários já criados não devem ter suas permissões alteradas na tabela `user_permissions`. Deve-se enviar uma notificação de nível `alert` para os usuários do tipo `admin` e `root`.

### Users

- Novos usuários (`users`) recebem as permissões padrão de acordo com o tipo de usuário (`role`), adicionando cada permissão na tabela `user_permissions`.
- A permissão padrão é verificada em runtime.
- É possível alterar as permissões de um usuário a qualquer momento, sempre criando uma notificação de nível `info` no sistema para o usuário em questão.

### Plans

- Ao criar um plano, deve-se adicionar as permissões referentes aos módulos que o plano inclui em `plan_permissions`.
  - Por exemplo: se o plano inclui o módulo de usuários, deve-se adicionar as permissões `users:create`, `users:view`, `users:update`, `users:delete`, `users:block`, `users:unblock` em `plan_permissions`.
- Caso o tenant altere seu plano, todos os seus usuários (`company_users`) devem ter suas permissões alteradas na tabela `user_permissions` conforme seus tipos (`role`) e o tipo padrão das permissões (`default_role`), devendo-se enviar uma notificação de nível `info` no sistema para os usuários em questão.

## Permissões por módulo

### Módulo de Configurações da Empresa

| Description                          | Name                         | Default Roles     |
| ------------------------------------ | ---------------------------- | ----------------- |
| Editar configurações de negócios     | `company_settings:edit`      | ['root', 'admin'] |
| Editar configurações de plano        | `plan_settings:edit`         | ['root', 'admin'] |
| Editar configurações de pagamento    | `payment_settings:edit`      | ['root', 'admin'] |
| Editar configurações de notificações | `notification_settings:edit` | ['root', 'admin'] |
| Editar configurações de integração   | `integration_settings:edit`  | ['root', 'admin'] |
| Editar configurações de segurança    | `security_settings:edit`     | ['root', 'admin'] |

### Módulo de Pessoas

| Description       | Name                | Default Roles               |
| ----------------- | ------------------- | --------------------------- |
| Criar pessoa      | `people:create`     | ['root', 'admin', 'system'] |
| Visualizar pessoa | `people:view`       | ['root', 'admin', 'system'] |
| Atualizar pessoa  | `people:update`     | ['root', 'admin', 'system'] |
| Deletar pessoa    | `people:delete`     | ['root', 'admin', 'system'] |
| Restaurar pessoa  | `people:restore`    | ['root', 'admin', 'system'] |
| Desativar pessoa  | `people:deactivate` | ['root', 'admin', 'system'] |
| Reativar pessoa   | `people:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Usuários

| Description         | Name            | Default Roles                 |
| ------------------- | --------------- | ----------------------------- |
| Criar usuário       | `users:create`  | ['root', 'internal', 'admin'] |
| Visualizar usuário  | `users:view`    | ['root', 'internal', 'admin'] |
| Atualizar usuário   | `users:update`  | ['root', 'internal', 'admin'] |
| Deletar usuário     | `users:delete`  | ['root', 'internal', 'admin'] |
| Restaurar usuário   | `users:restore` | ['root', 'internal', 'admin'] |
| Bloquear usuário    | `users:block`   | ['root', 'internal', 'admin'] |
| Desbloquear usuário | `users:unblock` | ['root', 'internal', 'admin'] |

### Módulo de Clientes

| Description        | Name                 | Default Roles               |
| ------------------ | -------------------- | --------------------------- |
| Criar cliente      | `clients:create`     | ['root', 'admin', 'system'] |
| Visualizar cliente | `clients:view`       | ['root', 'admin', 'system'] |
| Atualizar cliente  | `clients:update`     | ['root', 'admin', 'system'] |
| Deletar cliente    | `clients:delete`     | ['root', 'admin', 'system'] |
| Restaurar cliente  | `clients:restore`    | ['root', 'admin', 'system'] |
| Desativar cliente  | `clients:deactivate` | ['root', 'admin', 'system'] |
| Reativar cliente   | `clients:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Pets

| Description    | Name              | Default Roles                         |
| -------------- | ----------------- | ------------------------------------- |
| Criar pet      | `pets:create`     | ['root', 'admin', 'system', 'common'] |
| Visualizar pet | `pets:view`       | ['root', 'admin', 'system', 'common'] |
| Atualizar pet  | `pets:update`     | ['root', 'admin', 'system', 'common'] |
| Deletar pet    | `pets:delete`     | ['root', 'admin', 'system', 'common'] |
| Desativar pet  | `pets:deactivate` | ['root', 'admin', 'system', 'common'] |
| Reativar pet   | `pets:reactivate` | ['root', 'admin', 'system', 'common'] |

### Módulo de Agendamentos

| Description            | Name                   | Default Roles                         |
| ---------------------- | ---------------------- | ------------------------------------- |
| Criar agendamento      | `schedules:create`     | ['root', 'admin', 'system', 'common'] |
| Visualizar agendamento | `schedules:view`       | ['root', 'admin', 'system', 'common'] |
| Atualizar agendamento  | `schedules:update`     | ['root', 'admin', 'system', 'common'] |
| Deletar agendamento    | `schedules:delete`     | ['root', 'admin', 'system', 'common'] |
| Desativar agendamento  | `schedules:deactivate` | ['root', 'admin', 'system', 'common'] |
| Reativar agendamento   | `schedules:reactivate` | ['root', 'admin', 'system', 'common'] |

### Módulo de Produtos

| Description        | Name                  | Default Roles               |
| ------------------ | --------------------- | --------------------------- |
| Criar produto      | `products:create`     | ['root', 'admin', 'system'] |
| Visualizar produto | `products:view`       | ['root', 'admin', 'system'] |
| Atualizar produto  | `products:update`     | ['root', 'admin', 'system'] |
| Deletar produto    | `products:delete`     | ['root', 'admin', 'system'] |
| Desativar produto  | `products:deactivate` | ['root', 'admin', 'system'] |
| Reativar produto   | `products:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Serviços

| Description        | Name                  | Default Roles                         |
| ------------------ | --------------------- | ------------------------------------- |
| Criar serviço      | `services:create`     | ['root', 'admin', 'system']           |
| Visualizar serviço | `services:view`       | ['root', 'admin', 'system', 'common'] |
| Atualizar serviço  | `services:update`     | ['root', 'admin', 'system']           |
| Deletar serviço    | `services:delete`     | ['root', 'admin', 'system']           |
| Desativar serviço  | `services:deactivate` | ['root', 'admin', 'system']           |
| Reativar serviço   | `services:reactivate` | ['root', 'admin', 'system']           |

### Módulo de Logs

| Description     | Name        | Default Roles     |
| --------------- | ----------- | ----------------- |
| Visualizar logs | `logs:view` | ['root', 'admin'] |

### Módulo de Planos

| Description      | Name               | Default Roles |
| ---------------- | ------------------ | ------------- |
| Criar plano      | `plans:create`     | ['root']      |
| Visualizar plano | `plans:view`       | ['root']      |
| Atualizar plano  | `plans:update`     | ['root']      |
| Deletar plano    | `plans:delete`     | ['root']      |
| Restaurar plano  | `plans:restore`    | ['root']      |
| Desativar plano  | `plans:deactivate` | ['root']      |
| Reativar plano   | `plans:reactivate` | ['root']      |

### Módulo de Relatórios

| Description           | Name                 | Default Roles     |
| --------------------- | -------------------- | ----------------- |
| Criar relatório       | `reports:create`     | ['root']          |
| Visualizar relatórios | `reports:view`       | ['root', 'admin'] |
| Atualizar relatório   | `reports:update`     | ['root']          |
| Deletar relatório     | `reports:delete`     | ['root']          |
| Restaurar relatório   | `reports:restore`    | ['root']          |
| Desativar relatório   | `reports:deactivate` | ['root']          |
| Reativar relatório    | `reports:reactivate` | ['root']          |

### Módulo de Financeiro

| Description          | Name                  | Default Roles     |
| -------------------- | --------------------- | ----------------- |
| Criar transação      | `finances:create`     | ['root', 'admin'] |
| Visualizar transação | `finances:view`       | ['root', 'admin'] |
| Atualizar transação  | `finances:update`     | ['root', 'admin'] |
| Deletar transação    | `finances:delete`     | ['root', 'admin'] |
| Restaurar transação  | `finances:restore`    | ['root', 'admin'] |
| Desativar transação  | `finances:deactivate` | ['root', 'admin'] |
| Reativar transação   | `finances:reactivate` | ['root', 'admin'] |

### Módulo de Fornecedores

| Description           | Name                   | Default Roles               |
| --------------------- | ---------------------- | --------------------------- |
| Criar fornecedor      | `suppliers:create`     | ['root', 'admin', 'system'] |
| Visualizar fornecedor | `suppliers:view`       | ['root', 'admin', 'system'] |
| Atualizar fornecedor  | `suppliers:update`     | ['root', 'admin', 'system'] |
| Deletar fornecedor    | `suppliers:delete`     | ['root', 'admin', 'system'] |
| Restaurar fornecedor  | `suppliers:restore`    | ['root', 'admin', 'system'] |
| Desativar fornecedor  | `suppliers:deactivate` | ['root', 'admin', 'system'] |
| Reativar fornecedor   | `suppliers:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Acesso de Usuários Externos

| Description       | Name                         | Default Roles               |
| ----------------- | ---------------------------- | --------------------------- |
| Criar acesso      | `external_access:create`     | ['root', 'admin', 'system'] |
| Visualizar acesso | `external_access:view`       | ['root', 'admin', 'system'] |
| Atualizar acesso  | `external_access:update`     | ['root', 'admin', 'system'] |
| Deletar acesso    | `external_access:delete`     | ['root', 'admin', 'system'] |
| Desativar acesso  | `external_access:deactivate` | ['root', 'admin', 'system'] |
| Reativar acesso   | `external_access:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Estoque

| Description        | Name               | Default Roles               |
| ------------------ | ------------------ | --------------------------- |
| Criar estoque      | `stock:create`     | ['root', 'admin', 'system'] |
| Visualizar estoque | `stock:view`       | ['root', 'admin', 'system'] |
| Atualizar estoque  | `stock:update`     | ['root', 'admin', 'system'] |
| Deletar estoque    | `stock:delete`     | ['root', 'admin', 'system'] |
| Desativar estoque  | `stock:deactivate` | ['root', 'admin', 'system'] |
| Reativar estoque   | `stock:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Creche de Pets

| Description       | Name                 | Default Roles               |
| ----------------- | -------------------- | --------------------------- |
| Criar creche      | `daycare:create`     | ['root', 'admin', 'system'] |
| Visualizar creche | `daycare:view`       | ['root', 'admin', 'system'] |
| Atualizar creche  | `daycare:update`     | ['root', 'admin', 'system'] |
| Deletar creche    | `daycare:delete`     | ['root', 'admin', 'system'] |
| Restaurar creche  | `daycare:restore`    | ['root', 'admin', 'system'] |
| Desativar creche  | `daycare:deactivate` | ['root', 'admin', 'system'] |
| Reativar creche   | `daycare:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Hotel de Pets

| Description      | Name               | Default Roles               |
| ---------------- | ------------------ | --------------------------- |
| Criar hotel      | `hotel:create`     | ['root', 'admin', 'system'] |
| Visualizar hotel | `hotel:view`       | ['root', 'admin', 'system'] |
| Atualizar hotel  | `hotel:update`     | ['root', 'admin', 'system'] |
| Deletar hotel    | `hotel:delete`     | ['root', 'admin', 'system'] |
| Restaurar hotel  | `hotel:restore`    | ['root', 'admin', 'system'] |
| Desativar hotel  | `hotel:deactivate` | ['root', 'admin', 'system'] |
| Reativar hotel   | `hotel:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Chat

| Description     | Name              | Default Roles               |
| --------------- | ----------------- | --------------------------- |
| Criar chat      | `chat:create`     | ['root', 'admin', 'system'] |
| Visualizar chat | `chat:view`       | ['root', 'admin', 'system'] |
| Atualizar chat  | `chat:update`     | ['root', 'admin', 'system'] |
| Deletar chat    | `chat:delete`     | ['root', 'admin', 'system'] |
| Restaurar chat  | `chat:restore`    | ['root', 'admin', 'system'] |
| Desativar chat  | `chat:deactivate` | ['root', 'admin', 'system'] |
| Reativar chat   | `chat:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Notificações

| Description            | Name                       | Default Roles               |
| ---------------------- | -------------------------- | --------------------------- |
| Criar notificação      | `notifications:create`     | ['root', 'admin', 'system'] |
| Visualizar notificação | `notifications:view`       | ['root', 'admin', 'system'] |
| Atualizar notificação  | `notifications:update`     | ['root', 'admin', 'system'] |
| Deletar notificação    | `notifications:delete`     | ['root', 'admin', 'system'] |
| Restaurar notificação  | `notifications:restore`    | ['root', 'admin', 'system'] |
| Desativar notificação  | `notifications:deactivate` | ['root', 'admin', 'system'] |
| Reativar notificação   | `notifications:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Tele-busca/Entrega de Pets

| Description                      | Name                         | Default Roles               |
| -------------------------------- | ---------------------------- | --------------------------- |
| Criar serviço de tele-busca      | `pickup_delivery:create`     | ['root', 'admin', 'system'] |
| Visualizar serviço de tele-busca | `pickup_delivery:view`       | ['root', 'admin', 'system'] |
| Atualizar serviço de tele-busca  | `pickup_delivery:update`     | ['root', 'admin', 'system'] |
| Deletar serviço de tele-busca    | `pickup_delivery:delete`     | ['root', 'admin', 'system'] |
| Restaurar serviço de tele-busca  | `pickup_delivery:restore`    | ['root', 'admin', 'system'] |
| Desativar serviço de tele-busca  | `pickup_delivery:deactivate` | ['root', 'admin', 'system'] |
| Reativar serviço de tele-busca   | `pickup_delivery:reactivate` | ['root', 'admin', 'system'] |

### Módulo de Logs de Auditoria

| Description     | Name        | Default Roles     |
| --------------- | ----------- | ----------------- |
| Visualizar logs | `logs:view` | ['root', 'admin'] |

### Módulo de Logs de Autenticação

| Description     | Name        | Default Roles     |
| --------------- | ----------- | ----------------- |
| Visualizar logs | `logs:view` | ['root', 'admin'] |
