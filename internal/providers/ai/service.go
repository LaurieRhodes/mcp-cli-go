// Centralized AI Service Architecture
// This is the single point of AI initialization for all commands

package ai

import (
	"fmt"
	"os"
	"strings"

	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Service provides centralized AI provider management
type Service struct {
	configService domain.ConfigurationService
	factory       domain.ProviderFactory
}

// NewService creates a new AI service
func NewService() *Service {
	return &Service{
		configService: infraConfig.NewService(),
		factory:       NewProviderFactory(),
	}
}

// GetProviderFactory returns the provider factory for creating providers
func (s *Service) GetProviderFactory() domain.ProviderFactory {
	return s.factory
}

// InitializeProvider initializes an AI provider based on config and command-line overrides
func (s *Service) InitializeProvider(configFile, providerOverride, modelOverride string) (domain.LLMProvider, error) {
	logging.Debug("Initializing AI provider with config: %s, provider: %s, model: %s", 
		configFile, providerOverride, modelOverride)

	// Load configuration
	appConfig, err := s.configService.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine which provider to use
	providerName := providerOverride
	if providerName == "" {
		if appConfig.AI != nil && appConfig.AI.DefaultProvider != "" {
			providerName = appConfig.AI.DefaultProvider
		} else {
			providerName = "openai" // Final fallback
		}
	}

	logging.Info("Using AI provider: %s", providerName)

	// Get provider configuration from the modular config hierarchy
	providerConfig, interfaceType, err := s.getProviderConfiguration(appConfig, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider configuration for %s: %w", providerName, err)
	}

	// Override model if specified
	if modelOverride != "" {
		providerConfig.DefaultModel = modelOverride
		logging.Debug("Model overridden to: %s", modelOverride)
	}

	// Try environment variable for API key if not in config
	if providerConfig.APIKey == "" && interfaceType != config.OllamaNative {
		envKey := s.getAPIKeyFromEnv(providerName)
		if envKey != "" {
			providerConfig.APIKey = envKey
			logging.Debug("Using API key from environment for %s", providerName)
		}
	}

	// Validate we have required configuration
	if err := s.validateProviderConfig(providerName, providerConfig, interfaceType); err != nil {
		return nil, err
	}

	// Map provider name to provider type for the factory
	providerType, err := s.mapProviderNameToType(providerName)
	if err != nil {
		return nil, err
	}

	// Create the provider using the factory
	provider, err := s.factory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", providerName, err)
	}

	logging.Info("Successfully initialized AI provider: %s with model: %s", 
		providerName, providerConfig.DefaultModel)

	return provider, nil
}

// getProviderConfiguration retrieves provider config from the modular hierarchy
func (s *Service) getProviderConfiguration(appConfig *config.ApplicationConfig, providerName string) (*config.ProviderConfig, config.InterfaceType, error) {
	if appConfig.AI == nil {
		return nil, "", domain.ErrConfigNotFound.WithDetails("AI configuration missing")
	}

	// Search through interface hierarchy
	if appConfig.AI.Interfaces != nil {
		for interfaceType, interfaceConfig := range appConfig.AI.Interfaces {
			if providerConfig, exists := interfaceConfig.Providers[providerName]; exists {
				return &providerConfig, interfaceType, nil
			}
		}
	}

	// Fallback to legacy providers section
	if appConfig.AI.Providers != nil {
		if providerConfig, exists := appConfig.AI.Providers[providerName]; exists {
			interfaceType := s.inferInterfaceType(providerName)
			return &providerConfig, interfaceType, nil
		}
	}

	return nil, "", domain.ErrProviderNotFound.WithDetails(fmt.Sprintf("provider '%s' not found in configuration", providerName))
}

// inferInterfaceType determines interface type from provider name (fallback for legacy configs)
func (s *Service) inferInterfaceType(providerName string) config.InterfaceType {
	// This is a fallback for old configs that don't specify interface_type
	// New configs should always specify interface_type in the provider file
	switch strings.ToLower(providerName) {
	case "anthropic":
		return config.AnthropicNative
	case "ollama":
		return config.OllamaNative
	case "gemini":
		return config.GeminiNative
	default:
		// Safe default for OpenAI-compatible providers
		// This includes: openai, deepseek, openrouter, lmstudio, and any custom providers
		return config.OpenAICompatible
	}
}

// mapProviderNameToType converts config provider name to domain provider type
// For truly configuration-driven behavior, we use the provider name directly as the type
func (s *Service) mapProviderNameToType(providerName string) (domain.ProviderType, error) {
	// Simply use the provider name as-is - the factory will determine interface type from config
	return domain.ProviderType(providerName), nil
}

// getAPIKeyFromEnv retrieves API key from environment variables
func (s *Service) getAPIKeyFromEnv(providerName string) string {
	envVars := map[string]string{
		"openai":     "OPENAI_API_KEY",
		"anthropic":  "ANTHROPIC_API_KEY", 
		"gemini":     "GEMINI_API_KEY",
		"deepseek":   "DEEPSEEK_API_KEY",
		"openrouter": "OPENROUTER_API_KEY",
		// LMStudio doesn't need an API key (local service)
	}

	if envVar, exists := envVars[strings.ToLower(providerName)]; exists {
		return os.Getenv(envVar)
	}

	return ""
}

// validateProviderConfig ensures provider has required configuration
func (s *Service) validateProviderConfig(providerName string, cfg *config.ProviderConfig, interfaceType config.InterfaceType) error {
	if cfg.DefaultModel == "" {
		return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("provider '%s' missing default model", providerName))
	}

	// API key required for cloud providers (excluding Ollama and LMStudio)
	providerLower := strings.ToLower(providerName)
	if interfaceType != config.OllamaNative && providerLower != "lmstudio" && cfg.APIKey == "" {
		return domain.ErrProviderAuth.WithDetails(fmt.Sprintf("provider '%s' missing API key", providerName))
	}

	// Endpoint required for Ollama
	if interfaceType == config.OllamaNative && cfg.APIEndpoint == "" {
		return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("provider '%s' missing API endpoint", providerName))
	}

	return nil
}

// GetAvailableProviders returns list of configured providers
func (s *Service) GetAvailableProviders(configFile string) ([]string, error) {
	appConfig, err := s.configService.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	var providers []string
	seen := make(map[string]bool)

	if appConfig.AI != nil {
		// From interface hierarchy
		if appConfig.AI.Interfaces != nil {
			for _, interfaceConfig := range appConfig.AI.Interfaces {
				for providerName := range interfaceConfig.Providers {
					if !seen[providerName] {
						providers = append(providers, providerName)
						seen[providerName] = true
					}
				}
			}
		}

		// From legacy providers
		if appConfig.AI.Providers != nil {
			for providerName := range appConfig.AI.Providers {
				if !seen[providerName] {
					providers = append(providers, providerName)
					seen[providerName] = true
				}
			}
		}
	}

	return providers, nil
}

// GetDefaultProvider returns the configured default provider
func (s *Service) GetDefaultProvider(configFile string) (string, error) {
	appConfig, err := s.configService.LoadConfig(configFile)
	if err != nil {
		return "", err
	}

	if appConfig.AI != nil && appConfig.AI.DefaultProvider != "" {
		return appConfig.AI.DefaultProvider, nil
	}

	return "openai", nil // Safe fallback
}
