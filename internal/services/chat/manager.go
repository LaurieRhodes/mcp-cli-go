package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/mcp"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
)

// ChatManager manages the chat flow
type ChatManager struct {
	// LLM provider for chat completions (updated to use new domain interface)
	LLMProvider domain.LLMProvider
	
	// Server connections for tool execution
	Connections []*host.ServerConnection
	
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
			
			// If there was no textual response, tell the user we're still processing
			if response.Response == "" {
				// Get a new response from the model after processing the tools
				err = m.ProcessAfterToolExecution(userMessage.Content)
				if err != nil {
					m.UI.PrintError("Error getting follow-up response: %v", err)
				}
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
	stdioClient, ok := serverConn.Client.(*stdio.StdioClient)
	if !ok {
		return "", fmt.Errorf("server %s client is not stdio client type", serverName)
	}
	
	result, err := tools.SendToolsCall(stdioClient, stdioClient.GetDispatcher(), toolName, args)
	if err != nil {
		return "", fmt.Errorf("tool execution error: %w", err)
	}
	
	// Check for errors in the result
	if result.IsError {
		// Extract error message from content if Error field is empty
		errorMsg := result.Error
		if errorMsg == "" {
			// Try to extract from content array (MCP standard format)
			errorDetector := mcp.NewErrorDetector()
			errorMsg = errorDetector.ExtractTextFromContent(result.Content)
		}
		if errorMsg == "" {
			errorMsg = "unknown error"
		}
		return "", fmt.Errorf("tool execution failed: %s", errorMsg)
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
	var llmTools []domain.Tool
	var anyErrors error
	
	for _, conn := range m.Connections {
		serverTools, err := m.getServerTools(conn)
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
			anyErrors = err
			continue
		}
		
		for _, tool := range serverTools {
			// Format the tool name to be compatible with OpenAI's requirements
			formattedName := formatToolNameForOpenAI(conn.Name, tool.Name)
			
			// Debug log the name transformation
			logging.Debug("Transforming tool name for OpenAI: %s.%s -> %s", conn.Name, tool.Name, formattedName)
			
			// Create the tool with the formatted name using domain types
			llmTool := domain.Tool{
				Type: "function",
				Function: domain.ToolFunction{
					Name:        formattedName,
					Description: fmt.Sprintf("[%s] %s", conn.Name, tool.Description),
					Parameters:  tool.InputSchema,
				},
			}
			llmTools = append(llmTools, llmTool)
		}
	}
	
	if len(llmTools) == 0 && anyErrors != nil {
		return nil, fmt.Errorf("failed to get any tools: %w", anyErrors)
	}
	
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
		
		// Cache the tools
		m.toolsCache[conn.Name] = result.Tools
		serverTools = result.Tools
		
		logging.Info("Successfully got %d tools from server %s", len(serverTools), conn.Name)
		return serverTools, nil
	}
	
	return nil, lastErr
}

// StartChat starts the chat loop
func (m *ChatManager) StartChat() error {
	// Print welcome message
	m.UI.PrintWelcome()
	
	// Print connected servers
	var serverNames []string
	for _, conn := range m.Connections {
		serverNames = append(serverNames, conn.Name)
	}
	m.UI.PrintConnectedServers(serverNames)
	
	// Ensure UI cleanup on exit
	defer func() {
		if err := m.UI.Close(); err != nil {
			logging.Warn("Error closing UI: %v", err)
		}
	}()
	
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
			fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
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
	fmt.Printf("  Model: %v\n", stats["model"])
	fmt.Printf("  Messages: %v\n", stats["message_count"])
	fmt.Printf("  Tool Calls: %v\n", stats["tool_call_count"])
	fmt.Printf("  Token Management: %v\n", stats["token_management"])
	
	if stats["token_management"] == "enabled" {
		fmt.Printf("  Current Tokens: %v\n", stats["current_tokens"])
		fmt.Printf("  Max Tokens: %v\n", stats["max_tokens"])
		fmt.Printf("  Reserve Tokens: %v\n", stats["reserve_tokens"])
		fmt.Printf("  Effective Limit: %v\n", stats["effective_limit"])
		fmt.Printf("  Utilization: %.1f%%\n", stats["utilization_percent"])
		fmt.Printf("  Provider Configured: %v\n", stats["provider_configured"])
	} else {
		fmt.Printf("  Max History Size: %v\n", stats["max_history_size"])
	}
	
	fmt.Println()
}
