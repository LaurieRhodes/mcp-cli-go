package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// ToolDiscoveryConfig provides configuration for tool discovery and fallback strategies
type ToolDiscoveryConfig struct {
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	InitialWait time.Duration `json:"initial_wait"`
	ToolTimeout time.Duration `json:"tool_timeout"`
}

// DefaultToolDiscoveryConfig returns sensible defaults for tool discovery
func DefaultToolDiscoveryConfig() *ToolDiscoveryConfig {
	return &ToolDiscoveryConfig{
		MaxRetries:  5,
		RetryDelay:  500 * time.Millisecond,
		InitialWait: 1 * time.Second,
		ToolTimeout: 30 * time.Second,
	}
}

// ToolStrategy represents different approaches to execute vector search
type ToolStrategy string

const (
	StrategyNativeVectorSearch ToolStrategy = "native_vector_search"
	StrategyNativeSimilarity   ToolStrategy = "native_similarity"
	StrategyGenericSQL         ToolStrategy = "generic_sql"
	StrategyTextBasedSearch    ToolStrategy = "text_based_search"
)

// ToolMatcher provides intelligent tool matching for different server types
type ToolMatcher struct {
	serverName string
}

// NewToolMatcher creates a new tool matcher for the specified server
func NewToolMatcher(serverName string) *ToolMatcher {
	return &ToolMatcher{serverName: serverName}
}

// ProcessStep overrides the base ProcessStep to handle embeddings with robust tool discovery
func (es *EnhancedService) ProcessStep(ctx context.Context, step *domain.WorkflowStep, variables map[string]interface{}) (*domain.WorkflowStepResult, error) {
	result := &domain.WorkflowStepResult{
		Step:      step.Step,
		Name:      step.Name,
		Status:    domain.StepStatusRunning,
		Variables: make(map[string]interface{}),
	}
	
	logging.Info("🔥 ENHANCED ProcessStep called for step %d: %s", step.Step, step.Name)
	
	// Handle embedding generation if configured
	if step.Embedding != nil && step.Embedding.Generate {
		logging.Info("🚀 Step %d: Starting embedding generation", step.Step)
		
		embedding, err := es.generateStepEmbedding(ctx, step, variables)
		if err != nil {
			logging.Error("❌ Step %d: Embedding generation failed: %v", step.Step, err)
			result.Status = domain.StepStatusFailed
			result.Error = &domain.WorkflowError{
				Code:    "EMBEDDING_GENERATION_ERROR",
				Message: err.Error(),
				Step:    step.Step,
			}
			return result, fmt.Errorf("embedding generation failed: %w", err)
		}
		
		logging.Info("✅ Step %d: Embedding generation completed successfully", step.Step)
		
		// Store embedding metadata (NOT raw vectors) in variables
		if step.Embedding.StoreAs != "" {
			variables[step.Embedding.StoreAs] = embedding.GetMetadata()
			logging.Info("💾 Step %d: Stored embedding metadata in variable: %s", step.Step, step.Embedding.StoreAs)
		}
		
		// Execute embedding usage with robust tool discovery
		logging.Info("🔍 Step %d: Starting embedding usage execution", step.Step)
		usageResult, err := es.executeEmbeddingUsageWithDiscovery(ctx, embedding, step.Embedding.Usage, variables)
		if err != nil {
			logging.Error("❌ Step %d: Embedding usage failed: %v", step.Step, err)
			result.Status = domain.StepStatusFailed
			result.Error = &domain.WorkflowError{
				Code:    "EMBEDDING_USAGE_ERROR",
				Message: err.Error(),
				Step:    step.Step,
			}
			return result, fmt.Errorf("embedding usage failed: %w", err)
		}
		
		logging.Info("✅ Step %d: Embedding usage completed successfully", step.Step)
		
		// Store usage results in variables for LLM consumption
		es.storeEmbeddingUsageResults(variables, step.Embedding.Usage.Type, usageResult, step.Step)
		
		logging.Info("🎉 Step %d: Embedding processing completed", step.Step)
	} else {
		logging.Info("⏭️ Step %d: Skipping embedding processing (not configured or disabled)", step.Step)
	}
	
	// Continue with normal step processing (LLM call)
	logging.Info("🤖 Step %d: Proceeding to LLM processing", step.Step)
	return es.Service.ProcessStep(ctx, step, variables)
}

// executeEmbeddingUsageWithDiscovery executes embedding usage with robust tool discovery
func (es *EnhancedService) executeEmbeddingUsageWithDiscovery(ctx context.Context, embedding *domain.EmbeddingResult, usage domain.EmbeddingUsage, variables map[string]interface{}) (interface{}, error) {
	logging.Info("🔍 Executing embedding usage with discovery: %s", usage.Type)
	
	switch usage.Type {
	case domain.EmbeddingUsageVectorSearch:
		return es.executeVectorSearchWithDiscovery(ctx, embedding, usage.VectorSearch)
		
	case domain.EmbeddingUsageSimilarity:
		return es.executeSimilarityComparison(ctx, embedding, usage.Similarity, variables)
		
	case domain.EmbeddingUsageClustering:
		return es.executeClustering(ctx, embedding, usage.Clustering)
		
	case domain.EmbeddingUsageStorage:
		return map[string]interface{}{
			"stored":       true,
			"embedding_id": embedding.ID,
			"dimensions":   embedding.Dimensions,
			"chunk_count":  embedding.ChunkCount,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported embedding usage type: %s", usage.Type)
	}
}

// executeVectorSearchWithDiscovery performs vector search with robust tool discovery and fallback strategies
func (es *EnhancedService) executeVectorSearchWithDiscovery(ctx context.Context, embedding *domain.EmbeddingResult, config *domain.VectorSearchConfig) (interface{}, error) {
	vectors := embedding.GetVectors()
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no vectors available for search")
	}
	
	searchVector := vectors[0]
	logging.Info("🔍 Executing vector search with %d-dimensional vector", len(searchVector))
	
	// Use robust tool discovery
	discoveryConfig := DefaultToolDiscoveryConfig()
	toolMatcher := NewToolMatcher(config.Server)
	
	// Wait for initial tool registration to complete
	if discoveryConfig.InitialWait > 0 {
		logging.Debug("⏳ Waiting %v for tool registration to complete", discoveryConfig.InitialWait)
		time.Sleep(discoveryConfig.InitialWait)
	}
	
	var lastErr error
	for attempt := 1; attempt <= discoveryConfig.MaxRetries; attempt++ {
		logging.Debug("🔄 Tool discovery attempt %d/%d", attempt, discoveryConfig.MaxRetries)
		
		// Discover available tools
		availableTools, err := es.discoverAvailableTools(ctx)
		if err != nil {
			lastErr = fmt.Errorf("failed to discover tools on attempt %d: %w", attempt, err)
			logging.Warn("Tool discovery failed: %v", lastErr)
			
			if attempt < discoveryConfig.MaxRetries {
				time.Sleep(discoveryConfig.RetryDelay)
				continue
			}
			return nil, lastErr
		}
		
		if len(availableTools) == 0 {
			lastErr = fmt.Errorf("no tools available on attempt %d", attempt)
			logging.Warn("%v", lastErr)
			
			if attempt < discoveryConfig.MaxRetries {
				time.Sleep(discoveryConfig.RetryDelay)
				continue
			}
			return nil, lastErr
		}
		
		logging.Debug("📋 Discovered %d tools on attempt %d", len(availableTools), attempt)
		
		// Find the best tool for vector search
		toolChoice, strategy, err := toolMatcher.FindBestVectorSearchTool(availableTools)
		if err != nil {
			lastErr = fmt.Errorf("no suitable vector search tool found on attempt %d: %w", attempt, err)
			logging.Warn("%v", lastErr)
			
			if attempt < discoveryConfig.MaxRetries {
				time.Sleep(discoveryConfig.RetryDelay)
				continue
			}
			return nil, lastErr
		}
		
		logging.Info("🎯 Selected tool: %s (strategy: %s)", toolChoice.Function.Name, strategy)
		
		// Execute the vector search using the selected tool and strategy
		result, err := es.executeVectorSearchWithTool(ctx, searchVector, config, toolChoice, strategy, discoveryConfig.ToolTimeout)
		if err != nil {
			lastErr = fmt.Errorf("tool execution failed on attempt %d: %w", attempt, err)
			logging.Warn("%v", lastErr)
			
			// For tool execution failures, retry with a longer delay
			if attempt < discoveryConfig.MaxRetries {
				retryDelay := discoveryConfig.RetryDelay * time.Duration(attempt) // Exponential backoff
				logging.Debug("⏳ Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return nil, lastErr
		}
		
		logging.Info("✅ Vector search completed successfully on attempt %d", attempt)
		return result, nil
	}
	
	return nil, fmt.Errorf("vector search failed after %d attempts: %w", discoveryConfig.MaxRetries, lastErr)
}

// discoverAvailableTools robustly discovers available MCP tools
func (es *EnhancedService) discoverAvailableTools(ctx context.Context) ([]domain.Tool, error) {
	logging.Debug("🔍 Discovering available MCP tools...")
	
	toolCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	type toolResult struct {
		tools []domain.Tool
		err   error
	}
	
	resultCh := make(chan toolResult, 1)
	go func() {
		tools, err := es.serverManager.GetAvailableTools()
		resultCh <- toolResult{tools: tools, err: err}
	}()
	
	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, fmt.Errorf("failed to get available tools: %w", result.err)
		}
		
		logging.Debug("📋 Discovered %d tools", len(result.tools))
		for _, tool := range result.tools {
			logging.Debug("  - %s: %s", tool.Function.Name, tool.Function.Description)
		}
		
		return result.tools, nil
		
	case <-toolCtx.Done():
		return nil, fmt.Errorf("tool discovery timed out after 10 seconds")
	}
}

// FindBestVectorSearchTool finds the best tool for vector search using multiple strategies
func (tm *ToolMatcher) FindBestVectorSearchTool(availableTools []domain.Tool) (*domain.Tool, ToolStrategy, error) {
	logging.Debug("🔍 Finding best vector search tool among %d tools", len(availableTools))
	
	// Strategy 1: Look for native vector search tools (highest priority)
	if tool := tm.findToolByPattern(availableTools, []string{
		"similarity_search", "vector_search", "knn_search", "range_search",
	}); tool != nil {
		logging.Debug("✅ Found native vector search tool: %s", tool.Function.Name)
		return tool, StrategyNativeVectorSearch, nil
	}
	
	// Strategy 2: Look for text-based vector search tools
	if tool := tm.findToolByPattern(availableTools, []string{
		"vector_search_by_text", "text_search", "search_by_text",
	}); tool != nil {
		logging.Debug("✅ Found text-based vector search tool: %s", tool.Function.Name)
		return tool, StrategyTextBasedSearch, nil
	}
	
	// Strategy 3: Look for generic SQL execution tools
	if tool := tm.findToolByPattern(availableTools, []string{
		"execute_query", "query", "sql", "execute_sql", "run_query",
	}); tool != nil {
		logging.Debug("✅ Found generic SQL tool: %s", tool.Function.Name)
		return tool, StrategyGenericSQL, nil
	}
	
	// Strategy 4: Server-specific fallbacks
	serverSpecificTool, strategy := tm.findServerSpecificTool(availableTools)
	if serverSpecificTool != nil {
		logging.Debug("✅ Found server-specific tool: %s (strategy: %s)", serverSpecificTool.Function.Name, strategy)
		return serverSpecificTool, strategy, nil
	}
	
	// Create detailed error message
	toolNames := make([]string, len(availableTools))
	for i, tool := range availableTools {
		toolNames[i] = tool.Function.Name
	}
	
	return nil, "", fmt.Errorf("no suitable vector search tool found for server '%s'. Available tools: %v", tm.serverName, toolNames)
}

func (tm *ToolMatcher) findToolByPattern(tools []domain.Tool, patterns []string) *domain.Tool {
	for _, pattern := range patterns {
		for i, tool := range tools {
			toolNameLower := strings.ToLower(tool.Function.Name)
			if toolNameLower == pattern || strings.HasSuffix(toolNameLower, "_"+pattern) {
				return &tools[i]
			}
		}
	}
	
	for _, pattern := range patterns {
		for i, tool := range tools {
			toolNameLower := strings.ToLower(tool.Function.Name)
			if strings.Contains(toolNameLower, pattern) {
				return &tools[i]
			}
		}
	}
	
	return nil
}

func (tm *ToolMatcher) findServerSpecificTool(tools []domain.Tool) (*domain.Tool, ToolStrategy) {
	serverLower := strings.ToLower(tm.serverName)
	
	if strings.Contains(serverLower, "pgvector") || strings.Contains(serverLower, "postgres") {
		patterns := []string{"similarity_search", "knn_search", "vector_search"}
		if tool := tm.findToolByPattern(tools, patterns); tool != nil {
			return tool, StrategyNativeVectorSearch
		}
	}
	
	return nil, ""
}

func (es *EnhancedService) executeVectorSearchWithTool(ctx context.Context, searchVector []float32, config *domain.VectorSearchConfig, tool *domain.Tool, strategy ToolStrategy, timeout time.Duration) (interface{}, error) {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	params, err := es.prepareToolParameters(searchVector, config, tool, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tool parameters: %w", err)
	}
	
	result, err := es.executeToolWithRetry(execCtx, tool.Function.Name, params)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	
	if result == "" {
		return "No matching records found for the search query.", nil
	}
	
	return result, nil
}

func (es *EnhancedService) prepareToolParameters(searchVector []float32, config *domain.VectorSearchConfig, tool *domain.Tool, strategy ToolStrategy) (map[string]interface{}, error) {
	switch strategy {
	case StrategyNativeVectorSearch:
		return es.prepareNativeVectorSearchParams(searchVector, config)
	case StrategyGenericSQL:
		return es.prepareGenericSQLParams(searchVector, config)
	default:
		return es.prepareNativeVectorSearchParams(searchVector, config)
	}
}

func (es *EnhancedService) prepareNativeVectorSearchParams(searchVector []float32, config *domain.VectorSearchConfig) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"table_name":   config.Table,
		"query_vector": searchVector,
		"limit":        config.MaxResults,
		"threshold":    config.SimilarityThreshold,
	}
	
	if len(config.Filters) > 0 {
		params["filter"] = buildFiltersClause(config.Filters)
	}
	
	return params, nil
}

func (es *EnhancedService) prepareGenericSQLParams(searchVector []float32, config *domain.VectorSearchConfig) (map[string]interface{}, error) {
	vectorStr := vectorToPostgresArray(searchVector)
	
	var whereClause string
	if len(config.Filters) > 0 {
		filterStr := buildFiltersClause(config.Filters)
		whereClause = fmt.Sprintf("AND %s", filterStr)
	}
	
	sql := fmt.Sprintf(`SELECT %s, ROUND((1 - (%s <=> '%s'))::numeric, 4) as similarity
FROM %s 
WHERE 1 - (%s <=> '%s') > %f
%s
ORDER BY %s <=> '%s' 
LIMIT %d;`,
		strings.Join(config.TextColumns, ", "),
		config.VectorColumn,
		vectorStr,
		config.Table,
		config.VectorColumn,
		vectorStr,
		config.SimilarityThreshold,
		whereClause,
		config.VectorColumn,
		vectorStr,
		config.MaxResults)
	
	return map[string]interface{}{
		"query":     sql,
		"sql":       sql,
		"statement": sql,
		"command":   sql,
	}, nil
}

func (es *EnhancedService) executeToolWithRetry(ctx context.Context, toolName string, params map[string]interface{}) (string, error) {
	var lastErr error
	
	if sql, hasSql := params["sql"]; hasSql {
		if query, hasQuery := params["query"]; hasQuery && sql == query {
			paramVariations := []map[string]interface{}{
				{"query": sql},
				{"sql": sql},
				{"statement": sql},
				{"command": sql},
			}
			
			for _, paramSet := range paramVariations {
				result, err := es.serverManager.ExecuteTool(ctx, toolName, paramSet)
				if err == nil {
					return result, nil
				}
				
				lastErr = err
				if !isParameterFormatError(err) {
					break
				}
			}
		}
	}
	
	result, err := es.serverManager.ExecuteTool(ctx, toolName, params)
	if err != nil {
		if lastErr != nil {
			return "", fmt.Errorf("all parameter formats failed, last error: %w", lastErr)
		}
		return "", err
	}
	
	return result, nil
}
