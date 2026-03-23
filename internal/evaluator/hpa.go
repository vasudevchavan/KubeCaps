package evaluator

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// HPAEvaluator evaluates HPA configurations against real metrics.
type HPAEvaluator struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewHPAEvaluator creates a new HPA evaluator.
func NewHPAEvaluator(promClient *prometheus.Client) *HPAEvaluator {
	return &HPAEvaluator{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// Evaluate evaluates an HPA configuration and produces insights and a score.
func (e *HPAEvaluator) Evaluate(ctx context.Context, workload types.WorkloadInfo, hpa *types.HPAConfig, timeWindowHours int) (types.ComponentScore, []types.Insight, []types.RiskFlag, error) {
	if !hpa.Found {
		return types.ComponentScore{Component: "HPA", Score: 0, MaxScore: 10, Details: "No HPA configured"},
			nil, nil, nil
	}

	score := 10.0
	var insights []types.Insight
	var risks []types.RiskFlag

	end := time.Now()
	start := end.Add(-time.Duration(timeWindowHours) * time.Hour)
	step := time.Minute * 5

	// 1. Check scaling responsiveness
	desiredQuery := e.queries.HPAReplicaCount(workload.Namespace, hpa.Name)
	currentQuery := e.queries.HPACurrentReplicas(workload.Namespace, hpa.Name)
	desiredData, _ := e.promClient.QueryRange(ctx, desiredQuery, start, end, step)
	currentData, _ := e.promClient.QueryRange(ctx, currentQuery, start, end, step)

	if len(desiredData) > 0 && len(currentData) > 0 {
		lagCount := 0
		for i := 0; i < len(desiredData) && i < len(currentData); i++ {
			if desiredData[i].Value != currentData[i].Value {
				lagCount++
			}
		}
		lagPct := float64(lagCount) / float64(len(desiredData)) * 100
		if lagPct > 30 {
			score -= 2
			insights = append(insights, types.Insight{
				Severity:    "warning",
				Category:    "scaling",
				Title:       "High Scaling Lag",
				Description: fmt.Sprintf("%.0f%% of the time, desired replicas don't match current replicas", lagPct),
				Evidence:    fmt.Sprintf("Desired vs Current replica mismatch: %.0f%% of data points", lagPct),
				Suggestion:  "Consider adjusting HPA behavior policies or scaling thresholds",
			})
			risks = append(risks, types.RiskFlag{
				Level:       "medium",
				Type:        "scaling_lag",
				Description: "HPA is slow to converge to desired replicas",
				Impact:      "Potential performance degradation during scale-up",
			})
		}
	}

	// 2. Check target accuracy for CPU metric
	for _, metric := range hpa.Metrics {
		if metric.Name == "cpu" && metric.TargetType == "Utilization" {
			cpuUtilQuery := e.queries.CPUUtilizationPercent(workload.Namespace, workload.Name)
			cpuUtilData, _ := e.promClient.QueryRange(ctx, cpuUtilQuery, start, end, step)

			if len(cpuUtilData) > 0 {
				avgUtil := 0.0
				maxUtil := 0.0
				for _, dp := range cpuUtilData {
					avgUtil += dp.Value
					if dp.Value > maxUtil {
						maxUtil = dp.Value
					}
				}
				avgUtil /= float64(len(cpuUtilData))

				// Check if target is too far from actual utilization
				gap := math.Abs(avgUtil - metric.Target)
				if gap > 30 {
					score -= 2
					insights = append(insights, types.Insight{
						Severity:    "warning",
						Category:    "scaling",
						Title:       "CPU Target Mismatch",
						Description: fmt.Sprintf("Average CPU utilization (%.0f%%) is far from HPA target (%.0f%%)", avgUtil, metric.Target),
						Evidence:    fmt.Sprintf("Avg: %.0f%%, Max: %.0f%%, Target: %.0f%%", avgUtil, maxUtil, metric.Target),
						Suggestion:  fmt.Sprintf("Consider adjusting CPU target to ~%.0f%%", math.Min(avgUtil+15, 80)),
					})
				}

				// Check if max utilization exceeds target significantly
				if maxUtil > metric.Target*1.5 {
					score -= 1
					risks = append(risks, types.RiskFlag{
						Level:       "high",
						Type:        "scaling_lag",
						Description: fmt.Sprintf("Peak CPU utilization (%.0f%%) far exceeds HPA target (%.0f%%)", maxUtil, metric.Target),
						Impact:      "Possible request latency spikes during load bursts",
					})
				}
			}
		}
	}

	// 3. Check replica sufficiency
	if hpa.CurrentReplicas >= hpa.MaxReplicas {
		score -= 2
		insights = append(insights, types.Insight{
			Severity:    "critical",
			Category:    "scaling",
			Title:       "HPA at Maximum Replicas",
			Description: fmt.Sprintf("Current replicas (%d) equals max replicas (%d)", hpa.CurrentReplicas, hpa.MaxReplicas),
			Evidence:    fmt.Sprintf("currentReplicas=%d, maxReplicas=%d", hpa.CurrentReplicas, hpa.MaxReplicas),
			Suggestion:  "Increase maxReplicas or optimize the application to handle more load per replica",
		})
		risks = append(risks, types.RiskFlag{
			Level:       "high",
			Type:        "scaling_lag",
			Description: "HPA cannot scale further — at max replicas",
			Impact:      "Unable to handle additional load spikes",
		})
	}

	// 4. Check for oscillation (rapid scale up/down)
	if len(desiredData) > 10 {
		directionChanges := 0
		for i := 2; i < len(desiredData); i++ {
			prev := desiredData[i-1].Value - desiredData[i-2].Value
			curr := desiredData[i].Value - desiredData[i-1].Value
			if (prev > 0 && curr < 0) || (prev < 0 && curr > 0) {
				directionChanges++
			}
		}
		oscillationRate := float64(directionChanges) / float64(len(desiredData)) * 100
		if oscillationRate > 40 {
			score -= 1.5
			insights = append(insights, types.Insight{
				Severity:    "warning",
				Category:    "scaling",
				Title:       "Scaling Oscillation Detected",
				Description: fmt.Sprintf("Replica count changed direction %.0f%% of the time", oscillationRate),
				Evidence:    fmt.Sprintf("Direction changes: %d out of %d intervals", directionChanges, len(desiredData)),
				Suggestion:  "Configure HPA stabilization window or scaling behavior policies",
			})
		}
	}

	score = math.Max(score, 1)

	return types.ComponentScore{
		Component: "HPA",
		Score:     score,
		MaxScore:  10,
		Details:   fmt.Sprintf("Evaluated %d metrics, %d insights generated", len(hpa.Metrics), len(insights)),
	}, insights, risks, nil
}
