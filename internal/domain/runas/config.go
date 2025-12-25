package runas

import (
	"encoding/json"
	"fmt"
)

// RunAsType defines the type of server to run
type RunAsType string

const (
	// RunAsTypeMCP indicates this runs as an MCP server
	RunAsTypeMCP RunAsType = "mcp"
	
	// Future types
	// RunAsTypeAaA RunAsType = "aaa"  // Agent-as-a-API
	// RunAsTypeACP RunAsType = "acp"  // Agent Communication Protocol
)

// RunAsConfig defines how to expose templates as an MCP server
type RunAsConfig struct {
	// Type of server (mcp, aaa, acp, etc.)
	RunAsType RunAsType `yaml:"runas_type" json:"runas_type"`
	
	// Config version
	Version string `yaml:"version" json:"version"`
	
	// Server metadata
	ServerInfo ServerInfo `yaml:"server_info" json:"server_info"`
	
	// Tools to expose (templates mapped to MCP tools)
	Tools []ToolExposure `yaml:"tools" json:"tools"`
}

// ServerInfo contains metadata about the MCP server
type ServerInfo struct {
	// Server name (used in MCP initialize response)
	Name string `yaml:"name" json:"name"`
	
	// Server version
	Version string `yaml:"version" json:"version"`
	
	// Human-readable description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// ToolExposure defines how a template is exposed as an MCP tool
type ToolExposure struct {
	// Template name (must exist in application config)
	Template string `yaml:"template" json:"template"`
	
	// Tool name (exposed to MCP clients)
	Name string `yaml:"name" json:"name"`
	
	// Tool description (shown to MCP clients)
	Description string `yaml:"description" json:"description"`
	
	// Input schema (JSON Schema for tool parameters)
	InputSchema map[string]interface{} `yaml:"input_schema" json:"input_schema"`
	
	// Optional: Map tool inputs to template variables
	// Format: {"tool_param": "{{template_var}}"}
	InputMapping map[string]string `yaml:"input_mapping,omitempty" json:"input_mapping,omitempty"`
	
	// Optional: Override template settings
	Overrides *ToolOverrides `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

// ToolOverrides allows overriding template configuration per tool
type ToolOverrides struct {
	// Override provider
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
	
	// Override model
	Model string `yaml:"model,omitempty" json:"model,omitempty"`
	
	// Override max steps
	MaxSteps int `yaml:"max_steps,omitempty" json:"max_steps,omitempty"`
	
	// Override timeout (seconds)
	TimeoutSeconds int `yaml:"timeout_seconds,omitempty" json:"timeout_seconds,omitempty"`
}

// Validate validates the RunAs configuration
func (c *RunAsConfig) Validate() error {
	// Check runas type
	if c.RunAsType == "" {
		return fmt.Errorf("runas_type is required")
	}
	
	// Only MCP is supported for now
	if c.RunAsType != RunAsTypeMCP {
		return fmt.Errorf("unsupported runas_type: %s (only 'mcp' is currently supported)", c.RunAsType)
	}
	
	// Check version
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	// Validate server info
	if err := c.ServerInfo.Validate(); err != nil {
		return fmt.Errorf("invalid server_info: %w", err)
	}
	
	// Must have at least one tool
	if len(c.Tools) == 0 {
		return fmt.Errorf("at least one tool must be defined")
	}
	
	// Validate each tool
	toolNames := make(map[string]bool)
	for i, tool := range c.Tools {
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("invalid tool at index %d: %w", i, err)
		}
		
		// Check for duplicate tool names
		if toolNames[tool.Name] {
			return fmt.Errorf("duplicate tool name: %s", tool.Name)
		}
		toolNames[tool.Name] = true
	}
	
	return nil
}

// Validate validates the ServerInfo
func (s *ServerInfo) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("server name is required")
	}
	
	if s.Version == "" {
		return fmt.Errorf("server version is required")
	}
	
	return nil
}

// Validate validates the ToolExposure
func (t *ToolExposure) Validate() error {
	if t.Template == "" {
		return fmt.Errorf("template is required")
	}
	
	if t.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	
	if t.Description == "" {
		return fmt.Errorf("tool description is required")
	}
	
	// Validate input schema is a valid JSON Schema object
	if t.InputSchema == nil {
		return fmt.Errorf("input_schema is required")
	}
	
	// Check that it has a type field
	schemaType, ok := t.InputSchema["type"]
	if !ok {
		return fmt.Errorf("input_schema must have a 'type' field")
	}
	
	// Type should be "object" for MCP tools
	if schemaType != "object" {
		return fmt.Errorf("input_schema type must be 'object', got: %v", schemaType)
	}
	
	// Validate it's valid JSON (can be marshaled)
	if _, err := json.Marshal(t.InputSchema); err != nil {
		return fmt.Errorf("input_schema is not valid JSON: %w", err)
	}
	
	return nil
}

// GetToolByName retrieves a tool exposure by name
func (c *RunAsConfig) GetToolByName(name string) (*ToolExposure, bool) {
	for i := range c.Tools {
		if c.Tools[i].Name == name {
			return &c.Tools[i], true
		}
	}
	return nil, false
}

// ListToolNames returns all tool names
func (c *RunAsConfig) ListToolNames() []string {
	names := make([]string, len(c.Tools))
	for i, tool := range c.Tools {
		names[i] = tool.Name
	}
	return names
}
