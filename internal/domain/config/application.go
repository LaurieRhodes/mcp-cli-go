package config

import (
	"strings"
)

// ApplicationConfig represents the complete application configuration
type ApplicationConfig struct {
	Servers     map[string]ServerConfig      `yaml:"servers"`
	AI          *AIConfig                    `yaml:"ai,omitempty"`
	Embeddings  *EmbeddingsConfig            `yaml:"embeddings,omitempty"`
	Chat        *ChatConfig                  `yaml:"chat,omitempty"`
	Skills      *SkillsConfig                `yaml:"skills,omitempty"`
	Workflows   map[string]*WorkflowV2       `yaml:"-"` // Loaded separately from config/workflows/
}

// ValidateWorkflows validates all workflow v2 definitions
func (c *ApplicationConfig) ValidateWorkflows() error {
	if c.Workflows == nil {
		return nil
	}

	// Workflows are validated during loading by the Loader
	// This is a placeholder for additional validation if needed
	return nil
}

// GetWorkflow retrieves a workflow v2 by name
// GetWorkflow retrieves a workflow by name with directory-aware resolution
// If contextDir is provided, it will try to resolve relative to that directory first
func (c *ApplicationConfig) GetWorkflow(name string) (*WorkflowV2, bool) {
	if c.Workflows == nil {
		return nil, false
	}

	workflow, exists := c.Workflows[name]
	return workflow, exists
}

// GetWorkflowWithContext retrieves a workflow by name with directory context
// It tries to resolve in this order:
// 1. Exact match (supports directory notation like "dir/workflow")
// 2. Same directory as caller (if contextDir provided)
// 3. Root directory (workflow name only)
func (c *ApplicationConfig) GetWorkflowWithContext(name string, contextDir string) (*WorkflowV2, bool) {
	if c.Workflows == nil {
		return nil, false
	}

	// Try 1: Exact match first (supports explicit directory notation)
	if workflow, exists := c.Workflows[name]; exists {
		return workflow, true
	}

	// Try 2: If we have a context directory and name has no directory, try same directory
	if contextDir != "" && !strings.Contains(name, "/") {
		contextualName := contextDir + "/" + name
		if workflow, exists := c.Workflows[contextualName]; exists {
			return workflow, true
		}
	}

	// Try 3: Already tried root in step 1, so not found
	return nil, false
}

// ListWorkflows returns all available workflow v2 names
func (c *ApplicationConfig) ListWorkflows() []string {
	if c.Workflows == nil {
		return []string{}
	}

	names := make([]string, 0, len(c.Workflows))
	for name := range c.Workflows {
		names = append(names, name)
	}

	return names
}

// SkillsConfig represents skills-related configuration
type SkillsConfig struct {
	// OutputsDir is the directory where skill outputs are persisted
	OutputsDir string `yaml:"outputs_dir,omitempty"`
}

// GetOutputsDir returns the outputs directory with fallback to default
func (s *SkillsConfig) GetOutputsDir() string {
	if s == nil || s.OutputsDir == "" {
		return "/tmp/mcp-outputs"
	}
	return s.OutputsDir
}
