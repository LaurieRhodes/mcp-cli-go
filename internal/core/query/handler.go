package query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/ai"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
)

// Default maximum number of follow-up attempts to avoid infinite loops
const defaultMaxFollowUpAttempts = 2

// QueryHandler handles query execution
type QueryHandler struct {
	// Server connections for tool execution
	Connections []*host.ServerConnection
	
	// LLM provider for queries (updated to use new domain interface)
	LLMProvider domain.LLMProvider
	
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

// ✅ NEW: NewQueryHandlerWithProvider creates a query handler with a pre-initialized provider
func NewQueryHandlerWithProvider(connections []*host.ServerConnection, provider domain.LLMProvider, systemPrompt string) (*QueryHandler, error) {
	// Determine the server name
	var serverName string
	if len(connections) == 1 {
		serverName = connections[0].Name
	}
	
	// Use default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that answers questions concisely and accurately. You have access to tools and should use them when necessary to answer the question."
	}
	
	logging.Info("Creating query handler with pre-initialized provider")
	logging.Debug("System prompt: %s", systemPrompt)
	
	// Create dummy AI options for compatibility (since we already have the provider)
	aiOptions := &host.AIOptions{
		Provider: string(provider.GetProviderType()),
		Model:    "unknown", // Will be determined by provider
	}
	
	return &QueryHandler{
		Connections:         connections,
		LLMProvider:         provider,
		SystemPrompt:        systemPrompt,
		ContextMessages:     []domain.Message{},
		toolsCache:          make(map[string][]tools.Tool),
		AIOptions:           aiOptions,
		toolCalls:           []ToolCallInfo{},
		ServerName:          serverName,
		MaxFollowUpAttempts: defaultMaxFollowUpAttempts,
	}, nil
}

// NewQueryHandler creates a new query handler using the new provider factory
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
	
	// Create provider config from AI options
	providerConfig := &domain.ProviderConfig{
		APIKey:       aiOptions.APIKey,
		DefaultModel: aiOptions.Model,
		APIEndpoint:  aiOptions.APIEndpoint,
	}
	
	// Map provider string to domain type
	var providerType domain.ProviderType
	switch aiOptions.Provider {
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
		providerType = domain.ProviderOpenRouter // Added OpenRouter support
	default:
		return nil, fmt.Errorf("unsupported provider: %s", aiOptions.Provider)
	}
	
	// Create LLM provider using the new factory
	factory := ai.NewProviderFactory()
	provider, err := factory.CreateProvider(providerType, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}
	
	// Validate the provider configuration
	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("provider configuration validation failed: %w", err)
	}
	
	return &QueryHandler{
		Connections:         connections,
		LLMProvider:         provider,
		SystemPrompt:        systemPrompt,
		ContextMessages:     []domain.Message{},
		toolsCache:          make(map[string][]tools.Tool),
		AIOptions:           aiOptions,
		toolCalls:           []ToolCallInfo{},
		ServerName:          serverName,
		MaxFollowUpAttempts: defaultMaxFollowUpAttempts,
	}, nil
}

// NewQueryHandlerWithInterface creates a new query handler with interface type
func NewQueryHandlerWithInterface(connections []*host.ServerConnection, aiOptions *host.AIOptions, interfaceType config.InterfaceType, systemPrompt string) (*QueryHandler, error) {
	// Use the regular constructor since the new factory handles interface types automatically
	return NewQueryHandler(connections, aiOptions, systemPrompt)
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

// Execute executes the query and returns the result (simplified for the new architecture)
func (h *QueryHandler) Execute(question string) (*QueryResult, error) {
	startTime := time.Now()
	
	// Get available tools for the LLM
	logging.Info("Fetching available tools for LLM")
	llmTools, err := h.GetAvailableTools()
	if err != nil {
		return nil, fmt.Errorf("failed to get available tools: %w", err)
	}
	logging.Info("Successfully fetched %d tools for LLM", len(llmTools))
	
	// Create messages array with context + question
	messages := []domain.Message{}
	
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
	
	// Create completion request
	completionReq := &domain.CompletionRequest{
		Messages:     messages,
		Tools:        llmTools,
		SystemPrompt: h.SystemPrompt, // Use system prompt properly
		Temperature:  0.7, // Default temperature
		MaxTokens:    h.MaxTokens,
		Stream:       false,
	}
	
	// Execute the query
	logging.Info("Executing query: %s", question)
	response, err := h.LLMProvider.CreateCompletion(context.Background(), completionReq)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	
	logging.Debug("Initial response: %s", response.Response)

	// Handle tool calls with follow-ups
	followUpsUsed := 0
	for followUpsUsed < h.MaxFollowUpAttempts && len(response.ToolCalls) > 0 {
		logging.Info("Processing %d tool calls (follow-up #%d)", len(response.ToolCalls), followUpsUsed+1)
		
		// Add assistant message with tool calls
		assistantMessage := domain.Message{
			Role:      "assistant",
			Content:   response.Response,
			ToolCalls: response.ToolCalls,
		}
		messages = append(messages, assistantMessage)
		
		// Execute tools and get results
		if err := h.handleToolCalls(response.ToolCalls); err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", err)
		}
		
		// Add tool result messages
		for i, toolCall := range response.ToolCalls {
			if i < len(h.toolCalls) {
				toolInfo := h.toolCalls[len(h.toolCalls)-len(response.ToolCalls)+i]
				toolResultMessage := domain.Message{
					Role:       "tool",
					Content:    toolInfo.Result,
					ToolCallID: toolCall.ID,
				}
				messages = append(messages, toolResultMessage)
			}
		}
		
		// Get follow-up response
		completionReq.Messages = messages
		followUpResponse, err := h.LLMProvider.CreateCompletion(context.Background(), completionReq)
		if err != nil {
			return nil, fmt.Errorf("LLM follow-up request failed: %w", err)
		}
		
		response = followUpResponse
		followUpsUsed++
		
		logging.Debug("Follow-up response #%d: %s", followUpsUsed, response.Response)
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
		logging.Debug("Processing tool call with ID %s: %s", toolCall.ID, toolCall.Function.Name)
		
		result, err := h.executeToolCall(toolCall)
		
		// Record tool call info
		toolInfo := ToolCallInfo{
			Name:      toolCall.Function.Name,
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
	
	// Parse the function name to extract server and tool
	toolName := toolCall.Function.Name
	serverName := ""
	
	// Handle format: "server_name_tool_name"
	for _, conn := range h.Connections {
		if strings.HasPrefix(toolName, conn.Name+"_") {
			serverName = conn.Name
			toolName = strings.TrimPrefix(toolName, conn.Name+"_")
			break
		} else if strings.HasPrefix(toolName, strings.ReplaceAll(conn.Name, "-", "_")+"_") {
			serverName = conn.Name
			toolName = strings.TrimPrefix(toolName, strings.ReplaceAll(conn.Name, "-", "_")+"_")
			break
		}
	}
	
	// If no server found, use first available
	if serverName == "" && len(h.Connections) > 0 {
		serverName = h.Connections[0].Name
		logging.Warn("Could not determine server for tool %s, using default server %s", toolName, serverName)
	}
	
	// Find server connection
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
	
	logging.Debug("Calling tool %s on server %s", toolName, serverName)
	
	// Execute the tool call
	result, err := tools.SendToolsCall(serverConn.Client, toolName, args)
	if err != nil {
		return "", fmt.Errorf("tool execution error: %w", err)
	}
	
	if result.IsError {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}
	
	// Convert result to string
	var resultStr string
	switch content := result.Content.(type) {
	case string:
		resultStr = content
	default:
		resultBytes, _ := json.Marshal(content)
		resultStr = string(resultBytes)
	}
	
	return resultStr, nil
}

// formatToolNameForOpenAI formats the tool name to be compatible with OpenAI's requirements
func formatToolNameForOpenAI(serverName, toolName string) string {
	serverName = strings.ReplaceAll(serverName, ".", "_")
	serverName = strings.ReplaceAll(serverName, " ", "_")
	serverName = strings.ReplaceAll(serverName, "-", "_")
	
	toolName = strings.ReplaceAll(toolName, ".", "_")
	toolName = strings.ReplaceAll(toolName, " ", "_")
	
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
			formattedName := formatToolNameForOpenAI(conn.Name, tool.Name)
			
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
		return nil, fmt.Errorf("server connection error: %w", anyErrors)
	}
	
	return llmTools, nil
}

// getServerTools gets the tools from a server, using cache if available
func (h *QueryHandler) getServerTools(conn *host.ServerConnection) ([]tools.Tool, error) {
	// Check cache
	if cachedTools, ok := h.toolsCache[conn.Name]; ok {
		return cachedTools, nil
	}
	
	// Get tools with retry
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
