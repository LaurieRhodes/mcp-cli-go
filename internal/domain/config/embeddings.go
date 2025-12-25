package config

// ChunkingType represents different chunking strategies
type ChunkingType string

const (
	ChunkingSentence  ChunkingType = "sentence"
	ChunkingParagraph ChunkingType = "paragraph"
	ChunkingFixed     ChunkingType = "fixed"
	ChunkingSemantic  ChunkingType = "semantic"
	ChunkingSliding   ChunkingType = "sliding"
)

// EmbeddingsConfig represents the embeddings configuration section
type EmbeddingsConfig struct {
	DefaultProvider      string                                     `yaml:"default_provider,omitempty"`
	DefaultChunkStrategy ChunkingType                               `yaml:"default_chunk_strategy,omitempty"`
	DefaultMaxChunkSize  int                                        `yaml:"default_max_chunk_size,omitempty"`
	DefaultOverlap       int                                        `yaml:"default_overlap,omitempty"`
	OutputPrecision      int                                        `yaml:"output_precision,omitempty"`
	CacheEmbeddings      bool                                       `yaml:"cache_embeddings,omitempty"`
	MaxVectorMemoryMB    int                                        `yaml:"max_vector_memory_mb,omitempty"`
	CacheTTL             string                                     `yaml:"cache_ttl,omitempty"`
	Interfaces           map[InterfaceType]EmbeddingInterfaceConfig `yaml:"interfaces,omitempty"`
	Providers            map[string]EmbeddingProviderConfig         `yaml:"providers,omitempty"`
}

// EmbeddingInterfaceConfig represents configuration for an embedding interface
type EmbeddingInterfaceConfig struct {
	Providers map[string]EmbeddingProviderConfig `yaml:"providers"`
}

// EmbeddingProviderConfig represents configuration for an embedding provider
type EmbeddingProviderConfig struct {
	APIKey          string                          `yaml:"api_key"`
	APIEndpoint     string                          `yaml:"api_endpoint,omitempty"`
	DefaultModel    string                          `yaml:"default_model"`
	AvailableModels []string                        `yaml:"available_models,omitempty"`
	TimeoutSeconds  int                             `yaml:"timeout_seconds,omitempty"`
	MaxRetries      int                             `yaml:"max_retries,omitempty"`
	Models          map[string]EmbeddingModelConfig `yaml:"models,omitempty"`
}

// Note: EmbeddingModelConfig is defined in provider.go to avoid duplication
