package evaluator

import (
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// Scorer calculates per-component and overall optimization scores.
type Scorer struct{}

// NewScorer creates a new scorer.
func NewScorer() *Scorer {
	return &Scorer{}
}

// ComponentWeights defines the weight of each component in the overall score.
var ComponentWeights = map[string]float64{
	"HPA":  0.35,
	"VPA":  0.35,
	"KEDA": 0.30,
}

// CalculateOverallScore computes a weighted overall score from component scores.
func (s *Scorer) CalculateOverallScore(scores []types.ComponentScore) float64 {
	if len(scores) == 0 {
		return 0
	}

	totalWeight := 0.0
	weightedSum := 0.0
	activeComponents := 0

	for _, cs := range scores {
		if cs.Score == 0 {
			// Component not configured, skip
			continue
		}
		weight, ok := ComponentWeights[cs.Component]
		if !ok {
			weight = 0.33
		}
		weightedSum += cs.Score * weight
		totalWeight += weight
		activeComponents++
	}

	if totalWeight == 0 {
		return 0
	}

	// Normalize to 10-point scale
	overall := (weightedSum / totalWeight)

	// Bonus for having multiple scaling systems configured
	if activeComponents >= 2 {
		overall = overall * 1.05 // 5% bonus
		if overall > 10 {
			overall = 10
		}
	}

	return overall
}

// ScoreToGrade converts a numeric score to a letter grade.
func ScoreToGrade(score float64) string {
	switch {
	case score >= 9:
		return "A+"
	case score >= 8:
		return "A"
	case score >= 7:
		return "B"
	case score >= 6:
		return "C"
	case score >= 5:
		return "D"
	default:
		return "F"
	}
}

// ScoreToColor returns a color indicator for the score.
func ScoreToColor(score float64) string {
	switch {
	case score >= 8:
		return "green"
	case score >= 6:
		return "yellow"
	default:
		return "red"
	}
}
