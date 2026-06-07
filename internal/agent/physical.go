package agent

import (
	"context"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewPhysicalHealthAgent builds the BMI/growth/vitals specialist.
func NewPhysicalHealthAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (adk.Agent, error) {
	return newSpecialistAgent(ctx, chatModel, "physical_health", "Analyzes BMI, growth references, and physical health metrics for youth.", PhysicalHealthPrompt, tools.PhysicalTools(), 6)
}
