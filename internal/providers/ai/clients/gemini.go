package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// GeminiClient implements the domain.LLMProvider interface for Google's Gemini API
type GeminiClient struct {
	httpClient   *http.Client
	apiKey       string
	model        string
	providerType domain.ProviderType
	config       *config.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}


// Gemini API request/response types for embeddings
type geminiEmbeddingContent struct {
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

type geminiEmbeddingRequest struct {
	Model                string                 `json:"model"`
	Content              geminiEmbeddingContent `json:"content"`
	TaskType             string                 `json:"taskType,omitempty"`
	OutputDimensionality int                    `json:"outputDimensionality,omitempty"`
}

type geminiEmbeddingBatchRequest struct {
	Requests []geminiEmbeddingRequest `json:"requests"`
}

type geminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

type geminiEmbeddingBatchResponse struct {
	Embeddings []geminiEmbeddingResponse `json:"embeddings"`
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini")
	}

	// Get model or use default
	model := cfg.DefaultEmbeddingModel
	if model == "" {
		model = "models/embedding-001" // Default Gemini embedding model
		logging.Warn("No embedding model specified for Gemini, using default: %s", model)
	}

	logging.Info("Creating Gemini client with model: %s", model)

	// Set timeout from config or use default
	timeout := 60 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	// Set max retries from config or use default
	maxRetries := 3
	if cfg.MaxRetries > 0 {
		maxRetries = cfg.MaxRetries
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &GeminiClient{
		httpClient:   httpClient,
		apiKey:       cfg.APIKey,
		model:        model,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion - Not implemented for Gemini embedding-only client
func (c *GeminiClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	return nil, fmt.Errorf("completion not supported by Gemini embedding client")
}

// StreamCompletion - Not implemented for Gemini embedding-only client
func (c *GeminiClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	return nil, fmt.Errorf("streaming completion not supported by Gemini embedding client")
}

// CreateEmbeddings generates vector embeddings using the Gemini API
func (c *GeminiClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	// Use the configured embedding model or fallback to default
	model := req.Model
	if model == "" {
		model = c.model
	}

	// Ensure model has proper prefix for Gemini API
	if !strings.HasPrefix(model, "models/") {
		model = "models/" + model
	}

	logging.Info("Sending embeddings request to Gemini API with model %s for %d inputs", model, len(req.Input))

	// Gemini API supports batch requests, so we'll create individual requests for each input
	var requests []geminiEmbeddingRequest

	for _, input := range req.Input {
		geminiReq := geminiEmbeddingRequest{
			Model: model,
			Content: geminiEmbeddingContent{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: input},
				},
			},
		}

		// Set task type based on use case (optional)
		if req.User != "" {
			// Map user field to task type for Gemini
			geminiReq.TaskType = mapToGeminiTaskType(req.User)
		}

		// Set output dimensionality if specified
		if req.Dimensions > 0 {
			geminiReq.OutputDimensionality = req.Dimensions
		}

		requests = append(requests, geminiReq)
	}

	// Create batch request
	batchReq := geminiEmbeddingBatchRequest{
		Requests: requests,
	}

	// Marshal the request
	requestBody, err := json.Marshal(batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Construct the URL - use batchEmbedContent for multiple inputs
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:batchEmbedContent?key=%s",
		strings.TrimPrefix(model, "models/"), c.apiKey)

	// If only one input, use embedContent instead of batchEmbedContent
	if len(req.Input) == 1 {
		singleReq := requests[0]
		requestBody, err = json.Marshal(singleReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal single embedding request: %w", err)
		}
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
			strings.TrimPrefix(model, "models/"), c.apiKey)
	}

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Gemini embeddings API request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second) // Exponential backoff
		}

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			lastErr = fmt.Errorf("failed to create HTTP request: %w", err)
			continue
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")

		// Make the HTTP request
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("Gemini embeddings HTTP error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}
		defer resp.Body.Close()

		// Read response body
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			logging.Error("%v", lastErr)
			continue
		}

		// Check for HTTP errors
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("Gemini embeddings API returned status %d: %s", resp.StatusCode, string(responseBody))
			logging.Error("%v", lastErr)
			continue
		}

		// Parse the response
		var embeddings []geminiEmbeddingResponse

		if len(req.Input) == 1 {
			// Single embedding response
			var singleResp geminiEmbeddingResponse
			if err := json.Unmarshal(responseBody, &singleResp); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal single embedding response: %w", err)
				logging.Error("%v", lastErr)
				continue
			}
			embeddings = []geminiEmbeddingResponse{singleResp}
		} else {
			// Batch embedding response
			var batchResp geminiEmbeddingBatchResponse
			if err := json.Unmarshal(responseBody, &batchResp); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal batch embedding response: %w", err)
				logging.Error("%v", lastErr)
				continue
			}
			embeddings = batchResp.Embeddings
		}

		logging.Info("Successfully received embeddings response from Gemini API: %d embeddings", len(embeddings))

		// Convert Gemini response to domain format
		domainEmbeddings := make([]domain.Embedding, len(embeddings))
		for i, embedding := range embeddings {
			domainEmbeddings[i] = domain.Embedding{
				Object:    "embedding",
				Index:     i,
				Embedding: embedding.Embedding.Values,
			}
		}

		// Return domain EmbeddingResponse
		return &domain.EmbeddingResponse{
			Object: "list",
			Data:   domainEmbeddings,
			Model:  model,
			Usage: domain.Usage{
				PromptTokens:     0, // Gemini doesn't provide token usage in embedding responses
				CompletionTokens: 0,
				TotalTokens:      0,
			},
		}, nil
	}

	// If we get here, all retries failed
	return nil, fmt.Errorf("embeddings failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// mapToGeminiTaskType maps user-provided task hints to Gemini task types
func mapToGeminiTaskType(userHint string) string {
	switch strings.ToLower(userHint) {
	case "similarity", "semantic_similarity":
		return "SEMANTIC_SIMILARITY"
	case "classification":
		return "CLASSIFICATION"
	case "clustering":
		return "CLUSTERING"
	case "retrieval_document", "document":
		return "RETRIEVAL_DOCUMENT"
	case "retrieval_query", "query":
		return "RETRIEVAL_QUERY"
	case "code_retrieval_query", "code":
		return "CODE_RETRIEVAL_QUERY"
	case "question_answering", "qa":
		return "QUESTION_ANSWERING"
	case "fact_verification", "fact_check":
		return "FACT_VERIFICATION"
	default:
		return "SEMANTIC_SIMILARITY" // Default task type
	}
}

// GetSupportedEmbeddingModels returns a list of embedding models supported by Gemini
func (c *GeminiClient) GetSupportedEmbeddingModels() []string {
	if c.config.EmbeddingModels != nil && len(c.config.EmbeddingModels) > 0 {
		var models []string
		for model := range c.config.EmbeddingModels {
			models = append(models, model)
		}
		return models
	}

	// Default Gemini embedding models
	return []string{
		"embedding-001",
		"text-embedding-004", // Legacy
	}
}

// GetMaxEmbeddingTokens returns the maximum token limit for embeddings for the given model
func (c *GeminiClient) GetMaxEmbeddingTokens(model string) int {
	// Check if we have specific configuration for this model
	if c.config.EmbeddingModels != nil {
		if modelConfig, exists := c.config.EmbeddingModels[model]; exists {
			return modelConfig.MaxTokens
		}
	}

	// Default token limits for Gemini embedding models
	modelLower := strings.ToLower(model)
	switch {
	case strings.Contains(modelLower, "embedding-001"):
		return 2048 // Gemini embedding-001 token limit
	case strings.Contains(modelLower, "text-embedding-004"):
		return 2048 // Legacy model token limit
	default:
		return 2048 // Conservative default
	}
}

// GetProviderType returns the provider type
func (c *GeminiClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType returns the interface type
func (c *GeminiClient) GetInterfaceType() config.InterfaceType {
	return config.GeminiNative
}

// ValidateConfig validates the provider configuration
func (c *GeminiClient) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("configuration is required")
	}

	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required for Gemini")
	}

	return nil
}

// Close cleans up provider resources
func (c *GeminiClient) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}
