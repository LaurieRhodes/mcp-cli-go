package skills

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SkillImageMapping represents the skill-to-container-image mapping configuration
type SkillImageMapping struct {
	DefaultImage       string                 `yaml:"default_image"`
	Skills             map[string]string      `yaml:"skills"`
	SkillNetworkModes  map[string]string      `yaml:"skill_network_modes,omitempty"`
	ContainerConfig    ContainerConfig        `yaml:"container_config"`
	ImageInfo          map[string]interface{} `yaml:"image_info,omitempty"`
}

// ContainerConfig holds container runtime configuration
type ContainerConfig struct {
	MemoryLimit string `yaml:"memory_limit"`
	CPULimit    string `yaml:"cpu_limit"`
	Timeout     int    `yaml:"timeout"`
	Network     string `yaml:"network"`
	PidsLimit   int    `yaml:"pids_limit"`
}

// LoadSkillImageMapping loads the skill-to-image mapping from a YAML file
// If the file doesn't exist, returns a default mapping
func LoadSkillImageMapping(path string) (*SkillImageMapping, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default mapping if file doesn't exist
		return &SkillImageMapping{
			DefaultImage:      "python:3.11-slim",
			Skills:            make(map[string]string),
			SkillNetworkModes: make(map[string]string),
			ContainerConfig: ContainerConfig{
				MemoryLimit: "256m",
				CPULimit:    "0.5",
				Timeout:     60,
				Network:     "none",
				PidsLimit:   100,
			},
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
	if mapping.DefaultImage == "" {
		mapping.DefaultImage = "python:3.11-slim"
	}
	if mapping.Skills == nil {
		mapping.Skills = make(map[string]string)
	}
	if mapping.SkillNetworkModes == nil {
		mapping.SkillNetworkModes = make(map[string]string)
	}
	if mapping.ContainerConfig.MemoryLimit == "" {
		mapping.ContainerConfig.MemoryLimit = "256m"
	}
	if mapping.ContainerConfig.CPULimit == "" {
		mapping.ContainerConfig.CPULimit = "0.5"
	}
	if mapping.ContainerConfig.Timeout == 0 {
		mapping.ContainerConfig.Timeout = 60
	}
	if mapping.ContainerConfig.Network == "" {
		mapping.ContainerConfig.Network = "none"
	}
	if mapping.ContainerConfig.PidsLimit == 0 {
		mapping.ContainerConfig.PidsLimit = 100
	}

	return &mapping, nil
}

// GetImageForSkill returns the container image name for a given skill
// If no specific mapping exists, returns the default image
func (m *SkillImageMapping) GetImageForSkill(skillName string) string {
	if image, exists := m.Skills[skillName]; exists && image != "" {
		return image
	}
	return m.DefaultImage
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
// Returns skill-specific mode if defined, otherwise the default from container_config
func (m *SkillImageMapping) GetNetworkModeForSkill(skillName string) string {
	if mode, exists := m.SkillNetworkModes[skillName]; exists && mode != "" {
		return mode
	}
	return m.ContainerConfig.Network
}
