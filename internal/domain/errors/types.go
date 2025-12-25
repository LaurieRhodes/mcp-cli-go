package errors

import "fmt"

// DomainError represents a domain-level error
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause adds a cause to the error
func (e *DomainError) WithCause(err error) *DomainError {
	e.Cause = err
	return e
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code
	}
	return ErrCodeUnknown
}
