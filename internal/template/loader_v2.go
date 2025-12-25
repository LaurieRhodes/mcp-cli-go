package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// LoaderV2 handles loading template v2 YAML files
type LoaderV2 struct {
	configDir string
}

// NewLoaderV2 creates a new template v2 loader
func NewLoaderV2(configDir string) *LoaderV2 {
	return &LoaderV2{
		configDir: configDir,
	}
}

// LoadTemplate loads a single template from a YAML file
func (l *LoaderV2) LoadTemplate(path string) (*config.TemplateV2, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", path, err)
	}

	var template config.TemplateV2
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template YAML %s: %w", path, err)
	}

	// Validate template
	if err := l.validateTemplate(&template); err != nil {
		return nil, fmt.Errorf("template validation failed for %s: %w", path, err)
	}

	// Process includes (step libraries)
	if err := l.processIncludes(&template); err != nil {
		return nil, fmt.Errorf("failed to process includes for %s: %w", path, err)
	}

	return &template, nil
}

// LoadAllTemplates loads all templates from config/templates/*.yaml
func (l *LoaderV2) LoadAllTemplates() (map[string]*config.TemplateV2, error) {
	templatesDir := filepath.Join(l.configDir, "templates")
	templates := make(map[string]*config.TemplateV2)

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty map (not an error)
		return templates, nil
	}

	// Find all YAML files in templates directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob templates: %w", err)
	}

	// Also check for .yml extension
	ymlFiles, err := filepath.Glob(filepath.Join(templatesDir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob templates: %w", err)
	}
	files = append(files, ymlFiles...)

	for _, file := range files {
		template, err := l.LoadTemplate(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %s: %w", file, err)
		}

		// Use template name as key
		templates[template.Name] = template
	}

	return templates, nil
}

// validateTemplate validates a template structure
func (l *LoaderV2) validateTemplate(t *config.TemplateV2) error {
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if t.Description == "" {
		return fmt.Errorf("template description is required")
	}

	if len(t.Steps) == 0 {
		return fmt.Errorf("template must have at least one step")
	}

	// Validate step names are unique
	stepNames := make(map[string]bool)
	for i, step := range t.Steps {
		if err := l.validateStep(&step, i); err != nil {
			return err
		}

		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}

		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true
	}

	return nil
}

// validateStep validates a single step
func (l *LoaderV2) validateStep(step *config.WorkflowStepV2, index int) error {
	stepType := step.GetStepType()

	switch stepType {
	case config.StepTypeBasic:
		if step.Prompt == "" && step.Use == "" {
			return fmt.Errorf("step %d (%s): prompt is required for basic steps", index, step.Name)
		}

	case config.StepTypeParallel:
		if step.Parallel == nil || len(step.Parallel.Steps) == 0 {
			return fmt.Errorf("step %d (%s): parallel execution requires sub-steps", index, step.Name)
		}
		// Validate parallel sub-steps
		for i, subStep := range step.Parallel.Steps {
			if err := l.validateStep(&subStep, i); err != nil {
				return fmt.Errorf("step %d (%s), parallel sub-step %d: %w", index, step.Name, i, err)
			}
		}

	case config.StepTypeLoop:
		if step.ForEach == "" {
			return fmt.Errorf("step %d (%s): for_each is required for loop steps", index, step.Name)
		}
		if step.Prompt == "" {
			return fmt.Errorf("step %d (%s): prompt is required for loop steps", index, step.Name)
		}

	case config.StepTypeTransform:
		if step.Transform == nil {
			return fmt.Errorf("step %d (%s): transform config is required for transform steps", index, step.Name)
		}
		if step.Transform.Input == "" {
			return fmt.Errorf("step %d (%s): transform input is required", index, step.Name)
		}
		if len(step.Transform.Operations) == 0 {
			return fmt.Errorf("step %d (%s): transform must have at least one operation", index, step.Name)
		}

	case config.StepTypeUse:
		if step.Use == "" {
			return fmt.Errorf("step %d (%s): use field is required for reuse steps", index, step.Name)
		}

	case config.StepTypeNested:
		if len(step.Steps) == 0 {
			return fmt.Errorf("step %d (%s): nested steps are required", index, step.Name)
		}
		// Validate nested steps
		for i, nestedStep := range step.Steps {
			if err := l.validateStep(&nestedStep, i); err != nil {
				return fmt.Errorf("step %d (%s), nested step %d: %w", index, step.Name, i, err)
			}
		}
	}

	return nil
}

// processIncludes loads and merges step libraries
func (l *LoaderV2) processIncludes(t *config.TemplateV2) error {
	if len(t.Includes) == 0 {
		return nil
	}

	// Initialize step definitions if not exists
	if t.StepDefinitions == nil {
		t.StepDefinitions = make(map[string]*config.StepDefinition)
	}

	for _, includePath := range t.Includes {
		// Make path absolute relative to config directory
		fullPath := includePath
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(l.configDir, includePath)
		}

		// Load step library
		library, err := l.loadStepLibrary(fullPath)
		if err != nil {
			return fmt.Errorf("failed to load step library %s: %w", includePath, err)
		}

		// Merge step definitions
		for name, def := range library.Steps {
			// Check for conflicts
			if existing, exists := t.StepDefinitions[name]; exists {
				return fmt.Errorf("step definition conflict: %s already defined (existing: %v, new: %v)", 
					name, existing, def)
			}
			t.StepDefinitions[name] = def
		}
	}

	return nil
}

// loadStepLibrary loads a step library from a YAML file
func (l *LoaderV2) loadStepLibrary(path string) (*config.StepLibrary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read step library file: %w", err)
	}

	// Try parsing as step library first
	var library config.StepLibrary
	if err := yaml.Unmarshal(data, &library); err == nil && library.Steps != nil {
		return &library, nil
	}

	// Try parsing as simple steps map
	var stepsMap struct {
		Steps map[string]*config.StepDefinition `yaml:"steps"`
	}
	if err := yaml.Unmarshal(data, &stepsMap); err != nil {
		return nil, fmt.Errorf("failed to parse step library: %w", err)
	}

	return &config.StepLibrary{
		Steps: stepsMap.Steps,
	}, nil
}

// ListTemplates returns a list of available template names
func (l *LoaderV2) ListTemplates() ([]string, error) {
	templates, err := l.LoadAllTemplates()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}

	return names, nil
}

// GetTemplate retrieves a specific template by name
func (l *LoaderV2) GetTemplate(name string) (*config.TemplateV2, error) {
	templates, err := l.LoadAllTemplates()
	if err != nil {
		return nil, err
	}

	template, exists := templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return template, nil
}

// ValidateAllTemplates validates all templates in the templates directory
func (l *LoaderV2) ValidateAllTemplates() error {
	templates, err := l.LoadAllTemplates()
	if err != nil {
		return err
	}

	var errors []string
	for name, template := range templates {
		if err := l.validateTemplate(template); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("template validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
