package workflow

import (
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

func TestDependencyResolver_SimpleChain(t *testing.T) {
	// A → B → C
	steps := []*config.StepV2{
		{Name: "A", Needs: []string{}},
		{Name: "B", Needs: []string{"A"}},
		{Name: "C", Needs: []string{"B"}},
	}

	resolver := NewDependencyResolver(steps)

	// Initial: only A should be ready
	ready := resolver.GetReadySteps(map[string]bool{})
	if len(ready) != 1 || ready[0].Name != "A" {
		t.Errorf("Expected [A], got %v", getStepNames(ready))
	}

	// After A: only B should be ready
	ready = resolver.GetReadySteps(map[string]bool{"A": true})
	if len(ready) != 1 || ready[0].Name != "B" {
		t.Errorf("Expected [B], got %v", getStepNames(ready))
	}

	// After B: only C should be ready
	ready = resolver.GetReadySteps(map[string]bool{"A": true, "B": true})
	if len(ready) != 1 || ready[0].Name != "C" {
		t.Errorf("Expected [C], got %v", getStepNames(ready))
	}

	// All completed: nothing ready
	ready = resolver.GetReadySteps(map[string]bool{"A": true, "B": true, "C": true})
	if len(ready) != 0 {
		t.Errorf("Expected [], got %v", getStepNames(ready))
	}
}

func TestDependencyResolver_ParallelBranches(t *testing.T) {
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D

	steps := []*config.StepV2{
		{Name: "A", Needs: []string{}},
		{Name: "B", Needs: []string{"A"}},
		{Name: "C", Needs: []string{"A"}},
		{Name: "D", Needs: []string{"B", "C"}},
	}

	resolver := NewDependencyResolver(steps)

	// Initial: only A
	ready := resolver.GetReadySteps(map[string]bool{})
	if len(ready) != 1 || ready[0].Name != "A" {
		t.Errorf("Expected [A], got %v", getStepNames(ready))
	}

	// After A: B and C should both be ready (parallel!)
	ready = resolver.GetReadySteps(map[string]bool{"A": true})
	names := getStepNames(ready)
	if len(ready) != 2 || !contains(names, "B") || !contains(names, "C") {
		t.Errorf("Expected [B, C], got %v", names)
	}

	// After B only: D should NOT be ready
	ready = resolver.GetReadySteps(map[string]bool{"A": true, "B": true})
	if len(ready) != 1 || ready[0].Name != "C" {
		t.Errorf("Expected [C], got %v", getStepNames(ready))
	}

	// After both B and C: D should be ready
	ready = resolver.GetReadySteps(map[string]bool{"A": true, "B": true, "C": true})
	if len(ready) != 1 || ready[0].Name != "D" {
		t.Errorf("Expected [D], got %v", getStepNames(ready))
	}
}

func TestDependencyResolver_NoDependencies(t *testing.T) {
	// All steps independent
	steps := []*config.StepV2{
		{Name: "A", Needs: []string{}},
		{Name: "B", Needs: []string{}},
		{Name: "C", Needs: []string{}},
	}

	resolver := NewDependencyResolver(steps)

	// All should be ready initially
	ready := resolver.GetReadySteps(map[string]bool{})
	names := getStepNames(ready)
	if len(ready) != 3 || !contains(names, "A") || !contains(names, "B") || !contains(names, "C") {
		t.Errorf("Expected [A, B, C], got %v", names)
	}
}

func TestDependencyResolver_ValidateNoCycles(t *testing.T) {
	tests := []struct {
		name        string
		steps       []*config.StepV2
		expectError bool
	}{
		{
			name: "No cycles - simple chain",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{}},
				{Name: "B", Needs: []string{"A"}},
			},
			expectError: false,
		},
		{
			name: "Self-cycle",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{"A"}},
			},
			expectError: true,
		},
		{
			name: "Two-step cycle",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{"B"}},
				{Name: "B", Needs: []string{"A"}},
			},
			expectError: true,
		},
		{
			name: "Three-step cycle",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{"B"}},
				{Name: "B", Needs: []string{"C"}},
				{Name: "C", Needs: []string{"A"}},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDependencyResolver(tt.steps)
			err := resolver.ValidateNoCycles()

			if tt.expectError && err == nil {
				t.Error("Expected error for cycle detection, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestDependencyResolver_ValidateDependenciesExist(t *testing.T) {
	tests := []struct {
		name        string
		steps       []*config.StepV2
		expectError bool
	}{
		{
			name: "All dependencies exist",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{}},
				{Name: "B", Needs: []string{"A"}},
			},
			expectError: false,
		},
		{
			name: "Dependency does not exist",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{"NonExistent"}},
			},
			expectError: true,
		},
		{
			name: "Multiple dependencies, one missing",
			steps: []*config.StepV2{
				{Name: "A", Needs: []string{}},
				{Name: "B", Needs: []string{"A", "C"}},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDependencyResolver(tt.steps)
			err := resolver.ValidateDependenciesExist()

			if tt.expectError && err == nil {
				t.Error("Expected error for missing dependency, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestDependencyResolver_GetExecutionOrder(t *testing.T) {
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D

	steps := []*config.StepV2{
		{Name: "D", Needs: []string{"B", "C"}},
		{Name: "B", Needs: []string{"A"}},
		{Name: "C", Needs: []string{"A"}},
		{Name: "A", Needs: []string{}},
	}

	resolver := NewDependencyResolver(steps)
	order, err := resolver.GetExecutionOrder()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(order) != 4 {
		t.Fatalf("Expected 4 steps, got %d", len(order))
	}

	// A must come first
	if order[0].Name != "A" {
		t.Errorf("Expected A first, got %s", order[0].Name)
	}

	// D must come last
	if order[3].Name != "D" {
		t.Errorf("Expected D last, got %s", order[3].Name)
	}

	// B and C can be in either order (index 1 or 2)
	middleNames := []string{order[1].Name, order[2].Name}
	if !contains(middleNames, "B") || !contains(middleNames, "C") {
		t.Errorf("Expected B and C in middle positions, got %v", middleNames)
	}
}

// Helper functions

func getStepNames(steps []*config.StepV2) []string {
	names := make([]string, len(steps))
	for i, step := range steps {
		names[i] = step.Name
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
