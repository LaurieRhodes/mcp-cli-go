package cmd

import (
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/interactive"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// InteractiveCmd represents the interactive command
var InteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Enter interactive mode with slash commands",
	Long: `Interactive mode provides a command-line interface with slash commands for direct interaction with the server.
You can query server information, list available tools and resources, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse command configuration
		config := parseInteractiveConfig(cmd, args)

		// Display start message with all servers
		bold := color.New(color.Bold)
		serversText := strings.Join(config.ServerNames, ", ")
		if serversText == "" {
			serversText = "none"
		}
		bold.Printf("Starting interactive mode with servers: %s, provider: %s, model: %s\n\n", 
			serversText, config.ProviderName, config.ModelName)

		// Create interactive service and start session
		interactiveService := interactive.NewService()
		return interactiveService.StartInteractiveSession(config)
	},
}

// parseInteractiveConfig parses command line arguments into interactive service config
func parseInteractiveConfig(cmd *cobra.Command, args []string) *interactive.Config {
	// Process server configuration options
	serverNames, userSpecified := host.ProcessOptions(serverName, disableFilesystem, providerName, modelName)

	return &interactive.Config{
		ConfigFile:        configFile,
		ServerName:        serverName,
		ProviderName:      providerName,
		ModelName:         modelName,
		DisableFilesystem: disableFilesystem,
		ServerNames:       serverNames,
		UserSpecified:     userSpecified,
	}
}

func init() {
	// Interactive command doesn't need additional flags beyond the global ones
}
