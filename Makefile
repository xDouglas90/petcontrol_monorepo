ROOT_DIR := $(CURDIR)
API_DIR := $(ROOT_DIR)/apps/api
COMPOSE_FILE := $(ROOT_DIR)/infra/docker/docker-compose.yml

.PHONY: dev-api test-api sqlc docker-up docker-down docker-logs docker-ps go-work-sync

dev-api:
	cd $(API_DIR) && go run ./cmd/server

test-api:
	cd $(API_DIR) && go test ./...

sqlc:
	cd $(API_DIR) && sqlc generate

docker-up:
	docker compose -f $(COMPOSE_FILE) up -d

docker-down:
	docker compose -f $(COMPOSE_FILE) down

docker-logs:
	docker compose -f $(COMPOSE_FILE) logs -f

docker-ps:
	docker compose -f $(COMPOSE_FILE) ps

go-work-sync:
	go work sync