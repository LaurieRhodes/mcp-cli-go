package ports

import "github.com/LaurieRhodes/mcp-cli-go/internal/domain/models"

// OutputWriter defines the interface for writing output
type OutputWriter interface {
	// WriteInfo writes an informational message
	WriteInfo(format string, args ...interface{})
	
	// WriteSuccess writes a success message
	WriteSuccess(format string, args ...interface{})
	
	// WriteError writes an error message
	WriteError(format string, args ...interface{})
	
	// WriteWarning writes a warning message
	WriteWarning(format string, args ...interface{})
	
	// WriteDebug writes a debug message
	WriteDebug(format string, args ...interface{})
	
	// WriteLine writes a plain line
	WriteLine(format string, args ...interface{})
	
	// GetConfig returns the current output configuration
	GetConfig() *models.OutputConfig
	
	// SetConfig updates the output configuration
	SetConfig(config *models.OutputConfig)
}

// OutputManager manages output across the application
type OutputManager interface {
	// GetWriter returns the output writer
	GetWriter() OutputWriter
	
	// SetVerbosity sets the output verbosity level
	SetVerbosity(level models.OutputLevel)
	
	// GetVerbosity returns the current verbosity level
	GetVerbosity() models.OutputLevel
	
	// ShouldShowConnectionMessages returns true if connection messages should be shown
	ShouldShowConnectionMessages() bool
	
	// ShouldShowStartupInfo returns true if startup info should be shown
	ShouldShowStartupInfo() bool
	
	// ShouldSuppressServerStderr returns true if server stderr should be suppressed
	ShouldSuppressServerStderr() bool
}
