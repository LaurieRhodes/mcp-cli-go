package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

// UI manages the user interface for the chat mode
type UI struct {
	// Input scanner
	scanner *bufio.Scanner
	
	// Output colors
	userColor      *color.Color
	assistantColor *color.Color
	systemColor    *color.Color
	toolColor      *color.Color
	errorColor     *color.Color
	
	// Glamour renderer for markdown
	glamourRenderer *glamour.TermRenderer
	noColor         bool
	
	// Stream state tracking
	streamStarted bool
	streamEmpty   bool
	streamMutex   sync.Mutex
	
	// Buffer for content chunks
	contentBuffer string
}

// NewUI creates a new UI manager
func NewUI() *UI {
	// Check if colors are disabled
	noColor := os.Getenv("NO_COLOR") != ""
	
	// Initialize glamour renderer
	var renderer *glamour.TermRenderer
	if !noColor {
		var err error
		renderer, err = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err != nil {
			logging.Warn("Failed to initialize glamour renderer: %v, falling back to plain text", err)
			renderer = nil
		}
	}
	
	return &UI{
		scanner:         bufio.NewScanner(os.Stdin),
		userColor:       color.New(color.FgGreen, color.Bold),
		assistantColor:  color.New(color.FgCyan, color.Bold),
		systemColor:     color.New(color.FgYellow, color.Bold),
		toolColor:       color.New(color.FgBlue, color.Bold),
		errorColor:      color.New(color.FgRed, color.Bold),
		glamourRenderer: renderer,
		noColor:         noColor,
		streamStarted:   false,
		streamEmpty:     true,
		contentBuffer:   "",
	}
}

// ReadUserInput reads a line or multiline input from the user
func (u *UI) ReadUserInput() (string, error) {
	u.userColor.Print("You: ")
	
	var input strings.Builder
	
	// Basic implementation that reads until empty line
	for u.scanner.Scan() {
		line := u.scanner.Text()
		
		// Check if this is a command (single line starting with /)
		if strings.HasPrefix(line, "/") && input.Len() == 0 {
			return line, nil
		}
		
		// Empty line ends multiline input
		if line == "" && input.Len() > 0 {
			break
		}
		
		// Add line to input
		if input.Len() > 0 {
			input.WriteString("\n")
		}
		input.WriteString(line)
	}
	
	if err := u.scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	
	return input.String(), nil
}

// PrintAssistantResponse prints the assistant's response with markdown rendering
func (u *UI) PrintAssistantResponse(response string) {
	u.assistantColor.Println("\nAssistant:")
	
	// Render with Glamour if available
	if u.glamourRenderer != nil && !u.noColor {
		rendered, err := u.glamourRenderer.Render(response)
		if err != nil {
			// Fallback to plain text on error
			logging.Warn("Failed to render markdown: %v", err)
			fmt.Println(response)
		} else {
			fmt.Print(rendered)
		}
	} else {
		// Plain text output
		fmt.Println(response)
	}
	
	fmt.Println()
}

// StreamAssistantResponse prints the assistant's response in a streaming fashion
func (u *UI) StreamAssistantResponse(chunk string) {
	u.streamMutex.Lock()
	defer u.streamMutex.Unlock()
	
	// Print non-empty chunks
	if chunk != "" {
		// If this is the first non-empty chunk, mark the stream as non-empty
		u.streamEmpty = false
		
		// Add to buffer and print
		u.contentBuffer += chunk
		
		// Log the chunk for debugging
		logging.Debug("Received content chunk: %s", chunk)
		
		// Print the chunk directly to stdout
		fmt.Print(chunk)
	}
}

// StartStreamingResponse initializes the streaming response UI
func (u *UI) StartStreamingResponse() {
	u.streamMutex.Lock()
	defer u.streamMutex.Unlock()
	
	u.streamStarted = true
	u.streamEmpty = true
	u.contentBuffer = ""
	u.assistantColor.Println("\nAssistant:")
}

// EndStreamingResponse finalizes the streaming response UI
func (u *UI) EndStreamingResponse() {
	u.streamMutex.Lock()
	defer u.streamMutex.Unlock()
	
	// If we didn't actually receive any content but did get tool calls,
	// print a message to explain what's happening
	if u.streamEmpty {
		u.systemColor.Println("[Using tools to process your request...]")
	} else {
		// For streaming, we've been printing raw chunks
		// Now try to re-render the complete content with Glamour if possible
		// This is a bit of a workaround since streaming and markdown don't mix well
		// In practice, streaming output should already be formatted by the LLM
		if u.glamourRenderer != nil && !u.noColor && len(u.contentBuffer) > 0 {
			// Check if the content looks like markdown (has code fences, headers, etc)
			if strings.Contains(u.contentBuffer, "```") || 
			   strings.Contains(u.contentBuffer, "##") ||
			   strings.Contains(u.contentBuffer, "**") {
				// Clear the screen line and re-render with formatting
				fmt.Print("\r\033[K") // Clear current line
				rendered, err := u.glamourRenderer.Render(u.contentBuffer)
				if err == nil {
					fmt.Print(rendered)
				}
			}
		}
	}
	
	// Add a newline at the end
	fmt.Println()
	
	// Reset stream status
	u.streamStarted = false
	
	// Log the complete content for verification
	if len(u.contentBuffer) > 0 {
		logging.Debug("Complete assistant response: %s", u.contentBuffer)
	}
}

// PrintToolExecution prints information about a tool being executed
func (u *UI) PrintToolExecution(toolName, serverName string) {
	if u.noColor {
		fmt.Printf("\n⚡ Executing: %s on %s\n", toolName, serverName)
		return
	}
	
	// Use lipgloss for better formatting
	toolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Yellow
		Bold(true)
	
	serverStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")). // Gray
		Italic(true)
	
	fmt.Printf("\n%s %s %s\n", 
		toolStyle.Render("⚡ Executing:"),
		toolStyle.Render(toolName),
		serverStyle.Render("on "+serverName),
	)
}

// PrintToolResult prints the result of a tool execution
func (u *UI) PrintToolResult(result string) {
	// First check if this is JSON and try to format it
	formattedResult := u.formatToolResultForDisplay(result)
	
	// For long results, print a shortened version
	if len(formattedResult) > 400 {
		formattedResult = formattedResult[:400] + "... (truncated, full result will be sent to assistant)"
	}
	
	if u.noColor {
		fmt.Printf("Result: %s\n", formattedResult)
		return
	}
	
	// Use lipgloss box for tool results
	resultStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("240"))
	
	fmt.Println(resultStyle.Render(formattedResult))
}

// formatToolResultForDisplay formats a tool result for display
// This helps especially with Anthropic responses where we want to extract formatted text
func (u *UI) formatToolResultForDisplay(result string) string {
	// Check if this is JSON
	if !strings.HasPrefix(strings.TrimSpace(result), "[") && 
	   !strings.HasPrefix(strings.TrimSpace(result), "{") {
		// Not JSON, return as is
		return result
	}
	
	// Try to unmarshal the result to see if it's Anthropic-formatted JSON
	var jsonObj []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonObj); err == nil {
		// This is valid JSON array - check if it matches Anthropic's format
		if len(jsonObj) > 0 {
			for _, item := range jsonObj {
				// Check if this is Anthropic-style "text" content
				if textContent, ok := item["text"].(string); ok {
					return textContent
				}
			}
		}
	}
	
	// Try to unmarshal as a single object
	var singleObj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &singleObj); err == nil {
		// Check if this is Anthropic-style with "content" field containing "text"
		if content, ok := singleObj["content"].([]interface{}); ok {
			var extractedText strings.Builder
			for _, item := range content {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if textContent, ok := itemMap["text"].(string); ok {
						extractedText.WriteString(textContent)
					}
				}
			}
			if extractedText.Len() > 0 {
				return extractedText.String()
			}
		}
	}
	
	// If all else fails, pretty-print the JSON
	prettyJson, err := json.MarshalIndent(json.RawMessage(result), "", "  ")
	if err == nil {
		return string(prettyJson)
	}
	
	// Fallback to original
	return result
}

// PrintError prints an error message
func (u *UI) PrintError(format string, args ...interface{}) {
	u.errorColor.Printf("\nError: "+format+"\n", args...)
}

// PrintSystem prints a system message
func (u *UI) PrintSystem(format string, args ...interface{}) {
	u.systemColor.Printf(format+"\n", args...)
}

// PrintHelp prints the help message
func (u *UI) PrintHelp() {
	u.systemColor.Println("\nAvailable commands:")
	fmt.Println("  /exit, /quit - Exit chat mode")
	fmt.Println("  /help        - Show this help message")
	fmt.Println("  /clear       - Clear chat history")
	fmt.Println("  /system      - Set a custom system prompt")
	fmt.Println("  /tools       - List available tools")
	fmt.Println("  /history     - Show conversation history")
	fmt.Println()
}

// PrintWelcome prints the welcome message
func (u *UI) PrintWelcome() {
	if u.noColor {
		fmt.Println("Welcome to MCP CLI Chat Mode!")
		fmt.Println("Type your messages to chat with the assistant.")
		fmt.Println("Type '/exit' to quit, '/help' for commands.")
		fmt.Println()
		return
	}
	
	// Create styled welcome box
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")). // Cyan
		Padding(0, 1)
	
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")) // Gray
	
	fmt.Println()
	fmt.Println(titleStyle.Render("Welcome to MCP CLI Chat Mode!"))
	fmt.Println(infoStyle.Render("Type your messages to chat with the assistant."))
	fmt.Println(infoStyle.Render("Type '/exit' to quit, '/help' for commands."))
	fmt.Println()
}

// PrintConnectedServers prints information about connected servers
func (u *UI) PrintConnectedServers(connections []string) {
	if len(connections) == 0 {
		return
	}
	
	if u.noColor {
		fmt.Println("Connected to servers:")
		for _, conn := range connections {
			fmt.Printf("  - %s\n", conn)
		}
		fmt.Println()
		return
	}
	
	// Styled server list
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")) // Yellow
	
	serverStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")) // Green
	
	fmt.Println(headerStyle.Render("Connected to servers:"))
	for _, conn := range connections {
		fmt.Printf("  %s %s\n", serverStyle.Render("•"), conn)
	}
	fmt.Println()
}
