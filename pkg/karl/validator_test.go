package karl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRule(t *testing.T) {
	validator := &Validator{}

	tests := []struct {
		name        string
		rule        KARLRule
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid REQUIRE rule",
			rule: KARLRule{
				RuleType: RuleTypeRequire,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
				TopologyKey: TopologyNode,
			},
			expectError: false,
		},
		{
			name: "Valid PREFER rule with weight",
			rule: KARLRule{
				RuleType: RuleTypePrefer,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
				TopologyKey: TopologyZone,
				Weight:      80,
			},
			expectError: false,
		},
		{
			name: "Missing rule type",
			rule: KARLRule{
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
				TopologyKey: TopologyNode,
			},
			expectError: true,
			errorMsg:    "missing rule type",
		},
		{
			name: "Missing target selector",
			rule: KARLRule{
				RuleType:    RuleTypeRequire,
				TopologyKey: TopologyNode,
			},
			expectError: true,
			errorMsg:    "missing target selector",
		},
		{
			name: "Pods selector without labels",
			rule: KARLRule{
				RuleType: RuleTypeRequire,
				TargetSelector: TargetSelector{
					Type:           "pods",
					LabelSelectors: []LabelSelector{},
				},
				TopologyKey: TopologyNode,
			},
			expectError: true,
			errorMsg:    "pods selector requires labels",
		},
		{
			name: "Missing topology key",
			rule: KARLRule{
				RuleType: RuleTypeRequire,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
			},
			expectError: true,
			errorMsg:    "missing topology key",
		},
		{
			name: "Invalid weight for soft constraint",
			rule: KARLRule{
				RuleType: RuleTypePrefer,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
				TopologyKey: TopologyZone,
				Weight:      150, // Invalid weight > 100
			},
			expectError: true,
			errorMsg:    "soft constraint rule weight must be between 1 and 100",
		},
		{
			name: "Zero weight for soft constraint",
			rule: KARLRule{
				RuleType: RuleTypeRepel,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					},
				},
				TopologyKey: TopologyZone,
				Weight:      0, // Invalid weight = 0
			},
			expectError: true,
			errorMsg:    "soft constraint rule weight must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRule(tt.rule)

			if tt.expectError {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				assert.Equal(t, tt.errorMsg, err.Error(), "Error message mismatch")
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			}
		})
	}
}

func TestValidateLabelSelector(t *testing.T) {
	validator := &Validator{}

	tests := []struct {
		name        string
		selector    LabelSelector
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid equality selector",
			selector: LabelSelector{
				Key:       "app",
				Operation: LabelOpEquals,
				Values:    []string{"web"},
			},
			expectError: false,
		},
		{
			name: "Valid in selector",
			selector: LabelSelector{
				Key:       "env",
				Operation: LabelOpIn,
				Values:    []string{"prod", "staging"},
			},
			expectError: false,
		},
		{
			name: "Valid exists selector",
			selector: LabelSelector{
				Key:       "monitoring",
				Operation: LabelOpExists,
			},
			expectError: false,
		},
		{
			name: "Missing key",
			selector: LabelSelector{
				Operation: LabelOpEquals,
				Values:    []string{"web"},
			},
			expectError: true,
			errorMsg:    "label key cannot be empty",
		},
		{
			name: "Equality selector without values",
			selector: LabelSelector{
				Key:       "app",
				Operation: LabelOpEquals,
				Values:    []string{},
			},
			expectError: true,
			errorMsg:    "equality operation requires exactly one value",
		},
		{
			name: "In selector without values",
			selector: LabelSelector{
				Key:       "env",
				Operation: LabelOpIn,
				Values:    []string{},
			},
			expectError: true,
			errorMsg:    "in operation requires at least one value",
		},
		{
			name: "Exists selector with values",
			selector: LabelSelector{
				Key:       "monitoring",
				Operation: LabelOpExists,
				Values:    []string{"true"},
			},
			expectError: true,
			errorMsg:    "exists operation should not have values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLabelSelector(tt.selector)

			if tt.expectError {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				assert.Equal(t, tt.errorMsg, err.Error(), "Error message mismatch")
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
			}
		})
	}
}
