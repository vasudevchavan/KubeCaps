package predictor

import (
	"math"
	"testing"
)

func TestEWMA(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		alpha    float64
		expected float64
		delta    float64
	}{
		{
			name:     "empty values",
			values:   []float64{},
			alpha:    0.3,
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "single value",
			values:   []float64{10.0},
			alpha:    0.3,
			expected: 10.0,
			delta:    0.001,
		},
		{
			name:     "steady values",
			values:   []float64{10.0, 10.0, 10.0, 10.0},
			alpha:    0.3,
			expected: 10.0,
			delta:    0.001,
		},
		{
			name:     "increasing values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			alpha:    0.3,
			expected: 3.5, // Approximate
			delta:    0.5,
		},
		{
			name:     "decreasing values",
			values:   []float64{5.0, 4.0, 3.0, 2.0, 1.0},
			alpha:    0.3,
			expected: 2.5, // Approximate
			delta:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EWMA(tt.values, tt.alpha)
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("EWMA() = %v, want %v (±%v)", result, tt.expected, tt.delta)
			}
		})
	}
}

func TestEWMAWithDifferentAlpha(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	
	// Higher alpha should give more weight to recent values
	highAlpha := EWMA(values, 0.9)
	lowAlpha := EWMA(values, 0.1)
	
	// With increasing values, high alpha should be closer to the last value
	if highAlpha <= lowAlpha {
		t.Errorf("Expected high alpha (%v) > low alpha (%v) for increasing series", highAlpha, lowAlpha)
	}
}

func BenchmarkEWMA(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EWMA(values, 0.3)
	}
}

// Made with Bob
