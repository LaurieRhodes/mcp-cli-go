// File: internal/services/workflow/embedding_service.go
package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Enhanced Service with embedding capabilities
type EnhancedService struct {
	*Service // Embed the existing service
	
	// New fields for embedding support
	embeddingService domain.EmbeddingService
	vectorMemory     *VectorMemoryManager
	embeddingCache   *EmbeddingCache
	config           *config.ApplicationConfig
	serverManager    domain.MCPServerManager
}

// VectorMemoryManager manages embedding vectors in memory during workflow execution
type VectorMemoryManager struct {
	embeddings   map[string]*domain.EmbeddingResult
	maxMemoryMB  int
	currentUsage int64
	mutex        sync.RWMutex
	createdAt    map[string]time.Time
}

// EmbeddingCache caches embeddings to avoid regeneration
type EmbeddingCache struct {
	cache     map[string]*domain.EmbeddingResult
	ttl       time.Duration
	mutex     sync.RWMutex
}

// NewEnhancedService creates a new enhanced workflow service with embedding support
func NewEnhancedService(
	appConfig *config.ApplicationConfig,
	configService domain.ConfigurationService,
	providerService domain.LLMProvider,
	serverManager domain.MCPServerManager,
	embeddingService domain.EmbeddingService,
) (*EnhancedService, error) {
	
	baseService := NewService(appConfig, configService, providerService, serverManager)
	
	// Initialize embedding-specific components
	maxMemoryMB := 100 // Default 100MB
	if appConfig.Embeddings != nil && appConfig.Embeddings.MaxVectorMemoryMB > 0 {
		maxMemoryMB = appConfig.Embeddings.MaxVectorMemoryMB
	}
	
	cacheTTL := 1 * time.Hour // Default 1 hour
	if appConfig.Embeddings != nil && appConfig.Embeddings.CacheTTL != "" {
		if parsed, err := time.ParseDuration(appConfig.Embeddings.CacheTTL); err == nil {
			cacheTTL = parsed
		}
	}
	
	return &EnhancedService{
		Service:          baseService,
		embeddingService: embeddingService,
		vectorMemory: &VectorMemoryManager{
			embeddings:   make(map[string]*domain.EmbeddingResult),
			maxMemoryMB:  maxMemoryMB,
			createdAt:    make(map[string]time.Time),
		},
		embeddingCache: &EmbeddingCache{
			cache: make(map[string]*domain.EmbeddingResult),
			ttl:   cacheTTL,
		},
		config:        appConfig,
		serverManager: serverManager,
	}, nil
}

// ExecuteWorkflow executes a workflow template
func (es *EnhancedService) ExecuteWorkflow(ctx context.Context, templateName string, inputData string) (*domain.WorkflowResponse, error) {
	return es.Service.ProcessWorkflow(ctx, &domain.WorkflowRequest{
		TemplateName: templateName,
		InputData:    inputData,
		ExecutionID:  generateExecutionID(),
	})
}

func (es *EnhancedService) getWorkflowTemplate(step *domain.WorkflowStep) *domain.WorkflowTemplate {
	// Return a minimal template for variable processing
	return &domain.WorkflowTemplate{
		Variables: make(map[string]string),
	}
}

// generateStepEmbedding generates embeddings for a workflow step
func (es *EnhancedService) generateStepEmbedding(ctx context.Context, step *domain.WorkflowStep, variables map[string]interface{}) (*domain.EmbeddingResult, error) {
	// Validate embedding configuration
	if err := step.Embedding.Validate(); err != nil {
		return nil, err
	}
	
	// Process input text with variable substitution
	template := es.getWorkflowTemplate(step)
	inputText := template.ProcessVariables(step.Embedding.Input, variables)
	
	logging.Debug("Generating embedding for input: %s (length: %d)", truncateString(inputText, 100), len(inputText))
	
	// Check cache first
	if es.config.Embeddings != nil && es.config.Embeddings.CacheEmbeddings {
		if cached := es.embeddingCache.Get(inputText, step.Embedding); cached != nil {
			logging.Debug("Using cached embedding for input")
			return cached, nil
		}
	}
	
	// Get embedding configuration with defaults
	provider, model, err := es.getEmbeddingProviderAndModel(step.Embedding)
	if err != nil {
		return nil, err
	}
	
	chunkStrategy := step.Embedding.ChunkStrategy
	if chunkStrategy == "" && es.config.Embeddings != nil {
		chunkStrategy = string(es.config.Embeddings.DefaultChunkStrategy)
	}
	if chunkStrategy == "" {
		chunkStrategy = "sentence" // Default fallback
	}
	
	maxChunkSize := step.Embedding.MaxChunkSize
	if maxChunkSize == 0 && es.config.Embeddings != nil {
		maxChunkSize = es.config.Embeddings.DefaultMaxChunkSize
	}
	if maxChunkSize == 0 {
		maxChunkSize = 512 // Default fallback
	}
	
	overlap := step.Embedding.Overlap
	if overlap == 0 && es.config.Embeddings != nil {
		overlap = es.config.Embeddings.DefaultOverlap
	}
	
	// Generate embedding using existing embedding service
	embeddingJob, err := es.embeddingService.GenerateEmbeddings(ctx, &domain.EmbeddingJobRequest{
		Input:         inputText,
		Provider:      provider,
		Model:         model,
		ChunkStrategy: domain.ChunkingType(chunkStrategy),
		MaxChunkSize:  maxChunkSize,
		ChunkOverlap:  overlap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	// Convert to EmbeddingResult (metadata only, vectors stored privately)
	result := &domain.EmbeddingResult{
		ID:         generateEmbeddingID(),
		Input:      inputText,
		Provider:   provider,
		Model:      model,
		Dimensions: len(embeddingJob.Embeddings[0].Vector),
		ChunkCount: len(embeddingJob.Chunks),
		TokenCount: calculateTotalTokens(embeddingJob.Chunks),
		GeneratedAt: time.Now(),
		Metadata: map[string]interface{}{
			"chunk_strategy":  chunkStrategy,
			"max_chunk_size":  maxChunkSize,
			"overlap":         overlap,
			"generation_time": time.Since(embeddingJob.StartTime),
		},
	}
	
	// Extract and store vectors privately
	vectors := make([][]float32, len(embeddingJob.Embeddings))
	for i, embedding := range embeddingJob.Embeddings {
		vectors[i] = embedding.Vector
	}
	result.SetVectors(vectors)
	
	// Convert chunks without exposing vectors
	chunks := make([]domain.EmbeddingChunk, len(embeddingJob.Chunks))
	for i, chunk := range embeddingJob.Chunks {
		chunks[i] = domain.EmbeddingChunk{
			Text:       chunk.Text,
			Index:      chunk.Index,
			StartPos:   chunk.StartPos,
			EndPos:     chunk.EndPos,
			TokenCount: chunk.TokenCount,
			Vector:     embeddingJob.Embeddings[i].Vector, // Stored privately
		}
	}
	result.SetChunks(chunks)
	
	// Store in memory for workflow duration
	if err := es.vectorMemory.Store(result); err != nil {
		logging.Warn("Failed to store embedding in memory: %v", err)
	}
	
	// Cache the result if caching is enabled
	if es.config.Embeddings != nil && es.config.Embeddings.CacheEmbeddings {
		es.embeddingCache.Store(inputText, step.Embedding, result)
	}
	
	logging.Info("Generated embedding: ID=%s, dimensions=%d, chunks=%d, tokens=%d", 
		result.ID, result.Dimensions, result.ChunkCount, result.TokenCount)
	
	return result, nil
}

func (es *EnhancedService) storeEmbeddingUsageResults(variables map[string]interface{}, usageType domain.EmbeddingUsageType, result interface{}, stepNum int) {
	switch usageType {
	case domain.EmbeddingUsageVectorSearch:
		variables["vector_search_results"] = result
		variables[fmt.Sprintf("step%d_vector_search_results", stepNum)] = result
		
		if resultStr, ok := result.(string); ok {
			lines := strings.Split(strings.TrimSpace(resultStr), "\n")
			count := 0
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "-") && !strings.Contains(line, "identifier") {
					count++
				}
			}
			variables["vector_search_results_count"] = count
		}
		
	case domain.EmbeddingUsageSimilarity:
		variables["similarity_score"] = result
		variables[fmt.Sprintf("step%d_similarity_score", stepNum)] = result
		
	case domain.EmbeddingUsageClustering:
		variables["clustering_results"] = result
		variables[fmt.Sprintf("step%d_clustering_results", stepNum)] = result
		
	case domain.EmbeddingUsageStorage:
		variables["embedding_storage_result"] = result
		variables[fmt.Sprintf("step%d_embedding_storage", stepNum)] = result
	}
}

func (es *EnhancedService) getEmbeddingProviderAndModel(config *domain.EmbeddingConfig) (string, string, error) {
	provider := config.Provider
	if provider == "" && es.config.Embeddings != nil {
		provider = es.config.Embeddings.DefaultProvider
	}
	if provider == "" {
		return "", "", fmt.Errorf("no embedding provider specified and no default configured")
	}
	
	model := config.Model
	if model == "" && es.config.Embeddings != nil && es.config.Embeddings.Providers != nil {
		if providerConfig, exists := es.config.Embeddings.Providers[provider]; exists {
			model = providerConfig.DefaultModel
		}
	}
	if model == "" {
		return "", "", fmt.Errorf("no embedding model specified and no default configured for provider %s", provider)
	}
	
	return provider, model, nil
}

func (es *EnhancedService) executeSimilarityComparison(ctx context.Context, embedding *domain.EmbeddingResult, config *domain.SimilarityConfig, variables map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"similarity_comparison": "not implemented in this version",
	}, nil
}

func (es *EnhancedService) executeClustering(ctx context.Context, embedding *domain.EmbeddingResult, config *domain.ClusteringConfig) (interface{}, error) {
	return map[string]interface{}{
		"clustering": "not implemented in this version",
	}, nil
}

// VectorMemoryManager methods

func (vmm *VectorMemoryManager) Store(embedding *domain.EmbeddingResult) error {
	vmm.mutex.Lock()
	defer vmm.mutex.Unlock()
	
	estimatedSize := embedding.EstimateMemoryUsage()
	memoryLimitBytes := int64(vmm.maxMemoryMB) * 1024 * 1024
	
	if vmm.currentUsage+estimatedSize > memoryLimitBytes {
		logging.Warn("Vector memory limit reached, evicting old embeddings")
		for id := range vmm.embeddings {
			oldEmbedding := vmm.embeddings[id]
			vmm.currentUsage -= oldEmbedding.EstimateMemoryUsage()
			delete(vmm.embeddings, id)
			delete(vmm.createdAt, id)
			
			if vmm.currentUsage+estimatedSize <= memoryLimitBytes {
				break
			}
		}
	}
	
	vmm.embeddings[embedding.ID] = embedding
	vmm.createdAt[embedding.ID] = time.Now()
	vmm.currentUsage += estimatedSize
	
	logging.Debug("Stored embedding %s in memory (%.2f MB used)", embedding.ID, float64(vmm.currentUsage)/(1024*1024))
	
	return nil
}

func (vmm *VectorMemoryManager) Get(id string) (*domain.EmbeddingResult, bool) {
	vmm.mutex.RLock()
	defer vmm.mutex.RUnlock()
	
	embedding, exists := vmm.embeddings[id]
	return embedding, exists
}

func (vmm *VectorMemoryManager) Clear() {
	vmm.mutex.Lock()
	defer vmm.mutex.Unlock()
	
	vmm.embeddings = make(map[string]*domain.EmbeddingResult)
	vmm.createdAt = make(map[string]time.Time)
	vmm.currentUsage = 0
	
	logging.Debug("Cleared vector memory")
}

// EmbeddingCache methods

func (ec *EmbeddingCache) Get(input string, config *domain.EmbeddingConfig) *domain.EmbeddingResult {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	key := ec.createCacheKey(input, config)
	if cached, exists := ec.cache[key]; exists {
		if time.Since(cached.GeneratedAt) < ec.ttl {
			return cached
		}
	}
	
	return nil
}

func (ec *EmbeddingCache) Store(input string, config *domain.EmbeddingConfig, result *domain.EmbeddingResult) {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	key := ec.createCacheKey(input, config)
	ec.cache[key] = result
}

func (ec *EmbeddingCache) createCacheKey(input string, config *domain.EmbeddingConfig) string {
	inputLen := len(input)
	if inputLen > 50 {
		inputLen = 50
	}
	return fmt.Sprintf("%s:%s:%s:%s:%d:%d", 
		input[:inputLen], config.Provider, config.Model, config.ChunkStrategy, config.MaxChunkSize, config.Overlap)
}

// Utility functions

func generateExecutionID() string {
	return generateEmbeddingID()
}

func generateEmbeddingID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "emb_" + hex.EncodeToString(bytes)
}

func calculateTotalTokens(chunks []domain.Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += chunk.TokenCount
	}
	return total
}

func formatVectorForPostgres(vector []float32) string {
	var parts []string
	for _, v := range vector {
		parts = append(parts, strconv.FormatFloat(float64(v), 'f', 8, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
