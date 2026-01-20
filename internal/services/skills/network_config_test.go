package skills

import (
	"os"
	"testing"
)

func TestNetworkModeConfiguration(t *testing.T) {
	// Create temporary test config (V2 format)
	testConfig := `defaults:
  image: python:3.11-alpine
  network_mode: none

skills:
  docx:
    image: mcp-skills-docx
  
  test-secure:
    image: mcp-skills-docx
    network_mode: none
  
  test-network:
    image: mcp-skills-docx
    network_mode: bridge
`
	
	tmpFile, err := os.CreateTemp("", "test-skill-images-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tmpFile.Close()
	
	// Load configuration
	mapping, err := LoadSkillImageMapping(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Test cases
	tests := []struct{
		name     string
		skill    string
		expected string
	}{
		{"Default network mode", "docx", "none"},
		{"Explicitly set to none", "test-secure", "none"},
		{"Explicitly set to bridge", "test-network", "bridge"},
		{"Unknown skill uses default", "unknown", "none"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapping.GetNetworkModeForSkill(tt.skill)
			if result != tt.expected {
				t.Errorf("GetNetworkModeForSkill(%s) = %s; want %s", 
					tt.skill, result, tt.expected)
			}
		})
	}
}

func TestSkillImageMapping(t *testing.T) {
	// Create temporary test config (V2 format)
	testConfig := `defaults:
  image: python:3.11-alpine

skills:
  docx:
    image: mcp-skills-docx
  
  custom:
    image: my-custom-image
`
	
	tmpFile, err := os.CreateTemp("", "test-skill-images-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tmpFile.Close()
	
	// Load configuration
	mapping, err := LoadSkillImageMapping(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Test image mapping - use skill names directly
	tests := []struct{
		name     string
		skill    string
		expected string
	}{
		{"Mapped skill", "docx", "mcp-skills-docx"},
		{"Custom skill", "custom", "my-custom-image"},
		{"Unmapped skill uses default", "unknown", "python:3.11-alpine"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapping.GetImageForSkill(tt.skill)
			if result != tt.expected {
				t.Errorf("GetImageForSkill(%s) = %s; want %s", 
					tt.skill, result, tt.expected)
			}
		})
	}
}
