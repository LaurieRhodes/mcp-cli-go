package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkillImageMapping(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Test 1: Valid V2 mapping file
	t.Run("ValidMapping", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "valid-mapping.yaml")
		content := `defaults:
  image: mcp-skills-office
  network_mode: none
  memory: 512MB
  cpu: "1.0"
  timeout: 120s

skills:
  docx:
    image: mcp-skills-docx
  pptx:
    image: mcp-skills-pptx
  pdf:
    image: mcp-skills-pdf
`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		mapping, err := LoadSkillImageMapping(mappingFile)
		if err != nil {
			t.Fatalf("LoadSkillImageMapping failed: %v", err)
		}

		// Check defaults
		if mapping.Defaults.Image != "mcp-skills-office" {
			t.Errorf("Expected default image 'mcp-skills-office', got '%s'", mapping.Defaults.Image)
		}
		if mapping.Defaults.NetworkMode != "none" {
			t.Errorf("Expected network_mode 'none', got '%s'", mapping.Defaults.NetworkMode)
		}
		if mapping.Defaults.Memory != "512MB" {
			t.Errorf("Expected memory '512MB', got '%s'", mapping.Defaults.Memory)
		}

		// Check skill mappings
		if mapping.Skills["docx"].Image != "mcp-skills-docx" {
			t.Errorf("Expected docx -> mcp-skills-docx, got '%s'", mapping.Skills["docx"].Image)
		}
		if mapping.Skills["pptx"].Image != "mcp-skills-pptx" {
			t.Errorf("Expected pptx -> mcp-skills-pptx, got '%s'", mapping.Skills["pptx"].Image)
		}
		if mapping.Skills["pdf"].Image != "mcp-skills-pdf" {
			t.Errorf("Expected pdf -> mcp-skills-pdf, got '%s'", mapping.Skills["pdf"].Image)
		}
	})

	// Test 2: File doesn't exist
	t.Run("FileNotExists", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "does-not-exist.yaml")
		mapping, err := LoadSkillImageMapping(nonExistentFile)

		// Should return default mapping, not error
		if err != nil {
			t.Fatalf("Expected default mapping, got error: %v", err)
		}

		if mapping.Defaults.Image != "python:3.11-alpine" {
			t.Errorf("Expected default image 'python:3.11-alpine', got '%s'", mapping.Defaults.Image)
		}
	})

	// Test 3: Minimal mapping
	t.Run("MinimalMapping", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "minimal-mapping.yaml")
		content := `defaults:
  image: python:3.11-alpine

skills: {}
`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		mapping, err := LoadSkillImageMapping(mappingFile)
		if err != nil {
			t.Fatalf("LoadSkillImageMapping failed: %v", err)
		}

		// Should have default values filled in
		if mapping.Defaults.NetworkMode != "none" {
			t.Errorf("Expected default network_mode 'none', got '%s'", mapping.Defaults.NetworkMode)
		}
		if mapping.Defaults.Memory != "256MB" {
			t.Errorf("Expected default memory '256MB', got '%s'", mapping.Defaults.Memory)
		}
	})

	// Test 4: Invalid YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "invalid.yaml")
		content := `this is not: valid: yaml:: syntax`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := LoadSkillImageMapping(mappingFile)
		if err == nil {
			t.Fatal("Expected error for invalid YAML, got nil")
		}
	})
}

func TestGetImageForSkill(t *testing.T) {
	mapping := &SkillImageMapping{
		Defaults: SkillDefaults{
			Image: "python:3.11-alpine",
		},
		Skills: map[string]*SkillSpec{
			"docx": {Image: "mcp-skills-docx"},
			"pptx": {Image: "mcp-skills-pptx"},
			"xlsx": {Image: "mcp-skills-xlsx"},
			"pdf":  {Image: "mcp-skills-pdf"},
		},
	}

	tests := []struct {
		skillName string
		expected  string
	}{
		{"pptx", "mcp-skills-pptx"},
		{"docx", "mcp-skills-docx"},
		{"xlsx", "mcp-skills-xlsx"},
		{"pdf", "mcp-skills-pdf"},
		{"unknown-skill", "python:3.11-alpine"},
		{"", "python:3.11-alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.skillName, func(t *testing.T) {
			result := mapping.GetImageForSkill(tt.skillName)
			if result != tt.expected {
				t.Errorf("GetImageForSkill(%s) = %s; want %s", tt.skillName, result, tt.expected)
			}
		})
	}
}

func TestLoadSkillImageMappingFromSkillsDir(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	skillsDir := filepath.Join(configDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create skill-images.yaml
	mappingFile := filepath.Join(skillsDir, "skill-images.yaml")
	content := `defaults:
  image: python:3.11-alpine

skills:
  docx:
    image: mcp-skills-docx
`
	if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a config file in the config directory
	configFile := filepath.Join(configDir, "settings.yaml")
	if err := os.WriteFile(configFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load mapping from skills directory
	mapping, err := LoadSkillImageMappingFromSkillsDir(configFile)
	if err != nil {
		t.Fatalf("LoadSkillImageMappingFromSkillsDir failed: %v", err)
	}

	// Verify loaded mapping
	if mapping.Skills["docx"].Image != "mcp-skills-docx" {
		t.Errorf("Expected docx -> mcp-skills-docx, got '%s'", mapping.Skills["docx"].Image)
	}
}
