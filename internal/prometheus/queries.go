package prometheus

import "fmt"

// Queries provides PromQL query templates for various metrics.
type Queries struct{}

// NewQueries returns a new Queries instance.
func NewQueries() *Queries {
	return &Queries{}
}

// CPUUsageByPod returns PromQL for CPU usage rate by pod in a namespace.
func (q *Queries) CPUUsageByPod(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s.*", container!="", container!="POD"}[5m])) by (pod)`,
		namespace, podPrefix,
	)
}

// CPUUsageTotal returns PromQL for total CPU usage of a workload.
func (q *Queries) CPUUsageTotal(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s.*", container!="", container!="POD"}[5m]))`,
		namespace, podPrefix,
	)
}

// MemoryUsageByPod returns PromQL for memory working set by pod.
func (q *Queries) MemoryUsageByPod(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s", pod=~"%s.*", container!="", container!="POD"}) by (pod)`,
		namespace, podPrefix,
	)
}

// MemoryUsageTotal returns PromQL for total memory working set of a workload.
func (q *Queries) MemoryUsageTotal(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s", pod=~"%s.*", container!="", container!="POD"})`,
		namespace, podPrefix,
	)
}

// CPURequest returns PromQL for CPU requests.
func (q *Queries) CPURequest(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_container_resource_requests{namespace="%s", pod=~"%s.*", resource="cpu", container!=""})`,
		namespace, podPrefix,
	)
}

// CPULimit returns PromQL for CPU limits.
func (q *Queries) CPULimit(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_container_resource_limits{namespace="%s", pod=~"%s.*", resource="cpu", container!=""})`,
		namespace, podPrefix,
	)
}

// MemoryRequest returns PromQL for memory requests.
func (q *Queries) MemoryRequest(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_container_resource_requests{namespace="%s", pod=~"%s.*", resource="memory", container!=""})`,
		namespace, podPrefix,
	)
}

// MemoryLimit returns PromQL for memory limits.
func (q *Queries) MemoryLimit(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_container_resource_limits{namespace="%s", pod=~"%s.*", resource="memory", container!=""})`,
		namespace, podPrefix,
	)
}

// CPUThrottling returns PromQL for CPU throttling percentage.
func (q *Queries) CPUThrottling(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(rate(container_cpu_cfs_throttled_periods_total{namespace="%s", pod=~"%s.*", container!=""}[5m])) / sum(rate(container_cpu_cfs_periods_total{namespace="%s", pod=~"%s.*", container!=""}[5m])) * 100`,
		namespace, podPrefix, namespace, podPrefix,
	)
}

// OOMEvents returns PromQL for OOM killed events.
func (q *Queries) OOMEvents(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_container_status_last_terminated_reason{namespace="%s", pod=~"%s.*", reason="OOMKilled"})`,
		namespace, podPrefix,
	)
}

// RequestRate returns PromQL for HTTP request rate (requires service mesh / ingress metrics).
func (q *Queries) RequestRate(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(rate(http_requests_total{namespace="%s", pod=~"%s.*"}[5m]))`,
		namespace, podPrefix,
	)
}

// HPAReplicaCount returns PromQL for HPA desired replica count over time.
func (q *Queries) HPAReplicaCount(namespace, hpaName string) string {
	return fmt.Sprintf(
		`kube_horizontalpodautoscaler_status_desired_replicas{namespace="%s", horizontalpodautoscaler="%s"}`,
		namespace, hpaName,
	)
}

// HPACurrentReplicas returns PromQL for HPA current replica count over time.
func (q *Queries) HPACurrentReplicas(namespace, hpaName string) string {
	return fmt.Sprintf(
		`kube_horizontalpodautoscaler_status_current_replicas{namespace="%s", horizontalpodautoscaler="%s"}`,
		namespace, hpaName,
	)
}

// PodReadyCount returns PromQL for the number of ready pods for a workload.
func (q *Queries) PodReadyCount(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(kube_pod_status_ready{namespace="%s", pod=~"%s.*", condition="true"})`,
		namespace, podPrefix,
	)
}

// KafkaConsumerLag returns PromQL for Kafka consumer group lag.
func (q *Queries) KafkaConsumerLag(consumerGroup, topic string) string {
	return fmt.Sprintf(
		`sum(kafka_consumergroup_lag{consumergroup="%s", topic="%s"}) by (consumergroup)`,
		consumerGroup, topic,
	)
}

// CPUUtilizationPercent returns PromQL for CPU utilization as a percentage of requests.
func (q *Queries) CPUUtilizationPercent(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s.*", container!="", container!="POD"}[5m])) / sum(kube_pod_container_resource_requests{namespace="%s", pod=~"%s.*", resource="cpu", container!=""}) * 100`,
		namespace, podPrefix, namespace, podPrefix,
	)
}

// MemoryUtilizationPercent returns PromQL for memory utilization as a percentage of requests.
func (q *Queries) MemoryUtilizationPercent(namespace, podPrefix string) string {
	return fmt.Sprintf(
		`sum(container_memory_working_set_bytes{namespace="%s", pod=~"%s.*", container!="", container!="POD"}) / sum(kube_pod_container_resource_requests{namespace="%s", pod=~"%s.*", resource="memory", container!=""}) * 100`,
		namespace, podPrefix, namespace, podPrefix,
	)
}
