package config

// ChatConfig holds configuration for chat mode
type ChatConfig struct {
	// Default temperature for chat completions
	DefaultTemperature float64 `yaml:"default_temperature" json:"default_temperature"`

	// Maximum number of messages to keep in history
	MaxHistorySize int `yaml:"max_history_size" json:"max_history_size"`

	// Directory to store chat session logs (optional)
	// If set to a valid writable directory, sessions will be auto-saved
	// Format: YAML files named with session ID
	ChatLogsLocation string `yaml:"chat_logs_location" json:"chat_logs_location,omitempty"`

	// Whether to enable session logging (derived from ChatLogsLocation)
	SessionLoggingEnabled bool `yaml:"-" json:"-"`
}

// DefaultChatConfig returns default chat configuration
func DefaultChatConfig() *ChatConfig {
	return &ChatConfig{
		DefaultTemperature:    0.7,
		MaxHistorySize:        50,
		ChatLogsLocation:      "", // Empty = disabled
		SessionLoggingEnabled: false,
	}
}

// Validate checks if the chat config is valid
func (c *ChatConfig) Validate() error {
	if c.DefaultTemperature < 0 || c.DefaultTemperature > 2 {
		return NewConfigError("default_temperature must be between 0 and 2").
			WithContext("temperature", c.DefaultTemperature)
	}

	if c.MaxHistorySize < 1 {
		return NewConfigError("max_history_size must be greater than 0").
			WithContext("max_history_size", c.MaxHistorySize)
	}

	return nil
}
