package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const riskFlaggerToolName = "risk_flagger"

// RiskFlaggerInput describes a health risk found by an agent or tool.
type RiskFlaggerInput struct {
	RiskType    string `json:"risk_type"`
	Severity    string `json:"severity"`
	MetricName  string `json:"metric_name"`
	Value       string `json:"value"`
	Threshold   string `json:"threshold"`
	Description string `json:"description"`
}

// RiskFlaggerOutput is data only. It never triggers interrupt inside the tool.
type RiskFlaggerOutput struct {
	FlagID             string `json:"flag_id"`
	RiskType           string `json:"risk_type"`
	Severity           string `json:"severity"`
	MetricName         string `json:"metric_name"`
	Value              string `json:"value"`
	Threshold          string `json:"threshold"`
	Description        string `json:"description"`
	RequireHumanReview bool   `json:"require_human_review"`
	RecommendedAction  string `json:"recommended_action"`
}

// RiskFlagger returns structured risk data for the agent-level HITL hook.
type RiskFlagger struct{}

var _ einotool.InvokableTool = (*RiskFlagger)(nil)

func NewRiskFlagger() *RiskFlagger { return &RiskFlagger{} }

func (t *RiskFlagger) Flag(_ context.Context, input RiskFlaggerInput) (*RiskFlaggerOutput, error) {
	riskType, err := validateEnum("risk_type", input.RiskType, []string{"physical", "mental", "nutrition", "sleep", "exercise"})
	if err != nil {
		return nil, err
	}
	severity, err := validateEnum("severity", input.Severity, []string{"low", "medium", "high", "critical"})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.MetricName) == "" {
		return nil, fmt.Errorf("metric_name is required")
	}
	if strings.TrimSpace(input.Value) == "" {
		return nil, fmt.Errorf("value is required")
	}
	if strings.TrimSpace(input.Threshold) == "" {
		return nil, fmt.Errorf("threshold is required")
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, fmt.Errorf("description is required")
	}

	requireReview := severity == "high" || severity == "critical"
	action := "记录风险并继续观察；必要时咨询专业人员。"
	if requireReview {
		action = "停止给出该风险领域的进一步建议，转交人工/专业人员复核。"
	}

	return &RiskFlaggerOutput{
		FlagID:             fmt.Sprintf("%s:%s:%s", riskType, severity, strings.ToLower(strings.ReplaceAll(input.MetricName, " ", "_"))),
		RiskType:           riskType,
		Severity:           severity,
		MetricName:         strings.TrimSpace(input.MetricName),
		Value:              strings.TrimSpace(input.Value),
		Threshold:          strings.TrimSpace(input.Threshold),
		Description:        strings.TrimSpace(input.Description),
		RequireHumanReview: requireReview,
		RecommendedAction:  action,
	}, nil
}

func (t *RiskFlagger) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: riskFlaggerToolName,
		Desc: "Return structured health risk data. This tool never interrupts; agent-level AfterToolCallsHook handles human review.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"risk_type":   {Type: schema.String, Enum: []string{"physical", "mental", "nutrition", "sleep", "exercise"}, Required: true},
			"severity":    {Type: schema.String, Enum: []string{"low", "medium", "high", "critical"}, Required: true},
			"metric_name": {Type: schema.String, Desc: "Metric that triggered the flag.", Required: true},
			"value":       {Type: schema.String, Desc: "Observed value.", Required: true},
			"threshold":   {Type: schema.String, Desc: "Risk threshold or reference.", Required: true},
			"description": {Type: schema.String, Desc: "Short risk explanation.", Required: true},
		}),
	}, nil
}

func (t *RiskFlagger) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input RiskFlaggerInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", riskFlaggerToolName, err)
	}
	output, err := t.Flag(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", riskFlaggerToolName, err)
	}
	return string(data), nil
}

func validateEnum(field string, value string, allowed []string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if slices.Contains(allowed, normalized) {
		return normalized, nil
	}
	return "", fmt.Errorf("%s must be one of %s, got %q", field, strings.Join(allowed, ", "), value)
}
