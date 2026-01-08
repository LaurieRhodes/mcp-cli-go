package config

import "time"

// WorkflowV2 represents the v2.0 workflow schema with property inheritance
type WorkflowV2 struct {
	Schema      string            `yaml:"$schema"`
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Execution   ExecutionContext  `yaml:"execution"`
	Env         map[string]string `yaml:"env,omitempty"`
	Steps []StepV2 `yaml:"steps,omitempty"`
	Loops []LoopV2 `yaml:"loops,omitempty"`
}

// ExecutionContext defines workflow-level defaults for all steps
type ExecutionContext struct {
	// Provider configuration (fallback chain)
	Provider  string           `yaml:"provider,omitempty"`
	Model     string           `yaml:"model,omitempty"`
	Providers []ProviderFallback `yaml:"providers,omitempty"`

	// MCP servers
	Servers []string `yaml:"servers,omitempty"`
	
	// Anthropic Skills
	Skills []string `yaml:"skills,omitempty"`

	// Model parameters
	Temperature float64 `yaml:"temperature,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`

	// Execution control
	Timeout time.Duration `yaml:"timeout,omitempty"`

	// Logging
	Logging string `yaml:"logging,omitempty"` // normal, verbose, noisy
	NoColor bool   `yaml:"no_color,omitempty"`
}

// ProviderFallback represents a provider/model pair for fallback chains
type ProviderFallback struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

// StepV2 represents a workflow step with property inheritance
type StepV2 struct {
	Name string `yaml:"name"`
	ExecutionOrder int `yaml:"execution_order,omitempty"`

	// Core execution
	Run string `yaml:"run,omitempty"` // The prompt
	Loop       *LoopMode       `yaml:"loop,omitempty"`       // Loop execution

	// Provider override (inherits from execution if not specified)
	Provider  string           `yaml:"provider,omitempty"`
	Model     string           `yaml:"model,omitempty"`
	Providers []ProviderFallback `yaml:"providers,omitempty"`

	// Override execution context
	Servers     []string       `yaml:"servers,omitempty"`
	Skills      []string       `yaml:"skills,omitempty"`
	Temperature *float64       `yaml:"temperature,omitempty"` // Pointer to detect override
	MaxTokens   *int           `yaml:"max_tokens,omitempty"`
	Timeout     *time.Duration `yaml:"timeout,omitempty"`
	Logging     string         `yaml:"logging,omitempty"`
	NoColor     *bool          `yaml:"no_color,omitempty"`
	Input       interface{}    `yaml:"input,omitempty"`

	// Special modes
	Embeddings *EmbeddingsMode `yaml:"embeddings,omitempty"`
	Template   *TemplateMode   `yaml:"template,omitempty"`
	Consensus  *ConsensusMode  `yaml:"consensus,omitempty"`

	// Control flow
	If       string   `yaml:"if,omitempty"`
	Needs    []string `yaml:"needs,omitempty"`
}

// LoopV2 represents an iterative execution block
type LoopV2 struct {
	Name          string                 `yaml:"name"`
	Workflow      string                 `yaml:"workflow"`            // Required: workflow to call
	With          map[string]interface{} `yaml:"with,omitempty"`      // Input parameters
	MaxIterations int                    `yaml:"max_iterations"`      // Safety limit
	Until         string                 `yaml:"until"`               // Exit condition (LLM evaluates)
	OnFailure     string                 `yaml:"on_failure,omitempty"` // halt|continue|retry
	Accumulate    string                 `yaml:"accumulate,omitempty"` // Store iteration results
}

// LoopMode defines loop execution within a step
type LoopMode struct {
	Workflow      string                 `yaml:"workflow"`            // Required workflow to call
	With          map[string]interface{} `yaml:"with,omitempty"`      // Input parameters
	MaxIterations int                    `yaml:"max_iterations"`      // Safety limit (required)
	Until         string                 `yaml:"until"`               // Exit condition (LLM evaluates)
	OnFailure     string                 `yaml:"on_failure,omitempty"` // halt|continue|retry
	Accumulate    string                 `yaml:"accumulate,omitempty"` // Store iteration results
}




// EmbeddingsMode represents embeddings generation
type EmbeddingsMode struct {
	// Provider override (inherits from step/execution if not specified)
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`

	// Input source (one required)
	Input     interface{} `yaml:"input,omitempty"`      // string or array
	InputFile string      `yaml:"input_file,omitempty"` // alternative to Input

	// Chunking configuration
	ChunkStrategy  string `yaml:"chunk_strategy,omitempty"`   // sentence, paragraph, fixed
	MaxChunkSize   int    `yaml:"max_chunk_size,omitempty"`   // default: 512
	Overlap        int    `yaml:"overlap,omitempty"`          // overlap between chunks in tokens
	
	// Model configuration
	Dimensions int `yaml:"dimensions,omitempty"` // for supported models

	// Output configuration
	EncodingFormat  string `yaml:"encoding_format,omitempty"`  // float, base64
	IncludeMetadata *bool  `yaml:"include_metadata,omitempty"` // default: true
	OutputFormat    string `yaml:"output_format,omitempty"`    // json, csv, compact
	OutputFile      string `yaml:"output_file,omitempty"`      // output file path
}

// TemplateMode represents template execution
type TemplateMode struct {
	Name string                 `yaml:"name"`
	With map[string]interface{} `yaml:"with,omitempty"`
}

// ConsensusMode represents multi-provider consensus execution
type ConsensusMode struct {
	Prompt       string            `yaml:"prompt"`
	Executions   []ConsensusExec   `yaml:"executions"`
	Require      string            `yaml:"require"` // unanimous, 2/3, majority
	AllowPartial bool              `yaml:"allow_partial,omitempty"`
	Timeout      time.Duration     `yaml:"timeout,omitempty"`
}

// ConsensusExec represents a single provider execution in consensus
type ConsensusExec struct {
	Provider    string         `yaml:"provider"`
	Model       string         `yaml:"model"`
	Temperature *float64       `yaml:"temperature,omitempty"`
	MaxTokens   *int           `yaml:"max_tokens,omitempty"`
	Timeout     *time.Duration `yaml:"timeout,omitempty"`
}

// ConsensusResult represents the result of a consensus execution
type ConsensusResult struct {
	Success    bool              `json:"success"`
	Result     string            `json:"result"`
	Agreement  float64           `json:"agreement"`
	Votes      map[string]string `json:"votes"`
	Confidence string            `json:"confidence"` // high, good, medium, low
}
