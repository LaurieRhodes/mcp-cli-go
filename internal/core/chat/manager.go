
package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"time"
	appChat "github.com/LaurieRhodes/mcp-cli-go/internal/app/chat"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
    "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config" 
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	mcplib "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/mcp"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// ChatManager manages the chat flow
type ChatManager struct {
	// LLM provider for chat completions (updated to use new domain interface)
	LLMProvider domain.LLMProvider
	
	// Server connections for tool execution (legacy)
	Connections []*host.ServerConnection
	
	// Server manager for tool execution (new, supports built-in skills)
	ServerManager domain.MCPServerManager
	
	// Enabled skills
	EnabledSkills []string
	
	// Chat context
	Context *ChatContext
	
	// User interface manager
	UI *UI
	
	// Whether to stream responses
	StreamResponses bool
	
	// Available tools cache
	toolsCache map[string][]tools.Tool
	
	// Last assistant message with tool calls
	lastAssistantMessageWithToolCalls domain.Message

	// Session logging (optional)
	sessionLogger *appChat.SessionLogger
	session       *appChat.Session
	providerName  string
	modelName     string
}

// NewChatManager creates a new chat manager
func NewChatManager(provider domain.LLMProvider, connections []*host.ServerConnection) *ChatManager {
	return NewChatManagerWithConfig(provider, connections, nil, "")
}

// NewChatManagerWithConfig creates a new chat manager with provider configuration
func NewChatManagerWithConfig(provider domain.LLMProvider, connections []*host.ServerConnection, providerConfig *config.ProviderConfig, model string) *ChatManager {
	systemPrompt := `You are a helpful assistant with access to tools. Use the tools when necessary to fulfill user requests.

IMPORTANT - Using Skills:
Skills provide specialized capabilities through code execution. There are two ways to use skills:

1. PASSIVE MODE - Load documentation and reference materials:
   Call the skill tool directly (e.g., 'docx', 'pdf', 'pptx', 'xlsx')
   Use this to learn about a skill's capabilities before using it.

2. ACTIVE MODE - Execute code to perform tasks:
   Call 'execute_skill_code' with skill_name parameter
   Use this to CREATE, MODIFY, PROCESS, or GENERATE anything.

✅ CORRECT examples:
   - Create a document: execute_skill_code with skill_name='docx'
   - Generate a PDF: execute_skill_code with skill_name='pdf'
   - Process data: execute_skill_code with appropriate skill
   - Run analysis: execute_skill_code with appropriate skill

When writing code, save output files to /outputs/ directory:
   output.save('/outputs/result.docx')  ✅ CORRECT
   output.save('/home/result.docx')     ❌ WRONG - will be lost`
	return &ChatManager{
		LLMProvider:     provider,
		Connections:     connections,
		Context:         NewChatContextWithProvider(systemPrompt, model, providerConfig),
		UI:              NewUI(),
		StreamResponses: true,
		toolsCache:      make(map[string][]tools.Tool),
		modelName:       model,
	}
}

// NewChatManagerWithUI creates a new chat manager with a provided UI
func NewChatManagerWithUI(provider domain.LLMProvider, connections []*host.ServerConnection, ui *UI) *ChatManager {
	return NewChatManagerWithConfigAndUI(provider, connections, nil, "", ui)
}

// NewChatManagerWithConfigAndUI creates a new chat manager with provider configuration and provided UI
func NewChatManagerWithConfigAndUI(provider domain.LLMProvider, connections []*host.ServerConnection, providerConfig *config.ProviderConfig, model string, ui *UI) *ChatManager {
	systemPrompt := "You are a helpful assistant with access to tools. Use the tools when necessary to fulfill user requests."
	return &ChatManager{
		LLMProvider:     provider,
		Connections:     connections,
		Context:         NewChatContextWithProvider(systemPrompt, model, providerConfig),
		UI:              ui,
		StreamResponses: true,
		toolsCache:      make(map[string][]tools.Tool),
		modelName:       model,
	}
}


// NewChatManagerWithServerManagerAndUI creates a new chat manager with server manager (supports built-in skills)
func NewChatManagerWithServerManagerAndUI(provider domain.LLMProvider, serverManager domain.MCPServerManager, providerConfig *config.ProviderConfig, model string, ui *UI) *ChatManager {
	systemPrompt := `You are a helpful assistant with access to tools. Use the tools when necessary to fulfill user requests.

IMPORTANT - Using Skills:
Skills provide specialized capabilities through code execution. There are two ways to use skills:

1. PASSIVE MODE - Load documentation and reference materials:
   Call the skill tool directly (e.g., 'docx', 'pdf', 'pptx', 'xlsx')
   Use this to learn about a skill's capabilities before using it.

2. ACTIVE MODE - Execute code to perform tasks:
   Call 'execute_skill_code' with skill_name parameter
   Use this to CREATE, MODIFY, PROCESS, or GENERATE anything.

✅ CORRECT examples:
   - Create a document: execute_skill_code with skill_name='docx'
   - Generate a PDF: execute_skill_code with skill_name='pdf'
   - Process data: execute_skill_code with appropriate skill
   - Run analysis: execute_skill_code with appropriate skill

When writing code, save output files to /outputs/ directory:
   output.save('/outputs/result.docx')  ✅ CORRECT
   output.save('/home/result.docx')     ❌ WRONG - will be lost`
	
	return &ChatManager{
		LLMProvider:     provider,
		ServerManager:   serverManager,
		Context:         NewChatContextWithProvider(systemPrompt, model, providerConfig),
		UI:              ui,
		StreamResponses: true,
		toolsCache:      make(map[string][]tools.Tool),
		modelName:       model,
	}
}
// ProcessUserMessage processes a user message and returns the response
func (m *ChatManager) ProcessUserMessage(userInput string) error {
	// Add user message to context
	userMessage := domain.Message{
		Role:    "user",
		Content: userInput,
	}
	m.Context.AddMessage(userMessage)
	// Add to session if logging enabled
	if m.session != nil {
		m.session.AddMessage(convertDomainMessage(userMessage))
	}
	
	// Get available tools for the LLM
	logging.Info("Fetching available tools for LLM")
	llmTools, err := m.GetAvailableTools()
	if err != nil {
		m.UI.PrintError("Failed to get available tools: %v", err)
		// Continue without tools as a fallback
		llmTools = []domain.Tool{}
	}
	logging.Info("Successfully fetched %d tools for LLM", len(llmTools))
	
	// Get messages for the LLM
	messages := m.Context.GetMessagesForLLM()
	
	// Show indicator that we're working
	m.UI.PrintSystem("Thinking...")
	
	// Create completion request
	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		Tools:        llmTools,
		SystemPrompt: "", // Already included in messages
		Temperature:  0.7, // Default temperature for chat
		Stream:       m.StreamResponses,
	}
	
	var response *domain.CompletionResponse
	
	// Use streaming if supported and enabled
	if m.StreamResponses {
		// Start the streaming response UI
		m.UI.StartStreamingResponse()
		
		// Determine provider type for more accurate logging
		providerType := m.LLMProvider.GetProviderType()
		logging.Info("Starting streaming completion with %s", providerType)
		
		response, err = m.LLMProvider.StreamCompletion(context.Background(), completionReq, &streamingWriter{
			onChunk: func(chunk string) error {
				m.UI.StreamAssistantResponse(chunk)
				return nil
			},
		})
		
		// End the streaming response UI
		m.UI.EndStreamingResponse()
	} else {
		// Fallback to non-streaming
		logging.Info("Starting non-streaming completion")
		response, err = m.LLMProvider.CreateCompletion(context.Background(), completionReq)
		
		// Print the full response
		if err == nil && response != nil {
			m.UI.PrintAssistantResponse(response.Response)
		}
	}
	
	if err != nil {
		return fmt.Errorf("LLM completion error: %w", err)
	}
	
	// Add assistant message to context
	if response != nil {
		assistantMessage := domain.Message{
			Role:      "assistant",
			Content:   response.Response,
			ToolCalls: response.ToolCalls,
		}
		m.Context.AddMessage(assistantMessage)
		// Add to session if logging enabled
		if m.session != nil {
			m.session.AddMessage(convertDomainMessage(assistantMessage))
		}
		
		// Save this for tool responses if it has tool calls
		if len(response.ToolCalls) > 0 {
			m.lastAssistantMessageWithToolCalls = assistantMessage
		}
		
		// Handle tool calls if present
		if len(response.ToolCalls) > 0 {
			m.UI.PrintSystem("Executing tool calls...")
			err = m.HandleToolCalls(response.ToolCalls)
			if err != nil {
				m.UI.PrintError("Error executing tool calls: %v", err)
			}
			
			// ALWAYS get a follow-up response after tool execution
			// The LLM needs to synthesize the tool results into a final answer
			err = m.ProcessAfterToolExecution(userMessage.Content)
			if err != nil {
				m.UI.PrintError("Error getting follow-up response: %v", err)
			}
		}
	}
	
	return nil
}

// streamingWriter implements io.Writer for streaming responses
type streamingWriter struct {
	onChunk func(string) error
}

func (w *streamingWriter) Write(p []byte) (n int, err error) {
	if w.onChunk != nil {
		err = w.onChunk(string(p))
	}
	return len(p), err
}

// ProcessAfterToolExecution gets a follow-up response after tool execution
func (m *ChatManager) ProcessAfterToolExecution(userQuery string) error {
	// Get messages for the LLM - this will include the tool results now
	messages := m.Context.GetMessagesForLLM()
	
	// Get available tools for the LLM (might need more tools)
	llmTools, err := m.GetAvailableTools()
	if err != nil {
		llmTools = []domain.Tool{} // Continue without tools as fallback
	}
	
	// Show indicator that we're working on a response
	m.UI.PrintSystem("Generating response based on tool results...")
	
	// Create completion request
	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		Tools:        llmTools,
		SystemPrompt: "", // Already included in messages
		Temperature:  0.7, // Default temperature for chat
		Stream:       m.StreamResponses,
	}
	
	var response *domain.CompletionResponse
	
	// Use streaming if supported and enabled
	if m.StreamResponses {
		// Start the streaming response UI
		m.UI.StartStreamingResponse()
		
		// Determine provider type for more accurate logging
		providerType := m.LLMProvider.GetProviderType()
		logging.Info("Starting follow-up streaming completion with %s", providerType)
		
		response, err = m.LLMProvider.StreamCompletion(context.Background(), completionReq, &streamingWriter{
			onChunk: func(chunk string) error {
				m.UI.StreamAssistantResponse(chunk)
				return nil
			},
		})
		
		// End the streaming response UI
		m.UI.EndStreamingResponse()
	} else {
		// Fallback to non-streaming
		logging.Info("Starting follow-up non-streaming completion")
		response, err = m.LLMProvider.CreateCompletion(context.Background(), completionReq)
		
		// Print the full response
		if err == nil && response != nil {
			m.UI.PrintAssistantResponse(response.Response)
		}
	}
	
	if err != nil {
		return fmt.Errorf("follow-up completion error: %w", err)
	}
	
	// Add assistant message to context
	if response != nil {
		assistantMessage := domain.Message{
			Role:      "assistant",
			Content:   response.Response,
			ToolCalls: response.ToolCalls,
		}
		m.Context.AddMessage(assistantMessage)
		// Add to session if logging enabled
		if m.session != nil {
			m.session.AddMessage(convertDomainMessage(assistantMessage))
		}
		
		// Save this for tool responses if it has tool calls
		if len(response.ToolCalls) > 0 {
			m.lastAssistantMessageWithToolCalls = assistantMessage
		}
		
		// Handle any additional tool calls if present
		if len(response.ToolCalls) > 0 {
			m.UI.PrintSystem("Executing additional tool calls...")
			err = m.HandleToolCalls(response.ToolCalls)
			if err != nil {
				m.UI.PrintError("Error executing additional tool calls: %v", err)
				return err
			}
			
			// Recursively get final response after additional tool execution
			logging.Debug("Requesting final response after additional tool calls")
			return m.ProcessAfterToolExecution(userQuery)
		}
	}
	
	return nil
}

// HandleToolCalls executes tool calls and adds results to the context
func (m *ChatManager) HandleToolCalls(toolCalls []domain.ToolCall) error {
	for _, toolCall := range toolCalls {
		// Execute the tool call
		logging.Info("Executing tool call: %s", toolCall.Function.Name)
		
		// Log the arguments for debugging
		argString := string(toolCall.Function.Arguments)
		if argString == "" {
			logging.Warn("Tool call has empty arguments")
		} else {
			logging.Debug("Tool call arguments: %s", argString)
		}
		
		// Add default arguments if none provided
		if argString == "" || argString == "{}" || argString == "null" {
			// Try to provide default arguments based on the tool
			defaultArgs := m.getDefaultToolArguments(toolCall.Function.Name)
			if defaultArgs != "" {
				logging.Info("Using default arguments: %s", defaultArgs)
				toolCall.Function.Arguments = []byte(defaultArgs)
			}
		}
		
		// Execute the tool
		result, err := m.ExecuteToolCall(toolCall)
		
		// Add tool call to history
		m.Context.AddToolCall(toolCall, result, err)
		
		// Prepare tool result content (use error message if execution failed)
		var toolResultContent string
		if err != nil {
			m.UI.PrintError("Tool execution failed: %v", err)
			toolResultContent = fmt.Sprintf("Error: %v", err)
		} else {
			toolResultContent = result
		}
		
		// CRITICAL: Always add tool result message, even for errors
		// DeepSeek and other OpenAI-compatible APIs require a tool result for every tool_call_id
		toolResultMessage := domain.Message{
			Role:        "tool",
			Content:     toolResultContent,
			ToolCallID:  toolCall.ID,
		}
		m.Context.AddMessage(toolResultMessage)
		
		// Don't print raw tool results in chat mode - let the LLM synthesize them
		// The user will see the LLM's response after it processes the tool results
		// m.UI.PrintToolResult(result)  // Commented out to avoid showing raw tool output
	}
	
	return nil
}

// getDefaultToolArguments provides sensible defaults for common tools
func (m *ChatManager) getDefaultToolArguments(toolName string) string {
	// For List Directory, default to project root
	if strings.Contains(toolName, "list_directory") {
		return `{"path": "D:/Github/mcp-cli-go"}`
	}
	
	// For List Allowed Directories, empty args are fine
	if strings.Contains(toolName, "list_allowed_directories") {
		return `{}`
	}
	
	// For other tools, use an empty object
	return `{}`
}

// ExecuteToolCall executes a single tool call and returns the result
func (m *ChatManager) ExecuteToolCall(toolCall domain.ToolCall) (string, error) {
	// ARCHITECTURAL FIX: Use ServerManager if available (supports built-in skills)
	if m.ServerManager != nil {
		return m.executeToolCallWithServerManager(toolCall)
	}
	
	// Fall back to legacy Connections-based execution
	return m.executeToolCallWithConnections(toolCall)
}

// executeToolCallWithServerManager executes a tool call using the server manager
func (m *ChatManager) executeToolCallWithServerManager(toolCall domain.ToolCall) (string, error) {
	// Parse arguments
	var args map[string]interface{}
	err := json.Unmarshal(toolCall.Function.Arguments, &args)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	
	// Show what we're doing
	m.UI.PrintToolExecution(toolCall.Function.Name, "server-manager")
	
	// Execute tool using server manager
	logging.Debug("Executing tool %s using server manager", toolCall.Function.Name)
	result, err := m.ServerManager.ExecuteTool(context.Background(), toolCall.Function.Name, args)
	if err != nil {
		return "", fmt.Errorf("tool execution error: %w", err)
	}
	
	return result, nil
}

// executeToolCallWithConnections executes a tool call using legacy connections
func (m *ChatManager) executeToolCallWithConnections(toolCall domain.ToolCall) (string, error) {
	// Parse arguments
	var args map[string]interface{}
	err := json.Unmarshal(toolCall.Function.Arguments, &args)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	
	// Parse the function name to extract server and tool
	toolName := toolCall.Function.Name
	serverName := ""
	
	// Handle both formats: "server_name_tool_name" or "server-name-tool-name"
	for _, conn := range m.Connections {
		// Try different separators and variations
		if strings.HasPrefix(toolName, conn.Name+"_") {
			// Format: server_tool
			serverName = conn.Name
			toolName = strings.TrimPrefix(toolName, conn.Name+"_")
			break
		} else if strings.HasPrefix(toolName, conn.Name+"-") {
			// Format: server-tool
			serverName = conn.Name
			toolName = strings.TrimPrefix(toolName, conn.Name+"-")
			break
		} else if strings.HasPrefix(toolName, strings.ReplaceAll(conn.Name, "-", "_")+"_") {
			// Format: server_name_tool (when server has hyphen)
			serverName = conn.Name
			toolName = strings.TrimPrefix(toolName, strings.ReplaceAll(conn.Name, "-", "_")+"_")
			break
		}
	}
	
	// If we still don't have a server name, try to find a tool with this name on any server
	if serverName == "" {
		for _, conn := range m.Connections {
			serverTools, err := m.getServerTools(conn)
			if err != nil {
				logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
				continue
			}
			
			for _, tool := range serverTools {
				if tool.Name == toolName {
					serverName = conn.Name
					break
				}
			}
			
			if serverName != "" {
				break
			}
		}
	}
	
	// If we still don't have a server name, use the first available server
	if serverName == "" && len(m.Connections) > 0 {
		serverName = m.Connections[0].Name
		logging.Warn("Could not determine server for tool %s, using default server %s", toolName, serverName)
	}
	
	// Find the server connection
	var serverConn *host.ServerConnection
	for _, conn := range m.Connections {
		if conn.Name == serverName {
			serverConn = conn
			break
		}
	}
	
	if serverConn == nil {
		return "", fmt.Errorf("server not found: %s", serverName)
	}
	
	// Show what we're doing
	m.UI.PrintToolExecution(toolName, serverName)
	
	// Execute the tool call using the tools package
	logging.Info("Calling tool %s on server %s", toolName, serverName)
	
	// Type assert to stdio client
	stdioClient := serverConn.GetStdioClient()
	if stdioClient == nil {
		return "", fmt.Errorf("server %s does not support stdio protocol", serverName)
	}
	
	result, err := tools.SendToolsCall(stdioClient, stdioClient.GetDispatcher(), toolName, args)
	if err != nil {
		return "", fmt.Errorf("tool execution error: %w", err)
	}
	
	// Enhanced error detection - check both top-level and nested error formats
	// This follows MCP spec and handles legacy server implementations
	errorDetector := mcplib.NewErrorDetector()
	
	// Log detailed error info for debugging
	if logging.GetDefaultLevel() <= logging.DEBUG {
		errorDetector.LogErrorDetails(result)
	}
	
	// Check for errors using enhanced detection
	if errorDetector.IsMCPError(result) {
		// Try to get detailed error message
		if errMsg, hasMsg := errorDetector.GetErrorMessage(result); hasMsg {
			return "", fmt.Errorf("tool execution failed: %s", errMsg)
		}
		// Fallback to generic error
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}
	
	// Convert result to string if needed
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		resultStr = content
	default:
		// For domain providers, we need to handle structured responses differently
		// Check provider type to handle specific formatting
		providerType := m.LLMProvider.GetProviderType()
		if providerType == domain.ProviderAnthropic {
			// Try to extract proper text from the Anthropic response
			resultStr = m.formatAnthropicToolResult(content)
		} else {
			// Convert to JSON string as before for other clients
			resultBytes, err := json.MarshalIndent(content, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
			}
			resultStr = string(resultBytes)
		}
	}
	
	return resultStr, nil
}

// formatAnthropicToolResult formats tool results specifically for Anthropic
func (m *ChatManager) formatAnthropicToolResult(content interface{}) string {
	// Try to convert to JSON first
	resultBytes, err := json.Marshal(content)
	if err != nil {
		logging.Error("Failed to marshal Anthropic result to JSON: %v", err)
		resultBytes, _ = json.MarshalIndent(content, "", "  ")
		return string(resultBytes)
	}
	
	// Try to extract text content from Anthropic response format
	var resultArr []map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultArr); err == nil {
		// This is a valid JSON array - check for Anthropic's format
		for _, item := range resultArr {
			// Check for text field which is the actual content
			if text, ok := item["text"].(string); ok {
				return text
			}
		}
	}
	
	// If we can't extract from array format, try the object format
	var resultObj map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultObj); err == nil {
		// Try Anthropic's message format with content array
		if content, ok := resultObj["content"].([]interface{}); ok {
			var sb strings.Builder
			for _, item := range content {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						sb.WriteString(text)
					}
				}
			}
			if sb.Len() > 0 {
				return sb.String()
			}
		}
		
		// Try simpler format where text might be directly in the object
		if text, ok := resultObj["text"].(string); ok {
			return text
		}
	}
	
	// If all else fails, return pretty JSON
	resultBytes, _ = json.MarshalIndent(content, "", "  ")
	return string(resultBytes)
}

// formatToolNameForOpenAI formats the tool name to be compatible with OpenAI's requirements
// OpenAI only accepts names with alphanumeric characters, underscores, and hyphens
func formatToolNameForOpenAI(serverName, toolName string) string {
	// Replace any dots, spaces or invalid characters with underscores
	serverName = strings.ReplaceAll(serverName, ".", "_")
	serverName = strings.ReplaceAll(serverName, " ", "_")
	serverName = strings.ReplaceAll(serverName, "-", "_")
	
	// Make sure the tool name is valid too
	toolName = strings.ReplaceAll(toolName, ".", "_")
	toolName = strings.ReplaceAll(toolName, " ", "_")
	
	// Combine with underscore
	return fmt.Sprintf("%s_%s", serverName, toolName)
}

// GetAvailableTools returns the tools available for the LLM
func (m *ChatManager) GetAvailableTools() ([]domain.Tool, error) {
	// ARCHITECTURAL FIX: Use ServerManager if available (supports built-in skills)
	if m.ServerManager != nil {
		logging.Debug("Getting tools from ServerManager (includes built-in skills)")
		return m.ServerManager.GetAvailableTools()
	}
	
	// Fall back to legacy Connections-based tool retrieval
	var llmTools []domain.Tool
	var anyErrors error
	
	for _, conn := range m.Connections {
		serverTools, err := m.getServerTools(conn)
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
			anyErrors = err
			continue
		}
		
		logging.Debug("Processing %d tools from server %s for LLM provider", len(serverTools), conn.Name)
		
		for _, tool := range serverTools {
			// Format the tool name to be compatible with OpenAI's requirements
			formattedName := formatToolNameForOpenAI(conn.Name, tool.Name)
			
			// Debug log the name transformation
			logging.Debug("Transforming tool name for LLM: %s.%s -> %s", conn.Name, tool.Name, formattedName)
			
			// CRITICAL: Pass schema directly without transformation (Gemini CLI approach)
			// This ensures Gemini and other providers receive the schema exactly as MCP server provided it
			
			// Create the tool with the formatted name using domain types
			llmTool := domain.Tool{
				Type: "function",
				Function: domain.ToolFunction{
					Name:        formattedName,
					Description: fmt.Sprintf("[%s] %s", conn.Name, tool.Description),
					Parameters:  tool.InputSchema, // Direct pass-through - no transformation
				},
			}
			
			// Enhanced logging for debugging Gemini tool calling issues
			if logging.GetDefaultLevel() <= logging.DEBUG {
				logging.Debug("=== Tool Registration for LLM ===")
				logging.Debug("  Original: %s.%s", conn.Name, tool.Name)
				logging.Debug("  Formatted: %s", formattedName)
				logging.Debug("  Description: %s", tool.Description)
				if schemaJSON, err := json.Marshal(tool.InputSchema); err == nil {
					logging.Debug("  Schema (passed as-is): %s", string(schemaJSON))
				}
				logging.Debug("=================================")
			}
			
			llmTools = append(llmTools, llmTool)
		}
	}
	
	if len(llmTools) == 0 && anyErrors != nil {
		return nil, fmt.Errorf("failed to get any tools: %w", anyErrors)
	}
	
	logging.Info("Registered %d total tools for LLM provider", len(llmTools))
	return llmTools, nil
}

// getServerTools gets the tools from a server, using cache if available
func (m *ChatManager) getServerTools(conn *host.ServerConnection) ([]tools.Tool, error) {
	// Check if we have the tools in cache
	if cachedTools, ok := m.toolsCache[conn.Name]; ok {
		return cachedTools, nil
	}
	
	// Get the tools from the server with retry
	var serverTools []tools.Tool
	var lastErr error
	
	// Create lenient schema validator
	schemaValidator := mcplib.NewLenientSchemaValidator()
	
	for retries := 0; retries < 3; retries++ {
		if retries > 0 {
			logging.Warn("Retrying tools list request for server %s (attempt %d/3)", conn.Name, retries+1)
			time.Sleep(time.Duration(retries) * time.Second)
		}
		
		logging.Info("Getting tools list from server %s", conn.Name)
		
		// Type assert to stdio client
		stdioClient := conn.GetStdioClient()
		if stdioClient == nil {
			lastErr = fmt.Errorf("server %s does not support stdio protocol", conn.Name)
			break
		}
		
		result, err := tools.SendToolsList(stdioClient, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to get tools from server %s: %w", conn.Name, err)
			logging.Error("%v", lastErr)
			continue
		}
		
		// Validate and log schemas with lenient validation
		validatedTools := make([]tools.Tool, 0, len(result.Tools))
		for _, tool := range result.Tools {
			// Validate schema (lenient - logs warnings but doesn't reject)
			if err := schemaValidator.ValidateSchema(tool.InputSchema); err != nil {
				// This is a catastrophic error (not just validation failure)
				logging.Error("Catastrophic error validating schema for tool %s.%s: %v", 
					conn.Name, tool.Name, err)
				continue // Skip this tool
			}
			
			// Log schema for debugging if in debug mode
			if logging.GetDefaultLevel() <= logging.DEBUG {
				schemaValidator.LogSchemaForDebugging(
					fmt.Sprintf("%s.%s", conn.Name, tool.Name),
					tool.InputSchema,
				)
			}
			
			// Accept the tool
			validatedTools = append(validatedTools, tool)
		}
		
		logging.Info("Validated %d/%d tools from server %s", 
			len(validatedTools), len(result.Tools), conn.Name)
		
		// Cache the validated tools
		m.toolsCache[conn.Name] = validatedTools
		serverTools = validatedTools
		
		logging.Info("Successfully got %d tools from server %s", len(serverTools), conn.Name)
		return serverTools, nil
	}
	
	return nil, lastErr
}

// discoverAvailableSkills discovers which skill tools are available from connected servers
func (m *ChatManager) discoverAvailableSkills() []string {
	var skillNames []string
	skillsFound := make(map[string]bool) // Track unique skills
	
	// If EnabledSkills is set, use that as the filter
	var enabledSkillsMap map[string]bool
	if len(m.EnabledSkills) > 0 {
		fmt.Printf("[DEBUG] EnabledSkills filter: %v\n", m.EnabledSkills)
		enabledSkillsMap = make(map[string]bool)
		for _, skillName := range m.EnabledSkills {
			// Support both hyphenated skill names and underscored tool names
			enabledSkillsMap[skillName] = true
			// Also add the converted form
			if strings.Contains(skillName, "-") {
				enabledSkillsMap[strings.ReplaceAll(skillName, "-", "_")] = true
			} else if strings.Contains(skillName, "_") {
				enabledSkillsMap[strings.ReplaceAll(skillName, "_", "-")] = true
			}
		}
		fmt.Printf("[DEBUG] EnabledSkills map after conversion: %v\n", enabledSkillsMap)
	}
	
	// ARCHITECTURAL FIX: Use ServerManager if available
	if m.ServerManager != nil {
		// Get all tools from server manager
		tools, err := m.ServerManager.GetAvailableTools()
		if err != nil {
			logging.Debug("Could not get tools from server manager: %v", err)
			return skillNames
		}
		
		// Look for skill tools (prefixed with "skills_")
		for _, tool := range tools {
			toolName := tool.Function.Name
			
			// Check if this is a skill tool
			if strings.HasPrefix(toolName, "skills_") {
				// Strip the prefix
				skillTool := strings.TrimPrefix(toolName, "skills_")
				
				// Skip execute_skill_code (it's not a skill itself)
				if skillTool == "execute_skill_code" {
					continue
				}
				
				logging.Debug("Checking skill tool '%s'", skillTool)
				
				// If EnabledSkills is set, only include tools in that list
				if enabledSkillsMap != nil {
					if !enabledSkillsMap[skillTool] {
						logging.Debug("  Skill '%s' NOT in enabled skills map, skipping", skillTool)
						continue
					}
					logging.Debug("  Skill '%s' IS in enabled skills map", skillTool)
				}
				
				if !skillsFound[skillTool] {
					skillsFound[skillTool] = true
					// Convert tool name to display name (underscore to hyphen)
					displayName := strings.ReplaceAll(skillTool, "_", "-")
					skillNames = append(skillNames, displayName)
					logging.Debug("  Added skill: %s (display name: %s)", skillTool, displayName)
				}
			}
		}
		
		// Sort for consistent display
		sort.Strings(skillNames)
		return skillNames
	}
	
	// Fall back to legacy Connections-based discovery
	// Check each connected server
	for _, conn := range m.Connections {
		// Get tools from this server (may be cached)
		tools, err := m.getServerTools(conn)
		if err != nil {
			logging.Debug("Could not get tools from server %s: %v", conn.Name, err)
			continue
		}
		
		// Look for skill tools
		for _, tool := range tools {
			toolName := tool.Name
			
			logging.Debug("Checking tool '%s' from server '%s'", toolName, conn.Name)
			
			// If EnabledSkills is set, only include tools in that list
			if enabledSkillsMap != nil {
				if !enabledSkillsMap[toolName] {
					logging.Debug("  Tool '%s' NOT in enabled skills map, skipping", toolName)
					continue
				}
				logging.Debug("  Tool '%s' IS in enabled skills map", toolName)
			}
			
			// Skip execute_skill_code (it's not a skill itself)
			if toolName == "execute_skill_code" {
				continue
			}
			
			// Check if this tool is from the skills server
			if conn.Name == "skills" && !skillsFound[toolName] {
				skillsFound[toolName] = true
				// Convert tool name to display name (underscore to hyphen)
				displayName := strings.ReplaceAll(toolName, "_", "-")
				skillNames = append(skillNames, displayName)
				logging.Debug("  Added skill: %s (display name: %s)", toolName, displayName)
			}
		}
	}
	
	// Sort for consistent display
	sort.Strings(skillNames)
	
	return skillNames
}

// StartChat starts the chat loop

// SetSessionLogger sets the session logger for this chat manager
func (m *ChatManager) SetSessionLogger(logger *appChat.SessionLogger, providerName, modelName string) {
	m.sessionLogger = logger
	m.providerName = providerName
	m.modelName = modelName
	
	if logger != nil && logger.IsEnabled() {
		logging.Info("Session logging enabled for chat")
	}
}

// logSession logs the current session if session logging is enabled
func (m *ChatManager) logSession() {
	logging.Debug("logSession called - sessionLogger=%v, session=%v", m.sessionLogger != nil, m.session != nil)
	if m.sessionLogger == nil || !m.sessionLogger.IsEnabled() || m.session == nil {
		return
	}
	
	if err := m.sessionLogger.LogSession(m.session, m.providerName, m.modelName); err != nil {
		logging.Warn("Failed to log session: %v", err)
	}
}

func (m *ChatManager) StartChat() error {
	logging.Debug("Session logger status: enabled=%v", m.sessionLogger != nil && m.sessionLogger.IsEnabled())
	// Create session for logging
	if m.sessionLogger != nil && m.sessionLogger.IsEnabled() {
		m.session = appChat.NewSession(m.Context.SystemPrompt)
		logging.Info("Created chat session: %s", m.session.ID)
	}

	// Print welcome message
	m.UI.PrintWelcome()
	
	// Print connected servers
	var serverNames []string
	for _, conn := range m.Connections {
		serverNames = append(serverNames, conn.Name)
	}
	m.UI.PrintConnectedServers(serverNames)
	
	// Discover and print available skills
	availableSkills := m.discoverAvailableSkills()
	if len(availableSkills) > 0 {
		m.UI.PrintEnabledSkills(availableSkills)
	}
	
	// Main chat loop
	for {
		// Read user input
		userInput, err := m.UI.ReadUserInput()
		if err != nil {
			if err == io.EOF {
				m.UI.PrintSystem("Exiting chat mode.")
				return nil
			}
			return fmt.Errorf("error reading input: %w", err)
		}
		
		// Skip empty input
		if strings.TrimSpace(userInput) == "" {
			continue
		}
		
		// Process commands
		if strings.HasPrefix(userInput, "/") {
			cmd := strings.TrimSpace(userInput)
			switch cmd {
			case "/exit", "/quit":
				m.UI.PrintSystem("Exiting chat mode.")
				return nil
			case "/help":
				m.UI.PrintHelp()
				continue
			case "/clear":
				m.Context = NewChatContext(m.Context.SystemPrompt)
				m.UI.PrintSystem("Chat history cleared.")
				continue
			case "/tools":
				m.PrintAvailableTools()
				continue
			case "/history":
				m.PrintChatHistory()
				continue
			case "/system":
				// Handle system prompt setting
				// TODO: Implement this
				m.UI.PrintSystem("System prompt setting not implemented yet.")
				continue
			case "/context":
				// Print context statistics
				m.PrintContextStats()
				continue
			default:
				m.UI.PrintSystem("Unknown command: %s", cmd)
				continue
			}
		}
		
		// Process user message
		err = m.ProcessUserMessage(userInput)
		// Log session after processing message
		m.logSession()
		if err != nil {
			m.UI.PrintError("%v", err)
		}
	}
}

// PrintAvailableTools prints the available tools
func (m *ChatManager) PrintAvailableTools() {
	m.UI.PrintSystem("Available tools:")
	
	for _, conn := range m.Connections {
		serverTools, err := m.getServerTools(conn)
		if err != nil {
			m.UI.PrintError("Failed to get tools from server %s: %v", conn.Name, err)
			continue
		}
		
		m.UI.PrintSystem("Server: %s", conn.Name)
		
		for _, tool := range serverTools {
			fmt.Printf("  - %s: %s", tool.Name, tool.Description)
		}
	}
	
	fmt.Println()
}

// PrintChatHistory prints the chat history
func (m *ChatManager) PrintChatHistory() {
	m.UI.PrintSystem("Chat history:")
	
	for i, msg := range m.Context.Messages {
		switch msg.Role {
		case "user":
			m.UI.userColor.Printf("[%d] User: ", i+1)
			fmt.Println(msg.Content)
		case "assistant":
			m.UI.assistantColor.Printf("[%d] Assistant: ", i+1)
			// Truncate very long messages
			content := msg.Content
			if len(content) > 100 {
				content = content[:100] + "... (truncated)"
			}
			fmt.Println(content)
		case "tool":
			m.UI.toolColor.Printf("[%d] Tool Result (ID: %s): ", i+1, msg.ToolCallID)
			// Truncate very long results
			content := msg.Content
			if len(content) > 100 {
				content = content[:100] + "... (truncated)"
			}
			fmt.Println(content)
		}
	}
	
	fmt.Println()
}

// PrintContextStats prints context utilization statistics
func (m *ChatManager) PrintContextStats() {
	stats := m.Context.GetContextStats()
	
	m.UI.PrintSystem("Context Statistics:")
	fmt.Printf("  Model: %v", stats["model"])
	fmt.Printf("  Messages: %v", stats["message_count"])
	fmt.Printf("  Tool Calls: %v", stats["tool_call_count"])
	fmt.Printf("  Token Management: %v", stats["token_management"])
	
	if stats["token_management"] == "enabled" {
		fmt.Printf("  Current Tokens: %v", stats["current_tokens"])
		fmt.Printf("  Max Tokens: %v", stats["max_tokens"])
		fmt.Printf("  Reserve Tokens: %v", stats["reserve_tokens"])
		fmt.Printf("  Effective Limit: %v", stats["effective_limit"])
		fmt.Printf("  Utilization: %.1f%%", stats["utilization_percent"])
		fmt.Printf("  Provider Configured: %v", stats["provider_configured"])
	} else {
		fmt.Printf("  Max History Size: %v", stats["max_history_size"])
	}
	
	fmt.Println()
}



// convertDomainMessage converts a domain.Message to models.Message for session logging
func convertDomainMessage(msg domain.Message) models.Message {
	return models.Message{
		Role:      models.Role(msg.Role),
		Content:   msg.Content,
		ToolCalls: convertToolCalls(msg.ToolCalls),
	}
}

// convertToolCalls converts domain tool calls to models tool calls
func convertToolCalls(toolCalls []domain.ToolCall) []models.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}
	result := make([]models.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = models.ToolCall{
			ID:   tc.ID,
			Type: models.ToolType(tc.Type),
			Function: models.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}
