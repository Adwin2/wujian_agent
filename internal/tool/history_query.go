package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const historyQueryToolName = "history_query"

// HistoryQueryInput requests historical health data.
type HistoryQueryInput struct {
	UserID string `json:"user_id,omitempty"`
	Topic  string `json:"topic,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// HistoryQueryOutput reports availability of history data.
type HistoryQueryOutput struct {
	Available  bool     `json:"available"`
	UserID     string   `json:"user_id,omitempty"`
	Topic      string   `json:"topic,omitempty"`
	Records    []string `json:"records"`
	Message    string   `json:"message"`
	Source     string   `json:"source"`
	Disclaimer string   `json:"disclaimer"`
}

// HistoryQuery is a bounded placeholder until repository-backed history is wired.
type HistoryQuery struct{}

var _ einotool.InvokableTool = (*HistoryQuery)(nil)

func NewHistoryQuery() *HistoryQuery { return &HistoryQuery{} }

func (t *HistoryQuery) Query(_ context.Context, input HistoryQueryInput) (*HistoryQueryOutput, error) {
	if input.Limit < 0 {
		return nil, fmt.Errorf("limit must be greater than or equal to 0, got %d", input.Limit)
	}
	return &HistoryQueryOutput{
		Available:  false,
		UserID:     strings.TrimSpace(input.UserID),
		Topic:      strings.TrimSpace(input.Topic),
		Records:    []string{},
		Message:    "Phase 2 尚未接入按用户查询历史健康记录的仓储实现。",
		Source:     "phase2_history_query_placeholder",
		Disclaimer: "如果需要趋势判断，应使用真实、连续的历史体检/健康记录。",
	}, nil
}

func (t *HistoryQuery) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: historyQueryToolName,
		Desc: "Query historical youth health records when repository-backed history is available. In Phase 2 this returns explicit unavailable status.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {Type: schema.String, Desc: "Optional user ID."},
			"topic":   {Type: schema.String, Desc: "Optional health topic."},
			"limit":   {Type: schema.Integer, Desc: "Optional max records, non-negative."},
		}),
	}, nil
}

func (t *HistoryQuery) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input HistoryQueryInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", historyQueryToolName, err)
	}
	output, err := t.Query(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", historyQueryToolName, err)
	}
	return string(data), nil
}
