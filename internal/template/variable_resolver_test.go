package template

import (
	"testing"
)

func TestVariableResolver_SetAndGet(t *testing.T) {
	resolver := NewVariableResolver()

	// Test basic set/get
	resolver.SetVariable("name", "Alice")
	value, ok := resolver.GetVariable("name")
	
	if !ok {
		t.Fatal("Variable not found")
	}
	
	if value != "Alice" {
		t.Errorf("Expected 'Alice', got '%v'", value)
	}
}

func TestVariableResolver_ResolveString(t *testing.T) {
	resolver := NewVariableResolver()
	resolver.SetVariable("name", "Bob")
	resolver.SetVariable("age", 30)

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "Hello {{name}}",
			expected: "Hello Bob",
		},
		{
			template: "{{name}} is {{age}} years old",
			expected: "Bob is 30 years old",
		},
		{
			template: "No variables here",
			expected: "No variables here",
		},
	}

	for _, tt := range tests {
		result, err := resolver.ResolveString(tt.template)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestVariableResolver_NestedAccess(t *testing.T) {
	resolver := NewVariableResolver()
	
	// Set nested data
	incident := map[string]interface{}{
		"id":       "INC-001",
		"severity": "high",
		"details": map[string]interface{}{
			"title": "Security Alert",
			"count": 5,
		},
	}
	resolver.SetVariable("incident", incident)

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "Incident: {{incident.id}}",
			expected: "Incident: INC-001",
		},
		{
			template: "Severity: {{incident.severity}}",
			expected: "Severity: high",
		},
	}

	for _, tt := range tests {
		result, err := resolver.ResolveString(tt.template)
		if err != nil {
			t.Errorf("Unexpected error for '%s': %v", tt.template, err)
		}
		if result != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestVariableResolver_ArrayAccess(t *testing.T) {
	resolver := NewVariableResolver()
	
	// Set array data
	items := []interface{}{"first", "second", "third"}
	resolver.SetVariable("items", items)

	tests := []struct {
		template string
		expected string
	}{
		{
			template: "First: {{items[0]}}",
			expected: "First: first",
		},
		{
			template: "Second: {{items[1]}}",
			expected: "Second: second",
		},
	}

	for _, tt := range tests {
		result, err := resolver.ResolveString(tt.template)
		if err != nil {
			t.Errorf("Unexpected error for '%s': %v", tt.template, err)
		}
		if result != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, result)
		}
	}
}

func TestVariableResolver_EvaluateCondition(t *testing.T) {
	resolver := NewVariableResolver()
	resolver.SetVariable("severity", "critical")
	resolver.SetVariable("count", "5")

	tests := []struct {
		condition string
		expected  bool
	}{
		{
			condition: "{{severity == 'critical'}}",
			expected:  true,
		},
		{
			condition: "{{severity == 'low'}}",
			expected:  false,
		},
		{
			condition: "{{severity != 'low'}}",
			expected:  true,
		},
	}

	for _, tt := range tests {
		result, err := resolver.EvaluateCondition(tt.condition)
		if err != nil {
			t.Errorf("Unexpected error for condition '%s': %v", tt.condition, err)
		}
		if result != tt.expected {
			t.Errorf("Condition '%s': expected %v, got %v", tt.condition, tt.expected, result)
		}
	}
}

func TestVariableResolver_SetStepOutput(t *testing.T) {
	resolver := NewVariableResolver()
	
	// Test setting step output
	resolver.SetStepOutput("analyze", "Analysis complete")
	
	value, ok := resolver.GetVariable("analyze")
	if !ok {
		t.Fatal("Step output not found")
	}
	
	if value != "Analysis complete" {
		t.Errorf("Expected 'Analysis complete', got '%v'", value)
	}
}

func TestVariableResolver_SetMultiple(t *testing.T) {
	resolver := NewVariableResolver()
	
	vars := map[string]interface{}{
		"name":  "Charlie",
		"age":   25,
		"active": true,
	}
	
	resolver.SetMultiple(vars)
	
	// Check all variables were set
	for key, expectedValue := range vars {
		value, ok := resolver.GetVariable(key)
		if !ok {
			t.Errorf("Variable '%s' not found", key)
		}
		if value != expectedValue {
			t.Errorf("Variable '%s': expected %v, got %v", key, expectedValue, value)
		}
	}
}
