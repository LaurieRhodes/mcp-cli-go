package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

// ConfigValidateCmd validates the configuration file
var ConfigValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validates the configuration file for:
- Syntax errors
- Missing required fields
- Exposed API keys (security check)
- Template validation

Examples:
  mcp-cli config validate
  mcp-cli config validate --config custom-config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Validating configuration...")

		// Load configuration
		configService := config.NewService()
		appConfig, err := configService.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("❌ Failed to load config: %v\n", err)
			return err
		}

		// Validate configuration
		if err := configService.ValidateConfig(appConfig); err != nil {
			fmt.Printf("❌ Configuration validation failed: %v\n", err)
			return err
		}

		fmt.Println("✓ Configuration syntax is valid")

		// Security check: Look for exposed API keys
		hasExposedKeys := false

		// Check AI providers
		if appConfig.AI != nil && appConfig.AI.Interfaces != nil {
			for interfaceType, interfaceConfig := range appConfig.AI.Interfaces {
				for providerName, providerConfig := range interfaceConfig.Providers {
					if isExposedKey(providerConfig.APIKey) {
						fmt.Printf("⚠️  Warning: API key for %s/%s appears to be hardcoded\n",
							interfaceType, providerName)
						fmt.Println("   Consider moving to .env file: " + providerName + "_API_KEY")
						hasExposedKeys = true
					}
				}
			}
		}

		// Check embedding providers
		if appConfig.Embeddings != nil && appConfig.Embeddings.Interfaces != nil {
			for interfaceType, interfaceConfig := range appConfig.Embeddings.Interfaces {
				for providerName, providerConfig := range interfaceConfig.Providers {
					if isExposedKey(providerConfig.APIKey) {
						fmt.Printf("⚠️  Warning: Embedding API key for %s/%s appears to be hardcoded\n",
							interfaceType, providerName)
						fmt.Println("   Consider moving to .env file")
						hasExposedKeys = true
					}
				}
			}
		}

		if hasExposedKeys {
			fmt.Println("\n💡 Security Tip:")
			fmt.Println("   1. Create a .env file: cp .env.example .env")
			fmt.Println("   2. Add your keys: OPENAI_API_KEY=sk-...")
			fmt.Println("   3. Update config: \"api_key\": \"${OPENAI_API_KEY}\"")
			fmt.Println("   4. Add .env to .gitignore (already done)")
		} else {
			fmt.Println("✓ No exposed API keys found")
		}

		// Check for .env file
		envPath := ".env"
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			fmt.Println("\n💡 Tip: Create a .env file for API keys")
			fmt.Println("   cp .env.example .env")
		} else {
			fmt.Println("✓ .env file found")
		}

		// Summary
		fmt.Println("\n✅ Configuration is valid!")

		if hasExposedKeys {
			fmt.Println("\n⚠️  However, you should move hardcoded API keys to .env file for security")
			os.Exit(1)
		}

		return nil
	},
}

// isExposedKey checks if an API key appears to be hardcoded (not using env vars)
func isExposedKey(key string) bool {
	if key == "" {
		return false
	}

	// Check if it's an environment variable reference
	if strings.HasPrefix(key, "${") || strings.HasPrefix(key, "$") {
		return false
	}

	// Check if it's a placeholder
	placeholders := []string{
		"your-api-key",
		"your-key-here",
		"replace-me",
		"xxx",
		"REPLACE",
	}

	lowerKey := strings.ToLower(key)
	for _, placeholder := range placeholders {
		if strings.Contains(lowerKey, placeholder) {
			return false
		}
	}

	// If it looks like an actual key (starts with common prefixes)
	keyPrefixes := []string{
		"sk-",     // OpenAI
		"sk-ant-", // Anthropic
		"sk-or-",  // OpenRouter
		"AIza",    // Google/Gemini
	}

	for _, prefix := range keyPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// If it's longer than 20 chars and not a variable, likely a key
	if len(key) > 20 && !strings.Contains(key, "$") {
		return true
	}

	return false
}

func init() {
	ConfigValidateCmd.Flags().StringVar(&configFile, "config", "server_config.yaml", "Path to configuration file")
}
