package config


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
func (c *ApplicationConfig) GetWorkflow(name string) (*WorkflowV2, bool) {
	if c.Workflows == nil {
		return nil, false
	}

	workflow, exists := c.Workflows[name]
	return workflow, exists
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
