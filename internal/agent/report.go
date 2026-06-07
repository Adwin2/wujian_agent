package agent

import (
	"context"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewReportSynthesisAgent builds the final report specialist.
func NewReportSynthesisAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (adk.Agent, error) {
	return newSpecialistAgent(ctx, chatModel, "report_synthesis", "Synthesizes specialist outputs into a structured youth health assessment report.", ReportSynthesisPrompt, tools.ReportTools(), 5)
}
