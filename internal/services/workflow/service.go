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
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Service implements the WorkflowProcessor interface
type Service struct {
	config           *domain.ApplicationConfig
	configService    domain.ConfigurationService
	providerService  domain.LLMProvider
	serverManager    domain.MCPServerManager
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// NewService creates a new workflow service
func NewService(
	config *domain.ApplicationConfig,
	configService domain.ConfigurationService,
	providerService domain.LLMProvider,
	serverManager domain.MCPServerManager,
) (*Service, error) {
	if config == nil {
		return nil, domain.NewDomainError(domain.ErrCodeConfigInvalid, "configuration is required")
	}
	
	if configService == nil {
		return nil, domain.NewDomainError(domain.ErrCodeConfigInvalid, "configuration service is required")
	}
	
	if providerService == nil {
		return nil, domain.NewDomainError(domain.ErrCodeConfigInvalid, "provider service is required")
	}
	
	if serverManager == nil {
		return nil, domain.NewDomainError(domain.ErrCodeConfigInvalid, "server manager is required")
	}

	return &Service{
		config:          config,
		configService:   configService,
		providerService: providerService,
		serverManager:   serverManager,
	}, nil
}

// ProcessWorkflow executes a complete workflow template
func (s *Service) ProcessWorkflow(ctx context.Context, req *domain.WorkflowRequest) (*domain.WorkflowResponse, error) {
	startTime := time.Now()
	
	// Generate execution ID if not provided
	if req.ExecutionID == "" {
		req.ExecutionID = generateExecutionID()
	}
	
	logging.Info("Starting workflow execution: %s (template: %s)", req.ExecutionID, req.TemplateName)
	
	// Get workflow template
	template, exists := s.config.GetWorkflowTemplate(req.TemplateName)
	if !exists {
		return nil, domain.NewDomainError(domain.ErrCodeRequestInvalid, 
			fmt.Sprintf("workflow template '%s' not found", req.TemplateName))
	}
	
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
		// Check if we should start from a specific step
		if req.StartFromStep > 0 && step.Step < req.StartFromStep {
			stepResult := domain.WorkflowStepResult{
				Step:   step.Step,
				Name:   step.Name,
				Status: domain.StepStatusSkipped,
			}
			response.StepResults = append(response.StepResults, stepResult)
			continue
		}
		
		// Check if we should stop at a specific step
		if req.StopAtStep > 0 && step.Step > req.StopAtStep {
			break
		}
		
		// Check step conditions
		if !step.ShouldExecute(response.Variables) {
			stepResult := domain.WorkflowStepResult{
				Step:   step.Step,
				Name:   step.Name,
				Status: domain.StepStatusSkipped,
			}
			response.StepResults = append(response.StepResults, stepResult)
			logging.Debug("Step %d (%s) skipped due to conditions", step.Step, step.Name)
			continue
		}
		
		// Process the step
		stepResult, err := s.ProcessStep(ctx, &step, response.Variables)
		if err != nil {
			stepResult.Status = domain.StepStatusFailed
			stepResult.Error = &domain.WorkflowError{
				Code:      "STEP_EXECUTION_ERROR",
				Message:   err.Error(),
				Step:      step.Step,
				StepName:  step.Name,
				Retryable: s.isRetryableError(err),
			}
			
			response.StepResults = append(response.StepResults, *stepResult)
			
			// Check if we should fail the entire workflow
			if template.Settings != nil && template.Settings.FailOnStepError {
				response.Status = domain.WorkflowStatusFailed
				response.Error = stepResult.Error
				break
			} else {
				logging.Warn("Step %d (%s) failed but continuing: %v", step.Step, step.Name, err)
				continue
			}
		}
		
		response.StepResults = append(response.StepResults, *stepResult)
		
		// Save step output to variables if specified
		if step.OutputVariable != "" && stepResult.Output != "" {
			response.Variables[step.OutputVariable] = stepResult.Output
		}
		
		// Update variables from step result
		if stepResult.Variables != nil {
			for key, value := range stepResult.Variables {
				response.Variables[key] = value
			}
		}
		
		logging.Debug("Step %d (%s) completed successfully", step.Step, step.Name)
	}
	
	// Set final status if not already failed
	if response.Status != domain.WorkflowStatusFailed {
		response.Status = domain.WorkflowStatusCompleted
	}
	
	// Set final output (last step's output or specified output variable)
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

// ProcessStep executes a single workflow step with tool follow-up support (like query mode)
func (s *Service) ProcessStep(ctx context.Context, step *domain.WorkflowStep, variables map[string]interface{}) (*domain.WorkflowStepResult, error) {
	startTime := time.Now()
	
	result := &domain.WorkflowStepResult{
		Step:      step.Step,
		Name:      step.Name,
		Status:    domain.StepStatusRunning,
		Variables: make(map[string]interface{}),
	}
	
	logging.Debug("Processing step %d: %s", step.Step, step.Name)
	
	// Get timeout for this step
	timeout, err := step.GetTimeout()
	if err != nil {
		return result, err
	}
	
	// Create context with timeout
	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Process the prompt with variable substitution
	template := s.getWorkflowTemplate(step)
	if template == nil {
		return result, domain.NewDomainError(domain.ErrCodeRequestInvalid, "workflow template not found")
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
	
	// Execute initial request with retry policy if configured
	var completionResp *domain.CompletionResponse
	if step.RetryPolicy != nil {
		completionResp, err = s.executeWithRetry(stepCtx, completionReq, step.RetryPolicy)
	} else {
		completionResp, err = s.providerService.CreateCompletion(stepCtx, completionReq)
	}
	
	if err != nil {
		return result, fmt.Errorf("completion failed: %w", err)
	}
	
	logging.Debug("Initial LLM response: %s", completionResp.Response)
	
	// Handle tool calls with follow-ups (same logic as query mode)
	maxFollowUps := 2 // Same default as query mode
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
			output, err := s.serverManager.ExecuteTool(stepCtx, toolCall.Function.Name, args)
			if err != nil {
				logging.Error("Tool execution failed for %s: %v", toolCall.Function.Name, err)
				output = fmt.Sprintf("Error: %s", err.Error())
			}
			
			// Store tool result in variables
			if result.Variables == nil {
				result.Variables = make(map[string]interface{})
			}
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
		followUpResponse, err := s.providerService.CreateCompletion(stepCtx, completionReq)
		if err != nil {
			return result, fmt.Errorf("LLM follow-up request failed: %w", err)
		}
		
		completionResp = followUpResponse
		followUpsUsed++
		
		logging.Debug("Follow-up response #%d: %s", followUpsUsed, completionResp.Response)
	}
	
	// Set result data (final response after all tool calls)
	result.Output = completionResp.Response
	result.Provider = string(s.providerService.GetProviderType())
	result.Usage = completionResp.Usage
	result.ToolCalls = completionResp.ToolCalls // This will be the final set of tool calls
	result.ExecutionTime = time.Since(startTime)
	result.Status = domain.StepStatusCompleted
	
	// Add debug logging to understand what's happening
	logging.Debug("Step %d final output: %s", step.Step, result.Output)
	logging.Debug("Step %d final status: %s", step.Step, result.Status)
	
	// Process output configuration
	if step.Output != nil {
		err = s.processStepOutput(result, step.Output)
		if err != nil {
			return result, fmt.Errorf("output processing failed: %w", err)
		}
	}
	
	return result, nil
}

// ValidateWorkflow validates a workflow template
func (s *Service) ValidateWorkflow(template *domain.WorkflowTemplate) error {
	return template.ValidateWorkflowTemplate()
}

// getWorkflowTemplate retrieves the workflow template for a step
func (s *Service) getWorkflowTemplate(step *domain.WorkflowStep) *domain.WorkflowTemplate {
	// Find the template that contains this step
	for _, template := range s.config.Templates {
		for _, templateStep := range template.Steps {
			if templateStep.Step == step.Step && templateStep.Name == step.Name {
				return template
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
	
	// Add request variables (override template variables)
	if req.Variables != nil {
		for key, value := range req.Variables {
			variables[key] = value
		}
	}
	
	// Add input data if provided - support both {{stdin}} and {{input_data}} for flexibility
	if req.InputData != "" {
		variables["input_data"] = req.InputData
		variables["stdin"] = req.InputData  // Cleaner alias for function app use cases
	}
}

// getStepTools retrieves available tools for a workflow step following MCP principles
func (s *Service) getStepTools(step *domain.WorkflowStep) ([]domain.Tool, error) {
	// Get all available tools from servers
	allTools, err := s.serverManager.GetAvailableTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tools: %w", err)
	}
	
	// If no specific tools required, provide ALL tools from connected servers
	// This follows MCP principles of automatic tool discovery and advertising
	if len(step.ToolsRequired) == 0 {
		logging.Debug("No specific tools required, providing all %d available tools from connected servers", len(allTools))
		return allTools, nil
	}
	
	// If specific tools are required, filter to only those tools
	// This is for cases where you want to restrict tool access for security/performance
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
		
		return nil, domain.NewDomainError(domain.ErrCodeToolNotFound, 
			fmt.Sprintf("required tools not available: %s", strings.Join(missing, ", ")))
	}
	
	logging.Debug("Providing %d specific required tools", len(stepTools))
	return stepTools, nil
}

// executeWithRetry executes a completion request with retry policy
func (s *Service) executeWithRetry(ctx context.Context, req *domain.CompletionRequest, retryPolicy *domain.WorkflowRetryPolicy) (*domain.CompletionResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with backoff
			delay, err := retryPolicy.GetRetryDelay()
			if err != nil {
				return nil, err
			}
			
			if retryPolicy.BackoffFactor > 1.0 {
				delay = time.Duration(float64(delay) * retryPolicy.BackoffFactor * float64(attempt))
			}
			
			logging.Debug("Retrying completion request (attempt %d/%d) after %v", 
				attempt+1, retryPolicy.MaxRetries+1, delay)
			
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		
		resp, err := s.providerService.CreateCompletion(ctx, req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !s.isRetryableError(err) {
			break
		}
	}
	
	return nil, fmt.Errorf("completion failed after %d retries: %w", retryPolicy.MaxRetries, lastErr)
}

// isRetryableError determines if an error is retryable
func (s *Service) isRetryableError(err error) bool {
	// This is a simplified implementation
	// In production, you'd want more sophisticated error classification
	errStr := err.Error()
	
	// Network-related errors are typically retryable
	retryableErrors := []string{
		"timeout",
		"connection",
		"network",
		"rate limit",
		"429",
		"500",
		"502",
		"503",
		"504",
	}
	
	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}
	
	return false
}

// processStepOutput processes step output according to configuration
func (s *Service) processStepOutput(result *domain.WorkflowStepResult, outputConfig *domain.WorkflowOutputConfig) error {
	// Format output according to configuration
	switch outputConfig.Format {
	case "json":
		// Try to parse and reformat as JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(result.Output), &jsonData); err == nil {
			formatted, err := json.MarshalIndent(jsonData, "", "  ")
			if err == nil {
				result.Output = string(formatted)
			}
		}
	case "structured":
		// Create structured output with metadata
		structured := map[string]interface{}{
			"content": result.Output,
		}
		
		if outputConfig.IncludeMetadata {
			structured["metadata"] = map[string]interface{}{
				"step":           result.Step,
				"name":           result.Name,
				"provider":       result.Provider,
				"execution_time": result.ExecutionTime.String(),
				"usage":          result.Usage,
			}
		}
		
		formatted, err := json.MarshalIndent(structured, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format structured output: %w", err)
		}
		result.Output = string(formatted)
	}
	
	// Save to variable if specified
	if outputConfig.SaveToVariable != "" {
		if result.Variables == nil {
			result.Variables = make(map[string]interface{})
		}
		result.Variables[outputConfig.SaveToVariable] = result.Output
	}
	
	return nil
}

// ExecuteWorkflow is a convenience method that creates a workflow request and executes it
func (s *Service) ExecuteWorkflow(ctx context.Context, templateName string, inputData string) (*domain.WorkflowResponse, error) {
	req := &domain.WorkflowRequest{
		TemplateName: templateName,
		InputData:    inputData,
		ExecutionID:  generateExecutionID(),
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
	template, exists := s.config.GetWorkflowTemplate(templateName)
	if !exists {
		return nil, domain.NewDomainError(domain.ErrCodeRequestInvalid, 
			fmt.Sprintf("workflow template '%s' not found", templateName))
	}
	
	return template, nil
}

// ValidateTemplateConfiguration validates all workflow templates in the configuration
func (s *Service) ValidateTemplateConfiguration() error {
	return s.config.ValidateWorkflowTemplates()
}
