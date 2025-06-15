package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/sashabaranov/go-openai"
)

// OpenAICompatibleClient implements the domain.LLMProvider interface for OpenAI-compatible providers
type OpenAICompatibleClient struct {
	client       *openai.Client
	model        string
	providerType domain.ProviderType
	config       *domain.ProviderConfig
	timeout      time.Duration
	maxRetries   int
}

// Provider endpoint mappings
var providerEndpoints = map[domain.ProviderType]string{
	domain.ProviderOpenAI:     "", // Uses default OpenAI endpoint
	domain.ProviderDeepSeek:   "https://api.deepseek.com/v1",
	domain.ProviderGemini:     "https://generativelanguage.googleapis.com/v1beta/openai/",
	domain.ProviderOpenRouter: "https://openrouter.ai/api/v1", // Added OpenRouter
}

// Default models for each provider
var defaultModels = map[domain.ProviderType]string{
	domain.ProviderOpenAI:     "gpt-4o",
	domain.ProviderDeepSeek:   "deepseek-chat",
	domain.ProviderGemini:     "gemini-1.5-pro",
	domain.ProviderOpenRouter: "openai/gpt-4o", // Default OpenRouter model
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible client
func NewOpenAICompatibleClient(providerType domain.ProviderType, config *domain.ProviderConfig) (domain.LLMProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s", providerType)
	}

	// Get model or use default
	model := config.DefaultModel
	if model == "" {
		if defaultModel, exists := defaultModels[providerType]; exists {
			model = defaultModel
			logging.Warn("No model specified for %s, using default: %s", providerType, model)
		} else {
			return nil, fmt.Errorf("no model specified and no default available for %s", providerType)
		}
	}

	logging.Info("Creating %s client with model: %s", providerType, model)

	// Set timeout from config or use default
	timeout := 60 * time.Second
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	// Set max retries from config or use default
	maxRetries := 3
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}

	// Create OpenAI client configuration
	clientConfig := openai.DefaultConfig(config.APIKey)
	
	// Set custom endpoint if needed
	if endpoint, exists := providerEndpoints[providerType]; exists && endpoint != "" {
		clientConfig.BaseURL = endpoint
		logging.Debug("Using custom endpoint for %s: %s", providerType, endpoint)
	} else if config.APIEndpoint != "" {
		clientConfig.BaseURL = config.APIEndpoint
		logging.Debug("Using configured endpoint: %s", config.APIEndpoint)
	}

	// Set timeout for HTTP client
	clientConfig.HTTPClient.Timeout = timeout

	// Special handling for different providers
	switch providerType {
	case domain.ProviderDeepSeek:
		// DeepSeek uses custom endpoint and may need special headers
		if clientConfig.BaseURL == "" {
			clientConfig.BaseURL = providerEndpoints[domain.ProviderDeepSeek]
		}
	case domain.ProviderGemini:
		// Gemini may need special configuration
		if clientConfig.BaseURL == "" {
			clientConfig.BaseURL = providerEndpoints[domain.ProviderGemini]
		}
	case domain.ProviderOpenRouter:
		// OpenRouter uses custom endpoint
		if clientConfig.BaseURL == "" {
			clientConfig.BaseURL = providerEndpoints[domain.ProviderOpenRouter]
		}
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAICompatibleClient{
		client:       client,
		model:        model,
		providerType: providerType,
		config:       config,
		timeout:      timeout,
		maxRetries:   maxRetries,
	}, nil
}

// CreateCompletion generates a completion using the OpenAI-compatible API
func (c *OpenAICompatibleClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	// Convert domain messages to OpenAI format
	openaiMessages := c.convertToOpenAIMessages(req.Messages, req.SystemPrompt)
	
	// Convert domain tools to OpenAI format
	openaiTools := c.convertToOpenAITools(req.Tools)

	// Create the request
	request := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		request.Temperature = float32(req.Temperature)
	}

	// Add max tokens if specified
	if req.MaxTokens > 0 {
		request.MaxTokens = req.MaxTokens
	} else if c.config.MaxTokens > 0 {
		request.MaxTokens = c.config.MaxTokens
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Sending request to %s API with model %s", c.providerType, c.model)
	logging.Debug("Request details: %d messages, %d tools", len(req.Messages), len(req.Tools))
	
	// Implement retry logic
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying %s API request (attempt %d/%d)", c.providerType, retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second) // Exponential backoff
		}
		
		// Call the API
		response, err := c.client.CreateChatCompletion(ctx, request)
		if err != nil {
			lastErr = fmt.Errorf("%s API error (attempt %d/%d): %w", c.providerType, retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue // Try again
		}

		// Process the response
		if len(response.Choices) == 0 {
			lastErr = fmt.Errorf("no completion choices returned")
			logging.Error("%v", lastErr)
			continue // Try again
		}

		choice := response.Choices[0].Message

		// Convert OpenAI's tool calls to domain format
		toolCalls := c.convertFromOpenAIToolCalls(choice.ToolCalls)

		logging.Info("Successfully received response from %s API", c.providerType)
		
		// Return domain CompletionResponse
		return &domain.CompletionResponse{
			Response:  choice.Content,
			ToolCalls: toolCalls,
			Model:     c.model,
			Usage: &domain.Usage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			},
		}, nil
	}
	
	// If we get here, all retries failed
	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// StreamCompletion generates a streaming completion
func (c *OpenAICompatibleClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	// Convert domain messages to OpenAI format
	openaiMessages := c.convertToOpenAIMessages(req.Messages, req.SystemPrompt)
	
	// Convert domain tools to OpenAI format
	openaiTools := c.convertToOpenAITools(req.Tools)

	// Create the streaming request
	request := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
		Tools:    openaiTools,
		Stream:   true, // Enable streaming
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		request.Temperature = float32(req.Temperature)
	}

	// Add max tokens if specified
	if req.MaxTokens > 0 {
		request.MaxTokens = req.MaxTokens
	} else if c.config.MaxTokens > 0 {
		request.MaxTokens = c.config.MaxTokens
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	logging.Info("Sending streaming request to %s API with model %s", c.providerType, c.model)
	logging.Debug("Streaming request details: %d messages, %d tools", len(req.Messages), len(req.Tools))
	
	// Implement retry logic for streaming
	var lastErr error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			logging.Warn("Retrying %s streaming API request (attempt %d/%d)", c.providerType, retry, c.maxRetries)
			time.Sleep(time.Duration(retry) * 2 * time.Second) // Exponential backoff
		}
		
		// Call the streaming API
		stream, err := c.client.CreateChatCompletionStream(ctx, request)
		if err != nil {
			lastErr = fmt.Errorf("%s streaming API error (attempt %d/%d): %w", c.providerType, retry+1, c.maxRetries+1, err)
			logging.Error("%v", lastErr)
			continue // Try again
		}
		defer stream.Close()

		// Process the streaming response
		var fullResponse strings.Builder
		var toolCalls []domain.ToolCall

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Stream finished normally
					break
				}
				lastErr = fmt.Errorf("%s streaming receive error: %w", c.providerType, err)
				logging.Error("%v", lastErr)
				break // Exit inner loop to retry
			}

			// Process each choice in the response
			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				delta := choice.Delta

				// Handle content delta
				if delta.Content != "" {
					fullResponse.WriteString(delta.Content)
					if writer != nil {
						writer.Write([]byte(delta.Content))
					}
				}

				// Handle tool calls delta
				if len(delta.ToolCalls) > 0 {
					for _, tcDelta := range delta.ToolCalls {
						// Handle the index properly (it's a pointer)
						var index int
						if tcDelta.Index != nil {
							index = *tcDelta.Index
						} else {
							index = 0 // Default to 0 if index is nil
						}

						// Ensure we have enough space in toolCalls slice
						for len(toolCalls) <= index {
							toolCalls = append(toolCalls, domain.ToolCall{})
						}

						tc := &toolCalls[index]
						
						// Update tool call with delta
						if tcDelta.ID != "" {
							tc.ID = tcDelta.ID
						}
						if tcDelta.Type != "" {
							tc.Type = string(tcDelta.Type)
						}
						// Function is a struct, not a pointer, so check field values instead of nil
						if tcDelta.Function.Name != "" {
							tc.Function.Name = tcDelta.Function.Name
						}
						if tcDelta.Function.Arguments != "" {
							// Append arguments (they come in chunks)
							existingArgs := string(tc.Function.Arguments)
							tc.Function.Arguments = json.RawMessage(existingArgs + tcDelta.Function.Arguments)
						}
					}
				}
			}
		}

		// If we got here without an error from the inner loop, we succeeded
		if lastErr == nil || !strings.Contains(lastErr.Error(), "streaming receive error") {
			logging.Info("Successfully received streaming response from %s API", c.providerType)
			
			// Validate and fix tool call arguments
			for i := range toolCalls {
				if len(toolCalls[i].Function.Arguments) == 0 {
					toolCalls[i].Function.Arguments = json.RawMessage("{}")
				} else {
					// Validate JSON
					var jsonCheck map[string]interface{}
					if err := json.Unmarshal(toolCalls[i].Function.Arguments, &jsonCheck); err != nil {
						logging.Warn("Invalid JSON in streaming tool call arguments, using empty object: %v", err)
						toolCalls[i].Function.Arguments = json.RawMessage("{}")
					}
				}
			}
			
			// Return domain CompletionResponse (streaming doesn't provide usage info)
			return &domain.CompletionResponse{
				Response:  fullResponse.String(),
				ToolCalls: toolCalls,
				Model:     c.model,
				Usage:     nil, // Streaming responses typically don't include usage
			}, nil
		}
	}
	
	// If we get here, all retries failed
	return nil, fmt.Errorf("streaming failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// GetProviderType returns the provider type
func (c *OpenAICompatibleClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType returns the interface type
func (c *OpenAICompatibleClient) GetInterfaceType() domain.InterfaceType {
	return domain.OpenAICompatible
}

// ValidateConfig validates the provider configuration
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

// Close cleans up provider resources
func (c *OpenAICompatibleClient) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}

// Helper methods for converting between domain and OpenAI formats

// convertToOpenAIMessages converts domain messages to OpenAI format
func (c *OpenAICompatibleClient) convertToOpenAIMessages(messages []domain.Message, systemPrompt string) []openai.ChatCompletionMessage {
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	
	// Add system prompt if provided
	if systemPrompt != "" {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}
	
	// Convert messages
	for _, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Handle tool calls if present
		if len(msg.ToolCalls) > 0 {
			var openaiToolCalls []openai.ToolCall
			for _, toolCall := range msg.ToolCalls {
				openaiToolCalls = append(openaiToolCalls, openai.ToolCall{
					ID:   toolCall.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      toolCall.Function.Name,
						Arguments: string(toolCall.Function.Arguments),
					},
				})
			}
			openaiMsg.ToolCalls = openaiToolCalls
		}

		// Add tool call ID if present
		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}
	
	return openaiMessages
}

// convertToOpenAITools converts domain tools to OpenAI format
func (c *OpenAICompatibleClient) convertToOpenAITools(tools []domain.Tool) []openai.Tool {
	if len(tools) == 0 {
		return nil
	}
	
	openaiTools := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		// Marshal parameters to JSON
		paramsJSON, err := json.Marshal(tool.Function.Parameters)
		if err != nil {
			// Log error but continue
			logging.Warn("Error marshaling parameters for tool %s: %v", tool.Function.Name, err)
			continue
		}

		logging.Debug("Tool %d: %s", i, tool.Function.Name)

		// Convert the tool type string to the proper type
		toolType := openai.ToolTypeFunction
		if tool.Type != "function" {
			logging.Warn("Unsupported tool type: %s, defaulting to function", tool.Type)
		}

		openaiTools[i] = openai.Tool{
			Type: toolType,
			Function: openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  json.RawMessage(paramsJSON),
			},
		}
	}
	
	return openaiTools
}

// convertFromOpenAIToolCalls converts OpenAI tool calls to domain format
func (c *OpenAICompatibleClient) convertFromOpenAIToolCalls(openaiToolCalls []openai.ToolCall) []domain.ToolCall {
	if len(openaiToolCalls) == 0 {
		return nil
	}
	
	toolCalls := make([]domain.ToolCall, len(openaiToolCalls))
	for i, tc := range openaiToolCalls {
		// Ensure there's a valid JSON for arguments
		args := tc.Function.Arguments
		if args == "" {
			args = "{}"
		}
		
		// Try to validate the arguments as JSON
		var jsonCheck map[string]interface{}
		if err := json.Unmarshal([]byte(args), &jsonCheck); err != nil {
			logging.Warn("Invalid JSON in tool call arguments, using empty object: %v", err)
			args = "{}"
		}
		
		toolCalls[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: string(tc.Type), // Convert ToolType to string
			Function: domain.Function{
				Name:      tc.Function.Name,
				Arguments: json.RawMessage(args),
			},
		}
	}
	
	return toolCalls
}
