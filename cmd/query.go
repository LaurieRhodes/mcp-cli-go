package cmd

import (
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	queryservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/query"
	"github.com/spf13/cobra"
)

var (
	// Query-specific flags
	jsonOutput     bool
	contextFile    string
	systemPrompt   string
	maxTokens      int
	outputFile     string
	errorCodeOnly  bool
	noisy          bool   // Changed to be the opposite of quiet
	rawDataOutput  bool   // New flag for raw data output
)

// QueryCmd represents the query command
var QueryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Ask a single question and get a response",
	Long: `Query mode asks a single question to the AI model and receives a response without entering an interactive session.
It's designed for scripting, workflow automation, and multi-agent scenarios.

Example usage:
  mcp-cli query "What is the current time?"
  mcp-cli query --server filesystem,brave-search "Search for information about Model Context Protocol and summarize key points"
  mcp-cli query --context context.txt --system-prompt "You are a helpful coding assistant" "How do I implement a binary tree in Go?"
  mcp-cli query --json "List the top 5 cloud providers" > results.json
  mcp-cli query --noisy "What files are in this directory?" # Shows verbose logs and server messages
  mcp-cli query --raw-data "what are the top 10 latest low severity security incidents" # Shows raw data from tools`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse command configuration
		config := parseQueryConfig(cmd, args)

		// Create query service and execute
		queryService := queryservice.NewService()  
		return queryService.ExecuteQuery(config)
	},
}

// parseQueryConfig parses command line arguments into query service config
func parseQueryConfig(cmd *cobra.Command, args []string) *queryservice.Config {
	// Combine all args into a single question
	question := strings.Join(args, " ")

	// Process server configuration options
	serverNames, userSpecified := host.ProcessOptions(serverName, disableFilesystem, providerName, modelName)

	return &queryservice.Config{
		ConfigFile:        configFile,
		ServerName:        serverName,
		ProviderName:      providerName,
		ModelName:         modelName,
		DisableFilesystem: disableFilesystem,
		ServerNames:       serverNames,
		UserSpecified:     userSpecified,
		
		// Query-specific configuration
		Question:       question,
		JSONOutput:     jsonOutput,
		ContextFile:    contextFile,
		SystemPrompt:   systemPrompt,
		MaxTokens:      maxTokens,
		OutputFile:     outputFile,
		ErrorCodeOnly:  errorCodeOnly,
		Noisy:          noisy,
		RawDataOutput:  rawDataOutput,
	}
}

func init() {
	// Add query-specific flags
	QueryCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output response in JSON format")
	QueryCmd.Flags().StringVarP(&contextFile, "context", "c", "", "File containing additional context")
	QueryCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Custom system prompt")
	QueryCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "Maximum tokens in response (0 for default)")
	QueryCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default is stdout)")
	QueryCmd.Flags().BoolVar(&errorCodeOnly, "error-code-only", false, "Only return error codes, no error messages")
	QueryCmd.Flags().BoolVarP(&noisy, "noisy", "n", false, "Show detailed logs and server messages")
	QueryCmd.Flags().BoolVar(&rawDataOutput, "raw-data", false, "Output raw data from tools instead of AI summary")

	// Command is added to root in root.go
}
