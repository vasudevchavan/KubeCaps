package cli

import (
	"fmt"
	"os"

	"github.com/vasudevchavan/kubecaps/pkg/types"
	"sigs.k8s.io/yaml"
)

// LoadConfig loads the configuration from the provided path.
// If path is empty, it returns the default configuration.
func (f *GlobalFlags) LoadConfig() (types.Config, error) {
	config := types.DefaultConfig()

	if f.ConfigPath == "" {
		return config, nil
	}

	data, err := os.ReadFile(f.ConfigPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
