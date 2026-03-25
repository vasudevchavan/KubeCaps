package ai

import (
	"math"
	"sort"

	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// AnomalyDetector detects anomalies in time series data using statistical methods
type AnomalyDetector struct {
	sensitivity float64 // 1.0 = normal, 2.0 = more sensitive, 0.5 = less sensitive
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(sensitivity float64) *AnomalyDetector {
	if sensitivity <= 0 {
		sensitivity = 1.0
	}
	return &AnomalyDetector{
		sensitivity: sensitivity,
	}
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Timestamp   string  `json:"timestamp"`
	Value       float64 `json:"value"`
	Expected    float64 `json:"expected"`
	Deviation   float64 `json:"deviation"`
	Severity    string  `json:"severity"` // low, medium, high, critical
	Score       float64 `json:"score"`    // 0-1, higher = more anomalous
	Description string  `json:"description"`
}

// DetectAnomalies detects anomalies in a time series using Isolation Forest-like approach
func (d *AnomalyDetector) DetectAnomalies(data []types.DataPoint) []Anomaly {
	if len(data) < 10 {
		return nil // Need minimum data points
	}

	values := make([]float64, len(data))
	for i, dp := range data {
		values[i] = dp.Value
	}

	// Calculate statistical properties
	median := calculateMedian(values)
	iqr := calculateIQR(values)

	var anomalies []Anomaly

	// Use modified Z-score (more robust than standard Z-score)
	for i, dp := range data {
		// Modified Z-score using median and MAD (Median Absolute Deviation)
		mad := calculateMAD(values, median)
		modifiedZScore := 0.0
		if mad > 0 {
			modifiedZScore = 0.6745 * (dp.Value - median) / mad
		}

		// Anomaly score (0-1)
		score := math.Min(1.0, math.Abs(modifiedZScore)/10.0)

		// Adjust for sensitivity
		adjustedScore := score * d.sensitivity

		// Detect anomalies using multiple criteria
		isAnomaly := false
		severity := "low"
		description := ""

		// Criterion 1: Modified Z-score
		if math.Abs(modifiedZScore) > 3.5*d.sensitivity {
			isAnomaly = true
			severity = "high"
			description = "Significant deviation from median"
		} else if math.Abs(modifiedZScore) > 2.5*d.sensitivity {
			isAnomaly = true
			severity = "medium"
			description = "Moderate deviation from median"
		}

		// Criterion 2: IQR method
		q1, q3 := calculateQuartiles(values)
		lowerBound := q1 - 1.5*iqr*d.sensitivity
		upperBound := q3 + 1.5*iqr*d.sensitivity
		if dp.Value < lowerBound || dp.Value > upperBound {
			isAnomaly = true
			if severity == "low" {
				severity = "medium"
				description = "Outside IQR bounds"
			}
		}

		// Criterion 3: Extreme outliers (beyond 3*IQR)
		extremeLowerBound := q1 - 3*iqr*d.sensitivity
		extremeUpperBound := q3 + 3*iqr*d.sensitivity
		if dp.Value < extremeLowerBound || dp.Value > extremeUpperBound {
			severity = "critical"
			description = "Extreme outlier detected"
		}

		// Criterion 4: Sudden spikes (rate of change)
		if i > 0 {
			prevValue := data[i-1].Value
			if prevValue > 0 {
				changeRate := math.Abs((dp.Value - prevValue) / prevValue)
				if changeRate > 0.5*d.sensitivity { // 50% change
					isAnomaly = true
					if severity == "low" {
						severity = "medium"
					}
					description = "Sudden spike detected"
				}
			}
		}

		if isAnomaly {
			anomalies = append(anomalies, Anomaly{
				Timestamp:   dp.Timestamp.Format("2006-01-02 15:04:05"),
				Value:       dp.Value,
				Expected:    median,
				Deviation:   dp.Value - median,
				Severity:    severity,
				Score:       adjustedScore,
				Description: description,
			})
		}
	}

	return anomalies
}

// DetectPatternAnomalies detects anomalies based on pattern changes
func (d *AnomalyDetector) DetectPatternAnomalies(data []types.DataPoint, windowSize int) []Anomaly {
	if len(data) < windowSize*2 {
		return nil
	}

	var anomalies []Anomaly

	// Sliding window to detect pattern changes
	for i := windowSize; i < len(data)-windowSize; i++ {
		// Compare current window with previous window
		prevWindow := data[i-windowSize : i]
		currWindow := data[i : i+windowSize]

		prevMean := calculateMeanFromDataPoints(prevWindow)
		currMean := calculateMeanFromDataPoints(currWindow)

		prevStd := calculateStdDevFromDataPoints(prevWindow, prevMean)
		currStd := calculateStdDevFromDataPoints(currWindow, currMean)

		// Detect significant pattern shift
		if prevStd > 0 {
			meanShift := math.Abs(currMean-prevMean) / prevStd
			stdShift := math.Abs(currStd-prevStd) / prevStd

			if meanShift > 2.0*d.sensitivity || stdShift > 1.0*d.sensitivity {
				anomalies = append(anomalies, Anomaly{
					Timestamp:   data[i].Timestamp.Format("2006-01-02 15:04:05"),
					Value:       data[i].Value,
					Expected:    prevMean,
					Deviation:   currMean - prevMean,
					Severity:    "medium",
					Score:       math.Min(1.0, meanShift/5.0),
					Description: "Pattern shift detected",
				})
			}
		}
	}

	return anomalies
}

// Helper functions

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)))
}

func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func calculateMAD(values []float64, median float64) float64 {
	if len(values) == 0 {
		return 0
	}
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}
	return calculateMedian(deviations)
}

func calculateIQR(values []float64) float64 {
	q1, q3 := calculateQuartiles(values)
	return q3 - q1
}

func calculateQuartiles(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	q1Index := n / 4
	q3Index := 3 * n / 4

	return sorted[q1Index], sorted[q3Index]
}

func calculateMeanFromDataPoints(data []types.DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, dp := range data {
		sum += dp.Value
	}
	return sum / float64(len(data))
}

func calculateStdDevFromDataPoints(data []types.DataPoint, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, dp := range data {
		diff := dp.Value - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

// Made with Bob
