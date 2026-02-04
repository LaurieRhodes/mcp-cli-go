package config

import "fmt"

// ConfigError represents a configuration-related error
type ConfigError struct {
	Message string
	Context map[string]interface{}
	Cause   error
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) *ConfigError {
	return &ConfigError{
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *ConfigError) WithContext(key string, value interface{}) *ConfigError {
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *ConfigError) WithCause(cause error) *ConfigError {
	e.Cause = cause
	return e
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	msg := e.Message

	if len(e.Context) > 0 {
		msg += " ("
		first := true
		for k, v := range e.Context {
			if !first {
				msg += ", "
			}
			msg += fmt.Sprintf("%s: %v", k, v)
			first = false
		}
		msg += ")"
	}

	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause
func (e *ConfigError) Unwrap() error {
	return e.Cause
}
