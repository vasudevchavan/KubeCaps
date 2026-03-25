package ai

import (
	"fmt"
	"math"
	"sort"

	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// WorkloadDNA represents the behavioral fingerprint of a workload
type WorkloadDNA struct {
	WorkloadName string                 `json:"workload_name"`
	Namespace    string                 `json:"namespace"`
	Seasonality  []string               `json:"seasonality"`           // hourly, daily, weekly, monthly
	Volatility   float64                `json:"volatility"`            // 0-1, coefficient of variation
	GrowthRate   float64                `json:"growth_rate"`           // percentage per week
	Predictability float64              `json:"predictability"`        // 0-1, how predictable the pattern is
	CostSensitivity string              `json:"cost_sensitivity"`      // low, medium, high
	LatencySensitivity string           `json:"latency_sensitivity"`   // low, medium, high, critical
	ResourceProfile ResourceProfile     `json:"resource_profile"`
	TrafficPattern  TrafficPattern      `json:"traffic_pattern"`
	Characteristics map[string]float64  `json:"characteristics"`       // Additional metrics
}

// ResourceProfile describes resource usage patterns
type ResourceProfile struct {
	CPUIntensity    string  `json:"cpu_intensity"`     // low, medium, high
	MemoryIntensity string  `json:"memory_intensity"`  // low, medium, high
	CPUToMemoryRatio float64 `json:"cpu_to_memory_ratio"` // CPU cores per GB
	BurstCapability  float64 `json:"burst_capability"`  // 0-1, ability to handle bursts
}

// TrafficPattern describes traffic characteristics
type TrafficPattern struct {
	Type           string  `json:"type"`            // steady, periodic, bursty, random
	PeakToAvgRatio float64 `json:"peak_to_avg_ratio"` // How much peaks exceed average
	DailyVariation float64 `json:"daily_variation"` // 0-1, variation within a day
	WeeklyVariation float64 `json:"weekly_variation"` // 0-1, variation within a week
}

// DNAAnalyzer analyzes workload behavior and creates DNA profiles
type DNAAnalyzer struct {
	minDataPoints int
}

// NewDNAAnalyzer creates a new DNA analyzer
func NewDNAAnalyzer() *DNAAnalyzer {
	return &DNAAnalyzer{
		minDataPoints: 100,
	}
}

// AnalyzeDNA creates a DNA profile for a workload
func (a *DNAAnalyzer) AnalyzeDNA(
	workloadName, namespace string,
	cpuData, memData, trafficData []types.DataPoint,
) (*WorkloadDNA, error) {
	
	if len(cpuData) < a.minDataPoints {
		return nil, fmt.Errorf("insufficient data points: need at least %d, got %d", a.minDataPoints, len(cpuData))
	}

	dna := &WorkloadDNA{
		WorkloadName:    workloadName,
		Namespace:       namespace,
		Characteristics: make(map[string]float64),
	}

	// Analyze seasonality
	dna.Seasonality = a.detectSeasonality(cpuData)

	// Calculate volatility (coefficient of variation)
	dna.Volatility = a.calculateVolatility(cpuData)

	// Estimate growth rate
	dna.GrowthRate = a.estimateGrowthRate(cpuData)

	// Calculate predictability
	dna.Predictability = a.calculatePredictability(cpuData)

	// Determine sensitivities (heuristic-based for now)
	dna.CostSensitivity = a.determineCostSensitivity(dna.Volatility, dna.GrowthRate)
	dna.LatencySensitivity = a.determineLatencySensitivity(trafficData)

	// Analyze resource profile
	dna.ResourceProfile = a.analyzeResourceProfile(cpuData, memData)

	// Analyze traffic pattern
	dna.TrafficPattern = a.analyzeTrafficPattern(trafficData)

	// Additional characteristics
	dna.Characteristics["cpu_mean"] = calculateMeanFromDataPoints(cpuData)
	dna.Characteristics["cpu_p95"] = calculatePercentileFromDataPoints(cpuData, 95)
	dna.Characteristics["cpu_p99"] = calculatePercentileFromDataPoints(cpuData, 99)
	dna.Characteristics["mem_mean"] = calculateMeanFromDataPoints(memData)
	dna.Characteristics["mem_p95"] = calculatePercentileFromDataPoints(memData, 95)
	
	if len(trafficData) > 0 {
		dna.Characteristics["traffic_mean"] = calculateMeanFromDataPoints(trafficData)
		dna.Characteristics["traffic_p95"] = calculatePercentileFromDataPoints(trafficData, 95)
	}

	return dna, nil
}

// detectSeasonality detects periodic patterns in the data
func (a *DNAAnalyzer) detectSeasonality(data []types.DataPoint) []string {
	var patterns []string

	// Check for hourly pattern (if we have enough data)
	if len(data) >= 24 {
		if a.hasPeriodicPattern(data, 12) { // 12 * 5min = 1 hour
			patterns = append(patterns, "hourly")
		}
	}

	// Check for daily pattern
	if len(data) >= 288 { // 288 * 5min = 1 day
		if a.hasPeriodicPattern(data, 288) {
			patterns = append(patterns, "daily")
		}
	}

	// Check for weekly pattern
	if len(data) >= 2016 { // 2016 * 5min = 1 week
		if a.hasPeriodicPattern(data, 2016) {
			patterns = append(patterns, "weekly")
		}
	}

	if len(patterns) == 0 {
		patterns = append(patterns, "none")
	}

	return patterns
}

// hasPeriodicPattern checks if data has a periodic pattern with given period
func (a *DNAAnalyzer) hasPeriodicPattern(data []types.DataPoint, period int) bool {
	if len(data) < period*2 {
		return false
	}

	// Simple autocorrelation check
	correlation := a.calculateAutocorrelation(data, period)
	return correlation > 0.5 // Threshold for considering it periodic
}

// calculateAutocorrelation calculates autocorrelation at given lag
func (a *DNAAnalyzer) calculateAutocorrelation(data []types.DataPoint, lag int) float64 {
	if len(data) < lag+1 {
		return 0
	}

	values := make([]float64, len(data))
	for i, dp := range data {
		values[i] = dp.Value
	}

	mean := calculateMean(values)
	
	numerator := 0.0
	denominator := 0.0

	for i := 0; i < len(values)-lag; i++ {
		numerator += (values[i] - mean) * (values[i+lag] - mean)
	}

	for i := 0; i < len(values); i++ {
		denominator += (values[i] - mean) * (values[i] - mean)
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateVolatility calculates coefficient of variation
func (a *DNAAnalyzer) calculateVolatility(data []types.DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}

	mean := calculateMeanFromDataPoints(data)
	if mean == 0 {
		return 0
	}

	stdDev := calculateStdDevFromDataPoints(data, mean)
	return stdDev / mean // Coefficient of variation
}

// estimateGrowthRate estimates the growth rate using linear regression
func (a *DNAAnalyzer) estimateGrowthRate(data []types.DataPoint) float64 {
	if len(data) < 2 {
		return 0
	}

	// Simple linear regression
	n := float64(len(data))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, dp := range data {
		x := float64(i)
		y := dp.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Slope of the regression line
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// Convert to percentage growth per week (assuming 5-minute intervals)
	// 2016 data points = 1 week
	mean := sumY / n
	if mean == 0 {
		return 0
	}

	weeklyGrowth := (slope * 2016 / mean) * 100
	return weeklyGrowth
}

// calculatePredictability measures how predictable the pattern is
func (a *DNAAnalyzer) calculatePredictability(data []types.DataPoint) float64 {
	// Use R-squared from linear regression as a simple predictability measure
	if len(data) < 2 {
		return 0
	}

	mean := calculateMeanFromDataPoints(data)
	
	// Calculate total sum of squares
	ssTot := 0.0
	for _, dp := range data {
		ssTot += (dp.Value - mean) * (dp.Value - mean)
	}

	// Calculate residual sum of squares (using simple moving average as prediction)
	windowSize := 10
	ssRes := 0.0
	for i := windowSize; i < len(data); i++ {
		predicted := 0.0
		for j := i - windowSize; j < i; j++ {
			predicted += data[j].Value
		}
		predicted /= float64(windowSize)
		ssRes += (data[i].Value - predicted) * (data[i].Value - predicted)
	}

	if ssTot == 0 {
		return 0
	}

	rSquared := 1 - (ssRes / ssTot)
	return math.Max(0, math.Min(1, rSquared))
}

// determineCostSensitivity determines cost sensitivity based on volatility and growth
func (a *DNAAnalyzer) determineCostSensitivity(volatility, growthRate float64) string {
	// High volatility or high growth = high cost sensitivity
	if volatility > 0.5 || math.Abs(growthRate) > 10 {
		return "high"
	} else if volatility > 0.3 || math.Abs(growthRate) > 5 {
		return "medium"
	}
	return "low"
}

// determineLatencySensitivity determines latency sensitivity based on traffic patterns
func (a *DNAAnalyzer) determineLatencySensitivity(trafficData []types.DataPoint) string {
	if len(trafficData) == 0 {
		return "medium" // Default
	}

	// High traffic = likely user-facing = high latency sensitivity
	mean := calculateMeanFromDataPoints(trafficData)
	p95 := calculatePercentileFromDataPoints(trafficData, 95)

	if mean > 100 && p95/mean > 2 { // High traffic with spikes
		return "critical"
	} else if mean > 50 {
		return "high"
	} else if mean > 10 {
		return "medium"
	}
	return "low"
}

// analyzeResourceProfile analyzes CPU and memory usage patterns
func (a *DNAAnalyzer) analyzeResourceProfile(cpuData, memData []types.DataPoint) ResourceProfile {
	profile := ResourceProfile{}

	cpuMean := calculateMeanFromDataPoints(cpuData)
	memMean := calculateMeanFromDataPoints(memData)

	// Classify intensity
	profile.CPUIntensity = classifyIntensity(cpuMean, 0.5, 2.0)
	profile.MemoryIntensity = classifyIntensity(memMean/(1024*1024*1024), 0.5, 2.0) // Convert to GB

	// CPU to Memory ratio (cores per GB)
	if memMean > 0 {
		profile.CPUToMemoryRatio = cpuMean / (memMean / (1024 * 1024 * 1024))
	}

	// Burst capability (based on P95/mean ratio)
	cpuP95 := calculatePercentileFromDataPoints(cpuData, 95)
	if cpuMean > 0 {
		burstRatio := cpuP95 / cpuMean
		profile.BurstCapability = math.Min(1.0, (burstRatio-1.0)/2.0) // Normalize to 0-1
	}

	return profile
}

// analyzeTrafficPattern analyzes traffic patterns
func (a *DNAAnalyzer) analyzeTrafficPattern(trafficData []types.DataPoint) TrafficPattern {
	pattern := TrafficPattern{}

	if len(trafficData) == 0 {
		pattern.Type = "unknown"
		return pattern
	}

	mean := calculateMeanFromDataPoints(trafficData)
	stdDev := calculateStdDevFromDataPoints(trafficData, mean)
	peak := calculatePercentileFromDataPoints(trafficData, 99)

	// Classify pattern type
	cv := 0.0
	if mean > 0 {
		cv = stdDev / mean
		pattern.PeakToAvgRatio = peak / mean
	}

	if cv < 0.2 {
		pattern.Type = "steady"
	} else if cv < 0.5 {
		pattern.Type = "periodic"
	} else if cv < 1.0 {
		pattern.Type = "bursty"
	} else {
		pattern.Type = "random"
	}

	// Calculate variations (simplified)
	pattern.DailyVariation = math.Min(1.0, cv)
	pattern.WeeklyVariation = math.Min(1.0, cv*0.8)

	return pattern
}

// Helper functions

func classifyIntensity(value, lowThreshold, highThreshold float64) string {
	if value < lowThreshold {
		return "low"
	} else if value < highThreshold {
		return "medium"
	}
	return "high"
}

func calculatePercentileFromDataPoints(data []types.DataPoint, percentile float64) float64 {
	if len(data) == 0 {
		return 0
	}

	values := make([]float64, len(data))
	for i, dp := range data {
		values[i] = dp.Value
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * percentile / 100.0)
	return sorted[index]
}

// Made with Bob
