package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// ConsensusExecutor handles multi-provider consensus execution
type ConsensusExecutor struct {
	executor *Executor
	logger   *Logger
}

// NewConsensusExecutor creates a new consensus executor
func NewConsensusExecutor(executor *Executor) *ConsensusExecutor {
	return &ConsensusExecutor{
		executor: executor,
		logger:   executor.logger,
	}
}

// ProviderResult represents a single provider's response in consensus
type ProviderResult struct {
	Provider string
	Model    string
	Output   string
	Error    error
	Duration time.Duration
}

// ExecuteConsensus executes a consensus step with multiple providers
func (ce *ConsensusExecutor) ExecuteConsensus(
	ctx context.Context,
	step *config.StepV2,
) (*config.ConsensusResult, error) {
	consensus := step.Consensus
	if consensus == nil {
		return nil, fmt.Errorf("no consensus configuration")
	}

	if len(consensus.Executions) < 2 {
		return nil, fmt.Errorf("consensus requires at least 2 providers, got %d", len(consensus.Executions))
	}

	ce.logger.Info("Starting consensus with %d providers", len(consensus.Executions))

	// Execute all providers in parallel
	results := ce.executeParallel(ctx, step, consensus)

	// Count successful responses
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		} else {
			failCount++
			// Log API failures separately (not vote failures)
			ce.logger.Warn("Consensus: %s/%s failed - %v", r.Provider, r.Model, r.Error)
		}
	}

	ce.logger.Debug("Consensus results: %d successful, %d failed (API errors)",
		successCount, failCount)

	// Check if we have any successful responses
	if successCount == 0 {
		return nil, fmt.Errorf("all %d consensus providers failed (API errors, not votes)",
			len(consensus.Executions))
	}

	// Check if we have enough successful providers to meet requirement
	// For any requirement, we need at least 2 successful providers
	if successCount < 2 {
		return nil, fmt.Errorf("insufficient successful providers for consensus: only %d/%d succeeded (need at least 2)",
			successCount, len(consensus.Executions))
	}

	ce.logger.Info("Consensus voting with %d providers (ignoring %d API failures)",
		successCount, failCount)

	// Count votes from successful results only
	return ce.countVotes(results, consensus.Require)
}

// executeParallel executes all consensus providers in parallel
func (ce *ConsensusExecutor) executeParallel(
	ctx context.Context,
	step *config.StepV2,
	consensus *config.ConsensusMode,
) []*ProviderResult {
	// Channel for results
	resultsChan := make(chan *ProviderResult, len(consensus.Executions))

	// WaitGroup for goroutines
	var wg sync.WaitGroup

	// Resolve overall timeout
	timeout := consensus.Timeout
	if timeout == 0 {
		timeout = ce.executor.resolver.ResolveTimeout(step)
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Launch goroutine for each provider
	for _, exec := range consensus.Executions {
		wg.Add(1)
		go func(e config.ConsensusExec) {
			defer wg.Done()
			result := ce.executeConsensusProvider(execCtx, step, e, consensus.Prompt)
			resultsChan <- result
		}(exec)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []*ProviderResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// executeConsensusProvider executes a single provider in consensus
func (ce *ConsensusExecutor) executeConsensusProvider(
	ctx context.Context,
	step *config.StepV2,
	exec config.ConsensusExec,
	prompt string,
) *ProviderResult {
	ce.logger.Debug("Consensus: executing %s/%s", exec.Provider, exec.Model)

	startTime := time.Now()

	// Create a temporary step for this consensus execution
	tempStep := &config.StepV2{
		Name:     step.Name + "_consensus",
		Run:      prompt,
		Provider: exec.Provider,
		Model:    exec.Model,
	}

	// Apply consensus-level overrides
	tempStep.Temperature = exec.Temperature
	tempStep.MaxTokens = exec.MaxTokens
	tempStep.Timeout = exec.Timeout

	// Inherit other properties from original step
	tempStep.Servers = step.Servers
	tempStep.Logging = step.Logging
	tempStep.NoColor = step.NoColor

	// Execute with single provider (no fallback for consensus)
	providerConfig := config.ProviderFallback{
		Provider: exec.Provider,
		Model:    exec.Model,
	}

	result, err := ce.executor.executeWithProvider(ctx, tempStep, providerConfig)

	duration := time.Since(startTime)

	if err != nil {
		ce.logger.Warn("Consensus: %s/%s failed - %v", exec.Provider, exec.Model, err)
		return &ProviderResult{
			Provider: exec.Provider,
			Model:    exec.Model,
			Error:    err,
			Duration: duration,
		}
	}

	ce.logger.Info("Consensus: %s/%s succeeded (%.2fs)", exec.Provider, exec.Model, duration.Seconds())

	return &ProviderResult{
		Provider: exec.Provider,
		Model:    exec.Model,
		Output:   result.Output,
		Duration: duration,
	}
}

// countVotes counts votes and determines consensus
func (ce *ConsensusExecutor) countVotes(
	results []*ProviderResult,
	requirement string,
) (*config.ConsensusResult, error) {
	// Extract successful responses
	votes := make(map[string]string)
	counts := make(map[string]int)

	for _, r := range results {
		if r.Error == nil {
			// Normalize output (trim whitespace, lowercase for comparison)
			normalized := normalizeOutput(r.Output)
			votes[r.Provider+"/"+r.Model] = r.Output // Store original
			counts[normalized]++

			// Log what each provider voted (for debugging)
			ce.logger.Info("Provider %s/%s normalized vote: %s", r.Provider, r.Model, normalized)
		}
	}

	if len(votes) == 0 {
		return nil, fmt.Errorf("no successful votes to count")
	}

	// Find winner (most votes)
	var winner string
	var winnerOriginal string
	var maxCount int

	for normalized, count := range counts {
		if count > maxCount {
			maxCount = count
			winner = normalized
			// Find original output for this normalized version
			for _, output := range votes {
				if normalizeOutput(output) == normalized {
					winnerOriginal = output
					break
				}
			}
		}
	}

	// Calculate agreement
	totalVotes := len(votes)
	agreement := float64(maxCount) / float64(totalVotes)

	ce.logger.Debug("Vote counts: winner=%s with %d/%d votes (%.1f%%)",
		winner, maxCount, totalVotes, agreement*100)

	// Check requirement
	success := false
	switch requirement {
	case "unanimous":
		success = agreement == 1.0
	case "2/3":
		success = agreement >= 2.0/3.0
	case "majority":
		success = agreement > 0.5
	default:
		return nil, fmt.Errorf("invalid requirement: %s (must be unanimous, 2/3, or majority)", requirement)
	}

	// Determine confidence level
	var confidence string
	switch {
	case agreement == 1.0:
		confidence = "high"
	case agreement >= 0.75:
		confidence = "good"
	case agreement >= 0.6:
		confidence = "medium"
	default:
		confidence = "low"
	}

	ce.logger.Info("Consensus: %s (%.0f%% agreement, confidence: %s)",
		map[bool]string{true: "SUCCESS", false: "FAILED"}[success],
		agreement*100, confidence)

	return &config.ConsensusResult{
		Success:    success,
		Result:     winnerOriginal,
		Agreement:  agreement,
		Votes:      votes,
		Confidence: confidence,
	}, nil
}

// normalizeOutput normalizes output for comparison
// For validation steps, extracts SUCCESS or FAIL keywords
func normalizeOutput(output string) string {
	// Trim whitespace
	output = strings.TrimSpace(output)

	// Convert to uppercase for case-insensitive comparison
	outputUpper := strings.ToUpper(output)

	// For validation outputs, extract SUCCESS or FAIL
	// Check for SUCCESS (but not if FAIL is also present, which would indicate FAIL)
	if strings.Contains(outputUpper, "SUCCESS") && !strings.Contains(outputUpper, "FAIL") {
		return "SUCCESS"
	}

	// Check for FAIL
	if strings.Contains(outputUpper, "FAIL") {
		return "FAIL"
	}

	// For other consensus outputs, normalize the entire string
	normalized := removeExtraWhitespace(output)
	normalized = toUpperCase(normalized)

	return normalized
}

// Helper functions
func removeExtraWhitespace(s string) string {
	// Simple whitespace removal
	result := ""
	prevSpace := false

	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				result += " "
				prevSpace = true
			}
		} else {
			result += string(r)
			prevSpace = false
		}
	}

	// Trim
	if len(result) > 0 && result[0] == ' ' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}

	return result
}

func toUpperCase(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			result += string(r - 32)
		} else {
			result += string(r)
		}
	}
	return result
}
