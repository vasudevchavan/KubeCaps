package evaluator

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// KEDAEvaluator evaluates KEDA configurations against real metrics.
type KEDAEvaluator struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewKEDAEvaluator creates a new KEDA evaluator.
func NewKEDAEvaluator(promClient *prometheus.Client) *KEDAEvaluator {
	return &KEDAEvaluator{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// Evaluate evaluates a KEDA configuration and produces insights and a score.
func (e *KEDAEvaluator) Evaluate(ctx context.Context, workload types.WorkloadInfo, keda *types.KEDAConfig, timeWindowHours int) (types.ComponentScore, []types.Insight, []types.RiskFlag, error) {
	if !keda.Found {
		return types.ComponentScore{Component: "KEDA", Score: 0, MaxScore: 10, Details: "No KEDA ScaledObject configured"},
			nil, nil, nil
	}

	score := 10.0
	var insights []types.Insight
	var risks []types.RiskFlag

	end := time.Now()
	start := end.Add(-time.Duration(timeWindowHours) * time.Hour)
	step := time.Minute * 5

	// 1. Evaluate trigger accuracy
	for _, trigger := range keda.Triggers {
		switch trigger.Type {
		case "kafka":
			consumerGroup := trigger.Metadata["consumerGroup"]
			topic := trigger.Metadata["topic"]
			if consumerGroup != "" && topic != "" {
				lagQuery := e.queries.KafkaConsumerLag(consumerGroup, topic)
				lagData, _ := e.promClient.QueryRange(ctx, lagQuery, start, end, step)

				if len(lagData) > 0 {
					avgLag := 0.0
					maxLag := 0.0
					for _, dp := range lagData {
						avgLag += dp.Value
						if dp.Value > maxLag {
							maxLag = dp.Value
						}
					}
					avgLag /= float64(len(lagData))

					if avgLag > 1000 {
						score -= 2
						insights = append(insights, types.Insight{
							Severity:    "warning",
							Category:    "scaling",
							Title:       "High Kafka Consumer Lag",
							Description: fmt.Sprintf("Average consumer lag is %.0f for group '%s' on topic '%s'", avgLag, consumerGroup, topic),
							Evidence:    fmt.Sprintf("Avg lag: %.0f, Max lag: %.0f", avgLag, maxLag),
							Suggestion:  "Consider lowering the lag threshold or increasing maxReplicaCount",
						})
					}
				}
			}

		case "prometheus":
			// Check if the Prometheus trigger query returns reasonable data
			query := trigger.Metadata["query"]
			if query != "" {
				data, _ := e.promClient.QueryRange(ctx, query, start, end, step)
				if len(data) == 0 {
					score -= 1
					insights = append(insights, types.Insight{
						Severity:    "warning",
						Category:    "scaling",
						Title:       "KEDA Prometheus Trigger Returns No Data",
						Description: "The Prometheus query used by KEDA trigger returns no data",
						Evidence:    fmt.Sprintf("Query: %s", query),
						Suggestion:  "Verify the PromQL query is correct and the metric exists",
					})
				}
			}
		}
	}

	// 2. Evaluate polling interval
	if keda.PollingInterval > 60 {
		score -= 0.5
		insights = append(insights, types.Insight{
			Severity:    "info",
			Category:    "scaling",
			Title:       "Long Polling Interval",
			Description: fmt.Sprintf("KEDA polling interval is %ds — scaling reactions may be slow", keda.PollingInterval),
			Evidence:    fmt.Sprintf("pollingInterval: %ds", keda.PollingInterval),
			Suggestion:  "Consider reducing pollingInterval to 15-30s for faster scaling",
		})
	}

	// 3. Evaluate cooldown period
	if keda.CooldownPeriod > 600 {
		score -= 0.5
		insights = append(insights, types.Insight{
			Severity:    "info",
			Category:    "scaling",
			Title:       "Long Cooldown Period",
			Description: fmt.Sprintf("KEDA cooldown period is %ds — scale-down may be too slow", keda.CooldownPeriod),
			Evidence:    fmt.Sprintf("cooldownPeriod: %ds", keda.CooldownPeriod),
			Suggestion:  "Consider reducing cooldownPeriod if your workload has predictable traffic patterns",
		})
	} else if keda.CooldownPeriod < 60 {
		score -= 1
		insights = append(insights, types.Insight{
			Severity:    "warning",
			Category:    "scaling",
			Title:       "Very Short Cooldown Period",
			Description: fmt.Sprintf("KEDA cooldown period is only %ds — may cause scaling oscillation", keda.CooldownPeriod),
			Evidence:    fmt.Sprintf("cooldownPeriod: %ds", keda.CooldownPeriod),
			Suggestion:  "Consider increasing cooldownPeriod to 120-300s to avoid thrashing",
		})
	}

	// 4. Check scaling lag (compare pod readiness time)
	podReadyQuery := e.queries.PodReadyCount(workload.Namespace, workload.Name)
	podReadyData, _ := e.promClient.QueryRange(ctx, podReadyQuery, start, end, step)
	if len(podReadyData) > 10 {
		scaleEvents := 0
		for i := 1; i < len(podReadyData); i++ {
			if podReadyData[i].Value > podReadyData[i-1].Value {
				scaleEvents++
			}
		}
		if scaleEvents > 20 {
			score -= 1
			insights = append(insights, types.Insight{
				Severity:    "info",
				Category:    "scaling",
				Title:       "Frequent Scale-Up Events",
				Description: fmt.Sprintf("%d scale-up events detected in the analysis window", scaleEvents),
				Evidence:    fmt.Sprintf("Scale-up events: %d", scaleEvents),
				Suggestion:  "Review if triggers are too sensitive; consider adjusting thresholds",
			})
		}
	}

	// 5. Check replica bounds
	if keda.MaxReplicas <= keda.MinReplicas {
		score -= 2
		insights = append(insights, types.Insight{
			Severity:    "critical",
			Category:    "scaling",
			Title:       "Invalid Replica Bounds",
			Description: fmt.Sprintf("maxReplicaCount (%d) <= minReplicaCount (%d)", keda.MaxReplicas, keda.MinReplicas),
			Evidence:    fmt.Sprintf("min=%d, max=%d", keda.MinReplicas, keda.MaxReplicas),
			Suggestion:  "Set maxReplicaCount higher than minReplicaCount",
		})
	}

	score = math.Max(score, 1)

	return types.ComponentScore{
		Component: "KEDA",
		Score:     score,
		MaxScore:  10,
		Details:   fmt.Sprintf("Evaluated %d triggers, %d insights generated", len(keda.Triggers), len(insights)),
	}, insights, risks, nil
}
