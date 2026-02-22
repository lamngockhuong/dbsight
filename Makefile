.PHONY: build dev dev-docs build-docs docker-build docker-up docker-down migrate generate-key lint test fmt-md

BINARY=bin/dbsight
IMAGE=dbsight:latest

build:
	pnpm --filter web build
	go build -o $(BINARY) .

dev:
	docker-compose up -d postgres
	go run . serve &
	pnpm --filter web dev

dev-docs:
	pnpm --filter docs dev

build-docs:
	pnpm --filter docs build

docker-build:
	docker build -t $(IMAGE) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate:
	go run . migrate

generate-key:
	@openssl rand -hex 32

lint:
	go vet ./internal/...
	pnpm --filter web lint

test:
	go test ./internal/...

fmt-md:
	dprint fmt
