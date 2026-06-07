package agent

import (
	"context"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewMentalHealthAgent builds the mental-health screening specialist.
func NewMentalHealthAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (adk.Agent, error) {
	return newSpecialistAgent(ctx, chatModel, "mental_health", "Screens youth mental-health inputs with PHQ-A style tools and safety escalation.", MentalHealthPrompt, tools.MentalTools(), 6)
}
