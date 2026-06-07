package model

// ToolCall records a deterministic tool invocation surfaced to API callers.
type ToolCall struct {
	Name   string `json:"name"`
	Input  any    `json:"input,omitempty"`
	Output any    `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}
