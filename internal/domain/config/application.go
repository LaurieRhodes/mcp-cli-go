package config

// ApplicationConfig represents the complete application configuration
type ApplicationConfig struct {
	Servers     map[string]ServerConfig      `yaml:"servers"`
	AI          *AIConfig                    `yaml:"ai,omitempty"`
	Embeddings  *EmbeddingsConfig            `yaml:"embeddings,omitempty"`
	Chat        *ChatConfig                  `yaml:"chat,omitempty"`
	Skills      *SkillsConfig                `yaml:"skills,omitempty"`
	Templates   map[string]*WorkflowTemplate `yaml:"templates,omitempty"`
	TemplatesV2 map[string]*TemplateV2       `yaml:"-"` // Loaded separately from config/templates/
}

// ValidateWorkflowTemplates validates all workflow templates in the configuration
func (c *ApplicationConfig) ValidateWorkflowTemplates() error {
	if c.Templates == nil {
		return nil
	}

	for templateName, template := range c.Templates {
		if err := template.ValidateWorkflowTemplate(); err != nil {
			return NewConfigError("invalid workflow template").
				WithContext("template", templateName).
				WithCause(err)
		}
	}

	return nil
}

// GetWorkflowTemplate retrieves a workflow template by name
func (c *ApplicationConfig) GetWorkflowTemplate(name string) (*WorkflowTemplate, bool) {
	if c.Templates == nil {
		return nil, false
	}

	template, exists := c.Templates[name]
	return template, exists
}

// ListWorkflowTemplates returns all available workflow template names
func (c *ApplicationConfig) ListWorkflowTemplates() []string {
	if c.Templates == nil {
		return []string{}
	}

	names := make([]string, 0, len(c.Templates))
	for name := range c.Templates {
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
