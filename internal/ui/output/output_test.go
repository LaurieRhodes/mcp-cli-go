package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatterJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(FormatJSON, &buf)
	
	data := map[string]interface{}{
		"key": "value",
		"num": 42,
	}
	
	err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "key") {
		t.Error("Expected output to contain 'key'")
	}
}

func TestFormatterText(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(FormatText, &buf)
	
	err := formatter.Format("test output")
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "test output") {
		t.Error("Expected output to contain 'test output'")
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable(&buf, "Name", "Age", "City")
	
	table.AddRow("Alice", "30", "NYC")
	table.AddRow("Bob", "25", "LA")
	
	err := table.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Alice") {
		t.Error("Expected output to contain 'Alice'")
	}
	
	if !strings.Contains(output, "Name") {
		t.Error("Expected output to contain header 'Name'")
	}
}

func TestList(t *testing.T) {
	var buf bytes.Buffer
	list := NewList(&buf, ListStyleBullet)
	
	list.Add("Item 1")
	list.Add("Item 2")
	list.Add("Item 3")
	
	err := list.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Item 1") {
		t.Error("Expected output to contain 'Item 1'")
	}
	
	if !strings.Contains(output, "•") {
		t.Error("Expected bullet points")
	}
}

func TestListNumbered(t *testing.T) {
	var buf bytes.Buffer
	list := NewList(&buf, ListStyleNumbered)
	
	list.Add("First")
	list.Add("Second")
	
	err := list.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "1.") {
		t.Error("Expected numbered list")
	}
}

func TestKeyValue(t *testing.T) {
	var buf bytes.Buffer
	kv := NewKeyValue(&buf)
	
	kv.Add("Name", "Alice")
	kv.Add("Age", "30")
	kv.Add("City", "NYC")
	
	err := kv.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Name") {
		t.Error("Expected output to contain 'Name'")
	}
	
	if !strings.Contains(output, "Alice") {
		t.Error("Expected output to contain 'Alice'")
	}
}

func TestSuccessResponse(t *testing.T) {
	resp := NewSuccessResponse("Operation completed", map[string]string{"id": "123"})
	
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	
	if resp.Message != "Operation completed" {
		t.Errorf("Expected message 'Operation completed', got '%s'", resp.Message)
	}
}

func TestErrorResponse(t *testing.T) {
	err := &testError{msg: "test error"}
	resp := NewErrorResponse(err)
	
	if resp.Success {
		t.Error("Expected success to be false")
	}
	
	if resp.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", resp.Error)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
