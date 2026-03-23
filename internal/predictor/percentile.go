package predictor

import (
	"math"
	"sort"
)

// Percentile computes the p-th percentile of a data set using linear interpolation.
// p should be between 0 and 100.
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation between closest ranks
	rank := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return sorted[lower]
	}

	fraction := rank - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

// PercentileStats computes P50, P95, and P99 of a data set.
func PercentileStats(values []float64) (p50, p95, p99 float64) {
	return Percentile(values, 50), Percentile(values, 95), Percentile(values, 99)
}

// SafeLimit computes a safe resource limit based on percentile analysis.
// It uses P99 with an additional headroom factor.
func SafeLimit(values []float64, headroomFactor float64) float64 {
	if headroomFactor <= 0 {
		headroomFactor = 1.2
	}
	return Percentile(values, 99) * headroomFactor
}
