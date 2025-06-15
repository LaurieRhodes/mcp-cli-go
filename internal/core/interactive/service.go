package interactive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/presentation/console"
	"github.com/fatih/color"
)

// Service handles interactive command execution
type Service struct {
	formatter *console.InteractiveFormatter
}

// Config holds configuration for interactive execution
type Config struct {
	ConfigFile        string
	ServerName        string
	ProviderName      string
	ModelName         string
	DisableFilesystem bool
	ServerNames       []string
	UserSpecified     map[string]bool
}

// NewService creates a new interactive service
func NewService() *Service {
	return &Service{
		formatter: console.NewInteractiveFormatter(),
	}
}

// StartInteractiveSession starts an interactive session with the given configuration
func (s *Service) StartInteractiveSession(config *Config) error {
	logging.Info("Starting interactive mode")

	// Run the interactive session
	return host.RunCommand(s.runInteractiveSession, config.ConfigFile, config.ServerNames, config.UserSpecified)
}

// runInteractiveSession manages the interactive session with server connections
func (s *Service) runInteractiveSession(connections []*host.ServerConnection) error {
	logging.Info("Entering interactive mode with %d server connections", len(connections))

	// Display connection information
	s.displayConnections(connections)

	// Start interactive loop
	fmt.Println("\nInteractive Mode - Type '/help' for available commands or '/exit' to quit")
	
	reader := bufio.NewReader(os.Stdin)
	for {
		// Get user input
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			logging.Error("Error reading input: %v", err)
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		logging.Debug("Received input: %s", input)

		// Handle input
		if shouldExit := s.handleInput(input, connections); shouldExit {
			break
		}
	}

	return nil
}

// displayConnections shows information about connected servers
func (s *Service) displayConnections(connections []*host.ServerConnection) {
	fmt.Println("Connected to servers:")
	for _, conn := range connections {
		fmt.Printf("  - %s (%s v%s)\n", conn.Name, conn.ServerInfo.Name, conn.ServerInfo.Version)
	}
}

// handleInput processes user input and returns whether to exit
func (s *Service) handleInput(input string, connections []*host.ServerConnection) bool {
	// Check for empty input
	if input == "" {
		return false
	}

	// Check for exit commands
	if s.isExitCommand(input) {
		logging.Info("Exiting interactive mode")
		fmt.Println("Exiting interactive mode.")
		return true
	}

	// Process slash commands
	if strings.HasPrefix(input, "/") {
		s.handleSlashCommand(input, connections)
	} else {
		logging.Debug("Unknown input: %s", input)
		fmt.Println("Unknown input. Type '/help' for available commands.")
	}

	return false
}

// isExitCommand checks if the input is an exit command
func (s *Service) isExitCommand(input string) bool {
	exitCommands := []string{"/exit", "/quit", "exit", "quit"}
	for _, cmd := range exitCommands {
		if strings.EqualFold(input, cmd) {
			return true
		}
	}
	return false
}

// handleSlashCommand processes slash commands
func (s *Service) handleSlashCommand(command string, connections []*host.ServerConnection) {
	parts := strings.SplitN(command, " ", 3)
	cmd := strings.ToLower(parts[0])

	logging.Debug("Processing command: %s", cmd)

	switch cmd {
	case "/help":
		s.formatter.DisplayHelp()
	case "/ping":
		logging.Debug("Executing ping command")
		fmt.Println("Pong! Server is responsive.")
	case "/tools":
		logging.Debug("Executing tools command")
		s.listTools(connections)
	case "/tools-all":
		logging.Debug("Executing tools-all command")
		s.listToolsDetailed(connections)
	case "/tools-raw":
		logging.Debug("Executing tools-raw command")
		s.listToolsRaw(connections)
	case "/call":
		s.handleCallCommand(parts, connections)
	case "/cls", "/clear":
		logging.Debug("Executing clear screen command")
		s.formatter.ClearScreen()
	default:
		logging.Debug("Unknown command: %s", cmd)
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Type '/help' for available commands.")
	}
}

// handleCallCommand processes the /call command
func (s *Service) handleCallCommand(parts []string, connections []*host.ServerConnection) {
	if len(parts) < 2 {
		fmt.Println("Usage: /call <server_name> <tool_name> <json_arguments>")
		fmt.Println("Example: /call filesystem list_directory {\"path\": \"D:/Github/mcp-cli-go\"}")
		fmt.Println("For multi-line JSON input, use:")
		fmt.Println("/call <server_name> <tool_name>")
		fmt.Println("Then enter JSON on multiple lines, end with a line containing only '###'")
		return
	}

	serverName := parts[1]

	// Check if we have a complete command or need to enter multi-line mode
	if len(parts) == 2 {
		// Multi-line mode
		s.handleMultiLineCall(serverName, connections)
	} else {
		// Single-line mode
		remaining := strings.SplitN(parts[2], " ", 2)
		if len(remaining) < 2 {
			fmt.Println("Missing tool name or arguments")
			fmt.Println("Usage: /call <server_name> <tool_name> <json_arguments>")
			return
		}
		toolName := remaining[0]
		argsStr := remaining[1]

		logging.Debug("Executing call command: server=%s, tool=%s, args=%s", 
			serverName, toolName, argsStr)
		s.callTool(connections, serverName, toolName, argsStr)
	}
}

// handleMultiLineCall handles multi-line JSON input for the /call command
func (s *Service) handleMultiLineCall(serverName string, connections []*host.ServerConnection) {
	reader := bufio.NewReader(os.Stdin)

	// Get tool name
	fmt.Print("Enter tool name: ")
	toolName, err := reader.ReadString('\n')
	if err != nil {
		logging.Error("Error reading tool name: %v", err)
		fmt.Printf("Error reading tool name: %v\n", err)
		return
	}
	toolName = strings.TrimSpace(toolName)

	if toolName == "" {
		fmt.Println("Tool name cannot be empty.")
		return
	}

	// Get JSON arguments
	fmt.Println("Enter JSON arguments (end with a line containing only '###'):")
	jsonStr := s.readMultiLineJSON(reader)
	if jsonStr == "" {
		return
	}

	logging.Debug("Executing call command: server=%s, tool=%s, args=%s", 
		serverName, toolName, jsonStr)
	s.callTool(connections, serverName, toolName, jsonStr)
}

// readMultiLineJSON reads multi-line JSON input until '###' marker
func (s *Service) readMultiLineJSON(reader *bufio.Reader) string {
	var jsonLines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logging.Error("Error reading JSON input: %v", err)
			fmt.Printf("Error reading JSON input: %v\n", err)
			return ""
		}

		line = strings.TrimSpace(line)
		if line == "###" {
			break
		}

		jsonLines = append(jsonLines, line)
	}

	return strings.Join(jsonLines, "\n")
}

// listTools lists available tools from all connections
func (s *Service) listTools(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Tools from server %s:\n", conn.Name)
		logging.Debug("Listing tools from server: %s", conn.Name)

		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if len(result.Tools) == 0 {
			logging.Debug("No tools available from server: %s", conn.Name)
			fmt.Println("  No tools available")
		} else {
			logging.Debug("Found %d tools from server: %s", len(result.Tools), conn.Name)
			for _, tool := range result.Tools {
				fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			}
		}
		fmt.Println()
	}
}

// listToolsDetailed lists detailed information about all tools
func (s *Service) listToolsDetailed(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Tools from server %s:\n", conn.Name)
		logging.Debug("Listing detailed tools from server: %s", conn.Name)

		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting detailed tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if len(result.Tools) == 0 {
			logging.Debug("No tools available from server: %s", conn.Name)
			fmt.Println("  No tools available")
		} else {
			s.formatter.DisplayDetailedTools(result.Tools)
		}
		fmt.Println()
	}
}

// listToolsRaw shows the raw tool definitions in JSON
func (s *Service) listToolsRaw(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Raw tools from server %s:\n", conn.Name)
		logging.Debug("Listing raw tools from server: %s", conn.Name)

		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting raw tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		s.formatter.DisplayRawTools(result)
		fmt.Println()
	}
}

// callTool calls a tool on the specified server
func (s *Service) callTool(connections []*host.ServerConnection, serverName, toolName, argsStr string) {
	// Find the specified server
	var targetConn *host.ServerConnection
	for _, conn := range connections {
		if conn.Name == serverName {
			targetConn = conn
			break
		}
	}

	if targetConn == nil {
		logging.Error("Server not found: %s", serverName)
		fmt.Printf("Error: Server '%s' not found\n", serverName)
		return
	}

	// Parse arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		logging.Error("Failed to parse arguments: %v", err)
		fmt.Printf("Error: Failed to parse arguments: %v\n", err)
		return
	}

	// Validate tool and arguments
	toolDef, err := s.getToolDefinition(targetConn, toolName)
	if err != nil {
		logging.Error("Failed to get tool definition: %v", err)
		fmt.Printf("Error: Failed to get tool definition: %v\n", err)
		return
	}

	validationErrors := s.validateToolArguments(args, toolDef)
	if len(validationErrors) > 0 {
		logging.Error("Argument validation failed: %v", validationErrors)
		color.Red("Argument validation failed:")
		for _, errMsg := range validationErrors {
			fmt.Printf("  - %s\n", errMsg)
		}
		return
	}

	// Execute tool call
	logging.Info("Calling tool %s on server %s with arguments: %+v", 
		toolName, serverName, args)

	fmt.Printf("Calling tool '%s' on server '%s'...\n", toolName, serverName)

	result, err := tools.SendToolsCall(targetConn.Client, toolName, args)
	if err != nil {
		logging.Error("Error calling tool: %v", err)
		color.Red("Error calling tool: %v\n", err)
		return
	}

	// Check for errors in result
	if result.IsError {
		logging.Error("Tool execution failed: %s", result.Error)
		color.Red("Tool execution failed: %s\n", result.Error)
		return
	}

	// Display result
	logging.Info("Tool execution successful")
	color.Green("Tool execution successful!\n")
	fmt.Println("Result:")
	s.formatter.DisplayToolResult(result.Content)
}

// getToolDefinition fetches the tool definition from the server
func (s *Service) getToolDefinition(conn *host.ServerConnection, toolName string) (*tools.Tool, error) {
	logging.Debug("Fetching tool definition for: %s", toolName)

	result, err := tools.SendToolsList(conn.Client, []string{toolName})
	if err != nil {
		return nil, fmt.Errorf("failed to get tool definition: %w", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == toolName {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool '%s' not found on server", toolName)
}

// validateToolArguments validates tool arguments against the tool's schema
func (s *Service) validateToolArguments(args map[string]interface{}, tool *tools.Tool) []string {
	var errors []string

	// Get properties from input schema
	properties, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		logging.Error("Invalid input schema for tool: %s", tool.Name)
		return []string{"Tool has invalid input schema"}
	}

	// Get required properties
	var required []string
	if req, ok := tool.InputSchema["required"].([]interface{}); ok {
		for _, r := range req {
			if str, ok := r.(string); ok {
				required = append(required, str)
			}
		}
	}

	// Check for required properties
	for _, req := range required {
		if _, exists := args[req]; !exists {
			errors = append(errors, fmt.Sprintf("Missing required parameter: %s", req))
		}
	}

	// Validate each provided argument
	for name, value := range args {
		if propDef, exists := properties[name]; exists {
			if propMap, ok := propDef.(map[string]interface{}); ok {
				if validationError := s.validateParameterType(name, value, propMap); validationError != "" {
					errors = append(errors, validationError)
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Unknown parameter: %s", name))
		}
	}

	return errors
}

// validateParameterType validates a single parameter against its schema
func (s *Service) validateParameterType(name string, value interface{}, propMap map[string]interface{}) string {
	typeValue, hasType := propMap["type"]
	if !hasType {
		return ""
	}

	typeStr, ok := typeValue.(string)
	if !ok {
		return ""
	}

	if value == nil {
		return ""
	}

	switch typeStr {
	case "string":
		if fmt.Sprintf("%T", value) != "string" {
			return fmt.Sprintf("Parameter '%s' must be a string", name)
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// Valid number types
		default:
			return fmt.Sprintf("Parameter '%s' must be a number", name)
		}
	case "integer":
		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) {
				return fmt.Sprintf("Parameter '%s' must be an integer", name)
			}
		case int, int32, int64:
			// Valid integer types
		default:
			return fmt.Sprintf("Parameter '%s' must be an integer", name)
		}
	case "boolean":
		if fmt.Sprintf("%T", value) != "bool" {
			return fmt.Sprintf("Parameter '%s' must be a boolean", name)
		}
	case "array":
		if _, isArray := value.([]interface{}); !isArray {
			return fmt.Sprintf("Parameter '%s' must be an array", name)
		}
	case "object":
		if _, isObject := value.(map[string]interface{}); !isObject {
			return fmt.Sprintf("Parameter '%s' must be an object", name)
		}
	}

	// Validate enum values if specified
	if enumValues, hasEnum := propMap["enum"].([]interface{}); hasEnum {
		valid := false
		for _, enumVal := range enumValues {
			if fmt.Sprintf("%v", enumVal) == fmt.Sprintf("%v", value) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Sprintf("Parameter '%s' must be one of the allowed values", name)
		}
	}

	return ""
}
