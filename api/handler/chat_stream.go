package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/adwin2/youthvital/internal/agent"
	"github.com/adwin2/youthvital/internal/model"
)

// ChatStreamHandler handles SSE streaming chat responses.
type ChatStreamHandler struct {
	agent agent.ChatAgent
}

// NewChatStreamHandler creates a streaming chat HTTP handler.
func NewChatStreamHandler(a agent.ChatAgent) *ChatStreamHandler {
	return &ChatStreamHandler{agent: a}
}

// Register mounts streaming chat routes.
func (h *ChatStreamHandler) Register(group *route.RouterGroup) {
	group.POST("/chat/stream", h.Stream)
}

// Stream handles POST /v1/chat/stream with SSE.
func (h *ChatStreamHandler) Stream(ctx context.Context, c *app.RequestContext) {
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

	// Set SSE headers
	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Get response from agent
	var resp *model.ChatResponse
	var err error
	if sessionAgent, ok := h.agent.(agent.SessionChatAgent); ok {
		resp, err = sessionAgent.ChatWithOptions(ctx, req.Message, agent.ChatOptions{UserID: req.UserID, SessionID: req.SessionID})
	} else {
		resp, err = h.agent.Chat(ctx, req.Message)
	}
	if err != nil {
		h.sendSSEError(c, err.Error())
		return
	}

	// Stream the response in chunks
	h.streamResponse(c, resp)
}

type sseEvent struct {
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data"`
}

func (h *ChatStreamHandler) streamResponse(c *app.RequestContext, resp *model.ChatResponse) {
	// Send answer in chunks for streaming effect
	answer := resp.Answer
	chunkSize := 20 // Characters per chunk

	for i := 0; i < len(answer); i += chunkSize {
		end := i + chunkSize
		if end > len(answer) {
			end = len(answer)
		}

		chunk := answer[i:end]
		event := sseEvent{
			Event: "answer_chunk",
			Data:  chunk,
		}

		h.sendSSE(c, event)
		c.Flush()
		time.Sleep(20 * time.Millisecond) // Simulate streaming latency
	}

	// Send tool calls info
	if len(resp.ToolCalls) > 0 {
		h.sendSSE(c, sseEvent{
			Event: "tool_calls",
			Data:  resp.ToolCalls,
		})
		c.Flush()
	}

	// Send metadata
	h.sendSSE(c, sseEvent{
		Event: "metadata",
		Data: map[string]interface{}{
			"hitl_triggered":    resp.HITLTriggered,
			"safety_blocked":    resp.SafetyBlocked,
			"screening_blocked": resp.ScreeningBlocked,
			"agents_called":     resp.AgentsCalled,
		},
	})
	c.Flush()

	// Send done event
	h.sendSSE(c, sseEvent{
		Event: "done",
		Data:  nil,
	})
	c.Flush()
}

func (h *ChatStreamHandler) sendSSE(c *app.RequestContext, event sseEvent) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		slog.Error("failed to marshal SSE event", "error", err)
		return
	}

	// SSE format: data: <json>\n\n
	c.Response.BodyWriter().Write([]byte("data: "))
	c.Response.BodyWriter().Write(data)
	c.Response.BodyWriter().Write([]byte("\n\n"))
}

func (h *ChatStreamHandler) sendSSEError(c *app.RequestContext, message string) {
	h.sendSSE(c, sseEvent{
		Event: "error",
		Data: model.APIError{
			Code:    "chat_error",
			Message: message,
		},
	})
	c.Flush()
}
