package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// WorkflowTemplate represents a multi-step template with chaining capabilities
type WorkflowTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []WorkflowStep         `json:"steps"`
	Variables   map[string]string      `json:"variables,omitempty"`
	Settings    *WorkflowSettings      `json:"settings,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStep represents a single step in a workflow template
type WorkflowStep struct {
	Step            int                    `json:"step"`
	Name            string                 `json:"name"`
	BasePrompt      string                 `json:"base_prompt"`
	SystemPrompt    string                 `json:"system_prompt,omitempty"`
	Provider        string                 `json:"provider,omitempty"`
	Model           string                 `json:"model,omitempty"`
	Servers         []string               `json:"servers,omitempty"`
	ToolsRequired   []string               `json:"tools_required,omitempty"`
	InputHandling   *WorkflowInputConfig   `json:"input_handling,omitempty"`
	Output          *WorkflowOutputConfig  `json:"output,omitempty"`
	Conditions      *WorkflowConditions    `json:"conditions,omitempty"`
	Timeout         string                 `json:"timeout,omitempty"`
	RetryPolicy     *WorkflowRetryPolicy   `json:"retry_policy,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	Embedding       *EmbeddingConfig       `json:"embedding,omitempty"`     // Embedding configuration for RAG
}

// WorkflowInputConfig handles input processing for workflow steps
type WorkflowInputConfig struct {
	StdinAppend     bool   `json:"stdin_append"`
	StdinPrefix     string `json:"stdin_prefix"`
	MaxInputSize    string `json:"max_input_size"`
	RequireInput    bool   `json:"require_input"`
	InputValidation bool   `json:"input_validation"`
}

// WorkflowOutputConfig handles output formatting for workflow steps
type WorkflowOutputConfig struct {
	Format          string `json:"format"`           // json, text, structured
	IncludeMetadata bool   `json:"include_metadata"`
	ErrorHandling   string `json:"error_handling"`   // strict, continue, retry
	SaveToVariable  string `json:"save_to_variable,omitempty"`
	PassToNext      bool   `json:"pass_to_next"`     // Pass output to next step
}

// WorkflowConditions define when a step should execute
type WorkflowConditions struct {
	SkipIf       string   `json:"skip_if,omitempty"`        // Condition to skip step
	OnlyIf       string   `json:"only_if,omitempty"`        // Condition to execute step
	RequiredVars []string `json:"required_vars,omitempty"`  // Required variables to proceed
}

// WorkflowRetryPolicy defines retry behavior for failed steps
type WorkflowRetryPolicy struct {
	MaxRetries    int    `json:"max_retries"`
	RetryDelay    string `json:"retry_delay"`    // e.g., "5s", "1m"
	BackoffFactor float64 `json:"backoff_factor"`
}

// WorkflowSettings contains global workflow settings
type WorkflowSettings struct {
	MaxExecutionTime string `json:"max_execution_time"`
	FailOnStepError  bool   `json:"fail_on_step_error"`
	LogLevel         string `json:"log_level"`
	ConcurrentSteps  bool   `json:"concurrent_steps"`
}

// WorkflowProcessor defines the interface for processing workflow templates
type WorkflowProcessor interface {
	// ProcessWorkflow executes a complete workflow template
	ProcessWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowResponse, error)
	
	// ProcessStep executes a single workflow step
	ProcessStep(ctx context.Context, step *WorkflowStep, variables map[string]interface{}) (*WorkflowStepResult, error)
	
	// ValidateWorkflow validates a workflow template
	ValidateWorkflow(template *WorkflowTemplate) error
}

// WorkflowRequest represents a request to execute a workflow
type WorkflowRequest struct {
	TemplateName    string                 `json:"template_name"`
	InputData       string                 `json:"input_data"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	StartFromStep   int                    `json:"start_from_step,omitempty"`
	StopAtStep      int                    `json:"stop_at_step,omitempty"`
	ExecutionID     string                 `json:"execution_id"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowResponse represents the result of workflow execution
type WorkflowResponse struct {
	ExecutionID     string                   `json:"execution_id"`
	TemplateName    string                   `json:"template_name"`
	Status          WorkflowStatus           `json:"status"`
	FinalOutput     string                   `json:"final_output"`
	StepResults     []WorkflowStepResult     `json:"step_results"`
	Variables       map[string]interface{}   `json:"variables"`
	ExecutionTime   time.Duration            `json:"execution_time"`
	Timestamp       time.Time                `json:"timestamp"`
	Error           *WorkflowError           `json:"error,omitempty"`
	Metadata        map[string]interface{}   `json:"metadata,omitempty"`
}

// WorkflowStepResult represents the result of a single workflow step
type WorkflowStepResult struct {
	Step            int                    `json:"step"`
	Name            string                 `json:"name"`
	Status          WorkflowStepStatus     `json:"status"`
	Output          string                 `json:"output"`
	Provider        string                 `json:"provider"`
	Model           string                 `json:"model"`
	ToolCalls       []ToolCall            `json:"tool_calls,omitempty"`
	Usage           *Usage                `json:"usage,omitempty"`
	ExecutionTime   time.Duration         `json:"execution_time"`
	Error           *WorkflowError        `json:"error,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
}

// WorkflowError represents an error in workflow execution
type WorkflowError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Step        int                    `json:"step,omitempty"`
	StepName    string                 `json:"step_name,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Retryable   bool                   `json:"retryable"`
}

// WorkflowStatus represents the overall status of workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusPartial   WorkflowStatus = "partial" // Some steps completed
)

// WorkflowStepStatus represents the status of a single workflow step
type WorkflowStepStatus string

const (
	StepStatusPending   WorkflowStepStatus = "pending"
	StepStatusRunning   WorkflowStepStatus = "running"
	StepStatusCompleted WorkflowStepStatus = "completed"
	StepStatusFailed    WorkflowStepStatus = "failed"
	StepStatusSkipped   WorkflowStepStatus = "skipped"
	StepStatusRetrying  WorkflowStepStatus = "retrying"
)

// ValidateWorkflowTemplate validates a workflow template
func (wt *WorkflowTemplate) ValidateWorkflowTemplate() error {
	if wt.Name == "" {
		return NewDomainError(ErrCodeRequestInvalid, "workflow template name is required")
	}
	
	if len(wt.Steps) == 0 {
		return NewDomainError(ErrCodeRequestInvalid, "workflow template must have at least one step")
	}
	
	// Validate step sequence
	for i, step := range wt.Steps {
		if step.Step != i+1 {
			return NewDomainError(ErrCodeRequestInvalid, 
				fmt.Sprintf("step %d has incorrect step number %d", i+1, step.Step))
		}
		
		if step.Name == "" {
			return NewDomainError(ErrCodeRequestInvalid, 
				fmt.Sprintf("step %d is missing a name", step.Step))
		}
		
		if step.BasePrompt == "" {
			return NewDomainError(ErrCodeRequestInvalid, 
				fmt.Sprintf("step %d (%s) is missing base_prompt", step.Step, step.Name))
		}
	}
	
	return nil
}

// ProcessVariables processes variable substitution in a prompt
func (wt *WorkflowTemplate) ProcessVariables(prompt string, variables map[string]interface{}) string {
	result := prompt
	
	// Process template variables first
	for key, value := range wt.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Process runtime variables
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	
	return result
}

// GetTimeout returns the timeout duration for a workflow step
func (ws *WorkflowStep) GetTimeout() (time.Duration, error) {
	if ws.Timeout == "" {
		return 5 * time.Minute, nil // Default timeout
	}
	
	duration, err := time.ParseDuration(ws.Timeout)
	if err != nil {
		return 0, NewDomainError(ErrCodeRequestInvalid, 
			fmt.Sprintf("invalid timeout format '%s'", ws.Timeout))
	}
	
	return duration, nil
}

// GetRetryDelay returns the retry delay duration
func (wrp *WorkflowRetryPolicy) GetRetryDelay() (time.Duration, error) {
	if wrp.RetryDelay == "" {
		return 1 * time.Second, nil // Default delay
	}
	
	duration, err := time.ParseDuration(wrp.RetryDelay)
	if err != nil {
		return 0, NewDomainError(ErrCodeRequestInvalid, 
			fmt.Sprintf("invalid retry delay format '%s'", wrp.RetryDelay))
	}
	
	return duration, nil
}

// ShouldExecute determines if a step should execute based on conditions
func (ws *WorkflowStep) ShouldExecute(variables map[string]interface{}) bool {
	if ws.Conditions == nil {
		return true
	}
	
	// Check required variables
	if len(ws.Conditions.RequiredVars) > 0 {
		for _, required := range ws.Conditions.RequiredVars {
			if _, exists := variables[required]; !exists {
				return false
			}
		}
	}
	
	// Simple condition evaluation (could be enhanced with expression evaluation)
	if ws.Conditions.SkipIf != "" {
		// For now, just check if the condition variable exists and is truthy
		if val, exists := variables[ws.Conditions.SkipIf]; exists {
			if boolVal, ok := val.(bool); ok && boolVal {
				return false
			}
			if strVal, ok := val.(string); ok && strVal != "" {
				return false
			}
		}
	}
	
	if ws.Conditions.OnlyIf != "" {
		// Only execute if the condition variable exists and is truthy
		if val, exists := variables[ws.Conditions.OnlyIf]; exists {
			if boolVal, ok := val.(bool); ok {
				return boolVal
			}
			if strVal, ok := val.(string); ok {
				return strVal != ""
			}
		} else {
			return false
		}
	}
	
	return true
}
