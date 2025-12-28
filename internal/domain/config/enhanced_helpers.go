package config

import (
	"fmt"
)

// LoadEnhancedConfig loads the enhanced modular configuration
func LoadEnhancedConfig(configFile string) (*ApplicationConfig, error) {
	loader := NewLoader()
	return loader.Load(configFile)
}

// GetProviderFromEnhancedConfig retrieves a specific provider's config
func GetProviderFromEnhancedConfig(cfg *ApplicationConfig, providerName string) (*ProviderConfig, InterfaceType, error) {
	if cfg == nil || cfg.AI == nil || cfg.AI.Interfaces == nil {
		return nil, "", fmt.Errorf("invalid configuration")
	}

	// Search all interfaces for the provider
	for interfaceType, interfaceConfig := range cfg.AI.Interfaces {
		if provider, exists := interfaceConfig.Providers[providerName]; exists {
			return &provider, interfaceType, nil
		}
	}

	return nil, "", fmt.Errorf("provider '%s' not found", providerName)
}

// GetDefaultProviderFromEnhancedConfig retrieves the default provider's config
func GetDefaultProviderFromEnhancedConfig(cfg *ApplicationConfig) (string, *ProviderConfig, InterfaceType, error) {
	if cfg == nil || cfg.AI == nil {
		return "", nil, "", fmt.Errorf("invalid configuration")
	}

	defaultProvider := cfg.AI.DefaultProvider
	if defaultProvider == "" {
		return "", nil, "", fmt.Errorf("no default provider configured")
	}

	providerConfig, interfaceType, err := GetProviderFromEnhancedConfig(cfg, defaultProvider)
	if err != nil {
		return "", nil, "", err
	}

	return defaultProvider, providerConfig, interfaceType, nil
}
