package recommender

import (
	"math"

	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// GenerateHPA computes the optimal HPA configuration.
// By default, target utilization is around 70-80% for safety.
func GenerateHPA(workloadType types.WorkloadType, p50 float64, peak float64, currentReplicas int32, hasThrottling bool) types.HPASuggestion {
	enabled := false
	if workloadType == types.WorkloadTypeElastic || workloadType == types.WorkloadTypeBursty {
		enabled = true
	}

	targetValue := "80%"
	if hasThrottling {
		// If the workload is experiencing severe CPU throttling, we drop the HPA target
		// to force the HorizontalPodAutoscaler to spawn replicas earlier (at 60% load instead of 80%).
		targetValue = "60%"
	}

	// Calculate mathematical min/max replicas
	// P50 baseline gives us a sense of floor load
	minRep := int32(math.Max(1.0, math.Ceil(p50/0.8)))

	// Peak load dictates our max bounds
	maxRep := int32(math.Max(2.0, math.Ceil(peak/0.8)))

	if maxRep <= minRep {
		maxRep = minRep * 2
	}

	// Buffer for bursty workloads
	if workloadType == types.WorkloadTypeBursty {
		maxRep = int32(math.Ceil(float64(maxRep) * 1.5))
	}

	// Extra safety buffer if it's actively throttling
	if hasThrottling {
		maxRep = int32(math.Ceil(float64(maxRep) * 1.25))
	}

	return types.HPASuggestion{
		Enabled:      enabled,
		TargetMetric: "cpu",
		TargetValue:  targetValue,
		MinReplicas:  minRep,
		MaxReplicas:  maxRep,
	}
}
