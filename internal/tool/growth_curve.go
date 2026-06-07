package tool

import (
	"context"
	"encoding/json"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const growthCurveToolName = "growth_curve"

// GrowthCurveInput contains inputs for a youth growth reference lookup.
type GrowthCurveInput struct {
	Age      float64 `json:"age"`
	Sex      string  `json:"sex"`
	HeightCM float64 `json:"height_cm,omitempty"`
	WeightKG float64 `json:"weight_kg,omitempty"`
	BMI      float64 `json:"bmi,omitempty"`
}

// GrowthCurveOutput is intentionally conservative until authoritative tables are configured.
type GrowthCurveOutput struct {
	Available           bool     `json:"available"`
	PercentileAvailable bool     `json:"percentile_available"`
	Percentile          *float64 `json:"percentile,omitempty"`
	PercentileText      string   `json:"percentile_text"`
	Source              string   `json:"source"`
	Description         string   `json:"description"`
	Disclaimer          string   `json:"disclaimer"`
}

// GrowthCurve is an Eino-compatible placeholder for authoritative percentile data.
type GrowthCurve struct{}

var _ einotool.InvokableTool = (*GrowthCurve)(nil)

// NewGrowthCurve creates a growth curve lookup tool.
func NewGrowthCurve() *GrowthCurve {
	return &GrowthCurve{}
}

// Lookup validates inputs and returns a non-fabricated Phase 1 placeholder.
func (t *GrowthCurve) Lookup(_ context.Context, input GrowthCurveInput) (*GrowthCurveOutput, error) {
	if _, err := normalizeSex(input.Sex); err != nil {
		return nil, err
	}
	if err := validateAge(input.Age); err != nil {
		return nil, err
	}
	if input.HeightCM != 0 {
		if err := validateHeight(input.HeightCM); err != nil {
			return nil, err
		}
	}
	if input.WeightKG != 0 {
		if err := validateWeight(input.WeightKG); err != nil {
			return nil, err
		}
	}
	if input.BMI < 0 {
		return nil, fmt.Errorf("bmi must be greater than or equal to 0, got %.2f", input.BMI)
	}

	return &GrowthCurveOutput{
		Available:           false,
		PercentileAvailable: false,
		PercentileText:      "百分位参考表尚未配置，不能给出权威百分位。",
		Source:              "phase2_reference_placeholder",
		Description:         "Phase 2 validates growth-curve lookup inputs but does not include authoritative WHO/CDC or Chinese pediatric percentile tables yet, so it will not fabricate a percentile.",
		Disclaimer:          "青少年 BMI 和生长发育解读需要按年龄、性别和权威参考表综合判断；本结果不能替代医生诊断。",
	}, nil
}

// Info returns the Eino tool definition for growth reference lookup.
func (t *GrowthCurve) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: growthCurveToolName,
		Desc: "Validate youth growth reference lookup inputs. In Phase 1 this tool must not invent percentiles when authoritative growth tables are not configured.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"age": {
				Type:     schema.Number,
				Desc:     "Age in years. Must be between 2 and 20.",
				Required: true,
			},
			"sex": {
				Type:     schema.String,
				Desc:     "Biological sex for age/sex-adjusted reference lookup.",
				Enum:     []string{"male", "female"},
				Required: true,
			},
			"height_cm": {
				Type: schema.Number,
				Desc: "Optional height in centimeters.",
			},
			"weight_kg": {
				Type: schema.Number,
				Desc: "Optional weight in kilograms.",
			},
			"bmi": {
				Type: schema.Number,
				Desc: "Optional BMI value calculated by bmi_calculator.",
			},
		}),
	}, nil
}

// InvokableRun executes the growth curve tool from JSON arguments.
func (t *GrowthCurve) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input GrowthCurveInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", growthCurveToolName, err)
	}

	output, err := t.Lookup(ctx, input)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", growthCurveToolName, err)
	}
	return string(data), nil
}
