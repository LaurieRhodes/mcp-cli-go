package server

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	workflowservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// Service implements the MCP server message handler
type Service struct {
	runasConfig   *runas.RunAsConfig
	appConfig     *config.ApplicationConfig
	configService *infraConfig.Service
}

// NewService creates a new MCP server service
func NewService(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig, configService *infraConfig.Service) *Service {
	return &Service{
		runasConfig:   runasConfig,
		appConfig:     appConfig,
		configService: configService,
	}
}

// HandleInitialize handles the initialize request
func (s *Service) HandleInitialize(params map[string]interface{}) (map[string]interface{}, error) {
	logging.Info("Initialize request from client")
	
	// Log client info if provided
	if clientInfo, ok := params["clientInfo"].(map[string]interface{}); ok {
		if name, ok := clientInfo["name"].(string); ok {
			logging.Info("Client name: %s", name)
		}
		if version, ok := clientInfo["version"].(string); ok {
			logging.Info("Client version: %s", version)
		}
	}
	
	// Return server info and capabilities
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    s.runasConfig.ServerInfo.Name,
			"version": s.runasConfig.ServerInfo.Version,
		},
	}, nil
}

// HandleToolsList handles the tools/list request
func (s *Service) HandleToolsList(params map[string]interface{}) (map[string]interface{}, error) {
	logging.Info("Listing available tools")
	
	// Convert tool exposures to MCP tool format
	tools := make([]map[string]interface{}, 0, len(s.runasConfig.Tools))
	
	for _, toolExposure := range s.runasConfig.Tools {
		tool := map[string]interface{}{
			"name":        toolExposure.Name,
			"description": toolExposure.Description,
			"inputSchema": toolExposure.InputSchema,
		}
		
		tools = append(tools, tool)
		logging.Debug("Registered tool: %s (template: %s)", toolExposure.Name, toolExposure.Template)
	}
	
	logging.Info("Returning %d tools", len(tools))
	
	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// HandleToolsCall handles the tools/call request
func (s *Service) HandleToolsCall(params map[string]interface{}) (map[string]interface{}, error) {
	// Extract tool name
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'name' parameter")
	}
	
	logging.Info("Tool call request: %s", toolName)
	
	// Extract arguments
	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		// Arguments may be optional
		arguments = make(map[string]interface{})
	}
	
	logging.Debug("Tool arguments: %+v", arguments)
	
	// Find the tool exposure
	toolExposure, found := s.runasConfig.GetToolByName(toolName)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	
	// Execute the template
	result, err := s.executeTemplate(toolExposure, arguments)
	if err != nil {
		logging.Error("Template execution failed: %v", err)
		
		// Return error in MCP format
		return map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Template execution failed: %v", err),
				},
			},
			"isError": true,
		}, nil
	}
	
	// Return success result in MCP format
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

// executeTemplate executes a template with the given arguments
func (s *Service) executeTemplate(toolExposure *runas.ToolExposure, arguments map[string]interface{}) (string, error) {
	logging.Info("Executing template: %s", toolExposure.Template)
	
	// Check if template exists (v2 first, then v1)
	var isV2 bool
	var templateV2 *config.TemplateV2
	
	if tmpl, exists := s.appConfig.TemplatesV2[toolExposure.Template]; exists {
		isV2 = true
		templateV2 = tmpl
		logging.Debug("Using template v2: %s", toolExposure.Template)
	} else if _, exists := s.appConfig.Templates[toolExposure.Template]; !exists {
		return "", fmt.Errorf("template not found: %s", toolExposure.Template)
	}
	
	// Prepare input data by applying input mapping
	inputData, err := s.prepareInputData(toolExposure, arguments)
	if err != nil {
		return "", fmt.Errorf("failed to prepare input data: %w", err)
	}
	
	logging.Debug("Input data prepared: %s", inputData)
	
	// Execute template based on version
	if isV2 {
		return s.executeTemplateV2(templateV2, inputData, toolExposure)
	}
	
	return s.executeTemplateV1(toolExposure.Template, inputData, toolExposure)
}

// prepareInputData applies input mapping to convert tool arguments to template input
func (s *Service) prepareInputData(toolExposure *runas.ToolExposure, arguments map[string]interface{}) (string, error) {
	// If no input mapping, use first argument as-is
	if len(toolExposure.InputMapping) == 0 {
		// Try "data" field first
		if data, ok := arguments["data"]; ok {
			if str, ok := data.(string); ok {
				return str, nil
			}
		}
		
		// Try first argument
		for _, v := range arguments {
			return fmt.Sprintf("%v", v), nil
		}
		
		return "", nil
	}
	
	// Apply input mapping - simple replacement for now
	result := ""
	for argName := range toolExposure.InputMapping {
		if argValue, ok := arguments[argName]; ok {
			result = fmt.Sprintf("%v", argValue)
			break // Use first mapped argument
		}
	}
	
	return result, nil
}

// executeTemplateV1 executes a v1 workflow template
func (s *Service) executeTemplateV1(templateName string, inputData string, toolExposure *runas.ToolExposure) (string, error) {
	// TODO: Implement v1 template execution
	return fmt.Sprintf("V1 template execution not yet implemented for: %s", templateName), nil
}

// executeTemplateV2 executes a v2 template
func (s *Service) executeTemplateV2(tmpl *config.TemplateV2, inputData string, toolExposure *runas.ToolExposure) (string, error) {
	logging.Info("Executing template v2: %s", tmpl.Name)
	
	// Get provider configuration
	var providerName string
	var providerConfig *config.ProviderConfig
	var err error
	
	if toolExposure.Overrides != nil && toolExposure.Overrides.Provider != "" {
		providerName = toolExposure.Overrides.Provider
		providerConfig, _, err = s.configService.GetProviderConfig(providerName)
	} else if tmpl.Config != nil && tmpl.Config.Defaults != nil && tmpl.Config.Defaults.Provider != "" {
		providerName = tmpl.Config.Defaults.Provider
		providerConfig, _, err = s.configService.GetProviderConfig(providerName)
	} else {
		providerName, providerConfig, _, err = s.configService.GetDefaultProvider()
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to get provider config: %w", err)
	}
	
	// Override model if specified
	if toolExposure.Overrides != nil && toolExposure.Overrides.Model != "" {
		providerConfig.DefaultModel = toolExposure.Overrides.Model
	} else if tmpl.Config != nil && tmpl.Config.Defaults != nil && tmpl.Config.Defaults.Model != "" {
		providerConfig.DefaultModel = tmpl.Config.Defaults.Model
	}
	
	logging.Info("Using provider: %s (model: %s)", providerName, providerConfig.DefaultModel)
	
	// Import the provider factory and domain types to create the actual provider
	// This implementation mirrors the CLI's executeTemplateV2 function
	return s.executeTemplateV2WithProvider(tmpl, inputData, providerName, providerConfig)
}

// executeTemplateV2WithProvider executes a template with the actual provider
func (s *Service) executeTemplateV2WithProvider(tmpl *config.TemplateV2, inputData string, providerName string, providerConfig *config.ProviderConfig) (string, error) {
	// Map provider name to provider type
	var providerType domain.ProviderType
	switch strings.ToLower(providerName) {
	case "openai":
		providerType = domain.ProviderOpenAI
	case "anthropic":
		providerType = domain.ProviderAnthropic
	case "ollama":
		providerType = domain.ProviderOllama
	case "deepseek":
		providerType = domain.ProviderDeepSeek
	case "gemini":
		providerType = domain.ProviderGemini
	case "openrouter":
		providerType = domain.ProviderOpenRouter
	default:
		return "", fmt.Errorf("unsupported provider: %s", providerName)
	}
	
	logging.Debug("Creating provider: %s", providerType)
	
	// Create provider using factory
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, providerConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create provider: %w", err)
	}
	
	// For MCP server context, we don't manage external server connections
	// Templates should use their own server configurations
	// Create an empty server manager for now
	serverManager := NewEmptyServerManager()
	
	// Create workflow service v2
	workflowServiceV2 := workflowservice.NewServiceV2(s.appConfig, provider, serverManager)
	
	// Execute template
	ctx := context.Background()
	result, err := workflowServiceV2.ExecuteTemplate(ctx, tmpl.Name, inputData)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	
	// Check for execution error
	if result.Error != nil {
		return "", result.Error
	}
	
	// Return final output
	if result.FinalOutput != "" {
		return result.FinalOutput, nil
	}
	
	return fmt.Sprintf("Template '%s' completed but produced no output", tmpl.Name), nil
}

// EmptyServerManager implements a minimal MCPServerManager for templates that don't need servers
type EmptyServerManager struct{}

func NewEmptyServerManager() *EmptyServerManager {
	return &EmptyServerManager{}
}

func (esm *EmptyServerManager) StartServer(ctx context.Context, serverName string, cfg *config.ServerConfig) (domain.MCPServer, error) {
	return nil, fmt.Errorf("server management not available in MCP server mode")
}

func (esm *EmptyServerManager) StopServer(serverName string) error {
	return nil
}

func (esm *EmptyServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	return nil, false
}

func (esm *EmptyServerManager) ListServers() map[string]domain.MCPServer {
	return make(map[string]domain.MCPServer)
}

func (esm *EmptyServerManager) GetAvailableTools() ([]domain.Tool, error) {
	return []domain.Tool{}, nil
}

func (esm *EmptyServerManager) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	return "", fmt.Errorf("tool '%s' not found (no servers configured)", toolName)
}

func (esm *EmptyServerManager) StopAll() error {
	return nil
}
