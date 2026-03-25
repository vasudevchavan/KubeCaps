package evaluator

import (
	"testing"

	"github.com/vasudevchavan/kubecaps/pkg/types"
)

func TestCalculateOverallScore(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name     string
		scores   []types.ComponentScore
		expected float64
		delta    float64
	}{
		{
			name:     "empty scores",
			scores:   []types.ComponentScore{},
			expected: 0,
			delta:    0.001,
		},
		{
			name: "single component",
			scores: []types.ComponentScore{
				{Component: "HPA", Score: 8.0, MaxScore: 10},
			},
			expected: 8.0,
			delta:    0.1,
		},
		{
			name: "all components perfect",
			scores: []types.ComponentScore{
				{Component: "HPA", Score: 10.0, MaxScore: 10},
				{Component: "VPA", Score: 10.0, MaxScore: 10},
				{Component: "KEDA", Score: 10.0, MaxScore: 10},
			},
			expected: 10.0,
			delta:    0.5,
		},
		{
			name: "mixed scores",
			scores: []types.ComponentScore{
				{Component: "HPA", Score: 8.0, MaxScore: 10},
				{Component: "VPA", Score: 6.0, MaxScore: 10},
				{Component: "KEDA", Score: 7.0, MaxScore: 10},
			},
			expected: 7.0,
			delta:    0.5,
		},
		{
			name: "one component not configured",
			scores: []types.ComponentScore{
				{Component: "HPA", Score: 8.0, MaxScore: 10},
				{Component: "VPA", Score: 0, MaxScore: 10}, // Not configured
				{Component: "KEDA", Score: 7.0, MaxScore: 10},
			},
			expected: 7.5,
			delta:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.CalculateOverallScore(tt.scores)
			if result < tt.expected-tt.delta || result > tt.expected+tt.delta {
				t.Errorf("CalculateOverallScore() = %v, want %v (±%v)", result, tt.expected, tt.delta)
			}
		})
	}
}

func TestScoreToGrade(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{10.0, "A+"},
		{9.5, "A+"},
		{9.0, "A+"},
		{8.5, "A"},
		{8.0, "A"},
		{7.5, "B"},
		{7.0, "B"},
		{6.5, "C"},
		{6.0, "C"},
		{5.5, "D"},
		{5.0, "D"},
		{4.0, "F"},
		{0.0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ScoreToGrade(tt.score)
			if result != tt.expected {
				t.Errorf("ScoreToGrade(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestScoreToColor(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{10.0, "green"},
		{9.0, "green"},
		{8.0, "green"},
		{7.5, "yellow"},
		{7.0, "yellow"},
		{6.0, "yellow"},
		{5.5, "red"},
		{5.0, "red"},
		{0.0, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ScoreToColor(tt.score)
			if result != tt.expected {
				t.Errorf("ScoreToColor(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestComponentWeights(t *testing.T) {
	// Verify that component weights sum to 1.0
	total := ComponentWeights["HPA"] + ComponentWeights["VPA"] + ComponentWeights["KEDA"]
	expected := 1.0
	delta := 0.001

	if total < expected-delta || total > expected+delta {
		t.Errorf("Component weights sum = %v, want %v (±%v)", total, expected, delta)
	}
}

func TestMultiComponentBonus(t *testing.T) {
	scorer := NewScorer()

	// Single component
	singleScore := scorer.CalculateOverallScore([]types.ComponentScore{
		{Component: "HPA", Score: 8.0, MaxScore: 10},
	})

	// Multiple components with same average
	multiScore := scorer.CalculateOverallScore([]types.ComponentScore{
		{Component: "HPA", Score: 8.0, MaxScore: 10},
		{Component: "VPA", Score: 8.0, MaxScore: 10},
	})

	// Multi-component should have bonus (5%)
	if multiScore <= singleScore {
		t.Errorf("Expected multi-component bonus: multi=%v should be > single=%v", multiScore, singleScore)
	}
}

// Made with Bob
