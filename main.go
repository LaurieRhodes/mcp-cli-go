package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/LaurieRhodes/mcp-cli-go/cmd"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Version information - set at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func init() {
	// Set version information in cmd package
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.GitCommit = GitCommit
	
	// Setup signal handling
	setupSignalHandler()
}

func setupSignalHandler() {
	// Set up channel to receive signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		for sig := range c {
			fmt.Printf("\nInterrupted. Cleaning up...\n")
			logging.Info("Received signal: %v", sig)
			
			// This will let the OS know we handled the signal
			os.Exit(0)
		}
	}()
}

func main() {
	// Commands are automatically set up in their respective init() functions
	// and registered in cmd/root.go
	
	// Execute the root command
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
