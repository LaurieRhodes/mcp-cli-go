package chat

import (
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

func TestNewSession(t *testing.T) {
	session := NewSession("You are a helpful assistant")

	if session.ID == "" {
		t.Error("Expected session ID to be set")
	}

	if session.Conversation == nil {
		t.Fatal("Expected conversation to be initialized")
	}

	if session.Conversation.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("Expected system prompt to be set, got '%s'", session.Conversation.SystemPrompt)
	}

	if session.MessageCount() != 0 {
		t.Errorf("Expected 0 messages, got %d", session.MessageCount())
	}
}

func TestSessionAddMessage(t *testing.T) {
	session := NewSession("")

	msg := models.Message{
		Role:    models.RoleUser,
		Content: "Hello",
	}

	session.AddMessage(msg)

	if session.MessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", session.MessageCount())
	}

	last := session.GetLastMessage()
	if last == nil {
		t.Fatal("Expected last message to not be nil")
	}

	if last.Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", last.Content)
	}
}

func TestSessionClear(t *testing.T) {
	session := NewSession("")

	session.AddMessage(models.Message{Role: models.RoleUser, Content: "Test 1"})
	session.AddMessage(models.Message{Role: models.RoleAssistant, Content: "Test 2"})

	if session.MessageCount() != 2 {
		t.Fatalf("Expected 2 messages, got %d", session.MessageCount())
	}

	session.Clear()

	if session.MessageCount() != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", session.MessageCount())
	}
}

func TestSessionGetMessages(t *testing.T) {
	session := NewSession("")

	session.AddMessage(models.Message{Role: models.RoleUser, Content: "Hello"})
	session.AddMessage(models.Message{Role: models.RoleAssistant, Content: "Hi"})

	messages := session.GetMessages()

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected first message 'Hello', got '%s'", messages[0].Content)
	}

	if messages[1].Content != "Hi" {
		t.Errorf("Expected second message 'Hi', got '%s'", messages[1].Content)
	}
}

func TestSessionGetTotalTokens(t *testing.T) {
	session := NewSession("")

	// Add messages with known content
	session.AddMessage(models.Message{
		Role:    models.RoleUser,
		Content: "1234", // 1 token (approximate)
	})

	tokens := session.GetTotalTokens()

	if tokens == 0 {
		t.Error("Expected non-zero token count")
	}
}
