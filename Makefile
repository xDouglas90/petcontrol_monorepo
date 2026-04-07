ROOT_DIR := $(CURDIR)
API_DIR := $(ROOT_DIR)/apps/api
INFRA_DIR := $(ROOT_DIR)/infra
DOCKER_DIR := $(INFRA_DIR)/docker
COMPOSE_FILE := $(ROOT_DIR)/infra/docker/docker-compose.yml

.PHONY: dev-api test-api sqlc docker-up docker-down docker-logs docker-ps go-work-sync

dev-api:
	cd $(API_DIR) && go run ./cmd/server

test-api:
	cd $(API_DIR) && go test ./...

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