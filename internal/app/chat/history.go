package chat

import (
	"context"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// History manages conversation history with context window management
type History struct {
	session       *Session
	maxMessages   int
	maxTokens     int
	reserveTokens int
}

// NewHistory creates a new history manager
func NewHistory(session *Session, maxMessages int, maxTokens int) *History {
	return &History{
		session:       session,
		maxMessages:   maxMessages,
		maxTokens:     maxTokens,
		reserveTokens: 1000, // Reserve for response
	}
}

// AddMessage adds a message to history
func (h *History) AddMessage(msg models.Message) {
	h.session.AddMessage(msg)
	h.trimIfNeeded()
}

// GetMessages returns messages for the current request
func (h *History) GetMessages() []models.Message {
	return h.session.GetMessages()
}

// GetRecentMessages returns the N most recent messages
func (h *History) GetRecentMessages(n int) []models.Message {
	messages := h.session.GetMessages()
	if len(messages) <= n {
		return messages
	}
	return messages[len(messages)-n:]
}

// trimIfNeeded trims history if it exceeds limits
func (h *History) trimIfNeeded() {
	messages := h.session.GetMessages()

	// Trim by message count
	if h.maxMessages > 0 && len(messages) > h.maxMessages {
		// Keep system message if present
		var systemMsg *models.Message
		if len(messages) > 0 && messages[0].Role == models.RoleSystem {
			systemMsg = &messages[0]
			messages = messages[1:]
		}

		// Keep most recent messages
		trimCount := len(messages) - h.maxMessages + 1
		messages = messages[trimCount:]

		// Reconstruct with system message
		if systemMsg != nil {
			messages = append([]models.Message{*systemMsg}, messages...)
		}

		h.session.Conversation.Messages = messages
	}

	// TODO: Implement token-based trimming
	// This would require actual token counting per provider
}

// Clear clears all history
func (h *History) Clear() {
	h.session.Clear()
}

// GetTokenCount returns estimated token count
func (h *History) GetTokenCount() int {
	return h.session.GetTotalTokens()
}

// CanFitMessage checks if a message would fit in context
func (h *History) CanFitMessage(msg models.Message) bool {
	if h.maxTokens == 0 {
		return true
	}

	currentTokens := h.GetTokenCount()
	estimatedMsgTokens := len(msg.Content) / 4

	return (currentTokens + estimatedMsgTokens + h.reserveTokens) <= h.maxTokens
}

// SummarizeIfNeeded summarizes old messages if approaching limit
func (h *History) SummarizeIfNeeded(ctx context.Context, summarizer MessageSummarizer) error {
	// TODO: Implement summarization strategy
	// When history gets too long, summarize older messages
	return nil
}

// MessageSummarizer interface for summarizing messages
type MessageSummarizer interface {
	Summarize(ctx context.Context, messages []models.Message) (string, error)
}
