package main

import (
	"fmt"
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/cmd"
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

	// Note: Signal handling removed - Go runtime handles Ctrl-C naturally
	// and properly executes deferred cleanup functions (including terminal reset)
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
