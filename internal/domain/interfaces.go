package domain

import (
	"context"
	"encoding/json"
	"io"
)

// Message represents a message in a conversation
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a call to a tool
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function call within a tool call
type Function struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool defines a tool that can be used by the LLM
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction defines the function specification for a tool
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CompletionRequest contains the request parameters for LLM completion
type CompletionRequest struct {
	Messages      []Message `json:"messages"`
	Tools         []Tool    `json:"tools,omitempty"`
	SystemPrompt  string    `json:"system_prompt,omitempty"`
	Temperature   float64   `json:"temperature,omitempty"`
	MaxTokens     int       `json:"max_tokens,omitempty"`
	Stream        bool      `json:"stream,omitempty"`
}

// CompletionResponse contains the response from an LLM completion
type CompletionResponse struct {
	Response  string     `json:"response"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     *Usage     `json:"usage,omitempty"`
	Model     string     `json:"model,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
}

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI     ProviderType = "openai"
	ProviderAnthropic  ProviderType = "anthropic"
	ProviderOllama     ProviderType = "ollama"
	ProviderDeepSeek   ProviderType = "deepseek"
	ProviderGemini     ProviderType = "gemini"
	ProviderOpenRouter ProviderType = "openrouter"
)

// InterfaceType represents the API interface type that a provider uses
type InterfaceType string

const (
	OpenAICompatible InterfaceType = "openai_compatible"
	AnthropicNative  InterfaceType = "anthropic_native"
	OllamaNative     InterfaceType = "ollama_native"
	GeminiNative     InterfaceType = "gemini_native"
)

// LLMProvider defines the interface for interacting with Language Model providers
type LLMProvider interface {
	// CreateCompletion generates a completion using the specified request
	CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	
	// StreamCompletion generates a streaming completion
	StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)
	
	// GetProviderType returns the type of this provider
	GetProviderType() ProviderType
	
	// GetInterfaceType returns the interface type of this provider
	GetInterfaceType() InterfaceType
	
	// ValidateConfig validates the provider configuration
	ValidateConfig() error
	
	// Close cleans up provider resources
	Close() error
}

// ProviderConfig represents configuration for an LLM provider
type ProviderConfig struct {
	APIKey          string   `json:"api_key"`
	DefaultModel    string   `json:"default_model"`
	APIEndpoint     string   `json:"api_endpoint,omitempty"`
	AvailableModels []string `json:"available_models,omitempty"`
	TimeoutSeconds  int      `json:"timeout_seconds,omitempty"`
	MaxRetries      int      `json:"max_retries,omitempty"`
	Temperature     float64  `json:"temperature,omitempty"`
	MaxTokens       int      `json:"max_tokens,omitempty"`
}

// ServerConfig represents configuration for an MCP server
type ServerConfig struct {
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

// AIConfig represents the AI configuration
type AIConfig struct {
	DefaultProvider     string                            `json:"default_provider"`
	DefaultSystemPrompt string                            `json:"default_system_prompt,omitempty"`
	Interfaces          map[InterfaceType]InterfaceConfig `json:"interfaces"`
	
	// Legacy fields for backward compatibility
	Providers map[string]ProviderConfig `json:"providers,omitempty"`
}

// InterfaceConfig represents configuration for an API interface
type InterfaceConfig struct {
	Providers map[string]ProviderConfig `json:"providers"`
}

// ApplicationConfig represents the complete application configuration
type ApplicationConfig struct {
	Servers   map[string]ServerConfig      `json:"servers"`
	AI        *AIConfig                    `json:"ai,omitempty"`
	Settings  *GlobalSettings              `json:"settings,omitempty"`
	Templates map[string]*WorkflowTemplate `json:"templates,omitempty"`     // Workflow templates
}

// ValidateWorkflowTemplates validates all workflow templates in the configuration
func (c *ApplicationConfig) ValidateWorkflowTemplates() error {
	if c.Templates == nil {
		return nil
	}
	
	for templateName, template := range c.Templates {
		if err := template.ValidateWorkflowTemplate(); err != nil {
			return NewDomainError(ErrCodeConfigInvalid, "invalid workflow template").
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

// GlobalSettings contains global application settings
type GlobalSettings struct {
	LogLevel        string `json:"log_level,omitempty"`
	MaxToolFollowUp int    `json:"max_tool_follow_up,omitempty"`
	StrictMode      bool   `json:"strict_mode,omitempty"`
}

// GetMaxToolFollowUp returns the max tool follow-up setting
func (g *GlobalSettings) GetMaxToolFollowUp() int {
	if g == nil {
		return 0
	}
	return g.MaxToolFollowUp
}

// ConfigurationService defines the interface for configuration management
type ConfigurationService interface {
	// LoadConfig loads configuration from a file
	LoadConfig(filePath string) (*ApplicationConfig, error)
	
	// SaveConfig saves configuration to a file
	SaveConfig(config *ApplicationConfig, filePath string) error
	
	// GetProviderConfig retrieves provider configuration
	GetProviderConfig(providerName string) (*ProviderConfig, InterfaceType, error)
	
	// GetServerConfig retrieves server configuration
	GetServerConfig(serverName string) (*ServerConfig, error)
	
	// GetDefaultProvider returns the default provider configuration
	GetDefaultProvider() (string, *ProviderConfig, InterfaceType, error)
	
	// ValidateConfig validates the entire configuration
	ValidateConfig(config *ApplicationConfig) error
	
	// MigrateConfig migrates configuration from legacy format
	MigrateConfig(legacyConfigPath string) (*ApplicationConfig, error)
}

// MCPServer represents an MCP server instance
type MCPServer interface {
	// Start starts the MCP server
	Start(ctx context.Context) error
	
	// Stop stops the MCP server
	Stop() error
	
	// IsRunning returns true if the server is running
	IsRunning() bool
	
	// GetTools returns available tools from this server
	GetTools() ([]Tool, error)
	
	// ExecuteTool executes a tool on this server
	ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error)
	
	// GetServerName returns the name of this server
	GetServerName() string
	
	// GetConfig returns the server configuration
	GetConfig() *ServerConfig
}

// MCPServerManager defines the interface for managing MCP servers
type MCPServerManager interface {
	// StartServer starts an MCP server
	StartServer(ctx context.Context, serverName string, config *ServerConfig) (MCPServer, error)
	
	// StopServer stops an MCP server
	StopServer(serverName string) error
	
	// GetServer retrieves a running server
	GetServer(serverName string) (MCPServer, bool)
	
	// ListServers returns all running servers
	ListServers() map[string]MCPServer
	
	// GetAvailableTools returns all available tools from all servers
	GetAvailableTools() ([]Tool, error)
	
	// ExecuteTool executes a tool on the appropriate server
	ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error)
	
	// StopAll stops all running servers
	StopAll() error
}

// QueryRequest represents a query request
type QueryRequest struct {
	Query            string  `json:"query"`
	ServerName       string  `json:"server_name,omitempty"`
	ProviderName     string  `json:"provider_name,omitempty"`
	Model            string  `json:"model,omitempty"`
	SystemPrompt     string  `json:"system_prompt,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	MaxToolFollowUp  int     `json:"max_tool_follow_up,omitempty"`
	OutputFormat     string  `json:"output_format,omitempty"`
	OutputFile       string  `json:"output_file,omitempty"`
	ContextFile      string  `json:"context_file,omitempty"`
	Stream           bool    `json:"stream,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	Response     string                 `json:"response"`
	ToolCalls    []ToolCall            `json:"tool_calls,omitempty"`
	ToolResults  map[string]string     `json:"tool_results,omitempty"`
	Usage        *Usage                `json:"usage,omitempty"`
	Model        string                `json:"model,omitempty"`
	Provider     string                `json:"provider,omitempty"`
	Server       string                `json:"server,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// QueryService defines the interface for query processing
type QueryService interface {
	// ProcessQuery processes a query request
	ProcessQuery(ctx context.Context, req *QueryRequest) (*QueryResponse, error)
	
	// ProcessQueryWithStreaming processes a query with streaming output
	ProcessQueryWithStreaming(ctx context.Context, req *QueryRequest, writer io.Writer) (*QueryResponse, error)
	
	// ValidateRequest validates a query request
	ValidateRequest(req *QueryRequest) error
}

// ChatSession represents a chat session
type ChatSession interface {
	// AddMessage adds a message to the session
	AddMessage(message *Message) error
	
	// GetMessages returns all messages in the session
	GetMessages() []*Message
	
	// ProcessMessage processes a user message and returns the response
	ProcessMessage(ctx context.Context, userMessage string) (*Message, error)
	
	// ProcessMessageWithStreaming processes a message with streaming
	ProcessMessageWithStreaming(ctx context.Context, userMessage string, writer io.Writer) (*Message, error)
	
	// GetSessionID returns the session ID
	GetSessionID() string
	
	// Clear clears the session history
	Clear()
	
	// Save saves the session to a file
	Save(filePath string) error
	
	// Load loads the session from a file
	Load(filePath string) error
}

// ChatService defines the interface for chat management
type ChatService interface {
	// CreateSession creates a new chat session
	CreateSession(sessionID string, config *ChatConfig) (ChatSession, error)
	
	// GetSession retrieves an existing chat session
	GetSession(sessionID string) (ChatSession, bool)
	
	// DeleteSession deletes a chat session
	DeleteSession(sessionID string) error
	
	// ListSessions returns all active session IDs
	ListSessions() []string
}

// ChatConfig represents chat configuration
type ChatConfig struct {
	ServerName       string  `json:"server_name,omitempty"`
	ProviderName     string  `json:"provider_name,omitempty"`
	Model            string  `json:"model,omitempty"`
	SystemPrompt     string  `json:"system_prompt,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	MaxToolFollowUp  int     `json:"max_tool_follow_up,omitempty"`
	HistoryLimit     int     `json:"history_limit,omitempty"`
}

// ProviderFactory defines the interface for creating LLM providers
type ProviderFactory interface {
	// CreateProvider creates a new provider instance
	CreateProvider(providerType ProviderType, config *ProviderConfig) (LLMProvider, error)
	
	// GetSupportedProviders returns a list of supported provider types
	GetSupportedProviders() []ProviderType
	
	// GetProviderInterface returns the interface type for a provider
	GetProviderInterface(providerType ProviderType) InterfaceType
}
