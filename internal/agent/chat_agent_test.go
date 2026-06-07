package agent

import (
	"context"
	"testing"

	"github.com/adwin2/youthvital/internal/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase1ChatAgentAnswersBMIQuestion(t *testing.T) {
	t.Parallel()

	agent := NewPhase1ChatAgent(tool.NewRegistry())
	response, err := agent.Chat(context.Background(), "我女儿14岁158cm62kg的BMI是多少")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Contains(t, response.Answer, "24.84")
	assert.Contains(t, response.Answer, "24.8")
	require.Len(t, response.ToolCalls, 2)
	assert.Equal(t, "bmi_calculator", response.ToolCalls[0].Name)
	assert.Equal(t, "reference_lookup", response.ToolCalls[1].Name)
}

func TestPhase1ChatAgentAsksForMissingBMIData(t *testing.T) {
	t.Parallel()

	agent := NewPhase1ChatAgent(tool.NewRegistry())
	response, err := agent.Chat(context.Background(), "帮我算 BMI")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Contains(t, response.Answer, "年龄")
	assert.Contains(t, response.Answer, "性别")
	assert.Empty(t, response.ToolCalls)
}
