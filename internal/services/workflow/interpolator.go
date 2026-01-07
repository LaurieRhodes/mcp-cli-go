package workflow

import (
	"fmt"
	"regexp"
	"strings"
)

// Interpolator handles variable interpolation in workflow prompts
type Interpolator struct {
	variables map[string]string
}

// NewInterpolator creates a new interpolator with given variables
func NewInterpolator() *Interpolator {
	return &Interpolator{
		variables: make(map[string]string),
	}
}

// Set sets a variable value
func (i *Interpolator) Set(key, value string) {
	i.variables[key] = value
}

// SetStepResult sets a step's result
func (i *Interpolator) SetStepResult(stepName, result string) {
	i.variables[stepName] = result
}

// SetEnv sets environment variables
func (i *Interpolator) SetEnv(env map[string]string) {
	for k, v := range env {
		i.variables["env."+k] = v
	}
}

// Interpolate replaces all {{variable}} references in text
func (i *Interpolator) Interpolate(text string) (string, error) {
	// Regex to match {{variable}} or {{step.output}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := text
	missingVars := []string{}

	// Find all matches
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		// Get variable name (trim whitespace)
		varName := strings.TrimSpace(match[1])
		placeholder := match[0]

		// Look up value
		value, ok := i.variables[varName]
		if !ok {
			missingVars = append(missingVars, varName)
			continue
		}

		// Replace
		result = strings.Replace(result, placeholder, value, -1)
	}

	if len(missingVars) > 0 {
		return result, fmt.Errorf("undefined variables: %v", missingVars)
	}

	return result, nil
}

// HasVariable checks if a variable is defined
func (i *Interpolator) HasVariable(name string) bool {
	_, ok := i.variables[name]
	return ok
}

// GetVariable gets a variable value
func (i *Interpolator) GetVariable(name string) (string, bool) {
	val, ok := i.variables[name]
	return val, ok
}

// Clear clears all variables
func (i *Interpolator) Clear() {
	i.variables = make(map[string]string)
}

// Clone creates a copy of the interpolator
func (i *Interpolator) Clone() *Interpolator {
	clone := NewInterpolator()
	for k, v := range i.variables {
		clone.variables[k] = v
	}
	return clone
}

// SetLoopVars sets loop-specific variables for interpolation
func (i *Interpolator) SetLoopVars(iteration int, lastOutput string, allOutputs []string) {
	i.variables["loop.iteration"] = fmt.Sprintf("%d", iteration)
	i.variables["loop.output"] = lastOutput
	if iteration > 1 {
		i.variables["loop.last.output"] = lastOutput
	}
	if len(allOutputs) > 0 {
		i.variables["loop.history"] = strings.Join(allOutputs, "\n---\n")
	}
}

// SetLoopVars sets loop-specific variables for interpolation

// CopyLoopVars copies loop variables from this interpolator to another
func (i *Interpolator) CopyLoopVars(dest *Interpolator) {
	// Copy all loop.* variables
	for key, value := range i.variables {
		if strings.HasPrefix(key, "loop.") {
			dest.variables[key] = value
		}
	}
}
