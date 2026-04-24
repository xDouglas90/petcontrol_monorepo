# PetControl — Monorepo

> Documento vivo de arquitetura e decisões técnicas.
> Atualizar conforme o projeto evolui.

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Stack Tecnológica](#2-stack-tecnológica)
3. [Estrutura do Monorepo](#3-estrutura-do-monorepo)
4. [Libs Compartilhadas](#4-libs-compartilhadas)
5. [App: API](#5-app-api)
6. [App: Web](#6-app-web)
7. [App: Worker](#7-app-worker)
8. [Infraestrutura e DevOps](#8-infraestrutura-e-devops)
9. [Estratégia de Testes](#9-estratégia-de-testes)
10. [ADRs — Decisões de Arquitetura](#10-adrs--decisões-de-arquitetura)

---

## 1. Visão Geral

PetControl é uma plataforma SaaS **multi-tenant** voltada para pet shops e clínicas veterinárias. Cada empresa (tenant) possui seus próprios dados, planos e configurações. O sistema gerencia agendamentos, clientes, pets, serviços, funcionários, notificações e financeiro.

### Princípios arquiteturais

- **Isolamento de tenant por `company_id`** em todas as queries — nunca dados cruzados entre empresas.
- **Controle de acesso por plano** — funcionalidades liberadas conforme `module_package` da assinatura ativa.
- **Código compartilhado via libs internas** — sem duplicação entre API, Web e (futuramente) Mobile.
- **Auditoria imutável** — toda ação relevante é registrada sem possibilidade de deleção ou cascata.
- **Soft delete por padrão** — nenhuma entidade de negócio é removida fisicamente do banco.

### Credenciais de desenvolvimento (seed)

Após executar o seed, o ambiente local cria dois usuários prontos para login:

- `admin@petcontrol.local` / `password123` (compatível com o formulário padrão do Web)
- `root@petcontrol.local` / `password123` (com `must_change_password=true`)

### Estado atual vs alvo arquitetural (Abr/2026)

Implementado nesta etapa:

- API com autenticação JWT, multi-tenant por `company_id`, auditoria, correlation id e módulo real de `schedules`.
- API com módulo base de `clients` protegido por `CLI`, incluindo CRUD tenant-aware e auditoria das mutações.
- API com módulo base de `pets` protegido por `PET`, incluindo ownership validado por tenant, soft delete e payload com `owner_name`.
- API com módulo base de `services` protegido por `SCH`, incluindo catálogo por tenant, resolução de `service_types` e auditoria das mutações.
- `schedules` enriquecido com `client_name`, `pet_name`, `service_ids` e `service_titles`, além de persistência real em `schedule_services`.
- Worker com evento real `schedules:confirmed` (além do dummy legado), payload versionado em `v2` e callback HTTP de WhatsApp (`/webhook/whatsapp`).
- Web conectado aos fluxos reais de login, dashboard, `schedules`, `clients`, `pets` e `services`, com seletores reais de cliente, pet e serviço em vez de UUID digitado manualmente.
- Swagger operacional em `apps/api/docs`, com rota canônica em `/swagger/index.html` e alias compatível em `/api/v1/docs`.

Ainda planejado para próximos ciclos:

- Rotas Web e endpoints operacionais para `reports`, escolhido como próximo vertical recomendado.
- Endpoints adicionais documentados no Swagger conforme novos handlers forem estabilizados.
- Hardening de qualidade/CI e consolidação documental contínua.

O seed local agora também prepara massa operacional mínima para desenvolvimento:

- 1 cliente ativo vinculado ao tenant `petcontrol-dev`;
- 1 pet ativo desse cliente;
- 1 serviço ativo disponível para a empresa;
- 1 `schedule` de exemplo já confirmado para validar listagem e integração local.

---

## 2. Stack Tecnológica

### Backend (`/apps/api`)

| Categoria      | Tecnologia                      | Justificativa                                                                   |
| -------------- | ------------------------------- | ------------------------------------------------------------------------------- |
| Linguagem      | Go 1.26.1                       | Performance nativa, binário único, excelente concorrência via goroutines        |
| Framework HTTP | Gin                             | Roteamento rápido, middleware composável, ecossistema maduro                    |
| Banco de dados | PostgreSQL 18                   | Robusto, JSONB para auditoria, arrays nativos, UUID                             |
| Query layer    | SQLC                            | Geração de código type-safe a partir de SQL puro — zero ORM magic               |
| Migrations     | `golang-migrate`                | Migrations versionadas em SQL puro, aplicadas via CLI ou programaticamente      |
| Driver PG      | `pgx/v5`                        | Driver nativo PostgreSQL, máxima performance, suporte a pgxpool                 |
| Docs REST      | Swaggo (`swag` + `gin-swagger`) | Geração automática de OpenAPI a partir de anotações no código                   |
| Filas          | Asynq + Redis                   | Jobs assíncronos: notificações, expiração de planos, relatórios                 |
| Auth           | JWT (`golang-jwt`) + bcrypt     | Access token stateless, hash de senha com bcrypt; refresh token ainda planejado |
| Validação      | `go-playground/validator`       | Tags de validação struct-based, integradas ao binding do Gin                    |
| Storage        | Google Cloud Storage SDK        | Upload de arquivos (fotos de pets, documentos)                                  |
| Config         | `godotenv` + `os.Getenv`        | Leitura de `.env` em desenvolvimento; variáveis de ambiente em produção         |
| Testes         | `testify` + `testcontainers-go` | Unitários com mocks, integração com banco real em container                     |

### Frontend (`/apps/web`)

| Categoria     | Tecnologia             | Justificativa                                                        |
| ------------- | ---------------------- | -------------------------------------------------------------------- |
| Framework     | React 18+              | Ecossistema, compatibilidade com React Native no futuro              |
| Build         | Vite                   | HMR instantâneo, build rápido, ESM nativo                            |
| Roteamento    | TanStack Router        | Type-safe, file-based routing, excelente DX                          |
| Data fetching | TanStack Query         | Cache, revalidação, estados de loading/error                         |
| Estado global | Zustand                | Apenas estado de UI (tema, sidebar, modais) — sem estado de servidor |
| Estilo        | TailwindCSS 3+         | Utility-first, compatível com design system tokens                   |
| Componentes   | Shadcn/UI              | Acessível, não-opinado, baseado em Radix UI                          |
| Forms         | React Hook Form + Zod  | Performance e validação type-safe                                    |
| REST Client   | TanStack Query + fetch | Consome a API REST do backend Go                                     |

### Worker (`/apps/worker`)

| Categoria    | Tecnologia                       |
| ------------ | -------------------------------- |
| Linguagem    | Go 1.26.1                        |
| Filas        | Asynq (consumidor Redis)         |
| Scheduler    | Cron interno via `robfig/cron`   |
| Notificações | SDK WhatsApp Business API (HTTP) |
| Banco        | `pgx/v5` + SQLC (mesmo schema)   |

### Mobile (planejado — `apps/mobile`)

| Categoria     | Tecnologia                    |
| ------------- | ----------------------------- |
| Framework     | React Native (Expo)           |
| Navegação     | Expo Router (file-based)      |
| Estilo        | NativeWind (Tailwind para RN) |
| Componentes   | `libs/ui` (camada adaptada)   |
| Data fetching | TanStack Query                |

### Infra / DevOps

| Categoria  | Tecnologia                                |
| ---------- | ----------------------------------------- |
| Monorepo   | Makefile + scripts shell (sem Turborepo)  |
| Containers | Docker + Docker Compose                   |
| CI/CD      | GitHub Actions                            |
| Linting Go | `golangci-lint`                           |
| Linting JS | ESLint + Prettier                         |
| Commits    | Commitlint + Conventional Commits         |
| Secrets    | `.env` local + secret manager em produção |

---

## 3. Estrutura do Monorepo

```text
/petcontrol-monorepo
│
├── apps/
│   ├── api/                        # Backend Go + Gin
│   ├── web/                        # Frontend React + Vite
│   └── worker/                     # Jobs assíncronos Go + Asynq
│
├── libs/
│   ├── shared-types/               # Tipos, DTOs e enums compartilhados (TS — frontend/mobile)
│   ├── shared-utils/               # Funções puras sem dependência de framework (TS)
│   ├── shared-constants/           # Constantes de domínio e negócio (TS)
│   └── ui/                         # Design system: componentes cross-platform
│       ├── core/                   # Lógica e tokens sem dependência de plataforma
│       ├── web/                    # Componentes React (usa Tailwind + Radix)
│       └── native/                 # Componentes React Native (usa NativeWind)
│
├── docs/                           # Documentação do projeto
│   ├── adr/                        # Architecture Decision Records
│   ├── conventions/                # Convenções compartilhadas entre apps e libs
│   ├── plans/                      # Planos de execução por fases
│   ├── diagrams/                   # Diagramas de entidade e fluxo
│   └── CONTRIBUTING.md
│
├── infra/
│   ├── docker/
│   │   ├── docker-compose.yml      # PostgreSQL + Redis + pgAdmin
│   │   └── docker-compose.prod.yml
│   ├── migrations/                 # Migrations SQL (golang-migrate)
│   │   ├── 000001_init.up.sql
│   │   ├── 000001_init.down.sql
│   │   └── ...
│   ├── scripts/
│   │   ├── seed.sh                 # Seed de dados iniciais
│   │   └── migrate.sh              # Script de migration para CI
│   └── nginx/                      # Config de reverse proxy (produção)
│
├── .github/
│   └── workflows/
│       ├── go.yml                  # CI do backend Go
│       ├── frontend.yml            # CI do frontend e libs TS
│       └── branch-protection.yml   # Regras de proteção de branch
│
└── .env.example                    # Variáveis de ambiente documentadas
```

### 3.1 Taxonomia de `docs/`

Para evitar misturar documento operacional, decisão arquitetural e plano de execução, a pasta `docs/` segue esta divisão:

- `docs/plans/`: planos incrementais de implementação, com fases, checks e ordem recomendada de execução.
- `docs/adr/`: decisões arquiteturais relevantes e duradouras, registrando contexto, trade-offs e decisão final.
- `docs/conventions/`: convenções e regras compartilhadas de operação, navegação, contratos de uso e organização entre apps e libs.
- `docs/CONTRIBUTING.md`: onboarding local, fluxo de desenvolvimento e comandos principais do repositório.

---

## 4. Libs Compartilhadas

> As libs abaixo são **TypeScript** e servem exclusivamente ao frontend (`/apps/web`) e ao mobile futuro (`apps/mobile`). O backend Go possui seus próprios pacotes internos dentro de `/apps/api/internal/`.

### 4.1 `libs/shared-types`

Único source of truth para tipos usados entre Web e Worker. Evita dessincronização de contratos com a API REST.

```text
/shared-types
  /src
    /entities/          # Tipos espelhando as entidades do banco
      user.types.ts
      company.types.ts
      schedule.types.ts
      pet.types.ts
      ...
    /dtos/              # Data Transfer Objects (entrada e saída de API)
      auth.dto.ts
      schedule.dto.ts
      ...
    /enums/             # Espelho dos enums do banco
      user-role.enum.ts
      schedule-status.enum.ts
      plan-package.enum.ts
      ...
    index.ts
  package.json
  tsconfig.json
```

### 4.2 `libs/shared-utils`

Funções puras, sem estado, sem dependência de framework. Testáveis de forma isolada.

```text
/shared-utils
  /src
    /formatters/
      currency.ts         # Formatar NUMERIC(12,2) → "R$ 1.290,00"
      date.ts             # Wrappers de date-fns para pt-BR
      cpf-cnpj.ts         # Máscaras e validação
    /validators/
      cpf.ts
      cnpj.ts
      phone.ts
    /pagination/
      cursor.ts           # Helpers de paginação cursor-based
      offset.ts
    /tenant/
      guard.ts            # Utilitário de validação de company_id
    index.ts
```

### 4.3 `libs/shared-constants`

Constantes de domínio que precisam ser consistentes entre front e back.

```text
/shared-constants
  /src
    plans.constants.ts       # Limites por módulo/pacote (max_users, etc.)
    error-codes.constants.ts # Códigos de erro padronizados da API
    routes.constants.ts      # Rotas da API (evita magic strings no frontend)
    pagination.constants.ts  # PAGE_SIZE_DEFAULT = 20, MAX_PAGE_SIZE = 100
    index.ts
```

Além das rotas vivas da aplicação, `shared-constants` também pode expor padrões de navegação ainda em adoção incremental, como a convenção de rotas com `company_slug`. Quando isso acontecer, a regra funcional deve ser documentada em `docs/conventions/`, e o código deve conter apenas o contexto mínimo necessário para uso seguro.

### 4.4 `libs/ui` — Design System Cross-Platform

```text
/ui
  /core/                    # Lógica pura sem JSX de plataforma
    /hooks/
      useDebounce.ts
      usePagination.ts
      useForm.ts
    /tokens/
      colors.ts
      spacing.ts
      typography.ts
      breakpoints.ts
    /utils/
      cn.ts                 # clsx + tailwind-merge

  /web/                     # Componentes React DOM
    /components/
      /ui/                  # Primitivos (Button, Input, Badge, etc.)
      /layout/              # AppShell, Sidebar, Header
      /data-display/        # Table, Card, StatCard, Timeline
      /feedback/            # Toast, Alert, Skeleton, Spinner
      /forms/               # FormField, Select, DatePicker, FileUpload
      /charts/              # Wrappers de Recharts para dashboards

  /native/                  # Componentes React Native
    /components/
      /ui/                  # Button, Input, Badge (NativeWind)
      /layout/              # ScreenWrapper, BottomTabBar
      /feedback/            # Toast, Skeleton

  package.json
  tsconfig.json
```

---

## 5. App: API

### 5.1 Estrutura do projeto Go

```text
/apps/api
  /cmd
    /server
      main.go               # Entrypoint: carrega config, inicializa DB, router e HTTP server

  /internal
    /config/                # Leitura de variáveis de ambiente com godotenv
      config.go

    /db/                    # Camada de banco de dados
      /sqlc/                # Código gerado pelo SQLC (nunca editar manualmente)
        db.go
        models.go
        querier.go
        schedules.sql.go
        users.sql.go
        clients.sql.go
        pets.sql.go
        ...
      pool.go               # Inicialização do pgxpool.Pool

    /middleware/            # Middlewares Gin
      auth.go               # Valida JWT e injeta claims no context
      tenant.go             # Extrai e valida company_id do JWT
      plan.go               # Verifica módulo ativo para o tenant
      audit.go              # Intercepta mutations e registra em audit_logs

    /handler/               # Handlers HTTP por domínio
      auth.go
      users.go
      companies.go
      schedules.go
      clients.go
      pets.go
      services.go
      products.go
      notifications.go
      reports.go
      uploads.go
      audit.go

    /service/               # Lógica de negócio (chamada pelos handlers)
      auth.go
      schedules.go
      clients.go
      pets.go
      subscriptions.go
      reports.go

    /queue/                 # Publicação de jobs Asynq
      client.go             # asynq.Client singleton
      tasks.go              # Constantes de nome de task e payloads tipados

    /storage/               # Integração com Google Cloud Storage
      gcs.go

    /jwt/                   # Geração e validação de tokens JWT
      jwt.go

    /validator/             # Setup do go-playground/validator com regras customizadas
      validator.go

    /apperror/              # Erros de domínio padronizados → HTTP status codes
      errors.go

  /docs/                    # Gerado pelo swag (nunca editar manualmente)
    docs.go
    swagger.json
    swagger.yaml

  /query/                   # Arquivos SQL fonte do SQLC
    schedules.sql
    users.sql
    clients.sql
    pets.sql
    ...

  /test/
    /unit/
    /integration/           # TestContainers: PostgreSQL real

  sqlc.yaml                 # Configuração do SQLC
  go.mod
  go.sum
  Dockerfile
  .env.example
```

### 5.2 SQLC — configuração e uso

O SQLC lê arquivos `.sql` em `/query/` e gera código Go type-safe em `/internal/db/sqlc/`. Migrations são gerenciadas separadamente via `golang-migrate`.

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "./query"
    schema: "../../infra/migrations"
    gen:
      go:
        package: "sqlc"
        out: "./internal/db/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
```

Exemplo de query SQL:

```sql
-- query/schedules.sql

-- name: GetSchedulesByCompany :many
SELECT *
FROM schedules
WHERE company_id = @company_id
  AND deleted_at IS NULL
ORDER BY starts_at DESC
LIMIT @limit OFFSET @offset;

-- name: CreateSchedule :one
INSERT INTO schedules (
  id, company_id, client_id, pet_id, service_id,
  employee_id, starts_at, ends_at, status, notes
) VALUES (
  @id, @company_id, @client_id, @pet_id, @service_id,
  @employee_id, @starts_at, @ends_at, @status, @notes
)
RETURNING *;
```

Uso do código gerado no service:

```go
// internal/service/schedules.go
func (s *ScheduleService) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]sqlc.Schedule, error) {
    return s.queries.GetSchedulesByCompany(ctx, sqlc.GetSchedulesByCompanyParams{
        CompanyID: companyID,
        Limit:     limit,
        Offset:    offset,
    })
}
```

### 5.3 Roteamento com Gin

```go
// cmd/server/main.go (trecho)
r := gin.New()
r.Use(gin.Recovery(), middleware.Logger())

// Docs Swagger
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// Rotas públicas
r.POST("/auth/login", authHandler.Login)

// Rotas autenticadas
api := r.Group("/api/v1")
api.Use(middleware.Auth(jwtService))
api.Use(middleware.Tenant())
{
    // Agendamentos — requer módulo SCH
    sch := api.Group("/schedules")
    sch.Use(middleware.RequireModule("SCH"))
    sch.GET("", scheduleHandler.List)
    sch.POST("", scheduleHandler.Create)
    sch.GET("/:id", scheduleHandler.GetByID)
    sch.PUT("/:id", scheduleHandler.Update)
    sch.DELETE("/:id", scheduleHandler.SoftDelete)

    // Clientes — requer módulo CLI
    clients := api.Group("/clients")
    clients.Use(middleware.RequireModule(queries, "CLI"))
    clients.GET("", clientHandler.List)
    clients.POST("", clientHandler.Create)
    // ...
}
```

### 5.4 Multi-tenancy na API

Todo request autenticado carrega `company_id` no JWT. O middleware `tenant.go` extrai e valida esse valor, injetando-o no `gin.Context`. O middleware `plan.go` verifica se o módulo necessário está ativo para o tenant.

```go
// internal/middleware/tenant.go
func Tenant() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := GetClaims(c)
        if claims.CompanyID == uuid.Nil {
            c.AbortWithStatusJSON(http.StatusForbidden, apperror.ErrNoTenant)
            return
        }
        c.Set("company_id", claims.CompanyID)
        c.Next()
    }
}

// Helpers usados nos handlers
func CompanyID(c *gin.Context) uuid.UUID {
    return c.MustGet("company_id").(uuid.UUID)
}
```

### 5.5 Auditoria automática via Middleware

O middleware de auditoria persiste automaticamente em `audit_logs` as entradas acumuladas durante a request, incluindo `old_data` e `new_data` em JSONB.

No estado atual da implementação, os handlers de mutação ainda registram explicitamente cada `AuditEntry` via `middleware.AddAuditEntry(...)`, e o middleware cuida da persistência ao final da request.

### 5.6 Swagger com Swaggo

Swagger está integrado na API e usa anotações dos handlers de `auth`, `clients`, `pets`, `services` e `schedules`.

Rota pública local:

- Rota canônica: `GET /swagger/index.html`
- JSON bruto canônico: `GET /swagger/doc.json`
- Alias compatível: `GET /api/v1/docs`
- Alias do JSON: `GET /api/v1/docs/doc.json`

Gerar/atualizar docs:

```bash
cd apps/api
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/server,internal/handler,internal/middleware,internal/service --output docs
```

Arquivos gerados:

- `apps/api/docs/docs.go`
- `apps/api/docs/swagger.json`
- `apps/api/docs/swagger.yaml`

### 5.7 Dockerfile e Entrypoint da API

O backend em `apps/api` já possui conteinerização implementada com build multi-stage.

Resumo do que foi implementado:

- Stage de build em Go (`golang:1.26.1-alpine3.23`) para:
  - instalar `sqlc` e `migrate`;
  - executar `sqlc generate` dentro de `apps/api`;
  - compilar o binário `./cmd/server` em modo estático (`CGO_ENABLED=0`).
- Stage de runtime em Alpine (`alpine:3.23`) com:
  - `ca-certificates` e `tzdata`;
  - binário final da API;
  - binário `migrate`;
  - migrations copiadas de `infra/migrations`.
- `entrypoint.sh` como ponto de entrada do container:
  - se `RUN_MIGRATIONS=true`, roda `migrate up` antes da aplicação subir;
  - em seguida inicia a API (`/app/petcontrol-api`).

Arquivos relevantes:

- `apps/api/Dockerfile`
- `apps/api/entrypoint.sh`

Variáveis usadas no startup do container:

- `DATABASE_URL` (obrigatória para conexão com Postgres e migrations)
- `RUN_MIGRATIONS` (opcional, default `false`)

Exemplo de uso local:

```bash
# na raiz do monorepo
docker build -f apps/api/Dockerfile -t petcontrol-api .

docker run --rm \
  -e DATABASE_URL="postgres://petcontrol:petcontrol@host.docker.internal:5432/petcontrol?sslmode=disable" \
  -e RUN_MIGRATIONS=true \
  -p 8080:8080 \
  petcontrol-api
```

---

## 6. App: Web

### 6.1 Estrutura

```text
/apps/web
  /src
    /main.tsx
    /router.tsx                 # TanStack Router: definição de rotas

    /routes/                    # File-based routing (TanStack Router)
      /(auth)/
        login.tsx
        forgot-password.tsx
      /(app)/
        _layout.tsx             # AppShell com sidebar
        dashboard/
          index.tsx
        schedules/
          index.tsx
          $scheduleId.tsx
        clients/
        pets/
        services/
        reports/
      (admin)/
        companies/
        plans/
        users/

    /lib/
      /api/
        rest-client.ts          # fetch configurado com base URL e auth headers
      /auth/
        auth.store.ts           # Zustand: token, user, company_id
      /queries/                 # TanStack Query: hooks de data fetching
        schedules.queries.ts
        clients.queries.ts
        ...

    /stores/                    # Zustand: apenas estado de UI
      ui.store.ts

    /features/                  # Feature slices
      /schedules/
        ScheduleForm.tsx
        ScheduleCalendar.tsx
        useScheduleForm.ts
      /clients/
      /pets/
      /reports/

    /assets/
  /public/
  vite.config.ts
  tailwind.config.ts
  tsconfig.json
```

### 6.2 Gerenciamento de estado

| Tipo de estado                      | Solução                                    |
| ----------------------------------- | ------------------------------------------ |
| Dados do servidor                   | TanStack Query (cache, background refetch) |
| Estado de UI (sidebar, modal, tema) | Zustand                                    |
| Estado de formulário                | React Hook Form                            |
| Estado de URL (filtros, paginação)  | TanStack Router search params              |
| Auth (token, user)                  | Zustand + persistência em `localStorage`   |

> **Regra:** nenhum dado que vem do servidor entra no Zustand. Se veio da API, fica no TanStack Query.

### 6.3 Convenção de `company_slug` nas rotas autenticadas

Foi introduzida uma convenção de roteamento para explicitar o tenant também na URL do frontend Web, usando o `company_slug` antes dos recursos autenticados.

Rotas atualmente ativas:

- `/:companySlug/dashboard`
- `/:companySlug/schedules`
- `/:companySlug/clients`
- `/:companySlug/pets`
- `/:companySlug/services`

Essa convenção foi documentada em [docs/conventions/company-slug-routing.md](./docs/conventions/company-slug-routing.md) e detalhada no plano [docs/plans/0003-COMPANY_SLUG_ROUTING_PLAN.md](./docs/plans/0003-COMPANY_SLUG_ROUTING_PLAN.md).

Regras importantes:

- o `company_slug` na URL representa contexto de navegação e UX;
- a autorização continua pertencendo ao backend via JWT e `company_id`;
- `/login` permanece sem slug;
- a URL autenticada deve ser tratada como canônica em lowercase;
- o header da área autenticada deve deixar explícitos o tenant resolvido e o slug atual;
- a migração do router foi incremental; novas rotas autenticadas devem seguir o mesmo formato `/:companySlug/<recurso>`.

Direção futura para mudança de slug:

- o frontend continua resolvendo a empresa corrente pela sessão e corrige a URL quando houver divergência;
- se links antigos precisarem continuar válidos após troca de slug, a compatibilidade ideal deve ser tratada no backend ou em uma camada dedicada de redirecionamento.

Essa evolução não fazia parte do plano original de módulos funcionais da aplicação. Ela foi introduzida depois como melhoria de navegação, previsibilidade de URL e preparação para cenários multi-tenant mais explícitos no frontend.

---

## 7. App: Worker

### 7.1 Decisão de arquitetura

O Worker é um **processo Go standalone separado**, não embutido no `api`. Isso permite escalar workers independentemente, reiniciá-los sem afetar a API, e ter políticas de retry e concorrência configuradas por domínio.

A comunicação é via **filas Asynq no Redis** — a API publica tasks, o Worker as consome.

### 7.2 Estrutura

```text
/apps/worker
  /cmd
    /worker
      main.go               # Bootstrap: conecta Redis, registra handlers, inicia servidor Asynq

  /internal
    /config/
      config.go

    /db/                    # Reusa o SQLC gerado (via symlink ou módulo compartilhado)
      pool.go

    /processor/             # Handlers de tasks Asynq (um arquivo por domínio)
      notifications.go      # Processa envio de WhatsApp/e-mail/push
      subscriptions.go      # Expiração e renovação de assinaturas
      reports.go            # Geração assíncrona de PDFs/XLSX
      cleanup.go            # Sessões, tokens e soft-deletes antigos

    /scheduler/             # Cron jobs com robfig/cron
      subscription_checker.go   # Diário: assinaturas a vencer
      session_cleanup.go         # Horário: sessões expiradas
      report_aggregator.go       # Semanal: agrega métricas

    /whatsapp/              # Cliente HTTP para WhatsApp Business API
      client.go

  go.mod
  go.sum
  Dockerfile
```

### 7.3 Filas e responsabilidades

| Fila            | Publicado por         | Consumido pelo Worker    | Descrição                              |
| --------------- | --------------------- | ------------------------ | -------------------------------------- |
| `notifications` | API (qualquer módulo) | `NotificationsProcessor` | Envio de WhatsApp, e-mail, push        |
| `subscriptions` | API + Cron            | `SubscriptionsProcessor` | Expiração, renovação, trial-end        |
| `reports`       | API (módulo reports)  | `ReportsProcessor`       | Geração assíncrona de PDFs/XLSX        |
| `cleanup`       | Cron                  | `CleanupProcessor`       | Sessões, tokens e soft-deletes antigos |

### 7.4 Exemplo: publicar task na API

```go
// internal/queue/tasks.go
const (
    TypeScheduleConfirmed    = "schedule:confirmed"
    TypeSubscriptionExpiring = "subscription:expiring"
)

type ScheduleConfirmedPayload struct {
    ScheduleID uuid.UUID `json:"schedule_id"`
    CompanyID  uuid.UUID `json:"company_id"`
    ClientID   uuid.UUID `json:"client_id"`
}

// internal/service/schedules.go
func (s *ScheduleService) Confirm(ctx context.Context, scheduleID uuid.UUID) error {
    // ... lógica de negócio

    payload, _ := json.Marshal(queue.ScheduleConfirmedPayload{
        ScheduleID: scheduleID,
        CompanyID:  companyID,
        ClientID:   clientID,
    })

    task := asynq.NewTask(queue.TypeScheduleConfirmed, payload,
        asynq.MaxRetry(3),
        asynq.Timeout(30*time.Second),
    )

    _, err := s.queueClient.Enqueue(task)
    return err
}
```

### 7.5 Exemplo: consumir task no Worker

```go
// internal/processor/notifications.go
type NotificationsProcessor struct {
    whatsapp *whatsapp.Client
    queries  *sqlc.Queries
}

func (p *NotificationsProcessor) HandleScheduleConfirmed(ctx context.Context, t *asynq.Task) error {
    var payload queue.ScheduleConfirmedPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("json.Unmarshal: %w", err)
    }

    // buscar dados do agendamento e enviar WhatsApp
    schedule, err := p.queries.GetScheduleByID(ctx, payload.ScheduleID)
    if err != nil {
        return err
    }

    return p.whatsapp.SendConfirmation(ctx, schedule)
}

// cmd/worker/main.go — registro dos handlers
mux := asynq.NewServeMux()
mux.HandleFunc(queue.TypeScheduleConfirmed, notifProcessor.HandleScheduleConfirmed)
mux.HandleFunc(queue.TypeSubscriptionExpiring, subscProcessor.HandleSubscriptionExpiring)
```

---

## 8. Infraestrutura e DevOps

### 8.1 Docker Compose (desenvolvimento)

```yaml
# infra/docker/docker-compose.yml
services:
  postgres:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: petcontrol
      POSTGRES_USER: petcontrol
      POSTGRES_PASSWORD: petcontrol
    ports:
      - "5432:5432"
    volumes:
      - postgres18_data:/var/lib/postgresql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --save 60 1 --loglevel warning

  pgadmin:
    image: dpage/pgadmin4
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@petcontrol.dev
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"
    depends_on:
      - postgres

  asynqmon:
    image: hibiken/asynqmon    # UI para monitorar filas Asynq
    environment:
      REDIS_ADDR: redis:6379
    ports:
      - "8081:8081"
    depends_on:
      - redis

volumes:
  postgres18_data:
```

### 8.2 Makefile — pipeline de tasks

```makefile
# Makefile (raiz do monorepo)

.PHONY: dev test lint build migrate sqlc swagger

# Desenvolvimento
dev:
 # sobe api, worker e web juntos

dev-api:
 cd apps/api && go run ./cmd/server

dev-worker:
 cd apps/worker && go run ./cmd/worker

dev-web:
 cd apps/web && pnpm dev

# Build
build-api:
 cd apps/api && go build -o bin/server ./cmd/server

build-worker:
 cd apps/worker && go build -o bin/worker ./cmd/worker

# Testes
test-api:
 cd apps/api && go test ./... -v -race -count=1

test-web:
 cd apps/web && pnpm test

# Lint
lint-go:
 golangci-lint run ./apps/api/... ./apps/worker/...

lint-web:
 cd apps/web && pnpm lint

# Migrations
migrate-up:
 migrate -path infra/migrations -database "$$DATABASE_URL" up

migrate-down:
 migrate -path infra/migrations -database "$$DATABASE_URL" down 1

migrate-create:
 migrate create -ext sql -dir infra/migrations -seq $(name)

# SQLC
sqlc:
 cd apps/api && sqlc generate

# Swagger
swagger:
 cd apps/api && swag init -g cmd/server/main.go --output docs/
```

### 8.3 Variáveis de ambiente

```bash
# .env.example

# JWT
JWT_SECRET=change-me
JWT_ACCESS_TOKEN_TTL=15m

# App
API_PORT=8080
API_HOST=0.0.0.0
APP_ENV=development
CORS_ALLOWED_ORIGINS=http://localhost:5173
WORKER_CONCURRENCY=5
WORKER_QUEUE=default

# Database
DATABASE_URL=postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=petcontrol
POSTGRES_USER=petcontrol
POSTGRES_PASSWORD=petcontrol

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_HOST=localhost
REDIS_PORT=6379

# Web
VITE_API_URL=http://localhost:8080/api/v1
VITE_AUTH_MODE=api

# WhatsApp Business API
WHATSAPP_API_URL=
WHATSAPP_API_TOKEN=

# Google Cloud Storage
GCS_BUCKET_NAME=
GCS_UPLOADS_BASE_PATH=uploads
GCS_PUBLIC_BASE_URL=
GCS_SIGNED_URL_TTL_SECONDS=900
GCS_CREDENTIALS_FILE=
GCS_SIGNER_SERVICE_ACCOUNT_EMAIL=
GCS_SIGNER_PRIVATE_KEY=
```

Notas de configuração de upload:

- `GCS_BUCKET_NAME` ou `GCS_BUCKET`: nome do bucket alvo.
- `GCS_CREDENTIALS_FILE`: caminho para o JSON de credenciais usado pelo SDK e, quando necessário, como fallback para signer.
- `GCS_SIGNER_SERVICE_ACCOUNT_EMAIL` e `GCS_SIGNER_PRIVATE_KEY`: opcionais para assinar URLs explicitamente por env.
- `GCS_PUBLIC_BASE_URL`: opcional para servir URL canônica por CDN/domínio próprio.
- `GCS_UPLOADS_BASE_PATH`: prefixo lógico dos objetos no bucket.
- `GCS_SIGNED_URL_TTL_SECONDS`: TTL das signed URLs de upload.

### 8.4 CI/CD — GitHub Actions

```yaml
# .github/workflows/go.yml
name: Go CI
on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:18-alpine
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_USER: test
          POSTGRES_DB: test
        ports: ["5432:5432"]
      redis:
        image: redis:7-alpine
        ports: ["6379:6379"]

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/api/go.mod

      - name: Test API
        run: make test-api
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/test
          REDIS_URL: redis://localhost:6379/0

# .github/workflows/frontend.yml
name: Frontend CI
on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  lint-test-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
      - uses: actions/setup-node@v4
        with:
          node-version: lts/*
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm --filter @petcontrol/shared-types build
      - run: pnpm --filter @petcontrol/shared-utils build
      - run: pnpm --filter @petcontrol/shared-constants build
      - run: pnpm --filter @petcontrol/ui build
      - run: pnpm --filter web lint
      - run: pnpm --filter web test
      - run: pnpm --filter web build
```

---

## 9. Estratégia de Testes

### 9.1 Pirâmide de testes

```text
         /\
        /E2E\          Poucos, lentos, alto valor de confiança
       /------\
      / Integr.\       Handlers + banco real (TestContainers)
     /----------\
    /   Unitários\     Muitos, rápidos, lógica de negócio isolada
   /--------------\
```

### 9.2 Testes unitários

- Lógica dos services com dependências mockadas via interfaces Go
- Validators, formatters e utils de `shared-utils` (TS)
- Middlewares de autenticação e tenant

```go
// internal/service/schedules_test.go
type mockQuerier struct{ sqlc.Querier }

func (m *mockQuerier) GetSchedulesByCompany(ctx context.Context, arg sqlc.GetSchedulesByCompanyParams) ([]sqlc.Schedule, error) {
    // retorno mockado
}

func TestScheduleService_ListByCompany(t *testing.T) {
    svc := NewScheduleService(mockPool, sqlc.New(mockPool))
    result, err := svc.ListByCompany(context.Background(), testCompanyID, 20, 0)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 9.3 Testes de integração (TestContainers)

```go
// test/integration/schedules_integration_test.go
import (
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    pgContainer, _ := postgres.Run(ctx, "postgres:18-alpine",
        postgres.WithDatabase("petcontrol_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections"),
        ),
    )
    defer pgContainer.Terminate(ctx)

    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")

    // Rodar migrations no container
    runMigrations(connStr)

    // Inicializar pool e queries
    pool, _ := pgxpool.New(ctx, connStr)
    queries = sqlc.New(pool)

    os.Exit(m.Run())
}

func TestCreateSchedule(t *testing.T) {
    // teste com banco real
}
```

> **Importante:** usar `golang-migrate` com `migrate up` nos testes de integração — nunca criar schema manualmente nos testes.

### 9.4 Cobertura mínima por camada

| Camada                          | Meta de cobertura |
| ------------------------------- | ----------------- |
| `shared-utils` (TS)             | 95%+              |
| Services Go (lógica de negócio) | 80%+              |
| Handlers Go (integração)        | 70%+              |
| Middlewares Go                  | 80%+              |

---

## 10. ADRs — Decisões de Arquitetura

> Os ADRs documentam *por que* uma decisão foi tomada, não apenas *o que* foi decidido.
> Cada ADR fica em `docs/adr/ADR-NNN-titulo.md`.

### ADR-001: Go + Gin no lugar de NestJS + Fastify

**Contexto:** O projeto SaaS multi-tenant exige alto throughput, baixa latência e deploy eficiente em produção.

**Decisão:** Substituir NestJS/Fastify por Go com Gin como framework HTTP.

**Justificativa:** Go compila para um binário estático único (sem runtime Node), tem concorrência nativa via goroutines e consome significativamente menos memória que Node.js em carga equivalente. O modelo de erro explícito do Go favorece código mais previsível em sistemas de produção críticos.

**Consequências:** Curva de aprendizado para equipes habituadas a TypeScript. Ganho: performance, binary size, observabilidade e tempo de startup.

---

### ADR-002: SQLC no lugar de Prisma/ORM

**Contexto:** O projeto tem queries complexas com isolamento por `company_id`, paginação e filtros dinâmicos.

**Decisão:** Usar SQLC para geração de código type-safe a partir de SQL puro. Nenhum ORM é utilizado no backend Go.

**Justificativa:** SQL puro oferece controle total sobre as queries, sem surpresas de N+1, eager loading automático ou geração de SQL inesperado. SQLC garante type-safety em tempo de compilação sem o overhead de reflexão de um ORM. As queries ficam auditáveis e otimizáveis diretamente.

**Consequências:** Mais SQL a escrever manualmente. Ganho: previsibilidade total, queries otimizadas, sem "ORM magic".

---

### ADR-003: Worker separado ao invés de goroutines embutidas na API

**Contexto:** Jobs assíncronos precisam ser processados sem impactar a latência da API.

**Decisão:** `apps/worker` é um binário Go separado, comunicando-se com a API via Redis/Asynq.

**Justificativa:** Permite escalar workers independentemente, reiniciar sem afetar a API, e definir políticas de concorrência e retry por tipo de task. Asynq oferece UI de monitoramento (asynqmon) e garantias de at-least-once delivery.

**Consequências:** Dois processos para gerenciar em produção. Ganho: isolamento e escalabilidade.

---

### ADR-004: Soft delete por padrão em todas as entidades de negócio

**Contexto:** Dados de clientes, pets e histórico são sensíveis e regulados.

**Decisão:** Toda entidade de negócio tem `deleted_at TIMESTAMPTZ`. Deleção física é proibida em produção.

**Justificativa:** Auditabilidade, recuperação de dados acidental, conformidade com LGPD (direito ao esquecimento via anonimização, não deleção física imediata).

**Consequências:** Todas as queries SQLC precisam filtrar `WHERE deleted_at IS NULL`. Implementar via views no banco ou garantir que todos os arquivos `.sql` incluam o filtro.

---

### ADR-005: Zustand apenas para estado de UI (frontend)

**Contexto:** Risco de usar Zustand como cache de servidor, duplicando o papel do TanStack Query.

**Decisão:** Zustand gerencia exclusivamente estado de UI (sidebar, tema, modal stack). Nenhum dado de servidor entra no Zustand.

**Justificativa:** Manter uma única fonte de verdade para os dados da API (TanStack Query) previne bugs de sincronização, stale data e race conditions que surgem frequentemente ao copiar dados async para um state manager global.

**Consequências:** Qualquer componente React que precise de dados do servidor deve acessar via hooks do TanStack Query. A store global fica menor, previsível e dedicada apenas à experiência do usuário, tornando-se mais fácil de debugar e testar.

---

### ADR-006: Módulo `pets` sob a guarda do `PET`

**Contexto:** O projeto possui módulos dedicados no catálogo (`CLI`, `PET`, `SCH`, `SVC` etc.). O recurso `pets` semanticamente se relaciona com `schedules`, mas é um domínio cadastral próprio com autorização por módulo específico.

**Decisão:** O módulo `pets` está formalmente registrado e protegido no código pelo módulo `PET`, mantendo separação explícita no controle de acesso por módulo.

**Justificativa:** Em Petshops e Clínicas Veterinárias, um "Pet" é uma entidade dependente de um "Cliente" (tutor), porém possui ciclo de vida e permissões próprias. A guarda por `PET` elimina acoplamento legado e mantém o catálogo modular coerente com `modules`, `plan_modules` e `company_modules`.

**Consequências:** Todas as rotas base de `/pets` utilizarão `RequireModule(..., "PET")`. O módulo `SCH` continua consumindo dados de pets quando necessário, mas sem herdar a guarda de autorização do cadastro de pets.

### ADR-007: `libs/ui` com camada core agnóstica de plataforma

**Contexto:** Mobile com React Native é planejado para o futuro.

**Decisão:** Design tokens, hooks e lógica de componentes ficam em `libs/ui/core/`. Componentes React DOM ficam em `libs/ui/web/`. Componentes React Native ficam em `libs/ui/native/`.

**Justificativa:** Evita duplicação de lógica quando o mobile for implementado. Tokens compartilhados garantem consistência visual entre plataformas.

---

Documento criado em: 2026-04-07 | Versão: 2.0.0
