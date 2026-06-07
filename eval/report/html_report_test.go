package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractReportJSONFromGoTestOutput(t *testing.T) {
	input := `{"Output":"noise\n"}
{"Output":"PHASE4_EVAL_REPORT_JSON {\"summary\":{\"total\":1},\"results\":[]}\n"}`

	extracted := extractReportJSON(input)

	assert.Equal(t, `{"summary":{"total":1},"results":[]}`, extracted)
}

func TestHTMLReportRendersFullMetricColumns(t *testing.T) {
	cmd := exec.Command("go", "run", "./html_report.go")
	cmd.Stdin = strings.NewReader(`{
		"summary":{"total":1,"passed":1,"failed":0,"pass_rate":1,"safety_compliance":1},
		"results":[{
			"case_id":"E001",
			"tags":["phase-4"],
			"pass":true,
			"task_completion":{"completeness":1,"accuracy":1,"actionability":0.8,"safety":1,"tone":1,"weighted_score":0.97},
			"tool_correctness":1,
			"tool_recall":1,
			"argument_accuracy":1,
			"step_efficiency":1,
			"safety_compliance":true,
			"hallucination_free":true,
			"latency_ms":12,
			"tokens_used":42,
			"total_cost_usd":0.001
		}]
	}`)

	output, err := cmd.CombinedOutput()

	require.NoError(t, err, string(output))
	html := string(output)
	assert.Contains(t, html, "Arg Accuracy")
	assert.Contains(t, html, "Step Efficiency")
	assert.Contains(t, html, "Hallucination")
	assert.Contains(t, html, "Tokens")
	assert.Contains(t, html, "Cost")
	assert.Contains(t, html, "C 1.00 / A 1.00 / Act 0.80 / S 1.00 / T 1.00")
}
