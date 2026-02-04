package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
)

// UnixSocketServer implements an MCP server using Unix domain sockets
type UnixSocketServer struct {
	handler     MessageHandler
	socketPath  string
	listener    net.Listener
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	connections map[net.Conn]bool
	connMutex   sync.Mutex
}

// NewUnixSocketServer creates a new Unix socket-based MCP server
func NewUnixSocketServer(handler MessageHandler, socketPath string) *UnixSocketServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &UnixSocketServer{
		handler:     handler,
		socketPath:  socketPath,
		ctx:         ctx,
		cancel:      cancel,
		connections: make(map[net.Conn]bool),
	}
}

// Start starts the MCP server, listening on the Unix socket
func (s *UnixSocketServer) Start() error {
	logging.Info("Starting MCP server in Unix socket mode: %s", s.socketPath)

	// Remove old socket if it exists
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		logging.Warn("Failed to remove old socket: %v", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}
	s.listener = listener

	// Set secure permissions (owner only)
	if err := os.Chmod(s.socketPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	logging.Info("Unix socket created with secure permissions (0600): %s", s.socketPath)

	// Start accept loop in background
	s.wg.Add(1)
	go s.acceptLoop()

	// Return immediately - don't block!
	logging.Info("Unix socket server started successfully")
	return nil
}

// Stop stops the MCP server
func (s *UnixSocketServer) Stop() {
	logging.Info("Stopping MCP Unix socket server")
	s.cancel()

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all connections
	s.connMutex.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.connMutex.Unlock()

	// Wait for goroutines to finish
	s.wg.Wait()

	// Clean up socket file
	os.Remove(s.socketPath)

	logging.Info("MCP Unix socket server stopped")
}

// acceptLoop accepts incoming connections
func (s *UnixSocketServer) acceptLoop() {
	defer s.wg.Done()

	logging.Debug("Starting Unix socket accept loop")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				// Server is shutting down
				return
			default:
				logging.Error("Accept error: %v", err)
				continue
			}
		}

		// Track connection
		s.connMutex.Lock()
		s.connections[conn] = true
		s.connMutex.Unlock()

		logging.Info("New Unix socket connection accepted")

		// Handle connection in goroutine
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single Unix socket connection
func (s *UnixSocketServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		conn.Close()
		s.connMutex.Lock()
		delete(s.connections, conn)
		s.connMutex.Unlock()
		logging.Info("Unix socket connection closed")
	}()

	// Create connection handler
	handler := &connectionHandler{
		server:      s,
		conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriter(conn),
		initialized: false,
	}

	// Read loop
	handler.readLoop()
}

// connectionHandler handles a single connection's messages
type connectionHandler struct {
	server      *UnixSocketServer
	conn        net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	writeMutex  sync.Mutex
	initialized bool
}

// readLoop reads JSON-RPC messages from the connection
func (h *connectionHandler) readLoop() {
	logging.Debug("Starting connection read loop")

	// Increase buffer size for large messages
	buf := make([]byte, 1024*1024) // 1MB buffer
	scanner := bufio.NewScanner(h.reader)
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
			h.sendError(messages.NewRequestID(nil), -32700, "Parse error", nil)
			continue
		}

		// Handle the message
		h.handleMessage(&msg)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		logging.Error("Error reading from connection: %v", err)
	}

	logging.Debug("Connection read loop ended")
}

// handleMessage routes messages to appropriate handlers
func (h *connectionHandler) handleMessage(msg *messages.JSONRPCMessage) {
	logging.Debug("Handling message: method=%s, id=%v", msg.Method, msg.ID)

	// Check if this is a notification (no ID or null ID)
	isNotification := msg.ID.IsEmpty()

	switch msg.Method {
	case "initialize":
		h.handleInitialize(msg)
	case "initialized", "notifications/initialized":
		// Client notification that initialization is complete
		logging.Debug("Client initialized")
		h.initialized = true
		// No response needed for notification
	case "tools/list":
		h.handleToolsList(msg)
	case "tools/call":
		h.handleToolsCall(msg)
	case "tasks/get":
		h.handleTasksGet(msg)
	case "tasks/result":
		h.handleTasksResult(msg)
	case "tasks/list":
		h.handleTasksList(msg)
	case "tasks/cancel":
		h.handleTasksCancel(msg)
	default:
		logging.Warn("Unknown method: %s", msg.Method)
		// Only send error response if this is a request (has an ID)
		if !isNotification {
			h.sendError(msg.ID, -32601, "Method not found", map[string]interface{}{
				"method": msg.Method,
			})
		}
	}
}

// handleInitialize handles the initialize request
func (h *connectionHandler) handleInitialize(msg *messages.JSONRPCMessage) {
	logging.Info("Handling initialize request")

	// Parse params
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid initialize params: %v", err)
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	// Call handler
	result, err := h.server.handler.HandleInitialize(params)
	if err != nil {
		logging.Error("Initialize handler failed: %v", err)
		h.sendError(msg.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Send response
	h.sendResponse(msg.ID, result)
	logging.Info("Initialize request handled successfully")
}

// handleToolsList handles the tools/list request
func (h *connectionHandler) handleToolsList(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tools/list request")

	// Parse params (may be nil)
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid tools/list params: %v", err)
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	// Call handler
	result, err := h.server.handler.HandleToolsList(params)
	if err != nil {
		logging.Error("Tools list handler failed: %v", err)
		h.sendError(msg.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Send response
	h.sendResponse(msg.ID, result)
	logging.Debug("Tools list request handled successfully")
}

// handleToolsCall handles the tools/call request
func (h *connectionHandler) handleToolsCall(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tools/call request")

	// Parse params
	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			logging.Error("Invalid tools/call params: %v", err)
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	// Call handler
	result, err := h.server.handler.HandleToolsCall(params)
	if err != nil {
		logging.Error("Tools call handler failed: %v", err)

		// Return error in MCP tool result format
		h.sendResponse(msg.ID, map[string]interface{}{
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
	h.sendResponse(msg.ID, result)
	logging.Debug("Tools call request handled successfully")
}

// handleTasksGet handles tasks/get requests
func (h *connectionHandler) handleTasksGet(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tasks/get request")

	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.server.handler.HandleTasksGet(params)
	if err != nil {
		h.sendError(msg.ID, -32603, err.Error(), nil)
		return
	}

	h.sendResponse(msg.ID, result)
}

// handleTasksResult handles tasks/result requests
func (h *connectionHandler) handleTasksResult(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tasks/result request")

	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.server.handler.HandleTasksResult(params)
	if err != nil {
		h.sendError(msg.ID, -32603, err.Error(), nil)
		return
	}

	h.sendResponse(msg.ID, result)
}

// handleTasksList handles tasks/list requests
func (h *connectionHandler) handleTasksList(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tasks/list request")

	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.server.handler.HandleTasksList(params)
	if err != nil {
		h.sendError(msg.ID, -32603, err.Error(), nil)
		return
	}

	h.sendResponse(msg.ID, result)
}

// handleTasksCancel handles tasks/cancel requests
func (h *connectionHandler) handleTasksCancel(msg *messages.JSONRPCMessage) {
	logging.Info("Handling tasks/cancel request")

	params := make(map[string]interface{})
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			h.sendError(msg.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.server.handler.HandleTasksCancel(params)
	if err != nil {
		h.sendError(msg.ID, -32603, err.Error(), nil)
		return
	}

	h.sendResponse(msg.ID, result)
}

// sendResponse sends a JSON-RPC response
func (h *connectionHandler) sendResponse(id messages.RequestID, result interface{}) {
	// Marshal result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logging.Error("Failed to marshal result: %v", err)
		h.sendError(id, -32603, "Internal error", nil)
		return
	}

	response := messages.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}

	h.writeMessage(&response)
}

// sendError sends a JSON-RPC error response
func (h *connectionHandler) sendError(id messages.RequestID, code int, message string, data interface{}) {
	var dataJSON json.RawMessage
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			logging.Error("Failed to marshal error data: %v", err)
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

	h.writeMessage(&response)
}

// writeMessage writes a JSON-RPC message to the connection
func (h *connectionHandler) writeMessage(msg *messages.JSONRPCMessage) {
	h.writeMutex.Lock()
	defer h.writeMutex.Unlock()

	// Marshal to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		logging.Error("Failed to marshal response: %v", err)
		return
	}

	// Write with newline
	data = append(data, '\n')

	logging.Debug("Sending message: %s", string(data))

	if _, err := h.writer.Write(data); err != nil {
		logging.Error("Failed to write to connection: %v", err)
		return
	}

	if err := h.writer.Flush(); err != nil {
		logging.Error("Failed to flush connection: %v", err)
		return
	}
}

// SendProgressNotification sends a progress notification (not used in Unix socket mode for now)
func (s *UnixSocketServer) SendProgressNotification(progressToken string, progress float64, total int, message string) {
	// Progress notifications over Unix sockets would need to be sent to the right connection
	// For now, log a warning - this feature can be implemented if needed
	logging.Debug("Progress notifications not yet implemented for Unix socket mode: token=%s, progress=%.2f",
		progressToken, progress)
}
