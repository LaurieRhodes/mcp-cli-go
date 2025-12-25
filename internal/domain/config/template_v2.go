package config

import (
	"time"
)

// TemplateV2 represents the new YAML template structure with advanced features
type TemplateV2 struct {
	// Required metadata
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`

	// Optional metadata
	Metadata *TemplateMetadata `yaml:"metadata,omitempty"`

	// Configuration
	Config *TemplateConfig `yaml:"config,omitempty"`

	// Includes for step libraries
	Includes []string `yaml:"includes,omitempty"`

	// Reusable step definitions
	StepDefinitions map[string]*StepDefinition `yaml:"step_definitions,omitempty"`

	// Workflow steps (required)
	Steps []WorkflowStepV2 `yaml:"steps"`
}

// TemplateMetadata contains optional template metadata
type TemplateMetadata struct {
	Author   string    `yaml:"author,omitempty"`
	Created  time.Time `yaml:"created,omitempty"`
	Updated  time.Time `yaml:"updated,omitempty"`
	Tags     []string  `yaml:"tags,omitempty"`
	Category string    `yaml:"category,omitempty"`
}

// TemplateConfig contains template-level configuration
type TemplateConfig struct {
	// Template-level variables
	Variables map[string]interface{} `yaml:"variables,omitempty"`

	// Default step configuration
	Defaults *StepDefaults `yaml:"defaults,omitempty"`

	// Global error handling configuration
	ErrorHandling *ErrorHandlingConfig `yaml:"error_handling,omitempty"`
}

// StepDefaults provides default values for steps
type StepDefaults struct {
	Provider    string  `yaml:"provider,omitempty"`
	Model       string  `yaml:"model,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
	Timeout     string  `yaml:"timeout,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

// WorkflowStepV2 represents a step in the new template system
type WorkflowStepV2 struct {
	// Basic identification
	Name        string `yaml:"name"`                  // Required: unique step name
	Description string `yaml:"description,omitempty"` // Optional: documentation

	// Execution control
	Condition string   `yaml:"condition,omitempty"` // Optional: condition to execute
	DependsOn []string `yaml:"depends_on,omitempty"` // Optional: dependencies

	// Basic step fields
	Prompt       string  `yaml:"prompt,omitempty"`        // Main prompt
	SystemPrompt string  `yaml:"system_prompt,omitempty"` // System instructions
	Provider     string  `yaml:"provider,omitempty"`      // LLM provider
	Model        string  `yaml:"model,omitempty"`         // Model name
	Temperature  float64 `yaml:"temperature,omitempty"`   // Temperature setting
	MaxTokens    int     `yaml:"max_tokens,omitempty"`    // Max tokens
	Timeout      string  `yaml:"timeout,omitempty"`       // Timeout duration
	Servers      []string `yaml:"servers,omitempty"`      // MCP servers to use

	// Output configuration - can be string or map
	Output interface{} `yaml:"output,omitempty"` // Named output(s)

	// Step reuse
	Use    string                 `yaml:"use,omitempty"`    // Reference to step definition
	Inputs map[string]interface{} `yaml:"inputs,omitempty"` // Inputs when using step definition

	// Template composition (call another template)
	Template      string                 `yaml:"template,omitempty"`       // Template to call
	TemplateInput string                 `yaml:"template_input,omitempty"` // Input expression for template

	// Parallel execution
	Parallel *ParallelExecution `yaml:"parallel,omitempty"`

	// Loop execution
	ForEach  string `yaml:"for_each,omitempty"`  // Variable to iterate over
	ItemName string `yaml:"item_name,omitempty"` // Name for loop item (default: "item")

	// Transform step
	Transform *TransformConfig `yaml:"transform,omitempty"`

	// Nested steps (for conditionals)
	Steps []WorkflowStepV2 `yaml:"steps,omitempty"`

	// Step-level variables
	Variables map[string]interface{} `yaml:"variables,omitempty"`

	// Compute variables
	Compute map[string]string `yaml:"compute,omitempty"`

	// Error handling
	ErrorHandling *ErrorHandlingConfig `yaml:"error_handling,omitempty"`

	// Validation
	ValidateInput  *ValidationConfig `yaml:"validate_input,omitempty"`
	ValidateOutput *ValidationConfig `yaml:"validate_output,omitempty"`

	// Observability
	Observability *ObservabilityConfig `yaml:"observability,omitempty"`
}

// ParallelExecution defines parallel step execution
type ParallelExecution struct {
	Name          string           `yaml:"name,omitempty"`           // Optional group name
	Steps         []WorkflowStepV2 `yaml:"steps"`                    // Steps to run in parallel
	MaxConcurrent int              `yaml:"max_concurrent,omitempty"` // Limit concurrency
	Aggregate     *AggregateConfig `yaml:"aggregate,omitempty"`      // How to combine results
}

// AggregateConfig defines how to combine parallel results
type AggregateConfig struct {
	Output  string `yaml:"output"`              // Output variable name
	Combine string `yaml:"combine,omitempty"`   // merge | array | custom
}

// TransformConfig defines data transformation operations
type TransformConfig struct {
	Input      string                `yaml:"input"`      // Input variable/expression
	Operations []TransformOperation `yaml:"operations"` // Operations to apply
}

// TransformOperation represents a single transformation
type TransformOperation struct {
	Type      string      `yaml:"type"`                // Operation type
	Condition string      `yaml:"condition,omitempty"` // For filter operations
	Fields    interface{} `yaml:"fields,omitempty"`    // For map operations
	By        string      `yaml:"by,omitempty"`        // For sort/group operations
	Order     string      `yaml:"order,omitempty"`     // asc | desc
	Count     int         `yaml:"count,omitempty"`     // For limit operations
	Key       string      `yaml:"key,omitempty"`       // For groupBy operations
	Function  string      `yaml:"function,omitempty"`  // For custom operations
}

// ErrorHandlingConfig defines error handling behavior
type ErrorHandlingConfig struct {
	OnFailure      string          `yaml:"on_failure,omitempty"`       // stop | continue | retry | fallback
	MaxRetries     int             `yaml:"max_retries,omitempty"`      // Max retry attempts
	RetryBackoff   string          `yaml:"retry_backoff,omitempty"`    // linear | exponential
	InitialDelay   string          `yaml:"initial_delay,omitempty"`    // Initial delay duration
	FallbackStep   string          `yaml:"fallback_step,omitempty"`    // Step to run on failure
	FallbackChain  []FallbackStep  `yaml:"fallback_chain,omitempty"`   // Chain of fallback steps
	DefaultOutput  interface{}     `yaml:"default_output,omitempty"`   // Default output on failure
	Timeout        string          `yaml:"timeout,omitempty"`          // Step timeout
}

// FallbackStep represents a step in fallback chain
type FallbackStep struct {
	Step string `yaml:"step"` // Step name to try
}

// ValidationConfig defines input/output validation
type ValidationConfig struct {
	Schema                map[string]interface{} `yaml:"schema"`                            // JSON schema
	OnValidationFailure   string                 `yaml:"on_validation_failure,omitempty"`  // stop | warn | continue
}

// ObservabilityConfig defines logging and metrics
type ObservabilityConfig struct {
	LogLevel   string            `yaml:"log_level,omitempty"`    // debug | info | warn | error
	LogContext map[string]string `yaml:"log_context,omitempty"`  // Additional context
	Metrics    []string          `yaml:"metrics,omitempty"`      // Metrics to collect
	Tags       map[string]string `yaml:"tags,omitempty"`         // Tags for metrics
	Alerts     []AlertConfig     `yaml:"alerts,omitempty"`       // Alert conditions
}

// AlertConfig defines an alert condition
type AlertConfig struct {
	Condition string `yaml:"condition"` // Condition to trigger alert
	Message   string `yaml:"message"`   // Alert message
	Severity  string `yaml:"severity"`  // Alert severity
}

// StepDefinition represents a reusable step definition
type StepDefinition struct {
	Description  string                  `yaml:"description,omitempty"`
	Inputs       map[string]*InputSchema  `yaml:"inputs,omitempty"`
	Outputs      map[string]*OutputSchema `yaml:"outputs,omitempty"`
	Prompt       string                  `yaml:"prompt"`
	SystemPrompt string                  `yaml:"system_prompt,omitempty"`
	Provider     string                  `yaml:"provider,omitempty"`
	Model        string                  `yaml:"model,omitempty"`
	Temperature  float64                 `yaml:"temperature,omitempty"`
	Servers      []string                `yaml:"servers,omitempty"`
}

// InputSchema defines expected input structure
type InputSchema struct {
	Type        string      `yaml:"type"`                  // Data type
	Required    bool        `yaml:"required,omitempty"`    // Is required
	Default     interface{} `yaml:"default,omitempty"`     // Default value
	Description string      `yaml:"description,omitempty"` // Documentation
	Enum        []string    `yaml:"enum,omitempty"`        // Allowed values
}

// OutputSchema defines expected output structure
type OutputSchema struct {
	Type        string            `yaml:"type"`                  // Data type
	Description string            `yaml:"description,omitempty"` // Documentation
	Enum        []string          `yaml:"enum,omitempty"`        // Allowed values
	Fields      map[string]string `yaml:"fields,omitempty"`      // For object types
}

// StepType represents the type of workflow step
type StepType string

const (
	StepTypeBasic     StepType = "basic"
	StepTypeParallel  StepType = "parallel"
	StepTypeLoop      StepType = "loop"
	StepTypeTransform StepType = "transform"
	StepTypeUse       StepType = "use"
	StepTypeTemplate  StepType = "template"
	StepTypeNested    StepType = "nested"
)

// GetStepType determines the type of a step based on its fields
func (s *WorkflowStepV2) GetStepType() StepType {
	if s.Parallel != nil {
		return StepTypeParallel
	}
	if s.ForEach != "" {
		return StepTypeLoop
	}
	if s.Transform != nil {
		return StepTypeTransform
	}
	if s.Use != "" {
		return StepTypeUse
	}
	if s.Template != "" {
		return StepTypeTemplate
	}
	if len(s.Steps) > 0 {
		return StepTypeNested
	}
	return StepTypeBasic
}

// StepLibrary represents a library of reusable step definitions
type StepLibrary struct {
	Name        string                    `yaml:"name,omitempty"`
	Description string                    `yaml:"description,omitempty"`
	Version     string                    `yaml:"version,omitempty"`
	Steps       map[string]*StepDefinition `yaml:"steps"`
}
