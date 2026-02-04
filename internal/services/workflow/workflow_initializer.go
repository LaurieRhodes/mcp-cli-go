package workflow

import (
	"fmt"
	"os"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	infraSkills "github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// InitializeWorkflowServerManager initializes the server manager for a workflow
// This follows the exact same logic as standalone workflow execution in cmd/workflow.go
func InitializeWorkflowServerManager(
	workflow *config.WorkflowV2,
	appConfig *config.ApplicationConfig,
	configFile string,
) (domain.MCPServerManager, error) {
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] InitializeWorkflowServerManager called for: %s\n", workflow.Name)
	logging.Debug("[WORKFLOW_INIT] Initializing server manager for workflow: %s", workflow.Name)
	
	// Extract skills from workflow (same as collectSkillsFromWorkflow in cmd/workflow.go)
	skills := collectWorkflowSkills(workflow)
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Extracted skills: %v\n", skills)
	logging.Debug("[WORKFLOW_INIT] Extracted skills from workflow: %v", skills)
	
	if len(skills) == 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] No skills found\n")
		logging.Debug("[WORKFLOW_INIT] No skills found, returning nil server manager")
		// No skills needed
		return nil, nil
	}
	
	// Initialize built-in skills (same as executeWorkflowWithoutServers in cmd/workflow.go)
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Initializing skills: %v\n", skills)
	logging.Info("Initializing built-in skills for workflow: %v", skills)
	skillService, err := infraSkills.InitializeBuiltinSkills(configFile, appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize built-in skills: %w", err)
	}
	logging.Info("Built-in skills service initialized successfully")
	
	// Create server manager with skills (no external servers)
	logging.Info("Creating server manager with built-in skills only")
	serverManager := infraSkills.NewSkillsAwareServerManager(nil, skillService)
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Created server manager\n")
	
	// CRITICAL FIX: Verify tools are available before returning
	// This prevents race conditions with fast models in parallel execution
	// Each workflow is isolated, but initialization must be synchronous
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Verifying server manager readiness...\n")
	if err := verifyServerManagerReady(serverManager, skills); err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Verification FAILED: %v\n", err)
		return nil, fmt.Errorf("server manager not ready: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] ✅ Server manager verified ready!\n")
	logging.Info("✅ Server manager verified ready with all required tools")
	
	return serverManager, nil
}

// collectWorkflowSkills extracts all unique skill names from a workflow
// This is identical to collectSkillsFromWorkflow in cmd/workflow.go
func collectWorkflowSkills(wf *config.WorkflowV2) []string {
	skillSet := make(map[string]bool)
	
	// Collect from execution level
	for _, skill := range wf.Execution.Skills {
		skillSet[skill] = true
	}
	
	// Collect from steps
	for _, step := range wf.Steps {
		for _, skill := range step.Skills {
			skillSet[skill] = true
		}
	}
	
	// Convert to slice
	skills := make([]string, 0, len(skillSet))
	for skill := range skillSet {
		skills = append(skills, skill)
	}
	
	return skills
}

// verifyServerManagerReady ensures GetAvailableTools returns expected tools
// Retries with exponential backoff to handle async initialization (Docker startup, etc.)
// This is critical for fast models in parallel execution where race conditions can occur
func verifyServerManagerReady(serverManager domain.MCPServerManager, expectedSkills []string) error {
	const (
		maxRetries = 5
		baseDelay  = 100 * time.Millisecond
	)
	
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Starting readiness verification (max %d attempts)\n", maxRetries)
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt)) // Exponential backoff: 100ms, 200ms, 400ms, 800ms, 1600ms
			fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] Retry %d/%d after %v delay\n", attempt+1, maxRetries, delay)
			logging.Debug("[SKILL_INIT] Retry %d/%d after %v delay", attempt+1, maxRetries, delay)
			time.Sleep(delay)
		}
		
		// Attempt to get available tools
		tools, err := serverManager.GetAvailableTools()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] GetAvailableTools failed on attempt %d: %v\n", attempt+1, err)
			logging.Warn("[SKILL_INIT] GetAvailableTools failed (attempt %d/%d): %v", attempt+1, maxRetries, err)
			continue
		}
		
		// Check for essential tool: skills_run_helper_script
		// This tool must be available for skill-based workflows to function
		hasRunHelperScript := false
		skillToolCount := 0
		
		for _, tool := range tools {
			if tool.Function.Name == "skills_run_helper_script" {
				hasRunHelperScript = true
			}
			// Count all skills_ prefixed tools for diagnostics
			if len(tool.Function.Name) >= 7 && tool.Function.Name[:7] == "skills_" {
				skillToolCount++
			}
		}
		
		if hasRunHelperScript {
			fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] ✅ Found skills_run_helper_script! (%d skill tools, %d total tools)\n", skillToolCount, len(tools))
			logging.Info("[SKILL_INIT] ✅ Ready! Found skills_run_helper_script and %d total skill tools", skillToolCount)
			return nil
		}
		
		fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] ⚠️  Attempt %d/%d: Missing skills_run_helper_script (found %d tools, %d skill tools)\n", 
			attempt+1, maxRetries, len(tools), skillToolCount)
		logging.Warn("[SKILL_INIT] Not ready (attempt %d/%d): Found %d total tools, %d skill tools, but missing skills_run_helper_script", 
			attempt+1, maxRetries, len(tools), skillToolCount)
	}
	
	fmt.Fprintf(os.Stderr, "[DEBUG_PRINT] ❌ Verification FAILED after %d attempts\n", maxRetries)
	return fmt.Errorf("server manager not ready after %d attempts (max %v total delay) - tools not available", 
		maxRetries, baseDelay*(1<<uint(maxRetries-1)))
}
