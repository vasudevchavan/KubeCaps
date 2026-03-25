package config

// PredictionConfig holds configuration for prediction algorithms
type PredictionConfig struct {
	// Time window configurations
	DefaultTimeWindowHours int
	MaxTimeWindowHours     int
	MinTimeWindowHours     int
	
	// Data quality thresholds
	MinDataPointsForHighConfidence int
	MinDataPointsForMediumConfidence int
	MinDataPointsForLowConfidence int
	
	// Variance thresholds for workload classification
	SteadyWorkloadVarianceThreshold  float64
	ElasticWorkloadVarianceThreshold float64
	BurstyWorkloadVarianceThreshold  float64
	
	// Risk detection thresholds
	CPUThrottlingThresholdPercent float64
	HighVarianceThreshold         float64
	ModelDisagreementThreshold    float64
	
	// Optimization parameters
	DefaultCPULambda     float64
	DefaultMemoryLambda  float64
	ThrottlingCPULambda  float64
	OOMMemoryLambda      float64
	
	// Resource buffer multipliers
	CPULimitBufferMultiplier    float64
	MemoryLimitBufferMultiplier float64
	
	// EWMA parameters
	DefaultEWMAAlpha float64
	
	// Percentile configurations
	RequestPercentile float64
	LimitPercentile   float64
	SafetyPercentile  float64
}

// ScoringConfig holds configuration for scoring system
type ScoringConfig struct {
	// Component weights
	HPAWeight  float64
	VPAWeight  float64
	KEDAWeight float64
	
	// Score thresholds for grades
	GradeAPlus float64
	GradeA     float64
	GradeB     float64
	GradeC     float64
	GradeD     float64
	
	// Multi-component bonus
	MultiComponentBonusPercent float64
}

// PrometheusConfig holds configuration for Prometheus queries
type PrometheusConfig struct {
	// Query step duration in minutes
	DefaultStepMinutes int
	
	// Query timeout in seconds
	QueryTimeoutSeconds int
	
	// Retry configuration
	MaxRetries     int
	RetryDelayMs   int
	
	// Rate limiting
	MaxConcurrentQueries int
}

// DefaultPredictionConfig returns the default prediction configuration
func DefaultPredictionConfig() PredictionConfig {
	return PredictionConfig{
		// Time windows
		DefaultTimeWindowHours: 24,
		MaxTimeWindowHours:     720, // 30 days
		MinTimeWindowHours:     1,
		
		// Data quality
		MinDataPointsForHighConfidence:   100,
		MinDataPointsForMediumConfidence: 50,
		MinDataPointsForLowConfidence:    10,
		
		// Workload classification
		SteadyWorkloadVarianceThreshold:  0.20,
		ElasticWorkloadVarianceThreshold: 0.80,
		BurstyWorkloadVarianceThreshold:  0.80,
		
		// Risk detection
		CPUThrottlingThresholdPercent: 5.0,
		HighVarianceThreshold:         0.5,
		ModelDisagreementThreshold:    0.5,
		
		// Optimization
		DefaultCPULambda:    10.0,
		DefaultMemoryLambda: 50.0,
		ThrottlingCPULambda: 100.0,
		OOMMemoryLambda:     500.0,
		
		// Buffers
		CPULimitBufferMultiplier:    1.2,
		MemoryLimitBufferMultiplier: 1.2,
		
		// EWMA
		DefaultEWMAAlpha: 0.3,
		
		// Percentiles
		RequestPercentile: 50,
		LimitPercentile:   99,
		SafetyPercentile:  95,
	}
}

// DefaultScoringConfig returns the default scoring configuration
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		// Component weights (must sum to 1.0)
		HPAWeight:  0.35,
		VPAWeight:  0.35,
		KEDAWeight: 0.30,
		
		// Grade thresholds
		GradeAPlus: 9.0,
		GradeA:     8.0,
		GradeB:     7.0,
		GradeC:     6.0,
		GradeD:     5.0,
		
		// Bonuses
		MultiComponentBonusPercent: 5.0,
	}
}

// DefaultPrometheusConfig returns the default Prometheus configuration
func DefaultPrometheusConfig() PrometheusConfig {
	return PrometheusConfig{
		DefaultStepMinutes:       5,
		QueryTimeoutSeconds:      30,
		MaxRetries:               3,
		RetryDelayMs:             1000,
		MaxConcurrentQueries:     10,
	}
}

// Config holds all configuration
type Config struct {
	Prediction PredictionConfig
	Scoring    ScoringConfig
	Prometheus PrometheusConfig
}

// Default returns the default configuration
func Default() Config {
	return Config{
		Prediction: DefaultPredictionConfig(),
		Scoring:    DefaultScoringConfig(),
		Prometheus: DefaultPrometheusConfig(),
	}
}

// Made with Bob
