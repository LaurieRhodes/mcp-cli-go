package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
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
	Output   string
	Provider string
	Model    string
	Duration time.Duration
}

// ProviderError represents an error from a specific provider
type ProviderError struct {
	Provider string
	Model    string
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s/%s failed: %v", e.Provider, e.Model, e.Err)
}

// ExecuteStep executes a workflow step with provider fallback
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

// executeWithProvider executes a step with a specific provider
func (e *Executor) executeWithProvider(
	ctx context.Context,
	step *config.StepV2,
	pc config.ProviderFallback,
) (*StepResult, error) {
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
	temperature := e.resolver.ResolveTemperature(step)
	maxIterations := e.resolver.ResolveMaxIterations(step)
	timeout := e.resolver.ResolveTimeout(step)
	
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get available tools from servers specified in step configuration
	var tools []domain.Tool
	if e.serverManager != nil {
		serverNames := e.resolver.ResolveServers(step)
		e.logger.Info("Step '%s' using servers: %v", step.Name, serverNames)
		
		for _, serverName := range serverNames {
			server, exists := e.serverManager.GetServer(serverName)
			if !exists {
				e.logger.Warn("Server %s not found, skipping", serverName)
				continue
			}
			
			serverTools, err := server.GetTools()
			if err != nil {
				e.logger.Warn("Failed to get tools from server %s: %v", serverName, err)
				continue
			}
			
			e.logger.Info("Got %d tools from server '%s'", len(serverTools), serverName)
			tools = append(tools, serverTools...)
		}
		
		// Filter tools by skill names if specified
		if len(step.Skills) > 0 {
			e.logger.Info("Filtering to skills: %v", step.Skills)
			
			// Create a map of allowed skill names
			allowedSkills := make(map[string]bool)
			for _, skillName := range step.Skills {
				allowedSkills[skillName] = true
			}
			
			// Always allow execute_skill_code when skills are specified
			allowedSkills["execute_skill_code"] = true
			
			e.logger.Info("Allowed skills map: %v", allowedSkills)
			
			// Filter tools
			filteredTools := make([]domain.Tool, 0)
			for _, tool := range tools {
				toolName := tool.Function.Name
				
				// Strip server prefix (format is "servername_toolname")
				// Use Index (first underscore) not LastIndex to handle tool names with underscores
				unprefixedName := toolName
				if idx := strings.Index(toolName, "_"); idx > 0 {
					unprefixedName = toolName[idx+1:]
				}
				
				// Check if this tool is an allowed skill or execute_skill_code
				if allowedSkills[unprefixedName] {
					filteredTools = append(filteredTools, tool)
					e.logger.Info("  ✓ MATCHED: '%s' (unprefixed: '%s')", toolName, unprefixedName)
				} else {
					e.logger.Debug("  ✗ SKIPPED: '%s' (unprefixed: '%s')", toolName, unprefixedName)
				}
			}
			
			e.logger.Info("Filtered to %d tools (was %d)", len(filteredTools), len(tools))
			tools = filteredTools
		}
	}

	// Determine if we need skills-aware system prompt
	hasSkills := false
	for _, tool := range tools {
		// Check if any skill tools are loaded
		toolName := tool.Function.Name
		if strings.Contains(toolName, "skill") || 
		   toolName == "execute_skill_code" ||
		   toolName == "docx" || toolName == "pdf" || 
		   toolName == "pptx" || toolName == "xlsx" {
			hasSkills = true
			break
		}
	}

	// Build initial message history with system prompt
	messages := []domain.Message{}
	
	// Add system message if we have skills
	if hasSkills {
		systemPrompt := `You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question.

IMPORTANT - Using Skills:
Skills provide specialized capabilities through code execution. There are two ways to use skills:

1. PASSIVE MODE - Load documentation and reference materials:
   Call the skill tool directly (e.g., 'docx', 'pdf', 'pptx', 'xlsx')
   Use this to learn about a skill's capabilities before using it.

2. ACTIVE MODE - Execute code to perform tasks:
   Call 'execute_skill_code' with skill_name parameter
   Use this to CREATE, MODIFY, PROCESS, or GENERATE anything.

CRITICAL - File Paths:
When writing code, ALL output files MUST be saved to /outputs/ directory:
   doc.save('/outputs/result.docx')  ✅ CORRECT - File persists to host
   doc.save('/workspace/result.docx') ❌ WRONG - File deleted when container exits
   doc.save('result.docx') ❌ WRONG - Defaults to /workspace/

The /outputs/ directory is the ONLY location where files persist after execution.`

		messages = append(messages, domain.Message{
			Role:    "system",
			Content: systemPrompt,
		})
		e.logger.Debug("Added skills-aware system prompt")
	}
	
	// Add user message
	messages = append(messages, domain.Message{
		Role:    "user",
		Content: step.Run,
	})

	e.logger.Debug("Calling %s/%s with temp=%.2f, max_iterations=%d", 
		pc.Provider, pc.Model, temperature, maxIterations)

	// Make initial LLM call
	request := &domain.CompletionRequest{
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
	}

	response, err := provider.CreateCompletion(execCtx, request)
	if err != nil {
		return nil, &ProviderError{
			Provider: pc.Provider,
			Model:    pc.Model,
			Err:      err,
		}
	}

	e.logger.Debug("Initial response: %s", response.Response)

	// Agentic loop - handle tool calls
	iteration := 0
	for iteration < maxIterations {
		// Check if we have tool calls
		if len(response.ToolCalls) == 0 {
			// No tool calls - we're done
			e.logger.Debug("No tool calls, execution complete after %d iterations", iteration)
			break
		}

		e.logger.Info("Query resulted in %d tool calls (iteration #%d)", 
			len(response.ToolCalls), iteration+1)

		// Add assistant message with tool calls to history
		assistantMessage := domain.Message{
			Role:      "assistant",
			Content:   response.Response,
			ToolCalls: response.ToolCalls,
		}
		messages = append(messages, assistantMessage)

		// Execute each tool call
		for _, toolCall := range response.ToolCalls {
			e.logger.Debug("Executing tool: %s", toolCall.Function.Name)

			// Parse arguments
			var args map[string]interface{}
			if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
				e.logger.Warn("Failed to parse tool arguments: %v", err)
				// Add error as tool result
				messages = append(messages, domain.Message{
					Role:       "tool",
					Content:    fmt.Sprintf("Error: failed to parse arguments: %v", err),
					ToolCallID: toolCall.ID,
				})
				continue
			}

			// Execute tool via server manager
			result, err := e.serverManager.ExecuteTool(execCtx, toolCall.Function.Name, args)
			if err != nil {
				e.logger.Warn("Tool execution failed: %v", err)
				result = fmt.Sprintf("Error: %v", err)
			}

			// Add tool result to message history
			messages = append(messages, domain.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		// Get follow-up response
		e.logger.Info("Getting follow-up response #%d after tool execution", iteration+1)

		// Check if we have enough time left in the context
		if deadline, ok := execCtx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining < 30*time.Second {
				e.logger.Warn("Insufficient time remaining (%v) for follow-up request, stopping iterations", remaining)
				response.Response += fmt.Sprintf("\n\n[Note: Stopped after %d iterations due to timeout constraints. The result may be incomplete.]", iteration)
				break
			}
			e.logger.Debug("Time remaining for follow-up: %v", remaining)
		}

		followUpReq := &domain.CompletionRequest{
			Messages:    messages,
			Tools:       tools,
			Temperature: temperature,
		}

		response, err = provider.CreateCompletion(execCtx, followUpReq)
		if err != nil {
			// Check if it's a timeout error
			if execCtx.Err() == context.DeadlineExceeded {
				e.logger.Warn("Context deadline exceeded during follow-up, stopping iterations")
				response.Response += fmt.Sprintf("\n\n[Note: Stopped after %d iterations due to timeout. The result may be incomplete.]", iteration)
				break
			}
			return nil, &ProviderError{
				Provider: pc.Provider,
				Model:    pc.Model,
				Err:      fmt.Errorf("follow-up request failed: %w", err),
			}
		}

		e.logger.Debug("Received follow-up response #%d: %s", iteration+1, response.Response)

		iteration++
	}

	// Check if we hit max iterations with tool calls still pending
	if iteration >= maxIterations && len(response.ToolCalls) > 0 {
		e.logger.Warn("Reached maximum iterations (%d) but still have tool calls", maxIterations)
		response.Response += fmt.Sprintf("\n\n[Note: The maximum number of tool call iterations (%d) was reached. The result may be incomplete.]", maxIterations)
	}

	// Extract final response
	output := strings.TrimSpace(response.Response)

	return &StepResult{
		Output:   output,
		Provider: pc.Provider,
		Model:    pc.Model,
	}, nil
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

// SetAppConfig sets the application config for provider creation
func (e *Executor) SetAppConfig(appConfig *config.ApplicationConfig) {
	e.appConfig = appConfig
}

// SetProvider is deprecated - kept for compatibility
func (e *Executor) SetProvider(provider domain.LLMProvider) {
	// No-op - we create providers dynamically now
}

// SetServerManager sets the MCP server manager for the executor
func (e *Executor) SetServerManager(serverManager domain.MCPServerManager) {
	e.serverManager = serverManager
}
