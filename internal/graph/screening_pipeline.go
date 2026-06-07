package graph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

const (
	screeningCollectNode  = "collect_risk_hints"
	screeningClassifyNode = "classify_risk"
)

// ScreeningInput is the deterministic risk-screening input.
type ScreeningInput struct {
	Intake *IntakeOutput `json:"intake"`
}

// ScreeningOutput summarizes deterministic safety routing before agent advice.
type ScreeningOutput struct {
	RiskHints          []RiskHint `json:"risk_hints,omitempty"`
	RequireHumanReview bool       `json:"require_human_review"`
	HighestSeverity    string     `json:"highest_severity,omitempty"`
	BlockedAdviceAreas []string   `json:"blocked_advice_areas,omitempty"`
	Message            string     `json:"message,omitempty"`
}

// BuildScreeningPipeline composes deterministic risk collection and classification.
func BuildScreeningPipeline(ctx context.Context) (compose.Runnable[*ScreeningInput, *ScreeningOutput], error) {
	graph := compose.NewGraph[*ScreeningInput, *ScreeningOutput]()
	if err := graph.AddLambdaNode(screeningCollectNode, compose.InvokableLambda(collectRiskHints)); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode(screeningClassifyNode, compose.InvokableLambda(classifyRisk)); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, screeningCollectNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(screeningCollectNode, screeningClassifyNode); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(screeningClassifyNode, compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("screening_pipeline"))
}

func collectRiskHints(_ context.Context, input *ScreeningInput) (*ScreeningOutput, error) {
	if input == nil || input.Intake == nil {
		return nil, fmt.Errorf("intake is required")
	}
	return &ScreeningOutput{RiskHints: append([]RiskHint(nil), input.Intake.RiskHints...)}, nil
}

func classifyRisk(_ context.Context, output *ScreeningOutput) (*ScreeningOutput, error) {
	if output == nil {
		return nil, fmt.Errorf("screening output is required")
	}
	highest := ""
	blocked := make([]string, 0)
	for _, hint := range output.RiskHints {
		if severityRank(hint.Severity) > severityRank(highest) {
			highest = hint.Severity
		}
		if hint.Severity == "high" || hint.Severity == "critical" {
			output.RequireHumanReview = true
			blocked = append(blocked, hint.RiskType)
		}
	}
	output.HighestSeverity = highest
	output.BlockedAdviceAreas = uniqueStrings(blocked)
	if output.RequireHumanReview {
		output.Message = "检测到需要专业人员审查的健康风险，应停止对应领域自动建议并转交人工审核。"
	}
	return output, nil
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
