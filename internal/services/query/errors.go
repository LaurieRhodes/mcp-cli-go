package query

import (
	"errors"
	"fmt"
)

// Error codes as constants
const (
	ErrConfigNotFoundCode   = 10
	ErrProviderNotFoundCode = 11
	ErrModelNotFoundCode    = 12
	ErrServerConnectionCode = 13
	ErrToolExecutionCode    = 14
	ErrLLMRequestCode       = 15
	ErrContextNotFoundCode  = 16
	ErrInitializationCode   = 17
	ErrOutputFormatCode     = 18
	ErrOutputWriteCode      = 19
	ErrInvalidArgumentCode  = 20
)

// Error types with wrapped errors for error code mapping
var (
	ErrConfigNotFound   = errors.New("configuration not found")
	ErrProviderNotFound = errors.New("provider not found")
	ErrModelNotFound    = errors.New("model not found")
	ErrServerConnection = errors.New("server connection failed")
	ErrToolExecution    = errors.New("tool execution failed")
	ErrLLMRequest       = errors.New("LLM request failed")
	ErrContextNotFound  = errors.New("context file not found")
	ErrInitialization   = errors.New("query initialization failed")
	ErrOutputFormat     = errors.New("output formatting failed")
	ErrOutputWrite      = errors.New("output write failed")
	ErrInvalidArgument  = errors.New("invalid argument")
)

// Map errors to exit codes
var errorExitCodes = map[error]int{
	ErrConfigNotFound:   ErrConfigNotFoundCode,
	ErrProviderNotFound: ErrProviderNotFoundCode,
	ErrModelNotFound:    ErrModelNotFoundCode,
	ErrServerConnection: ErrServerConnectionCode,
	ErrToolExecution:    ErrToolExecutionCode,
	ErrLLMRequest:       ErrLLMRequestCode,
	ErrContextNotFound:  ErrContextNotFoundCode,
	ErrInitialization:   ErrInitializationCode,
	ErrOutputFormat:     ErrOutputFormatCode,
	ErrOutputWrite:      ErrOutputWriteCode,
	ErrInvalidArgument:  ErrInvalidArgumentCode,
}

// GetExitCode returns the appropriate exit code for an error
func GetExitCode(err error) int {
	// Check if the error is one of our known error types
	for errType, code := range errorExitCodes {
		if errors.Is(err, errType) {
			return code
		}
	}

	// Default error code for unknown errors
	return 1
}

// FormatError creates a formatted error with an appropriate wrapped error type
func FormatError(err error, errType error, format string, args ...interface{}) error {
	// Create the formatted error message
	formattedErr := fmt.Errorf(format, args...)

	// Wrap the original error with both the formatted message and the error type
	return fmt.Errorf("%w: %v: %w", errType, formattedErr, err)
}
