package chat

import (
	"context"
	"strings"
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	response *ports.CompletionResponse
	err      error
}

func (m *MockProvider) CreateCompletion(ctx context.Context, req *ports.CompletionRequest) (*ports.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockProvider) StreamCompletion(ctx context.Context, req *ports.CompletionRequest, writer interface{}) (*ports.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockProvider) CreateEmbeddings(ctx context.Context, req *models.EmbeddingRequest) (*models.EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockProvider) GetProviderType() ports.ProviderType {
	return ports.ProviderOpenAI
}

func (m *MockProvider) GetSupportedModels() []string {
	return []string{"test-model"}
}

func (m *MockProvider) GetMaxTokens(model string) int {
	return 4096
}

func (m *MockProvider) ValidateConfig() error {
	return nil
}

func (m *MockProvider) Close() error {
	return nil
}

func TestServiceSendMessage(t *testing.T) {
	mockProvider := &MockProvider{
		response: &ports.CompletionResponse{
			Content: "Hello, how can I help?",
			Model:   "test-model",
		},
	}

	service := NewService(mockProvider, ServiceConfig{})

	req := &MessageRequest{
		Messages: []models.Message{
			{Role: models.RoleUser, Content: "Hello"},
		},
	}

	resp, err := service.SendMessage(context.Background(), req)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if resp.Content != "Hello, how can I help?" {
		t.Errorf("Expected 'Hello, how can I help?', got '%s'", resp.Content)
	}

	if resp.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", resp.Model)
	}
}

func TestServiceStreamMessage(t *testing.T) {
	mockProvider := &MockProvider{
		response: &ports.CompletionResponse{
			Content: "Streamed response",
			Model:   "test-model",
		},
	}

	service := NewService(mockProvider, ServiceConfig{})

	req := &MessageRequest{
		Messages: []models.Message{
			{Role: models.RoleUser, Content: "Test"},
		},
	}

	var buf strings.Builder
	resp, err := service.StreamMessage(context.Background(), req, &buf)
	if err != nil {
		t.Fatalf("StreamMessage failed: %v", err)
	}

	if resp.Content != "Streamed response" {
		t.Errorf("Expected 'Streamed response', got '%s'", resp.Content)
	}
}

func TestServiceConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         ServiceConfig
		expectedMax    int
		expectedTokens int
	}{
		{
			name:           "default config",
			config:         ServiceConfig{},
			expectedMax:    50,
			expectedTokens: 4096,
		},
		{
			name: "custom config",
			config: ServiceConfig{
				MaxHistory: 100,
				MaxTokens:  8192,
			},
			expectedMax:    100,
			expectedTokens: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(&MockProvider{}, tt.config)

			if service.config.MaxHistory != tt.expectedMax {
				t.Errorf("Expected MaxHistory %d, got %d", tt.expectedMax, service.config.MaxHistory)
			}

			if service.config.MaxTokens != tt.expectedTokens {
				t.Errorf("Expected MaxTokens %d, got %d", tt.expectedTokens, service.config.MaxTokens)
			}
		})
	}
}
