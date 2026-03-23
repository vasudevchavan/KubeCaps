package cli

import "fmt"

// GlobalFlags holds CLI flags shared across all commands.
type GlobalFlags struct {
	Kubeconfig      string
	PrometheusURL   string
	Namespace       string
	OutputFormat    string
	TimeWindowHours int
}

// Validate validates the global flags.
func (f *GlobalFlags) Validate() error {
	if f.PrometheusURL == "" {
		return fmt.Errorf("--prometheus-url is required")
	}
	if f.Namespace == "" {
		return fmt.Errorf("--namespace is required")
	}
	if f.TimeWindowHours <= 0 {
		return fmt.Errorf("--time-window must be positive")
	}
	if f.OutputFormat != "table" && f.OutputFormat != "json" {
		return fmt.Errorf("--output must be 'table' or 'json'")
	}
	return nil
}
