package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/rag"
)

// executeRagStep executes a RAG retrieval step
func (o *Orchestrator) executeRagStep(ctx context.Context, step *config.StepV2) error {
	ragMode := step.Rag
	if ragMode == nil {
		return fmt.Errorf("rag mode is nil")
	}
	
	o.logger.Info("🔍 Executing RAG step: %s", step.Name)
	
	// Get RAG configuration from already-loaded app config
	if o.appConfig == nil || o.appConfig.RAG == nil {
		return fmt.Errorf("RAG configuration not loaded")
	}
	ragConfig := o.appConfig.RAG
	
	// Ensure we have a server manager
	if o.executor.serverManager == nil {
		return fmt.Errorf("server manager not initialized")
	}
	
	// Initialize RAG service with app config's RAG settings and embedding service
	ragService := rag.NewServiceWithConfig(ragConfig, o.executor.serverManager, o.embeddingService)
	
	// Interpolate query
	query, err := o.interpolator.Interpolate(ragMode.Query)
	if err != nil {
		return fmt.Errorf("failed to interpolate query: %w", err)
	}
	
	o.logger.Debug("RAG query: %s", query)
	
	// Single server search
	serverName := ragMode.Server
	if serverName == "" {
		serverName = ragConfig.DefaultServer
	}
	if serverName == "" {
		return fmt.Errorf("no server specified and no default server in RAG config")
	}
	
	req := rag.SearchRequest{
		Query:       query,
		Server:      serverName,
		Strategies:  ragMode.Strategies,
		TopK:        ragMode.TopK,
		Fusion:      ragMode.Fusion,
		ExpandQuery: ragMode.ExpandQuery,
	}
	
	response, err := ragService.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("RAG search failed: %w", err)
	}
	
	// Format output based on configuration
	var output string
	outputFormat := ragMode.OutputFormat
	if outputFormat == "" {
		outputFormat = "json"
	}
	
	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		output = string(data)
		
	case "compact":
		data, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		output = string(data)
		
	case "text":
		output = formatRagResultsAsText(response)
		
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	
	// Store results
	o.stepResults[step.Name] = output
	o.interpolator.SetStepResult(step.Name, output)
	
	// Also store structured results for easier access
	resultsJSON, _ := json.Marshal(response.Results)
	o.interpolator.Set(fmt.Sprintf("%s.results", step.Name), string(resultsJSON))
	o.interpolator.Set(fmt.Sprintf("%s.total_results", step.Name), fmt.Sprintf("%d", response.TotalResults))
	o.interpolator.Set(fmt.Sprintf("%s.fusion_method", step.Name), response.Fusion)
	
	o.logger.Info("✓ RAG step completed: %d results", response.TotalResults)
	
	return nil
}

// formatRagResultsAsText formats RAG results as human-readable text
func formatRagResultsAsText(response *rag.SearchResponse) string {
	var output string
	
	output += fmt.Sprintf("Query: %s\n", response.Query)
	output += fmt.Sprintf("Results: %d\n\n", response.TotalResults)
	
	for i, result := range response.Results {
		output += fmt.Sprintf("--- Result %d (score: %.4f) ---\n", i+1, result.CombinedScore)
		output += fmt.Sprintf("ID: %s\n", result.ID)
		
		if result.Source != "" {
			output += fmt.Sprintf("Source: %s\n", result.Source)
		}
		
		output += "Text:\n"
		for key, value := range result.Text {
			output += fmt.Sprintf("  %s: %v\n", key, value)
		}
		
		if len(result.ComponentScores) > 0 {
			output += "Component Scores:\n"
			for strategy, score := range result.ComponentScores {
				output += fmt.Sprintf("  %s: %.4f\n", strategy, score)
			}
		}
		
		output += "\n"
	}
	
	return output
}

