# YouthVital Go Backend

Phase 1 backend for YouthVital（青少年健康智能中心）.

## Stack

- Go 1.23+
- CloudWeGo Hertz
- CloudWeGo Eino ADK v0.9.4
- PostgreSQL 16 + pgvector
- Redis 7 for later session/rate-limit work
- Viper + `.env`
- `go test` + testify

## Layout

```text
cmd/server          Hertz server entrypoint
api/handler         HTTP handlers
internal/agent      Phase 1 chat agent and prompt
internal/config     Viper config loading
internal/model      API DTOs
internal/repository PostgreSQL connection helper
internal/tool       Deterministic health tools
migrations          PostgreSQL migrations
deploy              Local Docker assets
```

## Configuration

Environment variables:

```text
APP_ENV=local
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/youthvital?sslmode=disable
LLM_PROVIDER=openai-compatible
LLM_MODEL=gpt-4o
LLM_API_KEY=
LLM_BASE_URL=
```

The Phase 1 BMI verification path is deterministic and does not require LLM credentials.

## Development

```bash
go mod tidy
go test ./...
go run ./cmd/server
```

Or use Make targets:

```bash
make fmt
make test
make run
```

## Local database

```bash
docker compose -f deploy/docker-compose.yml up -d
psql "$DATABASE_URL" -f migrations/001_init.sql
```

## HTTP verification

Health check:

```bash
curl http://localhost:8080/healthz
```

Expected:

```json
{"status":"ok"}
```

BMI check:

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"我女儿14岁158cm62kg的BMI是多少"}'
```

Expected response includes a `bmi_calculator` tool call and BMI around `24.8` / `24.84`.

## Phase 1 scope

Implemented:

- Project module `github.com/adwin2/youthvital`
- Hertz server
- Viper configuration
- PostgreSQL schema
- `bmi_calculator`, `growth_curve`, `reference_lookup`
- Single Phase 1 chat agent boundary with deterministic BMI tool path

Deferred:

- Multi-agent supervisor
- Graph pipelines
- SSE streaming
- Auth/rate limiting
- Redis sessions
- HITL interrupts
- Full eval harness
- Admin dashboard integration
