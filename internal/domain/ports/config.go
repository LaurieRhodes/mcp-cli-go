package ports

import "github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"

// ConfigService manages application configuration
type ConfigService interface {
	// Loading
	Load(path string) (*AppConfig, error)
	LoadOrCreate(path string) (*AppConfig, bool, error)

	// Saving
	Save(config *AppConfig, path string) error

	// Providers
	GetProvider(name string) (ProviderConfig, error)
	GetDefaultProvider() (string, ProviderConfig, error)
	ListProviders() []string

	// Servers
	GetServer(name string) (ServerConfig, error)
	ListServers() []string

	// Validation
	Validate(config *AppConfig) error
}

// AppConfig represents the complete application configuration
type AppConfig struct {
	Servers    map[string]ServerConfig       `json:"servers"`
	Providers  map[string]ProviderConfig     `json:"providers"`
	Embeddings *EmbeddingsConfig             `json:"embeddings,omitempty"`
	// Workflows removed - use actual config.ApplicationConfig.Workflows (WorkflowV2)
}

// EmbeddingsConfig contains embedding-specific configuration
type EmbeddingsConfig struct {
	DefaultProvider      string                             `json:"default_provider,omitempty"`
	DefaultChunkStrategy models.ChunkingStrategy            `json:"default_chunk_strategy,omitempty"`
	DefaultMaxChunkSize  int                                `json:"default_max_chunk_size,omitempty"`
	DefaultOverlap       int                                `json:"default_overlap,omitempty"`
	Providers            map[string]EmbeddingProviderConfig `json:"providers,omitempty"`
}

// EmbeddingProviderConfig contains embedding provider configuration
type EmbeddingProviderConfig struct {
	APIKey       string   `json:"api_key"`
	APIEndpoint  string   `json:"api_endpoint,omitempty"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models,omitempty"`
}
