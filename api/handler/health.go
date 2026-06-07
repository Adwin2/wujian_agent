package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
)

// ReadinessChecker verifies downstream dependency readiness.
type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler exposes health and readiness probes.
type HealthHandler struct {
	checker ReadinessChecker
}

// NewHealthHandler creates a health handler.
func NewHealthHandler(checker ReadinessChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// Register mounts health routes.
func (h *HealthHandler) Register(engine route.IRouter) {
	engine.GET("/healthz", h.Healthz)
	engine.GET("/readyz", h.Readyz)
}

// Healthz reports process liveness.
func (h *HealthHandler) Healthz(_ context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}

// Readyz reports optional dependency readiness.
func (h *HealthHandler) Readyz(ctx context.Context, c *app.RequestContext) {
	if h.checker == nil {
		c.JSON(consts.StatusOK, map[string]string{"status": "ok", "database": "not_configured"})
		return
	}
	if err := h.checker.Ping(ctx); err != nil {
		c.JSON(consts.StatusServiceUnavailable, map[string]string{"status": "unavailable", "database": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}
