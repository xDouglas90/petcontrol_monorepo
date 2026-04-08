ROOT_DIR := $(CURDIR)
API_DIR := $(ROOT_DIR)/apps/api
WORKER_DIR := $(ROOT_DIR)/apps/worker
INFRA_DIR := $(ROOT_DIR)/infra
DOCKER_DIR := $(INFRA_DIR)/docker
COMPOSE_FILE := $(ROOT_DIR)/infra/docker/docker-compose.yml
MIGRATIONS_DIR := $(INFRA_DIR)/migrations
SCRIPTS_DIR := $(INFRA_DIR)/scripts
DATABASE_URL ?=

.PHONY: dev-api test-api dev-worker test-worker test lint lint-go lint-ts sqlc docker-up docker-down docker-logs docker-ps go-work-sync migrate-up migrate-down migrate-create seed

dev-api:
	cd $(API_DIR) && \
		set -a && \
		if [ -f $(ROOT_DIR)/.env ]; then . $(ROOT_DIR)/.env; else . $(ROOT_DIR)/.env.example; fi && \
		set +a && \
		go run ./cmd/server

test-api:
	cd $(API_DIR) && go test ./...

dev-worker:
	cd $(WORKER_DIR) && \
		set -a && \
		if [ -f $(ROOT_DIR)/.env ]; then . $(ROOT_DIR)/.env; else . $(ROOT_DIR)/.env.example; fi && \
		set +a && \
		go run ./cmd/worker

test-worker:
	cd $(WORKER_DIR) && go test ./...

test: test-api test-worker
	@if command -v pnpm >/dev/null 2>&1; then \
		pnpm --filter @petcontrol/shared-utils test && \
		pnpm --filter @petcontrol/shared-constants test && \
		pnpm --filter @petcontrol/ui test && \
		pnpm --filter web test; \
	else \
		echo "pnpm not found, skipping JS/TS tests"; \
	fi

lint: lint-go lint-ts

lint-go:
	@if [ -n "$(shell gofmt -l $(API_DIR) $(WORKER_DIR))" ]; then \
		echo "gofmt check failed:"; \
		gofmt -l $(API_DIR) $(WORKER_DIR); \
		exit 1; \
	fi
	cd $(API_DIR) && go vet ./...
	cd $(WORKER_DIR) && go vet ./...

lint-ts:
	@if command -v pnpm >/dev/null 2>&1; then \
		pnpm --filter web lint; \
	else \
		echo "pnpm not found, skipping TS lint"; \
	fi

sqlc:
	cd $(API_DIR) && sqlc generate

docker-up:
	cd $(DOCKER_DIR) && docker compose up -d

docker-down:
	cd $(DOCKER_DIR) && docker compose down

docker-logs:
	cd $(DOCKER_DIR) && docker compose logs -f

docker-ps:
	cd $(DOCKER_DIR) && docker compose ps

go-work-sync:
	go work sync

migrate-up:
	DATABASE_URL="$(DATABASE_URL)" MIGRATIONS_DIR="$(MIGRATIONS_DIR)" $(SCRIPTS_DIR)/migrate.sh up

migrate-down:
	DATABASE_URL="$(DATABASE_URL)" MIGRATIONS_DIR="$(MIGRATIONS_DIR)" $(SCRIPTS_DIR)/migrate.sh down 1

migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=your_migration_name"; exit 1; fi
	docker run --rm -v "$(MIGRATIONS_DIR):/migrations" migrate/migrate:v4.19.0 create -ext sql -dir /migrations -seq $(name)

seed:
	DATABASE_URL="$(DATABASE_URL)" $(SCRIPTS_DIR)/seed.sh