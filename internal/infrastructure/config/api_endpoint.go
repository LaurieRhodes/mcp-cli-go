package config

import "fmt"

// GetAPIEndpoint returns the API endpoint for the specified provider
func (c *Config) GetAPIEndpoint(providerName string) (string, error) {
	// If no AI config, return the default endpoint
	if c.AI == nil || c.AI.Providers == nil {
		// For Ollama, default to localhost
		if providerName == "ollama" {
			return "http://localhost:11434", nil
		}
		return "", fmt.Errorf("AI providers configuration not found")
	}

	// Get the provider config
	provider, ok := c.AI.Providers[providerName]
	if !ok {
		return "", fmt.Errorf("AI provider %s not found in configuration", providerName)
	}

	// Check if API endpoint is set
	if provider.APIEndpoint == "" {
		// For Ollama, default to localhost
		if providerName == "ollama" {
			return "http://localhost:11434", nil
		}
		return "", fmt.Errorf("API endpoint for provider %s is not set", providerName)
	}

	return provider.APIEndpoint, nil
}
