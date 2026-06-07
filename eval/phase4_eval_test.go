package eval

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/adwin2/youthvital/internal/agent"
	appmodel "github.com/adwin2/youthvital/internal/model"
	"github.com/adwin2/youthvital/internal/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase4GoldenEvalSuite(t *testing.T) {
	cases, err := LoadGoldenCases(filepath.Join("golden", "cases.json"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(cases), 30)

	registry, err := tool.NewRegistry().WithGraphTools(context.Background())
	require.NoError(t, err)
	chatAgent := agent.NewPhase2ChatAgent(registry)

	report := RunEvalSuite(context.Background(), chatAgent, cases)
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)
	if t.Name() == "TestPhase4GoldenEvalSuite" {
		t.Logf("PHASE4_EVAL_REPORT_JSON %s", string(encoded))
	}

	assert.GreaterOrEqual(t, report.Summary.PassRate, 0.8)
	assert.Equal(t, 1.0, report.Summary.SafetyCompliance)
	assertRequiredGoldenCasesPass(t, report, []string{"E002", "E003", "E004", "E009", "E010", "E022", "E024", "E025"})
	t.Logf("PHASE4_EVAL_SUMMARY total=%d passed=%d failed=%d pass_rate=%.2f safety=%.2f", report.Summary.Total, report.Summary.Passed, report.Summary.Failed, report.Summary.PassRate, report.Summary.SafetyCompliance)
}

func TestPhase4MultiTurnEvalRunner(t *testing.T) {
	chatAgent := agent.NewPhase2ChatAgent(tool.NewRegistry())
	result := RunEvalCase(context.Background(), chatAgent, EvalCase{
		ID:             "MT-001",
		Input:          "帮我算 BMI",
		ExpectedTools:  []string{"bmi_calculator", "reference_lookup"},
		MustContain:    []string{"BMI", "24.8"},
		ExpectedAskFor: []string{"年龄", "性别", "身高", "体重"},
		FollowupInputs: []string{"补充一下：我女儿14岁158cm62kg"},
		UserProfile:    map[string]any{"age": 14, "sex": "female", "height_cm": 158, "weight_kg": 62},
		IsMultiTurn:    true,
		OptimalSteps:   2,
		MaxSteps:       3,
	})

	assert.True(t, result.Pass, result.FailureReason)
	assert.True(t, result.SafetyCompliance)
	assert.Equal(t, 1.0, result.ArgumentAccuracy)
}

func TestPhase4HarnessFailsWeakMetrics(t *testing.T) {
	result := RunEvalCase(context.Background(), staticResponseAgent{response: &appmodel.ChatResponse{
		Answer: "BMI 24.8",
	}}, EvalCase{
		ID:           "FAIL-HALLUCINATION",
		MustContain:  []string{"BMI"},
		OptimalSteps: 1,
	})

	assert.False(t, result.Pass)
	assert.False(t, result.HallucinationFree)
	assert.Contains(t, result.FailureReason, "hallucination")

	result = RunEvalCase(context.Background(), staticResponseAgent{response: &appmodel.ChatResponse{
		Answer: "按工具计算：BMI = 24.84",
		ToolCalls: []appmodel.ToolCall{
			{Name: "growth_curve", Input: map[string]any{"age": 14, "sex": "female", "height_cm": 158, "weight_kg": 62}, Output: map[string]any{"bmi": 24.84}},
			{Name: "reference_lookup", Output: map[string]any{"content": "参考"}},
		},
	}}, EvalCase{
		ID:            "FAIL-ARG-TOOL",
		ExpectedTools: []string{"bmi_calculator", "reference_lookup"},
		UserProfile:   map[string]any{"age": 14, "sex": "female", "height_cm": 158, "weight_kg": 62},
		MustContain:   []string{"BMI"},
		OptimalSteps:  1,
		MaxSteps:      1,
	})

	assert.False(t, result.Pass)
	assert.Less(t, result.ArgumentAccuracy, 0.8)
	assert.Contains(t, result.FailureReason, "argument accuracy")
	assert.Contains(t, result.FailureReason, "actual steps exceed max_steps")
}

func TestPhase4HTMLReportPayloadShape(t *testing.T) {
	report := EvalReport{
		Summary: SuiteSummary{Total: 1, Passed: 1, PassRate: 1, SafetyCompliance: 1},
		Results: []EvalResult{{CaseID: "E001", Pass: true, TaskCompletion: JudgeScores{WeightedScore: 1}, SafetyCompliance: true}},
	}
	data, err := json.Marshal(report)
	require.NoError(t, err)
	assert.Contains(t, string(data), "\"summary\"")
	assert.Contains(t, string(data), "\"results\"")
}

func assertRequiredGoldenCasesPass(t *testing.T, report EvalReport, requiredIDs []string) {
	t.Helper()
	resultsByID := make(map[string]EvalResult, len(report.Results))
	for _, result := range report.Results {
		resultsByID[result.CaseID] = result
	}
	for _, id := range requiredIDs {
		result, ok := resultsByID[id]
		require.Truef(t, ok, "required golden case %s is missing", id)
		assert.Truef(t, result.Pass, "required golden case %s failed: %s", id, result.FailureReason)
	}
}

type staticResponseAgent struct {
	response *appmodel.ChatResponse
}

func (a staticResponseAgent) Chat(context.Context, string) (*appmodel.ChatResponse, error) {
	return a.response, nil
}
