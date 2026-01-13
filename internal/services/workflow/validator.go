package workflow

import (
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// ValidationError represents a workflow validation error
type ValidationError struct {
	Step    string
	Field   string
	Message string
	Hint    string
}

func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("Step '%s': %s", e.Step, e.Message)
	if e.Hint != "" {
		msg += fmt.Sprintf("\n  Hint: %s", e.Hint)
	}
	return msg
}

// WorkflowValidator validates workflow structure before execution
type WorkflowValidator struct {
	workflow *config.WorkflowV2
	errors   []ValidationError
}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator(workflow *config.WorkflowV2) *WorkflowValidator {
	return &WorkflowValidator{
		workflow: workflow,
		errors:   make([]ValidationError, 0),
	}
}

// Validate performs comprehensive validation and returns all errors
func (v *WorkflowValidator) Validate() error {
	// Validate each step
	for i := range v.workflow.Steps {
		v.validateStep(&v.workflow.Steps[i])
	}

	// Return errors if any
	if len(v.errors) > 0 {
		return v.formatErrors()
	}

	return nil
}

// validateStep validates a single step's structure
func (v *WorkflowValidator) validateStep(step *config.StepV2) {
	// Check that step has an execution mode
	executionModes := v.countExecutionModes(step)
	
	if executionModes == 0 {
		v.addError(step.Name, "", "no execution mode specified",
			"Steps must have ONE of: run, template, rag, embeddings, consensus, or loop")
	} else if executionModes > 1 {
		v.addError(step.Name, "", "multiple execution modes specified",
			"Steps can only have ONE execution mode (run, template, rag, embeddings, consensus, or loop)")
	}

	// Validate template mode
	if step.Template != nil {
		v.validateTemplateMode(step)
	}

	// Validate loop mode
	if step.Loop != nil {
		v.validateLoopMode(step)
	}

	// Validate consensus mode
	if step.Consensus != nil {
		v.validateConsensusMode(step)
	}

	// Validate rag mode
	if step.Rag != nil {
		v.validateRagMode(step)
	}

	// Validate dependencies
	v.validateDependencies(step)
}

// countExecutionModes counts how many execution modes are set
func (v *WorkflowValidator) countExecutionModes(step *config.StepV2) int {
	count := 0
	if step.Run != "" {
		count++
	}
	if step.Template != nil {
		count++
	}
	if step.Loop != nil {
		count++
	}
	if step.Embeddings != nil {
		count++
	}
	if step.Consensus != nil {
		count++
	}
	if step.Rag != nil {
		count++
	}
	return count
}

// validateTemplateMode validates template execution mode
func (v *WorkflowValidator) validateTemplateMode(step *config.StepV2) {
	if step.Template.Name == "" {
		v.addError(step.Name, "template.name", "template name is required",
			"Example: template:\n  name: my_workflow\n  with:\n    param: value")
	}
}

// validateLoopMode validates loop execution mode
func (v *WorkflowValidator) validateLoopMode(step *config.StepV2) {
	if step.Loop.Workflow == "" {
		v.addError(step.Name, "loop.workflow", "loop workflow name is required",
			"Example: loop:\n  workflow: child_workflow\n  max_iterations: 5")
	}
	
	if step.Loop.MaxIterations <= 0 {
		v.addError(step.Name, "loop.max_iterations", "max_iterations must be > 0",
			"Set a reasonable limit like max_iterations: 10")
	}

	// Validate parallel settings
	if step.Loop.Parallel && step.Loop.MaxWorkers <= 0 {
		v.addError(step.Name, "loop.max_workers", "max_workers must be > 0 when parallel is true",
			"Set max_workers to control concurrency (e.g., max_workers: 3)")
	}
	
	// Validate variable syntax in items
	if step.Loop.Items != "" {
		v.validateVariableSyntax(step, "loop.items", step.Loop.Items)
		v.validateLoopVariables(step)
	}
}

// validateConsensusMode validates consensus execution mode
func (v *WorkflowValidator) validateConsensusMode(step *config.StepV2) {
	if step.Consensus.Prompt == "" {
		v.addError(step.Name, "consensus.prompt", "consensus prompt is required",
			"Example: consensus:\n  prompt: \"Is this valid?\"\n  executions: [...]")
	}

	if len(step.Consensus.Executions) < 2 {
		v.addError(step.Name, "consensus.executions", "at least 2 executions required for consensus",
			"Add multiple provider/model combinations to get consensus")
	}
}

// validateRagMode validates RAG execution mode
func (v *WorkflowValidator) validateRagMode(step *config.StepV2) {
	if step.Rag.Server == "" {
		v.addError(step.Name, "rag.server", "RAG server name is required",
			"Example: rag:\n  server: pgvector\n  query: \"search terms\"")
	}

	if step.Rag.Query == "" {
		v.addError(step.Name, "rag.query", "RAG query is required",
			"Specify the search query for RAG retrieval")
	}
	
	// Validate variable syntax in query
	v.validateVariableSyntax(step, "rag.query", step.Rag.Query)
	v.validateRagVariables(step)
}

// validateDependencies validates step dependencies exist
func (v *WorkflowValidator) validateDependencies(step *config.StepV2) {
	if len(step.Needs) == 0 {
		return
	}

	// Build map of all step names
	stepNames := make(map[string]bool)
	for _, s := range v.workflow.Steps {
		stepNames[s.Name] = true
	}

	// Check each dependency exists
	for _, dep := range step.Needs {
		if !stepNames[dep] {
			v.addError(step.Name, "needs", fmt.Sprintf("dependency '%s' does not exist", dep),
				"Check that the step name matches exactly (case-sensitive)")
		}
	}
}

// addError adds a validation error
func (v *WorkflowValidator) addError(step, field, message, hint string) {
	v.errors = append(v.errors, ValidationError{
		Step:    step,
		Field:   field,
		Message: message,
		Hint:    hint,
	})
}

// formatErrors formats all errors into a single error message
func (v *WorkflowValidator) formatErrors() error {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Workflow validation failed with %d error(s):\n\n", len(v.errors)))
	
	for i, err := range v.errors {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Error()))
		if i < len(v.errors)-1 {
			sb.WriteString("\n")
		}
	}
	
	sb.WriteString("\n═══════════════════════════════════════════════════════════\n")
	sb.WriteString("Valid step execution modes:\n")
	sb.WriteString("  • run: \"LLM query with {{variables}}\"\n")
	sb.WriteString("  • template:\n")
	sb.WriteString("      name: workflow_name\n")
	sb.WriteString("      with:\n")
	sb.WriteString("        param: value\n")
	sb.WriteString("  • rag:\n")
	sb.WriteString("      server: pgvector\n")
	sb.WriteString("      query: \"search query\"\n")
	sb.WriteString("  • loop:\n")
	sb.WriteString("      workflow: child_workflow\n")
	sb.WriteString("      max_iterations: 10\n")
	sb.WriteString("  • embeddings: {...}\n")
	sb.WriteString("  • consensus: {...}\n")
	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	
	return fmt.Errorf("%s", sb.String())
}

// ValidateWorkflow is a convenience function to validate a workflow
func ValidateWorkflow(workflow *config.WorkflowV2) error {
	validator := NewWorkflowValidator(workflow)
	return validator.Validate()
}
