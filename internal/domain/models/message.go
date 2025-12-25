package models

import "time"

// Message represents a single message in a conversation
type Message struct {
	ID        string    `json:"id,omitempty"`
	Role      Role      `json:"role"`
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	
	// Tool-related fields
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	
	// Optional metadata
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Role represents the role of a message sender
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem, RoleTool:
		return true
	default:
		return false
	}
}

// Conversation represents a sequence of messages
type Conversation struct {
	ID           string    `json:"id"`
	Messages     []Message `json:"messages"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// GetLastMessage returns the most recent message
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// FilterByRole returns messages with the specified role
func (c *Conversation) FilterByRole(role Role) []Message {
	var filtered []Message
	for _, msg := range c.Messages {
		if msg.Role == role {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// MessageCount returns the number of messages
func (c *Conversation) MessageCount() int {
	return len(c.Messages)
}

// Clear removes all messages from the conversation
func (c *Conversation) Clear() {
	c.Messages = []Message{}
	c.UpdatedAt = time.Now()
}
