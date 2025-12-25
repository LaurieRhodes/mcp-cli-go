package chat

import (
	"context"
	"io"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// Service handles chat functionality
type Service struct {
	provider ports.LLMProvider
	config   ServiceConfig
}

// ServiceConfig contains chat service configuration
type ServiceConfig struct {
	MaxHistory      int
	SystemPrompt    string
	Temperature     float64
	MaxTokens       int
	EnableStreaming bool
}

// NewService creates a new chat service
func NewService(provider ports.LLMProvider, config ServiceConfig) *Service {
	// Set defaults
	if config.MaxHistory == 0 {
		config.MaxHistory = 50
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	return &Service{
		provider: provider,
		config:   config,
	}
}

// SendMessage sends a message and returns the response
func (s *Service) SendMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	// Build completion request
	completionReq := &ports.CompletionRequest{
		Messages:     req.Messages,
		Tools:        req.Tools,
		SystemPrompt: s.config.SystemPrompt,
		Temperature:  s.config.Temperature,
		MaxTokens:    s.config.MaxTokens,
	}

	// Override with request-specific settings
	if req.SystemPrompt != "" {
		completionReq.SystemPrompt = req.SystemPrompt
	}
	if req.Temperature > 0 {
		completionReq.Temperature = req.Temperature
	}

	// Send request
	resp, err := s.provider.CreateCompletion(ctx, completionReq)
	if err != nil {
		return nil, err
	}

	return &MessageResponse{
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
		Usage:     resp.Usage,
		Model:     resp.Model,
	}, nil
}

// StreamMessage sends a message and streams the response
func (s *Service) StreamMessage(ctx context.Context, req *MessageRequest, writer io.Writer) (*MessageResponse, error) {
	// Build completion request
	completionReq := &ports.CompletionRequest{
		Messages:     req.Messages,
		Tools:        req.Tools,
		SystemPrompt: s.config.SystemPrompt,
		Temperature:  s.config.Temperature,
		MaxTokens:    s.config.MaxTokens,
		Stream:       true,
	}

	// Override with request-specific settings
	if req.SystemPrompt != "" {
		completionReq.SystemPrompt = req.SystemPrompt
	}
	if req.Temperature > 0 {
		completionReq.Temperature = req.Temperature
	}

	// Send streaming request
	resp, err := s.provider.StreamCompletion(ctx, completionReq, writer)
	if err != nil {
		return nil, err
	}

	return &MessageResponse{
		Content:   resp.Content,
		ToolCalls: resp.ToolCalls,
		Usage:     resp.Usage,
		Model:     resp.Model,
	}, nil
}

// MessageRequest contains a chat message request
type MessageRequest struct {
	Messages     []models.Message
	Tools        []models.Tool
	SystemPrompt string
	Temperature  float64
}

// MessageResponse contains a chat message response
type MessageResponse struct {
	Content   string
	ToolCalls []models.ToolCall
	Usage     models.Usage
	Model     string
}
