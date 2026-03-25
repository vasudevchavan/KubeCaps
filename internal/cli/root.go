package cli

import (
	"github.com/spf13/cobra"
)

var (
	globalFlags GlobalFlags
)

var rootCmd = &cobra.Command{
	Use:   "kubecaps",
	Short: "KubeCaps — AI-Powered Kubernetes Resource & Autoscaling Advisor",
	Long: `KubeCaps is a CLI tool that predicts optimal resource configurations,
evaluates existing HPA/VPA/KEDA setups, and validates them against real
Prometheus metrics. It provides actionable insights and risk detection
to optimize your Kubernetes autoscaling.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalFlags.Kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: auto-detect)")
	rootCmd.PersistentFlags().StringVar(&globalFlags.PrometheusURL, "prometheus-url", "", "Prometheus server URL (required)")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.Namespace, "namespace", "n", "default", "Kubernetes namespace to analyze")
	rootCmd.PersistentFlags().StringVarP(&globalFlags.OutputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().IntVar(&globalFlags.TimeWindowHours, "time-window", 24, "Time window in hours for metric analysis")
	rootCmd.PersistentFlags().StringVar(&globalFlags.ConfigPath, "config", "", "Path to KubeCaps config file (YAML)")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
