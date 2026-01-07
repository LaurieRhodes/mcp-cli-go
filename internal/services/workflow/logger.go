package workflow

import (
	"fmt"
	"io"
	"os"
)

// LogLevel represents logging verbosity
type LogLevel int

const (
	LogNormal LogLevel = iota // Q&A only
	LogVerbose                // + operations
	LogNoisy                  // + everything
)

// Logger handles workflow logging at different verbosity levels
type Logger struct {
	level  LogLevel
	output io.Writer
}

// NewLogger creates a new logger with the specified level
func NewLogger(levelStr string) *Logger {
	var level LogLevel
	switch levelStr {
	case "verbose":
		level = LogVerbose
	case "noisy":
		level = LogNoisy
	default:
		level = LogNormal
	}

	return &Logger{
		level:  level,
		output: os.Stdout,
	}
}

// Info logs informational messages (visible at verbose level)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LogVerbose {
		fmt.Fprintf(l.output, "[INFO] "+format+"\n", args...)
	}
}

// Debug logs debug messages (visible at noisy level)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LogNoisy {
		fmt.Fprintf(l.output, "[DEBUG] "+format+"\n", args...)
	}
}

// Warn logs warning messages (visible at verbose level)
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level >= LogVerbose {
		fmt.Fprintf(l.output, "[WARN] "+format+"\n", args...)
	}
}

// Error logs error messages (always visible)
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(l.output, "[ERROR] "+format+"\n", args...)
}

// Output logs Q&A output (always visible)
func (l *Logger) Output(format string, args ...interface{}) {
	fmt.Fprintf(l.output, format+"\n", args...)
}

// SetOutput sets the output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}
