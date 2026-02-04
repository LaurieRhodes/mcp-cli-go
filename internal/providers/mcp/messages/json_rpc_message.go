package messages

import (
	"encoding/json"
	"fmt"
)

// RequestID can be either a string or number as per JSON-RPC spec
type RequestID struct {
	value interface{}
}

// UnmarshalJSON implements custom unmarshaling for RequestID
func (r *RequestID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a number first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		r.value = num
		return nil
	}

	// Try to unmarshal as a string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		r.value = str
		return nil
	}

	// Allow null/nil for notifications
	return nil
}

// MarshalJSON implements custom marshaling for RequestID
func (r RequestID) MarshalJSON() ([]byte, error) {
	if r.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(r.value)
}

// String returns the string representation of the ID
func (r RequestID) String() string {
	if r.value == nil {
		return ""
	}
	return fmt.Sprintf("%v", r.value)
}

// Equals compares two RequestIDs
func (r RequestID) Equals(other RequestID) bool {
	return r.value == other.value
}

// EqualsString compares RequestID with a string
func (r RequestID) EqualsString(s string) bool {
	if r.value == nil {
		return s == ""
	}
	return fmt.Sprintf("%v", r.value) == s
}

// IsEmpty returns true if the ID is empty/nil
func (r RequestID) IsEmpty() bool {
	return r.value == nil
}

// NewRequestID creates a new RequestID from any value
func NewRequestID(v interface{}) RequestID {
	return RequestID{value: v}
}

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      RequestID       `json:"id"` // Can be string, number, or null per JSON-RPC 2.0 spec
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// GetIDString returns the ID as a string for logging/comparison
func (m *JSONRPCMessage) GetIDString() string {
	return m.ID.String()
}

// JSONRPCError represents a JSON-RPC 2.0 error object
type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewRequest creates a new JSON-RPC request message
func NewRequest(id interface{}, method string, params interface{}) (*JSONRPCMessage, error) {
	var paramsJSON json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		paramsJSON = data
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      NewRequestID(id),
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

// NewResponse creates a new JSON-RPC response message
func NewResponse(id interface{}, result interface{}) (*JSONRPCMessage, error) {
	var resultJSON json.RawMessage
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		resultJSON = data
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      NewRequestID(id),
		Result:  resultJSON,
	}, nil
}

// NewError creates a new JSON-RPC error response
func NewError(id interface{}, code int, message string, data interface{}) (*JSONRPCMessage, error) {
	var dataJSON json.RawMessage
	if data != nil {
		d, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal error data: %w", err)
		}
		dataJSON = d
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      NewRequestID(id),
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    dataJSON,
		},
	}, nil
}

// IsRequest checks if the message is a request
func (m *JSONRPCMessage) IsRequest() bool {
	return m.Method != "" && m.Result == nil && m.Error == nil
}

// IsResponse checks if the message is a successful response
func (m *JSONRPCMessage) IsResponse() bool {
	return m.Method == "" && m.Result != nil && m.Error == nil
}

// IsError checks if the message is an error response
func (m *JSONRPCMessage) IsError() bool {
	return m.Method == "" && m.Result == nil && m.Error != nil
}

// UnmarshalParams unmarshals the params into the provided structure
func (m *JSONRPCMessage) UnmarshalParams(v interface{}) error {
	if m.Params == nil {
		return fmt.Errorf("no params to unmarshal")
	}
	return json.Unmarshal(m.Params, v)
}

// UnmarshalResult unmarshals the result into the provided structure
func (m *JSONRPCMessage) UnmarshalResult(v interface{}) error {
	if m.Result == nil {
		return fmt.Errorf("no result to unmarshal")
	}
	return json.Unmarshal(m.Result, v)
}
