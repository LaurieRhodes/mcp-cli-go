package validation

import (
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/errors"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		message *models.Message
		wantErr bool
		errCode errors.ErrorCode
	}{
		{
			name:    "nil message",
			message: nil,
			wantErr: true,
			errCode: errors.ErrCodeRequestInvalid,
		},
		{
			name: "valid user message",
			message: &models.Message{
				Role:    models.RoleUser,
				Content: "Hello",
			},
			wantErr: false,
		},
		{
			name: "valid assistant message with tool calls",
			message: &models.Message{
				Role: models.RoleAssistant,
				ToolCalls: []models.ToolCall{
					{ID: "call-1", Type: models.ToolTypeFunction},
				},
			},
			wantErr: false,
		},
		{
			name: "valid tool result",
			message: &models.Message{
				Role:       models.RoleTool,
				Content:    "result",
				ToolCallID: "call-1",
			},
			wantErr: false,
		},
		{
			name: "invalid role",
			message: &models.Message{
				Role:    models.Role("invalid"),
				Content: "test",
			},
			wantErr: true,
			errCode: errors.ErrCodeRequestInvalid,
		},
		{
			name: "empty message",
			message: &models.Message{
				Role: models.RoleUser,
			},
			wantErr: true,
			errCode: errors.ErrCodeRequestInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.message)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}

				domainErr, ok := err.(*errors.DomainError)
				if !ok {
					t.Errorf("Expected DomainError, got %T", err)
					return
				}

				if domainErr.Code != tt.errCode {
					t.Errorf("Expected error code %s, got %s", tt.errCode, domainErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateConversation(t *testing.T) {
	tests := []struct {
		name         string
		conversation *models.Conversation
		wantErr      bool
		errCode      errors.ErrorCode
	}{
		{
			name:         "nil conversation",
			conversation: nil,
			wantErr:      true,
			errCode:      errors.ErrCodeRequestInvalid,
		},
		{
			name: "empty conversation",
			conversation: &models.Conversation{
				ID:       "conv-1",
				Messages: []models.Message{},
			},
			wantErr: true,
			errCode: errors.ErrCodeRequestInvalid,
		},
		{
			name: "valid conversation",
			conversation: &models.Conversation{
				ID: "conv-1",
				Messages: []models.Message{
					{Role: models.RoleUser, Content: "Hello"},
					{Role: models.RoleAssistant, Content: "Hi there"},
				},
			},
			wantErr: false,
		},
		{
			name: "conversation with invalid message",
			conversation: &models.Conversation{
				ID: "conv-1",
				Messages: []models.Message{
					{Role: models.RoleUser, Content: "Hello"},
					{Role: models.Role("invalid"), Content: "Bad"},
				},
			},
			wantErr: true,
			errCode: errors.ErrCodeRequestInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConversation(tt.conversation)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}

				domainErr, ok := err.(*errors.DomainError)
				if !ok {
					t.Errorf("Expected DomainError, got %T", err)
					return
				}

				if domainErr.Code != tt.errCode {
					t.Errorf("Expected error code %s, got %s", tt.errCode, domainErr.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
