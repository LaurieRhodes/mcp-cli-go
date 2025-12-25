package console

import (
	"bytes"
	"testing"
)

func TestColors(t *testing.T) {
	// Disable colors for testing
	SetColorsEnabled(false)
	
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{"Bold", Bold, "test", "test"},
		{"Red", Red, "test", "test"},
		{"Green", Green, "test", "test"},
		{"Yellow", Yellow, "test", "test"},
		{"Blue", Blue, "test", "test"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
	
	// Re-enable colors
	SetColorsEnabled(true)
}

func TestColorsEnabled(t *testing.T) {
	SetColorsEnabled(true)
	
	result := Red("test")
	if result == "test" {
		t.Error("Expected colored output when colors enabled")
	}
	
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestFormatMessages(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		contains string
	}{
		{"Success", Success, "done", "✓"},
		{"Error", Error, "failed", "✗"},
		{"Warning", Warning, "careful", "⚠"},
		{"Info", Info, "note", "ℹ"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestPrintFunctions(t *testing.T) {
	// These just test that they don't panic
	PrintSuccess("Test %s", "success")
	PrintError("Test %s", "error")
	PrintWarning("Test %s", "warning")
	PrintInfo("Test %s", "info")
}

func TestChatUI(t *testing.T) {
	var buf bytes.Buffer
	ui := NewChatUI(&buf)
	
	ui.DisplayWelcome("openai", "gpt-4")
	
	if buf.Len() == 0 {
		t.Error("Expected output from DisplayWelcome")
	}
}

func TestChatUIMetadata(t *testing.T) {
	var buf bytes.Buffer
	ui := NewChatUI(&buf)
	
	// Test with metadata enabled
	ui.SetShowMeta(true)
	ui.DisplaySessionInfo("test-123", 5)
	
	if buf.Len() == 0 {
		t.Error("Expected output when metadata enabled")
	}
	
	// Test with metadata disabled
	buf.Reset()
	ui.SetShowMeta(false)
	ui.DisplaySessionInfo("test-123", 5)
	
	// Should still produce some output for session info
	// but usage display would be skipped
}
