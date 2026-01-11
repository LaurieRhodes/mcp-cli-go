package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// ExecuteIterateLoop executes a loop in iterate mode over an array of items
func (le *LoopExecutor) ExecuteIterateLoop(ctx context.Context, loop *config.LoopV2) (*config.LoopExecutionResult, error) {
	le.logger.Info("Starting iterate loop: %s", loop.Name)
	
	// Validate loop configuration
	if err := loop.Validate(); err != nil {
		return nil, fmt.Errorf("loop validation failed: %w", err)
	}
	
	// Parse items template to get array source
	itemsSource, err := le.interpolator.Interpolate(loop.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to interpolate items source '%s': %w", loop.Items, err)
	}
	
	// Parse array from source
	items, err := le.parseArrayInput(itemsSource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse array from items source: %w", err)
	}
	
	totalItems := len(items)
	if totalItems == 0 {
		le.logger.Warn("No items to process in iterate loop")
		return &config.LoopExecutionResult{
			TotalItems: 0,
			Success:    true,
			ExitReason: "no_items",
		}, nil
	}
	
	le.logger.Info("Processing %d items (max_iterations: %d)", totalItems, loop.MaxIterations)
	
	// Limit items to max_iterations
	if totalItems > loop.MaxIterations {
		le.logger.Warn("Item count (%d) exceeds max_iterations (%d), limiting to %d items", 
			totalItems, loop.MaxIterations, loop.MaxIterations)
		items = items[:loop.MaxIterations]
		totalItems = loop.MaxIterations
	}
	
	// Get the workflow to execute
	workflow, exists := le.appConfig.GetWorkflow(loop.Workflow)
	if !exists {
		return nil, fmt.Errorf("loop workflow '%s' not found", loop.Workflow)
	}
	
	// Initialize result tracking
	result := &config.LoopExecutionResult{
		TotalItems:  totalItems,
		Succeeded:   0,
		Failed:      0,
		Skipped:     0,
		FailedItems: make([]int, 0),
		AllOutputs:  make([]string, 0, totalItems),
	}
	
	startTime := time.Now()
	
	// Process each item
	for index, item := range items {
		itemResult := le.processIterationItem(ctx, loop, workflow, index, item, result)
		
		// Handle failure modes
		if !itemResult.Success {
			if loop.OnFailure == "halt" {
				result.ExitReason = "failure"
				result.Duration = time.Since(startTime)
				result.Success = false
				return result, fmt.Errorf("iteration %d failed, halting: %s", index, itemResult.Error)
			}
			// Continue mode - already tracked in result
		}
	}
	
	// Calculate final result
	result.Duration = time.Since(startTime)
	result.Iterations = result.Succeeded + result.Failed + result.Skipped
	
	// Set final output (last successful output, or empty if all failed)
	if len(result.AllOutputs) > 0 {
		result.FinalOutput = result.AllOutputs[len(result.AllOutputs)-1]
	} else {
		result.FinalOutput = ""
	}
	
	// Check success rate
	if loop.MinSuccessRate > 0 {
		result.Success = result.CheckSuccessRate(loop.MinSuccessRate)
		if !result.Success {
			result.ExitReason = "success_rate_not_met"
			actualRate := float64(result.Succeeded) / float64(result.TotalItems)
			le.logger.Warn("Loop failed: success rate %.2f%% < required %.2f%%", 
				actualRate*100, loop.MinSuccessRate*100)
		} else {
			result.ExitReason = "completed"
		}
	} else {
		result.Success = true
		result.ExitReason = "completed"
	}
	
	// Log summary
	le.logger.Info("Loop completed: %d/%d succeeded (%.1f%%), %d failed, duration: %s",
		result.Succeeded, result.TotalItems,
		float64(result.Succeeded)/float64(result.TotalItems)*100,
		result.Failed, result.Duration)
	
	// Store result for later access
	le.storeIterateLoopResult(loop, result)
	
	return result, nil
}

// processIterationItem processes a single item in the iteration
func (le *LoopExecutor) processIterationItem(
	ctx context.Context,
	loop *config.LoopV2,
	workflow *config.WorkflowV2,
	index int,
	item interface{},
	result *config.LoopExecutionResult,
) *itemExecutionResult {
	itemID := le.extractItemID(item, index)
	
	le.logger.Info("[LOOP] %s: Item %d/%d (%s) - started", 
		loop.Name, index+1, result.TotalItems, itemID)
	
	startTime := time.Now()
	
	// Set loop variables for interpolation
	le.interpolator.SetIterateLoopVars(index, item, result.TotalItems, result.Succeeded, result.Failed)
	
	// Prepare input for workflow (merge loop.With and current item)
	inputData, err := le.prepareIterateInput(loop, item)
	if err != nil {
		duration := time.Since(startTime)
		le.logger.Warn("[LOOP] %s: Item %d/%d (%s) - input preparation failed (%s): %v",
			loop.Name, index+1, result.TotalItems, itemID, duration, err)
		result.Failed++
		result.FailedItems = append(result.FailedItems, index)
		return &itemExecutionResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	// Execute workflow with retry logic
	var output string
	var execErr error
	
	maxAttempts := 1
	if loop.OnFailure == "retry" && loop.MaxRetries > 0 {
		maxAttempts = loop.MaxRetries + 1 // Initial attempt + retries
	}
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			le.logger.Info("[LOOP] %s: Item %d/%d (%s) - retrying (%d/%d)",
				loop.Name, index+1, result.TotalItems, itemID, attempt-1, loop.MaxRetries)
			
			// Apply retry delay
			if loop.RetryDelay != "" {
				if delay, err := time.ParseDuration(loop.RetryDelay); err == nil {
					time.Sleep(delay)
				}
			}
		}
		
		output, execErr = le.executeWorkflow(ctx, workflow, inputData)
		if execErr == nil {
			break // Success
		}
	}
	
	duration := time.Since(startTime)
	
	if execErr != nil {
		le.logger.Warn("[LOOP] %s: Item %d/%d (%s) - failed (%s): %v",
			loop.Name, index+1, result.TotalItems, itemID, duration, execErr)
		result.Failed++
		result.FailedItems = append(result.FailedItems, index)
		return &itemExecutionResult{
			Success: false,
			Error:   execErr.Error(),
		}
	}
	
	// Success
	le.logger.Info("[LOOP] %s: Item %d/%d (%s) - completed (%s)",
		loop.Name, index+1, result.TotalItems, itemID, duration)
	result.Succeeded++
	result.AllOutputs = append(result.AllOutputs, output)
	
	return &itemExecutionResult{
		Success: true,
		Output:  output,
	}
}

// itemExecutionResult tracks individual item execution
type itemExecutionResult struct {
	Success bool
	Output  string
	Error   string
}

// prepareIterateInput prepares input for iterate loop iteration
func (le *LoopExecutor) prepareIterateInput(loop *config.LoopV2, item interface{}) (string, error) {
	// Start with loop.With parameters
	inputMap := make(map[string]interface{})
	
	// Add all 'with' parameters
	for key, value := range loop.With {
		inputMap[key] = value
	}
	
	// Add current item (overrides any 'item' in With)
	inputMap["item"] = item
	
	// Convert to JSON for interpolation
	jsonBytes, err := json.Marshal(inputMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}
	
	// Interpolate
	interpolated, err := le.interpolator.Interpolate(string(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to interpolate input: %w", err)
	}
	
	return interpolated, nil
}

// extractItemID extracts an identifier from an item for logging
func (le *LoopExecutor) extractItemID(item interface{}, index int) string {
	// Try to extract 'id' field from item
	if itemMap, ok := item.(map[string]interface{}); ok {
		if id, ok := itemMap["id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		if id, ok := itemMap["control_id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		if id, ok := itemMap["name"]; ok {
			return fmt.Sprintf("%v", id)
		}
	}
	
	return fmt.Sprintf("ITEM-%03d", index)
}

// storeIterateLoopResult stores iterate loop result for later access
func (le *LoopExecutor) storeIterateLoopResult(loop *config.LoopV2, result *config.LoopExecutionResult) {
	// Store as loop.output
	le.interpolator.SetStepResult("loop.output", result.FinalOutput)
	le.interpolator.SetStepResult("loop.iterations", fmt.Sprintf("%d", result.Iterations))
	le.interpolator.SetStepResult("loop.succeeded", fmt.Sprintf("%d", result.Succeeded))
	le.interpolator.SetStepResult("loop.failed", fmt.Sprintf("%d", result.Failed))
	
	// Store with custom name if specified
	if loop.Accumulate != "" {
		history := strings.Join(result.AllOutputs, "\n---\n")
		le.interpolator.SetStepResult(loop.Accumulate, history)
	}
	
	// Store loop name result
	le.interpolator.SetStepResult(loop.Name, result.FinalOutput)
}

// parseArrayInput parses array input from various formats
func (le *LoopExecutor) parseArrayInput(data string) ([]interface{}, error) {
	data = strings.TrimSpace(data)
	
	if data == "" {
		return nil, fmt.Errorf("empty array input")
	}
	
	// Try JSON array format first (starts with '[')
	if strings.HasPrefix(data, "[") {
		if items, err := le.parseJSONArray(data); err == nil && len(items) > 0 {
			le.logger.Debug("Parsed %d items from JSON array format", len(items))
			return items, nil
		}
	}
	
	// Try JSONL format (one JSON object per line)
	if items, err := le.parseJSONL(data); err == nil && len(items) > 0 {
		le.logger.Debug("Parsed %d items from JSONL format", len(items))
		return items, nil
	}
	
	// Fallback to text lines
	items := le.parseTextLines(data)
	le.logger.Debug("Parsed %d items from text lines format", len(items))
	return items, nil
}

// parseJSONL parses JSONL format (one JSON object per line)
func (le *LoopExecutor) parseJSONL(data string) ([]interface{}, error) {
	lines := strings.Split(data, "\n")
	items := make([]interface{}, 0, len(lines))
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var item interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("line %d is not valid JSON: %w", i+1, err)
		}
		items = append(items, item)
	}
	
	if len(items) == 0 {
		return nil, fmt.Errorf("no valid JSONL items found")
	}
	
	return items, nil
}

// parseJSONArray parses JSON array format
func (le *LoopExecutor) parseJSONArray(data string) ([]interface{}, error) {
	var items []interface{}
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, fmt.Errorf("not a valid JSON array: %w", err)
	}
	
	if len(items) == 0 {
		return nil, fmt.Errorf("empty JSON array")
	}
	
	return items, nil
}

// parseTextLines parses plain text lines as items
func (le *LoopExecutor) parseTextLines(data string) []interface{} {
	lines := strings.Split(data, "\n")
	items := make([]interface{}, 0, len(lines))
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	
	return items
}
