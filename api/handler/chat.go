package handler

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/adwin2/youthvital/internal/agent"
	"github.com/adwin2/youthvital/internal/model"
	"github.com/adwin2/youthvital/internal/observability/metrics"
)

// ChatHandler adapts the Phase 1 chat agent to Hertz.
type ChatHandler struct {
	agent agent.ChatAgent
}

// NewChatHandler creates a chat HTTP handler.
func NewChatHandler(agent agent.ChatAgent) *ChatHandler {
	return &ChatHandler{agent: agent}
}

// Register mounts chat routes on a versioned group.
func (h *ChatHandler) Register(group *route.RouterGroup) {
	group.POST("/chat", h.Chat)
}

// Chat handles POST /v1/chat.
func (h *ChatHandler) Chat(ctx context.Context, c *app.RequestContext) {
	start := time.Now()
	var req model.ChatRequest
	if err := c.BindJSON(&req); err != nil {
		writeError(c, consts.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeError(c, consts.StatusBadRequest, "invalid_request", "message is required")
		return
	}

	var resp *model.ChatResponse
	var err error
	if sessionAgent, ok := h.agent.(agent.SessionChatAgent); ok {
		resp, err = sessionAgent.ChatWithOptions(ctx, req.Message, agent.ChatOptions{UserID: req.UserID, SessionID: req.SessionID})
	} else {
		resp, err = h.agent.Chat(ctx, req.Message)
	}
	if err != nil {
		metrics.RecordChat(ctx, time.Since(start), "error", false, false)
		writeError(c, consts.StatusBadRequest, "chat_error", err.Error())
		return
	}

	metrics.RecordChat(ctx, time.Since(start), "success", resp.HITLTriggered, resp.SafetyBlocked)
	c.JSON(consts.StatusOK, resp)
}

func writeError(c *app.RequestContext, status int, code string, message string) {
	c.JSON(status, model.ErrorResponse{
		Error: model.APIError{
			Code:    code,
			Message: message,
		},
	})
}
