package karl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpressiveLabelSelectors(t *testing.T) {
	interpreter := NewKARLInterpreter()

	tests := []struct {
		name        string
		rule        string
		expectError bool
	}{
		{
			name:        "Simple equality",
			rule:        "REQUIRE pods(app=web) on node",
			expectError: false,
		},
		{
			name:        "Multiple labels",
			rule:        "PREFER pods(app=web,tier=frontend) on zone weight=80",
			expectError: false,
		},
		{
			name:        "In operation",
			rule:        "REPEL pods(app in [web,api,frontend]) on zone weight=90",
			expectError: false,
		},
		{
			name:        "Not in operation",
			rule:        "AVOID pods(env not in [test,debug]) on node",
			expectError: false,
		},
		{
			name:        "Has operation",
			rule:        "REQUIRE pods(has monitoring) on zone",
			expectError: false,
		},
		{
			name:        "Not has operation",
			rule:        "REPEL pods(not has debug) on node weight=75",
			expectError: false,
		},
		{
			name:        "Mixed operations",
			rule:        "PREFER pods(app=web,env not in [test,debug],has monitoring) on zone weight=85",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interpreter.Parse(tt.rule)

			if tt.expectError {
				assert.Error(t, err, "Expected parse error for rule: %s", tt.rule)
			} else {
				require.NoError(t, err, "Unexpected parse error for rule: %s", tt.rule)

				// Validate the parsed rule
				err = interpreter.Validate()
				require.NoError(t, err, "Validation failed for rule: %s", tt.rule)

				// Test conversion to affinity
				_, err = interpreter.ToAffinity()
				assert.NoError(t, err, "ToAffinity conversion failed for rule: %s", tt.rule)
			}
		})
	}
}

func TestKARLInterpreterIntegration(t *testing.T) {
	interpreter := NewKARLInterpreter()

	// Test complete workflow
	rule := "REPEL pods(app in [web,api],has debug) on zone weight=85"

	err := interpreter.Parse(rule)
	require.NoError(t, err, "Failed to parse rule")

	err = interpreter.Validate()
	require.NoError(t, err, "Failed to validate rule")

	affinity, err := interpreter.ToAffinity()
	require.NoError(t, err, "Failed to convert to affinity")

	// Verify the result has pod anti-affinity
	require.NotNil(t, affinity.PodAntiAffinity, "Expected PodAntiAffinity to be set")

	assert.NotEmpty(t, affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		"Expected preferred anti-affinity terms")

	term := affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0]
	assert.Equal(t, int32(85), term.Weight, "Expected weight 85")

	assert.Equal(t, "topology.kubernetes.io/zone", term.PodAffinityTerm.TopologyKey,
		"Expected zone topology key")
}
