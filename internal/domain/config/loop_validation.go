package config

import (
	"fmt"
	"time"
)

// LoopIterationResult represents the result of a single loop iteration
type LoopIterationResult struct {
	Index    int           // Zero-based index
	ItemID   string        // Item identifier if available
	Status   string        // "succeeded", "failed", "skipped"
	Attempt  int           // Current retry attempt
	Duration time.Duration // Time taken
	Error    string        // Error message if failed
}

// LoopExecutionResult stores detailed results from loop execution
type LoopExecutionResult struct {
	// Iteration tracking
	TotalItems  int
	Succeeded   int
	Failed      int
	Skipped     int
	FailedItems []int // Indices of failed items
	
	// Timing
	Duration time.Duration
	
	// Success determination
	Success bool // Based on min_success_rate or completion
	
	// Legacy fields (for backward compatibility)
	Iterations  int
	FinalOutput string
	AllOutputs  []string
	ExitReason  string // "condition_met", "max_iterations", "failure", "success_rate_not_met"
}

// CheckSuccessRate validates if loop met minimum success rate
func (r *LoopExecutionResult) CheckSuccessRate(minRate float64) bool {
	if r.TotalItems == 0 {
		return true
	}
	actualRate := float64(r.Succeeded) / float64(r.TotalItems)
	return actualRate >= minRate
}

// Validate validates the LoopV2 configuration
func (l *LoopV2) Validate() error {
	// Mode defaults to "refine"
	if l.Mode == "" {
		l.Mode = "refine"
	}
	
	// Validate mode
	if l.Mode != "iterate" && l.Mode != "refine" {
		return fmt.Errorf("loop mode must be 'iterate' or 'refine', got '%s'", l.Mode)
	}
	
	// Mode-specific requirements
	if l.Mode == "iterate" {
		if l.Items == "" {
			return fmt.Errorf("iterate mode requires 'items' field")
		}
	} else if l.Mode == "refine" {
		if l.Until == "" {
			return fmt.Errorf("refine mode requires 'until' condition")
		}
	}
	
	// Workflow is required
	if l.Workflow == "" {
		return fmt.Errorf("workflow field is required")
	}
	
	// Max iterations must be positive
	if l.MaxIterations < 1 {
		l.MaxIterations = 100 // Default
	}
	
	// Validate success rate bounds
	if l.MinSuccessRate < 0 || l.MinSuccessRate > 1 {
		return fmt.Errorf("min_success_rate must be between 0.0 and 1.0, got %f", l.MinSuccessRate)
	}
	
	// Retry configuration validation
	if l.OnFailure == "retry" && l.MaxRetries < 1 {
		return fmt.Errorf("retry mode requires max_retries >= 1")
	}
	
	// Validate on_failure values
	if l.OnFailure != "" && l.OnFailure != "halt" && l.OnFailure != "continue" && l.OnFailure != "retry" {
		return fmt.Errorf("on_failure must be 'halt', 'continue', or 'retry', got '%s'", l.OnFailure)
	}
	
	return nil
}

// Validate validates the LoopMode configuration
func (l *LoopMode) Validate() error {
	// Mode defaults to "refine"
	if l.Mode == "" {
		l.Mode = "refine"
	}
	
	// Validate mode
	if l.Mode != "iterate" && l.Mode != "refine" {
		return fmt.Errorf("loop mode must be 'iterate' or 'refine', got '%s'", l.Mode)
	}
	
	// Mode-specific requirements
	if l.Mode == "iterate" {
		if l.Items == "" {
			return fmt.Errorf("iterate mode requires 'items' field")
		}
	} else if l.Mode == "refine" {
		if l.Until == "" {
			return fmt.Errorf("refine mode requires 'until' condition")
		}
	}
	
	// Workflow is required
	if l.Workflow == "" {
		return fmt.Errorf("workflow field is required")
	}
	
	// Max iterations must be positive
	if l.MaxIterations < 1 {
		l.MaxIterations = 100 // Default
	}
	
	// Validate success rate bounds
	if l.MinSuccessRate < 0 || l.MinSuccessRate > 1 {
		return fmt.Errorf("min_success_rate must be between 0.0 and 1.0, got %f", l.MinSuccessRate)
	}
	
	// Retry configuration validation
	if l.OnFailure == "retry" && l.MaxRetries < 1 {
		return fmt.Errorf("retry mode requires max_retries >= 1")
	}
	
	// Validate on_failure values
	if l.OnFailure != "" && l.OnFailure != "halt" && l.OnFailure != "continue" && l.OnFailure != "retry" {
		return fmt.Errorf("on_failure must be 'halt', 'continue', or 'retry', got '%s'", l.OnFailure)
	}
	
	return nil
}
