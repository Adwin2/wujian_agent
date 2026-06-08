package handler

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/adwin2/youthvital/internal/agent"
	appmodel "github.com/adwin2/youthvital/internal/model"
	"github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatHandlerRunsPhase3ScreeningBeforeResponse(t *testing.T) {
	registry, err := tool.NewRegistry().WithGraphTools(t.Context())
	require.NoError(t, err)
	chatAgent := agent.NewPhase2ChatAgent(registry)

	h := server.Default()
	v1 := h.Group("/v1")
	NewChatHandler(chatAgent).Register(v1)

	body := bytes.NewBufferString(`{"message":"孩子最近情绪很低落，经常哭，不想上学，还说不想活了","session_id":"s1"}`)
	w := ut.PerformRequest(h.Engine, consts.MethodPost, "/v1/chat", &ut.Body{Body: body, Len: body.Len()}, ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()

	require.Equal(t, consts.StatusOK, resp.StatusCode())
	var payload appmodel.ChatResponse
	require.NoError(t, json.Unmarshal(resp.Body(), &payload))
	assert.True(t, payload.HITLTriggered)
	assert.True(t, payload.ScreeningBlocked)
	require.Len(t, payload.ToolCalls, 2)
	assert.Equal(t, "intake_pipeline", payload.ToolCalls[0].Name)
	assert.Equal(t, "screening_pipeline", payload.ToolCalls[1].Name)
}

func TestChatStreamHandlerEmitsSSEEvents(t *testing.T) {
	chatAgent := agent.NewPhase2ChatAgent(tool.NewRegistry())
	h := server.Default()
	v1 := h.Group("/v1")
	NewChatStreamHandler(chatAgent).Register(v1)

	body := bytes.NewBufferString(`{"message":"帮我算 BMI"}`)
	w := ut.PerformRequest(h.Engine, consts.MethodPost, "/v1/chat/stream", &ut.Body{Body: body, Len: body.Len()}, ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()

	require.Equal(t, consts.StatusOK, resp.StatusCode())
	assert.Contains(t, string(resp.Header.ContentType()), "text/event-stream")
	assert.Contains(t, string(resp.Body()), "data:")
	assert.Contains(t, string(resp.Body()), "年龄")
}

func TestMetricsHandlerExposesPrometheusText(t *testing.T) {
	h := server.Default()
	NewMetricsHandler(nil).Register(h)

	w := ut.PerformRequest(h.Engine, consts.MethodGet, "/metrics", nil)
	resp := w.Result()

	require.Equal(t, consts.StatusOK, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), "# HELP")
}
