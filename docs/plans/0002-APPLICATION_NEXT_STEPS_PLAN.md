# Plano de Acao e Execucao - Proximos Passos da Aplicacao

## Objetivo

Definir a segunda etapa de desenvolvimento do PetControl apos a consolidacao da estrutura inicial do monorepo.

Este plano compara:

- o alvo arquitetural descrito no `README.md`;
- o que foi registrado e entregue no `docs/plans/0001-INITIAL_STRUCTURE_ACTION_PLAN.md`;
- o estado atual do repositório.

O foco agora deixa de ser "criar a base do monorepo" e passa a ser "evoluir a aplicacao para um produto funcional, coerente e expandivel".

## Resumo Executivo

### O que ja foi entregue com boa aderencia ao `README.md`

- Monorepo estruturado com `apps/api`, `apps/web`, `apps/worker`, `libs`, `infra`, `docs` e `.github`.
- API Go executavel com `health`, `ready`, SQLC, migrations, auth inicial, middlewares de tenant e modulo.
- Testes unitarios e de integracao no backend, incluindo Testcontainers e helper global de Postgres.
- Frontend Web base com React, Vite, TanStack Router, TanStack Query, Zustand, Tailwind, RHF e Zod.
- Libs compartilhadas em TypeScript consumidas via `workspace:*`.
- Worker separado com fila dummy via Redis/Asynq e testes.
- CI para Go, frontend, SQLC e validacao de Docker Compose.
- ADRs iniciais e `docs/CONTRIBUTING.md`.

### O que ainda esta abaixo do alvo descrito no `README.md`

- A API ainda cobre apenas uma fatia pequena do dominio real.
- O schema e amplo, mas a camada de aplicacao ainda nao implementa os modulos centrais do negocio, como schedules, clients, pets, services e reports.
- O Web ainda e um shell inicial com login e dashboard placeholder.
- As libs compartilhadas existem, mas ainda nao sao "source of truth" completo para enums, entidades e DTOs do dominio.
- Swagger ainda nao foi integrado de forma real.
- Auditoria automatica, refresh token, controle de plano por assinatura e modulos reais ainda nao estao fechados.
- O Worker ainda opera com task dummy, nao com eventos reais do negocio.

## Comparativo Consolidado entre README e Estado Atual

### Backend API

Alinhado:

- Estrutura modular em `cmd`, `internal`, `sqlc`, `handler`, `service`, `middleware`, `jwt`.
- SQLC com migrations centralizadas em `infra/migrations`.
- Healthcheck, readiness, login, middleware de auth, tenant e modulo.
- Testes de unidade e integracao.

Parcial:

- O `README.md` descreve handlers e services de dominios amplos; hoje ha implementacao real apenas para auth, users, verificacoes basicas e enqueue dummy do worker.
- Nao ha ainda camada consistente para schedules, clients, pets, services, products, reports ou uploads.
- Swagger esta no plano arquitetural, mas nao aparece ainda como entrega operacional da API.
- Auditoria imutavel via middleware ainda nao esta implementada.

### Frontend Web

Alinhado:

- Estrutura base em `src/main.tsx`, `router.tsx`, `routes`, `lib/api`, `lib/auth`, `stores`.
- Uso de Query para servidor e Zustand para auth/UI.
- Login funcional na arquitetura e testes do cliente HTTP e stores.

Parcial:

- O `README.md` projeta rotas e features para dashboard, schedules, clients, pets, services, reports e administracao. Hoje a aplicacao ainda esta concentrada em login, layout e dashboard inicial.
- Ainda nao existe camada de query por dominio nem formularios de negocio reais.

### Libs Compartilhadas

Alinhado:

- `shared-types`, `shared-utils`, `shared-constants` e `ui` foram criadas.
- O Web consome as libs via workspace.
- Ha build e testes nas libs basicas.

Parcial:

- As libs ainda nao refletem com profundidade o dominio descrito no README.
- Faltam enums, entidades, DTOs e constantes de modulos, erros e paginacao em nivel mais completo.
- `libs/ui` ainda e uma base utilitaria, nao um design system de componentes do dominio.

### Worker

Alinhado:

- Processo Go separado.
- Bootstrap, processor, scheduler e cliente placeholder.
- Integracao basica com Redis/Asynq.

Parcial:

- O Worker ainda nao processa eventos reais de notificacao, expiracao de assinaturas, relatórios ou limpeza.
- Ainda nao compartilha contratos ricos com a API para eventos de dominio.

### Infra, Qualidade e Documentacao

Alinhado:

- Docker Compose local para Postgres, Redis, pgAdmin e Asynqmon.
- Makefile com targets principais.
- CI para Go, frontend, SQLC e Compose.
- ADRs e guia de contribuicao.

Parcial:

- O `README.md` ainda fala em workflows `ci.yml` e `deploy.yml`; o repositório usa arquivos diferentes.
- Nao ha ainda documentacao operacional da API via Swagger.
- Nao ha fluxo de deploy por app implementado no escopo observado.

## Direcao da Fase 2 do Produto

O proximo ciclo deve atacar quatro objetivos em paralelo leve:

1. consolidar contratos e coerencia entre API, Web, Worker e libs;
2. implementar o primeiro conjunto real de modulos de negocio;
3. substituir placeholders por fluxos integrados ponta a ponta;
4. aumentar observabilidade, qualidade e previsibilidade da entrega.

## Fase 9 - Coerencia de Contratos e Ambiente

### 9.1 - Acoes

- Revisar divergencias entre `README.md`, `.env.example`, Compose, Makefile e implementacao atual.
- Padronizar versoes de runtime e imagens:
  - PostgreSQL;
  - porta padrao da API;
  - variaveis JWT;
  - variaveis do Web.
- Garantir coerencia entre seed, credenciais do Web e fluxo de login real.
- Revisar `shared-types` para refletir enums e payloads reais do backend.
- Corrigir queries SQLC que ainda geram tipos inadequados para filtros opcionais.

### 9.2 - Checks

- [x] `README.md`, `.env.example`, Makefile e Compose nao se contradizem em porta, versao e comandos principais.
- [x] Login seedado funciona ponta a ponta do Web para a API.
- [x] `shared-types` espelha os enums atuais de auth, usuario e tenant sem uso excessivo de `string`.
- [x] `sqlc generate` nao gera parametros incorretos para filtros opcionais relevantes.

## Fase 10 - Modulo Base de Empresas e Vinculos

### 10.1 - Acoes

- Implementar camada completa para:
  - `companies`;
  - `company_users`;
  - `modules`;
  - `plans`;
  - `company_modules`.
- Expor endpoints protegidos para:
  - listar empresa corrente;
  - listar usuarios da empresa;
  - consultar modulos ativos;
  - consultar plano atual.
- Ajustar seed para criar uma empresa de desenvolvimento funcional com:
  - company;
  - usuario admin;
  - company_user ativo;
  - company_modules compativeis com o plano.
- Substituir endpoints publicos sensiveis por versoes protegidas.

### 10.2 - Checks

- Usuario seedado recebe JWT com `company_id` valido.
- Endpoint de empresa corrente responde com base no tenant do token.
- Endpoint de modulos ativos reflete `company_modules`.
- Nenhuma rota de dados administrativos permanece publica sem necessidade.

## Fase 11 - Primeiro Modulo de Negocio Real: Schedules

### 11.1 - Acoes

- Implementar queries SQLC de `schedules` e tabelas associadas necessarias para o MVP:
  - listagem;
  - criacao;
  - consulta por id;
  - atualizacao;
  - soft delete;
  - historico de status, se viavel neste ciclo.
- Criar service e handler de schedules seguindo tenant por `company_id`.
- Adicionar validacoes de negocio minimas:
  - relacao com company;
  - status valido;
  - datas coerentes;
  - entidade nao deletada.
- Aplicar `RequireModule("SCH")` nas rotas do modulo.
- Adicionar testes unitarios e de integracao do fluxo principal.

### 11.2 - Checks

- `GET /api/v1/schedules` retorna somente registros do tenant.
- `POST /api/v1/schedules` cria registro com `company_id` derivado do token.
- `PUT` e `DELETE` respeitam soft delete e ownership do tenant.
- Testes de integracao cobrem ao menos listagem, criacao e isolamento multi-tenant.

## Fase 12 - Web do Primeiro Fluxo Real

### 12.1 - Acoes

- Criar camada de queries para empresa corrente e schedules.
- Implementar rotas do Web para:
  - dashboard conectado;
  - listagem de schedules;
  - formulario de criacao/edicao.
- Conectar login, sessao e tenant ao fluxo de dados real.
- Substituir dados mockados do dashboard por dados vindos da API, mesmo que ainda simples.
- Introduzir componentes iniciais de tabela, empty state, loading e erro no `libs/ui/web` quando houver reutilizacao clara.

### 12.2 - Checks

- Login leva a dashboard com dados reais da empresa corrente.
- Tela de schedules lista dados da API.
- Criacao de schedule atualiza cache do Query corretamente.
- Nenhum dado de servidor e salvo em Zustand fora da sessao/auth e UI.

## Fase 13 - Auditoria, Erros e Observabilidade Basica

### 13.1 - Acoes

- Implementar padrao unificado de erro HTTP no backend.
- Criar middleware de auditoria para mutacoes importantes.
- Registrar `audit_logs` ao menos para companies, company_users e schedules.
- Melhorar logs estruturados na API e no Worker.
- Adicionar correlation id basico por request, se o custo for baixo.

### 13.2 - Checks

- Respostas de erro da API seguem formato consistente.
- Mutacoes de schedules e company_users geram registros em `audit_logs`.
- Logs da API e do Worker identificam tipo de operacao, tenant e resultado.

## Fase 14 - Worker com Evento de Negocio Real

### 14.1 - Acoes

- Substituir ou complementar task dummy por evento real do dominio.
- Recomendacao: usar confirmacao de schedule como primeiro caso real.
- Na API:
  - publicar task ao confirmar/agendar;
  - versionar payload de task.
- No Worker:
  - processar payload;
  - resolver dados minimos da entidade;
  - enviar via cliente WhatsApp placeholder ou logger estruturado.
- Adicionar retentativa, timeout e testes de integracao.

### 14.2 - Checks

- Acao de negocio publica task real no Redis.
- Worker consome task real com payload validado.
- Falhas de payload e processamento sao tratadas sem derrubar o processo.

## Fase 15 - Swagger, Contribuicao e Consistencia Documental

### 15.1 - Acoes

- Integrar Swaggo quando os handlers reais estiverem estaveis.
- Gerar `apps/api/docs` e expor rota `/swagger/*any`.
- Atualizar `README.md` para refletir:
  - o que ja esta implementado;
  - o que ainda e alvo arquitetural;
  - paths e versoes realmente usados.
- Atualizar `docs/CONTRIBUTING.md` com fluxo de:
  - subir infraestrutura;
  - migrar;
  - seedar;
  - rodar API, Web e Worker;
  - executar testes.

### 15.2 - Checks

- Swagger abre localmente e documenta auth e schedules.
- README deixa claro o que e implementado vs planejado.
- CONTRIBUTING permite onboarding local sem conhecimento implícito.

## Fase 16 - Endurecimento de Qualidade

### 16.1 - Acoes

- Adicionar comandos agregados no Makefile para:
  - `test`;
  - `lint`;
  - `build`;
  - `dev`.
- Expandir testes:
  - middlewares;
  - schedules service/handler;
  - Web queries e formularios;
  - Worker com caso real.
- Adicionar verificacao de cobertura minima onde fizer sentido.
- Revisar CI para refletir o novo baseline de modulos reais.

### 16.2 - Checks

- CI bloqueia regressao em API, Web, Worker, SQLC e libs.
- Casos principais do primeiro modulo real estao cobertos por testes.
- Comandos de desenvolvimento local ficam previsiveis e curtos.

## Ordem Recomendada de Execucao

1. Fase 9: coerencia de contratos e ambiente.
2. Fase 10: empresas, vinculos e plano ativo.
3. Fase 11: modulo real de schedules na API.
4. Fase 12: Web conectado ao primeiro fluxo real.
5. Fase 13: auditoria, erros e observabilidade.
6. Fase 14: Worker com evento de negocio real.
7. Fase 15: Swagger e consolidacao documental.
8. Fase 16: endurecimento de qualidade.

## Checklist Consolidado do Proximo Ciclo

- [ ] Ambientes, versoes, portas e variaveis estao coerentes entre README, `.env.example`, Compose e codigo.
- [ ] Seed local cria tenant funcional e usuario utilizavel pelo Web.
- [ ] Contratos compartilhados refletem enums e payloads reais.
- [ ] API expõe modulo real de schedules com isolamento multi-tenant.
- [ ] Web consome dados reais de empresa e schedules.
- [ ] Worker processa ao menos um evento real do negocio.
- [ ] Swagger documenta os endpoints implementados.
- [ ] README e CONTRIBUTING refletem o estado real do projeto.
- [ ] Testes e CI cobrem os fluxos reais do primeiro modulo funcional.

## Riscos e Decisoes Pendentes

- O `README.md` ainda mistura arquitetura alvo e implementacao corrente. Isso pode induzir desenvolvimento fora de ordem se nao for revisado.
- O schema e muito maior do que a camada de aplicacao atual. O proximo ciclo deve escolher poucos modulos e fecha-los de ponta a ponta.
- As libs compartilhadas podem virar camada cosmetica se nao forem sincronizadas com contratos reais do backend.
- O Worker deve evoluir com eventos reais, mas sem criar dependencias prematuras em modulos ainda nao implementados.
- Auditoria e refresh token devem entrar com criterio, para evitar uma base de autenticacao parcialmente segura e parcialmente improvisada.
