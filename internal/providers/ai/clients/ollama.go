package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// OllamaClient implements the domain.LLMProvider interface for Ollama
type OllamaClient struct {
	client       *http.Client
	config       *domain.ProviderConfig
	providerType domain.ProviderType
}

// Internal structures for Ollama API communication
type ollamaChatMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	ToolCalls  []ollamaToolCall       `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunction `json:"function"`
}

type ollamaToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Tools    []ollamaTool        `json:"tools,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Model     string            `json:"model"`
	CreatedAt string            `json:"created_at"`
	Message   ollamaChatMessage `json:"message"`
	Done      bool              `json:"done"`
}

type ollamaChatChunk struct {
	Model     string            `json:"model"`
	CreatedAt string            `json:"created_at"`
	Message   ollamaChatMessage `json:"message"`
	Done      bool              `json:"done"`
}

// Internal structures for consistent tool call handling
type internalOllamaToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function internalOllamaFunction `json:"function"`
}

type internalOllamaFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Constants
const (
	defaultOllamaEndpoint = "http://localhost:11434"
	ollamaChatEndpoint    = "/api/chat"
)

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(config *domain.ProviderConfig) (domain.LLMProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("provider configuration is required")
	}

	// Set default endpoint if not provided
	endpoint := config.APIEndpoint
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}

	// Validate and fix endpoint format
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Set timeout
	timeout := time.Duration(config.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second // Default timeout for local models
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Fix the model name if needed
	model := fixOllamaModel(config.DefaultModel)
	config.DefaultModel = model

	logging.Info("Creating Ollama client with model: %s, endpoint: %s", model, endpoint)

	return &OllamaClient{
		client:       httpClient,
		config:       config,
		providerType: domain.ProviderOllama,
	}, nil
}

// CreateCompletion generates a completion using the specified request
func (c *OllamaClient) CreateCompletion(ctx context.Context, req *domain.CompletionRequest) (*domain.CompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("completion request is required")
	}

	// Convert domain request to Ollama format
	ollamaMessages := c.convertToOllamaMessages(req.Messages, req.SystemPrompt)
	ollamaTools := c.convertToOllamaTools(req.Tools)

	// Add tool follow-up clarification if needed
	if c.isToolResultFollowUp(req.Messages) {
		ollamaMessages = c.addToolFollowUpClarification(ollamaMessages, req.Messages)
	}

	// Prepare request
	ollamaReq := ollamaChatRequest{
		Model:    c.config.DefaultModel,
		Messages: ollamaMessages,
		Stream:   false,
		Tools:    ollamaTools,
		Options:  make(map[string]interface{}),
	}

	// Set temperature
	temperature := c.getTemperature(req.Temperature)
	if temperature > 0 {
		ollamaReq.Options["temperature"] = temperature
	}

	// Set max tokens if specified
	if req.MaxTokens > 0 {
		ollamaReq.Options["num_predict"] = req.MaxTokens
	}

	// Send request
	url := c.config.APIEndpoint + ollamaChatEndpoint
	response, err := c.sendRequest(ctx, url, ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Extract content and tool calls
	content, toolCalls := c.extractContentAndToolCalls(response)

	// Convert to domain format
	domainToolCalls := c.convertToDomainToolCalls(toolCalls)

	return &domain.CompletionResponse{
		Response:  content,
		ToolCalls: domainToolCalls,
		Model:     response.Model,
	}, nil
}

// StreamCompletion generates a streaming completion
func (c *OllamaClient) StreamCompletion(ctx context.Context, req *domain.CompletionRequest, writer io.Writer) (*domain.CompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("completion request is required")
	}

	// Convert domain request to Ollama format
	ollamaMessages := c.convertToOllamaMessages(req.Messages, req.SystemPrompt)
	ollamaTools := c.convertToOllamaTools(req.Tools)

	// Add tool follow-up clarification if needed
	if c.isToolResultFollowUp(req.Messages) {
		ollamaMessages = c.addToolFollowUpClarification(ollamaMessages, req.Messages)
	}

	// Prepare request
	ollamaReq := ollamaChatRequest{
		Model:    c.config.DefaultModel,
		Messages: ollamaMessages,
		Stream:   true,
		Tools:    ollamaTools,
		Options:  make(map[string]interface{}),
	}

	// Set temperature
	temperature := c.getTemperature(req.Temperature)
	if temperature > 0 {
		ollamaReq.Options["temperature"] = temperature
	}

	// Set max tokens if specified
	if req.MaxTokens > 0 {
		ollamaReq.Options["num_predict"] = req.MaxTokens
	}

	// Create callback for streaming
	callback := func(chunk string) error {
		if writer != nil {
			_, err := writer.Write([]byte(chunk))
			return err
		}
		return nil
	}

	// Send streaming request
	url := c.config.APIEndpoint + ollamaChatEndpoint
	content, toolCalls, err := c.sendStreamingRequest(ctx, url, ollamaReq, callback)
	if err != nil {
		return nil, fmt.Errorf("failed to send streaming request: %w", err)
	}

	// Convert to domain format
	domainToolCalls := c.convertToDomainToolCalls(toolCalls)

	return &domain.CompletionResponse{
		Response:  content,
		ToolCalls: domainToolCalls,
		Model:     c.config.DefaultModel,
	}, nil
}

// GetProviderType returns the type of this provider
func (c *OllamaClient) GetProviderType() domain.ProviderType {
	return c.providerType
}

// GetInterfaceType returns the interface type of this provider
func (c *OllamaClient) GetInterfaceType() domain.InterfaceType {
	return domain.OllamaNative
}

// ValidateConfig validates the provider configuration
func (c *OllamaClient) ValidateConfig() error {
	if c.config == nil {
		return fmt.Errorf("provider configuration is required")
	}

	if c.config.DefaultModel == "" {
		return fmt.Errorf("default model is required for Ollama provider")
	}

	return nil
}

// Close cleans up provider resources
func (c *OllamaClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}

// Helper methods

func (c *OllamaClient) getTemperature(requestTemp float64) float64 {
	if requestTemp > 0 {
		return requestTemp
	}
	if c.config.Temperature > 0 {
		return c.config.Temperature
	}
	return 0.7 // Default temperature
}

// fixOllamaModel ensures model names are handled correctly for Ollama
func fixOllamaModel(model string) string {
	logging.Debug("Fixing Ollama model name: %s", model)
	
	if model == "" {
		model = "ollama.com/ajindal/llama3.1-storm:8b"
		logging.Warn("Empty model name, using default: %s", model)
	} else if model == "llama3" {
		model = "ollama.com/ajindal/llama3.1-storm:8b"
		logging.Warn("Basic model 'llama3' doesn't support tools, using: %s", model)
	} else if !strings.Contains(model, ":") && !strings.Contains(model, "/") {
		model = model + ":8b"
		logging.Warn("Adding version to model: %s", model)
	}
	
	logging.Info("Using Ollama model: %s", model)
	return model
}

// convertToOllamaMessages converts domain messages to Ollama format
func (c *OllamaClient) convertToOllamaMessages(messages []domain.Message, systemPrompt string) []ollamaChatMessage {
	ollamaMessages := make([]ollamaChatMessage, 0)
	
	// Add system prompt if provided
	if systemPrompt != "" {
		ollamaMessages = append(ollamaMessages, ollamaChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			ollamaMessages = append(ollamaMessages, ollamaChatMessage{
				Role:    "system",
				Content: msg.Content,
			})
		case "tool":
			ollamaMessages = append(ollamaMessages, ollamaChatMessage{
				Role:       "tool",
				Content:    msg.Content,
				ToolCallID: msg.ToolCallID,
			})
		case "assistant":
			ollamaMessages = append(ollamaMessages, ollamaChatMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		default: // "user" or any other role
			ollamaMessages = append(ollamaMessages, ollamaChatMessage{
				Role:    "user",
				Content: msg.Content,
			})
		}
	}
	
	return ollamaMessages
}

// convertToOllamaTools converts domain tools to Ollama format
func (c *OllamaClient) convertToOllamaTools(tools []domain.Tool) []ollamaTool {
	ollamaTools := make([]ollamaTool, 0, len(tools))
	
	for _, tool := range tools {
		ollamaTool := ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
		
		ollamaTools = append(ollamaTools, ollamaTool)
	}
	
	return ollamaTools
}

// isToolResultFollowUp checks if the message sequence contains a tool result that needs follow-up
func (c *OllamaClient) isToolResultFollowUp(messages []domain.Message) bool {
	if len(messages) < 3 {
		return false
	}
	
	hasToolCall := false
	hasToolResult := false
	
	for i := 1; i < len(messages); i++ {
		if messages[i].Role == "assistant" && len(messages[i].ToolCalls) > 0 {
			hasToolCall = true
		}
		
		if messages[i].Role == "tool" {
			hasToolResult = true
		}
	}
	
	return hasToolCall && hasToolResult
}

// addToolFollowUpClarification adds clarification for tool follow-ups
func (c *OllamaClient) addToolFollowUpClarification(ollamaMessages []ollamaChatMessage, domainMessages []domain.Message) []ollamaChatMessage {
	var latestToolResult string
	var latestToolName string
	
	for i := len(domainMessages) - 1; i >= 0; i-- {
		if domainMessages[i].Role == "tool" {
			latestToolResult = domainMessages[i].Content
			
			// Search for the tool name in the previous assistant messages
			for j := i - 1; j >= 0; j-- {
				if domainMessages[j].Role == "assistant" && len(domainMessages[j].ToolCalls) > 0 {
					latestToolName = domainMessages[j].ToolCalls[0].Function.Name
					break
				}
			}
			break
		}
	}
	
	if latestToolResult != "" {
		clarificationContent := "I've executed the requested tool. "
		if latestToolName != "" {
			clarificationContent += fmt.Sprintf("Tool name: %s. ", latestToolName)
		}
		clarificationContent += "Please analyze this result and provide a complete response."
		
		clarificationMsg := ollamaChatMessage{
			Role:    "user",
			Content: clarificationContent,
		}
		
		ollamaMessages = append(ollamaMessages, clarificationMsg)
		logging.Debug("Added clarification message for tool follow-up")
	}
	
	return ollamaMessages
}

// sendRequest sends a request to the Ollama API
func (c *OllamaClient) sendRequest(ctx context.Context, url string, payload interface{}) (*ollamaChatResponse, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logging.Error("Ollama API error response body: %s", string(body))
		return nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result ollamaChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	return &result, nil
}

// sendStreamingRequest sends a streaming request to the Ollama API
func (c *OllamaClient) sendStreamingRequest(ctx context.Context, url string, payload interface{}, callback func(chunk string) error) (string, []internalOllamaToolCall, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("error marshaling request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("error sending streaming request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logging.Error("Ollama API streaming error response body: %s", string(body))
		return "", nil, fmt.Errorf("API returned error: %s - %s", resp.Status, string(body))
	}

	var fullContent strings.Builder
	var toolCalls []internalOllamaToolCall
	decoder := json.NewDecoder(resp.Body)
	
	for {
		var chunk ollamaChatChunk
		err := decoder.Decode(&chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullContent.String(), toolCalls, fmt.Errorf("error decoding stream: %w", err)
		}
		
		content := chunk.Message.Content
		if content != "" {
			fullContent.WriteString(content)
			
			if callback != nil {
				if err := callback(content); err != nil {
					return fullContent.String(), toolCalls, fmt.Errorf("callback error: %w", err)
				}
			}
		}
		
		// Process standard tool calls if they exist in the final chunk
		if len(chunk.Message.ToolCalls) > 0 && chunk.Done {
			for i, ollamaToolCall := range chunk.Message.ToolCalls {
				toolID := fmt.Sprintf("tc_%d", i)
				
				argsJSON, err := json.Marshal(ollamaToolCall.Function.Arguments)
				if err != nil {
					logging.Warn("Failed to marshal tool arguments to JSON: %v", err)
					argsJSON = []byte("{}")
				}
				
				toolCall := internalOllamaToolCall{
					ID:   toolID,
					Type: "function",
					Function: internalOllamaFunction{
						Name:      ollamaToolCall.Function.Name,
						Arguments: argsJSON,
					},
				}
				
				toolCalls = append(toolCalls, toolCall)
			}
		}
		
		if chunk.Done {
			break
		}
	}
	
	// Check for alternative tool call format if no standard tool calls were detected
	if len(toolCalls) == 0 && strings.Contains(fullContent.String(), "<tool_call>") {
		logging.Info("No standard tool calls found in streaming response, checking for alternative format")
		altToolCalls := c.parseAlternativeToolCalls(fullContent.String())
		
		if len(altToolCalls) > 0 {
			logging.Info("Successfully parsed %d alternative format tool calls from streaming content", len(altToolCalls))
			toolCalls = altToolCalls
			
			// Remove tool call text from content
			cleanContent := regexp.MustCompile(`<tool_call>.*?</tool_call>`).ReplaceAllString(fullContent.String(), "")
			cleanContent = strings.TrimSpace(cleanContent)
			
			if cleanContent != "" {
				fullContent = *new(strings.Builder)
				fullContent.WriteString(cleanContent)
			}
		}
	}
	
	return fullContent.String(), toolCalls, nil
}

// extractContentAndToolCalls extracts content and tool calls from an Ollama response
func (c *OllamaClient) extractContentAndToolCalls(response *ollamaChatResponse) (string, []internalOllamaToolCall) {
	var toolCalls []internalOllamaToolCall
	
	content := response.Message.Content
	
	// Extract tool calls if any - standard format
	if len(response.Message.ToolCalls) > 0 {
		logging.Debug("Found %d standard tool calls in response", len(response.Message.ToolCalls))
		
		for i, toolCall := range response.Message.ToolCalls {
			toolID := fmt.Sprintf("tc_%d", i)
			
			args, err := json.Marshal(toolCall.Function.Arguments)
			if err != nil {
				logging.Warn("Failed to marshal tool arguments to JSON: %v", err)
				args = []byte("{}")
			}
			
			newToolCall := internalOllamaToolCall{
				ID:   toolID,
				Type: "function",
				Function: internalOllamaFunction{
					Name:      toolCall.Function.Name,
					Arguments: args,
				},
			}
			
			toolCalls = append(toolCalls, newToolCall)
		}
	} else if strings.Contains(content, "<tool_call>") {
		// Fallback for alternative tool call format
		logging.Info("No standard tool calls found, checking for alternative format")
		altToolCalls := c.parseAlternativeToolCalls(content)
		
		if len(altToolCalls) > 0 {
			logging.Info("Successfully parsed %d alternative format tool calls", len(altToolCalls))
			toolCalls = altToolCalls
			
			// Remove tool call text from content
			content = regexp.MustCompile(`<tool_call>.*?</tool_call>`).ReplaceAllString(content, "")
			content = strings.TrimSpace(content)
		}
	}
	
	return content, toolCalls
}

// parseAlternativeToolCalls parses tool calls from alternative XML-like format
func (c *OllamaClient) parseAlternativeToolCalls(content string) []internalOllamaToolCall {
	var toolCalls []internalOllamaToolCall
	
	// Use regex to find tool calls in the content
	toolCallRegex := regexp.MustCompile(`<tool_call>\s*(\{[^}]*\})\s*</tool_call>`)
	matches := toolCallRegex.FindAllStringSubmatch(content, -1)
	
	for i, match := range matches {
		if len(match) > 1 {
			// Try to parse the JSON content
			var toolData map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &toolData); err != nil {
				logging.Warn("Failed to parse alternative tool call JSON: %v", err)
				continue
			}
			
			// Extract tool name and arguments
			name, _ := toolData["name"].(string)
			if name == "" {
				continue
			}
			
			// Convert arguments to JSON
			var argsJSON []byte
			if args, ok := toolData["arguments"].(map[string]interface{}); ok {
				var err error
				argsJSON, err = json.Marshal(args)
				if err != nil {
					argsJSON = []byte("{}")
				}
			} else {
				argsJSON = []byte("{}")
			}
			
			toolCall := internalOllamaToolCall{
				ID:   fmt.Sprintf("alt_tc_%d", i),
				Type: "function",
				Function: internalOllamaFunction{
					Name:      name,
					Arguments: argsJSON,
				},
			}
			
			toolCalls = append(toolCalls, toolCall)
		}
	}
	
	return toolCalls
}

// convertToDomainToolCalls converts internal tool calls to domain format
func (c *OllamaClient) convertToDomainToolCalls(toolCalls []internalOllamaToolCall) []domain.ToolCall {
	domainToolCalls := make([]domain.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		domainToolCalls[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: domain.Function{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return domainToolCalls
}
