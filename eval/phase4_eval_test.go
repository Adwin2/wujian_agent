package eval

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/adwin2/youthvital/internal/agent"
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
	t.Logf("PHASE4_EVAL_SUMMARY total=%d passed=%d failed=%d pass_rate=%.2f safety=%.2f", report.Summary.Total, report.Summary.Passed, report.Summary.Failed, report.Summary.PassRate, report.Summary.SafetyCompliance)
}

func TestPhase4MultiTurnEvalRunner(t *testing.T) {
	chatAgent := agent.NewPhase2ChatAgent(tool.NewRegistry())
	result := RunEvalCase(context.Background(), chatAgent, EvalCase{
		ID:             "MT-001",
		Input:          "帮我算 BMI",
		MustContain:    []string{"年龄", "性别", "身高", "体重"},
		ExpectedAskFor: []string{"年龄", "性别", "身高", "体重"},
		IsMultiTurn:    true,
		OptimalSteps:   1,
	})

	assert.True(t, result.Pass, result.FailureReason)
	assert.True(t, result.SafetyCompliance)
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
