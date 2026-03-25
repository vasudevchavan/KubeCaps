package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vasudevchavan/kubecaps/internal/evaluator"
	"github.com/vasudevchavan/kubecaps/internal/k8s"
	"github.com/vasudevchavan/kubecaps/internal/output"
	"github.com/vasudevchavan/kubecaps/internal/predictor"
	"github.com/vasudevchavan/kubecaps/internal/prometheus"
	"github.com/vasudevchavan/kubecaps/pkg/types"
)

var (
	evaluateAutoscaling bool
	compareCurrent      bool
	showMetrics         bool
	explain             bool
)

var evaluateCmd = &cobra.Command{
	Use:   "evaluate [workload-name]",
	Short: "Evaluate autoscaling configuration for a workload",
	Long: `Evaluate HPA, VPA, and KEDA autoscaling configurations for a workload
by validating them against real-time Prometheus metrics. Produces scored
insights, risk flags, and optimization recommendations.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEvaluate,
}

func init() {
	evaluateCmd.Flags().BoolVar(&evaluateAutoscaling, "evaluate-autoscaling", true, "Run full autoscaling evaluation")
	evaluateCmd.Flags().BoolVar(&compareCurrent, "compare-current", false, "Show current vs recommended config table")
	evaluateCmd.Flags().BoolVar(&showMetrics, "show-metrics", false, "Include raw Prometheus metric evidence")
	evaluateCmd.Flags().BoolVar(&explain, "explain", false, "Verbose explainability output")
	rootCmd.AddCommand(evaluateCmd)
}

func runEvaluate(cmd *cobra.Command, args []string) error {
	if err := globalFlags.Validate(); err != nil {
		return err
	}

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

	// If a specific workload is provided
	if len(args) > 0 {
		workload, err := k8sClient.GetWorkload(ctx, globalFlags.Namespace, args[0])
		if err != nil {
			return fmt.Errorf("failed to find workload: %w", err)
		}

		result, err := evaluateWorkload(ctx, k8sClient, promClient, *workload)
		if err != nil {
			return err
		}

		outputResult(result)
		return nil
	}

	// Evaluate all workloads in namespace
	workloads, err := k8sClient.ListAllWorkloads(ctx, globalFlags.Namespace)
	if err != nil {
		return fmt.Errorf("failed to list workloads: %w", err)
	}

	if len(workloads) == 0 {
		fmt.Printf("No workloads found in namespace %s\n", globalFlags.Namespace)
		return nil
	}

	fmt.Printf("🔍 Evaluating %d workloads in namespace '%s' (time window: %dh)...\n\n",
		len(workloads), globalFlags.Namespace, globalFlags.TimeWindowHours)

	for _, w := range workloads {
		result, err := evaluateWorkload(ctx, k8sClient, promClient, w)
		if err != nil {
			fmt.Printf("⚠️  Skipping %s/%s: %v\n\n", w.Kind, w.Name, err)
			continue
		}

		outputResult(result)
	}

	return nil
}

func evaluateWorkload(ctx context.Context, k8sClient *k8s.Client, promClient *prometheus.Client, workload types.WorkloadInfo) (*types.EvaluationResult, error) {
	result := &types.EvaluationResult{
		Workload:   workload,
		TimeWindow: fmt.Sprintf("%dh", globalFlags.TimeWindowHours),
		AnalyzedAt: time.Now(),
	}

	// Detect autoscaling configs
	hpaConfig, err := k8sClient.GetHPAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
	if err != nil {
		return nil, fmt.Errorf("HPA detection failed: %w", err)
	}
	result.HPAConfig = hpaConfig

	vpaConfig, err := k8sClient.GetVPAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
	if err != nil {
		return nil, fmt.Errorf("VPA detection failed: %w", err)
	}
	result.VPAConfig = vpaConfig

	kedaConfig, err := k8sClient.GetKEDAForWorkload(ctx, workload.Namespace, workload.Name, workload.Kind)
	if err != nil {
		return nil, fmt.Errorf("KEDA detection failed: %w", err)
	}
	result.KEDAConfig = kedaConfig

	// Run evaluators with proper error handling
	hpaEval := evaluator.NewHPAEvaluator(promClient)
	hpaScore, hpaInsights, hpaRisks, err := hpaEval.Evaluate(ctx, workload, hpaConfig, globalFlags.TimeWindowHours)
	if err != nil {
		fmt.Printf("⚠️  Warning: HPA evaluation failed: %v\n", err)
		// Add a warning insight instead of failing completely
		result.Insights = append(result.Insights, types.Insight{
			Severity:    "warning",
			Category:    "evaluation",
			Title:       "HPA Evaluation Failed",
			Description: fmt.Sprintf("Could not complete HPA evaluation: %v", err),
		})
	} else {
		result.Scores = append(result.Scores, hpaScore)
		result.Insights = append(result.Insights, hpaInsights...)
		result.Risks = append(result.Risks, hpaRisks...)
	}

	vpaEval := evaluator.NewVPAEvaluator(promClient)
	vpaScore, vpaInsights, vpaRisks, err := vpaEval.Evaluate(ctx, workload, vpaConfig, globalFlags.TimeWindowHours)
	if err != nil {
		fmt.Printf("⚠️  Warning: VPA evaluation failed: %v\n", err)
		result.Insights = append(result.Insights, types.Insight{
			Severity:    "warning",
			Category:    "evaluation",
			Title:       "VPA Evaluation Failed",
			Description: fmt.Sprintf("Could not complete VPA evaluation: %v", err),
		})
	} else {
		result.Scores = append(result.Scores, vpaScore)
		result.Insights = append(result.Insights, vpaInsights...)
		result.Risks = append(result.Risks, vpaRisks...)
	}

	kedaEval := evaluator.NewKEDAEvaluator(promClient)
	kedaScore, kedaInsights, kedaRisks, err := kedaEval.Evaluate(ctx, workload, kedaConfig, globalFlags.TimeWindowHours)
	if err != nil {
		fmt.Printf("⚠️  Warning: KEDA evaluation failed: %v\n", err)
		result.Insights = append(result.Insights, types.Insight{
			Severity:    "warning",
			Category:    "evaluation",
			Title:       "KEDA Evaluation Failed",
			Description: fmt.Sprintf("Could not complete KEDA evaluation: %v", err),
		})
	} else {
		result.Scores = append(result.Scores, kedaScore)
		result.Insights = append(result.Insights, kedaInsights...)
		result.Risks = append(result.Risks, kedaRisks...)
	}

	// Run correlation analysis
	corrAnalyzer := evaluator.NewCorrelationAnalyzer(promClient)
	_, corrInsights, err := corrAnalyzer.Analyze(ctx, workload, globalFlags.TimeWindowHours)
	if err != nil {
		fmt.Printf("⚠️  Warning: Correlation analysis failed: %v\n", err)
	} else {
		result.Insights = append(result.Insights, corrInsights...)
	}

	// Run risk detection
	riskDetector := evaluator.NewRiskDetector(promClient)
	additionalRisks, err := riskDetector.DetectRisks(ctx, workload, globalFlags.TimeWindowHours)
	if err != nil {
		fmt.Printf("⚠️  Warning: Risk detection failed: %v\n", err)
	} else {
		result.Risks = append(result.Risks, additionalRisks...)
	}

	// Calculate overall score
	scorer := evaluator.NewScorer()
	result.OverallScore = scorer.CalculateOverallScore(result.Scores)

	// Load config
	config, err := globalFlags.LoadConfig()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load config, using defaults: %v\n", err)
	}

	// Get resource recommendation
	engine := predictor.NewEngine(promClient, config)
	rec, err := engine.PredictResources(ctx, workload, globalFlags.TimeWindowHours)
	if err == nil {
		result.Recommendation = rec
	}

	return result, nil
}

func outputResult(result *types.EvaluationResult) {
	if globalFlags.OutputFormat == "json" {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		output.PrintEvaluationResult(result)
	}
}
