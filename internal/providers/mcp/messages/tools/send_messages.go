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
	
	// Method name for progress notifications
	progressNotificationMethod = "notifications/progress"

	// Default timeout for tools requests
	// This is the timeout between progress updates, not total execution time
	// If server sends progress notifications, timeout resets on each update
	defaultToolsTimeout = 120 * time.Second  // Increased from 30s for skill execution
	
	// Maximum total time for a single tool call (safety limit)
	maxTotalToolCallTime = 30 * time.Minute
)

// SendToolsList sends a tools/list request to the server and returns the result
func SendToolsList(client *stdio.StdioClient, names []string) (*ToolsListResult, error) {
	logging.Debug("Sending tools/list request")
	
	// Create request
	requestID := fmt.Sprintf("tools-list-%d", time.Now().UnixNano())
	request, err := messages.NewRequest(requestID, toolsListMethod, map[string]interface{}{})
	if err != nil {
		logging.Error("Failed to create tools/list request: %v", err)
		return nil, fmt.Errorf("failed to create tools/list request: %w", err)
	}
	
	// Get dispatcher
	dispatcher := client.GetDispatcher()
	if dispatcher == nil {
		return nil, fmt.Errorf("client dispatcher not initialized")
	}
	
	// Register for response BEFORE sending request
	responseCh := dispatcher.RegisterRequest(requestID)
	defer dispatcher.UnregisterRequest(requestID) // Clean up on timeout or error
	
	// Send the request
	logging.Debug("Sending tools/list request with ID: %s", requestID)
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

	case <-time.After(defaultToolsTimeout):
		logging.Error("Timed out waiting for tools/list response for request %s", requestID)
		logging.Debug("Pending requests in dispatcher: %d", dispatcher.GetPendingCount())
		return nil, fmt.Errorf("timed out waiting for tools/list response")
	}
}

func SendToolsCall(client *stdio.StdioClient, dispatcher *stdio.ResponseDispatcher, name string, arguments map[string]interface{}) (*ToolsCallResult, error) {
	logging.Debug("Sending tools/call request for tool: %s", name)
	logging.Debug("Tool arguments: %+v", arguments)
	
	// Create the parameters with progress token
	// Include _meta.progressToken so server sends progress notifications
	progressToken := fmt.Sprintf("progress_%d", time.Now().UnixNano())
	
	paramsMap := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
		"_meta": map[string]interface{}{
			"progressToken": progressToken,
		},
	}

	// Create the request message
	requestID := fmt.Sprintf("tools_call_%s_%d", name, time.Now().UnixNano())
	request, err := messages.NewRequest(requestID, toolsCallMethod, paramsMap)
	if err != nil {
		logging.Error("Failed to create tools/call request: %v", err)
		return nil, fmt.Errorf("failed to create tools/call request: %w", err)
	}
	
	logging.Debug("Created tools/call request with progress token: %s", progressToken)
	
	// Debug: Show what's actually being sent
	if requestJSON, err := json.Marshal(request); err == nil {
		logging.Debug("Sending tools/call JSON: %s", string(requestJSON))
	}

	// Register with dispatcher BEFORE sending request
	responseCh := dispatcher.RegisterRequest(requestID)

	// Send the request
	logging.Debug("Sending tools/call request")
	if err := client.Write(request); err != nil {
		logging.Error("Failed to send tools/call request: %v", err)
		return nil, fmt.Errorf("failed to send tools/call request: %w", err)
	}
	logging.Debug("Tools/call request sent successfully")

	// Wait for response with timeout that resets on progress notifications
	logging.Debug("Waiting for tools/call response (timeout: %v between updates, max total: %v)", 
		defaultToolsTimeout, maxTotalToolCallTime)
	
	// Timer for timeout between progress updates
	timeoutTimer := time.NewTimer(defaultToolsTimeout)
	defer timeoutTimer.Stop()
	
	// Timer for maximum total execution time
	maxTimer := time.NewTimer(maxTotalToolCallTime)
	defer maxTimer.Stop()
	
	startTime := time.Now()
	
	for {
		select {
		case response := <-responseCh:
			elapsed := time.Since(startTime)
			logging.Debug("Received tools/call response after %v", elapsed)
			
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

		case <-timeoutTimer.C:
			elapsed := time.Since(startTime)
			logging.Error("Timed out waiting for tools/call response (total time: %v)", elapsed)
			return nil, fmt.Errorf("timed out waiting for tools/call response after %v", elapsed)
		
		case <-maxTimer.C:
			elapsed := time.Since(startTime)
			logging.Error("Maximum execution time exceeded for tools/call (%v)", elapsed)
			return nil, fmt.Errorf("maximum execution time exceeded: %v", elapsed)
		}
	}
}
