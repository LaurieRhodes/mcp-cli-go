package transport

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
)

// MCPClient defines the interface that all MCP transport clients must implement
type MCPClient interface {
	// Start initiates the connection
	Start() error
	
	// Stop terminates the connection
	Stop() error
	
	// Read returns a channel of incoming messages
	Read() <-chan *messages.JSONRPCMessage
	
	// Write sends a message to the server
	Write(msg *messages.JSONRPCMessage) error
	
	// IsRunning returns whether the client is running
	IsRunning() bool
}
