package models

import "testing"

func TestUsageAdd(t *testing.T) {
	u1 := &Usage{
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
	}

	u2 := Usage{
		PromptTokens:     5,
		CompletionTokens: 10,
		TotalTokens:      15,
	}

	u1.Add(u2)

	if u1.PromptTokens != 15 {
		t.Errorf("Expected 15 prompt tokens, got %d", u1.PromptTokens)
	}

	if u1.CompletionTokens != 30 {
		t.Errorf("Expected 30 completion tokens, got %d", u1.CompletionTokens)
	}

	if u1.TotalTokens != 45 {
		t.Errorf("Expected 45 total tokens, got %d", u1.TotalTokens)
	}
}

func TestUsageIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
		empty bool
	}{
		{"empty", Usage{}, true},
		{"not empty", Usage{TotalTokens: 10}, false},
		{"only prompt", Usage{PromptTokens: 10}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.IsEmpty(); got != tt.empty {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.empty)
			}
		})
	}
}

func TestContextWindowAvailableTokens(t *testing.T) {
	cw := &ContextWindow{
		MaxTokens:     4096,
		ReserveTokens: 512,
	}

	available := cw.AvailableTokens()

	if available != 3584 {
		t.Errorf("Expected 3584 available tokens, got %d", available)
	}
}

func TestContextWindowCanFit(t *testing.T) {
	cw := &ContextWindow{
		MaxTokens:     1000,
		ReserveTokens: 200,
	}

	tests := []struct {
		name   string
		tokens int
		canFit bool
	}{
		{"fits", 500, true},
		{"exactly fits", 800, true},
		{"too large", 900, false},
		{"way too large", 2000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cw.CanFit(tt.tokens); got != tt.canFit {
				t.Errorf("CanFit(%d) = %v, want %v", tt.tokens, got, tt.canFit)
			}
		})
	}
}
