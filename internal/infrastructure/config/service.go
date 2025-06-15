package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
)

// Service implements the domain.ConfigurationService interface
type Service struct {
	config *domain.ApplicationConfig
}

// NewService creates a new configuration service
func NewService() *Service {
	return &Service{}
}

// LoadConfig loads configuration from a file
func (s *Service) LoadConfig(filePath string) (*domain.ApplicationConfig, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, domain.ErrConfigNotFound.WithDetails(fmt.Sprintf("file: %s", filePath))
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, domain.WrapError(err, domain.ErrCodeIORead, "failed to read config file").
			WithContext("file", filePath)
	}

	// Try to parse as new format first
	var config domain.ApplicationConfig
	if err := json.Unmarshal(data, &config); err == nil {
		// Validate the configuration
		if validationErr := s.ValidateConfig(&config); validationErr != nil {
			return nil, validationErr
		}
		
		s.config = &config
		return &config, nil
	}

	// If new format fails, try legacy format migration
	migratedConfig, err := s.migrateLegacyConfig(data)
	if err != nil {
		return nil, domain.WrapError(err, domain.ErrCodeConfigInvalid, "failed to parse configuration file").
			WithContext("file", filePath)
	}

	s.config = migratedConfig
	return migratedConfig, nil
}

// LoadConfigOrCreateExample loads configuration from a file, or creates an example if none exists
func (s *Service) LoadConfigOrCreateExample(filePath string) (*domain.ApplicationConfig, bool, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create example configuration
		if err := CreateExampleConfigWithComments(filePath); err != nil {
			return nil, false, fmt.Errorf("failed to create example config: %w", err)
		}
		
		// Load the created example config
		config, err := s.LoadConfig(filePath)
		if err != nil {
			return nil, true, fmt.Errorf("failed to load created example config: %w", err)
		}
		
		return config, true, nil
	}

	// File exists, load normally
	config, err := s.LoadConfig(filePath)
	return config, false, err
}

// SaveConfig saves configuration to a file
func (s *Service) SaveConfig(config *domain.ApplicationConfig, filePath string) error {
	// Validate the configuration first
	if err := s.ValidateConfig(config); err != nil {
		return err
	}

	// Marshal the JSON with proper formatting
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return domain.WrapError(err, domain.ErrCodeIOFormat, "failed to marshal configuration").
			WithContext("file", filePath)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domain.WrapError(err, domain.ErrCodeIOWrite, "failed to create config directory").
			WithContext("directory", dir)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return domain.WrapError(err, domain.ErrCodeIOWrite, "failed to write config file").
			WithContext("file", filePath)
	}

	return nil
}

// CreateExampleConfigWithComments creates a config file with helpful comments
func CreateExampleConfigWithComments(filePath string) error {
	exampleJSON := `{
  "_comment": "MCP CLI Configuration Example - Remove this comment line after editing",
  "_setup_instructions": [
    "1. Replace # prefixed API keys with your actual API keys",
    "2. For Windows users: Download exe servers from https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers",
    "3. Update server paths to point to your downloaded executables or use npx on Unix systems", 
    "4. Modify templates to match your workflow needs",
    "5. Remove these comment fields when ready"
  ],
  
  "servers": {
    "filesystem": {
      "command": "filesystem-mcp.exe",
      "system_prompt": "You are a helpful assistant with access to the local filesystem. You can read, write, and analyze files to help users with their tasks."
    },
    "brave-search": {
      "command": "brave-search-mcp.exe",
      "env": {
        "BRAVE_API_KEY": "your-brave-api-key-here"
      },
      "system_prompt": "You are a helpful assistant with access to web search capabilities. You can search for current information and provide up-to-date answers."
    }
  },
  
  "ai": {
    "default_provider": "openrouter",
    "default_system_prompt": "You are a helpful AI assistant that can use various tools to help users accomplish their tasks. Always be clear about what tools you're using and why.",
    "interfaces": {
      "openai_compatible": {
        "providers": {
          "openai": {
            "api_key": "# your-openai-api-key-here",
            "api_endpoint": "https://api.openai.com/v1",
            "default_model": "gpt-4o",
            "available_models": ["gpt-4o", "gpt-4o-mini"],
            "max_retries": 2
          },
          "openrouter": {
            "api_key": "# your-openrouter-api-key-here", 
            "api_endpoint": "https://openrouter.ai/api/v1",
            "default_model": "qwen/qwen3-30b-a3b",
            "available_models": ["qwen/qwen3-30b-a3b"],
            "max_retries": 2
          },
          "deepseek": {
            "api_key": "# your-deepseek-api-key-here",
            "api_endpoint": "https://api.deepseek.com/v1", 
            "default_model": "deepseek-chat",
            "available_models": ["deepseek-chat"],
            "max_retries": 2
          }
        }
      },
      "anthropic_native": {
        "providers": {
          "anthropic": {
            "api_key": "# your-anthropic-api-key-here",
            "default_model": "claude-3-5-sonnet-20241022", 
            "available_models": ["claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"],
            "max_retries": 2
          }
        }
      },
      "gemini_native": {
        "providers": {
          "gemini": {
            "api_key": "# your-gemini-api-key-here",
            "default_model": "gemini-1.5-pro",
            "available_models": ["gemini-1.5-pro", "gemini-1.5-flash", "gemini-1.0-pro", "gemini-pro"],
            "max_retries": 2
          }
        }
      },
      "ollama_native": {
        "providers": {
          "ollama": {
            "api_endpoint": "http://localhost:11434",
            "default_model": "ollama.com/ajindal/llama3.1-storm:8b",
            "available_models": ["ollama.com/ajindal/llama3.1-storm:8b", "qwen3:30b"],
            "max_retries": 2
          }
        }
      }
    }
  },
  
  "templates": {
    "analyze_file": {
      "name": "analyze_file",
      "description": "Read and analyze a file from the filesystem with detailed insights",
      "steps": [
        {
          "step": 1,
          "name": "read_file",
          "base_prompt": "Read and analyze the file at path: {{file_path}}. Provide a comprehensive analysis including file type, content summary, structure, and any notable patterns or issues.",
          "servers": ["filesystem"],
          "output_variable": "file_analysis", 
          "system_prompt": "You are a file analysis expert. When reading files, provide detailed insights about content, structure, potential issues, and suggestions for improvement."
        }
      ],
      "variables": {
        "file_path": "./README.md"
      }
    },
    "search_and_summarize": {
      "name": "search_and_summarize", 
      "description": "Search for information on a topic and create a summary report",
      "steps": [
        {
          "step": 1,
          "name": "web_search",
          "base_prompt": "Search for recent information about: {{search_topic}}. Focus on finding reliable, up-to-date sources.",
          "servers": ["brave-search"],
          "output_variable": "search_results",
          "system_prompt": "You are a research assistant. When searching, focus on finding credible, recent sources and gathering comprehensive information on the topic."
        },
        {
          "step": 2,
          "name": "create_summary",
          "base_prompt": "Based on the search results: {{search_results}}\\n\\nCreate a well-structured summary report with key findings, trends, and implications about {{search_topic}}.",
          "output_variable": "final_report",
          "system_prompt": "You are a professional analyst. Create clear, well-organized reports that highlight key insights and actionable information."
        }
      ],
      "variables": {
        "search_topic": "artificial intelligence trends 2024"
      }
    },
    "simple_analyze": {
      "name": "simple_analyze",
      "description": "Simple analysis using data from stdin or input-data flag",
      "steps": [
        {
          "step": 1,
          "name": "analyze_input",
          "base_prompt": "Analyze this data and provide insights: {{stdin}}",
          "system_prompt": "You are an analytical expert. Provide clear, actionable insights from the provided data."
        }
      ]
    }
  }
}`

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the example config with comments
	return os.WriteFile(filePath, []byte(exampleJSON), 0644)
}

// GetProviderConfig retrieves provider configuration
func (s *Service) GetProviderConfig(providerName string) (*domain.ProviderConfig, domain.InterfaceType, error) {
	if s.config == nil || s.config.AI == nil {
		return nil, "", domain.ErrConfigNotFound.WithDetails("AI configuration not loaded")
	}

	// First check the new interface-based configuration
	if s.config.AI.Interfaces != nil {
		for interfaceType, interfaceConfig := range s.config.AI.Interfaces {
			if providerConfig, exists := interfaceConfig.Providers[providerName]; exists {
				return &providerConfig, interfaceType, nil
			}
		}
	}

	// Fall back to legacy providers configuration
	if s.config.AI.Providers != nil {
		if providerConfig, exists := s.config.AI.Providers[providerName]; exists {
			// Determine the interface type based on provider name
			interfaceType := s.getProviderInterfaceType(providerName)
			return &providerConfig, interfaceType, nil
		}
	}

	return nil, "", domain.ErrProviderNotFound.WithDetails(fmt.Sprintf("provider: %s", providerName))
}

// GetServerConfig retrieves server configuration
func (s *Service) GetServerConfig(serverName string) (*domain.ServerConfig, error) {
	if s.config == nil || s.config.Servers == nil {
		return nil, domain.ErrConfigNotFound.WithDetails("server configuration not loaded")
	}

	serverConfig, exists := s.config.Servers[serverName]
	if !exists {
		return nil, domain.ErrServerNotFound.WithDetails(fmt.Sprintf("server: %s", serverName))
	}

	return &serverConfig, nil
}

// GetDefaultProvider returns the default provider configuration
func (s *Service) GetDefaultProvider() (string, *domain.ProviderConfig, domain.InterfaceType, error) {
	if s.config == nil || s.config.AI == nil {
		return "", nil, "", domain.ErrConfigNotFound.WithDetails("AI configuration not loaded")
	}

	providerName := s.config.AI.DefaultProvider
	if providerName == "" {
		// Find the first available provider
		if s.config.AI.Interfaces != nil {
			for interfaceType, interfaceConfig := range s.config.AI.Interfaces {
				for name, config := range interfaceConfig.Providers {
					return name, &config, interfaceType, nil
				}
			}
		}

		// Fall back to legacy providers
		if s.config.AI.Providers != nil {
			for name, config := range s.config.AI.Providers {
				interfaceType := s.getProviderInterfaceType(name)
				return name, &config, interfaceType, nil
			}
		}

		return "", nil, "", domain.ErrProviderNotFound.WithDetails("no providers configured")
	}

	// Get the specified default provider
	providerConfig, interfaceType, err := s.GetProviderConfig(providerName)
	if err != nil {
		return "", nil, "", err
	}

	return providerName, providerConfig, interfaceType, nil
}

// ValidateConfig validates the entire configuration
func (s *Service) ValidateConfig(config *domain.ApplicationConfig) error {
	if config == nil {
		return domain.ErrConfigInvalid.WithDetails("configuration is nil")
	}

	// Validate AI configuration
	if config.AI != nil {
		if err := s.validateAIConfig(config.AI); err != nil {
			return err
		}
	}

	// Validate server configurations
	if config.Servers != nil {
		for serverName, serverConfig := range config.Servers {
			if err := s.validateServerConfig(serverName, &serverConfig); err != nil {
				return err
			}
		}
	}

	// Validate workflow templates
	if err := config.ValidateWorkflowTemplates(); err != nil {
		return err
	}

	return nil
}

// MigrateConfig migrates configuration from legacy format
func (s *Service) MigrateConfig(legacyConfigPath string) (*domain.ApplicationConfig, error) {
	// Read the legacy config file
	data, err := os.ReadFile(legacyConfigPath)
	if err != nil {
		return nil, domain.WrapError(err, domain.ErrCodeIORead, "failed to read legacy config file").
			WithContext("file", legacyConfigPath)
	}

	// Migrate the configuration
	config, err := s.migrateLegacyConfig(data)
	if err != nil {
		return nil, domain.WrapError(err, domain.ErrCodeConfigMigration, "failed to migrate legacy configuration").
			WithContext("file", legacyConfigPath)
	}

	return config, nil
}

// migrateLegacyConfig migrates from legacy configuration format
func (s *Service) migrateLegacyConfig(data []byte) (*domain.ApplicationConfig, error) {
	// Try to parse as the old Config format
	var legacyConfig struct {
		Servers map[string]domain.ServerConfig `json:"servers"`
		AI      *struct {
			DefaultProvider     string                       `json:"default_provider"`
			DefaultSystemPrompt string                       `json:"default_system_prompt,omitempty"`
			Providers           map[string]domain.ProviderConfig `json:"providers,omitempty"`
		} `json:"ai,omitempty"`
		Settings *domain.GlobalSettings `json:"settings,omitempty"`
	}

	if err := json.Unmarshal(data, &legacyConfig); err != nil {
		return nil, fmt.Errorf("failed to parse legacy configuration: %w", err)
	}

	// Convert to new format
	config := &domain.ApplicationConfig{
		Servers:  legacyConfig.Servers,
		Settings: legacyConfig.Settings,
	}

	if legacyConfig.AI != nil {
		config.AI = &domain.AIConfig{
			DefaultProvider:     legacyConfig.AI.DefaultProvider,
			DefaultSystemPrompt: legacyConfig.AI.DefaultSystemPrompt,
			Interfaces:          make(map[domain.InterfaceType]domain.InterfaceConfig),
			Providers:           legacyConfig.AI.Providers, // Keep for backward compatibility
		}

		// Migrate providers to interface-based configuration
		if legacyConfig.AI.Providers != nil {
			s.migrateProvidersToInterfaces(config.AI, legacyConfig.AI.Providers)
		}
	}

	return config, nil
}

// migrateProvidersToInterfaces migrates legacy provider config to interface-based config
func (s *Service) migrateProvidersToInterfaces(aiConfig *domain.AIConfig, providers map[string]domain.ProviderConfig) {
	// Group providers by interface type
	interfaceProviders := make(map[domain.InterfaceType]map[string]domain.ProviderConfig)

	for providerName, providerConfig := range providers {
		interfaceType := s.getProviderInterfaceType(providerName)
		
		if interfaceProviders[interfaceType] == nil {
			interfaceProviders[interfaceType] = make(map[string]domain.ProviderConfig)
		}
		
		interfaceProviders[interfaceType][providerName] = providerConfig
	}

	// Create interface configurations
	for interfaceType, providers := range interfaceProviders {
		aiConfig.Interfaces[interfaceType] = domain.InterfaceConfig{
			Providers: providers,
		}
	}
}

// getProviderInterfaceType determines the interface type for a provider
func (s *Service) getProviderInterfaceType(providerName string) domain.InterfaceType {
	switch strings.ToLower(providerName) {
	case "anthropic":
		return domain.AnthropicNative
	case "ollama":
		return domain.OllamaNative
	case "gemini":
		return domain.GeminiNative
	case "openai", "deepseek", "openrouter":
		return domain.OpenAICompatible
	default:
		return domain.OpenAICompatible // Default to OpenAI compatible
	}
}

// validateAIConfig validates AI configuration
func (s *Service) validateAIConfig(aiConfig *domain.AIConfig) error {
	if aiConfig.DefaultProvider == "" && len(aiConfig.Interfaces) == 0 && len(aiConfig.Providers) == 0 {
		return domain.ErrConfigInvalid.WithDetails("no AI providers configured")
	}

	// Validate that the default provider exists
	if aiConfig.DefaultProvider != "" {
		found := false
		
		// Check in interfaces
		if aiConfig.Interfaces != nil {
			for _, interfaceConfig := range aiConfig.Interfaces {
				if _, exists := interfaceConfig.Providers[aiConfig.DefaultProvider]; exists {
					found = true
					break
				}
			}
		}
		
		// Check in legacy providers
		if !found && aiConfig.Providers != nil {
			if _, exists := aiConfig.Providers[aiConfig.DefaultProvider]; exists {
				found = true
			}
		}
		
		if !found {
			return domain.ErrProviderNotFound.WithDetails(fmt.Sprintf("default provider '%s' not found in configuration", aiConfig.DefaultProvider))
		}
	}

	// Validate individual provider configurations
	if aiConfig.Interfaces != nil {
		for interfaceType, interfaceConfig := range aiConfig.Interfaces {
			for providerName, providerConfig := range interfaceConfig.Providers {
				if err := s.validateProviderConfig(providerName, &providerConfig, interfaceType); err != nil {
					return err
				}
			}
		}
	}

	// Validate legacy providers
	if aiConfig.Providers != nil {
		for providerName, providerConfig := range aiConfig.Providers {
			interfaceType := s.getProviderInterfaceType(providerName)
			if err := s.validateProviderConfig(providerName, &providerConfig, interfaceType); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateProviderConfig validates a provider configuration
func (s *Service) validateProviderConfig(providerName string, config *domain.ProviderConfig, interfaceType domain.InterfaceType) error {
	if config.DefaultModel == "" {
		return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("provider '%s' missing default model", providerName))
	}

	// Validate based on interface type
	switch interfaceType {
	case domain.AnthropicNative, domain.OpenAICompatible, domain.GeminiNative:
		if config.APIKey == "" {
			return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("provider '%s' missing API key", providerName))
		}
	case domain.OllamaNative:
		if config.APIEndpoint == "" {
			return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("provider '%s' missing API endpoint", providerName))
		}
	}

	return nil
}

// validateServerConfig validates a server configuration
func (s *Service) validateServerConfig(serverName string, config *domain.ServerConfig) error {
	if config.Command == "" {
		return domain.ErrConfigInvalid.WithDetails(fmt.Sprintf("server '%s' missing command", serverName))
	}

	return nil
}

// GetConfig returns the current loaded configuration
func (s *Service) GetConfig() *domain.ApplicationConfig {
	return s.config
}

// GetSystemPrompt returns the system prompt for a server or the default
func (s *Service) GetSystemPrompt(serverName string) string {
	if s.config == nil {
		return ""
	}

	// Check for server-specific prompt
	if serverName != "" && s.config.Servers != nil {
		if server, exists := s.config.Servers[serverName]; exists && server.SystemPrompt != "" {
			return server.SystemPrompt
		}
	}

	// Check for default AI system prompt
	if s.config.AI != nil && s.config.AI.DefaultSystemPrompt != "" {
		return s.config.AI.DefaultSystemPrompt
	}

	return ""
}

// GetMaxToolFollowUp returns the maximum tool follow-up attempts
func (s *Service) GetMaxToolFollowUp(serverName string) int {
	if s.config == nil {
		return 2 // Default fallback
	}

	// Check server-specific setting
	if serverName != "" && s.config.Servers != nil {
		if server, exists := s.config.Servers[serverName]; exists && server.Settings != nil {
			if maxFollowUp := server.Settings.GetMaxToolFollowUp(); maxFollowUp > 0 {
				return maxFollowUp
			}
		}
	}

	// Check global setting
	if s.config.Settings != nil {
		if maxFollowUp := s.config.Settings.GetMaxToolFollowUp(); maxFollowUp > 0 {
			return maxFollowUp
		}
	}

	return 2 // Default fallback
}

// ListProviders returns all available provider names
func (s *Service) ListProviders() []string {
	if s.config == nil || s.config.AI == nil {
		return nil
	}

	var providers []string
	seen := make(map[string]bool)

	// Collect from interfaces
	if s.config.AI.Interfaces != nil {
		for _, interfaceConfig := range s.config.AI.Interfaces {
			for providerName := range interfaceConfig.Providers {
				if !seen[providerName] {
					providers = append(providers, providerName)
					seen[providerName] = true
				}
			}
		}
	}

	// Collect from legacy providers
	if s.config.AI.Providers != nil {
		for providerName := range s.config.AI.Providers {
			if !seen[providerName] {
				providers = append(providers, providerName)
				seen[providerName] = true
			}
		}
	}

	return providers
}

// ListServers returns all available server names
func (s *Service) ListServers() []string {
	if s.config == nil || s.config.Servers == nil {
		return nil
	}

	var servers []string
	for serverName := range s.config.Servers {
		servers = append(servers, serverName)
	}

	return servers
}
