package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG level for detailed troubleshooting
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for warnings that might require attention
	WARN
	// ERROR level for errors that don't terminate the application
	ERROR
	// FATAL level for critical errors that terminate the application
	FATAL
)

// ANSI color codes for different log levels
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"
	colorBright = "\033[1m"
)

var (
	// Default logger instance
	defaultLogger *Logger

	// Level names for display
	levelNames = map[LogLevel]string{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		FATAL: "FATAL",
	}

	// Color codes for each log level
	levelColors = map[LogLevel]string{
		DEBUG: colorGray,
		INFO:  colorBlue,
		WARN:  colorYellow,
		ERROR: colorRed,
		FATAL: colorRed + colorBright,
	}

	// Once ensures the default logger is initialized only once
	once sync.Once

	// colorEnabled controls whether color output is enabled
	colorEnabled = true
)

// Logger provides a simple logging facility
type Logger struct {
	level       LogLevel
	logger      *log.Logger
	mu          sync.Mutex
	colorOutput bool
}

// initDefaultLogger initializes the default logger
func initDefaultLogger() {
	defaultLogger = NewLogger(os.Stderr, INFO)
	// Enable colors by default on Windows 10+ and Unix systems
	defaultLogger.SetColorOutput(supportsColor())
}

// supportsColor determines if the current terminal supports ANSI colors
func supportsColor() bool {
	// Check for Windows
	if runtime.GOOS == "windows" {
		// Windows 10 version 1511 and later support ANSI colors
		// For simplicity, we'll assume modern Windows and enable colors
		// Users can disable if needed
		return true
	}
	
	// For Unix systems, check if we're outputting to a terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	return true
}

// NewLogger creates a new logger with the specified output and level
func NewLogger(out io.Writer, level LogLevel) *Logger {
	return &Logger{
		level:       level,
		logger:      log.New(out, "", log.LstdFlags),
		colorOutput: true,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetColorOutput enables or disables colored output
func (l *Logger) SetColorOutput(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorOutput = enabled
}

// IsColorOutputEnabled returns whether colored output is enabled
func (l *Logger) IsColorOutputEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.colorOutput
}

// formatLevel formats a log level with appropriate coloring
func (l *Logger) formatLevel(level LogLevel) string {
	levelName := levelNames[level]
	
	if !l.colorOutput {
		return fmt.Sprintf("[%s]", levelName)
	}
	
	color := levelColors[level]
	return fmt.Sprintf("%s[%s]%s", color, levelName, colorReset)
}

// log logs a message at the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	prefix := l.formatLevel(level) + " "
	msg := fmt.Sprintf(format, args...)
	l.logger.Print(prefix + msg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// GetDefaultLogger returns the default logger
func GetDefaultLogger() *Logger {
	once.Do(initDefaultLogger)
	return defaultLogger
}

// GetDefaultLevel returns the level of the default logger
func GetDefaultLevel() LogLevel {
	once.Do(initDefaultLogger)
	return defaultLogger.GetLevel()
}

// SetDefaultLevel sets the level for the default logger
func SetDefaultLevel(level LogLevel) {
	once.Do(initDefaultLogger)
	defaultLogger.SetLevel(level)
}

// SetColorOutput globally enables or disables colored output
func SetColorOutput(enabled bool) {
	once.Do(initDefaultLogger)
	defaultLogger.SetColorOutput(enabled)
	colorEnabled = enabled
}

// IsColorOutputEnabled returns whether colored output is globally enabled
func IsColorOutputEnabled() bool {
	once.Do(initDefaultLogger)
	return defaultLogger.IsColorOutputEnabled()
}

// Global logging functions

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	once.Do(initDefaultLogger)
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	once.Do(initDefaultLogger)
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	once.Do(initDefaultLogger)
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	once.Do(initDefaultLogger)
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(format string, args ...interface{}) {
	once.Do(initDefaultLogger)
	defaultLogger.Fatal(format, args...)
}
