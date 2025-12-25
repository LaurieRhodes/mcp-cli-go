package ai

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai/clients"
)

// ProviderFactory implements the domain.ProviderFactory interface
type ProviderFactory struct {
	// supportedProviders maps provider types to their interface types
	supportedProviders map[domain.ProviderType]config.InterfaceType
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() domain.ProviderFactory {
	factory := &ProviderFactory{
		supportedProviders: make(map[domain.ProviderType]config.InterfaceType),
	}

	// Initialize supported providers
	factory.supportedProviders[domain.ProviderOpenAI] = config.OpenAICompatible
	factory.supportedProviders[domain.ProviderAnthropic] = config.AnthropicNative
	factory.supportedProviders[domain.ProviderOllama] = config.OllamaNative
	factory.supportedProviders[domain.ProviderDeepSeek] = config.OpenAICompatible
	factory.supportedProviders[domain.ProviderGemini] = config.GeminiNative
	factory.supportedProviders[domain.ProviderOpenRouter] = config.OpenAICompatible
	factory.supportedProviders[domain.ProviderLMStudio] = config.OpenAICompatible

	return factory
}

// CreateProvider creates a new provider instance
func (f *ProviderFactory) CreateProvider(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider configuration is required")
	}

	logging.Info("Creating provider for type: %s", providerType)

	// Validate the provider type is supported
	interfaceType, exists := f.supportedProviders[providerType]
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	// Create the appropriate client based on the interface type
	switch interfaceType {
	case config.OpenAICompatible:
		return clients.NewOpenAICompatibleClient(providerType, cfg)
	case config.AnthropicNative:
		return clients.NewAnthropicClient(cfg)
	case config.OllamaNative:
		return clients.NewOllamaClient(cfg)
	case config.GeminiNative:
		return clients.NewGeminiClient(providerType, cfg)
	default:
		return nil, fmt.Errorf("unsupported interface type: %s", interfaceType)
	}
}

// GetSupportedProviders returns a list of supported provider types
func (f *ProviderFactory) GetSupportedProviders() []domain.ProviderType {
	providers := make([]domain.ProviderType, 0, len(f.supportedProviders))
	for providerType := range f.supportedProviders {
		providers = append(providers, providerType)
	}
	return providers
}

// GetProviderInterface returns the interface type for a provider
func (f *ProviderFactory) GetProviderInterface(providerType domain.ProviderType) config.InterfaceType {
	if interfaceType, exists := f.supportedProviders[providerType]; exists {
		return interfaceType
	}
	// Default to OpenAI compatible for unknown providers
	return config.OpenAICompatible
}
