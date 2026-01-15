package host

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	domainConfig "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/initialize"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/tools"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
)

// ServerConnection represents a connection to an MCP server
type ServerConnection struct {
	// Name of the server
	Name string

	// Client for communication with the server
	Client *stdio.StdioClient

	// Server info from initialize response
	ServerInfo initialize.ServerInfo

	// Server capabilities from initialize response
	Capabilities initialize.ServerCapabilities

	// Whether this server was explicitly requested by the user
	UserSpecified bool
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
		conn.Client.Stop()
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
		// Get tools from server using MCP protocol
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			logging.Warn("Failed to get tools from server %s: %v", conn.Name, err)
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
	}
	
	return allTools, nil
}

// ExecuteTool executes a tool on the appropriate server
func (m *ServerManager) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Find which server has this tool
	for _, conn := range m.connections {
		// Get tools list to check if this server has the tool
		result, err := tools.SendToolsList(conn.Client, nil)
		if err != nil {
			continue
		}
		
		// Check if this server has the tool
		for _, tool := range result.Tools {
			if tool.Name == toolName {
				// Execute the tool on this server
				logging.Debug("Executing tool %s on server %s", toolName, conn.Name)
				callResult, err := tools.SendToolsCall(conn.Client, conn.Client.GetDispatcher(), toolName, params)
				if err != nil {
					return "", fmt.Errorf("tool execution failed: %w", err)
				}
				
				// Check for error in result
				if callResult.IsError {
					return "", fmt.Errorf("tool error: %s", callResult.Error)
				}
				
				// Convert content to string
				// The content can be various types depending on the tool
				if callResult.Content == nil {
					return "", nil
				}
				
				// Try to convert content to a reasonable string representation
				switch v := callResult.Content.(type) {
				case string:
					return v, nil
				case map[string]interface{}, []interface{}:
					// Convert to JSON string
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						return "", fmt.Errorf("failed to marshal content: %w", err)
					}
					return string(jsonBytes), nil
				default:
					// For any other type, try JSON marshaling
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						// Fall back to string conversion
						return fmt.Sprintf("%v", v), nil
					}
					return string(jsonBytes), nil
				}
			}
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
			conn.Client.Stop()
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
