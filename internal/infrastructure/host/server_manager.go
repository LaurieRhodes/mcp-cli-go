package host

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	domainConfig "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/initialize"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/unixsocket"
)

// ServerConnection represents a connection to an MCP server
type ServerConnection struct {
	// Name of the server
	Name string

	// Client for communication with the server (can be stdio or Unix socket)
	Client interface{} // *stdio.StdioClient or *unixsocket.UnixSocketClient

	// Server info from initialize response
	ServerInfo initialize.ServerInfo

	// Server capabilities from initialize response
	Capabilities initialize.ServerCapabilities

	// Whether this server was explicitly requested by the user
	UserSpecified bool
}

// GetStdioClient returns the client as a stdio client if it is one, nil otherwise
func (sc *ServerConnection) GetStdioClient() *stdio.StdioClient {
	if stdioClient, ok := sc.Client.(*stdio.StdioClient); ok {
		return stdioClient
	}
	return nil
}

// GetUnixSocketClient returns the client as a Unix socket client if it is one, nil otherwise
func (sc *ServerConnection) GetUnixSocketClient() *unixsocket.UnixSocketClient {
	if socketClient, ok := sc.Client.(*unixsocket.UnixSocketClient); ok {
		return socketClient
	}
	return nil
}

// ServerManager manages connections to MCP servers
type ServerManager struct {
	connections     []*ServerConnection
	mu              sync.Mutex
	suppressConsole bool // Controls connection message visibility
}

// NewServerManager creates a new server manager
func NewServerManager() *ServerManager {
	logging.Debug("Creating new server manager")

	// Get output manager to determine console suppression
	outputMgr := output.GetGlobalManager()
	suppressConsole := !outputMgr.ShouldShowConnectionMessages()

	return &ServerManager{
		connections:     []*ServerConnection{},
		suppressConsole: suppressConsole,
	}
}

// NewServerManagerWithOptions creates a new server manager with custom options
func NewServerManagerWithOptions(suppressConsole bool) *ServerManager {
	logging.Debug("Creating new server manager with suppressConsole=%v", suppressConsole)
	return &ServerManager{
		connections:     []*ServerConnection{},
		suppressConsole: suppressConsole,
	}
}

// ConnectToServer connects to a server with the given configuration
func (m *ServerManager) ConnectToServer(serverName string, serverConfig domainConfig.ServerConfig, userSpecified bool) (*ServerConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Info("Connecting to server: %s", serverName)

	// NESTED MCP DETECTION: Check if we should use Unix socket instead of stdio
	if os.Getenv("MCP_NESTED") == "1" {
		logging.Info("Nested MCP context detected (MCP_NESTED=1)")

		// Try to get socket path from environment
		socketPath := os.Getenv(fmt.Sprintf("MCP_%s_SOCKET", strings.ToUpper(serverName)))

		if socketPath == "" {
			// Fallback: construct from socket directory
			socketDir := os.Getenv("MCP_SOCKET_DIR")
			if socketDir == "" {
				socketDir = "/tmp/mcp-sockets"
			}
			socketPath = filepath.Join(socketDir, serverName+".sock")
		}

		// Check if socket exists
		if _, err := os.Stat(socketPath); err == nil {
			logging.Info("Unix socket found: %s", socketPath)
			logging.Info("Attempting Unix socket connection (avoiding stdio conflict)")

			// Try Unix socket connection
			conn, err := m.connectViaUnixSocket(serverName, socketPath, userSpecified)
			if err != nil {
				logging.Warn("Unix socket connection failed: %v", err)
				logging.Info("Falling back to stdio connection")
				// Fall through to stdio connection attempt
			} else {
				logging.Info("Successfully connected via Unix socket")
				return conn, nil
			}
		} else {
			logging.Warn("Unix socket not found: %s", socketPath)
			logging.Info("Falling back to stdio connection")
		}
	}

	// Default: stdio connection
	logging.Debug("Using stdio connection for server: %s", serverName)
	logging.Debug("Server command: %s %v", serverConfig.Command, serverConfig.Args)

	// Get output manager for stderr suppression
	outputMgr := output.GetGlobalManager()
	suppressStderr := outputMgr.ShouldSuppressServerStderr()

	// Create the stdio client with intelligent stderr handling
	params := stdio.StdioServerParameters{
		Command: serverConfig.Command,
		Args:    serverConfig.Args,
		Env:     serverConfig.Env,
	}
	client := stdio.NewStdioClientWithStderrOption(params, suppressStderr)

	// Start the client
	logging.Debug("Starting stdio client for server: %s", serverName)
	if err := client.Start(); err != nil {
		logging.Error("Failed to start server %s: %v", serverName, err)
		return nil, fmt.Errorf("failed to start server %s: %w", serverName, err)
	}

	// Send initialize request
	logging.Debug("Sending initialize request to server: %s", serverName)
	initResult, err := initialize.SendInitialize(client, client.GetDispatcher())
	if err != nil {
		logging.Error("Failed to initialize server %s: %v", serverName, err)
		client.Stop()
		return nil, fmt.Errorf("failed to initialize server %s: %w", serverName, err)
	}

	// Create the connection
	conn := &ServerConnection{
		Name:          serverName,
		Client:        client,
		ServerInfo:    initResult.ServerInfo,
		Capabilities:  initResult.Capabilities,
		UserSpecified: userSpecified,
	}

	// Add to connections
	m.connections = append(m.connections, conn)
	logging.Info("Successfully connected to server: %s (%s v%s)",
		serverName, conn.ServerInfo.Name, conn.ServerInfo.Version)

	return conn, nil
}

// connectViaUnixSocket connects to a server via Unix domain socket
func (m *ServerManager) connectViaUnixSocket(serverName string, socketPath string, userSpecified bool) (*ServerConnection, error) {
	logging.Info("Connecting to %s via Unix socket: %s", serverName, socketPath)

	// Create Unix socket client
	client, err := unixsocket.NewUnixSocketClient(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Unix socket client: %w", err)
	}

	// Start the client
	logging.Debug("Starting Unix socket client for server: %s", serverName)
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Unix socket client: %w", err)
	}

	// Send initialize request using the Unix socket client's method
	logging.Debug("Sending initialize request via Unix socket to server: %s", serverName)
	initResponse, err := client.SendInitialize()
	if err != nil {
		logging.Error("Failed to initialize server %s via Unix socket: %v", serverName, err)
		client.Stop()
		return nil, fmt.Errorf("failed to initialize server %s: %w", serverName, err)
	}

	// Parse server info and capabilities from response
	var serverInfo initialize.ServerInfo
	var capabilities initialize.ServerCapabilities

	if si, ok := initResponse["serverInfo"].(map[string]interface{}); ok {
		if name, ok := si["name"].(string); ok {
			serverInfo.Name = name
		}
		if version, ok := si["version"].(string); ok {
			serverInfo.Version = version
		}
		if protocol, ok := si["protocolVersion"].(string); ok {
			serverInfo.ProtocolVersion = protocol
		}
	}

	if caps, ok := initResponse["capabilities"].(map[string]interface{}); ok {
		if tools, ok := caps["tools"].(map[string]interface{}); ok {
			capabilities.ProvidesTools = tools != nil
		}
		if prompts, ok := caps["prompts"].(map[string]interface{}); ok {
			capabilities.ProvidesPrompts = prompts != nil
		}
		if resources, ok := caps["resources"].(map[string]interface{}); ok {
			capabilities.ProvidesResources = resources != nil
		}
	}

	// Create the connection
	conn := &ServerConnection{
		Name:          serverName,
		Client:        client,
		ServerInfo:    serverInfo,
		Capabilities:  capabilities,
		UserSpecified: userSpecified,
	}

	// Add to connections
	m.connections = append(m.connections, conn)
	logging.Info("Successfully connected to server via Unix socket: %s (%s v%s)",
		serverName, conn.ServerInfo.Name, conn.ServerInfo.Version)

	return conn, nil
}

// ConnectToServers connects to multiple servers from the configuration
func (m *ServerManager) ConnectToServers(configFile string, serverNames []string, userSpecified map[string]bool) error {
	logging.Info("Connecting to servers from config file: %s", configFile)

	// Load the configuration using the new modular config service
	configService := config.NewService()
	appConfig, err := configService.LoadConfig(configFile)
	if err != nil {
		logging.Error("Failed to load configuration: %v", err)
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logging.Debug("Loaded configuration with %d server entries", len(appConfig.Servers))

	// Connect to each server
	for _, name := range serverNames {
		logging.Debug("Processing server: %s", name)

		// Get the server configuration
		serverConfig, exists := appConfig.Servers[name]
		if !exists {
			logging.Warn("Server configuration not found for %s", name)
			if !m.suppressConsole {
				fmt.Fprintf(os.Stderr, "Warning: server %s not found in configuration\n", name)
			}
			continue
		}

		// Connect to the server (now accepts domain config directly)
		isUserSpecified := userSpecified[name]
		_, err = m.ConnectToServer(name, serverConfig, isUserSpecified)
		if err != nil {
			logging.Warn("Failed to connect to server %s: %v", name, err)
			if !m.suppressConsole {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			continue
		}

		// Connection successful - no need to print message in normal mode
		// Message will be in logs if verbose mode is enabled
	}

	// Check if we have any connections
	// IMPORTANT: Allow zero connections when no servers were requested
	// This is valid for pure LLM queries that don't need MCP tools
	if len(serverNames) > 0 && len(m.connections) == 0 {
		logging.Error("Failed to connect to any of the requested servers")
		return fmt.Errorf("failed to connect to any of the requested servers")
	}

	if len(m.connections) == 0 {
		logging.Info("No server connections - running with LLM only")
	} else {
		logging.Info("Connected to %d server(s) successfully", len(m.connections))
	}

	return nil
}

// GetConnections returns all server connections
func (m *ServerManager) GetConnections() []*ServerConnection {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connections
}

// GetConnection returns the connection for the specified server name
func (m *ServerManager) GetConnection(name string) (*ServerConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, conn := range m.connections {
		if conn.Name == name {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("server %s not found", name)
}

// CloseConnections closes all connections
func (m *ServerManager) CloseConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Info("Closing all server connections")
	for _, conn := range m.connections {
		logging.Debug("Closing connection to server: %s", conn.Name)

		// Handle both stdio and Unix socket clients
		switch client := conn.Client.(type) {
		case *stdio.StdioClient:
			client.Stop()
		case *unixsocket.UnixSocketClient:
			client.Stop()
		default:
			logging.Warn("Unknown client type for server: %s", conn.Name)
		}
	}

	m.connections = []*ServerConnection{}
	logging.Debug("All server connections closed")
}

// SetSuppressConsole sets whether console output should be suppressed
func (m *ServerManager) SetSuppressConsole(suppress bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suppressConsole = suppress
}

// GetSuppressConsole returns whether console output is suppressed
func (m *ServerManager) GetSuppressConsole() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.suppressConsole
}

// Legacy method for backward compatibility - deprecated
// SetQuiet is deprecated, use SetSuppressConsole instead
func (m *ServerManager) SetQuiet(quiet bool) {
	logging.Warn("SetQuiet is deprecated, use SetSuppressConsole instead")
	m.SetSuppressConsole(quiet)
}

// GetAvailableTools returns all tools from all connected servers
func (m *ServerManager) GetAvailableTools() ([]domain.Tool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var allTools []domain.Tool

	for _, conn := range m.connections {
		// Handle both stdio and Unix socket clients
		var toolsList map[string]interface{}
		var err error

		switch client := conn.Client.(type) {
		case *stdio.StdioClient:
			// Get tools from server using MCP protocol
			result, e := tools.SendToolsList(client, nil)
			if e != nil {
				logging.Warn("Failed to get tools from server %s: %v", conn.Name, e)
				continue
			}
			// Convert MCP tools to domain tools
			for _, tool := range result.Tools {
				allTools = append(allTools, domain.Tool{
					Type: "function",
					Function: domain.ToolFunction{
						Name:        tool.Name,
						Description: tool.Description,
						Parameters:  tool.InputSchema,
					},
				})
			}
			continue

		case *unixsocket.UnixSocketClient:
			toolsList, err = client.SendToolsList(nil)
			if err != nil {
				logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
				continue
			}

		default:
			logging.Warn("Unknown client type for server: %s", conn.Name)
			continue
		}

		// Parse tools from Unix socket response
		if toolsArray, ok := toolsList["tools"].([]interface{}); ok {
			for _, t := range toolsArray {
				if toolMap, ok := t.(map[string]interface{}); ok {
					tool := domain.Tool{
						Type:     "function",
						Function: domain.ToolFunction{},
					}

					if name, ok := toolMap["name"].(string); ok {
						tool.Function.Name = name
					}
					if desc, ok := toolMap["description"].(string); ok {
						tool.Function.Description = desc
					}
					if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
						tool.Function.Parameters = schema
					}

					allTools = append(allTools, tool)
				}
			}
		}
	}

	return allTools, nil
}

// ExecuteTool executes a tool on the appropriate server
func (m *ServerManager) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find which server has this tool
	for _, conn := range m.connections {
		// Get tools list based on client type
		var hasToolResult bool

		switch client := conn.Client.(type) {
		case *stdio.StdioClient:
			result, err := tools.SendToolsList(client, nil)
			if err != nil {
				continue
			}
			// Check if this server has the tool
			for _, tool := range result.Tools {
				if tool.Name == toolName {
					hasToolResult = true
					break
				}
			}

		case *unixsocket.UnixSocketClient:
			result, err := client.SendToolsList(nil)
			if err != nil {
				continue
			}
			// Check if this server has the tool
			if toolsArray, ok := result["tools"].([]interface{}); ok {
				for _, t := range toolsArray {
					if toolMap, ok := t.(map[string]interface{}); ok {
						if name, ok := toolMap["name"].(string); ok && name == toolName {
							hasToolResult = true
							break
						}
					}
				}
			}

		default:
			continue
		}

		if !hasToolResult {
			continue
		}

		// Execute the tool on this server
		logging.Debug("Executing tool %s on server %s", toolName, conn.Name)

		switch client := conn.Client.(type) {
		case *stdio.StdioClient:
			callResult, err := tools.SendToolsCall(client, client.GetDispatcher(), toolName, params)
			if err != nil {
				return "", fmt.Errorf("tool execution failed: %w", err)
			}

			// Check for error in result
			if callResult.IsError {
				return "", fmt.Errorf("tool error: %s", callResult.Error)
			}

			// Convert content to string
			if callResult.Content == nil {
				return "", nil
			}

			// Try to convert content to a reasonable string representation
			switch v := callResult.Content.(type) {
			case string:
				return v, nil
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return "", fmt.Errorf("failed to marshal content: %w", err)
				}
				return string(jsonBytes), nil
			default:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return fmt.Sprintf("%v", v), nil
				}
				return string(jsonBytes), nil
			}

		case *unixsocket.UnixSocketClient:
			result, err := client.SendToolsCall(toolName, params)
			if err != nil {
				return "", fmt.Errorf("tool execution failed: %w", err)
			}

			// Check for error in result
			if isError, ok := result["isError"].(bool); ok && isError {
				if errMsg, ok := result["error"].(string); ok {
					return "", fmt.Errorf("tool error: %s", errMsg)
				}
				return "", fmt.Errorf("tool error (no message)")
			}

			// Convert content to string
			if content, ok := result["content"]; ok {
				switch v := content.(type) {
				case string:
					return v, nil
				case map[string]interface{}, []interface{}:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return "", fmt.Errorf("failed to marshal content: %w", err)
					}
					return string(jsonBytes), nil
				default:
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return fmt.Sprintf("%v", v), nil
					}
					return string(jsonBytes), nil
				}
			}

			return "", nil
		}
	}

	return "", fmt.Errorf("tool '%s' not found on any connected server", toolName)
}

// Additional methods to implement domain.MCPServerManager interface

// StartServer - not applicable for ServerManager (connections are established via ConnectToServer)
func (m *ServerManager) StartServer(ctx context.Context, serverName string, cfg *domainConfig.ServerConfig) (domain.MCPServer, error) {
	return nil, fmt.Errorf("StartServer not implemented - use ConnectToServer instead")
}

// StopServer stops a specific server connection
func (m *ServerManager) StopServer(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, conn := range m.connections {
		if conn.Name == serverName {
			// Handle both stdio and Unix socket clients
			switch client := conn.Client.(type) {
			case *stdio.StdioClient:
				client.Stop()
			case *unixsocket.UnixSocketClient:
				client.Stop()
			default:
				logging.Warn("Unknown client type for server: %s", serverName)
			}

			// Remove from connections slice
			m.connections = append(m.connections[:i], m.connections[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("server '%s' not found", serverName)
}

// GetServer retrieves a running server (ServerConnection implements MCPServer interface)
func (m *ServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		if conn.Name == serverName {
			// ServerConnection can be returned as MCPServer if it implements the interface
			// For now, return nil as ServerConnection doesn't implement MCPServer interface
			return nil, true
		}
	}

	return nil, false
}

// ListServers returns all running servers
func (m *ServerManager) ListServers() map[string]domain.MCPServer {
	m.mu.Lock()
	defer m.mu.Unlock()

	servers := make(map[string]domain.MCPServer)
	// ServerConnection doesn't implement MCPServer interface
	// This method is not critical for RAG functionality
	return servers
}

// StopAll stops all running servers
func (m *ServerManager) StopAll() error {
	m.CloseConnections()
	return nil
}
