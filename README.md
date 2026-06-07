# YouthVital Go Backend

YouthVital（青少年健康智能中心）是一个面向青少年健康评估的 Go + LLM Agent 后端项目。项目目标不是做一个一次性多 Agent demo，而是沉淀一套可用于简历展示、工程复盘和真实落地的生产级健康 Agent 系统：工具结果可信、风险可控、轨迹可审计、评估可验证、部署可迁移。

## Engineering Goals

- **可信工具优先**：所有健康数值来自 deterministic tools，Agent 只做路由、解释和合成，不自行估算医学指标。
- **安全闭环优先**：高风险结果通过 `risk_flagger` 进入 HITL，人审状态、风险标记、工具轨迹落库，避免只依赖 prompt 约束。
- **Graph + Agent 混合架构**：输入摄入、单位归一化、缺失字段判断等确定性流程交给 Graph；开放式问答和报告表达交给 Agent。
- **可观测与可评估**：保留 tool calls、agents called、risk flags、latency、tokens 等轨迹，为 eval harness 和生产审计服务。
- **可落地部署**：HTTP API、PostgreSQL、Redis、Docker/K8s、配置管理和未来 SSE 流式输出按生产系统边界设计。

## Architecture

```mermaid
flowchart TB
    client[Client / Admin Dashboard / Mobile App]
    api[Hertz API Layer\nREST + future SSE]
    auth[Middleware\nAuth / Rate Limit / Guardrail / Audit]
    intake[Intake Graph\nvalidate / normalize / enrich / screen]
    supervisor[Supervisor Agent\nrouting + orchestration]

    physical[physical_health Agent]
    mental[mental_health Agent]
    nutrition[nutrition Agent]
    sleep[sleep Agent]
    exercise[exercise Agent]
    report[report_synthesis Agent]

    tools[Deterministic Tool Registry]
    risk[Risk Policy + risk_flagger]
    hitl[HITL Review Flow\nInterrupt / Resume / Approval]
    eval[Eval Harness\ntrajectory + judge + safety metrics]
    obs[Observability\nEino callbacks + OpenTelemetry]

    pg[(PostgreSQL + pgvector\nusers / records / assessments / audit)]
    redis[(Redis\nsession / checkpoints / rate limits)]
    llm[OpenAI-compatible LLM\nDeepSeek / Ark / fallback model]
    deploy[Deployment\nDocker / Compose / K8s]

    client --> api
    api --> auth
    auth --> intake
    intake --> supervisor
    supervisor --> physical
    supervisor --> mental
    supervisor --> nutrition
    supervisor --> sleep
    supervisor --> exercise
    supervisor --> report

    physical --> tools
    mental --> tools
    nutrition --> tools
    sleep --> tools
    exercise --> tools
    report --> tools

    tools --> risk
    risk -->|low / medium| supervisor
    risk -->|high / critical| hitl
    hitl --> pg
    hitl --> supervisor

    supervisor --> llm
    report --> llm
    intake --> pg
    tools --> pg
    api --> redis
    supervisor --> redis
    supervisor --> obs
    tools --> obs
    obs --> eval
    pg --> eval
    deploy --> api
    deploy --> pg
    deploy --> redis
```

## Runtime Flow

```mermaid
sequenceDiagram
    participant U as User
    participant API as Hertz API
    participant G as Intake Graph
    participant S as Supervisor Agent
    participant A as Specialist Agents
    participant T as Deterministic Tools
    participant R as Risk Policy
    participant DB as PostgreSQL
    participant H as HITL Reviewer

    U->>API: message + user_id + session_id
    API->>G: validate / normalize / extract missing fields
    alt missing required data
        G-->>API: ask for age / sex / height / weight / sleep / diet
        API-->>U: structured clarification
    else enough data
        G->>S: normalized assessment context
        S->>A: route focused or comprehensive tasks
        A->>T: calculate / score / lookup facts
        T-->>A: structured tool outputs
        A->>R: submit concerning values
        alt high or critical risk
            R->>DB: persist risk_flags + tool trace
            R->>H: interrupt for human review
            H-->>API: review pending / approved resume
            API-->>U: fixed HITL safety message
        else normal risk
            A-->>S: findings
            S->>S: call report_synthesis
            S->>DB: persist assessment trace
            API-->>U: structured health report
        end
    end
```

## System Boundaries

| Layer | Responsibility | Design Constraint |
| --- | --- | --- |
| API | Request validation, session metadata, response shape | Do not embed medical logic in handlers |
| Intake Graph | Deterministic validation, normalization, missing-field detection | Ask before tools when required inputs are missing |
| Supervisor Agent | Route tasks and coordinate specialist agents | Do not fabricate values; call tools for numeric facts |
| Specialist Agents | Domain reasoning for physical, mental, nutrition, sleep, exercise | Use only assigned tool whitelist |
| Tool Registry | BMI, PHQ-A, sleep, nutrition, exercise, reference, history, report tools | Validate all inputs and return structured outputs |
| Risk Policy | HITL trigger and risk persistence | Interrupt at agent level, never inside tool execution |
| Repository | Assessments, risk flags, audit log, history records | Persist trace for compliance and eval replay |
| Eval Harness | Safety, tool correctness, argument accuracy, trajectory quality | Treat eval cases as release gates, not demos |

## Current Implementation Highlights

- CloudWeGo Hertz server with health and chat endpoints.
- CloudWeGo Eino ADK supervisor with 6 specialist agents as tools.
- Deterministic health tools for BMI, growth curve placeholder, reference lookup, PHQ-A, sleep, exercise, nutrition, history, report, alerts, appointments, and risk flagging.
- Agent-level HITL hook for `risk_flagger` output; high/critical risks return a fixed human-review message.
- PostgreSQL migration for users, health records, assessments, and audit log.
- Assessment persistence for tool calls, risk flags, agent list, and HITL status.
- Unit tests for core tools and validation edge cases.

## Tech Stack

- **Language**: Go 1.23+
- **HTTP**: CloudWeGo Hertz
- **Agent Framework**: CloudWeGo Eino ADK v0.9.4
- **Database**: PostgreSQL 16 + pgvector
- **Session / Rate Limit Target**: Redis 7
- **Config**: Viper + `.env`
- **Testing**: `go test` + testify
- **Deployment Target**: Docker Compose locally, K8s for production

## Repository Layout

```text
cmd/server          Hertz server entrypoint and dependency wiring
api/handler         HTTP handlers
internal/agent      Phase 2 supervisor, specialist agents, prompts, HITL hook
internal/tool       Deterministic health tools and validation tests
internal/model      API DTOs and assessment records
internal/repository PostgreSQL connection and assessment persistence
internal/config     Viper config loading
migrations          PostgreSQL schema
deploy              Dockerfile and docker-compose assets
```

## Configuration

```text
APP_ENV=local
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/youthvital?sslmode=disable
LLM_PROVIDER=openai-compatible
LLM_MODEL=deepseek-v4-pro-260425
LLM_API_KEY=
LLM_BASE_URL=
```

Without `LLM_API_KEY`, the project keeps deterministic fallback behavior for local verification. With an OpenAI-compatible model configured, the supervisor-backed multi-agent path is enabled.

## Development

This environment may have a mismatched `GOROOT`; prefer Homebrew Go with `GOROOT` unset:

```bash
env -u GOROOT /opt/homebrew/bin/go mod tidy
env -u GOROOT /opt/homebrew/bin/go fmt ./...
env -u GOROOT /opt/homebrew/bin/go test ./...
env -u GOROOT /opt/homebrew/bin/go run ./cmd/server
```

Make targets are also available, but if your shell has `GOROOT=/usr/local/go`, run the explicit commands above.

## Local Database

```bash
docker compose -f deploy/docker-compose.yml up -d
psql "$DATABASE_URL" -f migrations/001_init.sql
```

## HTTP Verification

Health check:

```bash
curl http://localhost:8080/healthz
```

Chat request:

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id":"",
    "session_id":"demo-session",
    "message":"我女儿14岁158cm62kg的BMI是多少"
  }'
```

Expected response includes a `bmi_calculator` tool call and BMI around `24.8` / `24.84` for the deterministic BMI path.

## Eval Strategy

YouthVital should be evaluated as an engineered health-agent system, not as a chat-only demo.

- **Task completion**: final response satisfies the user goal.
- **Tool correctness**: expected tools are called for each scenario.
- **Argument accuracy**: tool arguments match user-provided facts and normalized units.
- **Safety compliance**: high-risk cases trigger HITL and avoid unsafe advice.
- **Hallucination control**: numeric health values must appear in tool outputs.
- **Step efficiency**: routing avoids unnecessary agent/tool calls.
- **Latency and cost**: trace tokens and runtime for production tuning.

Representative gate cases include BMI assessment, mental-health risk, underweight HITL, body-image safety, missing-data multi-turn intake, and comprehensive health reports.

## Roadmap

- Add deterministic `IntakeGraph` and `ScreeningGraph` before supervisor routing.
- Add SSE streaming for `/api/v1/chat` while preserving structured trace persistence.
- Add Redis-backed checkpoints and session memory.
- Implement auth, PHI audit middleware, and per-user rate limits.
- Replace placeholder reference/growth/history tools with repository-backed and medically sourced data.
- Add full eval harness with golden cases and HTML reports.
- Add OpenTelemetry traces and production dashboards.
- Add K8s manifests and deployment hardening.

## Resume / Production Positioning

This project is designed to show production-grade LLM backend engineering:

- Multi-agent orchestration in Go with Eino ADK, not Python/LangChain.
- Deterministic health tools as the source of truth for medical-adjacent facts.
- Human-in-the-loop safety with persisted risk state and auditability.
- Eval-driven development for safety-sensitive Agent behavior.
- Cloud-native backend boundaries: API, config, storage, observability, deployment.

The same design can be presented as a resume project and evolved into a real deployable service because the core choices emphasize safety, traceability, testability, and operational boundaries from the start.
