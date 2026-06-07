package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceLookupBMIFormula(t *testing.T) {
	t.Parallel()

	lookup := NewReferenceLookup()
	output, err := lookup.Lookup(context.Background(), ReferenceLookupInput{Topic: "bmi_formula"})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "bmi_formula", output.Topic)
	assert.Contains(t, output.Content, "BMI")
	assert.NotEmpty(t, output.Source)
	assert.NotEmpty(t, output.Disclaimer)
}

func TestReferenceLookupRejectsUnknownTopic(t *testing.T) {
	t.Parallel()

	lookup := NewReferenceLookup()
	output, err := lookup.Lookup(context.Background(), ReferenceLookupInput{Topic: "diagnosis"})

	require.Error(t, err)
	assert.Nil(t, output)
}
