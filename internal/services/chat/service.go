package chat

import (
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/chat"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
)

// Service handles chat functionality and orchestrates the chat flow
type Service struct {
	aiService     *ai.Service
	configService domain.ConfigurationService
}

// Config holds configuration for chat execution
type Config struct {
	ConfigFile        string
	ServerName        string
	ProviderName      string
	ModelName         string
	DisableFilesystem bool
	ServerNames       []string
	UserSpecified     map[string]bool
}

// NewService creates a new chat service
func NewService() *Service {
	return &Service{
		aiService:     ai.NewService(),
		configService: infraConfig.NewService(),
	}
}

// StartChat starts a chat session with the given configuration
func (s *Service) StartChat(cfg *Config) error {
	logging.Info("Initializing chat mode...")
	
	// Load configuration to get provider config
	appConfig, err := s.configService.LoadConfig(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Get provider configuration for token management
	var providerConfig *config.ProviderConfig
	var interfaceType config.InterfaceType
	
	// Determine which provider to use (same logic as AI service)
	providerName := cfg.ProviderName
	if providerName == "" {
		if appConfig.AI != nil && appConfig.AI.DefaultProvider != "" {
			providerName = appConfig.AI.DefaultProvider
		} else {
			providerName = "openai" // Final fallback
		}
	}
	
	if providerName != "" {
		providerConfig, interfaceType, err = s.getProviderConfiguration(appConfig, providerName)
		if err != nil {
			logging.Warn("Failed to get provider config for %s: %v, proceeding without provider-specific token management", providerName, err)
		} else {
			logging.Info("Loaded provider configuration for %s (interface: %s)", providerName, interfaceType)
		}
	}
	
	// Get the model name (use override or provider default)
	modelName := cfg.ModelName
	if modelName == "" && providerConfig != nil {
		modelName = providerConfig.DefaultModel
	}
	
	// Initialize AI provider using the centralized AI service
	provider, err := s.aiService.InitializeProvider(cfg.ConfigFile, cfg.ProviderName, cfg.ModelName)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}
	defer provider.Close() // Clean up resources

	// Create UI at service level to ensure cleanup even on timeout
	ui := chat.NewUI()
	defer func() {
		if err := ui.Close(); err != nil {
			logging.Warn("Error closing UI: %v", err)
		}
	}()

	// Execute chat with server connections
	return host.RunCommand(func(conns []*host.ServerConnection) error {
		return s.runChat(conns, provider, providerConfig, modelName, ui)
	}, cfg.ConfigFile, cfg.ServerNames, cfg.UserSpecified)
}

// getProviderConfiguration retrieves provider config from the modular hierarchy
func (s *Service) getProviderConfiguration(appConfig *config.ApplicationConfig, providerName string) (*config.ProviderConfig, config.InterfaceType, error) {
	if appConfig.AI == nil {
		return nil, "", fmt.Errorf("AI configuration missing")
	}

	// Search through interface hierarchy
	if appConfig.AI.Interfaces != nil {
		for interfaceType, interfaceConfig := range appConfig.AI.Interfaces {
			if providerCfg, exists := interfaceConfig.Providers[providerName]; exists {
				return &providerCfg, interfaceType, nil
			}
		}
	}

	// Fallback to legacy providers section
	if appConfig.AI.Providers != nil {
		if providerCfg, exists := appConfig.AI.Providers[providerName]; exists {
			interfaceType := s.inferInterfaceType(providerName)
			return &providerCfg, interfaceType, nil
		}
	}

	return nil, "", fmt.Errorf("provider '%s' not found in configuration", providerName)
}

// inferInterfaceType determines interface type from provider name
func (s *Service) inferInterfaceType(providerName string) config.InterfaceType {
	switch strings.ToLower(providerName) {
	case "anthropic":
		return config.AnthropicNative
	case "ollama":
		return config.OllamaNative
	case "gemini":
		return config.GeminiNative
	case "openai", "deepseek", "openrouter", "lmstudio":
		return config.OpenAICompatible
	default:
		return config.OpenAICompatible // Safe default
	}
}

// runChat executes the chat session with server connections
func (s *Service) runChat(connections []*host.ServerConnection, provider domain.LLMProvider, providerConfig *config.ProviderConfig, model string, ui *chat.UI) error {
	// Validate provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider configuration validation failed: %w", err)
	}

	// Initialize and start chat using the enhanced chat manager with provider configuration
	var chatManager *chat.ChatManager
	if providerConfig != nil {
		chatManager = chat.NewChatManagerWithConfigAndUI(provider, connections, providerConfig, model, ui)
		logging.Info("Created chat manager with provider-aware token management for model: %s", model)
	} else {
		chatManager = chat.NewChatManagerWithUI(provider, connections, ui)
		logging.Info("Created chat manager with fallback token management")
	}
	
	if err := chatManager.StartChat(); err != nil {
		return fmt.Errorf("chat error: %w", err)
	}

	return nil
}
