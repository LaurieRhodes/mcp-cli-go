package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ServersCmd lists all available MCP servers
var ServersCmd = &cobra.Command{
	Use:   "servers",
	Short: "List all available MCP servers",
	Long: `List all MCP servers configured in the system.

This includes:
- Servers from config/servers/*.yaml
- RunAs servers from config/runas/*.yaml (templates as MCP servers)

Use these server names with --server flag in chat and query modes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listServers()
	},
}

func listServers() error {
	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if we have any servers
	if len(cfg.Servers) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println("\nTo add servers:")
		fmt.Println("  1. Create server configs in config/servers/*.yaml")
		fmt.Println("  2. Create runas configs in config/runas/*.yaml")
		fmt.Println("  3. Run 'mcp-cli init' to set up default configuration")
		return nil
	}

	// Sort server names for consistent output
	serverNames := make([]string, 0, len(cfg.Servers))
	for name := range cfg.Servers {
		serverNames = append(serverNames, name)
	}
	sort.Strings(serverNames)

	// Display servers
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	
	bold.Printf("\nConfigured MCP Servers (%d total):\n", len(cfg.Servers))
	fmt.Println(strings.Repeat("=", 50))

	for _, name := range serverNames {
		server := cfg.Servers[name]
		
		// Determine server type
		serverType := "Traditional MCP Server"
		if len(server.Args) > 0 && server.Args[0] == "serve" {
			serverType = "RunAs MCP Server"
			green.Printf("\n● %s", name)
			cyan.Printf(" (%s)\n", serverType)
			fmt.Printf("  Command: %s serve\n", server.Command)
			if len(server.Args) > 1 {
				fmt.Printf("  Config:  %s\n", server.Args[1])
			}
		} else {
			green.Printf("\n● %s", name)
			cyan.Printf(" (%s)\n", serverType)
			fmt.Printf("  Command: %s\n", server.Command)
			if len(server.Args) > 0 {
				fmt.Printf("  Args:    %v\n", server.Args)
			}
		}
		
		if server.SystemPrompt != "" {
			fmt.Printf("  Prompt:  %s\n", truncate(server.SystemPrompt, 60))
		}
	}

	// Usage examples
	fmt.Println("\n" + strings.Repeat("=", 50))
	bold.Println("\nUsage Examples:")
	fmt.Println("  # Use server in chat mode")
	fmt.Printf("  mcp-cli chat --server %s\n", serverNames[0])
	fmt.Println("\n  # Use server in query mode")
	fmt.Printf("  mcp-cli query --server %s \"your question\"\n", serverNames[0])
	fmt.Println("\n  # Chat with all servers")
	fmt.Println("  mcp-cli chat")
	
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
