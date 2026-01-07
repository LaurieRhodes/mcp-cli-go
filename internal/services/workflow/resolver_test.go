package workflow

import (
	"testing"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

func TestResolveProviders(t *testing.T) {
	tests := []struct {
		name      string
		execution *config.ExecutionContext
		step      *config.StepV2
		want      []config.ProviderFallback
	}{
		{
			name: "step providers override execution",
			execution: &config.ExecutionContext{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			step: &config.StepV2{
				Providers: []config.ProviderFallback{
					{Provider: "openai", Model: "gpt-4o"},
				},
			},
			want: []config.ProviderFallback{
				{Provider: "openai", Model: "gpt-4o"},
			},
		},
		{
			name: "step single provider",
			execution: &config.ExecutionContext{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			step: &config.StepV2{
				Provider: "openai",
				Model:    "gpt-4o",
			},
			want: []config.ProviderFallback{
				{Provider: "openai", Model: "gpt-4o"},
			},
		},
		{
			name: "inherit from execution providers array",
			execution: &config.ExecutionContext{
				Providers: []config.ProviderFallback{
					{Provider: "anthropic", Model: "claude-sonnet-4"},
					{Provider: "openai", Model: "gpt-4o"},
				},
			},
			step: &config.StepV2{},
			want: []config.ProviderFallback{
				{Provider: "anthropic", Model: "claude-sonnet-4"},
				{Provider: "openai", Model: "gpt-4o"},
			},
		},
		{
			name: "inherit from execution single provider",
			execution: &config.ExecutionContext{
				Provider: "anthropic",
				Model:    "claude-sonnet-4",
			},
			step: &config.StepV2{},
			want: []config.ProviderFallback{
				{Provider: "anthropic", Model: "claude-sonnet-4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPropertyResolver(tt.execution)
			got := resolver.ResolveProviders(tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveTemperature(t *testing.T) {
	tests := []struct {
		name      string
		execution *config.ExecutionContext
		step      *config.StepV2
		want      float64
	}{
		{
			name: "step override",
			execution: &config.ExecutionContext{
				Temperature: 0.9,
			},
			step: &config.StepV2{
				Temperature: ptrFloat64(0.3),
			},
			want: 0.3,
		},
		{
			name: "inherit from execution",
			execution: &config.ExecutionContext{
				Temperature: 0.9,
			},
			step: &config.StepV2{},
			want: 0.9,
		},
		{
			name: "use default",
			execution: &config.ExecutionContext{},
			step:      &config.StepV2{},
			want:      0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPropertyResolver(tt.execution)
			got := resolver.ResolveTemperature(tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveServers(t *testing.T) {
	tests := []struct {
		name      string
		execution *config.ExecutionContext
		step      *config.StepV2
		want      []string
	}{
		{
			name: "step override",
			execution: &config.ExecutionContext{
				Servers: []string{"filesystem", "brave-search"},
			},
			step: &config.StepV2{
				Servers: []string{"brave-search"},
			},
			want: []string{"brave-search"},
		},
		{
			name: "inherit from execution",
			execution: &config.ExecutionContext{
				Servers: []string{"filesystem", "brave-search"},
			},
			step: &config.StepV2{},
			want: []string{"filesystem", "brave-search"},
		},
		{
			name:      "no servers",
			execution: &config.ExecutionContext{},
			step:      &config.StepV2{},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPropertyResolver(tt.execution)
			got := resolver.ResolveServers(tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveConsensusTemperature(t *testing.T) {
	tests := []struct {
		name      string
		execution *config.ExecutionContext
		step      *config.StepV2
		exec      *config.ConsensusExec
		want      float64
	}{
		{
			name: "consensus exec override",
			execution: &config.ExecutionContext{
				Temperature: 0.9,
			},
			step: &config.StepV2{
				Temperature: ptrFloat64(0.5),
			},
			exec: &config.ConsensusExec{
				Temperature: ptrFloat64(0.0),
			},
			want: 0.0,
		},
		{
			name: "inherit from step",
			execution: &config.ExecutionContext{
				Temperature: 0.9,
			},
			step: &config.StepV2{
				Temperature: ptrFloat64(0.5),
			},
			exec: &config.ConsensusExec{},
			want: 0.5,
		},
		{
			name: "inherit from execution",
			execution: &config.ExecutionContext{
				Temperature: 0.9,
			},
			step: &config.StepV2{},
			exec: &config.ConsensusExec{},
			want: 0.9,
		},
		{
			name:      "use default",
			execution: &config.ExecutionContext{},
			step:      &config.StepV2{},
			exec:      &config.ConsensusExec{},
			want:      0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPropertyResolver(tt.execution)
			got := resolver.ResolveConsensusTemperature(tt.exec, tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveTimeout(t *testing.T) {
	tests := []struct {
		name      string
		execution *config.ExecutionContext
		step      *config.StepV2
		want      time.Duration
	}{
		{
			name: "step override",
			execution: &config.ExecutionContext{
				Timeout: 60 * time.Second,
			},
			step: &config.StepV2{
				Timeout: ptrDuration(10 * time.Second),
			},
			want: 10 * time.Second,
		},
		{
			name: "inherit from execution",
			execution: &config.ExecutionContext{
				Timeout: 60 * time.Second,
			},
			step: &config.StepV2{},
			want: 60 * time.Second,
		},
		{
			name:      "use default",
			execution: &config.ExecutionContext{},
			step:      &config.StepV2{},
			want:      30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewPropertyResolver(tt.execution)
			got := resolver.ResolveTimeout(tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt(i int) *int {
	return &i
}

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}

func ptrBool(b bool) *bool {
	return &b
}
