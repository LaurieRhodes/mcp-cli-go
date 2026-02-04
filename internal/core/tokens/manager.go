package tokens

import (
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

// Default fallback limits when provider doesn't specify context_window
const (
	DefaultContextWindow = 8000 // Conservative default context window
	DefaultReserveTokens = 1500 // Conservative default reserve tokens
	MinReserveTokens     = 500  // Minimum reserve tokens
	MaxReserveTokens     = 8000 // Maximum reserve tokens
)

// TokenManager handles token counting and context management
type TokenManager struct {
	model          string
	maxTokens      int
	reserveTokens  int
	codec          tokenizer.Codec
	providerConfig *config.ProviderConfig
}

// NewTokenManagerFromProvider creates a new token manager using provider configuration
func NewTokenManagerFromProvider(model string, cfg *config.ProviderConfig) (*TokenManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider configuration is required")
	}

	// Get context window from provider config or use fallback
	maxTokens := cfg.ContextWindow
	if maxTokens == 0 {
		maxTokens = DefaultContextWindow
		logging.Warn("No context window specified for model %s in provider config, using default: %d tokens", model, maxTokens)
	} else {
		logging.Info("Using provider-configured context limit %d tokens for model %s", maxTokens, model)
	}

	// Get reserve tokens from provider config or calculate default
	reserveTokens := cfg.ReserveTokens
	if reserveTokens == 0 {
		reserveTokens = DefaultReserveTokens
		// Ensure it's reasonable relative to context window
		if reserveTokens > maxTokens/3 {
			reserveTokens = maxTokens / 4 // Use 25% of context window
		}
		if reserveTokens < MinReserveTokens {
			reserveTokens = MinReserveTokens
		}
		if reserveTokens > MaxReserveTokens {
			reserveTokens = MaxReserveTokens
		}
		logging.Debug("Calculated default reserve tokens: %d for model %s", reserveTokens, model)
	} else {
		logging.Info("Using provider-configured reserve tokens: %d", reserveTokens)
	}

	// Try to get codec for the model
	codec, err := getCodecForModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to get token codec for model %s: %w", model, err)
	}

	logging.Info("Created provider-aware token manager for model %s: %d max tokens, %d reserve tokens", model, maxTokens, reserveTokens)

	return &TokenManager{
		model:          model,
		maxTokens:      maxTokens,
		reserveTokens:  reserveTokens,
		codec:          codec,
		providerConfig: cfg,
	}, nil
}

// NewTokenManagerFallback creates a token manager with minimal configuration (for backward compatibility)
func NewTokenManagerFallback(model string) (*TokenManager, error) {
	maxTokens := DefaultContextWindow
	reserveTokens := DefaultReserveTokens

	// Ensure reserve tokens are reasonable relative to context window
	if reserveTokens > maxTokens/3 {
		reserveTokens = maxTokens / 4
	}

	codec, err := getCodecForModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to get token codec for model %s: %w", model, err)
	}

	logging.Warn("Created fallback token manager for model %s (no provider config): %d max tokens, %d reserve tokens", model, maxTokens, reserveTokens)

	return &TokenManager{
		model:         model,
		maxTokens:     maxTokens,
		reserveTokens: reserveTokens,
		codec:         codec,
	}, nil
}

// getCodecForModel gets the appropriate tokenizer codec for a model
func getCodecForModel(model string) (tokenizer.Codec, error) {
	modelLower := strings.ToLower(model)

	// Try to get model-specific codec first - using encoding fallback since ForModel requires specific model constants
	// The tiktoken-go/tokenizer library doesn't have a simple string-based ForModel method
	// So we'll map model names to encodings

	var encoding tokenizer.Encoding
	switch {
	case strings.Contains(modelLower, "gpt-4o"):
		encoding = tokenizer.O200kBase
	case strings.Contains(modelLower, "gpt-4") || strings.Contains(modelLower, "gpt-3.5"):
		encoding = tokenizer.Cl100kBase
	case strings.Contains(modelLower, "claude"):
		encoding = tokenizer.Cl100kBase // Claude uses similar encoding
	case strings.Contains(modelLower, "gemini"):
		encoding = tokenizer.Cl100kBase // Gemini approximation
	case strings.Contains(modelLower, "text-davinci"):
		encoding = tokenizer.P50kBase
	case strings.Contains(modelLower, "codex"):
		encoding = tokenizer.P50kBase
	default:
		encoding = tokenizer.Cl100kBase // Default fallback
	}

	codec, err := tokenizer.Get(encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get codec for encoding %v: %w", encoding, err)
	}

	logging.Info("Using encoding %v for model %s", encoding, model)
	return codec, nil
}

// CountTokensInMessage counts tokens in a single message
func (tm *TokenManager) CountTokensInMessage(message domain.Message) int {
	// Convert domain message to OpenAI format for token counting
	openaiMsg := openai.ChatCompletionMessage{
		Role:    message.Role,
		Content: message.Content,
		Name:    message.Name,
	}

	// Handle tool calls
	if len(message.ToolCalls) > 0 {
		var openaiToolCalls []openai.ToolCall
		for _, toolCall := range message.ToolCalls {
			openaiToolCalls = append(openaiToolCalls, openai.ToolCall{
				ID:   toolCall.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: string(toolCall.Function.Arguments),
				},
			})
		}
		openaiMsg.ToolCalls = openaiToolCalls
	}

	// Handle tool call ID
	if message.ToolCallID != "" {
		openaiMsg.ToolCallID = message.ToolCallID
	}

	return tm.countTokensInOpenAIMessage(openaiMsg)
}

// CountTokensInMessages counts tokens in a slice of messages
func (tm *TokenManager) CountTokensInMessages(messages []domain.Message) int {
	// Convert to OpenAI format
	var openaiMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			var openaiToolCalls []openai.ToolCall
			for _, toolCall := range msg.ToolCalls {
				openaiToolCalls = append(openaiToolCalls, openai.ToolCall{
					ID:   toolCall.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      toolCall.Function.Name,
						Arguments: string(toolCall.Function.Arguments),
					},
				})
			}
			openaiMsg.ToolCalls = openaiToolCalls
		}

		// Handle tool call ID
		if msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	return tm.NumTokensFromMessages(openaiMessages)
}

// CountTokensInString counts tokens in a plain string (public method for chunking)
func (tm *TokenManager) CountTokensInString(text string) int {
	return tm.numTokensInString(text)
}

// GetMaxTokens returns the maximum token limit for this provider/model
func (tm *TokenManager) GetMaxTokens() int {
	return tm.maxTokens
}

// GetReserveTokens returns the number of tokens to reserve for response generation
func (tm *TokenManager) GetReserveTokens() int {
	return tm.reserveTokens
}

// GetModel returns the model name
func (tm *TokenManager) GetModel() string {
	return tm.model
}

// GetProviderConfig returns the provider configuration
func (tm *TokenManager) GetProviderConfig() *config.ProviderConfig {
	return tm.providerConfig
}

// TrimMessagesToFit trims messages to fit within the context limit while preserving important content
func (tm *TokenManager) TrimMessagesToFit(messages []domain.Message, customReserveTokens int) []domain.Message {
	// Use custom reserve tokens if provided, otherwise use configured reserve tokens
	reserveTokens := customReserveTokens
	if reserveTokens == 0 {
		reserveTokens = tm.reserveTokens
	}

	targetLimit := tm.maxTokens - reserveTokens
	if targetLimit <= 0 {
		logging.Warn("Reserve tokens (%d) exceeds provider context limit (%d)", reserveTokens, tm.maxTokens)
		return messages
	}

	currentTokens := tm.CountTokensInMessages(messages)
	if currentTokens <= targetLimit {
		// Already within limit
		return messages
	}

	logging.Info("Trimming messages: current=%d tokens, target=%d tokens (max=%d, reserve=%d)",
		currentTokens, targetLimit, tm.maxTokens, reserveTokens)

	// Strategy: Keep system message + recent messages that fit
	var trimmedMessages []domain.Message
	var systemMessage *domain.Message

	// Extract system message if present
	if len(messages) > 0 && messages[0].Role == "system" {
		systemMessage = &messages[0]
		messages = messages[1:] // Remove system message from processing
	}

	// Start from the end and work backwards, preserving tool call/response pairs
	currentTokenCount := 0

	// If we have a system message, include its tokens
	if systemMessage != nil {
		currentTokenCount = tm.CountTokensInMessage(*systemMessage)
	}

	// Process messages in reverse order
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		messageTokens := tm.CountTokensInMessage(message)

		// Check if adding this message would exceed the limit
		if currentTokenCount+messageTokens > targetLimit {
			// Check if this is part of a tool call/response pair that we should preserve
			if tm.isPartOfToolPair(message, messages, i) {
				// Try to include the pair if it fits
				pairTokens := tm.getToolPairTokens(message, messages, i)
				if currentTokenCount+pairTokens <= targetLimit {
					// Include the entire pair
					pairMessages := tm.getToolPairMessages(message, messages, i)
					trimmedMessages = append(pairMessages, trimmedMessages...)
					currentTokenCount += pairTokens

					// Skip the messages we just added
					if message.Role == "tool" {
						i-- // Skip the assistant message that preceded this tool message
					}
					continue
				}
			}

			// Can't fit this message or its pair - stop here
			logging.Debug("Stopped trimming at message %d (would exceed limit by %d tokens)",
				i, currentTokenCount+messageTokens-targetLimit)
			break
		}

		// Add this message
		trimmedMessages = append([]domain.Message{message}, trimmedMessages...)
		currentTokenCount += messageTokens
	}

	// Re-add system message at the beginning if it exists
	if systemMessage != nil {
		trimmedMessages = append([]domain.Message{*systemMessage}, trimmedMessages...)
	}

	finalTokens := tm.CountTokensInMessages(trimmedMessages)
	logging.Info("Trimming complete: %d messages, %d tokens (saved %d tokens)",
		len(trimmedMessages), finalTokens, currentTokens-finalTokens)

	return trimmedMessages
}

// isPartOfToolPair checks if a message is part of a tool call/response pair
func (tm *TokenManager) isPartOfToolPair(message domain.Message, messages []domain.Message, index int) bool {
	// Tool message should have a corresponding assistant message with tool calls
	if message.Role == "tool" && message.ToolCallID != "" {
		return true
	}

	// Assistant message with tool calls should have corresponding tool responses
	if message.Role == "assistant" && len(message.ToolCalls) > 0 {
		return true
	}

	return false
}

// getToolPairTokens calculates the token count for a tool call/response pair
func (tm *TokenManager) getToolPairTokens(message domain.Message, messages []domain.Message, index int) int {
	pairMessages := tm.getToolPairMessages(message, messages, index)
	return tm.CountTokensInMessages(pairMessages)
}

// getToolPairMessages retrieves the messages that form a tool call/response pair
func (tm *TokenManager) getToolPairMessages(message domain.Message, messages []domain.Message, index int) []domain.Message {
	var pairMessages []domain.Message

	if message.Role == "tool" && message.ToolCallID != "" {
		// Find the preceding assistant message with this tool call ID
		for i := index - 1; i >= 0; i-- {
			prevMsg := messages[i]
			if prevMsg.Role == "assistant" && len(prevMsg.ToolCalls) > 0 {
				// Check if any tool call matches
				for _, toolCall := range prevMsg.ToolCalls {
					if toolCall.ID == message.ToolCallID {
						pairMessages = append(pairMessages, prevMsg, message)
						return pairMessages
					}
				}
			}
		}
		// If no matching assistant message found, just return the tool message
		pairMessages = append(pairMessages, message)
	} else if message.Role == "assistant" && len(message.ToolCalls) > 0 {
		// Include the assistant message and any following tool responses
		pairMessages = append(pairMessages, message)

		// Look for tool responses that match the tool call IDs
		for _, toolCall := range message.ToolCalls {
			for i := index + 1; i < len(messages); i++ {
				nextMsg := messages[i]
				if nextMsg.Role == "tool" && nextMsg.ToolCallID == toolCall.ID {
					pairMessages = append(pairMessages, nextMsg)
				}
			}
		}
	} else {
		// Not part of a pair, just return the single message
		pairMessages = append(pairMessages, message)
	}

	return pairMessages
}

// countTokensInOpenAIMessage counts tokens in an OpenAI message format
func (tm *TokenManager) countTokensInOpenAIMessage(message openai.ChatCompletionMessage) int {
	// Use the tokenizer codec to count tokens
	return tm.numTokensFromMessage(message)
}

// NumTokensFromMessages calculates the total tokens in a slice of OpenAI messages
// Based on the OpenAI cookbook example for counting tokens
func (tm *TokenManager) NumTokensFromMessages(messages []openai.ChatCompletionMessage) int {
	var tokensPerMessage, tokensPerName int

	// Model-specific token calculations based on OpenAI's documentation
	model := strings.ToLower(tm.model)
	switch {
	case strings.Contains(model, "gpt-3.5-turbo-0613") ||
		strings.Contains(model, "gpt-3.5-turbo-16k-0613") ||
		strings.Contains(model, "gpt-4-0314") ||
		strings.Contains(model, "gpt-4-32k-0314") ||
		strings.Contains(model, "gpt-4-0613") ||
		strings.Contains(model, "gpt-4-32k-0613"):
		tokensPerMessage = 3
		tokensPerName = 1
	case strings.Contains(model, "gpt-3.5-turbo-0301"):
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\{content}<|end|>\
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") || strings.Contains(model, "gpt-4") {
			tokensPerMessage = 3
			tokensPerName = 1
		} else {
			// For non-OpenAI models, use conservative estimates
			tokensPerMessage = 4
			tokensPerName = 1
		}
	}

	numTokens := 0
	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += tm.numTokensInString(message.Content)
		numTokens += tm.numTokensInString(message.Role)

		if message.Name != "" {
			numTokens += tokensPerName
			numTokens += tm.numTokensInString(message.Name)
		}

		// Handle tool calls
		if len(message.ToolCalls) > 0 {
			for _, toolCall := range message.ToolCalls {
				numTokens += tm.numTokensInString(toolCall.Function.Name)
				numTokens += tm.numTokensInString(toolCall.Function.Arguments)
				numTokens += 3 // Overhead for tool call structure
			}
		}

		// Handle tool call ID
		if message.ToolCallID != "" {
			numTokens += tm.numTokensInString(message.ToolCallID)
		}
	}

	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>

	return numTokens
}

// numTokensFromMessage calculates tokens for a single message
func (tm *TokenManager) numTokensFromMessage(message openai.ChatCompletionMessage) int {
	return tm.NumTokensFromMessages([]openai.ChatCompletionMessage{message})
}

// numTokensInString counts tokens in a string using the codec
func (tm *TokenManager) numTokensInString(text string) int {
	if text == "" {
		return 0
	}

	// Use the Encode method and count tokens (tiktoken-go/tokenizer doesn't have Count method)
	tokens, _, err := tm.codec.Encode(text)
	if err != nil {
		// If encoding fails, use a rough approximation (4 chars per token)
		logging.Debug("Token encoding failed for text length %d, using approximation: %v", len(text), err)
		return len(text) / 4
	}

	return len(tokens)
}

// GetContextUtilization returns the current context utilization as a percentage
func (tm *TokenManager) GetContextUtilization(messages []domain.Message) float64 {
	currentTokens := tm.CountTokensInMessages(messages)
	return float64(currentTokens) / float64(tm.maxTokens) * 100.0
}

// GetEffectiveContextLimit returns the effective context limit (max tokens - reserve tokens)
func (tm *TokenManager) GetEffectiveContextLimit() int {
	return tm.maxTokens - tm.reserveTokens
}

// GetContextStats returns detailed context statistics
func (tm *TokenManager) GetContextStats(messages []domain.Message) map[string]interface{} {
	currentTokens := tm.CountTokensInMessages(messages)
	utilization := tm.GetContextUtilization(messages)

	return map[string]interface{}{
		"model":               tm.model,
		"max_tokens":          tm.maxTokens,
		"reserve_tokens":      tm.reserveTokens,
		"effective_limit":     tm.GetEffectiveContextLimit(),
		"current_tokens":      currentTokens,
		"available_tokens":    tm.maxTokens - currentTokens,
		"utilization_percent": utilization,
		"message_count":       len(messages),
		"provider_configured": tm.providerConfig != nil,
	}
}
