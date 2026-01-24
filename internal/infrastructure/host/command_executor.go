package host

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/output"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/transport/stdio"
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
	
	// Create a context that can be cancelled - use 30 minute timeout for workflow execution
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
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
	
	// IMPORTANT: Allow zero connections for pure LLM queries
	// Only log a warning, don't fail - the command function will decide if it needs servers
	if len(connections) == 0 {
		logging.Info("No server connections established - running with LLM only")
	}

	// Apply stderr suppression if requested (not recommended)
	if options != nil && options.SuppressStderr {
		logging.Warn("Suppressing server stderr - this may hide critical error information")
		for _, conn := range connections {
			// Only stdio clients support SetSuppressStderr
			if stdioClient, ok := conn.Client.(*stdio.StdioClient); ok {
				stdioClient.SetSuppressStderr(true)
			}
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
func ProcessOptions(configFile, serverFlag string, disableFilesystem bool, provider string, model string) ([]string, map[string]bool) {
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

	// If no servers specified and filesystem not disabled, load ALL servers from config
	if len(serverNames) == 0 && !disableFilesystem {
		// Use the new modular config service to load all servers
		configService := config.NewService()
		appConfig, err := configService.LoadConfig(configFile)
		if err == nil && appConfig != nil && len(appConfig.Servers) > 0 {
			// Add ALL configured servers
			for serverName := range appConfig.Servers {
				serverNames = append(serverNames, serverName)
				logging.Debug("Adding server from config: %s", serverName)
			}
			logging.Info("Loaded %d server(s) from config", len(serverNames))
			
			// Only show message if in verbose mode (to stderr, not stdout!)
			outputMgr := output.GetGlobalManager()
			if outputMgr.ShouldShowConnectionMessages() {
				fmt.Fprintf(os.Stderr, "Loading all %d configured servers.\n", len(serverNames))
			}
		} else {
			logging.Debug("No servers found in config or config load failed")
		}
		// If still no servers, leave empty - let caller handle it
	}

	// Create a map of user-specified servers
	userSpecified := make(map[string]bool)
	if serverFlag != "" {
		// Only servers explicitly specified via --server flag are marked as user-specified
		for _, name := range serverNames {
			userSpecified[name] = true
		}
	}
	// Auto-loaded servers from config are NOT marked as user-specified

	logging.Debug("Server names: %v", serverNames)
	return serverNames, userSpecified
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
