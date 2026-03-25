package predictor

import (
	"testing"
)

func TestPercentile(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		percentile float64
		expected   float64
	}{
		{
			name:       "empty values",
			values:     []float64{},
			percentile: 50,
			expected:   0,
		},
		{
			name:       "single value",
			values:     []float64{42.0},
			percentile: 50,
			expected:   42.0,
		},
		{
			name:       "p50 of sorted values",
			values:     []float64{1, 2, 3, 4, 5},
			percentile: 50,
			expected:   3.0,
		},
		{
			name:       "p95 of sorted values",
			values:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 95,
			expected:   9.55, // Interpolated value
		},
		{
			name:       "p99 of sorted values",
			values:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 99,
			expected:   9.91, // Interpolated value
		},
		{
			name:       "p0 should return minimum",
			values:     []float64{5, 3, 8, 1, 9},
			percentile: 0,
			expected:   1.0,
		},
		{
			name:       "p100 should return maximum",
			values:     []float64{5, 3, 8, 1, 9},
			percentile: 100,
			expected:   9.0,
		},
		{
			name:       "unsorted values",
			values:     []float64{9, 1, 5, 3, 7},
			percentile: 50,
			expected:   5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Percentile(tt.values, tt.percentile)
			// Use a small delta for floating point comparison
			delta := 0.01
			if result < tt.expected-delta || result > tt.expected+delta {
				t.Errorf("Percentile() = %v, want %v (±%v)", result, tt.expected, delta)
			}
		})
	}
}

func TestPercentileWithDuplicates(t *testing.T) {
	values := []float64{1, 1, 1, 2, 2, 3, 3, 3, 3}
	
	p50 := Percentile(values, 50)
	if p50 != 2.0 && p50 != 3.0 {
		t.Errorf("P50 with duplicates = %v, expected around 2-3", p50)
	}
}

func BenchmarkPercentile(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Percentile(values, 95)
	}
}

// Made with Bob
