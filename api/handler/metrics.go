package handler

import (
	"bytes"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// MetricsHandler exposes Prometheus metrics.
type MetricsHandler struct {
	gatherer prometheus.Gatherer
}

// NewMetricsHandler creates a Prometheus metrics handler.
func NewMetricsHandler(gatherer prometheus.Gatherer) *MetricsHandler {
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}
	return &MetricsHandler{gatherer: gatherer}
}

// Register mounts the metrics endpoint.
func (h *MetricsHandler) Register(engine route.IRouter) {
	engine.GET("/metrics", h.Metrics)
}

// Metrics writes Prometheus text exposition format.
func (h *MetricsHandler) Metrics(_ context.Context, c *app.RequestContext) {
	families, err := h.gatherer.Gather()
	if err != nil {
		writeError(c, consts.StatusInternalServerError, "metrics_error", err.Error())
		return
	}

	var body bytes.Buffer
	for _, family := range families {
		if _, err := expfmt.MetricFamilyToText(&body, family); err != nil {
			writeError(c, consts.StatusInternalServerError, "metrics_error", err.Error())
			return
		}
	}

	c.Response.Header.Set("Content-Type", string(expfmt.NewFormat(expfmt.TypeTextPlain)))
	c.String(consts.StatusOK, body.String())
}
