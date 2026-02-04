package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// VectorColumnConfig defines a vector column and its search configuration
type VectorColumnConfig struct {
	Name                string                 `json:"name"`
	Weight              float64                `json:"weight"`
	SimilarityThreshold float64                `json:"similarity_threshold,omitempty"`
	MaxResults          int                    `json:"max_results,omitempty"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
}

// VectorSearchConfig defines configuration for a single vector search operation
type VectorSearchConfig struct {
	Table               string                 `json:"table"`
	VectorColumn        string                 `json:"vector_column"`
	TextColumns         []string               `json:"text_columns"`
	SimilarityThreshold float64                `json:"similarity_threshold"`
	MaxResults          int                    `json:"max_results"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
}

// MultiVectorSearchConfig defines configuration for multi-vector search
type MultiVectorSearchConfig struct {
	Table             string               `json:"table"`
	VectorColumns     []VectorColumnConfig `json:"vector_columns"`
	TextColumns       []string             `json:"text_columns"`
	MetadataColumns   []string             `json:"metadata_columns,omitempty"`
	GlobalMaxResults  int                  `json:"global_max_results"`
	GlobalThreshold   float64              `json:"global_threshold"`
	CombinationMethod string               `json:"combination_method,omitempty"` // "weighted", "rrf", "max", "avg"
	RerankTopK        int                  `json:"rerank_top_k,omitempty"`
}

// SearchResult represents a single search result with combined score
type SearchResult struct {
	ID              string                 `json:"id"`
	Text            map[string]interface{} `json:"text"`
	Metadata        map[string]interface{} `json:"metadata"`
	CombinedScore   float64                `json:"combined_score"`
	ComponentScores map[string]float64     `json:"component_scores"`
	Source          string                 `json:"source"`
}

// MultiVectorRetriever provides advanced multi-vector retrieval capabilities
type MultiVectorRetriever struct {
	serverManager domain.MCPServerManager
	serverName    string
	ragConfig     *config.RagConfig // Full RAG config for looking up search tools
}

// NewMultiVectorRetriever creates a new multi-vector retriever (legacy constructor)
func NewMultiVectorRetriever(serverManager domain.MCPServerManager, serverName string) *MultiVectorRetriever {
	return &MultiVectorRetriever{
		serverManager: serverManager,
		serverName:    serverName,
		ragConfig:     nil,
	}
}

// NewMultiVectorRetrieverWithConfig creates a new multi-vector retriever with RAG config
// This allows the retriever to use configured search tool names instead of pattern matching
func NewMultiVectorRetrieverWithConfig(serverManager domain.MCPServerManager, ragConfig *config.RagConfig) *MultiVectorRetriever {
	return &MultiVectorRetriever{
		serverManager: serverManager,
		serverName:    ragConfig.DefaultServer,
		ragConfig:     ragConfig,
	}
}

// Search performs multi-vector search across configured vector columns
func (mvr *MultiVectorRetriever) Search(ctx context.Context, queryVector []float32, config MultiVectorSearchConfig) ([]SearchResult, error) {
	logging.Info("🔍 Starting multi-vector search across %d vector columns", len(config.VectorColumns))

	if len(config.VectorColumns) == 0 {
		return nil, fmt.Errorf("no vector columns configured for search")
	}

	// Validate configuration
	if err := mvr.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Execute searches across all vector columns
	var allResults []SearchResult
	var searchErrors []error

	for i, vectorCol := range config.VectorColumns {
		logging.Debug("🔎 Searching vector column %d/%d: %s (weight: %.2f)",
			i+1, len(config.VectorColumns), vectorCol.Name, vectorCol.Weight)

		results, err := mvr.searchSingleVectorColumn(ctx, queryVector, vectorCol, config)
		if err != nil {
			logging.Warn("❌ Search failed for vector column %s: %v", vectorCol.Name, err)
			searchErrors = append(searchErrors, fmt.Errorf("column %s: %w", vectorCol.Name, err))
			continue
		}

		// Add source information and apply column weight
		for j := range results {
			results[j].Source = vectorCol.Name
			if results[j].ComponentScores == nil {
				results[j].ComponentScores = make(map[string]float64)
			}
			results[j].ComponentScores[vectorCol.Name] = results[j].CombinedScore
			results[j].CombinedScore *= vectorCol.Weight
		}

		allResults = append(allResults, results...)
		logging.Debug("✅ Found %d results from column %s", len(results), vectorCol.Name)
	}

	// Check if we have any results
	if len(allResults) == 0 {
		if len(searchErrors) > 0 {
			return nil, fmt.Errorf("all vector column searches failed: %v", searchErrors)
		}
		logging.Info("No results found across any vector columns")
		return []SearchResult{}, nil
	}

	logging.Info("📊 Collected %d total results before combination", len(allResults))

	// Combine and rank results
	combinedResults := mvr.combineResults(allResults, config)

	// Apply global filtering and limits
	finalResults := mvr.applyGlobalFilters(combinedResults, config)

	logging.Info("✅ Multi-vector search completed: %d final results", len(finalResults))
	return finalResults, nil
}

// searchSingleVectorColumn searches a single vector column
func (mvr *MultiVectorRetriever) searchSingleVectorColumn(ctx context.Context, queryVector []float32, vectorCol VectorColumnConfig, globalConfig MultiVectorSearchConfig) ([]SearchResult, error) {
	// Prepare search parameters
	threshold := vectorCol.SimilarityThreshold
	if threshold == 0 {
		threshold = globalConfig.GlobalThreshold
	}

	maxResults := vectorCol.MaxResults
	if maxResults == 0 {
		maxResults = globalConfig.GlobalMaxResults
	}

	// Create vector search configuration
	searchConfig := &VectorSearchConfig{
		Table:               globalConfig.Table,
		VectorColumn:        vectorCol.Name,
		TextColumns:         globalConfig.TextColumns,
		SimilarityThreshold: threshold,
		MaxResults:          maxResults,
		Filters:             vectorCol.Filters,
	}

	// Discover available tools
	availableTools, err := mvr.serverManager.GetAvailableTools()
	if err != nil {
		return nil, fmt.Errorf("failed to discover tools: %w", err)
	}

	// Find suitable search tool
	tool, err := mvr.findBestSearchTool(availableTools)
	if err != nil {
		return nil, fmt.Errorf("no suitable search tool found: %w", err)
	}

	// Execute search
	params := mvr.prepareSearchParameters(queryVector, searchConfig, tool)

	rawResult, err := mvr.serverManager.ExecuteTool(ctx, tool.Function.Name, params)
	if err != nil {
		return nil, fmt.Errorf("search execution failed: %w", err)
	}

	// Parse results
	return mvr.parseSearchResults(rawResult, vectorCol.Name, globalConfig.TextColumns, globalConfig.MetadataColumns)
}

// combineResults combines results from multiple vector columns using the specified method
func (mvr *MultiVectorRetriever) combineResults(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	method := strings.ToLower(config.CombinationMethod)
	if method == "" {
		method = "weighted"
	}

	logging.Debug("🔄 Combining %d results using method: %s", len(results), method)

	switch method {
	case "rrf":
		return mvr.combineWithRRF(results, config)
	case "max":
		return mvr.combineWithMax(results, config)
	case "avg", "average":
		return mvr.combineWithAverage(results, config)
	case "weighted":
		fallthrough
	default:
		return mvr.combineWithWeighted(results, config)
	}
}

// combineWithWeighted combines results using weighted scores (default)
func (mvr *MultiVectorRetriever) combineWithWeighted(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	// Group results by ID (if they have unique identifiers)
	resultMap := make(map[string]*SearchResult)

	for _, result := range results {
		// Use text content as key for deduplication if no ID
		key := result.ID
		if key == "" {
			// Create a simple hash of text content for grouping
			key = mvr.createContentKey(result.Text)
		}

		if existing, exists := resultMap[key]; exists {
			// Combine scores from multiple vector columns
			existing.CombinedScore += result.CombinedScore

			// Merge component scores
			for colName, score := range result.ComponentScores {
				existing.ComponentScores[colName] = score
			}
		} else {
			resultMap[key] = &result
		}
	}

	// Convert back to slice and sort
	combined := make([]SearchResult, 0, len(resultMap))
	for _, result := range resultMap {
		combined = append(combined, *result)
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].CombinedScore > combined[j].CombinedScore
	})

	return combined
}

// combineWithRRF combines results using Reciprocal Rank Fusion
func (mvr *MultiVectorRetriever) combineWithRRF(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	k := 60.0 // RRF constant

	// Group results by vector column
	columnResults := make(map[string][]SearchResult)
	for _, result := range results {
		columnResults[result.Source] = append(columnResults[result.Source], result)
	}

	// Sort each column's results by score
	for _, columnRes := range columnResults {
		sort.Slice(columnRes, func(i, j int) bool {
			return columnRes[i].CombinedScore > columnRes[j].CombinedScore
		})
	}

	// Calculate RRF scores
	rrfScores := make(map[string]float64)
	for _, columnRes := range columnResults {
		for rank, result := range columnRes {
			key := mvr.createContentKey(result.Text)
			rrfScores[key] += 1.0 / (k + float64(rank+1))
		}
	}

	// Create combined results with RRF scores
	uniqueResults := make(map[string]SearchResult)
	for _, result := range results {
		key := mvr.createContentKey(result.Text)
		if _, exists := uniqueResults[key]; !exists {
			result.CombinedScore = rrfScores[key]
			uniqueResults[key] = result
		}
	}

	// Convert to slice and sort
	combined := make([]SearchResult, 0, len(uniqueResults))
	for _, result := range uniqueResults {
		combined = append(combined, result)
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].CombinedScore > combined[j].CombinedScore
	})

	return combined
}

// combineWithMax takes the maximum score across vector columns
func (mvr *MultiVectorRetriever) combineWithMax(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	uniqueResults := make(map[string]*SearchResult)

	for _, result := range results {
		key := mvr.createContentKey(result.Text)

		if existing, exists := uniqueResults[key]; exists {
			if result.CombinedScore > existing.CombinedScore {
				existing.CombinedScore = result.CombinedScore
			}
			// Merge component scores
			for colName, score := range result.ComponentScores {
				existing.ComponentScores[colName] = score
			}
		} else {
			uniqueResults[key] = &result
		}
	}

	// Convert to slice and sort
	combined := make([]SearchResult, 0, len(uniqueResults))
	for _, result := range uniqueResults {
		combined = append(combined, *result)
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].CombinedScore > combined[j].CombinedScore
	})

	return combined
}

// combineWithAverage takes the average score across vector columns
func (mvr *MultiVectorRetriever) combineWithAverage(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	type ScoreAccumulator struct {
		TotalScore float64
		Count      int
		Result     SearchResult
	}

	accumMap := make(map[string]*ScoreAccumulator)

	for _, result := range results {
		key := mvr.createContentKey(result.Text)

		if existing, exists := accumMap[key]; existing != nil && exists {
			existing.TotalScore += result.CombinedScore
			existing.Count++
			// Merge component scores
			for colName, score := range result.ComponentScores {
				existing.Result.ComponentScores[colName] = score
			}
		} else {
			accumMap[key] = &ScoreAccumulator{
				TotalScore: result.CombinedScore,
				Count:      1,
				Result:     result,
			}
		}
	}

	// Calculate averages and create final results
	combined := make([]SearchResult, 0, len(accumMap))
	for _, accum := range accumMap {
		accum.Result.CombinedScore = accum.TotalScore / float64(accum.Count)
		combined = append(combined, accum.Result)
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].CombinedScore > combined[j].CombinedScore
	})

	return combined
}

// applyGlobalFilters applies global threshold and result limits
func (mvr *MultiVectorRetriever) applyGlobalFilters(results []SearchResult, config MultiVectorSearchConfig) []SearchResult {
	var filtered []SearchResult

	for _, result := range results {
		if result.CombinedScore >= config.GlobalThreshold {
			filtered = append(filtered, result)
		}
	}

	// Apply global max results limit
	if config.GlobalMaxResults > 0 && len(filtered) > config.GlobalMaxResults {
		filtered = filtered[:config.GlobalMaxResults]
	}

	// Apply reranking if configured
	if config.RerankTopK > 0 && len(filtered) > config.RerankTopK {
		// For now, just take top K - could be enhanced with reranking models
		filtered = filtered[:config.RerankTopK]
	}

	return filtered
}

// Helper methods

func (mvr *MultiVectorRetriever) validateConfig(config MultiVectorSearchConfig) error {
	if config.Table == "" {
		return fmt.Errorf("table name is required")
	}

	if len(config.VectorColumns) == 0 {
		return fmt.Errorf("at least one vector column must be specified")
	}

	// Validate weights sum to reasonable value
	totalWeight := 0.0
	for _, col := range config.VectorColumns {
		if col.Weight <= 0 {
			return fmt.Errorf("vector column %s has invalid weight: %f", col.Name, col.Weight)
		}
		totalWeight += col.Weight
	}

	if totalWeight == 0 {
		return fmt.Errorf("total weight of vector columns cannot be zero")
	}

	return nil
}

func (mvr *MultiVectorRetriever) findBestSearchTool(availableTools []domain.Tool) (*domain.Tool, error) {
	// If we have RAG config, use the configured search tool name
	if mvr.ragConfig != nil {
		// Get the server config for the current server
		serverConfig, exists := mvr.ragConfig.Servers[mvr.serverName]
		if exists && serverConfig.SearchTool != "" {
			// Look for the specifically configured tool
			logging.Debug("Looking for configured search tool: %s", serverConfig.SearchTool)
			for _, tool := range availableTools {
				if tool.Function.Name == serverConfig.SearchTool {
					logging.Info("Found configured search tool: %s", serverConfig.SearchTool)
					return &tool, nil
				}
			}
			logging.Warn("Configured search tool '%s' not found, falling back to pattern matching", serverConfig.SearchTool)
		}
	}

	// Fall back to pattern matching if no RAG config or tool not found
	logging.Debug("Using pattern matching to find search tool")

	// Look for vector search tools in order of preference
	searchPatterns := []string{
		"search_vectors", // pgvector-mcp tool
		"similarity_search",
		"vector_search",
		"knn_search",
		"range_search",
		"search", // generic search tool
		"execute_query",
		"query",
		"sql",
	}

	for _, pattern := range searchPatterns {
		for _, tool := range availableTools {
			toolNameLower := strings.ToLower(tool.Function.Name)
			if strings.Contains(toolNameLower, pattern) {
				logging.Debug("Found search tool via pattern '%s': %s", pattern, tool.Function.Name)
				return &tool, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable search tool found")
}

func (mvr *MultiVectorRetriever) prepareSearchParameters(queryVector []float32, config *VectorSearchConfig, tool *domain.Tool) map[string]interface{} {
	// Convert vector to string format for database
	vectorStr := mvr.vectorToString(queryVector)

	params := map[string]interface{}{
		"table_name":   config.Table,
		"query_vector": queryVector,
		"vector":       vectorStr,
		"limit":        config.MaxResults,
		"threshold":    config.SimilarityThreshold,
	}

	// Add filters if present
	if len(config.Filters) > 0 {
		params["filter"] = config.Filters
		params["filters"] = config.Filters
	}

	// For SQL-based tools, prepare a full query
	toolNameLower := strings.ToLower(tool.Function.Name)
	if strings.Contains(toolNameLower, "query") || strings.Contains(toolNameLower, "sql") {
		sql := mvr.buildVectorSearchSQL(queryVector, config)
		params["query"] = sql
		params["sql"] = sql
		params["statement"] = sql
	}

	return params
}

func (mvr *MultiVectorRetriever) buildVectorSearchSQL(queryVector []float32, config *VectorSearchConfig) string {
	vectorStr := mvr.vectorToString(queryVector)

	selectColumns := strings.Join(config.TextColumns, ", ")
	if selectColumns == "" {
		selectColumns = "*"
	}

	var whereClause string
	if len(config.Filters) > 0 {
		conditions := make([]string, 0, len(config.Filters))
		for key, value := range config.Filters {
			switch v := value.(type) {
			case string:
				conditions = append(conditions, fmt.Sprintf("%s = '%s'", key, strings.ReplaceAll(v, "'", "''")))
			case bool:
				conditions = append(conditions, fmt.Sprintf("%s = %t", key, v))
			default:
				conditions = append(conditions, fmt.Sprintf("%s = '%v'", key, v))
			}
		}
		if len(conditions) > 0 {
			whereClause = "AND " + strings.Join(conditions, " AND ")
		}
	}

	sql := fmt.Sprintf(`
		SELECT %s, 
		       ROUND((1 - (%s <=> '%s'))::numeric, 4) as similarity
		FROM %s 
		WHERE 1 - (%s <=> '%s') > %f
		%s
		ORDER BY %s <=> '%s' 
		LIMIT %d`,
		selectColumns,
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

	return strings.TrimSpace(sql)
}

func (mvr *MultiVectorRetriever) parseSearchResults(rawResult, source string, textColumns, metadataColumns []string) ([]SearchResult, error) {
	if rawResult == "" {
		return []SearchResult{}, nil
	}

	// Try to parse JSON response - handle multiple formats
	var items []map[string]interface{}

	// First, try parsing as a wrapped response with success/results/total_results
	var searchResponse struct {
		Success      bool                     `json:"success"`
		Results      []map[string]interface{} `json:"results"`
		TotalResults int                      `json:"total_results"`
	}

	if err := json.Unmarshal([]byte(rawResult), &searchResponse); err == nil && searchResponse.Results != nil {
		// Successfully parsed as wrapped response
		if !searchResponse.Success {
			return nil, fmt.Errorf("search was not successful")
		}
		items = searchResponse.Results
		logging.Debug("Parsed %d results from wrapped response format", len(items))
	} else {
		// Try parsing as a raw array (pgvector MCP server format)
		if err := json.Unmarshal([]byte(rawResult), &items); err != nil {
			return nil, fmt.Errorf("failed to parse search results JSON (tried wrapped and array formats): %w", err)
		}
		logging.Debug("Parsed %d results from raw array format", len(items))
	}

	var results []SearchResult

	for i, item := range items {
		// Extract ID (try different field names)
		id := fmt.Sprintf("%s_%d", source, i)
		if itemID, ok := item["id"]; ok {
			id = fmt.Sprintf("%v", itemID)
		} else if identifier, ok := item["identifier"]; ok {
			id = fmt.Sprintf("%v", identifier)
		}

		// Extract similarity/distance score
		score := 0.5 // default
		if similarity, ok := item["similarity"].(float64); ok {
			score = similarity
		} else if distance, ok := item["distance"].(float64); ok {
			// Convert distance to similarity (smaller distance = higher similarity)
			score = 1.0 - distance
			if score < 0 {
				score = 0
			}
		}

		// Build text content from specified text columns
		textContent := make(map[string]interface{})
		for _, col := range textColumns {
			if val, ok := item[col]; ok {
				textContent[col] = val
			}
		}

		// If no text columns specified or found, use all non-metadata fields
		if len(textContent) == 0 {
			for key, val := range item {
				if key != "distance" && key != "similarity" && key != "id" {
					textContent[key] = val
				}
			}
		}

		// Build metadata from specified metadata columns
		metadata := map[string]interface{}{
			"source": source,
			"index":  i,
		}
		for _, col := range metadataColumns {
			if val, ok := item[col]; ok {
				metadata[col] = val
			}
		}
		// Also add any fields not in text columns as metadata
		for key, val := range item {
			if _, inText := textContent[key]; !inText && key != "distance" && key != "similarity" {
				metadata[key] = val
			}
		}

		result := SearchResult{
			ID:              id,
			Text:            textContent,
			Metadata:        metadata,
			CombinedScore:   score,
			ComponentScores: map[string]float64{source: score},
			Source:          source,
		}

		results = append(results, result)
	}

	return results, nil
}

func (mvr *MultiVectorRetriever) vectorToString(vector []float32) string {
	parts := make([]string, len(vector))
	for i, v := range vector {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			v = 0.0
		}
		parts[i] = fmt.Sprintf("%.6f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (mvr *MultiVectorRetriever) createContentKey(text map[string]interface{}) string {
	// Simple content-based key for deduplication
	var parts []string
	for _, value := range text {
		if str, ok := value.(string); ok && str != "" {
			parts = append(parts, str)
		}
	}
	content := strings.Join(parts, " ")
	if len(content) > 100 {
		content = content[:100]
	}
	return fmt.Sprintf("content_%x", []byte(content))
}
