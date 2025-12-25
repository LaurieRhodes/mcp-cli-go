package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/fatih/color"
)

// UI manages the user interface for the chat mode
type UI struct {
	// Input scanner
	scanner *bufio.Scanner
	
	// Output colors
	userColor     *color.Color
	assistantColor *color.Color
	systemColor   *color.Color
	toolColor     *color.Color
	errorColor    *color.Color
	
	// Stream state tracking
	streamStarted bool
	streamEmpty   bool
	streamMutex   sync.Mutex
	
	// Buffer for content chunks
	contentBuffer string
}

// NewUI creates a new UI manager
func NewUI() *UI {
	return &UI{
		scanner:        bufio.NewScanner(os.Stdin),
		userColor:      color.New(color.FgGreen, color.Bold),
		assistantColor: color.New(color.FgCyan, color.Bold),
		systemColor:    color.New(color.FgYellow, color.Bold),
		toolColor:      color.New(color.FgBlue, color.Bold),
		errorColor:     color.New(color.FgRed, color.Bold),
		streamStarted:  false,
		streamEmpty:    true,
		contentBuffer:  "",
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

// PrintAssistantResponse prints the assistant's response with streaming support
func (u *UI) PrintAssistantResponse(response string) {
	u.assistantColor.Println("\nAssistant:")
	fmt.Println(response)
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
		fmt.Println("[Using tools to process your request...]")
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
	u.toolColor.Printf("\nExecuting tool: %s on server: %s\n", toolName, serverName)
}

// PrintToolResult prints the result of a tool execution
func (u *UI) PrintToolResult(result string) {
	fmt.Print("Result: ")
	
	// First check if this is JSON and try to format it
	formattedResult := u.formatToolResultForDisplay(result)
	
	// For long results, print a shortened version
	if len(formattedResult) > 400 {
		fmt.Printf("%s... (truncated, full result will be sent to assistant)\n", formattedResult[:400])
	} else {
		fmt.Printf("%s\n", formattedResult)
	}
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
	u.systemColor.Println("Welcome to MCP CLI Chat Mode!")
	fmt.Println("Type your messages to chat with the assistant.")
	fmt.Println("Type '/exit' to quit, '/help' for commands.")
	fmt.Println()
}

// PrintConnectedServers prints information about connected servers
func (u *UI) PrintConnectedServers(connections []string) {
	u.systemColor.Println("Connected to servers:")
	for _, conn := range connections {
		fmt.Printf("  - %s\n", conn)
	}
	fmt.Println()
}
