package cmd

import (
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/chat"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ChatCmd represents the unified chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Enter interactive chat mode with the LLM",
	Long: `Chat mode provides a conversational interface with the LLM and is the primary way to interact with the client.
The LLM can execute queries, access data, and leverage other capabilities provided by the server.

This command uses the modern interface-based approach for LLM providers, supporting all configured
provider types including OpenAI, Anthropic, Ollama, and others.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse command configuration
		chatConfig := parseChatConfig(cmd, args)

		// Get output manager
		outputMgr := output.GetGlobalManager()

		// Only show startup info if verbose
		if outputMgr.ShouldShowStartupInfo() {
			bold := color.New(color.Bold)
			serversText := strings.Join(chatConfig.ServerNames, ", ")
			if serversText == "" {
				serversText = "none"
			}
			bold.Printf("Starting chat mode with servers: %s, provider: %s, model: %s\n\n",
				serversText, chatConfig.ProviderName, chatConfig.ModelName)
		}

		// Create chat service and start chat
		chatService := chat.NewService()
		return chatService.StartChat(chatConfig)
	},
}

// parseChatConfig parses command line arguments into chat service config
func parseChatConfig(_ *cobra.Command, _ []string) *chat.Config {
	// Process server configuration options - pass configFile
	serverNames, userSpecified := host.ProcessOptions(configFile, serverName, disableFilesystem, providerName, modelName)

	// Parse skill names if provided
	var skillNamesSlice []string
	if skillNames != "" {
		for _, s := range strings.Split(skillNames, ",") {
			skillNamesSlice = append(skillNamesSlice, strings.TrimSpace(s))
		}
	}

	return &chat.Config{
		ConfigFile:        configFile,
		ServerName:        serverName,
		ProviderName:      providerName,
		ModelName:         modelName,
		DisableFilesystem: disableFilesystem,
		ServerNames:       serverNames,
		UserSpecified:     userSpecified,
		SkillNames:        skillNamesSlice,
	}
}

func init() {
	// Chat command doesn't need additional flags beyond the global ones
	// All configuration is handled through global flags and config files
}
