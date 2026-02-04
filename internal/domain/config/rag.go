package config

// RagConfig represents the RAG configuration (loaded from config/rag/*.yaml)
type RagConfig struct {
	DefaultServer  string                     `yaml:"default_server,omitempty"`
	DefaultFusion  string                     `yaml:"default_fusion,omitempty"`
	DefaultTopK    int                        `yaml:"default_top_k,omitempty"`
	Servers        map[string]RagServerConfig `yaml:"servers,omitempty"`
	QueryExpansion QueryExpansionSettings     `yaml:"query_expansion,omitempty"`
	Fusion         FusionSettings             `yaml:"fusion,omitempty"`
}

// RagServerConfig defines configuration for a RAG-enabled MCP server
type RagServerConfig struct {
	ServerName      string                `yaml:"server_name"`                // Name of this RAG config
	MCPServer       string                `yaml:"mcp_server"`                 // Name of MCP server from servers config
	SearchTool      string                `yaml:"search_tool,omitempty"`      // Optional: specific tool name
	Strategies      []StrategyConfig      `yaml:"strategies"`                 // Vector column strategies
	Table           string                `yaml:"table"`                      // Table/collection name
	TextColumns     []string              `yaml:"text_columns"`               // Columns to return
	MetadataColumns []string              `yaml:"metadata_columns,omitempty"` // Metadata columns
	QueryEmbedding  *QueryEmbeddingConfig `yaml:"query_embedding,omitempty"`  // Default embedding config for queries
}

// QueryEmbeddingConfig defines how to generate query embeddings
type QueryEmbeddingConfig struct {
	Type          string                 `yaml:"type"`                     // "mcp_tool" or "service"
	ToolName      string                 `yaml:"tool_name,omitempty"`      // MCP tool name (if type=mcp_tool)
	DefaultParams map[string]interface{} `yaml:"default_params,omitempty"` // Parameters for tool/service
	Provider      string                 `yaml:"provider,omitempty"`       // Provider name (if type=service)
	Model         string                 `yaml:"model,omitempty"`          // Model name (if type=service)
}

// StrategyConfig defines a vector search strategy (column)
type StrategyConfig struct {
	Name           string                 `yaml:"name"`                      // Strategy name (e.g., "default", "technical")
	VectorColumn   string                 `yaml:"vector_column"`             // Name of vector column
	Weight         float64                `yaml:"weight"`                    // Weight for fusion (default: 1.0)
	Threshold      float64                `yaml:"threshold"`                 // Similarity threshold (default: 0.7)
	MaxResults     int                    `yaml:"max_results,omitempty"`     // Max results for this strategy
	Filters        map[string]interface{} `yaml:"filters,omitempty"`         // Optional filters
	QueryEmbedding *QueryEmbeddingConfig  `yaml:"query_embedding,omitempty"` // Override embedding config for this strategy
}

// QueryExpansionSettings defines query expansion configuration
type QueryExpansionSettings struct {
	Enabled       bool   `yaml:"enabled"`
	SynonymsFile  string `yaml:"synonyms_file,omitempty"`
	AcronymsFile  string `yaml:"acronyms_file,omitempty"`
	MaxExpansions int    `yaml:"max_expansions,omitempty"`
	CaseSensitive bool   `yaml:"case_sensitive,omitempty"`
}

// FusionSettings defines result fusion configuration
type FusionSettings struct {
	RRF      RRFSettings      `yaml:"rrf,omitempty"`
	Weighted WeightedSettings `yaml:"weighted,omitempty"`
}

// RRFSettings defines Reciprocal Rank Fusion parameters
type RRFSettings struct {
	K int `yaml:"k"` // RRF constant (default: 60)
}

// WeightedSettings defines weighted fusion parameters
type WeightedSettings struct {
	Normalize bool `yaml:"normalize"` // Normalize scores before fusion
}
