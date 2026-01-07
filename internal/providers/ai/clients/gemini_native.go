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

// GeminiNativeClient implements the domain.LLMProvider interface for Google's native Gemini API
type GeminiNativeClient struct {
	httpClient   *http.Client
	apiKey       string
	model        string
	providerType domain.ProviderType
	config       *config.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}

// Gemini native API structures
type geminiContent struct {
	Role  string        `json:"role"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text             string                 `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall    `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
	ThoughtSignature string                 `json:"thoughtSignature,omitempty"`
	Thought          bool                   `json:"thought,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent `json:"contents"`
	Tools            []geminiTool    `json:"tools,omitempty"`
	SystemInstruction *geminiContent `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsageMetadata `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content       geminiContent `json:"content"`
	FinishReason  string        `json:"finishReason,omitempty"`
	Index         int           `json:"index"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// NewGeminiNativeClient creates a new native Gemini client
func NewGeminiNativeClient(providerType domain.ProviderType, cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini")
	}

	// Get model or use default
	model := cfg.DefaultModel
	if model == "" {
		model = "gemini-2.0-flash-exp"
		logging.Warn("No model specified for Gemini, using default: %s", model)
	}

	logging.Info("Creating Gemini native client with model: %s", model)

	// Set timeout from config or use default
	timeout := 60 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	// Set max retries from config or use default
	maxRetries := 3
	if cfg.MaxRetries >= 0 {
		maxRetries = cfg.MaxRetries
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &GeminiNativeClient{
		httpClient:   httpClient,
		apiKey:       cfg.APIKey,
		model:        model,
		providerType: providerType,
		config:       cfg,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion implements domain.LLMProvider
func (c *GeminiNativeClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert domain messages to Gemini format
	contents, systemInstruction := convertToGeminiContents(req.Messages, req.SystemPrompt)
	
	// Convert domain tools to Gemini format
	var tools []geminiTool
	if len(req.Tools) > 0 {
		tools = []geminiTool{
			{FunctionDeclarations: convertToGeminiFunctionDeclarations(req.Tools)},
		}
	}

	// Create generation config
	var genConfig *geminiGenerationConfig
	if req.Temperature > 0 || req.MaxTokens > 0 {
		genConfig = &geminiGenerationConfig{}
		if req.Temperature > 0 {
			temp := req.Temperature
			genConfig.Temperature = &temp
		}
		if req.MaxTokens > 0 {
			genConfig.MaxOutputTokens = &req.MaxTokens
		}
	}

	// Create request payload
	payload := geminiGenerateContentRequest{
		Contents:         contents,
		Tools:            tools,
		SystemInstruction: systemInstruction,
		GenerationConfig: genConfig,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Sending request to Gemini native API with model %s", c.model)
	logging.Debug("Request details: %d contents, %d tools", len(contents), len(tools))

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Gemini API request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		response, err := c.sendRequest(ctx, payload, false)
		if err != nil {
			lastErr = fmt.Errorf("gemini API error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		if len(response.Candidates) == 0 {
			lastErr = fmt.Errorf("no candidates in response")
			logging.Error("%v", lastErr)
			continue
		}

		candidate := response.Candidates[0]
		
		// Convert response to domain format
		textContent, toolCalls := convertFromGeminiContent(candidate.Content)

		logging.Info("Successfully received response from Gemini API")

		return &domain.CompletionResponse{
			Response:  textContent,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion implements domain.LLMProvider
func (c *GeminiNativeClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// Convert domain messages to Gemini format
	contents, systemInstruction := convertToGeminiContents(req.Messages, req.SystemPrompt)
	
	// Convert domain tools to Gemini format
	var tools []geminiTool
	if len(req.Tools) > 0 {
		tools = []geminiTool{
			{FunctionDeclarations: convertToGeminiFunctionDeclarations(req.Tools)},
		}
	}

	// Create generation config
	var genConfig *geminiGenerationConfig
	if req.Temperature > 0 || req.MaxTokens > 0 {
		genConfig = &geminiGenerationConfig{}
		if req.Temperature > 0 {
			temp := req.Temperature
			genConfig.Temperature = &temp
		}
		if req.MaxTokens > 0 {
			genConfig.MaxOutputTokens = &req.MaxTokens
		}
	}

	// Create request payload
	payload := geminiGenerateContentRequest{
		Contents:         contents,
		Tools:            tools,
		SystemInstruction: systemInstruction,
		GenerationConfig: genConfig,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Starting streaming completion with Gemini native API")

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Gemini streaming API request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		textContent, toolCalls, err := c.processStreamingResponse(ctx, payload, writer)
		if err != nil {
			lastErr = fmt.Errorf("gemini API streaming error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		logging.Info("Successfully completed streaming response from Gemini API")

		return &domain.CompletionResponse{
			Response:  textContent,
			ToolCalls: toolCalls,
		}, nil
	}

	return nil, fmt.Errorf("streaming failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// sendRequest sends a request to the Gemini API
func (c *GeminiNativeClient) sendRequest(ctx context.Context, payload geminiGenerateContentRequest, stream bool) (*geminiGenerateContentResponse, error) {
	// Marshal the request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine endpoint
	endpoint := "generateContent"
	if stream {
		endpoint = "streamGenerateContent"
	}

	// Construct URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=%s",
		c.model, endpoint, c.apiKey)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%s): %s", resp.Status, string(responseBody))
	}

	// Parse the response
	var geminiResp geminiGenerateContentResponse
	if err := json.Unmarshal(responseBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &geminiResp, nil
}

// processStreamingResponse handles streaming responses from Gemini
func (c *GeminiNativeClient) processStreamingResponse(ctx context.Context, payload geminiGenerateContentRequest, writer io.Writer) (string, []domain.ToolCall, error) {
	// Marshal the request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Construct URL for streaming
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		c.model, c.apiKey)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", nil, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("API error (%s): %s", resp.Status, string(responseBody))
	}

	// Process SSE stream
	var fullText strings.Builder
	var toolCalls []domain.ToolCall
	var lastThoughtSignature string

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "" || jsonData == "[DONE]" {
			continue
		}

		// Parse the chunk
		var chunk geminiGenerateContentResponse
		if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
			logging.Warn("Failed to parse streaming chunk: %v", err)
			continue
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]

		// Process each part
		for _, part := range candidate.Content.Parts {
			// Handle text content
			if part.Text != "" && !part.Thought {
				fullText.WriteString(part.Text)
				if writer != nil {
					writer.Write([]byte(part.Text))
				}
			}

			// Handle function calls
			if part.FunctionCall != nil {
				toolCall := domain.ToolCall{
					ID:   fmt.Sprintf("call_%d", len(toolCalls)),
					Type: "function",
					Function: domain.Function{
						Name:      part.FunctionCall.Name,
						Arguments: marshalToRawJSON(part.FunctionCall.Args),
					},
				}
				toolCalls = append(toolCalls, toolCall)
			}

			// Save thought signature (needed for conversation continuity)
			if part.ThoughtSignature != "" {
				lastThoughtSignature = part.ThoughtSignature
				logging.Debug("Captured thought signature: %s", lastThoughtSignature)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, fmt.Errorf("error reading stream: %w", err)
	}

	return fullText.String(), toolCalls, nil
}

// convertToGeminiContents converts domain messages to Gemini format
func convertToGeminiContents(messages []domain.Message, systemPrompt string) ([]geminiContent, *geminiContent) {
	var contents []geminiContent
	var systemInstruction *geminiContent

	// Handle system prompt separately
	if systemPrompt != "" {
		systemInstruction = &geminiContent{
			Role: "user", // Gemini uses "user" role for system instructions
			Parts: []geminiPart{
				{Text: systemPrompt},
			},
		}
	}

	// Convert messages
	for _, msg := range messages {
		content := geminiContent{}

		// Map roles
		switch msg.Role {
		case "user":
			content.Role = "user"
		case "assistant":
			content.Role = "model"
		case "system":
			// System messages go to system instruction, skip in contents
			continue
		case "tool":
			// Tool results need special handling
			content.Role = "user"
			// Create function response part
			var responseData map[string]interface{}
			if msg.Content != "" {
				// Try to parse as JSON
				if err := json.Unmarshal([]byte(msg.Content), &responseData); err != nil {
					// If not JSON, wrap in response object
					responseData = map[string]interface{}{"result": msg.Content}
				}
			} else {
				responseData = map[string]interface{}{}
			}
			
			content.Parts = []geminiPart{
				{
					FunctionResponse: &geminiFunctionResponse{
						Name:     extractToolName(msg.ToolCallID),
						Response: responseData,
					},
				},
			}
			contents = append(contents, content)
			continue
		default:
			content.Role = "user"
		}

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			content.Role = "model"
			for _, tc := range msg.ToolCalls {
				var args map[string]interface{}
				if err := json.Unmarshal(tc.Function.Arguments, &args); err != nil {
					args = map[string]interface{}{}
				}
				
				part := geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: tc.Function.Name,
						Args: args,
					},
				}
				content.Parts = append(content.Parts, part)
			}
		} else if msg.Content != "" {
			// Regular text content
			content.Parts = []geminiPart{
				{Text: msg.Content},
			}
		}

		if len(content.Parts) > 0 {
			contents = append(contents, content)
		}
	}

	return contents, systemInstruction
}

// convertToGeminiFunctionDeclarations converts domain tools to Gemini function declarations
func convertToGeminiFunctionDeclarations(tools []domain.Tool) []geminiFunctionDeclaration {
	declarations := make([]geminiFunctionDeclaration, len(tools))
	
	for i, tool := range tools {
		declarations[i] = geminiFunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Parameters:  tool.Function.Parameters, // Direct pass-through - critical for Gemini
		}
		
		// Enhanced debugging for Gemini tool schema issues
		if logging.GetDefaultLevel() <= logging.DEBUG {
			logging.Debug("=== Gemini Tool Declaration ===")
			logging.Debug("  Name: %s", tool.Function.Name)
			logging.Debug("  Description: %s", tool.Function.Description)
			if schemaJSON, err := json.Marshal(tool.Function.Parameters); err == nil {
				logging.Debug("  Parameters (as-is from MCP): %s", string(schemaJSON))
			} else {
				logging.Warn("  Failed to marshal parameters: %v", err)
			}
			logging.Debug("===============================")
		}
	}
	
	return declarations
}

// convertFromGeminiContent converts Gemini response content to domain format
func convertFromGeminiContent(content geminiContent) (string, []domain.ToolCall) {
	var textParts []string
	var toolCalls []domain.ToolCall

	for _, part := range content.Parts {
		// Collect text parts (exclude thoughts)
		if part.Text != "" && !part.Thought {
			textParts = append(textParts, part.Text)
		}

		// Collect function calls
		if part.FunctionCall != nil {
			toolCall := domain.ToolCall{
				ID:   fmt.Sprintf("call_%d", len(toolCalls)),
				Type: "function",
				Function: domain.Function{
					Name:      part.FunctionCall.Name,
					Arguments: marshalToRawJSON(part.FunctionCall.Args),
				},
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	return strings.Join(textParts, ""), toolCalls
}

// Helper functions
func marshalToRawJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func extractToolName(toolCallID string) string {
	// Extract tool name from tool call ID if needed
	// For now, return a placeholder
	return "tool_function"
}

// CreateEmbeddings - Not supported by this client (use dedicated embedding client)
func (c *GeminiNativeClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Gemini native completion client, use dedicated embedding client")
}

// GetSupportedEmbeddingModels returns empty list (not supported)
func (c *GeminiNativeClient) GetSupportedEmbeddingModels() []string {
	return []string{}
}

// GetMaxEmbeddingTokens returns 0 (not supported)
func (c *GeminiNativeClient) GetMaxEmbeddingTokens(model string) int {
	return 0
}

// GetProviderType returns the provider type
func (c *GeminiNativeClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType returns the interface type
func (c *GeminiNativeClient) GetInterfaceType() config.InterfaceType {
	return config.GeminiNative
}

// ValidateConfig validates the provider configuration
func (c *GeminiNativeClient) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("configuration is required")
	}

	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required for Gemini")
	}

	return nil
}

// Close cleans up provider resources
func (c *GeminiNativeClient) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}
