package unixsocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// UnixSocketClient provides MCP communication over Unix domain sockets
type UnixSocketClient struct {
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	writeMutex sync.Mutex
	readMutex  sync.Mutex

	// Response handling
	pendingRequests map[string]chan json.RawMessage
	requestMutex    sync.Mutex

	// Connection state
	running  bool
	stopChan chan struct{}
}

// NewUnixSocketClient creates a new Unix socket MCP client
func NewUnixSocketClient(socketPath string) (*UnixSocketClient, error) {
	if socketPath == "" {
		return nil, fmt.Errorf("socket path cannot be empty")
	}

	// Check if socket exists
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("socket not found: %s: %w", socketPath, err)
	}

	return &UnixSocketClient{
		socketPath:      socketPath,
		pendingRequests: make(map[string]chan json.RawMessage),
		stopChan:        make(chan struct{}),
	}, nil
}

// Start connects to the Unix socket and starts the read loop
func (c *UnixSocketClient) Start() error {
	// Connect to Unix socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to socket %s: %w", c.socketPath, err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.running = true

	// Start read loop in background
	go c.readLoop()

	return nil
}

// Stop closes the Unix socket connection
func (c *UnixSocketClient) Stop() error {
	if !c.running {
		return nil
	}

	c.running = false
	close(c.stopChan)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// SendRequest sends a JSON-RPC request and waits for the response
func (c *UnixSocketClient) SendRequest(method string, params interface{}) (json.RawMessage, error) {
	if !c.running {
		return nil, fmt.Errorf("client not running")
	}

	// Generate request ID
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  method,
		"params":  params,
	}

	// Create response channel
	responseChan := make(chan json.RawMessage, 1)

	c.requestMutex.Lock()
	c.pendingRequests[requestID] = responseChan
	c.requestMutex.Unlock()

	// Serialize request
	data, err := json.Marshal(request)
	if err != nil {
		c.requestMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.requestMutex.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request (MCP uses newline-delimited JSON)
	c.writeMutex.Lock()
	_, err = c.writer.Write(append(data, '\n'))
	if err != nil {
		c.writeMutex.Unlock()
		c.requestMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.requestMutex.Unlock()
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	err = c.writer.Flush()
	c.writeMutex.Unlock()

	if err != nil {
		c.requestMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.requestMutex.Unlock()
		return nil, fmt.Errorf("failed to flush request: %w", err)
	}

	// Wait for response (with timeout)
	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(30 * time.Second):
		c.requestMutex.Lock()
		delete(c.pendingRequests, requestID)
		c.requestMutex.Unlock()
		return nil, fmt.Errorf("request timeout")
	case <-c.stopChan:
		return nil, fmt.Errorf("client stopped")
	}
}

// readLoop continuously reads responses from the socket
func (c *UnixSocketClient) readLoop() {
	for c.running {
		// Read line (newline-delimited JSON)
		c.readMutex.Lock()
		line, err := c.reader.ReadBytes('\n')
		c.readMutex.Unlock()

		if err != nil {
			if c.running {
				fmt.Fprintf(os.Stderr, "Unix socket read error: %v\n", err)
			}
			return
		}

		// Parse JSON-RPC response
		var response struct {
			ID     interface{}     `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  interface{}     `json:"error"`
		}

		if err := json.Unmarshal(line, &response); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
			continue
		}

		// Convert ID to string
		var requestID string
		switch id := response.ID.(type) {
		case string:
			requestID = id
		case float64:
			requestID = fmt.Sprintf("%.0f", id)
		default:
			continue
		}

		// Find pending request
		c.requestMutex.Lock()
		responseChan, exists := c.pendingRequests[requestID]
		if exists {
			delete(c.pendingRequests, requestID)
		}
		c.requestMutex.Unlock()

		if exists {
			// Check for error
			if response.Error != nil {
				// Send error as result (caller will handle)
				errorData, _ := json.Marshal(response.Error)
				responseChan <- errorData
			} else {
				responseChan <- response.Result
			}
		}
	}
}

// GetDispatcher returns a dispatcher for compatibility with existing code
// This is a compatibility shim - Unix socket client handles messaging directly
func (c *UnixSocketClient) GetDispatcher() interface{} {
	return c
}

// IsRunning returns whether the client is running
func (c *UnixSocketClient) IsRunning() bool {
	return c.running
}

// SendInitialize sends an MCP initialize request and parses the response
// This is a helper method to maintain compatibility with the existing code patterns
func (c *UnixSocketClient) SendInitialize() (map[string]interface{}, error) {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"clientInfo": map[string]interface{}{
			"name":    "mcp-cli-golang",
			"version": "0.1.0",
		},
		"capabilities": map[string]interface{}{},
	}

	response, err := c.SendRequest("initialize", params)
	if err != nil {
		return nil, fmt.Errorf("initialize request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse initialize response: %w", err)
	}

	return result, nil
}

// SendToolsList sends a tools/list request and parses the response
func (c *UnixSocketClient) SendToolsList(params interface{}) (map[string]interface{}, error) {
	response, err := c.SendRequest("tools/list", params)
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools/list response: %w", err)
	}

	return result, nil
}

// SendToolsCall sends a tools/call request and parses the response
func (c *UnixSocketClient) SendToolsCall(name string, arguments map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	response, err := c.SendRequest("tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools/call response: %w", err)
	}

	return result, nil
}
