package evaluator

import (
	"context"
	"fmt"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// RiskDetector detects operational risks from metrics.
type RiskDetector struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewRiskDetector creates a new risk detector.
func NewRiskDetector(promClient *prometheus.Client) *RiskDetector {
	return &RiskDetector{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// DetectRisks runs comprehensive risk detection on a workload.
func (rd *RiskDetector) DetectRisks(ctx context.Context, workload types.WorkloadInfo, timeWindowHours int) ([]types.RiskFlag, error) {
	var risks []types.RiskFlag

	end := time.Now()
	start := end.Add(-time.Duration(timeWindowHours) * time.Hour)
	step := time.Minute * 5

	// 1. OOM Risk
	oomCount, _ := rd.promClient.QueryInstant(ctx, rd.queries.OOMEvents(workload.Namespace, workload.Name))
	if oomCount > 0 {
		risks = append(risks, types.RiskFlag{
			Level:       "high",
			Type:        "oom",
			Description: fmt.Sprintf("%.0f OOM kill events detected in the analysis window", oomCount),
			Impact:      "Application restarts causing service disruption and potential data loss",
		})
	}

	// Check memory proximity to limits
	memLim, _ := rd.promClient.QueryInstant(ctx, rd.queries.MemoryLimit(workload.Namespace, workload.Name))
	memUsageData, _ := rd.promClient.QueryRange(ctx, rd.queries.MemoryUsageTotal(workload.Namespace, workload.Name), start, end, step)
	if len(memUsageData) > 0 && memLim > 0 {
		maxMem := 0.0
		for _, dp := range memUsageData {
			if dp.Value > maxMem {
				maxMem = dp.Value
			}
		}
		pct := (maxMem / memLim) * 100
		if pct > 85 {
			risks = append(risks, types.RiskFlag{
				Level:       "high",
				Type:        "oom",
				Description: fmt.Sprintf("Memory usage peaks at %.0f%% of limit (%.0f Mi / %.0f Mi)", pct, maxMem/1024/1024, memLim/1024/1024),
				Impact:      "Approaching OOM threshold — imminent risk of pod termination",
			})
		} else if pct > 70 {
			risks = append(risks, types.RiskFlag{
				Level:       "medium",
				Type:        "oom",
				Description: fmt.Sprintf("Memory usage reaches %.0f%% of limit during peaks", pct),
				Impact:      "Moderate OOM risk during traffic spikes",
			})
		}
	}

	// 2. CPU Throttling
	throttlePct, _ := rd.promClient.QueryInstant(ctx, rd.queries.CPUThrottling(workload.Namespace, workload.Name))
	if throttlePct > 50 {
		risks = append(risks, types.RiskFlag{
			Level:       "high",
			Type:        "throttling",
			Description: fmt.Sprintf("CPU throttling at %.0f%% — severely constrained", throttlePct),
			Impact:      "Significant request latency increase due to CPU starvation",
		})
	} else if throttlePct > 25 {
		risks = append(risks, types.RiskFlag{
			Level:       "medium",
			Type:        "throttling",
			Description: fmt.Sprintf("CPU throttling at %.0f%%", throttlePct),
			Impact:      "Noticeable latency impact from CPU contention",
		})
	}

	// 3. Over-Provisioning
	cpuReq, _ := rd.promClient.QueryInstant(ctx, rd.queries.CPURequest(workload.Namespace, workload.Name))
	cpuUsageData, _ := rd.promClient.QueryRange(ctx, rd.queries.CPUUsageTotal(workload.Namespace, workload.Name), start, end, step)
	if len(cpuUsageData) > 0 && cpuReq > 0 {
		avgCPU := 0.0
		for _, dp := range cpuUsageData {
			avgCPU += dp.Value
		}
		avgCPU /= float64(len(cpuUsageData))

		utilizationPct := (avgCPU / cpuReq) * 100
		if utilizationPct < 20 {
			risks = append(risks, types.RiskFlag{
				Level:       "medium",
				Type:        "over_provisioned",
				Description: fmt.Sprintf("Average CPU utilization is only %.0f%% of requested resources", utilizationPct),
				Impact:      "Wasting cluster resources — potential for significant cost savings",
			})
		}
	}

	memReq, _ := rd.promClient.QueryInstant(ctx, rd.queries.MemoryRequest(workload.Namespace, workload.Name))
	if len(memUsageData) > 0 && memReq > 0 {
		avgMem := 0.0
		for _, dp := range memUsageData {
			avgMem += dp.Value
		}
		avgMem /= float64(len(memUsageData))

		utilizationPct := (avgMem / memReq) * 100
		if utilizationPct < 20 {
			risks = append(risks, types.RiskFlag{
				Level:       "medium",
				Type:        "over_provisioned",
				Description: fmt.Sprintf("Average memory utilization is only %.0f%% of requested", utilizationPct),
				Impact:      "Over-provisioned memory — wasting cluster capacity",
			})
		}
	}

	return risks, nil
}
