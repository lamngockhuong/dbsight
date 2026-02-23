.PHONY: help build dev dev-docs build-docs docker-build docker-up docker-down migrate generate-key lint test fmt-md

BINARY=bin/dbsight
IMAGE=dbsight:latest

help: ## Show all available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "} {printf "  %-20s %s\n", $$1, $$2}' | sort

build: ## Build frontend and Go binary
	pnpm --filter web build
	go build -o $(BINARY) .

dev: ## Start development environment (PostgreSQL + Go server + Vite)
	docker-compose up -d postgres
	go run . serve &
	pnpm --filter web dev

dev-docs: ## Start documentation development server
	pnpm --filter docs dev

build-docs: ## Build documentation site
	pnpm --filter docs build

docker-build: ## Build Docker image
	docker build -t $(IMAGE) .

docker-up: ## Start Docker containers
	docker compose up -d

docker-down: ## Stop Docker containers
	docker compose down

migrate: ## Run database migrations
	go run . migrate

generate-key: ## Generate a new encryption key (64 hex characters)
	@openssl rand -hex 32

lint: ## Run linters for Go and React
	go vet ./internal/...
	pnpm --filter web lint

test: ## Run Go unit tests
	go test ./internal/...

fmt-md: ## Format markdown files
	dprint fmt
