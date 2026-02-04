package workflow

import (
	"testing"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

// TestBuildMCPCliArgs is disabled - buildMCPCliArgs method doesn't exist in current implementation
// TODO: Re-enable if this functionality is added back
/*
func TestBuildMCPCliArgs(t *testing.T) {
	tests := []struct {
		name      string
		workflow  *config.WorkflowV2
		step      *config.StepV2
		provider  config.ProviderFallback
		wantArgs  []string
	}{
		{
			name: "basic args",
			workflow: &config.WorkflowV2{
				Execution: config.ExecutionContext{},
			},
			step: &config.StepV2{
				Name: "test",
				Run:  "test prompt",
			},
			provider: config.ProviderFallback{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			wantArgs: []string{
				"query",
				"--provider", "anthropic",
				"--model", "claude-sonnet-4",
				"test prompt",
			},
		},
		{
			name: "with servers",
			workflow: &config.WorkflowV2{
				Execution: config.ExecutionContext{
					Servers: []string{"filesystem", "brave-search"},
				},
			},
			step: &config.StepV2{
				Name: "test",
				Run:  "test prompt",
			},
			provider: config.ProviderFallback{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			wantArgs: []string{
				"query",
				"--provider", "anthropic",
				"--model", "claude-sonnet-4",
				"--server", "filesystem,brave-search",
				"test prompt",
			},
		},
		{
			name: "with verbose logging",
			workflow: &config.WorkflowV2{
				Execution: config.ExecutionContext{
					Logging: "verbose",
				},
			},
			step: &config.StepV2{
				Name: "test",
				Run:  "test prompt",
			},
			provider: config.ProviderFallback{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			wantArgs: []string{
				"query",
				"--provider", "anthropic",
				"--model", "claude-sonnet-4",
				"--verbose",
				"test prompt",
			},
		},
		{
			name: "with no color",
			workflow: &config.WorkflowV2{
				Execution: config.ExecutionContext{
					NoColor: true,
				},
			},
			step: &config.StepV2{
				Name: "test",
				Run:  "test prompt",
			},
			provider: config.ProviderFallback{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			wantArgs: []string{
				"query",
				"--provider", "anthropic",
				"--model", "claude-sonnet-4",
				"--no-color",
				"test prompt",
			},
		},
		{
			name: "step overrides servers",
			workflow: &config.WorkflowV2{
				Execution: config.ExecutionContext{
					Servers: []string{"filesystem", "brave-search"},
				},
			},
			step: &config.StepV2{
				Name:    "test",
				Run:     "test prompt",
				Servers: []string{"brave-search"},
			},
			provider: config.ProviderFallback{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			wantArgs: []string{
				"query",
				"--provider", "anthropic",
				"--model", "claude-sonnet-4",
				"--server", "brave-search",
				"test prompt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger("normal", false)
			executor := NewExecutor(tt.workflow, logger)

			got := executor.buildMCPCliArgs(tt.step, tt.provider)

			assert.Equal(t, tt.wantArgs, got)
		})
	}
}
*/

// TestIsRetriableError is disabled - isRetriableError function and ExitCode field don't exist in current implementation
// TODO: Re-enable if error retry logic is added back
/*
func TestIsRetriableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "timeout error",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 124,
				Err:      fmt.Errorf("timeout"),
			},
			want: true,
		},
		{
			name: "network error",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 2,
				Err:      fmt.Errorf("network error"),
			},
			want: true,
		},
		{
			name: "rate limit",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 429,
				Err:      fmt.Errorf("rate limit exceeded"),
			},
			want: true,
		},
		{
			name: "service unavailable",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 503,
				Err:      fmt.Errorf("service unavailable"),
			},
			want: true,
		},
		{
			name: "invalid API key (non-retriable)",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 401,
				Err:      fmt.Errorf("invalid API key"),
			},
			want: false,
		},
		{
			name: "bad request (non-retriable)",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 400,
				Err:      fmt.Errorf("bad request"),
			},
			want: false,
		},
		{
			name: "timeout in error message",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 1,
				Err:      fmt.Errorf("request timeout"),
			},
			want: true,
		},
		{
			name: "connection refused",
			err: &ProviderError{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
				ExitCode: 1,
				Err:      fmt.Errorf("connection refused"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
*/

func TestExecutorCreation(t *testing.T) {
	workflow := &config.WorkflowV2{
		Name:    "test",
		Version: "1.0.0",
		Execution: config.ExecutionContext{
			Provider: "anthropic",
			Model:    "claude-sonnet-4",
		},
	}

	logger := NewLogger("normal", false)
	executor := NewExecutor(workflow, logger)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.resolver)
	assert.NotNil(t, executor.logger)
	assert.Equal(t, workflow, executor.workflow)
}

func TestProviderFallbackOrder(t *testing.T) {
	// Test that providers are tried in correct order
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Providers: []config.ProviderFallback{
				{Provider: "provider1", Model: "model1"},
				{Provider: "provider2", Model: "model2"},
				{Provider: "provider3", Model: "model3"},
			},
		},
	}

	step := &config.StepV2{
		Name: "test",
		Run:  "test",
	}

	logger := NewLogger("normal", false)
	executor := NewExecutor(workflow, logger)

	// Get resolved providers
	providers := executor.resolver.ResolveProviders(step)

	assert.Equal(t, 3, len(providers))
	assert.Equal(t, "provider1", providers[0].Provider)
	assert.Equal(t, "provider2", providers[1].Provider)
	assert.Equal(t, "provider3", providers[2].Provider)
}

func TestExecutorResolveTimeout(t *testing.T) {
	workflow := &config.WorkflowV2{
		Execution: config.ExecutionContext{
			Timeout: 45 * time.Second,
		},
	}

	step := &config.StepV2{
		Name: "test",
		Run:  "test",
	}

	logger := NewLogger("normal", false)
	executor := NewExecutor(workflow, logger)

	timeout := executor.resolver.ResolveTimeout(step)
	assert.Equal(t, 45*time.Second, timeout)
}
