package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const nutritionLookupToolName = "nutrition_lookup"

// NutritionLookupInput identifies a food or nutrition topic.
type NutritionLookupInput struct {
	Query string `json:"query"`
}

// NutritionLookupOutput contains bounded nutrition reference information.
type NutritionLookupOutput struct {
	Query      string   `json:"query"`
	Category   string   `json:"category"`
	Nutrients  []string `json:"nutrients"`
	Guidance   string   `json:"guidance"`
	Source     string   `json:"source"`
	Disclaimer string   `json:"disclaimer"`
}

// NutritionLookup is a bounded Phase 2 nutrition reference tool.
type NutritionLookup struct{}

var _ einotool.InvokableTool = (*NutritionLookup)(nil)

func NewNutritionLookup() *NutritionLookup { return &NutritionLookup{} }

func (t *NutritionLookup) Lookup(_ context.Context, input NutritionLookupInput) (*NutritionLookupOutput, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	lower := strings.ToLower(query)
	out := &NutritionLookupOutput{
		Query:      query,
		Category:   "general",
		Nutrients:  []string{"balanced_diet"},
		Guidance:   "建议保持谷物、优质蛋白、蔬菜水果和奶类等多样化饮食。",
		Source:     "phase2_bounded_nutrition_reference",
		Disclaimer: "营养建议需结合年龄、过敏、疾病和当地膳食指南；不能替代营养师或医生建议。",
	}
	switch {
	case strings.Contains(query, "蔬菜") || strings.Contains(query, "水果") || strings.Contains(lower, "vegetable") || strings.Contains(lower, "fruit"):
		out.Category = "vegetable_fruit"
		out.Nutrients = []string{"fiber", "vitamins", "minerals"}
		out.Guidance = "蔬菜水果通常提供膳食纤维、维生素和矿物质，可作为均衡饮食的一部分。"
	case strings.Contains(query, "奶") || strings.Contains(lower, "milk") || strings.Contains(lower, "dairy"):
		out.Category = "dairy"
		out.Nutrients = []string{"calcium", "protein"}
		out.Guidance = "奶类常提供钙和蛋白质；乳糖不耐受或过敏者需选择合适替代方案。"
	case strings.Contains(query, "蛋") || strings.Contains(query, "鱼") || strings.Contains(query, "肉") || strings.Contains(lower, "protein"):
		out.Category = "protein"
		out.Nutrients = []string{"protein", "iron"}
		out.Guidance = "优质蛋白有助于生长发育，但具体摄入量需结合年龄、活动量和健康状况。"
	case strings.Contains(query, "零食") || strings.Contains(query, "泡面") || strings.Contains(lower, "snack"):
		out.Category = "processed_food"
		out.Nutrients = []string{"sodium", "added_sugar_or_fat"}
		out.Guidance = "高盐、高糖或高脂加工食品建议控制频次，优先选择正餐和营养密度更高的食物。"
	}
	return out, nil
}

func (t *NutritionLookup) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: nutritionLookupToolName,
		Desc: "Look up bounded youth nutrition reference information for a food or nutrition topic. Do not provide weight-loss prescriptions.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {Type: schema.String, Desc: "Food or nutrition topic.", Required: true},
		}),
	}, nil
}

func (t *NutritionLookup) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input NutritionLookupInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", nutritionLookupToolName, err)
	}
	output, err := t.Lookup(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", nutritionLookupToolName, err)
	}
	return string(data), nil
}
