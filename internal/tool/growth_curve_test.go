package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrowthCurveLookupIsConservative(t *testing.T) {
	t.Parallel()

	lookup := NewGrowthCurve()
	output, err := lookup.Lookup(context.Background(), GrowthCurveInput{
		Age:      14,
		Sex:      "female",
		HeightCM: 158,
		WeightKG: 62,
		BMI:      24.84,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.Available)
	assert.Contains(t, output.Description, "does not include authoritative")
	assert.NotEmpty(t, output.Disclaimer)
}

func TestGrowthCurveValidation(t *testing.T) {
	t.Parallel()

	lookup := NewGrowthCurve()
	output, err := lookup.Lookup(context.Background(), GrowthCurveInput{Age: 30, Sex: "female"})

	require.Error(t, err)
	assert.Nil(t, output)
}
