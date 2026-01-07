package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// LoopExecutor handles loop execution
type LoopExecutor struct {
	appConfig     *config.ApplicationConfig
	logger        *Logger
	interpolator  *Interpolator
	executor      *Executor
	serverManager domain.MCPServerManager
}

// NewLoopExecutor creates a new loop executor
func NewLoopExecutor(
	appConfig *config.ApplicationConfig,
	logger *Logger,
	interpolator *Interpolator,
	executor *Executor,
	serverManager domain.MCPServerManager,
) *LoopExecutor {
	return &LoopExecutor{
		appConfig:     appConfig,
		logger:        logger,
		interpolator:  interpolator,
		executor:      executor,
		serverManager: serverManager,
	}
}

// LoopResult stores results from loop execution
type LoopResult struct {
	Iterations  int
	FinalOutput string
	AllOutputs  []string
	ExitReason  string // "condition_met", "max_iterations", "failure"
}

// ExecuteLoop executes a loop until condition is met or max iterations reached
func (le *LoopExecutor) ExecuteLoop(ctx context.Context, loop *config.LoopV2) (*LoopResult, error) {
	le.logger.Info("Starting loop: %s (max %d iterations)", loop.Name, loop.MaxIterations)
	
	if loop.MaxIterations <= 0 {
		return nil, fmt.Errorf("max_iterations must be > 0, got %d", loop.MaxIterations)
	}
	
	// Get the workflow to execute
	workflow, exists := le.appConfig.GetWorkflow(loop.Workflow)
	if !exists {
		return nil, fmt.Errorf("loop workflow '%s' not found", loop.Workflow)
	}
	
	result := &LoopResult{
		AllOutputs: make([]string, 0),
	}
	
	var lastOutput string
	
	for iteration := 1; iteration <= loop.MaxIterations; iteration++ {
		le.logger.Info("Loop iteration %d/%d", iteration, loop.MaxIterations)
		
		// Set loop variables for interpolation
		le.interpolator.SetLoopVars(iteration, lastOutput, result.AllOutputs)
		
		// Prepare input for workflow
		inputData, err := le.prepareLoopInput(loop, iteration, lastOutput)
		if err != nil {
			if loop.OnFailure == "halt" {
				return nil, fmt.Errorf("iteration %d input preparation failed: %w", iteration, err)
			}
			le.logger.Warn("Iteration %d input preparation failed: %v", iteration, err)
			continue
		}
		
		// Execute the workflow
		output, err := le.executeWorkflow(ctx, workflow, inputData)
		if err != nil {
			if loop.OnFailure == "halt" {
				result.ExitReason = "failure"
				return result, fmt.Errorf("iteration %d failed: %w", iteration, err)
			}
			le.logger.Warn("Iteration %d failed: %v", iteration, err)
			if loop.OnFailure == "retry" {
				iteration-- // Retry same iteration
			}
			continue
		}
		
		// Store result
		lastOutput = output
		result.AllOutputs = append(result.AllOutputs, output)
		result.Iterations = iteration
		result.FinalOutput = output
		
		le.logger.Debug("Iteration %d output: %s", iteration, truncate(output, 100))
		
		// Evaluate exit condition
		if loop.Until != "" {
			conditionMet, err := le.evaluateCondition(ctx, loop.Until, output)
			if err != nil {
				le.logger.Warn("Failed to evaluate condition: %v", err)
			} else if conditionMet {
				le.logger.Info("Loop exit condition met after %d iterations", iteration)
				result.ExitReason = "condition_met"
				
				// Store final result
				le.storeLoopResult(loop, result)
				return result, nil
			}
		}
	}
	
	// Max iterations reached
	le.logger.Info("Loop completed: max iterations (%d) reached", loop.MaxIterations)
	result.ExitReason = "max_iterations"
	
	// Store final result
	le.storeLoopResult(loop, result)
	return result, nil
}

// prepareLoopInput prepares input for loop iteration
func (le *LoopExecutor) prepareLoopInput(loop *config.LoopV2, iteration int, lastOutput string) (string, error) {
	// Build input from 'with' map
	if loop.With == nil || len(loop.With) == 0 {
		return lastOutput, nil
	}
	
	// If there's an 'input' key, use that
	if inputValue, ok := loop.With["input"]; ok {
		inputStr, ok := inputValue.(string)
		if !ok {
			return "", fmt.Errorf("loop input must be a string")
		}
		
		// Interpolate
		interpolated, err := le.interpolator.Interpolate(inputStr)
		if err != nil {
			return "", fmt.Errorf("failed to interpolate loop input: %w", err)
		}
		return interpolated, nil
	}
	
	// Otherwise, build from all 'with' parameters
	var parts []string
	for key, value := range loop.With {
		valueStr := fmt.Sprintf("%v", value)
		interpolated, err := le.interpolator.Interpolate(valueStr)
		if err != nil {
			return "", fmt.Errorf("failed to interpolate %s: %w", key, err)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", key, interpolated))
	}
	
	return strings.Join(parts, "\n"), nil
}

// executeWorkflow executes a workflow and returns its final output
func (le *LoopExecutor) executeWorkflow(ctx context.Context, workflow *config.WorkflowV2, inputData string) (string, error) {
	// Create sub-orchestrator
	subLogger := NewLogger(workflow.Execution.Logging)
	subOrchestrator := NewOrchestrator(workflow, subLogger)
	
	// Pass through dependencies
	subOrchestrator.executor.SetAppConfig(le.appConfig)
	if le.serverManager != nil {
		subOrchestrator.executor.SetServerManager(le.serverManager)
	}
	subOrchestrator.SetAppConfigForWorkflows(le.appConfig)
	
	// Copy loop variables to sub-workflow's interpolator
	le.interpolator.CopyLoopVars(subOrchestrator.interpolator)
	
	// Execute
	err := subOrchestrator.Execute(ctx, inputData)
	if err != nil {
		return "", err
	}
	
	// Get final result
	if len(workflow.Steps) > 0 {
		lastStepName := workflow.Steps[len(workflow.Steps)-1].Name
		if output, ok := subOrchestrator.GetStepResult(lastStepName); ok {
			return output, nil
		}
	}
	
	return "", fmt.Errorf("no output from workflow")
}

// evaluateCondition uses LLM to evaluate exit condition
func (le *LoopExecutor) evaluateCondition(ctx context.Context, condition string, output string) (bool, error) {
	// Interpolate condition
	interpolatedCondition, err := le.interpolator.Interpolate(condition)
	if err != nil {
		return false, fmt.Errorf("failed to interpolate condition: %w", err)
	}
	
	// Build evaluation prompt
	prompt := fmt.Sprintf(
		"Evaluate if this condition is satisfied. Answer only YES or NO.\n\n"+
			"Condition: %s\n\n"+
			"Output to evaluate:\n%s\n\n"+
			"Answer (YES or NO):",
		interpolatedCondition,
		truncate(output, 2000),
	)
	
	// Use executor's provider creation for evaluation
	// Get default provider from app config
	providerName := "deepseek" // Fallback
	if le.appConfig.AI != nil && le.appConfig.AI.DefaultProvider != "" {
		providerName = le.appConfig.AI.DefaultProvider
	}
	
	// Create provider
	provider, err := le.executor.createProvider(providerName, "")
	if err != nil {
		return false, fmt.Errorf("failed to create provider for condition evaluation: %w", err)
	}
	
	// Execute
	request := &domain.CompletionRequest{
		Messages: []domain.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	}
	
	response, err := provider.CreateCompletion(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate condition: %w", err)
	}
	
	// Check response
	answer := strings.ToUpper(strings.TrimSpace(response.Response))
	le.logger.Info("Condition evaluation: '%s' -> %s", condition, answer)
	
	return strings.Contains(answer, "YES"), nil
}

// storeLoopResult stores loop result for later access
func (le *LoopExecutor) storeLoopResult(loop *config.LoopV2, result *LoopResult) {
	// Store as loop.output
	le.interpolator.SetStepResult("loop.output", result.FinalOutput)
	le.interpolator.SetStepResult("loop.iteration", fmt.Sprintf("%d", result.Iterations))
	
	// Store with custom name if specified
	if loop.Accumulate != "" {
		history := strings.Join(result.AllOutputs, "\n---\n")
		le.interpolator.SetStepResult(loop.Accumulate, history)
	}
	
	// Store loop name result
	le.interpolator.SetStepResult(loop.Name, result.FinalOutput)
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
