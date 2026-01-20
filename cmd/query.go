package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/query"
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
	queryInputData string // Query-specific input data flag
)

// QueryCmd represents the query command
var QueryCmd = &cobra.Command{
	Use:   "query [question] or --input-data \"question\"",
	Short: "Ask a single question and get a response",
	Long: `Query mode asks a single question to the AI model and returns a response
without entering an interactive session. Perfect for scripting, automation,
and integration with other tools.

The question can be provided either as:
  • Positional argument: query "question"
  • --input-data flag: query --input-data "question"
  • stdin: echo "question" | query --input-data -

The query command supports:
  • Multiple MCP servers for tool access
  • Context from files (--context)
  • Custom system prompts (--system-prompt)
  • JSON output for parsing (--json)
  • Raw tool data output (--raw-data)
  • File output (--output)

Examples:
  # Basic query
  mcp-cli query "What is the current time?"
  
  # With specific servers and provider
  mcp-cli query --server filesystem,brave-search \
    --provider openai --model gpt-4o \
    "Search for MCP information and summarize"
  
  # With context file
  mcp-cli query --context context.txt \
    --system-prompt "You are a coding assistant" \
    "How do I implement a binary tree in Go?"
  
  # JSON output for parsing
  mcp-cli query --json "List the top 5 cloud providers" > results.json
  
  # Verbose mode (show all operations)
  mcp-cli query --noisy "What files are in this directory?"
  
  # Raw tool data (bypass AI summarization)
  mcp-cli query --raw-data "Show latest security incidents"
  
  # Output to file
  mcp-cli query "Analyze this code" --output analysis.txt
  
  # Using --input-data flag instead of positional argument
  mcp-cli query --input-data "What is the weather today?"
  
  # Both work the same way
  mcp-cli query "question" --provider anthropic
  mcp-cli query --provider anthropic --input-data "question"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// ARCHITECTURAL FIX: Handle noisy flag override for query command
		// This allows --noisy to override the default quiet behavior of query mode
		if noisy && !verbose {
			// Noisy flag: Show INFO and above logs (but not debug unless --verbose is also used)
			logging.SetDefaultLevel(logging.INFO)
			logging.Info("Noisy mode enabled for query command")
		}
		
		// Input validation
		if maxTokens != 0 && maxTokens < 1 {
			if errorCodeOnly {
				os.Exit(query.ErrInvalidArgumentCode)
			}
			return fmt.Errorf("--max-tokens must be positive, got %d", maxTokens)
		}
		
		// Validate context file exists if specified
		if contextFile != "" {
			if _, err := os.Stat(contextFile); os.IsNotExist(err) {
				if errorCodeOnly {
					os.Exit(query.ErrContextNotFoundCode)
				}
				return fmt.Errorf("context file does not exist: %s", contextFile)
			}
		}
		
		// Validate output file path is writable (check parent directory)
		if outputFile != "" {
			// Extract directory from output file path
			outputDir := outputFile
			if idx := strings.LastIndex(outputFile, "/"); idx != -1 {
				outputDir = outputFile[:idx]
			} else if idx := strings.LastIndex(outputFile, "\\"); idx != -1 {
				outputDir = outputFile[:idx]
			} else {
				outputDir = "." // Current directory
			}
			
			// Check if directory exists and is writable
			if stat, err := os.Stat(outputDir); err != nil {
				if errorCodeOnly {
					os.Exit(query.ErrOutputWriteCode)
				}
				return fmt.Errorf("output directory does not exist: %s", outputDir)
			} else if !stat.IsDir() {
				if errorCodeOnly {
					os.Exit(query.ErrOutputWriteCode)
				}
				return fmt.Errorf("output path is not a directory: %s", outputDir)
			}
		}
		
		// Get question from either positional args, query-specific --input-data, or root --input-data flag
		var question string
		if len(args) > 0 {
			// Use positional arguments if provided
			question = strings.Join(args, " ")
		} else if queryInputData != "" {
			// Use query-specific --input-data flag
			question = queryInputData
		} else if inputData != "" {
			// Fall back to root-level --input-data flag (for backward compatibility)
			question = inputData
		} else {
			// No question provided - show enhanced error
			cliErr := NewMissingArgumentError("question", "query", []string{
				`mcp-cli query "What is the capital of France?"`,
				`mcp-cli query --input-data "What is the capital of France?"`,
				`echo "What is the capital of France?" | mcp-cli query --input-data -`,
			})
			fmt.Fprintln(os.Stderr, cliErr.Format())
			
			// Exit immediately with proper code
			if errorCodeOnly {
				os.Exit(query.ErrInvalidArgumentCode)
			}
			os.Exit(1)
		}

		// Process server configuration options - use local ProcessOptions with configFile
		serverNames, userSpecified := ProcessOptions(configFile, serverName, disableFilesystem, providerName, modelName)
		logging.Debug("Server names: %v", serverNames)
		logging.Debug("Using provider from config: %s", providerName)

		// FIXED: Use enhanced AI options to support interface-based config format
		enhancedAIOptions, err := host.GetEnhancedAIOptions(configFile, providerName, modelName)
		if err != nil {
			if errorCodeOnly {
				os.Exit(query.ErrConfigNotFoundCode)
			}
			return fmt.Errorf("error loading enhanced AI options: %w", err)
		}

		// Convert enhanced options to standard options for backward compatibility
		aiOptions := &host.AIOptions{
			Provider:      enhancedAIOptions.Provider,
			Model:         enhancedAIOptions.Model,
			APIKey:        enhancedAIOptions.APIKey,
			APIEndpoint:   enhancedAIOptions.APIEndpoint,
			InterfaceType: enhancedAIOptions.Interface,
		}

		// Override from command-line flags if specified
		if providerName != "" {
			aiOptions.Provider = providerName
			enhancedAIOptions.Provider = providerName
		}
		if modelName != "" {
			aiOptions.Model = modelName
			enhancedAIOptions.Model = modelName
		}

		// Validate API key - the config system has already loaded and expanded environment variables
		// from the provider config files (e.g., api_key: ${OPENAI_API_KEY})
		// If the API key is still empty, provide a helpful error message
		if aiOptions.APIKey == "" {
			// Note: Some providers (Ollama, etc.) don't require API keys
			// The provider will return its own error if authentication fails
			logging.Debug("No API key configured for provider %s (may not be required)", aiOptions.Provider)
		}

		// Load context file if provided
		var contextContent string
		if contextFile != "" {
			content, err := os.ReadFile(contextFile)
			if err != nil {
				if errorCodeOnly {
					os.Exit(query.ErrContextNotFoundCode)
				}
				return fmt.Errorf("failed to read context file: %w", err)
			}
			contextContent = string(content)
		}

		// Load the configuration to check for system prompt and other settings
		oldCfg, err := config.LoadConfig(configFile)
		if err == nil {
			// If system prompt is not provided through command line, check config
			if systemPrompt == "" {
				// First try to get a server-specific prompt if a single server is specified
				if len(serverNames) == 1 {
					configPrompt := oldCfg.GetSystemPrompt(serverNames[0])
					if configPrompt != "" {
						systemPrompt = configPrompt
						logging.Debug("Using system prompt from config for server: %s", serverNames[0])
					}
				}
				
				// If no server-specific prompt, try the default system prompt
				if systemPrompt == "" {
					if oldCfg.AI != nil && oldCfg.AI.DefaultSystemPrompt != "" {
						systemPrompt = oldCfg.AI.DefaultSystemPrompt
						logging.Debug("Using default system prompt from config")
					}
				}
			}
		}

		// Check for raw data output setting in config file
		// This allows us to override the command-line flag
		serverRawDataOverride := make(map[string]bool)
		if oldCfg != nil {
			// Check for global settings
			settings := oldCfg.GetSettings()
			if settings != nil && settings.RawDataOverride {
				rawDataOutput = true
				logging.Debug("Raw data output enabled from global settings")
			}
			
			// Check for server-specific settings
			for _, name := range serverNames {
				serverSettings, err := oldCfg.GetServerSettings(name)
				if err == nil && serverSettings != nil && serverSettings.RawDataOverride {
					serverRawDataOverride[name] = true
					logging.Debug("Raw data output enabled for server: %s", name)
				}
			}
		}

		// ARCHITECTURAL FIX: Choose command options based on verbosity for clean output
		var commandOptions *host.CommandOptions
		if noisy || verbose {
			// Show connection messages and preserve server errors
			commandOptions = host.DefaultCommandOptions()
		} else {
			// DEFAULT: Clean user output (suppress console messages) but preserve server error handling
			commandOptions = host.QuietCommandOptions()
		}

		// Run the query command with the given options
		var result *query.QueryResult
		err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
			// Use AI service to create provider with full config
			aiService := ai.NewService()
			llmProvider, err := aiService.InitializeProvider(configFile, providerName, modelName)
			if err != nil {
				if errorCodeOnly {
					os.Exit(query.ErrInitializationCode)
				}
				return fmt.Errorf("failed to initialize AI provider: %w", err)
			}

			// Create query handler with pre-created provider
			handler, err := query.NewQueryHandlerWithProvider(conns, llmProvider, aiOptions, systemPrompt)
			if err != nil {
				if errorCodeOnly {
					os.Exit(query.ErrInitializationCode)
				}
				return fmt.Errorf("failed to initialize query: %w", err)
			}

			// Set context if provided
			if contextContent != "" {
				handler.AddContext(contextContent)
			}

			// Set max tokens if provided
			if maxTokens > 0 {
				handler.SetMaxTokens(maxTokens)
			}

			// Execute the query
			result, err = handler.Execute(question)
			if err != nil {
				// Return specific error code based on the error type
				if errorCodeOnly {
					exitCode := query.GetExitCode(err)
					os.Exit(exitCode)
				}
				return fmt.Errorf("query failed: %w", err)
			}

			return nil
		}, configFile, serverNames, userSpecified, commandOptions)

		if err != nil {
			return err
		}

		// Process the results if raw data output is enabled
		if result != nil && len(result.ToolCalls) > 0 {
			// Check if we need to use raw data output
			applyRawDataOutput := rawDataOutput
			
			// Also check for server-specific overrides
			for _, conn := range result.ServerConnections {
				if serverRawDataOverride[conn] {
					applyRawDataOutput = true
					break
				}
			}
			
			// Apply raw data output if needed
			if applyRawDataOutput {
				rawData := extractRawData(result.ToolCalls)
				if rawData != "" {
					// Replace the AI response with the raw data
					result.Response = rawData
				}
			}
		}

		// Format and output response
		if result != nil {
			if jsonOutput {
				// Output as JSON
				jsonData, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					if errorCodeOnly {
						os.Exit(query.ErrOutputFormatCode)
					}
					return fmt.Errorf("failed to format JSON response: %w", err)
				}

				// Write to file or stdout
				if outputFile != "" {
					err = os.WriteFile(outputFile, jsonData, 0644)
					if err != nil {
						if errorCodeOnly {
							os.Exit(query.ErrOutputWriteCode)
						}
						return fmt.Errorf("failed to write output file: %w", err)
					}
				} else {
					fmt.Println(string(jsonData))
				}
			} else {
				// Output as plain text
				if outputFile != "" {
					err = os.WriteFile(outputFile, []byte(result.Response), 0644)
					if err != nil {
						if errorCodeOnly {
							os.Exit(query.ErrOutputWriteCode)
						}
						return fmt.Errorf("failed to write output file: %w", err)
					}
				} else {
					// Use platform-aware output writer
					writer := output.NewWriter()
					defer writer.Close()
					writer.Println(result.Response)
				}
			}
		}

		return nil
	},
}

// ProcessOptions processes command-line options and returns the server names
func ProcessOptions(configFile, serverFlag string, disableFilesystem bool, provider string, model string) ([]string, map[string]bool) {
	logging.Debug("Processing options: server=%s, disableFilesystem=%v, provider=%s, model=%s",
		serverFlag, disableFilesystem, provider, model)
	
	// Parse the server list
	serverNames := []string{}
	if serverFlag != "" {
		// Split comma-separated list
		for _, s := range strings.Split(serverFlag, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				serverNames = append(serverNames, trimmed)
			}
		}
	}

	// If no servers specified and filesystem not disabled, load ALL servers from config
	if len(serverNames) == 0 && !disableFilesystem {
		// Use the new modular config service to load all servers
		configService := config.NewService()
		appConfig, err := configService.LoadConfig(configFile)
		if err == nil && appConfig != nil && len(appConfig.Servers) > 0 {
			// Add ALL configured servers
			for serverName := range appConfig.Servers {
				serverNames = append(serverNames, serverName)
				logging.Debug("Adding server from config: %s", serverName)
			}
			logging.Info("Loaded %d server(s) from config", len(serverNames))
		} else {
			logging.Debug("No servers found in config or config load failed")
		}
		// If no servers in config or config load failed, proceed with empty list
		// This allows query to work without MCP servers
	}

	// Create a map of user-specified servers
	userSpecified := make(map[string]bool)
	if serverFlag != "" {
		// Only servers explicitly specified via --server flag are marked as user-specified
		for _, name := range serverNames {
			userSpecified[name] = true
		}
	}
	// Auto-loaded servers from config are NOT marked as user-specified

	logging.Debug("Server names: %v", serverNames)
	return serverNames, userSpecified
}

// extractRawData extracts raw data from tool calls
func extractRawData(toolCalls []query.ToolCallInfo) string {
	if len(toolCalls) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString("RAW TOOL DATA:\n------------------------\n\n")
	
	for i, tc := range toolCalls {
		if tc.Success {
			result.WriteString(fmt.Sprintf("Tool Call #%d: %s\n", i+1, tc.Name))
			result.WriteString("Result:\n")
			
			// Try to format the result if it's JSON
			formattedResult := formatToolResult(tc.Result)
			if formattedResult != "" {
				result.WriteString(formattedResult)
			} else {
				result.WriteString(tc.Result)
			}
			
			result.WriteString("\n\n")
		}
	}
	
	return result.String()
}

// formatToolResult attempts to format JSON tool results
func formatToolResult(resultStr string) string {
	// First check if it contains a JSON object
	jsonStart := strings.Index(resultStr, "{")
	if jsonStart < 0 {
		return ""
	}
	
	// Try to parse and format the JSON
	var data interface{}
	err := json.Unmarshal([]byte(resultStr[jsonStart:]), &data)
	if err != nil {
		return ""
	}
	
	// Format the result based on type
	switch v := data.(type) {
	case map[string]interface{}:
		return formatJsonObject(v, 0)
	default:
		return ""
	}
}

// formatJsonObject formats a JSON object with indentation
func formatJsonObject(obj map[string]interface{}, indent int) string {
	var result strings.Builder
	indentStr := strings.Repeat("  ", indent)
	
	// Special handling for security incident data
	if val, ok := obj["result"].(map[string]interface{}); ok {
		if incidents, ok := val["value"].([]interface{}); ok {
			// Found security incidents, format them nicely
			result.WriteString(fmt.Sprintf("%sFound %d security incidents:\n\n", indentStr, len(incidents)))
			
			for i, inc := range incidents {
				if incident, ok := inc.(map[string]interface{}); ok {
					result.WriteString(fmt.Sprintf("%sIncident %d:\n", indentStr, i+1))
					
					// Format each field
					for field, value := range incident {
						result.WriteString(fmt.Sprintf("%s- %s: %v\n", indentStr+"  ", field, value))
					}
					result.WriteString("\n")
				}
			}
			
			return result.String()
		}
	}
	
	// Generic object formatting
	for key, value := range obj {
		result.WriteString(fmt.Sprintf("%s%s: ", indentStr, key))
		
		switch v := value.(type) {
		case map[string]interface{}:
			result.WriteString("\n")
			result.WriteString(formatJsonObject(v, indent+1))
		case []interface{}:
			result.WriteString("\n")
			for i, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					result.WriteString(fmt.Sprintf("%s  [%d]:\n", indentStr, i))
					result.WriteString(formatJsonObject(mapItem, indent+2))
				} else {
					result.WriteString(fmt.Sprintf("%s  [%d]: %v\n", indentStr, i, item))
				}
			}
		default:
			result.WriteString(fmt.Sprintf("%v\n", v))
		}
	}
	
	return result.String()
}

func init() {
	// Add query-specific flags
	QueryCmd.Flags().StringVar(&queryInputData, "input-data", "", "Question to ask (alternative to positional argument)")
	QueryCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output response in JSON format")
	QueryCmd.Flags().StringVarP(&contextFile, "context", "c", "", "File containing additional context")
	QueryCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Custom system prompt")
	QueryCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "Maximum tokens in response (0 for default)")
	QueryCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default is stdout)")
	QueryCmd.Flags().BoolVar(&errorCodeOnly, "error-code-only", false, "Only return error codes, no error messages")
	QueryCmd.Flags().BoolVarP(&noisy, "noisy", "n", false, "Show detailed logs and server messages")
	QueryCmd.Flags().BoolVar(&rawDataOutput, "raw-data", false, "Output raw data from tools instead of AI summary")

	// Note: QueryCmd is added to RootCmd in root.go init() with other commands
}
