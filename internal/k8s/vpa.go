package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vasudevchavan/kubecaps/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// VPA GVR for dynamic client access.
var vpaGVR = schema.GroupVersionResource{
	Group:    "autoscaling.k8s.io",
	Version:  "v1",
	Resource: "verticalpodautoscalers",
}

// GetVPAForWorkload finds VPAs targeting the given workload.
func (c *Client) GetVPAForWorkload(ctx context.Context, namespace, workloadName, workloadKind string) (*types.VPAConfig, error) {
	dynClient, err := c.getDynamicClient()
	if err != nil {
		return &types.VPAConfig{Found: false}, nil
	}

	vpaList, err := dynClient.Resource(vpaGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// VPA CRD may not be installed
		return &types.VPAConfig{Found: false}, nil
	}

	for _, item := range vpaList.Items {
		raw, _ := json.Marshal(item.Object)
		var vpa vpaObject
		if err := json.Unmarshal(raw, &vpa); err != nil {
			continue
		}

		if vpa.Spec.TargetRef.Name == workloadName && vpa.Spec.TargetRef.Kind == workloadKind {
			return vpaObjectToConfig(&vpa, namespace), nil
		}
	}

	return &types.VPAConfig{Found: false}, nil
}

// vpaObject is a simplified representation of a VPA for JSON unmarshaling.
type vpaObject struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		TargetRef struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
		} `json:"targetRef"`
		UpdatePolicy struct {
			UpdateMode string `json:"updateMode"`
		} `json:"updatePolicy"`
		ResourcePolicy struct {
			ContainerPolicies []struct {
				ContainerName string `json:"containerName"`
				MinAllowed    struct {
					CPU    string `json:"cpu"`
					Memory string `json:"memory"`
				} `json:"minAllowed"`
				MaxAllowed struct {
					CPU    string `json:"cpu"`
					Memory string `json:"memory"`
				} `json:"maxAllowed"`
			} `json:"containerPolicies"`
		} `json:"resourcePolicy"`
	} `json:"spec"`
	Status struct {
		Recommendation struct {
			ContainerRecommendations []struct {
				ContainerName  string `json:"containerName"`
				Target         map[string]string `json:"target"`
				LowerBound     map[string]string `json:"lowerBound"`
				UpperBound     map[string]string `json:"upperBound"`
				UncappedTarget map[string]string `json:"uncappedTarget"`
			} `json:"containerRecommendations"`
		} `json:"recommendation"`
	} `json:"status"`
}

// vpaObjectToConfig converts a VPA object to our VPAConfig type.
func vpaObjectToConfig(vpa *vpaObject, namespace string) *types.VPAConfig {
	config := &types.VPAConfig{
		Name:       vpa.Metadata.Name,
		Namespace:  namespace,
		UpdateMode: vpa.Spec.UpdatePolicy.UpdateMode,
		Found:      true,
	}

	for _, cp := range vpa.Spec.ResourcePolicy.ContainerPolicies {
		policy := types.VPAContainerPolicy{
			ContainerName: cp.ContainerName,
			MinCPU:        parseCPU(cp.MinAllowed.CPU),
			MaxCPU:        parseCPU(cp.MaxAllowed.CPU),
			MinMemory:     parseMemory(cp.MinAllowed.Memory),
			MaxMemory:     parseMemory(cp.MaxAllowed.Memory),
		}
		config.ContainerPolicies = append(config.ContainerPolicies, policy)
	}

	for _, cr := range vpa.Status.Recommendation.ContainerRecommendations {
		rec := types.VPARecommendation{
			ContainerName:  cr.ContainerName,
			TargetCPU:      parseCPU(cr.Target["cpu"]),
			TargetMemory:   parseMemory(cr.Target["memory"]),
			LowerCPU:       parseCPU(cr.LowerBound["cpu"]),
			LowerMemory:    parseMemory(cr.LowerBound["memory"]),
			UpperCPU:       parseCPU(cr.UpperBound["cpu"]),
			UpperMemory:    parseMemory(cr.UpperBound["memory"]),
			UncappedCPU:    parseCPU(cr.UncappedTarget["cpu"]),
			UncappedMemory: parseMemory(cr.UncappedTarget["memory"]),
		}
		config.Recommendations = append(config.Recommendations, rec)
	}

	return config
}

// parseCPU parses a CPU string (e.g., "250m", "1") to cores.
func parseCPU(s string) float64 {
	if s == "" {
		return 0
	}
	var val float64
	if n, _ := fmt.Sscanf(s, "%fm", &val); n == 1 {
		return val / 1000
	}
	if n, _ := fmt.Sscanf(s, "%f", &val); n == 1 {
		return val
	}
	return 0
}

// parseMemory parses a memory string (e.g., "128Mi", "1Gi") to bytes.
func parseMemory(s string) float64 {
	if s == "" {
		return 0
	}
	var val float64
	if n, _ := fmt.Sscanf(s, "%fGi", &val); n == 1 {
		return val * 1024 * 1024 * 1024
	}
	if n, _ := fmt.Sscanf(s, "%fMi", &val); n == 1 {
		return val * 1024 * 1024
	}
	if n, _ := fmt.Sscanf(s, "%fKi", &val); n == 1 {
		return val * 1024
	}
	if n, _ := fmt.Sscanf(s, "%f", &val); n == 1 {
		return val
	}
	return 0
}
