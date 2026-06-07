package tool

import (
	"context"
	"encoding/json"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const phqScorerToolName = "phq_scorer"

// PHQScorerInput contains PHQ-A item scores.
type PHQScorerInput struct {
	Items []int `json:"items"`
}

// PHQScorerOutput contains total score and category.
type PHQScorerOutput struct {
	Total          int    `json:"total"`
	Category       string `json:"category"`
	Interpretation string `json:"interpretation"`
	Source         string `json:"source"`
	Disclaimer     string `json:"disclaimer"`
}

// PHQScorer scores PHQ-A style item arrays.
type PHQScorer struct{}

var _ einotool.InvokableTool = (*PHQScorer)(nil)

func NewPHQScorer() *PHQScorer { return &PHQScorer{} }

func (t *PHQScorer) Score(_ context.Context, input PHQScorerInput) (*PHQScorerOutput, error) {
	if len(input.Items) != 9 {
		return nil, fmt.Errorf("items must contain exactly 9 PHQ-A scores, got %d", len(input.Items))
	}
	total := 0
	for i, score := range input.Items {
		if score < 0 || score > 3 {
			return nil, fmt.Errorf("items[%d] must be between 0 and 3, got %d", i, score)
		}
		total += score
	}
	category := "minimal"
	interpretation := "筛查分数较低。"
	switch {
	case total >= 20:
		category = "severe"
		interpretation = "筛查分数较高，需要专业人员进一步评估。"
	case total >= 15:
		category = "moderately_severe"
		interpretation = "筛查分数偏高，建议尽快寻求专业评估。"
	case total >= 10:
		category = "moderate"
		interpretation = "筛查分数提示需要关注，建议由专业人员进一步评估。"
	case total >= 5:
		category = "mild"
		interpretation = "筛查分数轻度升高，建议持续观察并关注情绪变化。"
	}
	return &PHQScorerOutput{
		Total:          total,
		Category:       category,
		Interpretation: interpretation,
		Source:         "phase2_phq_a_scoring_reference",
		Disclaimer:     "PHQ-A 是筛查工具，不能替代精神心理专业诊断；如有自伤风险应立即联系专业人员或紧急服务。",
	}, nil
}

func (t *PHQScorer) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: phqScorerToolName,
		Desc: "Score PHQ-A style screening items. Each item must be 0-3. Use only when item scores are supplied.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"items": {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.Integer}, Required: true, Desc: "PHQ-A item scores, each 0-3."},
		}),
	}, nil
}

func (t *PHQScorer) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input PHQScorerInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", phqScorerToolName, err)
	}
	output, err := t.Score(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", phqScorerToolName, err)
	}
	return string(data), nil
}
