package types

// Config holds the customizable optimization parameters for KubeCaps.
type Config struct {
	Optimization OptimizationConfig `json:"optimization" yaml:"optimization"`
}

// OptimizationConfig holds parameters for the prediction engine.
type OptimizationConfig struct {
	// Cost weights
	CPUWeight    float64 `json:"cpuWeight" yaml:"cpuWeight"`
	MemoryWeight float64 `json:"memoryWeight" yaml:"memoryWeight"`

	// Risk penalties
	CPURiskPenalty    float64 `json:"cpuRiskPenalty" yaml:"cpuRiskPenalty"`
	MemoryRiskPenalty float64 `json:"memoryRiskPenalty" yaml:"memoryRiskPenalty"`

	// Failure amplification
	OOMMultiplier        float64 `json:"oomMultiplier" yaml:"oomMultiplier"`
	ThrottlingMultiplier float64 `json:"throttlingMultiplier" yaml:"throttlingMultiplier"`

	// Buffers
	BufferMultiplier        float64 `json:"bufferMultiplier" yaml:"bufferMultiplier"`
	MemorySafetyMargin      float64 `json:"memorySafetyMargin" yaml:"memorySafetyMargin"`
	StartupBufferMultiplier float64 `json:"startupBufferMultiplier" yaml:"startupBufferMultiplier"`

	// Percentiles (critical)
	TargetCPUPercentile    float64 `json:"targetCPUPercentile" yaml:"targetCPUPercentile"`
	TargetMemoryPercentile float64 `json:"targetMemoryPercentile" yaml:"targetMemoryPercentile"`

	// Workload behavior
	BurstinessWeight float64 `json:"burstinessWeight" yaml:"burstinessWeight"`

	// Scaling interaction
	ScalingSensitivity float64 `json:"scalingSensitivity" yaml:"scalingSensitivity"`

	// Scheduling constraints
	BinPackingWeight float64 `json:"binPackingWeight" yaml:"binPackingWeight"`

	// SLA awareness
	LatencySensitivity float64 `json:"latencySensitivity" yaml:"latencySensitivity"`

	// Minimum bounds (CPU in millicores, Memory in MB)
	MinCPU    float64 `json:"minCPU" yaml:"minCPU"`
	MinMemory float64 `json:"minMemory" yaml:"minMemory"`
}

// DefaultConfig returns the default optimization parameters.
func DefaultConfig() Config {
	return Config{
		Optimization: OptimizationConfig{
			// Cost weights
			CPUWeight:    1.0,
			MemoryWeight: 1.0,

			// Risk penalties
			CPURiskPenalty:    10.0,
			MemoryRiskPenalty: 30.0,

			// Failure amplification
			OOMMultiplier:        10.0,
			ThrottlingMultiplier: 10.0,

			// Buffers
			BufferMultiplier:        1.2,
			MemorySafetyMargin:      1.3,
			StartupBufferMultiplier: 1.5,

			// Percentiles (critical)
			TargetCPUPercentile:    0.95,
			TargetMemoryPercentile: 0.99,

			// Workload behavior
			BurstinessWeight: 1.5,

			// Scaling interaction
			ScalingSensitivity: 1.2,

			// Scheduling constraints
			BinPackingWeight: 0.8,

			// SLA awareness
			LatencySensitivity: 1.5,

			// Minimum bounds
			MinCPU:    50, // millicores
			MinMemory: 64, // MB
		},
	}
}
