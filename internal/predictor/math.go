package predictor

import (
	"math"
	"sort"
)

// filterOutliersIQR removes outliers using the Interquartile Range (IQR) method.
func filterOutliersIQR(values []float64) []float64 {
	if len(values) < 4 {
		return values
	}
	
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	q1 := Percentile(sorted, 25)
	q3 := Percentile(sorted, 75)
	iqr := q3 - q1

	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	filtered := make([]float64, 0, len(values))
	for _, v := range values {
		if v >= lowerBound && v <= upperBound {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// HoltWinters predicts the peak for the next 'pointsToForecast' steps.
// Uses a simple Triple Exponential Smoothing (additive trend, no explicit seasonality for now as basic Holt works well for peaks).
func HoltWintersPeak(values []float64, pointsToForecast int) float64 {
	if len(values) < 2 {
		if len(values) == 1 {
			return values[0]
		}
		return 0
	}

	// Smoothing parameters
	alpha := 0.3 // level
	beta := 0.1  // trend

	level := values[0]
	trend := values[1] - values[0]

	for i := 1; i < len(values); i++ {
		val := values[i]
		lastLevel := level
		level = alpha*val + (1-alpha)*(level+trend)
		trend = beta*(level-lastLevel) + (1-beta)*trend
	}

	// Forecast future peak with damped trend (phi = 0.9)
	var peak float64
	phi := 0.90
	currentTrendTerm := trend

	for i := 1; i <= pointsToForecast; i++ {
		currentTrendTerm *= phi
		forecast := level + currentTrendTerm
		if forecast > peak {
			peak = forecast
		}
		// update level for next iteration mathematically
		level = forecast
	}
	
	// If the predicted peak is lower than our recent actual peak, fallback to current peak + slight trend
	currentPeak := Percentile(values, 100)
	if peak < currentPeak {
		return currentPeak * 1.05
	}
	return peak
}

// LogNormalOverflowProb fits a Log-Normal distribution and returns the probability that usage > request.
func LogNormalOverflowProb(values []float64, request float64) float64 {
	if len(values) == 0 || request <= 0 {
		return 0
	}

	var sumLog, sumLogSq float64
	var count int
	for _, v := range values {
		if v > 0 { // log(0) is undefined
			lv := math.Log(v)
			sumLog += lv
			sumLogSq += lv * lv
			count++
		}
	}

	if count == 0 {
		return 0
	}

	meanLog := sumLog / float64(count)
	varLog := (sumLogSq / float64(count)) - (meanLog * meanLog)

	if varLog <= 0 {
		// No variance, empirical check
		if request < math.Exp(meanLog) {
			return 1.0
		}
		return 0.0
	}

	stdLog := math.Sqrt(varLog)
	
	// P(X > request) = 1 - CDF_lognormal(request)
	// CDF_lognormal(x) = 0.5 + 0.5 * erf((ln(x) - \mu) / (\sigma * sqrt(2)))
	
	z := (math.Log(request) - meanLog) / (stdLog * math.Sqrt2)
	cdf := 0.5 + 0.5*math.Erf(z)
	
	prob := 1.0 - cdf
	if prob < 0 {
		return 0
	}
	return prob
}

// optimizeRequest finds the optimal request using Grid Search minimizing Cost + lambda*Risk.
func optimizeRequest(values []float64, minReq, maxReq float64, resourcePrice float64, lambda float64) float64 {
	if len(values) == 0 {
		return minReq
	}
	if minReq >= maxReq {
		return maxReq
	}

	bestReq := minReq
	minScore := math.MaxFloat64
	
	steps := 100
	stepSize := (maxReq - minReq) / float64(steps)
	if stepSize <= 0 {
		return maxReq
	}

	for r := minReq; r <= maxReq; r += stepSize {
		cost := r * resourcePrice
		riskProb := LogNormalOverflowProb(values, r)
		
		score := cost + lambda*riskProb
		
		if score < minScore {
			minScore = score
			bestReq = r
		} else if score > minScore * 1.5 {
			// Early stopping if score starts diverging significantly
			break
		}
	}

	return bestReq
}
