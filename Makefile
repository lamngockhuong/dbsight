.PHONY: build dev docker-build docker-up docker-down migrate generate-key lint test fmt-md

BINARY=bin/dbsight
IMAGE=dbsight:latest

build:
	cd web && pnpm run build
	go build -o $(BINARY) .

dev:
	docker-compose up -d postgres
	go run . serve &
	cd web && pnpm run dev

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
	cd web && pnpm run lint

test:
	go test ./internal/...

fmt-md:
	dprint fmt
