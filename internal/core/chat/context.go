package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/tokens"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
    "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config" 
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// ChatContext manages the state of a chat session
type ChatContext struct {
	// Messages in the conversation
	Messages []domain.Message

	// Tool call history
	ToolCalls []ToolCallHistory

	// System prompt template
	SystemPrompt string

	// Maximum number of messages to retain in history (fallback)
	MaxHistorySize int

	// Token manager for sophisticated context management
	TokenManager *tokens.TokenManager

	// Current model being used
	CurrentModel string

	// Current provider configuration
	ProviderConfig *config.ProviderConfig
}

// ToolCallHistory tracks the execution of a tool
type ToolCallHistory struct {
	ToolCall  domain.ToolCall
	Result    string
	Timestamp time.Time
	Error     string
}

// NewChatContext creates a new chat context
func NewChatContext(systemPrompt string) *ChatContext {
	return NewChatContextWithProvider(systemPrompt, "", nil)
}

// NewChatContextWithModel creates a new chat context for a specific model (backward compatibility)
func NewChatContextWithModel(systemPrompt string, model string) *ChatContext {
	return NewChatContextWithProvider(systemPrompt, model, nil)
}

// NewChatContextWithProvider creates a new chat context with provider configuration
func NewChatContextWithProvider(systemPrompt string, model string, providerConfig *config.ProviderConfig) *ChatContext {
	// If no system prompt provided, use default one
	if systemPrompt == "" {
		systemPrompt = `You are a helpful assistant with access to tools. The tools are provided by Model Context Protocol (MCP) servers.

When you need to perform actions such as searching for information, accessing files, or interacting with external systems, use the available tools.

To use a tool:
1. Consider if you need to use a tool to answer the user's request
2. Select the appropriate tool based on the tool description
3. Call the tool with the appropriate parameters
4. Wait for the tool to execute and return a result
5. Use the result to inform your response to the user

The tool names are in the format <server>_<tool>, for example 'filesystem_list_directory' for the list_directory tool on the filesystem server.

Format your response in a clear, helpful manner and always explain what tools you're using and why.

For file system interactions, make sure to respect file paths and check if operations succeeded.
`
	}

	context := &ChatContext{
		Messages:       []domain.Message{},
		ToolCalls:      []ToolCallHistory{},
		SystemPrompt:   systemPrompt,
		MaxHistorySize: 50, // Reasonable fallback for models without token management
		CurrentModel:   model,
		ProviderConfig: providerConfig,
	}

	// Initialize token manager if model and provider config are provided
	if model != "" && providerConfig != nil {
		tokenManager, err := tokens.NewTokenManagerFromProvider(model, providerConfig)
		if err != nil {
			logging.Warn("Failed to create provider-aware token manager for model %s: %v, falling back to simple message-based trimming", model, err)
		} else {
			context.TokenManager = tokenManager
			logging.Info("Initialized provider-aware token management for model %s", model)
		}
	} else if model != "" {
		// Fallback to model-only token manager for backward compatibility
		tokenManager, err := tokens.NewTokenManagerFallback(model)
		if err != nil {
			logging.Warn("Failed to create fallback token manager for model %s: %v, using simple message-based trimming", model, err)
		} else {
			context.TokenManager = tokenManager
			logging.Info("Initialized fallback token management for model %s", model)
		}
	}

	return context
}

// UpdateProvider updates the model and provider configuration and reinitializes token management
func (c *ChatContext) UpdateProvider(model string, providerConfig *config.ProviderConfig) error {
	if model == c.CurrentModel && providerConfig == c.ProviderConfig {
		return nil // No change needed
	}

	c.CurrentModel = model
	c.ProviderConfig = providerConfig
	
	if model == "" {
		c.TokenManager = nil
		return nil
	}

	var tokenManager *tokens.TokenManager
	var err error

	if providerConfig != nil {
		tokenManager, err = tokens.NewTokenManagerFromProvider(model, providerConfig)
		if err != nil {
			logging.Warn("Failed to create provider-aware token manager for model %s: %v, falling back to simple token manager", model, err)
			tokenManager, err = tokens.NewTokenManagerFallback(model)
		} else {
			logging.Info("Updated to provider-aware token management for model %s", model)
		}
	} else {
		tokenManager, err = tokens.NewTokenManagerFallback(model)
		if err == nil {
			logging.Info("Updated to fallback token management for model %s", model)
		}
	}

	if err != nil {
		logging.Warn("Failed to create token manager for model %s: %v, falling back to simple message-based trimming", model, err)
		c.TokenManager = nil
		return err
	}

	c.TokenManager = tokenManager
	
	// Immediately trim messages if we're over the new limit
	c.TrimHistory()
	
	return nil
}

// UpdateModel updates the model and reinitializes token management (backward compatibility)
func (c *ChatContext) UpdateModel(model string) error {
	return c.UpdateProvider(model, c.ProviderConfig)
}

// AddMessage adds a message to the context
func (c *ChatContext) AddMessage(message domain.Message) {
	c.Messages = append(c.Messages, message)
	c.TrimHistory()
}

// AddToolCall adds a tool call to the history
func (c *ChatContext) AddToolCall(toolCall domain.ToolCall, result string, err error) {
	history := ToolCallHistory{
		ToolCall:  toolCall,
		Result:    result,
		Timestamp: time.Now(),
	}
	
	if err != nil {
		history.Error = err.Error()
	}
	
	c.ToolCalls = append(c.ToolCalls, history)
}

// GetMessagesForLLM returns the messages to send to the LLM
func (c *ChatContext) GetMessagesForLLM() []domain.Message {
	// Start with system message
	messages := []domain.Message{
		{
			Role:    "system",
			Content: c.BuildSystemPrompt(),
		},
	}
	
	// Add all conversation messages (no validation yet)
	messages = append(messages, c.Messages...)

	// Apply token-based trimming FIRST (before validation)
	// This ensures we only validate messages that will actually be sent
	if c.TokenManager != nil {
		originalCount := len(messages)
		messages = c.TokenManager.TrimMessagesToFit(messages, 0) // Use default reserve tokens
		
		if len(messages) != originalCount {
			logging.Debug("Trimmed messages: %d -> %d", originalCount, len(messages))
		}
		
		// Log context utilization
		utilization := c.TokenManager.GetContextUtilization(messages)
		if utilization > 80.0 {
			logging.Warn("High context utilization: %.1f%% (%d/%d tokens)", 
				utilization, 
				c.TokenManager.CountTokensInMessages(messages),
				c.TokenManager.GetMaxTokens())
		} else {
			logging.Debug("Context utilization: %.1f%% (%d/%d tokens)", 
				utilization,
				c.TokenManager.CountTokensInMessages(messages),
				c.TokenManager.GetMaxTokens())
		}

	// Debug: Log ALL messages BEFORE validation
	logging.Debug("=== ALL MESSAGES BEFORE VALIDATION (count: %d) ===", len(messages))
	for i, msg := range messages {
		logging.Debug("  Pre-validation Message %d: role=%s, tool_call_id='%s', has_tool_calls=%v, content_len=%d", 
			i, msg.Role, msg.ToolCallID, len(msg.ToolCalls) > 0, len(msg.Content))
		if msg.Role == "tool" && msg.ToolCallID == "" {
			logging.Warn("  ⚠️  Tool message %d has EMPTY ToolCallID!", i)
		}
	}
	logging.Debug("=== END PRE-VALIDATION MESSAGES ===")
	}
	
	// CRITICAL FIX: Validate tool call/response pairing AFTER trimming
	// This prevents orphaned tool messages when trimming removes assistant messages
	validatedMessages := []domain.Message{messages[0]} // Keep system message
	toolCallIDToIndex := make(map[string]int)
	
	// Build mapping and validate pairing on the FINAL (post-trim) message set
	for _, msg := range messages[1:] {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			// Register tool call IDs from this assistant message
			for _, toolCall := range msg.ToolCalls {
				toolCallIDToIndex[toolCall.ID] = len(validatedMessages)
			}
			validatedMessages = append(validatedMessages, msg)
		} else if msg.Role == "tool" && msg.ToolCallID != "" {
			// Only include tool messages that have a corresponding assistant message
			if _, exists := toolCallIDToIndex[msg.ToolCallID]; exists {
				validatedMessages = append(validatedMessages, msg)
			} else {
				// This can happen if trimming removed the assistant message
				logging.Warn("Skipping orphaned tool message with ID %s (parent assistant message not in final set)", msg.ToolCallID)
			}
		} else {
			// Add all other messages (user, system, etc.)
			validatedMessages = append(validatedMessages, msg)
		}
	}
	
	// Log the message structure for debugging
	logging.Debug("Sending %d messages to LLM (after validation)", len(validatedMessages))
	for i, msg := range validatedMessages {
		if i == 0 {
			logging.Debug("Message %d: role=%s, content=[system prompt]", i, msg.Role)
		} else {
			logging.Debug("Message %d: role=%s, tool_call_id=%s, content_len=%d", i, msg.Role, msg.ToolCallID, len(msg.Content))
		}
	}
	
	return validatedMessages
}

// BuildSystemPrompt builds the system prompt including tool descriptions
func (c *ChatContext) BuildSystemPrompt() string {
	// Base system prompt
	prompt := c.SystemPrompt
	
	// Add recent tool history if available
	if len(c.ToolCalls) > 0 {
		toolHistory := c.FormatToolHistoryForLLM()
		if toolHistory != "" {
			prompt += "" + toolHistory
		}
	}
	
	logging.Debug("Built system prompt: %s", prompt)
	return prompt
}

// TrimHistory trims the history based on available token management or message count
func (c *ChatContext) TrimHistory() {
	if c.TokenManager != nil {
		// Use sophisticated token-based trimming
		originalCount := len(c.Messages)
		originalTokens := c.TokenManager.CountTokensInMessages(c.Messages)
		
		c.Messages = c.TokenManager.TrimMessagesToFit(c.Messages, 0) // Use default reserve tokens
		
		newCount := len(c.Messages)
		newTokens := c.TokenManager.CountTokensInMessages(c.Messages)
		
		if newCount != originalCount {
			logging.Info("Token-based trimming: %d→%d messages, %d→%d tokens", 
				originalCount, newCount, originalTokens, newTokens)
		}
	} else {
		// Fallback to simple message-based trimming
		if len(c.Messages) > c.MaxHistorySize {
			originalCount := len(c.Messages)
			c.Messages = c.Messages[len(c.Messages)-c.MaxHistorySize:]
			logging.Info("Message-based trimming: %d→%d messages", originalCount, len(c.Messages))
		}
	}
}

// GetContextStats returns context utilization statistics
func (c *ChatContext) GetContextStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["message_count"] = len(c.Messages)
	stats["tool_call_count"] = len(c.ToolCalls)
	stats["model"] = c.CurrentModel
	
	if c.TokenManager != nil {
		tokenStats := c.TokenManager.GetContextStats(c.Messages)
		for key, value := range tokenStats {
			stats[key] = value
		}
		stats["token_management"] = "enabled"
	} else {
		stats["max_history_size"] = c.MaxHistorySize
		stats["token_management"] = "disabled"
	}
	
	return stats
}

// FormatToolHistoryForLLM formats the tool history for the LLM
func (c *ChatContext) FormatToolHistoryForLLM() string {
	var history strings.Builder
	
	// Only include the most recent tool calls (last 5)
	recentCalls := c.ToolCalls
	if len(recentCalls) > 5 {
		recentCalls = recentCalls[len(recentCalls)-5:]
	}
	
	if len(recentCalls) == 0 {
		return ""
	}
	
	history.WriteString("Here are the results of recent tool calls:")
	
	for i, toolCall := range recentCalls {
		history.WriteString(fmt.Sprintf("Tool call %d:", i+1))
		history.WriteString(fmt.Sprintf("- Name: %s", toolCall.ToolCall.Function.Name))
		history.WriteString(fmt.Sprintf("- Arguments: %s", string(toolCall.ToolCall.Function.Arguments)))
		
		if toolCall.Error != "" {
			history.WriteString(fmt.Sprintf("- Error: %s", toolCall.Error))
		} else {
			// Truncate very long results
			result := toolCall.Result
			if len(result) > 500 {
				result = result[:500] + "... (truncated)"
			}
			history.WriteString(fmt.Sprintf("- Result: %s", result))
		}
		
		history.WriteString("")
	}
	
	return history.String()
}
