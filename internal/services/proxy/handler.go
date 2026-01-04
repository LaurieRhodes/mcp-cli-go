package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// ToolHandler handles HTTP requests for a specific tool
type ToolHandler struct {
	tool          *runas.ToolExposure
	appConfig     *config.ApplicationConfig
	proxyServer   *ProxyServer
}

// NewToolHandler creates a new tool handler
func NewToolHandler(tool *runas.ToolExposure, appConfig *config.ApplicationConfig, proxyServer *ProxyServer) (*ToolHandler, error) {
	// Only validate template existence for workflow tools (not MCP server tools)
	if tool.Template != "" {
		_, existsV1 := appConfig.Templates[tool.Template]
		_, existsV2 := appConfig.TemplatesV2[tool.Template]
		
		if !existsV1 && !existsV2 {
			return nil, fmt.Errorf("template not found: %s", tool.Template)
		}
	}
	
	return &ToolHandler{
		tool:        tool,
		appConfig:   appConfig,
		proxyServer: proxyServer,
	}, nil
}

// Handle processes an HTTP request for this tool
func (h *ToolHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		logging.Warn("Failed to parse request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate request against input schema
	if err := h.validateInput(requestData); err != nil {
		logging.Warn("Input validation failed: %v", err)
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}
	
	// Map tool inputs to template variables
	templateVars := h.mapInputsToTemplate(requestData)
	
	// Execute the workflow template
	result, err := h.executeTemplate(templateVars)
	if err != nil {
		logging.Error("Template execution failed: %v", err)
		http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
		"tool":    h.tool.Name,
	})
}

// validateInput validates the request data against the tool's input schema
func (h *ToolHandler) validateInput(data map[string]interface{}) error {
	// Get required fields from schema
	required, ok := h.tool.InputSchema["required"].([]interface{})
	if ok {
		for _, field := range required {
			fieldName := field.(string)
			if _, exists := data[fieldName]; !exists {
				return fmt.Errorf("missing required field: %s", fieldName)
			}
		}
	}
	
	// Basic type validation
	properties, ok := h.tool.InputSchema["properties"].(map[string]interface{})
	if ok {
		for fieldName, fieldValue := range data {
			propSchema, exists := properties[fieldName]
			if !exists {
				// Unknown field - log but don't fail
				logging.Debug("Unknown field in request: %s", fieldName)
				continue
			}
			
			// Validate type if specified
			if err := h.validateFieldType(fieldName, fieldValue, propSchema); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateFieldType validates a field's type against its schema
func (h *ToolHandler) validateFieldType(fieldName string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil // Can't validate without proper schema
	}
	
	expectedType, ok := schemaMap["type"].(string)
	if !ok {
		return nil // No type specified
	}
	
	// Basic type checking
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s must be a string", fieldName)
		}
	case "number", "integer":
		switch value.(type) {
		case float64, int, int64:
			// Valid number types
		default:
			return fmt.Errorf("field %s must be a number", fieldName)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s must be a boolean", fieldName)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field %s must be an object", fieldName)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("field %s must be an array", fieldName)
		}
	}
	
	return nil
}

// mapInputsToTemplate maps tool inputs to template variables using input_mapping
func (h *ToolHandler) mapInputsToTemplate(data map[string]interface{}) map[string]string {
	templateVars := make(map[string]string)
	
	if h.tool.InputMapping == nil || len(h.tool.InputMapping) == 0 {
		// No explicit mapping - use direct mapping
		for key, value := range data {
			templateVars[key] = fmt.Sprintf("%v", value)
		}
		return templateVars
	}
	
	// Apply input mapping
	for toolParam, templateVar := range h.tool.InputMapping {
		// Template vars can be like "{{variable_name}}"
		// Extract the variable name
		varName := templateVar
		if len(templateVar) > 4 && templateVar[:2] == "{{" && templateVar[len(templateVar)-2:] == "}}" {
			varName = templateVar[2 : len(templateVar)-2]
		}
		
		// Get value from request data
		if value, exists := data[toolParam]; exists {
			// Convert to string based on type
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case map[string]interface{}, []interface{}:
				// For complex types, marshal to JSON
				jsonBytes, _ := json.Marshal(v)
				strValue = string(jsonBytes)
			default:
				strValue = fmt.Sprintf("%v", v)
			}
			
			templateVars[varName] = strValue
		}
	}
	
	return templateVars
}

// executeTemplate executes the workflow template with the given variables
func (h *ToolHandler) executeTemplate(vars map[string]string) (string, error) {
	// Check if this is an MCP server tool (not a workflow template)
	if h.tool.MCPServer != "" && h.tool.MCPTool != "" {
		return h.executeMCPTool(vars)
	}
	
	// For proxy mode with workflow templates, use the workflow service if available
	if h.proxyServer.workflowService != nil {
		// Prepare input data - use the first variable or combine all
		var inputData string
		if len(vars) == 1 {
			for _, v := range vars {
				inputData = v
				break
			}
		} else if len(vars) > 1 {
			// Multiple variables - convert to JSON
			jsonBytes, _ := json.Marshal(vars)
			inputData = string(jsonBytes)
		}
		
		// Execute workflow using the workflow service
		logging.Info("Executing workflow %s for tool %s", h.tool.Template, h.tool.Name)
		ctx := context.Background()
		result, err := h.proxyServer.workflowService.ExecuteWorkflow(ctx, h.tool.Template, inputData)
		if err != nil {
			return "", fmt.Errorf("workflow execution failed: %w", err)
		}
		
		// Extract final output
		if result.FinalOutput != "" {
			return result.FinalOutput, nil
		}
		
		// Fallback to last step output
		if len(result.StepResults) > 0 {
			lastStep := result.StepResults[len(result.StepResults)-1]
			if lastStep.Output != "" {
				return lastStep.Output, nil
			}
		}
		
		return "", fmt.Errorf("no output generated from workflow")
	}
	
	// Fallback: If no workflow service, return error
	return "", fmt.Errorf("workflow service not available in proxy mode")
}

// executeMCPTool executes an MCP server tool
func (h *ToolHandler) executeMCPTool(vars map[string]string) (string, error) {
	// Find the MCP server connection
	var serverConn *host.ServerConnection
	for _, conn := range h.proxyServer.mcpServers {
		if conn.Name == h.tool.MCPServer {
			serverConn = conn
			break
		}
	}
	
	if serverConn == nil {
		return "", fmt.Errorf("MCP server not connected: %s", h.tool.MCPServer)
	}
	
	// Convert vars to arguments map
	arguments := make(map[string]interface{})
	for key, value := range vars {
		// Try to parse as JSON for complex types
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			arguments[key] = jsonValue
		} else {
			// Use as string if not valid JSON
			arguments[key] = value
		}
	}
	
	logging.Info("Executing MCP tool %s on server %s with args: %v", h.tool.MCPTool, h.tool.MCPServer, arguments)
	
	// Call the MCP tool
	result, err := tools.SendToolsCall(serverConn.Client, h.tool.MCPTool, arguments)
	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed: %w", err)
	}
	
	// Debug: log the full result
	resultJSON, _ := json.Marshal(result)
	logging.Debug("MCP tool result: %s", string(resultJSON))
	
	// Check for errors in result
	if result.IsError {
		errorMsg := result.Error
		
		// If Error field is empty but Content has error info, extract it
		if errorMsg == "" {
			if content, ok := result.Content.([]interface{}); ok && len(content) > 0 {
				if contentItem, ok := content[0].(map[string]interface{}); ok {
					if text, ok := contentItem["text"].(string); ok {
						errorMsg = text
					}
				}
			}
			
			// Still empty? Use generic message
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("tool returned error (Content: %v)", result.Content)
			}
		}
		
		logging.Error("MCP tool returned error: %s", errorMsg)
		return "", fmt.Errorf("MCP tool error: %s", errorMsg)
	}
	
	// Format the response - Content can be various types
	output := formatToolContent(result.Content)
	logging.Debug("MCP tool %s returned: %s", h.tool.MCPTool, output)
	
	return output, nil
}

// formatToolContent formats the content from an MCP tool result
// Content can be: string, array of content items, or other structures
func formatToolContent(content interface{}) string {
	if content == nil {
		return ""
	}
	
	// If it's already a string, return it
	if str, ok := content.(string); ok {
		return str
	}
	
	// If it's an array, process each item
	if arr, ok := content.([]interface{}); ok {
		var result strings.Builder
		for i, item := range arr {
			if i > 0 {
				result.WriteString("\n")
			}
			
			// Each item might be a map with type/text/etc
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemType, _ := itemMap["type"].(string)
				switch itemType {
				case "text":
					if text, ok := itemMap["text"].(string); ok {
						result.WriteString(text)
					}
				case "image":
					mimeType, _ := itemMap["mimeType"].(string)
					result.WriteString(fmt.Sprintf("[Image: %s]", mimeType))
				case "resource":
					uri, _ := itemMap["uri"].(string)
					result.WriteString(fmt.Sprintf("[Resource: %s]", uri))
				default:
					// Unknown type, try to get text
					if text, ok := itemMap["text"].(string); ok {
						result.WriteString(text)
					} else {
						// Marshal to JSON as fallback
						jsonBytes, _ := json.Marshal(item)
						result.WriteString(string(jsonBytes))
					}
				}
			} else {
				// Not a map, convert to string
				result.WriteString(fmt.Sprintf("%v", item))
			}
		}
		return result.String()
	}
	
	// For other types, try to marshal to JSON
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Sprintf("%v", content)
	}
	return string(jsonBytes)
}
