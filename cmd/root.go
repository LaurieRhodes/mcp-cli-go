package cmd

import (
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/spf13/cobra"
)

var (
	// Configuration options
	configFile        string
	serverName        string
	providerName      string
	modelName         string
	disableFilesystem bool
	verbose           bool
	noColor           bool
	
	// Template-based workflow flags
	templateName      string
	inputData         string
	listTemplates     bool

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "mcp-cli",
		Short: "MCP Command-Line Tool",
		Long: `MCP Command-Line Tool - A protocol-level CLI designed to interact with Model Context Provider servers.
The client allows users to send commands, query data, and interact with various resources provided by the server.

Workflow Templates:
  Use --template to execute predefined workflow templates that can chain multiple AI requests
  with different providers and pass data between steps.

Examples:
  mcp-cli --template "analyze_file"
  mcp-cli --template "search_and_summarize" --input-data "AI trends"
  mcp-cli --list-templates

First Time Setup:
  If no config file exists, an example will be created automatically with:
  - Sample servers (filesystem, brave-search)  
  - Multiple AI provider options (OpenAI, Anthropic, Ollama)
  - Example workflow templates for file analysis and web research`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set up color output based on --no-color flag
			if noColor {
				logging.SetColorOutput(false)
				logging.Debug("Color output disabled via --no-color flag")
			}
			
			// Different logging behavior based on command and flags
			isQueryCommand := cmd.Name() == "query"
			isTemplateMode := templateName != ""
			
			if verbose {
				// Verbose flag always enables debug logging regardless of command
				logging.SetDefaultLevel(logging.DEBUG)
				logging.Debug("Debug logging enabled via --verbose flag")
			} else if isQueryCommand || isTemplateMode {
				// Query mode and Template mode: Clean output for workflow automation (ERROR level only)
				// This suppresses INFO and DEBUG logs but preserves critical errors
				logging.SetDefaultLevel(logging.ERROR)
			} else {
				// Other commands (chat, interactive): Normal INFO level
				logging.SetDefaultLevel(logging.INFO)
			}
			
			// Try to load default provider from config if not specified
			if providerName == "" {
				// Try to load from config using auto-generation if needed
				configService := config.NewService()
				if appConfig, _, err := configService.LoadConfigOrCreateExample(configFile); err == nil {
					if appConfig.AI != nil && appConfig.AI.DefaultProvider != "" {
						providerName = appConfig.AI.DefaultProvider
						logging.Debug("Using default provider from config: %s", providerName)
					}
				}
			}
		},
		// If no subcommand is provided, check for template mode or run chat command
		Run: func(cmd *cobra.Command, args []string) {
			// Check for list templates flag
			if listTemplates {
				if err := executeListTemplates(); err != nil {
					logging.Error("Failed to list templates: %v", err)
					os.Exit(1)
				}
				return
			}
			
			// Check for template execution
			if templateName != "" {
				if err := executeTemplate(); err != nil {
					logging.Error("Template execution failed: %v", err)
					os.Exit(1)
				}
				return
			}
			
			// Execute the chat command by default
			if err := ChatCmd.RunE(cmd, args); err != nil {
				os.Exit(1)
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	// Global flags
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "server_config.json", "Path to the JSON configuration file")
	RootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "", "Specifies specific server(s) to use (comma-separated). If not specified, uses all configured servers.")
	RootCmd.PersistentFlags().StringVarP(&providerName, "provider", "p", "", "Specifies the AI provider to use (openai, anthropic, ollama)")
	RootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "Specifies the model to use")
	RootCmd.PersistentFlags().BoolVar(&disableFilesystem, "disable-filesystem", false, "Disable filesystem access for the LLM")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	RootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	
	// Template-based workflow flags
	RootCmd.PersistentFlags().StringVar(&templateName, "template", "", "Execute a specific workflow template")
	RootCmd.PersistentFlags().StringVar(&inputData, "input-data", "", "Input data for the workflow template (can be JSON or plain text)")
	RootCmd.PersistentFlags().BoolVar(&listTemplates, "list-templates", false, "List all available workflow templates")

	// Add subcommands
	RootCmd.AddCommand(ChatCmd)
	RootCmd.AddCommand(InteractiveCmd)
	RootCmd.AddCommand(QueryCmd)

	// Configuration-based initialization with auto-generation
	cobra.OnInitialize(func() {
		// Only load provider and model if not already specified on command line
		if providerName == "" || modelName == "" {
			// Try to load from config using auto-generation if needed
			configService := config.NewService()
			if appConfig, _, err := configService.LoadConfigOrCreateExample(configFile); err == nil && appConfig.AI != nil {
				// If provider not specified, use from config
				if providerName == "" && appConfig.AI.DefaultProvider != "" {
					providerName = appConfig.AI.DefaultProvider
					logging.Debug("Loaded default provider from config: %s", providerName)
				}
				
				// If model not specified and we have a provider, try to get default model
				if modelName == "" && providerName != "" {
					// Try to get provider config
					if providerConfig, _, err := configService.GetProviderConfig(providerName); err == nil && providerConfig.DefaultModel != "" {
						modelName = providerConfig.DefaultModel
						logging.Debug("Loaded default model from config for %s: %s", providerName, modelName)
					}
				}
			}
		}
	})
}
