package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
)

// Registry owns deterministic tools used by Phase 2 agents.
type Registry struct {
	BMI                *BMICalculator
	GrowthCurve        *GrowthCurve
	ReferenceLookup    *ReferenceLookup
	NutritionLookup    *NutritionLookup
	SleepScorer        *SleepScorer
	PHQScorer          *PHQScorer
	ExerciseCalculator *ExerciseCalculator
	RiskFlagger        *RiskFlagger
	HistoryQuery       *HistoryQuery
	ReportGenerator    *ReportGenerator
	AlertSender        *AlertSender
	AppointmentBooker  *AppointmentBooker
}

// NewRegistry constructs all YouthVital tools.
func NewRegistry() *Registry {
	return &Registry{
		BMI:                NewBMICalculator(),
		GrowthCurve:        NewGrowthCurve(),
		ReferenceLookup:    NewReferenceLookup(),
		NutritionLookup:    NewNutritionLookup(),
		SleepScorer:        NewSleepScorer(),
		PHQScorer:          NewPHQScorer(),
		ExerciseCalculator: NewExerciseCalculator(),
		RiskFlagger:        NewRiskFlagger(),
		HistoryQuery:       NewHistoryQuery(),
		ReportGenerator:    NewReportGenerator(),
		AlertSender:        NewAlertSender(),
		AppointmentBooker:  NewAppointmentBooker(),
	}
}

// EinoTools exposes every deterministic tool for Eino ADK tool registration.
func (r *Registry) EinoTools() []einotool.BaseTool {
	return []einotool.BaseTool{
		r.BMI,
		r.GrowthCurve,
		r.ReferenceLookup,
		r.NutritionLookup,
		r.SleepScorer,
		r.PHQScorer,
		r.ExerciseCalculator,
		r.RiskFlagger,
		r.HistoryQuery,
		r.ReportGenerator,
		r.AlertSender,
		r.AppointmentBooker,
	}
}

// PhysicalTools are used by the physical_health agent.
func (r *Registry) PhysicalTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.BMI, r.GrowthCurve, r.ReferenceLookup, r.RiskFlagger, r.HistoryQuery}
}

// SleepTools are used by the sleep agent.
func (r *Registry) SleepTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.SleepScorer, r.ReferenceLookup, r.RiskFlagger, r.HistoryQuery}
}

// MentalTools are used by the mental_health agent.
func (r *Registry) MentalTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.PHQScorer, r.ReferenceLookup, r.RiskFlagger, r.AlertSender, r.AppointmentBooker}
}

// NutritionTools are used by the nutrition agent.
func (r *Registry) NutritionTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.NutritionLookup, r.ReferenceLookup, r.RiskFlagger, r.HistoryQuery}
}

// ExerciseTools are used by the exercise agent.
func (r *Registry) ExerciseTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.ExerciseCalculator, r.ReferenceLookup, r.RiskFlagger, r.HistoryQuery}
}

// ReportTools are used by report_synthesis.
func (r *Registry) ReportTools() []einotool.BaseTool {
	return []einotool.BaseTool{r.ReportGenerator, r.ReferenceLookup}
}

// Invoke executes a registered tool by name using JSON arguments.
func (r *Registry) Invoke(ctx context.Context, name string, argumentsInJSON string) (string, error) {
	for _, candidate := range r.invokableTools() {
		info, err := candidate.Info(ctx)
		if err != nil {
			return "", err
		}
		if info.Name == name {
			return candidate.InvokableRun(ctx, argumentsInJSON)
		}
	}
	return "", fmt.Errorf("unknown tool %q", name)
}

func (r *Registry) invokableTools() []einotool.InvokableTool {
	return []einotool.InvokableTool{
		r.BMI,
		r.GrowthCurve,
		r.ReferenceLookup,
		r.NutritionLookup,
		r.SleepScorer,
		r.PHQScorer,
		r.ExerciseCalculator,
		r.RiskFlagger,
		r.HistoryQuery,
		r.ReportGenerator,
		r.AlertSender,
		r.AppointmentBooker,
	}
}
