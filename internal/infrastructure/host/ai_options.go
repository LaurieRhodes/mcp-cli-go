package host

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
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

	// Interface type (for determining which client to use)
	InterfaceType config.InterfaceType
}

// GetAIOptions loads AI options from the config file and command line flags
func GetAIOptions(configFile, cmdLineProvider, cmdLineModel string) (*AIOptions, error) {
	// Default options
	options := &AIOptions{
		Provider: cmdLineProvider,
		Model:    cmdLineModel,
	}

	// Set default values if still empty
	if options.Provider == "" {
		options.Provider = "openai"
		logging.Warn("No provider specified, defaulting to openai")
	}

	if options.Model == "" {
		logging.Warn("No model specified, using emergency fallback")
		options.Model = getEmergencyFallbackModel(options.Provider)
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
