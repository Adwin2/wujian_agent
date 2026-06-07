package tool

import (
	"context"
	"encoding/json"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const exerciseCalculatorToolName = "exercise_calculator"

// ExerciseCalculatorInput contains activity intensity and duration.
type ExerciseCalculatorInput struct {
	Activity        string   `json:"activity,omitempty"`
	MET             *float64 `json:"met"`
	DurationMinutes *float64 `json:"duration_minutes"`
	DaysPerWeek     float64  `json:"days_per_week,omitempty"`
}

// ExerciseCalculatorOutput contains MET-minute estimates.
type ExerciseCalculatorOutput struct {
	Activity         string  `json:"activity,omitempty"`
	MET              float64 `json:"met"`
	DurationMinutes  float64 `json:"duration_minutes"`
	DaysPerWeek      float64 `json:"days_per_week"`
	METMinutes       float64 `json:"met_minutes"`
	WeeklyMETMinutes float64 `json:"weekly_met_minutes"`
	Interpretation   string  `json:"interpretation"`
	Source           string  `json:"source"`
	Disclaimer       string  `json:"disclaimer"`
}

// ExerciseCalculator estimates MET-minutes from activity duration and MET.
type ExerciseCalculator struct{}

var _ einotool.InvokableTool = (*ExerciseCalculator)(nil)

func NewExerciseCalculator() *ExerciseCalculator { return &ExerciseCalculator{} }

func (t *ExerciseCalculator) Calculate(_ context.Context, input ExerciseCalculatorInput) (*ExerciseCalculatorOutput, error) {
	if input.MET == nil {
		return nil, fmt.Errorf("met is required")
	}
	if *input.MET < 0 {
		return nil, fmt.Errorf("met must be greater than or equal to 0, got %.2f", *input.MET)
	}
	if input.DurationMinutes == nil {
		return nil, fmt.Errorf("duration_minutes is required")
	}
	if *input.DurationMinutes < 0 {
		return nil, fmt.Errorf("duration_minutes must be greater than or equal to 0, got %.2f", *input.DurationMinutes)
	}
	if input.DaysPerWeek < 0 || input.DaysPerWeek > 7 {
		return nil, fmt.Errorf("days_per_week must be between 0 and 7, got %.2f", input.DaysPerWeek)
	}
	met := *input.MET
	durationMinutes := *input.DurationMinutes
	days := input.DaysPerWeek
	if days == 0 {
		days = 1
	}
	metMinutes := met * durationMinutes
	weekly := metMinutes * days
	interpretation := "已计算运动量；青少年运动建议还需结合年龄、健康状况和运动类型综合评估。"
	if weekly == 0 {
		interpretation = "未提供有效运动时长或 MET 值，无法形成运动量估计。"
	}
	return &ExerciseCalculatorOutput{
		Activity:         input.Activity,
		MET:              met,
		DurationMinutes:  durationMinutes,
		DaysPerWeek:      days,
		METMinutes:       roundTo(metMinutes, 2),
		WeeklyMETMinutes: roundTo(weekly, 2),
		Interpretation:   interpretation,
		Source:           "phase2_met_calculation",
		Disclaimer:       "运动建议不能替代医生或专业体能评估；如有疾病或不适，应先咨询专业人员。",
	}, nil
}

func (t *ExerciseCalculator) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: exerciseCalculatorToolName,
		Desc: "Calculate MET-minutes from MET value, duration, and optional days per week. Reject negative MET or duration.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"activity":         {Type: schema.String, Desc: "Optional activity name."},
			"met":              {Type: schema.Number, Desc: "MET value. Must be non-negative.", Required: true},
			"duration_minutes": {Type: schema.Number, Desc: "Duration in minutes. Must be non-negative.", Required: true},
			"days_per_week":    {Type: schema.Number, Desc: "Optional days per week, 0-7."},
		}),
	}, nil
}

func (t *ExerciseCalculator) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input ExerciseCalculatorInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", exerciseCalculatorToolName, err)
	}
	output, err := t.Calculate(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", exerciseCalculatorToolName, err)
	}
	return string(data), nil
}
