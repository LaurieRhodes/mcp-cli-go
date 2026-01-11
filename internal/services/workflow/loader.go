package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing workflow YAML files
type Loader struct{}

// NewLoader creates a new workflow loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromFile loads a workflow from a YAML file
func (l *Loader) LoadFromFile(path string) (*config.WorkflowV2, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return l.LoadFromBytes(data)
}

// LoadFromBytes loads a workflow from YAML bytes
func (l *Loader) LoadFromBytes(data []byte) (*config.WorkflowV2, error) {
	var workflow config.WorkflowV2

	// Parse YAML with strict mode (errors on unknown fields)
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true) // Enable strict mode
	
	if err := decoder.Decode(&workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Validate
	if err := l.Validate(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return &workflow, nil
}

// Validate validates a workflow
func (l *Loader) Validate(workflow *config.WorkflowV2) error {
	// Check required fields
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if workflow.Version == "" {
		return fmt.Errorf("workflow version is required")
	}

	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Validate execution context
	if err := l.validateExecution(&workflow.Execution); err != nil {
		return fmt.Errorf("execution context: %w", err)
	}

	// Validate steps
	stepNames := make(map[string]bool)
	for i, step := range workflow.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}

		// Check for duplicate names
		if stepNames[step.Name] {
			return fmt.Errorf("step %d: duplicate name '%s'", i, step.Name)
		}
		stepNames[step.Name] = true

		// Validate step
		if err := l.validateStep(&step, stepNames); err != nil {
			return fmt.Errorf("step %d (%s): %w", i, step.Name, err)
		}
	}

	return nil
}

// validateExecution validates the execution context
func (l *Loader) validateExecution(exec *config.ExecutionContext) error {
	// Must have at least one provider defined
	hasProvider := exec.Provider != "" && exec.Model != ""
	hasProviders := len(exec.Providers) > 0

	if !hasProvider && !hasProviders {
		return fmt.Errorf("must define at least one provider (provider+model or providers array)")
	}

	// Validate providers array
	for i, p := range exec.Providers {
		if p.Provider == "" {
			return fmt.Errorf("providers[%d]: provider is required", i)
		}
		if p.Model == "" {
			return fmt.Errorf("providers[%d]: model is required", i)
		}
	}

	return nil
}

// validateStep validates a single step
func (l *Loader) validateStep(step *config.StepV2, knownSteps map[string]bool) error {
	// Count execution modes
	modeCount := 0
	if step.Run != "" {
		modeCount++
	}
	if step.Embeddings != nil {
		modeCount++
	}
	if step.Template != nil {
		modeCount++
	}
	if step.Consensus != nil {
		modeCount++
	}

	if modeCount == 0 {
		return fmt.Errorf("must specify at least one execution mode (run, embeddings, template, or consensus)")
	}

	if modeCount > 1 {
		return fmt.Errorf("cannot specify multiple execution modes")
	}

	// Validate consensus
	if step.Consensus != nil {
		if err := l.validateConsensus(step.Consensus); err != nil {
			return fmt.Errorf("consensus: %w", err)
		}
	}

	// Validate dependencies
	for _, dep := range step.Needs {
		if !knownSteps[dep] {
			return fmt.Errorf("depends on unknown step: %s", dep)
		}
	}

	return nil
}

// validateConsensus validates consensus configuration
func (l *Loader) validateConsensus(consensus *config.ConsensusMode) error {
	if consensus.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(consensus.Executions) < 2 {
		return fmt.Errorf("requires at least 2 provider executions, got %d", len(consensus.Executions))
	}

	// Validate each execution
	for i, exec := range consensus.Executions {
		if exec.Provider == "" {
			return fmt.Errorf("executions[%d]: provider is required", i)
		}
		if exec.Model == "" {
			return fmt.Errorf("executions[%d]: model is required", i)
		}
	}

	// Validate requirement
	validRequirements := map[string]bool{
		"unanimous": true,
		"2/3":       true,
		"majority":  true,
	}

	if !validRequirements[consensus.Require] {
		return fmt.Errorf("invalid requirement '%s' (must be unanimous, 2/3, or majority)", consensus.Require)
	}

	return nil
}
