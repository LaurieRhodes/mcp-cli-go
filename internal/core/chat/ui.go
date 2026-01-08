package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
)

// UI manages the user interface for the chat mode
type UI struct {
	// Readline instance for input with history
	rl *readline.Instance
	
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
	
	// Multiline input buffer
	multilineBuffer strings.Builder
}

// NewUI creates a new UI manager
func NewUI() *UI {
	// Check if colors are disabled
	noColor := os.Getenv("NO_COLOR") != ""
	
	// Initialize glamour renderer
	var renderer *glamour.TermRenderer
	if !noColor {
		var err error
		// Use Dracula theme for excellent contrast on dark terminals
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dracula"),
			glamour.WithWordWrap(100),
		)
		if err != nil {
			logging.Warn("Failed to initialize glamour renderer: %v, falling back to plain text", err)
			renderer = nil
		} else {
			logging.Debug("Initialized glamour renderer with dracula style")
		}
	}
	
	// Get history file path
	historyFile := getHistoryFilePath()
	
	// Create readline configuration
	config := &readline.Config{
		Prompt:                 color.New(color.FgGreen, color.Bold).Sprint("You: "),
		HistoryFile:            historyFile,
		HistoryLimit:           1000,
		DisableAutoSaveHistory: false,
		InterruptPrompt:        "^C",
		EOFPrompt:              "/exit",
		HistorySearchFold:      true,
		VimMode:                false,
	}
	
	// Create readline instance
	rl, err := readline.NewEx(config)
	if err != nil {
		logging.Warn("Failed to initialize readline: %v, falling back to basic input", err)
		rl = nil
	}
	
	return &UI{
		rl:              rl,
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

// Close cleans up the UI resources
func (u *UI) Close() error {
	if u.rl != nil {
		return u.rl.Close()
	}
	return nil
}

// getHistoryFilePath returns the path to the history file
func getHistoryFilePath() string {
	// Try to use XDG config directory first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Last resort: current directory
			return ".mcp_cli_history"
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	// Create mcp-cli directory if it doesn't exist
	mcpDir := filepath.Join(configDir, "mcp-cli")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		logging.Warn("Failed to create config directory: %v", err)
		return ".mcp_cli_history"
	}
	
	return filepath.Join(mcpDir, "chat_history")
}

// ReadUserInput reads input from the user with readline support
func (u *UI) ReadUserInput() (string, error) {
	if u.rl == nil {
		// Fallback to basic input if readline failed
		return u.readBasicInput()
	}
	
	// Read a line
	line, err := u.rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			return "", io.EOF
		}
		if err == io.EOF {
			return "", err
		}
		return "", fmt.Errorf("error reading input: %w", err)
	}
	
	// Check for multiline continuation with backslash
	if strings.HasSuffix(strings.TrimSpace(line), "\\") {
		// Start multiline mode
		u.multilineBuffer.Reset()
		u.multilineBuffer.WriteString(strings.TrimSuffix(strings.TrimSpace(line), "\\"))
		
		// Read additional lines
		for {
			u.rl.SetPrompt(color.New(color.FgGreen).Sprint("  ... "))
			nextLine, err := u.rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt {
					fmt.Println("(multiline cancelled)")
					u.rl.SetPrompt(color.New(color.FgGreen, color.Bold).Sprint("You: "))
					return u.ReadUserInput() // Start over
				}
				return "", err
			}
			
			// Check if this line also continues
			if strings.HasSuffix(strings.TrimSpace(nextLine), "\\") {
				u.multilineBuffer.WriteString("\n")
				u.multilineBuffer.WriteString(strings.TrimSuffix(strings.TrimSpace(nextLine), "\\"))
				continue
			}
			
			// Final line
			u.multilineBuffer.WriteString("\n")
			u.multilineBuffer.WriteString(nextLine)
			break
		}
		
		// Reset prompt and return multiline result
		u.rl.SetPrompt(color.New(color.FgGreen, color.Bold).Sprint("You: "))
		result := u.multilineBuffer.String()
		u.multilineBuffer.Reset()
		return result, nil
	}
	
	// Single line - return immediately
	return line, nil
}

// readBasicInput provides fallback input without readline
func (u *UI) readBasicInput() (string, error) {
	fmt.Print(u.userColor.Sprint("You: "))
	
	var line string
	_, err := fmt.Scanln(&line)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	
	return line, nil
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
		
		// Add to buffer
		u.contentBuffer += chunk
		
		// Log the chunk for debugging
		logging.Debug("Received content chunk: %s", chunk)
		
		// For streaming, we collect chunks but don't print them yet
		// We'll render the complete markdown at the end
		// This prevents seeing raw markdown syntax during streaming
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
	
	// Show a subtle indicator that we're collecting the response
	if !u.noColor {
		fmt.Print(lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true).
			Render("Generating response..."))
	}
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
		// Clear the "Generating response..." message
		if !u.noColor {
			fmt.Print("\r\033[K") // Clear the current line
		}
		
		// Render the complete content with Glamour
		if u.glamourRenderer != nil && !u.noColor && len(u.contentBuffer) > 0 {
			// Render the complete markdown content
			rendered, err := u.glamourRenderer.Render(u.contentBuffer)
			if err != nil {
				// Fallback to plain text on error
				logging.Warn("Failed to render markdown: %v", err)
				fmt.Print(u.contentBuffer)
			} else {
				fmt.Print(rendered)
			}
		} else {
			// No glamour renderer or no color - print plain text
			fmt.Print(u.contentBuffer)
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
	
	// Check if this looks like markdown documentation
	isMarkdown := strings.Contains(formattedResult, "##") || 
		          strings.Contains(formattedResult, "```") ||
		          strings.Contains(formattedResult, "**") ||
		          strings.HasPrefix(strings.TrimSpace(formattedResult), "# ")
	
	// For markdown content, use glamour to render it beautifully
	if isMarkdown && u.glamourRenderer != nil && !u.noColor {
		rendered, err := u.glamourRenderer.Render(formattedResult)
		if err != nil {
			logging.Warn("Failed to render markdown tool result: %v", err)
			// Fall back to plain display
			u.printPlainToolResult(formattedResult)
			return
		}
		
		// Display rendered markdown with subtle separators
		separatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
		
		fmt.Println(separatorStyle.Render(strings.Repeat("─", 80)))
		fmt.Print(rendered)
		fmt.Println(separatorStyle.Render(strings.Repeat("─", 80)))
		return
	}
	
	// For non-markdown or when glamour unavailable, use plain display
	u.printPlainToolResult(formattedResult)
}

// printPlainToolResult displays tool results in a plain format with optional truncation
func (u *UI) printPlainToolResult(formattedResult string) {
	const maxChars = 2000
	
	var displayResult string
	if len(formattedResult) > maxChars {
		displayResult = formattedResult[:maxChars] + "\n... (truncated, full result sent to assistant)"
	} else {
		displayResult = formattedResult
	}
	
	if u.noColor {
		fmt.Printf("%s\n", displayResult)
		return
	}
	
	// Use lipgloss box for non-markdown results
	resultStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("240"))
	
	fmt.Println(resultStyle.Render(displayResult))
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
	fmt.Println("  /context     - Show context statistics")
	fmt.Println("  /system      - Set a custom system prompt")
	fmt.Println("  /tools       - List available tools")
	fmt.Println("  /history     - Show conversation history")
	fmt.Println()
	u.systemColor.Println("Input tips:")
	fmt.Println("  ↑/↓          - Navigate command history")
	fmt.Println("  Enter        - Send message")
	fmt.Println("  \\            - Continue input on next line (backslash at end)")
	fmt.Println("  Ctrl+C       - Cancel multiline input / interrupt")
	fmt.Println()
}

// PrintWelcome prints the welcome message
func (u *UI) PrintWelcome() {
	if u.noColor {
		fmt.Println("Welcome to MCP CLI Chat Mode!")
		fmt.Println("Type your messages and press Enter to send.")
		fmt.Println("Use \\ at the end of a line for multiline input.")
		fmt.Println("Type '/help' for commands, '/exit' to quit.")
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
	fmt.Println(infoStyle.Render("Type your messages and press Enter to send."))
	fmt.Println(infoStyle.Render("Use \\ at the end of a line for multiline input."))
	fmt.Println(infoStyle.Render("Type '/help' for commands, '/exit' to quit."))
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

// PrintEnabledSkills prints information about enabled skills
func (u *UI) PrintEnabledSkills(skills []string) {
	if len(skills) == 0 {
		return // Don't print if no skills enabled
	}
	
	if u.noColor {
		fmt.Println("Skills enabled:")
		for _, skill := range skills {
			fmt.Printf("  - %s\n", skill)
		}
		fmt.Println()
		return
	}
	
	// Styled skills list
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")) // Yellow
	
	skillStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")) // Purple
	
	fmt.Println(headerStyle.Render("Skills enabled:"))
	for _, skill := range skills {
		fmt.Printf("  %s %s\n", skillStyle.Render("•"), skill)
	}
	fmt.Println()
}
