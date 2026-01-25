package skills

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
)

// SeparateSkillsFromServers splits servers list into external servers and skills flag
// Returns: externalServers (non-skills servers), needsSkills (true if "skills" was in list)
func SeparateSkillsFromServers(servers []string) (externalServers []string, needsSkills bool) {
	for _, server := range servers {
		if server == "skills" {
			needsSkills = true
			logging.Info("Detected 'skills' in server list - will use built-in skills service")
		} else {
			externalServers = append(externalServers, server)
		}
	}
	return
}

// InitializeBuiltinSkills initializes the built-in skills service
// This uses the skills directory from config, with proper path resolution
func InitializeBuiltinSkills(configFile string, appConfig *config.ApplicationConfig) (*skillsvc.Service, error) {
	logging.Info("Initializing built-in skills service")
	
	skillService := skillsvc.NewService()
	skillService.SetConfig(appConfig)
	
	// Get skills directory from config (defaults to "config/skills")
	var skillsDir string
	if appConfig != nil && appConfig.Skills != nil {
		skillsDir = appConfig.Skills.GetSkillsDirectory()
	} else {
		skillsDir = "config/skills" // Fallback to convention
	}
	
	// Resolve relative paths relative to the config file's directory
	if !filepath.IsAbs(skillsDir) {
		// Get absolute path of config file first
		absConfigFile, err := filepath.Abs(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path of config file: %w", err)
		}
		
		configDir := filepath.Dir(absConfigFile)
		
		// If skills dir starts with "config/", it's relative to the project root
		// Otherwise it's relative to the config file's directory
		if strings.HasPrefix(skillsDir, "config/") {
			// Find project root by going up from config file directory
			// If config file is at root (e.g., /path/to/project/config.yaml),
			// configDir is already the project root
			// If config file is in a subdir (e.g., /path/to/project/config/config.yaml),
			// we need to go up one level
			projectRoot := configDir
			if filepath.Base(configDir) == "config" {
				projectRoot = filepath.Dir(configDir)
			}
			skillsDir = filepath.Join(projectRoot, skillsDir)
		} else {
			// Skills dir is relative to config file's directory
			skillsDir = filepath.Join(configDir, skillsDir)
		}
	}
	
	logging.Debug("Skills directory: %s", skillsDir)
	
	// Initialize with auto execution mode
	if err := skillService.Initialize(skillsDir, skills.ExecutionModeAuto); err != nil {
		return nil, fmt.Errorf("failed to initialize built-in skills: %w", err)
	}
	
	discoveredSkills := skillService.ListSkills()
	logging.Info("Initialized %d built-in skills: %v", len(discoveredSkills), discoveredSkills)
	
	return skillService, nil
}
