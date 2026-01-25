package skills

import (
	"context"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
)

// SkillScanner defines the interface for scanning and discovering skills
type SkillScanner interface {
	// ScanSkillsDirectory scans a directory for Anthropic-compatible skills
	// Returns a map of skill name -> Skill
	ScanSkillsDirectory(skillsDir string) (map[string]*Skill, error)
	
	// LoadSkill loads a single skill from a directory
	LoadSkill(skillDir string) (*Skill, error)
	
	// ValidateSkill validates a skill structure and content
	ValidateSkill(skill *Skill) error
	
	// ListSkills returns all discovered skill names
	ListSkills() []string
	
	// GetSkill retrieves a skill by name
	GetSkill(name string) (*Skill, bool)
}

// SkillLoader defines the interface for loading skill content
type SkillLoader interface {
	// LoadMainContent loads the main SKILL.md content (body only, not frontmatter)
	LoadMainContent(skill *Skill) (string, error)
	
	// LoadReference loads a specific reference file
	LoadReference(skill *Skill, referenceName string) (string, error)
	
	// LoadAllReferences loads all reference files for a skill
	LoadAllReferences(skill *Skill) (map[string]string, error)
	
	// LoadAsPassive loads skill in passive mode (as context)
	LoadAsPassive(skill *Skill, request *SkillLoadRequest) (*SkillLoadResult, error)
}

// SkillExecutor defines the interface for executing skill workflows
type SkillExecutor interface {
	// ExecuteWorkflow executes a skill's workflow.yaml
	ExecuteWorkflow(skill *Skill, inputData string) (*SkillLoadResult, error)
	
	// ExecuteScript executes a specific script from the skill
	ExecuteScript(skill *Skill, scriptName string, args []string) (string, error)
	
	// ExecuteCode executes arbitrary code with access to skill's helper libraries
	// This is the core capability that matches Anthropic's design
	ExecuteCode(request *CodeExecutionRequest) (*ExecutionResult, error)
	
	// LoadAsActive loads skill in active mode (executes workflow)
	LoadAsActive(skill *Skill, request *SkillLoadRequest) (*SkillLoadResult, error)
}

// SkillService combines all skill-related operations
type SkillService interface {
	SkillScanner
	SkillLoader
	SkillExecutor
	
	// Initialize initializes the skill service (scans skills directory)
	Initialize(skillsDir string, executionMode ExecutionMode) error
	
	// LoadSkillByRequest loads a skill according to the request
	LoadSkillByRequest(request *SkillLoadRequest) (*SkillLoadResult, error)
	
	// GenerateRunAsTools generates MCP tool definitions for all skills
	
	// GetSkillLanguage returns the configured language for a skill
	GetSkillLanguage(skillName string) string
	
	// GetSkillLanguages returns all supported languages for a skill
	GetSkillLanguages(skillName string) []string
	// Returns tools suitable for inclusion in a RunAsConfig
	GenerateRunAsTools() ([]map[string]interface{}, error)
	
	// ListTools returns all available skill tools as domain.Tool
	ListTools() ([]domain.Tool, error)
	
	// ExecuteTool executes a skill tool by name with arguments
	ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error)
}
