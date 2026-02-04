package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information - set from main package
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print detailed version information including build time and git commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-cli version %s\n", Version)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", GitCommit)
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)

	// Also add --version flag to root command
	RootCmd.Version = Version
	RootCmd.SetVersionTemplate(fmt.Sprintf("mcp-cli version %s (built: %s, commit: %s)\n", Version, BuildTime, GitCommit))
}
