package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	appmodel "github.com/adwin2/youthvital/internal/model"
	apptool "github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const hitlMessage = "⚠️ 检测到需要专业人员审查的健康风险，已转交人工审核。"

type phase2TraceKey struct{}

// AssessmentStore persists completed chat turns and risk flags.
type AssessmentStore interface {
	SaveAssessment(ctx context.Context, record appmodel.AssessmentRecord) error
}

// AuditStore persists PHI-related tool access records.
type AuditStore interface {
	SaveAuditLog(ctx context.Context, record appmodel.AuditLogRecord) error
}

type phase2Trace struct {
	mu            sync.Mutex
	toolCalls     []appmodel.ToolCall
	agentsCalled  map[string]struct{}
	riskFlags     []any
	hitlTriggered bool
	hitlReason    string
}

// Phase2ChatAgent executes the supervisor multi-agent graph when configured.
type Phase2ChatAgent struct {
	tools      *apptool.Registry
	store      AssessmentStore
	auditStore AuditStore
	runner     *adk.Runner
}

// NewPhase2ChatAgent creates a Phase 2 agent shell. It needs an Eino runner for
// real LLM verification; without one it falls back to the Phase 1 deterministic agent.
func NewPhase2ChatAgent(tools *apptool.Registry) *Phase2ChatAgent {
	return &Phase2ChatAgent{tools: tools}
}

// WithAssessmentStore wires persistence for completed assessments and risk flags.
func (a *Phase2ChatAgent) WithAssessmentStore(store AssessmentStore) *Phase2ChatAgent {
	a.store = store
	return a
}

// WithAuditStore wires PHI access audit logging.
func (a *Phase2ChatAgent) WithAuditStore(store AuditStore) *Phase2ChatAgent {
	a.auditStore = store
	return a
}

// NewEinoSupervisorChatAgent creates the real Phase 2 supervisor-backed agent.
func NewEinoSupervisorChatAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools *apptool.Registry) (*Phase2ChatAgent, error) {
	runner, err := NewSupervisorRunner(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	return &Phase2ChatAgent{tools: tools, runner: runner}, nil
}

// Chat runs the Phase 2 supervisor when available.
func (a *Phase2ChatAgent) Chat(ctx context.Context, message string) (*appmodel.ChatResponse, error) {
	return a.ChatWithOptions(ctx, message, ChatOptions{})
}

// ChatWithOptions runs the Phase 2 supervisor with request metadata.
func (a *Phase2ChatAgent) ChatWithOptions(ctx context.Context, message string, options ChatOptions) (*appmodel.ChatResponse, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errors.New("message is required")
	}
	if decision := evaluateGuardrail(message); decision.Blocked {
		response := &appmodel.ChatResponse{Answer: decision.Message, HITLTriggered: true}
		return response, a.persistAssessment(ctx, options, message, response)
	}
	if a.runner == nil {
		return NewPhase1ChatAgent(a.tools).Chat(ctx, message)
	}
	return a.chatWithSupervisor(ctx, message, options)
}

func (a *Phase2ChatAgent) chatWithSupervisor(ctx context.Context, message string, options ChatOptions) (*appmodel.ChatResponse, error) {
	trace := newPhase2Trace()
	runCtx := context.WithValue(ctx, phase2TraceKey{}, trace)
	iter := a.runner.Query(runCtx, message, adk.WithAfterToolCallsHook(phase2AfterToolCallsHook))

	var answer strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			trace.markHITL("agent interrupted for human review")
			response := trace.response(hitlMessage)
			return response, a.persistAssessment(ctx, options, message, response)
		}
		if event.Err != nil {
			if trace.isHITLTriggered() {
				return trace.response(hitlMessage), nil
			}
			return nil, event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		if event.Output.MessageOutput.Role != schema.Assistant {
			continue
		}
		msg, _, err := adk.TypedGetMessage(event)
		if err != nil {
			return nil, err
		}
		if msg != nil && strings.TrimSpace(msg.Content) != "" {
			answer.WriteString(msg.Content)
		}
	}

	finalAnswer := strings.TrimSpace(answer.String())
	if trace.isHITLTriggered() {
		finalAnswer = hitlMessage
	}
	if finalAnswer == "" {
		return nil, errors.New("agent returned empty response")
	}
	response := trace.response(finalAnswer)
	return response, a.persistAssessment(ctx, options, message, response)
}

func newPhase2Trace() *phase2Trace {
	return &phase2Trace{agentsCalled: map[string]struct{}{}}
}

func traceFromContext(ctx context.Context) *phase2Trace {
	trace, _ := ctx.Value(phase2TraceKey{}).(*phase2Trace)
	return trace
}

func (t *phase2Trace) recordToolCall(name string, inputJSON string, output string, err error) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if isAgentName(name) {
		t.agentsCalled[name] = struct{}{}
		return
	}

	call := appmodel.ToolCall{Name: name, Input: decodeJSON(inputJSON), Output: decodeJSON(output)}
	if err != nil {
		call.Error = err.Error()
	}
	t.toolCalls = append(t.toolCalls, call)

	if name == "risk_flagger" && output != "" {
		var riskOut apptool.RiskFlaggerOutput
		if json.Unmarshal([]byte(output), &riskOut) == nil {
			t.riskFlags = append(t.riskFlags, riskOut)
			if riskOut.RequireHumanReview {
				t.hitlTriggered = true
				t.hitlReason = fmt.Sprintf("Risk flag: %s severity=%s", riskOut.MetricName, riskOut.Severity)
			}
		}
	}
}

func (t *phase2Trace) markHITL(reason string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hitlTriggered = true
	if reason != "" {
		t.hitlReason = reason
	}
}

func (t *phase2Trace) isHITLTriggered() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.hitlTriggered
}

func (t *phase2Trace) response(answer string) *appmodel.ChatResponse {
	t.mu.Lock()
	defer t.mu.Unlock()

	agents := make([]string, 0, len(t.agentsCalled))
	for agentName := range t.agentsCalled {
		agents = append(agents, agentName)
	}
	sort.Strings(agents)

	toolCalls := make([]appmodel.ToolCall, len(t.toolCalls))
	copy(toolCalls, t.toolCalls)

	return &appmodel.ChatResponse{
		Answer:        answer,
		ToolCalls:     toolCalls,
		AgentsCalled:  agents,
		HITLTriggered: t.hitlTriggered,
	}
}

func (t *phase2Trace) riskFlagSnapshot() []any {
	t.mu.Lock()
	defer t.mu.Unlock()

	riskFlags := make([]any, len(t.riskFlags))
	copy(riskFlags, t.riskFlags)
	return riskFlags
}

func phase2AfterToolCallsHook(ctx context.Context) error {
	trace := traceFromContext(ctx)
	if trace != nil && trace.isHITLTriggered() {
		return compose.Interrupt(ctx, hitlMessage)
	}

	var shouldInterrupt bool
	err := compose.ProcessState(ctx, func(_ context.Context, st *adk.State) error {
		for _, msg := range st.Messages {
			if msg == nil || msg.Role != schema.Tool || msg.ToolName != "risk_flagger" {
				continue
			}
			var riskOut apptool.RiskFlaggerOutput
			if json.Unmarshal([]byte(msg.Content), &riskOut) == nil && riskOut.RequireHumanReview {
				shouldInterrupt = true
				if trace != nil {
					trace.markHITL(fmt.Sprintf("Risk flag: %s severity=%s", riskOut.MetricName, riskOut.Severity))
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("process agent state for HITL: %w", err)
	}
	if shouldInterrupt {
		return compose.Interrupt(ctx, hitlMessage)
	}
	return nil
}

func (a *Phase2ChatAgent) persistAssessment(ctx context.Context, options ChatOptions, input string, response *appmodel.ChatResponse) error {
	if a.store == nil || response == nil {
		return nil
	}
	userID := strings.TrimSpace(options.UserID)
	if err := a.store.SaveAssessment(ctx, appmodel.AssessmentRecord{
		UserID:        userID,
		SessionID:     strings.TrimSpace(options.SessionID),
		InputText:     input,
		OutputText:    response.Answer,
		AgentsCalled:  response.AgentsCalled,
		ToolCalls:     response.ToolCalls,
		RiskFlags:     traceRiskFlags(response.ToolCalls),
		HITLTriggered: response.HITLTriggered,
	}); err != nil {
		return err
	}
	return a.persistAuditLogs(ctx, userID, response.ToolCalls)
}

func traceRiskFlags(toolCalls []appmodel.ToolCall) []any {
	riskFlags := make([]any, 0)
	for _, call := range toolCalls {
		if call.Name == "risk_flagger" && call.Output != nil {
			riskFlags = append(riskFlags, call.Output)
		}
	}
	return riskFlags
}

func (a *Phase2ChatAgent) persistAuditLogs(ctx context.Context, userID string, toolCalls []appmodel.ToolCall) error {
	if a.auditStore == nil {
		return nil
	}
	for _, call := range toolCalls {
		if !isPHITool(call.Name) {
			continue
		}
		if err := a.auditStore.SaveAuditLog(ctx, appmodel.AuditLogRecord{
			UserID:       userID,
			Action:       "tool_access",
			ResourceType: "phi",
			ToolName:     call.Name,
			ToolInput:    call.Input,
			ToolOutput:   call.Output,
		}); err != nil {
			return err
		}
	}
	return nil
}

func isPHITool(name string) bool {
	switch name {
	case "bmi_calculator", "growth_curve", "sleep_scorer", "phq_scorer", "exercise_calculator", "risk_flagger", "history_query", "report_generator", "intake_pipeline", "screening_pipeline":
		return true
	default:
		return false
	}
}

func captureToolCallsMiddleware() compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				output, err := next(ctx, input)
				result := ""
				if output != nil {
					result = output.Result
				}
				if trace := traceFromContext(ctx); trace != nil && input != nil {
					trace.recordToolCall(input.Name, input.Arguments, result, err)
				}
				return output, err
			}
		},
	}
}

func decodeJSON(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return value
	}
	return decoded
}

func isAgentName(name string) bool {
	switch name {
	case "physical_health", "mental_health", "nutrition", "sleep", "exercise", "report_synthesis":
		return true
	default:
		return false
	}
}
