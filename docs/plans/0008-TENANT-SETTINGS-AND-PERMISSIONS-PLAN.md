# Plano de Ação e Execução - Tela de Configurações do Tenant e Gestão de Permissões

## Objetivo

Definir o escopo técnico, funcional e arquitetural da próxima PR do PetControl para implementar a tela de `Configurações` da aplicação Web, conectada aos dados reais do tenant e com gestão de permissões por usuário vinculada à empresa corrente.

Esta PR deve introduzir:

- tela real de configurações em `/:companySlug/settings`;
- três seções explícitas na tela:
  - `Configurações da empresa`;
  - `Configurações de negócios`;
  - `Permissões`;
- carregamento e edição das configurações da empresa com base na tabela `companies`;
- carregamento e edição das configurações de negócios com base na tabela `company_system_configs`;
- visibilidade da tela apenas para `admin` e para usuários do tipo `system` que tenham recebido permissão explícita de algum `admin`;
- comportamento readonly para usuários sem permissão de edição;
- seção administrativa para gestão de permissões dos usuários vinculados ao tenant, visível apenas para `admin`;
- bootstrap confiável de permissões padrão em `permissions` e `user_permissions` para usuários seedados.

## Contexto Atual

- O Web já possui rota autenticada de configurações, mas a tela atual é apenas placeholder.
- O backend já expõe:
  - `GET /companies/current`;
  - `PATCH /companies/current`;
  - `GET /company-system-configs/current`;
  - `GET /company-users`.
- O schema já possui a tabela `company_system_configs`, com campos suficientes para alimentar uma primeira versão útil da tela.
- O schema já possui as tabelas `permissions` e `user_permissions`.
- O documento [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1) já define a matriz inicial de permissões.
- Ainda não existe:
  - API específica para listar permissões efetivas por usuário do tenant;
  - API para conceder/revogar permissões de usuário;
  - middleware de autorização por `user_permissions` aplicado à tela de configurações;
  - seed consistente de `permissions` e `user_permissions` para usuários já seedados;
  - endpoint de update de `company_system_configs`.

## Referências Obrigatórias

- Schema inicial: [000001_init_schema.up.sql](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/infra/migrations/000001_init_schema.up.sql:1)
- Convenção de permissões: [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1)
- Placeholder atual do Web: [settings/index.tsx](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/apps/web/src/routes/(app)/settings/index.tsx:1)

## Resultado Esperado da PR

Ao final desta PR, o sistema deve apresentar:

- uma tela de configurações funcional, acessível a partir do shell autenticado;
- bloqueio de acesso para qualquer usuário que não seja `admin` nem `system` autorizado;
- visualização em modo readonly para usuários sem permissão de edição;
- leitura e edição dos dados pertinentes à tabela `companies`;
- leitura e edição de todos os itens pertinentes à tabela `company_system_configs`;
- uma área de gestão de permissões por usuário do tenant visível apenas para `admin`;
- todos os usuários seedados com permissões mínimas coerentes com seu `role`;
- base pronta para o futuro módulo de criação de usuários já nascer compatível com `user_permissions`.

## Escopo Funcional Confirmado

## 1. Controle de Acesso à Tela

Regras desta PR:

- usuários com `role = admin` sempre podem acessar a tela;
- usuários com `role = system` só podem acessar a tela se algum `admin` tiver concedido permissão explícita de configurações;
- usuários com `role` diferente de `admin` e `system` não acessam a tela;
- usuários sem permissões de edição devem visualizar os dados em modo readonly;
- a edição de cada seção deve respeitar a permissão correspondente.

Permissões de configurações já documentadas em [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1):

- `company_settings:edit`
- `plan_settings:edit`
- `payment_settings:edit`
- `notification_settings:edit`
- `integration_settings:edit`
- `security_settings:edit`

Decisão prática recomendada para esta PR:

- usar `role` e permissões efetivas para decidir visibilidade da tela;
- usar a permissão específica de cada bloco para habilitar a edição daquele bloco;
- manter a tela em readonly quando nenhuma permissão de edição estiver ativa;
- a seção de gestão de permissões de outros usuários aparece apenas para `admin`.

## 2. Seção `Configurações da Empresa`

Esta seção deve consumir `GET /companies/current` e `PATCH /companies/current`.

Escopo esperado:

- nome da empresa;
- nome fantasia;
- slug, se fizer sentido exibir como somente leitura;
- logo;
- demais campos da tabela `companies` que sejam pertinentes ao tenant e já estejam disponíveis no contrato atual;
- outras informações institucionais e cadastrais da empresa que façam parte da tabela `companies`.

Observação:

- a PR não deve inventar campos sem lastro no backend atual;
- se houver campos de `companies` desejados mas ainda não expostos pelo contrato atual, devem entrar como follow-up explícito;
- esta deve ser a seção responsável por todas as configurações institucionais da empresa.

## 3. Seção `Configurações de Negócios`

Esta seção deve ser baseada diretamente na tabela `company_system_configs` e concentrar todas as configurações operacionais e de negócio do tenant.

Campos confirmados no schema:

- `schedule_init_time`
- `schedule_pause_init_time`
- `schedule_pause_end_time`
- `schedule_end_time`
- `min_schedules_per_day`
- `max_schedules_per_day`
- `schedule_days`
- `dynamic_cages`
- `total_small_cages`
- `total_medium_cages`
- `total_large_cages`
- `total_giant_cages`
- `whatsapp_notifications`
- `whatsapp_conversation`
- `whatsapp_business_phone`

Agrupamento recomendado no Web:

- `Operação`
  - horários de funcionamento;
  - pausa operacional;
  - dias de atendimento;
  - limites mínimo e máximo de agendamentos por dia.
- `Capacidade`
  - `dynamic_cages`;
  - totais por porte de baia/gaiola.
- `WhatsApp`
  - notificações;
  - conversa;
  - telefone comercial.

Requisito técnico adicional:

- criar endpoint de update para `company_system_configs`, reaproveitando a query SQL já gerada para `UpdateCompanySystemConfig`;
- garantir que esta seção cubra todos os itens pertinentes da tabela `company_system_configs`.

## 4. Seção `Permissões`

Esta seção deve existir dentro da própria tela de configurações, mas deve ser renderizada apenas para usuários do tipo `admin`.

Fluxo funcional esperado:

- exibir um `select` com todos os usuários vinculados à empresa corrente;
- ao selecionar um usuário, carregar suas permissões atuais;
- apresentar as permissões de forma agrupada por módulo;
- permitir ao admin ativar ou desativar permissões;
- persistir a alteração com registro de `granted_by`, `revoked_by`, `granted_at` e `revoked_at` conforme a modelagem atual.

Escopo mínimo desta PR:

- garantir no mínimo a gestão das permissões do módulo `Configurações da Empresa`;
- permitir expansão futura para outros módulos sem refatoração estrutural.

Recomendação:

- a API pode já retornar todas as permissões catalogadas do usuário, mas a UI desta PR pode destacar primeiro as permissões de configurações;
- o estado da UI deve deixar claro quais permissões são padrão do papel e quais foram customizadas depois;
- usuários `system` autorizados podem acessar a tela de configurações conforme a regra geral, mas não devem ver este bloco no Web.

## 5. Permissões Básicas Automáticas por Usuário

O documento [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1) define que:

- novos usuários recebem permissões padrão conforme `role`;
- alterações posteriores vivem em `user_permissions`.

Como o módulo de criação de usuários ainda não foi implementado por completo, esta PR deve cobrir os usuários seedados.

Decisão desta PR:

- popular a tabela `permissions` via seed com base no catálogo atual do documento;
- popular `user_permissions` para todos os usuários seedados conforme `default_roles`;
- tornar o seed idempotente;
- garantir que não exista usuário seedado sem ao menos suas permissões básicas.

## Análise do Schema Relevante

## 1. `company_system_configs`

Pontos positivos:

- já há um registro único por `company_id`;
- a modelagem cobre o núcleo operacional da tela;
- a estrutura permite montar uma UX útil sem migração adicional imediata.

Pontos de atenção:

- `schedule_days` usa `week_day[]`, então o Web precisa de mapping claro entre enum técnico e labels amigáveis;
- os campos de horário exigem validação consistente no backend e no formulário;
- `min_schedules_per_day` e `max_schedules_per_day` precisam de regra para impedir inconsistências.

## 2. `user_permissions`

Pontos positivos:

- suporta ativação e revogação sem apagar histórico lógico;
- já registra quem concedeu ou revogou;
- já existe query gerada para insert, bulk insert e listagem por usuário.

Ponto crítico de modelagem:

- a tabela `user_permissions` é ligada a `user_id`, sem `company_id`;
- isso cria risco quando o mesmo usuário estiver vinculado a mais de uma empresa em `company_users`;
- uma alteração feita por um admin de um tenant pode vazar para o contexto de outro tenant do mesmo usuário.

Conclusão recomendada:

- esta PR deve explicitar esse risco como dívida arquitetural;
- se o produto ainda assumir, na prática, um usuário operacional vinculado a apenas um tenant, a PR pode seguir com a modelagem atual;
- se multi-tenant real por usuário já for requisito de curto prazo, a solução correta é revisar a modelagem para escopo por `company_user` ou por `company_id + user_id` antes de expandir gestão fina de permissões.

## Arquitetura Recomendada

## Backend

Adicionar uma trilha explícita para permissões e configurações:

- serviço para leitura e update de `company_system_configs`;
- serviço para leitura de permissões efetivas do usuário;
- serviço para concessão e revogação de permissões;
- middleware para checagem de permissão por código;
- endpoints dedicados para a tela de configurações.

Endpoints recomendados para esta PR:

- `GET /company-system-configs/current`
- `PATCH /company-system-configs/current`
- `GET /company-users`
- `GET /company-users/:user_id/permissions`
- `PATCH /company-users/:user_id/permissions`
- opcionalmente `GET /users/current/permissions` para facilitar gating no Web

Regras de autorização recomendadas:

- `admin` entra sempre na tela;
- `system` entra apenas quando tiver permissão explícita de configurações;
- apenas `admin` ou usuários com permissão apropriada editam um bloco;
- apenas `admin` visualiza a área de gestão de permissões de usuários;
- apenas `admin` pode conceder ou revogar permissões de outros usuários nesta primeira versão.

## Frontend

Separar a tela em blocos independentes:

- `Configurações da empresa`
- `Configurações de negócios`
- `Permissões`

Comportamentos esperados:

- carregamento paralelo dos blocos;
- skeleton ou estado de loading por seção;
- estados de somente leitura quando o usuário puder entrar na tela mas não puder editar certo bloco;
- feedback de sucesso/erro por formulário;
- cache invalidation com React Query após mutações.

## Fases Recomendadas de Execução

## Fase 0 - Fechamento de Escopo e Regras

### 0.1 Ações

- confirmar quais blocos da tela entram nesta PR;
- confirmar que, nesta fase, apenas `admin` gerencia permissões de outros usuários;
- confirmar quais blocos podem ser editados por `system` quando houver permissão correspondente.

### 0.2 Checks

- [ ] Regra de acesso à tela está fechada.
- [ ] Regra de edição por bloco está fechada.
- [ ] Escopo da gestão de permissões desta PR está fechado.

## Fase 1 - Catálogo de Permissões e Seeds

### 1.1 Ações

- criar seed idempotente para `permissions` com base em [permissions.md](/home/trapdev/go/src/github.com/xdouglas90/petcontrol_monorepo/docs/conventions/permissions.md:1);
- criar seed idempotente para `user_permissions` dos usuários já seedados;
- garantir que o seed respeite `default_roles`;
- preparar a mesma regra para reaproveitamento futuro no fluxo de criação de usuário.

### 1.2 Checks

- [ ] Todos os usuários seedados possuem permissões básicas.
- [ ] Nenhuma permissão é duplicada ao rodar seed mais de uma vez.

## Fase 2 - API de Configurações do Tenant

### 2.1 Ações

- implementar `PATCH /company-system-configs/current`;
- adicionar validações de payload e regras de consistência;
- manter `GET /companies/current` e `PATCH /companies/current` como fonte da seção institucional da empresa;
- documentar os contratos novos no swagger.

### 2.2 Checks

- [ ] O tenant consegue ler suas configurações atuais.
- [ ] O tenant consegue salvar mudanças válidas.
- [ ] O backend rejeita payload inconsistente.

## Fase 3 - API de Permissões por Usuário do Tenant

### 3.1 Ações

- criar endpoint para listar permissões de um usuário vinculado à empresa corrente;
- criar endpoint para atualizar permissões de um usuário vinculado à empresa corrente;
- validar que o usuário alvo pertence ao tenant do solicitante;
- registrar `granted_by` e `revoked_by`;
- impedir que usuários sem poder administrativo alterem permissões de terceiros.

### 3.2 Checks

- [ ] O admin consegue consultar permissões de qualquer usuário do tenant.
- [ ] O admin consegue conceder e revogar permissões.
- [ ] Usuários fora do tenant não podem ser alterados.

## Fase 4 - Middleware e Gating de Autorização

### 4.1 Ações

- introduzir middleware ou helper para checar permissão por `code`;
- integrar o gating nas rotas e handlers da área de configurações;
- expor ao Web as permissões efetivas do usuário atual para condicionar UI.

### 4.2 Checks

- [ ] Usuário que não seja `admin` nem `system` autorizado não acessa a tela.
- [ ] Usuário com permissão parcial enxerga apenas o que pode usar.
- [ ] Usuário sem permissão de edição vê a tela em readonly.
- [ ] A autorização não depende apenas do front-end.

## Fase 5 - Tela Web de Configurações

### 5.1 Ações

- substituir o placeholder atual por tela real;
- integrar queries para empresa atual, `company_system_configs`, usuários da empresa e permissões;
- criar formulários por seção;
- adicionar select de usuários na seção de permissões;
- esconder completamente o bloco de gestão de permissões para perfis diferentes de `admin`;
- refletir estado readonly quando a permissão estiver ausente.

### 5.2 Checks

- [ ] A tela carrega dados reais.
- [ ] A edição das configurações salva corretamente.
- [ ] A gestão de permissões funciona para `admin`.
- [ ] Usuários `system` autorizados não visualizam a área de gestão de permissões.

## Fase 6 - Testes e Robustez

### 6.1 Ações

- testes unitários de serviço para leitura e update de `company_system_configs`;
- testes de handler para autorização e update de permissões;
- testes de integração para seed de permissões;
- testes do Web para acesso negado, acesso permitido, edição e alteração de permissões.

### 6.2 Checks

- [ ] Existe cobertura de acesso e edição.
- [ ] Existe cobertura para usuários seedados com permissões padrão.
- [ ] Existe cobertura para tentativas inválidas entre tenants.

## Contrato Inicial de UX da Tela

A tela deve ser organizada explicitamente em três seções principais:

- `Configurações da empresa`
- `Configurações de negócios`
- `Permissões`

## Cabeçalho

- título da página `Configurações`;
- texto de apoio explicando que a tela concentra ajustes do tenant e gestão de acesso.

## Bloco 1 - Configurações da Empresa

- dados institucionais e cadastrais do tenant vindos de `companies`;
- campos como nome, nome fantasia, logo e demais informações pertinentes à empresa;
- ação de salvar separada.

## Bloco 2 - Configurações de Negócios

- todos os itens pertinentes de `company_system_configs`;
- horários, pausa, dias de atendimento;
- limites de agendamentos;
- capacidade física;
- canais de WhatsApp.

## Bloco 3 - Permissões

- visível apenas para `admin`;
- select de usuário do tenant;
- resumo do tipo do usuário;
- lista de permissões organizadas por módulo;
- destaque visual para permissões padrão versus customizadas;
- ação de salvar.

## Riscos e Cuidados

- `user_permissions` hoje não está claramente escopada por tenant;
- permissões do documento ainda estão marcadas como “Em Construção” e podem evoluir;
- o seed precisa ser idempotente para não poluir `user_permissions`;
- o Web não deve assumir que ter acesso à tela implica poder editar tudo;
- o Web não deve assumir que todo usuário autenticado pode visualizar a tela, pois a entrada fica restrita a `admin` e `system` autorizado;
- é importante evitar que o admin revogue de si mesmo o mínimo necessário para operar sem uma regra explícita.

## Decisões Recomendadas Para Manter a PR Saudável

- começar com as permissões do módulo de configurações, sem tentar resolver toda a matriz do sistema na UI;
- manter a gestão de permissões de outros usuários restrita a `admin` nesta primeira versão;
- tratar `company_system_configs` como o núcleo operacional da tela nesta PR;
- registrar explicitamente a dívida de escopo por tenant em `user_permissions`;
- preparar o seed agora para não bloquear o futuro módulo de criação de usuários.

## Ordem Recomendada de Execução

1. Fechar a regra de acesso e o escopo funcional da tela.
2. Seedar `permissions` e `user_permissions` para o ambiente atual.
3. Implementar update de `company_system_configs`.
4. Implementar APIs de leitura e edição de permissões por usuário do tenant.
5. Introduzir middleware e checagem de permissão.
6. Substituir o placeholder do Web pela tela real.
7. Cobrir com testes backend e frontend.

## Resultado Esperado

Se este plano for executado com sucesso, o PetControl deixará de tratar `Configurações` como rota placeholder e passará a oferecer uma área administrativa real do tenant, com edição de parâmetros operacionais e gestão controlada de permissões dos usuários vinculados à empresa, já alinhada ao schema existente e pronta para evoluções futuras.
