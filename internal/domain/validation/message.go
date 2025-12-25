package validation

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/errors"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// ValidateMessage validates a message
func ValidateMessage(msg *models.Message) error {
	if msg == nil {
		return errors.NewDomainError(errors.ErrCodeRequestInvalid, "message is nil")
	}

	if !msg.Role.IsValid() {
		return errors.NewDomainError(errors.ErrCodeRequestInvalid, "invalid role").
			WithContext("role", msg.Role)
	}

	// Must have either content or tool calls or be a tool result
	if msg.Content == "" && len(msg.ToolCalls) == 0 && msg.ToolCallID == "" {
		return errors.NewDomainError(errors.ErrCodeRequestInvalid,
			"message must have content, tool calls, or be a tool result")
	}

	return nil
}

// ValidateConversation validates a conversation
func ValidateConversation(conv *models.Conversation) error {
	if conv == nil {
		return errors.NewDomainError(errors.ErrCodeRequestInvalid, "conversation is nil")
	}

	if len(conv.Messages) == 0 {
		return errors.NewDomainError(errors.ErrCodeRequestInvalid, "conversation has no messages")
	}

	for i, msg := range conv.Messages {
		if err := ValidateMessage(&msg); err != nil {
			return errors.NewDomainError(errors.ErrCodeRequestInvalid,
				"invalid message in conversation").
				WithContext("index", i).
				WithCause(err)
		}
	}

	return nil
}
