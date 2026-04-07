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

## Comandos de desenvolvimento

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

## Banco de dados

Aplicar migrations:

```bash
make migrate-up DATABASE_URL="postgres://postgres:postgres@localhost:5432/petcontrol_dev?sslmode=disable"
```

Reverter ultima migration:

```bash
make migrate-down DATABASE_URL="postgres://postgres:postgres@localhost:5432/petcontrol_dev?sslmode=disable"
```

Seed inicial:

```bash
make seed DATABASE_URL="postgres://postgres:postgres@localhost:5432/petcontrol_dev?sslmode=disable"
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
