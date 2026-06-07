package agent

import (
	"context"
	"fmt"

	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewSupervisorRunner builds the Phase 2 supervisor and wraps sub-agents as tools.
func NewSupervisorRunner(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (*adk.Runner, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if tools == nil {
		return nil, fmt.Errorf("tool registry is required")
	}

	physicalAgent, err := NewPhysicalHealthAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	mentalAgent, err := NewMentalHealthAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	nutritionAgent, err := NewNutritionAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	sleepAgent, err := NewSleepAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	exerciseAgent, err := NewExerciseAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	reportAgent, err := NewReportSynthesisAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}

	supervisorTools := []einotool.BaseTool{
		adk.NewAgentTool(ctx, physicalAgent),
		adk.NewAgentTool(ctx, mentalAgent),
		adk.NewAgentTool(ctx, nutritionAgent),
		adk.NewAgentTool(ctx, sleepAgent),
		adk.NewAgentTool(ctx, exerciseAgent),
		adk.NewAgentTool(ctx, reportAgent),
		tools.RiskFlagger,
	}
	if tools.IntakePipeline != nil {
		supervisorTools = append(supervisorTools, tools.IntakePipeline)
	}
	if tools.ScreeningPipeline != nil {
		supervisorTools = append(supervisorTools, tools.ScreeningPipeline)
	}

	supervisor, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "YouthVital youth health assessment coordinator that routes to specialist agents and synthesizes reports.",
		Instruction: SupervisorPrompt,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               supervisorTools,
				ExecuteSequentially: true,
				ToolCallMiddlewares: []compose.ToolMiddleware{guardrailToolMiddleware(), captureToolCallsMiddleware()},
			},
			EmitInternalEvents: true,
		},
		MaxIterations: 14,
	})
	if err != nil {
		return nil, fmt.Errorf("create supervisor agent: %w", err)
	}

	return adk.NewRunner(ctx, adk.RunnerConfig{Agent: supervisor}), nil
}

func newSpecialistAgent(ctx context.Context, chatModel model.ToolCallingChatModel, name string, description string, instruction string, tools []einotool.BaseTool, maxIterations int) (adk.Agent, error) {
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        name,
		Description: description,
		Instruction: instruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               tools,
				ExecuteSequentially: true,
				ToolCallMiddlewares: []compose.ToolMiddleware{guardrailToolMiddleware(), captureToolCallsMiddleware()},
			},
		},
		MaxIterations: maxIterations,
	})
	if err != nil {
		return nil, fmt.Errorf("create %s agent: %w", name, err)
	}
	return agent, nil
}
