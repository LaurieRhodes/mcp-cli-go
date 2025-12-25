package config

// InterfaceType represents the API interface type that a provider uses
type InterfaceType string

const (
	OpenAICompatible InterfaceType = "openai_compatible"
	AnthropicNative  InterfaceType = "anthropic_native"
	OllamaNative     InterfaceType = "ollama_native"
	GeminiNative     InterfaceType = "gemini_native"
)

// AIConfig represents the AI configuration
type AIConfig struct {
	DefaultProvider     string                            `yaml:"default_provider"`
	DefaultSystemPrompt string                            `yaml:"default_system_prompt,omitempty"`
	MaxToolFollowUp     int                               `yaml:"max_tool_follow_up,omitempty"`
	Interfaces          map[InterfaceType]InterfaceConfig `yaml:"interfaces"`
	Providers           map[string]ProviderConfig         `yaml:"providers,omitempty"`
}

// GetMaxToolFollowUp returns the max tool follow-up setting from AI config
func (ai *AIConfig) GetMaxToolFollowUp() int {
	if ai == nil {
		return 0
	}
	return ai.MaxToolFollowUp
}

// InterfaceConfig represents configuration for an API interface
type InterfaceConfig struct {
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig represents configuration for an LLM provider
type ProviderConfig struct {
	APIKey                string                          `yaml:"api_key"`
	DefaultModel          string                          `yaml:"default_model"`
	APIEndpoint           string                          `yaml:"api_endpoint,omitempty"`
	AvailableModels       []string                        `yaml:"available_models,omitempty"`
	TimeoutSeconds        int                             `yaml:"timeout_seconds,omitempty"`
	MaxRetries            int                             `yaml:"max_retries,omitempty"`
	Temperature           float64                         `yaml:"temperature,omitempty"`
	MaxTokens             int                             `yaml:"max_tokens,omitempty"`
	ContextWindow         int                             `yaml:"context_window,omitempty"`
	ReserveTokens         int                             `yaml:"reserve_tokens,omitempty"`
	EmbeddingModels       map[string]EmbeddingModelConfig `yaml:"embedding_models,omitempty"`
	DefaultEmbeddingModel string                          `yaml:"default_embedding_model,omitempty"`
}

// EmbeddingModelConfig represents configuration for a specific embedding model
type EmbeddingModelConfig struct {
	MaxTokens       int     `yaml:"max_tokens"`
	Dimensions      int     `yaml:"dimensions"`
	CostPer1kTokens float64 `yaml:"cost_per_1k_tokens,omitempty"`
	Default         bool    `yaml:"default,omitempty"`
	Description     string  `yaml:"description,omitempty"`
}
