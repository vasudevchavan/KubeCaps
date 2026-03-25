package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobalFlagsValidate(t *testing.T) {
	tests := []struct {
		name    string
		flags   GlobalFlags
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid flags",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "default",
				OutputFormat:    "table",
				TimeWindowHours: 24,
			},
			wantErr: false,
		},
		{
			name: "missing prometheus URL",
			flags: GlobalFlags{
				Namespace:       "default",
				OutputFormat:    "table",
				TimeWindowHours: 24,
			},
			wantErr: true,
			errMsg:  "prometheus-url is required",
		},
		{
			name: "invalid prometheus URL",
			flags: GlobalFlags{
				PrometheusURL:   "://invalid-url-no-scheme",
				Namespace:       "default",
				OutputFormat:    "table",
				TimeWindowHours: 24,
			},
			wantErr: true,
		},
		{
			name: "empty namespace",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "",
				OutputFormat:    "table",
				TimeWindowHours: 24,
			},
			wantErr: true,
			errMsg:  "namespace cannot be empty",
		},
		{
			name: "invalid output format",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "default",
				OutputFormat:    "invalid",
				TimeWindowHours: 24,
			},
			wantErr: true,
			errMsg:  "invalid output format",
		},
		{
			name: "time window too small",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "default",
				OutputFormat:    "table",
				TimeWindowHours: 0,
			},
			wantErr: true,
			errMsg:  "time window must be at least 1 hour",
		},
		{
			name: "time window too large",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "default",
				OutputFormat:    "table",
				TimeWindowHours: 1000,
			},
			wantErr: true,
			errMsg:  "time window cannot exceed 720 hours",
		},
		{
			name: "valid json output",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "production",
				OutputFormat:    "json",
				TimeWindowHours: 168,
			},
			wantErr: false,
		},
		{
			name: "valid yaml output",
			flags: GlobalFlags{
				PrometheusURL:   "http://prometheus:9090",
				Namespace:       "staging",
				OutputFormat:    "yaml",
				TimeWindowHours: 72,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if err.Error() != tt.errMsg && len(tt.errMsg) > 0 {
					// Check if error message contains expected substring
					contains := false
					for i := 0; i <= len(err.Error())-len(tt.errMsg); i++ {
						if err.Error()[i:i+len(tt.errMsg)] == tt.errMsg {
							contains = true
							break
						}
					}
					if !contains {
						t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestValidateWorkloadName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid name",
			input:   "my-deployment",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "very long name",
			input:   string(make([]byte, 300)),
			wantErr: true,
		},
		{
			name:    "name with numbers",
			input:   "app-123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkloadName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkloadName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveKubeconfig(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", originalEnv)

	tests := []struct {
		name      string
		flagValue string
		envValue  string
		setup     func() string
		cleanup   func()
		wantEmpty bool
	}{
		{
			name:      "explicit flag takes precedence",
			flagValue: "/explicit/path",
			envValue:  "/env/path",
			setup:     func() string { return "" },
			cleanup:   func() {},
			wantEmpty: false,
		},
		{
			name:      "env variable when no flag",
			flagValue: "",
			envValue:  "/env/path",
			setup:     func() string { return "" },
			cleanup:   func() {},
			wantEmpty: false,
		},
		{
			name:      "empty when no flag or env",
			flagValue: "",
			envValue:  "",
			setup: func() string {
				// Create a temp kubeconfig file
				tmpDir := t.TempDir()
				kubeconfigPath := filepath.Join(tmpDir, ".kube", "config")
				os.MkdirAll(filepath.Dir(kubeconfigPath), 0755)
				os.WriteFile(kubeconfigPath, []byte("test"), 0644)
				return tmpDir
			},
			cleanup:   func() {},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			os.Setenv("KUBECONFIG", tt.envValue)
			result := ResolveKubeconfig(tt.flagValue)

			if tt.wantEmpty && result != "" {
				t.Errorf("ResolveKubeconfig() = %v, want empty string", result)
			}
			if !tt.wantEmpty && tt.flagValue != "" && result != tt.flagValue {
				t.Errorf("ResolveKubeconfig() = %v, want %v", result, tt.flagValue)
			}
		})
	}
}

// Made with Bob
