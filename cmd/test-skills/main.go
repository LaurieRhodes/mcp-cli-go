package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
)

func main() {
	fmt.Println("=== Skills System Test ===\n")

	// Create service
	service := skillsvc.NewService()

	// Initialize with skills directory
	skillsDir := "config/skills"
	fmt.Printf("Initializing from: %s\n", skillsDir)

	// Use ExecutionModeAuto to detect Docker/Podman availability
	err := service.Initialize(skillsDir, skills.ExecutionModeAuto)
	if err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}

	// List discovered skills
	skillNames := service.ListSkills()
	fmt.Printf("\n✅ Discovered %d skills:\n\n", len(skillNames))

	for i, name := range skillNames {
		skill, exists := service.GetSkill(name)
		if !exists {
			continue
		}

		fmt.Printf("%2d. %s\n", i+1, skill.Name)
		
		// Truncate description if too long
		desc := skill.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		fmt.Printf("    Description: %s\n", desc)
		
		// Show resources
		resources := []string{}
		if skill.HasReferences {
			resources = append(resources, fmt.Sprintf("references(%d)", len(skill.ReferenceFiles)))
		}
		if skill.HasScripts {
			resources = append(resources, fmt.Sprintf("scripts(%d)", len(skill.ScriptFiles)))
		}
		if skill.HasAssets {
			resources = append(resources, fmt.Sprintf("assets(%d)", len(skill.AssetFiles)))
		}
		if skill.HasWorkflow {
			resources = append(resources, "workflow")
		}
		
		if len(resources) > 0 {
			fmt.Printf("    Resources: %v\n", resources)
		}
		fmt.Println()
	}

	// Test loading a skill
	fmt.Println("=== Testing Passive Load ===\n")
	
	testSkill := "initialization"
	fmt.Printf("Loading skill: %s\n\n", testSkill)

	request := &skills.SkillLoadRequest{
		SkillName:         testSkill,
		Mode:              skills.SkillLoadModePassive,
		IncludeReferences: false,
	}

	result, err := service.LoadSkillByRequest(request)
	if err != nil {
		fmt.Printf("Error loading skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Loaded successfully\n")
	fmt.Printf("   Mode: %s\n", result.Mode)
	fmt.Printf("   Files loaded: %v\n", result.LoadedFiles)
	fmt.Printf("   Content length: %d characters\n", len(result.Content))
	fmt.Printf("\n   Content preview (first 300 chars):\n")
	preview := result.Content
	if len(preview) > 300 {
		preview = preview[:300] + "..."
	}
	fmt.Printf("   %s\n\n", preview)

	// Test MCP tool generation
	fmt.Println("=== Testing MCP Tool Generation ===\n")

	tools, err := service.GenerateRunAsTools()
	if err != nil {
		fmt.Printf("Error generating tools: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated %d MCP tools\n\n", len(tools))
	
	// Show first 2 tools as examples
	for i := 0; i < 2 && i < len(tools); i++ {
		toolJSON, _ := json.MarshalIndent(tools[i], "   ", "  ")
		fmt.Printf("   Tool %d:\n   %s\n\n", i+1, toolJSON)
	}

	fmt.Println("=== All Tests Passed! ===")
}
