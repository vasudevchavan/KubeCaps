package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vasudevchavan/kubecaps/internal/k8s"
	"github.com/vasudevchavan/kubecaps/internal/output"
	"github.com/vasudevchavan/kubecaps/internal/predictor"
	"github.com/vasudevchavan/kubecaps/internal/prometheus"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [workload-name]",
	Short: "Analyze a workload and predict optimal resource configuration",
	Long: `Analyze a Kubernetes workload by examining its historical resource usage
from Prometheus metrics and predict optimal CPU/Memory requests and limits
using statistical models (Linear Regression, EWMA, Percentile Analysis).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	if err := globalFlags.Validate(); err != nil {
		return err
	}

	ctx := context.Background()

	// Create Kubernetes client
	k8sClient, err := k8s.NewClient(globalFlags.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes: %w", err)
	}

	// Create Prometheus client
	promClient, err := prometheus.NewClient(globalFlags.PrometheusURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus: %w", err)
	}

	// Create prediction engine
	engine := predictor.NewEngine(promClient)

	// If a specific workload is provided, analyze just that one
	if len(args) > 0 {
		workload, err := k8sClient.GetWorkload(ctx, globalFlags.Namespace, args[0])
		if err != nil {
			return fmt.Errorf("failed to find workload: %w", err)
		}

		rec, err := engine.PredictResources(ctx, *workload, globalFlags.TimeWindowHours)
		if err != nil {
			return fmt.Errorf("prediction failed: %w", err)
		}

		if globalFlags.OutputFormat == "json" {
			data, _ := json.MarshalIndent(rec, "", "  ")
			fmt.Println(string(data))
		} else {
			output.PrintRecommendation(rec)
		}
		return nil
	}

	// Analyze all workloads in the namespace
	workloads, err := k8sClient.ListAllWorkloads(ctx, globalFlags.Namespace)
	if err != nil {
		return fmt.Errorf("failed to list workloads: %w", err)
	}

	if len(workloads) == 0 {
		fmt.Printf("No workloads found in namespace %s\n", globalFlags.Namespace)
		return nil
	}

	fmt.Printf("🔍 Analyzing %d workloads in namespace '%s' (time window: %dh)...\n\n",
		len(workloads), globalFlags.Namespace, globalFlags.TimeWindowHours)

	for _, w := range workloads {
		rec, err := engine.PredictResources(ctx, w, globalFlags.TimeWindowHours)
		if err != nil {
			fmt.Printf("⚠️  Skipping %s/%s: %v\n\n", w.Kind, w.Name, err)
			continue
		}

		if globalFlags.OutputFormat == "json" {
			data, _ := json.MarshalIndent(rec, "", "  ")
			fmt.Println(string(data))
		} else {
			output.PrintRecommendation(rec)
		}
	}
	return nil
}
