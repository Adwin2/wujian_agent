package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "youthvital"

var (
	// ChatLatency tracks chat request latency in milliseconds.
	ChatLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "chat_latency_milliseconds",
			Help:      "Latency of chat requests in milliseconds",
			Buckets:   prometheus.ExponentialBuckets(10, 2, 10), // 10ms to ~5s
		},
		[]string{"status", "hitl_triggered", "safety_blocked"},
	)

	// ChatRequestsTotal counts total chat requests.
	ChatRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "chat_requests_total",
			Help:      "Total number of chat requests",
		},
		[]string{"status", "hitl_triggered", "safety_blocked"},
	)

	// ToolCallsTotal counts tool invocations.
	ToolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "tool_calls_total",
			Help:      "Total number of tool calls",
		},
		[]string{"tool_name", "status"},
	)

	// ToolLatency tracks tool call latency.
	ToolLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "tool_latency_milliseconds",
			Help:      "Latency of tool calls in milliseconds",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1ms to ~500ms
		},
		[]string{"tool_name"},
	)

	// LLMTokensUsed tracks LLM token usage.
	LLMTokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "llm_tokens_used_total",
			Help:      "Total LLM tokens used",
		},
		[]string{"model", "type"}, // type: input, output
	)

	// LLMLatency tracks LLM call latency.
	LLMLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "llm_latency_milliseconds",
			Help:      "Latency of LLM calls in milliseconds",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 10), // 100ms to ~50s
		},
		[]string{"model", "status"},
	)

	// HITLTriggersTotal counts HITL interventions.
	HITLTriggersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "hitl_triggers_total",
			Help:      "Total HITL triggers by severity",
		},
		[]string{"severity"},
	)

	// SafetyBlocksTotal counts safety guardrail blocks.
	SafetyBlocksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "safety_blocks_total",
			Help:      "Total safety blocks by reason",
		},
		[]string{"reason"},
	)

	// AssessmentRecordsPersisted tracks assessment persistence.
	AssessmentRecordsPersisted = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "assessment_records_persisted_total",
			Help:      "Total assessment records persisted",
		},
	)

	// AuditLogsPersisted tracks audit log persistence.
	AuditLogsPersisted = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "audit_logs_persisted_total",
			Help:      "Total audit logs persisted",
		},
	)
)

// RecordChat records a chat request metric.
func RecordChat(ctx context.Context, duration time.Duration, status string, hitlTriggered, safetyBlocked bool) {
	ms := float64(duration.Milliseconds())

	ChatLatency.WithLabelValues(status, boolStr(hitlTriggered), boolStr(safetyBlocked)).Observe(ms)
	ChatRequestsTotal.WithLabelValues(status, boolStr(hitlTriggered), boolStr(safetyBlocked)).Inc()
}

// RecordToolCall records a tool call metric.
func RecordToolCall(ctx context.Context, toolName string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	ms := float64(duration.Milliseconds())
	ToolCallsTotal.WithLabelValues(toolName, status).Inc()
	ToolLatency.WithLabelValues(toolName).Observe(ms)
}

// RecordLLMUsage records LLM token usage.
func RecordLLMUsage(ctx context.Context, model string, inputTokens, outputTokens int) {
	if inputTokens > 0 {
		LLMTokensUsed.WithLabelValues(model, "input").Add(float64(inputTokens))
	}
	if outputTokens > 0 {
		LLMTokensUsed.WithLabelValues(model, "output").Add(float64(outputTokens))
	}
}

// RecordLLMLatency records LLM call latency.
func RecordLLMLatency(ctx context.Context, model string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	ms := float64(duration.Milliseconds())
	LLMLatency.WithLabelValues(model, status).Observe(ms)
}

// RecordHITLTrigger records a HITL trigger.
func RecordHITLTrigger(ctx context.Context, severity string) {
	HITLTriggersTotal.WithLabelValues(severity).Inc()
}

// RecordSafetyBlock records a safety block.
func RecordSafetyBlock(ctx context.Context, reason string) {
	SafetyBlocksTotal.WithLabelValues(reason).Inc()
}

// RecordAssessmentPersisted records an assessment persistence.
func RecordAssessmentPersisted(ctx context.Context) {
	AssessmentRecordsPersisted.Inc()
}

// RecordAuditLogPersisted records an audit log persistence.
func RecordAuditLogPersisted(ctx context.Context) {
	AuditLogsPersisted.Inc()
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
