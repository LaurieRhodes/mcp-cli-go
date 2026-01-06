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

// OpenAI API request/response structures
type openaiMessage struct {
	Role       string                   `json:"role"`
	Content    string                   `json:"content,omitempty"`
	Name       string                   `json:"name,omitempty"`
	ToolCalls  []openaiToolCall         `json:"tool_calls,omitempty"`
	ToolCallID string                   `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

type openaiChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Tools       []openaiTool    `json:"tools,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openaiChatResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []openaiChoice   `json:"choices"`
	Usage   openaiUsage      `json:"usage,omitempty"`
}

type openaiChoice struct {
	Index        int             `json:"index"`
	Message      openaiMessage   `json:"message,omitempty"`
	Delta        openaiMessage   `json:"delta,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openaiStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
}

type openaiEmbeddingRequest struct {
	Input          interface{} `json:"input"` // string or []string
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	User           string      `json:"user,omitempty"`
}

type openaiEmbeddingResponse struct {
	Object string                `json:"object"`
	Data   []openaiEmbeddingData `json:"data"`
	Model  string                `json:"model"`
	Usage  openaiUsage           `json:"usage"`
}

type openaiEmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type openaiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// OpenAICompatibleClient implements domain.LLMProvider for OpenAI-compatible APIs
type OpenAICompatibleClient struct {
	httpClient   *http.Client
	model        string
	apiKey       string
	apiEndpoint  string
	providerType domain.ProviderType
	config       *config.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible provider
func NewOpenAICompatibleClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s", providerType)
	}

	model := cfg.DefaultModel
	if model == "" {
		return nil, fmt.Errorf("no model specified for %s", providerType)
	}

	// Use provided endpoint or default to OpenAI
	apiEndpoint := cfg.APIEndpoint
	if apiEndpoint == "" {
		apiEndpoint = "https://api.openai.com/v1"
		logging.Warn("No API endpoint provided for %s, defaulting to OpenAI: %s", providerType, apiEndpoint)
	}

	// Remove trailing slash
	apiEndpoint = strings.TrimSuffix(apiEndpoint, "/")

	logging.Info("Creating %s client with model: %s, endpoint: %s", providerType, model, apiEndpoint)

	// Set timeout and retries from config
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

	return &OpenAICompatibleClient{
		httpClient:   httpClient,
		model:        model,
		apiKey:       cfg.APIKey,
		apiEndpoint:  apiEndpoint,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *OpenAICompatibleClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert domain messages to OpenAI format
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)
	
	// Convert domain tools to OpenAI format
	tools := convertToOpenAITools(req.Tools)

	// Create request payload
	payload := openaiChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Sending request to %s API with model %s", c.providerType, c.model)
	logging.Debug("Request details: %d messages, %d tools", len(req.Messages), len(req.Tools))

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying %s API request (attempt %d/%d)", c.providerType, retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		response, err := c.sendRequest(ctx, "/chat/completions", payload)
		if err != nil {
			lastErr = fmt.Errorf("%s API error (attempt %d/%d): %w", c.providerType, retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		// Parse response
		var chatResp openaiChatResponse
		if err := json.Unmarshal(response, &chatResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			logging.Error("%v", lastErr)
			continue
		}

		if len(chatResp.Choices) == 0 {
			lastErr = fmt.Errorf("no completion choices returned")
			logging.Error("%v", lastErr)
			continue
		}

		choice := chatResp.Choices[0].Message
		toolCalls := convertFromOpenAIToolCalls(choice.ToolCalls)

		logging.Info("Successfully received response from %s API", c.providerType)

		return &domain.CompletionResponse{
			Response:  choice.Content,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}


// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Retry on network/timeout errors
	if strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") {
		return true
	}
	
	// Retry on server errors (5xx) and rate limiting (429)
	if strings.Contains(errStr, "500 Internal Server Error") ||
		strings.Contains(errStr, "502 Bad Gateway") ||
		strings.Contains(errStr, "503 Service Unavailable") ||
		strings.Contains(errStr, "504 Gateway Timeout") ||
		strings.Contains(errStr, "429 Too Many Requests") {
		return true
	}
	
	// Do NOT retry on client errors (4xx except 429)
	if strings.Contains(errStr, "400 Bad Request") ||
		strings.Contains(errStr, "401 Unauthorized") ||
		strings.Contains(errStr, "403 Forbidden") ||
		strings.Contains(errStr, "404 Not Found") {
		return false
	}
	
	// Default: don't retry unknown errors
	return false
}

// StreamCompletion implements domain.LLMProvider
func (c *OpenAICompatibleClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
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

	logging.Info("Starting streaming request to %s API with model %s", c.providerType, c.model)

	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying %s API streaming request (attempt %d/%d)", c.providerType, retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		resp, err := c.sendStreamingRequest(ctx, "/chat/completions", payload)
		if err != nil {
			lastErr = fmt.Errorf("%s API streaming error (attempt %d/%d): %w", c.providerType, retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			
			// Don't retry client errors (4xx except 429)
			if !isRetryableError(err) {
				logging.Error("Non-retryable error detected, failing immediately")
				break
			}
			continue
		}

		// Process streaming response
		fullContent, toolCalls, streamErr := c.processStreamingResponse(resp, writer)
		if streamErr != nil {
			lastErr = streamErr
			
			// Only retry on retryable errors
			if isRetryableError(streamErr) {
				continue
			}
			break
		}

		logging.Info("Successfully completed streaming response from %s API", c.providerType)

		return &domain.CompletionResponse{
			Response:  fullContent,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// CreateEmbeddings implements domain.LLMProvider
func (c *OpenAICompatibleClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required for embeddings")
	}

	// Use configured embedding model or fallback to default
	model := req.Model
	if model == "" && c.config.DefaultEmbeddingModel != "" {
		model = c.config.DefaultEmbeddingModel
	}
	if model == "" {
		return nil, fmt.Errorf("no embedding model specified")
	}

	logging.Info("Sending embeddings request to %s API with model %s for %d inputs", c.providerType, model, len(req.Input))

	// Create embedding request
	payload := openaiEmbeddingRequest{
		Input: req.Input,
		Model: model,
	}

	if req.EncodingFormat != "" {
		payload.EncodingFormat = req.EncodingFormat
	}

	if req.User != "" {
		payload.User = req.User
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying %s embeddings API request (attempt %d/%d)", c.providerType, retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		responseData, err := c.sendRequest(ctx, "/embeddings", payload)
		if err != nil {
			lastErr = fmt.Errorf("%s embeddings API error (attempt %d/%d): %w", c.providerType, retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		// Parse response
		var embResp openaiEmbeddingResponse
		if err := json.Unmarshal(responseData, &embResp); err != nil {
			lastErr = fmt.Errorf("failed to parse embeddings response: %w", err)
			logging.Error("%v", lastErr)
			continue
		}

		logging.Info("Successfully received embeddings response from %s API: %d embeddings", c.providerType, len(embResp.Data))

		// Convert to domain format
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
func (c *OpenAICompatibleClient) GetSupportedEmbeddingModels() []string {
	if c.config.EmbeddingModels != nil && len(c.config.EmbeddingModels) > 0 {
		var models []string
		for model := range c.config.EmbeddingModels {
			models = append(models, model)
		}
		return models
	}

	// Default OpenAI embedding models
	return []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"text-embedding-ada-002",
	}
}

// GetMaxEmbeddingTokens implements domain.LLMProvider
func (c *OpenAICompatibleClient) GetMaxEmbeddingTokens(model string) int {
	if c.config.EmbeddingModels != nil {
		if modelConfig, exists := c.config.EmbeddingModels[model]; exists {
			return modelConfig.MaxTokens
		}
	}

	// Default token limits for OpenAI embedding models
	modelLower := strings.ToLower(model)
	switch {
	case strings.Contains(modelLower, "text-embedding-3"):
		return 8191
	case strings.Contains(modelLower, "ada-002"):
		return 8191
	default:
		return 8191
	}
}

// GetProviderType implements domain.LLMProvider
func (c *OpenAICompatibleClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType implements domain.LLMProvider
func (c *OpenAICompatibleClient) GetInterfaceType() config.InterfaceType {
	return config.OpenAICompatible
}

// ValidateConfig implements domain.LLMProvider
func (c *OpenAICompatibleClient) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("configuration is required")
	}

	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if c.config.DefaultModel == "" {
		return fmt.Errorf("default model is required")
	}

	return nil
}

// Close implements domain.LLMProvider
func (c *OpenAICompatibleClient) Close() error {
	return nil
}

// HTTP helper methods

func (c *OpenAICompatibleClient) sendRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.apiEndpoint + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Azure endpoints use "api-key" header, others use "Authorization: Bearer"
	if c.isAzureEndpoint() {
		req.Header.Set("api-key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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
		// Try to parse error response
		var errResp openaiErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error (%s): %s", resp.Status, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(body))
	}

	return body, nil
}

func (c *OpenAICompatibleClient) sendStreamingRequest(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.apiEndpoint + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Azure endpoints use "api-key" header, others use "Authorization: Bearer"
	if c.isAzureEndpoint() {
		req.Header.Set("api-key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	
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

// isAzureEndpoint checks if the endpoint is an Azure endpoint
func (c *OpenAICompatibleClient) isAzureEndpoint() bool {
	return strings.Contains(c.apiEndpoint, ".openai.azure.com") ||
		strings.Contains(c.apiEndpoint, ".services.ai.azure.com")
}

func (c *OpenAICompatibleClient) processStreamingResponse(resp *http.Response, writer io.Writer) (string, []domain.ToolCall, error) {
	defer resp.Body.Close()

	var fullContent string
	toolCallMap := make(map[int]*openaiToolCall)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp openaiStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			logging.Warn("Failed to parse streaming chunk: %v", err)
			continue
		}

		if len(streamResp.Choices) == 0 {
			continue
		}

		delta := streamResp.Choices[0].Delta

		// Handle content
		if delta.Content != "" {
			fullContent += delta.Content
			if writer != nil {
				writer.Write([]byte(delta.Content))
			}
		}

		// Handle tool calls
		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				idx := 0 // OpenAI doesn't provide index in streaming, use position
				
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

	// Convert accumulated tool calls
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

// Conversion helper methods (package-level, shared with Azure client)

func convertToOpenAIMessages(messages []domain.Message, systemPrompt string) []openaiMessage {
	openaiMessages := make([]openaiMessage, 0, len(messages)+1)

	if systemPrompt != "" {
		openaiMessages = append(openaiMessages, openaiMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, msg := range messages {
		openaiMsg := openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		if len(msg.ToolCalls) > 0 {
			var openaiToolCalls []openaiToolCall
			for _, toolCall := range msg.ToolCalls {
				openaiToolCalls = append(openaiToolCalls, openaiToolCall{
					ID:   toolCall.ID,
					Type: toolCall.Type,
					Function: openaiToolFunction{
						Name:      toolCall.Function.Name,
						Arguments: string(toolCall.Function.Arguments),
					},
				})
			}
			openaiMsg.ToolCalls = openaiToolCalls
		}

		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	return openaiMessages
}

func convertToOpenAITools(tools []domain.Tool) []openaiTool {
	if len(tools) == 0 {
		return nil
	}

	openaiTools := make([]openaiTool, len(tools))
	for i, tool := range tools {
		openaiTools[i] = openaiTool{
			Type: "function",
			Function: map[string]interface{}{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			},
		}
	}

	return openaiTools
}

func convertFromOpenAIToolCalls(openaiToolCalls []openaiToolCall) []domain.ToolCall {
	if len(openaiToolCalls) == 0 {
		return nil
	}

	toolCalls := make([]domain.ToolCall, len(openaiToolCalls))
	for i, tc := range openaiToolCalls {
		args := tc.Function.Arguments
		if args == "" {
			args = "{}"
		}

		// Validate JSON
		var jsonCheck map[string]interface{}
		if err := json.Unmarshal([]byte(args), &jsonCheck); err != nil {
			logging.Warn("Invalid JSON in tool call arguments, using empty object: %v", err)
			args = "{}"
		}

		toolCalls[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: domain.Function{
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(args),
			},
		}
	}

	return toolCalls
}
