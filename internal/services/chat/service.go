package chat

import (
	"fmt"
	"strings"

	appChat "github.com/LaurieRhodes/mcp-cli-go/internal/app/chat"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/chat"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	infraSkills "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
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
	SkillNames        []string // Filtered list of skills to expose
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
	if err == nil {
		logging.Debug("Loaded application config successfully")
		if appConfig.Chat != nil {
			logging.Debug("Chat config found: chat_logs_location=%s", appConfig.Chat.ChatLogsLocation)
		} else {
			logging.Debug("Chat config is nil in appConfig")
		}
	}
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

	// ARCHITECTURAL FIX: Separate built-in skills from external servers
	externalServers, needsSkills := infraSkills.SeparateSkillsFromServers(cfg.ServerNames)
	logging.Debug("External servers: %v, needs built-in skills: %v", externalServers, needsSkills)

	// Update userSpecified map to only include external servers
	externalUserSpecified := make(map[string]bool)
	for _, server := range externalServers {
		if cfg.UserSpecified[server] {
			externalUserSpecified[server] = true
		}
	}

	// Initialize built-in skills service if needed
	var skillService *skillsvc.Service
	if needsSkills {
		var err error
		skillService, err = infraSkills.InitializeBuiltinSkills(cfg.ConfigFile, appConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize built-in skills: %w", err)
		}
		logging.Info("Built-in skills service initialized successfully")
	}

	// Execute chat with server connections (ONLY external servers)
	return host.RunCommand(func(conns []*host.ServerConnection) error {
		// ARCHITECTURAL FIX: Create server manager (with skills if needed)
		var serverManager domain.MCPServerManager = infraSkills.NewHostServerManager(conns)
		if skillService != nil {
			logging.Info("Wrapping chat server manager with built-in skills support")
			serverManager = infraSkills.NewSkillsAwareServerManager(serverManager, skillService)
		}

		return s.runChat(serverManager, provider, providerConfig, modelName, ui, appConfig, cfg.SkillNames)
	}, cfg.ConfigFile, externalServers, externalUserSpecified)
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
func (s *Service) runChat(serverManager domain.MCPServerManager, provider domain.LLMProvider, providerConfig *config.ProviderConfig, model string, ui *chat.UI, appConfig *config.ApplicationConfig, skillNames []string) error {
	// Get chat configuration from loaded app config
	var chatConfig *config.ChatConfig
	if appConfig != nil && appConfig.Chat != nil {
		chatConfig = appConfig.Chat
		logging.Debug("Loaded chat config from settings: chat_logs_location=%s", chatConfig.ChatLogsLocation)
	} else {
		chatConfig = config.DefaultChatConfig()
		logging.Debug("Using default chat config (no session logging)")
	}

	// Create session logger if configured
	var sessionLogger *appChat.SessionLogger
	if chatConfig.ChatLogsLocation != "" {
		logger, err := appChat.NewSessionLogger(chatConfig.ChatLogsLocation)
		if err != nil {
			logging.Warn("Failed to create session logger: %v, continuing without session logging", err)
		} else {
			sessionLogger = logger
			logging.Info("Session logger created successfully for: %s", chatConfig.ChatLogsLocation)
			defer sessionLogger.Close()
		}
	}

	// Validate provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider configuration validation failed: %w", err)
	}

	// Initialize and start chat using the enhanced chat manager with provider configuration
	var chatManager *chat.ChatManager
	if providerConfig != nil {
		// ARCHITECTURAL FIX: Use ServerManager-based constructor if available
		chatManager = chat.NewChatManagerWithServerManagerAndUI(provider, serverManager, providerConfig, model, ui)
		logging.Info("Created chat manager with server manager and provider-aware token management for model: %s", model)
	} else {
		// Fallback for when providerConfig is nil (shouldn't happen but safe)
		chatManager = chat.NewChatManagerWithServerManagerAndUI(provider, serverManager, nil, model, ui)
		logging.Info("Created chat manager with server manager and fallback token management")
	}

	// Set enabled skills
	chatManager.EnabledSkills = skillNames

	// Configure session logging if enabled
	if sessionLogger != nil && sessionLogger.IsEnabled() {
		providerName := string(provider.GetProviderType())
		chatManager.SetSessionLogger(sessionLogger, providerName, model)
	}

	if err := chatManager.StartChat(); err != nil {
		return fmt.Errorf("chat error: %w", err)
	}

	return nil
}
