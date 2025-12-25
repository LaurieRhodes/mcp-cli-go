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
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai/streaming"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

const (
	// Base URLs for the Anthropic API
	anthropicBaseURL = "https://api.anthropic.com/v1/messages"
	
	// API version header
	anthropicAPIVersion = "2023-06-01"
	
	// Default settings
	defaultMaxTokens = 4096
	defaultTimeout   = 300 * time.Second
	defaultMaxRetries = 5
)

// List of supported Claude models
var supportedClaudeModels = map[string]bool{
	"claude-3-opus-20240229":     true,
	"claude-3-sonnet-20240229":   true,
	"claude-3-haiku-20240307":    true,
	"claude-3-5-sonnet-20240620": true,
	"claude-3-7-sonnet-20250219": true,
}

// AnthropicClient implements the domain.LLMProvider interface for Claude
type AnthropicClient struct {
	client     *http.Client
	model      string
	apiKey     string
	config     *config.ProviderConfig
	timeout    time.Duration
	maxRetries int
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(cfg *config.ProviderConfig) (domain.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if !strings.HasPrefix(cfg.APIKey, "sk-ant-") {
		logging.Warn("Anthropic API key should start with 'sk-ant-'")
	}

	// Format model name to ensure it uses the correct format
	model := formatClaudeModel(cfg.DefaultModel)
	
	// Verify that we're using a supported model
	if !supportedClaudeModels[model] {
		logging.Warn("Using possibly unsupported Claude model: %s", model)
		logging.Info("Supported models are: %v", supportedClaudeModels)
	}
	
	logging.Info("Creating Anthropic client with model: %s", model)

	// Set timeout from config or use default
	timeout := defaultTimeout
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	// Set max retries from config or use default
	maxRetries := defaultMaxRetries
	if cfg.MaxRetries > 0 {
		maxRetries = cfg.MaxRetries
	}

	// Create an HTTP client with extended timeouts
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &AnthropicClient{
		client:     httpClient,
		model:      model,
		apiKey:     cfg.APIKey,
		config:     cfg,
		timeout:    timeout,
		maxRetries: maxRetries,
	}, nil
}

// CreateCompletion generates a completion using the Anthropic API
func (c *AnthropicClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert domain request to internal format
	messages := convertDomainMessages(req.Messages)
	tools := convertDomainTools(req.Tools)

	// Convert our message format to Anthropic's format
	anthropicMessages, systemPrompt := c.convertToAnthropicMessages(messages, req.SystemPrompt)
	
	// Convert our tool format to Anthropic's format if tools are provided
	var anthropicTools []map[string]interface{}
	if len(tools) > 0 {
		anthropicTools = c.convertToAnthropicTools(tools)
		logging.Debug("Converted %d tools to Anthropic format", len(tools))
	}

	// Create the request payload
	payload := map[string]interface{}{
		"model":      c.model,
		"messages":   anthropicMessages,
		"max_tokens": c.getMaxTokens(req.MaxTokens),
	}

	// Add system prompt if present
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	// Add tools if provided
	if len(anthropicTools) > 0 {
		payload["tools"] = anthropicTools
		payload["tool_choice"] = map[string]interface{}{
			"type": "auto",
		}
		logging.Debug("Added tools and tool_choice to request")
	}

	logging.Info("Sending request to Anthropic API with model %s", c.model)
	logging.Debug("Request details: %d messages, %d tools", len(req.Messages), len(tools))

	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Anthropic API request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// Call the Anthropic API
		response, err := c.sendRequest(ctx, payload, false)
		if err != nil {
			lastErr = fmt.Errorf("Anthropic API error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
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

		logging.Info("Successfully received response from Anthropic API")
		logging.Debug("Response content length: %d, Tool calls: %d", len(content), len(toolCalls))

		// Convert back to domain format
		domainToolCalls := convertToDomainToolCalls(toolCalls)

		return &domain.CompletionResponse{
			Response:  content,
			ToolCalls: domainToolCalls,
			Model:     c.model,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion generates a streaming completion
func (c *AnthropicClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// Convert domain request to internal format
	messages := convertDomainMessages(req.Messages)
	tools := convertDomainTools(req.Tools)

	// Convert our message format to Anthropic's format
	anthropicMessages, systemPrompt := c.convertToAnthropicMessages(messages, req.SystemPrompt)
	
	// Convert our tool format to Anthropic's format if tools are provided
	var anthropicTools []map[string]interface{}
	if len(tools) > 0 {
		anthropicTools = c.convertToAnthropicTools(tools)
		logging.Debug("Converted %d tools to Claude format", len(anthropicTools))
	}

	// Create the request payload
	payload := map[string]interface{}{
		"model":       c.model,
		"messages":    anthropicMessages,
		"max_tokens":  c.getMaxTokens(req.MaxTokens),
		"stream":      true,
	}

	// Add system prompt if present
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	// Add tools if provided
	if len(anthropicTools) > 0 {
		payload["tools"] = anthropicTools
		payload["tool_choice"] = map[string]interface{}{
			"type": "auto",
		}
		logging.Debug("Added tools and tool_choice to streaming request")
	}

	logging.Info("Starting streaming request to Anthropic API with model %s", c.model)
	logging.Debug("Stream request details: %d messages, %d tools", len(req.Messages), len(tools))

	// Create callback for streaming processor
	callback := func(chunk string) error {
		if writer != nil {
			_, err := writer.Write([]byte(chunk))
			return err
		}
		return nil
	}

	// Implement retry logic for streaming
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying Anthropic API streaming request (attempt %d/%d)", retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second)
		}

		// Call the Anthropic API with streaming
		response, err := c.sendRequest(ctx, payload, true)
		if err != nil {
			lastErr = fmt.Errorf("Anthropic API streaming error (attempt %d/%d): %w", retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue
		}

		// Process the streaming response using the streaming processor
		processor := streaming.NewAnthropicProcessor()
		fullContent, streamingToolCalls, streamErr := processor.ProcessStreamingResponse(response, callback)
		
		if streamErr != nil {
			lastErr = streamErr
			logging.Error("%v", lastErr)
			
			if strings.Contains(streamErr.Error(), "context deadline exceeded") ||
			   strings.Contains(streamErr.Error(), "connection reset by peer") {
				continue
			}
			break
		}

		logging.Info("Successfully completed streaming response from Anthropic API")
		logging.Info("Full content length: %d, Tool calls: %d", len(fullContent), len(streamingToolCalls))

		// Convert streaming tool calls to domain format
		domainToolCalls := convertStreamingToDomainToolCalls(streamingToolCalls)

		return &domain.CompletionResponse{
			Response:  fullContent,
			ToolCalls: domainToolCalls,
			Model:     c.model,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// CreateEmbeddings - Anthropic doesn't support embeddings, return error
func (c *AnthropicClient) CreateEmbeddings(ctx context.Context, req *domain.EmbeddingRequest) (*domain.EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings are not supported by Anthropic provider - use OpenAI or compatible provider instead")
}

// GetSupportedEmbeddingModels returns empty list as Anthropic doesn't support embeddings
func (c *AnthropicClient) GetSupportedEmbeddingModels() []string {
	return []string{} // Anthropic doesn't support embeddings
}

// GetMaxEmbeddingTokens returns 0 as Anthropic doesn't support embeddings
func (c *AnthropicClient) GetMaxEmbeddingTokens(model string) int {
	return 0 // Anthropic doesn't support embeddings
}

// GetProviderType returns the provider type
func (c *AnthropicClient) GetProviderType() domain.ProviderType {
	return domain.ProviderAnthropic
}

// GetInterfaceType returns the interface type
func (c *AnthropicClient) GetInterfaceType() config.InterfaceType {
	return config.AnthropicNative
}

// ValidateConfig validates the provider configuration
func (c *AnthropicClient) ValidateConfig() error {
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
func (c *AnthropicClient) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}

// Helper methods

func (c *AnthropicClient) getMaxTokens(requestMaxTokens int) int {
	if requestMaxTokens > 0 {
		return requestMaxTokens
	}
	if c.config.MaxTokens > 0 {
		return c.config.MaxTokens
	}
	return defaultMaxTokens
}

// sendRequest sends a request to the Anthropic API
func (c *AnthropicClient) sendRequest(ctx context.Context, payload map[string]interface{}, stream bool) (interface{}, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicBaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", anthropicAPIVersion)

	logging.Debug("Anthropic API request headers set")

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		logging.Error("Anthropic API error response body: %s", string(body))
		return nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(body))
	}

	// Handle the response differently based on whether we're streaming or not
	if stream {
		return resp, nil // Return the raw response for streaming
	}

	// For non-streaming, read and parse the response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the response JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	return result, nil
}

// formatClaudeModel ensures the model name uses the correct format
func formatClaudeModel(model string) string {
	// First check if the model already matches one of our supported models
	if supportedClaudeModels[model] {
		return model
	}
	
	// If not, try to normalize the model name
	if strings.Contains(model, "claude") {
		// Try different patterns
		if !strings.Contains(model, "-") {
			// Add required hyphens
			model = strings.Replace(model, "claude", "claude-", 1)
			if strings.Contains(model, "claude-3") && !strings.Contains(model, "haiku") && 
			   !strings.Contains(model, "opus") && !strings.Contains(model, "sonnet") {
				model = strings.Replace(model, "claude-3", "claude-3-sonnet", 1)
			}
		}
		
		// Add date suffix if missing
		if !strings.Contains(model, "-202") {
			// Default to most compatible version
			if strings.Contains(model, "claude-3-opus") {
				model = "claude-3-opus-20240229"
			} else if strings.Contains(model, "claude-3-sonnet") {
				model = "claude-3-sonnet-20240229"
			} else if strings.Contains(model, "claude-3-haiku") {
				model = "claude-3-haiku-20240307"
			} else if strings.Contains(model, "claude-3-5-sonnet") {
				model = "claude-3-5-sonnet-20240620"
			} else if strings.Contains(model, "claude-3-7-sonnet") {
				model = "claude-3-7-sonnet-20250219"
			}
		}
	}
	
	return model
}

// extractContentAndToolCalls extracts content and tool calls from an Anthropic response
func (c *AnthropicClient) extractContentAndToolCalls(response interface{}) (string, []internalToolCall) {
	var content string
	var toolCalls []internalToolCall

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		logging.Error("Invalid response format, expected map[string]interface{}")
		return "", nil
	}

	// Extract the content and check for tool calls in content blocks
	if contentBlocks, ok := responseMap["content"].([]interface{}); ok {
		for _, block := range contentBlocks {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockMap["type"] == "text" {
					if text, ok := blockMap["text"].(string); ok {
						content += text
					}
				} else if blockMap["type"] == "tool_use" {
					// Found a tool_use in content blocks
					id, _ := blockMap["id"].(string)
					name, _ := blockMap["name"].(string)
					
					var arguments string
					if input, ok := blockMap["input"].(map[string]interface{}); ok {
						argsJSON, err := json.Marshal(input)
						if err == nil {
							arguments = string(argsJSON)
						} else {
							arguments = "{}"
						}
					} else {
						arguments = "{}"
					}
					
					logging.Debug("Found tool_use in content block: %s (%s) with args: %s", name, id, arguments)
					toolCalls = append(toolCalls, internalToolCall{
						ID:   id,
						Type: "function",
						Function: internalFunction{
							Name:      name,
							Arguments: json.RawMessage(arguments),
						},
					})
				}
			}
		}
	}

	return content, toolCalls
}

// convertToAnthropicMessages converts messages to Anthropic's format
func (c *AnthropicClient) convertToAnthropicMessages(messages []internalMessage, systemPrompt string) ([]map[string]interface{}, string) {
	anthropicMessages := make([]map[string]interface{}, 0)
	var systemContent string
	
	// Add system prompt from request if provided
	if systemPrompt != "" {
		systemContent = systemPrompt + ""
	}
	
	// Filter out system messages as they are handled differently in Anthropic's API
	var nonSystemMessages []internalMessage
	
	for _, msg := range messages {
		if msg.Role == "system" {
			systemContent += msg.Content + ""
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}
	
	// Convert non-system messages
	for _, msg := range nonSystemMessages {
		role := msg.Role
		
		// Handle tool results correctly for Claude
		if msg.Role == "tool" {
			// For Claude API, tool results must be in USER messages
			anthropicMsg := map[string]interface{}{
				"role": "user",
			}
			
			toolResult := map[string]interface{}{
				"type":        "tool_result",
				"tool_use_id": msg.ToolCallID,
				"content":     msg.Content,
			}
			
			anthropicMsg["content"] = []map[string]interface{}{toolResult}
			anthropicMessages = append(anthropicMessages, anthropicMsg)
			
			logging.Debug("Converting tool result to user message with tool_result content block")
		} else if msg.Role == "assistant" {
			// For assistant messages, handle both text content and tool calls
			anthropicMsg := map[string]interface{}{
				"role": role,
			}
			
			// Create content blocks array for the assistant message
			contentBlocks := []map[string]interface{}{}
			
			// Add text content if it exists
			if msg.Content != "" {
				contentBlocks = append(contentBlocks, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}
			
			// Add tool_use blocks for each tool call in the message
			for _, toolCall := range msg.ToolCalls {
				// Parse the function arguments as JSON
				var input map[string]interface{}
				if err := json.Unmarshal(toolCall.Function.Arguments, &input); err != nil {
					input = map[string]interface{}{}
					logging.Warn("Failed to parse tool arguments: %v", err)
				}
				
				// Create the tool_use block
				toolUseBlock := map[string]interface{}{
					"type":  "tool_use",
					"id":    toolCall.ID,
					"name":  toolCall.Function.Name,
					"input": input,
				}
				
				contentBlocks = append(contentBlocks, toolUseBlock)
				logging.Debug("Adding tool_use block to assistant message: %s (%s)", toolCall.Function.Name, toolCall.ID)
			}
			
			anthropicMsg["content"] = contentBlocks
			anthropicMessages = append(anthropicMessages, anthropicMsg)
		} else {
			// For regular messages (user)
			anthropicMsg := map[string]interface{}{
				"role": role,
			}
			
			if msg.Content != "" {
				anthropicMsg["content"] = []map[string]interface{}{
					{
						"type": "text",
						"text": msg.Content,
					},
				}
			}
			
			anthropicMessages = append(anthropicMessages, anthropicMsg)
		}
	}
	
	// Trim any trailing whitespace from system content
	systemContent = strings.TrimSpace(systemContent)
	
	return anthropicMessages, systemContent
}

// convertToAnthropicTools converts tools to Anthropic's format
func (c *AnthropicClient) convertToAnthropicTools(tools []internalTool) []map[string]interface{} {
	if len(tools) == 0 {
		return nil
	}
	
	anthropicTools := make([]map[string]interface{}, 0, len(tools))
	for i, tool := range tools {
		if tool.Type != "function" && tool.Type != "" {
			logging.Warn("Skipping tool with non-function type: %s", tool.Type)
			continue
		}
		
		// Get properties from parameters
		var properties map[string]interface{}
		var required []string
		
		if props, ok := tool.Function.Parameters["properties"].(map[string]interface{}); ok {
			properties = props
		}
		
		if req, ok := tool.Function.Parameters["required"].([]interface{}); ok {
			required = make([]string, len(req))
			for i, r := range req {
				if strValue, ok := r.(string); ok {
					required[i] = strValue
				}
			}
		} else if req, ok := tool.Function.Parameters["required"].([]string); ok {
			required = req
		}
		
		if properties == nil {
			properties = make(map[string]interface{})
		}
		
		anthropicTool := map[string]interface{}{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
			"input_schema": map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		}
		
		logging.Debug("Tool %d: %s", i, tool.Function.Name)
		anthropicTools = append(anthropicTools, anthropicTool)
	}
	
	return anthropicTools
}

// Internal types for compatibility
type internalMessage struct {
	Role       string               `json:"role"`
	Content    string               `json:"content,omitempty"`
	Name       string               `json:"name,omitempty"`
	ToolCalls  []internalToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
}

type internalToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function internalFunction `json:"function"`
}

type internalFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type internalTool struct {
	Type     string               `json:"type"`
	Function internalToolFunction `json:"function"`
}

type internalToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Conversion functions between domain and internal types
func convertDomainMessages(domainMessages []domain.Message) []internalMessage {
	messages := make([]internalMessage, len(domainMessages))
	for i, msg := range domainMessages {
		messages[i] = internalMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}
		
		// Convert tool calls
		if len(msg.ToolCalls) > 0 {
			messages[i].ToolCalls = make([]internalToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				messages[i].ToolCalls[j] = internalToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: internalFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}
	return messages
}

func convertDomainTools(domainTools []domain.Tool) []internalTool {
	tools := make([]internalTool, len(domainTools))
	for i, tool := range domainTools {
		tools[i] = internalTool{
			Type: tool.Type,
			Function: internalToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return tools
}

func convertToDomainToolCalls(internalToolCalls []internalToolCall) []domain.ToolCall {
	toolCalls := make([]domain.ToolCall, len(internalToolCalls))
	for i, tc := range internalToolCalls {
		toolCalls[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: domain.Function{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return toolCalls
}

// convertStreamingToDomainToolCalls converts streaming tool calls to domain format
func convertStreamingToDomainToolCalls(streamingToolCalls []streaming.ToolCall) []domain.ToolCall {
	toolCalls := make([]domain.ToolCall, len(streamingToolCalls))
	for i, tc := range streamingToolCalls {
		toolCalls[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: domain.Function{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return toolCalls
}
