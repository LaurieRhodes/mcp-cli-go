package models

import "time"

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	Dimensions     int      `json:"dimensions,omitempty"`
	User           string   `json:"user,omitempty"`
}

// EmbeddingResponse contains the generated embeddings
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single vector embedding
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingJob represents a complete embedding operation with metadata
type EmbeddingJob struct {
	ID         string              `json:"id"`
	Input      string              `json:"input"`
	Chunks     []TextChunk         `json:"chunks"`
	Embeddings []EmbeddingWithMeta `json:"embeddings"`
	Model      string              `json:"model"`
	Provider   string              `json:"provider"`
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	Metadata   map[string]any      `json:"metadata,omitempty"`
}

// TextChunk represents a chunk of text with position info
type TextChunk struct {
	Text       string `json:"text"`
	Index      int    `json:"index"`
	StartPos   int    `json:"start_pos"`
	EndPos     int    `json:"end_pos"`
	TokenCount int    `json:"token_count"`
}

// EmbeddingWithMeta combines an embedding with its source chunk
type EmbeddingWithMeta struct {
	Vector   []float32      `json:"vector"`
	Chunk    TextChunk      `json:"chunk"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ChunkingStrategy represents different text chunking approaches
type ChunkingStrategy string

const (
	ChunkingSentence  ChunkingStrategy = "sentence"
	ChunkingParagraph ChunkingStrategy = "paragraph"
	ChunkingFixed     ChunkingStrategy = "fixed"
	ChunkingSemantic  ChunkingStrategy = "semantic"
	ChunkingSliding   ChunkingStrategy = "sliding"
)

// IsValid checks if the chunking strategy is valid
func (cs ChunkingStrategy) IsValid() bool {
	switch cs {
	case ChunkingSentence, ChunkingParagraph, ChunkingFixed,
		ChunkingSemantic, ChunkingSliding:
		return true
	default:
		return false
	}
}
