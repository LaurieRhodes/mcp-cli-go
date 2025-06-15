package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
)

const (
	// Method names for tools requests
	toolsListMethod = "tools/list"
	toolsCallMethod = "tools/call"

	// Default timeout for tools requests
	defaultToolsTimeout = 30 * time.Second
)

// SendToolsList sends a tools/list request to the server and returns the result
func SendToolsList(client *stdio.StdioClient, names []string) (*ToolsListResult, error) {
	logging.Debug("Sending tools/list request with names: %v", names)
	
	// Create the parameters
	params := ToolsListParams{
		Names: names,
	}

	// Create the request message
	requestID := fmt.Sprintf("tools_list_%d", time.Now().UnixNano())
	request, err := messages.NewRequest(requestID, toolsListMethod, params)
	if err != nil {
		logging.Error("Failed to create tools/list request: %v", err)
		return nil, fmt.Errorf("failed to create tools/list request: %w", err)
	}

	// Send the request and wait for a response with a timeout
	responseCh := make(chan *messages.JSONRPCMessage, 1)
	errorCh := make(chan error, 1)
	doneCh := make(chan struct{})
	
	// Set up a goroutine to read responses
	go func() {
		defer close(doneCh)
		logging.Debug("Starting goroutine to listen for tools/list response")
		
		readCount := 0
		for msg := range client.Read() {
			readCount++
			logging.Debug("Received message #%d with ID: %s", readCount, msg.ID)
			if msg.ID == requestID {
				logging.Debug("Found matching response for request ID: %s", requestID)
				select {
				case responseCh <- msg:
					// Successfully sent the response
					return
				default:
					// This should never happen, but just in case
					logging.Error("Failed to send response to channel, it might be closed")
					return
				}
			} else {
				logging.Debug("Ignoring message with non-matching ID: %s (expected: %s)", msg.ID, requestID)
			}
		}
		logging.Error("Stdio client closed while waiting for tools/list response")
		select {
		case errorCh <- fmt.Errorf("stdio client closed while waiting for tools/list response"):
			// Successfully sent the error
		default:
			// Channel might be closed or full
			logging.Error("Failed to send error to channel, it might be closed")
		}
	}()

	// Send the request
	logging.Debug("Sending tools/list request")
	if err := client.Write(request); err != nil {
		logging.Error("Failed to send tools/list request: %v", err)
		return nil, fmt.Errorf("failed to send tools/list request: %w", err)
	}
	logging.Debug("Tools/list request sent successfully")

	// Wait for response with timeout
	logging.Debug("Waiting for tools/list response (timeout: %v)", defaultToolsTimeout)
	select {
	case response := <-responseCh:
		logging.Debug("Received tools/list response")
		
		// Check for errors
		if response.Error != nil {
			logging.Error("Server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
			return nil, fmt.Errorf("server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Parse the result
		var result ToolsListResult
		if err := json.Unmarshal(response.Result, &result); err != nil {
			logging.Error("Failed to parse tools/list result: %v", err)
			return nil, fmt.Errorf("failed to parse tools/list result: %w", err)
		}

		logging.Debug("Successfully received tools list with %d tools", len(result.Tools))
		return &result, nil

	case err := <-errorCh:
		logging.Error("Error during tools/list: %v", err)
		return nil, err

	case <-time.After(defaultToolsTimeout):
		logging.Error("Timed out waiting for tools/list response")
		return nil, fmt.Errorf("timed out waiting for tools/list response")
	}
}

// SendToolsCall sends a tools/call request to the server and returns the result
func SendToolsCall(client *stdio.StdioClient, name string, arguments map[string]interface{}) (*ToolsCallResult, error) {
	logging.Debug("Sending tools/call request for tool: %s", name)
	logging.Debug("Tool arguments: %+v", arguments)
	
	// Create the parameters
	params := ToolsCallParams{
		Name:      name,
		Arguments: arguments,
	}

	// Create the request message
	requestID := fmt.Sprintf("tools_call_%s_%d", name, time.Now().UnixNano())
	request, err := messages.NewRequest(requestID, toolsCallMethod, params)
	if err != nil {
		logging.Error("Failed to create tools/call request: %v", err)
		return nil, fmt.Errorf("failed to create tools/call request: %w", err)
	}

	// Send the request and wait for a response with a timeout
	responseCh := make(chan *messages.JSONRPCMessage, 1)
	errorCh := make(chan error, 1)
	doneCh := make(chan struct{})

	// Set up a goroutine to read responses
	go func() {
		defer close(doneCh)
		logging.Debug("Starting goroutine to listen for tools/call response")
		
		readCount := 0
		for msg := range client.Read() {
			readCount++
			logging.Debug("Received message #%d with ID: %s", readCount, msg.ID)
			if msg.ID == requestID {
				logging.Debug("Found matching response for request ID: %s", requestID)
				select {
				case responseCh <- msg:
					// Successfully sent the response
					return
				default:
					// This should never happen, but just in case
					logging.Error("Failed to send response to channel, it might be closed")
					return
				}
			} else {
				logging.Debug("Ignoring message with non-matching ID: %s (expected: %s)", msg.ID, requestID)
			}
		}
		logging.Error("Stdio client closed while waiting for tools/call response")
		select {
		case errorCh <- fmt.Errorf("stdio client closed while waiting for tools/call response"):
			// Successfully sent the error
		default:
			// Channel might be closed or full
			logging.Error("Failed to send error to channel, it might be closed")
		}
	}()

	// Send the request
	logging.Debug("Sending tools/call request")
	if err := client.Write(request); err != nil {
		logging.Error("Failed to send tools/call request: %v", err)
		return nil, fmt.Errorf("failed to send tools/call request: %w", err)
	}
	logging.Debug("Tools/call request sent successfully")

	// Wait for response with timeout
	logging.Debug("Waiting for tools/call response (timeout: %v)", defaultToolsTimeout)
	select {
	case response := <-responseCh:
		logging.Debug("Received tools/call response")
		
		// Check for errors
		if response.Error != nil {
			logging.Error("Server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
			return nil, fmt.Errorf("server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Parse the result
		var result ToolsCallResult
		if err := json.Unmarshal(response.Result, &result); err != nil {
			logging.Error("Failed to parse tools/call result: %v", err)
			return nil, fmt.Errorf("failed to parse tools/call result: %w", err)
		}

		if result.IsError {
			logging.Error("Tool execution failed: %s", result.Error)
		} else {
			logging.Debug("Tool execution successful")
		}

		return &result, nil

	case err := <-errorCh:
		logging.Error("Error during tools/call: %v", err)
		return nil, err

	case <-time.After(defaultToolsTimeout):
		logging.Error("Timed out waiting for tools/call response")
		return nil, fmt.Errorf("timed out waiting for tools/call response")
	}
}
