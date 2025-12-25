package chat

import (
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
)

// ServerStream represents a connection to a server
type ServerStream struct {
	Name   string
	Client *stdio.StdioClient
}

// ChatContext maintains the state for a chat session
type ChatContext struct {
	// Server connections
	ServerStreams []ServerStream

	// LLM configuration
	Provider string
	Model    string
	
	// LLM client
	LLMClient domain.LLMProvider
	
	// Chat history
	ConversationHistory []domain.Message
	
	// Tool information
	ToolHistory  []ToolHistoryEntry
	AvailableTools []domain.Tool
	
	// Session state
	ExitRequested bool
}

// ToolHistoryEntry represents a record of a tool call
type ToolHistoryEntry struct {
	// The tool call that was made
	ToolCall domain.ToolCall `json:"tool_call"`
	
	// The response from the tool
	Response string `json:"response"`
	
	// The server that handled the tool call
	Server string `json:"server"`
}

// NewChatContext creates a new chat context
func NewChatContext(serverStreams []ServerStream, provider, model string) *ChatContext {
	return &ChatContext{
		ServerStreams:       serverStreams,
		Provider:            provider,
		Model:               model,
		ConversationHistory: []domain.Message{},
		ToolHistory:         []ToolHistoryEntry{},
		AvailableTools:      []domain.Tool{},
		ExitRequested:       false,
	}
}

// InitializeLLM initializes the LLM client
func (c *ChatContext) InitializeLLM(apiKey string) error {
	// This method is deprecated - LLM client should be injected
	return fmt.Errorf("InitializeLLM is deprecated - use dependency injection")
}

// AddUserMessage adds a user message to the conversation history
func (c *ChatContext) AddUserMessage(content string) {
	c.ConversationHistory = append(c.ConversationHistory, domain.Message{
		Role:    "user",
		Content: content,
	})
}

// AddAssistantMessage adds an assistant message to the conversation history
func (c *ChatContext) AddAssistantMessage(content string) {
	c.ConversationHistory = append(c.ConversationHistory, domain.Message{
		Role:    "assistant",
		Content: content,
	})
}

// AddToolCall adds a tool call to the conversation history
func (c *ChatContext) AddToolCall(toolCall domain.ToolCall) {
	c.ConversationHistory = append(c.ConversationHistory, domain.Message{
		Role:      "assistant",
		Content:   "",
		ToolCalls: []domain.ToolCall{toolCall},
	})
}

// AddToolResponse adds a tool response to the conversation history
func (c *ChatContext) AddToolResponse(name, content, toolCallID string) {
	c.ConversationHistory = append(c.ConversationHistory, domain.Message{
		Role:       "tool",
		Name:       name,
		Content:    content,
		ToolCallID: toolCallID,
	})
}

// AddToToolHistory adds a tool call and its response to the tool history
func (c *ChatContext) AddToToolHistory(toolCall domain.ToolCall, response, server string) {
	c.ToolHistory = append(c.ToolHistory, ToolHistoryEntry{
		ToolCall: toolCall,
		Response: response,
		Server:   server,
	})
}

// ClearHistory clears the conversation and tool history
func (c *ChatContext) ClearHistory() {
	c.ConversationHistory = []domain.Message{}
	c.ToolHistory = []ToolHistoryEntry{}
}

// SetExitRequested signals that the chat should exit
func (c *ChatContext) SetExitRequested(value bool) {
	c.ExitRequested = value
}

// ToDict converts the chat context to a map for display purposes
func (c *ChatContext) ToDict() map[string]interface{} {
	servers := make([]string, len(c.ServerStreams))
	for i, s := range c.ServerStreams {
		servers[i] = s.Name
	}
	
	return map[string]interface{}{
		"servers":  servers,
		"provider": c.Provider,
		"model":    c.Model,
		"tools":    len(c.AvailableTools),
	}
}
