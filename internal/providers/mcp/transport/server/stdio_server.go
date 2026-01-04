package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
)

// MessageHandler handles incoming MCP messages
type MessageHandler interface {
	HandleInitialize(params map[string]interface{}) (map[string]interface{}, error)
	HandleToolsList(params map[string]interface{}) (map[string]interface{}, error)
	HandleToolsCall(params map[string]interface{}) (map[string]interface{}, error)
}

// StdioServer implements an MCP server using stdin/stdout
type StdioServer struct {
	handler    MessageHandler
	stdin      io.Reader
	stdout     io.Writer
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	writeMutex sync.Mutex
	initialized bool
}

// NewStdioServer creates a new stdio-based MCP server
func NewStdioServer(handler MessageHandler) *StdioServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StdioServer{
		handler: handler,
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the MCP server, listening for messages on stdin
func (s *StdioServer) Start() error {
	logging.Info("Starting MCP server in stdio mode")
	
	// Start reading loop
	s.wg.Add(1)
	go s.readLoop()
	
	// Wait for context cancellation
	<-s.ctx.Done()
	
	// Wait for goroutines to finish
	s.wg.Wait()
	
	logging.Info("MCP server stopped")
	return nil
}

// Stop stops the MCP server
func (s *StdioServer) Stop() {
	logging.Info("Stopping MCP server")
	s.cancel()
}

// readLoop reads JSON-RPC messages from stdin
func (s *StdioServer) readLoop() {
	defer s.wg.Done()
	
	logging.Debug("Starting stdin read loop")
	scanner := bufio.NewScanner(s.stdin)
	
	// Increase buffer size for large messages
	buf := make([]byte, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 1024*1024)
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		
		logging.Debug("Received message: %s", line)
		
		// Parse JSON-RPC message
		var msg messages.JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			logging.Error("Failed to parse JSON-RPC message: %v", err)
			s.sendError(messages.NewRequestID(nil), -32700, "Parse error", nil)
			continue
		}
		
		// Handle the message
		s.handleMessage(&msg)
	}
	
	if err := scanner.Err(); err != nil {
		logging.Error("Error reading from stdin: %v", err)
	}
	
	logging.Debug("Stdin read loop ended")
}

// handleMessage routes messages to appropriate handlers
func (s *StdioServer) handleMessage(msg *messages.JSONRPCMessage) {
	logging.Debug("Handling message: method=%s, id=%v", msg.Method, msg.ID)
	
	// Check if this is a notification (no ID or null ID)
	// Notifications should never receive a response
	isNotification := msg.ID.IsEmpty()
	
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized", "notifications/initialized":
		// Client notification that initialization is complete
		logging.Debug("Client initialized")
		s.initialized = true
		// No response needed for notification
	case "tools/list":
		s.handleToolsList(msg)
	case "tools/call":
		s.handleToolsCall(msg)
	default:
		logging.Warn("Unknown method: %s", msg.Method)
		// Only send error response if this is a request (has an ID), not a notification
		if !isNotification {
			s.sendError(msg.ID, -32601, "Method not found", map[string]interface{}{
				"method": msg.Method,
			})
		} else {
			logging.Debug("Ignoring unknown notification: %s", msg.Method)
		}
	}
}

// handleInitialize handles the initialize request
func (s *StdioServer) handleInitialize(msg *messages.JSONRPCMessage) {
	logging.Info("Handling initialize request")
	
	// Parse params
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid initialize params: %v", err)
			s.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}
	
	// Call handler
	result, err := s.handler.HandleInitialize(params)
	if err != nil {
		logging.Error("Initialize handler failed: %v", err)
		s.sendError(msg.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	// Send response
	s.sendResponse(msg.ID, result)
	logging.Info("Initialize request handled successfully")
}

// handleToolsList handles the tools/list request
func (s *StdioServer) handleToolsList(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tools/list request")
	
	// Parse params (may be nil)
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid tools/list params: %v", err)
			s.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}
	
	// Call handler
	result, err := s.handler.HandleToolsList(params)
	if err != nil {
		logging.Error("Tools list handler failed: %v", err)
		s.sendError(msg.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	// Send response
	s.sendResponse(msg.ID, result)
	logging.Debug("Tools list request handled successfully")
}

// handleToolsCall handles the tools/call request
func (s *StdioServer) handleToolsCall(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tools/call request")
	
	// Parse params
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid tools/call params: %v", err)
			s.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}
	
	// Debug: Log received params
	logging.Debug("Received tools/call params: %+v", params)
	if meta, ok := params["_meta"]; ok {
		logging.Debug("Found _meta in params: %+v", meta)
	} else {
		logging.Warn("No _meta found in tools/call params")
	}
	
	// Call handler
	result, err := s.handler.HandleToolsCall(params)
	if err != nil {
		logging.Error("Tools call handler failed: %v", err)
		
		// Return error in MCP tool result format
		s.sendResponse(msg.ID, map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error: %v", err),
				},
			},
			"isError": true,
		})
		return
	}
	
	// Send response
	s.sendResponse(msg.ID, result)
	logging.Debug("Tools call request handled successfully")
}

// sendResponse sends a JSON-RPC response
func (s *StdioServer) sendResponse(id messages.RequestID, result interface{}) {
	// Marshal result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logging.Error("Failed to marshal result: %v", err)
		s.sendError(id, -32603, "Internal error", nil)
		return
	}
	
	response := messages.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}
	
	s.writeMessage(&response)
}

// sendError sends a JSON-RPC error response
func (s *StdioServer) sendError(id messages.RequestID, code int, message string, data interface{}) {
	var dataJSON json.RawMessage
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			logging.Error("Failed to marshal error data: %v", err)
			// Send error without data
			dataJSON = nil
		} else {
			dataJSON = d
		}
	}
	
	response := messages.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &messages.JSONRPCError{
			Code:    code,
			Message: message,
			Data:    dataJSON,
		},
	}
	
	s.writeMessage(&response)
}

// writeMessage writes a JSON-RPC message to stdout
func (s *StdioServer) writeMessage(msg *messages.JSONRPCMessage) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	
	// Marshal to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		logging.Error("Failed to marshal response: %v", err)
		return
	}
	
	// Write to stdout with newline
	data = append(data, '\n')
	
	logging.Debug("Sending message: %s", string(data))
	
	if _, err := s.stdout.Write(data); err != nil {
		logging.Error("Failed to write to stdout: %v", err)
		return
	}
}

// SendProgressNotification sends a progress notification to the client
// This is a one-way notification (no response expected)
func (s *StdioServer) SendProgressNotification(progressToken string, progress float64, total int, message string) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	
	// Create progress notification
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/progress",
		"params": map[string]interface{}{
			"progressToken": progressToken,
			"progress":      progress,
		},
	}
	
	// Add optional fields
	if total > 0 {
		notification["params"].(map[string]interface{})["total"] = total
	}
	if message != "" {
		notification["params"].(map[string]interface{})["message"] = message
	}
	
	// Marshal to JSON
	data, err := json.Marshal(notification)
	if err != nil {
		logging.Error("Failed to marshal progress notification: %v", err)
		return
	}
	
	// Write to stdout with newline
	data = append(data, '\n')
	
	logging.Debug("Sending progress notification: token=%s, progress=%.2f, message=%s", 
		progressToken, progress, message)
	
	if _, err := s.stdout.Write(data); err != nil {
		logging.Error("Failed to write progress notification: %v", err)
		return
	}
}

// IsInitialized returns whether the server has been initialized
func (s *StdioServer) IsInitialized() bool {
	return s.initialized
}
