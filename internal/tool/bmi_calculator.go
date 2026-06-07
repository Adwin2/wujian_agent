package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const bmiToolName = "bmi_calculator"

// BMICalculatorInput contains the required anthropometric values for BMI.
type BMICalculatorInput struct {
	Age      float64 `json:"age"`
	Sex      string  `json:"sex"`
	HeightCM float64 `json:"height_cm"`
	WeightKG float64 `json:"weight_kg"`
}

// BMICalculatorOutput contains the computed BMI and display metadata.
type BMICalculatorOutput struct {
	BMI     float64 `json:"bmi"`
	Rounded float64 `json:"rounded"`
	Formula string  `json:"formula"`
	Unit    string  `json:"unit"`
}

// BMICalculator is a deterministic Eino-compatible BMI tool.
type BMICalculator struct{}

var _ einotool.InvokableTool = (*BMICalculator)(nil)

// NewBMICalculator creates a BMI calculator tool.
func NewBMICalculator() *BMICalculator {
	return &BMICalculator{}
}

// Calculate validates input and computes BMI = weight_kg / height_m².
func (t *BMICalculator) Calculate(_ context.Context, input BMICalculatorInput) (*BMICalculatorOutput, error) {
	sex, err := normalizeSex(input.Sex)
	if err != nil {
		return nil, err
	}
	if err := validateAge(input.Age); err != nil {
		return nil, err
	}
	if err := validateHeight(input.HeightCM); err != nil {
		return nil, err
	}
	if err := validateWeight(input.WeightKG); err != nil {
		return nil, err
	}

	input.Sex = sex
	heightM := input.HeightCM / 100
	bmi := input.WeightKG / (heightM * heightM)

	return &BMICalculatorOutput{
		BMI:     roundTo(bmi, 2),
		Rounded: roundTo(bmi, 1),
		Formula: "BMI = weight_kg / (height_cm / 100)^2",
		Unit:    "kg/m^2",
	}, nil
}

// Info returns the Eino tool definition for model tool calling.
func (t *BMICalculator) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: bmiToolName,
		Desc: "Calculate BMI for youth ages 2-20 from age, sex, height_cm and weight_kg. Use this for any BMI or body mass index calculation; do not calculate BMI mentally.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"age": {
				Type:     schema.Number,
				Desc:     "Age in years. Must be between 2 and 20.",
				Required: true,
			},
			"sex": {
				Type:     schema.String,
				Desc:     "Biological sex for age/sex-adjusted interpretation.",
				Enum:     []string{"male", "female"},
				Required: true,
			},
			"height_cm": {
				Type:     schema.Number,
				Desc:     "Height in centimeters. Must be greater than 0 and no more than 250.",
				Required: true,
			},
			"weight_kg": {
				Type:     schema.Number,
				Desc:     "Weight in kilograms. Must be greater than 0 and no more than 300.",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun executes the BMI tool from JSON arguments.
func (t *BMICalculator) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input BMICalculatorInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", bmiToolName, err)
	}

	output, err := t.Calculate(ctx, input)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", bmiToolName, err)
	}
	return string(data), nil
}

func normalizeSex(sex string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(sex)) {
	case "male", "m", "男", "男孩", "儿子":
		return "male", nil
	case "female", "f", "女", "女孩", "女儿":
		return "female", nil
	default:
		return "", fmt.Errorf("sex must be male or female, got %q", sex)
	}
}

func validateAge(age float64) error {
	if age < 2 || age > 20 {
		return fmt.Errorf("age must be between 2 and 20 years, got %.2f", age)
	}
	return nil
}

func validateHeight(heightCM float64) error {
	if heightCM <= 0 || heightCM > 250 {
		return fmt.Errorf("height_cm must be greater than 0 and no more than 250, got %.2f", heightCM)
	}
	return nil
}

func validateWeight(weightKG float64) error {
	if weightKG <= 0 || weightKG > 300 {
		return fmt.Errorf("weight_kg must be greater than 0 and no more than 300, got %.2f", weightKG)
	}
	return nil
}

func roundTo(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}
