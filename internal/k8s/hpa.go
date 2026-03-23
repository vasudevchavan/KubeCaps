package k8s

import (
	"context"
	"fmt"

	"github.com/vasudevchavan/kubecaps/pkg/types"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetHPAForWorkload finds HPAs targeting the given workload.
func (c *Client) GetHPAForWorkload(ctx context.Context, namespace, workloadName, workloadKind string) (*types.HPAConfig, error) {
	hpaList, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list HPAs in namespace %s: %w", namespace, err)
	}

	for _, hpa := range hpaList.Items {
		ref := hpa.Spec.ScaleTargetRef
		if ref.Name == workloadName && ref.Kind == workloadKind {
			return hpaToConfig(&hpa), nil
		}
	}

	return &types.HPAConfig{Found: false}, nil
}

// hpaToConfig converts a Kubernetes HPA to our HPAConfig type.
func hpaToConfig(hpa *autoscalingv2.HorizontalPodAutoscaler) *types.HPAConfig {
	config := &types.HPAConfig{
		Name:            hpa.Name,
		Namespace:       hpa.Namespace,
		MaxReplicas:     hpa.Spec.MaxReplicas,
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		Found:           true,
	}

	if hpa.Spec.MinReplicas != nil {
		config.MinReplicas = *hpa.Spec.MinReplicas
	} else {
		config.MinReplicas = 1
	}

	for _, metric := range hpa.Spec.Metrics {
		hpaMetric := convertHPAMetric(metric)
		if hpaMetric != nil {
			config.Metrics = append(config.Metrics, *hpaMetric)
		}
	}

	return config
}

// convertHPAMetric converts a Kubernetes HPA metric spec to our HPAMetric type.
func convertHPAMetric(metric autoscalingv2.MetricSpec) *types.HPAMetric {
	switch metric.Type {
	case autoscalingv2.ResourceMetricSourceType:
		if metric.Resource == nil {
			return nil
		}
		m := &types.HPAMetric{
			Type: "Resource",
			Name: string(metric.Resource.Name),
		}
		if metric.Resource.Target.Type == autoscalingv2.UtilizationMetricType && metric.Resource.Target.AverageUtilization != nil {
			m.TargetType = "Utilization"
			m.Target = float64(*metric.Resource.Target.AverageUtilization)
		} else if metric.Resource.Target.Type == autoscalingv2.AverageValueMetricType && metric.Resource.Target.AverageValue != nil {
			m.TargetType = "AverageValue"
			m.Target = float64(metric.Resource.Target.AverageValue.MilliValue()) / 1000
		}
		return m

	case autoscalingv2.PodsMetricSourceType:
		if metric.Pods == nil {
			return nil
		}
		m := &types.HPAMetric{
			Type: "Pods",
			Name: metric.Pods.Metric.Name,
		}
		if metric.Pods.Target.AverageValue != nil {
			m.TargetType = "AverageValue"
			m.Target = float64(metric.Pods.Target.AverageValue.MilliValue()) / 1000
		}
		return m

	case autoscalingv2.ObjectMetricSourceType:
		if metric.Object == nil {
			return nil
		}
		return &types.HPAMetric{
			Type: "Object",
			Name: metric.Object.Metric.Name,
		}

	case autoscalingv2.ExternalMetricSourceType:
		if metric.External == nil {
			return nil
		}
		m := &types.HPAMetric{
			Type: "External",
			Name: metric.External.Metric.Name,
		}
		if metric.External.Target.Value != nil {
			m.TargetType = "Value"
			m.Target = float64(metric.External.Target.Value.MilliValue()) / 1000
		} else if metric.External.Target.AverageValue != nil {
			m.TargetType = "AverageValue"
			m.Target = float64(metric.External.Target.AverageValue.MilliValue()) / 1000
		}
		return m
	}

	return nil
}
