package evaluator

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// CorrelationAnalyzer performs cross-metric correlation analysis.
type CorrelationAnalyzer struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewCorrelationAnalyzer creates a new correlation analyzer.
func NewCorrelationAnalyzer(promClient *prometheus.Client) *CorrelationAnalyzer {
	return &CorrelationAnalyzer{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// Analyze performs cross-correlation analysis between traffic, CPU, memory, and scaling events.
func (ca *CorrelationAnalyzer) Analyze(ctx context.Context, workload types.WorkloadInfo, timeWindowHours int) ([]types.CorrelationResult, []types.Insight, error) {
	end := time.Now()
	start := end.Add(-time.Duration(timeWindowHours) * time.Hour)
	step := time.Minute * 5

	var correlations []types.CorrelationResult
	var insights []types.Insight

	// Fetch all metric series
	cpuQuery := ca.queries.CPUUsageTotal(workload.Namespace, workload.Name)
	memQuery := ca.queries.MemoryUsageTotal(workload.Namespace, workload.Name)
	rpsQuery := ca.queries.RequestRate(workload.Namespace, workload.Name)
	podQuery := ca.queries.PodReadyCount(workload.Namespace, workload.Name)

	cpuData, _ := ca.promClient.QueryRange(ctx, cpuQuery, start, end, step)
	memData, _ := ca.promClient.QueryRange(ctx, memQuery, start, end, step)
	rpsData, _ := ca.promClient.QueryRange(ctx, rpsQuery, start, end, step)
	podData, _ := ca.promClient.QueryRange(ctx, podQuery, start, end, step)

	// CPU vs RPS correlation
	if len(cpuData) > 5 && len(rpsData) > 5 {
		cpuVals := extractVals(cpuData)
		rpsVals := extractVals(rpsData)
		corr := pearsonCorrelation(cpuVals, rpsVals)

		interpretation := interpretCorrelation("CPU usage", "request rate", corr)
		correlations = append(correlations, types.CorrelationResult{
			MetricA:        "CPU Usage",
			MetricB:        "Request Rate (RPS)",
			Correlation:    corr,
			Interpretation: interpretation,
		})

		if corr > 0.7 {
			insights = append(insights, types.Insight{
				Severity:    "info",
				Category:    "scaling",
				Title:       "CPU Driven by Traffic",
				Description: fmt.Sprintf("Strong correlation (%.2f) between CPU and RPS — CPU scales with traffic", corr),
				Evidence:    fmt.Sprintf("Pearson correlation: %.2f", corr),
				Suggestion:  "HPA with CPU metric is a good scaling strategy for this workload",
			})
		}
	}

	// Memory vs RPS correlation
	if len(memData) > 5 && len(rpsData) > 5 {
		memVals := extractVals(memData)
		rpsVals := extractVals(rpsData)
		corr := pearsonCorrelation(memVals, rpsVals)

		correlations = append(correlations, types.CorrelationResult{
			MetricA:        "Memory Usage",
			MetricB:        "Request Rate (RPS)",
			Correlation:    corr,
			Interpretation: interpretCorrelation("Memory usage", "request rate", corr),
		})

		if corr > 0.7 {
			insights = append(insights, types.Insight{
				Severity:    "info",
				Category:    "scaling",
				Title:       "Memory Driven by Traffic",
				Description: fmt.Sprintf("Strong correlation (%.2f) between memory and RPS", corr),
				Evidence:    fmt.Sprintf("Pearson correlation: %.2f", corr),
				Suggestion:  "Consider memory-based scaling triggers if not already configured",
			})
		}
	}

	// CPU vs Replica count correlation
	if len(cpuData) > 5 && len(podData) > 5 {
		cpuVals := extractVals(cpuData)
		podVals := extractVals(podData)
		corr := pearsonCorrelation(cpuVals, podVals)

		correlations = append(correlations, types.CorrelationResult{
			MetricA:        "CPU Usage",
			MetricB:        "Replica Count",
			Correlation:    corr,
			Interpretation: interpretCorrelation("CPU usage", "replica count", corr),
		})

		if corr < 0.3 && corr > -0.3 {
			insights = append(insights, types.Insight{
				Severity:    "warning",
				Category:    "scaling",
				Title:       "Weak CPU-Scaling Correlation",
				Description: "Scaling events don't appear to correlate with CPU usage patterns",
				Evidence:    fmt.Sprintf("CPU vs Replicas correlation: %.2f", corr),
				Suggestion:  "Review scaling triggers — CPU may not be the right metric for this workload",
			})
		}
	}

	return correlations, insights, nil
}

// pearsonCorrelation calculates the Pearson correlation coefficient between two series.
func pearsonCorrelation(x, y []float64) float64 {
	n := math.Min(float64(len(x)), float64(len(y)))
	if n < 3 {
		return 0
	}

	count := int(n)
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0
	for i := 0; i < count; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// interpretCorrelation provides a human-readable interpretation.
func interpretCorrelation(metricA, metricB string, corr float64) string {
	switch {
	case corr > 0.8:
		return fmt.Sprintf("Very strong positive correlation: %s increases with %s", metricA, metricB)
	case corr > 0.5:
		return fmt.Sprintf("Moderate positive correlation: %s tends to increase with %s", metricA, metricB)
	case corr > 0.2:
		return fmt.Sprintf("Weak positive correlation between %s and %s", metricA, metricB)
	case corr > -0.2:
		return fmt.Sprintf("No significant correlation between %s and %s", metricA, metricB)
	case corr > -0.5:
		return fmt.Sprintf("Weak negative correlation: %s tends to decrease as %s increases", metricA, metricB)
	case corr > -0.8:
		return fmt.Sprintf("Moderate negative correlation between %s and %s", metricA, metricB)
	default:
		return fmt.Sprintf("Very strong negative correlation: %s decreases as %s increases", metricA, metricB)
	}
}

// extractVals extracts float64 values from DataPoints.
func extractVals(data []types.DataPoint) []float64 {
	vals := make([]float64, len(data))
	for i, dp := range data {
		vals[i] = dp.Value
	}
	return vals
}
