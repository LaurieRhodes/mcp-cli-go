package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	domainSkills "github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	skillsvc "github.com/LaurieRhodes/mcp-cli-go/internal/services/skills"
)

// SkillsAwareServerManager wraps external servers and adds built-in skills
// It implements the MCPServerManager interface by combining:
// - External MCP servers (from externalServers)
// - Built-in skills service (from skillService)
type SkillsAwareServerManager struct {
	externalServers domain.MCPServerManager
	skillService    *skillsvc.Service
}

// NewSkillsAwareServerManager creates a new server manager that includes built-in skills
func NewSkillsAwareServerManager(external domain.MCPServerManager, skills *skillsvc.Service) domain.MCPServerManager {
	logging.Info("Creating SkillsAwareServerManager with built-in skills")
	return &SkillsAwareServerManager{
		externalServers: external,
		skillService:    skills,
	}
}

// GetAvailableTools returns all tools from external servers + built-in skills
func (sm *SkillsAwareServerManager) GetAvailableTools() ([]domain.Tool, error) {
	// Get tools from external servers (may be empty)
	externalTools := []domain.Tool{}
	if sm.externalServers != nil {
		tools, err := sm.externalServers.GetAvailableTools()
		if err == nil {
			externalTools = tools
		}
	}
	
	// Generate tools from built-in skills
	skillTools := sm.generateSkillTools()
	
	allTools := append(externalTools, skillTools...)
	logging.Debug("Total tools available: %d (external: %d, skills: %d)", 
		len(allTools), len(externalTools), len(skillTools))
	
	return allTools, nil
}

// ExecuteTool routes tool execution to either built-in skills or external servers
func (sm *SkillsAwareServerManager) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// Check if this is a skill tool (prefixed with "skills_")
	if strings.HasPrefix(toolName, "skills_") {
		logging.Debug("Routing tool '%s' to built-in skills service", toolName)
		return sm.executeSkillTool(ctx, toolName, arguments)
	}
	
	// Otherwise delegate to external servers
	if sm.externalServers == nil {
		return "", fmt.Errorf("tool '%s' not found (no external servers available)", toolName)
	}
	
	logging.Debug("Routing tool '%s' to external servers", toolName)
	return sm.externalServers.ExecuteTool(ctx, toolName, arguments)
}

// generateSkillTools creates MCP tools from all available skills
// This matches the logic in cmd/serve.go
func (sm *SkillsAwareServerManager) generateSkillTools() []domain.Tool {
	tools := []domain.Tool{}
	
	// Generate a tool for each skill
	for _, skillName := range sm.skillService.ListSkills() {
		skill, exists := sm.skillService.GetSkill(skillName)
		if !exists {
			continue
		}
		
		tool := domain.Tool{
			Type: "function",
			Function: domain.ToolFunction{
				Name:        "skills_" + skill.GetMCPToolName(),
				Description: skill.GetToolDescription(),
				Parameters:  skill.GetMCPInputSchema(),
			},
		}
		tools = append(tools, tool)
		logging.Debug("Generated tool '%s' for skill '%s'", tool.Function.Name, skillName)
	}
	
	// Add execute_skill_code tool for dynamic code execution
	executeCodeTool := domain.Tool{
		Type: "function",
		Function: domain.ToolFunction{
			Name: "skills_execute_skill_code",
			Description: "[SKILL CODE EXECUTION] Execute code with access to a skill's helper libraries. " +
				"Use this to: (1) Create documents dynamically, (2) Process files with custom logic, " +
				"(3) Use skill helper libraries (e.g., Document class from docx skill). " +
				"The code executes in a sandboxed environment with the skill's scripts/ directory available for imports via PYTHONPATH.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"skill_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of skill whose helper libraries to use (e.g., 'docx', 'pdf', 'xlsx')",
					},
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Code to execute (Python or Bash). Can import from 'scripts' module to use skill helper libraries.",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"python", "bash"},
						"default":     "python",
						"description": "Programming language ('python' or 'bash')",
					},
					"files": map[string]interface{}{
						"type":        "object",
						"description": "Optional files to make available in workspace (filename -> base64 content)",
					},
				},
				"required": []string{"skill_name", "code"},
			},
		},
	}
	tools = append(tools, executeCodeTool)
	
	logging.Info("Generated %d tools from built-in skills", len(tools))
	
	// Add run_helper_script tool for direct script execution
	runHelperScriptTool := domain.Tool{
		Type: "function",
		Function: domain.ToolFunction{
			Name: "skills_run_helper_script",
			Description: "[HELPER SCRIPT EXECUTION] Directly execute a pre-written helper script from a skill's scripts/ directory. " +
				"This is more efficient than execute_skill_code for running existing scripts. " +
				"The script must exist in /skill/scripts/ and can be Python (.py) or Bash (.sh). " +
				"Input/output files use /outputs/ directory.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"skill_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of skill containing the script (e.g., 'python-context-builder')",
					},
					"script_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of script file in /skill/scripts/ (e.g., 'process_chunk.py')",
					},
					"args": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command-line arguments to pass to the script",
					},
				},
				"required": []string{"skill_name", "script_name"},
			},
		},
	}
	tools = append(tools, runHelperScriptTool)
	return tools
}

// executeSkillTool handles execution of skill-related tools
func (sm *SkillsAwareServerManager) executeSkillTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	// Strip "skills_" prefix to get actual tool name
	actualToolName := strings.TrimPrefix(toolName, "skills_")
	
	// Special handling for execute_skill_code
	if actualToolName == "execute_skill_code" {
		return sm.executeSkillCode(ctx, arguments)
	}
	
	// Special handling for run_helper_script
	if actualToolName == "run_helper_script" {
		return sm.runHelperScript(ctx, arguments)
	}
	
	// Find the skill that matches this tool
	for _, skillName := range sm.skillService.ListSkills() {
		skill, exists := sm.skillService.GetSkill(skillName)
		if !exists {
			continue
		}
		
		if skill.GetMCPToolName() == actualToolName {
			// This is a skill load request
			return sm.loadSkill(ctx, skill, arguments)
		}
	}
	
	return "", fmt.Errorf("skill tool '%s' not found", toolName)
}

// loadSkill loads a skill (passive or active mode based on arguments)
func (sm *SkillsAwareServerManager) loadSkill(ctx context.Context, skill *domainSkills.Skill, arguments map[string]interface{}) (string, error) {
	// Determine mode from arguments
	mode := "passive" // default
	if modeArg, ok := arguments["mode"].(string); ok {
		mode = modeArg
	}
	
	// Create skill load request
	request := &domainSkills.SkillLoadRequest{
		Mode: domainSkills.SkillLoadMode(mode),
	}
	
	// Add include_references if specified
	if includeRefs, ok := arguments["include_references"]; ok {
		if refsBool, ok := includeRefs.(bool); ok {
			request.IncludeReferences = refsBool
		}
	}
	
	// Add reference_files if specified
	if refFiles, ok := arguments["reference_files"].([]interface{}); ok {
		refs := []string{}
		for _, ref := range refFiles {
			if refStr, ok := ref.(string); ok {
				refs = append(refs, refStr)
			}
		}
		request.ReferenceFiles = refs
	}
	
	// Add input_data if specified
	if inputData, ok := arguments["input_data"].(string); ok {
		request.InputData = inputData
	}
	
	// Load the skill
	var result *domainSkills.SkillLoadResult
	var err error
	
	if mode == "active" {
		result, err = sm.skillService.LoadAsActive(skill, request)
	} else {
		result, err = sm.skillService.LoadAsPassive(skill, request)
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to load skill '%s': %w", skill.Name, err)
	}
	
	// Format result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal skill result: %w", err)
	}
	
	return string(resultJSON), nil
}

// executeSkillCode handles the execute_skill_code tool
func (sm *SkillsAwareServerManager) executeSkillCode(ctx context.Context, arguments map[string]interface{}) (string, error) {
	// Extract required parameters
	skillName, ok := arguments["skill_name"].(string)
	if !ok {
		return "", fmt.Errorf("skill_name is required - example: execute_skill_code with skill_name=python-context-builder and code=your_python_code")
	}
	
	code, ok := arguments["code"].(string)
	if !ok {
		return "", fmt.Errorf("code is required")
	}
	
	// Extract optional parameters
	language := "python" // default
	if lang, ok := arguments["language"].(string); ok {
		language = lang
	}
	
	var files map[string][]byte
	if filesArg, ok := arguments["files"].(map[string]interface{}); ok {
		files = make(map[string][]byte)
		for k, v := range filesArg {
			if vStr, ok := v.(string); ok {
				files[k] = []byte(vStr)
			}
		}
	}
	
	// Create code execution request
	request := &domainSkills.CodeExecutionRequest{
		SkillName: skillName,
		Code:      code,
		Language:  language,
		Files:     files,
	}
	
	// Execute the code
	result, err := sm.skillService.ExecuteCode(request)
	if err != nil {
		return "", fmt.Errorf("code execution failed: %w", err)
	}
	
	// Format result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal execution result: %w", err)
	}
	
	return string(resultJSON), nil
}

// runHelperScript handles the run_helper_script tool
func (sm *SkillsAwareServerManager) runHelperScript(ctx context.Context, arguments map[string]interface{}) (string, error) {
	// Extract required parameters
	skillName, ok := arguments["skill_name"].(string)
	if !ok {
		return "", fmt.Errorf("skill_name is required - example: run_helper_script with skill_name=python-context-builder, script_name=process_chunk.py, args=array")
	}
	
	scriptName, ok := arguments["script_name"].(string)
	if !ok {
		return "", fmt.Errorf("script_name is required")
	}
	
	// Extract optional args
	var args []string
	if argsArg, ok := arguments["args"].([]interface{}); ok {
		for _, arg := range argsArg {
			if argStr, ok := arg.(string); ok {
				args = append(args, argStr)
			}
		}
	}
	
	// Create script execution request
	request := &domainSkills.HelperScriptRequest{
		SkillName:  skillName,
		ScriptName: scriptName,
		Args:       args,
	}
	
	// Execute the script
	result, err := sm.skillService.RunHelperScript(request)
	if err != nil {
		return "", fmt.Errorf("script execution failed: %w", err)
	}
	
	// Format result as JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal execution result: %w", err)
	}
	
	return string(resultJSON), nil
}

// Remaining MCPServerManager interface methods - delegate to external servers

func (sm *SkillsAwareServerManager) StartServer(ctx context.Context, serverName string, cfg *config.ServerConfig) (domain.MCPServer, error) {
	if sm.externalServers == nil {
		return nil, fmt.Errorf("no external servers available")
	}
	return sm.externalServers.StartServer(ctx, serverName, cfg)
}

func (sm *SkillsAwareServerManager) StopServer(serverName string) error {
	if sm.externalServers == nil {
		return fmt.Errorf("no external servers available")
	}
	return sm.externalServers.StopServer(serverName)
}

func (sm *SkillsAwareServerManager) GetServer(serverName string) (domain.MCPServer, bool) {
	if sm.externalServers == nil {
		return nil, false
	}
	return sm.externalServers.GetServer(serverName)
}

func (sm *SkillsAwareServerManager) ListServers() map[string]domain.MCPServer {
	if sm.externalServers == nil {
		return make(map[string]domain.MCPServer)
	}
	return sm.externalServers.ListServers()
}

func (sm *SkillsAwareServerManager) StopAll() error {
	if sm.externalServers == nil {
		return nil
	}
	return sm.externalServers.StopAll()
}
