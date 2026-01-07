package workflow

import (
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// PropertyResolver resolves properties using 3-level inheritance:
// execution (workflow) → step → consensus.execution
type PropertyResolver struct {
	execution *config.ExecutionContext
}

// NewPropertyResolver creates a new property resolver
func NewPropertyResolver(exec *config.ExecutionContext) *PropertyResolver {
	return &PropertyResolver{execution: exec}
}

// ResolveProviders resolves the provider chain for a step
// Returns array of providers to try in order (fallback chain)
func (r *PropertyResolver) ResolveProviders(step *config.StepV2) []config.ProviderFallback {
	// Step level providers (highest priority)
	if len(step.Providers) > 0 {
		return step.Providers
	}

	// Step level single provider
	if step.Provider != "" && step.Model != "" {
		return []config.ProviderFallback{{
			Provider: step.Provider,
			Model:    step.Model,
		}}
	}

	// Execution context providers
	if len(r.execution.Providers) > 0 {
		return r.execution.Providers
	}

	// Execution context single provider
	if r.execution.Provider != "" && r.execution.Model != "" {
		return []config.ProviderFallback{{
			Provider: r.execution.Provider,
			Model:    r.execution.Model,
		}}
	}

	// No providers configured
	return nil
}

// ResolveServers resolves MCP servers for a step
func (r *PropertyResolver) ResolveServers(step *config.StepV2) []string {
	// Step override
	if len(step.Servers) > 0 {
		return step.Servers
	}

	// Execution default
	return r.execution.Servers
}

// ResolveTemperature resolves temperature setting
func (r *PropertyResolver) ResolveTemperature(step *config.StepV2) float64 {
	// Step override
	if step.Temperature != nil {
		return *step.Temperature
	}

	// Execution default
	if r.execution.Temperature != 0 {
		return r.execution.Temperature
	}

	// mcp-cli default
	return 0.7
}

// ResolveMaxTokens resolves max tokens setting
func (r *PropertyResolver) ResolveMaxTokens(step *config.StepV2) int {
	// Step override
	if step.MaxTokens != nil {
		return *step.MaxTokens
	}

	// Execution default
	if r.execution.MaxTokens != 0 {
		return r.execution.MaxTokens
	}

	// mcp-cli default
	return 4096
}

// ResolveTimeout resolves timeout duration
func (r *PropertyResolver) ResolveTimeout(step *config.StepV2) time.Duration {
	// Step override
	if step.Timeout != nil {
		return *step.Timeout
	}

	// Execution default
	if r.execution.Timeout != 0 {
		return r.execution.Timeout
	}

	// mcp-cli default
	return 30 * time.Second
}

// ResolveLogging resolves logging level
func (r *PropertyResolver) ResolveLogging(step *config.StepV2) string {
	// Step override
	if step.Logging != "" {
		return step.Logging
	}

	// Execution default
	if r.execution.Logging != "" {
		return r.execution.Logging
	}

	// mcp-cli default
	return "normal"
}

// ResolveNoColor resolves no color setting
func (r *PropertyResolver) ResolveNoColor(step *config.StepV2) bool {
	// Step override
	if step.NoColor != nil {
		return *step.NoColor
	}

	// Execution default
	return r.execution.NoColor
}

// ResolveConsensusTemperature resolves temperature for consensus execution
// Follows 3-level hierarchy: consensus exec → step → execution
func (r *PropertyResolver) ResolveConsensusTemperature(
	exec *config.ConsensusExec,
	step *config.StepV2,
) float64 {
	// Consensus execution level (highest priority)
	if exec.Temperature != nil {
		return *exec.Temperature
	}

	// Step level
	if step.Temperature != nil {
		return *step.Temperature
	}

	// Execution context level
	if r.execution.Temperature != 0 {
		return r.execution.Temperature
	}

	// Default
	return 0.7
}

// ResolveConsensusMaxTokens resolves max tokens for consensus execution
func (r *PropertyResolver) ResolveConsensusMaxTokens(
	exec *config.ConsensusExec,
	step *config.StepV2,
) int {
	// Consensus execution level
	if exec.MaxTokens != nil {
		return *exec.MaxTokens
	}

	// Step level
	if step.MaxTokens != nil {
		return *step.MaxTokens
	}

	// Execution context level
	if r.execution.MaxTokens != 0 {
		return r.execution.MaxTokens
	}

	// Default
	return 4096
}

// ResolveConsensusTimeout resolves timeout for consensus execution
func (r *PropertyResolver) ResolveConsensusTimeout(
	exec *config.ConsensusExec,
	step *config.StepV2,
) time.Duration {
	// Consensus execution level
	if exec.Timeout != nil {
		return *exec.Timeout
	}

	// Step level
	if step.Timeout != nil {
		return *step.Timeout
	}

	// Execution context level
	if r.execution.Timeout != 0 {
		return r.execution.Timeout
	}

	// Default
	return 30 * time.Second
}
