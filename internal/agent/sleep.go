package agent

import (
	"context"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewSleepAgent builds the sleep specialist.
func NewSleepAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (adk.Agent, error) {
	return newSpecialistAgent(ctx, chatModel, "sleep", "Assesses adolescent sleep duration, sleep quality, and fatigue-related sleep factors.", SleepPrompt, tools.SleepTools(), 6)
}
