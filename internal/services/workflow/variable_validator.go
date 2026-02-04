package workflow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// VariableValidator validates variable references in workflow steps
type VariableValidator struct {
	workflow *config.WorkflowV2
	stepMap  map[string]bool
	loopMap  map[string]bool
}

// NewVariableValidator creates a new variable validator
func NewVariableValidator(workflow *config.WorkflowV2) *VariableValidator {
	stepMap := make(map[string]bool)
	for i := range workflow.Steps {
		stepMap[workflow.Steps[i].Name] = true
	}

	loopMap := make(map[string]bool)
	for i := range workflow.Loops {
		loopMap[workflow.Loops[i].Name] = true
	}

	return &VariableValidator{
		workflow: workflow,
		stepMap:  stepMap,
		loopMap:  loopMap,
	}
}

// ValidateAll validates all steps in the workflow
func (v *VariableValidator) ValidateAll() []error {
	var errors []error

	for i := range v.workflow.Steps {
		step := &v.workflow.Steps[i]
		if errs := v.ValidateStep(step); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	return errors
}

// ValidateStep validates variable references in a single step
func (v *VariableValidator) ValidateStep(step *config.StepV2) []error {
	var errors []error

	// Extract text to validate from various step modes
	textsToValidate := v.extractTextsFromStep(step)

	for _, text := range textsToValidate {
		refs := v.extractVariableReferences(text)

		for _, ref := range refs {
			// Skip built-in variables
			if v.isBuiltInVariable(ref) {
				continue
			}

			// Check if reference exists as a step or loop
			if !v.stepMap[ref] && !v.loopMap[ref] {
				errors = append(errors, fmt.Errorf(
					"step '%s' references non-existent variable '{{%s}}'",
					step.Name, ref,
				))
				continue
			}

			// Check if reference is in needs array
			if !v.isInNeeds(step, ref) {
				errors = append(errors, fmt.Errorf(
					"step '%s' references '{{%s}}' but '%s' is not in needs: array (add 'needs: [%s]' to ensure correct execution order)",
					step.Name, ref, ref, ref,
				))
			}
		}
	}

	return errors
}

// extractTextsFromStep extracts all text fields that might contain variable references
func (v *VariableValidator) extractTextsFromStep(step *config.StepV2) []string {
	var texts []string

	// Run mode
	if step.Run != "" {
		texts = append(texts, step.Run)
	}

	// Consensus mode
	if step.Consensus != nil && step.Consensus.Prompt != "" {
		texts = append(texts, step.Consensus.Prompt)
	}

	// Template mode (with parameters)
	if step.Template != nil && step.Template.With != nil {
		for _, value := range step.Template.With {
			if str, ok := value.(string); ok {
				texts = append(texts, str)
			}
		}
	}

	// Loop mode (with parameters)
	if step.Loop != nil && step.Loop.With != nil {
		for _, value := range step.Loop.With {
			if str, ok := value.(string); ok {
				texts = append(texts, str)
			}
		}
	}

	return texts
}

// extractVariableReferences extracts all {{variable}} references from text
func (v *VariableValidator) extractVariableReferences(text string) []string {
	// Match {{variable_name}} pattern
	re := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_\.]*)\}\}`)
	matches := re.FindAllStringSubmatch(text, -1)

	var refs []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			ref := match[1]

			// Extract base variable name (before any dots)
			base := strings.Split(ref, ".")[0]

			if !seen[base] {
				refs = append(refs, base)
				seen[base] = true
			}
		}
	}

	return refs
}

// isBuiltInVariable checks if a variable is a built-in variable
func (v *VariableValidator) isBuiltInVariable(name string) bool {
	builtIns := map[string]bool{
		"input":     true,
		"loop":      true,
		"env":       true,
		"iteration": true,
		"item":      true,
		"index":     true,
		"consensus": true,
	}

	return builtIns[name]
}

// isInNeeds checks if a variable reference is in the step's needs array
func (v *VariableValidator) isInNeeds(step *config.StepV2, ref string) bool {
	for _, need := range step.Needs {
		if need == ref {
			return true
		}
	}
	return false
}
