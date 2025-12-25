package models

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Add combines two usage statistics
func (u *Usage) Add(other Usage) {
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalTokens += other.TotalTokens
}

// IsEmpty returns true if no tokens were used
func (u *Usage) IsEmpty() bool {
	return u.TotalTokens == 0
}

// ContextWindow represents context window constraints
type ContextWindow struct {
	MaxTokens     int `json:"max_tokens"`
	ReserveTokens int `json:"reserve_tokens"`
}

// AvailableTokens calculates tokens available for input
func (cw *ContextWindow) AvailableTokens() int {
	return cw.MaxTokens - cw.ReserveTokens
}

// CanFit checks if the given token count fits in the window
func (cw *ContextWindow) CanFit(tokens int) bool {
	return tokens <= cw.AvailableTokens()
}
