package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// RAG command flags
var (
	ragShowConfig bool
)

// RagCmd represents the rag command
var RagCmd = &cobra.Command{
	Use:   "rag",
	Short: "RAG (Retrieval-Augmented Generation) operations",
	Long: `Perform RAG operations using MCP vector servers.

RAG enables semantic search across vector databases connected via MCP.
Supports multi-strategy search, query expansion, and result fusion.

Primary usage is through workflows. See examples/rag/ for workflow examples.

Examples:
  # Show RAG configuration
  mcp-cli rag config
  
  # Use RAG in workflows
  mcp-cli workflow run examples/rag/search.yaml`,
}

// RagConfigCmd shows RAG configuration
var RagConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show RAG configuration",
	Long:  `Display the current RAG configuration including servers, strategies, and fusion settings.`,
	RunE:  executeRagConfig,
}

// RagSearchCmd shows how to use RAG search
var RagSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Show how to use RAG search (via workflows)",
	Long: `RAG search is primarily used through workflows for maximum flexibility.

Direct search support will be added in a future release.

Example workflow (examples/rag/search.yaml):

  steps:
    - name: retrieve
      rag:
        query: "{{user_query}}"
        server: pgvector
        strategies: [default, technical]
        fusion: rrf
        top_k: 5
        expand_query: true
        
    - name: generate
      llm:
        provider: anthropic
        model: claude-sonnet-4
        prompt: |
          Context: {{step.retrieve.results}}
          Question: {{user_query}}

Run workflow:
  mcp-cli workflow run examples/rag/search.yaml --var user_query="MFA requirements"`,
	RunE: executeRagSearchInfo,
}

func init() {
	RagConfigCmd.Flags().BoolVar(&ragShowConfig, "verbose", false, "Show detailed configuration")
	
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
	
	for name, server := range ragConfig.Servers {
		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  MCP Server: %s\n", server.MCPServer)
		if server.SearchTool != "" {
			fmt.Printf("  Search Tool: %s\n", server.SearchTool)
		}
		fmt.Printf("  Table: %s\n", server.Table)
		fmt.Printf("  Strategies: %d configured\n", len(server.Strategies))
		
		if ragShowConfig {
			for _, strategy := range server.Strategies {
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
	fmt.Println("=== Usage ===")
	fmt.Println("RAG is primarily used through workflows.")
	fmt.Println("Example workflow step:")
	fmt.Println()
	fmt.Println("  steps:")
	fmt.Println("    - name: retrieve")
	fmt.Println("      rag:")
	fmt.Println("        query: \"{{user_query}}\"")
	fmt.Println("        server: pgvector")
	fmt.Println("        strategies: [default, technical]")
	fmt.Println("        fusion: rrf")
	fmt.Println("        top_k: 5")
	fmt.Println()
	
	logging.Info("✓ RAG configuration loaded successfully")
	
	return nil
}

func executeRagSearchInfo(cmd *cobra.Command, args []string) error {
	fmt.Println("RAG Search is used through workflows for maximum flexibility.")
	fmt.Println()
	fmt.Println("Example workflow file (save as rag-search.yaml):")
	fmt.Println()
	fmt.Println("---")
	fmt.Println("name: RAG Search Example")
	fmt.Println("version: 1.0.0")
	fmt.Println()
	fmt.Println("execution:")
	fmt.Println("  provider: anthropic")
	fmt.Println("  model: claude-sonnet-4")
	fmt.Println()
	fmt.Println("steps:")
	fmt.Println("  - name: retrieve")
	fmt.Println("    rag:")
	fmt.Println("      query: \"{{user_query}}\"")
	fmt.Println("      server: pgvector")
	fmt.Println("      strategies:")
	fmt.Println("        - default")
	fmt.Println("        - technical")
	fmt.Println("      fusion: rrf")
	fmt.Println("      top_k: 5")
	fmt.Println("      expand_query: true")
	fmt.Println()
	fmt.Println("  - name: generate")
	fmt.Println("    llm:")
	fmt.Println("      prompt: |")
	fmt.Println("        Based on these retrieved documents:")
	fmt.Println("        {{step.retrieve.results}}")
	fmt.Println()
	fmt.Println("        Answer this question: {{user_query}}")
	fmt.Println()
	fmt.Println("Run with:")
	fmt.Println("  mcp-cli workflow run rag-search.yaml --var user_query=\"What are the MFA requirements?\"")
	fmt.Println()
	fmt.Println("See examples/rag/ directory for more examples.")
	fmt.Println()
	
	return nil
}
