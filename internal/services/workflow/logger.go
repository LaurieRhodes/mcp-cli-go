package workflow

import (
	"fmt"
	"io"
	"os"
)

// LogLevel represents logging verbosity
type LogLevel int

const (
	LogError LogLevel = iota // Errors only
	LogWarn                  // + warnings
	LogInfo                  // + info
	LogSteps                 // + step-level workflow events (clean, semantic output)
	LogDebug                 // + debug messages
	LogVerbose               // + all internal operations (noisy)
)

// Logger handles workflow logging at different verbosity levels
type Logger struct {
	level  LogLevel
	output io.Writer
}

// NewLogger creates a new logger with the specified level
// If levelStr is empty and cliVerbose is true, uses verbose level
func NewLogger(levelStr string, cliVerbose bool) *Logger {
	var level LogLevel
	
	// Parse level string
	if levelStr != "" {
		switch levelStr {
		case "error":
			level = LogError
		case "warn":
			level = LogWarn
		case "info":
			level = LogInfo
		case "step", "steps":  // Accept both singular and plural
			level = LogSteps
		case "debug":
			level = LogDebug
		case "verbose":
			level = LogVerbose
		case "noisy": // Legacy alias for verbose
			level = LogVerbose
		default:
			level = LogInfo // Default to info for unknown levels
		}
	} else if cliVerbose {
		// CLI --verbose flag enables verbose logging
		level = LogVerbose
	} else {
		// Default: info
		level = LogInfo
	}

	return &Logger{
		level:  level,
		output: os.Stdout,
	}
}

// Error logs error messages (always visible except at level < error)
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level >= LogError {
		fmt.Fprintf(l.output, "[ERROR] "+format+"\n", args...)
	}
}

// Warn logs warning messages (visible at warn level and above)
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level >= LogWarn {
		fmt.Fprintf(l.output, "[WARN] "+format+"\n", args...)
	}
}

// Info logs informational messages (visible at info level and above)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LogInfo {
		fmt.Fprintf(l.output, "[INFO] "+format+"\n", args...)
	}
}

// Step logs step-level workflow events (visible at steps level and above)
// This provides clean, semantic output focused on workflow steps
func (l *Logger) Step(format string, args ...interface{}) {
	if l.level >= LogSteps {
		fmt.Fprintf(l.output, format+"\n", args...)
	}
}

// Debug logs debug messages (visible at debug level and above)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LogDebug {
		fmt.Fprintf(l.output, "[DEBUG] "+format+"\n", args...)
	}
}

// Verbose logs verbose internal operations (visible at verbose level only)
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.level >= LogVerbose {
		fmt.Fprintf(l.output, "[VERBOSE] "+format+"\n", args...)
	}
}

// Output logs Q&A output (always visible at all levels)
func (l *Logger) Output(format string, args ...interface{}) {
	fmt.Fprintf(l.output, format+"\n", args...)
}

// SetOutput sets the output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}
