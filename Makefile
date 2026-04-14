ROOT_DIR := $(CURDIR)
API_DIR := $(ROOT_DIR)/apps/api
WORKER_DIR := $(ROOT_DIR)/apps/worker
INFRA_DIR := $(ROOT_DIR)/infra
DOCKER_DIR := $(INFRA_DIR)/docker
COMPOSE_FILE := $(ROOT_DIR)/infra/docker/docker-compose.yml
MIGRATIONS_DIR := $(INFRA_DIR)/migrations
SCRIPTS_DIR := $(INFRA_DIR)/scripts
DATABASE_URL ?=
GO_CACHE_DIR ?= /tmp/petcontrol-go-cache

# Coverage gates intentionally focus on deterministic unit-level packages.
API_COVERAGE_MIN ?= 70
WORKER_COVERAGE_MIN ?= 60
API_COVERAGE_PACKAGES := ./internal/apperror ./internal/jwt ./internal/middleware
WORKER_COVERAGE_PACKAGES := ./internal/config ./internal/queue ./internal/whatsapp

.PHONY: \
	build \
	dev \
	dev-api \
	dev-worker \
	diagrams \
	diagrams-check \
	docker-down \
	docker-logs \
	docker-ps \
	docker-up \
	go-work-sync \
	lint \
	lint-go \
	lint-ts \
	migrate-create \
	migrate-down \
	migrate-up \
	seed \
	sqlc \
	test \
	test-api \
	test-worker \
	coverage \
	coverage-api \
	coverage-worker
	db-reset

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
		pnpm --filter @petcontrol/shared-types test && \
		pnpm --filter @petcontrol/shared-utils test && \
		pnpm --filter @petcontrol/shared-constants test && \
		pnpm --filter @petcontrol/ui test && \
		pnpm --filter web test; \
	else \
		echo "pnpm not found, skipping JS/TS tests"; \
	fi

build:
	cd $(API_DIR) && go build ./...
	cd $(WORKER_DIR) && go build ./...
	@if command -v pnpm >/dev/null 2>&1; then \
		pnpm --filter @petcontrol/shared-types build && \
		pnpm --filter @petcontrol/shared-utils build && \
		pnpm --filter @petcontrol/shared-constants build && \
		pnpm --filter @petcontrol/ui build && \
		pnpm --filter web build; \
	else \
		echo "pnpm not found, skipping TS/JS builds"; \
	fi

dev:
	@echo "Starting API, Worker and Web (use separate terminals for logs)"
	@echo "Run 'make dev-api', 'make dev-worker' and 'pnpm --filter web dev' in separate terminals for local development."

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

coverage: coverage-api coverage-worker

coverage-api:
	cd $(API_DIR) && \
		go test -coverprofile=coverage.out $(API_COVERAGE_PACKAGES) && \
		total="$$(go tool cover -func=coverage.out | awk '/^total:/ {print substr($$3, 1, length($$3)-1)}')" && \
		awk -v total="$$total" -v min="$(API_COVERAGE_MIN)" 'BEGIN { \
			if (total + 0 < min + 0) { \
				printf "API coverage %.1f%% is below minimum %.1f%%\n", total, min; \
				exit 1; \
			} \
			printf "API coverage %.1f%% meets minimum %.1f%%\n", total, min; \
		}'

coverage-worker:
	cd $(WORKER_DIR) && \
		go test -coverprofile=coverage.out $(WORKER_COVERAGE_PACKAGES) && \
		total="$$(go tool cover -func=coverage.out | awk '/^total:/ {print substr($$3, 1, length($$3)-1)}')" && \
		awk -v total="$$total" -v min="$(WORKER_COVERAGE_MIN)" 'BEGIN { \
			if (total + 0 < min + 0) { \
				printf "Worker coverage %.1f%% is below minimum %.1f%%\n", total, min; \
				exit 1; \
			} \
			printf "Worker coverage %.1f%% meets minimum %.1f%%\n", total, min; \
		}'

sqlc:
	cd $(API_DIR) && sqlc generate

diagrams:
	cd $(API_DIR) && GOCACHE=$(GO_CACHE_DIR) go run ./cmd/erdiagram -input ../../infra/migrations/000001_init_schema.up.sql > $(ROOT_DIR)/docs/diagrams/er-diagram.mmd

diagrams-check:
	GOCACHE=$(GO_CACHE_DIR) go run github.com/sammcj/go-mermaid/cmd/go-mermaid@v0.0.2 docs/diagrams/er-diagram.mmd

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

db-reset:
	@$(MAKE) migrate-down DATABASE_URL="$(DATABASE_URL)"
	@$(MAKE) migrate-up DATABASE_URL="$(DATABASE_URL)"
	@$(MAKE) seed DATABASE_URL="$(DATABASE_URL)"
