package workflow

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// DependencyResolver handles step dependency resolution
type DependencyResolver struct {
	steps         []*config.StepV2
	stepsByName   map[string]*config.StepV2
	dependencies  map[string][]string // step -> dependencies
	dependents    map[string][]string // step -> steps that depend on it
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(steps []*config.StepV2) *DependencyResolver {
	resolver := &DependencyResolver{
		steps:        steps,
		stepsByName:  make(map[string]*config.StepV2),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}

	// Build indices
	for _, step := range steps {
		resolver.stepsByName[step.Name] = step
		resolver.dependencies[step.Name] = step.Needs

		// Build reverse index (dependents)
		for _, dep := range step.Needs {
			resolver.dependents[dep] = append(resolver.dependents[dep], step.Name)
		}
	}

	return resolver
}

// GetReadySteps returns all steps whose dependencies are satisfied
func (r *DependencyResolver) GetReadySteps(completed map[string]bool) []*config.StepV2 {
	var ready []*config.StepV2

	for _, step := range r.steps {
		// Skip if already completed
		if completed[step.Name] {
			continue
		}

		// Check if all dependencies are met
		if r.areDependenciesMet(step.Name, completed) {
			ready = append(ready, step)
		}
	}

	return ready
}

// areDependenciesMet checks if all dependencies for a step are satisfied
func (r *DependencyResolver) areDependenciesMet(stepName string, completed map[string]bool) bool {
	deps := r.dependencies[stepName]
	
	for _, dep := range deps {
		if !completed[dep] {
			return false
		}
	}
	
	return true
}

// ValidateNoCycles checks for circular dependencies
func (r *DependencyResolver) ValidateNoCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, step := range r.steps {
		if !visited[step.Name] {
			if r.hasCycle(step.Name, visited, recStack) {
				return fmt.Errorf("circular dependency detected involving step: %s", step.Name)
			}
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles
func (r *DependencyResolver) hasCycle(stepName string, visited, recStack map[string]bool) bool {
	visited[stepName] = true
	recStack[stepName] = true

	for _, dep := range r.dependencies[stepName] {
		if !visited[dep] {
			if r.hasCycle(dep, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			return true
		}
	}

	recStack[stepName] = false
	return false
}

// ValidateDependenciesExist checks that all referenced dependencies exist
func (r *DependencyResolver) ValidateDependenciesExist() error {
	for stepName, deps := range r.dependencies {
		for _, dep := range deps {
			if _, exists := r.stepsByName[dep]; !exists {
				return fmt.Errorf("step %s depends on non-existent step: %s", stepName, dep)
			}
		}
	}
	return nil
}

// GetDependents returns all steps that depend on the given step
func (r *DependencyResolver) GetDependents(stepName string) []string {
	return r.dependents[stepName]
}

// GetDependencies returns all dependencies for the given step
func (r *DependencyResolver) GetDependencies(stepName string) []string {
	return r.dependencies[stepName]
}

// GetExecutionOrder returns a valid execution order (topological sort)
// This is useful for sequential fallback mode
func (r *DependencyResolver) GetExecutionOrder() ([]*config.StepV2, error) {
	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int)
	
	// Calculate in-degree for each step
	for _, step := range r.steps {
		inDegree[step.Name] = len(r.dependencies[step.Name])
	}

	// Queue for steps with no dependencies
	var queue []*config.StepV2
	for _, step := range r.steps {
		if inDegree[step.Name] == 0 {
			queue = append(queue, step)
		}
	}

	var order []*config.StepV2

	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		// Reduce in-degree for dependents
		for _, dependent := range r.dependents[current.Name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, r.stepsByName[dependent])
			}
		}
	}

	// If we didn't process all steps, there's a cycle
	if len(order) != len(r.steps) {
		return nil, fmt.Errorf("circular dependency detected - cannot determine execution order")
	}

	return order, nil
}
