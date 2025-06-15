package config

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
)

// Backward compatibility types and functions
// These maintain compatibility with existing code while the transition happens

// ConfigWrapper wraps ApplicationConfig to add backward compatibility methods
type ConfigWrapper struct {
	*domain.ApplicationConfig
}

// Legacy types for backward compatibility
type Config = ConfigWrapper  // Now points to wrapper instead of direct alias
type ServerConfig = domain.ServerConfig
type ProviderConfig = domain.ProviderConfig
type AIConfig = domain.AIConfig
type InterfaceConfig = domain.InterfaceConfig
type ServerSettings = domain.ServerSettings
type SettingsConfig = domain.GlobalSettings
type InterfaceType = domain.InterfaceType
type EnhancedConfig = domain.ApplicationConfig
type EnhancedAIConfig = domain.AIConfig
type EnhancedProviderConfig = domain.ProviderConfig

// Legacy constants for backward compatibility
const (
	OpenAICompatible = domain.OpenAICompatible
	AnthropicNative  = domain.AnthropicNative
	OllamaNative     = domain.OllamaNative
)

// Global service instance for backward compatibility
var defaultService *Service

// init initializes the default service
func init() {
	defaultService = NewService()
}

// Legacy functions for backward compatibility

// LoadConfig loads configuration using the default service
func LoadConfig(file string) (*Config, error) {
	appConfig, err := defaultService.LoadConfig(file)
	if err != nil {
		return nil, err
	}
	return &ConfigWrapper{ApplicationConfig: appConfig}, nil
}

// LoadEnhancedConfig loads enhanced configuration (same as LoadConfig now)
func LoadEnhancedConfig(file string) (*EnhancedConfig, error) {
	return defaultService.LoadConfig(file)
}

// SaveConfig saves configuration using the default service
func SaveConfig(config *Config, file string) error {
	return defaultService.SaveConfig(config.ApplicationConfig, file)
}

// SaveEnhancedConfig saves enhanced configuration (same as SaveConfig now)
func SaveEnhancedConfig(config *EnhancedConfig, file string) error {
	return defaultService.SaveConfig(config, file)
}

// GetConfigPath returns the absolute path to the configuration file
func GetConfigPath(configFile string) (string, error) {
	// This function remains unchanged for backward compatibility
	if configFile[0] == '/' || configFile[1] == ':' { // Unix absolute or Windows absolute
		return configFile, nil
	}
	
	// For relative paths, keep the original behavior
	return configFile, nil
}

// MigrateConfigFile migrates a config file from legacy format to new format
func MigrateConfigFile(file string) error {
	config, err := defaultService.MigrateConfig(file)
	if err != nil {
		return err
	}
	
	return defaultService.SaveConfig(config, file)
}

// GetProviderFromEnhancedConfig retrieves a provider config from the enhanced config
func GetProviderFromEnhancedConfig(config *EnhancedConfig, providerName string) (ProviderConfig, InterfaceType, error) {
	service := NewService()
	service.config = config
	
	providerConfig, interfaceType, err := service.GetProviderConfig(providerName)
	if err != nil {
		return ProviderConfig{}, "", err
	}
	
	return *providerConfig, interfaceType, nil
}

// GetDefaultProviderFromEnhancedConfig gets the default provider from the enhanced config
func GetDefaultProviderFromEnhancedConfig(config *EnhancedConfig) (string, ProviderConfig, InterfaceType, error) {
	service := NewService()
	service.config = config
	
	name, providerConfig, interfaceType, err := service.GetDefaultProvider()
	if err != nil {
		return "", ProviderConfig{}, "", err
	}
	
	return name, *providerConfig, interfaceType, nil
}

// ConvertToEnhancedConfig converts a regular Config to an EnhancedConfig
func ConvertToEnhancedConfig(config *Config) *EnhancedConfig {
	return config.ApplicationConfig
}

// GenerateNewConfig generates a new configuration file with interfaces
func GenerateNewConfig(config *Config) *EnhancedConfig {
	return config.ApplicationConfig
}

// Compatibility wrapper functions that work with Config instances

// GetServerConfigFromConfig returns the configuration for the specified server
func GetServerConfigFromConfig(c *Config, name string) (ServerConfig, error) {
	if c.Servers == nil {
		return ServerConfig{}, domain.ErrServerNotFound.WithDetails("no servers configured")
	}
	
	server, ok := c.Servers[name]
	if !ok {
		return ServerConfig{}, domain.ErrServerNotFound.WithDetails("server: " + name)
	}
	return server, nil
}

// GetProviderConfigFromConfig returns the configuration for the specified AI provider
func GetProviderConfigFromConfig(c *Config, name string) (ProviderConfig, error) {
	service := NewService()
	service.config = c.ApplicationConfig
	
	providerConfig, _, err := service.GetProviderConfig(name)
	if err != nil {
		return ProviderConfig{}, err
	}
	
	return *providerConfig, nil
}

// GetAPIKeyFromConfig returns the API key for the specified provider
func GetAPIKeyFromConfig(c *Config, providerName string) (string, error) {
	providerConfig, err := GetProviderConfigFromConfig(c, providerName)
	if err != nil {
		return "", err
	}
	
	if providerConfig.APIKey == "" {
		return "", domain.ErrProviderAuth.WithDetails("API key not configured for provider: " + providerName)
	}
	
	return providerConfig.APIKey, nil
}

// GetDefaultProviderFromConfig returns the default AI provider from config
func GetDefaultProviderFromConfig(c *Config) string {
	if c.AI == nil || c.AI.DefaultProvider == "" {
		return "openai" // Fallback to openai
	}
	return c.AI.DefaultProvider
}

// GetDefaultModelFromConfig returns the default model for the specified provider
func GetDefaultModelFromConfig(c *Config, providerName string) (string, error) {
	providerConfig, err := GetProviderConfigFromConfig(c, providerName)
	if err != nil {
		return "", err
	}
	
	if providerConfig.DefaultModel == "" {
		return "", domain.ErrModelNotFound.WithDetails("default model not configured for provider: " + providerName)
	}
	
	return providerConfig.DefaultModel, nil
}

// GetSystemPromptFromConfig returns the system prompt for the specified server or the default one
func GetSystemPromptFromConfig(c *Config, serverName string) string {
	service := NewService()
	service.config = c.ApplicationConfig
	return service.GetSystemPrompt(serverName)
}

// GetMaxToolFollowUpFromConfig returns the maximum tool follow-up attempts from configuration
func GetMaxToolFollowUpFromConfig(c *Config, serverName string) int {
	service := NewService()
	service.config = c.ApplicationConfig
	return service.GetMaxToolFollowUp(serverName)
}

// GetSettingsFromConfig returns the global settings from the config
func GetSettingsFromConfig(c *Config) *SettingsConfig {
	return c.Settings
}

// GetServerSettingsFromConfig returns settings for a specific server
func GetServerSettingsFromConfig(c *Config, serverName string) (*ServerSettings, error) {
	serverConfig, err := GetServerConfigFromConfig(c, serverName)
	if err != nil {
		return nil, err
	}
	
	return serverConfig.Settings, nil
}

// METHOD RECEIVERS for ConfigWrapper struct

// GetMaxToolFollowUp returns the maximum tool follow-up attempts from configuration
func (c *ConfigWrapper) GetMaxToolFollowUp(serverName string) int {
	service := NewService()
	service.config = c.ApplicationConfig
	return service.GetMaxToolFollowUp(serverName)
}

// GetSystemPrompt returns the system prompt for the specified server or the default one
func (c *ConfigWrapper) GetSystemPrompt(serverName string) string {
	service := NewService()
	service.config = c.ApplicationConfig
	return service.GetSystemPrompt(serverName)
}

// GetSettings returns the global settings from the config
func (c *ConfigWrapper) GetSettings() *SettingsConfig {
	return c.Settings
}

// GetServerSettings returns settings for a specific server
func (c *ConfigWrapper) GetServerSettings(serverName string) (*ServerSettings, error) {
	if c.Servers == nil {
		return nil, domain.ErrServerNotFound.WithDetails("no servers configured")
	}
	
	server, ok := c.Servers[serverName]
	if !ok {
		return nil, domain.ErrServerNotFound.WithDetails("server: " + serverName)
	}
	
	return server.Settings, nil
}

// GetDefaultProvider returns the default AI provider from config
func (c *ConfigWrapper) GetDefaultProvider() string {
	if c.AI == nil || c.AI.DefaultProvider == "" {
		return "openai" // Fallback to openai
	}
	return c.AI.DefaultProvider
}

// GetDefaultModel returns the default model for the specified provider
func (c *ConfigWrapper) GetDefaultModel(providerName string) (string, error) {
	service := NewService()
	service.config = c.ApplicationConfig
	
	providerConfig, _, err := service.GetProviderConfig(providerName)
	if err != nil {
		return "", err
	}
	
	if providerConfig.DefaultModel == "" {
		return "", domain.ErrModelNotFound.WithDetails("default model not configured for provider: " + providerName)
	}
	
	return providerConfig.DefaultModel, nil
}
