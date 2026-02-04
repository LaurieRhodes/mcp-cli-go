package console

import (
	"fmt"
	"io"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"
)

// ChatUI handles chat interface display
type ChatUI struct {
	writer   io.Writer
	showMeta bool
}

// NewChatUI creates a new chat UI
func NewChatUI(writer io.Writer) *ChatUI {
	return &ChatUI{
		writer:   writer,
		showMeta: true,
	}
}

// SetShowMeta sets whether to show metadata
func (cui *ChatUI) SetShowMeta(show bool) {
	cui.showMeta = show
}

// DisplayMessage displays a chat message
func (cui *ChatUI) DisplayMessage(msg models.Message) {
	switch msg.Role {
	case models.RoleUser:
		cui.displayUserMessage(msg)
	case models.RoleAssistant:
		cui.displayAssistantMessage(msg)
	case models.RoleSystem:
		cui.displaySystemMessage(msg)
	case models.RoleTool:
		cui.displayToolMessage(msg)
	}
}

// displayUserMessage displays a user message
func (cui *ChatUI) displayUserMessage(msg models.Message) {
	fmt.Fprintf(cui.writer, "\n%s\n", Bold(Cyan("You:")))
	fmt.Fprintf(cui.writer, "%s\n", msg.Content)
}

// displayAssistantMessage displays an assistant message
func (cui *ChatUI) displayAssistantMessage(msg models.Message) {
	fmt.Fprintf(cui.writer, "\n%s\n", Bold(Green("Assistant:")))
	fmt.Fprintf(cui.writer, "%s\n", msg.Content)

	// Display tool calls if present
	if len(msg.ToolCalls) > 0 {
		cui.displayToolCalls(msg.ToolCalls)
	}
}

// displaySystemMessage displays a system message
func (cui *ChatUI) displaySystemMessage(msg models.Message) {
	fmt.Fprintf(cui.writer, "\n%s\n", Dim("[System] "+msg.Content))
}

// displayToolMessage displays a tool result message
func (cui *ChatUI) displayToolMessage(msg models.Message) {
	fmt.Fprintf(cui.writer, "\n%s\n", Dim("[Tool Result]"))

	// Try to format JSON content
	if strings.HasPrefix(msg.Content, "{") || strings.HasPrefix(msg.Content, "[") {
		fmt.Fprintf(cui.writer, "%s\n", Dim(msg.Content))
	} else {
		fmt.Fprintf(cui.writer, "%s\n", msg.Content)
	}
}

// displayToolCalls displays tool calls
func (cui *ChatUI) displayToolCalls(toolCalls []models.ToolCall) {
	fmt.Fprintf(cui.writer, "\n%s\n", Yellow("Calling tools:"))

	for _, tc := range toolCalls {
		fmt.Fprintf(cui.writer, "  • %s", Bold(tc.Function.Name))

		if len(tc.Function.Arguments) > 0 {
			fmt.Fprintf(cui.writer, " %s", Dim(string(tc.Function.Arguments)))
		}

		fmt.Fprintln(cui.writer)
	}
}

// DisplayUsage displays token usage
func (cui *ChatUI) DisplayUsage(usage models.Usage) {
	if !cui.showMeta {
		return
	}

	fmt.Fprintf(cui.writer, "\n%s\n", Dim(fmt.Sprintf(
		"Tokens: %d prompt + %d completion = %d total",
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
	)))
}

// DisplaySessionInfo displays session information
func (cui *ChatUI) DisplaySessionInfo(sessionID string, messageCount int) {
	fmt.Fprintf(cui.writer, "\n%s\n", Dim(fmt.Sprintf(
		"Session: %s | Messages: %d",
		sessionID,
		messageCount,
	)))
}

// DisplayWelcome displays welcome message
func (cui *ChatUI) DisplayWelcome(provider, model string) {
	fmt.Fprintln(cui.writer, Bold("Chat Mode"))
	fmt.Fprintf(cui.writer, "Provider: %s | Model: %s\n", Cyan(provider), Cyan(model))
	fmt.Fprintln(cui.writer, Dim("Type 'exit' or 'quit' to end the chat"))
	fmt.Fprintln(cui.writer, strings.Repeat("─", 50))
}

// DisplayPrompt displays the input prompt
func (cui *ChatUI) DisplayPrompt() {
	fmt.Fprint(cui.writer, "\n"+Bold(Cyan("You: ")))
}

// DisplayStreaming displays streaming indicator
func (cui *ChatUI) DisplayStreaming() {
	fmt.Fprintf(cui.writer, "\n%s ", Bold(Green("Assistant:")))
}

// DisplayError displays an error message
func (cui *ChatUI) DisplayError(err error) {
	fmt.Fprintf(cui.writer, "\n%s %s\n", Red("Error:"), err.Error())
}

// DisplayThinking displays a thinking indicator
func (cui *ChatUI) DisplayThinking() {
	fmt.Fprint(cui.writer, Dim("Thinking..."))
}

// ClearThinking clears the thinking indicator
func (cui *ChatUI) ClearThinking() {
	fmt.Fprint(cui.writer, "\r"+strings.Repeat(" ", 20)+"\r")
}

// DisplaySeparator displays a visual separator
func (cui *ChatUI) DisplaySeparator() {
	fmt.Fprintln(cui.writer, Dim(strings.Repeat("─", 50)))
}

// DisplayHistory displays conversation history
func (cui *ChatUI) DisplayHistory(messages []models.Message) {
	fmt.Fprintln(cui.writer, Bold("Conversation History:"))
	fmt.Fprintln(cui.writer)

	for _, msg := range messages {
		cui.DisplayMessage(msg)
	}
}

// DisplayHelp displays help information
func (cui *ChatUI) DisplayHelp() {
	fmt.Fprintln(cui.writer, Bold("Available Commands:"))
	fmt.Fprintln(cui.writer, "  /help    - Show this help message")
	fmt.Fprintln(cui.writer, "  /clear   - Clear conversation history")
	fmt.Fprintln(cui.writer, "  /history - Show conversation history")
	fmt.Fprintln(cui.writer, "  /tokens  - Show token usage")
	fmt.Fprintln(cui.writer, "  /exit    - Exit chat mode")
}
