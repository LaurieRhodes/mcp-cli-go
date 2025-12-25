package cmd

import (
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long: `Manage mcp-cli configuration files.

Available subcommands:
  validate - Validate configuration file and check for security issues

Examples:
  mcp-cli config validate
  mcp-cli config validate --config custom-config.yaml`,
}

func init() {
	// Add subcommands
	ConfigCmd.AddCommand(ConfigValidateCmd)
}
