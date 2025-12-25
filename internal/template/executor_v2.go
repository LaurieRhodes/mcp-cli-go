package template

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// ExecutorV2 executes template v2 workflows
type ExecutorV2 struct {
	template         *config.TemplateV2
	resolver         *VariableResolver
	aiProvider       AIProvider
	mcpManager       MCPManager
	executedSteps    map[string]bool // Track which steps have been executed
	templateRegistry TemplateRegistry // Not a pointer - interfaces don't need pointers
	recursionDepth   int
	maxRecursion     int
}

// TemplateRegistry provides access to all available templates
type TemplateRegistry interface {
	GetTemplate(name string) (*config.TemplateV2, error)
	ListTemplates() []string
}

// AIProvider interface for LLM interactions
type AIProvider interface {
	CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error)
	GetProviderType() domain.ProviderType
}

// MCPManager interface for MCP server interactions
type MCPManager interface {
	GetAvailableTools() ([]domain.Tool, error)
	ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (string, error)
}

// NewExecutorV2 creates a new template v2 executor
func NewExecutorV2(
	template *config.TemplateV2,
	aiProvider AIProvider,
	mcpManager MCPManager,
) *ExecutorV2 {
	return &ExecutorV2{
		template:         template,
		resolver:         NewVariableResolver(),
		aiProvider:       aiProvider,
		mcpManager:       mcpManager,
		executedSteps:    make(map[string]bool),
		templateRegistry: nil, // Will be set if needed
		recursionDepth:   0,
		maxRecursion:     10, // Default max recursion depth
	}
}

// NewExecutorV2WithRegistry creates an executor with template registry support
func NewExecutorV2WithRegistry(
	template *config.TemplateV2,
	aiProvider AIProvider,
	mcpManager MCPManager,
	registry TemplateRegistry,
) *ExecutorV2 {
	executor := NewExecutorV2(template, aiProvider, mcpManager)
	executor.templateRegistry = registry
	return executor
}

// Execute runs the complete template workflow
func (e *ExecutorV2) Execute(ctx context.Context, input string) (*ExecutionResult, error) {
	startTime := time.Now()

	logging.Info("Starting template v2 execution: %s (version: %s)", e.template.Name, e.template.Version)

	// Initialize built-in variables
	e.initializeBuiltInVariables(input)

	// Initialize template-level variables
	if e.template.Config != nil && e.template.Config.Variables != nil {
		e.resolver.SetMultiple(e.template.Config.Variables)
	}

	// Execute steps
	stepResults := make([]*StepResult, 0, len(e.template.Steps))
	var finalOutput string

	for i, step := range e.template.Steps {
		logging.Debug("Executing step %d: %s (type: %s)", i+1, step.Name, step.GetStepType())

		stepResult, err := e.executeStep(ctx, &step)
		stepResults = append(stepResults, stepResult)

		if err != nil {
			logging.Error("Step %s failed: %v", step.Name, err)
			
			// Check error handling strategy
			if e.shouldStopOnError(&step, err) {
				return &ExecutionResult{
					Success:       false,
					FinalOutput:   finalOutput,
					StepResults:   stepResults,
					ExecutionTime: time.Since(startTime),
					Error:         err,
				}, err
			}
			
			// Continue to next step
			continue
		}

		// Store step output
		finalOutput = stepResult.Output
	}

	logging.Info("Template execution completed: %s (time: %v)", e.template.Name, time.Since(startTime))

	return &ExecutionResult{
		Success:       true,
		FinalOutput:   finalOutput,
		StepResults:   stepResults,
		ExecutionTime: time.Since(startTime),
	}, nil
}

// executeStep executes a single step based on its type
func (e *ExecutorV2) executeStep(ctx context.Context, step *config.WorkflowStepV2) (*StepResult, error) {
	startTime := time.Now()

	result := &StepResult{
		StepName:  step.Name,
		StartTime: startTime,
	}

	// Check condition
	if step.Condition != "" {
		shouldExecute, err := e.evaluateCondition(step.Condition)
		if err != nil {
			return result, fmt.Errorf("condition evaluation failed: %w", err)
		}
		if !shouldExecute {
			logging.Debug("Step %s skipped due to condition: %s", step.Name, step.Condition)
			result.Skipped = true
			result.ExecutionTime = time.Since(startTime)
			// Mark skipped steps as executed so they don't break dependencies
			e.executedSteps[step.Name] = true
			return result, nil
		}
	}

	// Check dependencies
	if len(step.DependsOn) > 0 {
		if err := e.checkDependencies(step.DependsOn); err != nil {
			return result, fmt.Errorf("dependency check failed: %w", err)
		}
	}

	// Execute based on step type
	var output string
	var err error

	switch step.GetStepType() {
	case config.StepTypeBasic:
		output, err = e.executeBasicStep(ctx, step)
	
	case config.StepTypeParallel:
		output, err = e.executeParallel(ctx, step)
	
	case config.StepTypeLoop:
		output, err = e.executeLoop(ctx, step)
	
	case config.StepTypeTransform:
		output, err = e.executeTransform(ctx, step)
	
	case config.StepTypeUse:
		output, err = e.executeReuse(ctx, step)
	
	case config.StepTypeTemplate:
		output, err = e.executeTemplateCall(ctx, step)
	
	case config.StepTypeNested:
		output, err = e.executeNested(ctx, step)
	
	default:
		err = fmt.Errorf("unknown step type: %s", step.GetStepType())
	}

	result.ExecutionTime = time.Since(startTime)
	result.Output = output
	result.Error = err

	if err != nil {
		result.Success = false
		return result, err
	}

	// Store output(s)
	if err := e.storeOutput(step, output); err != nil {
		return result, fmt.Errorf("failed to store output: %w", err)
	}

	// Mark step as successfully executed
	e.executedSteps[step.Name] = true

	result.Success = true
	return result, nil
}

// executeBasicStep executes a basic LLM request step
func (e *ExecutorV2) executeBasicStep(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	// Resolve prompt
	prompt, err := e.resolver.ResolveString(step.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to resolve prompt: %w", err)
	}

	systemPrompt := ""
	if step.SystemPrompt != "" {
		systemPrompt, err = e.resolver.ResolveString(step.SystemPrompt)
		if err != nil {
			return "", fmt.Errorf("failed to resolve system prompt: %w", err)
		}
	}

	// Get tools from MCP servers
	tools, err := e.getStepTools(step)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	// Create completion request
	messages := []domain.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		SystemPrompt: systemPrompt,
		Tools:        tools,
		Temperature:  e.getTemperature(step),
		MaxTokens:    e.getMaxTokens(step),
	}

	// Execute with retry if configured
	response, err := e.executeWithRetry(ctx, completionReq, step.ErrorHandling)
	if err != nil {
		return "", err
	}

	return response.Response, nil
}

// initializeBuiltInVariables sets up built-in template variables
func (e *ExecutorV2) initializeBuiltInVariables(input string) {
	e.resolver.SetVariable("stdin", input)
	e.resolver.SetVariable("input_data", input)
	e.resolver.SetVariable("template.name", e.template.Name)
	e.resolver.SetVariable("template.version", e.template.Version)
	e.resolver.SetVariable("execution.timestamp", time.Now().Format(time.RFC3339))
}

// storeOutput stores step output using configured output mapping
func (e *ExecutorV2) storeOutput(step *config.WorkflowStepV2, output string) error {
	if step.Output == nil {
		// No output configuration, just store with step name
		e.resolver.SetStepOutput(step.Name, output)
		return nil
	}

	switch out := step.Output.(type) {
	case string:
		// Simple output - store with given name
		e.resolver.SetVariable(out, output)
		
	case map[string]interface{}:
		// Complex output - extract fields
		// For Phase 1, we'll store the whole output for each field
		// Full implementation would parse output and extract specific fields
		for fieldName := range out {
			e.resolver.SetVariable(fieldName, output)
		}
		
	default:
		return fmt.Errorf("invalid output configuration type: %T", step.Output)
	}

	return nil
}

// evaluateCondition evaluates a step condition
func (e *ExecutorV2) evaluateCondition(condition string) (bool, error) {
	return e.resolver.EvaluateCondition(condition)
}

// checkDependencies verifies all dependencies are satisfied
func (e *ExecutorV2) checkDependencies(dependencies []string) error {
	for _, dep := range dependencies {
		if !e.executedSteps[dep] {
			return fmt.Errorf("dependency not satisfied: %s (step not executed)", dep)
		}
	}
	return nil
}

// getStepTools retrieves tools for a step based on configured servers
func (e *ExecutorV2) getStepTools(step *config.WorkflowStepV2) ([]domain.Tool, error) {
	if step.Servers == nil {
		// No specific servers, get all available tools
		return e.mcpManager.GetAvailableTools()
	}

	if len(step.Servers) == 0 {
		// Empty array means no tools
		return []domain.Tool{}, nil
	}

	// For Phase 1, return all available tools
	// Full implementation would filter by server names
	return e.mcpManager.GetAvailableTools()
}

// getTemperature returns the temperature for a step
func (e *ExecutorV2) getTemperature(step *config.WorkflowStepV2) float64 {
	if step.Temperature != 0 {
		return step.Temperature
	}
	if e.template.Config != nil && e.template.Config.Defaults != nil {
		return e.template.Config.Defaults.Temperature
	}
	return 0.7 // Default
}

// getMaxTokens returns max tokens for a step
func (e *ExecutorV2) getMaxTokens(step *config.WorkflowStepV2) int {
	if step.MaxTokens != 0 {
		return step.MaxTokens
	}
	if e.template.Config != nil && e.template.Config.Defaults != nil {
		return e.template.Config.Defaults.MaxTokens
	}
	return 0 // Use provider default
}

// executeWithRetry executes a completion request with retry logic
func (e *ExecutorV2) executeWithRetry(
	ctx context.Context,
	req *domain.CompletionRequest,
	errorConfig *config.ErrorHandlingConfig,
) (*domain.CompletionResponse, error) {
	if errorConfig == nil || errorConfig.MaxRetries == 0 {
		// No retry configured
		return e.aiProvider.CreateCompletion(ctx, req)
	}

	maxRetries := errorConfig.MaxRetries
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := e.calculateBackoff(attempt-1, errorConfig)
			logging.Debug("Retry attempt %d after %v", attempt, delay)
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		response, err := e.aiProvider.CreateCompletion(ctx, req)
		if err == nil {
			return response, nil
		}

		lastErr = err
		logging.Warn("Attempt %d failed: %v", attempt+1, err)
	}

	// All retries failed
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// calculateBackoff calculates retry delay with backoff
func (e *ExecutorV2) calculateBackoff(attempt int, config *config.ErrorHandlingConfig) time.Duration {
	initialDelay := 1 * time.Second
	if config.InitialDelay != "" {
		if d, err := time.ParseDuration(config.InitialDelay); err == nil {
			initialDelay = d
		}
	}

	if config.RetryBackoff == "exponential" {
		return initialDelay * time.Duration(1<<uint(attempt))
	}

	// Linear backoff
	return initialDelay * time.Duration(attempt+1)
}

// shouldStopOnError determines if execution should stop on error
func (e *ExecutorV2) shouldStopOnError(step *config.WorkflowStepV2, err error) bool {
	if step.ErrorHandling == nil {
		// No error handling config, stop by default
		return true
	}

	switch step.ErrorHandling.OnFailure {
	case "continue":
		return false
	case "stop":
		return true
	default:
		// Default is to stop
		return true
	}
}

// executeParallel executes multiple steps in parallel
func (e *ExecutorV2) executeParallel(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	if step.Parallel == nil || len(step.Parallel.Steps) == 0 {
		return "", fmt.Errorf("parallel execution requires sub-steps")
	}

	parallelSteps := step.Parallel.Steps
	logging.Debug("Parallel step %s: executing %d steps concurrently", step.Name, len(parallelSteps))

	// Determine max concurrency
	maxConcurrent := step.Parallel.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = len(parallelSteps) // No limit
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, maxConcurrent)

	// Results collection
	results := make([]parallelResult, len(parallelSteps))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Execute each step in parallel
	for i, parallelStep := range parallelSteps {
		wg.Add(1)
		go func(idx int, pStep config.WorkflowStepV2) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			logging.Debug("Starting parallel sub-step %d: %s", idx+1, pStep.Name)

			// Execute the step
			stepResult, err := e.executeStep(ctx, &pStep)

			// Store result
			mu.Lock()
			if err != nil {
				results[idx] = parallelResult{index: idx, err: err}
			} else {
				results[idx] = parallelResult{index: idx, output: stepResult.Output}
			}
			mu.Unlock()

			logging.Debug("Completed parallel sub-step %d: %s", idx+1, pStep.Name)
		}(i, parallelStep)
	}

	// Wait for all to complete
	wg.Wait()

	// Check for errors
	var firstError error
	for _, result := range results {
		if result.err != nil && firstError == nil {
			firstError = result.err
		}
	}

	if firstError != nil {
		// Check error handling
		if step.ErrorHandling != nil && step.ErrorHandling.OnFailure == "continue" {
			logging.Warn("Parallel execution had errors but continuing: %v", firstError)
		} else {
			return "", fmt.Errorf("parallel execution failed: %w", firstError)
		}
	}

	// Aggregate results
	if step.Parallel.Aggregate != nil {
		return e.aggregateParallelResults(results, step.Parallel.Aggregate)
	}

	// Default: return results as JSON array
	outputs := make([]string, len(results))
	for i, result := range results {
		outputs[i] = result.output
	}

	jsonOutput, err := json.Marshal(outputs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal parallel results: %w", err)
	}

	return string(jsonOutput), nil
}

// aggregateParallelResults combines parallel execution results
func (e *ExecutorV2) aggregateParallelResults(results []parallelResult, config *config.AggregateConfig) (string, error) {
	switch config.Combine {
	case "merge":
		// Merge all outputs into single string
		var merged string
		for _, result := range results {
			if result.err == nil {
				merged += result.output + "\n"
			}
		}
		return merged, nil

	case "array":
		// Return as JSON array (default)
		outputs := make([]string, len(results))
		for i, result := range results {
			outputs[i] = result.output
		}
		jsonOutput, err := json.Marshal(outputs)
		if err != nil {
			return "", fmt.Errorf("failed to marshal array results: %w", err)
		}
		return string(jsonOutput), nil

	default:
		// Default to array
		outputs := make([]string, len(results))
		for i, result := range results {
			outputs[i] = result.output
		}
		jsonOutput, err := json.Marshal(outputs)
		if err != nil {
			return "", fmt.Errorf("failed to marshal results: %w", err)
		}
		return string(jsonOutput), nil
	}
}

// executeLoop executes a step with iteration over a collection
func (e *ExecutorV2) executeLoop(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	// Resolve the collection to iterate over
	collectionExpr := step.ForEach
	collection, err := e.resolver.ResolveExpression(collectionExpr)
	if err != nil {
		return "", fmt.Errorf("failed to resolve for_each collection: %w", err)
	}

	// Convert to array
	var items []interface{}
	switch v := collection.(type) {
	case []interface{}:
		items = v
	case string:
		// Try to parse as JSON array
		var parsed []interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			items = parsed
		} else {
			// Treat as single item
			items = []interface{}{v}
		}
	default:
		return "", fmt.Errorf("for_each value is not iterable: %T", collection)
	}

	if len(items) == 0 {
		logging.Debug("Loop step %s: no items to iterate", step.Name)
		return "[]", nil
	}

	logging.Debug("Loop step %s: iterating over %d items", step.Name, len(items))

	// Determine item variable name (default: "item")
	itemName := step.ItemName
	if itemName == "" {
		itemName = "item"
	}

	// Execute step for each item
	results := make([]string, 0, len(items))
	for i, item := range items {
		logging.Debug("Loop iteration %d/%d for step %s", i+1, len(items), step.Name)

		// Set loop variables
		e.resolver.SetVariable(itemName, item)
		e.resolver.SetVariable("index", i)
		e.resolver.SetVariable("first", i == 0)
		e.resolver.SetVariable("last", i == len(items)-1)

		// Resolve prompt with current item
		prompt, err := e.resolver.ResolveString(step.Prompt)
		if err != nil {
			return "", fmt.Errorf("iteration %d: failed to resolve prompt: %w", i, err)
		}

		systemPrompt := ""
		if step.SystemPrompt != "" {
			systemPrompt, err = e.resolver.ResolveString(step.SystemPrompt)
			if err != nil {
				return "", fmt.Errorf("iteration %d: failed to resolve system prompt: %w", i, err)
			}
		}

		// Get tools
		tools, err := e.getStepTools(step)
		if err != nil {
			return "", fmt.Errorf("iteration %d: failed to get tools: %w", i, err)
		}

		// Create completion request
		messages := []domain.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		completionReq := &domain.CompletionRequest{
			Messages:     messages,
			SystemPrompt: systemPrompt,
			Tools:        tools,
			Temperature:  e.getTemperature(step),
			MaxTokens:    e.getMaxTokens(step),
		}

		// Execute with retry if configured
		response, err := e.executeWithRetry(ctx, completionReq, step.ErrorHandling)
		if err != nil {
			if step.ErrorHandling != nil && step.ErrorHandling.OnFailure == "continue" {
				logging.Warn("Loop iteration %d failed but continuing: %v", i, err)
				results = append(results, fmt.Sprintf("Error: %v", err))
				continue
			}
			return "", fmt.Errorf("iteration %d failed: %w", i, err)
		}

		results = append(results, response.Response)
	}

	// Clean up loop variables
	e.resolver.DeleteVariable(itemName)
	e.resolver.DeleteVariable("index")
	e.resolver.DeleteVariable("first")
	e.resolver.DeleteVariable("last")

	// Return results as JSON array
	jsonResults, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal loop results: %w", err)
	}

	return string(jsonResults), nil
}

// executeTransform executes data transformation operations
func (e *ExecutorV2) executeTransform(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	if step.Transform == nil {
		return "", fmt.Errorf("transform config is required for transform steps")
	}

	// Resolve input data
	inputData, err := e.resolver.ResolveExpression(step.Transform.Input)
	if err != nil {
		return "", fmt.Errorf("failed to resolve transform input: %w", err)
	}

	// If input is a JSON string, try to parse it
	if strData, ok := inputData.(string); ok {
		var parsed interface{}
		if err := json.Unmarshal([]byte(strData), &parsed); err == nil {
			inputData = parsed
		}
	}

	logging.Debug("Transform step %s: processing %d operations", step.Name, len(step.Transform.Operations))

	// Apply each operation in sequence
	current := inputData
	for i, op := range step.Transform.Operations {
		logging.Debug("Applying transform operation %d: %s", i+1, op.Type)
		current, err = e.applyTransformOperation(current, &op)
		if err != nil {
			return "", fmt.Errorf("transform operation %d (%s) failed: %w", i+1, op.Type, err)
		}
	}

	// Convert result to string
	result := e.resolver.valueToString(current)
	return result, nil
}

// applyTransformOperation applies a single transformation operation
func (e *ExecutorV2) applyTransformOperation(data interface{}, op *config.TransformOperation) (interface{}, error) {
	switch op.Type {
	case "filter":
		return e.transformFilter(data, op)
	case "map":
		return e.transformMap(data, op)
	case "sort":
		return e.transformSort(data, op)
	case "limit":
		return e.transformLimit(data, op)
	case "pluck":
		return e.transformPluck(data, op)
	case "group":
		return e.transformGroup(data, op)
	default:
		return nil, fmt.Errorf("unknown transform operation: %s", op.Type)
	}
}

// transformFilter filters array items based on condition
func (e *ExecutorV2) transformFilter(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter requires an array, got %T", data)
	}

	if op.Condition == "" {
		return arr, nil // No condition, return all
	}

	filtered := make([]interface{}, 0)
	for _, item := range arr {
		// Set item as temporary variable for condition evaluation
		oldItem, _ := e.resolver.GetVariable("item")
		e.resolver.SetVariable("item", item)
		
		// If item is a map, also set each field as a variable for easy access
		var savedFields map[string]interface{}
		if itemMap, ok := item.(map[string]interface{}); ok {
			savedFields = make(map[string]interface{})
			for key, val := range itemMap {
				if old, exists := e.resolver.GetVariable(key); exists {
					savedFields[key] = old
				}
				e.resolver.SetVariable(key, val)
			}
		}

		// Evaluate condition
		matches, err := e.resolver.EvaluateCondition(op.Condition)
		
		// Restore variables
		if oldItem != nil {
			e.resolver.SetVariable("item", oldItem)
		} else {
			e.resolver.DeleteVariable("item")
		}
		
		// Restore field variables
		if savedFields != nil {
			for key, oldVal := range savedFields {
				e.resolver.SetVariable(key, oldVal)
			}
			// Delete any fields that weren't there before
			if itemMap, ok := item.(map[string]interface{}); ok {
				for key := range itemMap {
					if _, wasSaved := savedFields[key]; !wasSaved {
						e.resolver.DeleteVariable(key)
					}
				}
			}
		}

		if err != nil {
			logging.Warn("Filter condition evaluation failed for item: %v", err)
			continue
		}

		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// transformMap applies a transformation to each item
func (e *ExecutorV2) transformMap(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("map requires an array, got %T", data)
	}

	// For now, implement field extraction (pluck-like behavior)
	// Full expression evaluation would require more complex parsing
	if op.Fields == nil {
		return arr, nil
	}

	mapped := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// Extract specified fields
			if fieldList, ok := op.Fields.([]interface{}); ok {
				result := make(map[string]interface{})
				for _, f := range fieldList {
					if fieldName, ok := f.(string); ok {
						if val, exists := itemMap[fieldName]; exists {
							result[fieldName] = val
						}
					}
				}
				mapped = append(mapped, result)
			}
		} else {
			mapped = append(mapped, item)
		}
	}

	return mapped, nil
}

// transformSort sorts an array
func (e *ExecutorV2) transformSort(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort requires an array, got %T", data)
	}

	// For Phase 2, implement basic sorting
	// Full implementation would support custom comparators and nested field access
	// For now, return as-is (sorting complex data structures requires more work)
	logging.Debug("Sort operation: returning data as-is (full sort not yet implemented)")
	return arr, nil
}

// transformLimit limits the number of items
func (e *ExecutorV2) transformLimit(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("limit requires an array, got %T", data)
	}

	if op.Count <= 0 || op.Count >= len(arr) {
		return arr, nil
	}

	return arr[:op.Count], nil
}

// transformPluck extracts a specific field from each item
func (e *ExecutorV2) transformPluck(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("pluck requires an array, got %T", data)
	}

	if op.Key == "" {
		return nil, fmt.Errorf("pluck requires a key")
	}

	plucked := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if val, exists := itemMap[op.Key]; exists {
				plucked = append(plucked, val)
			}
		}
	}

	return plucked, nil
}

// transformGroup groups items by a key
func (e *ExecutorV2) transformGroup(data interface{}, op *config.TransformOperation) (interface{}, error) {
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("group requires an array, got %T", data)
	}

	if op.Key == "" {
		return nil, fmt.Errorf("group requires a key")
	}

	grouped := make(map[string][]interface{})
	for _, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if keyVal, exists := itemMap[op.Key]; exists {
				keyStr := fmt.Sprintf("%v", keyVal)
				grouped[keyStr] = append(grouped[keyStr], item)
			}
		}
	}

	return grouped, nil
}

// executeReuse executes a reusable step definition
func (e *ExecutorV2) executeReuse(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	if step.Use == "" {
		return "", fmt.Errorf("use field is required for reuse steps")
	}

	// Get step definition
	stepDef, ok := e.template.StepDefinitions[step.Use]
	if !ok {
		return "", fmt.Errorf("step definition not found: %s", step.Use)
	}

	logging.Debug("Reuse step %s: using definition %s", step.Name, step.Use)

	// Prepare inputs from step configuration
	if step.Inputs != nil {
		for key, value := range step.Inputs {
			// Resolve input value
			if strVal, ok := value.(string); ok {
				resolved, err := e.resolver.ResolveString(strVal)
				if err != nil {
					return "", fmt.Errorf("failed to resolve input %s: %w", key, err)
				}
				e.resolver.SetVariable(key, resolved)
			} else {
				e.resolver.SetVariable(key, value)
			}
		}
	}

	// Resolve prompt from step definition
	prompt, err := e.resolver.ResolveString(stepDef.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to resolve prompt: %w", err)
	}

	systemPrompt := ""
	if stepDef.SystemPrompt != "" {
		systemPrompt, err = e.resolver.ResolveString(stepDef.SystemPrompt)
		if err != nil {
			return "", fmt.Errorf("failed to resolve system prompt: %w", err)
		}
	}

	// Get tools - use servers from step definition
	var tools []domain.Tool
	if len(stepDef.Servers) > 0 {
		// Filter tools by specified servers (simplified for Phase 2)
		tools, err = e.mcpManager.GetAvailableTools()
		if err != nil {
			return "", fmt.Errorf("failed to get tools: %w", err)
		}
	}

	// Create completion request
	messages := []domain.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Use temperature and model from step definition
	temperature := stepDef.Temperature
	if temperature == 0 {
		temperature = e.getTemperature(step)
	}

	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		SystemPrompt: systemPrompt,
		Tools:        tools,
		Temperature:  temperature,
		MaxTokens:    e.getMaxTokens(step),
	}

	// Execute with retry if configured
	response, err := e.executeWithRetry(ctx, completionReq, step.ErrorHandling)
	if err != nil {
		return "", err
	}

	return response.Response, nil
}

// executeTemplateCall executes another template as a sub-workflow
func (e *ExecutorV2) executeTemplateCall(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	if step.Template == "" {
		return "", fmt.Errorf("template field is required for template call steps")
	}

	// Check if template registry is available
	if e.templateRegistry == nil {
		return "", fmt.Errorf("template registry not available - cannot call template: %s", step.Template)
	}

	// Check recursion depth
	if e.recursionDepth >= e.maxRecursion {
		return "", fmt.Errorf("maximum template recursion depth exceeded (%d)", e.maxRecursion)
	}

	logging.Info("Calling template: %s (depth: %d)", step.Template, e.recursionDepth+1)

	// Get the target template
	targetTemplate, err := e.templateRegistry.GetTemplate(step.Template)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", step.Template, err)
	}

	// Resolve template input
	var templateInput string
	if step.TemplateInput != "" {
		// Use specified input expression
		templateInput, err = e.resolver.ResolveString(step.TemplateInput)
		if err != nil {
			return "", fmt.Errorf("failed to resolve template input: %w", err)
		}
	} else if step.Inputs != nil {
		// Use inputs map (legacy support)
		if inputData, ok := step.Inputs["input_data"]; ok {
			if strData, ok := inputData.(string); ok {
				templateInput, err = e.resolver.ResolveString(strData)
				if err != nil {
					return "", fmt.Errorf("failed to resolve input_data: %w", err)
				}
			} else {
				templateInput = fmt.Sprintf("%v", inputData)
			}
		}
	} else {
		// Default: pass current stdin/input_data
		if val, exists := e.resolver.GetVariable("stdin"); exists {
			templateInput = fmt.Sprintf("%v", val)
		}
	}

	logging.Debug("Template %s input: %s", step.Template, templateInput)

	// Create executor for sub-template with incremented recursion depth
	subExecutor := &ExecutorV2{
		template:         targetTemplate,
		resolver:         NewVariableResolver(), // New resolver for isolated variable scope
		aiProvider:       e.aiProvider,          // Share AI provider
		mcpManager:       e.mcpManager,          // Share MCP manager
		executedSteps:    make(map[string]bool),
		templateRegistry: e.templateRegistry,    // Share template registry
		recursionDepth:   e.recursionDepth + 1,  // Increment depth
		maxRecursion:     e.maxRecursion,
	}

	// Execute the sub-template
	result, err := subExecutor.Execute(ctx, templateInput)
	if err != nil {
		return "", fmt.Errorf("template %s execution failed: %w", step.Template, err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("template %s returned error: %w", step.Template, result.Error)
	}

	logging.Info("Template %s completed successfully", step.Template)

	// Return the final output from the sub-template
	return result.FinalOutput, nil
}

func (e *ExecutorV2) executeNested(ctx context.Context, step *config.WorkflowStepV2) (string, error) {
	return "", fmt.Errorf("nested steps not implemented in Phase 1")
}

// ExecutionResult represents the result of template execution
type ExecutionResult struct {
	Success       bool
	FinalOutput   string
	StepResults   []*StepResult
	ExecutionTime time.Duration
	Error         error
}

// StepResult represents the result of a single step
type StepResult struct {
	StepName      string
	Success       bool
	Skipped       bool
	Output        string
	StartTime     time.Time
	ExecutionTime time.Duration
	Error         error
}

// parallelResult represents the result of a parallel step execution
type parallelResult struct {
	index  int
	output string
	err    error
}
