package recommender

import (
	"fmt"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// GenerateVPA computes the optimal VPA configuration with safety buffers.
func GenerateVPA(cpuRaw float64, memRaw float64, existingHPA bool, isCritical bool) types.VPASuggestion {
	// Add 15% safety buffer
	bufferedCPU := cpuRaw * 1.15
	bufferedMem := memRaw * 1.15

	mode := "Auto"
	if existingHPA {
		mode = "Off" // Conflict prevention: HPA controls scaling, VPA only advises
	} else if isCritical {
		mode = "Initial" // Critical workloads shouldn't be disrupted mid-flight
	}

	return types.VPASuggestion{
		Mode:   mode,
		CPU:    formatCPU(bufferedCPU),
		Memory: formatMemory(bufferedMem),
	}
}

// formatCPU formats CPU cores to Kubernetes format (e.g., "250m" or "1.5").
func formatCPU(cores float64) string {
	if cores < 1.0 {
		return fmt.Sprintf("%dm", int(cores*1000))
	}
	return fmt.Sprintf("%.1f", cores)
}

// formatMemory formats bytes to Kubernetes format (e.g., "128Mi", "1Gi").
func formatMemory(bytes float64) string {
	gi := bytes / (1024 * 1024 * 1024)
	if gi >= 1.0 {
		return fmt.Sprintf("%.1fGi", gi)
	}
	mi := bytes / (1024 * 1024)
	return fmt.Sprintf("%.0fMi", mi)
}
