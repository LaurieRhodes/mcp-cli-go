package chat

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/chat"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
)

// Service handles chat functionality and orchestrates the chat flow
type Service struct {
	aiService *ai.Service
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
		aiService: ai.NewService(),
	}
}

// StartChat starts a chat session with the given configuration
func (s *Service) StartChat(config *Config) error {
	logging.Info("Initializing chat mode...")
	
	// Initialize AI provider using the centralized AI service
	provider, err := s.aiService.InitializeProvider(config.ConfigFile, config.ProviderName, config.ModelName)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}
	defer provider.Close() // Clean up resources

	// Execute chat with server connections
	return host.RunCommand(func(conns []*host.ServerConnection) error {
		return s.runChat(conns, provider)
	}, config.ConfigFile, config.ServerNames, config.UserSpecified)
}

// runChat executes the chat session with server connections
func (s *Service) runChat(connections []*host.ServerConnection, provider domain.LLMProvider) error {
	// Validate provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider configuration validation failed: %w", err)
	}

	// Initialize and start chat using the existing chat manager
	// The chat.Manager from internal/core/chat handles the actual chat loop
	chatManager := chat.NewChatManager(provider, connections)
	if err := chatManager.StartChat(); err != nil {
		return fmt.Errorf("chat error: %w", err)
	}

	return nil
}
