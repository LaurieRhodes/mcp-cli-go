package domain

import (
	"errors"
	"fmt"
)

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// Configuration errors (1000-1099)
	ErrCodeConfigNotFound  ErrorCode = 1001
	ErrCodeConfigInvalid   ErrorCode = 1002
	ErrCodeConfigMigration ErrorCode = 1003

	// Provider errors (1100-1199)
	ErrCodeProviderNotFound      ErrorCode = 1101
	ErrCodeProviderUnsupported   ErrorCode = 1102
	ErrCodeProviderAuth          ErrorCode = 1103
	ErrCodeProviderConnection    ErrorCode = 1104
	ErrCodeProviderRateLimit     ErrorCode = 1105
	ErrCodeProviderQuotaExceeded ErrorCode = 1106

	// Model errors (1200-1299)
	ErrCodeModelNotFound     ErrorCode = 1201
	ErrCodeModelUnsupported  ErrorCode = 1202
	ErrCodeModelTokenLimit   ErrorCode = 1203
	ErrCodeModelContextLimit ErrorCode = 1204

	// Server errors (1300-1399)
	ErrCodeServerNotFound     ErrorCode = 1301
	ErrCodeServerStartFailed  ErrorCode = 1302
	ErrCodeServerConnection   ErrorCode = 1303
	ErrCodeServerTimeout      ErrorCode = 1304
	ErrCodeServerUnresponsive ErrorCode = 1305

	// Tool errors (1400-1499)
	ErrCodeToolNotFound    ErrorCode = 1401
	ErrCodeToolExecution   ErrorCode = 1402
	ErrCodeToolTimeout     ErrorCode = 1403
	ErrCodeToolInvalidArgs ErrorCode = 1404
	ErrCodeToolPermission  ErrorCode = 1405

	// Request errors (1500-1599)
	ErrCodeRequestInvalid   ErrorCode = 1501
	ErrCodeRequestTimeout   ErrorCode = 1502
	ErrCodeRequestTooLarge  ErrorCode = 1503
	ErrCodeRequestRateLimit ErrorCode = 1504

	// Response errors (1600-1699)
	ErrCodeResponseInvalid   ErrorCode = 1601
	ErrCodeResponseTimeout   ErrorCode = 1602
	ErrCodeResponseTooLarge  ErrorCode = 1603
	ErrCodeResponseMalformed ErrorCode = 1604

	// Session errors (1700-1799)
	ErrCodeSessionNotFound ErrorCode = 1701
	ErrCodeSessionExpired  ErrorCode = 1702
	ErrCodeSessionLimit    ErrorCode = 1703
	ErrCodeSessionInvalid  ErrorCode = 1704

	// System errors (1800-1899)
	ErrCodeSystemInitialization ErrorCode = 1801
	ErrCodeSystemShutdown       ErrorCode = 1802
	ErrCodeSystemResource       ErrorCode = 1803
	ErrCodeSystemPermission     ErrorCode = 1804

	// IO errors (1900-1999)
	ErrCodeIORead    ErrorCode = 1901
	ErrCodeIOWrite   ErrorCode = 1902
	ErrCodeIOFormat  ErrorCode = 1903
	ErrCodeIONetwork ErrorCode = 1904

	// Generic errors (2000+)
	ErrCodeInternal ErrorCode = 2000
	ErrCodeUnknown  ErrorCode = 2001
)

// DomainError represents a domain-specific error with structured information
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Retryable  bool                   `json:"retryable"`
	UserFacing bool                   `json:"user_facing"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison
func (e *DomainError) Is(target error) bool {
	if t, ok := target.(*DomainError); ok {
		return e.Code == t.Code
	}
	return false
}

// WithContext adds context information to the error
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *DomainError) WithCause(cause error) *DomainError {
	e.Cause = cause
	return e
}

// WithDetails adds additional details
func (e *DomainError) WithDetails(details string) *DomainError {
	e.Details = details
	return e
}

// GetExitCode returns the appropriate exit code for CLI usage
func (e *DomainError) GetExitCode() int {
	switch {
	case e.Code >= 1000 && e.Code < 1100: // Config errors
		return 10
	case e.Code >= 1100 && e.Code < 1200: // Provider errors
		return 11
	case e.Code >= 1200 && e.Code < 1300: // Model errors
		return 12
	case e.Code >= 1300 && e.Code < 1400: // Server errors
		return 13
	case e.Code >= 1400 && e.Code < 1500: // Tool errors
		return 14
	case e.Code >= 1500 && e.Code < 1600: // Request errors
		return 15
	case e.Code >= 1600 && e.Code < 1700: // Response errors
		return 16
	case e.Code >= 1700 && e.Code < 1800: // Session errors
		return 17
	case e.Code >= 1800 && e.Code < 1900: // System errors
		return 18
	case e.Code >= 1900 && e.Code < 2000: // IO errors
		return 19
	default: // Generic errors
		return 1
	}
}

// Predefined error instances for common cases
var (
	// Configuration errors
	ErrConfigNotFound  = &DomainError{Code: ErrCodeConfigNotFound, Message: "configuration not found", UserFacing: true}
	ErrConfigInvalid   = &DomainError{Code: ErrCodeConfigInvalid, Message: "configuration is invalid", UserFacing: true}
	ErrConfigMigration = &DomainError{Code: ErrCodeConfigMigration, Message: "configuration migration failed", UserFacing: true}

	// Provider errors
	ErrProviderNotFound      = &DomainError{Code: ErrCodeProviderNotFound, Message: "provider not found", UserFacing: true}
	ErrProviderUnsupported   = &DomainError{Code: ErrCodeProviderUnsupported, Message: "provider not supported", UserFacing: true}
	ErrProviderAuth          = &DomainError{Code: ErrCodeProviderAuth, Message: "provider authentication failed", UserFacing: true}
	ErrProviderConnection    = &DomainError{Code: ErrCodeProviderConnection, Message: "provider connection failed", Retryable: true, UserFacing: true}
	ErrProviderRateLimit     = &DomainError{Code: ErrCodeProviderRateLimit, Message: "provider rate limit exceeded", Retryable: true, UserFacing: true}
	ErrProviderQuotaExceeded = &DomainError{Code: ErrCodeProviderQuotaExceeded, Message: "provider quota exceeded", UserFacing: true}

	// Model errors
	ErrModelNotFound     = &DomainError{Code: ErrCodeModelNotFound, Message: "model not found", UserFacing: true}
	ErrModelUnsupported  = &DomainError{Code: ErrCodeModelUnsupported, Message: "model not supported", UserFacing: true}
	ErrModelTokenLimit   = &DomainError{Code: ErrCodeModelTokenLimit, Message: "model token limit exceeded", UserFacing: true}
	ErrModelContextLimit = &DomainError{Code: ErrCodeModelContextLimit, Message: "model context limit exceeded", UserFacing: true}

	// Server errors
	ErrServerNotFound     = &DomainError{Code: ErrCodeServerNotFound, Message: "server not found", UserFacing: true}
	ErrServerStartFailed  = &DomainError{Code: ErrCodeServerStartFailed, Message: "server start failed", Retryable: true, UserFacing: true}
	ErrServerConnection   = &DomainError{Code: ErrCodeServerConnection, Message: "server connection failed", Retryable: true, UserFacing: true}
	ErrServerTimeout      = &DomainError{Code: ErrCodeServerTimeout, Message: "server timeout", Retryable: true, UserFacing: true}
	ErrServerUnresponsive = &DomainError{Code: ErrCodeServerUnresponsive, Message: "server unresponsive", Retryable: true, UserFacing: true}

	// Tool errors
	ErrToolNotFound    = &DomainError{Code: ErrCodeToolNotFound, Message: "tool not found", UserFacing: true}
	ErrToolExecution   = &DomainError{Code: ErrCodeToolExecution, Message: "tool execution failed", Retryable: true, UserFacing: true}
	ErrToolTimeout     = &DomainError{Code: ErrCodeToolTimeout, Message: "tool timeout", Retryable: true, UserFacing: true}
	ErrToolInvalidArgs = &DomainError{Code: ErrCodeToolInvalidArgs, Message: "tool invalid arguments", UserFacing: true}
	ErrToolPermission  = &DomainError{Code: ErrCodeToolPermission, Message: "tool permission denied", UserFacing: true}

	// Request errors
	ErrRequestInvalid   = &DomainError{Code: ErrCodeRequestInvalid, Message: "request invalid", UserFacing: true}
	ErrRequestTimeout   = &DomainError{Code: ErrCodeRequestTimeout, Message: "request timeout", Retryable: true, UserFacing: true}
	ErrRequestTooLarge  = &DomainError{Code: ErrCodeRequestTooLarge, Message: "request too large", UserFacing: true}
	ErrRequestRateLimit = &DomainError{Code: ErrCodeRequestRateLimit, Message: "request rate limit exceeded", Retryable: true, UserFacing: true}

	// Response errors
	ErrResponseInvalid   = &DomainError{Code: ErrCodeResponseInvalid, Message: "response invalid", Retryable: true}
	ErrResponseTimeout   = &DomainError{Code: ErrCodeResponseTimeout, Message: "response timeout", Retryable: true, UserFacing: true}
	ErrResponseTooLarge  = &DomainError{Code: ErrCodeResponseTooLarge, Message: "response too large", UserFacing: true}
	ErrResponseMalformed = &DomainError{Code: ErrCodeResponseMalformed, Message: "response malformed", Retryable: true}

	// Session errors
	ErrSessionNotFound = &DomainError{Code: ErrCodeSessionNotFound, Message: "session not found", UserFacing: true}
	ErrSessionExpired  = &DomainError{Code: ErrCodeSessionExpired, Message: "session expired", UserFacing: true}
	ErrSessionLimit    = &DomainError{Code: ErrCodeSessionLimit, Message: "session limit exceeded", UserFacing: true}
	ErrSessionInvalid  = &DomainError{Code: ErrCodeSessionInvalid, Message: "session invalid", UserFacing: true}

	// System errors
	ErrSystemInitialization = &DomainError{Code: ErrCodeSystemInitialization, Message: "system initialization failed"}
	ErrSystemShutdown       = &DomainError{Code: ErrCodeSystemShutdown, Message: "system shutdown failed"}
	ErrSystemResource       = &DomainError{Code: ErrCodeSystemResource, Message: "system resource unavailable", Retryable: true}
	ErrSystemPermission     = &DomainError{Code: ErrCodeSystemPermission, Message: "system permission denied", UserFacing: true}

	// IO errors
	ErrIORead    = &DomainError{Code: ErrCodeIORead, Message: "IO read failed", Retryable: true}
	ErrIOWrite   = &DomainError{Code: ErrCodeIOWrite, Message: "IO write failed", Retryable: true}
	ErrIOFormat  = &DomainError{Code: ErrCodeIOFormat, Message: "IO format error", UserFacing: true}
	ErrIONetwork = &DomainError{Code: ErrCodeIONetwork, Message: "network IO failed", Retryable: true, UserFacing: true}

	// Generic errors
	ErrInternal = &DomainError{Code: ErrCodeInternal, Message: "internal error"}
	ErrUnknown  = &DomainError{Code: ErrCodeUnknown, Message: "unknown error"}
)

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an existing error as a domain error
func WrapError(err error, code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if domainErr, ok := IsDomainError(err); ok {
		return domainErr.Retryable
	}
	return false
}

// IsUserFacing checks if an error should be shown to users
func IsUserFacing(err error) bool {
	if domainErr, ok := IsDomainError(err); ok {
		return domainErr.UserFacing
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if domainErr, ok := IsDomainError(err); ok {
		return domainErr.Code
	}
	return ErrCodeUnknown
}

// GetExitCode extracts the exit code from an error
func GetExitCode(err error) int {
	if domainErr, ok := IsDomainError(err); ok {
		return domainErr.GetExitCode()
	}
	return 1
}
