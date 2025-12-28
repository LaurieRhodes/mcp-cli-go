package host

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// EnhancedAIOptions contains the options for the AI provider with interface type
type EnhancedAIOptions struct {
	// Provider name (openai, anthropic, ollama, etc.)
	Provider string

	// Interface type (openai_compatible, anthropic_native, ollama_native)
	Interface config.InterfaceType

	// Model name
	Model string

	// API key (for providers that require it)
	APIKey string

	// API endpoint
	APIEndpoint string

	// Available models for this provider
	AvailableModels []string
}

// GetEnhancedAIOptions loads AI options from the enhanced config file and command line flags
func GetEnhancedAIOptions(configFile, cmdLineProvider, cmdLineModel string) (*EnhancedAIOptions, error) {
	// Default options
	options := &EnhancedAIOptions{
		Provider: cmdLineProvider,
		Model:    cmdLineModel,
	}

	// Try to load from enhanced config file
	enhancedCfg, err := config.LoadEnhancedConfig(configFile)
	if err != nil {
		logging.Warn("Failed to load enhanced config file: %v", err)
		logging.Info("Using default AI options")

		// Determine the interface type based on provider
		if cmdLineProvider != "" {
			switch cmdLineProvider {
			case "openai", "deepseek", "openrouter", "gemini":
				options.Interface = config.OpenAICompatible
			case "anthropic":
				options.Interface = config.AnthropicNative
			case "ollama":
				options.Interface = config.OllamaNative
			default:
				options.Interface = config.OpenAICompatible // Default to OpenAI-compatible
			}
		}

		return options, nil
	}

	// If provider not specified on command line, try to get default from config
	if cmdLineProvider == "" {
		provider, providerConfig, interfaceType, err := config.GetDefaultProviderFromEnhancedConfig(enhancedCfg)
		if err != nil {
			logging.Warn("Failed to get default provider from config: %v", err)
		} else {
			options.Provider = provider
			options.Interface = interfaceType
			options.APIKey = providerConfig.APIKey
			options.APIEndpoint = providerConfig.APIEndpoint

			// FIXED: Use configured model, don't override if it exists
			if cmdLineModel == "" && providerConfig.DefaultModel != "" {
				options.Model = providerConfig.DefaultModel
				logging.Info("Using configured default model for %s: %s", options.Provider, options.Model)
			}

			logging.Debug("Using provider from config: %s with interface: %s", options.Provider, options.Interface)
		}
	} else {
		// Provider specified on command line, get the config for it
		providerConfig, interfaceType, err := config.GetProviderFromEnhancedConfig(enhancedCfg, cmdLineProvider)
		if err != nil {
			logging.Warn("Provider %s not found in config: %v", cmdLineProvider, err)

			// Determine the interface type based on provider
			switch cmdLineProvider {
			case "openai", "deepseek", "openrouter", "gemini":
				options.Interface = config.OpenAICompatible
			case "anthropic":
				options.Interface = config.AnthropicNative
			case "ollama":
				options.Interface = config.OllamaNative
			default:
				options.Interface = config.OpenAICompatible // Default to OpenAI-compatible
			}
		} else {
			options.Interface = interfaceType
			options.APIKey = providerConfig.APIKey
			options.APIEndpoint = providerConfig.APIEndpoint

			// FIXED: Use configured model, don't override if it exists
			if cmdLineModel == "" && providerConfig.DefaultModel != "" {
				options.Model = providerConfig.DefaultModel
				logging.Info("Using configured default model for %s: %s", options.Provider, options.Model)
			}

			logging.Debug("Using provider config for %s with interface %s", options.Provider, options.Interface)
		}
	}

	// Set default values if still empty
	if options.Provider == "" {
		options.Provider = "openai"
		options.Interface = config.OpenAICompatible
		logging.Warn("No provider specified, defaulting to openai")
	}

	// FIXED: Only use emergency fallbacks if no model is configured at all
	if options.Model == "" {
		logging.Warn("No model specified in config or command line for provider %s, using emergency fallback", options.Provider)
		options.Model = getEnhancedEmergencyFallbackModel(options.Provider)
	}

	// Set default API endpoint for Ollama if not specified
	if options.Provider == "ollama" && options.APIEndpoint == "" {
		options.APIEndpoint = "http://localhost:11434"
	}

	// Important: Log the final model choice so we can confirm it's working
	logging.Info("Final model selection: %s for provider %s", options.Model, options.Provider)

	return options, nil
}

// getEnhancedEmergencyFallbackModel returns emergency fallback models ONLY when config is missing
// This should rarely be used - configuration should always specify models
func getEnhancedEmergencyFallbackModel(provider string) string {
	logging.Warn("Using emergency fallback model for provider %s - configuration should be updated", provider)

	switch provider {
	case "openai":
		return "gpt-4o" // Changed from gpt-4o-mini - use better default
	case "anthropic":
		return "claude-3-sonnet-20240229"
	case "deepseek":
		return "deepseek-chat"
	case "ollama":
		return "llama3.1:8b" // Simple fallback for ollama
	case "openrouter":
		return "phi-3-medium-128k-instruct:free"
	case "gemini":
		return "gemini-1.5-pro"
	default:
		return "gpt-4o" // Changed from gpt-4o-mini - use better default
	}
}

// ProcessEnhancedOptions processes the server options for enhanced commands
// Renamed from ProcessOptions to avoid conflict
func ProcessEnhancedOptions(serverName string, disableFilesystem bool, providerName, modelName string) ([]string, map[string]bool) {
	// Return the server names and user specified map
	var serverNames []string
	userSpecified := make(map[string]bool)

	// Add the specified server
	if serverName != "" {
		serverNames = append(serverNames, serverName)
		userSpecified[serverName] = true
	}

	// If we don't have a server and filesystem is not disabled, add filesystem
	if len(serverNames) == 0 && !disableFilesystem {
		serverNames = append(serverNames, "filesystem")
		userSpecified["filesystem"] = false
	}

	return serverNames, userSpecified
}
