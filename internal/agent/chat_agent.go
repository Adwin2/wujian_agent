package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	appmodel "github.com/adwin2/youthvital/internal/model"
	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// ChatOptions carries per-request session metadata used by Phase 2 persistence.
type ChatOptions struct {
	UserID    string
	SessionID string
}

// ChatAgent is the Phase 1 boundary consumed by HTTP handlers.
type ChatAgent interface {
	Chat(ctx context.Context, message string) (*appmodel.ChatResponse, error)
}

// SessionChatAgent is implemented by agents that can consume request metadata.
type SessionChatAgent interface {
	ChatWithOptions(ctx context.Context, message string, options ChatOptions) (*appmodel.ChatResponse, error)
}

// Phase1ChatAgent wires deterministic tools and, when configured, an Eino ADK
// ChatModelAgent. The deterministic BMI path keeps Phase 1 verifiable without
// model credentials while the Eino construction remains isolated here.
type Phase1ChatAgent struct {
	tools  *apptool.Registry
	runner *adk.Runner
}

// NewPhase1ChatAgent creates a Phase 1 agent with deterministic tool fallback.
func NewPhase1ChatAgent(tools *apptool.Registry) *Phase1ChatAgent {
	return &Phase1ChatAgent{tools: tools}
}

// NewEinoChatAgent creates a Phase 1 ADK ChatModelAgent with the core tools.
func NewEinoChatAgent(ctx context.Context, chatModel model.BaseChatModel, tools *apptool.Registry) (*Phase1ChatAgent, error) {
	if chatModel == nil {
		return nil, errors.New("chat model is required")
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "phase1_chat",
		Description: "Phase 1 YouthVital chat agent with BMI, growth curve, and reference lookup tools.",
		Instruction: SystemPrompt,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               tools.EinoTools(),
				ExecuteSequentially: true,
			},
		},
		MaxIterations: 6,
	})
	if err != nil {
		return nil, fmt.Errorf("create Eino chat agent: %w", err)
	}

	return &Phase1ChatAgent{
		tools:  tools,
		runner: adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent}),
	}, nil
}

// Chat answers a user message. The required Phase 1 BMI query is handled by
// deterministic parsing and tool invocation so tests do not require LLM access.
func (a *Phase1ChatAgent) Chat(ctx context.Context, message string) (*appmodel.ChatResponse, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errors.New("message is required")
	}

	if input, ok := parseBMIQuestion(message); ok {
		return a.answerBMI(ctx, input)
	}

	if a.runner != nil {
		return a.chatWithEino(ctx, message)
	}

	return &appmodel.ChatResponse{
		Answer: "我可以先帮你计算 BMI。请提供孩子的年龄、性别、身高(cm)和体重(kg)。如果需要生长曲线百分位，Phase 1 还需要接入权威参考表后才能给出。",
	}, nil
}

func (a *Phase1ChatAgent) answerBMI(ctx context.Context, input apptool.BMICalculatorInput) (*appmodel.ChatResponse, error) {
	bmiOutput, err := a.tools.BMI.Calculate(ctx, input)
	if err != nil {
		return nil, err
	}

	refOutput, err := a.tools.ReferenceLookup.Lookup(ctx, apptool.ReferenceLookupInput{Topic: "bmi_interpretation_limitations"})
	if err != nil {
		return nil, err
	}

	answer := fmt.Sprintf("按工具计算：年龄 %.0f 岁、身高 %.0fcm、体重 %.0fkg，BMI = %.2f，约为 %.1f kg/m²。%s%s",
		input.Age,
		input.HeightCM,
		input.WeightKG,
		bmiOutput.BMI,
		bmiOutput.Rounded,
		refOutput.Content,
		refOutput.Disclaimer,
	)

	return &appmodel.ChatResponse{
		Answer: answer,
		ToolCalls: []appmodel.ToolCall{
			{
				Name:   "bmi_calculator",
				Input:  input,
				Output: bmiOutput,
			},
			{
				Name:   "reference_lookup",
				Input:  apptool.ReferenceLookupInput{Topic: "bmi_interpretation_limitations"},
				Output: refOutput,
			},
		},
	}, nil
}

func (a *Phase1ChatAgent) chatWithEino(ctx context.Context, message string) (*appmodel.ChatResponse, error) {
	iter := a.runner.Query(ctx, message)
	var answer strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Err != nil {
			return nil, event.Err
		}
		msg, _, err := adk.TypedGetMessage(event)
		if err != nil {
			return nil, err
		}
		if msg != nil && msg.Content != "" {
			answer.WriteString(msg.Content)
		}
	}

	if strings.TrimSpace(answer.String()) == "" {
		return nil, errors.New("agent returned empty response")
	}
	return &appmodel.ChatResponse{Answer: answer.String()}, nil
}

func parseBMIQuestion(message string) (apptool.BMICalculatorInput, bool) {
	age, ok := extractNumberBeforeAny(message, []string{"岁", "周岁"})
	if !ok {
		return apptool.BMICalculatorInput{}, false
	}
	height, ok := extractNumberBeforeAny(message, []string{"cm", "厘米", "CM"})
	if !ok {
		return apptool.BMICalculatorInput{}, false
	}
	weight, ok := extractNumberBeforeAny(message, []string{"kg", "公斤", "千克", "KG"})
	if !ok {
		return apptool.BMICalculatorInput{}, false
	}

	sex := ""
	switch {
	case strings.Contains(message, "女儿") || strings.Contains(message, "女孩") || strings.Contains(message, "女童"):
		sex = "female"
	case strings.Contains(message, "儿子") || strings.Contains(message, "男孩") || strings.Contains(message, "男童"):
		sex = "male"
	}
	if sex == "" {
		return apptool.BMICalculatorInput{}, false
	}

	return apptool.BMICalculatorInput{
		Age:      age,
		Sex:      sex,
		HeightCM: height,
		WeightKG: weight,
	}, true
}

func extractNumberBeforeAny(text string, units []string) (float64, bool) {
	for _, unit := range units {
		pattern := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*` + regexp.QuoteMeta(unit))
		matches := pattern.FindStringSubmatch(text)
		if len(matches) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			continue
		}
		return value, true
	}
	return 0, false
}

// MarshalToolInput supports trace/debug use when wiring future Eino callbacks.
func MarshalToolInput(input any) string {
	data, err := json.Marshal(input)
	if err != nil {
		return "{}"
	}
	return string(data)
}
