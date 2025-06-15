package query

import (
	"encoding/json"
	"time"
)

// QueryResult contains the response from a query execution
type QueryResult struct {
	// The text response from the LLM
	Response string `json:"response"`
	
	// Any tool calls that were made during execution
	ToolCalls []ToolCallInfo `json:"tool_calls,omitempty"`
	
	// Time taken to complete the query
	TimeTaken time.Duration `json:"time_taken"`
	
	// The provider and model used for the query
	Provider string `json:"provider"`
	Model    string `json:"model"`
	
	// List of server names connected for this query
	ServerConnections []string `json:"server_connections,omitempty"`
}

// ToolCallInfo contains information about a tool call that was made
type ToolCallInfo struct {
	// The name of the tool
	Name string `json:"name"`
	
	// The arguments passed to the tool
	Arguments json.RawMessage `json:"arguments"`
	
	// The result returned by the tool
	Result string `json:"result"`
	
	// Indicates if the tool call was successful
	Success bool `json:"success"`
	
	// Error message if the tool call failed
	Error string `json:"error,omitempty"`
}
