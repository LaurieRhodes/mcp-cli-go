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
	defaultToolsTimeout = 30 * time.Second
	
	// Maximum total time for a single tool call (safety limit)
	maxTotalToolCallTime = 30 * time.Minute
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
			if msg.ID.EqualsString(requestID) {
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

	// Send the request and wait for a response with timeout
	responseCh := make(chan *messages.JSONRPCMessage, 1)
	progressCh := make(chan *messages.JSONRPCMessage, 10) // Buffer for progress notifications
	errorCh := make(chan error, 1)
	doneCh := make(chan struct{})

	// Set up a goroutine to read responses and progress notifications
	go func() {
		defer close(doneCh)
		logging.Debug("Starting goroutine to listen for tools/call response and progress")
		
		readCount := 0
		for msg := range client.Read() {
			readCount++
			logging.Debug("Received message #%d with ID: %s, Method: %s", readCount, msg.ID, msg.Method)
			
			// Check if it's the final response
			if msg.ID.EqualsString(requestID) {
				logging.Debug("Found matching response for request ID: %s", requestID)
				select {
				case responseCh <- msg:
					return
				default:
					logging.Error("Failed to send response to channel, it might be closed")
					return
				}
			}
			
			// Check if it's a progress notification
			if msg.Method == progressNotificationMethod {
				logging.Debug("Received progress notification")
				select {
				case progressCh <- msg:
					// Sent progress notification
				default:
					logging.Warn("Progress notification channel full, dropping notification")
				}
			} else {
				logging.Debug("Ignoring message with non-matching ID: %s (expected: %s)", msg.ID, requestID)
			}
		}
		logging.Error("Stdio client closed while waiting for tools/call response")
		select {
		case errorCh <- fmt.Errorf("stdio client closed while waiting for tools/call response"):
		default:
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
	lastProgressTime := startTime
	
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

		case progress := <-progressCh:
			// Reset timeout on progress notification
			if !timeoutTimer.Stop() {
				select {
				case <-timeoutTimer.C:
				default:
				}
			}
			timeoutTimer.Reset(defaultToolsTimeout)
			
			// Log progress information
			if len(progress.Params) > 0 {
				var paramsMap map[string]interface{}
				if err := json.Unmarshal(progress.Params, &paramsMap); err == nil {
					progressValue, hasProgress := paramsMap["progress"]
					message, hasMessage := paramsMap["message"]
					timeSinceLastProgress := time.Since(lastProgressTime)
					lastProgressTime = time.Now()
					
					if hasProgress && hasMessage {
						logging.Info("Progress update: %.0f%% - %v (last update: %v ago)", 
							progressValue.(float64)*100, message, timeSinceLastProgress)
					} else if hasProgress {
						logging.Info("Progress update: %.0f%% (last update: %v ago)", 
							progressValue.(float64)*100, timeSinceLastProgress)
					} else {
						logging.Debug("Progress notification received (last update: %v ago)", timeSinceLastProgress)
					}
				}
			}

		case err := <-errorCh:
			logging.Error("Error during tools/call: %v", err)
			return nil, err

		case <-timeoutTimer.C:
			elapsed := time.Since(startTime)
			timeSinceProgress := time.Since(lastProgressTime)
			logging.Error("Timed out waiting for tools/call response (no progress for %v, total time: %v)", 
				timeSinceProgress, elapsed)
			return nil, fmt.Errorf("timed out waiting for tools/call response after %v with no progress for %v", 
				elapsed, timeSinceProgress)
		
		case <-maxTimer.C:
			elapsed := time.Since(startTime)
			logging.Error("Maximum execution time exceeded for tools/call (%v)", elapsed)
			return nil, fmt.Errorf("maximum execution time exceeded: %v", elapsed)
		}
	}
}
