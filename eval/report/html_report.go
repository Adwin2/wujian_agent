package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"strings"
)

type evalReport struct {
	Summary struct {
		Total            int     `json:"total"`
		Passed           int     `json:"passed"`
		Failed           int     `json:"failed"`
		PassRate         float64 `json:"pass_rate"`
		SafetyCompliance float64 `json:"safety_compliance"`
	} `json:"summary"`
	Results []struct {
		CaseID         string   `json:"case_id"`
		Tags           []string `json:"tags"`
		Pass           bool     `json:"pass"`
		FailureReason  string   `json:"failure_reason"`
		TaskCompletion struct {
			Completeness  float64 `json:"completeness"`
			Accuracy      float64 `json:"accuracy"`
			Actionability float64 `json:"actionability"`
			Safety        float64 `json:"safety"`
			Tone          float64 `json:"tone"`
			WeightedScore float64 `json:"weighted_score"`
		} `json:"task_completion"`
		ToolCorrectness   float64 `json:"tool_correctness"`
		ToolRecall        float64 `json:"tool_recall"`
		ArgumentAccuracy  float64 `json:"argument_accuracy"`
		StepEfficiency    float64 `json:"step_efficiency"`
		SafetyCompliance  bool    `json:"safety_compliance"`
		HallucinationFree bool    `json:"hallucination_free"`
		LatencyMs         int64   `json:"latency_ms"`
		TokensUsed        int     `json:"tokens_used"`
		TotalCost         float64 `json:"total_cost_usd"`
	} `json:"results"`
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatal(err)
	}
	reportJSON := extractReportJSON(string(data))
	var report evalReport
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		fatal(err)
	}
	fmt.Println("<!doctype html><html><head><meta charset=\"utf-8\"><title>YouthVital Eval Report</title><style>body{font-family:-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;margin:32px;color:#172033}table{border-collapse:collapse;width:100%;margin-top:20px}th,td{border:1px solid #d7dce5;padding:8px;text-align:left}th{background:#f5f7fb}.pass{color:#067647;font-weight:700}.fail{color:#b42318;font-weight:700}.metric{display:inline-block;margin-right:20px;padding:12px 16px;background:#f5f7fb;border-radius:10px}</style></head><body>")
	fmt.Println("<h1>YouthVital Eval Report</h1>")
	fmt.Printf("<div class=\"metric\">Total: %d</div><div class=\"metric\">Passed: %d</div><div class=\"metric\">Pass Rate: %.0f%%</div><div class=\"metric\">Safety: %.0f%%</div>", report.Summary.Total, report.Summary.Passed, report.Summary.PassRate*100, report.Summary.SafetyCompliance*100)
	fmt.Println("<table><thead><tr><th>Case</th><th>Status</th><th>Weighted</th><th>Judge Dimensions</th><th>Tool Precision</th><th>Tool Recall</th><th>Arg Accuracy</th><th>Step Efficiency</th><th>Safety</th><th>Hallucination</th><th>Tokens</th><th>Cost</th><th>Latency</th><th>Tags</th><th>Failure</th></tr></thead><tbody>")
	for _, result := range report.Results {
		status := "FAIL"
		className := "fail"
		if result.Pass {
			status = "PASS"
			className = "pass"
		}
		judgeDimensions := fmt.Sprintf("C %.2f / A %.2f / Act %.2f / S %.2f / T %.2f", result.TaskCompletion.Completeness, result.TaskCompletion.Accuracy, result.TaskCompletion.Actionability, result.TaskCompletion.Safety, result.TaskCompletion.Tone)
		fmt.Printf("<tr><td>%s</td><td class=\"%s\">%s</td><td>%.2f</td><td>%s</td><td>%.2f</td><td>%.2f</td><td>%.2f</td><td>%.2f</td><td>%t</td><td>%t</td><td>%d</td><td>$%.4f</td><td>%dms</td><td>%s</td><td>%s</td></tr>", html.EscapeString(result.CaseID), className, status, result.TaskCompletion.WeightedScore, html.EscapeString(judgeDimensions), result.ToolCorrectness, result.ToolRecall, result.ArgumentAccuracy, result.StepEfficiency, result.SafetyCompliance, result.HallucinationFree, result.TokensUsed, result.TotalCost, result.LatencyMs, html.EscapeString(fmt.Sprint(result.Tags)), html.EscapeString(result.FailureReason))
	}
	fmt.Println("</tbody></table></body></html>")
}

func extractReportJSON(input string) string {
	marker := "PHASE4_EVAL_REPORT_JSON "
	output := collectGoTestOutput(input)
	for _, line := range strings.Split(output, "\n") {
		idx := strings.Index(line, marker)
		if idx >= 0 {
			return strings.TrimSpace(line[idx+len(marker):])
		}
	}
	return strings.TrimSpace(input)
}

func collectGoTestOutput(input string) string {
	var builder strings.Builder
	for _, line := range strings.Split(input, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event struct {
			Output string `json:"Output"`
		}
		if json.Unmarshal([]byte(line), &event) == nil && event.Output != "" {
			builder.WriteString(event.Output)
		}
	}
	if builder.Len() == 0 {
		return input
	}
	return builder.String()
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "generate html report: %v\n", err)
	os.Exit(1)
}
