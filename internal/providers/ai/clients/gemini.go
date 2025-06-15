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
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

const (
	// Base URL for the Gemini API
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"
	
	// Default settings
	defaultGeminiTimeout   = 60 * time.Second
	defaultGeminiMaxRetries = 3
)

// List of supported Gemini models
var supportedGeminiModels = map[string]bool{
	"gemini-1.5-pro":        true,
	"gemini-1.5-flash":      true,
	"gemini-1.0-pro":        true,
	"gemini-pro":            true,
	"gemini-pro-vision":     true,
}

// GeminiClient implements the domain.LLMProvider interface for Google's Gemini
type GeminiClient struct {
	client     *http.Client
	model      string
	apiKey     string
	config     *domain.ProviderConfig
	timeout    time.Duration
	maxRetries int
}

// Gemini API request/response structures
type GeminiRequest struct {
	Contents         []GeminiContent      `json:"contents"`
	Tools            []GeminiTool         `json:"tools,omitempty"`
	GenerationConfig *GeminiGenConfig     `json:"generationConfig,omitempty"`
	SystemInstruction *GeminiContent      `json:"systemInstruction,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text         string                `json:"text,omitempty"`
	FunctionCall *GeminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
}

type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type GeminiGenConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP           *float64 `json:"topP,omitempty"`
	TopK           *int     `json:"topK,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	UsageMetadata *GeminiUsage   `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content       GeminiContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
	Index         int           `json:"index"`
	SafetyRatings []interface{} `json:"safetyRatings,omitempty"`
}

type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(config *domain.ProviderConfig) (domain.LLMProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	// Format model name
	model := formatGeminiModel(config.DefaultModel)
	
	// Verify that we're using a supported model
	if !supportedGeminiModels[model] {
		logging.Warn("Using possibly unsupported Gemini model: %s", model)
		logging.Info("Supported models are: %v", getSupportedGeminiModelsList())
	}
	
	logging.Info("Creating Gemini client with model: %s", model)

	// Set timeout from config or use default
	timeout := defaultGeminiTimeout
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	// Set max retries from config or use default
	maxRetries := defaultGeminiMaxRetries
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}

	// Create an HTTP client with timeouts
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &GeminiClient{
		client:     httpClient,
		model:      model,
		apiKey:     config.APIKey,
		config:     config,
		timeout:    timeout,
		maxRetries: maxRetries,
	}, nil
}

// CreateCompletion generates a completion using the Gemini API
func (c *GeminiClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert domain request to Gemini format
	geminiReq := c.convertToGeminiRequest(req)

	logging.Info("Sending request to Gemini API with model %s", c.model)
	logging.Debug("Request details: %d contents, %d tools", len(geminiReq.Contents), len(geminiReq.Tools))

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Gemini API request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// Call the Gemini API
		response, err := c.sendRequest(ctx, geminiReq, false)
		if err != nil {
			lastErr = fmt.Errorf("Gemini API error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		// Process the response
		content, toolCalls := c.extractContentAndToolCalls(response)
		if content == "" && len(toolCalls) == 0 {
			lastErr = fmt.Errorf("no content or tool calls in response")
			logging.Error("%v", lastErr)
			continue
		}

		logging.Info("Successfully received response from Gemini API")
		logging.Debug("Response content length: %d, Tool calls: %d", len(content), len(toolCalls))

		// Convert usage information
		var usage *domain.Usage
		if response.UsageMetadata != nil {
			usage = &domain.Usage{
				PromptTokens:     response.UsageMetadata.PromptTokenCount,
				CompletionTokens: response.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      response.UsageMetadata.TotalTokenCount,
			}
		}

		return &domain.CompletionResponse{
			Response:  content,
			ToolCalls: toolCalls,
			Model:     c.model,
			Usage:     usage,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion generates a streaming completion (Gemini supports streaming)
func (c *GeminiClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// For now, implement streaming by calling the non-streaming version and writing output
	// Gemini does support streaming, but implementing the full SSE parsing would be complex
	// TODO: Implement native Gemini streaming support
	
	response, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	// Write the response to the writer if provided
	if writer != nil && response.Response != "" {
		writer.Write([]byte(response.Response))
	}

	return response, nil
}

// GetProviderType returns the provider type
func (c *GeminiClient) GetProviderType() domain.ProviderType {
	return domain.ProviderGemini
}

// GetInterfaceType returns the interface type
func (c *GeminiClient) GetInterfaceType() domain.InterfaceType {
	return domain.GeminiNative
}

// ValidateConfig validates the provider configuration
func (c *GeminiClient) ValidateConfig() error {
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

// Close cleans up provider resources
func (c *GeminiClient) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}

// Helper methods

// sendRequest sends a request to the Gemini API
func (c *GeminiClient) sendRequest(ctx context.Context, payload *GeminiRequest, stream bool) (*GeminiResponse, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request payload: %w", err)
	}

	// Construct the URL
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiBaseURL, c.model, c.apiKey)
	if stream {
		url += "&alt=sse"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")

	logging.Debug("Gemini API request URL: %s", url)

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logging.Error("Gemini API error response body: %s", string(body))
		return nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(body))
	}

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the response JSON
	var result GeminiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	return &result, nil
}

// convertToGeminiRequest converts domain request to Gemini format
func (c *GeminiClient) convertToGeminiRequest(req *domain.CompletionRequest) *GeminiRequest {
	geminiReq := &GeminiRequest{
		Contents: make([]GeminiContent, 0),
	}

	// Add system instruction if provided
	if req.SystemPrompt != "" {
		geminiReq.SystemInstruction = &GeminiContent{
			Parts: []GeminiPart{{Text: req.SystemPrompt}},
		}
	}

	// Convert messages to Gemini contents
	for _, msg := range req.Messages {
		geminiContent := c.convertMessageToGeminiContent(msg)
		if geminiContent != nil {
			geminiReq.Contents = append(geminiReq.Contents, *geminiContent)
		}
	}

	// Convert tools to Gemini format
	if len(req.Tools) > 0 {
		geminiReq.Tools = c.convertToGeminiTools(req.Tools)
	}

	// Set generation config
	if req.Temperature > 0 || req.MaxTokens > 0 {
		geminiReq.GenerationConfig = &GeminiGenConfig{}
		
		if req.Temperature > 0 {
			temp := req.Temperature
			geminiReq.GenerationConfig.Temperature = &temp
		}
		
		if req.MaxTokens > 0 {
			geminiReq.GenerationConfig.MaxOutputTokens = &req.MaxTokens
		}
	}

	return geminiReq
}

// convertMessageToGeminiContent converts a domain message to Gemini content
func (c *GeminiClient) convertMessageToGeminiContent(msg domain.Message) *GeminiContent {
	switch msg.Role {
	case "system":
		// System messages are handled separately in SystemInstruction
		return nil
	case "user":
		return &GeminiContent{
			Role:  "user",
			Parts: []GeminiPart{{Text: msg.Content}},
		}
	case "assistant":
		content := &GeminiContent{
			Role:  "model", // Gemini uses "model" instead of "assistant"
			Parts: make([]GeminiPart, 0),
		}
		
		// Add text content if present
		if msg.Content != "" {
			content.Parts = append(content.Parts, GeminiPart{Text: msg.Content})
		}
		
		// Add function calls if present
		for _, toolCall := range msg.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
				logging.Warn("Failed to parse tool call arguments: %v", err)
				args = make(map[string]interface{})
			}
			
			content.Parts = append(content.Parts, GeminiPart{
				FunctionCall: &GeminiFunctionCall{
					Name: toolCall.Function.Name,
					Args: args,
				},
			})
		}
		
		return content
	case "tool":
		// Convert tool results to function responses
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Content), &response); err != nil {
			// If content is not JSON, wrap it as text
			response = map[string]interface{}{
				"result": msg.Content,
			}
		}
		
		return &GeminiContent{
			Role: "function", // Gemini uses "function" for tool responses
			Parts: []GeminiPart{{
				FunctionResponse: &GeminiFunctionResponse{
					Name:     c.extractToolNameFromID(msg.ToolCallID),
					Response: response,
				},
			}},
		}
	default:
		logging.Warn("Unknown message role: %s", msg.Role)
		return nil
	}
}

// convertToGeminiTools converts domain tools to Gemini format
func (c *GeminiClient) convertToGeminiTools(tools []domain.Tool) []GeminiTool {
	if len(tools) == 0 {
		return nil
	}
	
	geminiTool := GeminiTool{
		FunctionDeclarations: make([]GeminiFunctionDeclaration, len(tools)),
	}
	
	for i, tool := range tools {
		geminiTool.FunctionDeclarations[i] = GeminiFunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Parameters:  tool.Function.Parameters,
		}
	}
	
	return []GeminiTool{geminiTool}
}

// extractContentAndToolCalls extracts content and tool calls from a Gemini response
func (c *GeminiClient) extractContentAndToolCalls(response *GeminiResponse) (string, []domain.ToolCall) {
	var content strings.Builder
	var toolCalls []domain.ToolCall

	if len(response.Candidates) == 0 {
		logging.Warn("No candidates in Gemini response")
		return "", nil
	}

	candidate := response.Candidates[0]
	
	for i, part := range candidate.Content.Parts {
		if part.Text != "" {
			content.WriteString(part.Text)
		}
		
		if part.FunctionCall != nil {
			// Convert function call to domain tool call
			argsJSON, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				logging.Warn("Failed to marshal function call args: %v", err)
				argsJSON = []byte("{}")
			}
			
			toolCall := domain.ToolCall{
				ID:   fmt.Sprintf("call_%d", i),
				Type: "function",
				Function: domain.Function{
					Name:      part.FunctionCall.Name,
					Arguments: json.RawMessage(argsJSON),
				},
			}
			
			toolCalls = append(toolCalls, toolCall)
		}
	}

	return content.String(), toolCalls
}

// extractToolNameFromID extracts tool name from tool call ID (simplified approach)
func (c *GeminiClient) extractToolNameFromID(toolCallID string) string {
	// This is a simplified approach - in a real implementation, you'd need to
	// maintain a mapping of tool call IDs to tool names
	return "unknown_tool"
}

// formatGeminiModel ensures the model name uses the correct format
func formatGeminiModel(model string) string {
	// If already a valid model, return as is
	if supportedGeminiModels[model] {
		return model
	}
	
	// Add some common transformations
	if model == "gemini" || model == "gemini-pro" {
		return "gemini-1.5-pro"
	}
	
	if model == "gemini-flash" {
		return "gemini-1.5-flash"
	}
	
	// If it contains "gemini" but not a recognized format, default to gemini-1.5-pro
	if strings.Contains(strings.ToLower(model), "gemini") {
		return "gemini-1.5-pro"
	}
	
	return model
}

// getSupportedGeminiModelsList returns a list of supported models
func getSupportedGeminiModelsList() []string {
	models := make([]string, 0, len(supportedGeminiModels))
	for model := range supportedGeminiModels {
		models = append(models, model)
	}
	return models
}
