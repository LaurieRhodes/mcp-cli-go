package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkillImageMapping(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	
	// Test 1: Valid mapping file
	t.Run("ValidMapping", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "valid-mapping.yaml")
		content := `default_image: mcp-skills-office

skills:
  docx: mcp-skills-docx
  pptx: mcp-skills-pptx
  pdf: mcp-skills-pdf

container_config:
  memory_limit: "512m"
  cpu_limit: "1.0"
  timeout: 120
  network: "none"
  pids_limit: 200
`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		mapping, err := LoadSkillImageMapping(mappingFile)
		if err != nil {
			t.Fatalf("LoadSkillImageMapping failed: %v", err)
		}

		// Check default image
		if mapping.DefaultImage != "mcp-skills-office" {
			t.Errorf("Expected default_image 'mcp-skills-office', got '%s'", mapping.DefaultImage)
		}

		// Check skill mappings
		if mapping.Skills["docx"] != "mcp-skills-docx" {
			t.Errorf("Expected docx -> mcp-skills-docx, got '%s'", mapping.Skills["docx"])
		}
		if mapping.Skills["pptx"] != "mcp-skills-pptx" {
			t.Errorf("Expected pptx -> mcp-skills-pptx, got '%s'", mapping.Skills["pptx"])
		}
		if mapping.Skills["pdf"] != "mcp-skills-pdf" {
			t.Errorf("Expected pdf -> mcp-skills-pdf, got '%s'", mapping.Skills["pdf"])
		}

		// Check container config
		if mapping.ContainerConfig.MemoryLimit != "512m" {
			t.Errorf("Expected memory_limit '512m', got '%s'", mapping.ContainerConfig.MemoryLimit)
		}
		if mapping.ContainerConfig.CPULimit != "1.0" {
			t.Errorf("Expected cpu_limit '1.0', got '%s'", mapping.ContainerConfig.CPULimit)
		}
		if mapping.ContainerConfig.Timeout != 120 {
			t.Errorf("Expected timeout 120, got %d", mapping.ContainerConfig.Timeout)
		}
	})

	// Test 2: File doesn't exist (should return defaults)
	t.Run("FileNotExists", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "does-not-exist.yaml")
		
		mapping, err := LoadSkillImageMapping(nonExistentFile)
		if err != nil {
			t.Fatalf("LoadSkillImageMapping should not fail for missing file: %v", err)
		}

		// Check defaults
		if mapping.DefaultImage != "python:3.11-slim" {
			t.Errorf("Expected default 'python:3.11-slim', got '%s'", mapping.DefaultImage)
		}
		if mapping.ContainerConfig.MemoryLimit != "256m" {
			t.Errorf("Expected default memory '256m', got '%s'", mapping.ContainerConfig.MemoryLimit)
		}
		if mapping.ContainerConfig.Timeout != 60 {
			t.Errorf("Expected default timeout 60, got %d", mapping.ContainerConfig.Timeout)
		}
	})

	// Test 3: Minimal file (defaults should be filled in)
	t.Run("MinimalMapping", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "minimal-mapping.yaml")
		content := `default_image: test-image

skills:
  test-skill: test-image-specific
`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		mapping, err := LoadSkillImageMapping(mappingFile)
		if err != nil {
			t.Fatalf("LoadSkillImageMapping failed: %v", err)
		}

		// Check that defaults were filled in
		if mapping.ContainerConfig.MemoryLimit != "256m" {
			t.Errorf("Expected default memory '256m', got '%s'", mapping.ContainerConfig.MemoryLimit)
		}
		if mapping.ContainerConfig.Network != "none" {
			t.Errorf("Expected default network 'none', got '%s'", mapping.ContainerConfig.Network)
		}
		if mapping.ContainerConfig.PidsLimit != 100 {
			t.Errorf("Expected default pids_limit 100, got %d", mapping.ContainerConfig.PidsLimit)
		}
	})

	// Test 4: Invalid YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		mappingFile := filepath.Join(tmpDir, "invalid-mapping.yaml")
		content := `this is not valid yaml: [[[`
		if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := LoadSkillImageMapping(mappingFile)
		if err == nil {
			t.Error("LoadSkillImageMapping should fail for invalid YAML")
		}
	})
}

func TestGetImageForSkill(t *testing.T) {
	mapping := &SkillImageMapping{
		DefaultImage: "default-image",
		Skills: map[string]string{
			"pptx":      "mcp-skills-pptx",
			"docx":      "mcp-skills-docx",
			"xlsx":      "mcp-skills-xlsx",
			"pdf":       "mcp-skills-pdf",
		},
	}

	tests := []struct {
		skillName     string
		expectedImage string
	}{
		{"pptx", "mcp-skills-pptx"},
		{"docx", "mcp-skills-docx"},
		{"xlsx", "mcp-skills-xlsx"},
		{"pdf", "mcp-skills-pdf"},
		{"unknown-skill", "default-image"},
		{"", "default-image"},
	}

	for _, tt := range tests {
		t.Run(tt.skillName, func(t *testing.T) {
			image := mapping.GetImageForSkill(tt.skillName)
			if image != tt.expectedImage {
				t.Errorf("GetImageForSkill(%s) = %s, want %s", tt.skillName, image, tt.expectedImage)
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
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create mapping file
	mappingFile := filepath.Join(skillsDir, "skill-images.yaml")
	content := `default_image: test-default
skills:
  test: test-image
`
	if err := os.WriteFile(mappingFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create mapping file: %v", err)
	}

	// Create a fake config file
	configFile := filepath.Join(configDir, "test-config.yaml")
	if err := os.WriteFile(configFile, []byte("test: config"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading from skills dir
	mapping, err := LoadSkillImageMappingFromSkillsDir(configFile)
	if err != nil {
		t.Fatalf("LoadSkillImageMappingFromSkillsDir failed: %v", err)
	}

	if mapping.DefaultImage != "test-default" {
		t.Errorf("Expected default_image 'test-default', got '%s'", mapping.DefaultImage)
	}
	if mapping.Skills["test"] != "test-image" {
		t.Errorf("Expected test -> test-image, got '%s'", mapping.Skills["test"])
	}
}
