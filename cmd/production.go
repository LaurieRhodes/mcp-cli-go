// production.go contains production environment settings
//go:build production
// +build production

package cmd

import (
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

func init() {
	// Set production mode for logging in production builds
	logging.SetProductionMode(true)
}
