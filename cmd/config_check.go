package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// checkConfigExists checks if the configuration file exists and shows a helpful error if not
func checkConfigExists(configPath string) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		printNoConfigError()
		os.Exit(1)
	}
}

// printNoConfigError prints a helpful error message when no configuration is found
func printNoConfigError() {
	errorColor := color.New(color.FgRed, color.Bold)
	infoColor := color.New(color.FgCyan)
	
	fmt.Fprintln(os.Stderr)
	errorColor.Fprintln(os.Stderr, "⚠️  No configuration found")
	fmt.Fprintln(os.Stderr)
	
	infoColor.Fprintln(os.Stderr, "Get started in 30 seconds:")
	fmt.Fprintln(os.Stderr, "  ./mcp-cli init --quick       (recommended, uses local ollama)")
	fmt.Fprintln(os.Stderr)
	
	infoColor.Fprintln(os.Stderr, "Or for guided setup:")
	fmt.Fprintln(os.Stderr, "  ./mcp-cli init               (interactive)")
	fmt.Fprintln(os.Stderr, "  ./mcp-cli init --full        (all options)")
	fmt.Fprintln(os.Stderr)
	
	infoColor.Fprintln(os.Stderr, "Need help?")
	fmt.Fprintln(os.Stderr, "  ./mcp-cli init --help")
	fmt.Fprintln(os.Stderr)
}

// isInitCommand checks if the current command is the init command
func isInitCommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	return os.Args[1] == "init"
}
