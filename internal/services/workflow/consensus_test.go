package workflow

import (
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

func TestCountVotes(t *testing.T) {
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal")
	executor := NewExecutor(workflow, logger)
	ce := NewConsensusExecutor(executor)

	tests := []struct {
		name        string
		results     []*ProviderResult
		requirement string
		wantSuccess bool
		wantResult  string
		wantAgree   float64
	}{
		{
			name: "unanimous success",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
			},
			requirement: "unanimous",
			wantSuccess: true,
			wantResult:  "YES",
			wantAgree:   1.0,
		},
		{
			name: "unanimous failure (split)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "NO"},
			},
			requirement: "unanimous",
			wantSuccess: false,
			wantAgree:   0.5,
		},
		{
			name: "2/3 success",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "NO"},
			},
			requirement: "2/3",
			wantSuccess: true,
			wantResult:  "YES",
			wantAgree:   0.67,
		},
		{
			name: "2/3 failure (not enough)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "NO"},
				{Provider: "gemini", Model: "pro", Output: "MAYBE"},
			},
			requirement: "2/3",
			wantSuccess: false,
			wantAgree:   0.33,
		},
		{
			name: "majority success",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "NO"},
			},
			requirement: "majority",
			wantSuccess: true,
			wantResult:  "YES",
			wantAgree:   0.67,
		},
		{
			name: "majority failure (tie)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "NO"},
			},
			requirement: "majority",
			wantSuccess: false,
			wantAgree:   0.5,
		},
		{
			name: "case insensitive matching",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "yes"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "Yes"},
			},
			requirement: "unanimous",
			wantSuccess: true,
			wantAgree:   1.0,
		},
		{
			name: "whitespace normalization",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "  YES  "},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "YES\n"},
			},
			requirement: "unanimous",
			wantSuccess: true,
			wantAgree:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ce.countVotes(tt.results, tt.requirement)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSuccess, result.Success)
			assert.InDelta(t, tt.wantAgree, result.Agreement, 0.01)
			
			if tt.wantSuccess && tt.wantResult != "" {
				// Check that result matches (case-insensitive, whitespace-normalized)
				assert.Contains(t, toUpperCase(result.Result), toUpperCase(tt.wantResult))
			}
		})
	}
}

func TestCountVotesErrors(t *testing.T) {
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal")
	executor := NewExecutor(workflow, logger)
	ce := NewConsensusExecutor(executor)

	tests := []struct {
		name        string
		results     []*ProviderResult
		requirement string
		wantErr     bool
	}{
		{
			name: "invalid requirement",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
			},
			requirement: "invalid",
			wantErr:     true,
		},
		{
			name: "no successful votes",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Error: assert.AnError},
				{Provider: "openai", Model: "gpt4", Error: assert.AnError},
			},
			requirement: "unanimous",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ce.countVotes(tt.results, tt.requirement)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeOutput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "YES",
			want:  "YES",
		},
		{
			input: "yes",
			want:  "YES",
		},
		{
			input: "  YES  ",
			want:  "YES",
		},
		{
			input: "YES\n",
			want:  "YES",
		},
		{
			input: "  yes  \n",
			want:  "YES",
		},
		{
			input: "Y E S",
			want:  "Y E S",
		},
		{
			input: "Hello World",
			want:  "HELLO WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeOutput(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConsensusConfidence(t *testing.T) {
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal")
	executor := NewExecutor(workflow, logger)
	ce := NewConsensusExecutor(executor)

	tests := []struct {
		name           string
		results        []*ProviderResult
		wantConfidence string
	}{
		{
			name: "high confidence (unanimous)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "YES"},
			},
			wantConfidence: "high",
		},
		{
			name: "good confidence (75%)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "YES"},
				{Provider: "mistral", Model: "large", Output: "NO"},
			},
			wantConfidence: "good",
		},
		{
			name: "medium confidence (67%)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "YES"},
				{Provider: "gemini", Model: "pro", Output: "NO"},
			},
			wantConfidence: "medium",
		},
		{
			name: "low confidence (50%)",
			results: []*ProviderResult{
				{Provider: "anthropic", Model: "claude", Output: "YES"},
				{Provider: "openai", Model: "gpt4", Output: "NO"},
			},
			wantConfidence: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ce.countVotes(tt.results, "majority")
			assert.NoError(t, err)
			assert.Equal(t, tt.wantConfidence, result.Confidence)
		})
	}
}

func TestConsensusExecutorCreation(t *testing.T) {
	workflow := &config.WorkflowV2{
		Name:    "test",
		Version: "1.0.0",
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal")
	executor := NewExecutor(workflow, logger)
	ce := NewConsensusExecutor(executor)

	assert.NotNil(t, ce)
	assert.NotNil(t, ce.executor)
	assert.NotNil(t, ce.logger)
}

func TestProviderResultWithError(t *testing.T) {
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal")
	executor := NewExecutor(workflow, logger)
	ce := NewConsensusExecutor(executor)

	// Test that results with errors are excluded from voting
	results := []*ProviderResult{
		{Provider: "anthropic", Model: "claude", Output: "YES"},
		{Provider: "openai", Model: "gpt4", Error: assert.AnError}, // Error - should be excluded
		{Provider: "gemini", Model: "pro", Output: "YES"},
	}

	result, err := ce.countVotes(results, "unanimous")
	assert.NoError(t, err)
	assert.True(t, result.Success) // 2/2 successful votes are unanimous
	assert.Equal(t, 1.0, result.Agreement)
}
