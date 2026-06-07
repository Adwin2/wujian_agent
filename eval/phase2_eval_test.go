package eval

import (
	"context"
	"strings"
	"testing"

	"github.com/adwin2/youthvital/internal/agent"
	"github.com/adwin2/youthvital/internal/tool"
)

type phase2EvalResult struct {
	ID     string
	Tags   []string
	Pass   bool
	Detail string
}

func TestEvalSuite(t *testing.T) {
	cases := []struct {
		id   string
		tags []string
		run  func(context.Context) phase2EvalResult
	}{
		{id: "P2-001", tags: []string{"phase-2", "chat", "tool-trace"}, run: evalDeterministicBMIChat},
		{id: "P2-002", tags: []string{"phase-2", "hitl", "risk"}, run: evalRiskFlaggerHITLSignal},
		{id: "P2-003", tags: []string{"phase-2", "sleep", "validation"}, run: evalSleepZeroHours},
		{id: "P2-004", tags: []string{"phase-2", "exercise", "validation"}, run: evalExerciseMissingRequiredArgs},
		{id: "P2-005", tags: []string{"phase-2", "mental-health", "validation"}, run: evalPHQRequiresNineItems},
	}

	ctx := context.Background()
	results := make([]phase2EvalResult, 0, len(cases))
	for _, c := range cases {
		if !hasTag(c.tags, "phase-2") {
			continue
		}
		result := c.run(ctx)
		result.ID = c.id
		result.Tags = c.tags
		results = append(results, result)
		if !result.Pass {
			t.Errorf("%s failed: %s", result.ID, result.Detail)
		}
	}

	passed := 0
	for _, result := range results {
		if result.Pass {
			passed++
		}
		status := "FAIL"
		if result.Pass {
			status = "PASS"
		}
		t.Logf("EVAL_RESULT id=%s status=%s tags=%s detail=%s", result.ID, status, strings.Join(result.Tags, ","), result.Detail)
	}
	t.Logf("EVAL_SUMMARY tag=phase-2 total=%d passed=%d failed=%d", len(results), passed, len(results)-passed)
}

func evalDeterministicBMIChat(ctx context.Context) phase2EvalResult {
	registry := tool.NewRegistry()
	chatAgent := agent.NewPhase2ChatAgent(registry)
	response, err := chatAgent.Chat(ctx, "我女儿14岁158cm62kg的BMI是多少")
	if err != nil {
		return phase2EvalResult{Pass: false, Detail: err.Error()}
	}
	if !strings.Contains(response.Answer, "24.8") {
		return phase2EvalResult{Pass: false, Detail: "expected BMI answer around 24.8"}
	}
	if len(response.ToolCalls) != 2 {
		return phase2EvalResult{Pass: false, Detail: "expected bmi_calculator and reference_lookup tool calls"}
	}
	if response.ToolCalls[0].Name != "bmi_calculator" || response.ToolCalls[1].Name != "reference_lookup" {
		return phase2EvalResult{Pass: false, Detail: "unexpected tool call sequence"}
	}
	return phase2EvalResult{Pass: true, Detail: "deterministic BMI path produced expected tool trace"}
}

func evalRiskFlaggerHITLSignal(ctx context.Context) phase2EvalResult {
	flagger := tool.NewRiskFlagger()
	output, err := flagger.Flag(ctx, tool.RiskFlaggerInput{
		RiskType:    "physical",
		Severity:    "high",
		MetricName:  "BMI percentile",
		Value:       "<5th percentile",
		Threshold:   "5th percentile",
		Description: "BMI percentile is below the review threshold.",
	})
	if err != nil {
		return phase2EvalResult{Pass: false, Detail: err.Error()}
	}
	if !output.RequireHumanReview {
		return phase2EvalResult{Pass: false, Detail: "expected high severity risk to require human review"}
	}
	return phase2EvalResult{Pass: true, Detail: "high severity risk requires human review"}
}

func evalSleepZeroHours(ctx context.Context) phase2EvalResult {
	hours := 0.0
	output, err := tool.NewSleepScorer().Score(ctx, tool.SleepScorerInput{Age: 15, Hours: &hours})
	if err != nil {
		return phase2EvalResult{Pass: false, Detail: err.Error()}
	}
	if output.Category != "very_insufficient" {
		return phase2EvalResult{Pass: false, Detail: "expected zero sleep hours to be very_insufficient"}
	}
	return phase2EvalResult{Pass: true, Detail: "explicit zero sleep hours accepted and classified"}
}

func evalExerciseMissingRequiredArgs(ctx context.Context) phase2EvalResult {
	_, err := tool.NewExerciseCalculator().Calculate(ctx, tool.ExerciseCalculatorInput{DurationMinutes: floatPtr(30)})
	if err == nil || !strings.Contains(err.Error(), "met is required") {
		return phase2EvalResult{Pass: false, Detail: "expected missing met validation error"}
	}
	return phase2EvalResult{Pass: true, Detail: "missing required MET is rejected"}
}

func evalPHQRequiresNineItems(ctx context.Context) phase2EvalResult {
	_, err := tool.NewPHQScorer().Score(ctx, tool.PHQScorerInput{Items: []int{3, 2, 1}})
	if err == nil || !strings.Contains(err.Error(), "exactly 9") {
		return phase2EvalResult{Pass: false, Detail: "expected PHQ-A item count validation error"}
	}
	return phase2EvalResult{Pass: true, Detail: "incomplete PHQ-A item list is rejected"}
}

func hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func floatPtr(value float64) *float64 {
	return &value
}
