package workflow

import (
	"context"
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

	// Resolve temperature
	temperature := e.resolver.ResolveTemperature(step)
	
	// Resolve timeout
	timeout := e.resolver.ResolveTimeout(step)
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build the request
	request := &domain.CompletionRequest{
		Messages: []domain.Message{
			{
				Role:    "user",
				Content: step.Run,
			},
		},
		Temperature: temperature,
	}

	// Add tools if server manager is available
	if e.serverManager != nil {
		tools, err := e.serverManager.GetAvailableTools()
		if err == nil && len(tools) > 0 {
			request.Tools = tools
		}
	}

	e.logger.Debug("Calling %s/%s with temp=%.2f", pc.Provider, pc.Model, temperature)

	// Make the actual LLM call
	response, err := provider.CreateCompletion(execCtx, request)
	if err != nil {
		return nil, &ProviderError{
			Provider: pc.Provider,
			Model:    pc.Model,
			Err:      err,
		}
	}

	// Extract response
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
