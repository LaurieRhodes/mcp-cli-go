package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/env"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// getColorizedHelp returns a colorized help message for the CLI
func getColorizedHelp() string {
	// Define colors
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	magenta := color.New(color.FgMagenta, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	
	header := cyan("================================================================================") + "\n" +
		cyan("                          MCP Command-Line Tool") + "\n" +
		cyan("          Protocol-level CLI for Model Context Provider servers") + "\n" +
		cyan("================================================================================") + "\n\n" +
		white("A versatile command-line interface for interacting with AI models through the\n") +
		white("Model Context Protocol (MCP). Supports multiple AI providers, workflow templates,\n") +
		white("embeddings generation, and can run as an MCP server itself.\n\n")
	
	setup := yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + magenta("First Time Setup") + yellow("                                                           |\n") +
		yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + green("mcp-cli init --quick") + "         Quick setup (30 seconds)                     " + yellow("|\n") +
		yellow("| ") + green("mcp-cli init") + "                 Interactive guided setup                     " + yellow("|\n") +
		yellow("| ") + green("mcp-cli init --full") + "          Complete setup with all options              " + yellow("|\n") +
		yellow("+----------------------------------------------------------------------------+\n\n")
	
	usage := blue("+----------------------------------------------------------------------------+\n") +
		blue("| ") + magenta("Basic Usage") + blue("                                                                |\n") +
		blue("+----------------------------------------------------------------------------+\n") +
		blue("| ") + green("mcp-cli") + "                      Start interactive chat (default)             " + blue("|\n") +
		blue("| ") + green("mcp-cli chat") + "                 Explicitly start chat mode                   " + blue("|\n") +
		blue("| ") + green("mcp-cli query \"question\"") + "     Ask a single question                        " + blue("|\n") +
		blue("| ") + green("mcp-cli interactive") + "          Interactive mode with slash commands         " + blue("|\n") +
		blue("|                                                                            |\n") +
		blue("| ") + yellow("For query options: ") + green("mcp-cli query --help") + "                                   " + blue("|\n") +
		blue("+----------------------------------------------------------------------------+\n\n")
	
	templates := cyan("+----------------------------------------------------------------------------+\n") +
		cyan("| ") + magenta("Workflow Templates") + cyan("                                                         |\n") +
		cyan("+----------------------------------------------------------------------------+\n") +
		cyan("| ") + white("Templates chain multiple AI requests with different providers and pass    ") + cyan("|\n") +
		cyan("| ") + white("data between steps for complex, automated workflows.                      ") + cyan("|\n") +
		cyan("|                                                                            |\n") +
		cyan("| ") + green("mcp-cli workflows") + "                        List available workflows          " + cyan("|\n") +
		cyan("| ") + green("mcp-cli --workflow analyze") + "              Run 'analyze' workflow            " + cyan("|\n") +
		cyan("| ") + green("mcp-cli --workflow analyze --input-data \"data\"") + "  With input data           " + cyan("|\n") +
		cyan("| ") + green("echo \"data\" | mcp-cli --workflow analyze") + "        From stdin                " + cyan("|\n") +
		cyan("|                                                                            |\n") +
		cyan("| ") + white("Resume from specific step (skips previous steps, validation still runs):  ") + cyan("|\n") +
		cyan("| ") + green("mcp-cli --workflow pipeline --start-from process") + "   Resume from 'process'  " + cyan("|\n") +
		cyan("|                                                                            |\n") +
		cyan("| ") + yellow("Workflow flags: ") + white("--start-from, --end-at, --log-level                      ") + cyan("|\n") +
		cyan("+----------------------------------------------------------------------------+\n\n")
	
	server := yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + magenta("MCP Server Mode") + yellow("                                                            |\n") +
		yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + white("Run mcp-cli as an MCP server, exposing workflow templates as callable     ") + yellow("|\n") +
		yellow("| ") + white("tools that other applications (like Claude Desktop) can use.              ") + yellow("|\n") +
		yellow("|                                                                            |\n") +
		yellow("| ") + green("mcp-cli serve config/runas/agent.yaml") + "   Start MCP server                  " + yellow("|\n") +
		yellow("| ") + green("mcp-cli serve --verbose agent.yaml") + "      With detailed logging             " + yellow("|\n") +
		yellow("+----------------------------------------------------------------------------+\n\n")
	
	embeddings := blue("+----------------------------------------------------------------------------+\n") +
		blue("| ") + magenta("Embeddings & Vector Search") + blue("                                                 |\n") +
		blue("+----------------------------------------------------------------------------+\n") +
		blue("| ") + green("mcp-cli embeddings \"text\"") + "                    Generate embeddings          " + blue("|\n") +
		blue("| ") + green("mcp-cli embeddings --input-file doc.txt") + "      From file                    " + blue("|\n") +
		blue("| ") + green("mcp-cli embeddings --model text-embedding-3-large") + "  Specific model         " + blue("|\n") +
		blue("| ") + green("echo \"text\" | mcp-cli embeddings") + "             From stdin                   " + blue("|\n") +
		blue("|                                                                            |\n") +
		blue("| ") + yellow("For chunking, output options: ") + green("mcp-cli embeddings --help") + "                   " + blue("|\n") +
		blue("+----------------------------------------------------------------------------+\n\n")
	
	rag := yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + magenta("RAG Operations") + yellow("                                                             |\n") +
		yellow("+----------------------------------------------------------------------------+\n") +
		yellow("| ") + green("mcp-cli rag search \"query\" --server pgvector --top-k 10") + "                   " + yellow("|\n") +
		yellow("| ") + green("mcp-cli rag search \"query\" --strategies default,hybrid --fusion rrf") + "       " + yellow("|\n") +
		yellow("| ") + green("mcp-cli rag config") + "                           Show RAG configuration       " + yellow("|\n") +
		yellow("|                                                                            |\n") +
		yellow("| ") + yellow("For all RAG options: ") + green("mcp-cli rag --help") + "                                   " + yellow("|\n") +
		yellow("+----------------------------------------------------------------------------+\n\n")
	
	configSection := cyan("+----------------------------------------------------------------------------+\n") +
		cyan("| ") + magenta("Configuration") + cyan("                                                              |\n") +
		cyan("+----------------------------------------------------------------------------+\n") +
		cyan("| ") + green("mcp-cli config validate") + "                      Validate configuration       " + cyan("|\n") +
		cyan("| ") + green("mcp-cli config --help") + "                        See all config commands      " + cyan("|\n") +
		cyan("+----------------------------------------------------------------------------+")
	
	return header + setup + usage + templates + server + embeddings + rag + configSection
}

var (
	// Configuration options
	configFile        string
	serverName        string
	providerName      string
	modelName         string
	skillNames        string
	disableFilesystem bool
	verbose           bool
	logLevel          string
	noColor           bool
	
	// Template-based workflow flags
	workflowName      string
	startFromStep     string
	endAtStep         string
	inputData         string

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "mcp-cli",
		Short: "MCP Command-Line Tool - Interact with AI models and MCP servers",
		Long:  getColorizedHelp(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip config check for init command, help, and serve (serve handles config loading internally)
			cmdName := cmd.Name()
			if cmdName == "init" || cmdName == "help" || cmdName == "completion" || cmdName == "serve" {
				return
			}
			
			// Check if config exists (except for init command)
			checkConfigExists(configFile)
			
			// Determine output configuration based on command and flags
			var outputConfig *models.OutputConfig
			
			isQueryCommand := cmd.Name() == "query"
			isTemplateMode := workflowName != ""
			isEmbeddingsCommand := cmd.Name() == "embeddings"
			
			if verbose {
				// Verbose flag: Show everything
				outputConfig = models.NewVerboseOutputConfig()
			} else if isQueryCommand || isTemplateMode || isEmbeddingsCommand {
				// Query/template/embeddings: Quiet mode for clean output
				outputConfig = models.NewQuietOutputConfig()
			} else {
				// Chat and other commands: Normal mode
				outputConfig = models.NewDefaultOutputConfig()
			}
			
			// Apply no-color flag
			if noColor {
				outputConfig.ShowColors = false
			}
			
			// Set global output manager
			outputManager := output.NewManager(outputConfig)
			output.SetGlobalManager(outputManager)
			
			// Configure legacy logging to match output level
			configureLegacyLogging(outputConfig)
			
			// Try to load default provider from config if not specified
			if providerName == "" {
				configService := config.NewService()
				if appConfig, err := configService.LoadConfig(configFile); err == nil {
					if appConfig.AI != nil && appConfig.AI.DefaultProvider != "" {
						providerName = appConfig.AI.DefaultProvider
						logging.Debug("Using default provider from config: %s", providerName)
					}
				}
			}
		},
		// If no subcommand is provided, check for template mode or run chat command
		Run: func(cmd *cobra.Command, args []string) {
			// Check for template execution
			if workflowName != "" {
				if err := executeWorkflow(); err != nil {
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

// setupErrorHandlers configures custom error handling for better UX
func setupErrorHandlers() {
	// Custom flag error handler
	RootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		errStr := err.Error()

		// Handle unknown flag errors
		if strings.Contains(errStr, "unknown flag:") || strings.Contains(errStr, "unknown shorthand flag:") {
			flagName := ExtractFlagName(errStr)
			if flagName != "" {
				cliErr := NewUnknownFlagError(flagName, cmd)
				fmt.Fprintln(os.Stderr, cliErr.Format())
				os.Exit(1)
			}
		}

		// Fall back to default behavior
		return err
	})

	// Enable command suggestions with low threshold
	RootCmd.SuggestionsMinimumDistance = 2

	// NOTE: Removed custom UsageFunc that was causing infinite recursion
	// The default cobra usage handler works fine
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	// Load .env file first - before any config loading
	// This ensures environment variables are available for config substitution
	// Silently ignores if .env doesn't exist
	_ = env.LoadDotEnv()
	
	// Ensure standard system paths are in PATH (ALWAYS runs, independent of .env)
	// Fixes issues when running from non-interactive shells (Claude Desktop, systemd, cron)
	// that may have minimal PATH set
	env.EnsureStandardPaths()
	
	// Global flags
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "Path to configuration file (YAML/JSON)")
	RootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "", "MCP server(s) to use (comma-separated, e.g., 'filesystem,brave-search')")
	RootCmd.PersistentFlags().StringVar(&skillNames, "skills", "", "Skill(s) to expose (comma-separated, e.g., 'docx,pdf,xlsx')")
	RootCmd.PersistentFlags().StringVarP(&providerName, "provider", "p", "", "AI provider (openai, anthropic, ollama, deepseek, gemini, openrouter)")
	RootCmd.PersistentFlags().StringVarP(&modelName, "model", "m", "", "Model to use (e.g., gpt-4o, claude-sonnet-4, qwen2.5:32b)")
	RootCmd.PersistentFlags().BoolVar(&disableFilesystem, "disable-filesystem", false, "Disable filesystem server (prevents file access)")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (shortcut for --log-level verbose)")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "Set log level: error, warn, info, step, steps, debug, verbose, noisy (default: info)")
	RootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output (for piping or logging)")
	
	// Template-based workflow flags (only for root command, not subcommands)
	RootCmd.Flags().StringVar(&workflowName, "workflow", "", "Execute workflow by name")
	RootCmd.Flags().StringVar(&startFromStep, "start-from", "", "Start workflow from specific step (skips previous steps)")
	RootCmd.Flags().StringVar(&endAtStep, "end-at", "", "End workflow at specific step (skips steps after)")
	RootCmd.Flags().StringVar(&inputData, "input-data", "", "Input data for template (JSON or plain text)")

	// Custom error handlers for better UX
	setupErrorHandlers()

	// Add subcommands
	RootCmd.AddCommand(ChatCmd)
	RootCmd.AddCommand(InteractiveCmd)
	RootCmd.AddCommand(QueryCmd)
	RootCmd.AddCommand(ServersCmd)
	RootCmd.AddCommand(WorkflowsCmd)  // List workflows
	RootCmd.AddCommand(SkillsCmd)     // List skills
	RootCmd.AddCommand(EmbeddingsCmd)
	RootCmd.AddCommand(RagCmd)  // RAG operations
	RootCmd.AddCommand(ConfigCmd)
	RootCmd.AddCommand(InitCmd)  // Setup wizard
	// Note: ServeCmd is added in serve.go's init() function

	// Configuration-based initialization
	cobra.OnInitialize(func() {
		// Skip if running init command
		if isInitCommand() {
			return
		}
		
		// Only load provider and model if not already specified on command line
		if providerName == "" || modelName == "" {
			configService := config.NewService()
			if appConfig, err := configService.LoadConfig(configFile); err == nil && appConfig.AI != nil {
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
	
	// Set custom help template with colors
	setColoredHelpTemplate()
}

// setColoredHelpTemplate configures colored output for help text
func setColoredHelpTemplate() {
	// Define color functions
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	
	// Custom usage template with colors
	cobra.AddTemplateFunc("styleHeading", func(s string) string {
		return cyan(s)
	})
	cobra.AddTemplateFunc("styleCommand", func(s string) string {
		return green(s)
	})
	cobra.AddTemplateFunc("styleFlag", func(s string) string {
		return yellow(s)
	})
	cobra.AddTemplateFunc("styleExample", func(s string) string {
		return blue(s)
	})
	
	usageTemplate := `{{styleHeading "Usage:"}}{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{styleHeading "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{styleHeading "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{styleHeading "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad (styleCommand .Name) .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{styleHeading "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{styleHeading "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{styleHeading "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	
	RootCmd.SetUsageTemplate(usageTemplate)
}

// configureLegacyLogging configures the legacy logging system to match OutputConfig
func configureLegacyLogging(config *models.OutputConfig) {
	// Map OutputLevel to logging level
	switch config.Level {
	case models.OutputQuiet:
		logging.SetDefaultLevel(logging.ERROR)
	case models.OutputNormal:
		logging.SetDefaultLevel(logging.WARN)
	case models.OutputVerbose:
		logging.SetDefaultLevel(logging.DEBUG)
	}
	
	// Configure color output
	logging.SetColorOutput(config.ShowColors)
}
