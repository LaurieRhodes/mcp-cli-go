package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	workflowservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// Service implements the MCP server message handler
// ProgressNotifier interface for sending progress notifications
type ProgressNotifier interface {
	SendProgressNotification(progressToken string, progress float64, total int, message string)
}

type Service struct {
	runasConfig        *runas.RunAsConfig
	appConfig          *config.ApplicationConfig
	configService      *infraConfig.Service
	skillService       skills.SkillService
	progressNotifier   ProgressNotifier
}

// NewService creates a new MCP server service
func NewService(runasConfig *runas.RunAsConfig, appConfig *config.ApplicationConfig, configService *infraConfig.Service, skillService skills.SkillService) *Service {
	return &Service{
		runasConfig:   runasConfig,
		appConfig:     appConfig,
		configService: configService,
		skillService:  skillService,
	}
}

// SetProgressNotifier sets the progress notifier for sending progress updates
func (s *Service) SetProgressNotifier(notifier ProgressNotifier) {
	s.progressNotifier = notifier
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
	
	// Log params for debugging
	logging.Debug("Raw params received: %+v", params)
	
	// Extract progress token if present (MCP protocol support)
	var progressToken string
	if meta, ok := params["_meta"].(map[string]interface{}); ok {
		if token, ok := meta["progressToken"].(string); ok {
			progressToken = token
			logging.Info("Progress token extracted: %s", progressToken)
		} else {
			logging.Warn("_meta exists but progressToken not found or not string")
		}
	} else {
		logging.Warn("No _meta field in params (progress notifications disabled)")
	}
	
	// Find the tool exposure
	toolExposure, found := s.runasConfig.GetToolByName(toolName)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	
	// CHECK: Is this the execute_skill_code tool? (identified by template)
	if toolExposure.Template == "execute_skill_code" {
		return s.handleExecuteSkillCode(arguments)
	}
	
	// CHECK: Is this a skill tool (uses load_skill template)?
	if toolExposure.Template == "load_skill" {
		return s.handleSkillToolCall(toolExposure, arguments)
	}
	
	// Execute the template with progress token
	result, err := s.executeTemplateWithProgress(toolExposure, arguments, progressToken)
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

// executeTemplateWithProgress executes a template and sends progress notifications
func (s *Service) executeTemplateWithProgress(toolExposure *runas.ToolExposure, arguments map[string]interface{}, progressToken string) (string, error) {
	logging.Info("Executing template with progress support: token=%s, hasNotifier=%v", 
		progressToken, s.progressNotifier != nil)
	
	// Send initial progress (0%)
	if progressToken != "" && s.progressNotifier != nil {
		logging.Info("Sending initial progress notification (0%%)")
		s.progressNotifier.SendProgressNotification(progressToken, 0.0, 0, fmt.Sprintf("Starting %s", toolExposure.Name))
	} else {
		if progressToken == "" {
			logging.Warn("No progress token provided - progress notifications disabled")
		}
		if s.progressNotifier == nil {
			logging.Warn("No progress notifier available - progress notifications disabled")
		}
	}
	
	// Start heartbeat goroutine to send periodic progress updates
	// This keeps the client alive during long-running template execution
	done := make(chan bool)
	if progressToken != "" && s.progressNotifier != nil {
		go func() {
			ticker := time.NewTicker(20 * time.Second)  // Send heartbeat every 20 seconds
			defer ticker.Stop()
			
			for {
				select {
				case <-ticker.C:
					// Send a "still working" notification with same progress value
					// This resets the client's timeout without implying actual progress
					s.progressNotifier.SendProgressNotification(
						progressToken, 
						0.5,  // Mid-point to indicate "in progress"
						0, 
						fmt.Sprintf("Executing %s...", toolExposure.Name),
					)
					logging.Debug("Sent heartbeat progress notification")
				case <-done:
					return
				}
			}
		}()
	}
	
	// Execute the template (this blocks)
	result, err := s.executeTemplate(toolExposure, arguments)
	
	// Stop heartbeat
	close(done)
	
	// Send completion progress (100%)
	if progressToken != "" && s.progressNotifier != nil {
		if err != nil {
			logging.Info("Sending failure progress notification (100%%)")
			s.progressNotifier.SendProgressNotification(progressToken, 1.0, 0, fmt.Sprintf("Failed: %v", err))
		} else {
			logging.Info("Sending completion progress notification (100%%)")
			s.progressNotifier.SendProgressNotification(progressToken, 1.0, 0, fmt.Sprintf("Completed %s", toolExposure.Name))
		}
	}
	
	return result, err
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
	// Convert provider name to ProviderType (configuration-driven)
	providerType := domain.ProviderType(providerName)
	
	logging.Debug("Creating provider: %s", providerType)
	
	// Get interface type from configuration
	_, interfaceType, err := s.configService.GetProviderConfig(providerName)
	if err != nil {
		return "", fmt.Errorf("failed to get provider interface type: %w", err)
	}
	
	// Create provider using factory
	providerFactory := ai.NewProviderFactory()
	provider, err := providerFactory.CreateProvider(providerType, providerConfig, interfaceType)
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

// handleSkillToolCall handles calls to skill tools (tools using load_skill template)
func (s *Service) handleSkillToolCall(toolExposure *runas.ToolExposure, arguments map[string]interface{}) (map[string]interface{}, error) {
	logging.Info("Handling skill tool call: %s", toolExposure.Name)
	
	// Extract skill name from input mapping
	skillName := ""
	if mapping, ok := toolExposure.InputMapping["skill_name"]; ok {
		skillName = mapping
	} else {
		// Fallback: convert tool name (python_best_practices -> python-best-practices)
		skillName = strings.ReplaceAll(toolExposure.Name, "_", "-")
	}
	
	// Build skill load request
	request := &skills.SkillLoadRequest{
		SkillName: skillName,
		Mode:      skills.SkillLoadModePassive, // Default
	}
	
	// Extract parameters from arguments
	if mode, ok := arguments["mode"].(string); ok {
		request.Mode = skills.SkillLoadMode(mode)
	}
	
	if includeRefs, ok := arguments["include_references"].(bool); ok {
		request.IncludeReferences = includeRefs
	}
	
	if refFiles, ok := arguments["reference_files"].([]interface{}); ok {
		for _, ref := range refFiles {
			if refStr, ok := ref.(string); ok {
				request.ReferenceFiles = append(request.ReferenceFiles, refStr)
			}
		}
	}
	
	if inputData, ok := arguments["input_data"].(string); ok {
		request.InputData = inputData
	}
	
	// Load the skill
	result, err := s.skillService.LoadSkillByRequest(request)
	if err != nil {
		logging.Error("Failed to load skill: %v", err)
		return map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Failed to load skill %s: %v", skillName, err),
				},
			},
			"isError": true,
		}, nil
	}
	
	// Return skill content
	content := result.Content
	if result.Mode == skills.SkillLoadModeActive {
		content = result.Result
	}
	
	logging.Info("Successfully loaded skill: %s (%d chars)", skillName, len(content))
	
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": content,
			},
		},
	}, nil
}

// handleExecuteSkillCode handles the execute_skill_code tool
func (s *Service) handleExecuteSkillCode(arguments map[string]interface{}) (map[string]interface{}, error) {
	logging.Info("Handling execute_skill_code request")
	
	// Extract skill_name
	skillName, ok := arguments["skill_name"].(string)
	if !ok || skillName == "" {
		return s.errorResponse("skill_name parameter is required"), nil
	}
	
	// Extract language (default to python)
	language := "python"
	if lang, ok := arguments["language"].(string); ok && lang != "" {
		language = lang
	}
	
	// Extract code
	code, ok := arguments["code"].(string)
	if !ok || code == "" {
		return s.errorResponse("code parameter is required"), nil
	}
	
	logging.Info("Executing code for skill: %s (language: %s, code length: %d)", skillName, language, len(code))
	
	// Extract files (optional)
	files := make(map[string][]byte)
	if filesArg, ok := arguments["files"].(map[string]interface{}); ok {
		for filename, content := range filesArg {
			// Content should be base64 encoded
			if contentStr, ok := content.(string); ok {
				// Try to decode as base64
				decoded, err := base64.StdEncoding.DecodeString(contentStr)
				if err != nil {
					// If not base64, treat as plain text
					decoded = []byte(contentStr)
				}
				files[filename] = decoded
				logging.Debug("Added file: %s (%d bytes)", filename, len(decoded))
			}
		}
	}
	
	// Create execution request
	request := &skills.CodeExecutionRequest{
		SkillName: skillName,
		Language:  language,
		Code:      code,
		Files:     files,
		Timeout:   60, // 60 second timeout
	}
	
	// Execute code
	result, err := s.skillService.ExecuteCode(request)
	if err != nil {
		logging.Error("Code execution failed: %v", err)
		return s.errorResponse(fmt.Sprintf("Code execution failed: %v", err)), nil
	}
	
	// Check if execution had an error
	if result.Error != nil {
		logging.Warn("Code execution completed with error: %v", result.Error)
		return s.errorResponse(fmt.Sprintf("Code execution error: %v\n\nOutput:\n%s", result.Error, result.Output)), nil
	}
	
	// Success!
	logging.Info("Code executed successfully (exit code: %d, duration: %dms)", result.ExitCode, result.Duration)
	
	// Format response with output
	responseText := result.Output
	if result.Duration > 0 {
		responseText = fmt.Sprintf("%s\n\n[Executed in %dms]", result.Output, result.Duration)
	}
	
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": responseText,
			},
		},
	}, nil
}

// errorResponse creates an error response in MCP format
func (s *Service) errorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": message,
			},
		},
		"isError": true,
	}
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
