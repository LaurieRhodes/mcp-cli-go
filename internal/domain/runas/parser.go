package runas

import (
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// ExposureParser parses the expose field and converts it to ToolExposure structs
type ExposureParser struct {
	appConfig  *config.ApplicationConfig
	mcpServers []*host.ServerConnection
}

// NewExposureParser creates a new exposure parser
func NewExposureParser(appConfig *config.ApplicationConfig, mcpServers []*host.ServerConnection) *ExposureParser {
	return &ExposureParser{
		appConfig:  appConfig,
		mcpServers: mcpServers,
	}
}

// ParseExposure parses the expose field and returns a list of ToolExposure items
func (p *ExposureParser) ParseExposure(expose []interface{}) ([]ToolExposure, error) {
	var tools []ToolExposure
	
	for i, item := range expose {
		itemTools, err := p.parseExposureItem(item, i)
		if err != nil {
			return nil, err
		}
		tools = append(tools, itemTools...)
	}
	
	return tools, nil
}

// parseExposureItem parses a single exposure item
func (p *ExposureParser) parseExposureItem(item interface{}, index int) ([]ToolExposure, error) {
	switch v := item.(type) {
	case string:
		return p.parseStringExposure(v, index)
	case map[string]interface{}:
		return p.parseMapExposure(v, index)
	case map[interface{}]interface{}:
		// YAML sometimes returns this format
		converted := make(map[string]interface{})
		for k, val := range v {
			if strKey, ok := k.(string); ok {
				converted[strKey] = val
			}
		}
		return p.parseMapExposure(converted, index)
	default:
		return nil, fmt.Errorf("expose item %d: invalid type %T (must be string or map)", index, v)
	}
}

// parseStringExposure parses a string exposure item
func (p *ExposureParser) parseStringExposure(name string, index int) ([]ToolExposure, error) {
	// Check if it's server.tool notation
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("expose item %d: invalid server.tool format: %s", index, name)
		}
		
		serverName := parts[0]
		toolName := parts[1]
		
		// Verify server exists
		if _, exists := p.appConfig.Servers[serverName]; !exists {
			return nil, fmt.Errorf("expose item %d: server not found: %s", index, serverName)
		}
		
		logging.Debug("Exposing specific tool: %s from server %s", toolName, serverName)
		return []ToolExposure{{
			MCPServer: serverName,
			MCPTool:   toolName,
			Name:      toolName,
		}}, nil
	}
	
	// Check if it's a template
	_, existsV2 := p.appConfig.TemplatesV2[name]
	_, existsV1 := p.appConfig.Templates[name]
	
	if existsV2 || existsV1 {
		logging.Debug("Exposing template: %s", name)
		return []ToolExposure{{
			Template: name,
			Name:     name,
		}}, nil
	}
	
	// Check if it's a server (expose all tools)
	if _, exists := p.appConfig.Servers[name]; exists {
		logging.Debug("Exposing all tools from server: %s", name)
		return p.autoDiscoverServerTools(name)
	}
	
	// Check if it's a map with server and tools list
	// This handles YAML parsing where a map might be in the list
	
	// Not found
	return nil, fmt.Errorf("expose item %d: '%s' not found (not a server, template, or server.tool)", index, name)
}

// parseMapExposure parses a map exposure item for customization
func (p *ExposureParser) parseMapExposure(m map[string]interface{}, index int) ([]ToolExposure, error) {
	// Check if it's {server: name, tools: [list]} format
	if serverName, ok := m["server"].(string); ok {
		// Check for tools list
		if toolsList, ok := m["tools"]; ok {
			// Convert to string slice
			var toolNames []string
			switch t := toolsList.(type) {
			case []interface{}:
				for _, tool := range t {
					if toolStr, ok := tool.(string); ok {
						toolNames = append(toolNames, toolStr)
					}
				}
			case []string:
				toolNames = t
			default:
				return nil, fmt.Errorf("expose item %d: 'tools' must be a list of strings", index)
			}
			
			// Create tool exposure for each tool
			var tools []ToolExposure
			for _, toolName := range toolNames {
				tools = append(tools, ToolExposure{
					MCPServer: serverName,
					MCPTool:   toolName,
					Name:      toolName,
				})
			}
			return tools, nil
		}
		
		// No tools list - expose all tools from server
		return p.autoDiscoverServerTools(serverName)
	}
	
	// Check if it's {template: name, as: custom_name} format
	if templateName, ok := m["template"].(string); ok {
		tool := ToolExposure{
			Template: templateName,
			Name:     templateName,
		}
		
		// Check for custom name
		if asName, ok := m["as"].(string); ok {
			tool.Name = asName
		}
		
		// Check for custom description
		if desc, ok := m["description"].(string); ok {
			tool.Description = desc
		}
		
		return []ToolExposure{tool}, nil
	}
	
	// Check if it's {server: name, tool: name, as: custom_name} format
	if serverName, ok := m["server"].(string); ok {
		if toolName, ok := m["tool"].(string); ok {
			tool := ToolExposure{
				MCPServer: serverName,
				MCPTool:   toolName,
				Name:      toolName,
			}
			
			// Check for custom name
			if asName, ok := m["as"].(string); ok {
				tool.Name = asName
			}
			
			// Check for custom description
			if desc, ok := m["description"].(string); ok {
				tool.Description = desc
			}
			
			return []ToolExposure{tool}, nil
		}
	}
	
	return nil, fmt.Errorf("expose item %d: invalid map format (use {server: name} or {template: name} or {server: name, tool: name})", index)
}

// autoDiscoverServerTools discovers all tools from a server
func (p *ExposureParser) autoDiscoverServerTools(serverName string) ([]ToolExposure, error) {
	// Find the server connection
	var serverConn *host.ServerConnection
	for _, conn := range p.mcpServers {
		if conn.Name == serverName {
			serverConn = conn
			break
		}
	}
	
	if serverConn == nil {
		return nil, fmt.Errorf("server %s not connected (required for auto-discovery)", serverName)
	}
	
	// Get all tools from the server
	logging.Debug("Auto-discovering tools from server: %s", serverName)
	result, err := tools.SendToolsList(serverConn.Client, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools from server %s: %w", serverName, err)
	}
	
	// Convert to ToolExposure
	var exposures []ToolExposure
	for _, tool := range result.Tools {
		exposures = append(exposures, ToolExposure{
			MCPServer: serverName,
			MCPTool:   tool.Name,
			Name:      tool.Name,
		})
		logging.Debug("  - Discovered tool: %s", tool.Name)
	}
	
	logging.Info("Auto-discovered %d tools from server %s", len(exposures), serverName)
	return exposures, nil
}

// ParseServer parses the server shorthand field
func (p *ExposureParser) ParseServer(serverName string) ([]ToolExposure, error) {
	// Verify server exists
	if _, exists := p.appConfig.Servers[serverName]; !exists {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}
	
	logging.Debug("Exposing all tools from server (shorthand): %s", serverName)
	return p.autoDiscoverServerTools(serverName)
}
