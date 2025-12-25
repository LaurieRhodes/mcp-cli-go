package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Service implements the WorkflowProcessor interface
type Service struct {
	config           *config.ApplicationConfig
	configService    domain.ConfigurationService
	provider         domain.LLMProvider
	serverManager    domain.MCPServerManager
}

// NewService creates a new workflow service
func NewService(
	appConfig *config.ApplicationConfig,
	configService domain.ConfigurationService,
	provider domain.LLMProvider,
	serverManager domain.MCPServerManager,
) *Service {
	return &Service{
		config:        appConfig,
		configService: configService,
		provider:      provider,
		serverManager: serverManager,
	}
}

// ProcessWorkflow executes a complete workflow template
func (s *Service) ProcessWorkflow(ctx context.Context, req *domain.WorkflowRequest) (*domain.WorkflowResponse, error) {
	startTime := time.Now()
	
	// Generate execution ID if not provided
	if req.ExecutionID == "" {
		bytes := make([]byte, 16)
		rand.Read(bytes)
		req.ExecutionID = hex.EncodeToString(bytes)
	}
	
	logging.Info("Starting workflow execution: %s (template: %s)", req.ExecutionID, req.TemplateName)
	
	// Get workflow template from config
	configTemplate, exists := s.config.GetWorkflowTemplate(req.TemplateName)
	if !exists {
		return nil, fmt.Errorf("workflow template '%s' not found", req.TemplateName)
	}
	
	// Convert config template to domain template
	template := convertConfigTemplateToDomain(configTemplate)
	
	// Validate workflow template
	if err := s.ValidateWorkflow(template); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}
	
	// Initialize response
	response := &domain.WorkflowResponse{
		ExecutionID:   req.ExecutionID,
		TemplateName:  req.TemplateName,
		Status:        domain.WorkflowStatusRunning,
		StepResults:   make([]domain.WorkflowStepResult, 0, len(template.Steps)),
		Variables:     make(map[string]interface{}),
		Timestamp:     startTime,
		Metadata:      req.Metadata,
	}
	
	// Initialize variables from template and request
	s.initializeVariables(response.Variables, template, req)
	
	// Process each step
	for _, step := range template.Steps {
		// Process the step
		stepResult, err := s.ProcessStep(ctx, &step, response.Variables)
		if err != nil {
			stepResult.Status = domain.StepStatusFailed
			stepResult.Error = &domain.WorkflowError{
				Code:    "STEP_EXECUTION_ERROR",
				Message: err.Error(),
				Step:    step.Step,
			}
			
			response.StepResults = append(response.StepResults, *stepResult)
			response.Status = domain.WorkflowStatusFailed
			response.Error = stepResult.Error
			break
		}
		
		response.StepResults = append(response.StepResults, *stepResult)
		
		// Store step output in variables
		response.Variables[fmt.Sprintf("step%d_output", step.Step)] = stepResult.Output
	}
	
	// Set final status if not already failed
	if response.Status != domain.WorkflowStatusFailed {
		response.Status = domain.WorkflowStatusCompleted
	}
	
	// Set final output
	if len(response.StepResults) > 0 {
		lastStep := response.StepResults[len(response.StepResults)-1]
		if lastStep.Status == domain.StepStatusCompleted {
			response.FinalOutput = lastStep.Output
		}
	}
	
	response.ExecutionTime = time.Since(startTime)
	
	logging.Info("Workflow execution completed: %s (status: %s, time: %v)", 
		req.ExecutionID, response.Status, response.ExecutionTime)
	
	return response, nil
}

// ProcessStep executes a single workflow step with tool follow-up support
func (s *Service) ProcessStep(ctx context.Context, step *domain.WorkflowStep, variables map[string]interface{}) (*domain.WorkflowStepResult, error) {
	startTime := time.Now()
	
	result := &domain.WorkflowStepResult{
		Step:      step.Step,
		Name:      step.Name,
		Status:    domain.StepStatusRunning,
		Variables: make(map[string]interface{}),
	}
	
	logging.Debug("Processing step %d: %s", step.Step, step.Name)
	
	// Process the prompt with variable substitution
	template := s.getWorkflowTemplate(step)
	if template == nil {
		return result, fmt.Errorf("workflow template not found")
	}
	
	processedPrompt := template.ProcessVariables(step.BasePrompt, variables)
	
	// Create initial messages array
	messages := []domain.Message{
		{
			Role:    "user",
			Content: processedPrompt,
		},
	}
	
	// Get available tools for this step
	tools, err := s.getStepTools(step)
	if err != nil {
		return result, fmt.Errorf("failed to get tools for step: %w", err)
	}
	
	logging.Debug("Step %d has %d tools available", step.Step, len(tools))
	
	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		Tools:        tools,
		SystemPrompt: step.SystemPrompt,
		Temperature:  step.Temperature,
		MaxTokens:    step.MaxTokens,
	}
	
	// Execute initial request
	completionResp, err := s.provider.CreateCompletion(ctx, completionReq)
	if err != nil {
		return result, fmt.Errorf("completion failed: %w", err)
	}
	
	logging.Debug("Initial LLM response: %s", completionResp.Response)
	
	// Handle tool calls with follow-ups
	maxFollowUps := 2
	followUpsUsed := 0
	
	for followUpsUsed < maxFollowUps && len(completionResp.ToolCalls) > 0 {
		logging.Info("Processing %d tool calls (follow-up #%d)", len(completionResp.ToolCalls), followUpsUsed+1)
		
		// Add assistant message with tool calls
		assistantMessage := domain.Message{
			Role:      "assistant",
			Content:   completionResp.Response,
			ToolCalls: completionResp.ToolCalls,
		}
		messages = append(messages, assistantMessage)
		
		// Execute tools and collect results
		for _, toolCall := range completionResp.ToolCalls {
			logging.Debug("Executing tool call: %s", toolCall.Function.Name)
			
			// Parse tool arguments
			var args map[string]interface{}
			if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
				return result, fmt.Errorf("failed to parse tool arguments: %w", err)
			}
			
			// Execute tool
			output, err := s.serverManager.ExecuteTool(ctx, toolCall.Function.Name, args)
			if err != nil {
				logging.Error("Tool execution failed for %s: %v", toolCall.Function.Name, err)
				output = fmt.Sprintf("Error: %s", err.Error())
			}
			
			// Store tool result in variables
			result.Variables[fmt.Sprintf("tool_%s_result", toolCall.Function.Name)] = output
			
			// Add tool result message
			toolResultMessage := domain.Message{
				Role:       "tool",
				Content:    output,
				ToolCallID: toolCall.ID,
			}
			messages = append(messages, toolResultMessage)
			
			logging.Debug("Tool %s executed successfully", toolCall.Function.Name)
		}
		
		// Get follow-up response from LLM
		completionReq.Messages = messages
		followUpResponse, err := s.provider.CreateCompletion(ctx, completionReq)
		if err != nil {
			return result, fmt.Errorf("LLM follow-up request failed: %w", err)
		}
		
		completionResp = followUpResponse
		followUpsUsed++
		
		logging.Debug("Follow-up response #%d: %s", followUpsUsed, completionResp.Response)
	}
	
	// Set result data
	result.Output = completionResp.Response
	result.Provider = string(s.provider.GetProviderType())
	result.Usage = completionResp.Usage
	result.ToolCalls = completionResp.ToolCalls
	result.ExecutionTime = time.Since(startTime)
	result.Status = domain.StepStatusCompleted
	
	logging.Debug("Step %d final output: %s", step.Step, result.Output)
	
	return result, nil
}

// ValidateWorkflow validates a workflow template
func (s *Service) ValidateWorkflow(template *domain.WorkflowTemplate) error {
	return template.ValidateWorkflowTemplate()
}

// getWorkflowTemplate retrieves the workflow template for a step
func (s *Service) getWorkflowTemplate(step *domain.WorkflowStep) *domain.WorkflowTemplate {
	// Find the template that contains this step
	for _, configTemplate := range s.config.Templates {
		for _, templateStep := range configTemplate.Steps {
			if templateStep.Step == step.Step && templateStep.Name == step.Name {
				return convertConfigTemplateToDomain(configTemplate)
			}
		}
	}
	return nil
}

// initializeVariables initializes workflow variables
func (s *Service) initializeVariables(variables map[string]interface{}, template *domain.WorkflowTemplate, req *domain.WorkflowRequest) {
	// Add template variables
	if template.Variables != nil {
		for key, value := range template.Variables {
			variables[key] = value
		}
	}
	
	// Add request variables
	if req.Variables != nil {
		for key, value := range req.Variables {
			variables[key] = value
		}
	}
	
	// Add input data
	if req.InputData != "" {
		variables["input_data"] = req.InputData
		variables["stdin"] = req.InputData
	}
}

// getStepTools retrieves available tools for a workflow step
func (s *Service) getStepTools(step *domain.WorkflowStep) ([]domain.Tool, error) {
	// Get all available tools from servers
	allTools, err := s.serverManager.GetAvailableTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tools: %w", err)
	}
	
	// If no specific tools required, provide all tools
	if len(step.ToolsRequired) == 0 {
		logging.Debug("No specific tools required, providing all %d available tools", len(allTools))
		return allTools, nil
	}
	
	// Filter to only required tools
	var stepTools []domain.Tool
	for _, requiredTool := range step.ToolsRequired {
		for _, tool := range allTools {
			if tool.Function.Name == requiredTool {
				stepTools = append(stepTools, tool)
				break
			}
		}
	}
	
	// Check if all required tools are available
	if len(stepTools) != len(step.ToolsRequired) {
		missing := make([]string, 0)
		for _, required := range step.ToolsRequired {
			found := false
			for _, tool := range stepTools {
				if tool.Function.Name == required {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, required)
			}
		}
		
		return nil, fmt.Errorf("required tools not available: %s", strings.Join(missing, ", "))
	}
	
	logging.Debug("Providing %d specific required tools", len(stepTools))
	return stepTools, nil
}

// ExecuteWorkflow is a convenience method that creates a workflow request and executes it
func (s *Service) ExecuteWorkflow(ctx context.Context, templateName string, inputData string) (*domain.WorkflowResponse, error) {
	// Generate execution ID
	bytes := make([]byte, 16)
	rand.Read(bytes)
	executionID := hex.EncodeToString(bytes)
	
	req := &domain.WorkflowRequest{
		TemplateName: templateName,
		InputData:    inputData,
		ExecutionID:  executionID,
		Variables:    make(map[string]interface{}),
		Metadata: map[string]interface{}{
			"created_at": time.Now(),
		},
	}
	
	return s.ProcessWorkflow(ctx, req)
}

// ListAvailableTemplates returns a list of available workflow templates
func (s *Service) ListAvailableTemplates() []string {
	return s.config.ListWorkflowTemplates()
}

// GetTemplateInfo returns information about a specific workflow template
func (s *Service) GetTemplateInfo(templateName string) (*domain.WorkflowTemplate, error) {
	configTemplate, exists := s.config.GetWorkflowTemplate(templateName)
	if !exists {
		return nil, fmt.Errorf("workflow template '%s' not found", templateName)
	}
	
	return convertConfigTemplateToDomain(configTemplate), nil
}

// ValidateTemplateConfiguration validates all workflow templates in the configuration
func (s *Service) ValidateTemplateConfiguration() error {
	return s.config.ValidateWorkflowTemplates()
}

// convertConfigTemplateToDomain converts config.WorkflowTemplate to domain.WorkflowTemplate
func convertConfigTemplateToDomain(configTemplate *config.WorkflowTemplate) *domain.WorkflowTemplate {
	if configTemplate == nil {
		return nil
	}
	
	// Convert steps
	domainSteps := make([]domain.WorkflowStep, len(configTemplate.Steps))
	for i, configStep := range configTemplate.Steps {
		domainSteps[i] = domain.WorkflowStep{
			Step:          configStep.Step,
			Name:          configStep.Name,
			BasePrompt:    configStep.BasePrompt,
			SystemPrompt:  configStep.SystemPrompt,
			Provider:      configStep.Provider,
			Model:         configStep.Model,
			Servers:       configStep.Servers,
			ToolsRequired: configStep.ToolsRequired,
			Temperature:   configStep.Temperature,
			MaxTokens:     configStep.MaxTokens,
		}
	}
	
	return &domain.WorkflowTemplate{
		Name:        configTemplate.Name,
		Description: configTemplate.Description,
		Steps:       domainSteps,
		Variables:   configTemplate.Variables,
	}
}
