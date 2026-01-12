package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/host"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// InteractiveCmd represents the interactive command
var InteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Enter interactive mode with slash commands",
	Long: `Interactive mode provides a command-line interface with slash commands for direct interaction with the server.
You can query server information, list available tools and resources, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Process configuration options - pass configFile
		serverNames, userSpecified := host.ProcessOptions(configFile, serverName, disableFilesystem, providerName, modelName)
		
		// Display the start message
		bold := color.New(color.Bold)
		bold.Printf("Starting interactive mode with server: %s, provider: %s, model: %s\n\n", serverName, providerName, modelName)
		
		logging.Info("Starting interactive mode")
		
		// Run the interactive command
		err := host.RunCommand(runInteractiveMode, configFile, serverNames, userSpecified)
		if err != nil {
			logging.Error("Error in interactive mode: %v", err)
			fmt.Fprintf(os.Stderr, "Error in interactive mode: %v\n", err)
			return err
		}
		
		return nil
	},
}

// runInteractiveMode is the function that runs when the interactive command is executed
func runInteractiveMode(connections []*host.ServerConnection) error {
	logging.Info("Entering interactive mode with %d server connections", len(connections))
	
	fmt.Println("Connected to servers:")
	for _, conn := range connections {
		fmt.Printf("  - %s (%s v%s)\n", conn.Name, conn.ServerInfo.Name, conn.ServerInfo.Version)
	}
	
	// Simple interactive loop
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
		
		// Check for empty input
		if input == "" {
			continue
		}
		
		// Check for exit command
		if strings.EqualFold(input, "/exit") || strings.EqualFold(input, "/quit") || 
		   strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			logging.Info("Exiting interactive mode")
			fmt.Println("Exiting interactive mode.")
			break
		}
		
		// Process commands
		if strings.HasPrefix(input, "/") {
			handleInteractiveCommand(input, connections)
		} else {
			logging.Debug("Unknown input: %s", input)
			fmt.Println("Unknown input. Type '/help' for available commands.")
		}
	}
	
	return nil
}

// handleInteractiveCommand processes slash commands in interactive mode
func handleInteractiveCommand(command string, connections []*host.ServerConnection) {
	// Split the command into parts
	parts := strings.SplitN(command, " ", 3)
	cmd := strings.ToLower(parts[0])
	
	logging.Debug("Processing command: %s", cmd)
	
	// Handle different commands
	switch cmd {
	case "/help":
		displayInteractiveHelp()
	case "/ping":
		logging.Debug("Executing ping command")
		fmt.Println("Pong! Server is responsive.")
	case "/tools":
		logging.Debug("Executing tools command")
		listTools(connections)
	case "/tools-all":
		logging.Debug("Executing tools-all command")
		listToolsDetailed(connections)
	case "/tools-raw":
		logging.Debug("Executing tools-raw command")
		listToolsRaw(connections)
	case "/call":
		if len(parts) < 2 {
			fmt.Println("Usage: /call <server_name> <tool_name> <json_arguments>")
			fmt.Println("Example: /call filesystem list_directory {\"path\": \"D:/Github/mcp-cli-golang\"}")
			fmt.Println("For multi-line JSON input, use:")
			fmt.Println("/call <server_name> <tool_name>")
			fmt.Println("Then enter JSON on multiple lines, end with a line containing only '###'")
			return
		}
		
		serverName := parts[1]
		
		// Check if we have a complete command or need to enter multi-line mode
		if len(parts) == 2 {
			// Multi-line mode
			handleMultiLineCall(serverName, connections)
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
			callTool(connections, serverName, toolName, argsStr)
		}
	case "/cls", "/clear":
		logging.Debug("Executing clear screen command")
		clearScreen()
	default:
		logging.Debug("Unknown command: %s", cmd)
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Type '/help' for available commands.")
	}
}

// handleMultiLineCall handles multi-line JSON input for the /call command
func handleMultiLineCall(serverName string, connections []*host.ServerConnection) {
	reader := bufio.NewReader(os.Stdin)
	
	// Prompt for tool name
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
	
	// Prompt for JSON arguments
	fmt.Println("Enter JSON arguments (end with a line containing only '###'):")
	
	var jsonLines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logging.Error("Error reading JSON input: %v", err)
			fmt.Printf("Error reading JSON input: %v\n", err)
			return
		}
		
		line = strings.TrimSpace(line)
		
		// Check for end marker
		if line == "###" {
			break
		}
		
		jsonLines = append(jsonLines, line)
	}
	
	// Combine JSON lines
	jsonStr := strings.Join(jsonLines, "\n")
	
	// Validate JSON
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &args); err != nil {
		logging.Error("Invalid JSON: %v", err)
		fmt.Printf("Invalid JSON: %v\n", err)
		return
	}
	
	logging.Debug("Executing call command: server=%s, tool=%s, args=%s", 
		serverName, toolName, jsonStr)
	callTool(connections, serverName, toolName, jsonStr)
}

// displayInteractiveHelp shows help information for interactive mode
func displayInteractiveHelp() {
	logging.Debug("Displaying help information")
	fmt.Println("Available Commands:")
	fmt.Println("  /help         - Show this help message")
	fmt.Println("  /ping         - Check if server is responsive")
	fmt.Println("  /tools        - List available tools")
	fmt.Println("  /tools-all    - Show detailed tool information")
	fmt.Println("  /tools-raw    - Show raw tool definitions in JSON")
	fmt.Println("  /call <server> <tool> <args> - Call a tool with arguments")
	fmt.Println("  /call <server>              - Call a tool with multi-line JSON input")
	fmt.Println("  /cls, /clear  - Clear the screen")
	fmt.Println("  /exit, /quit  - Exit interactive mode")
}

// listTools lists the available tools from all connections
func listTools(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Tools from server %s:\n", conn.Name)
		logging.Debug("Listing tools from server: %s", conn.Name)
		
		// Get the tools list from the server
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		
		// Display the tools
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
func listToolsDetailed(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Tools from server %s:\n", conn.Name)
		logging.Debug("Listing detailed tools from server: %s", conn.Name)
		
		// Get the tools list from the server
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting detailed tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		
		// Display the tools with detailed information
		if len(result.Tools) == 0 {
			logging.Debug("No tools available from server: %s", conn.Name)
			fmt.Println("  No tools available")
		} else {
			logging.Debug("Found %d tools from server: %s", len(result.Tools), conn.Name)
			for _, tool := range result.Tools {
				fmt.Printf("  - %s:\n", tool.Name)
				fmt.Printf("    Description: %s\n", tool.Description)
				fmt.Println("    Parameters:")
				
				// Display input schema
				if schema, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
					for paramName, paramDetails := range schema {
						if details, ok := paramDetails.(map[string]interface{}); ok {
							paramType := details["type"]
							paramDesc := details["description"]
							fmt.Printf("      - %s (%v): %v\n", paramName, paramType, paramDesc)
						}
					}
				} else {
					fmt.Println("      No parameters defined")
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}
}

// listToolsRaw shows the raw tool definitions in JSON
func listToolsRaw(connections []*host.ServerConnection) {
	for _, conn := range connections {
		fmt.Printf("Raw tools from server %s:\n", conn.Name)
		logging.Debug("Listing raw tools from server: %s", conn.Name)
		
		// Get the tools list from the server
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Error("Error getting raw tools list from %s: %v", conn.Name, err)
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		
		// Display the raw tool definitions
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			logging.Error("Error marshaling tools to JSON: %v", err)
			fmt.Printf("  Error marshaling to JSON: %v\n", err)
			continue
		}
		
		// Print the JSON with syntax highlighting
		printSyntaxHighlightedJSON(string(jsonData))
		fmt.Println()
	}
}

// callTool calls a tool on the specified server
func callTool(connections []*host.ServerConnection, serverName, toolName, argsStr string) {
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
	
	// Parse the arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		logging.Error("Failed to parse arguments: %v", err)
		fmt.Printf("Error: Failed to parse arguments: %v\n", err)
		return
	}
	
	// Fetch tool definition to validate against schema
	toolDef, err := getToolDefinition(targetConn, toolName)
	if err != nil {
		logging.Error("Failed to get tool definition: %v", err)
		fmt.Printf("Error: Failed to get tool definition: %v\n", err)
		return
	}
	
	// Validate arguments against tool schema
	validationErrors := validateToolArguments(args, toolDef)
	if len(validationErrors) > 0 {
		logging.Error("Argument validation failed: %v", validationErrors)
		color.Red("Argument validation failed:")
		for _, errMsg := range validationErrors {
			fmt.Printf("  - %s\n", errMsg)
		}
		return
	}
	
	// Call the tool
	logging.Info("Calling tool %s on server %s with arguments: %+v", 
		toolName, serverName, args)
	
	fmt.Printf("Calling tool '%s' on server '%s'...\n", toolName, serverName)
	
	result, err := tools.SendToolsCall(targetConn.Client, targetConn.Client.GetDispatcher(), toolName, args)
	if err != nil {
		logging.Error("Error calling tool: %v", err)
		color.Red("Error calling tool: %v\n", err)
		return
	}
	// Check for errors in the result
	if result.IsError {
		logging.Error("Tool execution failed: %s", result.Error)
		color.Red("Tool execution failed: %s\n", result.Error)
		return
	}
	
	// Display the result
	logging.Info("Tool execution successful")
	color.Green("Tool execution successful!\n")
	fmt.Println("Result:")
	
	// Format the result for display
	printFormattedResult(result.Content)
}

// getToolDefinition fetches the tool definition from the server
func getToolDefinition(conn *host.ServerConnection, toolName string) (*tools.Tool, error) {
	logging.Debug("Fetching tool definition for: %s", toolName)
	
	// Get the tools list from the server
	result, err := tools.SendToolsList(conn.Client, []string{toolName})
	if err != nil {
		return nil, fmt.Errorf("failed to get tool definition: %w", err)
	}
	
	// Find the requested tool
	for _, tool := range result.Tools {
		if tool.Name == toolName {
			return &tool, nil
		}
	}
	
	return nil, fmt.Errorf("tool '%s' not found on server", toolName)
}

// validateToolArguments validates tool arguments against the tool's schema
func validateToolArguments(args map[string]interface{}, tool *tools.Tool) []string {
	var errors []string
	
	// Get the properties from the input schema
	properties, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		logging.Error("Invalid input schema for tool: %s", tool.Name)
		return []string{"Tool has invalid input schema"}
	}
	
	// Get the required properties
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
		// Check if this parameter exists in the schema
		propDef, exists := properties[name]
		if !exists {
			errors = append(errors, fmt.Sprintf("Unknown parameter: %s", name))
			continue
		}
		
		// Get the parameter definition
		propMap, ok := propDef.(map[string]interface{})
		if !ok {
			continue // Skip if parameter definition is invalid
		}
		
		// Get the parameter type
		typeValue, hasType := propMap["type"]
		if !hasType {
			continue // Skip if no type information
		}
		
		// Validate the parameter type
		typeStr, ok := typeValue.(string)
		if !ok {
			continue // Skip if type is not a string
		}
		
		// Validate based on type
		switch typeStr {
		case "string":
			if value != nil && fmt.Sprintf("%T", value) != "string" {
				errors = append(errors, fmt.Sprintf("Parameter '%s' must be a string", name))
			}
		case "number":
			if value != nil {
				switch value.(type) {
				case float64, float32, int, int32, int64:
					// These are valid number types
				default:
					errors = append(errors, fmt.Sprintf("Parameter '%s' must be a number", name))
				}
			}
		case "integer":
			if value != nil {
				switch v := value.(type) {
				case float64:
					// JSON unmarshals numbers as float64, check if it's an integer
					if v != float64(int(v)) {
						errors = append(errors, fmt.Sprintf("Parameter '%s' must be an integer", name))
					}
				case int, int32, int64:
					// These are valid integer types
				default:
					errors = append(errors, fmt.Sprintf("Parameter '%s' must be an integer", name))
				}
			}
		case "boolean":
			if value != nil && fmt.Sprintf("%T", value) != "bool" {
				errors = append(errors, fmt.Sprintf("Parameter '%s' must be a boolean", name))
			}
		case "array":
			if value != nil {
				_, isArray := value.([]interface{})
				if !isArray {
					errors = append(errors, fmt.Sprintf("Parameter '%s' must be an array", name))
				}
			}
		case "object":
			if value != nil {
				_, isObject := value.(map[string]interface{})
				if !isObject {
					errors = append(errors, fmt.Sprintf("Parameter '%s' must be an object", name))
				}
			}
		}
		
		// Validate enum values if specified
		if enumValues, hasEnum := propMap["enum"].([]interface{}); hasEnum && value != nil {
			valid := false
			for _, enumVal := range enumValues {
				if fmt.Sprintf("%v", enumVal) == fmt.Sprintf("%v", value) {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, fmt.Sprintf("Parameter '%s' must be one of the allowed values", name))
			}
		}
	}
	
	return errors
}

// printFormattedResult formats and prints a tool result with syntax highlighting
func printFormattedResult(content interface{}) {
	switch v := content.(type) {
	case string:
		// Check if the string is JSON
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(v), &jsonObj); err == nil {
			// It's valid JSON, format it
			printJSON(jsonObj)
		} else {
			// Plain string
			fmt.Println(v)
		}
		
	case []interface{}:
		// Array
		printArray(v)
		
	case map[string]interface{}:
		// Object
		printObject(v)
		
	default:
		// For other types, pretty print as JSON
		printJSON(content)
	}
}

// printJSON pretty prints JSON with syntax highlighting
func printJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", data)
		return
	}
	
	// Print formatted JSON with syntax highlighting
	printSyntaxHighlightedJSON(string(jsonBytes))
}

// printArray formats and prints an array with indices
func printArray(arr []interface{}) {
	if len(arr) == 0 {
		fmt.Println("[]")
		return
	}
	
	fmt.Println("[")
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			// For objects in arrays, indent and format
			fmt.Printf("  [%d] {\n", i)
			for key, val := range v {
				fmt.Printf("    %s: ", key)
				printValue(val, 4)
				fmt.Println()
			}
			fmt.Println("  }")
		default:
			// For other types
			fmt.Printf("  [%d] ", i)
			printValue(item, 4)
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// printObject formats and prints an object (map)
func printObject(obj map[string]interface{}) {
	if len(obj) == 0 {
		fmt.Println("{}")
		return
	}
	
	fmt.Println("{")
	for key, val := range obj {
		fmt.Printf("  %s: ", key)
		printValue(val, 2)
		fmt.Println()
	}
	fmt.Println("}")
}

// printValue formats and prints a value with proper indentation
func printValue(val interface{}, indent int) {
	indentStr := strings.Repeat(" ", indent)
	
	switch v := val.(type) {
	case string:
		color.Green("%q", v)
	case bool:
		if v {
			color.Cyan("true")
		} else {
			color.Cyan("false")
		}
	case float64:
		color.Yellow("%v", v)
	case int:
		color.Yellow("%d", v)
	case nil:
		color.Cyan("null")
	case []interface{}:
		if len(v) == 0 {
			fmt.Print("[]")
			return
		}
		
		fmt.Printf("[\n%s  ", indentStr)
		for i, item := range v {
			printValue(item, indent+2)
			if i < len(v)-1 {
				fmt.Printf(",\n%s  ", indentStr)
			}
		}
		fmt.Printf("\n%s]", indentStr)
	case map[string]interface{}:
		if len(v) == 0 {
			fmt.Print("{}")
			return
		}
		
		fmt.Printf("{\n")
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		for i, k := range keys {
			fmt.Printf("%s  %s: ", indentStr, k)
			printValue(v[k], indent+2)
			if i < len(keys)-1 {
				fmt.Printf(",\n")
			} else {
				fmt.Printf("\n")
			}
		}
		fmt.Printf("%s}", indentStr)
	default:
		fmt.Printf("%v", v)
	}
}

// printSyntaxHighlightedJSON prints JSON string with syntax highlighting
func printSyntaxHighlightedJSON(jsonStr string) {
	// Split by newlines to process each line
	lines := strings.Split(jsonStr, "\n")
	
	for _, line := range lines {
		// Process each part of the line
		if strings.TrimSpace(line) == "" {
			fmt.Println()
			continue
		}
		
		// Check for different JSON elements
		if strings.Contains(line, ":") {
			// Key-value pair
			parts := strings.SplitN(line, ":", 2)
			key := parts[0]
			value := parts[1]
			
			// Highlight key
			keyParts := strings.Split(key, "\"")
			if len(keyParts) >= 3 {
				fmt.Print(keyParts[0])
				color.New(color.FgBlue).Print("\"" + keyParts[1] + "\"")
				fmt.Print(keyParts[2] + ":")
			} else {
				fmt.Print(key + ":")
			}
			
			// Highlight value
			highlightJSONValue(value)
		} else {
			// Single element (brackets, braces, etc.)
			highlightJSONValue(line)
		}
		
		fmt.Println()
	}
}

// highlightJSONValue applies color highlighting to a JSON value
func highlightJSONValue(value string) {
	value = strings.TrimSpace(value)
	
	// Determine value type and highlight accordingly
	switch {
	case value == "{" || value == "}" || value == "[" || value == "]" || value == "," || value == "":
		// Structural elements
		fmt.Print(value)
	case strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""):
		// String value
		color.New(color.FgGreen).Print(value)
	case value == "true" || value == "false":
		// Boolean
		color.New(color.FgCyan).Print(value)
	case value == "null":
		// Null
		color.New(color.FgCyan).Print(value)
	case regexp.MustCompile(`^-?\d+(\.\d+)?`).MatchString(value):
		// Number
		color.New(color.FgYellow).Print(value)
	default:
		// Other
		fmt.Print(value)
	}
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func init() {
	// Interactive command flags (if any specific to interactive mode)
}
