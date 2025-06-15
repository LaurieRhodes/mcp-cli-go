package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Manager handles tool processing and execution across all providers
type Manager struct {
	// Future: Could add tool registrations, caching, etc.
}

// NewManager creates a new tool manager
func NewManager() *Manager {
	return &Manager{}
}

// ParseToolCalls parses and validates tool calls from any provider format
func (m *Manager) ParseToolCalls(providerType domain.ProviderType, data interface{}) ([]domain.ToolCall, error) {
	switch providerType {
	case domain.ProviderAnthropic:
		return m.parseAnthropicToolCalls(data)
	case domain.ProviderOpenAI, domain.ProviderDeepSeek, domain.ProviderGemini:
		return m.parseOpenAIToolCalls(data)
	case domain.ProviderOllama:
		return m.parseOllamaToolCalls(data)
	default:
		return nil, fmt.Errorf("unsupported provider type for tool call parsing: %s", providerType)
	}
}

// ConvertToProviderFormat converts tool calls between provider formats
func (m *Manager) ConvertToProviderFormat(toolCalls []domain.ToolCall, providerType domain.ProviderType) (interface{}, error) {
	switch providerType {
	case domain.ProviderAnthropic:
		return m.convertToAnthropicFormat(toolCalls)
	case domain.ProviderOpenAI, domain.ProviderDeepSeek, domain.ProviderGemini:
		return m.convertToOpenAIFormat(toolCalls)
	case domain.ProviderOllama:
		return m.convertToOllamaFormat(toolCalls)
	default:
		return nil, fmt.Errorf("unsupported provider type for tool call conversion: %s", providerType)
	}
}

// ValidateToolCall validates a tool call against schema
func (m *Manager) ValidateToolCall(toolCall domain.ToolCall, schema map[string]interface{}) error {
	if toolCall.ID == "" {
		return fmt.Errorf("tool call missing ID")
	}
	
	if toolCall.Type != "function" {
		return fmt.Errorf("unsupported tool call type: %s", toolCall.Type)
	}
	
	if toolCall.Function.Name == "" {
		return fmt.Errorf("tool call missing function name")
	}
	
	// Validate that arguments is valid JSON
	var args map[string]interface{}
	if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
		return fmt.Errorf("tool call arguments are not valid JSON: %w", err)
	}
	
	// Validate against schema if provided
	if schema != nil {
		if err := m.validateAgainstSchema(args, schema); err != nil {
			return fmt.Errorf("tool call arguments validation failed: %w", err)
		}
	}
	
	return nil
}

// GenerateToolCallID generates consistent tool call IDs
func (m *Manager) GenerateToolCallID(providerType domain.ProviderType, index int) string {
	switch providerType {
	case domain.ProviderAnthropic:
		return fmt.Sprintf("toolu_%s", m.generateRandomID(8))
	case domain.ProviderOpenAI, domain.ProviderDeepSeek, domain.ProviderGemini:
		return fmt.Sprintf("call_%s", m.generateRandomID(8))
	case domain.ProviderOllama:
		return fmt.Sprintf("tc_%d", index)
	default:
		return fmt.Sprintf("tool_%d", index)
	}
}

// ValidateTools validates a list of tool definitions
func (m *Manager) ValidateTools(tools []domain.Tool) error {
	for i, tool := range tools {
		if err := m.ValidateTool(tool); err != nil {
			return fmt.Errorf("tool %d validation failed: %w", i, err)
		}
	}
	return nil
}

// ValidateTool validates a single tool definition
func (m *Manager) ValidateTool(tool domain.Tool) error {
	if tool.Type != "function" && tool.Type != "" {
		return fmt.Errorf("unsupported tool type: %s", tool.Type)
	}
	
	if tool.Function.Name == "" {
		return fmt.Errorf("tool missing function name")
	}
	
	if tool.Function.Description == "" {
		logging.Warn("Tool %s missing description", tool.Function.Name)
	}
	
	// Validate parameters schema
	if tool.Function.Parameters != nil {
		if _, ok := tool.Function.Parameters["type"]; !ok {
			return fmt.Errorf("tool %s parameters missing 'type' field", tool.Function.Name)
		}
	}
	
	return nil
}

// ProcessToolCalls processes a list of tool calls and returns results
// This is a placeholder for future unified tool execution logic
func (m *Manager) ProcessToolCalls(ctx context.Context, toolCalls []domain.ToolCall) (map[string]string, error) {
	results := make(map[string]string)
	
	for _, toolCall := range toolCalls {
		if err := m.ValidateToolCall(toolCall, nil); err != nil {
			logging.Error("Invalid tool call %s: %v", toolCall.ID, err)
			results[toolCall.ID] = fmt.Sprintf("Error: %v", err)
			continue
		}
		
		// For now, just log the tool call
		logging.Info("Processing tool call: %s (%s)", toolCall.Function.Name, toolCall.ID)
		results[toolCall.ID] = "Tool execution not implemented in consolidated manager"
	}
	
	return results, nil
}

// CreateToolResponseMessage creates a tool response message
func (m *Manager) CreateToolResponseMessage(toolCallID, result string) domain.Message {
	return domain.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	}
}

// FormatToolError formats a tool execution error
func (m *Manager) FormatToolError(toolName string, err error) string {
	return fmt.Sprintf("Error executing tool %s: %v", toolName, err)
}

// Provider-specific parsing methods

// parseAnthropicToolCalls parses Anthropic-format tool calls
func (m *Manager) parseAnthropicToolCalls(data interface{}) ([]domain.ToolCall, error) {
	// Anthropic uses "tool_use" blocks in content
	switch v := data.(type) {
	case []interface{}:
		var toolCalls []domain.ToolCall
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "tool_use" {
					toolCall, err := m.parseAnthropicToolUse(itemMap)
					if err != nil {
						logging.Warn("Failed to parse Anthropic tool use: %v", err)
						continue
					}
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}
		return toolCalls, nil
	case map[string]interface{}:
		if v["type"] == "tool_use" {
			toolCall, err := m.parseAnthropicToolUse(v)
			if err != nil {
				return nil, err
			}
			return []domain.ToolCall{toolCall}, nil
		}
	}
	return nil, fmt.Errorf("unsupported Anthropic tool call format")
}

// parseAnthropicToolUse parses a single Anthropic tool_use block
func (m *Manager) parseAnthropicToolUse(data map[string]interface{}) (domain.ToolCall, error) {
	id, _ := data["id"].(string)
	name, _ := data["name"].(string)
	input := data["input"]
	
	if id == "" || name == "" {
		return domain.ToolCall{}, fmt.Errorf("missing required fields in Anthropic tool use")
	}
	
	// Convert input to JSON
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return domain.ToolCall{}, fmt.Errorf("failed to marshal Anthropic tool input: %w", err)
	}
	
	return domain.ToolCall{
		ID:   id,
		Type: "function",
		Function: domain.Function{
			Name:      name,
			Arguments: inputBytes,
		},
	}, nil
}

// parseOpenAIToolCalls parses OpenAI-format tool calls
func (m *Manager) parseOpenAIToolCalls(data interface{}) ([]domain.ToolCall, error) {
	switch v := data.(type) {
	case []interface{}:
		var toolCalls []domain.ToolCall
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				toolCall, err := m.parseOpenAIToolCall(itemMap)
				if err != nil {
					logging.Warn("Failed to parse OpenAI tool call: %v", err)
					continue
				}
				toolCalls = append(toolCalls, toolCall)
			}
		}
		return toolCalls, nil
	case map[string]interface{}:
		toolCall, err := m.parseOpenAIToolCall(v)
		if err != nil {
			return nil, err
		}
		return []domain.ToolCall{toolCall}, nil
	}
	return nil, fmt.Errorf("unsupported OpenAI tool call format")
}

// parseOpenAIToolCall parses a single OpenAI tool call
func (m *Manager) parseOpenAIToolCall(data map[string]interface{}) (domain.ToolCall, error) {
	id, _ := data["id"].(string)
	toolType, _ := data["type"].(string)
	
	functionData, ok := data["function"].(map[string]interface{})
	if !ok {
		return domain.ToolCall{}, fmt.Errorf("missing function data in OpenAI tool call")
	}
	
	name, _ := functionData["name"].(string)
	arguments, _ := functionData["arguments"].(string)
	
	if id == "" || name == "" {
		return domain.ToolCall{}, fmt.Errorf("missing required fields in OpenAI tool call")
	}
	
	return domain.ToolCall{
		ID:   id,
		Type: toolType,
		Function: domain.Function{
			Name:      name,
			Arguments: json.RawMessage(arguments),
		},
	}, nil
}

// parseOllamaToolCalls parses Ollama-format tool calls
func (m *Manager) parseOllamaToolCalls(data interface{}) ([]domain.ToolCall, error) {
	switch v := data.(type) {
	case []interface{}:
		var toolCalls []domain.ToolCall
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				toolCall, err := m.parseOllamaToolCall(itemMap)
				if err != nil {
					logging.Warn("Failed to parse Ollama tool call: %v", err)
					continue
				}
				toolCalls = append(toolCalls, toolCall)
			}
		}
		return toolCalls, nil
	case map[string]interface{}:
		toolCall, err := m.parseOllamaToolCall(v)
		if err != nil {
			return nil, err
		}
		return []domain.ToolCall{toolCall}, nil
	case string:
		// Handle XML format parsing
		return m.parseOllamaXMLToolCalls(v)
	}
	return nil, fmt.Errorf("unsupported Ollama tool call format")
}

// parseOllamaToolCall parses a single Ollama tool call
func (m *Manager) parseOllamaToolCall(data map[string]interface{}) (domain.ToolCall, error) {
	id, _ := data["id"].(string)
	toolType, _ := data["type"].(string)
	
	functionData, ok := data["function"].(map[string]interface{})
	if !ok {
		return domain.ToolCall{}, fmt.Errorf("missing function data in Ollama tool call")
	}
	
	name, _ := functionData["name"].(string)
	arguments := functionData["arguments"]
	
	if id == "" || name == "" {
		return domain.ToolCall{}, fmt.Errorf("missing required fields in Ollama tool call")
	}
	
	// Convert arguments to JSON
	argBytes, err := json.Marshal(arguments)
	if err != nil {
		return domain.ToolCall{}, fmt.Errorf("failed to marshal Ollama tool arguments: %w", err)
	}
	
	return domain.ToolCall{
		ID:   id,
		Type: toolType,
		Function: domain.Function{
			Name:      name,
			Arguments: argBytes,
		},
	}, nil
}

// parseOllamaXMLToolCalls parses Ollama XML-format tool calls
func (m *Manager) parseOllamaXMLToolCalls(content string) ([]domain.ToolCall, error) {
	var toolCalls []domain.ToolCall
	
	// Find all <tool_call> blocks
	start := 0
	for {
		startTag := strings.Index(content[start:], "<tool_call>")
		if startTag == -1 {
			break
		}
		startTag += start
		
		endTag := strings.Index(content[startTag:], "</tool_call>")
		if endTag == -1 {
			break
		}
		endTag += startTag
		
		// Extract the JSON content
		jsonContent := strings.TrimSpace(content[startTag+11 : endTag])
		
		// Parse the JSON
		var toolData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &toolData); err != nil {
			logging.Warn("Failed to parse Ollama XML tool call JSON: %v", err)
			start = endTag + 12
			continue
		}
		
		// Extract tool information
		name, _ := toolData["name"].(string)
		if name == "" {
			start = endTag + 12
			continue
		}
		
		// Generate ID
		id := fmt.Sprintf("alt_tc_%d", len(toolCalls))
		
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
		
		toolCall := domain.ToolCall{
			ID:   id,
			Type: "function",
			Function: domain.Function{
				Name:      name,
				Arguments: argsJSON,
			},
		}
		
		toolCalls = append(toolCalls, toolCall)
		start = endTag + 12
	}
	
	return toolCalls, nil
}

// Provider-specific conversion methods

// convertToAnthropicFormat converts tool calls to Anthropic format
func (m *Manager) convertToAnthropicFormat(toolCalls []domain.ToolCall) (interface{}, error) {
	var result []map[string]interface{}
	
	for _, toolCall := range toolCalls {
		var input interface{}
		if err := json.Unmarshal(toolCall.Function.Arguments, &input); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments for Anthropic format: %w", err)
		}
		
		anthropicToolCall := map[string]interface{}{
			"type":  "tool_use",
			"id":    toolCall.ID,
			"name":  toolCall.Function.Name,
			"input": input,
		}
		
		result = append(result, anthropicToolCall)
	}
	
	return result, nil
}

// convertToOpenAIFormat converts tool calls to OpenAI format
func (m *Manager) convertToOpenAIFormat(toolCalls []domain.ToolCall) (interface{}, error) {
	var result []map[string]interface{}
	
	for _, toolCall := range toolCalls {
		openaiToolCall := map[string]interface{}{
			"id":   toolCall.ID,
			"type": toolCall.Type,
			"function": map[string]interface{}{
				"name":      toolCall.Function.Name,
				"arguments": string(toolCall.Function.Arguments),
			},
		}
		
		result = append(result, openaiToolCall)
	}
	
	return result, nil
}

// convertToOllamaFormat converts tool calls to Ollama format
func (m *Manager) convertToOllamaFormat(toolCalls []domain.ToolCall) (interface{}, error) {
	var result []map[string]interface{}
	
	for _, toolCall := range toolCalls {
		var arguments interface{}
		if err := json.Unmarshal(toolCall.Function.Arguments, &arguments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments for Ollama format: %w", err)
		}
		
		ollamaToolCall := map[string]interface{}{
			"id":   toolCall.ID,
			"type": toolCall.Type,
			"function": map[string]interface{}{
				"name":      toolCall.Function.Name,
				"arguments": arguments,
			},
		}
		
		result = append(result, ollamaToolCall)
	}
	
	return result, nil
}

// Helper methods

// validateAgainstSchema validates arguments against a JSON schema
func (m *Manager) validateAgainstSchema(args map[string]interface{}, schema map[string]interface{}) error {
	// Basic schema validation - can be extended for more complex schemas
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil // No properties to validate
	}
	
	required, _ := schema["required"].([]interface{})
	requiredFields := make(map[string]bool)
	for _, field := range required {
		if fieldStr, ok := field.(string); ok {
			requiredFields[fieldStr] = true
		}
	}
	
	// Check required fields
	for field := range requiredFields {
		if _, exists := args[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	
	// Basic type checking
	for field, value := range args {
		if propSchema, exists := properties[field]; exists {
			if err := m.validateFieldType(field, value, propSchema); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateFieldType validates a field against its schema type
func (m *Manager) validateFieldType(field string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil // Skip validation if schema is not a map
	}
	
	expectedType, ok := schemaMap["type"].(string)
	if !ok {
		return nil // Skip validation if no type specified
	}
	
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' should be string, got %T", field, value)
		}
	case "number":
		switch value.(type) {
		case float64, int, int64:
			// Valid number types
		default:
			return fmt.Errorf("field '%s' should be number, got %T", field, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' should be boolean, got %T", field, value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("field '%s' should be array, got %T", field, value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' should be object, got %T", field, value)
		}
	}
	
	return nil
}

// generateRandomID generates a random ID for tool calls
func (m *Manager) generateRandomID(length int) string {
	// Simple implementation - in production, use a proper random generator
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
