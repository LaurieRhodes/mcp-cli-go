package runas

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles loading RunAs configurations
type Loader struct{}

// NewLoader creates a new RunAs config loader
func NewLoader() *Loader {
	return &Loader{}
}

// Load loads a RunAs configuration from a YAML file
func (l *Loader) Load(path string) (*RunAsConfig, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read runas config file: %w", err)
	}

	// Parse YAML
	var config RunAsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse runas config YAML: %w", err)
	}

	// Expand environment variables in proxy config
	l.expandEnvVars(&config)

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid runas config: %w", err)
	}

	return &config, nil
}

// expandEnvVars expands environment variables in the config
func (l *Loader) expandEnvVars(config *RunAsConfig) {
	// Expand in proxy config if present
	if config.ProxyConfig != nil {
		config.ProxyConfig.APIKey = expandEnvVar(config.ProxyConfig.APIKey)
	}
}

// expandEnvVar expands environment variables in a string
// Supports ${VAR_NAME} and $VAR_NAME formats
// Only expands if the string looks like an environment variable reference
func expandEnvVar(s string) string {
	// Check if string contains environment variable patterns
	hasEnvPattern := strings.Contains(s, "${") ||
		(strings.Contains(s, "$") && len(s) > 1 && (s[0] == '$' || strings.Contains(s, " $")))

	if !hasEnvPattern {
		return s
	}

	return os.ExpandEnv(s)
}

// LoadOrDefault loads a config or returns a default example
func (l *Loader) LoadOrDefault(path string) (*RunAsConfig, bool, error) {
	// Try to load
	config, err := l.Load(path)
	if err == nil {
		return config, false, nil
	}

	// If file doesn't exist, create example
	if os.IsNotExist(err) {
		example := l.CreateExample()
		if saveErr := l.Save(example, path); saveErr != nil {
			return nil, false, fmt.Errorf("failed to create example config: %w", saveErr)
		}
		return example, true, nil
	}

	// Other error
	return nil, false, err
}

// Save saves a RunAs configuration to a YAML file
func (l *Loader) Save(config *RunAsConfig, path string) error {
	// Validate before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateExample creates an example RunAs configuration
func (l *Loader) CreateExample() *RunAsConfig {
	return &RunAsConfig{
		RunAsType: RunAsTypeMCP,
		Version:   "1.0",
		ServerInfo: ServerInfo{
			Name:        "example_agent",
			Version:     "1.0.0",
			Description: "Example MCP server exposing workflow templates",
		},
		Tools: []ToolExposure{
			{
				Template:    "simple_analysis",
				Name:        "analyze",
				Description: "Analyzes input data using AI",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{
							"type":        "string",
							"description": "Data to analyze",
						},
					},
					"required": []interface{}{"data"},
				},
				InputMapping: map[string]string{
					"data": "{{input_data}}",
				},
			},
		},
	}
}
