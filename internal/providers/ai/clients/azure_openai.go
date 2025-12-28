package clients

import (
	"bufio"
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

// AzureOpenAIClient implements domain.LLMProvider for Azure OpenAI Service
// Uses the same request/response format as OpenAI but with different authentication
type AzureOpenAIClient struct {
	httpClient   *http.Client
	model        string
	apiKey       string
	apiEndpoint  string
	apiVersion   string
	deploymentID string
	providerType domain.ProviderType
	config       *config.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}

// NewAzureOpenAIClient creates a new Azure OpenAI Service provider
func NewAzureOpenAIClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Azure OpenAI")
	}

	if cfg.APIEndpoint == "" {
		return nil, fmt.Errorf("API endpoint is required for Azure OpenAI")
	}

	deploymentID := cfg.DefaultModel
	if deploymentID == "" {
		return nil, fmt.Errorf("deployment ID (model name) is required for Azure OpenAI")
	}

	// Default API version
	apiVersion := "2024-02-15-preview"
	
	// Clean endpoint
	apiEndpoint := strings.TrimSuffix(cfg.APIEndpoint, "/")

	logging.Info("Creating Azure OpenAI client with deployment: %s, endpoint: %s", deploymentID, apiEndpoint)

	timeout := 45 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	maxRetries := 3
	if cfg.MaxRetries > 0 {
		maxRetries = cfg.MaxRetries
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &AzureOpenAIClient{
		httpClient:   httpClient,
		model:        deploymentID,
		apiKey:       cfg.APIKey,
		apiEndpoint:  apiEndpoint,
		apiVersion:   apiVersion,
		deploymentID: deploymentID,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *AzureOpenAIClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)
	tools := convertToOpenAITools(req.Tools)

	payload := openaiChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Sending request to Azure OpenAI with deployment %s", c.deploymentID)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Azure OpenAI request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		response, err := c.sendRequest(ctx, "/chat/completions", payload)
		if err != nil {
			lastErr = err
			logging.Error("Azure OpenAI error: %v", err)
			continue
		}

		var chatResp openaiChatResponse
		if err := json.Unmarshal(response, &chatResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		if len(chatResp.Choices) == 0 {
			lastErr = fmt.Errorf("no completion choices returned")
			continue
		}

		choice := chatResp.Choices[0].Message
		toolCalls := convertFromOpenAIToolCalls(choice.ToolCalls)

		logging.Info("Successfully received response from Azure OpenAI")

		return &domain.CompletionResponse{
			Response:  choice.Content,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion implements domain.LLMProvider
func (c *AzureOpenAIClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)
	tools := convertToOpenAITools(req.Tools)

	payload := openaiChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Starting streaming request to Azure OpenAI")

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Azure OpenAI streaming (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		resp, err := c.sendStreamingRequest(ctx, "/chat/completions", payload)
		if err != nil {
			lastErr = err
			continue
		}

		fullContent, toolCalls, streamErr := c.processStreamingResponse(resp, writer)
		if streamErr != nil {
			lastErr = streamErr
			if strings.Contains(streamErr.Error(), "context deadline exceeded") {
				continue
			}
			break
		}

		return &domain.CompletionResponse{
			Response:  fullContent,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// CreateEmbeddings implements domain.LLMProvider
func (c *AzureOpenAIClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	model := req.Model
	if model == "" {
		model = c.config.DefaultEmbeddingModel
	}
	if model == "" {
		return nil, fmt.Errorf("no embedding model specified")
	}

	payload := openaiEmbeddingRequest{
		Input:          req.Input,
		Model:          model,
		EncodingFormat: req.EncodingFormat,
		User:           req.User,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		responseData, err := c.sendRequest(ctx, "/embeddings", payload)
		if err != nil {
			lastErr = err
			continue
		}

		var embResp openaiEmbeddingResponse
		if err := json.Unmarshal(responseData, &embResp); err != nil {
			lastErr = fmt.Errorf("failed to parse embeddings response: %w", err)
			continue
		}

		domainEmbeddings := make([]domain.Embedding, len(embResp.Data))
		for i, embedding := range embResp.Data {
			domainEmbeddings[i] = domain.Embedding{
				Object:    embedding.Object,
				Index:     embedding.Index,
				Embedding: embedding.Embedding,
			}
		}

		return &domain.EmbeddingResponse{
			Object: embResp.Object,
			Data:   domainEmbeddings,
			Model:  embResp.Model,
			Usage: domain.Usage{
				PromptTokens:     embResp.Usage.PromptTokens,
				CompletionTokens: embResp.Usage.CompletionTokens,
				TotalTokens:      embResp.Usage.TotalTokens,
			},
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// GetSupportedEmbeddingModels implements domain.LLMProvider
func (c *AzureOpenAIClient) GetSupportedEmbeddingModels() []string {
	if c.config.EmbeddingModels != nil && len(c.config.EmbeddingModels) > 0 {
		var models []string
		for model := range c.config.EmbeddingModels {
			models = append(models, model)
		}
		return models
	}
	return []string{"text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002"}
}

// GetMaxEmbeddingTokens implements domain.LLMProvider
func (c *AzureOpenAIClient) GetMaxEmbeddingTokens(model string) int {
	if c.config.EmbeddingModels != nil {
		if modelConfig, exists := c.config.EmbeddingModels[model]; exists {
			return modelConfig.MaxTokens
		}
	}
	return 8191
}

// GetProviderType implements domain.LLMProvider
func (c *AzureOpenAIClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType implements domain.LLMProvider
func (c *AzureOpenAIClient) GetInterfaceType() config.InterfaceType {
	return config.AzureOpenAI
}

// ValidateConfig implements domain.LLMProvider
func (c *AzureOpenAIClient) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("configuration is required")
	}
	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.config.APIEndpoint == "" {
		return fmt.Errorf("API endpoint is required")
	}
	if c.config.DefaultModel == "" {
		return fmt.Errorf("deployment ID is required")
	}
	return nil
}

// Close implements domain.LLMProvider
func (c *AzureOpenAIClient) Close() error {
	return nil
}

// buildAzureURL constructs Azure-specific URL with deployment and API version
func (c *AzureOpenAIClient) buildAzureURL(endpoint string) string {
	baseURL := c.apiEndpoint
	
	// Build: {endpoint}/openai/deployments/{deployment-id}{endpoint}?api-version={version}
	if !strings.Contains(baseURL, "/openai/deployments/") {
		baseURL = baseURL + "/openai/deployments/" + c.deploymentID
	}
	
	url := baseURL + endpoint
	
	// Add API version
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}
	url = url + separator + "api-version=" + c.apiVersion
	
	return url
}

// sendRequest sends HTTP request with Azure-specific authentication
func (c *AzureOpenAIClient) sendRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.buildAzureURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Azure uses "api-key" header instead of "Authorization: Bearer"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp openaiErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error (%s): %s", resp.Status, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(body))
	}

	return body, nil
}

// sendStreamingRequest sends streaming HTTP request with Azure authentication
func (c *AzureOpenAIClient) sendStreamingRequest(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.buildAzureURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(body))
	}

	return resp, nil
}

// processStreamingResponse processes streaming response (same as OpenAI)
func (c *AzureOpenAIClient) processStreamingResponse(resp *http.Response, writer io.Writer) (string, []domain.ToolCall, error) {
	defer resp.Body.Close()

	var fullContent string
	toolCallMap := make(map[int]*openaiToolCall)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp openaiStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		delta := streamResp.Choices[0].Delta

		if delta.Content != "" {
			fullContent += delta.Content
			if writer != nil {
				writer.Write([]byte(delta.Content))
			}
		}

		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				idx := 0
				
				if _, exists := toolCallMap[idx]; !exists {
					toolCallMap[idx] = &openaiToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: openaiToolFunction{
							Name:      "",
							Arguments: "",
						},
					}
				}

				currentCall := toolCallMap[idx]
				if tc.ID != "" {
					currentCall.ID = tc.ID
				}
				if tc.Type != "" {
					currentCall.Type = tc.Type
				}
				if tc.Function.Name != "" {
					currentCall.Function.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					currentCall.Function.Arguments += tc.Function.Arguments
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent, nil, fmt.Errorf("streaming error: %w", err)
	}

	var toolCalls []domain.ToolCall
	if len(toolCallMap) > 0 {
		openaiToolCalls := make([]openaiToolCall, 0, len(toolCallMap))
		for _, tc := range toolCallMap {
			if tc.Function.Name != "" {
				openaiToolCalls = append(openaiToolCalls, *tc)
			}
		}
		toolCalls = convertFromOpenAIToolCalls(openaiToolCalls)
	}

	return fullContent, toolCalls, nil
}
