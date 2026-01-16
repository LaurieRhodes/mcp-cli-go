package workflow

import (
	"strings"
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

func TestVariableValidator_ValidStep(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Do step 1"},
			{Name: "step2", Needs: []string{"step1"}, Run: "Use {{step1}} in step 2"},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_MissingNeeds(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Do step 1"},
			{Name: "step2", Run: "Use {{step1}} in step 2"}, // Missing needs: [step1]
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errs), errs)
	}

	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "step2") || !strings.Contains(errMsg, "step1") || !strings.Contains(errMsg, "needs:") {
		t.Errorf("Error message doesn't contain expected information: %s", errMsg)
	}
}

func TestVariableValidator_NonExistentReference(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Use {{nonexistent}} data"},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errs), errs)
	}

	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "non-existent") || !strings.Contains(errMsg, "nonexistent") {
		t.Errorf("Error message doesn't mention non-existent variable: %s", errMsg)
	}
}

func TestVariableValidator_BuiltInVariables(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Use {{input}}, {{env.VAR}}, and {{loop.iteration}}"},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors for built-in variables, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_DottedReferences(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Generate JSON"},
			{Name: "step2", Needs: []string{"step1"}, Run: "Use {{step1.field.nested}} data"},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors for dotted references, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_ConsensusMode(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Generate data"},
			{
				Name:  "step2",
				Needs: []string{"step1"},
				Consensus: &config.ConsensusMode{
					Prompt: "Validate {{step1}}",
				},
			},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_TemplateMode(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Generate data"},
			{
				Name:  "step2",
				Needs: []string{"step1"},
				Template: &config.TemplateMode{
					Name: "other_workflow",
					With: map[string]interface{}{
						"data": "{{step1}}",
					},
				},
			},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_LoopReference(t *testing.T) {
	workflow := &config.WorkflowV2{
		Loops: []config.LoopV2{
			{Name: "my_loop", Workflow: "child", MaxIterations: 5},
		},
		Steps: []config.StepV2{
			{Name: "step1", Needs: []string{"my_loop"}, Run: "Use {{my_loop}} result"},
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	if len(errs) != 0 {
		t.Errorf("Expected no errors for loop reference, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_MultipleErrors(t *testing.T) {
	workflow := &config.WorkflowV2{
		Steps: []config.StepV2{
			{Name: "step1", Run: "Do step 1"},
			{Name: "step2", Run: "Use {{step1}} and {{nonexistent}}"}, // 2 errors: missing needs, nonexistent
			{Name: "step3", Run: "Use {{step1}}"}, // 1 error: missing needs
		},
	}

	validator := NewVariableValidator(workflow)
	errs := validator.ValidateAll()

	// step2 should have 2 errors, step3 should have 1 error
	if len(errs) < 2 {
		t.Errorf("Expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestVariableValidator_ExtractReferences(t *testing.T) {
	validator := &VariableValidator{}

	tests := []struct {
		text     string
		expected []string
	}{
		{"No variables here", []string{}},
		{"Use {{step1}} data", []string{"step1"}},
		{"{{step1}} and {{step2}}", []string{"step1", "step2"}},
		{"{{step1.field.nested}}", []string{"step1"}},
		{"{{step1}} {{step1}}", []string{"step1"}}, // Duplicates removed
		{"{{input}} {{env.VAR}} {{step1}}", []string{"input", "env", "step1"}},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			refs := validator.extractVariableReferences(tt.text)
			if len(refs) != len(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, refs)
				return
			}
			for _, exp := range tt.expected {
				if !contains(refs, exp) {
					t.Errorf("Expected %s in %v", exp, refs)
				}
			}
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
