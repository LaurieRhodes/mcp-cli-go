package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"gopkg.in/yaml.v3"
)

// Service provides high-level RAG operations
type Service struct {
	configService   *infraConfig.Service
	serverManager   domain.MCPServerManager
	embeddingService domain.EmbeddingService
	retriever       *MultiVectorRetriever
	expander        *QueryExpander
	ragConfig       *config.RagConfig
}

// NewService creates a new RAG service
func NewService(configService *infraConfig.Service, serverManager domain.MCPServerManager, embeddingService domain.EmbeddingService) (*Service, error) {
	// Load RAG configuration
	ragConfig := configService.GetRagConfig()

	return NewServiceWithConfig(ragConfig, serverManager, embeddingService), nil
}

// NewServiceWithConfig creates a new RAG service with provided config
func NewServiceWithConfig(ragConfig *config.RagConfig, serverManager domain.MCPServerManager, embeddingService domain.EmbeddingService) *Service {
	// Initialize retriever
	retriever := NewMultiVectorRetriever(serverManager, ragConfig.DefaultServer)

	// Load expansion dictionaries
	expansionConfig, err := loadExpansionConfig(ragConfig.QueryExpansion)
	if err != nil {
		logging.Warn("Failed to load query expansion config: %v", err)
		// Continue with empty expansion config
		expansionConfig = TermExpansionConfig{}
	}

	// Initialize expander
	expander := NewQueryExpander(expansionConfig)

	return &Service{
		configService:   nil, // Not needed when config is provided directly
		serverManager:   serverManager,
		embeddingService: embeddingService,
		retriever:       retriever,
		expander:        expander,
		ragConfig:       ragConfig,
	}
}

// SearchRequest represents a RAG search request
type SearchRequest struct {
	Query       string   // Search query
	Server      string   // Server name (from config)
	Strategies  []string // Strategy names to use
	TopK        int      // Number of results
	Fusion      string   // Fusion method (rrf, weighted, max, avg)
	ExpandQuery bool     // Enable query expansion
	Filters     map[string]interface{} // Additional filters
}

// SearchResponse represents a RAG search response
type SearchResponse struct {
	Query            string                 `json:"query"`
	ExpandedQuery    *ExpandedQuery        `json:"expanded_query,omitempty"`
	Results          []SearchResult         `json:"results"`
	Strategy         string                 `json:"strategy,omitempty"`
	Fusion           string                 `json:"fusion,omitempty"`
	TotalResults     int                    `json:"total_results"`
	ExecutionTimeMs  int64                  `json:"execution_time_ms"`
}

// Search performs a RAG search
func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	logging.Info("🔍 RAG Search: query=%s, server=%s, strategies=%v", req.Query, req.Server, req.Strategies)

	// Use defaults if not specified
	if req.Server == "" {
		req.Server = s.ragConfig.DefaultServer
	}
	if req.TopK == 0 {
		req.TopK = s.ragConfig.DefaultTopK
	}
	if req.Fusion == "" {
		req.Fusion = s.ragConfig.DefaultFusion
	}

	// Get server config
	serverConfig, exists := s.ragConfig.Servers[req.Server]
	if !exists {
		return nil, fmt.Errorf("server %s not found in RAG config", req.Server)
	}

	// Query expansion
	var expandedQuery *ExpandedQuery
	if req.ExpandQuery && s.ragConfig.QueryExpansion.Enabled {
		logging.Debug("🔄 Expanding query...")
		expansionConfig := QueryExpansionConfig{
			EnableSynonymExpansion: true,
			EnableAcronymExpansion: true,
			TermExpansion: TermExpansionConfig{
				MaxExpansions: s.ragConfig.QueryExpansion.MaxExpansions,
			},
		}
		
		var err error
		expandedQuery, err = s.expander.ExpandQuery(ctx, req.Query, expansionConfig)
		if err != nil {
			logging.Warn("Query expansion failed: %v", err)
		} else {
			logging.Debug("✅ Query expanded: %d variants", len(expandedQuery.ExpandedVariants))
		}
	}

	// Build multi-vector search config
	searchConfig := s.buildSearchConfig(req, serverConfig)

	// Generate query embedding using configured method
	queryVector, err := s.generateQueryEmbedding(ctx, req.Query, serverConfig, req.Strategies)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Execute search
	results, err := s.retriever.Search(ctx, queryVector, searchConfig)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	logging.Info("✅ RAG Search completed: %d results", len(results))

	return &SearchResponse{
		Query:         req.Query,
		ExpandedQuery: expandedQuery,
		Results:       results,
		Fusion:        req.Fusion,
		TotalResults:  len(results),
	}, nil
}

// generateQueryEmbedding generates an embedding vector for the query text
func (s *Service) generateQueryEmbedding(ctx context.Context, query string, serverConfig config.RagServerConfig, strategies []string) ([]float32, error) {
	// Determine which embedding config to use
	var embeddingConfig *config.QueryEmbeddingConfig
	
	// If using a single strategy, check if it has a specific embedding config
	if len(strategies) == 1 {
		for _, strategy := range serverConfig.Strategies {
			if strategy.Name == strategies[0] && strategy.QueryEmbedding != nil {
				embeddingConfig = strategy.QueryEmbedding
				logging.Debug("Using strategy-specific embedding config for: %s", strategy.Name)
				break
			}
		}
	}
	
	// Fall back to server-level default
	if embeddingConfig == nil {
		embeddingConfig = serverConfig.QueryEmbedding
		logging.Debug("Using server-level embedding config")
	}
	
	// If still no config, return error
	if embeddingConfig == nil {
		return nil, fmt.Errorf("no query embedding configuration found for server %s", serverConfig.MCPServer)
	}
	
	// Generate embedding based on configured type
	switch embeddingConfig.Type {
	case "mcp_tool":
		return s.generateEmbeddingViaMCPTool(ctx, query, embeddingConfig)
	case "service":
		return s.generateEmbeddingViaService(ctx, query, embeddingConfig)
	default:
		return nil, fmt.Errorf("unknown embedding type: %s (must be 'mcp_tool' or 'service')", embeddingConfig.Type)
	}
}

// generateEmbeddingViaMCPTool generates embedding using an MCP tool
func (s *Service) generateEmbeddingViaMCPTool(ctx context.Context, query string, config *config.QueryEmbeddingConfig) ([]float32, error) {
	if config.ToolName == "" {
		return nil, fmt.Errorf("tool_name is required when type=mcp_tool")
	}
	
	// Build tool parameters
	params := map[string]interface{}{
		"texts": []string{query},
	}
	
	// Add default parameters from config
	for k, v := range config.DefaultParams {
		params[k] = v
	}
	
	logging.Debug("Calling MCP tool %s with params: %v", config.ToolName, params)
	
	// Execute tool via server manager
	rawResult, err := s.serverManager.ExecuteTool(ctx, config.ToolName, params)
	if err != nil {
		return nil, fmt.Errorf("MCP tool %s failed: %w", config.ToolName, err)
	}
	
	// Parse JSON result
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rawResult), &result); err != nil {
		return nil, fmt.Errorf("failed to parse result from %s: %w", config.ToolName, err)
	}
	
	// Expected: { "embeddings": [ { "vector": [...], ... } ] }
	embeddings, ok := result["embeddings"].([]interface{})
	if !ok || len(embeddings) == 0 {
		return nil, fmt.Errorf("unexpected response structure from %s: no embeddings array", config.ToolName)
	}
	
	firstEmbedding, ok := embeddings[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected embedding structure from %s", config.ToolName)
	}
	
	vectorInterface, ok := firstEmbedding["vector"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no vector field in embedding result from %s", config.ToolName)
	}
	
	// Convert []interface{} of floats to []float32
	vector := make([]float32, len(vectorInterface))
	for i, v := range vectorInterface {
		switch val := v.(type) {
		case float64:
			vector[i] = float32(val)
		case float32:
			vector[i] = val
		default:
			return nil, fmt.Errorf("unexpected vector element type at index %d: %T", i, v)
		}
	}
	
	logging.Debug("Generated embedding vector of dimension: %d", len(vector))
	
	return vector, nil
}

// generateEmbeddingViaService generates embedding using mcp-cli's embedding service
func (s *Service) generateEmbeddingViaService(ctx context.Context, query string, config *config.QueryEmbeddingConfig) ([]float32, error) {
	if s.embeddingService == nil {
		return nil, fmt.Errorf("embedding service not available")
	}
	
	// Build embedding request
	req := &domain.EmbeddingJobRequest{
		Input:    query,
		Provider: config.Provider,
		Model:    config.Model,
	}
	
	// Add optional parameters
	if dimensions, ok := config.DefaultParams["dimensions"].(int); ok {
		req.Dimensions = dimensions
	}
	if encodingFormat, ok := config.DefaultParams["encoding_format"].(string); ok {
		req.EncodingFormat = encodingFormat
	}
	
	logging.Debug("Generating embedding via service with provider=%s, model=%s", req.Provider, req.Model)
	
	// Generate embedding
	job, err := s.embeddingService.GenerateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding service failed: %w", err)
	}
	
	// Extract vector from first embedding
	if len(job.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	
	vector := job.Embeddings[0].Vector
	logging.Debug("Generated embedding vector of dimension: %d via service", len(vector))
	
	return vector, nil
}

// buildSearchConfig builds a MultiVectorSearchConfig from request and server config
func (s *Service) buildSearchConfig(req SearchRequest, serverConfig config.RagServerConfig) MultiVectorSearchConfig {
	searchConfig := MultiVectorSearchConfig{
		Table:              serverConfig.Table,
		TextColumns:        serverConfig.TextColumns,
		MetadataColumns:    serverConfig.MetadataColumns,
		GlobalMaxResults:   req.TopK,
		CombinationMethod:  req.Fusion,
	}

	// Build vector column configs
	for _, strategyName := range req.Strategies {
		// Find strategy in server config
		for _, strategy := range serverConfig.Strategies {
			if strategy.Name == strategyName {
				vectorCol := VectorColumnConfig{
					Name:                strategy.VectorColumn,
					Weight:              strategy.Weight,
					SimilarityThreshold: strategy.Threshold,
					MaxResults:          req.TopK,
					Filters:             strategy.Filters,
				}
				
				// Apply request-level filters
				if req.Filters != nil {
					if vectorCol.Filters == nil {
						vectorCol.Filters = make(map[string]interface{})
					}
					for k, v := range req.Filters {
						vectorCol.Filters[k] = v
					}
				}
				
				searchConfig.VectorColumns = append(searchConfig.VectorColumns, vectorCol)
				break
			}
		}
	}

	return searchConfig
}

// ExpandQuery expands a query using configured expansion strategies
func (s *Service) ExpandQuery(ctx context.Context, query string) (*ExpandedQuery, error) {
	if !s.ragConfig.QueryExpansion.Enabled {
		return &ExpandedQuery{
			Original:         query,
			ExpandedVariants: []string{query},
		}, nil
	}

	expansionConfig := QueryExpansionConfig{
		EnableSynonymExpansion: true,
		EnableAcronymExpansion: true,
		TermExpansion: TermExpansionConfig{
			MaxExpansions: s.ragConfig.QueryExpansion.MaxExpansions,
		},
	}

	return s.expander.ExpandQuery(ctx, query, expansionConfig)
}

// GetServerConfig returns the RAG config for a server
func (s *Service) GetServerConfig(serverName string) (*config.RagServerConfig, error) {
	serverConfig, exists := s.ragConfig.Servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server %s not found in RAG config", serverName)
	}
	return &serverConfig, nil
}

// ListServers returns list of configured RAG servers
func (s *Service) ListServers() []string {
	servers := make([]string, 0, len(s.ragConfig.Servers))
	for name := range s.ragConfig.Servers {
		servers = append(servers, name)
	}
	return servers
}

// ListStrategies returns list of strategies for a server
func (s *Service) ListStrategies(serverName string) ([]string, error) {
	serverConfig, err := s.GetServerConfig(serverName)
	if err != nil {
		return nil, err
	}

	strategies := make([]string, 0, len(serverConfig.Strategies))
	for _, strategy := range serverConfig.Strategies {
		strategies = append(strategies, strategy.Name)
	}
	return strategies, nil
}

// loadExpansionConfig loads synonyms and acronyms for query expansion
func loadExpansionConfig(settings config.QueryExpansionSettings) (TermExpansionConfig, error) {
	expansionConfig := TermExpansionConfig{
		MaxExpansions: settings.MaxExpansions,
		CaseSensitive: settings.CaseSensitive,
		Synonyms:      make(map[string][]string),
		Acronyms:      make(map[string][]string),
	}

	// Load synonyms
	if settings.SynonymsFile != "" {
		synonyms, err := loadDictionary(settings.SynonymsFile)
		if err != nil {
			logging.Warn("Failed to load synonyms from %s: %v", settings.SynonymsFile, err)
		} else {
			expansionConfig.Synonyms = synonyms
			logging.Debug("Loaded %d synonym entries", len(synonyms))
		}
	}

	// Load acronyms
	if settings.AcronymsFile != "" {
		acronyms, err := loadDictionary(settings.AcronymsFile)
		if err != nil {
			logging.Warn("Failed to load acronyms from %s: %v", settings.AcronymsFile, err)
		} else {
			expansionConfig.Acronyms = acronyms
			logging.Debug("Loaded %d acronym entries", len(acronyms))
		}
	}

	return expansionConfig, nil
}

// loadDictionary loads a YAML dictionary file
func loadDictionary(filepath string) (map[string][]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var dict map[string][]string
	if err := yaml.Unmarshal(data, &dict); err != nil {
		return nil, err
	}

	return dict, nil
}
