package agent

import (
	"context"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewNutritionAgent builds the nutrition specialist.
func NewNutritionAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (adk.Agent, error) {
	return newSpecialistAgent(ctx, chatModel, "nutrition", "Provides bounded youth nutrition references and deficiency checks.", NutritionPrompt, tools.NutritionTools(), 5)
}
