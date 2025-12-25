package ports

import (
	"context"
	"io"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// LLMProvider defines the contract for LLM integrations
type LLMProvider interface {
	// Core operations
	CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	StreamCompletion(ctx context.Context, req *CompletionRequest, writer io.Writer) (*CompletionResponse, error)

	// Embeddings
	CreateEmbeddings(ctx context.Context, req *models.EmbeddingRequest) (*models.EmbeddingResponse, error)

	// Metadata
	GetProviderType() ProviderType
	GetSupportedModels() []string
	GetMaxTokens(model string) int

	// Lifecycle
	ValidateConfig() error
	Close() error
}

// CompletionRequest contains parameters for text generation
type CompletionRequest struct {
	Messages     []models.Message `json:"messages"`
	Tools        []models.Tool    `json:"tools,omitempty"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Model        string           `json:"model,omitempty"`
	Temperature  float64          `json:"temperature,omitempty"`
	MaxTokens    int              `json:"max_tokens,omitempty"`
	Stream       bool             `json:"stream,omitempty"`
}

// CompletionResponse contains the LLM's response
type CompletionResponse struct {
	Content   string              `json:"content"`
	ToolCalls []models.ToolCall   `json:"tool_calls,omitempty"`
	Usage     models.Usage        `json:"usage,omitempty"`
	Model     string              `json:"model"`
}

// ProviderType identifies the LLM provider
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

// ProviderFactory creates LLM provider instances
type ProviderFactory interface {
	Create(providerType ProviderType, config ProviderConfig) (LLMProvider, error)
	GetSupportedTypes() []ProviderType
}

// ProviderConfig contains provider-specific configuration
type ProviderConfig struct {
	APIKey        string   `json:"api_key"`
	APIEndpoint   string   `json:"api_endpoint,omitempty"`
	DefaultModel  string   `json:"default_model"`
	Models        []string `json:"models,omitempty"`
	Timeout       int      `json:"timeout_seconds,omitempty"`
	MaxRetries    int      `json:"max_retries,omitempty"`
	ContextWindow int      `json:"context_window,omitempty"`
	ReserveTokens int      `json:"reserve_tokens,omitempty"`
}
