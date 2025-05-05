package forgeron

import (
	"testing"
)

// createTestNetwork creates a simple test headerGeneratorNetwork
func createTestNetwork() *bayesianNetwork {
	network := newBayesianNetwork()

	// Create a simple headerGeneratorNetwork with two nodes
	// node A has no parents
	// node B depends on node A
	nodeA := &node{
		Name:           "A",
		PossibleValues: []string{"a1", "a2"},
		ConditionalProbs: map[string]interface{}{
			"a1": 0.6,
			"a2": 0.4,
		},
	}

	nodeB := &node{
		Name:           "B",
		ParentNames:    []string{"A"},
		PossibleValues: []string{"b1", "b2"},
		ConditionalProbs: map[string]interface{}{
			"deeper": map[string]interface{}{
				"a1": map[string]interface{}{
					"b1": 0.7,
					"b2": 0.3,
				},
				"a2": map[string]interface{}{
					"b1": 0.2,
					"b2": 0.8,
				},
			},
		},
	}

	network.NodesByName = map[string]*node{
		"A": nodeA,
		"B": nodeB,
	}
	network.NodesInSamplingOrder = []*node{nodeA, nodeB}

	// Set up parent-child relationships
	nodeB.parents = []*node{nodeA}
	nodeA.children = []*node{nodeB}

	return network
}

func TestGetProbability(t *testing.T) {
	network := createTestNetwork()

	tests := []struct {
		name     string
		nodeName string
		value    string
		evidence map[string]string
		want     float64
	}{
		{
			name:     "node A value a1",
			nodeName: "A",
			value:    "a1",
			evidence: map[string]string{},
			want:     0.6,
		},
		{
			name:     "node B value b1 given A=a1",
			nodeName: "B",
			value:    "b1",
			evidence: map[string]string{"A": "a1"},
			want:     0.7,
		},
		{
			name:     "node B value b2 given A=a2",
			nodeName: "B",
			value:    "b2",
			evidence: map[string]string{"A": "a2"},
			want:     0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := network.getProbability(tt.nodeName, tt.value, tt.evidence)
			if got != tt.want {
				t.Errorf("getProbability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSample(t *testing.T) {
	network := createTestNetwork()

	// Test sampling from node A
	parentValues := map[string]string{}
	value := network.NodesByName["A"].sample(parentValues)
	if value != "a1" && value != "a2" {
		t.Errorf("sample() returned invalid value: %v", value)
	}

	// Test sampling from node B given A=a1
	parentValues = map[string]string{"A": "a1"}
	value = network.NodesByName["B"].sample(parentValues)
	if value != "b1" && value != "b2" {
		t.Errorf("sample() returned invalid value: %v", value)
	}
}

func TestGenerateSample(t *testing.T) {
	network := createTestNetwork()

	sample := network.generateSample(nil)

	// Check that all nodes are present
	if _, exists := sample["A"]; !exists {
		t.Error("sample missing node A")
	}
	if _, exists := sample["B"]; !exists {
		t.Error("sample missing node B")
	}

	// Check that values are valid
	if sample["A"] != "a1" && sample["A"] != "a2" {
		t.Errorf("Invalid value for node A: %v", sample["A"])
	}
	if sample["B"] != "b1" && sample["B"] != "b2" {
		t.Errorf("Invalid value for node B: %v", sample["B"])
	}
}

func TestGenerateConsistentSampleWhenPossible(t *testing.T) {
	network := createTestNetwork()

	// Test with valid restrictions
	restrictions := map[string][]string{
		"A": {"a1"},
		"B": {"b1", "b2"},
	}

	sample, success := network.generateConsistentSampleWhenPossible(restrictions)
	if !success {
		t.Error("Failed to generate consistent sample")
	}

	if sample["A"] != "a1" {
		t.Errorf("sample does not respect restriction: got %v, want a1", sample["A"])
	}

	// Test with impossible restrictions
	restrictions = map[string][]string{
		"A": {"a1"},
		"B": {"invalid_value"},
	}

	_, success = network.generateConsistentSampleWhenPossible(restrictions)
	if success {
		t.Error("Should fail with impossible restrictions")
	}
}
