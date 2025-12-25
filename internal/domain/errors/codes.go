package errors

// ErrorCode represents a specific error category
type ErrorCode string

const (
	// Configuration errors
	ErrCodeConfigInvalid     ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigNotFound    ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigParseFailed ErrorCode = "CONFIG_PARSE_FAILED"

	// Provider errors
	ErrCodeProviderNotFound ErrorCode = "PROVIDER_NOT_FOUND"
	ErrCodeProviderInvalid  ErrorCode = "PROVIDER_INVALID"
	ErrCodeProviderTimeout  ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAPIError ErrorCode = "PROVIDER_API_ERROR"

	// Tool errors
	ErrCodeToolNotFound       ErrorCode = "TOOL_NOT_FOUND"
	ErrCodeToolExecutionError ErrorCode = "TOOL_EXECUTION_ERROR"
	ErrCodeToolInvalidArgs    ErrorCode = "TOOL_INVALID_ARGS"

	// Request errors
	ErrCodeRequestInvalid  ErrorCode = "REQUEST_INVALID"
	ErrCodeRequestTooLarge ErrorCode = "REQUEST_TOO_LARGE"

	// Server errors
	ErrCodeServerNotFound    ErrorCode = "SERVER_NOT_FOUND"
	ErrCodeServerStartFailed ErrorCode = "SERVER_START_FAILED"
	ErrCodeServerStopped     ErrorCode = "SERVER_STOPPED"

	// Generic errors
	ErrCodeUnknown  ErrorCode = "UNKNOWN"
	ErrCodeInternal ErrorCode = "INTERNAL"
)

// IsRetryable returns true if the error code indicates a retryable error
func (ec ErrorCode) IsRetryable() bool {
	switch ec {
	case ErrCodeProviderTimeout, ErrCodeProviderAPIError:
		return true
	default:
		return false
	}
}

// HTTPStatusCode maps error codes to HTTP status codes
func (ec ErrorCode) HTTPStatusCode() int {
	switch ec {
	case ErrCodeConfigInvalid, ErrCodeRequestInvalid, ErrCodeToolInvalidArgs:
		return 400
	case ErrCodeProviderNotFound, ErrCodeToolNotFound, ErrCodeServerNotFound:
		return 404
	case ErrCodeProviderTimeout:
		return 408
	case ErrCodeRequestTooLarge:
		return 413
	case ErrCodeProviderAPIError, ErrCodeToolExecutionError:
		return 502
	default:
		return 500
	}
}
