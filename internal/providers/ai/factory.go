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
	// No hardcoded provider mapping - fully configuration-driven
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() domain.ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates a new provider instance based on interface type in config
func (f *ProviderFactory) CreateProvider(providerType domain.ProviderType, cfg *config.ProviderConfig, interfaceType config.InterfaceType) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider configuration is required")
	}

	logging.Info("Creating provider '%s' with interface type '%s'", providerType, interfaceType)

	// Create the appropriate client based on the interface type from configuration
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

// GetSupportedProviders returns supported interface types (not hardcoded providers)
func (f *ProviderFactory) GetSupportedProviders() []domain.ProviderType {
	// This method is deprecated in favor of configuration-driven approach
	return []domain.ProviderType{}
}

// GetProviderInterface is deprecated - interface type comes from configuration
func (f *ProviderFactory) GetProviderInterface(providerType domain.ProviderType) config.InterfaceType {
	// Return default - actual interface type should come from config
	return config.OpenAICompatible
}
