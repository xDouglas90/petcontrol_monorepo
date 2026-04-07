# Plano de Ação e Execução - Estrutura Inicial

## Objetivo

Estabelecer a estrutura inicial executável do monorepo PetControl, alinhada ao `README.md`, cobrindo backend (`apps/api`), frontend (`apps/web`), libs compartilhadas (`libs/*`) e infraestrutura local (`infra/docker` com Docker Compose).

O foco deste plano e criar uma base consistente para desenvolvimento incremental antes de implementar os módulos de negocio completos.

## Estado Atual Observado

- `README.md` documenta a arquitetura alvo do monorepo, stack, estrutura de apps, libs, infra, CI e testes.
- `apps/api` ja existe com `go.mod`, `go.sum`, `sqlc.yaml`, migrations SQL, query `users.sql` e código SQLC gerado.
- `go.work` foi adicionado na raiz para o workspace Go enxergar `apps/api` a partir da raiz do monorepo.
- `Makefile`, `.env.example`, `infra/` e `libs/` foram iniciados na raiz para padronização da estrutura.
- `infra/migrations` agora contem as migrations completas da base inicial e o `sqlc.yaml` de `apps/api` aponta para essa pasta.
- `infra/docker` agora contem a base local e de producao do Docker Compose para Postgres, Redis, pgAdmin e Asynqmon.
- `docs` existe, mas estava sem documentos versionados.
- `.github/workflows` ja existe com workflows relacionados a Go, frontend e proteção de branch.
- Ainda nao existem `apps/web` nem `apps/worker` na raiz.
- `schema.sql` permanece como copia de referencia do schema inicial, mas a fonte operacional agora e `infra/migrations`.

## Princípios de Execução

- Priorizar estrutura executável e verificável sobre implementação extensa de regra de negocio.
- Manter isolamento multi-tenant como requisito desde as primeiras queries e rotas protegidas.
- Evitar duplicação entre frontend/mobile via libs TypeScript compartilhadas.
- Tratar migrations e queries SQLC como fonte de verdade do backend.
- Cada fase deve terminar com checks objetivos de build, lint, teste ou validação manual.

## Fase 0 - Padronização Inicial do Monorepo

### 0.1 - Ações

- Definir a estrutura final de diretórios da raiz: `apps`, `libs`, `docs`, `infra`, `.github`.
- Criar ou ajustar `Makefile` na raiz com comandos padrão para API, Web, SQLC, migrations, Docker e testes.
- Criar `.env.example` na raiz com variáveis documentadas para API, Worker, Postgres, Redis, JWT, GCS e WhatsApp.
- Decidir a localização oficial das migrations:
  - mover para `infra/migrations` e ajustar `apps/api/sqlc.yaml`.
- Padronizar versões de runtime: Go `1.26.1`, Node LTS, pnpm e PostgreSQL `17`.

### 0.2 - Checks

- [x] `go work sync` executa sem erro.
- [x] `go test ./...` dentro de `apps/api` executa sem falha estrutural.
- [x] `sqlc generate` dentro de `apps/api` gera código sem alterações inesperadas.
- [x] `Makefile` possui pelo menos `dev-api`, `test-api`, `sqlc`, `docker-up`, `docker-down`.
- [x] `.env.example` cobre todas as variáveis usadas pelos comandos de desenvolvimento.

## Fase 1 - Infra Local com Docker e Docker Compose

### 1.1 - Ações

- Criar `infra/docker/docker-compose.yml` para desenvolvimento local com:
  - `postgres` usando `postgres:16-alpine`.
  - `redis` usando `redis:7-alpine`.
  - `pgadmin` para inspeção local.
  - `asynqmon` para monitorar filas quando o Worker existir.
- Criar `infra/docker/docker-compose.prod.yml` apenas com estrutura base, sem segredos hardcoded.
- Adicionar volumes nomeados para Postgres e Redis quando necessário.
- Adicionar healthchecks para Postgres e Redis.
- Criar comandos `docker-up`, `docker-down`, `docker-logs` e `docker-ps` no `Makefile`.
- Documentar a `DATABASE_URL` local e a porta dos serviços no `.env.example`.

### 1.2 - Checks

- [x] `docker compose -f infra/docker/docker-compose.yml config` valida a sintaxe.
- [x] `make docker-up` sobe Postgres e Redis.
- [x] `docker compose -f infra/docker/docker-compose.yml ps` mostra Postgres e Redis saudáveis.
- [x] Conexão Postgres local funciona com a `DATABASE_URL` documentada.
- [x] Redis responde via `redis-cli ping` ou check equivalente no container.

## Fase 2 - Backend API Base (`apps/api`)

### 2.1 - Ações

- Criar estrutura minima do Go conforme README:
  - `cmd/server/main.go`.
  - `internal/config/config.go`.
  - `internal/db/pool.go`.
  - `internal/apperror/errors.go`.
  - `internal/middleware`.
  - `internal/handler`.
  - `internal/service`.
  - `internal/jwt`.
  - `internal/validator`.
- Criar `apps/api/Dockerfile` (multi-stage) e `entrypoint.sh` da API, compatíveis com `cmd/server/main.go` e `infra/migrations`.
- Conectar `pgxpool` usando `DATABASE_URL`.
- Expor rotas iniciais:
  - `GET /health`.
  - `GET /ready` validando conexão com Postgres.
  - `GET /api/v1/users` ou rota equivalente usando SQLC, protegida quando auth estiver pronta.
- Adicionar dependências planejadas no `go.mod` de forma incremental: Gin, godotenv, validator, JWT, bcrypt, golang-migrate, swaggo, testify e testcontainers apenas quando usados.
- Garantir que SQLC leia queries e migrations do local padronizado na Fase 0.
- Implementar middleware de erro e formato padrão de resposta da API.

### 2.2 - Checks

- [] `go mod tidy` em `apps/api` nao remove dependências necessárias nem deixa pacotes quebrados.
- [] `go run ./cmd/server` inicia o servidor local.
- [] `curl http://localhost:<API_PORT>/health` retorna sucesso.
- [] `curl http://localhost:<API_PORT>/ready` retorna sucesso com Postgres ativo.
- [] `go test ./...` em `apps/api` passa.
- [] `sqlc generate` passa e mantém `internal/db/sqlc` consistente.

## Fase 3 - Migrations, Seed e Persistência

### 3.1 - Ações

- Adicionar comandos de migration no `Makefile` usando `golang-migrate`.
- Criar script `infra/scripts/migrate.sh` ou comando make equivalente para CI/local.
- Criar seed mínimo para dados essenciais:
  - Módulos.
  - Tipos de planos.
  - Plano inicial.
  - Usuário root/admin de desenvolvimento, se aplicável.
- Validar que o schema multi-tenant possui indices para `company_id` nas entidades que usam tenant.
- Criar queries SQLC iniciais para domínios base: users, companies, company_users, modules, plans.

### 3.2 - Checks

- [] `make migrate-up` aplica a migration inicial em banco limpo.
- [] `make migrate-down` reverte a ultima migration sem erro.
- [] `make seed` cria dados mínimos idempotentes ou falha de forma controlada.
- [] `sqlc generate` gera os arquivos esperados para as novas queries.
- [] Queries tenant-aware sempre recebem `company_id` quando a tabela pertence ao tenant.

## Fase 4 - Autenticação, Tenant e Permissões

### 4.1 - Ações

- Implementar fluxo inicial de auth:
  - Login com e-mail e senha.
  - Hash de senha com bcrypt.
  - Access token JWT com `user_id`, `company_id`, `role` e `kind`.
  - Refresh token persistido ou controlado via Redis conforme decisão técnica.
- Criar middleware `Auth` para validar JWT.
- Criar middleware `Tenant` para injetar `company_id` no contexto.
- Criar middleware `RequireModule` para validar modulo ativo da empresa.
- Registrar logs de login em `login_history` ou tabela equivalente do schema.
- Definir padrão de erros para `401`, `403`, `404`, `409` e `422`.

### 4.2 - Checks

- [] Login com credenciais validas retorna JWT.
- [] Login invalido retorna erro padronizado e nao vaza detalhe sensível.
- [] Rota autenticada sem token retorna `401`.
- [] Rota autenticada com token sem `company_id` retorna `403`.
- [] Queries de dados de tenant usam `company_id` obtido do middleware, nao do body do request.
- [] Testes unitários cobrem middlewares `Auth` e `Tenant`.

## Fase 5 - Frontend Web Base (`apps/web`)

### 5.1 - Ações

- Criar `apps/web` com React, Vite e TypeScript.
- Configurar pnpm workspace na raiz, se o projeto optar por workspaces JS/TS.
- Instalar e configurar:
  - TanStack Router.
  - TanStack Query.
  - Zustand.
  - TailwindCSS.
  - React Hook Form.
  - Zod.
  - ESLint e Prettier.
- Criar estrutura inicial:
  - `src/main.tsx`.
  - `src/router.tsx`.
  - `src/routes/(auth)/login.tsx`.
  - `src/routes/(app)/_layout.tsx`.
  - `src/routes/(app)/dashboard/index.tsx`.
  - `src/lib/api/rest-client.ts`.
  - `src/lib/auth/auth.store.ts`.
  - `src/stores/ui.store.ts`.
- Criar `VITE_API_URL` no `.env.example`.
- Implementar tela de login conectada ao endpoint inicial da API ou mock controlado enquanto a API nao estiver pronta.

### 5.2 - Checks

- [] `pnpm install` executa sem conflito de workspace.
- [] `pnpm --filter web dev` inicia o Vite.
- [] `pnpm --filter web build` gera build de produção.
- [] `pnpm --filter web lint` passa.
- [] Login chama `VITE_API_URL` configurado.
- [] Estado vindo da API fica no TanStack Query; Zustand fica restrito a auth/UI.

## Fase 6 - Libs Compartilhadas (`libs/*`)

### 6.1 - Ações

- Criar `libs/shared-types` para entidades, DTOs e enums usados pelo Web e Mobile futuro.
- Criar `libs/shared-utils` para formatadores, validadores e helpers puros.
- Criar `libs/shared-constants` para rotas, códigos de erro, paginação e limites de plano.
- Criar `libs/ui` com separação:
  - `core` para tokens, hooks e utils sem dependência de plataforma.
  - `web` para componentes React DOM.
  - `native` apenas como estrutura futura, sem acoplamento prematuro.
- Configurar `package.json`, `tsconfig.json` e exports para cada lib.
- Atualizar `apps/web` para consumir libs via workspace em vez de paths relativos profundos.

### 6.2 - Checks

- []`pnpm -r build` ou comando equivalente compila libs e web.
- [] `shared-utils` possui testes unitários para validadores e formatadores iniciais.
- [] `shared-types` exporta DTOs usados pelo `rest-client`.
- [] `shared-constants` evita magic strings de rotas no Web.
- [] `libs/ui/core` nao importa React DOM nem React Native.

## Fase 7 - Worker Base (`apps/worker`)

### 7.1 - Ações

- Criar `apps/worker` como processo Go separado.
- Definir se o Worker sera modulo Go independente ou parte do mesmo modulo/workspace Go.
- Criar estrutura:
  - `cmd/worker/main.go`.
  - `internal/config`.
  - `internal/db`.
  - `internal/processor`.
  - `internal/scheduler`.
  - `internal/whatsapp`.
- Configurar Asynq com Redis.
- Criar fila inicial `notifications` com task dummy verificável.
- Compartilhar tipos de task com a API via pacote Go interno comum ou duplicação minima documentada.

### 7.2 - Checks

- [] `go run ./cmd/worker` inicia e conecta no Redis.
- [] API consegue publicar task dummy.
- [] Worker consome task dummy e registra log estruturado.
- [] `go test ./...` no Worker passa.
- [] Worker pode ser desligado sem afetar a API.

## Fase 8 - Qualidade, CI e Documentação

### 8.1 - Ações

- Revisar `.github/workflows/go.yml` e `.github/workflows/frontend.yml` para refletir os comandos reais.
- Garantir que CI rode:
  - Build e teste da API.
  - SQLC generation check.
  - Build e lint do Web.
  - Testes das libs.
  - Docker Compose config validation.
- Adicionar `docs/CONTRIBUTING.md` com setup local.
- Criar `docs/adr` e registrar ADRs iniciais:
  - Go + Gin.
  - SQLC em vez de ORM.
  - Worker separado.
  - Estrategia de multi-tenancy por `company_id`.
- Adicionar Swagger quando houver handlers reais suficientes para justificar a geração.

### 8.2 - Checks

- [] Workflows do GitHub Actions executam em pull request.
- [] `make test` ou comando equivalente roda a suite minima local.
- [] `make lint` ou comando equivalente cobre Go e TypeScript.
- [] `docs/CONTRIBUTING.md` permite setup local sem depender de conhecimento oral.
- [] ADRs iniciais explicam o motivo das decisões, nao apenas a tecnologia escolhida.

## Ordem Recomendada de Execução

1. Fase 0: padronizar estrutura e paths.
2. Fase 1: subir infra local com Docker Compose.
3. Fase 2: tornar API executável com health/readiness.
4. Fase 3: estabilizar migrations, seed e SQLC.
5. Fase 4: implementar auth, tenant e permissões base.
6. Fase 5: criar Web com login/layout/rest-client.
7. Fase 6: extrair contratos, constantes e utils para libs.
8. Fase 7: adicionar Worker com fila dummy.
9. Fase 8: fechar CI, docs e ADRs.

## Checklist Consolidado

- [ ] Estrutura raiz possui `apps`, `libs`, `docs`, `infra` e `.github`.
- [ ] `Makefile` centraliza comandos de dev, build, test, lint, sqlc, migrations e Docker.
- [ ] `.env.example` documenta todas as variáveis locais.
- [ ] Docker Compose sobe Postgres, Redis, pgAdmin e asynqmon.
- [ ] API inicia com `go run ./cmd/server`.
- [ ] API expõe `/health` e `/ready`.
- [ ] SQLC gera código a partir das queries e migrations padronizadas.
- [ ] Migrations aplicam e revertem em banco limpo.
- [ ] Auth JWT e middleware de tenant estão implementados antes de rotas multi-tenant reais.
- [ ] Web inicia com Vite e consome `VITE_API_URL`.
- [ ] Libs TS exportam tipos, constantes, utils e base de UI sem acoplamento indevido.
- [ ] Worker inicia separado e consome task dummy do Redis.
- [ ] CI executa checks de Go, TypeScript, SQLC e Docker Compose.
- [ ] Docs de contribuição e ADRs iniciais estão versionados.

## Riscos e Decisões Pendentes

- O README menciona `apps/worker` e `libs`, mas eles ainda nao existem. Criar a estrutura sem implementar regra complexa reduz risco de acoplamento prematuro.
- O Web depende de contratos estáveis da API. Enquanto auth e endpoints base nao estiverem prontos, usar mocks deve ser temporário e explicitamente marcado.
- O Worker deve ser criado depois da API publicar uma task minima; caso contrario, a integração com Redis fica pouco verificável.
- Swagger deve entrar depois de handlers reais, para evitar documentação automática vazia ou enganosa.
