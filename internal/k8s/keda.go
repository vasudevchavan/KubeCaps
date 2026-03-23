package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vasudevchavan/kubecaps/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KEDA GVR for dynamic client access.
var kedaScaledObjectGVR = schema.GroupVersionResource{
	Group:    "keda.sh",
	Version:  "v1alpha1",
	Resource: "scaledobjects",
}

// GetKEDAForWorkload finds KEDA ScaledObjects targeting the given workload.
func (c *Client) GetKEDAForWorkload(ctx context.Context, namespace, workloadName, workloadKind string) (*types.KEDAConfig, error) {
	dynClient, err := c.getDynamicClient()
	if err != nil {
		return &types.KEDAConfig{Found: false}, nil
	}

	soList, err := dynClient.Resource(kedaScaledObjectGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// KEDA CRD may not be installed
		return &types.KEDAConfig{Found: false}, nil
	}

	for _, item := range soList.Items {
		raw, _ := json.Marshal(item.Object)
		var so scaledObject
		if err := json.Unmarshal(raw, &so); err != nil {
			continue
		}

		if so.Spec.ScaleTargetRef.Name == workloadName {
			return scaledObjectToConfig(&so, namespace), nil
		}
	}

	return &types.KEDAConfig{Found: false}, nil
}

// scaledObject is a simplified representation of a KEDA ScaledObject.
type scaledObject struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		ScaleTargetRef struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"scaleTargetRef"`
		MinReplicaCount *int32 `json:"minReplicaCount"`
		MaxReplicaCount *int32 `json:"maxReplicaCount"`
		PollingInterval *int32 `json:"pollingInterval"`
		CooldownPeriod  *int32 `json:"cooldownPeriod"`
		Triggers        []struct {
			Type     string            `json:"type"`
			Metadata map[string]string `json:"metadata"`
			AuthenticationRef *struct {
				Name string `json:"name"`
			} `json:"authenticationRef"`
		} `json:"triggers"`
	} `json:"spec"`
}

// scaledObjectToConfig converts a KEDA ScaledObject to our KEDAConfig type.
func scaledObjectToConfig(so *scaledObject, namespace string) *types.KEDAConfig {
	config := &types.KEDAConfig{
		Name:      so.Metadata.Name,
		Namespace: namespace,
		Found:     true,
	}

	if so.Spec.MinReplicaCount != nil {
		config.MinReplicas = *so.Spec.MinReplicaCount
	}
	if so.Spec.MaxReplicaCount != nil {
		config.MaxReplicas = *so.Spec.MaxReplicaCount
	} else {
		config.MaxReplicas = 100 // KEDA default
	}
	if so.Spec.PollingInterval != nil {
		config.PollingInterval = *so.Spec.PollingInterval
	} else {
		config.PollingInterval = 30 // KEDA default
	}
	if so.Spec.CooldownPeriod != nil {
		config.CooldownPeriod = *so.Spec.CooldownPeriod
	} else {
		config.CooldownPeriod = 300 // KEDA default
	}

	for _, t := range so.Spec.Triggers {
		trigger := types.KEDATrigger{
			Type:     t.Type,
			Metadata: t.Metadata,
		}
		if t.AuthenticationRef != nil {
			trigger.AuthRef = t.AuthenticationRef.Name
		}
		config.Triggers = append(config.Triggers, trigger)
	}

	return config
}

// getDynamicClient returns a dynamic Kubernetes client for CRD access.
func (c *Client) getDynamicClient() (dynamicClient, error) {
	if c.dynClient != nil {
		return c.dynClient, nil
	}

	dc, err := newDynamicClient(c.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	c.dynClient = dc
	return dc, nil
}
