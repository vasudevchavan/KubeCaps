package recommender

import (
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// ResolveConflicts orchestrates the recommendations to prevent production issues.
// Modifies the suggestions in-place and returns a list of architectural insights/warnings.
func ResolveConflicts(hpa *types.HPASuggestion, vpa *types.VPASuggestion, keda *types.KEDASuggestion) []string {
	var insights []string

	// Rule 1: HPA + VPA Conflict
	// If both HPA and VPA are turned on, Kubernetes components will fight over CPU/Memory.
	// VPA will evict pods to resize them while HPA creates new ones to meet utilization.
	if hpa.Enabled && (vpa.Mode == "Auto" || vpa.Mode == "Initial") {
		vpa.Mode = "Off"
		insights = append(insights, "⚠️ HPA + VPA Conflict Detected: Safely forcing VPA Mode to 'Off'. VPA will only recommend limits, allowing HPA to exclusively scale pod counts.")
	}

	// Rule 2: HPA + KEDA Overlap
	// If KEDA is enabled on CPU/Memory metrics alongside HPA, they compete directly for the horizontal scaling replica bounds.
	if hpa.Enabled && keda.Enabled {
		if keda.Trigger == "prometheus" || keda.Trigger == "cpu" || keda.Trigger == "memory" {
			keda.Enabled = false
			insights = append(insights, "⚠️ HPA + KEDA Conflict Detected: Safely disabling KEDA. HPA is already scaling on resource metrics. Retaining both risks oscillating replica counts.")
		}
	}

	// General Architecture Confirmations (No conflict)
	if hpa.Enabled && !keda.Enabled {
		insights = append(insights, "✅ Valid Architecture: Standalone HPA managing elastic load variance.")
	} else if !hpa.Enabled && keda.Enabled {
		insights = append(insights, "✅ Valid Architecture: Standalone KEDA managing event-driven scale.")
	} else if !hpa.Enabled && !keda.Enabled && (vpa.Mode == "Auto" || vpa.Mode == "Initial") {
		insights = append(insights, "✅ Valid Architecture: Standalone VPA rightsizing steady workloads.")
	} else if !hpa.Enabled && !keda.Enabled && vpa.Mode == "Off" {
		insights = append(insights, "ℹ️ No active scaler suggested. Workload relies on static resources.")
	}

	return insights
}
