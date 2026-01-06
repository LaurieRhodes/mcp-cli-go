package config

// SettingsConfig represents global application settings
type SettingsConfig struct {
	// RawDataOverride determines if raw data output should override AI responses
	RawDataOverride bool `yaml:"raw_data_override,omitempty"`
	
	// MaxToolFollowUp sets the maximum number of follow-up tool calling requests (default: 2)
	MaxToolFollowUp int `yaml:"max_tool_follow_up,omitempty"`
	
	// OutputsDir is the directory where skill outputs are persisted
	OutputsDir string `yaml:"outputs_dir,omitempty"`
}

// ServerSettings represents server-specific settings
type ServerSettings struct {
	// RawDataOverride determines if raw data output should override AI responses for this server
	RawDataOverride bool `yaml:"raw_data_override,omitempty"`
	
	// MaxToolFollowUp sets the maximum number of follow-up tool calling requests for this server (default: 2)
	MaxToolFollowUp int `yaml:"max_tool_follow_up,omitempty"`
}

// GetMaxToolFollowUp returns the maximum tool follow-up attempts from settings, with fallback to default
func (s *SettingsConfig) GetMaxToolFollowUp() int {
	if s == nil || s.MaxToolFollowUp <= 0 {
		return 2 // Default value
	}
	return s.MaxToolFollowUp
}

// GetOutputsDir returns the outputs directory from settings, with fallback to default
func (s *SettingsConfig) GetOutputsDir() string {
	if s == nil || s.OutputsDir == "" {
		return "/tmp/mcp-outputs" // Default value
	}
	return s.OutputsDir
}

// GetMaxToolFollowUp returns the maximum tool follow-up attempts from server settings, with fallback to default
func (s *ServerSettings) GetMaxToolFollowUp() int {
	if s == nil || s.MaxToolFollowUp <= 0 {
		return 2 // Default value
	}
	return s.MaxToolFollowUp
}

// GetSettings returns the global settings from the config
func (c *Config) GetSettings() *SettingsConfig {
	return c.Settings
}

// GetServerSettings returns settings for a specific server
func (c *Config) GetServerSettings(serverName string) (*ServerSettings, error) {
	// Check if the server exists
	serverConfig, err := c.GetServerConfig(serverName)
	if err != nil {
		return nil, err
	}
	
	return serverConfig.Settings, nil
}
