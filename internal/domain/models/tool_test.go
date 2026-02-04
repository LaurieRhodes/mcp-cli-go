package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestFunctionCallParseArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			args:    `{"key": "value", "num": 42}`,
			wantErr: false,
		},
		{
			name:    "empty object",
			args:    `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			args:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FunctionCall{
				Name:      "test_function",
				Arguments: json.RawMessage(tt.args),
			}

			args, err := fc.ParseArguments()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if args == nil {
				t.Error("Expected non-nil args")
			}
		})
	}
}

func TestToolResultIsSuccess(t *testing.T) {
	tests := []struct {
		name    string
		result  *ToolResult
		success bool
	}{
		{
			name: "success",
			result: &ToolResult{
				ToolCallID: "call-1",
				Output:     "success output",
				Error:      nil,
			},
			success: true,
		},
		{
			name: "failure",
			result: &ToolResult{
				ToolCallID: "call-2",
				Output:     "",
				Error:      errors.New("test error"),
			},
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsSuccess(); got != tt.success {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.success)
			}
		})
	}
}

func TestToolCreation(t *testing.T) {
	tool := Tool{
		Type: ToolTypeFunction,
		Function: ToolFunction{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	if tool.Type != ToolTypeFunction {
		t.Errorf("Expected type %s, got %s", ToolTypeFunction, tool.Type)
	}

	if tool.Function.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got %s", tool.Function.Name)
	}
}

func TestToolCallCreation(t *testing.T) {
	args := `{"param1": "value1"}`

	toolCall := ToolCall{
		ID:   "call-123",
		Type: ToolTypeFunction,
		Function: FunctionCall{
			Name:      "my_function",
			Arguments: json.RawMessage(args),
		},
	}

	if toolCall.ID != "call-123" {
		t.Errorf("Expected ID 'call-123', got %s", toolCall.ID)
	}

	if toolCall.Function.Name != "my_function" {
		t.Errorf("Expected function name 'my_function', got %s", toolCall.Function.Name)
	}

	parsedArgs, err := toolCall.Function.ParseArguments()
	if err != nil {
		t.Fatalf("Failed to parse arguments: %v", err)
	}

	if parsedArgs["param1"] != "value1" {
		t.Errorf("Expected param1='value1', got %v", parsedArgs["param1"])
	}
}
