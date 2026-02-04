package initialize

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
)

const (
	// CurrentProtocolVersion is the version of the MCP protocol that this client implements
	CurrentProtocolVersion = "2024-05-01"

	// The method name for initialize requests
	initializeMethod = "initialize"

	// Default timeout for initialize requests
	defaultInitializeTimeout = 10 * time.Second
)

// DefaultClientInfo contains default information about this client
var DefaultClientInfo = ClientInfo{
	Name:    "mcp-cli-golang",
	Version: "0.1.0",
}

// SendInitialize sends an initialize request to the server and returns the result
func SendInitialize(client *stdio.StdioClient, dispatcher *stdio.ResponseDispatcher) (*InitializeResult, error) {
	logging.Info("Initializing MCP server connection")

	// Create initialize parameters
	params := InitializeParams{
		ProtocolVersion: CurrentProtocolVersion,
		ClientInfo:      DefaultClientInfo,
		Capabilities: ClientCapabilities{
			SupportsConfigurationChange: true,
			SupportsProgressReporting:   true,
			SupportsCancellation:        true,
		},
	}

	logging.Debug("Initialize parameters: protocolVersion=%s, clientInfo=%s/%s",
		params.ProtocolVersion, params.ClientInfo.Name, params.ClientInfo.Version)

	// Create the request message
	requestID := fmt.Sprintf("initialize_%d", time.Now().UnixNano())
	request, err := messages.NewRequest(requestID, initializeMethod, params)
	if err != nil {
		logging.Error("Failed to create initialize request: %v", err)
		return nil, fmt.Errorf("failed to create initialize request: %w", err)
	}

	logging.Debug("Created initialize request with ID: %s", requestID)

	// Register with dispatcher BEFORE sending request
	responseCh := dispatcher.RegisterRequest(requestID)

	// Send the request
	logging.Debug("Sending initialize request")
	if err := client.Write(request); err != nil {
		logging.Error("Failed to send initialize request: %v", err)
		return nil, fmt.Errorf("failed to send initialize request: %w", err)
	}
	logging.Debug("Initialize request sent successfully")

	// Wait for response with timeout
	logging.Debug("Waiting for initialize response (timeout: %v)", defaultInitializeTimeout)
	select {
	case response := <-responseCh:
		logging.Debug("Received initialize response")

		// Check for errors
		if response.Error != nil {
			logging.Error("Server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
			return nil, fmt.Errorf("server returned error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Parse the result
		var result InitializeResult
		if err := json.Unmarshal(response.Result, &result); err != nil {
			logging.Error("Failed to parse initialize result: %v", err)
			return nil, fmt.Errorf("failed to parse initialize result: %w", err)
		}

		logging.Info("Server initialized successfully: %s v%s (protocol: %s)",
			result.ServerInfo.Name, result.ServerInfo.Version, result.ServerInfo.ProtocolVersion)
		logging.Debug("Server capabilities: tools=%v, prompts=%v, resources=%v",
			result.Capabilities.ProvidesTools, result.Capabilities.ProvidesPrompts,
			result.Capabilities.ProvidesResources)

		return &result, nil

	case <-time.After(defaultInitializeTimeout):
		logging.Error("Timed out waiting for initialize response")
		return nil, fmt.Errorf("timed out waiting for initialize response")
	}
}
