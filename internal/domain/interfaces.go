package domain

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
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
	Messages     []Message `json:"messages"`
	Tools        []Tool    `json:"tools,omitempty"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	Temperature  float64   `json:"temperature,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	Stream       bool      `json:"stream,omitempty"`
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

// EMBEDDING TYPES

// EmbeddingRequest represents a request for vector embeddings
type EmbeddingRequest struct {
	Input          []string `json:"input"`                     // Text chunks to embed
	Model          string   `json:"model"`                     // Embedding model
	EncodingFormat string   `json:"encoding_format,omitempty"` // "float" or "base64"
	Dimensions     int      `json:"dimensions,omitempty"`      // For models that support it
	User           string   `json:"user,omitempty"`            // User identifier
}

// EmbeddingResponse represents the response from embedding API
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding vector
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingJob represents a complete embedding job with metadata
type EmbeddingJob struct {
	ID         string                 `json:"id"`
	Input      string                 `json:"input"`
	Chunks     []TextChunk            `json:"chunks"`
	Embeddings []EmbeddingWithMeta    `json:"embeddings"`
	Model      string                 `json:"model"`
	Provider   string                 `json:"provider"`
	StartTime  time.Time              `json:"start_time"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TextChunk represents a chunk of text with metadata (alias for compatibility)
type TextChunk = Chunk

// Chunk represents a chunk of text with metadata
type Chunk struct {
	Text       string `json:"text"`
	Index      int    `json:"index"`
	StartPos   int    `json:"start_pos"`
	EndPos     int    `json:"end_pos"`
	TokenCount int    `json:"token_count"`
}

// EmbeddingWithMeta combines embedding vector with chunk metadata
type EmbeddingWithMeta struct {
	Vector   []float32              `json:"vector"`
	Chunk    Chunk                  `json:"chunk"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ChunkingStrategy defines interface for text chunking strategies
type ChunkingStrategy interface {
	ChunkText(text string, maxTokens int) ([]Chunk, error)
	GetName() string
	GetDescription() string
}

// ChunkingType represents different chunking strategies
type ChunkingType string

const (
	ChunkingSentence  ChunkingType = "sentence"
	ChunkingParagraph ChunkingType = "paragraph"
	ChunkingFixed     ChunkingType = "fixed"
	ChunkingSemantic  ChunkingType = "semantic"
	ChunkingSliding   ChunkingType = "sliding"
)

// EmbeddingJobRequest represents a request to generate embeddings for text
type EmbeddingJobRequest struct {
	Input          string                 `json:"input"`
	Provider       string                 `json:"provider,omitempty"`
	Model          string                 `json:"model,omitempty"`
	ChunkStrategy  ChunkingType           `json:"chunk_strategy,omitempty"`
	MaxChunkSize   int                    `json:"max_chunk_size,omitempty"`
	ChunkOverlap   int                    `json:"chunk_overlap,omitempty"`
	EncodingFormat string                 `json:"encoding_format,omitempty"`
	Dimensions     int                    `json:"dimensions,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
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
	ProviderLMStudio   ProviderType = "lmstudio"
)

// LLMProvider defines the interface for interacting with Language Model providers
type LLMProvider interface {
	// CreateCompletion generates a completion using the specified request
	CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// StreamCompletion generates a streaming completion
	StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)

	// CreateEmbeddings generates vector embeddings for the given input
	CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetSupportedEmbeddingModels returns a list of embedding models supported by this provider
	GetSupportedEmbeddingModels() []string

	// GetMaxEmbeddingTokens returns the maximum token limit for embeddings for the given model
	GetMaxEmbeddingTokens(model string) int

	// GetProviderType returns the type of this provider
	GetProviderType() ProviderType

	// GetInterfaceType returns the interface type of this provider
	GetInterfaceType() config.InterfaceType

	// ValidateConfig validates the provider configuration
	ValidateConfig() error

	// Close cleans up provider resources
	Close() error
}

// EmbeddingModelConfig is imported from config package
// EmbeddingProviderConfig is imported from config package  
// EmbeddingsConfig is imported from config package
// ProviderConfig is imported from config package
// ServerConfig is imported from config package
// AIConfig is imported from config package
// ApplicationConfig is imported from config package
// InterfaceType and constants are imported from config package

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
	GetConfig() *config.ServerConfig
}

// MCPServerManager defines the interface for managing MCP servers
type MCPServerManager interface {
	// StartServer starts an MCP server
	StartServer(ctx context.Context, serverName string, cfg *config.ServerConfig) (MCPServer, error)

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
	Query           string  `json:"query"`
	ServerName      string  `json:"server_name,omitempty"`
	ProviderName    string  `json:"provider_name,omitempty"`
	Model           string  `json:"model,omitempty"`
	SystemPrompt    string  `json:"system_prompt,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	MaxTokens       int     `json:"max_tokens,omitempty"`
	MaxToolFollowUp int     `json:"max_tool_follow_up,omitempty"`
	OutputFormat    string  `json:"output_format,omitempty"`
	OutputFile      string  `json:"output_file,omitempty"`
	ContextFile     string  `json:"context_file,omitempty"`
	Stream          bool    `json:"stream,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	Response    string                 `json:"response"`
	ToolCalls   []ToolCall             `json:"tool_calls,omitempty"`
	ToolResults map[string]string      `json:"tool_results,omitempty"`
	Usage       *Usage                 `json:"usage,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Provider    string                 `json:"provider,omitempty"`
	Server      string                 `json:"server,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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
	ServerName      string  `json:"server_name,omitempty"`
	ProviderName    string  `json:"provider_name,omitempty"`
	Model           string  `json:"model,omitempty"`
	SystemPrompt    string  `json:"system_prompt,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	MaxTokens       int     `json:"max_tokens,omitempty"`
	MaxToolFollowUp int     `json:"max_tool_follow_up,omitempty"`
	HistoryLimit    int     `json:"history_limit,omitempty"`
}

// ProviderFactory defines the interface for creating LLM providers
type ProviderFactory interface {
	// CreateProvider creates a new provider instance
	CreateProvider(providerType ProviderType, cfg *config.ProviderConfig) (LLMProvider, error)

	// GetSupportedProviders returns a list of supported provider types
	GetSupportedProviders() []ProviderType

	// GetProviderInterface returns the interface type for a provider
	GetProviderInterface(providerType ProviderType) config.InterfaceType
}

// EmbeddingService defines the interface for embedding operations
type EmbeddingService interface {
	// GenerateEmbeddings processes text input and returns embeddings
	GenerateEmbeddings(ctx context.Context, req *EmbeddingJobRequest) (*EmbeddingJob, error)

	// GetAvailableChunkingStrategies returns available chunking strategies
	GetAvailableChunkingStrategies() []ChunkingType

	// ValidateEmbeddingRequest validates an embedding request
	ValidateEmbeddingRequest(req *EmbeddingJobRequest) error
}

// ConfigurationService defines the interface for configuration management
type ConfigurationService interface {
	// LoadConfig loads configuration from a file
	LoadConfig(filePath string) (*config.ApplicationConfig, error)

	// SaveConfig saves configuration to a file
	SaveConfig(config *config.ApplicationConfig, filePath string) error

	// GetProvider creates and returns a provider instance
	GetProvider(providerName string) (LLMProvider, error)

	// GetProviderConfig retrieves provider configuration
	GetProviderConfig(providerName string) (*config.ProviderConfig, config.InterfaceType, error)

	// GetEmbeddingProviderConfig retrieves embedding provider configuration
	GetEmbeddingProviderConfig(providerName string) (*config.EmbeddingProviderConfig, config.InterfaceType, error)

	// GetServerConfig retrieves server configuration
	GetServerConfig(serverName string) (*config.ServerConfig, error)

	// GetDefaultProvider returns the default provider configuration
	GetDefaultProvider() (string, *config.ProviderConfig, config.InterfaceType, error)

	// GetDefaultEmbeddingProvider returns the default embedding provider configuration
	GetDefaultEmbeddingProvider() (string, *config.EmbeddingProviderConfig, config.InterfaceType, error)

	// ListServers returns a list of configured server names
	ListServers() []string

	// ListEmbeddingProviders returns a list of configured embedding provider names
	ListEmbeddingProviders() []string

	// ValidateConfig validates the entire configuration
	ValidateConfig(config *config.ApplicationConfig) error
}
