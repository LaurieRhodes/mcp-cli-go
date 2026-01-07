package proxy

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
)

// SchemaGenerator generates OpenAPI-style schemas for MCP tools
type SchemaGenerator struct {
	appConfig *config.ApplicationConfig
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(appConfig *config.ApplicationConfig) *SchemaGenerator {
	return &SchemaGenerator{
		appConfig: appConfig,
	}
}

// GenerateForTool generates a schema for a tool exposure
func (g *SchemaGenerator) GenerateForTool(tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// If tool has explicit input schema, use it
	if tool.InputSchema != nil {
		return tool.InputSchema, tool.Description, nil
	}

	// Otherwise generate from workflow template
	return g.generateFromTemplate(tool)
}

// generateFromTemplate generates schema from a workflow template
func (g *SchemaGenerator) generateFromTemplate(tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// Look for workflow
	if workflow, exists := g.appConfig.Workflows[tool.Template]; exists {
		return g.generateFromWorkflowV2(workflow, tool)
	}
	
	return nil, "", fmt.Errorf("workflow not found: %s", tool.Template)
}

// generateFromWorkflowV2 generates schema from WorkflowV2 (stub for now)
func (g *SchemaGenerator) generateFromWorkflowV2(template *config.WorkflowV2, tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// Simple default schema for all workflows
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input_data": map[string]interface{}{
				"type":        "string",
				"description": "Input data for the workflow",
			},
		},
		"required": []string{"input_data"},
	}
	
	description := template.Description
	if description == "" {
		description = fmt.Sprintf("Execute workflow: %s", template.Name)
	}
	
	return schema, description, nil
}
