package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/ports"
)

// ToolExecutor handles tool call execution
type ToolExecutor struct {
	mcpManager ports.MCPManager
	maxRounds  int
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(mcpManager ports.MCPManager, maxRounds int) *ToolExecutor {
	if maxRounds == 0 {
		maxRounds = 5 // Default max rounds
	}

	return &ToolExecutor{
		mcpManager: mcpManager,
		maxRounds:  maxRounds,
	}
}

// ExecuteToolCalls executes tool calls and returns results
func (te *ToolExecutor) ExecuteToolCalls(ctx context.Context, toolCalls []models.ToolCall) ([]models.Message, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}

	results := make([]models.Message, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		result, err := te.executeToolCall(ctx, toolCall)
		if err != nil {
			// Create error result
			results = append(results, models.Message{
				Role:       models.RoleTool,
				Content:    fmt.Sprintf("Error: %v", err),
				ToolCallID: toolCall.ID,
			})
			continue
		}

		// Create success result
		results = append(results, models.Message{
			Role:       models.RoleTool,
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	return results, nil
}

// executeToolCall executes a single tool call
func (te *ToolExecutor) executeToolCall(ctx context.Context, toolCall models.ToolCall) (string, error) {
	// Parse arguments
	args, err := toolCall.Function.ParseArguments()
	if err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via MCP manager
	result, err := te.mcpManager.ExecuteTool(ctx, toolCall.Function.Name, args)
	if err != nil {
		return "", err
	}

	return result, nil
}

// GetAvailableTools returns all available tools from MCP servers
func (te *ToolExecutor) GetAvailableTools() ([]models.Tool, error) {
	return te.mcpManager.GetAllTools()
}

// ToolCallRound represents one round of tool calls and responses
type ToolCallRound struct {
	ToolCalls []models.ToolCall
	Results   []models.Message
	Error     error
}

// ExecuteWithFollowUp executes tool calls with automatic follow-up
func (te *ToolExecutor) ExecuteWithFollowUp(
	ctx context.Context,
	service *Service,
	session *Session,
	initialToolCalls []models.ToolCall,
) ([]ToolCallRound, error) {
	rounds := make([]ToolCallRound, 0, te.maxRounds)

	currentToolCalls := initialToolCalls

	for i := 0; i < te.maxRounds; i++ {
		if len(currentToolCalls) == 0 {
			break
		}

		// Execute current round
		results, err := te.ExecuteToolCalls(ctx, currentToolCalls)
		round := ToolCallRound{
			ToolCalls: currentToolCalls,
			Results:   results,
			Error:     err,
		}
		rounds = append(rounds, round)

		if err != nil {
			return rounds, err
		}

		// Add results to session
		for _, result := range results {
			session.AddMessage(result)
		}

		// Send follow-up request
		resp, err := service.SendMessage(ctx, &MessageRequest{
			Messages: session.GetMessages(),
		})
		if err != nil {
			return rounds, fmt.Errorf("follow-up request failed: %w", err)
		}

		// Add assistant response
		session.AddMessage(models.Message{
			Role:      models.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Check if there are more tool calls
		currentToolCalls = resp.ToolCalls
	}

	return rounds, nil
}

// FormatToolResult formats a tool result for display
func FormatToolResult(result string) string {
	// Try to pretty-print JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err == nil {
		if formatted, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			return string(formatted)
		}
	}
	return result
}
