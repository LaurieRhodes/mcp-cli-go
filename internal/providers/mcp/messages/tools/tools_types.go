package tools

// Tool represents a tool that can be used by the LLM
type Tool struct {
	// The name of the tool
	Name string `json:"name"`

	// A description of the tool
	Description string `json:"description,omitempty"`

	// The JSON Schema for the input parameters
	InputSchema map[string]interface{} `json:"inputSchema"`

	// Schema information for the output
	OutputSchema map[string]interface{} `json:"outputSchema,omitempty"`
}

// ToolsListParams represents the parameters for a tools/list request
type ToolsListParams struct {
	// Optional filter for specific tool names
	Names []string `json:"names,omitempty"`
}

// ToolsListResult represents the result of a tools/list request
type ToolsListResult struct {
	// The tools available on the server
	Tools []Tool `json:"tools"`
}

// ToolsCallParams represents the parameters for a tools/call request
type ToolsCallParams struct {
	// The name of the tool to call
	Name string `json:"name"`

	// The arguments for the tool call
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolsCallResult represents the result of a tools/call request
type ToolsCallResult struct {
	// Whether an error occurred
	IsError bool `json:"isError"`

	// Error message if isError is true
	Error string `json:"error,omitempty"`

	// Content returned from the tool
	Content interface{} `json:"content"`
}
