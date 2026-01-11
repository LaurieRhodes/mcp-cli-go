package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	domainConfig "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Service implements the ConfigurationService interface
type Service struct {
	config    *domainConfig.ApplicationConfig
	configDir string
	loader    *domainConfig.Loader
}

// NewService creates a new configuration service
func NewService() *Service {
	return &Service{}
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// loadEnvFile loads .env file from the config directory
func (s *Service) loadEnvFile(configPath string) error {
	// Determine directory to search for .env
	configDir := filepath.Dir(configPath)
	
	// Try .env in config directory
	envPath := filepath.Join(configDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		return nil
	}
	
	// Try .env in executable directory
	execDir := getExecutableDir()
	envPath = filepath.Join(execDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		return nil
	}
	
	// .env file is optional, not an error if missing
	return nil
}

// expandEnvVars expands environment variables in a string
// Supports ${VAR_NAME} and $VAR_NAME formats
// Only expands if the string looks like an environment variable reference
func expandEnvVars(s string) string {
	// Don't expand if empty
	if s == "" {
		return s
	}
	
	// Check if it looks like an environment variable reference
	// Pattern: ${VAR} or $VAR (where VAR starts with letter/underscore)
	hasEnvPattern := strings.Contains(s, "${") || 
		(strings.Contains(s, "$") && len(s) > 1 && (s[0] == '$' || strings.Contains(s, " $")))
	
	if !hasEnvPattern {
		return s
	}
	
	return os.ExpandEnv(s)
}

// expandEnvVarsInConfig recursively expands environment variables in the config
func (s *Service) expandEnvVarsInConfig(config *domainConfig.ApplicationConfig) {
	// Expand in AI providers
	if config.AI != nil && config.AI.Interfaces != nil {
		for interfaceType, interfaceConfig := range config.AI.Interfaces {
			for providerName, providerConfig := range interfaceConfig.Providers {
				providerConfig.APIKey = expandEnvVars(providerConfig.APIKey)
				providerConfig.APIEndpoint = expandEnvVars(providerConfig.APIEndpoint)
				// AWS Bedrock specific fields
				providerConfig.AWSRegion = expandEnvVars(providerConfig.AWSRegion)
				providerConfig.AWSAccessKeyID = expandEnvVars(providerConfig.AWSAccessKeyID)
				providerConfig.AWSSecretAccessKey = expandEnvVars(providerConfig.AWSSecretAccessKey)
				providerConfig.AWSSessionToken = expandEnvVars(providerConfig.AWSSessionToken)
				// GCP Vertex AI specific fields
				providerConfig.ProjectID = expandEnvVars(providerConfig.ProjectID)
				providerConfig.Location = expandEnvVars(providerConfig.Location)
				providerConfig.CredentialsPath = expandEnvVars(providerConfig.CredentialsPath)
				interfaceConfig.Providers[providerName] = providerConfig
			}
			config.AI.Interfaces[interfaceType] = interfaceConfig
		}
	}
	
	// Expand in legacy providers
	if config.AI != nil && config.AI.Providers != nil {
		for providerName, providerConfig := range config.AI.Providers {
			providerConfig.APIKey = expandEnvVars(providerConfig.APIKey)
			providerConfig.APIEndpoint = expandEnvVars(providerConfig.APIEndpoint)
			// AWS Bedrock specific fields
			providerConfig.AWSRegion = expandEnvVars(providerConfig.AWSRegion)
			providerConfig.AWSAccessKeyID = expandEnvVars(providerConfig.AWSAccessKeyID)
			providerConfig.AWSSecretAccessKey = expandEnvVars(providerConfig.AWSSecretAccessKey)
			providerConfig.AWSSessionToken = expandEnvVars(providerConfig.AWSSessionToken)
			// GCP Vertex AI specific fields
			providerConfig.ProjectID = expandEnvVars(providerConfig.ProjectID)
			providerConfig.Location = expandEnvVars(providerConfig.Location)
			providerConfig.CredentialsPath = expandEnvVars(providerConfig.CredentialsPath)
			config.AI.Providers[providerName] = providerConfig
		}
	}
	
	// Expand in embedding providers
	if config.Embeddings != nil && config.Embeddings.Interfaces != nil {
		for interfaceType, interfaceConfig := range config.Embeddings.Interfaces {
			for providerName, providerConfig := range interfaceConfig.Providers {
				providerConfig.APIKey = expandEnvVars(providerConfig.APIKey)
				providerConfig.APIEndpoint = expandEnvVars(providerConfig.APIEndpoint)
				interfaceConfig.Providers[providerName] = providerConfig
			}
			config.Embeddings.Interfaces[interfaceType] = interfaceConfig
		}
	}
	
	// Expand in servers
	if config.Servers != nil {
		for serverName, serverConfig := range config.Servers {
			serverConfig.Command = expandEnvVars(serverConfig.Command)
			for i, arg := range serverConfig.Args {
				serverConfig.Args[i] = expandEnvVars(arg)
			}
			if serverConfig.Env != nil {
				for key, value := range serverConfig.Env {
					serverConfig.Env[key] = expandEnvVars(value)
				}
			}
			config.Servers[serverName] = serverConfig
		}
	}
}

// LoadConfig loads configuration from a file (supports both monolithic and modular)
func (s *Service) LoadConfig(filePath string) (*domainConfig.ApplicationConfig, error) {
	// Load .env file first
	s.loadEnvFile(filePath)
	
	// Initialize loader if needed
	if s.loader == nil {
		s.loader = domainConfig.NewLoader()
	}
	
	// Use loader (handles both single file and modular)
	config, err := s.loader.Load(filePath)
	if err != nil {
		return nil, err
	}

	// Expand environment variables in config
	s.expandEnvVarsInConfig(config)
	
	// Store config directory for future use
	s.configDir = filepath.Dir(filePath)
	s.config = config
	
	return config, nil
}

// LoadConfigOrCreateExample loads config or creates an example if it doesn't exist
func (s *Service) LoadConfigOrCreateExample(filePath string) (*domainConfig.ApplicationConfig, bool, error) {
	config, err := s.LoadConfig(filePath)
	if err != nil {
	} else {
		return config, false, nil
	}

	// Create a basic example config
	exampleConfig := &domainConfig.ApplicationConfig{
		Servers: make(map[string]domainConfig.ServerConfig),
		AI: &domainConfig.AIConfig{
			DefaultProvider: "ollama",
			Interfaces: map[domainConfig.InterfaceType]domainConfig.InterfaceConfig{
				domainConfig.OllamaNative: {
					Providers: map[string]domainConfig.ProviderConfig{
						"ollama": {
							APIKey:       "",
							DefaultModel: "llama2",
							APIEndpoint:  "http://localhost:11434",
						},
					},
				},
			},
		},
	}

	// Save example config
	if s.loader == nil {
		s.loader = domainConfig.NewLoader()
	}
	if err := s.loader.Save(exampleConfig, filePath); err != nil {
		return nil, false, fmt.Errorf("failed to create example config: %w", err)
	}

	config, err = s.LoadConfig(filePath)
	if err != nil {
		return nil, true, fmt.Errorf("failed to load created example config: %w", err)
	}

	return config, true, nil
}

// SaveConfig saves configuration to a file (format determined by loader)
func (s *Service) SaveConfig(config *domainConfig.ApplicationConfig, filePath string) error {
	if s.loader == nil {
		s.loader = domainConfig.NewLoader()
	}
	return s.loader.Save(config, filePath)
}

// GetProvider creates and returns a provider instance (placeholder implementation)
func (s *Service) GetProvider(providerName string) (domain.LLMProvider, error) {
	return nil, fmt.Errorf("GetProvider is a placeholder - use provider factory instead for creating provider instances")
}

// GetProviderConfig retrieves provider configuration
func (s *Service) GetProviderConfig(providerName string) (*domainConfig.ProviderConfig, domainConfig.InterfaceType, error) {
	if s.config == nil || s.config.AI == nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "AI configuration not loaded")
	}

	for interfaceType, interfaceConfig := range s.config.AI.Interfaces {
		if provider, exists := interfaceConfig.Providers[providerName]; exists {
			return &provider, interfaceType, nil
		}
	}

	if s.config.AI.Providers != nil {
		if provider, exists := s.config.AI.Providers[providerName]; exists {
			return &provider, domainConfig.OpenAICompatible, nil
		}
	}

	return nil, "", domain.NewDomainError(domain.ErrCodeProviderNotFound, fmt.Sprintf("provider '%s' not found", providerName))
}

// GetEmbeddingProviderConfig retrieves embedding provider configuration
func (s *Service) GetEmbeddingProviderConfig(providerName string) (*domainConfig.EmbeddingProviderConfig, domainConfig.InterfaceType, error) {
	if s.config == nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "configuration not loaded")
	}

	if s.config.Embeddings != nil && s.config.Embeddings.Interfaces != nil {
		for interfaceType, interfaceConfig := range s.config.Embeddings.Interfaces {
			if provider, exists := interfaceConfig.Providers[providerName]; exists {
				return &provider, interfaceType, nil
			}
		}
	}

	if s.config.AI != nil {
		for interfaceType, interfaceConfig := range s.config.AI.Interfaces {
			if aiProvider, exists := interfaceConfig.Providers[providerName]; exists {
				if aiProvider.EmbeddingModels != nil && len(aiProvider.EmbeddingModels) > 0 {
					embeddingProvider := &domainConfig.EmbeddingProviderConfig{
						APIKey:         aiProvider.APIKey,
						APIEndpoint:    aiProvider.APIEndpoint,
						DefaultModel:   aiProvider.DefaultEmbeddingModel,
						TimeoutSeconds: aiProvider.TimeoutSeconds,
						MaxRetries:     aiProvider.MaxRetries,
						Models:         aiProvider.EmbeddingModels,
					}
					
					availableModels := make([]string, 0, len(aiProvider.EmbeddingModels))
					for modelName := range aiProvider.EmbeddingModels {
						availableModels = append(availableModels, modelName)
					}
					embeddingProvider.AvailableModels = availableModels
					
					return embeddingProvider, interfaceType, nil
				}
			}
		}
	}

	return nil, "", domain.NewDomainError(domain.ErrCodeProviderNotFound, fmt.Sprintf("embedding provider '%s' not found", providerName))
}

// GetServerConfig retrieves server configuration
func (s *Service) GetServerConfig(serverName string) (*domainConfig.ServerConfig, error) {
	if s.config == nil || s.config.Servers == nil {
		return nil, domain.NewDomainError(domain.ErrCodeConfigInvalid, "server configuration not loaded")
	}

	server, exists := s.config.Servers[serverName]
	if !exists {
		return nil, domain.NewDomainError(domain.ErrCodeServerNotFound, fmt.Sprintf("server '%s' not found", serverName))
	}

	return &server, nil
}

// GetDefaultProvider returns the default provider configuration
func (s *Service) GetDefaultProvider() (string, *domainConfig.ProviderConfig, domainConfig.InterfaceType, error) {
	if s.config == nil || s.config.AI == nil {
		return "", nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "AI configuration not loaded")
	}

	defaultProviderName := s.config.AI.DefaultProvider
	if defaultProviderName == "" {
		return "", nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "default provider not specified")
	}

	providerConfig, interfaceType, err := s.GetProviderConfig(defaultProviderName)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to get default provider config: %w", err)
	}

	return defaultProviderName, providerConfig, interfaceType, nil
}

// GetDefaultEmbeddingProvider returns the default embedding provider configuration
func (s *Service) GetDefaultEmbeddingProvider() (string, *domainConfig.EmbeddingProviderConfig, domainConfig.InterfaceType, error) {
	if s.config == nil {
		return "", nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "configuration not loaded")
	}

	var defaultProviderName string
	
	if s.config.Embeddings != nil && s.config.Embeddings.DefaultProvider != "" {
		defaultProviderName = s.config.Embeddings.DefaultProvider
	} else if s.config.AI != nil && s.config.AI.DefaultProvider != "" {
		defaultProviderName = s.config.AI.DefaultProvider
	} else {
		return "", nil, "", domain.NewDomainError(domain.ErrCodeConfigInvalid, "default embedding provider not specified")
	}

	providerConfig, interfaceType, err := s.GetEmbeddingProviderConfig(defaultProviderName)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to get default embedding provider config: %w", err)
	}

	return defaultProviderName, providerConfig, interfaceType, nil
}

// ListServers returns a list of configured server names
func (s *Service) ListServers() []string {
	if s.config == nil || s.config.Servers == nil {
		return []string{}
	}

	names := make([]string, 0, len(s.config.Servers))
	for name := range s.config.Servers {
		names = append(names, name)
	}

	return names
}

// ListEmbeddingProviders returns a list of configured embedding provider names
func (s *Service) ListEmbeddingProviders() []string {
	providers := make(map[string]bool)
	
	if s.config != nil && s.config.Embeddings != nil && s.config.Embeddings.Interfaces != nil {
		for _, interfaceConfig := range s.config.Embeddings.Interfaces {
			for providerName := range interfaceConfig.Providers {
				providers[providerName] = true
			}
		}
	}
	
	if s.config != nil && s.config.AI != nil && s.config.AI.Interfaces != nil {
		for _, interfaceConfig := range s.config.AI.Interfaces {
			for providerName, providerConfig := range interfaceConfig.Providers {
				if len(providerConfig.EmbeddingModels) > 0 {
					providers[providerName] = true
				}
			}
		}
	}
	
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	
	return names
}

// ValidateConfig validates the entire configuration
func (s *Service) ValidateConfig(config *domainConfig.ApplicationConfig) error {
	if config.AI == nil {
		return domain.NewDomainError(domain.ErrCodeConfigInvalid, "AI configuration is required")
	}

	if config.AI.DefaultProvider == "" {
		return domain.NewDomainError(domain.ErrCodeConfigInvalid, "default provider is required")
	}

	found := false
	for _, interfaceConfig := range config.AI.Interfaces {
		if _, exists := interfaceConfig.Providers[config.AI.DefaultProvider]; exists {
			found = true
			break
		}
	}

	if !found && config.AI.Providers != nil {
		if _, exists := config.AI.Providers[config.AI.DefaultProvider]; exists {
			found = true
		}
	}

	if !found {
		return domain.NewDomainError(domain.ErrCodeProviderNotFound, fmt.Sprintf("default provider '%s' not found in configuration", config.AI.DefaultProvider))
	}

	if err := config.ValidateWorkflows(); err != nil {
		return fmt.Errorf("workflow template validation failed: %w", err)
	}

	return nil
}

// MigrateConfig migrates configuration from legacy format
func (s *Service) MigrateConfig(legacyConfigPath string) (*domainConfig.ApplicationConfig, error) {
	// Use loader to handle migration
	if s.loader == nil {
		s.loader = domainConfig.NewLoader()
	}
	return s.loader.Load(legacyConfigPath)
}

// LoadYAMLFile loads a YAML file into the provided config structure
func (s *Service) LoadYAMLFile(filePath string, config interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, config)
}

// GetRagConfig returns the RAG configuration with defaults
// RAG config is loaded via includes in the main config
func (s *Service) GetRagConfig() *domainConfig.RagConfig {
	// If already loaded in application config, return it
	if s.config != nil && s.config.RAG != nil {
		return s.config.RAG
	}
	
	// Return defaults if not configured
	return &domainConfig.RagConfig{
		DefaultFusion: "rrf",
		DefaultTopK:   5,
		QueryExpansion: domainConfig.QueryExpansionSettings{
			Enabled:       true,
			MaxExpansions: 5,
			CaseSensitive: false,
		},
		Fusion: domainConfig.FusionSettings{
			RRF: domainConfig.RRFSettings{K: 60},
			Weighted: domainConfig.WeightedSettings{Normalize: true},
		},
		Servers: make(map[string]domainConfig.RagServerConfig),
	}
}

