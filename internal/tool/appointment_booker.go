package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const appointmentBookerToolName = "appointment_booker"

// AppointmentBookerInput describes an appointment request. Phase 2 is dry-run.
type AppointmentBookerInput struct {
	Department string `json:"department"`
	Reason     string `json:"reason"`
	Preferred  string `json:"preferred,omitempty"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

// AppointmentBookerOutput reports dry-run booking status.
type AppointmentBookerOutput struct {
	Booked     bool   `json:"booked"`
	DryRun     bool   `json:"dry_run"`
	Department string `json:"department"`
	Status     string `json:"status"`
	Disclaimer string `json:"disclaimer"`
}

// AppointmentBooker prepares appointment guidance without external booking.
type AppointmentBooker struct{}

var _ einotool.InvokableTool = (*AppointmentBooker)(nil)

func NewAppointmentBooker() *AppointmentBooker { return &AppointmentBooker{} }

func (t *AppointmentBooker) Book(_ context.Context, input AppointmentBookerInput) (*AppointmentBookerOutput, error) {
	department := strings.TrimSpace(input.Department)
	if department == "" {
		return nil, fmt.Errorf("department is required")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, fmt.Errorf("reason is required")
	}
	return &AppointmentBookerOutput{
		Booked:     false,
		DryRun:     true,
		Department: department,
		Status:     "Phase 2 dry-run only; no real appointment was booked.",
		Disclaimer: "如症状明显、持续或紧急，请直接联系当地医疗机构；本系统当前不执行真实预约。",
	}, nil
}

func (t *AppointmentBooker) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: appointmentBookerToolName,
		Desc: "Prepare appointment guidance in dry-run mode. Does not book real appointments in Phase 2.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"department": {Type: schema.String, Required: true},
			"reason":     {Type: schema.String, Required: true},
			"preferred":  {Type: schema.String, Desc: "Optional preferred time or hospital."},
			"dry_run":    {Type: schema.Boolean, Desc: "Always treated as true in Phase 2."},
		}),
	}, nil
}

func (t *AppointmentBooker) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var input AppointmentBookerInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("decode %s arguments: %w", appointmentBookerToolName, err)
	}
	output, err := t.Book(ctx, input)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("encode %s output: %w", appointmentBookerToolName, err)
	}
	return string(data), nil
}
