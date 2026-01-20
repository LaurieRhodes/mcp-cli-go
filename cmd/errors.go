package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/agext/levenshtein"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CLIError represents a user-friendly CLI error
type CLIError struct {
	Type        string   // "unknown_flag", "missing_arg", "invalid_value", "unknown_command"
	Message     string   // The error message
	Suggestions []string // Possible corrections
	Examples    []string // Usage examples
	HelpCommand string   // Command to get more help
}

// Format returns a user-friendly error message
func (e *CLIError) Format() string {
	var sb strings.Builder

	// Check if we should use colors/emojis
	useColors := !noColor && isTerminal()
	useEmojis := useColors // Only use emojis if colors are enabled

	// Error header (red, prominent)
	if useColors {
		red := color.New(color.FgRed, color.Bold).SprintFunc()
		if useEmojis {
			sb.WriteString(fmt.Sprintf("%s %s\n\n", "❌", red("Error: "+e.Message)))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\n", red("Error: "+e.Message)))
		}
	} else {
		sb.WriteString(fmt.Sprintf("Error: %s\n\n", e.Message))
	}

	// Suggestions (if any)
	if len(e.Suggestions) > 0 {
		if useColors {
			yellow := color.New(color.FgYellow).SprintFunc()
			if useEmojis {
				sb.WriteString(yellow("💡 Did you mean:\n"))
			} else {
				sb.WriteString(yellow("Did you mean:\n"))
			}
		} else {
			sb.WriteString("Did you mean:\n")
		}

		for _, suggestion := range e.Suggestions {
			if useColors {
				green := color.New(color.FgGreen).SprintFunc()
				sb.WriteString(fmt.Sprintf("   %s\n", green(suggestion)))
			} else {
				sb.WriteString(fmt.Sprintf("   %s\n", suggestion))
			}
		}
		sb.WriteString("\n")
	}

	// Examples (if any)
	if len(e.Examples) > 0 {
		if useColors {
			cyan := color.New(color.FgCyan).SprintFunc()
			if useEmojis {
				sb.WriteString(cyan("✨ Examples:\n"))
			} else {
				sb.WriteString(cyan("Examples:\n"))
			}
		} else {
			sb.WriteString("Examples:\n")
		}

		for _, example := range e.Examples {
			sb.WriteString(fmt.Sprintf("   %s\n", example))
		}
		sb.WriteString("\n")
	}

	// Help command
	if e.HelpCommand != "" {
		if useColors {
			blue := color.New(color.FgBlue).SprintFunc()
			if useEmojis {
				sb.WriteString(fmt.Sprintf("%s %s\n", "📖", blue("For more info: "+e.HelpCommand)))
			} else {
				sb.WriteString(fmt.Sprintf("%s\n", blue("For more info: "+e.HelpCommand)))
			}
		} else {
			sb.WriteString(fmt.Sprintf("For more info: %s\n", e.HelpCommand))
		}
	}

	return sb.String()
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// FindSimilarFlags finds flags similar to the given name using Levenshtein distance
func FindSimilarFlags(flagName string, cmd *cobra.Command) []string {
	const maxDistance = 3 // Maximum Levenshtein distance to consider

	// Clean the flag name (remove leading dashes)
	cleanFlag := strings.TrimLeft(flagName, "-")

	var similar []string
	var bestDistance int = maxDistance + 1

	// Check local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		distance := levenshtein.Distance(cleanFlag, flag.Name, nil)
		if distance <= maxDistance {
			flagStr := fmt.Sprintf("--%s", flag.Name)
			if flag.Shorthand != "" {
				flagStr += fmt.Sprintf(", -%s", flag.Shorthand)
			}
			if flag.Usage != "" {
				flagStr += fmt.Sprintf("  (%s)", flag.Usage)
			}

			// Track best matches
			if distance < bestDistance {
				similar = []string{flagStr}
				bestDistance = distance
			} else if distance == bestDistance {
				similar = append(similar, flagStr)
			}
		}
	})

	// Check inherited flags if no close matches found
	if len(similar) == 0 {
		cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
			distance := levenshtein.Distance(cleanFlag, flag.Name, nil)
			if distance <= maxDistance {
				flagStr := fmt.Sprintf("--%s", flag.Name)
				if flag.Shorthand != "" {
					flagStr += fmt.Sprintf(", -%s", flag.Shorthand)
				}
				if flag.Usage != "" {
					flagStr += fmt.Sprintf("  (%s)", flag.Usage)
				}

				if distance < bestDistance {
					similar = []string{flagStr}
					bestDistance = distance
				} else if distance == bestDistance {
					similar = append(similar, flagStr)
				}
			}
		})
	}

	return similar
}

// FindSimilarCommands finds commands similar to the given name
func FindSimilarCommands(commandName string, cmd *cobra.Command) []string {
	const maxDistance = 2 // Stricter for commands

	var similar []string
	var bestDistance int = maxDistance + 1

	for _, subcmd := range cmd.Commands() {
		if !subcmd.Hidden {
			distance := levenshtein.Distance(commandName, subcmd.Name(), nil)
			if distance <= maxDistance {
				cmdStr := subcmd.Name()
				if subcmd.Short != "" {
					cmdStr += fmt.Sprintf("  (%s)", subcmd.Short)
				}

				if distance < bestDistance {
					similar = []string{cmdStr}
					bestDistance = distance
				} else if distance == bestDistance {
					similar = append(similar, cmdStr)
				}
			}

			// Also check aliases
			for _, alias := range subcmd.Aliases {
				distance := levenshtein.Distance(commandName, alias, nil)
				if distance <= maxDistance {
					cmdStr := fmt.Sprintf("%s (alias for %s)", alias, subcmd.Name())
					if distance < bestDistance {
						similar = []string{cmdStr}
						bestDistance = distance
					} else if distance == bestDistance {
						similar = append(similar, cmdStr)
					}
				}
			}
		}
	}

	return similar
}

// NewUnknownFlagError creates an error for unknown flags
func NewUnknownFlagError(flagName string, cmd *cobra.Command) *CLIError {
	similar := FindSimilarFlags(flagName, cmd)

	return &CLIError{
		Type:        "unknown_flag",
		Message:     fmt.Sprintf("Unknown flag: %s", flagName),
		Suggestions: similar,
		Examples: []string{
			fmt.Sprintf("mcp-cli %s --help", cmd.Name()),
		},
		HelpCommand: fmt.Sprintf("mcp-cli %s --help", cmd.Name()),
	}
}

// NewMissingArgumentError creates an error for missing arguments
func NewMissingArgumentError(argumentName string, commandName string, examples []string) *CLIError {
	return &CLIError{
		Type:        "missing_arg",
		Message:     fmt.Sprintf("Missing required argument: %s", argumentName),
		Examples:    examples,
		HelpCommand: fmt.Sprintf("mcp-cli %s --help", commandName),
	}
}

// NewUnknownCommandError creates an error for unknown commands
func NewUnknownCommandError(commandName string, parentCmd *cobra.Command) *CLIError {
	similar := FindSimilarCommands(commandName, parentCmd)

	// Get a few common commands to show
	commonCommands := []string{}
	for _, cmd := range parentCmd.Commands() {
		if !cmd.Hidden && len(commonCommands) < 5 {
			cmdStr := fmt.Sprintf("%-12s %s", cmd.Name(), cmd.Short)
			commonCommands = append(commonCommands, cmdStr)
		}
	}

	examples := []string{
		fmt.Sprintf("mcp-cli %s --help", parentCmd.Name()),
	}

	// Add common commands to examples if we have them
	if len(commonCommands) > 0 {
		examples = append([]string{"Common commands:"}, commonCommands...)
	}

	return &CLIError{
		Type:        "unknown_command",
		Message:     fmt.Sprintf("Unknown command: %s", commandName),
		Suggestions: similar,
		Examples:    examples,
		HelpCommand: "mcp-cli --help",
	}
}

// NewInvalidValueError creates an error for invalid flag values
func NewInvalidValueError(flagName string, value string, validValues []string, commandName string) *CLIError {
	suggestions := []string{}
	for _, v := range validValues {
		suggestions = append(suggestions, fmt.Sprintf("--%s %s", flagName, v))
	}

	return &CLIError{
		Type:        "invalid_value",
		Message:     fmt.Sprintf("Invalid value '%s' for flag --%s", value, flagName),
		Suggestions: suggestions,
		Examples: []string{
			fmt.Sprintf("mcp-cli %s --help", commandName),
		},
		HelpCommand: fmt.Sprintf("mcp-cli %s --help", commandName),
	}
}

// ExtractFlagName extracts the flag name from an error string
func ExtractFlagName(errStr string) string {
	// Handle "unknown flag: --flag-name" format
	if idx := strings.Index(errStr, "unknown flag:"); idx >= 0 {
		remaining := strings.TrimSpace(errStr[idx+13:])
		if len(remaining) > 0 {
			return strings.Fields(remaining)[0]
		}
	}

	// Handle "unknown shorthand flag: 'x' in -xyz" format
	if idx := strings.Index(errStr, "unknown shorthand flag:"); idx >= 0 {
		remaining := errStr[idx+23:]
		if len(remaining) > 2 {
			return "-" + string(remaining[2])
		}
	}

	return ""
}
