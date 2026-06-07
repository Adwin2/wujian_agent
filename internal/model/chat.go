package model

// ChatRequest is the JSON payload accepted by POST /v1/chat.
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

// ChatResponse is returned by the YouthVital chat agent.
type ChatResponse struct {
	Answer           string     `json:"answer"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	AgentsCalled     []string   `json:"agents_called,omitempty"`
	HITLTriggered    bool       `json:"hitl_triggered,omitempty"`
	SafetyBlocked    bool       `json:"safety_blocked,omitempty"`
	ScreeningBlocked bool       `json:"screening_blocked,omitempty"`
}

// AssessmentRecord captures the completed chat turn for persistence and audit.
type AssessmentRecord struct {
	UserID        string
	SessionID     string
	InputText     string
	OutputText    string
	AgentsCalled  []string
	ToolCalls     []ToolCall
	RiskFlags     []any
	HITLTriggered bool
}

// AuditLogRecord captures PHI-related tool access for compliance review.
type AuditLogRecord struct {
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	ToolName     string
	ToolInput    any
	ToolOutput   any
}

// ErrorResponse keeps API error responses stable for clients and tests.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// APIError describes an HTTP API error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
