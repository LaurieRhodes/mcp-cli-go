package host

import (
	"os"

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
	
	// Load config to get default provider and settings
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logging.Warn("Failed to load config: %v", err)
		// Use fallbacks if config fails
		if cmdLineProvider == "" {
			options.Provider = "openai"
		}
		if cmdLineModel == "" {
			options.Model = getEmergencyFallbackModel(options.Provider)
		}
		return options, nil
	}
	
	// Try to get default provider if not specified on command line
	if cmdLineProvider == "" {
		defaultProvider := config.GetDefaultProviderFromConfig(cfg)
		options.Provider = defaultProvider
		logging.Debug("Using provider from config: %s", options.Provider)
	}
	
	// Try to get default model if not specified on command line
	if cmdLineModel == "" {
		// Get the model for the current provider
		defaultModel, err := config.GetDefaultModelFromConfig(cfg, options.Provider)
		if err != nil {
			logging.Warn("Failed to get default model for %s: %v", options.Provider, err)
			options.Model = getEmergencyFallbackModel(options.Provider)
		} else {
			options.Model = defaultModel
			logging.Info("Using configured default model for %s: %s", options.Provider, options.Model)
		}
	}
	
	// Handle provider-specific options
	if options.Provider == "ollama" {
		// For Ollama, try to get the API endpoint from config
		providerConfig, err := config.GetProviderConfigFromConfig(cfg, options.Provider)
		if err == nil && providerConfig.APIEndpoint != "" {
			options.APIEndpoint = providerConfig.APIEndpoint
			logging.Debug("Using API endpoint from config for provider: %s", options.Provider)
		} else {
			options.APIEndpoint = "http://localhost:11434" // Default Ollama endpoint
			logging.Debug("Using default Ollama endpoint")
		}
	} else {
		// For other providers, get the API key
		apiKey, err := config.GetAPIKeyFromConfig(cfg, options.Provider)
		if err != nil {
			// Try environment variables as fallback
			envKey := getAPIKeyFromEnv(options.Provider)
			if envKey != "" {
				options.APIKey = envKey
				logging.Debug("Using API key from environment for provider: %s", options.Provider)
			} else {
				logging.Warn("Failed to get API key from config or environment: %v", err)
			}
		} else {
			options.APIKey = apiKey
			logging.Debug("Using API key from config for provider: %s", options.Provider)
		}
	}
	
	return options, nil
}

// getAPIKeyFromEnv gets API key from environment variables
func getAPIKeyFromEnv(provider string) string {
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "gemini":
		return os.Getenv("GEMINI_API_KEY")
	case "deepseek":
		return os.Getenv("DEEPSEEK_API_KEY")
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY")
	default:
		return ""
	}
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
