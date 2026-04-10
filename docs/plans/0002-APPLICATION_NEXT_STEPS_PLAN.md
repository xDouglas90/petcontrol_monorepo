# Plano de Ação e Execução - Próximos Passos da Aplicação

## Objetivo

Definir a segunda etapa de desenvolvimento do PetControl apos a consolidação da estrutura inicial do monorepo.

Este plano compara:

- o alvo arquitetural descrito no `*README*.md`;
- o que foi registrado e entregue no `docs/plans/0001-INITIAL_STRUCTURE_ACTION_PLAN.md`;
- o estado atual do repositório.

O foco agora deixa de ser "criar a base do monorepo" e passa a ser "evoluir a aplicação para um produto funcional, coerente e expandível".

## Resumo Executivo

### O que ja foi entregue com boa aderência ão `*README*.md`

- Monorepo estruturado com `apps/api`, `apps/web`, `apps/worker`, `libs`, `infra`, `docs` e `.github`.
- API Go executável com `health`, `ready`, SQLC, migrations, auth inicial, middlewares de tenant e modulo.
- Testes unitarios e de integração no backend, incluindo Testcontainers e helper global de Postgres.
- Frontend Web base com React, Vite, TanStack Router, TanStack Query, Zustand, Tailwind, RHF e Zod.
- Libs compartilhadas em TypeScript consumidas via `workspace:*`.
- Worker separado com fila dummy via Redis/Asynq e testes.
- CI para Go, frontend, SQLC e validação de Docker Compose.
- ADRs iniciais e `docs/*CONTRIBUTING*.md`.

### O que ainda esta abaixo do alvo descrito no `*README*.md`

- A API ainda cobre apenas uma fatia pequena do domínio real.
- O schema e amplo, mas a camada de aplicação ainda não implementa os módulos centrais do negocio, como `schedules`, clients, pets, services e reports.
- O Web ainda e um shell inicial com login e dashboard placeholder.
- As libs compartilhadas existem, mas ainda não são "source of truth" completo para enums, entidades e DTOs do domínio.
- Swagger ainda não foi integrado de forma real.
- Auditoria automática, refresh token, controle de plano por assinatura e módulos reais ainda não estão fechados.
- O Worker ainda opera com task dummy, não com eventos reais do negocio.

## Comparativo Consolidado entre *README* e Estado Atual

### Backend API

Alinhado:

- Estrutura modular em `cmd`, `internal`, `sqlc`, `handler`, `service`, `middleware`, `jwt`.
- SQLC com migrations centralizadas em `infra/migrations`.
- Healthcheck, readiness, login, middleware de auth, tenant e modulo.
- Testes de unidade e integração.

Parcial:

- O `*README*.md` descreve handlers e services de domínios amplos; hoje ha implementação real apenas para auth, users, verificações básicas e enqueue dummy do worker.
- Não ha ainda camada consistente para `schedules`, clients, pets, services, products, reports ou uploads.
- Swagger esta no plano arquitetural, mas não aparece ainda como entrega operacional da API.
- Auditoria imutável via middleware ainda não esta implementada.

### Frontend Web

Alinhado:

- Estrutura base em `src/main.tsx`, `router.tsx`, `routes`, `lib/api`, `lib/auth`, `stores`.
- Uso de Query para servidor e Zustand para auth/UI.
- Login funcional na arquitetura e testes do cliente HTTP e stores.

Parcial:

- O `*README*.md` projeta rotas e features para dashboard, `schedules`, clients, pets, services, reports e administração. Hoje a aplicação ainda esta concentrada em login, layout e dashboard inicial.
- Ainda não existe camada de query por domínio nem formulários de negocio reais.

### Libs Compartilhadas

Alinhado:

- `shared-types`, `shared-utils`, `shared-constants` e `ui` foram criadas.
- O Web consome as libs via workspace.
- Ha build e testes nas libs básicas.

Parcial:

- As libs ainda não refletem com profundidade o domínio descrito no *README*.
- Faltam enums, entidades, DTOs e constantes de módulos, erros e paginação em nível mais completo.
- `libs/ui` ainda e uma base utilitária, não um design system de componentes do domínio.

### Worker

Alinhado:

- Processo Go separado.
- Bootstrap, processor, scheduler e cliente placeholder.
- Integração básica com Redis/Asynq.

Parcial:

- O Worker ainda não processa eventos reais de notificação, expiração de assinaturas, relatórios ou limpeza.
- Ainda não compartilha contratos ricos com a API para eventos de domínio.

### Infra, Qualidade e Documentação

Alinhado:

- Docker Compose local para Postgres, Redis, pgAdmin e Asynqmon.
- Makefile com targets principais.
- CI para Go, frontend, SQLC e Compose.
- ADRs e guia de contribuição.

Parcial:

- O `*README*.md` ainda fala em workflows `ci.yml` e `deploy.yml`; o repositório usa arquivos diferentes.
- Não ha ainda documentação operacional da API via Swagger.
- Não ha fluxo de deploy por app implementado no escopo observado.

## Direção da Fase 2 do Produto

O próximo ciclo deve atacar quatro objetivos em paralelo leve:

1. consolidar contratos e coerência entre API, Web, Worker e libs;
2. implementar o primeiro conjunto real de módulos de negocio;
3. substituir placeholders por fluxos integrados ponta a ponta;
4. aumentar observabilidade, qualidade e previsibilidade da entrega.

## Fase 9 - Coerência de Contratos e Ambiente

### 9.1 - Ações

- Revisar divergências entre `*README*.md`, `.env.example`, Compose, Makefile e implementação atual.
- Padronizar versões de runtime e imagens:
  - PostgreSQL;
  - porta padrão da API;
  - variaveis JWT;
  - variaveis do Web.
- Garantir coerência entre seed, credenciais do Web e fluxo de login real.
- Revisar `shared-types` para refletir enums e payloads reais do backend.
- Corrigir queries SQLC que ainda geram tipos inadequados para filtros opcionais.

### 9.2 - Checks

- [x] `*README*.md`, `.env.example`, Makefile e Compose não se contradizem em porta, versão e comandos principais.
- [x] Login "seedado" funciona ponta a ponta do Web para a API.
- [x] `shared-types` espelha os enums atuais de auth, usuário e tenant sem uso excessivo de `string`.
- [x] `sqlc generate` não gera parâmetros incorretos para filtros opcionais relevantes.

## Fase 10 - Modulo Base de Empresas e Vínculos

### 10.1 - Ações

- Implementar camada completa para:
  - `companies`;
  - `company_users`;
  - `modules`;
  - `plans`;
  - `company_modules`.
- Expor endpoints protegidos para:
  - listar empresa corrente;
  - listar usuários da empresa;
  - consultar módulos ativos;
  - consultar plano atual.
- Ajustar seed para criar uma empresa de desenvolvimento funcional com:
  - company;
  - usuário admin;
  - company_user ativo;
  - company_modules compatíveis com o plano.
- Substituir endpoints públicos sensíveis por versões protegidas.

Nota: o seed da fase 10 agora inclui uma assinatura ativa para o plano atual da empresa de desenvolvimento.

### 10.2 - Checks

- [x] Usuário seedado recebe JWT com `company_id` valido.
- [x] Endpoint de empresa corrente responde com base no tenant do token.
- [x] Endpoint de módulos ativos reflete `company_modules`.
- [x] Nenhuma rota de dados administrativos permanece publica sem necessidade.

## Fase 11 - Primeiro Modulo de Negocio Real: `schedules`

### 11.1 - Ações

- Implementar queries SQLC de ``schedules`` e tabelas associadas necessárias para o MVP:
  - listagem;
  - criação;
  - consulta por id;
  - atualização;
  - soft delete;
  - histórico de status, se viável neste ciclo.
- Criar service e handler de `schedules` seguindo tenant por `company_id`.
- Adicionar validações de negocio minímas:
  - relação com company;
  - status valido;
  - datas coerentes;
  - entidade não deletada.
- Aplicar `RequireModule("SCH")` nas rotas do modulo.
- Adicionar testes unitarios e de integração do fluxo principal.

### 11.2 - Checks

- [x] `GET /api/v1/`schedules`` retorna somente registros do tenant.
- [x] `POST /api/v1/`schedules`` cria registro com `company_id` derivado do token.
- [x]x `PUT` e `DELETE` respeitam soft delete e ownership do tenant.
- [x] Testes de integração cobrem ão menos listagem, criação e isolamento multi-tenant.

Observação: além dos checks originais, a fase passou a expor `GET /api/v1/`schedules`/:id/history` para consulta do histórico de status por tenant e ganhou cobertura explicita para bloqueio de acesso quando o modulo `SCH` nao esta ativo.

## Fase 12 - Web do Primeiro Fluxo Real

### 12.1 - Ações

- Criar camada de queries para empresa corrente e `schedules`.
- Implementar rotas do Web para:
  - dashboard conectado;
  - listagem de `schedules`;
  - formulário de criação/edição.
- Conectar login, sessão e tenant ão fluxo de dados real.
- Substituir dados mockados do dashboard por dados vindos da API, mesmo que ainda simples.
- Introduzir componentes iniciais de tabela, empty state, loading e erro no `libs/ui/web` quando houver reutilização clara.

### 12.2 - Checks

- [x] Login leva a dashboard com dados reais da empresa corrente.
- [x] Tela de `schedules` lista dados da API.
- [x] Criação de schedule atualiza cache do Query corretamente.
- [x] Nenhum dado de servidor e salvo em Zustand fora da sessão/auth e UI.

Observação: além dos checks originais, a fase passou a contar com testes de componente para a dashboard conectada e testes de integração dos hooks de domínio para validar invalidação e recarga do cache de `schedules` via TanStack Query.

## Fase 13 - Auditoria, Erros e Observabilidade Básica

### 13.1 - Ações

- Implementar padrão unificado de erro HTTP no backend.
- Criar middleware de auditoria para mutações importantes.
- Registrar `audit_logs` ao menos para `companies`, `company_users` e `schedules`.
- Melhorar logs estruturados na API e no Worker.
- Adicionar correlation id básico por request, se o custo for baixo.

### 13.2 - Checks

- [x] Respostas de erro da API seguem formato consistente.
- [x] Mutações de `schedules` e `company_users` geram registros em `audit_logs`.
- [x] Logs da API e do Worker identificam tipo de operação, tenant e resultado.

Observação: a fase passou a incluir middleware de `correlation_id` por request, payload de erro unificado com `code` e `correlation_id`, e cobertura integrada para persistência de auditoria em `schedules` e `company_users`.

Observação: embora a ação cite `companies`, o escopo efetivamente entregue nesta fase concentrou a auditoria nos fluxos mutáveis hoje expostos pela API, com cobertura aplicada em `schedules` e `company_users`.

## Fase 14 - Worker com Evento de Negocio Real

### 14.1 - Ações

- Substituir ou complementar task dummy por evento real do domínio.
- Recomendação: usar confirmação de schedule como primeiro caso real.
- Na API:
  - publicar task ao confirmar/agendar;
  - versionar payload de task.
- No Worker:
  - processar payload;
  - resolver dados mínimos da entidade;
  - criar rota pública para `callback URL` necessária para o fluxo de notificação real do WhatsApp e add o `WHATSAPP_VERIFY_TOKEN`;
  - enviar via cliente WhatsApp placeholder ou logger estruturado.
- Adicionar retentativa, timeout e testes de integração.

### 14.2 - Checks

- [x] Ação de negocio publica task real no Redis.
- [x] Worker consome task real com payload validado.
- [x] Falhas de payload e processamento são tratadas sem derrubar o processo.

Observação: a Fase 14 foi implementada com o evento de confirmação de schedule como primeiro fluxo real, incluindo task versionada na API, consumer dedicado no worker, callback HTTP de WhatsApp com `WHATSAPP_VERIFY_TOKEN` e testes unitários/integrados para publicação, consumo e verificação de webhook.

## Fase 15 - Swagger, Contribuição e Consistencia Documental

### 15.1 - Ações

- Integrar Swaggo quando os handlers reais estiverem estáveis.
- Gerar `apps/api/docs` e expor rota `/swagger/*any`.
- Atualizar `*README*.md` para refletir:
  - o que ja esta implementado;
  - o que ainda e alvo arquitetural;
  - paths e versões realmente usados.
- Atualizar `docs/*CONTRIBUTING*.md` com fluxo de:
  - subir infraestrutura;
  - migrar;
  - seedar;
  - rodar API, Web e Worker;
  - executar testes.

### 15.2 - Checks

- [x] Swagger abre localmente e documenta auth e `schedules`.
- [x] *README* deixa claro o que e implementado vs planejado.
- [x] *CONTRIBUTING* permite onboarding local sem conhecimento implícito.

Observação: a fase foi concluída com integração do Swaggo na API (rota canônica `/swagger/*any`), geração versionada em `apps/api/docs`, alias compatível em `/api/v1/docs` com cobertura de testes para evitar regressão, além da revisão documental de onboarding e estado atual do projeto.

## Fase 16 - Endurecimento de Qualidade
\
Observação: a Fase 16 foi concluída com a consolidação dos comandos agregados no Makefile (`test`, `lint`, `build`, `dev`), expansão e validação dos testes unitários e de integração para middlewares, handlers e worker, além da revisão dos workflows de CI para garantir bloqueio de regressão em todos os módulos principais. O baseline de qualidade e previsibilidade foi consolidado, com verificação de cobertura mínima aplicada na CI para API e Worker em pacotes unitários determinísticos, e com comandos de desenvolvimento local padronizados e documentados.

### 16.1 - Ações

- Adicionar comandos agregados no Makefile para:
  - `test`;
  - `lint`;
  - `build`;
  - `dev`.
- Expandir testes:
  - middlewares;
  - `schedules` service/handler;
  - Web queries e formulários;
  - Worker com caso real.
- Adicionar verificação de cobertura minima onde fizer sentido.
- Revisar CI para refletir o novo baseline de módulos reais.

### 16.2 - Checks

- [x] CI bloqueia regressão em API, Web, Worker, SQLC e libs.
- [x] Casos principais do primeiro modulo real estão cobertos por testes.
- [x] Comandos de desenvolvimento local ficam previsíveis e curtos.

## Ordem Recomendada de Execução

1. Fase 9: coerência de contratos e ambiente.
2. Fase 10: empresas, vínculos e plano ativo.
3. Fase 11: modulo real de `schedules` na API.
4. Fase 12: Web conectado ão primeiro fluxo real.
5. Fase 13: auditoria, erros e observabilidade.
6. Fase 14: Worker com evento de negocio real.
7. Fase 15: Swagger e consolidação documental.
8. Fase 16: endurecimento de qualidade.

## Checklist Consolidado do Próximo Ciclo
\
Observação: todos os itens do ciclo foram concluídos, consolidando o baseline de qualidade, integração contínua e previsibilidade de entrega para os módulos implementados. O projeto está pronto para expansão de domínios e endurecimento adicional conforme novas fases.

- [x] Ambientes, versões, portas e variáveis estão coerentes entre *README*, `.env.example`, Compose e código.
- [x] Seed local cria tenant funcional e usuário utilizável pelo Web.
- [x] Contratos compartilhados refletem enums e payloads reais.
- [x] API expõe módulo real de `schedules` com isolamento multi-tenant.
- [x] Web consome dados reais de empresa e `schedules`.
- [x] Worker processa ao menos um evento real do negócio.
- [x] Swagger documenta os endpoints implementados.
- [x] *README* e *CONTRIBUTING* refletem o estado real do projeto.
- [x] Testes e CI cobrem os fluxos reais do primeiro modulo funcional.

## Riscos e Decisões Pendentes

- O `README.md` ainda mistura arquitetura alvo e implementação corrente. Isso pode induzir desenvolvimento fora de ordem se não for revisado.
- O schema e muito maior do que a camada de aplicação atual. O próximo ciclo deve escolher poucos módulos e fecha-los de ponta a ponta.
- As libs compartilhadas podem virar camada cosmética se não forem sincronizadas com contratos reais do backend.
- O Worker deve evoluir com eventos reais, mas sem criar dependencies prematuras em módulos ainda não implementados.
- Auditoria e refresh token devem entrar com critério, para evitar uma base de autenticação parcialmente segura e parcialmente improvisada.
