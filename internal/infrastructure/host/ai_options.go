package host

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// AIOptions contains the options for the AI provider
type AIOptions struct {
	// Provider name (openai, anthropic, ollama)
	Provider string
	
	// Model name
	Model string
	
	// API key (for OpenAI, Anthropic)
	APIKey string
	
	// API endpoint (for Ollama)
	APIEndpoint string
}

// GetAIOptions loads AI options from the config file and command line flags
func GetAIOptions(configFile, cmdLineProvider, cmdLineModel string) (*AIOptions, error) {
	// Default options
	options := &AIOptions{
		Provider: cmdLineProvider,
		Model:    cmdLineModel,
	}
	
	// Try to get default provider if not specified on command line
	if cmdLineProvider == "" {
		defaultProvider, err := config.UpdateGetDefaultProvider(configFile)
		if err != nil {
			logging.Warn("Failed to get default provider: %v", err)
		} else {
			options.Provider = defaultProvider
			logging.Debug("Using provider from config: %s", options.Provider)
		}
	}
	
	// Try to get default model if not specified on command line
	if cmdLineModel == "" {
		// Load all providers to find the model
		providers, err := config.UpdateLoadAllProviders(configFile)
		if err != nil {
			logging.Warn("Failed to load providers: %v", err)
		} else {
			// Get the model for the current provider
			if providerConfig, ok := providers[options.Provider]; ok {
				if providerConfig.DefaultModel != "" {
					options.Model = providerConfig.DefaultModel
					logging.Info("Using configured default model for %s: %s", options.Provider, options.Model)
				} else {
					logging.Warn("No default model configured for provider %s", options.Provider)
				}
			}
		}
	}
	
	// FIXED: Only use emergency fallbacks if no model is configured at all
	if options.Model == "" {
		logging.Warn("No model specified in config or command line for provider %s, using emergency fallback", options.Provider)
		options.Model = getEmergencyFallbackModel(options.Provider)
	}
	
	// Handle provider-specific options
	if options.Provider == "ollama" {
		// For Ollama, get the API endpoint
		apiEndpoint, err := config.GetAPIEndpoint(options.Provider, configFile)
		if err != nil {
			logging.Warn("Failed to get API endpoint from config: %v", err)
			options.APIEndpoint = "http://localhost:11434" // Default Ollama endpoint
		} else {
			options.APIEndpoint = apiEndpoint
			logging.Debug("Using API endpoint from config for provider: %s", options.Provider)
		}
	} else {
		// For other providers, get the API key
		apiKey, err := config.UpdatedGetAPIKey(options.Provider, configFile)
		if err != nil {
			logging.Warn("Failed to get API key from config: %v", err)
		} else {
			options.APIKey = apiKey
			logging.Debug("Using API key from config for provider: %s", options.Provider)
		}
	}
	
	return options, nil
}

// getEmergencyFallbackModel returns emergency fallback models ONLY when config is missing
// This should rarely be used - configuration should always specify models
func getEmergencyFallbackModel(provider string) string {
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

// ProcessAIOptions processes the server options for chat and interactive commands
// Using a different name to avoid conflict with ProcessOptions in command_executor.go
func ProcessAIOptions(serverName string, disableFilesystem bool, providerName, modelName string) ([]string, map[string]bool) {
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
