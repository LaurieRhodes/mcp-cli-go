package ports

import (
	"context"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// MCPServer represents a single MCP server instance
type MCPServer interface {
	// Lifecycle
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool

	// Tool operations
	GetTools() ([]models.Tool, error)
	ExecuteTool(ctx context.Context, name string, args map[string]any) (string, error)

	// Metadata
	GetName() string
	GetConfig() ServerConfig
}

// MCPManager manages multiple MCP servers
type MCPManager interface {
	// Server lifecycle
	StartServer(ctx context.Context, name string, config ServerConfig) (MCPServer, error)
	StopServer(name string) error
	StopAll() error

	// Server access
	GetServer(name string) (MCPServer, bool)
	ListServers() []string

	// Tool operations across servers
	GetAllTools() ([]models.Tool, error)
	ExecuteTool(ctx context.Context, toolName string, args map[string]any) (string, error)
}

// ServerConfig contains MCP server configuration
type ServerConfig struct {
	Name         string            `json:"name"`
	Command      string            `json:"command"`
	Args         []string          `json:"args"`
	Env          map[string]string `json:"env,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Settings     *ServerSettings   `json:"settings,omitempty"`
}

// ServerSettings contains server-specific settings
type ServerSettings struct {
	MaxToolFollowUp int  `json:"max_tool_follow_up,omitempty"`
	StrictMode      bool `json:"strict_mode,omitempty"`
}

// GetMaxToolFollowUp returns the max tool follow-up setting
func (s *ServerSettings) GetMaxToolFollowUp() int {
	if s == nil {
		return 0
	}
	return s.MaxToolFollowUp
}
