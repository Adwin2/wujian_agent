package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const sleepScorerToolName = "sleep_scorer"

// SleepScorerInput contains sleep timing or already calculated sleep hours.
type SleepScorerInput struct {
	Age      float64  `json:"age"`
	Bedtime  string   `json:"bedtime,omitempty"`
	WakeTime string   `json:"wake_time,omitempty"`
	Hours    *float64 `json:"hours,omitempty"`
}

// SleepScorerOutput contains a bounded sleep-duration assessment.
type SleepScorerOutput struct {
	Hours          float64 `json:"hours"`
	RecommendedMin float64 `json:"recommended_min"`
	RecommendedMax float64 `json:"recommended_max"`
	Score          int     `json:"score"`
	Category       string  `json:"category"`
	Interpretation string  `json:"interpretation"`
	Source         string  `json:"source"`
	Disclaimer     string  `json:"disclaimer"`
}

// SleepScorer scores youth sleep duration.
type SleepScorer struct{}

var _ einotool.InvokableTool = (*SleepScorer)(nil)

func NewSleepScorer() *SleepScorer { return &SleepScorer{} }

func (t *SleepScorer) Score(_ context.Context, input SleepScorerInput) (*SleepScorerOutput, error) {
	if err := validateAge(input.Age); err != nil {
		return nil, err
	}

	var hours float64
	if input.Hours != nil {
		hours = *input.Hours
	} else {
		parsed, err := hoursBetween(input.Bedtime, input.WakeTime)
		if err != nil {
			return nil, err
		}
		hours = parsed
	}
	if hours < 0 || hours > 24 {
		return nil, fmt.Errorf("sleep hours must be between 0 and 24, got %.2f", hours)
	}

	recommendedMin, recommendedMax := recommendedSleepRange(input.Age)
	category := "adequate"
	score := 100
	interpretation := "睡眠时长在该年龄段常见建议范围内。"
	switch {
	case hours < 6:
		category = "very_insufficient"
		score = 40
		interpretation = "睡眠时长明显不足；如果持续出现，建议家长记录并咨询专业人员。"
	case hours < recommendedMin:
		category = "insufficient"
		score = 70
		interpretation = "睡眠时长低于该年龄段常见建议范围，可能与疲劳、注意力下降相关。"
	case hours > recommendedMax+1:
		category = "excessive"
		score = 75
		interpretation = "睡眠时长高于常见建议范围；如果伴随持续疲劳，应结合生活习惯和健康状况评估。"
	}

	return &SleepScorerOutput{
		Hours:          roundTo(hours, 2),
		RecommendedMin: recommendedMin,
		RecommendedMax: recommendedMax,
		Score:          score,
		Category:       category,
		Interpretation: interpretation,
		Source:         "phase2_adolescent_sleep_reference",
		Disclaimer:     "睡眠评估不能替代医生诊断；如疲劳持续或影响学习生活，请咨询儿科或学校卫生专业人员。",
	}, nil
}

func (t *SleepScorer) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: sleepScorerToolName,
		Desc: "Score youth sleep duration from age and either hours or bedtime/wake_time. Use this for sleep duration and fatigue-related sleep assessment.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"age":       {Type: schema.Number, Desc: "Age in years, 2-20.", Required: true},
			"bedtime":   {Type: schema.String, Desc: "Optional bedtime in HH:MM, e.g. 23:00."},
			"wake_time": {Type: schema.String, Desc: "Optional wake time in HH:MM, e.g. 06:00."},
			"hours":     {Type: schema.Number, Desc: "Optional sleep hours. Must be between 0 and 24."},
		}),
	}, nil
}

func (t *SleepScorer) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input SleepScorerInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", sleepScorerToolName, err)
	}
	output, err := t.Score(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", sleepScorerToolName, err)
	}
	return string(data), nil
}

func recommendedSleepRange(age float64) (float64, float64) {
	if age >= 14 && age <= 17 {
		return 8, 10
	}
	if age >= 6 && age <= 13 {
		return 9, 11
	}
	return 7, 9
}

func hoursBetween(bedtime string, wakeTime string) (float64, error) {
	bedtime = normalizeClock(bedtime)
	wakeTime = normalizeClock(wakeTime)
	if bedtime == "" || wakeTime == "" {
		return 0, fmt.Errorf("either hours or both bedtime and wake_time are required")
	}
	bed, err := parseClock(bedtime)
	if err != nil {
		return 0, fmt.Errorf("invalid bedtime %q: %w", bedtime, err)
	}
	wake, err := parseClock(wakeTime)
	if err != nil {
		return 0, fmt.Errorf("invalid wake_time %q: %w", wakeTime, err)
	}
	if !wake.After(bed) {
		wake = wake.Add(24 * time.Hour)
	}
	return math.Round(wake.Sub(bed).Hours()*100) / 100, nil
}

func normalizeClock(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.Contains(value, ":") {
		return value
	}
	if n, err := strconv.Atoi(value); err == nil {
		return fmt.Sprintf("%02d:00", n)
	}
	return value
}

func parseClock(value string) (time.Time, error) {
	return time.Parse("15:04", value)
}
