package sandbox

import (
	"path/filepath"
	"testing"
)

// testMapper is a test implementation of the ImageMapper interface
type testMapper struct {
	skills       map[string]string
	defaultImage string
}

func (tm testMapper) GetImageForSkill(skillName string) string {
	if img, ok := tm.skills[skillName]; ok {
		return img
	}
	return tm.defaultImage
}

func TestGetImageForSkill(t *testing.T) {
	
	mapping := testMapper{
		skills: map[string]string{
			"docx": "mcp-skills-docx",
			"pptx": "mcp-skills-pptx",
			"xlsx": "mcp-skills-xlsx",
			"pdf":  "mcp-skills-pdf",
		},
		defaultImage: "mcp-skills-office",
	}
	
	config := ExecutorConfig{
		PythonImage:  "python:3.11-slim",
		ImageMapping: mapping,
	}
	
	tests := []struct {
		name          string
		skillPath     string
		expectedImage string
	}{
		{
			name:          "DOCX skill uses docx image",
			skillPath:     "/path/to/skills/docx",
			expectedImage: "mcp-skills-docx",
		},
		{
			name:          "PPTX skill uses pptx image",
			skillPath:     "/config/skills/pptx",
			expectedImage: "mcp-skills-pptx",
		},
		{
			name:          "XLSX skill uses xlsx image",
			skillPath:     "skills/xlsx",
			expectedImage: "mcp-skills-xlsx",
		},
		{
			name:          "PDF skill uses pdf image",
			skillPath:     "/skills/pdf",
			expectedImage: "mcp-skills-pdf",
		},
		{
			name:          "Unknown skill uses default",
			skillPath:     "/skills/unknown-skill",
			expectedImage: "mcp-skills-office",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image := config.GetImageForSkill(tt.skillPath)
			if image != tt.expectedImage {
				t.Errorf("GetImageForSkill(%s) = %s, want %s", 
					filepath.Base(tt.skillPath), image, tt.expectedImage)
			}
		})
	}
}

func TestExecutorConfigWithoutMapping(t *testing.T) {
	config := ExecutorConfig{
		PythonImage:  "python:3.11-slim",
		ImageMapping: nil,
	}
	
	// Should always return default when no mapping
	tests := []string{
		"/skills/docx",
		"/skills/pptx",
		"/skills/anything",
	}
	
	for _, skillPath := range tests {
		image := config.GetImageForSkill(skillPath)
		if image != "python:3.11-slim" {
			t.Errorf("Without mapping, expected default image python:3.11-slim, got %s", image)
		}
	}
}

