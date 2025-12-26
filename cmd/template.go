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
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	workflowservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
	"github.com/LaurieRhodes/mcp-cli-go/internal/template"
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
	configService := infraConfig.NewService()
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
	
	// 2. Check for template v2 first
	if templateV2, exists := appConfig.TemplatesV2[templateName]; exists {
		logging.Info("Executing template v2: %s", templateName)
		return executeTemplateV2(appConfig, configService, templateV2)
	}
	
	// 3. Fallback to old template system
	// Validate template exists and get template configuration
	workflowTemplate, exists := appConfig.GetWorkflowTemplate(templateName)
	if !exists {
		availableTemplates := appConfig.ListWorkflowTemplates()
		availableV2 := make([]string, 0, len(appConfig.TemplatesV2))
		for name := range appConfig.TemplatesV2 {
			availableV2 = append(availableV2, name)
		}
		
		if len(availableTemplates) == 0 && len(availableV2) == 0 {
			return handleTemplateError(fmt.Errorf("no workflow templates are configured"))
		}
		
		allTemplates := append(availableTemplates, availableV2...)
		return handleTemplateError(fmt.Errorf("workflow template '%s' not found. Available templates: %v", templateName, allTemplates))
	}
	
	// 3. Determine servers to use from template configuration
	var serverNames []string
	var userSpecified map[string]bool
	
	// Use servers specified in the template steps, or command-line override
	if serverName != "" {
		// Command-line server override - pass configFile
		serverNames, userSpecified = host.ProcessOptions(configFile, serverName, disableFilesystem, providerName, modelName)
		logging.Debug("Using command-line server override: %v", serverNames)
	} else {
		// Extract servers from template steps
		serverSet := make(map[string]bool)
		for _, step := range workflowTemplate.Steps {
			for _, server := range step.Servers {
				serverSet[server] = true
			}
		}
		
		if len(serverSet) == 0 {
			// No servers specified in template, don't connect to any servers
			serverNames = []string{}
			userSpecified = make(map[string]bool)
			logging.Info("No servers specified in template, workflow will run without MCP tools")
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
	
	// Servers are optional for templates that only use LLM without tools
	logging.Debug("Template will use %d servers", len(serverNames))
	
	// 4. Initialize provider using the provider factory
	actualProviderName := providerName
	var providerConfig *config.ProviderConfig
	var interfaceType config.InterfaceType
	
	if actualProviderName == "" {
		// Use default provider
		var err error
		actualProviderName, providerConfig, interfaceType, err = configService.GetDefaultProvider()
		if err != nil {
			return handleTemplateError(fmt.Errorf("failed to get default provider: %w", err))
		}
	} else {
		// Use specified provider
		var err error
		providerConfig, interfaceType, err = configService.GetProviderConfig(actualProviderName)
		if err != nil {
			return handleTemplateError(fmt.Errorf("failed to get provider config: %w", err))
		}
	}
	
	// Override model if specified
	if modelName != "" {
		providerConfig.DefaultModel = modelName
	}
	
	// Map provider name to provider type
	providerType := domain.ProviderType(actualProviderName)
	
	// Create provider using factory
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return handleTemplateError(fmt.Errorf("failed to create provider: %w", err))
	}
	
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
	
	// 6. Execute workflow
	var workflowResponse *domain.WorkflowResponse
	
	if len(serverNames) == 0 {
		// No servers needed - execute workflow directly without server connections
		logging.Debug("Executing workflow without MCP servers")
		
		// Create empty server manager
		serverManager := NewHostServerManager([]*host.ServerConnection{})
		
		// Create workflow service with domain.LLMProvider
		workflowService := workflowservice.NewService(appConfig, configService, provider, serverManager)
		
		// Execute workflow
		ctx := context.Background()
		workflowResponse, err = workflowService.ExecuteWorkflow(ctx, templateName, processedInputData)
		if err != nil {
			return handleTemplateError(fmt.Errorf("workflow execution failed: %w", err))
		}
	} else {
		// Servers needed - use RunCommandWithOptions to connect
		err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
			// Create server manager
			serverManager := NewHostServerManager(conns)
			
			// Create workflow service with domain.LLMProvider
			workflowService := workflowservice.NewService(appConfig, configService, provider, serverManager)
			
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
	}
	
	// 7. Output response
	return outputTemplateResponse(workflowResponse)
}

// executeListTemplates lists all available workflow templates
func executeListTemplates() error {
	// Load configuration using auto-generation if needed
	configService := infraConfig.NewService()
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
	
	// Get available templates (both v1 and v2)
	templatesV1 := appConfig.ListWorkflowTemplates()
	templatesV2 := make([]string, 0, len(appConfig.TemplatesV2))
	for name := range appConfig.TemplatesV2 {
		templatesV2 = append(templatesV2, name)
	}
	
	totalCount := len(templatesV1) + len(templatesV2)
	
	if totalCount == 0 {
		fmt.Println("No workflow templates are configured.")
		fmt.Println("\nTo add templates, add YAML files to config/templates/ directory.")
		return nil
	}
	
	// Create template list response
	templateList := map[string]interface{}{
		"templates_v1":        templatesV1,
		"templates_v2":        templatesV2,
		"total_count":        totalCount,
		"timestamp":          time.Now(),
	}
	
	// Add template details if verbose mode is enabled
	if verbose {
		templateDetailsV1 := make(map[string]interface{})
		for _, templateName := range templatesV1 {
			if workflowTemplate, exists := appConfig.GetWorkflowTemplate(templateName); exists {
				templateDetailsV1[templateName] = map[string]interface{}{
					"version":     "1.0 (legacy)",
					"description": workflowTemplate.Description,
					"steps":       len(workflowTemplate.Steps),
					"variables":   len(workflowTemplate.Variables),
				}
			}
		}
		
		templateDetailsV2 := make(map[string]interface{})
		for _, templateName := range templatesV2 {
			if tmpl, exists := appConfig.TemplatesV2[templateName]; exists {
				templateDetailsV2[templateName] = map[string]interface{}{
					"version":     tmpl.Version,
					"description": tmpl.Description,
					"steps":       len(tmpl.Steps),
					"category":    func() string { if tmpl.Metadata != nil { return tmpl.Metadata.Category }; return "" }(),
					"tags":        func() []string { if tmpl.Metadata != nil { return tmpl.Metadata.Tags }; return nil }(),
				}
			}
		}
		
		templateList["template_details_v1"] = templateDetailsV1
		templateList["template_details_v2"] = templateDetailsV2
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
		if len(templatesV2) > 0 {
			fmt.Printf("   mcp-cli --template %s\n", templatesV2[0])
		}
		if len(templatesV1) > 0 {
			fmt.Printf("   mcp-cli --template %s\n", templatesV1[0])
		}
		fmt.Println("   echo 'some data' | mcp-cli --template simple_analysis")
		fmt.Println()
		fmt.Println("💡 Remember to configure your API keys and server paths first!")
		fmt.Println("📦 Download Golang MCP servers from: https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers")
	}
	
	return nil
}

// outputTemplateResponse outputs the workflow response
func outputTemplateResponse(response *domain.WorkflowResponse) error {
	// If there's an error, output structured error for debugging
	if response.Error != nil {
		output, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal error response: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}
	
	// For successful execution, output clean text
	if response.FinalOutput != "" {
		fmt.Println(response.FinalOutput)
	} else if len(response.StepResults) > 0 {
		lastStep := response.StepResults[len(response.StepResults)-1]
		if lastStep.Output != "" {
			fmt.Println(lastStep.Output)
		} else {
			fmt.Printf("Template '%s' completed but produced no output\n", response.TemplateName)
		}
	} else {
		fmt.Printf("Template '%s' completed but produced no output\n", response.TemplateName)
	}
	
	return nil
}

// executeTemplateV2 executes a template v2 workflow
func executeTemplateV2(appConfig *config.ApplicationConfig, configService *infraConfig.Service, tmpl *config.TemplateV2) error {
	logging.Info("Executing template v2: %s (version: %s)", tmpl.Name, tmpl.Version)
	
	// Initialize provider
	actualProviderName := providerName
	var providerConfig *config.ProviderConfig
	var interfaceType config.InterfaceType
	
	if actualProviderName == "" {
		// Use default from template or config
		if tmpl.Config != nil && tmpl.Config.Defaults != nil && tmpl.Config.Defaults.Provider != "" {
			actualProviderName = tmpl.Config.Defaults.Provider
			var err error
			providerConfig, interfaceType, err = configService.GetProviderConfig(actualProviderName)
			if err != nil {
				return handleTemplateError(fmt.Errorf("failed to get template provider config: %w", err))
			}
		} else {
			// Use system default
			var err error
			actualProviderName, providerConfig, interfaceType, err = configService.GetDefaultProvider()
			if err != nil {
				return handleTemplateError(fmt.Errorf("failed to get default provider: %w", err))
			}
		}
	} else {
		// Use specified provider
		var err error
		providerConfig, interfaceType, err = configService.GetProviderConfig(actualProviderName)
		if err != nil {
			return handleTemplateError(fmt.Errorf("failed to get provider config: %w", err))
		}
	}
	
	// Override model if specified
	if modelName != "" {
		providerConfig.DefaultModel = modelName
	} else if tmpl.Config != nil && tmpl.Config.Defaults != nil && tmpl.Config.Defaults.Model != "" {
		providerConfig.DefaultModel = tmpl.Config.Defaults.Model
	}
	
	// Create provider
	providerType := domain.ProviderType(actualProviderName)
	
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return handleTemplateError(fmt.Errorf("failed to create provider: %w", err))
	}
	
	// Prepare input data
	var processedInputData string
	if inputData != "" {
		processedInputData = inputData
	} else {
		// Check stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return handleTemplateError(fmt.Errorf("failed to read stdin: %w", err))
			}
			processedInputData = string(stdinData)
		}
	}
	
	// Collect all servers needed from all steps
	serverSet := make(map[string]bool)
	for _, step := range tmpl.Steps {
		for _, server := range step.Servers {
			serverSet[server] = true
		}
	}
	
	serverNames := make([]string, 0, len(serverSet))
	for server := range serverSet {
		serverNames = append(serverNames, server)
	}
	
	// Execute template v2
	var result *template.ExecutionResult
	
	if len(serverNames) == 0 {
		// No servers - execute without connections
		logging.Debug("Executing template v2 without MCP servers")
		
		serverManager := NewHostServerManager([]*host.ServerConnection{})
		workflowServiceV2 := workflowservice.NewServiceV2(appConfig, provider, serverManager)
		
		ctx := context.Background()
		result, err = workflowServiceV2.ExecuteTemplate(ctx, tmpl.Name, processedInputData)
		if err != nil {
			return handleTemplateError(fmt.Errorf("template v2 execution failed: %w", err))
		}
	} else {
		// With servers
		userSpecified := make(map[string]bool)
		for _, s := range serverNames {
			userSpecified[s] = true
		}
		
		err = host.RunCommandWithOptions(func(conns []*host.ServerConnection) error {
			serverManager := NewHostServerManager(conns)
			workflowServiceV2 := workflowservice.NewServiceV2(appConfig, provider, serverManager)
			
			ctx := context.Background()
			result, err = workflowServiceV2.ExecuteTemplate(ctx, tmpl.Name, processedInputData)
			return err
		}, configFile, serverNames, userSpecified, host.QuietCommandOptions())
		
		if err != nil {
			return handleTemplateError(err)
		}
	}
	
	// Output result
	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "Template v2 execution failed: %v\n", result.Error)
		return result.Error
	}
	
	if result.FinalOutput != "" {
		fmt.Println(result.FinalOutput)
	} else {
		fmt.Printf("Template '%s' completed but produced no output\n", tmpl.Name)
	}
	
	return nil
}

// handleTemplateError handles template execution errors
func handleTemplateError(err error) error {
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
	
	if outputErr := outputTemplateResponse(errorResponse); outputErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	
	return err
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
		
		for _, tool := range toolsList {
			if tool.Function.Name == toolName {
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

func (hsa *HostServerAdapter) GetServerName() string {
	return hsa.connection.Name
}

func (hsa *HostServerAdapter) GetConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Command: "mock",
		Args:    []string{},
	}
}


