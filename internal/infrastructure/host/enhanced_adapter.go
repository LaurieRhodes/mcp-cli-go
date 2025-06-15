package host

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// EnhancedAIOptionsAdapter adapts EnhancedAIOptions to AIOptions
func EnhancedAIOptionsAdapter(enhancedOptions *EnhancedAIOptions) *AIOptions {
	return &AIOptions{
		Provider:    enhancedOptions.Provider,
		Model:       enhancedOptions.Model,
		APIKey:      enhancedOptions.APIKey,
		APIEndpoint: enhancedOptions.APIEndpoint,
	}
}

// GetProviderInterface returns the interface type for a given provider
func GetProviderInterface(provider string, configFile string) config.InterfaceType {
	// Try to determine from the config file first
	cfg, err := config.LoadConfig(configFile)
	if err == nil && cfg != nil && cfg.AI != nil && cfg.AI.Interfaces != nil {
		// Look for the provider in each interface
		for interfaceType, interfaceConfig := range cfg.AI.Interfaces {
			if _, ok := interfaceConfig.Providers[provider]; ok {
				logging.Debug("Found provider %s in interface %s from config", provider, interfaceType)
				return interfaceType
			}
		}
	}
	
	// If not found in config, use default mappings
	switch provider {
	case "openai", "deepseek", "openrouter", "gemini":
		return config.OpenAICompatible
	case "anthropic":
		return config.AnthropicNative
	case "ollama":
		return config.OllamaNative
	default:
		// Default to OpenAI-compatible
		logging.Warn("Unknown provider %s, defaulting to OpenAI-compatible interface", provider)
		return config.OpenAICompatible
	}
}

// ProcessOptionsEnhanced processes the server options for enhanced commands
// This version ensures the interface type is properly set for the provider
func ProcessOptionsEnhanced(configFile string, serverName string, disableFilesystem bool, providerName, modelName string) ([]string, map[string]bool, config.InterfaceType) {
	// First get the basic server names and user specified map
	serverNames, userSpecified := ProcessOptions(serverName, disableFilesystem, providerName, modelName)
	
	// Now determine the interface type
	var interfaceType config.InterfaceType
	
	// If provider name is specified, get its interface
	if providerName != "" {
		interfaceType = GetProviderInterface(providerName, configFile)
	} else {
		// Try to get the default provider from config and then its interface
		cfg, err := config.LoadConfig(configFile)
		if err == nil && cfg != nil {
			defaultProvider := config.GetDefaultProviderFromConfig(cfg)
			if defaultProvider != "" {
				providerName = defaultProvider
				logging.Debug("Using default provider from config: %s", providerName)
				interfaceType = GetProviderInterface(defaultProvider, configFile)
			}
		}
	}
	
	// If no interface type determined yet, use a reasonable default
	if interfaceType == "" {
		interfaceType = config.OpenAICompatible
	}
	
	return serverNames, userSpecified, interfaceType
}
