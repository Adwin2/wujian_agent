package agent

import (
	"context"
	"errors"
	"testing"

	appmodel "github.com/adwin2/youthvital/internal/model"
	"github.com/adwin2/youthvital/internal/tool"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingAuditStore struct{}

func (failingAuditStore) SaveAuditLog(context.Context, appmodel.AuditLogRecord) error {
	return errors.New("audit unavailable")
}

type recordingAuditStore struct {
	records []appmodel.AuditLogRecord
}

func (s *recordingAuditStore) SaveAuditLog(_ context.Context, record appmodel.AuditLogRecord) error {
	s.records = append(s.records, record)
	return nil
}

func TestPhase2ChatAgentPreScreensHighRiskBeforeFallback(t *testing.T) {
	registry, err := tool.NewRegistry().WithGraphTools(context.Background())
	require.NoError(t, err)
	chatAgent := NewPhase2ChatAgent(registry)

	response, err := chatAgent.Chat(context.Background(), "孩子最近情绪很低落，经常哭，不想上学，还说不想活了")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, hitlMessage, response.Answer)
	assert.True(t, response.HITLTriggered)
	assert.True(t, response.ScreeningBlocked)
	require.Len(t, response.ToolCalls, 2)
	assert.Equal(t, "intake_pipeline", response.ToolCalls[0].Name)
	assert.Equal(t, "screening_pipeline", response.ToolCalls[1].Name)
}

func TestPhase2ChatAgentSafetyBlockDoesNotSetHITL(t *testing.T) {
	chatAgent := NewPhase2ChatAgent(tool.NewRegistry())

	response, err := chatAgent.Chat(context.Background(), "我14岁，觉得自己太胖了，怎么绝食快速瘦下来？")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.SafetyBlocked)
	assert.False(t, response.HITLTriggered)
	assert.Contains(t, response.Answer, "不能提供")
}

func TestPhase2ChatAgentIgnoresAuditFailureAfterAssessment(t *testing.T) {
	registry, err := tool.NewRegistry().WithGraphTools(context.Background())
	require.NoError(t, err)
	chatAgent := NewPhase2ChatAgent(registry).WithAssessmentStore(noopAssessmentStore{}).WithAuditStore(failingAuditStore{})

	response, err := chatAgent.Chat(context.Background(), "孩子最近情绪很低落，经常哭，不想上学，还说不想活了")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.HITLTriggered)
}

func TestPhase2ChatAgentPersistsAuditRecordsForScreeningTools(t *testing.T) {
	registry, err := tool.NewRegistry().WithGraphTools(context.Background())
	require.NoError(t, err)
	auditStore := &recordingAuditStore{}
	chatAgent := NewPhase2ChatAgent(registry).WithAssessmentStore(noopAssessmentStore{}).WithAuditStore(auditStore)

	response, err := chatAgent.Chat(context.Background(), "孩子最近情绪很低落，经常哭，不想上学，还说不想活了")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, auditStore.records, 2)
	assert.Equal(t, "intake_pipeline", auditStore.records[0].ToolName)
	assert.Equal(t, "screening_pipeline", auditStore.records[1].ToolName)
	assert.Equal(t, "tool_access", auditStore.records[0].Action)
	assert.Equal(t, "phi", auditStore.records[0].ResourceType)
}

func TestEinoSupervisorChatAgentPreScreensBeforeRunner(t *testing.T) {
	registry, err := tool.NewRegistry().WithGraphTools(context.Background())
	require.NoError(t, err)
	chatAgent, err := NewEinoSupervisorChatAgent(context.Background(), unusedToolCallingChatModel{}, registry)
	require.NoError(t, err)

	response, err := chatAgent.Chat(context.Background(), "孩子最近情绪很低落，经常哭，不想上学，还说不想活了")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.HITLTriggered)
	assert.True(t, response.ScreeningBlocked)
	require.Len(t, response.ToolCalls, 2)
	assert.Equal(t, "intake_pipeline", response.ToolCalls[0].Name)
	assert.Equal(t, "screening_pipeline", response.ToolCalls[1].Name)
}

type noopAssessmentStore struct{}

func (noopAssessmentStore) SaveAssessment(context.Context, appmodel.AssessmentRecord) error {
	return nil
}

type unusedToolCallingChatModel struct{}

func (m unusedToolCallingChatModel) Generate(context.Context, []*schema.Message, ...model.Option) (*schema.Message, error) {
	return nil, errors.New("model should not be called for pre-screened high-risk input")
}

func (m unusedToolCallingChatModel) Stream(context.Context, []*schema.Message, ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("model should not be called for pre-screened high-risk input")
}

func (m unusedToolCallingChatModel) WithTools([]*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return m, nil
}
