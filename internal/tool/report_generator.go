package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const reportGeneratorToolName = "report_generator"

// ReportGeneratorInput contains already-computed findings for synthesis.
type ReportGeneratorInput struct {
	Title       string   `json:"title,omitempty"`
	Findings    []string `json:"findings"`
	Limitations []string `json:"limitations,omitempty"`
	NextSteps   []string `json:"next_steps,omitempty"`
}

// ReportGeneratorOutput contains a structured report composed from inputs.
type ReportGeneratorOutput struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Findings    []string `json:"findings"`
	Limitations []string `json:"limitations,omitempty"`
	NextSteps   []string `json:"next_steps,omitempty"`
	Disclaimer  string   `json:"disclaimer"`
}

// ReportGenerator synthesizes tool-derived findings without new calculations.
type ReportGenerator struct{}

var _ einotool.InvokableTool = (*ReportGenerator)(nil)

func NewReportGenerator() *ReportGenerator { return &ReportGenerator{} }

func (t *ReportGenerator) Generate(_ context.Context, input ReportGeneratorInput) (*ReportGeneratorOutput, error) {
	findings := trimStrings(input.Findings)
	if len(findings) == 0 {
		return nil, fmt.Errorf("findings are required; report_generator only synthesizes existing tool outputs")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "YouthVital 结构化健康评估摘要"
	}
	limitations := trimStrings(input.Limitations)
	nextSteps := trimStrings(input.NextSteps)
	return &ReportGeneratorOutput{
		Title:       title,
		Summary:     strings.Join(findings, " "),
		Findings:    findings,
		Limitations: limitations,
		NextSteps:   nextSteps,
		Disclaimer:  "本报告仅汇总工具结果和已提供信息，不能替代医生诊断。",
	}, nil
}

func (t *ReportGenerator) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: reportGeneratorToolName,
		Desc: "Generate a structured health report only from already computed tool findings. Do not invent numeric values.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title":       {Type: schema.String, Desc: "Optional report title."},
			"findings":    {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}, Required: true},
			"limitations": {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}},
			"next_steps":  {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}},
		}),
	}, nil
}

func (t *ReportGenerator) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input ReportGeneratorInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", reportGeneratorToolName, err)
	}
	output, err := t.Generate(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", reportGeneratorToolName, err)
	}
	return string(data), nil
}

func trimStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
