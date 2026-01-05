package chat

import (
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// Session represents an active chat session
type Session struct {
	ID           string
	Conversation *models.Conversation
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Metadata     map[string]interface{}
	
	// For future multi-user support
	UserID   string // Identifies the user (for authentication/auditing)
	ClientID string // Identifies the client connection (for multi-session per user)
}

// NewSession creates a new chat session
func NewSession(systemPrompt string) *Session {
	now := time.Now()
	conv := &models.Conversation{
		ID:           generateSessionID(),
		Messages:     []models.Message{},
		SystemPrompt: systemPrompt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return &Session{
		ID:           conv.ID,
		Conversation: conv,
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]interface{}),
		UserID:       "", // Will be set when authentication is implemented
		ClientID:     "", // Will be set when multi-session support is added
	}
}

// SetUser sets the user ID for this session
func (s *Session) SetUser(userID string) {
	s.UserID = userID
	s.Metadata["user_id"] = userID
	s.UpdatedAt = time.Now()
}

// SetClient sets the client ID for this session
func (s *Session) SetClient(clientID string) {
	s.ClientID = clientID
	s.Metadata["client_id"] = clientID
	s.UpdatedAt = time.Now()
}

// AddMessage adds a message to the session
func (s *Session) AddMessage(msg models.Message) {
	s.Conversation.AddMessage(msg)
	s.UpdatedAt = time.Now()
}

// GetMessages returns all messages in the session
func (s *Session) GetMessages() []models.Message {
	return s.Conversation.Messages
}

// GetLastMessage returns the last message in the session
func (s *Session) GetLastMessage() *models.Message {
	return s.Conversation.GetLastMessage()
}

// Clear clears all messages from the session
func (s *Session) Clear() {
	s.Conversation.Clear()
	s.UpdatedAt = time.Now()
}

// MessageCount returns the number of messages
func (s *Session) MessageCount() int {
	return s.Conversation.MessageCount()
}

// GetTotalTokens calculates total tokens used in session
func (s *Session) GetTotalTokens() int {
	total := 0
	for _, msg := range s.Conversation.Messages {
		// Approximate token count (will be replaced with actual usage tracking)
		total += len(msg.Content) / 4
	}
	return total
}

// Helper function to generate session IDs
func generateSessionID() string {
	return time.Now().Format("20060102-150405")
}
