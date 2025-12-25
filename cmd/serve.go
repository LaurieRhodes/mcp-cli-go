package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/server"
	serverService "github.com/LaurieRhodes/mcp-cli-go/internal/services/server"
	"github.com/spf13/cobra"
)

var (
	// Serve command flags
	serveConfig string
)

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve [runas-config]",
	Short: "Run as an MCP server exposing workflow templates as tools",
	Long: `Serve mode runs mcp-cli as an MCP server, exposing your workflow templates
as callable MCP tools that other applications can use.

This allows applications like Claude Desktop, IDEs, or other MCP clients to:
  • Execute your custom workflow templates as tools
  • Chain multiple AI operations together
  • Access your configured AI providers and MCP servers

The serve command requires a "runas" configuration file that defines:
  • Server name and version
  • Which templates to expose as tools
  • Input/output mappings for each tool
  • Optional provider/model overrides

Example usage:
  # Start MCP server with specific config
  mcp-cli serve config/runas/research_agent.yaml
  
  # With verbose logging for debugging
  mcp-cli serve --verbose config/runas/code_reviewer.yaml
  
  # Using the --serve flag
  mcp-cli --serve config/runas/data_analyst.yaml

Claude Desktop Configuration:
  Add to your Claude Desktop config (claude_desktop_config.json):
  
  {
    "mcpServers": {
      "research-agent": {
        "command": "/path/to/mcp-cli",
        "args": ["serve", "/path/to/config/runas/research_agent.yaml"]
      }
    }
  }`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine runas config path
		runasConfigPath := serveConfig
		if len(args) > 0 {
			runasConfigPath = args[0]
		}
		
		if runasConfigPath == "" {
			return fmt.Errorf("runas config file is required")
		}
		
		// Set logging to ERROR by default for clean MCP protocol
		if !verbose {
			logging.SetDefaultLevel(logging.ERROR)
		}
		
		logging.Info("Starting MCP server mode with config: %s", runasConfigPath)
		
		// Load runas config
		runasLoader := runas.NewLoader()
		runasConfig, created, err := runasLoader.LoadOrDefault(runasConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load runas config: %w", err)
		}
		
		if created {
			fmt.Fprintf(os.Stderr, "Created example runas config at: %s\n", runasConfigPath)
			fmt.Fprintf(os.Stderr, "Please edit the file to configure your MCP server.\n")
			return nil
		}
		
		logging.Info("Loaded runas config: %s", runasConfig.ServerInfo.Name)
		
		// Determine config file location - always relative to the binary
		actualConfigFile := configFile
		if actualConfigFile == "config.yaml" {
			// Default value - look in same directory as binary
			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to determine executable path: %w", err)
			}
			exeDir := filepath.Dir(exePath)
			actualConfigFile = filepath.Join(exeDir, "config.yaml")
			logging.Info("Using config file: %s", actualConfigFile)
		}
		
		// Load application config
		configService := infraConfig.NewService()
		appConfig, err := configService.LoadConfig(actualConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load application config from %s: %w", actualConfigFile, err)
		}
		
		// Validate templates exist
		for i, tool := range runasConfig.Tools {
			_, existsV1 := appConfig.Templates[tool.Template]
			_, existsV2 := appConfig.TemplatesV2[tool.Template]
			
			if !existsV1 && !existsV2 {
				return fmt.Errorf("tool %d (%s) references unknown template: %s", 
					i, tool.Name, tool.Template)
			}
		}
		
		// Create server service
		service := serverService.NewService(runasConfig, appConfig, configService)
		
		// Create stdio server
		stdioServer := server.NewStdioServer(service)
		
		// Start server
		logging.Info("MCP server starting...")
		if err := stdioServer.Start(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		
		return nil
	},
}

func init() {
	ServeCmd.Flags().StringVar(&serveConfig, "serve", "", "Path to runas config file")
	RootCmd.AddCommand(ServeCmd)
}
