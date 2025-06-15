package queryservice

import (
	"fmt"
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/query"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/presentation/console"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
)

// Service handles query command execution
type Service struct {
	formatter *console.QueryFormatter
}

// Config holds configuration for query execution
type Config struct {
	ConfigFile        string
	ServerName        string
	ProviderName      string
	ModelName         string
	DisableFilesystem bool
	ServerNames       []string
	UserSpecified     map[string]bool
	
	// Query-specific configuration
	Question       string
	JSONOutput     bool
	ContextFile    string
	SystemPrompt   string
	MaxTokens      int
	OutputFile     string
	ErrorCodeOnly  bool
	Noisy          bool
	RawDataOutput  bool
}

// NewService creates a new query service
func NewService() *Service {
	return &Service{
		formatter: console.NewQueryFormatter(),
	}
}

// ExecuteQuery executes a query with the given configuration
func (s *Service) ExecuteQuery(config *Config) error {
	// Handle noisy flag override for query command
	if config.Noisy {
		logging.SetDefaultLevel(logging.INFO)
		logging.Info("Noisy mode enabled for query command")
	}

	// Initialize AI service
	aiService := ai.NewService()
	provider, err := aiService.InitializeProvider(config.ConfigFile, config.ProviderName, config.ModelName)
	if err != nil {
		if config.ErrorCodeOnly {
			os.Exit(query.ErrProviderNotFoundCode)
		}
		return fmt.Errorf("failed to initialize AI provider: %w", err)
	}
	defer provider.Close() // Clean up resources

	// Load context file if provided
	var contextContent string
	if config.ContextFile != "" {
		content, err := os.ReadFile(config.ContextFile)
		if err != nil {
			if config.ErrorCodeOnly {
				os.Exit(query.ErrContextNotFoundCode)
			}
			return fmt.Errorf("failed to read context file: %w", err)
		}
		contextContent = string(content)
	}

	// Load configuration for additional settings
	maxToolFollowUp, systemPrompt := s.loadConfigSettings(config)

	// Choose command options based on verbosity
	var commandOptions *host.CommandOptions
	if config.Noisy {
		commandOptions = host.DefaultCommandOptions()
	} else {
		commandOptions = host.QuietCommandOptions()
	}

	// Execute query with server connections
	var result *query.QueryResult
	err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
		// Create query handler with the AI provider
		handler, err := query.NewQueryHandlerWithProvider(conns, provider, systemPrompt)
		if err != nil {
			if config.ErrorCodeOnly {
				os.Exit(query.ErrInitializationCode)
			}
			return fmt.Errorf("failed to initialize query: %w", err)
		}

		handler.SetMaxFollowUpAttempts(maxToolFollowUp)

		if contextContent != "" {
			handler.AddContext(contextContent)
		}

		if config.MaxTokens > 0 {
			handler.SetMaxTokens(config.MaxTokens)
		}

		result, err = handler.Execute(config.Question)
		if err != nil {
			if config.ErrorCodeOnly {
				exitCode := query.GetExitCode(err)
				os.Exit(exitCode)
			}
			return fmt.Errorf("query failed: %w", err)
		}

		return nil
	}, config.ConfigFile, config.ServerNames, config.UserSpecified, commandOptions)

	if err != nil {
		return err
	}

	// Process and output results
	return s.processResults(result, config)
}

// loadConfigSettings loads configuration settings and returns max tool follow-up and system prompt
func (s *Service) loadConfigSettings(cfg *Config) (int, string) {
	maxToolFollowUp := 2 // default
	systemPrompt := cfg.SystemPrompt

	appCfg, err := config.LoadConfig(cfg.ConfigFile)
	if err == nil {
		// Get the maximum tool follow-up attempts from configuration
		var primaryServerName string
		if len(cfg.ServerNames) == 1 {
			primaryServerName = cfg.ServerNames[0]
		}
		
		maxToolFollowUp = appCfg.GetMaxToolFollowUp(primaryServerName)
		logging.Debug("Using max tool follow-up attempts from config: %d", maxToolFollowUp)
		
		// Get system prompt from config if not provided
		if systemPrompt == "" {
			if len(cfg.ServerNames) == 1 {
				configPrompt := appCfg.GetSystemPrompt(cfg.ServerNames[0])
				if configPrompt != "" {
					systemPrompt = configPrompt
					logging.Debug("Using system prompt from config for server: %s", cfg.ServerNames[0])
				}
			}
			
			if systemPrompt == "" && appCfg.AI != nil && appCfg.AI.DefaultSystemPrompt != "" {
				systemPrompt = appCfg.AI.DefaultSystemPrompt
				logging.Debug("Using default system prompt from config")
			}
		}
	} else {
		logging.Debug("Config loading failed, using default max tool follow-up attempts: %d", maxToolFollowUp)
	}

	return maxToolFollowUp, systemPrompt
}

// processResults processes and outputs the query results
func (s *Service) processResults(result *query.QueryResult, config *Config) error {
	if result == nil {
		return nil
	}

	// Process raw data output if enabled
	if len(result.ToolCalls) > 0 && config.RawDataOutput {
		rawData := s.formatter.FormatRawData(result.ToolCalls)
		if rawData != "" {
			result.Response = rawData
		}
	}

	// Format and output response
	if config.JSONOutput {
		return s.outputJSON(result, config)
	} else {
		return s.outputText(result, config)
	}
}

// outputJSON outputs the result in JSON format
func (s *Service) outputJSON(result *query.QueryResult, config *Config) error {
	jsonData, err := s.formatter.FormatAsJSON(result)
	if err != nil {
		if config.ErrorCodeOnly {
			os.Exit(query.ErrOutputFormatCode)
		}
		return fmt.Errorf("failed to format JSON response: %w", err)
	}

	if config.OutputFile != "" {
		err = os.WriteFile(config.OutputFile, jsonData, 0644)
		if err != nil {
			if config.ErrorCodeOnly {
				os.Exit(query.ErrOutputWriteCode)
			}
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// outputText outputs the result in text format
func (s *Service) outputText(result *query.QueryResult, config *Config) error {
	text := s.formatter.FormatAsText(result)
	
	if config.OutputFile != "" {
		err := os.WriteFile(config.OutputFile, []byte(text), 0644)
		if err != nil {
			if config.ErrorCodeOnly {
				os.Exit(query.ErrOutputWriteCode)
			}
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Println(text)
	}

	return nil
}
