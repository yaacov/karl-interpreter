package karl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToAffinity(t *testing.T) {
	converter := &Converter{}

	tests := []struct {
		name                 string
		rule                 KARLRule
		expectPodAffinity    bool
		expectAntiAffinity   bool
		expectHardConstraint bool
		expectSoftConstraint bool
		expectedWeight       int32
	}{
		{
			name: "REQUIRE rule creates hard affinity",
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
			expectPodAffinity:    true,
			expectAntiAffinity:   false,
			expectHardConstraint: true,
			expectSoftConstraint: false,
		},
		{
			name: "PREFER rule creates soft affinity",
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
			expectPodAffinity:    true,
			expectAntiAffinity:   false,
			expectHardConstraint: false,
			expectSoftConstraint: true,
			expectedWeight:       80,
		},
		{
			name: "AVOID rule creates hard anti-affinity",
			rule: KARLRule{
				RuleType: RuleTypeAvoid,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"test"}},
					},
				},
				TopologyKey: TopologyNode,
			},
			expectPodAffinity:    false,
			expectAntiAffinity:   true,
			expectHardConstraint: true,
			expectSoftConstraint: false,
		},
		{
			name: "REPEL rule creates soft anti-affinity",
			rule: KARLRule{
				RuleType: RuleTypeRepel,
				TargetSelector: TargetSelector{
					Type: "pods",
					LabelSelectors: []LabelSelector{
						{Key: "app", Operation: LabelOpEquals, Values: []string{"batch"}},
					},
				},
				TopologyKey: TopologyZone,
				Weight:      90,
			},
			expectPodAffinity:    false,
			expectAntiAffinity:   true,
			expectHardConstraint: false,
			expectSoftConstraint: true,
			expectedWeight:       90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			affinity, err := converter.ToAffinity(tt.rule)
			require.NoError(t, err, "ToAffinity() should not return error")

			// Check pod affinity vs anti-affinity
			if tt.expectPodAffinity {
				assert.NotNil(t, affinity.PodAffinity, "Expected PodAffinity to be set")
			}

			if tt.expectAntiAffinity {
				assert.NotNil(t, affinity.PodAntiAffinity, "Expected PodAntiAffinity to be set")
			}

			// Check hard vs soft constraints
			if tt.expectHardConstraint {
				if tt.expectPodAffinity {
					assert.NotEmpty(t, affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "Expected hard affinity constraint")
				} else if tt.expectAntiAffinity {
					assert.NotEmpty(t, affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, "Expected hard anti-affinity constraint")
				}
			}

			if tt.expectSoftConstraint {
				if tt.expectPodAffinity {
					require.NotEmpty(t, affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, "Expected soft affinity constraint")
					weight := affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight
					assert.Equal(t, tt.expectedWeight, weight, "Weight mismatch")
				} else if tt.expectAntiAffinity {
					require.NotEmpty(t, affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, "Expected soft anti-affinity constraint")
					weight := affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight
					assert.Equal(t, tt.expectedWeight, weight, "Weight mismatch")
				}
			}
		})
	}
}

func TestGetTopologyKey(t *testing.T) {
	converter := &Converter{}

	tests := []struct {
		topologyKey TopologyKey
		expected    string
	}{
		{TopologyNode, "kubernetes.io/hostname"},
		{TopologyZone, "topology.kubernetes.io/zone"},
		{TopologyRegion, "topology.kubernetes.io/region"},
		{TopologyRack, "topology.kubernetes.io/rack"},
	}

	for _, tt := range tests {
		t.Run(string(tt.topologyKey), func(t *testing.T) {
			rule := KARLRule{TopologyKey: tt.topologyKey}
			result := converter.getTopologyKey(rule)
			if result != tt.expected {
				t.Errorf("getTopologyKey() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCreateLabelSelector(t *testing.T) {
	converter := &Converter{}

	tests := []struct {
		name                string
		target              TargetSelector
		expectMatchLabels   bool
		expectExpressions   bool
		expectedLabelsCount int
		expectedExpCount    int
	}{
		{
			name: "Simple equality selector",
			target: TargetSelector{
				Type: "pods",
				LabelSelectors: []LabelSelector{
					{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
				},
			},
			expectMatchLabels:   true,
			expectExpressions:   false,
			expectedLabelsCount: 1,
			expectedExpCount:    0,
		},
		{
			name: "In operation selector",
			target: TargetSelector{
				Type: "pods",
				LabelSelectors: []LabelSelector{
					{Key: "env", Operation: LabelOpIn, Values: []string{"prod", "staging"}},
				},
			},
			expectMatchLabels:   false,
			expectExpressions:   true,
			expectedLabelsCount: 0,
			expectedExpCount:    1,
		},
		{
			name: "Mixed selectors",
			target: TargetSelector{
				Type: "pods",
				LabelSelectors: []LabelSelector{
					{Key: "app", Operation: LabelOpEquals, Values: []string{"web"}},
					{Key: "env", Operation: LabelOpIn, Values: []string{"prod", "staging"}},
					{Key: "monitoring", Operation: LabelOpExists},
				},
			},
			expectMatchLabels:   true,
			expectExpressions:   true,
			expectedLabelsCount: 1,
			expectedExpCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labelSelector, err := converter.createLabelSelector(tt.target)
			if err != nil {
				t.Errorf("createLabelSelector() error = %v", err)
				return
			}

			if tt.expectMatchLabels {
				if labelSelector.MatchLabels == nil {
					t.Error("Expected MatchLabels to be set")
				} else if len(labelSelector.MatchLabels) != tt.expectedLabelsCount {
					t.Errorf("Expected %d match labels, got %d", tt.expectedLabelsCount, len(labelSelector.MatchLabels))
				}
			} else {
				if len(labelSelector.MatchLabels) > 0 {
					t.Error("Expected MatchLabels to be empty")
				}
			}

			if tt.expectExpressions {
				if len(labelSelector.MatchExpressions) != tt.expectedExpCount {
					t.Errorf("Expected %d match expressions, got %d", tt.expectedExpCount, len(labelSelector.MatchExpressions))
				}
			} else {
				if len(labelSelector.MatchExpressions) > 0 {
					t.Error("Expected MatchExpressions to be empty")
				}
			}
		})
	}
}

func TestLabelOperationToKubernetes(t *testing.T) {
	tests := []struct {
		operation LabelOperation
		expected  metav1.LabelSelectorOperator
	}{
		{LabelOpIn, metav1.LabelSelectorOpIn},
		{LabelOpNotIn, metav1.LabelSelectorOpNotIn},
		{LabelOpExists, metav1.LabelSelectorOpExists},
		{LabelOpNotExists, metav1.LabelSelectorOpDoesNotExist},
	}

	for _, tt := range tests {
		t.Run(string(tt.operation), func(t *testing.T) {
			converter := &Converter{}
			target := TargetSelector{
				Type: "pods",
				LabelSelectors: []LabelSelector{
					{Key: "test", Operation: tt.operation, Values: []string{"value"}},
				},
			}

			labelSelector, err := converter.createLabelSelector(target)
			if err != nil {
				t.Errorf("createLabelSelector() error = %v", err)
				return
			}

			if len(labelSelector.MatchExpressions) == 0 {
				t.Error("Expected at least one match expression")
				return
			}

			if labelSelector.MatchExpressions[0].Operator != tt.expected {
				t.Errorf("Expected operator %v, got %v", tt.expected, labelSelector.MatchExpressions[0].Operator)
			}
		})
	}
}
