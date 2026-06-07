.PHONY: tidy fmt test run docker-up docker-down migrate

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/youthvital?sslmode=disable

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

run:
	go run ./cmd/server

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql
