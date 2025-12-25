package workflow

import (
	"context"
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// Service handles workflow execution
type Service struct {
	providerFactory ports.ProviderFactory
	mcpManager      ports.MCPManager
}

// NewService creates a new workflow service
func NewService(factory ports.ProviderFactory, mcpManager ports.MCPManager) *Service {
	return &Service{
		providerFactory: factory,
		mcpManager:      mcpManager,
	}
}

// Execute executes a workflow
func (s *Service) Execute(ctx context.Context, workflow *Workflow, input string) (*WorkflowResult, error) {
	result := &WorkflowResult{
		WorkflowName: workflow.Name,
		Steps:        make([]StepResult, 0, len(workflow.Steps)),
	}

	// Process variables
	variables := make(map[string]string)
	variables["input"] = input

	// Execute each step
	for i, step := range workflow.Steps {
		stepResult, err := s.executeStep(ctx, &step, variables)
		if err != nil {
			return nil, fmt.Errorf("step %d (%s) failed: %w", i+1, step.Name, err)
		}

		result.Steps = append(result.Steps, *stepResult)

		// Store step output in variables
		variables[fmt.Sprintf("step%d_output", i+1)] = stepResult.Output
		variables[step.Name] = stepResult.Output
	}

	// Final output is last step's output
	if len(result.Steps) > 0 {
		result.FinalOutput = result.Steps[len(result.Steps)-1].Output
	}

	return result, nil
}

// executeStep executes a single workflow step
func (s *Service) executeStep(ctx context.Context, step *WorkflowStep, variables map[string]string) (*StepResult, error) {
	// Create provider for this step
	provider, err := s.providerFactory.Create(
		ports.ProviderType(step.Provider),
		ports.ProviderConfig{
			APIKey:       step.APIKey,
			DefaultModel: step.Model,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	defer provider.Close()

	// Process prompt with variables
	prompt := replaceVariables(step.Prompt, variables)

	// Get tools if specified
	var tools []models.Tool
	if len(step.Servers) > 0 {
		tools, err = s.mcpManager.GetAllTools()
		if err != nil {
			return nil, fmt.Errorf("failed to get tools: %w", err)
		}
	}

	// Execute completion
	resp, err := provider.CreateCompletion(ctx, &ports.CompletionRequest{
		Messages: []models.Message{{
			Role:    models.RoleUser,
			Content: prompt,
		}},
		Tools:        tools,
		SystemPrompt: step.SystemPrompt,
		Temperature:  step.Temperature,
		MaxTokens:    step.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	return &StepResult{
		StepName: step.Name,
		Output:   resp.Content,
		Usage:    resp.Usage,
		Model:    resp.Model,
	}, nil
}

// Workflow represents a multi-step workflow
type Workflow struct {
	Name        string
	Description string
	Steps       []WorkflowStep
	Variables   map[string]string
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name         string
	Prompt       string
	SystemPrompt string
	Provider     string
	Model        string
	APIKey       string
	Temperature  float64
	MaxTokens    int
	Servers      []string
}

// WorkflowResult contains the result of workflow execution
type WorkflowResult struct {
	WorkflowName string
	Steps        []StepResult
	FinalOutput  string
}

// StepResult contains the result of a single step
type StepResult struct {
	StepName string
	Output   string
	Usage    models.Usage
	Model    string
}

// Helper function to replace variables in text
func replaceVariables(text string, variables map[string]string) string {
	result := text
	for key, value := range variables {
		placeholder := "{" + key + "}"
		result = replaceAll(result, placeholder, value)
	}
	return result
}

func replaceAll(s, old, new string) string {
	// Simple string replacement
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
