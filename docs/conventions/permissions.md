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
| Criar pessoa      | `people:create`     | ['root', 'admin'] |
| Visualizar pessoa | `people:view`       | ['root', 'admin'] |
| Atualizar pessoa  | `people:update`     | ['root', 'admin'] |
| Deletar pessoa    | `people:delete`     | ['root', 'admin'] |
| Restaurar pessoa  | `people:restore`    | ['root', 'admin'] |
| Desativar pessoa  | `people:deactivate` | ['root', 'admin'] |
| Reativar pessoa   | `people:reactivate` | ['root', 'admin'] |

### Módulo de Usuários

| Description         | Name            | Default Roles                 |
| ------------------- | --------------- | ----------------------------- |
| Criar usuário       | `users:create`  | ['root', 'internal', 'admin'] |
| Visualizar usuário  | `users:view`    | ['root', 'internal', 'admin', 'system'] |
| Atualizar usuário   | `users:update`  | ['root', 'internal', 'admin'] |
| Deletar usuário     | `users:delete`  | ['root', 'internal', 'admin'] |
| Restaurar usuário   | `users:restore` | ['root', 'internal', 'admin'] |
| Bloquear usuário    | `users:block`   | ['root', 'internal', 'admin'] |
| Desbloquear usuário | `users:unblock` | ['root', 'internal', 'admin'] |

### Módulo de Clientes

| Description        | Name                 | Default Roles               |
| ------------------ | -------------------- | --------------------------- |
| Criar cliente      | `clients:create`     | ['root', 'admin'] |
| Visualizar cliente | `clients:view`       | ['root', 'admin'] |
| Atualizar cliente  | `clients:update`     | ['root', 'admin'] |
| Deletar cliente    | `clients:delete`     | ['root', 'admin'] |
| Restaurar cliente  | `clients:restore`    | ['root', 'admin'] |
| Desativar cliente  | `clients:deactivate` | ['root', 'admin'] |
| Reativar cliente   | `clients:reactivate` | ['root', 'admin'] |

### Módulo de Pets

| Description    | Name              | Default Roles                         |
| -------------- | ----------------- | ------------------------------------- |
| Criar pet      | `pets:create`     | ['root', 'admin', 'common'] |
| Visualizar pet | `pets:view`       | ['root', 'admin', 'common'] |
| Atualizar pet  | `pets:update`     | ['root', 'admin', 'common'] |
| Deletar pet    | `pets:delete`     | ['root', 'admin', 'common'] |
| Desativar pet  | `pets:deactivate` | ['root', 'admin', 'common'] |
| Reativar pet   | `pets:reactivate` | ['root', 'admin', 'common'] |

### Módulo de Agendamentos

| Description            | Name                   | Default Roles                         |
| ---------------------- | ---------------------- | ------------------------------------- |
| Criar agendamento      | `schedules:create`     | ['root', 'admin', 'common'] |
| Visualizar agendamento | `schedules:view`       | ['root', 'admin', 'common'] |
| Atualizar agendamento  | `schedules:update`     | ['root', 'admin', 'common'] |
| Deletar agendamento    | `schedules:delete`     | ['root', 'admin', 'common'] |
| Desativar agendamento  | `schedules:deactivate` | ['root', 'admin', 'common'] |
| Reativar agendamento   | `schedules:reactivate` | ['root', 'admin', 'common'] |

### Módulo de Produtos

| Description        | Name                  | Default Roles               |
| ------------------ | --------------------- | --------------------------- |
| Criar produto      | `products:create`     | ['root', 'admin'] |
| Visualizar produto | `products:view`       | ['root', 'admin'] |
| Atualizar produto  | `products:update`     | ['root', 'admin'] |
| Deletar produto    | `products:delete`     | ['root', 'admin'] |
| Desativar produto  | `products:deactivate` | ['root', 'admin'] |
| Reativar produto   | `products:reactivate` | ['root', 'admin'] |

### Módulo de Serviços

| Description        | Name                  | Default Roles                         |
| ------------------ | --------------------- | ------------------------------------- |
| Criar serviço      | `services:create`     | ['root', 'admin']           |
| Visualizar serviço | `services:view`       | ['root', 'admin', 'system', 'common'] |
| Atualizar serviço  | `services:update`     | ['root', 'admin']           |
| Deletar serviço    | `services:delete`     | ['root', 'admin']           |
| Desativar serviço  | `services:deactivate` | ['root', 'admin']           |
| Reativar serviço   | `services:reactivate` | ['root', 'admin']           |

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
| Criar fornecedor      | `suppliers:create`     | ['root', 'admin'] |
| Visualizar fornecedor | `suppliers:view`       | ['root', 'admin'] |
| Atualizar fornecedor  | `suppliers:update`     | ['root', 'admin'] |
| Deletar fornecedor    | `suppliers:delete`     | ['root', 'admin'] |
| Restaurar fornecedor  | `suppliers:restore`    | ['root', 'admin'] |
| Desativar fornecedor  | `suppliers:deactivate` | ['root', 'admin'] |
| Reativar fornecedor   | `suppliers:reactivate` | ['root', 'admin'] |

### Módulo de Acesso de Usuários Externos

| Description       | Name                         | Default Roles               |
| ----------------- | ---------------------------- | --------------------------- |
| Criar acesso      | `external_access:create`     | ['root', 'admin'] |
| Visualizar acesso | `external_access:view`       | ['root', 'admin'] |
| Atualizar acesso  | `external_access:update`     | ['root', 'admin'] |
| Deletar acesso    | `external_access:delete`     | ['root', 'admin'] |
| Desativar acesso  | `external_access:deactivate` | ['root', 'admin'] |
| Reativar acesso   | `external_access:reactivate` | ['root', 'admin'] |

### Módulo de Estoque

| Description        | Name               | Default Roles               |
| ------------------ | ------------------ | --------------------------- |
| Criar estoque      | `stock:create`     | ['root', 'admin'] |
| Visualizar estoque | `stock:view`       | ['root', 'admin'] |
| Atualizar estoque  | `stock:update`     | ['root', 'admin'] |
| Deletar estoque    | `stock:delete`     | ['root', 'admin'] |
| Desativar estoque  | `stock:deactivate` | ['root', 'admin'] |
| Reativar estoque   | `stock:reactivate` | ['root', 'admin'] |

### Módulo de Creche de Pets

| Description       | Name                 | Default Roles               |
| ----------------- | -------------------- | --------------------------- |
| Criar creche      | `daycare:create`     | ['root', 'admin'] |
| Visualizar creche | `daycare:view`       | ['root', 'admin'] |
| Atualizar creche  | `daycare:update`     | ['root', 'admin'] |
| Deletar creche    | `daycare:delete`     | ['root', 'admin'] |
| Restaurar creche  | `daycare:restore`    | ['root', 'admin'] |
| Desativar creche  | `daycare:deactivate` | ['root', 'admin'] |
| Reativar creche   | `daycare:reactivate` | ['root', 'admin'] |

### Módulo de Hotel de Pets

| Description      | Name               | Default Roles               |
| ---------------- | ------------------ | --------------------------- |
| Criar hotel      | `hotel:create`     | ['root', 'admin'] |
| Visualizar hotel | `hotel:view`       | ['root', 'admin'] |
| Atualizar hotel  | `hotel:update`     | ['root', 'admin'] |
| Deletar hotel    | `hotel:delete`     | ['root', 'admin'] |
| Restaurar hotel  | `hotel:restore`    | ['root', 'admin'] |
| Desativar hotel  | `hotel:deactivate` | ['root', 'admin'] |
| Reativar hotel   | `hotel:reactivate` | ['root', 'admin'] |

### Módulo de Chat

| Description     | Name              | Default Roles               |
| --------------- | ----------------- | --------------------------- |
| Criar chat      | `chat:create`     | ['root', 'admin'] |
| Visualizar chat | `chat:view`       | ['root', 'admin'] |
| Atualizar chat  | `chat:update`     | ['root', 'admin'] |
| Deletar chat    | `chat:delete`     | ['root', 'admin'] |
| Restaurar chat  | `chat:restore`    | ['root', 'admin'] |
| Desativar chat  | `chat:deactivate` | ['root', 'admin'] |
| Reativar chat   | `chat:reactivate` | ['root', 'admin'] |

### Módulo de Notificações

| Description            | Name                       | Default Roles               |
| ---------------------- | -------------------------- | --------------------------- |
| Criar notificação      | `notifications:create`     | ['root', 'admin'] |
| Visualizar notificação | `notifications:view`       | ['root', 'admin'] |
| Atualizar notificação  | `notifications:update`     | ['root', 'admin'] |
| Deletar notificação    | `notifications:delete`     | ['root', 'admin'] |
| Restaurar notificação  | `notifications:restore`    | ['root', 'admin'] |
| Desativar notificação  | `notifications:deactivate` | ['root', 'admin'] |
| Reativar notificação   | `notifications:reactivate` | ['root', 'admin'] |

### Módulo de Tele-busca/Entrega de Pets

| Description                      | Name                         | Default Roles               |
| -------------------------------- | ---------------------------- | --------------------------- |
| Criar serviço de tele-busca      | `pickup_delivery:create`     | ['root', 'admin'] |
| Visualizar serviço de tele-busca | `pickup_delivery:view`       | ['root', 'admin'] |
| Atualizar serviço de tele-busca  | `pickup_delivery:update`     | ['root', 'admin'] |
| Deletar serviço de tele-busca    | `pickup_delivery:delete`     | ['root', 'admin'] |
| Restaurar serviço de tele-busca  | `pickup_delivery:restore`    | ['root', 'admin'] |
| Desativar serviço de tele-busca  | `pickup_delivery:deactivate` | ['root', 'admin'] |
| Reativar serviço de tele-busca   | `pickup_delivery:reactivate` | ['root', 'admin'] |

### Módulo de Logs de Auditoria

| Description     | Name        | Default Roles     |
| --------------- | ----------- | ----------------- |
| Visualizar logs | `logs:view` | ['root', 'admin'] |

### Módulo de Logs de Autenticação

| Description     | Name        | Default Roles     |
| --------------- | ----------- | ----------------- |
| Visualizar logs | `logs:view` | ['root', 'admin'] |
