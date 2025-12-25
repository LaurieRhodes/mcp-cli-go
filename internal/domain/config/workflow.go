package config

import (
	"fmt"
	"strings"
)

// WorkflowTemplate represents a workflow template
type WorkflowTemplate struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Steps       []WorkflowStep `yaml:"steps"`
	Variables   map[string]string `yaml:"variables,omitempty"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Step           int      `yaml:"step"`
	Name           string   `yaml:"name"`
	BasePrompt     string   `yaml:"base_prompt"`
	SystemPrompt   string   `yaml:"system_prompt,omitempty"`
	Provider       string   `yaml:"provider,omitempty"`
	Model          string   `yaml:"model,omitempty"`
	Servers        []string `yaml:"servers,omitempty"`
	ToolsRequired  []string `yaml:"tools_required,omitempty"`
	Temperature    float64  `yaml:"temperature,omitempty"`
	MaxTokens      int      `yaml:"max_tokens,omitempty"`
}

// ValidateWorkflowTemplate validates the workflow template structure
func (w *WorkflowTemplate) ValidateWorkflowTemplate() error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	for i, step := range w.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i+1)
		}
		if step.BasePrompt == "" {
			return fmt.Errorf("step %d: base_prompt is required", i+1)
		}
	}

	return nil
}

// ProcessVariables replaces variables in text with their values
func (w *WorkflowTemplate) ProcessVariables(text string, variables map[string]interface{}) string {
	result := text
	
	// Process template variables first
	for key, value := range w.Variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	// Process runtime variables
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	return result
}
