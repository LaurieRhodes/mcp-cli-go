package errors

import (
	"errors"
	"testing"
)

func TestNewDomainError(t *testing.T) {
	err := NewDomainError(ErrCodeConfigInvalid, "test error")

	if err.Code != ErrCodeConfigInvalid {
		t.Errorf("Expected code %s, got %s", ErrCodeConfigInvalid, err.Code)
	}

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", err.Message)
	}

	if err.Context == nil {
		t.Error("Expected context to be initialized")
	}
}

func TestDomainErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		contains string
	}{
		{
			name: "without cause",
			err: &DomainError{
				Code:    ErrCodeConfigInvalid,
				Message: "config error",
			},
			contains: "[CONFIG_INVALID] config error",
		},
		{
			name: "with cause",
			err: &DomainError{
				Code:    ErrCodeConfigInvalid,
				Message: "config error",
				Cause:   errors.New("underlying error"),
			},
			contains: "underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			if errStr == "" {
				t.Error("Expected non-empty error string")
			}
		})
	}
}

func TestDomainErrorWithContext(t *testing.T) {
	err := NewDomainError(ErrCodeProviderInvalid, "test error")
	err.WithContext("provider", "openai").
		WithContext("model", "gpt-4")

	if err.Context["provider"] != "openai" {
		t.Errorf("Expected provider 'openai', got %v", err.Context["provider"])
	}

	if err.Context["model"] != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %v", err.Context["model"])
	}
}

func TestDomainErrorWithCause(t *testing.T) {
	cause := errors.New("root cause")
	err := NewDomainError(ErrCodeInternal, "wrapped error").
		WithCause(cause)

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Error("Expected Unwrap to return cause")
	}
}

func TestIsDomainError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		isDomain bool
	}{
		{
			name:     "domain error",
			err:      NewDomainError(ErrCodeInternal, "test"),
			isDomain: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			isDomain: false,
		},
		{
			name:     "nil error",
			err:      nil,
			isDomain: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDomainError(tt.err); got != tt.isDomain {
				t.Errorf("IsDomainError() = %v, want %v", got, tt.isDomain)
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code ErrorCode
	}{
		{
			name: "domain error",
			err:  NewDomainError(ErrCodeProviderNotFound, "not found"),
			code: ErrCodeProviderNotFound,
		},
		{
			name: "standard error",
			err:  errors.New("standard"),
			code: ErrCodeUnknown,
		},
		{
			name: "nil error",
			err:  nil,
			code: ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCode(tt.err); got != tt.code {
				t.Errorf("GetCode() = %v, want %v", got, tt.code)
			}
		})
	}
}
