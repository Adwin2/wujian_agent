package tracing

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/adwin2/youthvital")

// ContextKey is used to store trace context values.
type ContextKey string

const (
	SessionIDKey  ContextKey = "session_id"
	UserIDKey     ContextKey = "user_id"
	TraceIDKey    ContextKey = "trace_id"
	SpanIDKey     ContextKey = "span_id"
)

// StartSpan creates a new span from context and returns a child context with the span.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// StartSpanFromRequest extracts session/user context and starts a span.
func StartSpanFromRequest(ctx context.Context, c *app.RequestContext, name string) (context.Context, trace.Span) {
	sessionID := string(c.FormValue("session_id"))
	userID := string(c.FormValue("user_id"))

	spanCtx, span := tracer.Start(ctx, name,
		trace.WithAttributes(
			attribute.String("session.id", sessionID),
			attribute.String("user.id", userID),
			attribute.String("http.method", string(c.Method())),
			attribute.String("http.path", string(c.Path())),
		),
	)

	if sessionID != "" {
		spanCtx = context.WithValue(spanCtx, SessionIDKey, sessionID)
	}
	if userID != "" {
		spanCtx = context.WithValue(spanCtx, UserIDKey, userID)
	}

	return spanCtx, span
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() && err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetAttributes sets attributes on the current span.
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// GetTraceID returns the current trace ID as a string.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
