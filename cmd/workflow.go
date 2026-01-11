package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/services/embeddings"
	workflow "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// resolveLogLevel determines the effective log level from CLI flags and workflow config
// Priority: 1) --log-level flag, 2) --verbose flag, 3) workflow config, 4) default
func resolveLogLevel(workflowConfigLevel string) string {
	// Highest priority: --log-level flag
	if logLevel != "" {
		return logLevel
	}
	
	// Second priority: --verbose flag (convert to "verbose")
	if verbose {
		return "verbose"
	}
	
	// Third priority: workflow configuration
	if workflowConfigLevel != "" {
		return workflowConfigLevel
	}
	
	// Default: info
	return "info"
}

// executeWorkflow executes a workflow by name using the new v2.0 system
func executeWorkflow() error {
	logging.Info("Executing workflow: %s", workflowName)
	
	// 1. Load configuration
	configService := infraConfig.NewService()
	appConfig, exampleCreated, err := configService.LoadConfigOrCreateExample(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// If we created an example config, inform the user
	if exampleCreated {
		fmt.Printf("📋 Created example configuration file: %s\n", configFile)
		fmt.Println("🔧 Please edit the file to:")
		fmt.Println("   1. Replace placeholder API keys with your actual keys")
		fmt.Println("   2. Add workflow files to config/workflows/")
		fmt.Println("   3. Run the command again")
		fmt.Println()
		fmt.Printf("💡 Try: mcp-cli --list-workflows\n")
		return nil
	}
	
	// 2. Get workflow
	wf, exists := appConfig.GetWorkflow(workflowName)
	if !exists {
		available := appConfig.ListWorkflows()
		if len(available) == 0 {
			return fmt.Errorf("no workflows configured. Add YAML files to config/workflows/")
		}
		return fmt.Errorf("workflow '%s' not found. Available workflows: %v", workflowName, available)
	}
	
	// 2.5. Validate workflow structure BEFORE execution
	if err := workflow.ValidateWorkflow(wf); err != nil {
		return fmt.Errorf("workflow validation failed:\n%w", err)
	}
	logging.Debug("Workflow validation passed")
	
	// 3. Get input data
	inputData, err := getInputData()
	if err != nil {
		return fmt.Errorf("failed to get input data: %w", err)
	}
	
	// 4. Collect servers needed from workflow steps
	servers := collectServersFromWorkflow(wf, appConfig)
	
	// 5. Collect skills needed from workflow steps
	skills := collectSkillsFromWorkflow(wf)
	
	// Override with command-line flag if provided
	if skillNames != "" {
		skills = strings.Split(skillNames, ",")
		// Trim whitespace from each skill name
		for i := range skills {
			skills[i] = strings.TrimSpace(skills[i])
		}
		logging.Info("Using skills from command-line flag: %v", skills)
	}
	
	// 6. Execute workflow (with or without servers)
	if len(servers) == 0 {
		return executeWorkflowWithoutServers(wf, workflowName, inputData, appConfig, skills)
	}
	return executeWorkflowWithServers(wf, workflowName, inputData, appConfig, servers, skills)
}

// initializeProvider creates the LLM provider for the workflow
func initializeProvider(appConfig *config.ApplicationConfig, configService *infraConfig.Service, wf *config.WorkflowV2) (domain.LLMProvider, error) {
	// Determine provider to use
	var actualProviderName string
	
	// Check command-line overrides first
	if providerName != "" {
		actualProviderName = providerName
	} else if wf.Execution.Provider != "" {
		actualProviderName = wf.Execution.Provider
	} else if len(wf.Execution.Providers) > 0 {
		// Use first provider from chain
		actualProviderName = wf.Execution.Providers[0].Provider
	}
	
	// Get provider config
	var providerConfig *config.ProviderConfig
	var interfaceType config.InterfaceType
	var err error
	
	if actualProviderName == "" {
		// Use default provider
		actualProviderName, providerConfig, interfaceType, err = configService.GetDefaultProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get default provider: %w", err)
		}
	} else {
		// Use specified provider
		providerConfig, interfaceType, err = configService.GetProviderConfig(actualProviderName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider config for %s: %w", actualProviderName, err)
		}
	}
	
	// Override model if specified
	if modelName != "" {
		providerConfig.DefaultModel = modelName
	} else if wf.Execution.Model != "" {
		providerConfig.DefaultModel = wf.Execution.Model
	} else if len(wf.Execution.Providers) > 0 && wf.Execution.Providers[0].Model != "" {
		providerConfig.DefaultModel = wf.Execution.Providers[0].Model
	}
	
	// Create provider
	providerType := domain.ProviderType(actualProviderName)
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	
	return provider, nil
}

// getInputData retrieves input from flag or stdin
func getInputData() (string, error) {
	if inputData != "" {
		return inputData, nil
	}
	
	// Check stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}
	
	return "", nil // Empty input is OK
}

// collectServersFromWorkflow extracts all unique server names from workflow steps
func collectServersFromWorkflow(wf *config.WorkflowV2, appConfig *config.ApplicationConfig) []string {
	serverSet := make(map[string]bool)
	
	// Get RAG config from already-loaded app config
	var ragConfig *config.RagConfig
	if appConfig != nil && appConfig.RAG != nil {
		ragConfig = appConfig.RAG
	}
	
	// Collect from execution level
	for _, server := range wf.Execution.Servers {
		serverSet[server] = true
	}
	
	// Collect from steps
	for _, step := range wf.Steps {
		// Regular server references
		for _, server := range step.Servers {
			serverSet[server] = true
		}
		
		// RAG step servers - need to resolve to actual MCP server names
		if step.Rag != nil && ragConfig != nil {
			// Helper function to resolve RAG server name to MCP server name
			resolveRagServer := func(ragServerName string) string {
				if ragServerConfig, exists := ragConfig.Servers[ragServerName]; exists {
					logging.Debug("Resolved RAG server '%s' to MCP server '%s'", ragServerName, ragServerConfig.MCPServer)
					return ragServerConfig.MCPServer
				}
				// If not found in RAG config, assume it's an MCP server name directly
				logging.Debug("RAG server '%s' not in config, using as-is", ragServerName)
				return ragServerName
			}
			
			if step.Rag.Server != "" {
				// Single server mode
				mcpServer := resolveRagServer(step.Rag.Server)
				if mcpServer != "" {
					serverSet[mcpServer] = true
				}
			}
			// Multi-server mode
			for _, ragServer := range step.Rag.Servers {
				mcpServer := resolveRagServer(ragServer)
				if mcpServer != "" {
					serverSet[mcpServer] = true
				}
			}
		}
	}
	
	// Convert to slice
	servers := make([]string, 0, len(serverSet))
	for server := range serverSet {
		servers = append(servers, server)
	}
	
	return servers
}

// collectSkillsFromWorkflow extracts all unique skill names from workflow steps
func collectSkillsFromWorkflow(wf *config.WorkflowV2) []string {
	skillSet := make(map[string]bool)
	
	// Collect from execution level
	for _, skill := range wf.Execution.Skills {
		skillSet[skill] = true
	}
	
	// Collect from steps
	for _, step := range wf.Steps {
		for _, skill := range step.Skills {
			skillSet[skill] = true
		}
	}
	
	// Convert to slice
	skills := make([]string, 0, len(skillSet))
	for skill := range skillSet {
		skills = append(skills, skill)
	}
	
	return skills
}

// executeWorkflowWithoutServers executes a workflow that doesn't need MCP servers
func executeWorkflowWithoutServers(wf *config.WorkflowV2, workflowKey string, inputData string, appConfig *config.ApplicationConfig, skills []string) error {
	logging.Debug("Executing workflow without MCP servers")
	
	// Note: Skills are typically exposed through MCP servers, so this path wouldn't use skills
	// But we keep the parameter for consistency
	if len(skills) > 0 {
		logging.Debug("Skills specified but not used (no MCP server mode): %v", skills)
	}
	
	// Create services
	configService := infraConfig.NewService()
	
	// Load AI provider configurations (needed for embedding service)
	// Always load config.yaml so configService has provider configs
	_, err := configService.LoadConfig("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load AI provider config: %w", err)
	}
	
	providerFactory := ai.NewProviderFactory()
	embeddingService := embeddings.NewService(configService, providerFactory)
	
	// Create logger with resolved log level
	effectiveLogLevel := resolveLogLevel(wf.Execution.Logging)
	logger := workflow.NewLogger(effectiveLogLevel, false) // verbose handled by resolveLogLevel
	
	// Create orchestrator with workflow key for directory-aware resolution
	orchestrator := workflow.NewOrchestratorWithKey(wf, workflowKey, logger)
	
	// Set provider on executor
	orchestrator.SetAppConfig(appConfig)
	orchestrator.SetAppConfigForWorkflows(appConfig)
	orchestrator.SetEmbeddingService(embeddingService)
	
	// Execute
	ctx := context.Background()
	if err := orchestrator.Execute(ctx, inputData); err != nil {
		return handleWorkflowError(wf.Name, err)
	}
	
	// Output results
	return outputWorkflowResults(orchestrator, wf)
}

// executeWorkflowWithServers executes a workflow that needs MCP servers
func executeWorkflowWithServers(wf *config.WorkflowV2, workflowKey string, inputData string, appConfig *config.ApplicationConfig, servers []string, skills []string) error {
	logging.Debug("Executing workflow with MCP servers: %v", servers)
	if len(skills) > 0 {
		logging.Info("Skills filter enabled: %v", skills)
	}
	
	// Mark all servers as user-specified
	userSpecified := make(map[string]bool)
	for _, server := range servers {
		userSpecified[server] = true
	}
	
	// Execute with server connections
	var execErr error
	err := host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
		// Create config service and load AI provider configurations
		// This is needed for the embedding service to work
		configService := infraConfig.NewService()
		_, err := configService.LoadConfig("config.yaml")
		if err != nil {
			return fmt.Errorf("failed to load AI provider config: %w", err)
		}
		
		providerFactory := ai.NewProviderFactory()
		embeddingService := embeddings.NewService(configService, providerFactory)
		
		// Create server manager
		serverManager := NewHostServerManager(conns)
		
		// Create logger with resolved log level
		effectiveLogLevel := resolveLogLevel(wf.Execution.Logging)
		logger := workflow.NewLogger(effectiveLogLevel, false) // verbose handled by resolveLogLevel
		
		// Create orchestrator with workflow key for directory-aware resolution
		orchestrator := workflow.NewOrchestratorWithKey(wf, workflowKey, logger)
		
		// Set provider and server manager
		orchestrator.SetAppConfig(appConfig)
		orchestrator.SetAppConfigForWorkflows(appConfig)
		orchestrator.SetServerManager(serverManager)
		orchestrator.SetEmbeddingService(embeddingService)
		
		// Execute
		ctx := context.Background()
		if err := orchestrator.Execute(ctx, inputData); err != nil {
			execErr = handleWorkflowError(wf.Name, err)
			return execErr
		}
		
		// Output results
		execErr = outputWorkflowResults(orchestrator, wf)
		return execErr
	}, configFile, servers, userSpecified, host.QuietCommandOptions())
	
	if err != nil {
		return err
	}
	return execErr
}

// outputWorkflowResults outputs the final results from orchestrator
func outputWorkflowResults(orchestrator *workflow.Orchestrator, wf *config.WorkflowV2) error {
	// Get final step result
	if len(wf.Steps) == 0 {
		fmt.Println("Workflow completed (no steps)")
		return nil
	}
	
	lastStepName := wf.Steps[len(wf.Steps)-1].Name
	finalResult, ok := orchestrator.GetStepResult(lastStepName)
	
	if !ok {
		fmt.Printf("Workflow '%s' completed but produced no output\n", wf.Name)
		return nil
	}
	
	// Clean output
	fmt.Println(strings.TrimSpace(finalResult))
	
	return nil
}

// handleWorkflowError formats workflow execution errors
func handleWorkflowError(workflowName string, err error) error {
	errorResponse := map[string]interface{}{
		"workflow":  workflowName,
		"status":    "failed",
		"timestamp": time.Now().Format(time.RFC3339),
		"error":     err.Error(),
	}
	
	output, _ := json.MarshalIndent(errorResponse, "", "  ")
	fmt.Fprintln(os.Stderr, string(output))
	
	return err
}

// executeListWorkflows lists all available workflows
func executeListWorkflows() error {
	// Load configuration
	configService := infraConfig.NewService()
	appConfig, exampleCreated, err := configService.LoadConfigOrCreateExample(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// If we created an example config, inform the user
	if exampleCreated {
		fmt.Printf("📋 Created example configuration file: %s\n", configFile)
		fmt.Println("🔧 Please edit the file and add workflow files to config/workflows/")
		fmt.Println()
		fmt.Println("📝 Example workflow structure:")
		fmt.Println("   config/workflows/my_workflow.yaml")
		return nil
	}
	
	// Get available workflows
	workflows := appConfig.ListWorkflows()
	
	if len(workflows) == 0 {
		fmt.Println("No workflows configured.")
		fmt.Println("\nTo add workflows:")
		fmt.Println("  1. Create YAML files in config/workflows/")
		fmt.Println("  2. Use schema: workflow/v2.0")
		fmt.Println("  3. See examples in config/workflows/")
		return nil
	}
	
	// Create workflow list response
	workflowList := map[string]interface{}{
		"workflows": workflows,
		"count":     len(workflows),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Add workflow details if verbose mode
	if verbose {
		workflowDetails := make(map[string]interface{})
		for _, name := range workflows {
			if wf, exists := appConfig.GetWorkflow(name); exists {
				details := map[string]interface{}{
					"version":     wf.Version,
					"description": wf.Description,
					"steps":       len(wf.Steps),
				}
				
				// Add execution info
				execInfo := make(map[string]interface{})
				if len(wf.Execution.Providers) > 0 {
					execInfo["providers"] = wf.Execution.Providers
				} else if wf.Execution.Provider != "" {
					execInfo["provider"] = wf.Execution.Provider
				}
				if wf.Execution.Temperature != 0 {
					execInfo["temperature"] = wf.Execution.Temperature
				}
				if len(execInfo) > 0 {
					details["execution"] = execInfo
				}
				
				workflowDetails[name] = details
			}
		}
		workflowList["details"] = workflowDetails
	}
	
	// Output as JSON
	output, err := json.MarshalIndent(workflowList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow list: %w", err)
	}
	
	fmt.Println(string(output))
	
	// Show usage examples
	if !verbose {
		fmt.Println()
		fmt.Println("💡 Usage examples:")
		if len(workflows) > 0 {
			fmt.Printf("   mcp-cli --workflow %s --input-data \"your data\"\n", workflows[0])
			fmt.Printf("   echo \"data\" | mcp-cli --workflow %s\n", workflows[0])
		}
		fmt.Println()
		fmt.Println("📖 For detailed info: mcp-cli --list-workflows --verbose")
	}
	
	return nil
}

// HostServerManager adapts host.ServerConnection to domain.MCPServerManager interface
type HostServerManager struct {
	connections []*host.ServerConnection
}

func NewHostServerManager(connections []*host.ServerConnection) *HostServerManager {
	return &HostServerManager{connections: connections}
}

func (hsm *HostServerManager) StartServer(ctx context.Context, serverName string, cfg *config.ServerConfig) (domain.MCPServer, error) {
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, nil
		}
	}
	return nil, fmt.Errorf("server '%s' not found in host connections", serverName)
}

func (hsm *HostServerManager) StopServer(serverName string) error {
	return nil
}

func (hsm *HostServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, true
		}
	}
	return nil, false
}

func (hsm *HostServerManager) ListServers() map[string]domain.MCPServer {
	servers := make(map[string]domain.MCPServer)
	for _, conn := range hsm.connections {
		servers[conn.Name] = &HostServerAdapter{connection: conn}
	}
	return servers
}

func (hsm *HostServerManager) GetAvailableTools() ([]domain.Tool, error) {
	var toolsList []domain.Tool
	
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		serverTools, err := adapter.GetTools()
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
			continue
		}
		toolsList = append(toolsList, serverTools...)
	}
	
	return toolsList, nil
}

func (hsm *HostServerManager) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		toolsList, err := adapter.GetTools()
		if err != nil {
			continue
		}
		
		// Check both prefixed and unprefixed tool names
		serverPrefix := conn.Name + "_"
		serverPrefixUnderscore := strings.ReplaceAll(conn.Name, "-", "_") + "_"
		
		for _, tool := range toolsList {
			// Extract original tool name (strip server prefix if present)
			originalName := tool.Function.Name
			if strings.HasPrefix(originalName, serverPrefix) {
				originalName = strings.TrimPrefix(originalName, serverPrefix)
			} else if strings.HasPrefix(originalName, serverPrefixUnderscore) {
				originalName = strings.TrimPrefix(originalName, serverPrefixUnderscore)
			}
			
			// Match against both original name and prefixed name
			if tool.Function.Name == toolName || originalName == toolName {
				return adapter.ExecuteTool(ctx, toolName, arguments)
			}
		}
	}
	
	return "", fmt.Errorf("tool '%s' not found on any server", toolName)
}

func (hsm *HostServerManager) StopAll() error {
	return nil
}

// HostServerAdapter adapts host.ServerConnection to domain.MCPServer interface
type HostServerAdapter struct {
	connection  *host.ServerConnection
	toolsCache  []domain.Tool
	toolsCached bool
}

func (hsa *HostServerAdapter) Start(ctx context.Context) error {
	return nil
}

func (hsa *HostServerAdapter) Stop() error {
	return nil
}

func (hsa *HostServerAdapter) IsRunning() bool {
	return hsa.connection.Client != nil
}

func formatToolNameForOpenAI(serverName, toolName string) string {
	serverName = strings.ReplaceAll(serverName, ".", "_")
	serverName = strings.ReplaceAll(serverName, " ", "_")
	serverName = strings.ReplaceAll(serverName, "-", "_")
	
	toolName = strings.ReplaceAll(toolName, ".", "_")
	toolName = strings.ReplaceAll(toolName, " ", "_")
	
	return fmt.Sprintf("%s_%s", serverName, toolName)
}

func (hsa *HostServerAdapter) GetTools() ([]domain.Tool, error) {
	if hsa.toolsCached {
		return hsa.toolsCache, nil
	}

	result, err := tools.SendToolsList(hsa.connection.Client, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP server %s: %w", hsa.connection.Name, err)
	}

	var domainTools []domain.Tool
	for _, tool := range result.Tools {
		formattedName := formatToolNameForOpenAI(hsa.connection.Name, tool.Name)
		
		domainTool := domain.Tool{
			Type: "function",
			Function: domain.ToolFunction{
				Name:        formattedName,
				Description: fmt.Sprintf("[%s] %s", hsa.connection.Name, tool.Description),
				Parameters:  tool.InputSchema,
			},
		}
		domainTools = append(domainTools, domainTool)
	}

	hsa.toolsCache = domainTools
	hsa.toolsCached = true

	logging.Debug("Successfully got %d tools from server %s", len(domainTools), hsa.connection.Name)
	return domainTools, nil
}

func (hsa *HostServerAdapter) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	actualToolName := toolName
	serverPrefix := hsa.connection.Name + "_"
	serverPrefixUnderscore := strings.ReplaceAll(hsa.connection.Name, "-", "_") + "_"
	
	if strings.HasPrefix(toolName, serverPrefix) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefix)
	} else if strings.HasPrefix(toolName, serverPrefixUnderscore) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefixUnderscore)
	}

	logging.Debug("Executing tool %s (actual: %s) on server %s", toolName, actualToolName, hsa.connection.Name)

	result, err := tools.SendToolsCall(hsa.connection.Client, actualToolName, arguments)
	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed for %s: %w", actualToolName, err)
	}

	if result.IsError {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}

	// Extract text from content blocks
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		// Direct string response
		resultStr = content
	case []interface{}:
		// Content blocks array (standard MCP format)
		// Extract text from the first text-type content block
		for _, item := range content {
			if block, ok := item.(map[string]interface{}); ok {
				if blockType, hasType := block["type"].(string); hasType && blockType == "text" {
					if text, hasText := block["text"].(string); hasText {
						resultStr = text
						break
					}
				}
			}
		}
		if resultStr == "" {
			// No text content found, marshal the whole thing as fallback
			resultBytes, err := json.Marshal(content)
			if err != nil {
				return "", fmt.Errorf("failed to marshal tool result: %w", err)
			}
			resultStr = string(resultBytes)
		}
	default:
		// Unknown format, marshal it
		resultBytes, err := json.Marshal(content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tool result: %w", err)
		}
		resultStr = string(resultBytes)
	}

	logging.Debug("Tool %s executed successfully on server %s", actualToolName, hsa.connection.Name)
	return resultStr, nil
}

func (hsa *HostServerAdapter) GetServerName() string {
	return hsa.connection.Name
}

func (hsa *HostServerAdapter) GetConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Command: "mock",
		Args:    []string{},
	}
}
