package query

import (
	"context"
	"io"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// Service handles one-shot query execution
type Service struct {
	provider ports.LLMProvider
	config   ServiceConfig
}

// ServiceConfig contains query service configuration
type ServiceConfig struct {
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
}

// NewService creates a new query service
func NewService(provider ports.LLMProvider, config ServiceConfig) *Service {
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	return &Service{
		provider: provider,
		config:   config,
	}
}

// Execute executes a one-shot query
func (s *Service) Execute(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// Build messages
	messages := []models.Message{{
		Role:    models.RoleUser,
		Content: req.Query,
	}}

	// Build completion request
	completionReq := &ports.CompletionRequest{
		Messages:     messages,
		Tools:        req.Tools,
		SystemPrompt: req.SystemPrompt,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
	}

	// Use service defaults if not provided
	if completionReq.SystemPrompt == "" {
		completionReq.SystemPrompt = s.config.SystemPrompt
	}
	if completionReq.Temperature == 0 {
		completionReq.Temperature = s.config.Temperature
	}
	if completionReq.MaxTokens == 0 {
		completionReq.MaxTokens = s.config.MaxTokens
	}

	// Execute query
	resp, err := s.provider.CreateCompletion(ctx, completionReq)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{
		Answer:    resp.Content,
		ToolCalls: resp.ToolCalls,
		Usage:     resp.Usage,
		Model:     resp.Model,
	}, nil
}

// ExecuteStream executes a query with streaming
func (s *Service) ExecuteStream(ctx context.Context, req *QueryRequest, writer io.Writer) (*QueryResponse, error) {
	// Build messages
	messages := []models.Message{{
		Role:    models.RoleUser,
		Content: req.Query,
	}}

	// Build completion request
	completionReq := &ports.CompletionRequest{
		Messages:     messages,
		Tools:        req.Tools,
		SystemPrompt: req.SystemPrompt,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Stream:       true,
	}

	// Use service defaults if not provided
	if completionReq.SystemPrompt == "" {
		completionReq.SystemPrompt = s.config.SystemPrompt
	}
	if completionReq.Temperature == 0 {
		completionReq.Temperature = s.config.Temperature
	}
	if completionReq.MaxTokens == 0 {
		completionReq.MaxTokens = s.config.MaxTokens
	}

	// Execute streaming query
	resp, err := s.provider.StreamCompletion(ctx, completionReq, writer)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{
		Answer:    resp.Content,
		ToolCalls: resp.ToolCalls,
		Usage:     resp.Usage,
		Model:     resp.Model,
	}, nil
}

// QueryRequest contains a query request
type QueryRequest struct {
	Query        string
	Tools        []models.Tool
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
}

// QueryResponse contains a query response
type QueryResponse struct {
	Answer    string
	ToolCalls []models.ToolCall
	Usage     models.Usage
	Model     string
}
