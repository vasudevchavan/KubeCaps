package predictor

// EWMA computes the Exponentially Weighted Moving Average of a time series.
// Alpha is the smoothing factor (0 < alpha <= 1). Higher alpha gives more weight
// to recent observations.
func EWMA(values []float64, alpha float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3 // default
	}

	ewma := values[0]
	for _, v := range values[1:] {
		ewma = alpha*v + (1-alpha)*ewma
	}
	return ewma
}

// EWMATimeSeries computes the EWMA for each point, returning the smoothed series.
func EWMATimeSeries(values []float64, alpha float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3
	}

	result := make([]float64, len(values))
	result[0] = values[0]
	for i := 1; i < len(values); i++ {
		result[i] = alpha*values[i] + (1-alpha)*result[i-1]
	}
	return result
}
