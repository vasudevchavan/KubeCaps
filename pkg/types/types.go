package types

import "time"

// WorkloadInfo holds basic information about a Kubernetes workload.
type WorkloadInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Kind      string            `json:"kind"` // Deployment, StatefulSet, DaemonSet
	Labels    map[string]string `json:"labels"`
	Replicas  int32             `json:"replicas"`
}

// DataPoint represents a single metric data point in a time series.
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ResourceUsage holds time-series resource usage data.
type ResourceUsage struct {
	CPUUsage    []DataPoint `json:"cpuUsage"`    // CPU cores
	MemoryUsage []DataPoint `json:"memoryUsage"` // Bytes
	CPURequest  float64     `json:"cpuRequest"`  // CPU cores requested
	CPULimit    float64     `json:"cpuLimit"`     // CPU cores limit
	MemRequest  float64     `json:"memRequest"`   // Memory bytes requested
	MemLimit    float64     `json:"memLimit"`     // Memory bytes limit
}

// HPAMetric describes a single HPA metric configuration.
type HPAMetric struct {
	Type       string  `json:"type"`       // Resource, Pods, Object, External
	Name       string  `json:"name"`       // cpu, memory, custom metric name
	TargetType string  `json:"targetType"` // Utilization, Value, AverageValue
	Target     float64 `json:"target"`     // Target value
}

// HPAConfig holds HPA configuration for a workload.
type HPAConfig struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	MinReplicas     int32       `json:"minReplicas"`
	MaxReplicas     int32       `json:"maxReplicas"`
	CurrentReplicas int32       `json:"currentReplicas"`
	DesiredReplicas int32       `json:"desiredReplicas"`
	Metrics         []HPAMetric `json:"metrics"`
	Found           bool        `json:"found"`
}

// VPAContainerPolicy holds VPA resource policy for a container.
type VPAContainerPolicy struct {
	ContainerName string  `json:"containerName"`
	MinCPU        float64 `json:"minCPU"`    // cores
	MaxCPU        float64 `json:"maxCPU"`    // cores
	MinMemory     float64 `json:"minMemory"` // bytes
	MaxMemory     float64 `json:"maxMemory"` // bytes
}

// VPARecommendation holds VPA recommendation for a container.
type VPARecommendation struct {
	ContainerName string  `json:"containerName"`
	TargetCPU     float64 `json:"targetCPU"`     // cores
	TargetMemory  float64 `json:"targetMemory"`  // bytes
	LowerCPU      float64 `json:"lowerCPU"`      // cores
	LowerMemory   float64 `json:"lowerMemory"`   // bytes
	UpperCPU      float64 `json:"upperCPU"`      // cores
	UpperMemory   float64 `json:"upperMemory"`   // bytes
	UncappedCPU   float64 `json:"uncappedCPU"`   // cores
	UncappedMemory float64 `json:"uncappedMemory"` // bytes
}

// VPAConfig holds VPA configuration for a workload.
type VPAConfig struct {
	Name             string               `json:"name"`
	Namespace        string               `json:"namespace"`
	UpdateMode       string               `json:"updateMode"` // Off, Initial, Auto
	ContainerPolicies []VPAContainerPolicy `json:"containerPolicies"`
	Recommendations  []VPARecommendation  `json:"recommendations"`
	Found            bool                 `json:"found"`
}

// KEDATrigger holds a KEDA trigger configuration.
type KEDATrigger struct {
	Type       string            `json:"type"`       // kafka, prometheus, cron, etc.
	Metadata   map[string]string `json:"metadata"`   // Trigger-specific metadata
	AuthRef    string            `json:"authRef"`     // TriggerAuthentication reference
}

// KEDAConfig holds KEDA ScaledObject configuration.
type KEDAConfig struct {
	Name            string        `json:"name"`
	Namespace       string        `json:"namespace"`
	MinReplicas     int32         `json:"minReplicas"`
	MaxReplicas     int32         `json:"maxReplicas"`
	PollingInterval int32         `json:"pollingInterval"` // seconds
	CooldownPeriod  int32         `json:"cooldownPeriod"`  // seconds
	Triggers        []KEDATrigger `json:"triggers"`
	Found           bool          `json:"found"`
}

// Recommendation holds a resource or autoscaling recommendation.
type Recommendation struct {
	WorkloadName    string  `json:"workloadName"`
	WorkloadKind    string  `json:"workloadKind"`
	Namespace       string  `json:"namespace"`
	CPURequest      string  `json:"cpuRequest"`      // e.g. "250m"
	CPULimit        string  `json:"cpuLimit"`         // e.g. "500m"
	MemoryRequest   string  `json:"memoryRequest"`    // e.g. "128Mi"
	MemoryLimit     string  `json:"memoryLimit"`      // e.g. "256Mi"
	CPURequestRaw   float64 `json:"cpuRequestRaw"`    // cores
	CPULimitRaw     float64 `json:"cpuLimitRaw"`      // cores
	MemRequestRaw   float64 `json:"memRequestRaw"`    // bytes
	MemLimitRaw     float64 `json:"memLimitRaw"`      // bytes
	Confidence      float64 `json:"confidence"`       // 0.0 - 1.0
	Explanation     string  `json:"explanation"`
	TimeWindowHours int     `json:"timeWindowHours"`
}

// Insight represents a single actionable insight from evaluation.
type Insight struct {
	Severity    string `json:"severity"`    // critical, warning, info
	Category    string `json:"category"`    // scaling, resources, cost, reliability
	Title       string `json:"title"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`    // Prometheus metric evidence
	Suggestion  string `json:"suggestion"`
}

// RiskFlag represents a detected risk.
type RiskFlag struct {
	Level       string `json:"level"`       // high, medium, low
	Type        string `json:"type"`        // oom, throttling, scaling_lag, over_provisioned
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// ComponentScore holds the evaluation score for a single component.
type ComponentScore struct {
	Component   string  `json:"component"`   // HPA, VPA, KEDA
	Score       float64 `json:"score"`       // 1-10
	MaxScore    float64 `json:"maxScore"`    // 10
	Details     string  `json:"details"`
}

// EvaluationResult holds the complete evaluation output.
type EvaluationResult struct {
	Workload        WorkloadInfo     `json:"workload"`
	HPAConfig       *HPAConfig       `json:"hpaConfig,omitempty"`
	VPAConfig       *VPAConfig       `json:"vpaConfig,omitempty"`
	KEDAConfig      *KEDAConfig      `json:"kedaConfig,omitempty"`
	Recommendation  *Recommendation  `json:"recommendation,omitempty"`
	Scores          []ComponentScore `json:"scores"`
	OverallScore    float64          `json:"overallScore"`
	Insights        []Insight        `json:"insights"`
	Risks           []RiskFlag       `json:"risks"`
	TimeWindow      string           `json:"timeWindow"`
	AnalyzedAt      time.Time        `json:"analyzedAt"`
}

// ComparisonRow holds a single row of current vs recommended config comparison.
type ComparisonRow struct {
	Aspect       string `json:"aspect"`
	CurrentValue string `json:"currentValue"`
	Recommended  string `json:"recommended"`
	Delta        string `json:"delta"`
}

// CorrelationResult holds cross-metric correlation analysis.
type CorrelationResult struct {
	MetricA        string  `json:"metricA"`
	MetricB        string  `json:"metricB"`
	Correlation    float64 `json:"correlation"`    // -1.0 to 1.0
	Interpretation string  `json:"interpretation"`
}
