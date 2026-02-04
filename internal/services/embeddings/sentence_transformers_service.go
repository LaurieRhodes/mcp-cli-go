package embeddings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// SentenceTransformersConfig defines configuration for SentenceTransformers service
type SentenceTransformersConfig struct {
	APIEndpoint         string                 `json:"api_endpoint"`
	DefaultModel        string                 `json:"default_model"`
	AvailableModels     map[string]ModelConfig `json:"available_models"`
	TimeoutSeconds      int                    `json:"timeout_seconds"`
	MaxRetries          int                    `json:"max_retries"`
	BatchSize           int                    `json:"batch_size"`
	PoolingStrategy     string                 `json:"pooling_strategy,omitempty"` // "mean", "cls", "max"
	NormalizeEmbeddings bool                   `json:"normalize_embeddings"`
}

// ModelConfig defines configuration for a specific SentenceTransformers model
type ModelConfig struct {
	ModelName           string   `json:"model_name"`
	MaxTokens           int      `json:"max_tokens"`
	Dimensions          int      `json:"dimensions"`
	Description         string   `json:"description"`
	OptimalDomains      []string `json:"optimal_domains,omitempty"`
	MultiLingual        bool     `json:"multi_lingual"`
	SupportsInstruction bool     `json:"supports_instruction"`
	Default             bool     `json:"default,omitempty"`
}

// SentenceTransformersRequest represents a request to the SentenceTransformers API
type SentenceTransformersRequest struct {
	Sentences           []string `json:"sentences"`
	Model               string   `json:"model,omitempty"`
	PoolingStrategy     string   `json:"pooling_strategy,omitempty"`
	NormalizeEmbeddings bool     `json:"normalize_embeddings,omitempty"`
	InstructionPrefix   string   `json:"instruction_prefix,omitempty"`
	MaxLength           int      `json:"max_length,omitempty"`
}

// SentenceTransformersResponse represents the response from SentenceTransformers API
type SentenceTransformersResponse struct {
	Embeddings     [][]float32 `json:"embeddings"`
	Model          string      `json:"model"`
	Dimensions     int         `json:"dimensions"`
	TokensUsed     int         `json:"tokens_used,omitempty"`
	ProcessingTime string      `json:"processing_time,omitempty"`
}

// SentenceTransformersService provides advanced embedding capabilities using SentenceTransformers
type SentenceTransformersService struct {
	config SentenceTransformersConfig
}

// NewSentenceTransformersService creates a new SentenceTransformers embedding service
func NewSentenceTransformersService(config SentenceTransformersConfig) *SentenceTransformersService {
	// Set defaults
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.BatchSize == 0 {
		config.BatchSize = 32
	}
	if config.PoolingStrategy == "" {
		config.PoolingStrategy = "mean"
	}

	return &SentenceTransformersService{
		config: config,
	}
}

// GenerateEmbeddings generates embeddings using SentenceTransformers models
func (sts *SentenceTransformersService) GenerateEmbeddings(ctx context.Context, request *domain.EmbeddingJobRequest) (*domain.EmbeddingJob, error) {
	startTime := time.Now()

	logging.Info("🧠 Generating SentenceTransformers embeddings for input length: %d", len(request.Input))

	// Choose the optimal model for the request
	model, modelConfig, err := sts.selectOptimalModel(request)
	if err != nil {
		return nil, fmt.Errorf("failed to select model: %w", err)
	}

	logging.Debug("Selected SentenceTransformers model: %s (dimensions: %d)", model, modelConfig.Dimensions)

	// Chunk the input text
	chunks, err := sts.chunkText(request.Input, request.ChunkStrategy, request.MaxChunkSize, request.ChunkOverlap, modelConfig.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk text: %w", err)
	}

	logging.Debug("Created %d chunks for embedding generation", len(chunks))

	// Generate embeddings for all chunks
	embeddings, err := sts.generateEmbeddingsForChunks(ctx, chunks, model, modelConfig, request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create embedding job result
	job := &domain.EmbeddingJob{
		ID:         generateJobID(),
		Input:      request.Input,
		Provider:   "sentence-transformers",
		Model:      model,
		StartTime:  startTime,
		Timestamp:  time.Now(),
		Chunks:     chunks,
		Embeddings: embeddings,
		Metadata: map[string]interface{}{
			"model_config":         modelConfig,
			"pooling_strategy":     sts.config.PoolingStrategy,
			"normalize_embeddings": sts.config.NormalizeEmbeddings,
			"batch_size":           sts.config.BatchSize,
			"total_tokens":         calculateTotalTokens(chunks),
		},
	}

	processingTime := time.Since(startTime)

	logging.Info("✅ SentenceTransformers embedding generation completed: %d embeddings in %v",
		len(embeddings), processingTime)

	return job, nil
}

// selectOptimalModel selects the best SentenceTransformers model for the request
func (sts *SentenceTransformersService) selectOptimalModel(request *domain.EmbeddingJobRequest) (string, ModelConfig, error) {
	requestedModel := request.Model
	if requestedModel == "" {
		requestedModel = sts.config.DefaultModel
	}

	// Check if requested model is available
	if modelConfig, exists := sts.config.AvailableModels[requestedModel]; exists {
		return requestedModel, modelConfig, nil
	}

	// If specific model not found, try to find optimal model based on use case
	return sts.findOptimalModel(request)
}

// findOptimalModel finds the optimal model based on request characteristics
func (sts *SentenceTransformersService) findOptimalModel(request *domain.EmbeddingJobRequest) (string, ModelConfig, error) {
	// Domain-specific model selection
	domain := sts.inferDomain(request.Input)

	// Look for domain-optimized models
	for modelName, config := range sts.config.AvailableModels {
		for _, optimalDomain := range config.OptimalDomains {
			if optimalDomain == domain {
				logging.Debug("Selected domain-optimized model %s for domain: %s", modelName, domain)
				return modelName, config, nil
			}
		}
	}

	// Fall back to default model
	for modelName, config := range sts.config.AvailableModels {
		if config.Default {
			return modelName, config, nil
		}
	}

	// If no default found, use the first available model
	for modelName, config := range sts.config.AvailableModels {
		return modelName, config, nil
	}

	return "", ModelConfig{}, fmt.Errorf("no SentenceTransformers models available")
}

// inferDomain attempts to infer the domain of the input text
func (sts *SentenceTransformersService) inferDomain(input string) string {
	inputLower := strings.ToLower(input)

	// Cybersecurity domain indicators
	securityTerms := []string{
		"security", "cyber", "vulnerability", "threat", "attack", "malware",
		"firewall", "encryption", "authentication", "authorization", "compliance",
		"audit", "governance", "risk", "incident", "breach", "phishing",
	}

	// Medical domain indicators
	medicalTerms := []string{
		"medical", "health", "patient", "diagnosis", "treatment", "medication",
		"hospital", "clinic", "doctor", "nurse", "disease", "symptom",
	}

	// Legal domain indicators
	legalTerms := []string{
		"legal", "law", "court", "contract", "agreement", "regulation",
	}

	// Count term matches for each domain
	securityCount := countMatches(inputLower, securityTerms)
	medicalCount := countMatches(inputLower, medicalTerms)
	legalCount := countMatches(inputLower, legalTerms)

	// Return domain with highest match count
	if securityCount > medicalCount && securityCount > legalCount {
		return "cybersecurity"
	} else if medicalCount > legalCount {
		return "medical"
	} else if legalCount > 0 {
		return "legal"
	}

	return "general"
}

// countMatches counts how many terms from the list appear in the text
func countMatches(text string, terms []string) int {
	count := 0
	for _, term := range terms {
		if strings.Contains(text, term) {
			count++
		}
	}
	return count
}

// Stub methods to make the file compile
func (sts *SentenceTransformersService) chunkText(input string, strategy domain.ChunkingType, maxSize, overlap, maxTokens int) ([]domain.Chunk, error) {
	// Placeholder implementation
	return []domain.Chunk{{Text: input, Index: 0, TokenCount: len(input) / 4}}, nil
}

func (sts *SentenceTransformersService) generateEmbeddingsForChunks(ctx context.Context, chunks []domain.Chunk, model string, config ModelConfig, request *domain.EmbeddingJobRequest) ([]domain.EmbeddingWithMeta, error) {
	// Placeholder implementation
	return []domain.EmbeddingWithMeta{}, nil
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().Unix())
}

func calculateTotalTokens(chunks []domain.Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += chunk.TokenCount
	}
	return total
}
