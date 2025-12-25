package embeddings

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/chunking"
	"github.com/LaurieRhodes/mcp-cli-go/internal/core/tokens"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Service implements the domain.EmbeddingService interface
type Service struct {
	configService   domain.ConfigurationService
	providerFactory domain.ProviderFactory
	chunkingManager *chunking.ChunkingManager
}

// NewService creates a new embeddings service
func NewService(configService domain.ConfigurationService, providerFactory domain.ProviderFactory) domain.EmbeddingService {
	return &Service{
		configService:   configService,
		providerFactory: providerFactory,
		chunkingManager: chunking.NewChunkingManager(),
	}
}

// GenerateEmbeddings processes text input and returns embeddings
func (s *Service) GenerateEmbeddings(ctx context.Context, req *domain.EmbeddingJobRequest) (*domain.EmbeddingJob, error) {
	logging.Info("Starting embedding generation for %d characters of input text", len(req.Input))

	// Validate the request
	if err := s.ValidateEmbeddingRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Apply defaults
	req = s.applyDefaults(req)

	// Get provider configuration
	providerName := req.Provider
	if providerName == "" {
		// Use default provider
		defaultProviderName, _, _, err := s.configService.GetDefaultProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get default provider: %w", err)
		}
		providerName = defaultProviderName
	}

	providerConfig, interfaceType, err := s.configService.GetProviderConfig(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config for %s: %w", providerName, err)
	}

	// Create provider instance - FIXED: Added missing gemini case
	var providerType domain.ProviderType
	switch providerName {
	case "openai":
		providerType = domain.ProviderOpenAI
	case "deepseek":
		providerType = domain.ProviderDeepSeek
	case "openrouter":
		providerType = domain.ProviderOpenRouter
	case "gemini":
		providerType = domain.ProviderGemini
	case "lmstudio":
		providerType = domain.ProviderLMStudio
	default:
		providerType = domain.ProviderOpenAI // Default fallback
	}

	logging.Info("Creating provider %s (type: %s, interface: %s)", providerName, providerType, interfaceType)

	provider, err := s.providerFactory.CreateProvider(providerType, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", providerName, err)
	}
	defer provider.Close()

	// Determine embedding model
	embeddingModel := req.Model
	if embeddingModel == "" {
		if providerConfig.DefaultEmbeddingModel != "" {
			embeddingModel = providerConfig.DefaultEmbeddingModel
		} else {
			// Use first available model
			supportedModels := provider.GetSupportedEmbeddingModels()
			if len(supportedModels) > 0 {
				embeddingModel = supportedModels[0]
				logging.Info("No embedding model specified, using first available: %s", embeddingModel)
			} else {
				return nil, fmt.Errorf("no embedding models available for provider %s", providerName)
			}
		}
	}

	// Get token limits for the model
	maxTokens := provider.GetMaxEmbeddingTokens(embeddingModel)
	if req.MaxChunkSize > 0 && req.MaxChunkSize < maxTokens {
		maxTokens = req.MaxChunkSize
	}

	logging.Info("Using embedding model %s with max %d tokens per chunk", embeddingModel, maxTokens)

	// Create token manager for the embedding model
	tokenManager, err := tokens.NewTokenManagerFromProvider(embeddingModel, providerConfig)
	if err != nil {
		// Fallback to basic token manager
		tokenManager, err = tokens.NewTokenManagerFallback(embeddingModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create token manager: %w", err)
		}
	}

	// Get chunking strategy
	chunkingStrategy, err := s.chunkingManager.GetStrategy(req.ChunkStrategy, tokenManager, req.ChunkOverlap)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunking strategy: %w", err)
	}

	logging.Info("Using chunking strategy: %s with overlap: %d", chunkingStrategy.GetName(), req.ChunkOverlap)

	// Chunk the input text
	chunks, err := chunkingStrategy.ChunkText(req.Input, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk text: %w", err)
	}

	logging.Info("Text chunked into %d chunks", len(chunks))

	// Prepare input for embedding API
	var inputTexts []string
	for _, chunk := range chunks {
		inputTexts = append(inputTexts, chunk.Text)
	}

	// Create embedding request
	embeddingReq := &domain.EmbeddingRequest{
		Input:          inputTexts,
		Model:          embeddingModel,
		EncodingFormat: req.EncodingFormat,
		Dimensions:     req.Dimensions,
	}

	// Generate embeddings
	logging.Info("Generating embeddings for %d chunks using provider %s", len(inputTexts), providerType)
	embeddingResp, err := provider.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Combine embeddings with chunk metadata
	var embeddingsWithMeta []domain.EmbeddingWithMeta
	for i, embedding := range embeddingResp.Data {
		if i < len(chunks) {
			embeddingMeta := domain.EmbeddingWithMeta{
				Vector: embedding.Embedding,
				Chunk:  chunks[i],
				Metadata: map[string]interface{}{
					"model_dimensions": len(embedding.Embedding),
					"chunk_strategy":   string(req.ChunkStrategy),
					"provider":         providerName,
					"model":            embeddingModel,
				},
			}

			// Add any custom metadata
			if req.Metadata != nil {
				for key, value := range req.Metadata {
					embeddingMeta.Metadata[key] = value
				}
			}

			embeddingsWithMeta = append(embeddingsWithMeta, embeddingMeta)
		}
	}

	// Generate job ID
	jobID := s.generateJobID()

	// Create job metadata
	jobMetadata := map[string]interface{}{
		"total_chunks":   len(chunks),
		"total_tokens":   embeddingResp.Usage.TotalTokens,
		"chunk_strategy": string(req.ChunkStrategy),
		"max_chunk_size": maxTokens,
		"chunk_overlap":  req.ChunkOverlap,
		"provider":       providerName,
		"interface_type": string(interfaceType),
		"input_length":   len(req.Input),
	}

	// Add custom metadata
	if req.Metadata != nil {
		for key, value := range req.Metadata {
			jobMetadata[key] = value
		}
	}

	// Create the embedding job
	job := &domain.EmbeddingJob{
		ID:         jobID,
		Input:      req.Input,
		Chunks:     chunks,
		Embeddings: embeddingsWithMeta,
		Model:      embeddingModel,
		Provider:   providerName,
		Timestamp:  time.Now().UTC(),
		Metadata:   jobMetadata,
	}

	logging.Info("Embedding generation completed: job %s with %d embeddings", jobID, len(embeddingsWithMeta))

	return job, nil
}

// GetAvailableChunkingStrategies returns available chunking strategies
func (s *Service) GetAvailableChunkingStrategies() []domain.ChunkingType {
	return s.chunkingManager.GetAvailableStrategies()
}

// ValidateEmbeddingRequest validates an embedding request
func (s *Service) ValidateEmbeddingRequest(req *domain.EmbeddingJobRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Input == "" {
		return fmt.Errorf("input text is required")
	}

	if len(req.Input) > 1000000 { // 1MB limit
		return fmt.Errorf("input text too large (max 1MB)")
	}

	if req.MaxChunkSize < 0 {
		return fmt.Errorf("max chunk size cannot be negative")
	}

	if req.ChunkOverlap < 0 {
		return fmt.Errorf("chunk overlap cannot be negative")
	}

	if req.Dimensions < 0 {
		return fmt.Errorf("dimensions cannot be negative")
	}

	// Validate chunking strategy
	if req.ChunkStrategy != "" {
		availableStrategies := s.GetAvailableChunkingStrategies()
		valid := false
		for _, strategy := range availableStrategies {
			if strategy == req.ChunkStrategy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported chunking strategy: %s", req.ChunkStrategy)
		}
	}

	return nil
}

// applyDefaults applies default values to the request
func (s *Service) applyDefaults(req *domain.EmbeddingJobRequest) *domain.EmbeddingJobRequest {
	// Make a copy to avoid modifying the original
	result := *req

	// Apply chunking strategy default
	if result.ChunkStrategy == "" {
		result.ChunkStrategy = domain.ChunkingSentence
	}

	// Apply max chunk size default
	if result.MaxChunkSize == 0 {
		result.MaxChunkSize = 512
	}

	// Apply chunk overlap default
	if result.ChunkOverlap == 0 {
		result.ChunkOverlap = 0 // No overlap by default
	}

	// Apply encoding format default
	if result.EncodingFormat == "" {
		result.EncodingFormat = "float"
	}

	return &result
}

// generateJobID generates a unique job identifier
func (s *Service) generateJobID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "emb_" + hex.EncodeToString(bytes)
}
