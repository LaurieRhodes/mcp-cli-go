package config

// InterfaceType represents the API interface type that a provider uses
type InterfaceType string

const (
	// OpenAICompatible represents the OpenAI-compatible API interface
	OpenAICompatible InterfaceType = "openai_compatible"

	// AnthropicNative represents the Anthropic-native API interface
	AnthropicNative InterfaceType = "anthropic_native"

	// OllamaNative represents the Ollama-native API interface
	OllamaNative InterfaceType = "ollama_native"
)

// InterfaceConfig represents the configuration for an API interface
type InterfaceConfig struct {
	// Providers maps provider names to their configurations
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// EnhancedProviderConfig represents an enhanced configuration for an AI provider with interface type
type EnhancedProviderConfig struct {
	// APIKey is the API key for the provider
	APIKey string `yaml:"api_key"`

	// DefaultModel is the default model to use for the provider
	DefaultModel string `yaml:"default_model"`

	// APIEndpoint is an optional endpoint URL (Used for Ollama, etc.)
	APIEndpoint string `yaml:"api_endpoint,omitempty"`

	// Interface is the type of API interface this provider uses
	Interface InterfaceType `yaml:"interface,omitempty"`

	// AvailableModels is the list of models this provider offers
	AvailableModels []string `yaml:"available_models,omitempty"`
}

// EnhancedAIConfig represents the enhanced AI configuration with interfaces
type EnhancedAIConfig struct {
	// DefaultProvider is the default AI provider to use
	DefaultProvider string `yaml:"default_provider"`

	// DefaultInterface is the default interface to use (for backward compatibility)
	DefaultInterface InterfaceType `yaml:"default_interface,omitempty"`

	// Providers maps provider names to their configurations (backward compatibility)
	Providers map[string]ProviderConfig `yaml:"providers,omitempty"`

	// Interfaces maps interface types to their configurations
	Interfaces map[InterfaceType]InterfaceConfig `yaml:"interfaces,omitempty"`
}

// EnhancedConfig represents the enhanced application configuration
type EnhancedConfig struct {
	// Servers maps server names to their configurations
	Servers map[string]ServerConfig `yaml:"servers"`

	// AI contains the AI-related configuration
	AI *EnhancedAIConfig `yaml:"ai,omitempty"`
}

// ConvertToEnhancedConfig converts a regular Config to an EnhancedConfig
func ConvertToEnhancedConfig(config *Config) *EnhancedConfig {
	if config == nil {
		return nil
	}

	// Create the enhanced config
	enhancedConfig := &EnhancedConfig{
		Servers: config.Servers,
		AI: &EnhancedAIConfig{
			DefaultProvider: config.AI.DefaultProvider,
			Interfaces:      make(map[InterfaceType]InterfaceConfig),
		},
	}

	// Map of provider names to their interface types
	providerInterfaces := map[string]InterfaceType{
		"openai":     OpenAICompatible,
		"deepseek":   OpenAICompatible,
		"openrouter": OpenAICompatible,
		"anthropic":  AnthropicNative,
		"ollama":     OllamaNative,
		"gemini":     OpenAICompatible, // For now, we'll treat Gemini as OpenAI-compatible
		"kimik2":     OpenAICompatible, // Kimi K2 (Moonshot AI)
		"lmstudio":   OpenAICompatible, // LM Studio
	}

	// Create the interfaces configuration
	for providerName, providerConfig := range config.AI.Providers {
		// Get the interface type for this provider
		interfaceType, ok := providerInterfaces[providerName]
		if !ok {
			// Default to OpenAI-compatible if not specified
			interfaceType = OpenAICompatible
		}

		// Make sure the interface exists
		if _, ok := enhancedConfig.AI.Interfaces[interfaceType]; !ok {
			enhancedConfig.AI.Interfaces[interfaceType] = InterfaceConfig{
				Providers: make(map[string]ProviderConfig),
			}
		}

		// Add the provider to the interface
		enhancedConfig.AI.Interfaces[interfaceType].Providers[providerName] = ProviderConfig{
			APIKey:       providerConfig.APIKey,
			DefaultModel: providerConfig.DefaultModel,
			APIEndpoint:  providerConfig.APIEndpoint,
		}
	}

	return enhancedConfig
}

// GenerateNewConfig generates a new configuration file with interfaces
func GenerateNewConfig(config *Config) *EnhancedConfig {
	// Base conversion
	enhancedConfig := ConvertToEnhancedConfig(config)

	// Define interface configurations
	openaiCompatible := InterfaceConfig{
		Providers: make(map[string]ProviderConfig),
	}

	anthropicNative := InterfaceConfig{
		Providers: make(map[string]ProviderConfig),
	}

	ollamaNative := InterfaceConfig{
		Providers: make(map[string]ProviderConfig),
	}

	// Map providers to their interfaces
	for providerName, providerConfig := range config.AI.Providers {
		switch providerName {
		case "openai", "deepseek", "openrouter", "kimik2", "lmstudio":
			openaiCompatible.Providers[providerName] = providerConfig
		case "anthropic":
			anthropicNative.Providers[providerName] = providerConfig
		case "ollama":
			ollamaNative.Providers[providerName] = providerConfig
		case "gemini":
			// For now, we'll treat Gemini as OpenAI-compatible
			openaiCompatible.Providers[providerName] = providerConfig
		}
	}

	// Set the interfaces
	enhancedConfig.AI.Interfaces = map[InterfaceType]InterfaceConfig{
		OpenAICompatible: openaiCompatible,
		AnthropicNative:  anthropicNative,
		OllamaNative:     ollamaNative,
	}

	return enhancedConfig
}
