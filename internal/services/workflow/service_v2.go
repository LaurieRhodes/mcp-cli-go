package workflow

import (
	"context"
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/template"
)

// ServiceV2 provides template v2 execution capabilities
type ServiceV2 struct {
	config        *config.ApplicationConfig
	provider      domain.LLMProvider
	serverManager domain.MCPServerManager
}

// NewServiceV2 creates a new template v2 service
func NewServiceV2(
	appConfig *config.ApplicationConfig,
	provider domain.LLMProvider,
	serverManager domain.MCPServerManager,
) *ServiceV2 {
	return &ServiceV2{
		config:        appConfig,
		provider:      provider,
		serverManager: serverManager,
	}
}

// ExecuteTemplate executes a template v2 workflow
func (s *ServiceV2) ExecuteTemplate(ctx context.Context, templateName string, inputData string) (*template.ExecutionResult, error) {
	logging.Info("Executing template v2: %s", templateName)

	// Get template
	tmpl, exists := s.config.TemplatesV2[templateName]
	if !exists {
		return nil, fmt.Errorf("template v2 not found: %s", templateName)
	}

	// Create executor with template registry support
	executor := template.NewExecutorV2WithRegistry(tmpl, s.provider, s.serverManager, s)

	// Execute
	result, err := executor.Execute(ctx, inputData)
	if err != nil {
		logging.Error("Template v2 execution failed for %s: %v", templateName, err)
		return result, err
	}

	logging.Info("Template v2 execution completed: %s", templateName)
	return result, nil
}

// ListTemplates returns available template v2 names
func (s *ServiceV2) ListTemplates() []string {
	names := make([]string, 0, len(s.config.TemplatesV2))
	for name := range s.config.TemplatesV2 {
		names = append(names, name)
	}
	return names
}

// GetTemplate retrieves a template v2 by name
func (s *ServiceV2) GetTemplate(name string) (*config.TemplateV2, error) {
	tmpl, exists := s.config.TemplatesV2[name]
	if !exists {
		return nil, fmt.Errorf("template v2 not found: %s", name)
	}
	return tmpl, nil
}

// ValidateTemplates validates all template v2 configurations
func (s *ServiceV2) ValidateTemplates() error {
	loader := template.NewLoaderV2(s.config.AI.DefaultSystemPrompt) // Temporary - needs proper config dir
	return loader.ValidateAllTemplates()
}
