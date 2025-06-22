package karl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRule(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name         string
		rule         string
		expectError  bool
		expectedType RuleType
	}{
		{
			name:         "REQUIRE rule",
			rule:         "REQUIRE pods(app=web) on node",
			expectError:  false,
			expectedType: RuleTypeRequire,
		},
		{
			name:         "PREFER rule with weight",
			rule:         "PREFER pods(app=web) on zone weight=80",
			expectError:  false,
			expectedType: RuleTypePrefer,
		},
		{
			name:         "AVOID rule",
			rule:         "AVOID pods(app=test) on node",
			expectError:  false,
			expectedType: RuleTypeAvoid,
		},
		{
			name:         "REPEL rule with weight",
			rule:         "REPEL pods(app=batch) on zone weight=90",
			expectError:  false,
			expectedType: RuleTypeRepel,
		},
		{
			name:        "Invalid rule type",
			rule:        "INVALID pods(app=web) on node",
			expectError: true,
		},
		{
			name:        "Too few tokens",
			rule:        "REQUIRE pods(app=web)",
			expectError: true,
		},
		{
			name:        "Invalid topology",
			rule:        "REQUIRE pods(app=web) on invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := parser.ParseRule(tt.rule)

			if tt.expectError {
				assert.Error(t, err, "Expected error for rule: %s", tt.rule)
			} else {
				require.NoError(t, err, "Unexpected error for rule: %s", tt.rule)
				assert.Equal(t, tt.expectedType, rule.RuleType, "Rule type mismatch")
			}
		})
	}
}

func TestParseLabelSelectors(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name     string
		input    string
		expected int // number of label selectors expected
	}{
		{
			name:     "Simple equality",
			input:    "app=web",
			expected: 1,
		},
		{
			name:     "Multiple labels",
			input:    "app=web,tier=frontend",
			expected: 2,
		},
		{
			name:     "In operation",
			input:    "app in [web,api]",
			expected: 1,
		},
		{
			name:     "Not in operation",
			input:    "env not in [test,debug]",
			expected: 1,
		},
		{
			name:     "Has operation",
			input:    "has monitoring",
			expected: 1,
		},
		{
			name:     "Not has operation",
			input:    "not has debug",
			expected: 1,
		},
		{
			name:     "Mixed operations",
			input:    "app=web,env not in [test],has prod",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selectors, err := parser.parseLabelSelectors(tt.input)
			require.NoError(t, err, "Unexpected error parsing label selectors")
			assert.Len(t, selectors, tt.expected, "Number of selectors mismatch for input: %s", tt.input)
		})
	}
}

func TestParseSingleLabelSelector(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name              string
		input             string
		expectedOperation LabelOperation
		expectedKey       string
		expectedValues    []string
	}{
		{
			name:              "Simple equality",
			input:             "app=web",
			expectedOperation: LabelOpEquals,
			expectedKey:       "app",
			expectedValues:    []string{"web"},
		},
		{
			name:              "In operation",
			input:             "app in [web,api]",
			expectedOperation: LabelOpIn,
			expectedKey:       "app",
			expectedValues:    []string{"web", "api"},
		},
		{
			name:              "Not in operation",
			input:             "env not in [test,debug]",
			expectedOperation: LabelOpNotIn,
			expectedKey:       "env",
			expectedValues:    []string{"test", "debug"},
		},
		{
			name:              "Has operation",
			input:             "has monitoring",
			expectedOperation: LabelOpExists,
			expectedKey:       "monitoring",
			expectedValues:    nil,
		},
		{
			name:              "Not has operation",
			input:             "not has debug",
			expectedOperation: LabelOpNotExists,
			expectedKey:       "debug",
			expectedValues:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := parser.parseSingleLabelSelector(tt.input)
			require.NoError(t, err, "Unexpected error parsing single label selector")

			assert.Equal(t, tt.expectedOperation, selector.Operation, "Operation mismatch")
			assert.Equal(t, tt.expectedKey, selector.Key, "Key mismatch")
			assert.Equal(t, tt.expectedValues, selector.Values, "Values mismatch")
		})
	}
}

func TestParseTargetSelector(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name           string
		input          string
		expectedType   string
		expectedLabels int
		expectError    bool
	}{
		{
			name:           "Simple pods selector",
			input:          "pods(app=web)",
			expectedType:   "pods",
			expectedLabels: 1,
			expectError:    false,
		},
		{
			name:           "Multiple labels",
			input:          "pods(app=web,tier=frontend)",
			expectedType:   "pods",
			expectedLabels: 2,
			expectError:    false,
		},
		{
			name:        "Unsupported target type",
			input:       "services(my-service)",
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := parser.parseTargetSelector(tt.input)

			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expectedType, target.Type, "Target type mismatch")
				assert.Len(t, target.LabelSelectors, tt.expectedLabels, "Label selectors count mismatch")
			}
		})
	}
}
