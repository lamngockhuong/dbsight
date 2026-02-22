.PHONY: build dev docker-build migrate fmt-md

build:
	cd web && pnpm run build
	go build -o bin/dbsight .

dev:
	docker-compose up -d postgres
	go run . serve &
	cd web && pnpm run dev

docker-build:
	docker build -t dbsight:latest .

migrate:
	go run . migrate

fmt-md:
	dprint fmt
