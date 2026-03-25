package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vasudevchavan/kubecaps/internal/ai"
	"github.com/vasudevchavan/kubecaps/internal/k8s"
	"github.com/vasudevchavan/kubecaps/internal/output"
	"github.com/vasudevchavan/kubecaps/internal/predictor"
	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

var (
	aiDetectAnomalies     bool
	aiGenerateDNA         bool
	aiAnalyzeAutoscalers  bool
	aiSensitivity         float64
	aiVerbose             bool
)

var aiCmd = &cobra.Command{
	Use:   "ai [workload-name]",
	Short: "AI-powered analysis of workload behavior",
	Long: `Use AI/ML techniques to analyze workload behavior including:
- Anomaly detection in resource usage
- Workload DNA profiling (behavioral fingerprinting)
- Pattern recognition and seasonality detection
- Autoscaler correlation analysis (HPA/VPA/KEDA)`,
	Args: cobra.ExactArgs(1),
	RunE: runAI,
}

func init() {
	aiCmd.Flags().BoolVar(&aiDetectAnomalies, "detect-anomalies", true, "Detect anomalies in resource usage")
	aiCmd.Flags().BoolVar(&aiGenerateDNA, "generate-dna", true, "Generate workload DNA profile")
	aiCmd.Flags().BoolVar(&aiAnalyzeAutoscalers, "analyze-autoscalers", true, "Analyze HPA/VPA/KEDA correlation with anomalies")
	aiCmd.Flags().Float64Var(&aiSensitivity, "sensitivity", 1.0, "Anomaly detection sensitivity (0.5-2.0)")
	aiCmd.Flags().BoolVarP(&aiVerbose, "verbose", "v", false, "Show detailed calculation values and math")
	rootCmd.AddCommand(aiCmd)
}

func runAI(cmd *cobra.Command, args []string) error {
	if err := globalFlags.Validate(); err != nil {
		return err
	}

	workloadName := args[0]
	ctx := context.Background()

	// Create clients
	k8sClient, err := k8s.NewClient(globalFlags.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes: %w", err)
	}

	promClient, err := prometheus.NewClient(globalFlags.PrometheusURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus: %w", err)
	}

	// Get workload info
	workload, err := k8sClient.GetWorkload(ctx, globalFlags.Namespace, workloadName)
	if err != nil {
		return fmt.Errorf("failed to find workload: %w", err)
	}

	fmt.Printf("🤖 AI Analysis for %s/%s (%s)\n\n", workload.Kind, workload.Name, workload.Namespace)

	// Fetch autoscaler configurations
	var hpaConfig, vpaConfig, kedaConfig interface{}
	if aiAnalyzeAutoscalers {
		hpaConfig, _ = k8sClient.GetHPAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
		vpaConfig, _ = k8sClient.GetVPAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
		kedaConfig, _ = k8sClient.GetKEDAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
	}

	// Fetch metrics
	end := time.Now()
	start := end.Add(-time.Duration(globalFlags.TimeWindowHours) * time.Hour)
	step := time.Minute * 5

	queries := prometheus.NewQueries()

	cpuQ := queries.CPUUsageTotal(workload.Namespace, workload.Name)
	cpuData, err := promClient.QueryRange(ctx, cpuQ, start, end, step)
	if err != nil {
		return fmt.Errorf("failed to fetch CPU metrics: %w", err)
	}

	memQ := queries.MemoryUsageTotal(workload.Namespace, workload.Name)
	memData, err := promClient.QueryRange(ctx, memQ, start, end, step)
	if err != nil {
		return fmt.Errorf("failed to fetch memory metrics: %w", err)
	}

	rpsQ := queries.RequestRate(workload.Namespace, workload.Name)
	trafficData, _ := promClient.QueryRange(ctx, rpsQ, start, end, step)

	// Anomaly Detection
	if aiDetectAnomalies {
		fmt.Println("🔍 Anomaly Detection")
		fmt.Println(output.Separator())

		detector := ai.NewAnomalyDetector(aiSensitivity)

		// Detect CPU anomalies
		cpuAnomalies := detector.DetectAnomalies(cpuData)
		if len(cpuAnomalies) > 0 {
			fmt.Printf("\n📊 CPU Anomalies Detected: %d\n\n", len(cpuAnomalies))
			printAnomalies(cpuAnomalies, "CPU")
		} else {
			fmt.Println("\n✅ No CPU anomalies detected")
		}

		// Detect Memory anomalies
		memAnomalies := detector.DetectAnomalies(memData)
		if len(memAnomalies) > 0 {
			fmt.Printf("\n📊 Memory Anomalies Detected: %d\n\n", len(memAnomalies))
			printAnomalies(memAnomalies, "Memory")
		} else {
			fmt.Println("\n✅ No memory anomalies detected")
		}

		// Pattern anomalies
		patternAnomalies := detector.DetectPatternAnomalies(cpuData, 12) // 1-hour windows
		if len(patternAnomalies) > 0 {
			fmt.Printf("\n📈 Pattern Shifts Detected: %d\n\n", len(patternAnomalies))
			printAnomalies(patternAnomalies, "Pattern")
		}

		fmt.Println()
	}

	// Workload DNA
	if aiGenerateDNA {
		fmt.Println("🧬 Workload DNA Profile")
		fmt.Println(output.Separator())

		analyzer := ai.NewDNAAnalyzer()
		dna, err := analyzer.AnalyzeDNA(workload.Name, workload.Namespace, cpuData, memData, trafficData)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not generate DNA profile: %v\n", err)
		} else {
			if globalFlags.OutputFormat == "json" {
				data, _ := json.MarshalIndent(dna, "", "  ")
				fmt.Println(string(data))
			} else {
				printDNA(dna)
			}
		}
	}

	// Load config
	config, _ := globalFlags.LoadConfig()

	// Create prediction engine
	engine := predictor.NewEngine(promClient, config)
	workloadTypes := types.WorkloadInfo{
		Name:      workload.Name,
		Namespace: workload.Namespace,
		Kind:      workload.Kind,
		Replicas:  workload.Replicas,
	}
	kubecapsRec, _ := engine.PredictResources(ctx, workloadTypes, globalFlags.TimeWindowHours)

	// Autoscaler Correlation Analysis
	if aiAnalyzeAutoscalers && (hpaConfig != nil || vpaConfig != nil || kedaConfig != nil) {
		fmt.Println("⚙️  Autoscaler Correlation Analysis")
		fmt.Println(output.Separator())
		
		printAutoscalerAnalysis(hpaConfig, vpaConfig, kedaConfig, cpuData, memData, config, kubecapsRec)
		fmt.Println()
	}

	return nil
}

func formatMemoryMBi(bytes float64) string {
	const MBi = 1024 * 1024
	return fmt.Sprintf("%.1f MBi", bytes/MBi)
}

func printAutoscalerAnalysis(hpaConfig, vpaConfig, kedaConfig interface{}, cpuData, memData []types.DataPoint, config types.Config, kubecapsRec *types.Recommendation) {
	label := output.NewLabelPrinter()
	value := output.NewValuePrinter()
	
	fmt.Println()

	if aiVerbose && kubecapsRec != nil {
		label.Println("  🧮 KubeCaps Optimization Math (Verbose):")
		fmt.Println()
		opt := config.Optimization
		calc := kubecapsRec.Calculation

		label.Print("     CPU Strategy:        ")
		value.Printf("Baseline(P%.0f: %.3f cores) → Optimized for risk up to Peak(%.3f)*%.1fx\n", 
			opt.TargetCPUPercentile*100, calc.CPUBaseline, calc.CPUForecastPeak, calc.CPUBuffer*opt.ScalingSensitivity)
		label.Print("     CPU Risk Penalty:    ")
		value.Printf("%.1f (Lambda) * %.1f (SLA) = %.1f (Total Penalty)\n", 
			opt.CPURiskPenalty, opt.LatencySensitivity, calc.CPUPenalty)
		
		label.Print("     Mem Strategy:        ")
		value.Printf("Baseline(P%.0f: %s) → Optimized for risk up to Peak(%s)*%.1fx\n", 
			opt.TargetMemoryPercentile*100, formatMemoryMBi(calc.MemBaseline), formatMemoryMBi(calc.MemForecastPeak), calc.MemBuffer*opt.ScalingSensitivity)
		label.Print("     Mem Risk Penalty:    ")
		value.Printf("%.1f (Lambda) = %.1f (Total Penalty)\n", 
			opt.MemoryRiskPenalty, calc.MemPenalty)

		if calc.CPUForecastPeak > 0 {
			label.Print("     Confidence:          ")
			value.Printf("%.0f%% (based on %d data points)\n", kubecapsRec.Confidence*100, len(cpuData))
		}
		fmt.Println()
	}
	
	// HPA Analysis
	if hpa, ok := hpaConfig.(*types.HPAConfig); ok && hpa.Found {
		label.Println("  📊 HPA Configuration Detected:")
		fmt.Println()
		label.Print("     Name:                ")
		value.Printf("%s\n", hpa.Name)
		label.Print("     Replicas:            ")
		value.Printf("Min: %d, Max: %d, Current: %d\n", hpa.MinReplicas, hpa.MaxReplicas, hpa.CurrentReplicas)
		
		if len(hpa.Metrics) > 0 {
			label.Print("     Metrics:             ")
			for i, m := range hpa.Metrics {
				if i > 0 {
					fmt.Print("                          ")
				}
				value.Printf("%s (%s): %.0f%%\n", m.Name, m.Type, m.Target)
			}
		}
		
		// AI Insight: Check if HPA targets align with actual usage
		if len(cpuData) > 0 {
			avgCPU := calculateAverage(cpuData)
			for _, m := range hpa.Metrics {
				if m.Name == "cpu" && m.TargetType == "Utilization" {
					deviation := ((avgCPU - m.Target) / m.Target) * 100
					fmt.Println()
					label.Print("     AI Insight:          ")
					if deviation > 20 {
						value.Printf("⚠️  Actual CPU (%.0f%%) is %.0f%% above target - HPA may scale frequently\n", avgCPU, deviation)
					} else if deviation < -20 {
						value.Printf("💡 Actual CPU (%.0f%%) is %.0f%% below target - Consider lowering target\n", avgCPU, -deviation)
					} else {
						value.Printf("✅ HPA target well-aligned with actual usage (%.0f%%)\n", avgCPU)
					}
				}
			}
		}
		fmt.Println()
	}
	
	// VPA Analysis
	if vpa, ok := vpaConfig.(*types.VPAConfig); ok && vpa.Found {
		label.Println("  📊 VPA Configuration Detected:")
		fmt.Println()
		label.Print("     Name:                ")
		value.Printf("%s\n", vpa.Name)
		label.Print("     Update Mode:         ")
		value.Printf("%s\n", vpa.UpdateMode)
		
		if len(vpa.Recommendations) > 0 {
			fmt.Println()
			label.Println("     Recommendations:")
			for _, rec := range vpa.Recommendations {
				fmt.Printf("       Container: %s\n", rec.ContainerName)
				label.Print("         CPU Target:      ")
				value.Printf("%.3f cores\n", rec.TargetCPU)
				label.Print("         Memory Target:   ")
				// Display in MBi/GBi for better clarity (binary units used in K8s)
				memBytes := float64(rec.TargetMemory)

				const (
					KBi = 1024.0
					MBi = KBi * 1024
					GBi = MBi * 1024
				)

				if memBytes >= GBi {
					value.Printf("%.2f GBi\n", memBytes/GBi)
				} else {
					value.Printf("%.1f MBi\n", memBytes/MBi)
				}
			}
		}
		
		// AI Insight: Compare VPA recommendations with actual usage
		if len(cpuData) > 0 && len(vpa.Recommendations) > 0 {
			avgCPU := calculateAverage(cpuData)
			rec := vpa.Recommendations[0]
			deviation := ((avgCPU - rec.TargetCPU) / rec.TargetCPU) * 100
			fmt.Println()
			label.Print("     AI Insight:          ")
			if deviation > 30 {
				value.Printf("⚠️  Actual usage exceeds VPA recommendation by %.0f%% - Risk of throttling\n", deviation)
			} else if deviation < -30 {
				value.Printf("💡 VPA recommendation %.0f%% higher than actual - Over-provisioned\n", -deviation)
			} else {
				value.Printf("✅ VPA recommendations align well with actual usage\n")
			}
		}
		fmt.Println()
	}
	
	// KEDA Analysis
	if keda, ok := kedaConfig.(*types.KEDAConfig); ok && keda.Found {
		label.Println("  📊 KEDA Configuration Detected:")
		fmt.Println()
		label.Print("     Name:                ")
		value.Printf("%s\n", keda.Name)
		label.Print("     Replicas:            ")
		value.Printf("Min: %d, Max: %d\n", keda.MinReplicas, keda.MaxReplicas)
		label.Print("     Polling Interval:    ")
		value.Printf("%ds\n", keda.PollingInterval)
		label.Print("     Cooldown Period:     ")
		value.Printf("%ds\n", keda.CooldownPeriod)
		
		if len(keda.Triggers) > 0 {
			fmt.Println()
			label.Println("     Triggers:")
			for _, t := range keda.Triggers {
				fmt.Printf("       - Type: %s\n", t.Type)
				if len(t.Metadata) > 0 {
					for k, v := range t.Metadata {
						fmt.Printf("         %s: %s\n", k, v)
					}
				}
			}
		}
		
		// AI Insight: KEDA polling efficiency
		fmt.Println()
		label.Print("     AI Insight:          ")
		if keda.PollingInterval < 10 {
			value.Printf("⚠️  Very frequent polling (%ds) - May cause API throttling\n", keda.PollingInterval)
		} else if keda.PollingInterval > 60 {
			value.Printf("💡 Slow polling (%ds) - May miss rapid traffic changes\n", keda.PollingInterval)
		} else {
			value.Printf("✅ Polling interval (%ds) is well-balanced\n", keda.PollingInterval)
		}
		fmt.Println()
	}
	
	// Overall AI Recommendation
	if hpaConfig == nil && vpaConfig == nil && kedaConfig == nil {
		fmt.Println()
		label.Print("  💡 Recommendation:   ")
		value.Println("No autoscalers detected - Consider implementing HPA or KEDA for dynamic scaling")
		fmt.Println()
	}
}

func calculateAverage(data []types.DataPoint) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, d := range data {
		sum += d.Value
	}
	return sum / float64(len(data))
}

func printAnomalies(anomalies []ai.Anomaly, metricType string) {
	// Show top 10 most severe anomalies
	count := len(anomalies)
	if count > 10 {
		count = 10
	}

	for i := 0; i < count; i++ {
		a := anomalies[i]
		severityColor := output.GetSeverityColor(a.Severity)
		
		fmt.Printf("  %s [%s] %s\n", 
			severityColor.Sprint("●"), 
			a.Timestamp, 
			a.Description)
		fmt.Printf("    Value: %.2f | Expected: %.2f | Deviation: %.2f | Score: %.2f\n",
			a.Value, a.Expected, a.Deviation, a.Score)
		fmt.Println()
	}

	if len(anomalies) > 10 {
		fmt.Printf("  ... and %d more anomalies\n", len(anomalies)-10)
	}
}

func printDNA(dna *ai.WorkloadDNA) {
	label := output.NewLabelPrinter()
	value := output.NewValuePrinter()

	fmt.Println()
	label.Print("  Workload:        ")
	value.Printf("%s/%s\n", dna.WorkloadName, dna.Namespace)

	fmt.Println()
	label.Println("  📊 Behavioral Characteristics:")
	fmt.Println()

	label.Print("     Seasonality:         ")
	value.Printf("%v\n", dna.Seasonality)

	label.Print("     Volatility:          ")
	printVolatility(dna.Volatility)

	label.Print("     Growth Rate:         ")
	printGrowthRate(dna.GrowthRate)

	label.Print("     Predictability:      ")
	printPredictability(dna.Predictability)

	label.Print("     Cost Sensitivity:    ")
	printSensitivity(dna.CostSensitivity)

	label.Print("     Latency Sensitivity: ")
	printSensitivity(dna.LatencySensitivity)

	fmt.Println()
	label.Println("  💻 Resource Profile:")
	fmt.Println()

	label.Print("     CPU Intensity:       ")
	value.Printf("%s\n", dna.ResourceProfile.CPUIntensity)

	label.Print("     Memory Intensity:    ")
	value.Printf("%s\n", dna.ResourceProfile.MemoryIntensity)

	label.Print("     CPU/Memory Ratio:    ")
	value.Printf("%.2f cores/GB\n", dna.ResourceProfile.CPUToMemoryRatio)

	label.Print("     Burst Capability:    ")
	printCapability(dna.ResourceProfile.BurstCapability)

	fmt.Println()
	label.Println("  🌐 Traffic Pattern:")
	fmt.Println()

	label.Print("     Type:                ")
	value.Printf("%s\n", dna.TrafficPattern.Type)

	label.Print("     Peak/Avg Ratio:      ")
	value.Printf("%.2fx\n", dna.TrafficPattern.PeakToAvgRatio)

	label.Print("     Daily Variation:     ")
	printVariation(dna.TrafficPattern.DailyVariation)

	label.Print("     Weekly Variation:    ")
	printVariation(dna.TrafficPattern.WeeklyVariation)

	fmt.Println()
	label.Println("  📈 Key Metrics:")
	fmt.Println()

	if cpuMean, ok := dna.Characteristics["cpu_mean"]; ok {
		label.Print("     CPU Mean:            ")
		value.Printf("%.3f cores\n", cpuMean)
	}
	if cpuP95, ok := dna.Characteristics["cpu_p95"]; ok {
		label.Print("     CPU P95:             ")
		value.Printf("%.3f cores\n", cpuP95)
	}
	if memMean, ok := dna.Characteristics["mem_mean"]; ok {
		label.Print("     Memory Mean:         ")
		memBytes := memMean
		const (
			KBi = 1024.0
			MBi = KBi * 1024
			GBi = MBi * 1024
		)
		if memBytes >= GBi {
			value.Printf("%.2f GBi\n", memBytes/GBi)
		} else {
			value.Printf("%.1f MBi\n", memBytes/MBi)
		}
	}
	if trafficMean, ok := dna.Characteristics["traffic_mean"]; ok {
		label.Print("     Traffic Mean:        ")
		value.Printf("%.2f RPS\n", trafficMean)
	}

	fmt.Println()
}

func printVolatility(v float64) {
	value := output.NewValuePrinter()
	if v < 0.2 {
		value.Printf("%.2f (Low - Stable)\n", v)
	} else if v < 0.5 {
		value.Printf("%.2f (Medium - Moderate)\n", v)
	} else {
		value.Printf("%.2f (High - Volatile)\n", v)
	}
}

func printGrowthRate(g float64) {
	value := output.NewValuePrinter()
	if g > 5 {
		value.Printf("+%.1f%%/week (Growing)\n", g)
	} else if g < -5 {
		value.Printf("%.1f%%/week (Declining)\n", g)
	} else {
		value.Printf("%.1f%%/week (Stable)\n", g)
	}
}

func printPredictability(p float64) {
	value := output.NewValuePrinter()
	percentage := p * 100
	if p > 0.7 {
		value.Printf("%.0f%% (Highly Predictable)\n", percentage)
	} else if p > 0.4 {
		value.Printf("%.0f%% (Moderately Predictable)\n", percentage)
	} else {
		value.Printf("%.0f%% (Unpredictable)\n", percentage)
	}
}

func printSensitivity(s string) {
	value := output.NewValuePrinter()
	value.Printf("%s\n", s)
}

func printCapability(c float64) {
	value := output.NewValuePrinter()
	percentage := c * 100
	if c > 0.7 {
		value.Printf("%.0f%% (High)\n", percentage)
	} else if c > 0.3 {
		value.Printf("%.0f%% (Medium)\n", percentage)
	} else {
		value.Printf("%.0f%% (Low)\n", percentage)
	}
}

func printVariation(v float64) {
	value := output.NewValuePrinter()
	percentage := v * 100
	value.Printf("%.0f%%\n", percentage)
}

// Made with Bob
