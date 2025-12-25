package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServerConfig represents a configuration for a server
type ServerConfig struct {
	// Command is the executable to run
	Command string `yaml:"command"`
	
	// Args are the command-line arguments to pass to the command
	Args []string `yaml:"args"`
	
	// Env is the environment variables to set for the process
	Env map[string]string `yaml:"env,omitempty"`
	
	// SystemPrompt is an optional custom system prompt to use when this server is selected
	SystemPrompt string `yaml:"system_prompt,omitempty"`
	
	// Settings contains server-specific settings
	Settings *ServerSettings `yaml:"settings,omitempty"`
}

// ProviderConfig represents a configuration for an AI provider
type ProviderConfig struct {
	// APIKey is the API key for the provider
	APIKey string `yaml:"api_key"`
	
	// DefaultModel is the default model to use for the provider
	DefaultModel string `yaml:"default_model"`
	
	// APIEndpoint is an optional endpoint URL (used for Ollama, etc.)
	APIEndpoint string `yaml:"api_endpoint,omitempty"`
	
	// AvailableModels is a list of available models for this provider
	AvailableModels []string `yaml:"available_models,omitempty"`
	
	// TimeoutSeconds is the timeout for API calls
	TimeoutSeconds int `yaml:"timeout_seconds,omitempty"`
	
	// MaxRetries is the maximum number of retries for API calls
	MaxRetries int `yaml:"max_retries,omitempty"`
}

// AIConfig represents the AI configuration
type AIConfig struct {
	// DefaultProvider is the default AI provider to use
	DefaultProvider string `yaml:"default_provider"`
	
	// Providers maps provider names to their configurations
	Providers map[string]ProviderConfig `yaml:"providers"`
	
	// DefaultSystemPrompt is an optional system prompt to use by default
	DefaultSystemPrompt string `yaml:"default_system_prompt,omitempty"`
}

// Config represents the application configuration
type Config struct {
	// Servers maps server names to their configurations
	Servers map[string]ServerConfig `yaml:"servers"`
	
	// AI contains the AI-related configuration
	AI *AIConfig `yaml:"ai,omitempty"`
	
	// Settings contains global application settings
	Settings *SettingsConfig `yaml:"settings,omitempty"`
}

// LoadConfig loads the configuration from the specified file (supports both JSON and YAML)
func LoadConfig(file string) (*Config, error) {
	// Check if the file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", file)
	}

	// Read the file
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	
	// Detect format by file extension
	ext := strings.ToLower(filepath.Ext(file))
	
	if ext == ".yaml" || ext == ".yml" {
		// Parse as YAML
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	} else {
		// Parse as JSON
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	}

	// Ensure servers is initialized
	if config.Servers == nil {
		config.Servers = make(map[string]ServerConfig)
	}

	return &config, nil
}

// GetServerConfig returns the configuration for the specified server
func (c *Config) GetServerConfig(name string) (ServerConfig, error) {
	server, ok := c.Servers[name]
	if !ok {
		return ServerConfig{}, fmt.Errorf("server %s not found in configuration", name)
	}
	return server, nil
}

// GetProviderConfig returns the configuration for the specified AI provider
func (c *Config) GetProviderConfig(name string) (ProviderConfig, error) {
	if c.AI == nil || c.AI.Providers == nil {
		return ProviderConfig{}, fmt.Errorf("AI providers configuration not found")
	}
	
	provider, ok := c.AI.Providers[name]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("AI provider %s not found in configuration", name)
	}
	
	return provider, nil
}

// GetAPIKey returns the API key for the specified provider
func (c *Config) GetAPIKey(providerName string) (string, error) {
	// If no AI config, return empty
	if c.AI == nil || c.AI.Providers == nil {
		return "", fmt.Errorf("AI providers configuration not found")
	}
	
	// Get the provider config
	provider, ok := c.AI.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("AI provider %s not found in configuration", providerName)
	}
	
	// Check if API key is set
	if provider.APIKey == "" {
		return "", fmt.Errorf("API key for provider %s is not set", providerName)
	}
	
	return provider.APIKey, nil
}

// GetDefaultProvider returns the default AI provider from config
func (c *Config) GetDefaultProvider() string {
	if c.AI == nil || c.AI.DefaultProvider == "" {
		return "openai" // Fallback to openai
	}
	return c.AI.DefaultProvider
}

// GetDefaultModel returns the default model for the specified provider
func (c *Config) GetDefaultModel(providerName string) (string, error) {
	// If no AI config, return error
	if c.AI == nil || c.AI.Providers == nil {
		return "", fmt.Errorf("AI providers configuration not found")
	}
	
	// Get the provider config
	provider, ok := c.AI.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("AI provider %s not found in configuration", providerName)
	}
	
	// FIXED: Return the configured model, don't use hardcoded defaults
	if provider.DefaultModel == "" {
		return "", fmt.Errorf("default model for provider %s is not configured", providerName)
	}
	
	return provider.DefaultModel, nil
}

// GetSystemPrompt returns the system prompt for the specified server or the default one
func (c *Config) GetSystemPrompt(serverName string) string {
	// Check if we have a specific server prompt
	if serverName != "" {
		if server, ok := c.Servers[serverName]; ok && server.SystemPrompt != "" {
			return server.SystemPrompt
		}
	}
	
	// Check for default AI system prompt
	if c.AI != nil && c.AI.DefaultSystemPrompt != "" {
		return c.AI.DefaultSystemPrompt
	}
	
	// Return empty string if not found
	return ""
}

// GetMaxToolFollowUp returns the maximum tool follow-up attempts from configuration
func (c *Config) GetMaxToolFollowUp(serverName string) int {
	// First check for server-specific setting
	if serverName != "" {
		if server, ok := c.Servers[serverName]; ok && server.Settings != nil {
			if maxFollowUp := server.Settings.GetMaxToolFollowUp(); maxFollowUp > 0 {
				return maxFollowUp
			}
		}
	}
	
	// Fall back to global setting
	if c.Settings != nil {
		if maxFollowUp := c.Settings.GetMaxToolFollowUp(); maxFollowUp > 0 {
			return maxFollowUp
		}
	}
	
	// Default fallback
	return 2
}

// GetConfigPath returns the absolute path to the configuration file
func GetConfigPath(configFile string) (string, error) {
	// If the path is absolute, use it directly
	if filepath.IsAbs(configFile) {
		return configFile, nil
	}

	// Otherwise, make it relative to the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(wd, configFile), nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *Config, file string) error {
	// Marshal the JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the file
	if err := os.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
