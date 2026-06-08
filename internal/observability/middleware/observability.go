package middleware

import (
	"context"
	"time"

	"github.com/adwin2/youthvital/internal/observability/metrics"
	"github.com/adwin2/youthvital/internal/observability/tracing"
	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ObservabilityMiddleware adds tracing and metrics to requests.
func ObservabilityMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Path())
		method := string(c.Method())

		// Start span
		spanCtx, span := tracing.StartSpanFromRequest(ctx, c, "HTTP "+method+" "+path)
		defer span.End()

		// Store trace ID in context for logging
		traceID := tracing.GetTraceID(spanCtx)
		c.Set(string(tracing.TraceIDKey), traceID)

		// Process request
		c.Next(spanCtx)

		// Record metrics
		status := "success"
		if c.Response.StatusCode() >= 400 {
			status = "error"
			span.SetStatus(codes.Error, "HTTP error")
			span.SetAttributes(attribute.Int("http.status_code", c.Response.StatusCode()))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		duration := time.Since(start)
		metrics.RecordChat(spanCtx, duration, status, false, false)
	}
}
