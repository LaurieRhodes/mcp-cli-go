package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadEnhancedConfig loads the enhanced configuration from the specified file (supports both JSON and YAML)
func LoadEnhancedConfig(file string) (*EnhancedConfig, error) {
	// Use the new Service to load the config
	service := NewService()
	appConfig, err := service.LoadConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// Convert ApplicationConfig to EnhancedConfig for backward compatibility
	enhancedConfig := &EnhancedConfig{
		Servers: make(map[string]ServerConfig),
		AI: &EnhancedAIConfig{
			Interfaces: make(map[InterfaceType]InterfaceConfig),
		},
	}
	
	// Copy servers
	for name, domainServer := range appConfig.Servers {
		// Convert ServerSettings if present
		var settings *ServerSettings
		if domainServer.Settings != nil {
			settings = &ServerSettings{
				MaxToolFollowUp: domainServer.Settings.MaxToolFollowUp,
				// RawDataOverride is not in domain config, leave as default (false)
			}
		}
		
		enhancedConfig.Servers[name] = ServerConfig{
			Command:      domainServer.Command,
			Args:         domainServer.Args,
			Env:          domainServer.Env,
			SystemPrompt: domainServer.SystemPrompt,
			Settings:     settings,
		}
	}
	
	// Copy AI config
	if appConfig.AI != nil {
		enhancedConfig.AI.DefaultProvider = appConfig.AI.DefaultProvider
		
		// Copy interfaces - convert domain types to infrastructure types
		for domainInterfaceType, domainInterfaceConfig := range appConfig.AI.Interfaces {
			// Convert interface type
			var infraInterfaceType InterfaceType
			switch domainInterfaceType {
			case "openai_compatible":
				infraInterfaceType = OpenAICompatible
			case "anthropic_native":
				infraInterfaceType = AnthropicNative
			case "ollama_native":
				infraInterfaceType = OllamaNative
			case "gemini_native":
				// Gemini actually uses OpenAI-compatible interface
				infraInterfaceType = OpenAICompatible
			default:
				infraInterfaceType = InterfaceType(domainInterfaceType)
			}
			
			// Convert providers
			infraProviders := make(map[string]ProviderConfig)
			for providerName, domainProvider := range domainInterfaceConfig.Providers {
				infraProviders[providerName] = ProviderConfig{
					APIKey:          domainProvider.APIKey,
					DefaultModel:    domainProvider.DefaultModel,
					APIEndpoint:     domainProvider.APIEndpoint,
					AvailableModels: domainProvider.AvailableModels,
					TimeoutSeconds:  domainProvider.TimeoutSeconds,
					MaxRetries:      domainProvider.MaxRetries,
				}
			}
			
			// Merge into existing interface if it already exists (e.g., gemini merging into openai_compatible)
			if existing, ok := enhancedConfig.AI.Interfaces[infraInterfaceType]; ok {
				for name, provider := range infraProviders {
					existing.Providers[name] = provider
				}
				enhancedConfig.AI.Interfaces[infraInterfaceType] = existing
			} else {
				enhancedConfig.AI.Interfaces[infraInterfaceType] = InterfaceConfig{
					Providers: infraProviders,
				}
			}
		}
		
		// Copy legacy providers if any
		if appConfig.AI.Providers != nil {
			enhancedConfig.AI.Providers = make(map[string]ProviderConfig)
			for providerName, domainProvider := range appConfig.AI.Providers {
				enhancedConfig.AI.Providers[providerName] = ProviderConfig{
					APIKey:          domainProvider.APIKey,
					DefaultModel:    domainProvider.DefaultModel,
					APIEndpoint:     domainProvider.APIEndpoint,
					AvailableModels: domainProvider.AvailableModels,
					TimeoutSeconds:  domainProvider.TimeoutSeconds,
					MaxRetries:      domainProvider.MaxRetries,
				}
			}
		}
	}
	
	return enhancedConfig, nil
}

// SaveEnhancedConfig saves the enhanced configuration to the specified file (format determined by extension)
func SaveEnhancedConfig(config *EnhancedConfig, file string) error {
	var data []byte
	var err error
	
	// Detect format by file extension
	ext := strings.ToLower(filepath.Ext(file))
	
	if ext == ".yaml" || ext == ".yml" {
		// Save as YAML
		data, err = yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML config: %w", err)
		}
	} else {
		// Save as JSON
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
	}

	// Write the file
	if err := os.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// MigrateConfigFile migrates a config file from legacy format to enhanced format
func MigrateConfigFile(file string) error {
	// Load the legacy config
	legacyConfig, err := LoadConfig(file)
	if err != nil {
		return fmt.Errorf("failed to load legacy config: %w", err)
	}
	
	// Generate the enhanced config
	enhancedConfig := GenerateNewConfig(legacyConfig)
	
	// Save the enhanced config
	return SaveEnhancedConfig(enhancedConfig, file)
}

// GetProviderFromEnhancedConfig retrieves a provider config from the enhanced config
func GetProviderFromEnhancedConfig(config *EnhancedConfig, providerName string) (ProviderConfig, InterfaceType, error) {
	if config == nil || config.AI == nil || config.AI.Interfaces == nil {
		return ProviderConfig{}, "", fmt.Errorf("AI interfaces configuration not found")
	}
	
	// Look for the provider in each interface
	for interfaceType, interfaceConfig := range config.AI.Interfaces {
		if providerConfig, ok := interfaceConfig.Providers[providerName]; ok {
			return providerConfig, interfaceType, nil
		}
	}
	
	// If not found in interfaces, check the legacy providers
	if config.AI.Providers != nil {
		if providerConfig, ok := config.AI.Providers[providerName]; ok {
			// Determine the interface type
			var interfaceType InterfaceType
			switch providerName {
			case "openai", "deepseek", "openrouter", "gemini":
				interfaceType = OpenAICompatible
			case "anthropic":
				interfaceType = AnthropicNative
			case "ollama":
				interfaceType = OllamaNative
			default:
				interfaceType = OpenAICompatible // Default
			}
			
			return providerConfig, interfaceType, nil
		}
	}
	
	return ProviderConfig{}, "", fmt.Errorf("provider %s not found in configuration", providerName)
}

// GetDefaultProviderFromEnhancedConfig gets the default provider from the enhanced config
func GetDefaultProviderFromEnhancedConfig(config *EnhancedConfig) (string, ProviderConfig, InterfaceType, error) {
	if config == nil || config.AI == nil {
		return "", ProviderConfig{}, "", fmt.Errorf("AI configuration not found")
	}
	
	providerName := config.AI.DefaultProvider
	if providerName == "" {
		// Look for the first provider in any interface
		for interfaceType, interfaceConfig := range config.AI.Interfaces {
			for name, config := range interfaceConfig.Providers {
				return name, config, interfaceType, nil
			}
		}
		
		// If no interfaces, check legacy providers
		if config.AI.Providers != nil {
			for name, config := range config.AI.Providers {
				var interfaceType InterfaceType
				switch name {
				case "openai", "deepseek", "openrouter", "gemini":
					interfaceType = OpenAICompatible
				case "anthropic":
					interfaceType = AnthropicNative
				case "ollama":
					interfaceType = OllamaNative
				default:
					interfaceType = OpenAICompatible
				}
				return name, config, interfaceType, nil
			}
		}
		
		return "", ProviderConfig{}, "", fmt.Errorf("no providers found in configuration")
	}
	
	// Get the provider config
	providerConfig, interfaceType, err := GetProviderFromEnhancedConfig(config, providerName)
	if err != nil {
		return "", ProviderConfig{}, "", err
	}
	
	return providerName, providerConfig, interfaceType, nil
}
