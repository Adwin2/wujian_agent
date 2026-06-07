package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBMICalculatorCalculate(t *testing.T) {
	t.Parallel()

	calculator := NewBMICalculator()
	output, err := calculator.Calculate(context.Background(), BMICalculatorInput{
		Age:      14,
		Sex:      "female",
		HeightCM: 158,
		WeightKG: 62,
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.InDelta(t, 24.84, output.BMI, 0.01)
	assert.Equal(t, 24.8, output.Rounded)
	assert.Equal(t, "kg/m^2", output.Unit)
}

func TestBMICalculatorValidation(t *testing.T) {
	t.Parallel()

	valid := BMICalculatorInput{Age: 14, Sex: "female", HeightCM: 158, WeightKG: 62}
	tests := []struct {
		name  string
		input BMICalculatorInput
	}{
		{name: "age too low", input: BMICalculatorInput{Age: 1, Sex: valid.Sex, HeightCM: valid.HeightCM, WeightKG: valid.WeightKG}},
		{name: "age too high", input: BMICalculatorInput{Age: 21, Sex: valid.Sex, HeightCM: valid.HeightCM, WeightKG: valid.WeightKG}},
		{name: "height zero", input: BMICalculatorInput{Age: valid.Age, Sex: valid.Sex, HeightCM: 0, WeightKG: valid.WeightKG}},
		{name: "height too high", input: BMICalculatorInput{Age: valid.Age, Sex: valid.Sex, HeightCM: 251, WeightKG: valid.WeightKG}},
		{name: "weight zero", input: BMICalculatorInput{Age: valid.Age, Sex: valid.Sex, HeightCM: valid.HeightCM, WeightKG: 0}},
		{name: "weight too high", input: BMICalculatorInput{Age: valid.Age, Sex: valid.Sex, HeightCM: valid.HeightCM, WeightKG: 301}},
		{name: "invalid sex", input: BMICalculatorInput{Age: valid.Age, Sex: "unknown", HeightCM: valid.HeightCM, WeightKG: valid.WeightKG}},
	}

	calculator := NewBMICalculator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := calculator.Calculate(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, output)
		})
	}
}

func TestBMICalculatorInvokableRun(t *testing.T) {
	t.Parallel()

	calculator := NewBMICalculator()
	result, err := calculator.InvokableRun(context.Background(), `{"age":14,"sex":"女儿","height_cm":158,"weight_kg":62}`)

	require.NoError(t, err)
	assert.Contains(t, result, `"bmi":24.84`)
}
