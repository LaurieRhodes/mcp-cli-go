package host

import (
	"fmt"
	"os"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages/initialize"
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
	connections      []*ServerConnection
	mu               sync.Mutex
	suppressConsole  bool  // Controls connection message visibility
}

// NewServerManager creates a new server manager
func NewServerManager() *ServerManager {
	logging.Debug("Creating new server manager")
	
	// Determine console suppression based on logging level
	// If logging is ERROR or above, suppress connection messages for clean output
	suppressConsole := logging.GetDefaultLevel() >= logging.ERROR
	
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
func (m *ServerManager) ConnectToServer(serverName string, serverConfig config.ServerConfig, userSpecified bool) (*ServerConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Info("Connecting to server: %s", serverName)
	logging.Debug("Server command: %s %v", serverConfig.Command, serverConfig.Args)

	// Create the stdio client with intelligent stderr handling
	params := stdio.StdioServerParameters{
		Command: serverConfig.Command,
		Args:    serverConfig.Args,
		Env:     serverConfig.Env,
	}
	client := stdio.NewStdioClientWithOptions(params, m.suppressConsole)

	// Start the client
	logging.Debug("Starting stdio client for server: %s", serverName)
	if err := client.Start(); err != nil {
		logging.Error("Failed to start server %s: %v", serverName, err)
		return nil, fmt.Errorf("failed to start server %s: %w", serverName, err)
	}

	// Send initialize request
	logging.Debug("Sending initialize request to server: %s", serverName)
	initResult, err := initialize.SendInitialize(client)
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
	
	// Load the configuration using the service
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logging.Error("Failed to load configuration: %v", err)
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logging.Debug("Loaded configuration with %d server entries", len(cfg.Servers))

	// Connect to each server
	for _, name := range serverNames {
		logging.Debug("Processing server: %s", name)
		
		// Get the server configuration using the wrapper function
		serverConfig, err := config.GetServerConfigFromConfig(cfg, name)
		if err != nil {
			logging.Warn("Server configuration not found for %s: %v", name, err)
			if !m.suppressConsole {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			continue
		}

		// Connect to the server
		isUserSpecified := userSpecified[name]
		_, err = m.ConnectToServer(name, serverConfig, isUserSpecified)
		if err != nil {
			logging.Warn("Failed to connect to server %s: %v", name, err)
			if !m.suppressConsole {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			continue
		}

		// Only print the connection message if console output is not suppressed
		if !m.suppressConsole {
			fmt.Printf("Connected to server: %s\n", name)
		}
	}

	// Check if we have any connections
	if len(m.connections) == 0 {
		logging.Error("Failed to connect to any servers")
		return fmt.Errorf("failed to connect to any servers")
	}

	logging.Info("Connected to %d servers successfully", len(m.connections))
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
