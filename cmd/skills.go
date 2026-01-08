package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	infraConfig "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
)

// executeListSkills lists all available skills
func executeListSkills() error {
	// Determine skills directory
	var skillsDir string
	
	// Try to get from config first
	configService := infraConfig.NewService()
	appConfig, _, err := configService.LoadConfigOrCreateExample(configFile)
	if err == nil && appConfig != nil && appConfig.Skills != nil {
		// Check if there's a configured skills directory in the future
		// For now, use default
	}
	
	// Default skills directory relative to config file location
	if skillsDir == "" {
		// Get absolute path of config file
		absConfigPath, err := filepath.Abs(configFile)
		if err == nil {
			configFile = absConfigPath
		}
		
		// Config directory (where config.yaml lives)
		configBaseDir := filepath.Dir(configFile)
		
		// Skills are in config/skills subdirectory
		skillsDir = filepath.Join(configBaseDir, "config", "skills")
	}
	
	logging.Debug("Skills directory: %s", skillsDir)
	
	// Initialize skills service
	skillService := skillsvc.NewService()
	if appConfig != nil {
		skillService.SetConfig(appConfig)
	}
	
	// Try to initialize with auto mode to detect Docker/Podman availability
	// This won't show warnings but will tell us if executor is available
	if err := skillService.Initialize(skillsDir, skills.ExecutionModePassive); err != nil {
		return fmt.Errorf("failed to initialize skills: %w", err)
	}
	
	// Get skill names
	skillNames := skillService.ListSkills()
	
	if len(skillNames) == 0 {
		fmt.Println("No skills found.")
		fmt.Println("\nSkills directory: " + skillsDir)
		fmt.Println("\nTo add skills:")
		fmt.Println("  1. Create a subdirectory in config/skills/")
		fmt.Println("  2. Add a SKILL.md file with YAML frontmatter")
		fmt.Println("  3. See config/skills/README.md for documentation")
		return nil
	}
	
	// Categorize skills
	activeSkills := make([]*skills.Skill, 0)
	passiveSkills := make([]*skills.Skill, 0)
	
	for _, name := range skillNames {
		if skill, exists := skillService.GetSkill(name); exists {
			if skill.HasScripts {
				activeSkills = append(activeSkills, skill)
			} else {
				passiveSkills = append(passiveSkills, skill)
			}
		}
	}
	
	// Sort by name
	sort.Slice(activeSkills, func(i, j int) bool {
		return activeSkills[i].Name < activeSkills[j].Name
	})
	sort.Slice(passiveSkills, func(i, j int) bool {
		return passiveSkills[i].Name < passiveSkills[j].Name
	})
	
	// Check Docker/Podman availability for active skills
	dockerAvailable := checkDockerAvailability()
	
	// Define color styles
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)
	gray := color.New(color.FgHiBlack)
	
	// Display header
	fmt.Println()
	cyan.Println("═══════════════════════════════════════════════════════════════")
	cyan.Println("                     ANTHROPIC SKILLS")
	cyan.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("\nDirectory: %s\n", gray.Sprint(skillsDir))
	fmt.Println()
	
	// Display Active Skills
	if len(activeSkills) > 0 {
		cyan.Println("┌─────────────────────────────────────────────────────────────┐")
		cyan.Printf("│ ACTIVE SKILLS (%d)                                           │\n", len(activeSkills))
		cyan.Println("└─────────────────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("  These skills have executable scripts for dynamic workflows.")
		
		if !dockerAvailable {
			fmt.Println()
			yellow.Println("  ⚠️  WARNING: Docker/Podman not available")
			yellow.Println("     These skills will run in PASSIVE mode (documentation only)")
			fmt.Println()
			fmt.Println("  To enable script execution:")
			fmt.Println("    • Install Docker: https://docs.docker.com/get-docker/")
			fmt.Println("    • OR Podman: https://podman.io/getting-started/installation")
		}
		
		fmt.Println()
		
		for _, skill := range activeSkills {
			var icon string
			var status string
			var statusColor *color.Color
			
			if dockerAvailable {
				icon = green.Sprint("✓")
				status = "ready"
				statusColor = green
			} else {
				icon = red.Sprint("✗")
				status = "unavailable"
				statusColor = red
			}
			
			fmt.Printf("  %s %s (%s)\n", icon, bold.Sprint(skill.Name), statusColor.Sprint(status))
			
			// Description (wrapped if needed)
			desc := skill.Description
			if len(desc) > 70 {
				desc = wrapText(desc, 70, "     ")
			}
			fmt.Printf("     %s\n", desc)
			
			if verbose {
				fmt.Printf("     %s %s\n", gray.Sprint("Directory:"), skill.DirectoryPath)
				fmt.Printf("     %s %d\n", gray.Sprint("Scripts:"), len(skill.ScriptFiles))
				if skill.License != "" {
					fmt.Printf("     %s %s\n", gray.Sprint("License:"), skill.License)
				}
			}
			fmt.Println()
		}
	}
	
	// Display Passive Skills
	if len(passiveSkills) > 0 {
		cyan.Println("┌─────────────────────────────────────────────────────────────┐")
		cyan.Printf("│ PASSIVE SKILLS (%d)                                          │\n", len(passiveSkills))
		cyan.Println("└─────────────────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("  These skills provide documentation and guidance.")
		fmt.Println()
		
		for _, skill := range passiveSkills {
			fmt.Printf("  %s %s\n", green.Sprint("✓"), bold.Sprint(skill.Name))
			
			// Description (wrapped if needed)
			desc := skill.Description
			if len(desc) > 70 {
				desc = wrapText(desc, 70, "     ")
			}
			fmt.Printf("     %s\n", desc)
			
			if verbose {
				fmt.Printf("     %s %s\n", gray.Sprint("Directory:"), skill.DirectoryPath)
				if skill.HasReferences {
					fmt.Printf("     %s %d\n", gray.Sprint("References:"), len(skill.ReferenceFiles))
				}
				if skill.License != "" {
					fmt.Printf("     %s %s\n", gray.Sprint("License:"), skill.License)
				}
			}
			fmt.Println()
		}
	}
	
	// Summary
	cyan.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("SUMMARY: %s total | %s active | %s passive\n", 
		bold.Sprint(len(skillNames)), 
		bold.Sprint(len(activeSkills)), 
		bold.Sprint(len(passiveSkills)))
	
	if len(activeSkills) > 0 {
		if dockerAvailable {
			fmt.Printf("Status: %s\n", green.Sprint("✓ Script execution enabled"))
		} else {
			fmt.Printf("Status: %s\n", yellow.Sprint("⚠  Script execution unavailable (Docker/Podman not found)"))
		}
	}
	cyan.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	
	return nil
}

// outputSkillsJSON outputs skills in JSON format (for programmatic use)
func outputSkillsJSON(skillService *skillsvc.Service, skillNames []string, skillsDir string) error {
	// Sort skills alphabetically
	sort.Strings(skillNames)
	
	// Create skill list response
	skillList := map[string]interface{}{
		"skills":    make([]map[string]interface{}, 0),
		"count":     len(skillNames),
		"timestamp": time.Now().Format(time.RFC3339),
		"directory": skillsDir,
	}
	
	// Add skill details if verbose mode
	if verbose {
		skillDetails := make([]map[string]interface{}, 0, len(skillNames))
		for _, name := range skillNames {
			if skill, exists := skillService.GetSkill(name); exists {
				detail := map[string]interface{}{
					"name":        skill.Name,
					"description": skill.Description,
					"directory":   skill.DirectoryPath,
				}
				
				// Add optional fields
				if skill.License != "" {
					detail["license"] = skill.License
				}
				if skill.HasScripts {
					detail["has_scripts"] = true
					detail["script_count"] = len(skill.ScriptFiles)
				}
				if skill.HasReferences {
					detail["has_references"] = true
					detail["reference_count"] = len(skill.ReferenceFiles)
				}
				
				skillDetails = append(skillDetails, detail)
			}
		}
		skillList["skills"] = skillDetails
	} else {
		// Simple mode: just names and descriptions
		skillDetails := make([]map[string]interface{}, 0, len(skillNames))
		for _, name := range skillNames {
			if skill, exists := skillService.GetSkill(name); exists {
				skillDetails = append(skillDetails, map[string]interface{}{
					"name":        skill.Name,
					"description": skill.Description,
				})
			}
		}
		skillList["skills"] = skillDetails
	}
	
	// Output as JSON
	output, err := json.MarshalIndent(skillList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}
	
	fmt.Println(string(output))
	
	return nil
}

// checkDockerAvailability checks if Docker or Podman is available
func checkDockerAvailability() bool {
	// Try to create a native executor config and check availability
	config := infraConfig.NewService()
	appConfig, _, _ := config.LoadConfigOrCreateExample(configFile)
	
	skillService := skillsvc.NewService()
	if appConfig != nil {
		skillService.SetConfig(appConfig)
	}
	
	// Try to initialize with auto mode
	skillsDir, _ := filepath.Abs(filepath.Join(filepath.Dir(configFile), "config", "skills"))
	err := skillService.Initialize(skillsDir, skills.ExecutionModeAuto)
	
	// If no error and we have an executor, Docker/Podman is available
	// This is a bit hacky - we're re-initializing, but it's the cleanest way
	// to check availability without exposing executor internals
	return err == nil
}

// wrapText wraps text to specified width with indent for continuation lines
func wrapText(text string, width int, indent string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	
	var lines []string
	var currentLine []string
	currentLength := 0
	
	for _, word := range words {
		wordLen := len(word)
		
		// +1 for space before word
		if currentLength > 0 && currentLength+1+wordLen > width {
			// Start new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = wordLen
		} else {
			currentLine = append(currentLine, word)
			if currentLength > 0 {
				currentLength += 1 + wordLen
			} else {
				currentLength = wordLen
			}
		}
	}
	
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	
	// Join with newline and indent
	for i := 1; i < len(lines); i++ {
		lines[i] = indent + lines[i]
	}
	
	return strings.Join(lines, "\n")
}
