package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/server"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/proxy"
	serverService "github.com/LaurieRhodes/mcp-cli-go/internal/services/server"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
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
		
		// === Process templates array (convert to tools) ===
		// For MCP types using the new templates config_source pattern
		if len(runasConfig.Templates) > 0 {
			logging.Info("Processing %d template source(s)...", len(runasConfig.Templates))
			
			for _, templateSrc := range runasConfig.Templates {
				// Extract template name from config_source path
				basename := filepath.Base(templateSrc.ConfigSource)
				templateName := strings.TrimSuffix(basename, filepath.Ext(basename))
				
				// Verify template exists
				_, existsV1 := appConfig.Workflows[templateName]
				templateV2, existsV2 := appConfig.Workflows[templateName]
				
				if !existsV1 && !existsV2 {
					return fmt.Errorf("template source '%s' points to unknown template: %s", 
						templateSrc.ConfigSource, templateName)
				}
				
				// Use custom name if provided, otherwise use template name
				toolName := templateSrc.Name
				if toolName == "" {
					toolName = templateName
				}
				
				// Use custom description if provided, otherwise derive from template
				toolDescription := templateSrc.Description
				if toolDescription == "" && existsV2 {
					toolDescription = templateV2.Description
				}
				
				// Standard input schema for all templates
				// Templates receive input_data as their primary parameter
				inputSchema := map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input_data": map[string]interface{}{
							"type":        "string",
							"description": "Input data for the template workflow",
						},
					},
					"required": []string{"input_data"},
				}
				
				// Create ToolExposure from template source
				tool := runas.ToolExposure{
					Template:    templateName,
					Name:        toolName,
					Description: toolDescription,
					InputSchema: inputSchema,
					InputMapping: map[string]string{
						"input_data": "{{input_data}}",
					},
				}
				
				// Add to tools array
				runasConfig.Tools = append(runasConfig.Tools, tool)
				logging.Info("Created tool '%s' from template '%s' (source: %s)", 
					toolName, templateName, templateSrc.ConfigSource)
			}
			
			logging.Info("Processed %d template(s) into %d total tool(s)", 
				len(runasConfig.Templates), len(runasConfig.Tools))
		}
		
		// Initialize skills service
		skillService := skillsvc.NewService()
		skillService.SetConfig(appConfig)
		
		// === Handle mcp-skills type: Auto-discover and generate tools ===
		if runasConfig.RunAsType == runas.RunAsTypeMCPSkills || runasConfig.RunAsType == runas.RunAsTypeProxySkills {
			logging.Info("Auto-discovering skills for mcp-skills server type")
			
			// Determine skills directory
			var skillsDir string
			if runasConfig.SkillsConfig != nil && runasConfig.SkillsConfig.SkillsDirectory != "" {
				skillsDir = runasConfig.SkillsConfig.SkillsDirectory
				// If it's a relative path, resolve it relative to the config file's directory
				if !filepath.IsAbs(skillsDir) {
					skillsDir = filepath.Join(filepath.Dir(runasConfigPath), skillsDir)
				}
			} else {
				// Default: Get directory containing the RunAs config
				runasDir := filepath.Dir(runasConfigPath)
				// Go up to config directory (from config/runasMCP to config)
				configDir := filepath.Dir(runasDir)
				// Default skills directory is config/skills
				skillsDir = filepath.Join(configDir, "skills")
			}
			
			// Determine execution mode (default: auto)
			execMode := skills.ExecutionModeAuto
			if runasConfig.SkillsConfig != nil && runasConfig.SkillsConfig.ExecutionMode != "" {
				execMode = skills.ExecutionMode(runasConfig.SkillsConfig.ExecutionMode)
			}
			
			logging.Info("Skills directory: %s", skillsDir)
			
			// Initialize skill service (discovers skills)
			if err := skillService.Initialize(skillsDir, execMode); err != nil {
				return fmt.Errorf("failed to initialize skills service: %w", err)
			}
			
			// Get list of discovered skills
			discoveredSkills := skillService.ListSkills()
			logging.Info("Discovered %d skills from %s", len(discoveredSkills), skillsDir)
			
			// Override with command-line flag if provided
			if skillNames != "" {
				// Parse comma-separated skill names
				requestedSkills := strings.Split(skillNames, ",")
				for i := range requestedSkills {
					requestedSkills[i] = strings.TrimSpace(requestedSkills[i])
				}
				
				// Create temporary SkillsConfig to override
				if runasConfig.SkillsConfig == nil {
					runasConfig.SkillsConfig = &runas.SkillsConfig{}
				}
				runasConfig.SkillsConfig.IncludeSkills = requestedSkills
				runasConfig.SkillsConfig.ExcludeSkills = nil // Clear excludes when using explicit include
				
				logging.Info("Using skills from command-line flag: %v", requestedSkills)
			}
			
			// Filter skills based on include/exclude lists
			var filteredSkills []string
			for _, skillName := range discoveredSkills {
				if runasConfig.ShouldIncludeSkill(skillName) {
					filteredSkills = append(filteredSkills, skillName)
				} else {
					logging.Info("Excluding skill: %s", skillName)
				}
			}
			
			logging.Info("Exposing %d skills as MCP tools", len(filteredSkills))
			
			// Generate MCP tools from skills
			// For each skill, create a tool with load_skill template
			runasConfig.Tools = make([]runas.ToolExposure, 0, len(filteredSkills)+1)
			
			for _, skillName := range filteredSkills {
				skill, exists := skillService.GetSkill(skillName)
				if !exists {
					continue
				}
				
				// Create tool for this skill
				tool := runas.ToolExposure{
					Name:        skill.GetMCPToolName(),
					Description: skill.GetToolDescription(),
					Template:    "load_skill", // Special marker for skill loading
					InputSchema: skill.GetMCPInputSchema(),
					InputMapping: map[string]string{
						"skill_name": skillName,
					},
				}
				
				runasConfig.Tools = append(runasConfig.Tools, tool)
				logging.Info("Created tool '%s' for skill '%s'", tool.Name, skillName)
			}
			
			// Add execute_skill_code tool for dynamic code execution
			executeCodeTool := runas.ToolExposure{
				Name: "execute_skill_code",
				Description: "[SKILL CODE EXECUTION] Execute code with access to a skill's helper libraries. " +
					"Use this to: (1) Create documents dynamically, (2) Process files with custom logic, " +
					"(3) Use skill helper libraries (e.g., Document class from docx skill). " +
					"The code executes in a sandboxed environment with the skill's scripts/ directory " +
					"available for imports via PYTHONPATH.",
				Template: "execute_skill_code", // Special marker for code execution
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skill_name": map[string]interface{}{
							"type":        "string",
							"description": "Name of skill whose helper libraries to use (e.g., 'docx', 'pdf', 'xlsx')",
						},
						"language": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"python"},
							"description": "Programming language (currently only 'python' supported)",
							"default":     "python",
						},
						"code": map[string]interface{}{
							"type":        "string",
							"description": "Python code to execute. Can import from 'scripts' module to use skill helper libraries.",
						},
						"files": map[string]interface{}{
							"type":        "object",
							"description": "Optional files to make available in workspace (filename -> base64 content)",
						},
					},
					"required": []string{"skill_name", "code"},
				},
			}
			
			runasConfig.Tools = append(runasConfig.Tools, executeCodeTool)
			
			logging.Info("Generated %d MCP tools from skills (including execute_skill_code)", len(runasConfig.Tools))
		}
		
		// Validate templates exist (skip for special skill templates)
		for i, tool := range runasConfig.Tools {
			// Skip validation for special skill-related templates
			if tool.Template == "load_skill" || tool.Template == "execute_skill_code" {
				continue
			}
			
			_, existsV1 := appConfig.Workflows[tool.Template]
			_, existsV2 := appConfig.Workflows[tool.Template]
			
			if !existsV1 && !existsV2 {
				return fmt.Errorf("tool %d (%s) references unknown template: %s", 
					i, tool.Name, tool.Template)
			}
		}
		
		// Check runas type and start appropriate server
		if runasConfig.RunAsType == runas.RunAsTypeProxy || runasConfig.RunAsType == runas.RunAsTypeProxySkills {
			// Start HTTP proxy server
			return startProxyServer(runasConfig, appConfig, configService, skillService)
		}
		
		// Default: Start stdio MCP server
		return startStdioServer(runasConfig, appConfig, configService, skillService)
	},
}

// startProxyServer starts an HTTP proxy server
func startProxyServer(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig, configService *infraConfig.Service, skillService *skillsvc.Service) error {
	logging.Info("Starting HTTP proxy server on port %d", runasConfig.ProxyConfig.Port)
	
	// Create proxy server
	proxyServer := proxy.NewProxyServer(runasConfig, appConfig)
	
	// Start proxy server (blocks until shutdown)
	if err := proxyServer.Start(); err != nil {
		return fmt.Errorf("proxy server error: %w", err)
	}
	
	return nil
}

// startStdioServer starts a stdio MCP server
func startStdioServer(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig, configService *infraConfig.Service, skillService *skillsvc.Service) error {
	logging.Info("Starting MCP server in stdio mode")
	
	// Create server service
	service := serverService.NewService(runasConfig, appConfig, configService, skillService)
	
	// Create stdio server
	stdioServer := server.NewStdioServer(service)
	
	// Wire up progress notifier so service can send progress updates
	service.SetProgressNotifier(stdioServer)
	
	// Start server
	logging.Info("MCP server starting...")
	if err := stdioServer.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	
	return nil
}

func init() {
	ServeCmd.Flags().StringVar(&serveConfig, "serve", "", "Path to runas config file")
	RootCmd.AddCommand(ServeCmd)
}
