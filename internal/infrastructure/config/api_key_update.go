package config

import (
	"fmt"
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// UpdatedGetAPIKey returns the API key for the specified provider, checking both legacy and new formats
func UpdatedGetAPIKey(providerName, configFile string) (string, error) {
	// First try to read as enhanced config
	enhancedCfg, err := LoadEnhancedConfig(configFile)
	if err == nil && enhancedCfg != nil && enhancedCfg.AI != nil {
		// Try to get from interfaces
		if enhancedCfg.AI.Interfaces != nil {
			// Search in each interface
			for _, interfaceConfig := range enhancedCfg.AI.Interfaces {
				if provider, ok := interfaceConfig.Providers[providerName]; ok {
					if provider.APIKey != "" {
						logging.Debug("Found API key for %s in interface-based config", providerName)
						return provider.APIKey, nil
					}
				}
			}
		}
		
		// Try legacy providers section
		if enhancedCfg.AI.Providers != nil {
			if provider, ok := enhancedCfg.AI.Providers[providerName]; ok {
				if provider.APIKey != "" {
					logging.Debug("Found API key for %s in legacy providers section", providerName)
					return provider.APIKey, nil
				}
			}
		}
	}
	
	// If that fails, try the legacy config
	legacyCfg, err := LoadConfig(configFile)
	if err == nil && legacyCfg != nil && legacyCfg.AI != nil && legacyCfg.AI.Providers != nil {
		if provider, ok := legacyCfg.AI.Providers[providerName]; ok {
			if provider.APIKey != "" {
				logging.Debug("Found API key for %s in legacy config", providerName)
				return provider.APIKey, nil
			}
		}
	}
	
	// If still not found, look in environment variables
	switch providerName {
	case "openai":
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			return apiKey, nil
		}
	case "anthropic":
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			return apiKey, nil
		}
	case "gemini":
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			return apiKey, nil
		}
	case "deepseek":
		if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
			return apiKey, nil
		}
	case "openrouter":
		if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
			return apiKey, nil
		}
	}
	
	return "", fmt.Errorf("API key for provider %s not found in configuration or environment variables", providerName)
}

// GetAPIEndpoint returns the API endpoint for the specified provider
func GetAPIEndpoint(providerName, configFile string) (string, error) {
	// First try to read as enhanced config
	enhancedCfg, err := LoadEnhancedConfig(configFile)
	if err == nil && enhancedCfg != nil && enhancedCfg.AI != nil {
		// Try to get from interfaces
		if enhancedCfg.AI.Interfaces != nil {
			// Search in each interface
			for _, interfaceConfig := range enhancedCfg.AI.Interfaces {
				if provider, ok := interfaceConfig.Providers[providerName]; ok {
					if provider.APIEndpoint != "" {
						return provider.APIEndpoint, nil
					}
				}
			}
		}
		
		// Try legacy providers section
		if enhancedCfg.AI.Providers != nil {
			if provider, ok := enhancedCfg.AI.Providers[providerName]; ok {
				if provider.APIEndpoint != "" {
					return provider.APIEndpoint, nil
				}
			}
		}
	}
	
	// If that fails, try the legacy config
	legacyCfg, err := LoadConfig(configFile)
	if err == nil && legacyCfg != nil && legacyCfg.AI != nil && legacyCfg.AI.Providers != nil {
		if provider, ok := legacyCfg.AI.Providers[providerName]; ok {
			if provider.APIEndpoint != "" {
				return provider.APIEndpoint, nil
			}
		}
	}
	
	// Return default endpoints based on provider
	switch providerName {
	case "ollama":
		return "http://localhost:11434", nil
	case "openai":
		return "https://api.openai.com/v1", nil
	case "anthropic":
		return "https://api.anthropic.com", nil
	case "deepseek":
		return "https://api.deepseek.com/v1", nil
	case "openrouter":
		return "https://openrouter.ai/api/v1", nil
	}
	
	return "", fmt.Errorf("API endpoint for provider %s not found in configuration", providerName)
}

// UpdateGetDefaultProvider returns the default provider from the config
func UpdateGetDefaultProvider(configFile string) (string, error) {
	// First try to read as enhanced config
	enhancedCfg, err := LoadEnhancedConfig(configFile)
	if err == nil && enhancedCfg != nil && enhancedCfg.AI != nil && enhancedCfg.AI.DefaultProvider != "" {
		logging.Debug("Found default provider %s in enhanced config", enhancedCfg.AI.DefaultProvider)
		return enhancedCfg.AI.DefaultProvider, nil
	}
	
	// If that fails, try the legacy config
	legacyCfg, err := LoadConfig(configFile)
	if err == nil && legacyCfg != nil && legacyCfg.AI != nil && legacyCfg.AI.DefaultProvider != "" {
		logging.Debug("Found default provider %s in legacy config", legacyCfg.AI.DefaultProvider)
		return legacyCfg.AI.DefaultProvider, nil
	}
	
	// Default to openai
	logging.Debug("No default provider found in config, using openai")
	return "openai", nil
}

// UpdateLoadAllProviders loads all providers from the config file
func UpdateLoadAllProviders(configFile string) (map[string]ProviderConfig, error) {
	result := make(map[string]ProviderConfig)
	
	// First try to read as enhanced config
	enhancedCfg, err := LoadEnhancedConfig(configFile)
	if err == nil && enhancedCfg != nil && enhancedCfg.AI != nil {
		// Get from interfaces
		if enhancedCfg.AI.Interfaces != nil {
			for _, interfaceConfig := range enhancedCfg.AI.Interfaces {
				for name, provider := range interfaceConfig.Providers {
					result[name] = provider
					logging.Debug("Found provider %s in interface-based config", name)
				}
			}
		}
		
		// Get from legacy providers
		if enhancedCfg.AI.Providers != nil {
			for name, provider := range enhancedCfg.AI.Providers {
				result[name] = provider
				logging.Debug("Found provider %s in legacy providers section", name)
			}
		}
		
		if len(result) > 0 {
			return result, nil
		}
	}
	
	// If that fails, try the legacy config
	legacyCfg, err := LoadConfig(configFile)
	if err == nil && legacyCfg != nil && legacyCfg.AI != nil && legacyCfg.AI.Providers != nil {
		for name, provider := range legacyCfg.AI.Providers {
			result[name] = provider
			logging.Debug("Found provider %s in legacy config", name)
		}
		
		if len(result) > 0 {
			return result, nil
		}
	}
	
	return result, fmt.Errorf("no providers found in configuration")
}

// GetInterfaceTypeForProvider returns the interface type for a provider
func GetInterfaceTypeForProvider(providerName, configFile string) (InterfaceType, error) {
	// First try to read as enhanced config
	enhancedCfg, err := LoadEnhancedConfig(configFile)
	if err == nil && enhancedCfg != nil && enhancedCfg.AI != nil && enhancedCfg.AI.Interfaces != nil {
		// Search in each interface
		for interfaceType, interfaceConfig := range enhancedCfg.AI.Interfaces {
			if _, ok := interfaceConfig.Providers[providerName]; ok {
				logging.Debug("Found interface type %s for provider %s in config", interfaceType, providerName)
				return interfaceType, nil
			}
		}
	}
	
	// If not found in config, use default mappings
	switch providerName {
	case "openai", "deepseek", "openrouter", "gemini":
		return OpenAICompatible, nil
	case "anthropic":
		return AnthropicNative, nil
	case "ollama":
		return OllamaNative, nil
	default:
		// Default to OpenAI-compatible
		logging.Warn("Unknown provider %s, defaulting to OpenAI-compatible interface", providerName)
		return OpenAICompatible, nil
	}
}
