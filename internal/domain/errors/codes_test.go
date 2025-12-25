package errors

import "testing"

func TestErrorCodeIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		code      ErrorCode
		retryable bool
	}{
		{"provider timeout", ErrCodeProviderTimeout, true},
		{"provider API error", ErrCodeProviderAPIError, true},
		{"config invalid", ErrCodeConfigInvalid, false},
		{"tool not found", ErrCodeToolNotFound, false},
		{"request invalid", ErrCodeRequestInvalid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.IsRetryable(); got != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.retryable)
			}
		})
	}
}

func TestErrorCodeHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		code   ErrorCode
		status int
	}{
		{"config invalid", ErrCodeConfigInvalid, 400},
		{"request invalid", ErrCodeRequestInvalid, 400},
		{"tool invalid args", ErrCodeToolInvalidArgs, 400},
		{"provider not found", ErrCodeProviderNotFound, 404},
		{"tool not found", ErrCodeToolNotFound, 404},
		{"server not found", ErrCodeServerNotFound, 404},
		{"provider timeout", ErrCodeProviderTimeout, 408},
		{"request too large", ErrCodeRequestTooLarge, 413},
		{"provider API error", ErrCodeProviderAPIError, 502},
		{"tool execution error", ErrCodeToolExecutionError, 502},
		{"unknown", ErrCodeUnknown, 500},
		{"internal", ErrCodeInternal, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.HTTPStatusCode(); got != tt.status {
				t.Errorf("HTTPStatusCode() = %v, want %v", got, tt.status)
			}
		})
	}
}
