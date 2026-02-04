package workflow

import (
	"testing"
)

func TestParseJSONL(t *testing.T) {
	le := &LoopExecutor{logger: NewLogger("normal", false)}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid JSONL",
			input: `{"id": "1", "name": "Alice"}
{"id": "2", "name": "Bob"}
{"id": "3", "name": "Charlie"}`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "JSONL with empty lines",
			input: `{"id": "1"}

{"id": "2"}

{"id": "3"}`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "single JSONL item",
			input:     `{"id": "1", "value": "test"}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid JSON line",
			input:     `{"id": "1"}\n{invalid json}`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := le.parseJSONL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) != tt.wantCount {
				t.Errorf("parseJSONL() got %d items, want %d", len(items), tt.wantCount)
			}
		})
	}
}

func TestParseJSONArray(t *testing.T) {
	le := &LoopExecutor{logger: NewLogger("normal", false)}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid JSON array",
			input: `[
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
				{"id": "3", "name": "Charlie"}
			]`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "compact JSON array",
			input:     `[{"id":"1"},{"id":"2"},{"id":"3"}]`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "single item array",
			input:     `[{"id": "1"}]`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty array",
			input:     `[]`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "invalid JSON",
			input:     `{not an array}`,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := le.parseJSONArray(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) != tt.wantCount {
				t.Errorf("parseJSONArray() got %d items, want %d", len(items), tt.wantCount)
			}
		})
	}
}

func TestParseTextLines(t *testing.T) {
	le := &LoopExecutor{logger: NewLogger("normal", false)}

	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{
			name: "multiple lines",
			input: `line 1
line 2
line 3`,
			wantCount: 3,
		},
		{
			name: "lines with empty lines",
			input: `line 1

line 2

line 3`,
			wantCount: 3,
		},
		{
			name:      "single line",
			input:     "just one line",
			wantCount: 1,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name: "lines with whitespace",
			input: `  line 1  
  line 2  
  line 3  `,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := le.parseTextLines(tt.input)
			if len(items) != tt.wantCount {
				t.Errorf("parseTextLines() got %d items, want %d", len(items), tt.wantCount)
			}
		})
	}
}

func TestParseArrayInput(t *testing.T) {
	le := &LoopExecutor{logger: NewLogger("normal", false)}

	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantErr    bool
		wantFormat string // "jsonl", "json", "text"
	}{
		{
			name: "detects JSONL",
			input: `{"id": "1"}
{"id": "2"}`,
			wantCount:  2,
			wantErr:    false,
			wantFormat: "jsonl",
		},
		{
			name:       "detects JSON array",
			input:      `[{"id": "1"}, {"id": "2"}]`,
			wantCount:  2,
			wantErr:    false,
			wantFormat: "json",
		},
		{
			name: "falls back to text lines",
			input: `item 1
item 2
item 3`,
			wantCount:  3,
			wantErr:    false,
			wantFormat: "text",
		},
		{
			name:      "empty input error",
			input:     "",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := le.parseArrayInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArrayInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) != tt.wantCount {
				t.Errorf("parseArrayInput() got %d items, want %d", len(items), tt.wantCount)
			}
		})
	}
}

func TestExtractItemID(t *testing.T) {
	le := &LoopExecutor{logger: NewLogger("normal", false)}

	tests := []struct {
		name     string
		item     interface{}
		index    int
		expected string
	}{
		{
			name: "extract id field",
			item: map[string]interface{}{
				"id":   "ITEM-123",
				"data": "test",
			},
			index:    5,
			expected: "ITEM-123",
		},
		{
			name: "extract control_id field",
			item: map[string]interface{}{
				"control_id": "CTRL-456",
				"data":       "test",
			},
			index:    5,
			expected: "CTRL-456",
		},
		{
			name: "extract name field",
			item: map[string]interface{}{
				"name": "MyItem",
				"data": "test",
			},
			index:    5,
			expected: "MyItem",
		},
		{
			name:     "no id field - use index",
			item:     map[string]interface{}{"data": "test"},
			index:    7,
			expected: "ITEM-007",
		},
		{
			name:     "string item - use index",
			item:     "plain string",
			index:    3,
			expected: "ITEM-003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := le.extractItemID(tt.item, tt.index)
			if result != tt.expected {
				t.Errorf("extractItemID() = %v, want %v", result, tt.expected)
			}
		})
	}
}
