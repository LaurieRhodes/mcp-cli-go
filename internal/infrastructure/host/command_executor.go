package host

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// CommandOptions provides configuration for command execution
type CommandOptions struct {
	SuppressConsole bool  // Suppress console output (connection messages, etc.)
	SuppressStderr  bool  // Suppress server stderr (not recommended - use only for truly quiet operation)
}

// DefaultCommandOptions returns the default command options
func DefaultCommandOptions() *CommandOptions {
	return &CommandOptions{
		SuppressConsole: false,  // Show connection messages by default
		SuppressStderr:  false,  // Always preserve server stderr for error handling
	}
}

// QuietCommandOptions returns options for quiet operation (suppresses console but preserves server errors)
func QuietCommandOptions() *CommandOptions {
	return &CommandOptions{
		SuppressConsole: true,   // Hide connection messages
		SuppressStderr:  false,  // Still preserve server stderr for error handling
	}
}

// SilentCommandOptions returns options for completely silent operation (not recommended)
func SilentCommandOptions() *CommandOptions {
	return &CommandOptions{
		SuppressConsole: true,  // Hide connection messages
		SuppressStderr:  true,  // Suppress server stderr (DANGEROUS - may hide critical errors)
	}
}

// RunCommand executes a function with connections to the specified servers
func RunCommand(commandFunc func([]*ServerConnection) error, configFile string, serverNames []string, userSpecified map[string]bool) error {
	return RunCommandWithOptions(commandFunc, configFile, serverNames, userSpecified, DefaultCommandOptions())
}

// RunCommandWithOptions executes a function with connections to the specified servers using custom options
func RunCommandWithOptions(commandFunc func([]*ServerConnection) error, configFile string, serverNames []string, userSpecified map[string]bool, options *CommandOptions) error {
	logging.Info("Running command with servers: %v", serverNames)
	
	// Create a context that can be cancelled - use a longer timeout for chat mode
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create the server manager with appropriate options
	var manager *ServerManager
	if options != nil {
		manager = NewServerManagerWithOptions(options.SuppressConsole)
	} else {
		manager = NewServerManager()
	}

	// Connect to the servers
	logging.Debug("Connecting to servers")
	if err := manager.ConnectToServers(configFile, serverNames, userSpecified); err != nil {
		logging.Error("Failed to connect to servers: %v", err)
		return err
	}

	// Ensure servers are closed when done
	logging.Debug("Setting up deferred server connection cleanup")
	defer manager.CloseConnections()

	// Get the connections
	connections := manager.GetConnections()
	if len(connections) == 0 {
		logging.Error("No valid server connections established")
		return fmt.Errorf("no valid server connections established")
	}

	// Apply stderr suppression if requested (not recommended)
	if options != nil && options.SuppressStderr {
		logging.Warn("Suppressing server stderr - this may hide critical error information")
		for _, conn := range connections {
			conn.Client.SetSuppressStderr(true)
		}
	}

	// Run the command with the connections
	logging.Debug("Starting command execution")
	errCh := make(chan error, 1)
	go func() {
		errCh <- commandFunc(connections)
	}()

	// Wait for the command to complete or context to be cancelled
	logging.Debug("Waiting for command to complete (timeout: 10m)")
	select {
	case err := <-errCh:
		if err != nil {
			logging.Error("Command execution failed: %v", err)
		} else {
			logging.Info("Command executed successfully")
		}
		return err
	case <-ctx.Done():
		logging.Error("Command execution timed out")
		return fmt.Errorf("command timed out")
	}
}

// ProcessOptions processes command-line options and returns the server names
// FIXED: Now dynamically loads available servers from configuration instead of hardcoding "filesystem"
func ProcessOptions(serverFlag string, disableFilesystem bool, provider string, model string) ([]string, map[string]bool) {
	logging.Debug("Processing options: server=%s, disableFilesystem=%v, provider=%s, model=%s",
		serverFlag, disableFilesystem, provider, model)
	
	// Parse the server list
	serverNames := []string{}
	if serverFlag != "" {
		// Split comma-separated list
		for _, s := range splitAndTrim(serverFlag, ",") {
			if s != "" {
				serverNames = append(serverNames, s)
			}
		}
	}

	// If no servers specified, dynamically load all available servers from configuration
	if len(serverNames) == 0 {
		serverNames = getAvailableServersFromConfig()
		if len(serverNames) == 0 {
			logging.Warn("No servers configured in configuration file")
			// Only show message if not in error-only logging mode
			if logging.GetDefaultLevel() < logging.ERROR {
				fmt.Fprintln(os.Stdout, "Warning: No servers configured in configuration file")
			}
		} else {
			logging.Info("No servers specified. Using all available servers from configuration: %v", serverNames)
			// Only show message if not in error-only logging mode
			if logging.GetDefaultLevel() < logging.ERROR {
				fmt.Fprintf(os.Stdout, "No servers specified. Using all available servers: %s\n", strings.Join(serverNames, ", "))
			}
		}
	}

	// Create a map of user-specified servers
	userSpecified := make(map[string]bool)
	for _, name := range serverNames {
		userSpecified[name] = true
	}

	logging.Debug("Final server names: %v", serverNames)
	return serverNames, userSpecified
}

// getAvailableServersFromConfig dynamically loads all available servers from configuration
func getAvailableServersFromConfig() []string {
	// Try to load from default config location
	configService := config.NewService()
	
	// Try standard config file locations
	configFiles := []string{
		"server_config.json",
		"config.json",
		"mcp-config.json",
	}
	
	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			_, err := configService.LoadConfig(configFile)
			if err != nil {
				logging.Debug("Failed to load config from %s: %v", configFile, err)
				continue
			}
			
			servers := configService.ListServers()
			if len(servers) > 0 {
				logging.Debug("Found %d servers in %s: %v", len(servers), configFile, servers)
				return servers
			}
		}
	}
	
	logging.Debug("No servers found in any configuration file")
	return []string{}
}

// Helper to split and trim a string
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
