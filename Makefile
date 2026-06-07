.PHONY: tidy fmt test test-real-judge run docker-up docker-down migrate

GO ?= env -u GOROOT /opt/homebrew/bin/go
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/youthvital?sslmode=disable

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-real-judge:
	PHASE4_REAL_JUDGE=1 $(GO) test ./eval -run TestArkJudgeIntegration -count=1 -timeout=90s -v

run:
	$(GO) run ./cmd/server

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql
