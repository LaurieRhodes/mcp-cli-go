package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// ChatContext manages the state of a chat session
type ChatContext struct {
	// Messages in the conversation
	Messages []domain.Message

	// Tool call history
	ToolCalls []ToolCallHistory

	// System prompt template
	SystemPrompt string

	// Maximum number of messages to retain in history
	MaxHistorySize int

	// Maximum tokens to retain in context
	MaxTokens int
}

// ToolCallHistory tracks the execution of a tool
type ToolCallHistory struct {
	ToolCall  domain.ToolCall
	Result    string
	Timestamp time.Time
	Error     string
}

// NewChatContext creates a new chat context
func NewChatContext(systemPrompt string) *ChatContext {
	// If no system prompt provided, use default one
	if systemPrompt == "" {
		systemPrompt = `You are a helpful assistant with access to tools. The tools are provided by Model Context Protocol (MCP) servers.

When you need to perform actions such as searching for information, accessing files, or interacting with external systems, use the available tools.

To use a tool:
1. Consider if you need to use a tool to answer the user's request
2. Select the appropriate tool based on the tool description
3. Call the tool with the appropriate parameters
4. Wait for the tool to execute and return a result
5. Use the result to inform your response to the user

The tool names are in the format <server>_<tool>, for example 'filesystem_list_directory' for the list_directory tool on the filesystem server.

Format your response in a clear, helpful manner and always explain what tools you're using and why.

For file system interactions, make sure to respect file paths and check if operations succeeded.
`
	}

	return &ChatContext{
		Messages:      []domain.Message{},
		ToolCalls:     []ToolCallHistory{},
		SystemPrompt:  systemPrompt,
		MaxHistorySize: 50,  // Reasonable default for most models
		MaxTokens:     8000, // Approximating a 16k context window
	}
}

// AddMessage adds a message to the context
func (c *ChatContext) AddMessage(message domain.Message) {
	c.Messages = append(c.Messages, message)
	c.TrimHistory()
}

// AddToolCall adds a tool call to the history
func (c *ChatContext) AddToolCall(toolCall domain.ToolCall, result string, err error) {
	history := ToolCallHistory{
		ToolCall:  toolCall,
		Result:    result,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		history.Error = err.Error()
	}
	
	c.ToolCalls = append(c.ToolCalls, history)
}

// GetMessagesForLLM returns the messages to send to the LLM
func (c *ChatContext) GetMessagesForLLM() []domain.Message {
	// Start with system message
	messages := []domain.Message{
		{
			Role:    "system",
			Content: c.BuildSystemPrompt(),
		},
	}
	
	// Process conversation history to ensure correct message format
	// We need to make sure that 'tool' role messages properly reference their parent assistant messages
	var processedMessages []domain.Message
	
	// Track tool calls from assistant messages to associate them with tool responses
	toolCallIDToAssistantIndex := make(map[string]int)
	
	// First pass: add all non-tool messages and build the tool call ID mapping
	for _, msg := range c.Messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			// Register the tool call IDs with this assistant message
			for _, toolCall := range msg.ToolCalls {
				toolCallIDToAssistantIndex[toolCall.ID] = len(processedMessages)
			}
		}
		
		// For tool messages, ensure they reference a valid assistant message
		if msg.Role == "tool" && msg.ToolCallID != "" {
			// Find the assistant message index that contains this tool call
			if assistantIndex, exists := toolCallIDToAssistantIndex[msg.ToolCallID]; exists {
				// Only include the tool message if its parent assistant message is included
				if assistantIndex >= 0 {
					processedMessages = append(processedMessages, msg)
				} else {
					logging.Warn("Skipping tool message with ID %s as its parent assistant message was not found", msg.ToolCallID)
				}
			} else {
				// Skip tool messages that don't have a corresponding assistant message with tool calls
				logging.Warn("Skipping tool message with ID %s as no parent assistant message was found", msg.ToolCallID)
				continue
			}
		} else {
			// Add all non-tool messages
			processedMessages = append(processedMessages, msg)
		}
	}
	
	// Add the processed messages
	messages = append(messages, processedMessages...)
	
	// Log the message structure for debugging
	logging.Debug("Sending %d messages to LLM", len(messages))
	for i, msg := range messages {
		if i == 0 {
			logging.Debug("Message %d: role=%s, content=[system prompt]", i, msg.Role)
		} else {
			logging.Debug("Message %d: role=%s, tool_call_id=%s, content_len=%d", i, msg.Role, msg.ToolCallID, len(msg.Content))
		}
	}
	
	return messages
}

// BuildSystemPrompt builds the system prompt including tool descriptions
func (c *ChatContext) BuildSystemPrompt() string {
	// Base system prompt
	prompt := c.SystemPrompt
	
	// Add recent tool history if available
	if len(c.ToolCalls) > 0 {
		toolHistory := c.FormatToolHistoryForLLM()
		if toolHistory != "" {
			prompt += "\n\n" + toolHistory
		}
	}
	
	logging.Debug("Built system prompt: %s", prompt)
	return prompt
}

// TrimHistory trims the history if it exceeds the maximum size
func (c *ChatContext) TrimHistory() {
	// Simple implementation that truncates based on message count
	if len(c.Messages) > c.MaxHistorySize {
		logging.Debug("Trimming history from %d messages to %d messages", 
			len(c.Messages), c.MaxHistorySize)
		c.Messages = c.Messages[len(c.Messages)-c.MaxHistorySize:]
	}
	
	// TODO: Implement more sophisticated token counting and trimming
}

// FormatToolHistoryForLLM formats the tool history for the LLM
func (c *ChatContext) FormatToolHistoryForLLM() string {
	var history strings.Builder
	
	// Only include the most recent tool calls (last 5)
	recentCalls := c.ToolCalls
	if len(recentCalls) > 5 {
		recentCalls = recentCalls[len(recentCalls)-5:]
	}
	
	if len(recentCalls) == 0 {
		return ""
	}
	
	history.WriteString("Here are the results of recent tool calls:\n\n")
	
	for i, toolCall := range recentCalls {
		history.WriteString(fmt.Sprintf("Tool call %d:\n", i+1))
		history.WriteString(fmt.Sprintf("- Name: %s\n", toolCall.ToolCall.Function.Name))
		history.WriteString(fmt.Sprintf("- Arguments: %s\n", string(toolCall.ToolCall.Function.Arguments)))
		
		if toolCall.Error != "" {
			history.WriteString(fmt.Sprintf("- Error: %s\n", toolCall.Error))
		} else {
			// Truncate very long results
			result := toolCall.Result
			if len(result) > 500 {
				result = result[:500] + "... (truncated)"
			}
			history.WriteString(fmt.Sprintf("- Result: %s\n", result))
		}
		
		history.WriteString("\n")
	}
	
	return history.String()
}
