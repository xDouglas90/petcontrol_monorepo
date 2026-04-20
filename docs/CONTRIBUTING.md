# Contribuindo para o PetControl Monorepo

Este guia descreve o fluxo mínimo para desenvolver e validar alterações localmente.

## Requisitos

- Go 1.26.1
- Node.js LTS
- pnpm 10+
- Docker e Docker Compose

## Setup inicial

1. Copie as variáveis de ambiente:

```bash
cp .env.example .env
```

1. Sincronize o workspace Go:

```bash
go work sync
```

1. Instale dependências JS/TS:

```bash
pnpm install
```

1. Suba a infraestrutura local:

```bash
make docker-up
```

1. Aplique migrations e seed para obter um tenant funcional:

```bash
make migrate-up DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
make seed DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
```

Após o seed, o ambiente local fica com uma massa mínima pronta para uso:

- tenant `petcontrol-dev`;
- usuários `admin@petcontrol.local`, `root@petcontrol.local` e `system@petcontrol.local` com senha `password123`;
- 1 cliente ativo com pet vinculado;
- 1 serviço ativo de catálogo;
- configuração operacional em `company_system_configs`;
- conjunto de agendamentos distribuídos entre hoje, ontem, mês atual e mês anterior para alimentar o dashboard `admin`;
- conversa persistida inicial entre `admin@petcontrol.local` e `system@petcontrol.local`.

Com essa massa, o Web já permite validar:

- o shell autenticado com branding do tenant e card de upgrade;
- a home `admin` em `/:companySlug/dashboard`;
- o núcleo operacional completo em `/:companySlug/clients`, `/:companySlug/pets`, `/:companySlug/services` e `/:companySlug/schedules`.

Credenciais recomendadas para validar a home rica do tenant:

- `admin@petcontrol.local` / `password123` para o dashboard `admin`;
- `system@petcontrol.local` / `password123` para validar o papel `system` seedado;
- `root@petcontrol.local` / `password123` para bootstrap administrativo.

## Comandos de desenvolvimento

Tudo em um unico terminal:

```bash
make dev
```

API:

```bash
make dev-api
```

Worker:

```bash
make dev-worker
```

Frontend:

```bash
pnpm --filter web dev
```

Sequência recomendada para desenvolvimento local:

1. `make dev`
2. Se preferir logs separados, use `make dev-api`, `make dev-worker` e `pnpm --filter web dev`

## Banco de dados

Aplicar migrations:

```bash
make migrate-up DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
```

Reverter ultima migration:

```bash
make migrate-down DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
```

Seed inicial:

```bash
make seed DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
```

Observação sobre rede Docker:

- No Linux, os scripts de migration e seed usam `--network host` por padrão.
- Em macOS/Windows (Docker Desktop), o script ajusta automaticamente `localhost` para `host.docker.internal`.
- Se você precisar forçar uma rede específica do Docker, defina `DOCKER_NETWORK`:

```bash
DOCKER_NETWORK=petcontrol_monorepo_default make migrate-up DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"

DOCKER_NETWORK=petcontrol_monorepo_default make seed DATABASE_URL="postgres://petcontrol:petcontrol@localhost:5432/petcontrol?sslmode=disable"
```

## Qualidade local

Suite minima (Go + JS/TS quando pnpm estiver disponível):

```bash
make test
```

Lint padrão (Go e TypeScript):

```bash
make lint
```

Testes por app (execução direta):

```bash
cd apps/api && go test ./... -count=1
cd apps/worker && go test ./... -count=1
pnpm --filter web test
```

## Swagger (API)

Gerar/atualizar especificação OpenAPI:

```bash
cd apps/api
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/server,internal/handler,internal/middleware,internal/service --output docs
```

A especificação versionada cobre os handlers reais de `auth`, `clients`, `pets`, `services` e `schedules`.

Após subir a API, validar localmente:

- `http://localhost:8080/swagger/index.html`
- `http://localhost:8080/swagger/doc.json`
- `http://localhost:8080/api/v1/docs` (`alias` compatível do Swagger UI)
- `http://localhost:8080/api/v1/docs/doc.json` (`alias` compatível do JSON)

## CI

Workflows principais:

- Go CI: lint/build/test para API e Worker, verificação de sqlc e docker compose.
- Frontend CI: lint/type-check/test/build do app web e testes das libs compartilhadas.

## Padrão de commits

Use Conventional Commits, por exemplo:

- feat(api): adiciona endpoint de notificação
- fix(worker): corrige parsing de payload da task
- docs(plan): atualiza checklists da fase

## Pull Requests

Antes de abrir PR:

1. Rode make lint.
2. Rode make test.
3. Garanta que nao ha segredos em arquivos versionados.
4. Descreva impacto funcional e plano de rollback (se aplicável).
