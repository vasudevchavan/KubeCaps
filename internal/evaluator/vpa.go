package evaluator

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// VPAEvaluator evaluates VPA configurations against real metrics.
type VPAEvaluator struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewVPAEvaluator creates a new VPA evaluator.
func NewVPAEvaluator(promClient *prometheus.Client) *VPAEvaluator {
	return &VPAEvaluator{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// Evaluate evaluates a VPA configuration and produces insights and a score.
func (e *VPAEvaluator) Evaluate(ctx context.Context, workload types.WorkloadInfo, vpa *types.VPAConfig, timeWindowHours int) (types.ComponentScore, []types.Insight, []types.RiskFlag, error) {
	if !vpa.Found {
		return types.ComponentScore{Component: "VPA", Score: 0, MaxScore: 10, Details: "No VPA configured"},
			nil, nil, nil
	}

	score := 10.0
	var insights []types.Insight
	var risks []types.RiskFlag

	end := time.Now()
	start := end.Add(-time.Duration(timeWindowHours) * time.Hour)
	step := time.Minute * 5

	// 1. Check request vs actual usage gap
	cpuUsageQuery := e.queries.CPUUsageTotal(workload.Namespace, workload.Name)
	cpuUsageData, _ := e.promClient.QueryRange(ctx, cpuUsageQuery, start, end, step)

	memUsageQuery := e.queries.MemoryUsageTotal(workload.Namespace, workload.Name)
	memUsageData, _ := e.promClient.QueryRange(ctx, memUsageQuery, start, end, step)

	cpuReq, _ := e.promClient.QueryInstant(ctx, e.queries.CPURequest(workload.Namespace, workload.Name))
	memReq, _ := e.promClient.QueryInstant(ctx, e.queries.MemoryRequest(workload.Namespace, workload.Name))

	// Analyze CPU request vs usage gap
	if len(cpuUsageData) > 0 && cpuReq > 0 {
		avgCPU := 0.0
		maxCPU := 0.0
		for _, dp := range cpuUsageData {
			avgCPU += dp.Value
			if dp.Value > maxCPU {
				maxCPU = dp.Value
			}
		}
		avgCPU /= float64(len(cpuUsageData))

		overProvisionPct := ((cpuReq - avgCPU) / cpuReq) * 100
		if overProvisionPct > 50 {
			score -= 2
			insights = append(insights, types.Insight{
				Severity:    "warning",
				Category:    "resources",
				Title:       "CPU Over-Provisioned",
				Description: fmt.Sprintf("CPU request (%.3f cores) is %.0f%% higher than average usage (%.3f cores)", cpuReq, overProvisionPct, avgCPU),
				Evidence:    fmt.Sprintf("Request: %.3f, Avg Usage: %.3f, Max: %.3f", cpuReq, avgCPU, maxCPU),
				Suggestion:  "VPA should adjust CPU requests downward to save resources",
			})
			risks = append(risks, types.RiskFlag{
				Level:       "low",
				Type:        "over_provisioned",
				Description: fmt.Sprintf("CPU over-provisioned by %.0f%%", overProvisionPct),
				Impact:      "Wasted cluster resources and higher cost",
			})
		}

		underProvisionPct := ((avgCPU - cpuReq) / cpuReq) * 100
		if underProvisionPct > 20 {
			score -= 2
			insights = append(insights, types.Insight{
				Severity:    "critical",
				Category:    "resources",
				Title:       "CPU Under-Provisioned",
				Description: fmt.Sprintf("Average CPU usage (%.3f cores) exceeds request (%.3f cores) by %.0f%%", avgCPU, cpuReq, underProvisionPct),
				Evidence:    fmt.Sprintf("Request: %.3f, Avg Usage: %.3f, Max: %.3f", cpuReq, avgCPU, maxCPU),
				Suggestion:  "VPA should increase CPU requests to prevent throttling",
			})
		}
	}

	// Analyze Memory request vs usage gap
	if len(memUsageData) > 0 && memReq > 0 {
		avgMem := 0.0
		maxMem := 0.0
		for _, dp := range memUsageData {
			avgMem += dp.Value
			if dp.Value > maxMem {
				maxMem = dp.Value
			}
		}
		avgMem /= float64(len(memUsageData))

		overProvisionPct := ((memReq - avgMem) / memReq) * 100
		if overProvisionPct > 50 {
			score -= 1.5
			insights = append(insights, types.Insight{
				Severity:    "warning",
				Category:    "resources",
				Title:       "Memory Over-Provisioned",
				Description: fmt.Sprintf("Memory request (%.0f Mi) is %.0f%% higher than average usage (%.0f Mi)", memReq/1024/1024, overProvisionPct, avgMem/1024/1024),
				Evidence:    fmt.Sprintf("Request: %.0f Mi, Avg: %.0f Mi, Max: %.0f Mi", memReq/1024/1024, avgMem/1024/1024, maxMem/1024/1024),
				Suggestion:  "VPA should adjust memory requests downward",
			})
		}
	}

	// 2. Check OOM risk
	memLim, _ := e.promClient.QueryInstant(ctx, e.queries.MemoryLimit(workload.Namespace, workload.Name))
	if len(memUsageData) > 0 && memLim > 0 {
		maxMem := 0.0
		for _, dp := range memUsageData {
			if dp.Value > maxMem {
				maxMem = dp.Value
			}
		}
		memUsagePct := (maxMem / memLim) * 100
		if memUsagePct > 90 {
			score -= 2
			risks = append(risks, types.RiskFlag{
				Level:       "high",
				Type:        "oom",
				Description: fmt.Sprintf("Peak memory usage (%.0f Mi) is %.0f%% of limit (%.0f Mi)", maxMem/1024/1024, memUsagePct, memLim/1024/1024),
				Impact:      "High risk of OOM kill events",
			})
			insights = append(insights, types.Insight{
				Severity:    "critical",
				Category:    "reliability",
				Title:       "OOM Risk Detected",
				Description: fmt.Sprintf("Peak memory reaches %.0f%% of limit", memUsagePct),
				Evidence:    fmt.Sprintf("Peak: %.0f Mi, Limit: %.0f Mi", maxMem/1024/1024, memLim/1024/1024),
				Suggestion:  "Increase memory limits or optimize application memory usage",
			})
		}
	}

	// Check OOM events
	oomQuery := e.queries.OOMEvents(workload.Namespace, workload.Name)
	oomCount, _ := e.promClient.QueryInstant(ctx, oomQuery)
	if oomCount > 0 {
		score -= 3
		risks = append(risks, types.RiskFlag{
			Level:       "high",
			Type:        "oom",
			Description: fmt.Sprintf("%.0f OOM kill events detected", oomCount),
			Impact:      "Application restarts causing service disruption",
		})
	}

	// 3. Check CPU throttling
	throttleQuery := e.queries.CPUThrottling(workload.Namespace, workload.Name)
	throttlePct, _ := e.promClient.QueryInstant(ctx, throttleQuery)
	if throttlePct > 25 {
		score -= 2
		risks = append(risks, types.RiskFlag{
			Level:       "medium",
			Type:        "throttling",
			Description: fmt.Sprintf("CPU throttling at %.0f%%", throttlePct),
			Impact:      "Increased request latency due to CPU throttling",
		})
		insights = append(insights, types.Insight{
			Severity:    "warning",
			Category:    "resources",
			Title:       "CPU Throttling Detected",
			Description: fmt.Sprintf("CPU throttling rate is %.0f%%", throttlePct),
			Evidence:    fmt.Sprintf("Throttle percentage: %.0f%%", throttlePct),
			Suggestion:  "Increase CPU limits or optimize CPU usage",
		})
	}

	// 4. Check VPA update mode
	if vpa.UpdateMode == "Off" {
		score -= 1
		insights = append(insights, types.Insight{
			Severity:    "info",
			Category:    "scaling",
			Title:       "VPA in Recommendation-Only Mode",
			Description: "VPA updateMode is 'Off' — recommendations are generated but not applied",
			Evidence:    "updateMode: Off",
			Suggestion:  "Consider setting updateMode to 'Auto' or 'Initial' for automatic adjustments",
		})
	}

	score = math.Max(score, 1)

	return types.ComponentScore{
		Component: "VPA",
		Score:     score,
		MaxScore:  10,
		Details:   fmt.Sprintf("Analyzed resource gaps, OOM risk, throttling; %d insights generated", len(insights)),
	}, insights, risks, nil
}
