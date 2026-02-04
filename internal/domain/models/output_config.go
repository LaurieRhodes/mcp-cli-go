package models

// OutputLevel defines the verbosity level for application output
type OutputLevel int

const (
	// OutputQuiet shows only essential output (errors, final results)
	OutputQuiet OutputLevel = iota

	// OutputNormal shows standard output (user-facing messages, results)
	OutputNormal

	// OutputVerbose shows detailed output (connection info, progress, debug)
	OutputVerbose
)

// String returns the string representation of the output level
func (l OutputLevel) String() string {
	switch l {
	case OutputQuiet:
		return "quiet"
	case OutputNormal:
		return "normal"
	case OutputVerbose:
		return "verbose"
	default:
		return "unknown"
	}
}

// OutputConfig defines how the application should handle output
type OutputConfig struct {
	// Level controls verbosity
	Level OutputLevel

	// ShowColors enables/disables colored output
	ShowColors bool

	// ShowProgress enables/disables progress indicators
	ShowProgress bool

	// ShowTimestamps adds timestamps to output
	ShowTimestamps bool

	// SuppressServerStderr suppresses MCP server stderr output
	SuppressServerStderr bool
}

// NewDefaultOutputConfig creates a default output configuration
func NewDefaultOutputConfig() *OutputConfig {
	return &OutputConfig{
		Level:                OutputNormal,
		ShowColors:           true,
		ShowProgress:         true,
		ShowTimestamps:       false,
		SuppressServerStderr: true, // Suppress server debug output in normal mode
	}
}

// NewQuietOutputConfig creates a quiet output configuration
func NewQuietOutputConfig() *OutputConfig {
	return &OutputConfig{
		Level:                OutputQuiet,
		ShowColors:           true,
		ShowProgress:         false,
		ShowTimestamps:       false,
		SuppressServerStderr: true,
	}
}

// NewVerboseOutputConfig creates a verbose output configuration
func NewVerboseOutputConfig() *OutputConfig {
	return &OutputConfig{
		Level:                OutputVerbose,
		ShowColors:           true,
		ShowProgress:         true,
		ShowTimestamps:       true,
		SuppressServerStderr: false,
	}
}

// ShouldShow determines if a message at the given level should be shown
func (c *OutputConfig) ShouldShow(level OutputLevel) bool {
	return level <= c.Level
}

// ShouldShowConnectionMessages returns true if connection messages should be displayed
func (c *OutputConfig) ShouldShowConnectionMessages() bool {
	return c.Level >= OutputVerbose
}

// ShouldShowStartupInfo returns true if startup information should be displayed
func (c *OutputConfig) ShouldShowStartupInfo() bool {
	return c.Level >= OutputVerbose
}

// ShouldShowProgress returns true if progress indicators should be shown
func (c *OutputConfig) ShouldShowProgress() bool {
	return c.ShowProgress && c.Level >= OutputNormal
}

// Clone creates a copy of the output configuration
func (c *OutputConfig) Clone() *OutputConfig {
	return &OutputConfig{
		Level:                c.Level,
		ShowColors:           c.ShowColors,
		ShowProgress:         c.ShowProgress,
		ShowTimestamps:       c.ShowTimestamps,
		SuppressServerStderr: c.SuppressServerStderr,
	}
}
