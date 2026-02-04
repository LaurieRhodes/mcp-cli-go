package models

import "testing"

func TestOutputLevel_String(t *testing.T) {
	tests := []struct {
		level    OutputLevel
		expected string
	}{
		{OutputQuiet, "quiet"},
		{OutputNormal, "normal"},
		{OutputVerbose, "verbose"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("OutputLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewDefaultOutputConfig(t *testing.T) {
	config := NewDefaultOutputConfig()

	if config.Level != OutputNormal {
		t.Errorf("Expected OutputNormal, got %v", config.Level)
	}

	if !config.ShowColors {
		t.Error("Expected ShowColors to be true")
	}

	if !config.ShowProgress {
		t.Error("Expected ShowProgress to be true")
	}
}

func TestNewQuietOutputConfig(t *testing.T) {
	config := NewQuietOutputConfig()

	if config.Level != OutputQuiet {
		t.Errorf("Expected OutputQuiet, got %v", config.Level)
	}

	if config.ShowProgress {
		t.Error("Expected ShowProgress to be false in quiet mode")
	}

	if !config.SuppressServerStderr {
		t.Error("Expected SuppressServerStderr to be true in quiet mode")
	}
}

func TestNewVerboseOutputConfig(t *testing.T) {
	config := NewVerboseOutputConfig()

	if config.Level != OutputVerbose {
		t.Errorf("Expected OutputVerbose, got %v", config.Level)
	}

	if !config.ShowTimestamps {
		t.Error("Expected ShowTimestamps to be true in verbose mode")
	}
}

func TestOutputConfig_ShouldShow(t *testing.T) {
	tests := []struct {
		name     string
		config   OutputLevel
		check    OutputLevel
		expected bool
	}{
		{"quiet shows quiet", OutputQuiet, OutputQuiet, true},
		{"quiet hides normal", OutputQuiet, OutputNormal, false},
		{"quiet hides verbose", OutputQuiet, OutputVerbose, false},
		{"normal shows quiet", OutputNormal, OutputQuiet, true},
		{"normal shows normal", OutputNormal, OutputNormal, true},
		{"normal hides verbose", OutputNormal, OutputVerbose, false},
		{"verbose shows all", OutputVerbose, OutputVerbose, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OutputConfig{Level: tt.config}
			if got := config.ShouldShow(tt.check); got != tt.expected {
				t.Errorf("ShouldShow() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOutputConfig_ShouldShowConnectionMessages(t *testing.T) {
	tests := []struct {
		level    OutputLevel
		expected bool
	}{
		{OutputQuiet, false},
		{OutputNormal, false},
		{OutputVerbose, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			config := &OutputConfig{Level: tt.level}
			if got := config.ShouldShowConnectionMessages(); got != tt.expected {
				t.Errorf("ShouldShowConnectionMessages() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOutputConfig_Clone(t *testing.T) {
	original := &OutputConfig{
		Level:                OutputVerbose,
		ShowColors:           false,
		ShowProgress:         true,
		ShowTimestamps:       true,
		SuppressServerStderr: true,
	}

	clone := original.Clone()

	// Verify values match
	if clone.Level != original.Level {
		t.Error("Clone Level mismatch")
	}
	if clone.ShowColors != original.ShowColors {
		t.Error("Clone ShowColors mismatch")
	}
	if clone.ShowProgress != original.ShowProgress {
		t.Error("Clone ShowProgress mismatch")
	}
	if clone.ShowTimestamps != original.ShowTimestamps {
		t.Error("Clone ShowTimestamps mismatch")
	}
	if clone.SuppressServerStderr != original.SuppressServerStderr {
		t.Error("Clone SuppressServerStderr mismatch")
	}

	// Verify it's a different instance
	clone.Level = OutputQuiet
	if original.Level == OutputQuiet {
		t.Error("Clone modified original")
	}
}
