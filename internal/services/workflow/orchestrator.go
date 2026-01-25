package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
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
	stepResultsMu    sync.RWMutex // Protects stepResults for parallel execution
	consensusResults map[string]*config.ConsensusResult
	appConfig        *config.ApplicationConfig
	loopExecutor     *LoopExecutor
	embeddingService domain.EmbeddingService
	ragServerManager *host.ServerManager // Dedicated manager for RAG servers (internal, not exposed to LLM)
	startFrom        string // Step name to start workflow from (skips previous steps)
	endAt            string // Step name to end workflow at (skips steps after)
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
	// Validate workflow before execution
	if err := ValidateWorkflow(o.workflow); err != nil {
		return fmt.Errorf("workflow validation failed:\n%w", err)
	}
	
	// Set initial input
	o.interpolator.Set("input", input)

	// Log start-from if specified
	if o.startFrom != "" {
		o.logger.Info("Resuming from step: %s", o.startFrom)
	}
	
	// Log end-at if specified
	if o.endAt != "" {
		o.logger.Info("Ending at step: %s", o.endAt)
	}

	o.logger.Info("Starting workflow: %s v%s", o.workflow.Name, o.workflow.Version)
	o.logger.Step("\n[WORKFLOW] %s v%s", o.workflow.Name, o.workflow.Version)

	// Check if parallel execution is enabled
	if o.workflow.Execution.Parallel {
		maxWorkers := o.workflow.Execution.MaxWorkers
		if maxWorkers <= 0 {
			maxWorkers = 3 // Default
		}
		o.logger.Info("Parallel execution enabled (max_workers: %d, policy: %s)", 
			maxWorkers, o.getErrorPolicy())
	}

	// Connect to RAG servers if workflow uses RAG (separate from LLM-exposed servers)
	if err := o.connectRAGServersIfNeeded(ctx); err != nil {
		return fmt.Errorf("failed to connect RAG servers: %w", err)
	}
	
	// Ensure RAG server connections are cleaned up
	if o.ragServerManager != nil {
		defer func() {
			o.logger.Debug("Closing RAG server connections")
			o.ragServerManager.CloseConnections()
		}()
	}

	// Initialize loop executor if we have appConfig and loops
	if o.appConfig != nil && len(o.workflow.Loops) > 0 {
		o.loopExecutor = NewLoopExecutor(
			o.appConfig,
			o.logger,
			o.interpolator,
			o.executor,
			o.executor.serverManager,
			o.embeddingService,
		)
	}

	// Choose execution mode
	if o.workflow.Execution.Parallel {
		return o.executeParallel(ctx)
	}
	
	return o.executeSequential(ctx)
}

// getErrorPolicy returns the error policy with fallback to default
func (o *Orchestrator) getErrorPolicy() string {
	if o.workflow.Execution.OnError == "" {
		return "cancel_all"
	}
	return o.workflow.Execution.OnError
}

// executeSequential executes workflow steps sequentially (original behavior)
func (o *Orchestrator) executeSequential(ctx context.Context) error {
	// Track completed steps and loops
	completed := make(map[string]bool)

	// Pre-mark steps as completed if using start-from or end-at
	if o.startFrom != "" || o.endAt != "" {
		if err := o.markStepsAsCompleted(completed); err != nil {
			return err
		}
	}

	// Execute steps and loops in dependency order
	stepsRemaining := make(map[string]*config.StepV2)
	for i := range o.workflow.Steps {
		step := &o.workflow.Steps[i]
		
		// Skip steps that were marked completed
		if completed[step.Name] {
			o.logger.Debug("Skipping step: %s", step.Name)
			continue
		}
		
		stepsRemaining[step.Name] = step
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

// executeParallel executes workflow steps in parallel with dependency awareness
func (o *Orchestrator) executeParallel(ctx context.Context) error {
	// Create cancellable context for error handling
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Validate variable references
	varValidator := NewVariableValidator(o.workflow)
	if errs := varValidator.ValidateAll(); len(errs) > 0 {
		for _, err := range errs {
			o.logger.Error("Variable validation error: %v", err)
		}
		return fmt.Errorf("variable validation failed with %d error(s)", len(errs))
	}

	// Convert steps to pointers for dependency resolver
	stepPtrs := make([]*config.StepV2, len(o.workflow.Steps))
	for i := range o.workflow.Steps {
		stepPtrs[i] = &o.workflow.Steps[i]
	}

	// Create dependency resolver
	resolver := NewDependencyResolver(stepPtrs)
	
	// Validate dependencies
	if err := resolver.ValidateDependenciesExist(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}
	if err := resolver.ValidateNoCycles(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	// Create worker pool
	pool := NewWorkerPool(
		o.workflow.Execution.MaxWorkers,
		o.getErrorPolicy(),
		o,
	)
	pool.SetCancelFunc(cancel)
	
	// Start timeline tracking
	pool.timeline.Start()

	// Track completion
	completed := make(map[string]bool)
	
	// Pre-mark steps as completed if using start-from or end-at
	if o.startFrom != "" || o.endAt != "" {
		if err := o.markStepsAsCompleted(completed); err != nil {
			return err
		}
	}

	// Get initial ready steps (no dependencies)
	readySteps := resolver.GetReadySteps(completed)
	
	o.logger.Debug("Initial ready steps: %d", len(readySteps))
	for _, step := range readySteps {
		o.logger.Debug("  - %s", step.Name)
	}

	// Submit initial ready steps
	for _, step := range readySteps {
		if err := pool.SubmitStep(ctx, step); err != nil {
			return fmt.Errorf("failed to submit step %s: %w", step.Name, err)
		}
		completed[step.Name] = true // Mark as submitted
	}

	// Track remaining steps and loops
	stepsRemaining := make(map[string]*config.StepV2)
	for i := range o.workflow.Steps {
		step := &o.workflow.Steps[i]
		if !completed[step.Name] {
			stepsRemaining[step.Name] = step
		}
	}
	
	// Loops can run independently - submit them all at once
	loopsRemaining := make(map[string]*config.LoopV2)
	for i := range o.workflow.Loops {
		loop := &o.workflow.Loops[i]
		loopsRemaining[loop.Name] = loop
		
		// Submit loop immediately (loops have no dependencies)
		o.logger.Debug("Submitting loop: %s", loop.Name)
		if err := pool.SubmitLoop(ctx, loop); err != nil {
			return fmt.Errorf("failed to submit loop %s: %w", loop.Name, err)
		}
		completed[loop.Name] = true // Mark as submitted
	}

	totalRemaining := len(stepsRemaining) + len(loopsRemaining)

	// Event-driven coordination loop
	for totalRemaining > 0 {
		select {
		case completedName := <-pool.notifyCompletion:
			// Step or loop completed
			o.logger.Debug("Element completed: %s", completedName)
			
			// Check for error
			if err, hasError := pool.GetError(completedName); hasError {
				o.logger.Error("Element %s failed: %v", completedName, err)
				
				// Wait for in-flight elements based on error policy
				pool.Wait()
				
				// Copy results to orchestrator
				o.copyPoolResults(pool)
				
				return fmt.Errorf("element %s failed: %w", completedName, err)
			}
			
			// Mark as completed
			completed[completedName] = true
			
			// Remove from remaining (could be step or loop)
			delete(stepsRemaining, completedName)
			delete(loopsRemaining, completedName)
			totalRemaining = len(stepsRemaining) + len(loopsRemaining)
			
			// Find newly ready steps
			newReadySteps := resolver.GetReadySteps(completed)
			
			if len(newReadySteps) > 0 {
				o.logger.Debug("New ready steps: %d", len(newReadySteps))
				for _, step := range newReadySteps {
					o.logger.Debug("  - %s", step.Name)
				}
			}
			
			// Submit newly ready steps
			for _, step := range newReadySteps {
				if err := pool.SubmitStep(ctx, step); err != nil {
					pool.Wait()
					o.copyPoolResults(pool)
					return fmt.Errorf("failed to submit step %s: %w", step.Name, err)
				}
				completed[step.Name] = true // Mark as submitted
			}

		case <-ctx.Done():
			// Context cancelled (error or timeout)
			pool.Wait()
			o.copyPoolResults(pool)
			return ctx.Err()
		}
	}

	// Wait for all workers to complete
	pool.Wait()
	
	// End timeline tracking
	pool.timeline.End()
	
	// Copy results from pool to orchestrator
	o.copyPoolResults(pool)
	
	// Check for any errors
	allErrors := pool.GetAllErrors()
	if len(allErrors) > 0 {
		// Return first error
		for name, err := range allErrors {
			return fmt.Errorf("element %s failed: %w", name, err)
		}
	}

	// Output execution summary and timeline
	o.logger.Info("\n")
	o.logger.Info(pool.bufferedLogger.GetExecutionSummary())
	o.logger.Info(pool.timeline.GenerateGanttChart())
	
	// Calculate and display speedup
	speedup := pool.timeline.GetSpeedup()
	if speedup > 1.0 {
		sequential := pool.timeline.GetSequentialEstimate()
		parallel := pool.timeline.GetTotalDuration()
		o.logger.Info("Performance: %.2fx speedup (Sequential: %v, Parallel: %v)\n", 
			speedup, sequential.Round(time.Millisecond), parallel.Round(time.Millisecond))
	}

	o.logger.Step("\n[SUCCESS] Workflow completed (parallel mode)")
	return nil
}

// copyPoolResults copies results from worker pool to orchestrator
func (o *Orchestrator) copyPoolResults(pool *WorkflowWorkerPool) {
	results := pool.GetAllResults()
	for stepName, result := range results {
		o.stepResults[stepName] = result
		o.interpolator.Set(stepName, result)
	}
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
	} else if step.Loop != nil {
		err = o.executeLoopStep(ctx, step)
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
		// Apply error handling policy
		return o.handleStepError(step, err)
	}
	
	// Store result
	o.stepResults[step.Name] = result.Output
	o.interpolator.SetStepResult(step.Name, result.Output)

	o.logger.Output("Step %s result: %s", step.Name, result.Output)

	return nil
}

// handleStepError applies error handling policy for failed steps
func (o *Orchestrator) handleStepError(step *config.StepV2, err error) error {
	// Determine error policy
	onFailure := step.OnFailure
	if onFailure == "" {
		// Use workflow-level default
		onFailure = o.workflow.Execution.OnError
	}
	if onFailure == "" {
		// Ultimate default: halt
		onFailure = "halt"
	}
	
	o.logger.Warn("Step '%s' failed: %v", step.Name, err)
	o.logger.Warn("Error policy: %s", onFailure)
	
	switch onFailure {
	case "continue":
		// Log warning but continue workflow
		o.logger.Warn("Continuing workflow despite step failure (policy: continue)")
		// Store empty result
		o.stepResults[step.Name] = ""
		o.interpolator.SetStepResult(step.Name, "")
		return nil
		
	case "retry":
		// Retry logic would go here (future enhancement)
		o.logger.Warn("Retry not yet implemented, treating as halt")
		return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		
	case "halt", "cancel_all":
		fallthrough
	default:
		// Halt workflow execution
		o.logger.Error("Halting workflow due to step failure")
		return fmt.Errorf("step '%s' failed: %w", step.Name, err)
	}
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

	// Output consensus details with individual votes
	o.logger.Output("Step %s consensus result: %s", step.Name, result.Result)
	o.logger.Output("  Agreement: %.0f%%, Confidence: %s", result.Agreement*100, result.Confidence)
	
	// Show individual provider votes for transparency
	if len(result.Votes) > 0 {
		o.logger.Output("  Provider votes:")
		for provider, vote := range result.Votes {
			o.logger.Output("    - %s: %s", provider, vote)
		}
	}

	if !result.Success {
		return fmt.Errorf("consensus failed to reach agreement")
	}

	// Check validation result (for validation steps using SUCCESS/FAIL)
	// If the step name contains "validate", check if the result is SUCCESS
	if strings.Contains(step.Name, "validate") {
		// Strip markdown formatting and normalize
		cleaned := strings.TrimSpace(result.Result)
		cleaned = strings.ReplaceAll(cleaned, "**", "")  // Remove bold
		cleaned = strings.ReplaceAll(cleaned, "*", "")   // Remove italic
		cleaned = strings.ReplaceAll(cleaned, "`", "")   // Remove code
		resultUpper := strings.ToUpper(cleaned)
		
		if resultUpper != "SUCCESS" {
			return fmt.Errorf("validation failed: %s", result.Result)
		}
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

// SetStartFrom sets the step to start workflow from, skipping previous steps
func (o *Orchestrator) SetStartFrom(stepName string) {
	o.startFrom = stepName
}

// SetEndAt sets the step to end workflow at, skipping steps after
func (o *Orchestrator) SetEndAt(stepName string) {
	o.endAt = stepName
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
	// CRITICAL: Inherit output from parent logger (stdout in CLI, stderr in MCP serve mode)
	subLogger.SetOutput(o.logger.GetOutput())
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
	
	// Initialize loop executor if not already done
	if o.loopExecutor == nil {
		o.loopExecutor = NewLoopExecutor(
			o.appConfig,
			o.logger,
			o.interpolator,
			o.executor,
			o.executor.serverManager,
			o.embeddingService,
		)
	}
	
	o.logger.Info("Starting loop: %s", step.Name)
	
	// Convert step.Loop to LoopV2 config
	loopConfig := &config.LoopV2{
		Name:           step.Name,
		Workflow:       step.Loop.Workflow,
		Mode:           step.Loop.Mode,
		Items:          step.Loop.Items,
		With:           step.Loop.With,
		MaxIterations:  step.Loop.MaxIterations,
		Until:          step.Loop.Until,
		OnFailure:      step.Loop.OnFailure,
		MaxRetries:     step.Loop.MaxRetries,
		RetryDelay:     step.Loop.RetryDelay,
		MinSuccessRate: step.Loop.MinSuccessRate,
		TimeoutPerItem: step.Loop.TimeoutPerItem,
		TotalTimeout:   step.Loop.TotalTimeout,
		Accumulate:     step.Loop.Accumulate,
		Parallel:       step.Loop.Parallel,
		MaxWorkers:     step.Loop.MaxWorkers,
	}
	
	// Execute the loop using LoopExecutor
	result, err := o.loopExecutor.ExecuteLoop(ctx, loopConfig)
	if err != nil {
		o.logger.Warn("Loop %s failed: %v", step.Name, err)
		return err
	}
	
	o.logger.Info("Loop %s completed: %d iterations, exit: %s", 
		step.Name, result.Iterations, result.ExitReason)
	
	// Store result for access by subsequent steps
	o.interpolator.SetStepResult(step.Name, result.FinalOutput)
	
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
	
	// Extract directory from current workflow key for directory-aware resolution
	var contextDir string
	if o.workflowKey != "" {
		if idx := strings.LastIndex(o.workflowKey, "/"); idx != -1 {
			contextDir = o.workflowKey[:idx]
		}
	}
	
	// Use contextual lookup to support relative workflow references
	wf, exists := o.appConfig.GetWorkflowWithContext(workflow, contextDir)
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
	// CRITICAL: Inherit output from parent logger (stdout in CLI, stderr in MCP serve mode)
	subLogger.SetOutput(o.logger.GetOutput())
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

// markStepsAsCompleted marks steps before start-from and after end-at as completed
func (o *Orchestrator) markStepsAsCompleted(completed map[string]bool) error {
	var startStepIndex int = -1
	var endStepIndex int = -1
	
	// Find start-from step index
	if o.startFrom != "" {
		startStepExists := false
		
		for i, step := range o.workflow.Steps {
			if step.Name == o.startFrom {
				startStepExists = true
				startStepIndex = i
				break
			}
		}
		
		// Check loops too for start-from
		if !startStepExists {
			for _, loop := range o.workflow.Loops {
				if loop.Name == o.startFrom {
					startStepExists = true
					// For loops, mark all steps as skipped
					for _, step := range o.workflow.Steps {
						completed[step.Name] = true
					}
					o.logger.Info("Starting from loop: %s (all steps skipped)", o.startFrom)
					return nil
				}
			}
		}
		
		if !startStepExists {
			availableSteps := make([]string, 0, len(o.workflow.Steps)+len(o.workflow.Loops))
			for _, step := range o.workflow.Steps {
				availableSteps = append(availableSteps, step.Name)
			}
			for _, loop := range o.workflow.Loops {
				availableSteps = append(availableSteps, loop.Name)
			}
			return fmt.Errorf("start-from step '%s' not found in workflow. Available steps: %v", o.startFrom, availableSteps)
		}
		
		// Mark all steps before start-from as completed
		for i := 0; i < startStepIndex; i++ {
			step := o.workflow.Steps[i]
			completed[step.Name] = true
			o.logger.Debug("Skipped (before start-from): %s", step.Name)
		}
		
		o.logger.Info("Starting from step %d/%d: %s", startStepIndex+1, len(o.workflow.Steps), o.startFrom)
	}
	
	// Find end-at step index
	if o.endAt != "" {
		endStepExists := false
		
		for i, step := range o.workflow.Steps {
			if step.Name == o.endAt {
				endStepExists = true
				endStepIndex = i
				break
			}
		}
		
		// Check loops too for end-at
		if !endStepExists {
			for _, loop := range o.workflow.Loops {
				if loop.Name == o.endAt {
					endStepExists = true
					// For end-at on a loop, we can't really handle it well
					// So just log a warning
					o.logger.Warn("End-at specified for loop: %s (loops cannot be partially executed)", o.endAt)
					return nil
				}
			}
		}
		
		if !endStepExists {
			availableSteps := make([]string, 0, len(o.workflow.Steps)+len(o.workflow.Loops))
			for _, step := range o.workflow.Steps {
				availableSteps = append(availableSteps, step.Name)
			}
			for _, loop := range o.workflow.Loops {
				availableSteps = append(availableSteps, loop.Name)
			}
			return fmt.Errorf("end-at step '%s' not found in workflow. Available steps: %v", o.endAt, availableSteps)
		}
		
		// Mark all steps after end-at as completed
		for i := endStepIndex + 1; i < len(o.workflow.Steps); i++ {
			step := o.workflow.Steps[i]
			completed[step.Name] = true
			o.logger.Debug("Skipped (after end-at): %s", step.Name)
		}
		
		o.logger.Info("Ending at step %d/%d: %s", endStepIndex+1, len(o.workflow.Steps), o.endAt)
	}
	
	// Validate that start-from comes before end-at if both are specified
	if startStepIndex != -1 && endStepIndex != -1 {
		if startStepIndex > endStepIndex {
			return fmt.Errorf("start-from step '%s' (index %d) comes after end-at step '%s' (index %d)", 
				o.startFrom, startStepIndex+1, o.endAt, endStepIndex+1)
		}
	}
	
	return nil
}

// connectRAGServersIfNeeded detects RAG steps in workflow and connects required servers
func (o *Orchestrator) connectRAGServersIfNeeded(ctx context.Context) error {
	// Check if workflow uses RAG
	hasRAG := o.workflowUsesRAG()
	if !hasRAG {
		o.logger.Debug("No RAG steps detected, skipping RAG server connections")
		return nil
	}
	
	// Get RAG configuration
	if o.appConfig == nil || o.appConfig.RAG == nil {
		return fmt.Errorf("workflow uses RAG but no RAG configuration found")
	}
	
	o.logger.Info("Workflow uses RAG, connecting to RAG servers...")
	
	// Create dedicated server manager for RAG (internal connections only)
	o.ragServerManager = host.NewServerManagerWithOptions(true) // suppress console
	
	// Connect to all servers referenced in RAG config
	connectedServers := make(map[string]bool)
	for _, ragServerConfig := range o.appConfig.RAG.Servers {
		mcpServerName := ragServerConfig.MCPServer
		
		// Skip if already connected
		if connectedServers[mcpServerName] {
			continue
		}
		
		// Get server definition from app config
		serverDef, exists := o.appConfig.Servers[mcpServerName]
		if !exists {
			o.logger.Warn("RAG server '%s' not found in servers config", mcpServerName)
			continue
		}
		
		o.logger.Info("Connecting RAG server (internal): %s", mcpServerName)
		
		// Connect to server (internal, not exposed to LLM)
		_, err := o.ragServerManager.ConnectToServer(mcpServerName, serverDef, false)
		if err != nil {
			o.logger.Warn("Failed to connect RAG server '%s': %v", mcpServerName, err)
			continue
		}
		
		connectedServers[mcpServerName] = true
		o.logger.Info("✓ Connected RAG server: %s", mcpServerName)
	}
	
	if len(connectedServers) == 0 {
		return fmt.Errorf("workflow uses RAG but failed to connect to any RAG servers")
	}
	
	o.logger.Info("✓ Connected %d RAG server(s)", len(connectedServers))
	return nil
}

// workflowUsesRAG checks if the workflow or any child workflows use RAG
func (o *Orchestrator) workflowUsesRAG() bool {
	// Check steps in current workflow
	for _, step := range o.workflow.Steps {
		if step.Rag != nil {
			return true
		}
		
		// Check if step uses a loop with child workflow
		if step.Loop != nil && step.Loop.Workflow != "" {
			childWorkflow, exists := o.appConfig.GetWorkflow(step.Loop.Workflow)
			if exists {
				// Check child workflow steps for RAG
				for _, childStep := range childWorkflow.Steps {
					if childStep.Rag != nil {
						return true
					}
				}
			}
		}
	}
	
	return false
}
