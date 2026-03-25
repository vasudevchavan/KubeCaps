package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// GlobalFlags holds all global CLI flags
type GlobalFlags struct {
	Kubeconfig      string
	PrometheusURL   string
	Namespace       string
	OutputFormat    string
	TimeWindowHours int
	ConfigPath      string
}

// Validate validates all global flags
func (f *GlobalFlags) Validate() error {
	// Validate Prometheus URL
	if f.PrometheusURL == "" {
		return fmt.Errorf("--prometheus-url is required")
	}
	
	if _, err := url.Parse(f.PrometheusURL); err != nil {
		return fmt.Errorf("invalid prometheus URL: %w", err)
	}
	
	// Validate namespace
	if f.Namespace == "" {
		return fmt.Errorf("--namespace cannot be empty")
	}
	
	// Validate output format
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"yaml":  true,
	}
	if !validFormats[f.OutputFormat] {
		return fmt.Errorf("invalid output format '%s': must be one of: table, json, yaml", f.OutputFormat)
	}
	
	// Validate time window
	if f.TimeWindowHours < 1 {
		return fmt.Errorf("time window must be at least 1 hour, got %d", f.TimeWindowHours)
	}
	if f.TimeWindowHours > 720 { // 30 days
		return fmt.Errorf("time window cannot exceed 720 hours (30 days), got %d", f.TimeWindowHours)
	}
	
	// Validate kubeconfig if provided
	if f.Kubeconfig != "" {
		if _, err := os.Stat(f.Kubeconfig); err != nil {
			return fmt.Errorf("kubeconfig file not found: %w", err)
		}
	}
	
	return nil
}

// ValidateWorkloadName validates a workload name
func ValidateWorkloadName(name string) error {
	if name == "" {
		return fmt.Errorf("workload name cannot be empty")
	}
	
	// Kubernetes name validation (simplified)
	if len(name) > 253 {
		return fmt.Errorf("workload name too long: %d characters (max 253)", len(name))
	}
	
	return nil
}

// ResolveKubeconfig resolves the kubeconfig path using the following priority:
// 1. Explicit flag value
// 2. KUBECONFIG environment variable
// 3. Default ~/.kube/config
// 4. Empty string (will use in-cluster config)
func ResolveKubeconfig(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		return envPath
	}
	
	home, err := os.UserHomeDir()
	if err == nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath
		}
	}
	
	return ""
}

// ValidatePrometheusConnection validates that Prometheus is reachable
// This is a placeholder for future implementation
func ValidatePrometheusConnection(promURL string) error {
	// TODO: Implement actual connection check
	// For now, just validate URL format
	_, err := url.Parse(promURL)
	return err
}

// Made with Bob
