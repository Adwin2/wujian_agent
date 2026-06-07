package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSleepScorerAcceptsExplicitZeroHours(t *testing.T) {
	t.Parallel()

	hours := 0.0
	scorer := NewSleepScorer()
	output, err := scorer.Score(context.Background(), SleepScorerInput{Age: 15, Hours: &hours})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 0.0, output.Hours)
	assert.Equal(t, "very_insufficient", output.Category)
}

func TestExerciseCalculatorRequiresMETAndDuration(t *testing.T) {
	t.Parallel()

	calculator := NewExerciseCalculator()
	_, err := calculator.Calculate(context.Background(), ExerciseCalculatorInput{DurationMinutes: floatPtr(30)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "met is required")

	_, err = calculator.Calculate(context.Background(), ExerciseCalculatorInput{MET: floatPtr(3.5)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duration_minutes is required")
}

func TestPHQScorerRequiresNineItems(t *testing.T) {
	t.Parallel()

	scorer := NewPHQScorer()
	_, err := scorer.Score(context.Background(), PHQScorerInput{Items: []int{3, 2, 1}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 9")
}

func floatPtr(value float64) *float64 {
	return &value
}
