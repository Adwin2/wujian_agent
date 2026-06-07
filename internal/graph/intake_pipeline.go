package graph

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/compose"
)

const (
	intakeValidateNode  = "validate"
	intakeNormalizeNode = "normalize"
	intakeEnrichNode    = "enrich"
	intakeScreenNode    = "screen"
)

// IntakeInput is the raw user input and optional profile entering deterministic pre-processing.
type IntakeInput struct {
	Message string      `json:"message"`
	Profile UserProfile `json:"user_profile,omitempty"`
}

// UserProfile contains known youth demographic and physical data.
type UserProfile struct {
	Age      float64 `json:"age,omitempty"`
	Sex      string  `json:"sex,omitempty"`
	HeightCM float64 `json:"height_cm,omitempty"`
	WeightKG float64 `json:"weight_kg,omitempty"`
}

// IntakeOutput is the normalized input context used by the supervisor or exposed as a tool.
type IntakeOutput struct {
	Message       string      `json:"message"`
	Profile       UserProfile `json:"user_profile,omitempty"`
	MissingFields []string    `json:"missing_fields,omitempty"`
	FocusAreas    []string    `json:"focus_areas,omitempty"`
	RiskHints     []RiskHint  `json:"risk_hints,omitempty"`
	ReadyForTools bool        `json:"ready_for_tools"`
	Clarification string      `json:"clarification,omitempty"`
}

// RiskHint is a deterministic signal passed to screening and risk policy.
type RiskHint struct {
	RiskType    string `json:"risk_type"`
	Severity    string `json:"severity"`
	MetricName  string `json:"metric_name"`
	Value       string `json:"value"`
	Threshold   string `json:"threshold"`
	Description string `json:"description"`
}

// BuildIntakePipeline composes validation, normalization, enrichment, and initial screening.
func BuildIntakePipeline(ctx context.Context) (compose.Runnable[*IntakeInput, *IntakeOutput], error) {
	graph := compose.NewGraph[*IntakeInput, *IntakeOutput]()
	if err := graph.AddLambdaNode(intakeValidateNode, compose.InvokableLambda(validateInput)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(intakeNormalizeNode, compose.InvokableLambda(normalizeUnits)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(intakeEnrichNode, compose.InvokableLambda(enrichWithFocusAreas)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(intakeScreenNode, compose.InvokableLambda(initialScreening)); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, intakeValidateNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(intakeValidateNode, intakeNormalizeNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(intakeNormalizeNode, intakeEnrichNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(intakeEnrichNode, intakeScreenNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(intakeScreenNode, compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("intake_pipeline"))
}

func validateInput(_ context.Context, input *IntakeInput) (*IntakeOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("intake input is required")
	}
	message := strings.TrimSpace(input.Message)
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	return &IntakeOutput{Message: message, Profile: input.Profile}, nil
}

func normalizeUnits(_ context.Context, input *IntakeOutput) (*IntakeOutput, error) {
	out := cloneIntakeOutput(input)
	message := out.Message
	if out.Profile.Age == 0 {
		out.Profile.Age = extractNumberBeforeAny(message, []string{"岁", "周岁"})
	}
	if out.Profile.HeightCM == 0 {
		out.Profile.HeightCM = extractNumberBeforeAny(message, []string{"cm", "厘米", "CM"})
	}
	if out.Profile.WeightKG == 0 {
		out.Profile.WeightKG = extractNumberBeforeAny(message, []string{"kg", "公斤", "千克", "KG"})
	}
	if out.Profile.Sex == "" {
		out.Profile.Sex = inferSex(message)
	}
	out.Profile.Sex = normalizeSex(out.Profile.Sex)
	return out, nil
}

func enrichWithFocusAreas(_ context.Context, input *IntakeOutput) (*IntakeOutput, error) {
	out := cloneIntakeOutput(input)
	message := out.Message
	focusAreas := make([]string, 0)
	if containsAny(message, []string{"bmi", "BMI", "身高", "体重", "胖", "瘦", "头疼", "肚子疼"}) {
		focusAreas = append(focusAreas, "physical")
	}
	if containsAny(message, []string{"情绪", "低落", "哭", "不想上学", "厌学", "自伤", "焦虑", "体像", "减肥"}) {
		focusAreas = append(focusAreas, "mental")
	}
	if containsAny(message, []string{"饮食", "营养", "蔬菜", "水果", "泡面", "零食", "食欲"}) {
		focusAreas = append(focusAreas, "nutrition")
	}
	if containsAny(message, []string{"睡", "熬夜", "凌晨", "起不来", "疲劳", "累"}) {
		focusAreas = append(focusAreas, "sleep")
	}
	if containsAny(message, []string{"运动", "跑步", "锻炼", "活动"}) {
		focusAreas = append(focusAreas, "exercise")
	}
	out.FocusAreas = uniqueStrings(focusAreas)
	return out, nil
}

func initialScreening(_ context.Context, input *IntakeOutput) (*IntakeOutput, error) {
	out := cloneIntakeOutput(input)
	missing := make([]string, 0)
	if containsStringAny(out.FocusAreas, []string{"physical"}) {
		if out.Profile.Age == 0 {
			missing = append(missing, "age")
		}
		if out.Profile.Sex == "" {
			missing = append(missing, "sex")
		}
		if out.Profile.HeightCM == 0 {
			missing = append(missing, "height_cm")
		}
		if out.Profile.WeightKG == 0 {
			missing = append(missing, "weight_kg")
		}
	}
	if containsStringAny(out.FocusAreas, []string{"sleep"}) && !containsAny(out.Message, []string{"小时", "点", "凌晨", "早上", "晚上"}) {
		missing = append(missing, "sleep_hours")
	}
	if len(out.FocusAreas) == 0 && containsAny(out.Message, []string{"整体", "健康情况", "全面"}) {
		missing = append(missing, "height_cm", "weight_kg", "sleep_hours", "diet_description")
		out.FocusAreas = []string{"physical", "nutrition", "sleep", "exercise"}
	}
	out.MissingFields = uniqueStrings(missing)
	out.ReadyForTools = len(out.MissingFields) == 0
	if len(out.MissingFields) > 0 {
		out.Clarification = "需要补充信息后再调用健康工具：" + strings.Join(out.MissingFields, ", ")
	}
	out.RiskHints = initialRiskHints(out)
	return out, nil
}

func initialRiskHints(out *IntakeOutput) []RiskHint {
	message := out.Message
	hints := make([]RiskHint, 0)
	if containsAny(message, []string{"自伤", "轻生", "不想活"}) {
		hints = append(hints, RiskHint{RiskType: "mental", Severity: "critical", MetricName: "self_harm_language", Value: "present", Threshold: "any self-harm signal", Description: "用户文本包含自伤或轻生相关表达。"})
	}
	if containsAny(message, []string{"情绪很低落", "经常哭", "不想上学"}) {
		hints = append(hints, RiskHint{RiskType: "mental", Severity: "high", MetricName: "depressive_symptoms", Value: "present", Threshold: "persistent low mood or school refusal", Description: "用户文本包含持续低落、哭泣或拒学等心理风险信号。"})
	}
	if containsAny(message, []string{"BMI只有", "瘦了很多"}) {
		hints = append(hints, RiskHint{RiskType: "physical", Severity: "high", MetricName: "underweight_or_weight_loss", Value: "present", Threshold: "rapid weight loss or very low BMI", Description: "用户文本包含明显低体重或近期消瘦风险信号。"})
	}
	return hints
}

func cloneIntakeOutput(input *IntakeOutput) *IntakeOutput {
	if input == nil {
		return &IntakeOutput{}
	}
	out := *input
	out.MissingFields = append([]string(nil), input.MissingFields...)
	out.FocusAreas = append([]string(nil), input.FocusAreas...)
	out.RiskHints = append([]RiskHint(nil), input.RiskHints...)
	return &out
}

func extractNumberBeforeAny(text string, units []string) float64 {
	for _, unit := range units {
		pattern := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*` + regexp.QuoteMeta(unit))
		matches := pattern.FindStringSubmatch(text)
		if len(matches) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return value
		}
	}
	return 0
}

func inferSex(message string) string {
	switch {
	case containsAny(message, []string{"女儿", "女孩", "女童"}):
		return "female"
	case containsAny(message, []string{"儿子", "男孩", "男童"}):
		return "male"
	default:
		return ""
	}
}

func normalizeSex(sex string) string {
	switch strings.ToLower(strings.TrimSpace(sex)) {
	case "male", "m", "男", "男孩", "儿子":
		return "male"
	case "female", "f", "女", "女孩", "女儿":
		return "female"
	default:
		return ""
	}
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func containsStringAny(values []string, targets []string) bool {
	for _, value := range values {
		for _, target := range targets {
			if value == target {
				return true
			}
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
