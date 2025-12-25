package models

import "encoding/json"

// Tool represents a callable tool/function
type Tool struct {
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolType represents the type of tool
type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

// ToolFunction defines a function that can be called
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a request to call a tool
type ToolCall struct {
	ID       string       `json:"id"`
	Type     ToolType     `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall contains the function name and arguments
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ParseArguments parses the function arguments into a map
func (fc *FunctionCall) ParseArguments() (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(fc.Arguments, &args); err != nil {
		return nil, err
	}
	return args, nil
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Output     string                 `json:"output"`
	Error      error                  `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// IsSuccess returns true if the tool execution succeeded
func (tr *ToolResult) IsSuccess() bool {
	return tr.Error == nil
}
