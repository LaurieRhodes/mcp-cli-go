package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/query"
)

// Executor executes workflow steps with provider fallback
type Executor struct {
	workflow        *config.WorkflowV2
	resolver        *PropertyResolver
	logger          *Logger
	appConfig       *config.ApplicationConfig
	configService   interface{} // infraConfig.Service
	serverManager   domain.MCPServerManager
}

// NewExecutor creates a new workflow executor
func NewExecutor(workflow *config.WorkflowV2, logger *Logger) *Executor {
	return &Executor{
		workflow: workflow,
		resolver: NewPropertyResolver(&workflow.Execution),
		logger:   logger,
	}
}

// StepResult represents the result of a step execution
type StepResult struct {
	Output    string
	Messages  []domain.Message
	ToolsUsed bool
	Success   bool
	Duration  time.Duration
}

// ProviderError represents a provider-specific error
type ProviderError struct {
	Provider string
	Model    string
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s/%s: %v", e.Provider, e.Model, e.Err)
}

// ExecuteStep executes a single workflow step with provider fallback
func (e *Executor) ExecuteStep(ctx context.Context, step *config.StepV2) (*StepResult, error) {
	// Resolve provider chain
	providers := e.resolver.ResolveProviders(step)
	
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured for step %s", step.Name)
	}

	e.logger.Debug("Step: %s", step.Name)
	e.logger.Debug("Provider chain: %d providers", len(providers))

	// Try each provider in order
	var lastErr error
	for i, pc := range providers {
		e.logger.Debug("Attempting provider %d/%d: %s/%s", i+1, len(providers), pc.Provider, pc.Model)

		startTime := time.Now()
		result, err := e.executeWithProvider(ctx, step, pc)
		duration := time.Since(startTime)

		if err == nil {
			e.logger.Info("Success: %s/%s (%.2fs)", pc.Provider, pc.Model, duration.Seconds())
			result.Duration = duration
			return result, nil
		}

		// Log failure
		e.logger.Warn("Failed: %s/%s - %v", pc.Provider, pc.Model, err)
		lastErr = err

		// Continue to next provider in chain
	}

	// All providers failed
	return nil, fmt.Errorf("all %d providers failed, last error: %v", len(providers), lastErr)
}

// executeWithProvider executes a step with a specific provider using the query service
func (e *Executor) executeWithProvider(
	ctx context.Context,
	step *config.StepV2,
	pc config.ProviderFallback,
) (*StepResult, error) {
	// ARCHITECTURAL FIX: Delegate to query service instead of reimplementing
	// This ensures workflows behave identically to `mcp-cli query` calls
	
	// Create provider for this specific execution
	provider, err := e.createProvider(pc.Provider, pc.Model)
	if err != nil {
		return nil, &ProviderError{
			Provider: pc.Provider,
			Model:    pc.Model,
			Err:      fmt.Errorf("failed to create provider: %w", err),
		}
	}

	// Resolve configuration
	maxIterations := e.resolver.ResolveMaxIterations(step)
	
	// Build AI options (minimal - provider already configured)
	aiOptions := &host.AIOptions{
		Provider: pc.Provider,
		Model:    pc.Model,
	}

	// Determine system prompt based on whether skills are requested
	systemPrompt := ""
	if len(step.Skills) > 0 {
		// Skills-aware system prompt
		systemPrompt = `You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question.

IMPORTANT - Using Skills:
Skills provide specialized capabilities through code execution. There are three ways to use skills:

1. PASSIVE MODE - Load documentation and reference materials:
   Call the skill tool directly (e.g., 'docx', 'pdf', 'pptx', 'xlsx')
   Use this to learn about a skill's capabilities before using it.

2. RUN HELPER SCRIPT - Execute pre-written scripts (RECOMMENDED):
   Call 'run_helper_script' with skill_name, script_name, and args parameters
   Use this for direct execution of existing scripts in the skill's scripts/ directory
   This is the most efficient method - no code generation needed

3. EXECUTE CUSTOM CODE - Write and execute custom code:
   Call 'execute_skill_code' with skill_name parameter
   Use this to CREATE, MODIFY, PROCESS, or GENERATE anything with custom logic
   Use when you need flexibility beyond what helper scripts provide

CRITICAL - File Paths:
When working with files, ALL output files MUST be saved to /outputs/ directory:
   doc.save('/outputs/result.docx')  ✅ CORRECT - File persists to host
   doc.save('/workspace/result.docx') ❌ WRONG - File deleted when container exits
   doc.save('result.docx') ❌ WRONG - Defaults to /workspace/

The /outputs/ directory is the ONLY location where files persist after execution.`
	}

	// Create query handler with server manager (includes skills)
	handler := query.NewQueryHandlerWithServerManager(
		e.serverManager,
		provider,
		aiOptions,
		systemPrompt,
	)

	
	// Set max iterations
	handler.SetMaxFollowUpAttempts(maxIterations)

	// Execute query
	e.logger.Debug("Executing step via query service: %s/%s with max_iterations=%d", 
		pc.Provider, pc.Model, maxIterations)
	
	queryResult, err := handler.Execute(step.Run)
	if err != nil {
		return nil, &ProviderError{
			Provider: pc.Provider,
			Model:    pc.Model,
			Err:      err,
		}
	}

	// Check for failure indicators in the response
	failed := e.detectStepFailure(queryResult.Response, nil)
	
	// Convert query result to step result
	result := &StepResult{
		Output:       queryResult.Response,
		Messages:     nil, // Query service doesn't expose message history
		ToolsUsed:    len(queryResult.ToolCalls) > 0,
		Success:      !failed,
	}

	e.logger.Debug("Step result: %s", result.Output)
	return result, nil
}

// createProvider creates a provider instance
func (e *Executor) createProvider(providerName, modelName string) (domain.LLMProvider, error) {
	if e.appConfig == nil {
		return nil, fmt.Errorf("no app config available")
	}

	// Get provider config from app config
	var providerConfig *config.ProviderConfig
	var interfaceType config.InterfaceType

	// Search through AI interfaces for this provider
	if e.appConfig.AI != nil {
		for iType, iface := range e.appConfig.AI.Interfaces {
			if pConfig, exists := iface.Providers[providerName]; exists {
				providerConfig = &pConfig
				interfaceType = iType
				break
			}
		}
	}

	if providerConfig == nil {
		return nil, fmt.Errorf("provider '%s' not found in configuration", providerName)
	}

	// Clone the config and override settings
	configCopy := *providerConfig
	if modelName != "" {
		configCopy.DefaultModel = modelName
	}
	
	// For failover chains: disable retries at provider level
	// The executor handles retries by trying the next provider
	configCopy.MaxRetries = 0  // No retries - fail fast for failover
	
	e.logger.Debug("Creating provider %s with model=%s, max_retries=0 (failover mode)", 
		providerName, configCopy.DefaultModel)

	// Create provider
	providerType := domain.ProviderType(providerName)
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, &configCopy, interfaceType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// SetAppConfig sets the application configuration
func (e *Executor) SetAppConfig(appConfig *config.ApplicationConfig) {
	e.appConfig = appConfig
}

// SetProvider is deprecated - kept for compatibility
func (e *Executor) SetProvider(provider domain.LLMProvider) {
	// No-op - we create providers dynamically now
}

// SetServerManager sets the server manager for tool execution
func (e *Executor) SetServerManager(serverManager domain.MCPServerManager) {
	e.serverManager = serverManager
}

// detectStepFailure analyzes LLM output and tool results for failure indicators
func (e *Executor) detectStepFailure(output string, messages []domain.Message) bool {
	outputLower := strings.ToLower(output)
	
	// Check for explicit failure phrases in output
	failureIndicators := []string{
		"failed to",
		"could not",
		"unable to",
		"error:",
		"exception:",
		"traceback",
		"syntax error",
		"name error",
		"type error",
		"value error",
		"file not found",
		"permission denied",
	}
	
	for _, indicator := range failureIndicators {
		if strings.Contains(outputLower, indicator) {
			e.logger.Debug("Detected failure indicator: '%s'", indicator)
			return true
		}
	}
	
	// Check tool results for errors
	for _, msg := range messages {
		if msg.Role == "tool" {
			if e.isToolErrorResponse(msg.Content) {
				e.logger.Debug("Detected tool error in messages")
				return true
			}
		}
	}
	
	return false
}

// isToolErrorResponse checks if a tool response indicates an error
func (e *Executor) isToolErrorResponse(toolOutput string) bool {
	lowerOutput := strings.ToLower(toolOutput)
	
	errorIndicators := []string{
		"error:",
		"exception:",
		"traceback",
		"failed:",
		"could not",
		"unable to",
		"syntax error",
		"name error",
		"type error",
		"import error",
		"file not found",
		"no such file",
		"permission denied",
	}
	
	for _, indicator := range errorIndicators {
		if strings.Contains(lowerOutput, indicator) {
			return true
		}
	}
	
	return false
}

// extractFailureReason extracts a concise failure reason from output
func (e *Executor) extractFailureReason(output string) string {
	lines := strings.Split(output, "\n")
	
	// Look for lines containing error indicators
	errorIndicators := []string{"error:", "exception:", "traceback:", "failed:"}
	
	for idx, line := range lines {
		lineLower := strings.ToLower(line)
		for _, indicator := range errorIndicators {
			if strings.Contains(lineLower, indicator) {
				// Found an error line - extract context
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				
				// Get the error line plus a few lines of context
				reason := []string{trimmed}
				for i := idx + 1; i < len(lines) && i < idx+3; i++ {
					if contextLine := strings.TrimSpace(lines[i]); contextLine != "" {
						reason = append(reason, contextLine)
					}
				}
				return strings.Join(reason, " ")
			}
		}
	}
	
	// Fallback: return first 200 chars
	if len(output) > 200 {
		return output[:200] + "..."
	}
	return output
}
