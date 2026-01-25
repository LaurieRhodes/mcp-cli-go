package cmd

import (
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"golang.org/x/term"
)

// redirectStdinIfNotTerminal redirects stdin to /dev/null if it's not a terminal
// This prevents blocking when called via MCP tools or other non-interactive contexts
func redirectStdinIfNotTerminal() {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Not a terminal - redirect to /dev/null to prevent blocking
		devNull, err := os.Open("/dev/null")
		if err != nil {
			logging.Warn("Failed to redirect stdin: %v", err)
			return
		}
		os.Stdin = devNull
		logging.Debug("Redirected stdin to /dev/null (non-terminal context detected)")
	}
}
