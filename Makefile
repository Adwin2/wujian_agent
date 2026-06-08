.PHONY: tidy fmt test test-real-judge run docker-build docker-up docker-down migrate k8s-apply

GO ?= env -u GOROOT /opt/homebrew/bin/go
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/youthvital?sslmode=disable
IMAGE_NAME ?= youthvital
IMAGE_TAG ?= latest

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

docker-build:
	docker build -f deploy/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-build-full:
	docker compose -f deploy/docker-compose.yml build

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init.sql

k8s-apply:
	kubectl apply -f deploy/k8s/
