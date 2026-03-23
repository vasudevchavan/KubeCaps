package predictor

import "math"

// LinearRegressionSlope computes the slope of the best-fit line using least squares.
func LinearRegressionSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, v := range values {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

// LinearRegressionPeak predicts the peak value by projecting the linear trend
// forward by 20% of the data window.
func LinearRegressionPeak(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		if len(values) == 1 {
			return values[0]
		}
		return 0
	}

	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, v := range values {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return sumY / n
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n

	// Project forward by 20% of the window
	futureX := n + n*0.2
	projected := intercept + slope*futureX

	// Peak is the max of: current max, projected value
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	return math.Max(maxVal, projected)
}
