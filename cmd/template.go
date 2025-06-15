package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	workflowservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// executeTemplate executes a workflow template
func executeTemplate() error {
	logging.Info("Executing workflow template: %s", templateName)
	
	// 1. Load configuration using auto-generation if needed
	configService := config.NewService()
	appConfig, exampleCreated, err := configService.LoadConfigOrCreateExample(configFile)
	if err != nil {
		return handleTemplateError(fmt.Errorf("failed to load configuration: %w", err))
	}
	
	// If we created an example config, inform the user
	if exampleCreated {
		fmt.Printf("📋 Created example configuration file: %s\n", configFile)
		fmt.Println("🔧 Please edit the file to:")
		fmt.Println("   1. Replace '# your-api-key-here' with your actual API keys")
		fmt.Println("   2. For Windows users: Download MCP server executables from:")
		fmt.Println("      https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers")
		fmt.Println("   3. Update server paths to point to your downloaded .exe files")
		fmt.Println("   4. Remove comment fields when ready")
		fmt.Println("   5. Run the command again")
		fmt.Println()
		fmt.Printf("💡 Try running: mcp-cli --list-templates\n")
		return nil
	}
	
	// 2. Validate template exists and get template configuration
	template, exists := appConfig.GetWorkflowTemplate(templateName)
	if !exists {
		availableTemplates := appConfig.ListWorkflowTemplates()
		if len(availableTemplates) == 0 {
			return handleTemplateError(fmt.Errorf("no workflow templates are configured"))
		}
		return handleTemplateError(fmt.Errorf("workflow template '%s' not found. Available templates: %v", templateName, availableTemplates))
	}
	
	// 3. Determine servers to use from template configuration
	var serverNames []string
	var userSpecified map[string]bool
	
	// Use servers specified in the template steps, or command-line override
	if serverName != "" {
		// Command-line server override
		serverNames, userSpecified = host.ProcessOptions(serverName, disableFilesystem, providerName, modelName)
		logging.Debug("Using command-line server override: %v", serverNames)
	} else {
		// Extract servers from template steps
		serverSet := make(map[string]bool)
		for _, step := range template.Steps {
			for _, server := range step.Servers {
				serverSet[server] = true
			}
		}
		
		if len(serverSet) == 0 {
			// No servers specified in template, use all available
			serverNames, userSpecified = host.ProcessOptions("", disableFilesystem, providerName, modelName)
			logging.Debug("No servers specified in template, using all available: %v", serverNames)
		} else {
			// Use servers from template
			serverNames = make([]string, 0, len(serverSet))
			userSpecified = make(map[string]bool)
			for server := range serverSet {
				serverNames = append(serverNames, server)
				userSpecified[server] = true
			}
			logging.Info("Using servers from template configuration: %v", serverNames)
		}
	}
	
	if len(serverNames) == 0 {
		return handleTemplateError(fmt.Errorf("no servers available for workflow execution"))
	}
	
	// 4. Initialize AI provider service
	aiService := ai.NewService()
	provider, err := aiService.InitializeProvider(configFile, providerName, modelName)
	if err != nil {
		return handleTemplateError(fmt.Errorf("failed to initialize AI provider: %w", err))
	}
	defer provider.Close()
	
	// 5. Prepare input data
	var processedInputData string
	if inputData != "" {
		processedInputData = inputData
	} else {
		// Check if there's stdin data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data available on stdin
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return handleTemplateError(fmt.Errorf("failed to read stdin: %w", err))
			}
			processedInputData = string(stdinData)
		}
	}
	
	// 6. Execute workflow using the same pattern as query mode
	var workflowResponse *domain.WorkflowResponse
	err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
		// Create enhanced server manager that wraps the host connections
		serverManager := NewHostServerManager(conns)
		
		// Create workflow service
		workflowService, err := workflowservice.NewService(appConfig, configService, provider, serverManager)
		if err != nil {
			return fmt.Errorf("failed to create workflow service: %w", err)
		}
		
		// Execute workflow
		ctx := context.Background()
		workflowResponse, err = workflowService.ExecuteWorkflow(ctx, templateName, processedInputData)
		if err != nil {
			return fmt.Errorf("workflow execution failed: %w", err)
		}
		
		return nil
	}, configFile, serverNames, userSpecified, host.QuietCommandOptions())
	
	if err != nil {
		return handleTemplateError(err)
	}
	
	// 7. Output response
	return outputTemplateResponse(workflowResponse)
}

// executeListTemplates lists all available workflow templates
func executeListTemplates() error {
	// Load configuration using auto-generation if needed
	configService := config.NewService()
	appConfig, exampleCreated, err := configService.LoadConfigOrCreateExample(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// If we created an example config, inform the user and show example templates
	if exampleCreated {
		fmt.Printf("📋 Created example configuration file: %s\n", configFile)
		fmt.Println("🔧 Please edit the file to:")
		fmt.Println("   1. Replace '# your-api-key-here' with your actual API keys")
		fmt.Println("   2. For Windows users: Download MCP server executables from:")
		fmt.Println("      https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers")
		fmt.Println("   3. Update server paths to point to your downloaded .exe files")
		fmt.Println("   4. Remove comment fields when ready")
		fmt.Println()
		fmt.Println("📝 Example templates included:")
	}
	
	// Get available templates
	templates := appConfig.ListWorkflowTemplates()
	
	if len(templates) == 0 {
		fmt.Println("No workflow templates are configured.")
		fmt.Println("\nTo add templates, add a 'templates' section to your configuration file.")
		return nil
	}
	
	// Create template list response
	templateList := map[string]interface{}{
		"available_templates": templates,
		"total_count":        len(templates),
		"timestamp":          time.Now(),
	}
	
	// Add template details if verbose mode is enabled
	if verbose {
		templateDetails := make(map[string]interface{})
		for _, templateName := range templates {
			if workflowTemplate, exists := appConfig.GetWorkflowTemplate(templateName); exists {
				templateDetails[templateName] = map[string]interface{}{
					"description": workflowTemplate.Description,
					"steps":       len(workflowTemplate.Steps),
					"variables":   len(workflowTemplate.Variables),
				}
			}
		}
		templateList["template_details"] = templateDetails
	}
	
	// Output as JSON
	output, err := json.MarshalIndent(templateList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template list: %w", err)
	}
	
	fmt.Println(string(output))
	
	// If example was created, show usage instructions
	if exampleCreated {
		fmt.Println()
		fmt.Println("🚀 Usage examples:")
		fmt.Println("   mcp-cli --template analyze_file")
		fmt.Println("   mcp-cli --template search_and_summarize")
		fmt.Println("   echo 'some data' | mcp-cli --template simple_analyze")
		fmt.Println()
		fmt.Println("💡 Remember to configure your API keys and server paths first!")
		fmt.Println("📦 Download Golang MCP servers from: https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers")
	}
	
	return nil
}

// outputTemplateResponse outputs the workflow response like query mode (clean by default)
func outputTemplateResponse(response *domain.WorkflowResponse) error {
	// Check if --json flag was used (if we need to support it later)
	// For now, default to clean text output like query mode
	
	// If there's an error, output structured error for debugging
	if response.Error != nil {
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal error response: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}
	
	// For successful execution, output clean text like query mode
	if response.FinalOutput != "" {
		// Just output the final LLM response, clean and simple
		fmt.Println(response.FinalOutput)
	} else if len(response.StepResults) > 0 {
		// Fallback to last step output if no final output
		lastStep := response.StepResults[len(response.StepResults)-1]
		if lastStep.Output != "" {
			fmt.Println(lastStep.Output)
		} else {
			// If still no output, something went wrong - show minimal info
			fmt.Printf("Template '%s' completed but produced no output\n", response.TemplateName)
		}
	} else {
		fmt.Printf("Template '%s' completed but produced no output\n", response.TemplateName)
	}
	
	return nil
}

// handleTemplateError handles template execution errors with structured output
func handleTemplateError(err error) error {
	// Create structured error response for template mode
	errorResponse := &domain.WorkflowResponse{
		ExecutionID:  "error",
		TemplateName: templateName,
		Status:       domain.WorkflowStatusFailed,
		Timestamp:    time.Now(),
		Error: &domain.WorkflowError{
			Code:    "TEMPLATE_EXECUTION_ERROR",
			Message: err.Error(),
		},
		Metadata: map[string]interface{}{
			"error_occurred": true,
			"command_line":   os.Args,
		},
	}
	
	// Output structured error
	if outputErr := outputTemplateResponse(errorResponse); outputErr != nil {
		// Fallback to plain text error if JSON marshaling fails
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	
	return err
}

// HostServerManager adapts host.ServerConnection to domain.MCPServerManager interface
type HostServerManager struct {
	connections []*host.ServerConnection
}

// NewHostServerManager creates a new host server manager
func NewHostServerManager(connections []*host.ServerConnection) *HostServerManager {
	return &HostServerManager{
		connections: connections,
	}
}

// StartServer starts an MCP server (not applicable for host connections)
func (hsm *HostServerManager) StartServer(ctx context.Context, serverName string, config *domain.ServerConfig) (domain.MCPServer, error) {
	// Find existing connection
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, nil
		}
	}
	return nil, fmt.Errorf("server '%s' not found in host connections", serverName)
}

// StopServer stops an MCP server (not applicable for host connections)
func (hsm *HostServerManager) StopServer(serverName string) error {
	// Host connections are managed by the host package
	return nil
}

// GetServer retrieves a running server
func (hsm *HostServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	for _, conn := range hsm.connections {
		if conn.Name == serverName {
			return &HostServerAdapter{connection: conn}, true
		}
	}
	return nil, false
}

// ListServers returns all running servers
func (hsm *HostServerManager) ListServers() map[string]domain.MCPServer {
	servers := make(map[string]domain.MCPServer)
	for _, conn := range hsm.connections {
		servers[conn.Name] = &HostServerAdapter{connection: conn}
	}
	return servers
}

// GetAvailableTools returns all available tools from all servers
func (hsm *HostServerManager) GetAvailableTools() ([]domain.Tool, error) {
	var tools []domain.Tool
	
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		serverTools, err := adapter.GetTools()
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
			continue
		}
		tools = append(tools, serverTools...)
	}
	
	return tools, nil
}

// ExecuteTool executes a tool on the appropriate server
func (hsm *HostServerManager) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// Find the server that has this tool
	for _, conn := range hsm.connections {
		adapter := &HostServerAdapter{connection: conn}
		tools, err := adapter.GetTools()
		if err != nil {
			continue
		}
		
		for _, tool := range tools {
			if tool.Function.Name == toolName {
				return adapter.ExecuteTool(ctx, toolName, arguments)
			}
		}
	}
	
	return "", fmt.Errorf("tool '%s' not found on any server", toolName)
}

// StopAll stops all running servers (not applicable for host connections)
func (hsm *HostServerManager) StopAll() error {
	// Host connections are managed by the host package
	return nil
}

// HostServerAdapter adapts host.ServerConnection to domain.MCPServer interface
type HostServerAdapter struct {
	connection  *host.ServerConnection
	toolsCache  []domain.Tool // Cache tools to avoid repeated calls
	toolsCached bool
}

// Start starts the MCP server (already started by host)
func (hsa *HostServerAdapter) Start(ctx context.Context) error {
	return nil
}

// Stop stops the MCP server (managed by host)
func (hsa *HostServerAdapter) Stop() error {
	return nil
}

// IsRunning returns true if the server is running
func (hsa *HostServerAdapter) IsRunning() bool {
	return hsa.connection.Client != nil
}

// formatToolNameForOpenAI formats the tool name to be compatible with OpenAI's requirements
// (copied from query mode for consistency)
func formatToolNameForOpenAI(serverName, toolName string) string {
	serverName = strings.ReplaceAll(serverName, ".", "_")
	serverName = strings.ReplaceAll(serverName, " ", "_")
	serverName = strings.ReplaceAll(serverName, "-", "_")
	
	toolName = strings.ReplaceAll(toolName, ".", "_")
	toolName = strings.ReplaceAll(toolName, " ", "_")
	
	return fmt.Sprintf("%s_%s", serverName, toolName)
}

// GetTools returns available tools from this server using real MCP protocol
func (hsa *HostServerAdapter) GetTools() ([]domain.Tool, error) {
	// Use cache if available
	if hsa.toolsCached {
		return hsa.toolsCache, nil
	}

	// Get tools using the same method as query mode
	result, err := tools.SendToolsList(hsa.connection.Client, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP server %s: %w", hsa.connection.Name, err)
	}

	// Convert MCP tools to domain tools (same format as query mode)
	var domainTools []domain.Tool
	for _, tool := range result.Tools {
		// Format tool name for consistency (same as query mode)
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

	// Cache the results
	hsa.toolsCache = domainTools
	hsa.toolsCached = true

	logging.Debug("Successfully got %d tools from server %s", len(domainTools), hsa.connection.Name)
	return domainTools, nil
}

// ExecuteTool executes a tool on this server using real MCP protocol
func (hsa *HostServerAdapter) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// Parse the tool name to remove server prefix (same logic as query mode)
	actualToolName := toolName
	serverPrefix := hsa.connection.Name + "_"
	serverPrefixUnderscore := strings.ReplaceAll(hsa.connection.Name, "-", "_") + "_"
	
	if strings.HasPrefix(toolName, serverPrefix) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefix)
	} else if strings.HasPrefix(toolName, serverPrefixUnderscore) {
		actualToolName = strings.TrimPrefix(toolName, serverPrefixUnderscore)
	}

	logging.Debug("Executing tool %s (actual: %s) on server %s", toolName, actualToolName, hsa.connection.Name)

	// Execute the tool using real MCP protocol (same as query mode)
	result, err := tools.SendToolsCall(hsa.connection.Client, actualToolName, arguments)
	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed for %s: %w", actualToolName, err)
	}

	if result.IsError {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}

	// Convert result to string (same logic as query mode)
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		resultStr = content
	default:
		resultBytes, err := json.Marshal(content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tool result: %w", err)
		}
		resultStr = string(resultBytes)
	}

	logging.Debug("Tool %s executed successfully on server %s", actualToolName, hsa.connection.Name)
	return resultStr, nil
}

// GetServerName returns the name of this server
func (hsa *HostServerAdapter) GetServerName() string {
	return hsa.connection.Name
}

// GetConfig returns the server configuration
func (hsa *HostServerAdapter) GetConfig() *domain.ServerConfig {
	// Return a basic config based on the connection
	return &domain.ServerConfig{
		Command: "mock",
		Args:    []string{},
	}
}
