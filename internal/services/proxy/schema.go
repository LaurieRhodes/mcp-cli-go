package proxy

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// SchemaGenerator generates input schemas for tools
type SchemaGenerator struct {
	appConfig  *config.ApplicationConfig
	mcpServers []*host.ServerConnection
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(appConfig *config.ApplicationConfig, mcpServers []*host.ServerConnection) *SchemaGenerator {
	return &SchemaGenerator{
		appConfig:  appConfig,
		mcpServers: mcpServers,
	}
}

// GenerateSchema generates an input schema for a tool
// Returns the schema and description (both can be auto-generated)
func (g *SchemaGenerator) GenerateSchema(tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// If schema is already provided, use it
	if tool.InputSchema != nil && len(tool.InputSchema) > 0 {
		description := tool.Description
		if description == "" {
			description = fmt.Sprintf("Execute %s", tool.Name)
		}
		return tool.InputSchema, description, nil
	}
	
	// Auto-generate based on source
	if tool.Template != "" {
		return g.generateFromTemplate(tool)
	}
	
	if tool.MCPServer != "" && tool.MCPTool != "" {
		return g.generateFromMCPTool(tool)
	}
	
	return nil, "", fmt.Errorf("cannot generate schema: no template or mcp_server specified")
}

// generateFromTemplate generates schema from a workflow template
func (g *SchemaGenerator) generateFromTemplate(tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// Look for template in V2 format first
	if templateV2, exists := g.appConfig.TemplatesV2[tool.Template]; exists {
		return g.generateFromTemplateV2(templateV2, tool)
	}
	
	// Fallback to V1 format
	if templateV1, exists := g.appConfig.Templates[tool.Template]; exists {
		return g.generateFromTemplateV1(templateV1, tool)
	}
	
	return nil, "", fmt.Errorf("template not found: %s", tool.Template)
}

// generateFromTemplateV2 generates schema from a V2 template
func (g *SchemaGenerator) generateFromTemplateV2(template *config.TemplateV2, tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	description := tool.Description
	if description == "" {
		description = template.Description
		if description == "" {
			description = fmt.Sprintf("Execute %s workflow", template.Name)
		}
	}
	
	// Check if template has step definitions with input schemas
	if template.StepDefinitions != nil && len(template.StepDefinitions) > 0 {
		// Look for first step definition with inputs
		for _, stepDef := range template.StepDefinitions {
			if stepDef.Inputs != nil && len(stepDef.Inputs) > 0 {
				schema := g.convertInputSchemaFromStepDef(stepDef.Inputs)
				logging.Debug("Generated schema from step definition: %v", schema)
				return schema, description, nil
			}
		}
	}
	
	// Check first step for validation schema
	if len(template.Steps) > 0 {
		firstStep := template.Steps[0]
		if firstStep.ValidateInput != nil && firstStep.ValidateInput.Schema != nil {
			logging.Debug("Generated schema from step validation: %v", firstStep.ValidateInput.Schema)
			return firstStep.ValidateInput.Schema, description, nil
		}
	}
	
	// Default: generic input_data schema
	return g.generateGenericSchema(), description, nil
}

// generateFromTemplateV1 generates schema from a V1 template
func (g *SchemaGenerator) generateFromTemplateV1(template *config.WorkflowTemplate, tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	description := tool.Description
	if description == "" {
		description = template.Description
		if description == "" {
			description = fmt.Sprintf("Execute %s workflow", template.Name)
		}
	}
	
	// V1 templates don't have schema metadata, use generic
	return g.generateGenericSchema(), description, nil
}

// generateFromMCPTool generates schema by querying an MCP server
func (g *SchemaGenerator) generateFromMCPTool(tool *runas.ToolExposure) (map[string]interface{}, string, error) {
	// Find the MCP server connection
	var serverConn *host.ServerConnection
	for _, conn := range g.mcpServers {
		if conn.Name == tool.MCPServer {
			serverConn = conn
			break
		}
	}
	
	if serverConn == nil {
		return nil, "", fmt.Errorf("MCP server not found: %s", tool.MCPServer)
	}
	
	// Get tool list from the server
	logging.Debug("Fetching tools from MCP server: %s", tool.MCPServer)
	result, err := tools.SendToolsList(serverConn.Client, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get tools from MCP server %s: %w", tool.MCPServer, err)
	}
	
	// Find the specific tool
	var mcpTool *tools.Tool
	for i := range result.Tools {
		if result.Tools[i].Name == tool.MCPTool {
			mcpTool = &result.Tools[i]
			break
		}
	}
	
	if mcpTool == nil {
		return nil, "", fmt.Errorf("tool %s not found on MCP server %s", tool.MCPTool, tool.MCPServer)
	}
	
	// Use the MCP tool's schema and description
	description := tool.Description
	if description == "" {
		description = mcpTool.Description
		if description == "" {
			description = fmt.Sprintf("Execute %s from %s", tool.MCPTool, tool.MCPServer)
		}
	}
	
	schema := mcpTool.InputSchema
	if schema == nil {
		// If MCP tool doesn't provide schema, use generic
		logging.Warn("MCP tool %s on server %s has no input schema, using generic", tool.MCPTool, tool.MCPServer)
		schema = g.generateGenericSchema()
	}
	
	logging.Info("Auto-generated schema from MCP server %s tool %s", tool.MCPServer, tool.MCPTool)
	return schema, description, nil
}

// generateGenericSchema creates a generic input schema
func (g *SchemaGenerator) generateGenericSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input_data": map[string]interface{}{
				"type":        "string",
				"description": "Input data for the workflow",
			},
		},
		"required": []interface{}{"input_data"},
	}
}

// convertInputSchemaFromStepDef converts step definition inputs to JSON schema
func (g *SchemaGenerator) convertInputSchemaFromStepDef(inputs map[string]*config.InputSchema) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []interface{}{}
	
	for name, input := range inputs {
		prop := make(map[string]interface{})
		prop["type"] = input.Type
		
		if input.Description != "" {
			prop["description"] = input.Description
		}
		
		if input.Default != nil {
			prop["default"] = input.Default
		}
		
		if input.Enum != nil && len(input.Enum) > 0 {
			prop["enum"] = input.Enum
		}
		
		properties[name] = prop
		
		if input.Required {
			required = append(required, name)
		}
	}
	
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	
	if len(required) > 0 {
		schema["required"] = required
	}
	
	return schema
}
