package query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// Default maximum number of follow-up attempts to avoid infinite loops
const defaultMaxFollowUpAttempts = 2

// QueryHandler handles query execution
type QueryHandler struct {
	// Server connections for tool execution
	Connections []*host.ServerConnection
	
	// LLM client for queries
	LLMClient domain.LLMProvider
	
	// System prompt for the query
	SystemPrompt string
	
	// Additional context messages
	ContextMessages []domain.Message
	
	// Maximum tokens in the response
	MaxTokens int
	
	// Available tools cache
	toolsCache map[string][]tools.Tool
	
	// AI options
	AIOptions *host.AIOptions
	
	// Interface type (for interface-based providers)
	InterfaceType config.InterfaceType
	
	// Tool calls made during execution
	toolCalls []ToolCallInfo
	
	// Server name - needed to check for GraphSecurityIncidents
	ServerName string
	
	// Maximum number of follow-up attempts (configurable)
	MaxFollowUpAttempts int
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(connections []*host.ServerConnection, aiOptions *host.AIOptions, systemPrompt string) (*QueryHandler, error) {
	// Determine the server name
	var serverName string
	if len(connections) == 1 {
		serverName = connections[0].Name
	}
	
	// Use default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question."
	}
	
	// DEBUGGING: Log the exact system prompt being used
	logging.Info("SYSTEM_PROMPT_DEBUG: Using system prompt: %s", systemPrompt)
	
	// Convert AIOptions to ProviderConfig
	providerConfig := &config.ProviderConfig{
		APIKey:       aiOptions.APIKey,
		APIEndpoint:  aiOptions.APIEndpoint,
		DefaultModel: aiOptions.Model,
	}
	
	// Determine provider type
	providerType := domain.ProviderType(aiOptions.Provider)
	
	// Use interface type from AIOptions if set, otherwise infer
	interfaceType := aiOptions.InterfaceType
	if interfaceType == "" {
		switch strings.ToLower(aiOptions.Provider) {
		case "anthropic":
			interfaceType = config.AnthropicNative
		case "ollama":
			interfaceType = config.OllamaNative
		case "gemini":
			interfaceType = config.GeminiNative
		case "bedrock":
			interfaceType = config.AWSBedrock
		case "azure-openai":
			interfaceType = config.AzureOpenAI
		case "vertex-ai":
			interfaceType = config.GCPVertexAI
		default:
			interfaceType = config.OpenAICompatible
		}
	}
	
	logging.Debug("Using interface type %s for provider %s", interfaceType, aiOptions.Provider)
	
	// Create LLM provider using factory
	factory := ai.NewProviderFactory()
	client, err := factory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	
	return &QueryHandler{
		Connections:         connections,
		LLMClient:           client,
		SystemPrompt:        systemPrompt,
		ContextMessages:     []domain.Message{},
		toolsCache:          make(map[string][]tools.Tool),
		AIOptions:           aiOptions,
		InterfaceType:       interfaceType,
		toolCalls:           []ToolCallInfo{},
		ServerName:          serverName,
		MaxFollowUpAttempts: defaultMaxFollowUpAttempts, // Use default value
	}, nil
}

// NewQueryHandlerWithProvider creates a new query handler with a pre-created LLM provider
func NewQueryHandlerWithProvider(connections []*host.ServerConnection, llmProvider domain.LLMProvider, aiOptions *host.AIOptions, systemPrompt string) (*QueryHandler, error) {
	// Determine the server name
	var serverName string
	if len(connections) == 1 {
		serverName = connections[0].Name
	}
	
	// Use default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question."
	}
	
	// DEBUGGING: Log the exact system prompt being used
	logging.Info("SYSTEM_PROMPT_DEBUG: Using system prompt: %s", systemPrompt)
	
	return &QueryHandler{
		Connections:         connections,
		LLMClient:           llmProvider,
		SystemPrompt:        systemPrompt,
		ContextMessages:     []domain.Message{},
		toolsCache:          make(map[string][]tools.Tool),
		AIOptions:           aiOptions,
		InterfaceType:       aiOptions.InterfaceType,
		toolCalls:           []ToolCallInfo{},
		ServerName:          serverName,
		MaxFollowUpAttempts: defaultMaxFollowUpAttempts,
	}, nil
}

// NewQueryHandlerWithInterface creates a new query handler with interface type
func NewQueryHandlerWithInterface(connections []*host.ServerConnection, aiOptions *host.AIOptions, interfaceType config.InterfaceType, systemPrompt string) (*QueryHandler, error) {
	// Determine the server name
	var serverName string
	if len(connections) == 1 {
		serverName = connections[0].Name
	}
	
	// Use default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question."
	}
	
	// DEBUGGING: Log the exact system prompt being used
	logging.Info("SYSTEM_PROMPT_DEBUG: Using system prompt: %s", systemPrompt)
	
	// If no interface type specified, determine from provider
	if interfaceType == "" {
		switch aiOptions.Provider {
		case "openai", "deepseek", "openrouter", "gemini":
			interfaceType = config.OpenAICompatible
		case "anthropic":
			interfaceType = config.AnthropicNative
		case "ollama":
			interfaceType = config.OllamaNative
		default:
			interfaceType = config.OpenAICompatible // Default to OpenAI-compatible
		}
	}
	
	logging.Debug("Using interface type %s for provider %s", interfaceType, aiOptions.Provider)
	
	// Convert AIOptions to ProviderConfig
	providerConfig := &config.ProviderConfig{
		APIKey:       aiOptions.APIKey,
		APIEndpoint:  aiOptions.APIEndpoint,
		DefaultModel: aiOptions.Model,
	}
	
	// Determine provider type
	providerType := domain.ProviderType(aiOptions.Provider)
	
	// Create LLM provider using factory
	factory := ai.NewProviderFactory()
	client, err := factory.CreateProvider(providerType, providerConfig, interfaceType)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	
	return &QueryHandler{
		Connections:         connections,
		LLMClient:           client,
		SystemPrompt:        systemPrompt,
		ContextMessages:     []domain.Message{},
		toolsCache:          make(map[string][]tools.Tool),
		AIOptions:           aiOptions,
		InterfaceType:       interfaceType,
		toolCalls:           []ToolCallInfo{},
		ServerName:          serverName,
		MaxFollowUpAttempts: defaultMaxFollowUpAttempts, // Use default value
	}, nil
}

// SetMaxFollowUpAttempts sets the maximum number of follow-up attempts
func (h *QueryHandler) SetMaxFollowUpAttempts(maxAttempts int) {
	if maxAttempts <= 0 {
		h.MaxFollowUpAttempts = defaultMaxFollowUpAttempts
	} else {
		h.MaxFollowUpAttempts = maxAttempts
	}
	logging.Debug("Set maximum follow-up attempts to: %d", h.MaxFollowUpAttempts)
}

// AddContext adds context to the query
func (h *QueryHandler) AddContext(context string) {
	// Add as a user message with a special prefix
	contextMessage := domain.Message{
		Role:    "user",
		Content: "Context information (use this to help answer my question):\n\n" + context,
	}
	h.ContextMessages = append(h.ContextMessages, contextMessage)
	
	// DEBUGGING: Log context being added
	logging.Info("CONTEXT_DEBUG: Added context message: %s", contextMessage.Content)
}

// SetMaxTokens sets the maximum tokens in the response
func (h *QueryHandler) SetMaxTokens(maxTokens int) {
	h.MaxTokens = maxTokens
}

// Execute executes the query and returns the result
func (h *QueryHandler) Execute(question string) (*QueryResult, error) {
	startTime := time.Now()
	
	// Get available tools for the LLM
	logging.Info("Fetching available tools for LLM")
	llmTools, err := h.GetAvailableTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tools: %w", err)
	}
	logging.Info("Successfully fetched %d tools for LLM", len(llmTools))
	
	// Create messages array with system prompt + context + question
	messages := []domain.Message{
		{
			Role:    "system",
			Content: h.SystemPrompt,
		},
	}
	
	// Add context messages if any
	if len(h.ContextMessages) > 0 {
		messages = append(messages, h.ContextMessages...)
	}
	
	// Add user question
	userMessage := domain.Message{
		Role:    "user",
		Content: question,
	}
	messages = append(messages, userMessage)
	
	// DEBUGGING: Log all messages being sent to LLM - THIS IS KEY!
	logging.Info("=== CRITICAL DEBUG: Messages being sent to LLM ===")
	for i, msg := range messages {
		logging.Info("MESSAGE_DEBUG[%d]: Role=%s, Content=%s", i, msg.Role, msg.Content)
	}
	logging.Info("=== End Messages Debug ===")
	
	// DEBUGGING: Log tools being sent to LLM
	logging.Info("TOOLS_DEBUG: Sending %d tools to LLM", len(llmTools))
	for i, tool := range llmTools {
		logging.Info("TOOL_DEBUG[%d]: Name=%s, Desc=%s", i, tool.Function.Name, tool.Function.Description)
	}
	
	// Execute the query
	logging.Info("Executing query: %s", question)
	
	// Create completion request
	req := &domain.CompletionRequest{
		Messages:     messages,
		Tools:        llmTools,
		SystemPrompt: "", // Already in messages
	}
	
	response, err := h.LLMClient.CreateCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLLMRequest, err)
	}
	
	logging.Debug("Initial response: %s", response.Response)

	// Keep track of number of follow-up attempts to avoid infinite loops
	followUpsUsed := 0
	
	// Detect if we're using Ollama with alternative format
	usingOllamaAlternativeFormat := h.AIOptions.Provider == "ollama" && 
		                          (strings.Contains(response.Response, "<tool_call>") || len(response.ToolCalls) > 0 && len(response.Response) == 0)
	
	if usingOllamaAlternativeFormat {
		logging.Info("Detected Ollama using alternative tool call format or empty content")
	}
	
	// Log the maximum follow-up attempts being used
	logging.Debug("Using maximum follow-up attempts: %d", h.MaxFollowUpAttempts)
	
	// Handle tool calls if present
	for followUpsUsed < h.MaxFollowUpAttempts {
		// Check if we have tool calls in the response
		if response != nil && len(response.ToolCalls) > 0 {
			logging.Info("Query resulted in %d tool calls (follow-up #%d)", len(response.ToolCalls), followUpsUsed+1)
			
			// DEBUGGING: Log each tool call in detail
			for i, toolCall := range response.ToolCalls {
				logging.Info("TOOL_CALL_DEBUG[%d]: ID=%s, Name=%s, Args=%s", 
					i, toolCall.ID, toolCall.Function.Name, string(toolCall.Function.Arguments))
			}
			
			// Add assistant message with tool calls to conversation history
			assistantMessage := domain.Message{
				Role:      "assistant",
				Content:   response.Response, // Even if empty, we need to include Content field
				ToolCalls: response.ToolCalls,
			}
			
			// For providers that require content, ensure the assistant message has at least an empty string
			if assistantMessage.Content == "" {
				assistantMessage.Content = " " // Use a space instead of empty string for better compatibility
				logging.Debug("Setting empty assistant message content to space character for API compatibility")
			}
			
			messages = append(messages, assistantMessage)
			
			// Execute tools and get results
			if err := h.handleToolCalls(response.ToolCalls); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrToolExecution, err)
			}
			
			// Add tool result messages to conversation history
			for i, toolCall := range response.ToolCalls {
				toolInfo := h.toolCalls[i]
				
				// Add tool result message
				toolResultMessage := domain.Message{
					Role:       "tool",
					Content:    toolInfo.Result,
					ToolCallID: toolCall.ID,
				}
				messages = append(messages, toolResultMessage)
				logging.Debug("Added tool result message for %s with ID %s: %s", 
					toolCall.Function.Name, toolCall.ID, toolInfo.Result)
			}
			
			// For Ollama alternative format, add a special clarification message to help process the results
			if usingOllamaAlternativeFormat && followUpsUsed > 0 {
				clarificationMsg := domain.Message{
					Role:    "user",
					Content: "I've executed the requested tool. Please provide a final answer without using additional tools.",
				}
				messages = append(messages, clarificationMsg)
				logging.Debug("Added clarification message for Ollama alternative format")
				
				// For second follow-up with Ollama, don't include tools
				if followUpsUsed >= 1 {
					llmTools = []domain.Tool{}
					logging.Debug("Removed tools for final Ollama follow-up to prevent looping")
				}
			}
			
			// Get follow-up response
			logging.Info("Getting follow-up response #%d after tool execution", followUpsUsed+1)
			
			followUpReq := &domain.CompletionRequest{
				Messages:     messages,
				Tools:        llmTools,
				SystemPrompt: "", // Already in messages
			}
			
			followUpResponse, err := h.LLMClient.CreateCompletion(context.Background(), followUpReq)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrLLMRequest, err)
			}
			
			// Log the follow-up response
			logging.Debug("Received follow-up response #%d: %s", followUpsUsed+1, followUpResponse.Response)
			
			// Update our response to the follow-up
			response = followUpResponse
			
			// If this response doesn't have tool calls, we're done
			if len(response.ToolCalls) == 0 {
				// Add this final response to conversation history
				messages = append(messages, domain.Message{
					Role:    "assistant",
					Content: response.Response,
				})
				
				break
			}
			
			// Increment our follow-up counter
			followUpsUsed++
		} else {
			// If we have a response with content but no tool calls, we're done
			break
		}
	}
	
	// A special case: check if we need one final follow-up due to intent to use tools
	if followUpsUsed < h.MaxFollowUpAttempts {
		needsFinalFollowUp := false
		responseContent := strings.ToLower(response.Response)
		
		// Check for phrases that indicate intent to use tools
		if strings.Contains(responseContent, "let me use") || 
		   strings.Contains(responseContent, "i'll use") || 
		   strings.Contains(responseContent, "i will use") ||
		   strings.Contains(responseContent, "let's use") {
			needsFinalFollowUp = true
		}
		
		// Add a special additional message asking to provide a final response
		if needsFinalFollowUp {
			logging.Info("Response indicates intent to use tools, getting final response")
			
			// Add a special message explicitly requesting a final response
			messages = append(messages, domain.Message{
				Role:    "user",
				Content: "Please just provide your final answer based on the information you have.",
			})
			
			// Get final response
			finalReq := &domain.CompletionRequest{
				Messages:     messages,
				Tools:        []domain.Tool{}, // No tools in final request
				SystemPrompt: "",
			}
			
			finalResponse, err := h.LLMClient.CreateCompletion(context.Background(), finalReq)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrLLMRequest, err)
			}
			
			logging.Debug("Received final answer response: %s", finalResponse.Response)
			response = finalResponse
		}
	}
	
	// If we've reached the max follow-ups but still don't have a good response,
	// append a note about that in the response
	if followUpsUsed >= h.MaxFollowUpAttempts && len(response.ToolCalls) > 0 {
		logging.Warn("Reached maximum number of follow-ups (%d) but still getting tool calls", h.MaxFollowUpAttempts)
		response.Response += fmt.Sprintf("\n\n[Note: The maximum number of tool call iterations (%d) was reached. The result may be incomplete.]", h.MaxFollowUpAttempts)
	}
	
	// Calculate time taken
	timeTaken := time.Since(startTime)
	
	// Collect server connection names
	serverConnections := make([]string, 0, len(h.Connections))
	for _, conn := range h.Connections {
		serverConnections = append(serverConnections, conn.Name)
	}
	
	// Create the result
	result := &QueryResult{
		Response:          response.Response,
		ToolCalls:         h.toolCalls,
		TimeTaken:         timeTaken,
		Provider:          h.AIOptions.Provider,
		Model:             h.AIOptions.Model,
		ServerConnections: serverConnections,
	}
	
	return result, nil
}

// handleToolCalls executes tool calls and records the results
func (h *QueryHandler) handleToolCalls(toolCalls []domain.ToolCall) error {
	for _, toolCall := range toolCalls {
		// Log the tool call ID for debugging
		logging.Debug("Processing tool call with ID %s: %s", toolCall.ID, toolCall.Function.Name)
		
		// Parse the function name
		toolName := toolCall.Function.Name
		
		// Execute the tool call
		logging.Info("Executing tool call: %s", toolName)
		
		result, err := h.executeToolCall(toolCall)
		
		// Record tool call info
		toolInfo := ToolCallInfo{
			Name:      toolName,
			Arguments: toolCall.Function.Arguments,
			Success:   err == nil,
		}
		
		if err != nil {
			toolInfo.Error = err.Error()
			toolInfo.Result = fmt.Sprintf("Error: %s", err.Error())
		} else {
			toolInfo.Result = result
		}
		
		h.toolCalls = append(h.toolCalls, toolInfo)
		
		// If there's an error, continue with other tool calls
		if err != nil {
			logging.Error("Tool execution failed: %v", err)
			continue
		}
	}
	
	return nil
}

// executeToolCall executes a single tool call and returns the result
func (h *QueryHandler) executeToolCall(toolCall domain.ToolCall) (string, error) {
	// Parse arguments
	var args map[string]interface{}
	err := json.Unmarshal(toolCall.Function.Arguments, &args)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}
	
	// Special handling for filesystem tools with paths
	if strings.Contains(toolCall.Function.Name, "list_directory") || 
	   strings.Contains(toolCall.Function.Name, "search_files") {
		
		// Make sure a path is set
		if _, ok := args["path"]; !ok || args["path"] == "" {
			// Default to d:\Github for filesystem operations
			args["path"] = "d:\\Github"
			logging.Debug("Added default path for %s: %s", toolCall.Function.Name, args["path"])
		} else {
			// Fix any path issues
			pathStr, ok := args["path"].(string)
			if ok && strings.Contains(pathStr, "Githib") {
				// Fix common typo: Githib -> Github
				newPath := strings.Replace(pathStr, "Githib", "Github", 1)
				args["path"] = newPath
				logging.Debug("Fixed path typo in %s: %s -> %s", toolCall.Function.Name, pathStr, newPath)
			}
		}
		
		// For search_files, also ensure pattern is set
		if strings.Contains(toolCall.Function.Name, "search_files") {
			if _, ok := args["pattern"]; !ok || args["pattern"] == "" {
				args["pattern"] = "*"
				logging.Debug("Added default pattern for %s: %s", toolCall.Function.Name, args["pattern"])
			}
		}
	}
	
	// Parse the function name to extract server and tool
	toolName := toolCall.Function.Name
	serverName := ""
	
	// Handle both formats: "server_name_tool_name" or "server-name-tool-name"
	for _, conn := range h.Connections {
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
		for _, conn := range h.Connections {
			serverTools, err := h.getServerTools(conn)
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
	if serverName == "" && len(h.Connections) > 0 {
		serverName = h.Connections[0].Name
		logging.Warn("Could not determine server for tool %s, using default server %s", toolName, serverName)
	}
	
	// Find the server connection
	var serverConn *host.ServerConnection
	for _, conn := range h.Connections {
		if conn.Name == serverName {
			serverConn = conn
			break
		}
	}
	
	if serverConn == nil {
		return "", fmt.Errorf("server not found: %s", serverName)
	}
	
	// Log the arguments for debugging
	argsJSON, _ := json.MarshalIndent(args, "", "  ")
	logging.Debug("Calling tool %s on server %s with args: %s", toolName, serverName, string(argsJSON))
	
	// Execute the tool call using the tools package
	result, err := tools.SendToolsCall(serverConn.Client, toolName, args)
	if err != nil {
		return "", fmt.Errorf("tool execution error: %w", err)
	}
	
	// Check for errors in the result
	if result.IsError {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}
	
	// Convert result to string if needed
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		resultStr = content
	default:
		// Try to extract text content from structured response
		resultBytes, _ := json.Marshal(content)
		rawJSON := string(resultBytes)
        
        // Look for text content in the JSON structure
        var extractedText string
        
        // Try parsing as an array of content blocks (common format)
        var contentBlocks []map[string]interface{}
        if err := json.Unmarshal(resultBytes, &contentBlocks); err == nil {
            // Try to find text fields in the content blocks
            for _, block := range contentBlocks {
                if textContent, ok := block["text"].(string); ok {
                    extractedText = textContent
                    break
                }
            }
        }
        
        // If we couldn't extract text from the array format, try other formats
        if extractedText == "" {
            // Try as a single content block
            var contentBlock map[string]interface{}
            if err := json.Unmarshal(resultBytes, &contentBlock); err == nil {
                if textContent, ok := contentBlock["text"].(string); ok {
                    extractedText = textContent
                }
            }
        }
        
        // If we successfully extracted text, use it; otherwise use the original JSON
        if extractedText != "" {
            resultStr = extractedText
            logging.Debug("Successfully extracted text content from structured response")
        } else {
            // Fall back to the full JSON
            resultStr = rawJSON
            logging.Debug("Using full JSON for tool result as text couldn't be extracted")
        }
	}
	
	return resultStr, nil
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
func (h *QueryHandler) GetAvailableTools() ([]domain.Tool, error) {
	var llmTools []domain.Tool
	var anyErrors error
	
	for _, conn := range h.Connections {
		serverTools, err := h.getServerTools(conn)
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
			
			// Create the tool with the formatted name
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
		return nil, fmt.Errorf("%w: %v", ErrServerConnection, anyErrors)
	}
	
	return llmTools, nil
}

// getServerTools gets the tools from a server, using cache if available
func (h *QueryHandler) getServerTools(conn *host.ServerConnection) ([]tools.Tool, error) {
	// Check if we have the tools in cache
	if cachedTools, ok := h.toolsCache[conn.Name]; ok {
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
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to get tools from server %s: %w", conn.Name, err)
			logging.Error("%v", lastErr)
			continue
		}
		
		// Cache the tools
		h.toolsCache[conn.Name] = result.Tools
		serverTools = result.Tools
		
		logging.Info("Successfully got %d tools from server %s", len(serverTools), conn.Name)
		return serverTools, nil
	}
	
	return nil, lastErr
}
