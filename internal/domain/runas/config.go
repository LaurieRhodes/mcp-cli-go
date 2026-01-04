package runas

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// RunAsType defines the type of server to run
type RunAsType string

const (
	// RunAsTypeMCP indicates this runs as an MCP server
	RunAsTypeMCP RunAsType = "mcp"
	
	// RunAsTypeMCPSkills indicates auto-generated MCP server from skills directory
	RunAsTypeMCPSkills RunAsType = "mcp-skills"
	
	// RunAsTypeProxy indicates HTTP proxy for MCP workflows
	RunAsTypeProxy RunAsType = "proxy"
	
	// RunAsTypeProxySkills indicates HTTP proxy for auto-discovered skills
	RunAsTypeProxySkills RunAsType = "proxy-skills"
	
	// Future types
	// RunAsTypeAaA RunAsType = "aaa"  // Agent-as-a-API
	// RunAsTypeACP RunAsType = "acp"  // Agent Communication Protocol
)

// RunAsConfig defines how to expose templates as an MCP server
type RunAsConfig struct {
	// Type of server (mcp, mcp-skills, proxy, proxy-skills)
	RunAsType RunAsType `yaml:"runas_type" json:"runas_type"`
	
	// Config version
	Version string `yaml:"version" json:"version"`
	
	// Server metadata
	ServerInfo ServerInfo `yaml:"server_info,omitempty" json:"server_info,omitempty"`
	
	// === Proxy type: Config source ===
	// Path to the config file to proxy (MCP server or workflow template)
	// Examples: "config/servers/brave-search.yaml", "config/templates/deep_research.yaml"
	ConfigSource string `yaml:"config_source,omitempty" json:"config_source,omitempty"`
	
	// === MCP type: Template sources ===
	// For MCP stdio servers - list of templates to expose with config_source
	// Derives tool information from template configs (no duplication)
	Templates []TemplateSource `yaml:"templates,omitempty" json:"templates,omitempty"`
	
	// === Proxy type: Simple exposure (legacy/alternative) ===
	// For proxy type - shorthand to expose one MCP server
	Server string `yaml:"server,omitempty" json:"server,omitempty"`
	
	// For proxy type - explicit list of what to expose (servers, templates, or both)
	// Can be strings ("filesystem", "template_name", "filesystem.read_file")
	// Or maps for customization
	Expose []interface{} `yaml:"expose,omitempty" json:"expose,omitempty"`
	
	// === Advanced: Explicit tool definitions ===
	// Tools to expose (templates mapped to MCP tools)
	// Optional - use Templates with config_source instead for simplicity
	// Not used for runas_type: mcp-skills or proxy-skills (auto-generated)
	Tools []ToolExposure `yaml:"tools,omitempty" json:"tools,omitempty"`
	
	// Skills configuration (for runas_type: mcp-skills, proxy-skills)
	SkillsConfig *SkillsConfig `yaml:"skills_config,omitempty" json:"skills_config,omitempty"`
	
	// Proxy configuration (for runas_type: proxy, proxy-skills)
	ProxyConfig *ProxyConfig `yaml:"proxy_config,omitempty" json:"proxy_config,omitempty"`
}

// TemplateSource specifies a template to expose with its config source
type TemplateSource struct {
	// Path to template config file
	ConfigSource string `yaml:"config_source" json:"config_source"`
	
	// Optional custom tool name (defaults to template name from config)
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	
	// Optional custom description (defaults to template description from config)
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
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
	// Template name (must exist in application config) - used for workflows
	Template string `yaml:"template,omitempty" json:"template,omitempty"`
	
	// MCP Server and tool name - used for proxying MCP server tools
	MCPServer string `yaml:"mcp_server,omitempty" json:"mcp_server,omitempty"`
	MCPTool   string `yaml:"mcp_tool,omitempty" json:"mcp_tool,omitempty"`
	
	// Tool name (exposed to MCP clients)
	Name string `yaml:"name" json:"name"`
	
	// Tool description (shown to MCP clients)
	// Optional - will be auto-generated from template/MCP tool if not provided
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	
	// Input schema (JSON Schema for tool parameters)
	// Optional - will be auto-generated from template/MCP tool if not provided
	InputSchema map[string]interface{} `yaml:"input_schema,omitempty" json:"input_schema,omitempty"`
	
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

// SkillsConfig defines configuration for auto-discovered skills
type SkillsConfig struct {
	// Directory containing skills (defaults to config/skills)
	SkillsDirectory string `yaml:"skills_directory,omitempty" json:"skills_directory,omitempty"`
	
	// Execution mode: "passive" (documentation only), "active" (execute scripts), or "auto" (detect)
	ExecutionMode string `yaml:"execution_mode,omitempty" json:"execution_mode,omitempty"`
	
	// Optional: only include these skills
	IncludeSkills []string `yaml:"include_skills,omitempty" json:"include_skills,omitempty"`
	
	// Optional: exclude these skills
	ExcludeSkills []string `yaml:"exclude_skills,omitempty" json:"exclude_skills,omitempty"`
}

// ProxyConfig defines configuration for HTTP proxy server
type ProxyConfig struct {
	// Port to listen on (defaults to 8080)
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
	
	// Host to bind to (defaults to "0.0.0.0")
	Host string `yaml:"host,omitempty" json:"host,omitempty"`
	
	// API key for authentication (required for proxy types)
	APIKey string `yaml:"api_key" json:"api_key"`
	
	// CORS allowed origins (defaults to ["*"])
	CORSOrigins []string `yaml:"cors_origins,omitempty" json:"cors_origins,omitempty"`
	
	// Enable OpenAPI documentation endpoint at /docs (defaults to true)
	EnableDocs bool `yaml:"enable_docs,omitempty" json:"enable_docs,omitempty"`
	
	// Base path for all endpoints (e.g., "/api/v1")
	BasePath string `yaml:"base_path,omitempty" json:"base_path,omitempty"`
	
	// TLS configuration (optional)
	TLS *TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// TLSConfig defines TLS/HTTPS configuration
type TLSConfig struct {
	// Path to certificate file
	CertFile string `yaml:"cert_file" json:"cert_file"`
	
	// Path to private key file
	KeyFile string `yaml:"key_file" json:"key_file"`
}

// Validate validates the RunAs configuration
func (c *RunAsConfig) Validate() error {
	// Check runas type
	if c.RunAsType == "" {
		return fmt.Errorf("runas_type is required")
	}
	
	// Validate supported types
	if c.RunAsType != RunAsTypeMCP && 
	   c.RunAsType != RunAsTypeMCPSkills && 
	   c.RunAsType != RunAsTypeProxy && 
	   c.RunAsType != RunAsTypeProxySkills {
		return fmt.Errorf("unsupported runas_type: %s (supported: 'mcp', 'mcp-skills', 'proxy', 'proxy-skills')", c.RunAsType)
	}
	
	// Check version
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	// Validate server info (optional for proxy types)
	if c.RunAsType == RunAsTypeMCP || c.RunAsType == RunAsTypeMCPSkills {
		if err := c.ServerInfo.Validate(); err != nil {
			return fmt.Errorf("invalid server_info: %w", err)
		}
	} else {
		// For proxy types, server_info is optional
		// If provided, validate it, but don't require it
		if c.ServerInfo.Name != "" || c.ServerInfo.Version != "" {
			if err := c.ServerInfo.Validate(); err != nil {
				return fmt.Errorf("invalid server_info: %w", err)
			}
		}
	}
	
	// Type-specific validation
	if c.RunAsType == RunAsTypeMCP {
		// MCP type requires either tools array or templates array
		if len(c.Tools) == 0 && len(c.Templates) == 0 {
			return fmt.Errorf("runas_type 'mcp' requires at least one tool or template")
		}
		
		// Validate templates if provided
		for i, template := range c.Templates {
			if template.ConfigSource == "" {
				return fmt.Errorf("template at index %d missing config_source", i)
			}
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
	} else if c.RunAsType == RunAsTypeMCPSkills {
		// MCP-Skills type - tools auto-generated from skills directory
		// No tools array validation needed
		// SkillsConfig is optional (uses defaults if not provided)
	} else if c.RunAsType == RunAsTypeProxy {
		// Proxy type requires proxy config
		if c.ProxyConfig == nil {
			return fmt.Errorf("runas_type 'proxy' requires proxy_config")
		}
		
		if err := c.ProxyConfig.Validate(); err != nil {
			return fmt.Errorf("invalid proxy_config: %w", err)
		}
		
		// Must have one of: ConfigSource, Server, Expose, or Tools
		// ConfigSource is the preferred explicit method
		hasConfigSource := c.ConfigSource != ""
		hasServer := c.Server != ""
		hasExpose := len(c.Expose) > 0
		hasTools := len(c.Tools) > 0
		
		// Allow empty if all are empty - will be inferred from filename in serve command
		// This enables minimal configs where the server name is inferred from the filename
		if !hasConfigSource && !hasServer && !hasExpose && !hasTools {
			// This is OK - inference will happen in serve command
			// Just ensure proxy_config is valid
		}
		
		// ConfigSource takes precedence - cannot combine with Server or Expose
		if hasConfigSource && (hasServer || hasExpose) {
			return fmt.Errorf("cannot use 'config_source' with 'server' or 'expose' - config_source is explicit")
		}
		
		// Cannot have both Server and Expose (legacy validation)
		if hasServer && hasExpose {
			return fmt.Errorf("cannot use both 'server' and 'expose' - choose one")
		}
		
		// Validate tools if provided (for backward compatibility)
		if hasTools {
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
		}
		
		// Note: Expose field validation happens at runtime when we have access to appConfig
		// to detect whether items are servers or templates
	} else if c.RunAsType == RunAsTypeProxySkills {
		// Proxy-Skills type - tools auto-generated, requires proxy config
		if c.ProxyConfig == nil {
			return fmt.Errorf("runas_type 'proxy-skills' requires proxy_config")
		}
		
		if err := c.ProxyConfig.Validate(); err != nil {
			return fmt.Errorf("invalid proxy_config: %w", err)
		}
		
		// SkillsConfig is optional (uses defaults if not provided)
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
	// Must have either template or MCP server+tool
	hasTemplate := t.Template != ""
	hasMCPTool := t.MCPServer != "" && t.MCPTool != ""
	
	if !hasTemplate && !hasMCPTool {
		return fmt.Errorf("tool must specify either 'template' or both 'mcp_server' and 'mcp_tool'")
	}
	
	if hasTemplate && hasMCPTool {
		return fmt.Errorf("tool cannot specify both 'template' and 'mcp_server/mcp_tool' - choose one")
	}
	
	if t.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	
	// Description is optional - will be auto-generated if not provided
	// InputSchema is optional - will be auto-generated if not provided
	
	// If InputSchema is provided, validate it
	if t.InputSchema != nil {
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

// GetSkillsDirectory returns the skills directory with default fallback
func (c *RunAsConfig) GetSkillsDirectory(configDir string) string {
	if c.SkillsConfig != nil && c.SkillsConfig.SkillsDirectory != "" {
		return c.SkillsConfig.SkillsDirectory
	}
	// Default: config/skills relative to config directory
	return filepath.Join(configDir, "skills")
}

// ShouldIncludeSkill checks if a skill should be included based on filters
func (c *RunAsConfig) ShouldIncludeSkill(skillName string) bool {
	if c.SkillsConfig == nil {
		return true
	}
	
	// Check exclude list first
	for _, excluded := range c.SkillsConfig.ExcludeSkills {
		if excluded == skillName {
			return false
		}
	}
	
	// If include list is specified, skill must be in it
	if len(c.SkillsConfig.IncludeSkills) > 0 {
		for _, included := range c.SkillsConfig.IncludeSkills {
			if included == skillName {
				return true
			}
		}
		return false
	}
	
	return true
}

// Validate validates the ProxyConfig
func (p *ProxyConfig) Validate() error {
	if p.APIKey == "" {
		return fmt.Errorf("api_key is required for proxy types. Use a direct value or environment variable like ${MCP_PROXY_API_KEY}")
	}
	
	// Set defaults
	if p.Port == 0 {
		p.Port = 8080
	}
	
	if p.Host == "" {
		p.Host = "0.0.0.0"
	}
	
	if p.CORSOrigins == nil {
		p.CORSOrigins = []string{"*"}
	}
	
	// Validate port range
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", p.Port)
	}
	
	// Validate TLS config if provided
	if p.TLS != nil {
		if p.TLS.CertFile == "" {
			return fmt.Errorf("tls.cert_file is required when TLS is enabled")
		}
		if p.TLS.KeyFile == "" {
			return fmt.Errorf("tls.key_file is required when TLS is enabled")
		}
	}
	
	return nil
}
