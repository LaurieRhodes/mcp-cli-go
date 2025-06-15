package messages

import (
	"encoding/json"
	"fmt"
)

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object
type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewRequest creates a new JSON-RPC request message
func NewRequest(id, method string, params interface{}) (*JSONRPCMessage, error) {
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
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

// NewResponse creates a new JSON-RPC response message
func NewResponse(id string, result interface{}) (*JSONRPCMessage, error) {
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
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// NewError creates a new JSON-RPC error response
func NewError(id string, code int, message string, data interface{}) (*JSONRPCMessage, error) {
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
		ID:      id,
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
