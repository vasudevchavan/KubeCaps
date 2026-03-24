package predictor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/internal/recommender"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

// Engine is the main prediction engine that aggregates multiple prediction models.
type Engine struct {
	promClient *prometheus.Client
	queries    *prometheus.Queries
}

// NewEngine creates a new prediction engine.
func NewEngine(promClient *prometheus.Client) *Engine {
	return &Engine{
		promClient: promClient,
		queries:    prometheus.NewQueries(),
	}
}

// PredictResources predicts optimal resource configuration for a workload.
func (e *Engine) PredictResources(ctx context.Context, workload types.WorkloadInfo, timeWindowHours int) (*types.Recommendation, error) {
	end := time.Now()
	step := time.Minute * 5

	// Fetch CPU usage
	var cpuData, memData []types.DataPoint
	var effectiveWindowHours int

	// Evaluate time windows based on data availability: 1 Month (720h), 1 Week (168h), 1 Day (24h)
	// If timeWindowHours is explicitly set (not 24 default or if 24 is set we still try this logic if it's for recommendation?)
	// Actually, just try exactly in this order:
	windowsToTry := []int{720, 168, 24}
	// If user specifically requested a window other than the windows, we can just use it.
	// But let's just override and use the windows logic to find max available up to the requested one, or just hardcode the 3 tiers.
	if timeWindowHours != 24 && timeWindowHours != 720 && timeWindowHours != 168 {
		windowsToTry = []int{timeWindowHours}
	}

	for _, w := range windowsToTry {
		start := end.Add(-time.Duration(w) * time.Hour)
		cpuQ := e.queries.CPUUsageTotal(workload.Namespace, workload.Name)
		cpuD, errC := e.promClient.QueryRange(ctx, cpuQ, start, end, step)
		
		memQ := e.queries.MemoryUsageTotal(workload.Namespace, workload.Name)
		memD, errM := e.promClient.QueryRange(ctx, memQ, start, end, step)

		if errC == nil && errM == nil && len(cpuD) > 0 && len(memD) > 0 {
			// Calculate actual data duration available
			actualDuration := end.Sub(cpuD[0].Timestamp).Hours()
			if w == 720 && actualDuration >= 168 { // We have more than a week, consider it 1 Month tier
				cpuData, memData = cpuD, memD
				effectiveWindowHours = 720
				break
			} else if w == 168 && actualDuration >= 24 { // We have more than a day, consider it 1 Week tier
				cpuData, memData = cpuD, memD
				effectiveWindowHours = 168
				break
			} else if w == 24 {
				cpuData, memData = cpuD, memD
				effectiveWindowHours = 24
				break
			}
			// If we asked for 720 but only got 2 days, we will loop to 168.
			// If we asked for 168 but only got 12 hours, we will loop to 24.
		}
	}

	if len(cpuData) == 0 && len(memData) == 0 {
		return nil, fmt.Errorf("no metric data found for workload %s/%s", workload.Namespace, workload.Name)
	}

	// Fetch current requests/limits
	cpuReq, _ := e.promClient.QueryInstant(ctx, e.queries.CPURequest(workload.Namespace, workload.Name))
	cpuLim, _ := e.promClient.QueryInstant(ctx, e.queries.CPULimit(workload.Namespace, workload.Name))
	memReq, _ := e.promClient.QueryInstant(ctx, e.queries.MemoryRequest(workload.Namespace, workload.Name))
	memLim, _ := e.promClient.QueryInstant(ctx, e.queries.MemoryLimit(workload.Namespace, workload.Name))

	// Fetch OOM and Throttling metrics for Feedback Loop
	oomQ := e.queries.OOMEvents(workload.Namespace, workload.Name)
	oomVal, _ := e.promClient.QueryInstant(ctx, oomQ)
	hasOOM := oomVal > 0

	throttleQ := e.queries.CPUThrottling(workload.Namespace, workload.Name)
	throttleVal, _ := e.promClient.QueryInstant(ctx, throttleQ)
	hasThrottling := throttleVal > 5.0 // > 5% throttling is significant

	// Fetch RPS for classification using effective window
	startEffective := end.Add(-time.Duration(effectiveWindowHours) * time.Hour)
	rpsQ := e.queries.RequestRate(workload.Namespace, workload.Name)
	rpsData, _ := e.promClient.QueryRange(ctx, rpsQ, startEffective, end, step)

	// Run prediction models
	cpuPrediction := e.predictMetric(cpuData)
	memPrediction := e.predictMetric(memData)
	
	// Classify workload behavior
	workloadType := e.ClassifyWorkload(cpuPrediction, extractValues(rpsData))

	// Forecast Peak usage for next 24h (288 5-minute points)
	cpuForecastPeak := HoltWintersPeak(cpuPrediction.values, 288)
	memForecastPeak := HoltWintersPeak(memPrediction.values, 288)

	// Optimization Layer: Cost vs Risk
	// Define risk penalties. Heavily amplify if Prometheus feedback loop detected issues.
	lambdaCPU := 10.0
	lambdaMem := 50.0
	if hasThrottling {
		lambdaCPU = 100.0 // Heavy penalty for CPU risk
	}
	if hasOOM {
		lambdaMem = 500.0 // Huge penalty for OOM risk
	}

	// Optimize Request: Grid search from P50 to 1.2 * Forecast Peak
	// CPU price weight = 1.0/core, Mem price weight = 1.0/GB
	cpuRequestRaw := optimizeRequest(cpuPrediction.values, cpuPrediction.p50, cpuForecastPeak*1.2, 1.0, lambdaCPU)
	memRequestRaw := optimizeRequest(memPrediction.values, memPrediction.p50, memForecastPeak*1.2, 1.0/(1024*1024*1024), lambdaMem)

	// Limits
	cpuLimitRaw := math.Max(cpuPrediction.p99, cpuForecastPeak) * 1.2
	memLimitRaw := math.Max(memPrediction.p99, memForecastPeak) * 1.2

	// Per-replica values
	replicas := float64(workload.Replicas)
	if replicas < 1 {
		replicas = 1
	}
	cpuRequestRaw /= replicas
	cpuLimitRaw /= replicas
	memRequestRaw /= replicas
	memLimitRaw /= replicas

	// Calculate confidence based on data quality
	confidence := e.calculateConfidence(cpuData, memData, cpuPrediction, memPrediction)

	// Build explanation
	explanation := e.buildExplanation(cpuReq, cpuLim, memReq, memLim, cpuRequestRaw, cpuLimitRaw, memRequestRaw, memLimitRaw)

	// Calculate diff percentages
	var cpuReqDiff, cpuLimDiff, memReqDiff, memLimDiff float64
	if cpuReq > 0 {
		cpuReqDiff = ((cpuRequestRaw - cpuReq) / cpuReq) * 100
	}
	if cpuLim > 0 {
		cpuLimDiff = ((cpuLimitRaw - cpuLim) / cpuLim) * 100
	}
	if memReq > 0 {
		memReqDiff = ((memRequestRaw - memReq) / memReq) * 100
	}
	if memLim > 0 {
		memLimDiff = ((memLimitRaw - memLim) / memLim) * 100
	}

	// Recommender generation
	hpaRec := recommender.GenerateHPA(workloadType, cpuPrediction.p50, cpuForecastPeak, workload.Replicas, "80%")
	vpaRec := recommender.GenerateVPA(cpuRequestRaw, memRequestRaw, false, false) 
	kedaRec := recommender.GenerateKEDA(workloadType, false, false, 0)

	// Conflict Coordinator
	archInsights := recommender.ResolveConflicts(&hpaRec, &vpaRec, &kedaRec)
	
	// Format Insights
	insights := archInsights
	if cpuReqDiff < -20 {
		insights = append(insights, fmt.Sprintf("CPU request can be reduced by %.0f%% (over-provisioned).", math.Abs(cpuReqDiff)))
	}
	if memReqDiff < -20 {
		insights = append(insights, fmt.Sprintf("Memory request can be reduced by %.0f%% (over-provisioned).", math.Abs(memReqDiff)))
	}

	// Calculate Risk
	riskLevel := "low"
	var riskNotes []string
	if hasOOM {
		riskLevel = "high"
		riskNotes = append(riskNotes, "Critical OOMs detected: Do not constrain memory scaling.")
	}
	if hasThrottling {
		if riskLevel != "high" {
			riskLevel = "medium"
		}
		riskNotes = append(riskNotes, "CPU Throttling detected: Scaling up limits rapidly.")
	}
	if len(riskNotes) == 0 {
		riskNotes = append(riskNotes, "Safe to apply gradually")
	}

	return &types.Recommendation{
		WorkloadName:    workload.Name,
		WorkloadKind:    workload.Kind,
		Namespace:       workload.Namespace,
		Type:            workloadType,
		Recommendations: types.AutoscalingRecommendations{
			HPA:  hpaRec,
			VPA:  vpaRec,
			KEDA: kedaRec,
		},
		Insights:        insights,
		Risk: types.RiskProfile{
			Level: riskLevel,
			Notes: riskNotes,
		},
		CPURequest:      formatCPU(cpuRequestRaw),
		CPULimit:        formatCPU(cpuLimitRaw),
		MemoryRequest:   formatMemory(memRequestRaw),
		MemoryLimit:     formatMemory(memLimitRaw),
		CPURequestRaw:   cpuRequestRaw,
		CPULimitRaw:     cpuLimitRaw,
		MemRequestRaw:   memRequestRaw,
		MemLimitRaw:     memLimitRaw,

		CurrentCPURequestRaw: cpuReq,
		CurrentCPULimitRaw:   cpuLim,
		CurrentMemRequestRaw: memReq,
		CurrentMemLimitRaw:   memLim,

		CPUReqDiffPercent:   cpuReqDiff,
		CPULimitDiffPercent: cpuLimDiff,
		MemReqDiffPercent:   memReqDiff,
		MemLimitDiffPercent: memLimDiff,

		Confidence:      confidence,
		Explanation:     explanation,
		TimeWindowHours: effectiveWindowHours,
	}, nil
}

// metricPrediction holds aggregated prediction from all models.
type metricPrediction struct {
	ewma       float64
	linearPeak float64
	p50        float64
	p95        float64
	p99        float64
	trend      float64 // slope direction
	variance   float64
	values     []float64 // the filtered raw values
}

// predictMetric runs all prediction models on a time series.
func (e *Engine) predictMetric(data []types.DataPoint) metricPrediction {
	values := extractValues(data)
	if len(values) == 0 {
		return metricPrediction{}
	}

	// Layer 1: Robust Statistical Baseline (Filter outliers)
	filteredValues := filterOutliersIQR(values)
	if len(filteredValues) < 4 {
		filteredValues = values // Fallback if too few points
	}

	return metricPrediction{
		ewma:       EWMA(filteredValues, 0.3),
		linearPeak: LinearRegressionPeak(filteredValues),
		p50:        Percentile(filteredValues, 50),
		p95:        Percentile(filteredValues, 95),
		p99:        Percentile(filteredValues, 99),
		trend:      LinearRegressionSlope(filteredValues),
		variance:   variance(filteredValues),
		values:     filteredValues,
	}
}

// calculateConfidence calculates the confidence level (0-1) of the prediction.
func (e *Engine) calculateConfidence(cpuData, memData []types.DataPoint, cpuPred, memPred metricPrediction) float64 {
	confidence := 1.0

	// Deduct for insufficient data points
	dataPoints := math.Min(float64(len(cpuData)), float64(len(memData)))
	if dataPoints < 10 {
		confidence -= 0.4
	} else if dataPoints < 50 {
		confidence -= 0.2
	} else if dataPoints < 100 {
		confidence -= 0.1
	}

	// Deduct for high variance (unstable workloads)
	if cpuPred.variance > 0.5 {
		confidence -= 0.15
	}
	if memPred.variance > 0.5 {
		confidence -= 0.15
	}

	// Deduct if models disagree significantly
	cpuDisagreement := math.Abs(cpuPred.ewma-cpuPred.p95) / math.Max(cpuPred.p95, 0.001)
	if cpuDisagreement > 0.5 {
		confidence -= 0.1
	}

	return math.Max(confidence, 0.1)
}

// buildExplanation creates a human-readable explanation of the recommendation.
func (e *Engine) buildExplanation(cpuReq, cpuLim, memReq, memLim, recCPUReq, recCPULim, recMemReq, recMemLim float64) string {
	explanation := "Prediction models used: IQR Filtering, Holt-Winters Peak Forecasting, Log-Normal Probabilistic Optimization.\n"

	if cpuReq > 0 {
		cpuReqDelta := ((recCPUReq - cpuReq) / cpuReq) * 100
		if cpuReqDelta < -10 {
			explanation += fmt.Sprintf("CPU request can be reduced by %.0f%% (over-provisioned).\n", -cpuReqDelta)
		} else if cpuReqDelta > 10 {
			explanation += fmt.Sprintf("CPU request should be increased by %.0f%% (under-provisioned).\n", cpuReqDelta)
		}
	} else {
		explanation += "CPU request is not currently set.\n"
	}

	if memReq > 0 {
		memReqDelta := ((recMemReq - memReq) / memReq) * 100
		if memReqDelta < -10 {
			explanation += fmt.Sprintf("Memory request can be reduced by %.0f%% (over-provisioned).\n", -memReqDelta)
		} else if memReqDelta > 10 {
			explanation += fmt.Sprintf("Memory request should be increased by %.0f%% (under-provisioned).\n", memReqDelta)
		}
	} else {
		explanation += "Memory request is not currently set.\n"
	}

	return explanation
}

// extractValues extracts float64 values from DataPoints.
func extractValues(data []types.DataPoint) []float64 {
	values := make([]float64, len(data))
	for i, dp := range data {
		values[i] = dp.Value
	}
	return values
}

// variance calculates the coefficient of variation (normalized variance).
func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))
	if mean == 0 {
		return 0
	}

	sumSqDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSqDiff += diff * diff
	}
	stddev := math.Sqrt(sumSqDiff / float64(len(values)))
	return stddev / mean // coefficient of variation
}

// ClassifyWorkload determines the behavioral type of the workload.
func (e *Engine) ClassifyWorkload(cpuPred metricPrediction, rpsValues []float64) types.WorkloadType {
	cv := cpuPred.variance // Coefficient of Variation
	
	hasTraffic := false
	for _, r := range rpsValues {
		if r > 0.1 {
			hasTraffic = true
			break
		}
	}

	// 1. If variance is very low, it's Steady
	if cv < 0.20 {
		return types.WorkloadTypeSteady
	}

	// 2. If it has traffic and moderate/high variance, it's Elastic
	if hasTraffic && cv >= 0.20 && cv < 0.80 {
		return types.WorkloadTypeElastic
	}

	// 3. If variance is extremely high, it's Bursty
	if cv >= 0.80 {
		return types.WorkloadTypeBursty
	}

	// Default fallback if no traffic but high variance (maybe event-driven, but we lack queue metrics here)
	return types.WorkloadTypeBursty
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
