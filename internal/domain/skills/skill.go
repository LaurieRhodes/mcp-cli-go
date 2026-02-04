package skills

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Skill represents an Anthropic-compatible skill
type Skill struct {
	// Parsed from YAML frontmatter
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Language    string `yaml:"-" json:"language,omitempty"` // Required language (bash, python, etc.)
	License     string `yaml:"license,omitempty" json:"license,omitempty"`

	// Skill metadata (not from YAML)
	DirectoryPath  string   `yaml:"-" json:"directory_path"`
	SkillMDPath    string   `yaml:"-" json:"skill_md_path"`
	HasReferences  bool     `yaml:"-" json:"has_references"`
	HasScripts     bool     `yaml:"-" json:"has_scripts"`
	HasAssets      bool     `yaml:"-" json:"has_assets"`
	ReferenceFiles []string `yaml:"-" json:"reference_files,omitempty"`
	ScriptFiles    []string `yaml:"-" json:"script_files,omitempty"`
	Scripts        []string `yaml:"-" json:"scripts,omitempty"`     // Executable scripts only
	ScriptsDir     string   `yaml:"-" json:"scripts_dir,omitempty"` // Full path to scripts directory
	AssetFiles     []string `yaml:"-" json:"asset_files,omitempty"`

	// Skill content (loaded on demand)
	MainContent string            `yaml:"-" json:"-"` // SKILL.md body
	References  map[string]string `yaml:"-" json:"-"` // reference filename -> content

	// For active mode (optional)
	WorkflowPath string `yaml:"-" json:"workflow_path,omitempty"`
	HasWorkflow  bool   `yaml:"-" json:"has_workflow"`
}

// SkillFrontmatter represents the YAML frontmatter in SKILL.md
type SkillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Language    string `yaml:"-" json:"language,omitempty"` // Required language (bash, python, etc.)
	License     string `yaml:"license,omitempty"`
}

// Validate validates the skill
func (s *Skill) Validate() error {
	// Validate name
	if err := ValidateSkillName(s.Name); err != nil {
		return fmt.Errorf("invalid skill name: %w", err)
	}

	// Validate description
	if s.Description == "" {
		return fmt.Errorf("skill description is required")
	}

	if len(s.Description) > 1024 {
		return fmt.Errorf("skill description too long (max 1024 characters, got %d)", len(s.Description))
	}

	// Validate paths
	if s.DirectoryPath == "" {
		return fmt.Errorf("skill directory path is required")
	}

	if s.SkillMDPath == "" {
		return fmt.Errorf("SKILL.md path is required")
	}

	return nil
}

// GetToolDescription generates an MCP tool description from this skill
// Optimized for small LLMs with concrete, action-oriented language
func (s *Skill) GetToolDescription() string {
	return fmt.Sprintf("[SKILL] %s\n\n"+
		"CALL THIS FIRST to see:\n"+
		"• Available scripts and how to use them\n"+
		"• Example commands with correct file paths\n"+
		"• Required parameters and output formats\n\n"+
		"After reading this, use 'execute_skill_code' tool with skill_name='%s' to run the commands.",
		s.Description, s.Name)
}

// GetMCPToolName returns the MCP tool name for this skill
// Follows MCP naming conventions (lowercase with underscores)
func (s *Skill) GetMCPToolName() string {
	// Convert skill-name to skill_name
	return strings.ReplaceAll(s.Name, "-", "_")
}

// GetMCPInputSchema returns the JSON Schema for this skill's MCP tool
func (s *Skill) GetMCPInputSchema() map[string]interface{} {
	// Basic schema for skill loading
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "Loading mode: 'passive' (load as context) or 'active' (execute workflow)",
				"enum":        []string{"passive", "active"},
				"default":     "passive",
			},
			"include_references": map[string]interface{}{
				"type":        "boolean",
				"description": "Include reference files in initial load (passive mode only)",
				"default":     false,
			},
			"reference_files": map[string]interface{}{
				"type":        "array",
				"description": "Specific reference files to load (passive mode only)",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"input_data": map[string]interface{}{
				"type":        "string",
				"description": "Input data for workflow execution (active mode only)",
			},
		},
	}
}

// ValidateSkillName validates a skill name
// Per Anthropic spec: lowercase letters, numbers, and hyphens only, max 64 chars
func ValidateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 64 {
		return fmt.Errorf("name too long (max 64 characters, got %d)", len(name))
	}

	// Must match pattern: lowercase letters, numbers, and hyphens
	validPattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("name must contain only lowercase letters, numbers, and hyphens (got: %s)", name)
	}

	// Cannot start or end with hyphen
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("name cannot start or end with hyphen")
	}

	return nil
}

// GetSkillNameFromDirectory extracts the skill name from a directory path
func GetSkillNameFromDirectory(dirPath string) string {
	return filepath.Base(dirPath)
}

// SkillLoadMode represents how a skill should be loaded
type SkillLoadMode string

const (
	// SkillLoadModePassive loads skill content as context
	SkillLoadModePassive SkillLoadMode = "passive"

	// SkillLoadModeActive executes skill workflow
	SkillLoadModeActive SkillLoadMode = "active"
)

// SkillLoadRequest represents a request to load a skill
type SkillLoadRequest struct {
	SkillName         string        `json:"skill_name"`
	Mode              SkillLoadMode `json:"mode"`
	IncludeReferences bool          `json:"include_references"`
	ReferenceFiles    []string      `json:"reference_files,omitempty"`
	InputData         string        `json:"input_data,omitempty"` // For active mode
}

// SkillLoadResult represents the result of loading a skill
type SkillLoadResult struct {
	SkillName   string        `json:"skill_name"`
	Mode        SkillLoadMode `json:"mode"`
	Content     string        `json:"content"` // For passive mode
	Result      string        `json:"result"`  // For active mode
	Error       error         `json:"error,omitempty"`
	LoadedFiles []string      `json:"loaded_files,omitempty"`
}

// HelperScriptRequest represents a request to run a helper script
type HelperScriptRequest struct {
	SkillName  string   // Skill containing the script
	ScriptName string   // Script filename (e.g., "process_chunk.py")
	Args       []string // Command-line arguments
}
