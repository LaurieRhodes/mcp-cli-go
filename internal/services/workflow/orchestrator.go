package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// Orchestrator orchestrates workflow execution with dependency resolution
type Orchestrator struct {
	workflow         *config.WorkflowV2
	workflowKey      string // Full workflow key (e.g., "iterative_dev/dev_cycle" or "simple")
	executor         *Executor
	consensusExec    *ConsensusExecutor
	interpolator     *Interpolator
	logger           *Logger
	stepResults      map[string]string
	consensusResults map[string]*config.ConsensusResult
	appConfig        *config.ApplicationConfig
	loopExecutor     *LoopExecutor
	embeddingService domain.EmbeddingService
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator(workflow *config.WorkflowV2, logger *Logger) *Orchestrator {
	return NewOrchestratorWithKey(workflow, "", logger)
}

// NewOrchestratorWithKey creates a new workflow orchestrator with a workflow key for directory context
func NewOrchestratorWithKey(workflow *config.WorkflowV2, workflowKey string, logger *Logger) *Orchestrator {
	executor := NewExecutor(workflow, logger)
	consensusExec := NewConsensusExecutor(executor)
	interpolator := NewInterpolator()

	// Set environment variables
	interpolator.SetEnv(workflow.Env)

	return &Orchestrator{
		workflow:         workflow,
		workflowKey:      workflowKey,
		executor:         executor,
		consensusExec:    consensusExec,
		interpolator:     interpolator,
		logger:           logger,
		stepResults:      make(map[string]string),
		consensusResults: make(map[string]*config.ConsensusResult),
	}
}

// Execute executes the entire workflow
func (o *Orchestrator) Execute(ctx context.Context, input string) error {
	// Set initial input
	o.interpolator.Set("input", input)

	o.logger.Info("Starting workflow: %s v%s", o.workflow.Name, o.workflow.Version)
	o.logger.Step("\n[WORKFLOW] %s v%s", o.workflow.Name, o.workflow.Version)

	// Initialize loop executor if we have appConfig and loops
	if o.appConfig != nil && len(o.workflow.Loops) > 0 {
		o.loopExecutor = NewLoopExecutor(
			o.appConfig,
			o.logger,
			o.interpolator,
			o.executor,
			o.executor.serverManager,
		)
	}

	// Track completed steps and loops
	completed := make(map[string]bool)

	// Execute steps and loops in dependency order
	stepsRemaining := make(map[string]*config.StepV2)
	for i := range o.workflow.Steps {
		stepsRemaining[o.workflow.Steps[i].Name] = &o.workflow.Steps[i]
	}
	
	loopsRemaining := make(map[string]*config.LoopV2)
	for i := range o.workflow.Loops {
		loopsRemaining[o.workflow.Loops[i].Name] = &o.workflow.Loops[i]
	}

	maxIterations := 100
	iteration := 0
	
	for len(stepsRemaining) > 0 || len(loopsRemaining) > 0 {
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("dependency resolution exceeded max iterations (possible circular dependency)")
		}
		
		progressMade := false
		
		// Try to execute steps whose dependencies are met
		for name, step := range stepsRemaining {
			if o.checkDependencies(step, completed) == nil {
				// Dependencies met, execute
				if err := o.executeStep(ctx, step); err != nil {
					return fmt.Errorf("step %s failed: %w", step.Name, err)
				}
				completed[step.Name] = true
				delete(stepsRemaining, name)
				progressMade = true
			}
		}
		
		// Try to execute loops (loops have no dependencies - they just run)
		// Loops can be referenced by steps via needs: [loop_name]
		for name, loop := range loopsRemaining {
			// Loops execute immediately (no dependency checking for loops themselves)
			if err := o.executeLoop(ctx, loop); err != nil {
				return fmt.Errorf("loop %s failed: %w", loop.Name, err)
			}
			completed[loop.Name] = true
			delete(loopsRemaining, name)
			progressMade = true
			break // Execute one loop at a time, then re-check steps
		}
		
		// If no progress was made, we have a deadlock
		if !progressMade {
			var pending []string
			for name, step := range stepsRemaining {
				pending = append(pending, fmt.Sprintf("step:%s (needs: %v)", name, step.Needs))
			}
			for name := range loopsRemaining {
				pending = append(pending, fmt.Sprintf("loop:%s", name))
			}
			return fmt.Errorf("dependency deadlock: cannot execute remaining elements: %v", pending)
		}
	}

	o.logger.Info("Workflow completed successfully")
	o.logger.Step("\n[SUCCESS] Workflow completed")
	return nil
}

// checkDependencies checks if all dependencies are met
func (o *Orchestrator) checkDependencies(step *config.StepV2, completed map[string]bool) error {
	for _, dep := range step.Needs {
		if !completed[dep] {
			return fmt.Errorf("dependency not met: %s", dep)
		}
	}
	return nil
}

// executeStep executes a single step
func (o *Orchestrator) executeStep(ctx context.Context, step *config.StepV2) error {
	o.logger.Info("Executing step: %s", step.Name)
	
	// Track step timing for steps level logging
	stepStart := time.Now()
	
	// Find step index for steps level logging
	stepIndex := 0
	totalSteps := len(o.workflow.Steps)
	for i, s := range o.workflow.Steps {
		if s.Name == step.Name {
			stepIndex = i + 1
			break
		}
	}
	
	o.logger.Step("\n[STEP %d/%d] %s", stepIndex, totalSteps, step.Name)

	// Check condition
	if step.If != "" {
		if !o.evaluateIfCondition(step.If) {
			o.logger.Info("Step skipped (condition not met)")
			o.logger.Step("  ⊘ Skipped (condition not met)")
			return nil
		}
	}

	// Determine step type and execute
	var err error
	if step.Consensus != nil {
		err = o.executeConsensusStep(ctx, step)
	} else if step.Run != "" {
		err = o.executeRegularStep(ctx, step)
	} else if step.Embeddings != nil {
		err = o.executeEmbeddingsStep(ctx, step)
	} else if step.Rag != nil {
		err = o.executeRagStep(ctx, step)
	} else if step.Template != nil {
		err = o.executeWorkflowStep(ctx, step)
	} else {
		err = fmt.Errorf("no execution mode specified")
	}
	
	// Log step completion with timing
	duration := time.Since(stepStart)
	if err != nil {
		o.logger.Step("  ✗ Failed (%.1fs): %v", duration.Seconds(), err)
		return err
	}
	
	o.logger.Step("  ✓ Completed (%.1fs)", duration.Seconds())
	return nil
}

// executeRegularStep executes a regular (non-consensus) step
func (o *Orchestrator) executeRegularStep(ctx context.Context, step *config.StepV2) error {
	// Interpolate prompt
	prompt, _ := o.interpolator.Interpolate(step.Run)

	// Create temp step with interpolated prompt
	tempStep := *step
	tempStep.Run = prompt

	// Execute
	result, err := o.executor.ExecuteStep(ctx, &tempStep)

	if err != nil {
		return err
	}
	// Store result
	o.stepResults[step.Name] = result.Output
	o.interpolator.SetStepResult(step.Name, result.Output)

	o.logger.Output("Step %s result: %s", step.Name, result.Output)

	return nil
}

// executeConsensusStep executes a consensus step
func (o *Orchestrator) executeConsensusStep(ctx context.Context, step *config.StepV2) error {
	// Interpolate consensus prompt
	prompt, _ := o.interpolator.Interpolate(step.Consensus.Prompt)

	// Create temp step with interpolated prompt
	tempStep := *step
	tempConsensus := *step.Consensus
	tempConsensus.Prompt = prompt
	tempStep.Consensus = &tempConsensus

	// Execute consensus
	result, err := o.consensusExec.ExecuteConsensus(ctx, &tempStep)
	if err != nil {
		return fmt.Errorf("consensus execution failed: %w", err)
	}
	
	// Check if result is nil (shouldn't happen if no error, but defensive)
	if result == nil {
		return fmt.Errorf("consensus returned nil result")
	}

	// Store results
	o.consensusResults[step.Name] = result
	o.stepResults[step.Name] = result.Result
	o.interpolator.SetStepResult(step.Name, result.Result)

	// Output consensus details
	o.logger.Output("Step %s consensus result: %s", step.Name, result.Result)
	o.logger.Output("  Agreement: %.0f%%, Confidence: %s", result.Agreement*100, result.Confidence)

	if !result.Success {
		return fmt.Errorf("consensus failed to reach agreement")
	}

	return nil
}

// executeEmbeddingsStep executes an embeddings generation step
func (o *Orchestrator) executeEmbeddingsStep(ctx context.Context, step *config.StepV2) error {
	emb := step.Embeddings
	if emb == nil {
		return fmt.Errorf("embeddings configuration is nil")
	}

	// Check if embeddings service is available
	if o.embeddingService == nil {
		return fmt.Errorf("embeddings service not initialized")
	}

	// Determine input
	var inputText string
	if emb.InputFile != "" {
		// Read from file
		interpolatedPath, _ := o.interpolator.Interpolate(emb.InputFile)
		data, err := os.ReadFile(interpolatedPath)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		inputText = string(data)
		o.logger.Info("Read %d characters from file: %s", len(inputText), interpolatedPath)
	} else if emb.Input != nil {
		// Interpolate and join inputs
		var inputs []string
		switch v := emb.Input.(type) {
		case string:
			interpolated, _ := o.interpolator.Interpolate(v)
			inputs = []string{interpolated}
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					interpolated, _ := o.interpolator.Interpolate(str)
					inputs = append(inputs, interpolated)
				}
			}
		case []string:
			for _, str := range v {
				interpolated, _ := o.interpolator.Interpolate(str)
				inputs = append(inputs, interpolated)
			}
		default:
			return fmt.Errorf("invalid input type for embeddings: %T", v)
		}
		// Join multiple inputs with newlines
		inputText = strings.Join(inputs, "\n\n")
	} else {
		return fmt.Errorf("either input or input_file required for embeddings")
	}

	if strings.TrimSpace(inputText) == "" {
		return fmt.Errorf("input text is empty")
	}

	// Get provider (inherit from step/execution or use override)
	provider := emb.Provider
	if provider == "" {
		provider = step.Provider
	}
	if provider == "" {
		provider = o.workflow.Execution.Provider
	}

	// Get model (inherit from step/execution or use override)
	model := emb.Model
	if model == "" {
		model = step.Model
	}
	if model == "" {
		model = o.workflow.Execution.Model
	}

	if provider == "" || model == "" {
		return fmt.Errorf("provider and model required for embeddings")
	}

	// Set defaults for optional parameters
	chunkStrategy := emb.ChunkStrategy
	if chunkStrategy == "" {
		chunkStrategy = "sentence"
	}

	maxChunkSize := emb.MaxChunkSize
	if maxChunkSize == 0 {
		maxChunkSize = 512
	}

	encodingFormat := emb.EncodingFormat
	if encodingFormat == "" {
		encodingFormat = "float"
	}

	outputFormat := emb.OutputFormat
	if outputFormat == "" {
		outputFormat = "json"
	}

	includeMetadata := true
	if emb.IncludeMetadata != nil {
		includeMetadata = *emb.IncludeMetadata
	}

	o.logger.Info("Generating embeddings with %s/%s (strategy: %s, chunk size: %d)", 
		provider, model, chunkStrategy, maxChunkSize)

	// Create embedding request
	req := &domain.EmbeddingJobRequest{
		Input:          inputText,
		Provider:       provider,
		Model:          model,
		ChunkStrategy:  domain.ChunkingType(chunkStrategy),
		MaxChunkSize:   maxChunkSize,
		ChunkOverlap:   emb.Overlap,
		EncodingFormat: encodingFormat,
		Dimensions:     emb.Dimensions,
		Metadata: map[string]interface{}{
			"workflow": o.workflow.Name,
			"step":     step.Name,
		},
	}

	// Generate embeddings
	job, err := o.embeddingService.GenerateEmbeddings(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	o.logger.Info("Generated embeddings: %d chunks, %d vectors", 
		len(job.Chunks), len(job.Embeddings))

	// Format output
	var outputData []byte
	var result string

	if includeMetadata {
		// Full job with metadata
		outputData, err = json.MarshalIndent(job, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal embeddings: %w", err)
		}
		result = string(outputData)
	} else {
		// Minimal format - just vectors
		vectors := make([][]float32, len(job.Embeddings))
		for i, embedding := range job.Embeddings {
			vectors[i] = embedding.Vector
		}
		
		minimal := map[string]interface{}{
			"model":   job.Model,
			"vectors": vectors,
			"count":   len(vectors),
		}
		
		outputData, err = json.MarshalIndent(minimal, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal embeddings: %w", err)
		}
		result = string(outputData)
	}

	// Write to output file if specified
	if emb.OutputFile != "" {
		interpolatedPath, _ := o.interpolator.Interpolate(emb.OutputFile)
		err = os.WriteFile(interpolatedPath, outputData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		o.logger.Info("Embeddings written to: %s", interpolatedPath)
		
		// Store file path in results
		result = fmt.Sprintf("Embeddings saved to: %s (%d vectors)", interpolatedPath, len(job.Embeddings))
	}
	
	// Store result for interpolation
	o.stepResults[step.Name] = result
	o.interpolator.SetStepResult(step.Name, result)

	o.logger.Output("Step %s result: Generated %d embeddings", step.Name, len(job.Embeddings))

	return nil
}

// evaluateCondition evaluates a conditional expression
func (o *Orchestrator) evaluateCondition(condition string) bool {
	// Simple condition evaluation
	// For now, support: ${{ stepName == "value" }} or ${{ stepName.result == "value" }}

	// Extract condition components
	// This is a simplified implementation
	// TODO: Implement full expression evaluator

	// For MVP, check if step result equals a value
	// Format: ${{ stepName == "value" }}
	
	// Remove ${{ and }}
	condition = strings.TrimSpace(condition)
	condition = strings.TrimPrefix(condition, "${{")
	condition = strings.TrimSuffix(condition, "}}")
	condition = strings.TrimSpace(condition)

	// Split by ==
	parts := strings.Split(condition, "==")
	if len(parts) != 2 {
		o.logger.Warn("Invalid condition format: %s", condition)
		return false
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	// Remove quotes from right side
	right = strings.Trim(right, "\"'")

	// Handle step.result format
	if strings.Contains(left, ".result") {
		left = strings.TrimSuffix(left, ".result")
		left = strings.TrimSpace(left)
	}

	// Get step result
	value, ok := o.stepResults[left]
	if !ok {
		o.logger.Warn("Condition references unknown step: %s", left)
		return false
	}

	// Compare (case-insensitive, trimmed)
	leftVal := strings.TrimSpace(strings.ToUpper(value))
	rightVal := strings.TrimSpace(strings.ToUpper(right))

	return leftVal == rightVal
}

// GetStepResult gets a step's result
func (o *Orchestrator) GetStepResult(stepName string) (string, bool) {
	result, ok := o.stepResults[stepName]
	return result, ok
}

// GetConsensusResult gets a step's consensus result
func (o *Orchestrator) GetConsensusResult(stepName string) (*config.ConsensusResult, bool) {
	result, ok := o.consensusResults[stepName]
	return result, ok
}

// SetAppConfig sets the application config for provider creation
func (o *Orchestrator) SetAppConfig(appConfig *config.ApplicationConfig) {
	o.executor.SetAppConfig(appConfig)
}

// SetEmbeddingService sets the embedding service for embeddings steps
func (o *Orchestrator) SetEmbeddingService(service domain.EmbeddingService) {
	o.embeddingService = service
}

// SetProvider is deprecated - kept for compatibility
func (o *Orchestrator) SetProvider(provider domain.LLMProvider) {
	// No-op - we create providers dynamically now
}

// SetServerManager sets the MCP server manager for the orchestrator
func (o *Orchestrator) SetServerManager(serverManager domain.MCPServerManager) {
	o.executor.SetServerManager(serverManager)
}

// SetAppConfigForWorkflows sets the app config for workflow-to-workflow calls
func (o *Orchestrator) SetAppConfigForWorkflows(appConfig *config.ApplicationConfig) {
	o.appConfig = appConfig
}

// executeWorkflowStep executes a step that calls another workflow
func (o *Orchestrator) executeWorkflowStep(ctx context.Context, step *config.StepV2) error {
	workflowName := step.Template.Name
	
	o.logger.Info("Calling workflow: %s", workflowName)
	
	// Extract directory from current workflow key for directory-aware resolution
	var contextDir string
	if o.workflowKey != "" {
		if idx := strings.LastIndex(o.workflowKey, "/"); idx != -1 {
			contextDir = o.workflowKey[:idx]
		}
	}
	
	// Get the sub-workflow with directory-aware resolution
	subWorkflow, exists := o.appConfig.GetWorkflowWithContext(workflowName, contextDir)
	if !exists {
		// Try to provide helpful error message
		if contextDir != "" {
			return fmt.Errorf("workflow '%s' not found (searched in '%s' directory and root)", workflowName, contextDir)
		}
		return fmt.Errorf("workflow '%s' not found", workflowName)
	}
	
	// Determine the full key of the resolved workflow for sub-workflow context
	var subWorkflowKey string
	for key, wf := range o.appConfig.Workflows {
		if wf == subWorkflow {
			subWorkflowKey = key
			break
		}
	}
	
	// Prepare input for sub-workflow
	var inputData string
	if inputValue, ok := step.Template.With["input"]; ok {
		// Interpolate the input value
		inputStr, ok := inputValue.(string)
		if !ok {
			return fmt.Errorf("workflow input must be a string")
		}
		interpolated, _ := o.interpolator.Interpolate(inputStr)
		inputData = interpolated
	}
	
	// Create a new orchestrator for the sub-workflow with its key for directory context
	subLogger := NewLogger(subWorkflow.Execution.Logging, false)
	subOrchestrator := NewOrchestratorWithKey(subWorkflow, subWorkflowKey, subLogger)
	
	// Pass through app config and server manager
	subOrchestrator.executor.SetAppConfig(o.executor.appConfig)
	if o.executor.serverManager != nil {
		subOrchestrator.executor.SetServerManager(o.executor.serverManager)
	}
	
	// Pass app config to sub-orchestrator for nested workflow calls
	subOrchestrator.SetAppConfigForWorkflows(o.appConfig)
	
	// Execute the sub-workflow
	err := subOrchestrator.Execute(ctx, inputData)
 if err != nil {
 	return fmt.Errorf("execution failed: %w", err)
 }
	
	// Get the final result from the sub-workflow
	var result string
	if len(subWorkflow.Steps) > 0 {
		lastStepName := subWorkflow.Steps[len(subWorkflow.Steps)-1].Name
		finalResult, ok := subOrchestrator.GetStepResult(lastStepName)
		if ok {
			result = finalResult
		}
	}
	
	// Store result (same as executeRegularStep)
	o.stepResults[step.Name] = result
	o.interpolator.SetStepResult(step.Name, result)
	
	o.logger.Info("Workflow '%s' completed, result available as {{%s}}", workflowName, step.Name)
	
	return nil
}
// parseExecutionOrder determines execution order from YAML structure


// executeLoopElement executes a loop element


func (o *Orchestrator) executeStepElement(ctx context.Context, step *config.StepV2) error {
	// Check dependencies
	if !o.dependenciesMet(step) {
		return fmt.Errorf("dependencies not met for step: %s", step.Name)
	}

	// Check condition
	if step.If != "" {
		shouldRun := o.evaluateIfCondition(step.If)
		if !shouldRun {
			o.logger.Info("Skipping step %s (condition not met)", step.Name)
			return nil
		}
	}

	o.logger.Info("Executing step: %s", step.Name)

	// Route to appropriate executor
	if step.Consensus != nil {
		return o.executeConsensusStep(ctx, step)
	} else if step.Embeddings != nil {
		return o.executeEmbeddingsStep(ctx, step)
	} else if step.Rag != nil {
		return o.executeRagStep(ctx, step)
	} else if step.Template != nil {
		return o.executeWorkflowStep(ctx, step)
	} else if step.Loop != nil {
		return o.executeLoopStep(ctx, step)
	}

	return fmt.Errorf("no execution mode specified")
}

func (o *Orchestrator) executeLoopStep(ctx context.Context, step *config.StepV2) error {
	if o.appConfig == nil {
		return fmt.Errorf("loop executor not initialized (appConfig missing)")
	}
	
	o.logger.Info("Starting loop: %s", step.Name)
	
	// Execute the loop
	result, _ := o.executeLoopInternal(
		ctx,
		step.Name,
		step.Loop.Workflow,
		step.Loop.With,
		step.Loop.MaxIterations,
		step.Loop.Until,
		step.Loop.OnFailure,
		step.Loop.Accumulate,
	)
	
	o.logger.Info("Loop %s completed: %d iterations, exit: %s", 
		step.Name, result.Iterations, result.ExitReason)
	
	return nil
}

// loopResult stores results from loop execution
type loopResult struct {
	Iterations  int
	FinalOutput string
	AllOutputs  []string
	ExitReason  string
}



// executeLoopInternal contains the actual loop execution logic
func (o *Orchestrator) executeLoopInternal(ctx context.Context, name string, workflow string, with map[string]interface{}, maxIterations int, until string, onFailure string, accumulate string) (*loopResult, error) {
	if maxIterations <= 0 {
		return nil, fmt.Errorf("max_iterations must be > 0, got %d", maxIterations)
	}
	
	wf, exists := o.appConfig.GetWorkflow(workflow)
	if !exists {
		return nil, fmt.Errorf("loop workflow '%s' not found", workflow)
	}
	
	result := &loopResult{
		AllOutputs: make([]string, 0),
	}
	
	var lastOutput string
	
	for iteration := 1; iteration <= maxIterations; iteration++ {
		o.logger.Info("Loop iteration %d/%d", iteration, maxIterations)
		
		o.interpolator.SetLoopVars(iteration, lastOutput, result.AllOutputs)
		
		inputData, err := o.prepareLoopInput(with, lastOutput)
		if err != nil {
			if onFailure == "halt" {
				return nil, fmt.Errorf("iteration %d input prep failed: %w", iteration, err)
			}
			o.logger.Warn("Iteration %d input prep failed: %v", iteration, err)
			continue
		}
		
		output, err := o.executeLoopWorkflow(ctx, wf, inputData)
		if err != nil {
			if onFailure == "halt" {
				result.ExitReason = "failure"
				return result, fmt.Errorf("iteration %d failed: %w", iteration, err)
			}
			o.logger.Warn("Iteration %d failed: %v", iteration, err)
			if onFailure == "retry" {
				iteration--
			}
			continue
		}
		
		lastOutput = output
		result.AllOutputs = append(result.AllOutputs, output)
		result.Iterations = iteration
		result.FinalOutput = output
		
		if until != "" {
			conditionMet, err := o.evaluateLoopCondition(ctx, until, output)
			if err != nil {
				o.logger.Warn("Failed to evaluate condition: %v", err)
			} else if conditionMet {
				o.logger.Info("Loop exit condition met after %d iterations", iteration)
				result.ExitReason = "condition_met"
				o.storeLoopResult(name, accumulate, result)
				return result, nil
			}
		}
	}
	
	o.logger.Info("Loop completed: max iterations (%d) reached", maxIterations)
	result.ExitReason = "max_iterations"
	o.storeLoopResult(name, accumulate, result)
	return result, nil
}


func (o *Orchestrator) prepareLoopInput(with map[string]interface{}, lastOutput string) (string, error) {
	if with == nil || len(with) == 0 {
		return lastOutput, nil
	}
	
	if inputValue, ok := with["input"]; ok {
		inputStr, ok := inputValue.(string)
		if !ok {
			return "", fmt.Errorf("loop input must be a string")
		}
		interpolated, _ := o.interpolator.Interpolate(inputStr)
		return interpolated, nil
	}
	
	var parts []string
	for key, value := range with {
		valueStr := fmt.Sprintf("%v", value)
		interpolated, _ := o.interpolator.Interpolate(valueStr)
		parts = append(parts, fmt.Sprintf("%s: %s", key, interpolated))
	}
	
	return strings.Join(parts, "\n"), nil
}

func (o *Orchestrator) executeLoopWorkflow(ctx context.Context, workflow *config.WorkflowV2, inputData string) (string, error) {
	subLogger := NewLogger(workflow.Execution.Logging, false)
	subOrchestrator := NewOrchestrator(workflow, subLogger)
	
	subOrchestrator.executor.SetAppConfig(o.executor.appConfig)
	if o.executor.serverManager != nil {
		subOrchestrator.executor.SetServerManager(o.executor.serverManager)
	}
	subOrchestrator.SetAppConfigForWorkflows(o.appConfig)
	
	err := subOrchestrator.Execute(ctx, inputData)
 if err != nil {
 	return "", fmt.Errorf("execution failed: %w", err)
 }
	
	if len(workflow.Steps) > 0 {
		lastStepName := workflow.Steps[len(workflow.Steps)-1].Name
		if output, ok := subOrchestrator.GetStepResult(lastStepName); ok {
			return output, nil
		}
	}
	
	return "", fmt.Errorf("no output from workflow")
}

func (o *Orchestrator) evaluateLoopCondition(ctx context.Context, condition string, output string) (bool, error) {
	interpolatedCondition, _ := o.interpolator.Interpolate(condition)
	
	prompt := fmt.Sprintf(
		"Evaluate if this condition is satisfied. Answer only YES or NO.\n\n"+
			"Condition: %s\n\n"+
			"Output to evaluate:\n%s\n\n"+
			"Answer (YES or NO):",
		interpolatedCondition,
		truncateString(output, 2000),
	)
	
	providerName := "deepseek"
	if o.appConfig.AI != nil && o.appConfig.AI.DefaultProvider != "" {
		providerName = o.appConfig.AI.DefaultProvider
	}
	
	provider, _ := o.executor.createProvider(providerName, "")
	
	request := &domain.CompletionRequest{
		Messages: []domain.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	}
	
	response, _ := provider.CreateCompletion(ctx, request)
	
	answer := strings.ToUpper(strings.TrimSpace(response.Response))
	o.logger.Debug("Condition evaluation: '%s' -> %s", condition, answer)
	
	return strings.Contains(answer, "YES"), nil
}

func (o *Orchestrator) storeLoopResult(name string, accumulate string, result *loopResult) {
	o.interpolator.SetStepResult("loop.output", result.FinalOutput)
	o.interpolator.SetStepResult("loop.iteration", fmt.Sprintf("%d", result.Iterations))
	
	if accumulate != "" {
		history := strings.Join(result.AllOutputs, "\n---\n")
		o.interpolator.SetStepResult(accumulate, history)
	}
	
	o.interpolator.SetStepResult(name, result.FinalOutput)
}


// dependenciesMet checks if all dependencies for a step are satisfied
func (o *Orchestrator) dependenciesMet(step *config.StepV2) bool {
	if len(step.Needs) == 0 {
		return true
	}
	
	for _, depName := range step.Needs {
		if _, exists := o.stepResults[depName]; !exists {
			return false
		}
	}
	return true
}

// evaluateIfCondition evaluates a conditional expression
func (o *Orchestrator) evaluateIfCondition(condition string) bool {
	// Simple evaluation for now: check if variables are set and non-empty
	interpolated, err := o.interpolator.Interpolate(condition)
	if err != nil {
		return false
	}
	
	// Basic truthy check
	return interpolated != "" && interpolated != "false" && interpolated != "0"
}

// executeLoop executes a loop element
func (o *Orchestrator) executeLoop(ctx context.Context, loop *config.LoopV2) error {
	o.logger.Info("Executing loop: %s", loop.Name)
	
	if o.loopExecutor == nil {
		return fmt.Errorf("loop executor not initialized (appConfig missing)")
	}
	
	result, err := o.loopExecutor.ExecuteLoop(ctx, loop)
	if err != nil {
		return fmt.Errorf("loop %s failed: %w", loop.Name, err)
	}
	
	o.logger.Info("Loop %s completed: %d iterations, exit: %s", 
		loop.Name, result.Iterations, result.ExitReason)
	
	return nil
}
