package proxy

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/mcp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/runas"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	workflowservice "github.com/LaurieRhodes/mcp-cli-go/internal/services/workflow"
)

// ToolHandler handles HTTP requests for a specific MCP tool
type ToolHandler struct {
	tool        *runas.ToolExposure
	proxyServer *ProxyServer
}

// NewToolHandler creates a new tool handler
func NewToolHandler(tool *runas.ToolExposure, proxyServer *ProxyServer) *ToolHandler {
	return &ToolHandler{
		tool:        tool,
		proxyServer: proxyServer,
	}
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

	// Convert to string map for template execution
	vars := make(map[string]string)
	for k, v := range requestData {
		vars[k] = fmt.Sprintf("%v", v)
	}

	// Execute the template/tool
	result, err := h.executeTemplate(vars)
	if err != nil {
		logging.Warn("Template execution failed: %v", err)
		http.Error(w, fmt.Sprintf("Execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result as JSON
	response := map[string]interface{}{
		"result": result,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeTemplate executes the workflow template with the given variables
func (h *ToolHandler) executeTemplate(vars map[string]string) (string, error) {
	// Check if this is an MCP server tool (not a workflow template)
	if h.tool.MCPServer != "" && h.tool.MCPTool != "" {
		return h.executeMCPTool(vars)
	}

	// For workflow templates, execute using the orchestrator
	workflow, exists := h.proxyServer.appConfig.Workflows[h.tool.Template]
	if !exists {
		return "", fmt.Errorf("workflow not found: %s", h.tool.Template)
	}

	// Prepare input data - use first variable or combine all
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

	// Create logger
	logger := workflowservice.NewLogger(workflow.Execution.Logging)
	
	// Create orchestrator
	orchestrator := workflowservice.NewOrchestrator(workflow, logger)
	orchestrator.SetAppConfigForWorkflows(h.proxyServer.appConfig)
	
	// Execute workflow
	ctx := context.Background()
	err := orchestrator.Execute(ctx, inputData)
	if err != nil {
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}
	
	// Get result from last step
	result := ""
	if len(workflow.Steps) > 0 {
		lastStepName := workflow.Steps[len(workflow.Steps)-1].Name
		if output, ok := orchestrator.GetStepResult(lastStepName); ok {
			result = output
		}
	}
	
	if result == "" {
		return fmt.Sprintf("Workflow '%s' completed but produced no output", workflow.Name), nil
	}
	
	return result, nil
}

// executeMCPTool executes an MCP tool call by proxying to the MCP server
func (h *ToolHandler) executeMCPTool(vars map[string]string) (string, error) {
	// Find the MCP server connection by name
	var server *host.ServerConnection
	for _, conn := range h.proxyServer.mcpServers {
		if conn.Name == h.tool.MCPServer {
			server = conn
			break
		}
	}
	
	if server == nil {
		return "", fmt.Errorf("MCP server not found: %s", h.tool.MCPServer)
	}

	// Convert vars to arguments map
	args := make(map[string]interface{})
	for k, v := range vars {
		args[k] = v
	}

	// Call the MCP tool
	result, err := tools.SendToolsCall(server.Client, h.tool.MCPTool, args)
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Check for error
	if result.IsError {
		return "", fmt.Errorf("MCP tool call failed: %s", result.Error)
	}

	// Extract text content from result
	errorDetector := mcp.NewErrorDetector()
	text := errorDetector.ExtractTextFromContent(result.Content)
	
	if text == "" {
		return fmt.Sprintf("Tool completed: %v", result), nil
	}
	
	return text, nil
}
