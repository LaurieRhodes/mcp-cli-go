package models

import (
	"testing"
	"time"
)

func TestRoleIsValid(t *testing.T) {
	tests := []struct {
		name  string
		role  Role
		valid bool
	}{
		{"user role", RoleUser, true},
		{"assistant role", RoleAssistant, true},
		{"system role", RoleSystem, true},
		{"tool role", RoleTool, true},
		{"invalid role", Role("invalid"), false},
		{"empty role", Role(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.valid {
				t.Errorf("Role.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestConversationAddMessage(t *testing.T) {
	conv := &Conversation{
		ID:        "test-conv",
		Messages:  []Message{},
		CreatedAt: time.Now(),
	}

	msg := Message{
		Role:    RoleUser,
		Content: "Hello",
	}

	conv.AddMessage(msg)

	if len(conv.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(conv.Messages))
	}

	if conv.Messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", conv.Messages[0].Content)
	}

	if conv.Messages[0].Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestConversationGetLastMessage(t *testing.T) {
	conv := &Conversation{Messages: []Message{}}

	// Test empty conversation
	if msg := conv.GetLastMessage(); msg != nil {
		t.Error("Expected nil for empty conversation")
	}

	// Add messages
	conv.AddMessage(Message{Role: RoleUser, Content: "First"})
	conv.AddMessage(Message{Role: RoleAssistant, Content: "Second"})

	last := conv.GetLastMessage()
	if last == nil {
		t.Fatal("Expected last message to not be nil")
	}

	if last.Content != "Second" {
		t.Errorf("Expected 'Second', got %s", last.Content)
	}
}

func TestConversationFilterByRole(t *testing.T) {
	conv := &Conversation{Messages: []Message{}}

	conv.AddMessage(Message{Role: RoleUser, Content: "User1"})
	conv.AddMessage(Message{Role: RoleAssistant, Content: "Assistant1"})
	conv.AddMessage(Message{Role: RoleUser, Content: "User2"})
	conv.AddMessage(Message{Role: RoleSystem, Content: "System1"})

	userMsgs := conv.FilterByRole(RoleUser)
	if len(userMsgs) != 2 {
		t.Errorf("Expected 2 user messages, got %d", len(userMsgs))
	}

	assistantMsgs := conv.FilterByRole(RoleAssistant)
	if len(assistantMsgs) != 1 {
		t.Errorf("Expected 1 assistant message, got %d", len(assistantMsgs))
	}
}

func TestConversationClear(t *testing.T) {
	conv := &Conversation{Messages: []Message{}}

	conv.AddMessage(Message{Role: RoleUser, Content: "Test"})
	conv.AddMessage(Message{Role: RoleAssistant, Content: "Response"})

	if len(conv.Messages) != 2 {
		t.Fatalf("Expected 2 messages before clear, got %d", len(conv.Messages))
	}

	conv.Clear()

	if len(conv.Messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(conv.Messages))
	}
}

func TestConversationMessageCount(t *testing.T) {
	conv := &Conversation{Messages: []Message{}}

	if count := conv.MessageCount(); count != 0 {
		t.Errorf("Expected 0, got %d", count)
	}

	conv.AddMessage(Message{Role: RoleUser, Content: "Test"})

	if count := conv.MessageCount(); count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}
}
