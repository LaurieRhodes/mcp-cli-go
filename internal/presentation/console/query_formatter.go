package console

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/core/query"
)

// QueryFormatter handles formatting of query results for console output
type QueryFormatter struct{}

// NewQueryFormatter creates a new query result formatter
func NewQueryFormatter() *QueryFormatter {
	return &QueryFormatter{}
}

// FormatAsJSON formats a query result as JSON
func (f *QueryFormatter) FormatAsJSON(result *query.QueryResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

// FormatAsText formats a query result as plain text
func (f *QueryFormatter) FormatAsText(result *query.QueryResult) string {
	if result == nil {
		return ""
	}
	return result.Response
}

// FormatRawData extracts and formats raw data from tool calls
func (f *QueryFormatter) FormatRawData(toolCalls []query.ToolCallInfo) string {
	if len(toolCalls) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("RAW TOOL DATA:\n------------------------\n\n")

	for i, tc := range toolCalls {
		if tc.Success {
			result.WriteString(fmt.Sprintf("Tool Call #%d: %s\n", i+1, tc.Name))
			result.WriteString("Result:\n")

			// Try to format the result if it's JSON
			formattedResult := f.formatToolResult(tc.Result)
			if formattedResult != "" {
				result.WriteString(formattedResult)
			} else {
				result.WriteString(tc.Result)
			}

			result.WriteString("\n\n")
		}
	}

	return result.String()
}

// formatToolResult attempts to format JSON tool results
func (f *QueryFormatter) formatToolResult(resultStr string) string {
	// First check if it contains a JSON object
	jsonStart := strings.Index(resultStr, "{")
	if jsonStart < 0 {
		return ""
	}

	// Try to parse and format the JSON
	var data interface{}
	err := json.Unmarshal([]byte(resultStr[jsonStart:]), &data)
	if err != nil {
		return ""
	}

	// Format the result based on type
	switch v := data.(type) {
	case map[string]interface{}:
		return f.formatJsonObject(v, 0)
	default:
		return ""
	}
}

// formatJsonObject formats a JSON object with indentation
func (f *QueryFormatter) formatJsonObject(obj map[string]interface{}, indent int) string {
	var result strings.Builder
	indentStr := strings.Repeat("  ", indent)

	// Special handling for security incident data
	if val, ok := obj["result"].(map[string]interface{}); ok {
		if incidents, ok := val["value"].([]interface{}); ok {
			// Found security incidents, format them nicely
			result.WriteString(fmt.Sprintf("%sFound %d security incidents:\n\n", indentStr, len(incidents)))

			for i, inc := range incidents {
				if incident, ok := inc.(map[string]interface{}); ok {
					result.WriteString(fmt.Sprintf("%sIncident %d:\n", indentStr, i+1))

					// Format each field
					for field, value := range incident {
						result.WriteString(fmt.Sprintf("%s- %s: %v\n", indentStr+"  ", field, value))
					}
					result.WriteString("\n")
				}
			}

			return result.String()
		}
	}

	// Generic object formatting
	for key, value := range obj {
		result.WriteString(fmt.Sprintf("%s%s: ", indentStr, key))

		switch v := value.(type) {
		case map[string]interface{}:
			result.WriteString("\n")
			result.WriteString(f.formatJsonObject(v, indent+1))
		case []interface{}:
			result.WriteString("\n")
			for i, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					result.WriteString(fmt.Sprintf("%s  [%d]:\n", indentStr, i))
					result.WriteString(f.formatJsonObject(mapItem, indent+2))
				} else {
					result.WriteString(fmt.Sprintf("%s  [%d]: %v\n", indentStr, i, item))
				}
			}
		default:
			result.WriteString(fmt.Sprintf("%v\n", v))
		}
	}

	return result.String()
}
