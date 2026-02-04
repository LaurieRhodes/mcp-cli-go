package skills

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SkillDefaults contains default values inherited by all skills
type SkillDefaults struct {
	Language    string `yaml:"language,omitempty"`
	Image       string `yaml:"image"`
	NetworkMode string `yaml:"network_mode"`
	Memory      string `yaml:"memory"`
	CPU         string `yaml:"cpu"`
	Timeout     string `yaml:"timeout"`
	OutputsDir  string `yaml:"outputs_dir"`
}

// SkillSpec contains the complete configuration for a skill
type SkillSpec struct {
	Image                string   `yaml:"image"`
	Language             string   `yaml:"language,omitempty"`
	Languages            []string `yaml:"languages,omitempty"`
	Description          string   `yaml:"description,omitempty"`
	NetworkMode          string   `yaml:"network_mode,omitempty"`
	Dockerfile           string   `yaml:"dockerfile,omitempty"`
	Memory               string   `yaml:"memory,omitempty"`
	CPU                  string   `yaml:"cpu,omitempty"`
	Timeout              string   `yaml:"timeout,omitempty"`
	Mounts               []string `yaml:"mounts,omitempty"`
	Environment          []string `yaml:"environment,omitempty"`
	NetworkJustification string   `yaml:"network_justification,omitempty"`
}

// SkillImageMapping maps skill names to their configurations (V2 format)
type SkillImageMapping struct {
	Defaults SkillDefaults         `yaml:"defaults"`
	Skills   map[string]*SkillSpec `yaml:"skills"`
}

// LoadSkillImageMapping loads the skill-to-image mapping from a YAML file
// If the file doesn't exist, returns a default mapping
func LoadSkillImageMapping(path string) (*SkillImageMapping, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default mapping if file doesn't exist
		return &SkillImageMapping{
			Defaults: SkillDefaults{
				Image:       "python:3.11-alpine",
				NetworkMode: "none",
				Memory:      "256MB",
				CPU:         "0.5",
				Timeout:     "60s",
				OutputsDir:  "/tmp/mcp-outputs",
			},
			Skills: make(map[string]*SkillSpec),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var mapping SkillImageMapping
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if mapping.Defaults.Image == "" {
		mapping.Defaults.Image = "python:3.11-alpine"
	}
	if mapping.Defaults.Language == "" {
		mapping.Defaults.Language = "python"
	}
	if mapping.Defaults.NetworkMode == "" {
		mapping.Defaults.NetworkMode = "none"
	}
	if mapping.Defaults.Memory == "" {
		mapping.Defaults.Memory = "256MB"
	}
	if mapping.Defaults.CPU == "" {
		mapping.Defaults.CPU = "0.5"
	}
	if mapping.Defaults.Timeout == "" {
		mapping.Defaults.Timeout = "60s"
	}
	if mapping.Defaults.OutputsDir == "" {
		mapping.Defaults.OutputsDir = "/tmp/mcp-outputs"
	}
	if mapping.Skills == nil {
		mapping.Skills = make(map[string]*SkillSpec)
	}

	return &mapping, nil
}

// GetImageForSkill returns the container image name for a given skill
// If no specific mapping exists, returns the default image
func (m *SkillImageMapping) GetImageForSkill(skillName string) string {
	if spec, exists := m.Skills[skillName]; exists && spec != nil && spec.Image != "" {
		return spec.Image
	}
	return m.Defaults.Image
}

// LoadSkillImageMappingFromSkillsDir loads the mapping from the standard skills directory
// Looks for config/skills/skill-images.yaml relative to a config file
func LoadSkillImageMappingFromSkillsDir(configFilePath string) (*SkillImageMapping, error) {
	// Get directory of config file
	configDir := filepath.Dir(configFilePath)

	// Look for skill-images.yaml in config/skills/
	mappingPath := filepath.Join(configDir, "skills", "skill-images.yaml")

	return LoadSkillImageMapping(mappingPath)
}

// LoadSkillImageMappingDefault loads the mapping from the default location
// Tries: ./config/skills/skill-images.yaml
func LoadSkillImageMappingDefault() (*SkillImageMapping, error) {
	return LoadSkillImageMapping("config/skills/skill-images.yaml")
}

// GetNetworkModeForSkill returns the network mode for a given skill
// Returns skill-specific mode if defined, otherwise the default
func (m *SkillImageMapping) GetNetworkModeForSkill(skillName string) string {
	if spec, exists := m.Skills[skillName]; exists && spec != nil && spec.NetworkMode != "" {
		return spec.NetworkMode
	}
	return m.Defaults.NetworkMode
}
