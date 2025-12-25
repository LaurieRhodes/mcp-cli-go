package models

import "testing"

func TestChunkingStrategyIsValid(t *testing.T) {
	tests := []struct {
		name     string
		strategy ChunkingStrategy
		valid    bool
	}{
		{"sentence", ChunkingSentence, true},
		{"paragraph", ChunkingParagraph, true},
		{"fixed", ChunkingFixed, true},
		{"semantic", ChunkingSemantic, true},
		{"sliding", ChunkingSliding, true},
		{"invalid", ChunkingStrategy("invalid"), false},
		{"empty", ChunkingStrategy(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.strategy.IsValid(); got != tt.valid {
				t.Errorf("ChunkingStrategy.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestEmbeddingRequestCreation(t *testing.T) {
	req := &EmbeddingRequest{
		Input:          []string{"text1", "text2"},
		Model:          "text-embedding-ada-002",
		EncodingFormat: "float",
		Dimensions:     1536,
	}

	if len(req.Input) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(req.Input))
	}

	if req.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model 'text-embedding-ada-002', got %s", req.Model)
	}

	if req.Dimensions != 1536 {
		t.Errorf("Expected 1536 dimensions, got %d", req.Dimensions)
	}
}

func TestEmbeddingResponseCreation(t *testing.T) {
	resp := &EmbeddingResponse{
		Object: "list",
		Data: []Embedding{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: []float32{0.1, 0.2, 0.3},
			},
		},
		Model: "text-embedding-ada-002",
		Usage: Usage{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}

	if resp.Object != "list" {
		t.Errorf("Expected object 'list', got %s", resp.Object)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 embedding, got %d", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(resp.Data[0].Embedding))
	}
}

func TestTextChunkCreation(t *testing.T) {
	chunk := TextChunk{
		Text:       "This is a test chunk",
		Index:      0,
		StartPos:   0,
		EndPos:     20,
		TokenCount: 5,
	}

	if chunk.Text != "This is a test chunk" {
		t.Errorf("Unexpected text: %s", chunk.Text)
	}

	if chunk.TokenCount != 5 {
		t.Errorf("Expected 5 tokens, got %d", chunk.TokenCount)
	}
}

func TestEmbeddingJobCreation(t *testing.T) {
	job := &EmbeddingJob{
		ID:       "job-123",
		Input:    "Test input",
		Model:    "test-model",
		Provider: "test-provider",
		Chunks: []TextChunk{
			{Text: "chunk1", Index: 0},
			{Text: "chunk2", Index: 1},
		},
	}

	if job.ID != "job-123" {
		t.Errorf("Expected ID 'job-123', got %s", job.ID)
	}

	if len(job.Chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(job.Chunks))
	}

	if job.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", job.Provider)
	}
}
