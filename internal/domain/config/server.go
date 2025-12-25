package config

// ServerConfig represents configuration for an MCP server
type ServerConfig struct {
	Command      string            `yaml:"command"`
	Args         []string          `yaml:"args"`
	Env          map[string]string `yaml:"env,omitempty"`
	SystemPrompt string            `yaml:"system_prompt,omitempty"`
	Settings     *ServerSettings   `yaml:"settings,omitempty"`
}

// ServerSettings contains server-specific settings
type ServerSettings struct {
	MaxToolFollowUp int  `yaml:"max_tool_follow_up,omitempty"`
	StrictMode      bool `yaml:"strict_mode,omitempty"`
}

// GetMaxToolFollowUp returns the max tool follow-up setting
func (s *ServerSettings) GetMaxToolFollowUp() int {
	if s == nil {
		return 0
	}
	return s.MaxToolFollowUp
}
