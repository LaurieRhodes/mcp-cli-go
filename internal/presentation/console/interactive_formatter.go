package console

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/fatih/color"
)

// InteractiveFormatter handles formatting for interactive mode
type InteractiveFormatter struct{}

// NewInteractiveFormatter creates a new interactive formatter
func NewInteractiveFormatter() *InteractiveFormatter {
	return &InteractiveFormatter{}
}

// DisplayHelp shows help information for interactive mode
func (f *InteractiveFormatter) DisplayHelp() {
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

// DisplayDetailedTools shows detailed information about tools
func (f *InteractiveFormatter) DisplayDetailedTools(toolList []tools.Tool) {
	for _, tool := range toolList {
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

// DisplayRawTools shows raw tool definitions in JSON format
func (f *InteractiveFormatter) DisplayRawTools(result *tools.ToolsListResult) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("  Error marshaling to JSON: %v\n", err)
		return
	}

	f.printSyntaxHighlightedJSON(string(jsonData))
}

// DisplayToolResult formats and displays a tool execution result
func (f *InteractiveFormatter) DisplayToolResult(content interface{}) {
	switch v := content.(type) {
	case string:
		// Check if the string is JSON
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(v), &jsonObj); err == nil {
			// It's valid JSON, format it
			f.printJSON(jsonObj)
		} else {
			// Plain string
			fmt.Println(v)
		}
	case []interface{}:
		// Array
		f.printArray(v)
	case map[string]interface{}:
		// Object
		f.printObject(v)
	default:
		// For other types, pretty print as JSON
		f.printJSON(content)
	}
}

// ClearScreen clears the terminal screen
func (f *InteractiveFormatter) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// printJSON pretty prints JSON with syntax highlighting
func (f *InteractiveFormatter) printJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", data)
		return
	}

	f.printSyntaxHighlightedJSON(string(jsonBytes))
}

// printArray formats and prints an array with indices
func (f *InteractiveFormatter) printArray(arr []interface{}) {
	if len(arr) == 0 {
		fmt.Println("[]")
		return
	}

	fmt.Println("[")
	for i, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			fmt.Printf("  [%d] {\n", i)
			for key, val := range v {
				fmt.Printf("    %s: ", key)
				f.printValue(val, 4)
				fmt.Println()
			}
			fmt.Println("  }")
		default:
			fmt.Printf("  [%d] ", i)
			f.printValue(item, 4)
			fmt.Println()
		}
	}
	fmt.Println("]")
}

// printObject formats and prints an object (map)
func (f *InteractiveFormatter) printObject(obj map[string]interface{}) {
	if len(obj) == 0 {
		fmt.Println("{}")
		return
	}

	fmt.Println("{")
	for key, val := range obj {
		fmt.Printf("  %s: ", key)
		f.printValue(val, 2)
		fmt.Println()
	}
	fmt.Println("}")
}

// printValue formats and prints a value with proper indentation
func (f *InteractiveFormatter) printValue(val interface{}, indent int) {
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
			f.printValue(item, indent+2)
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
			f.printValue(v[k], indent+2)
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
func (f *InteractiveFormatter) printSyntaxHighlightedJSON(jsonStr string) {
	lines := strings.Split(jsonStr, "\n")

	for _, line := range lines {
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
			f.highlightJSONValue(value)
		} else {
			// Single element (brackets, braces, etc.)
			f.highlightJSONValue(line)
		}

		fmt.Println()
	}
}

// highlightJSONValue applies color highlighting to a JSON value
func (f *InteractiveFormatter) highlightJSONValue(value string) {
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
