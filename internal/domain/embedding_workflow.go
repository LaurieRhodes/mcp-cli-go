// File: internal/domain/embedding_workflow.go
package domain

import (
	"context"
	"fmt"
	"time"
)

// EmbeddingConfig defines how embeddings should be generated and used in workflow steps
type EmbeddingConfig struct {
	// Generation settings
	Generate bool   `json:"generate,omitempty"` // Whether to generate embeddings
	Input    string `json:"input,omitempty"`    // Text to embed: "{{stdin}}", "{{step1_output}}", etc.

	// Provider settings (falls back to config defaults if not specified)
	Provider string `json:"provider,omitempty"` // openai, gemini, deepseek, etc.
	Model    string `json:"model,omitempty"`    // text-embedding-3-small, etc.

	// Chunking settings
	ChunkStrategy string `json:"chunk_strategy,omitempty"` // sentence, paragraph, fixed
	MaxChunkSize  int    `json:"max_chunk_size,omitempty"` // Default from config
	Overlap       int    `json:"overlap,omitempty"`        // Chunk overlap

	// Usage configuration - HOW the embedding will be used (not sent to LLM)
	Usage EmbeddingUsage `json:"usage"`

	// Variable management
	StoreAs string `json:"store_as,omitempty"` // Variable name for metadata
}

// EmbeddingUsage defines how embeddings are used (vectors never sent to LLM)
type EmbeddingUsage struct {
	Type EmbeddingUsageType `json:"type"`

	// Configuration for different usage types
	VectorSearch *VectorSearchConfig `json:"vector_search,omitempty"`
	Similarity   *SimilarityConfig   `json:"similarity,omitempty"`
	Clustering   *ClusteringConfig   `json:"clustering,omitempty"`
}

// EmbeddingUsageType defines the available embedding usage patterns
type EmbeddingUsageType string

const (
	EmbeddingUsageVectorSearch EmbeddingUsageType = "vector_search" // Search vector database
	EmbeddingUsageSimilarity   EmbeddingUsageType = "similarity"    // Compare with other embeddings
	EmbeddingUsageClustering   EmbeddingUsageType = "clustering"    // Group similar embeddings
	EmbeddingUsageStorage      EmbeddingUsageType = "storage"       // Just store for later use
)

// VectorSearchConfig configures database similarity search
type VectorSearchConfig struct {
	Server              string                 `json:"server"`               // MCP server name
	Table               string                 `json:"table"`                // Database table name
	VectorColumn        string                 `json:"vector_column"`        // Column containing vectors
	TextColumns         []string               `json:"text_columns"`         // Columns to return in results
	SimilarityThreshold float64                `json:"similarity_threshold"` // Minimum similarity (0.0-1.0)
	MaxResults          int                    `json:"max_results"`          // Maximum number of results
	Filters             map[string]interface{} `json:"filters,omitempty"`    // Additional WHERE conditions
	OrderBy             string                 `json:"order_by,omitempty"`   // Custom ordering
}

// SimilarityConfig configures embedding similarity comparison
type SimilarityConfig struct {
	CompareWith  []string `json:"compare_with"`            // Variable names of other embeddings
	Method       string   `json:"method"`                  // cosine, euclidean, dot_product
	Threshold    float64  `json:"threshold"`               // Similarity threshold
	StoreResults bool     `json:"store_results,omitempty"` // Store detailed comparison results
}

// ClusteringConfig configures embedding clustering
type ClusteringConfig struct {
	Method      string  `json:"method"`                 // kmeans, hierarchical, dbscan
	NumClusters int     `json:"num_clusters,omitempty"` // For k-means
	Threshold   float64 `json:"threshold,omitempty"`    // For hierarchical/dbscan
	MinSamples  int     `json:"min_samples,omitempty"`  // For dbscan
}

// EmbeddingResult contains embedding metadata WITHOUT raw vectors (for LLM consumption)
type EmbeddingResult struct {
	ID          string    `json:"id"`
	Input       string    `json:"input"`        // Original input text
	Provider    string    `json:"provider"`     // Provider used
	Model       string    `json:"model"`        // Model used
	Dimensions  int       `json:"dimensions"`   // Vector dimensions
	ChunkCount  int       `json:"chunk_count"`  // Number of chunks created
	TokenCount  int       `json:"token_count"`  // Total tokens processed
	GeneratedAt time.Time `json:"generated_at"` // When generated

	// Metadata for LLM consumption (NO raw vectors)
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Private fields - NEVER exposed to LLM or JSON serialization
	vectors [][]float32      `json:"-"` // Raw vectors stored privately
	chunks  []EmbeddingChunk `json:"-"` // Chunk details stored privately
}

// EmbeddingChunk contains chunk details WITHOUT exposing vectors
type EmbeddingChunk struct {
	Text       string    `json:"text"`
	Index      int       `json:"index"`
	StartPos   int       `json:"start_pos"`
	EndPos     int       `json:"end_pos"`
	TokenCount int       `json:"token_count"`
	Vector     []float32 `json:"-"` // NEVER serialized - stored privately
}

// GetVectors returns the private vectors (for internal use only)
func (er *EmbeddingResult) GetVectors() [][]float32 {
	return er.vectors
}

// GetChunks returns the private chunks (for internal use only)
func (er *EmbeddingResult) GetChunks() []EmbeddingChunk {
	return er.chunks
}

// SetVectors sets the private vectors (for internal use only)
func (er *EmbeddingResult) SetVectors(vectors [][]float32) {
	er.vectors = vectors
}

// SetChunks sets the private chunks (for internal use only)
func (er *EmbeddingResult) SetChunks(chunks []EmbeddingChunk) {
	er.chunks = chunks
}

// GetMetadata returns only the safe metadata for LLM consumption
func (er *EmbeddingResult) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"id":           er.ID,
		"provider":     er.Provider,
		"model":        er.Model,
		"dimensions":   er.Dimensions,
		"chunk_count":  er.ChunkCount,
		"token_count":  er.TokenCount,
		"generated_at": er.GeneratedAt,
		"metadata":     er.Metadata,
	}
}

// EstimateMemoryUsage calculates approximate memory usage in bytes
func (er *EmbeddingResult) EstimateMemoryUsage() int64 {
	if len(er.vectors) == 0 {
		return 0
	}

	// Each float32 = 4 bytes
	vectorBytes := int64(len(er.vectors) * len(er.vectors[0]) * 4)

	// Estimate text and metadata overhead
	textBytes := int64(len(er.Input))
	for _, chunk := range er.chunks {
		textBytes += int64(len(chunk.Text))
	}

	// Add metadata overhead estimate
	metadataBytes := int64(500) // Conservative estimate

	return vectorBytes + textBytes + metadataBytes
}

// Validation methods
func (ec *EmbeddingConfig) Validate() error {
	if ec.Generate && ec.Input == "" {
		return NewDomainError(ErrCodeRequestInvalid, "embedding input is required when generate=true")
	}

	if ec.Usage.Type == "" {
		return NewDomainError(ErrCodeRequestInvalid, "embedding usage type is required")
	}

	return ec.Usage.Validate()
}

func (eu *EmbeddingUsage) Validate() error {
	switch eu.Type {
	case EmbeddingUsageVectorSearch:
		if eu.VectorSearch == nil {
			return NewDomainError(ErrCodeRequestInvalid, "vector_search config required for vector_search usage")
		}
		return eu.VectorSearch.Validate()

	case EmbeddingUsageSimilarity:
		if eu.Similarity == nil {
			return NewDomainError(ErrCodeRequestInvalid, "similarity config required for similarity usage")
		}
		return eu.Similarity.Validate()

	case EmbeddingUsageClustering:
		if eu.Clustering == nil {
			return NewDomainError(ErrCodeRequestInvalid, "clustering config required for clustering usage")
		}
		return eu.Clustering.Validate()

	case EmbeddingUsageStorage:
		// Storage usage requires no additional config
		return nil

	default:
		return NewDomainError(ErrCodeRequestInvalid, fmt.Sprintf("unsupported embedding usage type: %s", eu.Type))
	}
}

func (vsc *VectorSearchConfig) Validate() error {
	if vsc.Server == "" {
		return NewDomainError(ErrCodeRequestInvalid, "server is required for vector search")
	}

	if vsc.Table == "" {
		return NewDomainError(ErrCodeRequestInvalid, "table is required for vector search")
	}

	if vsc.VectorColumn == "" {
		return NewDomainError(ErrCodeRequestInvalid, "vector_column is required for vector search")
	}

	if len(vsc.TextColumns) == 0 {
		return NewDomainError(ErrCodeRequestInvalid, "at least one text_column is required for vector search")
	}

	if vsc.SimilarityThreshold < 0 || vsc.SimilarityThreshold > 1 {
		return NewDomainError(ErrCodeRequestInvalid, "similarity_threshold must be between 0 and 1")
	}

	if vsc.MaxResults <= 0 {
		return NewDomainError(ErrCodeRequestInvalid, "max_results must be greater than 0")
	}

	return nil
}

func (sc *SimilarityConfig) Validate() error {
	if len(sc.CompareWith) == 0 {
		return NewDomainError(ErrCodeRequestInvalid, "compare_with must contain at least one embedding variable")
	}

	validMethods := []string{"cosine", "euclidean", "dot_product"}
	if sc.Method == "" {
		sc.Method = "cosine" // Default
	} else {
		found := false
		for _, valid := range validMethods {
			if sc.Method == valid {
				found = true
				break
			}
		}
		if !found {
			return NewDomainError(ErrCodeRequestInvalid, fmt.Sprintf("similarity method must be one of: %v", validMethods))
		}
	}

	return nil
}

func (cc *ClusteringConfig) Validate() error {
	validMethods := []string{"kmeans", "hierarchical", "dbscan"}
	if cc.Method == "" {
		return NewDomainError(ErrCodeRequestInvalid, "clustering method is required")
	}

	found := false
	for _, valid := range validMethods {
		if cc.Method == valid {
			found = true
			break
		}
	}
	if !found {
		return NewDomainError(ErrCodeRequestInvalid, fmt.Sprintf("clustering method must be one of: %v", validMethods))
	}

	switch cc.Method {
	case "kmeans":
		if cc.NumClusters <= 0 {
			return NewDomainError(ErrCodeRequestInvalid, "num_clusters must be greater than 0 for kmeans")
		}
	case "dbscan":
		if cc.MinSamples <= 0 {
			return NewDomainError(ErrCodeRequestInvalid, "min_samples must be greater than 0 for dbscan")
		}
	}

	return nil
}

// Extended interfaces for embedding support
type EmbeddingWorkflowProcessor interface {
	WorkflowProcessor

	// Generate embeddings for a workflow step
	GenerateStepEmbedding(ctx context.Context, step *WorkflowStep, input string, variables map[string]interface{}) (*EmbeddingResult, error)

	// Execute embedding usage (vector search, similarity, etc.)
	ExecuteEmbeddingUsage(ctx context.Context, embedding *EmbeddingResult, usage EmbeddingUsage, variables map[string]interface{}) (interface{}, error)

	// Manage embedding memory
	GetEmbeddingFromMemory(id string) (*EmbeddingResult, bool)
	StoreEmbeddingInMemory(embedding *EmbeddingResult) error
	ClearEmbeddingMemory()
}
