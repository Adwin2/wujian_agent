package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const alertSenderToolName = "alert_sender"

// AlertSenderInput describes an alert request. Phase 2 only supports dry-run.
type AlertSenderInput struct {
	Recipient string `json:"recipient"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

// AlertSenderOutput reports dry-run status.
type AlertSenderOutput struct {
	Sent       bool   `json:"sent"`
	DryRun     bool   `json:"dry_run"`
	Severity   string `json:"severity"`
	Status     string `json:"status"`
	Disclaimer string `json:"disclaimer"`
}

// AlertSender is a dry-run alert tool for Phase 2.
type AlertSender struct{}

var _ einotool.InvokableTool = (*AlertSender)(nil)

func NewAlertSender() *AlertSender { return &AlertSender{} }

func (t *AlertSender) Send(_ context.Context, input AlertSenderInput) (*AlertSenderOutput, error) {
	severity, err := validateEnum("severity", input.Severity, []string{"low", "medium", "high", "critical"})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Recipient) == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	if strings.TrimSpace(input.Message) == "" {
		return nil, fmt.Errorf("message is required")
	}
	return &AlertSenderOutput{
		Sent:       false,
		DryRun:     true,
		Severity:   severity,
		Status:     "Phase 2 dry-run only; no external alert was sent.",
		Disclaimer: "紧急情况请直接联系当地急救或专业人员；本系统当前不发送真实通知。",
	}, nil
}

func (t *AlertSender) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: alertSenderToolName,
		Desc: "Prepare an emergency/guardian alert in dry-run mode. Does not send external messages in Phase 2.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"recipient": {Type: schema.String, Required: true},
			"severity":  {Type: schema.String, Enum: []string{"low", "medium", "high", "critical"}, Required: true},
			"message":   {Type: schema.String, Required: true},
			"dry_run":   {Type: schema.Boolean, Desc: "Always treated as true in Phase 2."},
		}),
	}, nil
}

func (t *AlertSender) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input AlertSenderInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", alertSenderToolName, err)
	}
	output, err := t.Send(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", alertSenderToolName, err)
	}
	return string(data), nil
}
