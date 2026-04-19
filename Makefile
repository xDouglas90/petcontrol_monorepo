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
	dev-web \
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
	@cd $(API_DIR) && \
		set -a && \
		if [ -f $(ROOT_DIR)/.env ]; then . $(ROOT_DIR)/.env; else . $(ROOT_DIR)/.env.example; fi && \
		set +a && \
		go run ./cmd/server

test-api:
	cd $(API_DIR) && go test ./...

dev-worker:
	@cd $(WORKER_DIR) && \
		set -a && \
		if [ -f $(ROOT_DIR)/.env ]; then . $(ROOT_DIR)/.env; else . $(ROOT_DIR)/.env.example; fi && \
		set +a && \
		go run ./cmd/worker

dev-web:
	@if command -v pnpm >/dev/null 2>&1; then \
		cd $(ROOT_DIR) && pnpm --filter web dev; \
	else \
		echo "pnpm not found, unable to start web"; \
		exit 1; \
	fi

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
	@printf '%s\n' "Starting API, Worker and Web in the same terminal" "Press Ctrl+C to stop all services."
	@pids=""; interrupted=""; \
	set -a; \
	if [ -f "$(ROOT_DIR)/.env" ]; then . "$(ROOT_DIR)/.env"; else . "$(ROOT_DIR)/.env.example"; fi; \
	set +a; \
	api_port="$${API_PORT:-8080}"; \
	worker_addr="$${WORKER_HTTP_ADDR:-:8091}"; \
	worker_port="$${worker_addr##*:}"; \
	port_in_use() { \
		port="$$1"; \
		if command -v lsof >/dev/null 2>&1; then \
			lsof -nP -iTCP:"$$port" -sTCP:LISTEN >/dev/null 2>&1; \
			return $$?; \
		fi; \
		if command -v ss >/dev/null 2>&1; then \
			ss -ltn | awk 'NR > 1 {print $$4}' | grep -Eq "(^|:)$$port$$"; \
			return $$?; \
		fi; \
		return 1; \
	}; \
	if port_in_use "$$api_port"; then \
		printf 'Cannot start API: port %s is already in use.\n' "$$api_port"; \
		exit 1; \
	fi; \
	if [ -n "$$worker_port" ] && port_in_use "$$worker_port"; then \
		printf 'Cannot start Worker webhook: port %s is already in use.\n' "$$worker_port"; \
		exit 1; \
	fi; \
	start_service() { \
		target="$$1"; \
		if command -v setsid >/dev/null 2>&1; then \
			setsid $(MAKE) --no-print-directory "$$target" & \
		else \
			$(MAKE) --no-print-directory "$$target" & \
		fi; \
		pid="$$!"; \
		pids="$$pids $$pid"; \
	}; \
	stop_service() { \
		pid="$$1"; \
		if command -v setsid >/dev/null 2>&1; then \
			kill -TERM -$$pid 2>/dev/null || kill $$pid 2>/dev/null || true; \
		else \
			kill $$pid 2>/dev/null || true; \
		fi; \
		wait $$pid 2>/dev/null || true; \
	}; \
	wait_for_stable_start() { \
		pid="$$1"; \
		label="$$2"; \
		seconds="$$3"; \
		count=0; \
		while [ "$$count" -lt "$$seconds" ]; do \
			if ! kill -0 $$pid 2>/dev/null; then \
				wait $$pid; \
				status="$$?"; \
				printf '%s failed to start.\n' "$$label"; \
				exit "$$status"; \
			fi; \
			sleep 1; \
			count=$$((count + 1)); \
		done; \
	}; \
	cleanup() { \
		status=$$?; \
		if [ -n "$$interrupted" ]; then \
			status=0; \
		fi; \
		if [ -n "$$pids" ]; then \
			for pid in $$pids; do \
				stop_service $$pid; \
			done; \
		fi; \
		trap - INT TERM EXIT; \
		exit $$status; \
	}; \
	trap 'interrupted=1; cleanup' INT TERM; \
	trap cleanup EXIT; \
	start_service dev-api; \
	wait_for_stable_start "$$pid" "API" 2; \
	start_service dev-worker; \
	wait_for_stable_start "$$pid" "Worker" 2; \
	start_service dev-web; \
	wait_for_stable_start "$$pid" "Web" 2; \
	while :; do \
		for pid in $$pids; do \
			if ! kill -0 $$pid 2>/dev/null; then \
				wait $$pid; \
				exit $$?; \
			fi; \
		done; \
		sleep 1; \
	done

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
