package forgeron

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// node represents a node in the Bayesian network
type node struct {
	Name             string                 `json:"name"`
	ParentNames      []string               `json:"parentNames"`
	PossibleValues   []string               `json:"possibleValues"`
	ConditionalProbs map[string]interface{} `json:"conditionalProbabilities"`
	parents          []*node
	children         []*node
}

// bayesianNetwork represents the entire network
type bayesianNetwork struct {
	NodesInSamplingOrder []*node
	NodesByName          map[string]*node
}

// newBayesianNetwork creates a new Bayesian network
func newBayesianNetwork() *bayesianNetwork {
	return &bayesianNetwork{
		NodesByName: make(map[string]*node),
	}
}

// loadNetwork loads a network definition from JSON data
func (bn *bayesianNetwork) loadNetwork(data []byte) error {
	var networkDef struct {
		Nodes []node `json:"nodes"`
	}
	if err := json.Unmarshal(data, &networkDef); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Add nodes to network
	for i := range networkDef.Nodes {
		node := &networkDef.Nodes[i]
		bn.NodesByName[node.Name] = node
		bn.NodesInSamplingOrder = append(bn.NodesInSamplingOrder, node)
	}

	// Set up parent-child relationships
	for _, node := range bn.NodesByName {
		for _, parentName := range node.ParentNames {
			if parent, exists := bn.NodesByName[parentName]; exists {
				node.parents = append(node.parents, parent)
				parent.children = append(parent.children, node)
			}
		}
	}
	return nil
}

// getProbabilitiesGivenKnownValues extracts unconditional probabilities of node values given parent values
func (n *node) getProbabilitiesGivenKnownValues(parentValues map[string]string) map[string]float64 {
	probabilities := n.ConditionalProbs
	for _, parentName := range n.ParentNames {
		parentValue := parentValues[parentName]
		if deeper, ok := probabilities["deeper"].(map[string]interface{}); ok {
			if next, exists := deeper[parentValue]; exists {
				probabilities = next.(map[string]interface{})
			} else {
				probabilities = probabilities["skip"].(map[string]interface{})
			}
		}
	}

	result := make(map[string]float64)
	for k, v := range probabilities {
		if prob, ok := v.(float64); ok {
			result[k] = prob
		}
	}
	return result
}

// sampleRandomValueFromPossibilities randomly samples from given values using probabilities
func (n *node) sampleRandomValueFromPossibilities(possibleValues []string, probabilities map[string]float64) string {
	anchor := rand.Float64()
	cumulativeProbability := 0.0
	for _, value := range possibleValues {
		cumulativeProbability += probabilities[value]
		if cumulativeProbability > anchor {
			return value
		}
	}
	return possibleValues[0]
}

// sample randomly samples from the conditional distribution given parent values
func (n *node) sample(parentValues map[string]string) string {
	probabilities := n.getProbabilitiesGivenKnownValues(parentValues)
	possibleValues := make([]string, 0, len(probabilities))
	for value := range probabilities {
		possibleValues = append(possibleValues, value)
	}
	return n.sampleRandomValueFromPossibilities(possibleValues, probabilities)
}

// sampleAccordingToRestrictions samples with restrictions on possible values
func (n *node) sampleAccordingToRestrictions(
	parentValues map[string]string,
	valuePossibilities []string,
	bannedValues []string,
) (string, bool) {
	probabilities := n.getProbabilitiesGivenKnownValues(parentValues)
	validValues := make([]string, 0)

	// Create a map of banned values for quick lookup
	bannedMap := make(map[string]struct{})
	for _, v := range bannedValues {
		bannedMap[v] = struct{}{}
	}

	// Filter valid values
	for _, value := range valuePossibilities {
		if _, isBanned := bannedMap[value]; !isBanned {
			if _, hasProb := probabilities[value]; hasProb {
				validValues = append(validValues, value)
			}
		}
	}

	if len(validValues) == 0 {
		return "", false
	}

	// Create probability map for valid values
	validProbs := make(map[string]float64)
	for _, value := range validValues {
		validProbs[value] = probabilities[value]
	}

	return n.sampleRandomValueFromPossibilities(validValues, validProbs), true
}

// generateSample generates a random sample from the network
func (bn *bayesianNetwork) generateSample(inputValues map[string]string) map[string]string {
	if inputValues == nil {
		inputValues = make(map[string]string)
	}
	sample := make(map[string]string)
	for k, v := range inputValues {
		sample[k] = v
	}

	for _, node := range bn.NodesInSamplingOrder {
		if _, exists := sample[node.Name]; !exists {
			sample[node.Name] = node.sample(sample)
		}
	}
	return sample
}

// generateConsistentSampleWhenPossible generates a sample consistent with value restrictions
func (bn *bayesianNetwork) generateConsistentSampleWhenPossible(
	valuePossibilities map[string][]string,
) (map[string]string, bool) {
	return bn.recursivelyGenerateConsistentSampleWhenPossible(
		make(map[string]string),
		valuePossibilities,
		0,
	)
}

func (bn *bayesianNetwork) recursivelyGenerateConsistentSampleWhenPossible(
	sampleSoFar map[string]string,
	valuePossibilities map[string][]string,
	depth int,
) (map[string]string, bool) {
	if depth == len(bn.NodesInSamplingOrder) {
		return sampleSoFar, true
	}

	node := bn.NodesInSamplingOrder[depth]
	bannedValues := make([]string, 0)

	for {
		possibilities := valuePossibilities[node.Name]
		if possibilities == nil {
			possibilities = node.PossibleValues
		}

		sampleValue, ok := node.sampleAccordingToRestrictions(
			sampleSoFar,
			possibilities,
			bannedValues,
		)

		if !ok {
			break
		}

		sampleSoFar[node.Name] = sampleValue
		nextSample, success := bn.recursivelyGenerateConsistentSampleWhenPossible(
			sampleSoFar,
			valuePossibilities,
			depth+1,
		)

		if success {
			return nextSample, true
		}

		bannedValues = append(bannedValues, sampleValue)
		delete(sampleSoFar, node.Name)
	}

	return nil, false
}

// getProbability calculates the probability of a value given evidence
func (bn *bayesianNetwork) getProbability(nodeName string, value string, evidence map[string]string) float64 {
	node, exists := bn.NodesByName[nodeName]
	if !exists {
		return 0.0
	}

	// If no parents, return marginal probability
	if len(node.parents) == 0 {
		if prob, ok := node.ConditionalProbs[value].(float64); ok {
			return prob
		}
		return 0.0
	}

	// Get parent values from evidence
	parentValues := make([]string, len(node.parents))
	for i, parent := range node.parents {
		if val, exists := evidence[parent.Name]; exists {
			parentValues[i] = val
		} else {
			return 0.0 // Missing evidence for parent
		}
	}

	// Navigate through conditional probability structure
	current := node.ConditionalProbs
	for _, parentValue := range parentValues {
		if deeper, ok := current["deeper"].(map[string]interface{}); ok {
			if next, exists := deeper[parentValue]; exists {
				current = next.(map[string]interface{})
			} else {
				return 0.0 // Invalid parent value
			}
		} else {
			return 0.0 // Invalid structure
		}
	}

	// Get the final probability
	if prob, ok := current[value].(float64); ok {
		return prob
	}

	return 0.0
}

// infer calculates the probability distribution for a node given evidence
func (bn *bayesianNetwork) infer(nodeName string, evidence map[string]string) map[string]float64 {
	node, exists := bn.NodesByName[nodeName]
	if !exists {
		return nil
	}

	distribution := make(map[string]float64)
	for _, value := range node.PossibleValues {
		distribution[value] = bn.getProbability(nodeName, value, evidence)
	}

	return distribution
}
