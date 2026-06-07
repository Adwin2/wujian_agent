package graph

import (
	"context"
	"encoding/json"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	intakePipelineToolName    = "intake_pipeline"
	screeningPipelineToolName = "screening_pipeline"
)

// IntakePipelineTool exposes the deterministic intake graph as an Eino tool.
type IntakePipelineTool struct {
	runnable compose.Runnable[*IntakeInput, *IntakeOutput]
}

var _ einotool.InvokableTool = (*IntakePipelineTool)(nil)

func NewIntakePipelineTool(ctx context.Context) (*IntakePipelineTool, error) {
	runnable, err := BuildIntakePipeline(ctx)
	if err != nil {
		return nil, fmt.Errorf("build intake pipeline: %w", err)
	}
	return &IntakePipelineTool{runnable: runnable}, nil
}

func (t *IntakePipelineTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: intakePipelineToolName,
		Desc: "Validate, normalize, enrich, and pre-screen youth health input before specialist routing.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"message": {Type: schema.String, Required: true, Desc: "Raw user health message."},
			"user_profile": {Type: schema.Object, Desc: "Optional known profile with age, sex, height_cm, weight_kg.",
				SubParams: map[string]*schema.ParameterInfo{
					"age":       {Type: schema.Number, Desc: "Age in years."},
					"sex":       {Type: schema.String, Desc: "male or female."},
					"height_cm": {Type: schema.Number, Desc: "Height in centimeters."},
					"weight_kg": {Type: schema.Number, Desc: "Weight in kilograms."},
				}},
		}),
	}, nil
}

func (t *IntakePipelineTool) Run(ctx context.Context, input IntakeInput) (*IntakeOutput, error) {
	output, err := t.runnable.Invoke(ctx, &input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (t *IntakePipelineTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input IntakeInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", intakePipelineToolName, err)
	}
	output, err := t.Run(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", intakePipelineToolName, err)
	}
	return string(data), nil
}

// ScreeningPipelineTool exposes deterministic safety screening as an Eino tool.
type ScreeningPipelineTool struct {
	runnable compose.Runnable[*ScreeningInput, *ScreeningOutput]
}

var _ einotool.InvokableTool = (*ScreeningPipelineTool)(nil)

func NewScreeningPipelineTool(ctx context.Context) (*ScreeningPipelineTool, error) {
	runnable, err := BuildScreeningPipeline(ctx)
	if err != nil {
		return nil, fmt.Errorf("build screening pipeline: %w", err)
	}
	return &ScreeningPipelineTool{runnable: runnable}, nil
}

func (t *ScreeningPipelineTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: screeningPipelineToolName,
		Desc: "Classify deterministic risk hints and decide whether human review is required.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"intake": {Type: schema.Object, Required: true, Desc: "Output from intake_pipeline."},
		}),
	}, nil
}

func (t *ScreeningPipelineTool) Run(ctx context.Context, input ScreeningInput) (*ScreeningOutput, error) {
	output, err := t.runnable.Invoke(ctx, &input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (t *ScreeningPipelineTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input ScreeningInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", screeningPipelineToolName, err)
	}
	output, err := t.Run(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", screeningPipelineToolName, err)
	}
	return string(data), nil
}
