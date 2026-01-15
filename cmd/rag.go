package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/rag"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/embeddings"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
)

// RAG command flags
var (
	ragShowConfig  bool
	ragServer      string
	ragStrategies  []string
	ragTopK        int
	ragFusion      string
	ragExpandQuery bool
)

// RagCmd represents the rag command
var RagCmd = &cobra.Command{
	Use:   "rag",
	Short: "RAG (Retrieval-Augmented Generation) operations",
	Long: `Perform RAG operations using MCP vector servers.

RAG enables semantic search across vector databases connected via MCP.
Supports multi-strategy search, query expansion, and result fusion.

Examples:
  # Show RAG configuration
  mcp-cli rag config
  
  # Search directly
  mcp-cli rag search "What are the MFA requirements?"`,
}

// RagConfigCmd shows RAG configuration
var RagConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show RAG configuration",
	Long:  `Display the current RAG configuration including servers, strategies, and fusion settings.`,
	RunE:  executeRagConfig,
}

// RagSearchCmd performs RAG search
var RagSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search using RAG",
	Long: `Perform a RAG search against configured vector databases.

RAG (Retrieval-Augmented Generation) uses semantic similarity to find
relevant documents from vector databases. Results are ranked by relevance.

Examples:
  # Basic search
  mcp-cli rag search "authentication requirements"
  
  # Search with more results
  mcp-cli rag search "access control policies" --top-k 10
  
  # Use specific server
  mcp-cli rag search "encryption" --server pgvector
  
  # Multi-strategy search with fusion
  mcp-cli rag search "security controls" --strategies default,context --fusion rrf
  
  # Enable query expansion
  mcp-cli rag search "MFA" --expand

Output:
  Returns JSON with query, results, and metadata including:
  - Matched document identifiers and text
  - Similarity scores (higher = more relevant)
  - Total results and execution time`,
	Args:  cobra.ExactArgs(1),
	RunE:  executeRagSearch,
}

func init() {
	RagConfigCmd.Flags().BoolVar(&ragShowConfig, "verbose", false, "Show detailed configuration")
	
	RagSearchCmd.Flags().StringVar(&ragServer, "server", "", "RAG server to use (default from config)")
	RagSearchCmd.Flags().StringSliceVar(&ragStrategies, "strategies", []string{"default"}, "Strategies to use")
	RagSearchCmd.Flags().IntVar(&ragTopK, "top-k", 5, "Number of results")
	RagSearchCmd.Flags().StringVar(&ragFusion, "fusion", "", "Fusion method (rrf, weighted, max, avg)")
	RagSearchCmd.Flags().BoolVar(&ragExpandQuery, "expand", false, "Enable query expansion")
	
	RagCmd.AddCommand(RagConfigCmd)
	RagCmd.AddCommand(RagSearchCmd)
}

func executeRagConfig(cmd *cobra.Command, args []string) error {
	// Initialize config service
	configService := config.NewService()
	
	// Load configuration
	_, err := configService.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Load RAG configuration
	ragConfig := configService.GetRagConfig()
	
	// Display configuration
	fmt.Println("=== RAG Configuration ===")
	fmt.Printf("Default Server: %s\n", ragConfig.DefaultServer)
	fmt.Printf("Default Fusion: %s\n", ragConfig.DefaultFusion)
	fmt.Printf("Default Top-K: %d\n", ragConfig.DefaultTopK)
	fmt.Println()
	
	fmt.Println("=== Configured Servers ===")
	if len(ragConfig.Servers) == 0 {
		fmt.Println("No servers configured. Create config/rag.yaml to configure RAG.")
		fmt.Println("See config/rag.yaml.example for template.")
		return nil
	}
	
	for name, srv := range ragConfig.Servers {
		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  MCP Server: %s\n", srv.MCPServer)
		if srv.SearchTool != "" {
			fmt.Printf("  Search Tool: %s\n", srv.SearchTool)
		}
		fmt.Printf("  Table: %s\n", srv.Table)
		fmt.Printf("  Strategies: %d configured\n", len(srv.Strategies))
		
		if ragShowConfig {
			for _, strategy := range srv.Strategies {
				fmt.Printf("    - %s (column: %s, weight: %.2f, threshold: %.2f)\n",
					strategy.Name, strategy.VectorColumn, strategy.Weight, strategy.Threshold)
			}
		}
	}
	
	fmt.Println()
	fmt.Println("=== Query Expansion ===")
	fmt.Printf("Enabled: %v\n", ragConfig.QueryExpansion.Enabled)
	if ragConfig.QueryExpansion.Enabled {
		fmt.Printf("Max Expansions: %d\n", ragConfig.QueryExpansion.MaxExpansions)
		if ragConfig.QueryExpansion.SynonymsFile != "" {
			fmt.Printf("Synonyms: %s\n", ragConfig.QueryExpansion.SynonymsFile)
		}
		if ragConfig.QueryExpansion.AcronymsFile != "" {
			fmt.Printf("Acronyms: %s\n", ragConfig.QueryExpansion.AcronymsFile)
		}
	}
	
	fmt.Println()
	logging.Info("✓ RAG configuration loaded successfully")
	
	return nil
}

func executeRagSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	
	// Load config
	configService := config.NewService()
	_, err := configService.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Get RAG configuration
	ragConfig := configService.GetRagConfig()
	if ragConfig == nil {
		return fmt.Errorf("no RAG configuration found")
	}
	
	// Determine server
	serverName := ragServer
	if serverName == "" {
		serverName = ragConfig.DefaultServer
	}
	if serverName == "" {
		return fmt.Errorf("no server specified and no default server in config")
	}
	
	var searchErr error
	
	// Determine which servers to connect
	servers := []string{serverName}
	userSpecifiedServers := make(map[string]bool)
	userSpecifiedServers[serverName] = true
	
	// Run with host server connections
	err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
		// Create server manager
		serverManager := NewHostServerManager(conns)
		
		// Create embedding service
		providerFactory := ai.NewProviderFactory()
		embeddingService := embeddings.NewService(configService, providerFactory)
		
		// Create RAG service with embedding service
		ragService := rag.NewServiceWithConfig(ragConfig, serverManager, embeddingService)
		
		// Create search request
		req := rag.SearchRequest{
			Query:       query,
			Server:      serverName,
			Strategies:  ragStrategies,
			TopK:        ragTopK,
			Fusion:      ragFusion,
			ExpandQuery: ragExpandQuery,
		}
		
		// Execute search
		logging.Info("🔍 Searching: %s", query)
		startTime := time.Now()
		
		ctx := context.Background()
		response, err := ragService.Search(ctx, req)
		if err != nil {
			searchErr = fmt.Errorf("search failed: %w", err)
			return searchErr
		}
		
		elapsed := time.Since(startTime)
		
		// Display results
		fmt.Println()
		fmt.Printf("Found %d results in %v\n", response.TotalResults, elapsed)
		fmt.Println()
		
		if response.TotalResults == 0 {
			fmt.Println("No results found.")
			return nil
		}
		
		// Format as JSON
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			searchErr = fmt.Errorf("failed to format results: %w", err)
			return searchErr
		}
		
		fmt.Println(string(output))
		
		return nil
	}, configFile, servers, userSpecifiedServers, &host.CommandOptions{
		
		SuppressConsole: false,
	})
	
	if err != nil {
		return err
	}
	if searchErr != nil {
		return searchErr
	}
	
	return nil
}
