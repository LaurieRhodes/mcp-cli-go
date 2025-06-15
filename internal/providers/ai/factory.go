package ai

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai/clients"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// ProviderFactory implements the domain.ProviderFactory interface
type ProviderFactory struct {
	// supportedProviders maps provider types to their interface types
	supportedProviders map[domain.ProviderType]domain.InterfaceType
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() domain.ProviderFactory {
	factory := &ProviderFactory{
		supportedProviders: make(map[domain.ProviderType]domain.InterfaceType),
	}
	
	// Initialize supported providers
	factory.supportedProviders[domain.ProviderOpenAI] = domain.OpenAICompatible
	factory.supportedProviders[domain.ProviderAnthropic] = domain.AnthropicNative
	factory.supportedProviders[domain.ProviderOllama] = domain.OllamaNative
	factory.supportedProviders[domain.ProviderDeepSeek] = domain.OpenAICompatible
	factory.supportedProviders[domain.ProviderGemini] = domain.GeminiNative // Changed to native
	factory.supportedProviders[domain.ProviderOpenRouter] = domain.OpenAICompatible
	
	return factory
}

// CreateProvider creates a new provider instance
func (f *ProviderFactory) CreateProvider(providerType domain.ProviderType, config *domain.ProviderConfig) (domain.LLMProvider, error) {
	if config == nil {
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
	case domain.OpenAICompatible:
		return clients.NewOpenAICompatibleClient(providerType, config)
	case domain.AnthropicNative:
		return clients.NewAnthropicClient(config)
	case domain.OllamaNative:
		return clients.NewOllamaClient(config)
	case domain.GeminiNative:
		return clients.NewGeminiClient(config)
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
func (f *ProviderFactory) GetProviderInterface(providerType domain.ProviderType) domain.InterfaceType {
	if interfaceType, exists := f.supportedProviders[providerType]; exists {
		return interfaceType
	}
	// Default to OpenAI compatible for unknown providers
	return domain.OpenAICompatible
}
