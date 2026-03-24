package recommender

import (
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// GenerateKEDA computes KEDA ScaledObject configurations.
func GenerateKEDA(workloadType types.WorkloadType, cpuUsageLow bool, latencyHigh bool, queueLag float64) types.KEDASuggestion {
	enabled := false
	trigger := ""
	threshold := "100" // Default generic threshold
	var cooldown int32 = 300

	// Event-driven heuristics
	if workloadType == types.WorkloadTypeEventDriven || queueLag > 0 {
		enabled = true
		trigger = "kafka"       // For queue lag, we assume kafka as default metric proxy unless told otherwise
		threshold = "50"        // Lower threshold for high queue lag
		cooldown = 60           // Fast cooldown for pure event queues
	} else if cpuUsageLow && latencyHigh {
		// CPU isn't correlated with scale but latency is spiking (classic HTTP burst)
		enabled = true
		trigger = "prometheus"  // Recommend scaling on custom prometheus HTTP metrics
		threshold = "1000"      // RPS threshold
		cooldown = 120
	// } else if workloadType == types.WorkloadTypeBursty {
	//	enabled = true
	//	trigger = "cron"
	}

	return types.KEDASuggestion{
		Enabled:        enabled,
		Trigger:        trigger,
		Threshold:      threshold,
		CooldownPeriod: cooldown,
	}
}
