package embeddings

import (
	"context"
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// Service handles embedding generation
type Service struct {
	provider ports.LLMProvider
	config   ServiceConfig
}

// ServiceConfig contains embedding service configuration
type ServiceConfig struct {
	Model         string
	ChunkSize     int
	ChunkOverlap  int
	ChunkStrategy models.ChunkingStrategy
	BatchSize     int
}

// NewService creates a new embeddings service
func NewService(provider ports.LLMProvider, config ServiceConfig) *Service {
	// Set defaults
	if config.ChunkSize == 0 {
		config.ChunkSize = 512
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 50
	}
	if config.ChunkStrategy == "" {
		config.ChunkStrategy = models.ChunkingFixed
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}

	return &Service{
		provider: provider,
		config:   config,
	}
}

// GenerateEmbedding generates an embedding for a single text
func (s *Service) GenerateEmbedding(ctx context.Context, text string) (*models.Embedding, error) {
	req := &models.EmbeddingRequest{
		Input: []string{text},
		Model: s.config.Model,
	}

	resp, err := s.provider.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return &resp.Data[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *Service) GenerateEmbeddings(ctx context.Context, texts []string) ([]models.Embedding, error) {
	// Process in batches
	var allEmbeddings []models.Embedding

	for i := 0; i < len(texts); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]

		req := &models.EmbeddingRequest{
			Input: batch,
			Model: s.config.Model,
		}

		resp, err := s.provider.CreateEmbeddings(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("batch %d failed: %w", i/s.config.BatchSize, err)
		}

		allEmbeddings = append(allEmbeddings, resp.Data...)
	}

	return allEmbeddings, nil
}

// GenerateWithChunking generates embeddings with text chunking
func (s *Service) GenerateWithChunking(ctx context.Context, text string) (*models.EmbeddingJob, error) {
	// Chunk the text
	chunks := s.chunkText(text)

	// Generate embeddings for chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Text
	}

	embeddings, err := s.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return nil, err
	}

	// Build embedding job
	job := &models.EmbeddingJob{
		Input:      text,
		Chunks:     chunks,
		Embeddings: make([]models.EmbeddingWithMeta, len(embeddings)),
		Model:      s.config.Model,
		Provider:   string(s.provider.GetProviderType()),
	}

	for i, emb := range embeddings {
		job.Embeddings[i] = models.EmbeddingWithMeta{
			Vector: emb.Embedding,
			Chunk:  chunks[i],
		}
	}

	return job, nil
}

// chunkText splits text into chunks based on strategy
func (s *Service) chunkText(text string) []models.TextChunk {
	switch s.config.ChunkStrategy {
	case models.ChunkingFixed:
		return s.chunkFixed(text)
	case models.ChunkingSentence:
		return s.chunkSentence(text)
	case models.ChunkingParagraph:
		return s.chunkParagraph(text)
	default:
		return s.chunkFixed(text)
	}
}

// chunkFixed chunks text by fixed size
func (s *Service) chunkFixed(text string) []models.TextChunk {
	var chunks []models.TextChunk
	chunkSize := s.config.ChunkSize
	overlap := s.config.ChunkOverlap

	for i := 0; i < len(text); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunks = append(chunks, models.TextChunk{
			Text:     text[i:end],
			Index:    len(chunks),
			StartPos: i,
			EndPos:   end,
		})

		if end == len(text) {
			break
		}
	}

	return chunks
}

// chunkSentence chunks text by sentences (simplified)
func (s *Service) chunkSentence(text string) []models.TextChunk {
	// Simple sentence splitting - could be enhanced
	// For now, just use fixed chunking
	return s.chunkFixed(text)
}

// chunkParagraph chunks text by paragraphs
func (s *Service) chunkParagraph(text string) []models.TextChunk {
	// Simple paragraph splitting - could be enhanced
	// For now, just use fixed chunking
	return s.chunkFixed(text)
}
