package skills

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	domainConfig "github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/sandbox"
	"gopkg.in/yaml.v3"
)

// Service implements the SkillService interface
type Service struct {
	skillsDir               string
	skills                  map[string]*skills.Skill
	enabledSkills           map[string]bool // Track which skills are enabled (nil = all enabled)
	executor                sandbox.Executor
	executionMode           skills.ExecutionMode
	imageMapping            *SkillImageMapping
	appConfig               *domainConfig.ApplicationConfig
	attemptedInitialization bool // Track if we tried to initialize executor
}

// NewService creates a new skill service
func NewService() *Service {
	return &Service{
		skills: make(map[string]*skills.Skill),
	}
}

// SetConfig sets the application configuration for the service
func (s *Service) SetConfig(config *domainConfig.ApplicationConfig) {
	s.appConfig = config
}

// Initialize scans the skills directory and loads all skills
// executionMode can be "passive", "active", or "auto"
func (s *Service) Initialize(skillsDir string, executionMode skills.ExecutionMode) error {
	logging.Info("Initializing skill service from directory: %s", skillsDir)
	logging.Info("Execution mode: %s", executionMode)
	
	// Convert skills directory to absolute path (required for Docker bind mounts)
	absSkillsDir, err := filepath.Abs(skillsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for skills directory: %w", err)
	}
	
	s.skillsDir = absSkillsDir
	logging.Debug("Absolute skills directory: %s", absSkillsDir)
	s.executionMode = executionMode
	
	// Load skill image mapping
	mappingPath := filepath.Join(absSkillsDir, "skill-images.yaml")
	mapping, err := LoadSkillImageMapping(mappingPath)
	if err != nil {
		logging.Warn("Failed to load skill image mapping: %v", err)
		// Continue with default mapping
	} else {
		s.imageMapping = mapping
		logging.Info("✅ Loaded skill image mappings: %d skills, default: %s", 
			len(mapping.Skills), mapping.DefaultImage)
	}
	
	// Initialize executor if needed
	if executionMode == skills.ExecutionModeActive || executionMode == skills.ExecutionModeAuto {
		s.attemptedInitialization = true
		if err := s.initializeExecutor(); err != nil {
			if executionMode == skills.ExecutionModeActive {
				return fmt.Errorf("active mode requires Docker/Podman: %w", err)
			}
			logging.Warn("Docker/Podman not available, falling back to passive mode")
			s.executionMode = skills.ExecutionModePassive
		}
	}
	
	// Check if directory exists
	if _, err := os.Stat(absSkillsDir); os.IsNotExist(err) {
		logging.Warn("Skills directory does not exist: %s", absSkillsDir)
		return nil // Not an error, just no skills available
	}
	
	// Scan directory
	discovered, err := s.ScanSkillsDirectory(absSkillsDir)
	if err != nil {
		return fmt.Errorf("failed to scan skills directory: %w", err)
	}
	
	s.skills = discovered
	
	// Log execution status
	s.logExecutionStatus()
	
	logging.Info("Initialized skill service with %d skills", len(s.skills))
	
	return nil
}

// initializeExecutor sets up the script executor
func (s *Service) initializeExecutor() error {
	config := sandbox.DefaultConfig()
	
	// Configure persistent outputs directory from settings
	if s.appConfig != nil && s.appConfig.Skills != nil {
		config.OutputsDir = s.appConfig.Skills.GetOutputsDir()
		logging.Info("Using outputs directory from config: %s", config.OutputsDir)
	} else {
		// Fallback to default if no config provided
		config.OutputsDir = "/tmp/mcp-outputs"
		logging.Warn("No config provided, using default outputs directory: %s", config.OutputsDir)
	}
	
	// Ensure outputs directory exists
	if err := os.MkdirAll(config.OutputsDir, 0755); err != nil {
		return fmt.Errorf("failed to create outputs directory: %w", err)
	}
	logging.Debug("Outputs directory ready: %s", config.OutputsDir)
	
	// Pass image mapping to executor if available
	if s.imageMapping != nil {
		config.ImageMapping = s.imageMapping
	}
	
	executor, err := sandbox.DetectExecutor(config)
	if err != nil {
		return err
	}
	
	if !executor.IsAvailable() {
		return fmt.Errorf("executor not available")
	}
	
	s.executor = executor
	logging.Info("✅ Executor initialized: %s", executor.GetInfo())
	
	return nil
}

// logExecutionStatus logs the current execution status
func (s *Service) logExecutionStatus() {
	scriptsCount := 0
	skillsWithScripts := []string{}
	
	for _, skill := range s.skills {
		if skill.HasScripts {
			scriptsCount++
			skillsWithScripts = append(skillsWithScripts, skill.Name)
		}
	}
	
	if scriptsCount == 0 {
		return // No skills with scripts
	}
	
	// Only show warning if we actually attempted to initialize executor and failed
	// Don't show warning for explicit passive mode (like --list-skills)
	if s.executor == nil && !s.attemptedInitialization {
		logging.Debug("Skills in passive mode: %d skills with scripts available", scriptsCount)
		return
	}
	
	if s.executor == nil {
		logging.Warn("")
		logging.Warn("╔════════════════════════════════════════════════════════════╗")
		logging.Warn("║  ⚠️  Docker/Podman Not Available                          ║")
		logging.Warn("╚════════════════════════════════════════════════════════════╝")
		logging.Warn("")
		logging.Warn("  %d skills have scripts that require Docker/Podman:", scriptsCount)
		for _, name := range skillsWithScripts {
			logging.Warn("    - %s", name)
		}
		logging.Warn("")
		logging.Warn("  These skills will run in PASSIVE mode (documentation only)")
		logging.Warn("")
		logging.Warn("  To enable script execution:")
		logging.Warn("    1. Install Docker: https://docs.docker.com/get-docker/")
		logging.Warn("       OR Podman: https://podman.io/getting-started/installation")
		logging.Warn("    2. Restart mcp-cli")
		logging.Warn("")
	} else {
		logging.Info("")
		logging.Info("✅ Script execution enabled for %d skills", scriptsCount)
		logging.Info("")
	}
}

// ScanSkillsDirectory scans a directory for Anthropic-compatible skills
func (s *Service) ScanSkillsDirectory(skillsDir string) (map[string]*skills.Skill, error) {
	discovered := make(map[string]*skills.Skill)
	
	// Read directory entries
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}
	
	// Process each subdirectory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		// Skip hidden directories and special files
		if strings.HasPrefix(entry.Name(), ".") || 
		   entry.Name() == "README.md" || 
		   strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		
		skillDir := filepath.Join(skillsDir, entry.Name())
		
		// Try to load the skill
		skill, err := s.LoadSkill(skillDir)
		if err != nil {
			logging.Warn("Failed to load skill from %s: %v", skillDir, err)
			continue
		}
		
		// Validate skill
		if err := s.ValidateSkill(skill); err != nil {
			logging.Warn("Invalid skill %s: %v", skill.Name, err)
			continue
		}
		
		discovered[skill.Name] = skill
		logging.Debug("Discovered skill: %s (%s)", skill.Name, skill.Description)
	}
	
	return discovered, nil
}

// LoadSkill loads a single skill from a directory
func (s *Service) LoadSkill(skillDir string) (*skills.Skill, error) {
	// Check for SKILL.md
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SKILL.md not found in %s", skillDir)
	}
	
	// Parse SKILL.md frontmatter
	frontmatter, err := s.parseSkillFrontmatter(skillMDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Create skill object
	skill := &skills.Skill{
		Name:          frontmatter.Name,
		Description:   frontmatter.Description,
		License:       frontmatter.License,
		DirectoryPath: skillDir,
		SkillMDPath:   skillMDPath,
	}
	
	// Detect resources
	if err := s.detectSkillResources(skill); err != nil {
		return nil, fmt.Errorf("failed to detect resources: %w", err)
	}
	
	return skill, nil
}

// parseSkillFrontmatter parses the YAML frontmatter from SKILL.md
func (s *Service) parseSkillFrontmatter(skillMDPath string) (*skills.SkillFrontmatter, error) {
	file, err := os.Open(skillMDPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SKILL.md: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	// First line should be "---"
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("SKILL.md must start with '---'")
	}
	
	// Read frontmatter content until closing "---"
	var frontmatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SKILL.md: %w", err)
	}
	
	// Parse YAML
	frontmatterYAML := strings.Join(frontmatterLines, "\n")
	
	var frontmatter skills.SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}
	
	return &frontmatter, nil
}

// detectSkillResources detects references, scripts, and assets in a skill directory
func (s *Service) detectSkillResources(skill *skills.Skill) error {
	// Check for references/ directory
	referencesDir := filepath.Join(skill.DirectoryPath, "references")
	if info, err := os.Stat(referencesDir); err == nil && info.IsDir() {
		skill.HasReferences = true
		
		// List reference files
		entries, err := os.ReadDir(referencesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
					skill.ReferenceFiles = append(skill.ReferenceFiles, entry.Name())
				}
			}
		}
	}
	
	// Check for scripts/ directory
	scriptsDir := filepath.Join(skill.DirectoryPath, "scripts")
	if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
		skill.HasScripts = true
		skill.ScriptsDir = scriptsDir
		
		// List script files
		entries, err := os.ReadDir(scriptsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					// Detect script type by extension
					name := entry.Name()
					if strings.HasSuffix(name, ".py") || 
					   strings.HasSuffix(name, ".sh") || 
					   strings.HasSuffix(name, ".bash") {
						skill.ScriptFiles = append(skill.ScriptFiles, name)
						skill.Scripts = append(skill.Scripts, name)
					}
				}
			}
		}
	}
	
	// Also check for scripts in root directory (some Anthropic skills have this structure)
	if !skill.HasScripts {
		entries, err := os.ReadDir(skill.DirectoryPath)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					name := entry.Name()
					// Skip non-script files
					if name == "SKILL.md" || name == "LICENSE.txt" || name == "README.md" {
						continue
					}
					// Check for script extensions
					if strings.HasSuffix(name, ".py") || 
					   strings.HasSuffix(name, ".sh") || 
					   strings.HasSuffix(name, ".bash") {
						skill.HasScripts = true
						skill.ScriptsDir = skill.DirectoryPath
						skill.ScriptFiles = append(skill.ScriptFiles, name)
						skill.Scripts = append(skill.Scripts, name)
					}
				}
			}
		}
	}
	
	// Check for assets/ directory
	assetsDir := filepath.Join(skill.DirectoryPath, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		skill.HasAssets = true
		
		// List asset files (just count, don't load)
		entries, err := os.ReadDir(assetsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					skill.AssetFiles = append(skill.AssetFiles, entry.Name())
				}
			}
		}
	}
	
	// Check for workflow.yaml
	workflowPath := filepath.Join(skill.DirectoryPath, "workflow.yaml")
	if _, err := os.Stat(workflowPath); err == nil {
		skill.HasWorkflow = true
		skill.WorkflowPath = workflowPath
	}
	
	return nil
}

// ValidateSkill validates a skill structure and content
func (s *Service) ValidateSkill(skill *skills.Skill) error {
	return skill.Validate()
}

// ListSkills returns all discovered skill names
func (s *Service) ListSkills() []string {
	names := make([]string, 0, len(s.skills))
	for name := range s.skills {
		names = append(names, name)
	}
	return names
}

// GetSkill retrieves a skill by name
func (s *Service) GetSkill(name string) (*skills.Skill, bool) {
	skill, exists := s.skills[name]
	return skill, exists
}

// LoadMainContent loads the main SKILL.md content (body only, not frontmatter)
func (s *Service) LoadMainContent(skill *skills.Skill) (string, error) {
	// If already loaded, return cached content
	if skill.MainContent != "" {
		return skill.MainContent, nil
	}
	
	file, err := os.Open(skill.SkillMDPath)
	if err != nil {
		return "", fmt.Errorf("failed to open SKILL.md: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	// Skip frontmatter (until second "---")
	foundFirst := false
	foundSecond := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if !foundFirst {
				foundFirst = true
			} else {
				foundSecond = true
				break
			}
		}
	}
	
	if !foundSecond {
		return "", fmt.Errorf("SKILL.md frontmatter not properly closed")
	}
	
	// Read the rest as main content
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading SKILL.md content: %w", err)
	}
	
	content := strings.Join(contentLines, "\n")
	
	// Cache it
	skill.MainContent = content
	
	return content, nil
}

// LoadReference loads a specific reference file
func (s *Service) LoadReference(skill *skills.Skill, referenceName string) (string, error) {
	// Check if reference exists
	found := false
	for _, ref := range skill.ReferenceFiles {
		if ref == referenceName {
			found = true
			break
		}
	}
	
	if !found {
		return "", fmt.Errorf("reference file not found: %s", referenceName)
	}
	
	// Load from cache if available
	if skill.References != nil {
		if content, exists := skill.References[referenceName]; exists {
			return content, nil
		}
	}
	
	// Read the reference file
	refPath := filepath.Join(skill.DirectoryPath, "references", referenceName)
	content, err := os.ReadFile(refPath)
	if err != nil {
		return "", fmt.Errorf("failed to read reference file: %w", err)
	}
	
	// Cache it
	if skill.References == nil {
		skill.References = make(map[string]string)
	}
	skill.References[referenceName] = string(content)
	
	return string(content), nil
}

// LoadAllReferences loads all reference files for a skill
func (s *Service) LoadAllReferences(skill *skills.Skill) (map[string]string, error) {
	if !skill.HasReferences {
		return map[string]string{}, nil
	}
	
	references := make(map[string]string)
	
	for _, refName := range skill.ReferenceFiles {
		content, err := s.LoadReference(skill, refName)
		if err != nil {
			logging.Warn("Failed to load reference %s: %v", refName, err)
			continue
		}
		references[refName] = content
	}
	
	return references, nil
}

// LoadAsPassive loads skill in passive mode (as context)
func (s *Service) LoadAsPassive(skill *skills.Skill, request *skills.SkillLoadRequest) (*skills.SkillLoadResult, error) {
	logging.Info("Loading skill '%s' in passive mode", skill.Name)
	
	result := &skills.SkillLoadResult{
		SkillName:   skill.Name,
		Mode:        skills.SkillLoadModePassive,
		LoadedFiles: []string{},
	}
	
	// Build content
	var contentParts []string
	
	// Add skill header
	contentParts = append(contentParts, fmt.Sprintf("# Skill: %s\n", skill.Name))
	contentParts = append(contentParts, fmt.Sprintf("**Description:** %s\n", skill.Description))
	
	// Load main content
	mainContent, err := s.LoadMainContent(skill)
	if err != nil {
		return nil, fmt.Errorf("failed to load main content: %w", err)
	}
	
	contentParts = append(contentParts, "\n## Main Content\n")
	contentParts = append(contentParts, mainContent)
	result.LoadedFiles = append(result.LoadedFiles, "SKILL.md")
	
	// Load references if requested
	if request.IncludeReferences {
		if skill.HasReferences {
			contentParts = append(contentParts, "\n## References\n")
			
			refs, err := s.LoadAllReferences(skill)
			if err != nil {
				logging.Warn("Failed to load all references: %v", err)
			} else {
				for refName, refContent := range refs {
					contentParts = append(contentParts, fmt.Sprintf("\n### %s\n", refName))
					contentParts = append(contentParts, refContent)
					result.LoadedFiles = append(result.LoadedFiles, filepath.Join("references", refName))
				}
			}
		}
	} else if len(request.ReferenceFiles) > 0 {
		// Load specific reference files
		contentParts = append(contentParts, "\n## References\n")
		
		for _, refName := range request.ReferenceFiles {
			refContent, err := s.LoadReference(skill, refName)
			if err != nil {
				logging.Warn("Failed to load reference %s: %v", refName, err)
				continue
			}
			
			contentParts = append(contentParts, fmt.Sprintf("\n### %s\n", refName))
			contentParts = append(contentParts, refContent)
			result.LoadedFiles = append(result.LoadedFiles, filepath.Join("references", refName))
		}
	}
	
	// Combine all content
	result.Content = strings.Join(contentParts, "\n")
	
	logging.Info("Loaded skill '%s' with %d files", skill.Name, len(result.LoadedFiles))
	
	return result, nil
}

// ExecuteWorkflow executes a skill's workflow.yaml (stub for now)
func (s *Service) ExecuteWorkflow(skill *skills.Skill, inputData string) (*skills.SkillLoadResult, error) {
	if !skill.HasWorkflow {
		return nil, fmt.Errorf("skill %s does not have a workflow.yaml", skill.Name)
	}
	
	// TODO: Implement workflow execution
	// This would integrate with the existing workflow service
	
	return &skills.SkillLoadResult{
		SkillName: skill.Name,
		Mode:      skills.SkillLoadModeActive,
		Result:    fmt.Sprintf("Workflow execution not yet implemented for skill: %s", skill.Name),
	}, nil
}

// ExecuteScript executes a specific script from the skill
func (s *Service) ExecuteScript(skill *skills.Skill, scriptName string, args []string) (string, error) {
	// Check if skill has scripts
	if !skill.HasScripts {
		return "", fmt.Errorf("skill %s does not have scripts", skill.Name)
	}
	
	// Check if executor is available
	if s.executor == nil {
		return "", fmt.Errorf("script execution not available (Docker/Podman not found)")
	}
	
	// Verify script exists
	found := false
	for _, script := range skill.ScriptFiles {
		if script == scriptName {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("script not found: %s", scriptName)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Determine script type and execute
	startTime := time.Now()
	var output string
	var err error
	
	logging.Info("Executing script: %s/%s", skill.Name, scriptName)
	
	if strings.HasSuffix(scriptName, ".py") {
		output, err = s.executor.ExecutePython(ctx, skill.DirectoryPath, "scripts/"+scriptName, args)
	} else if strings.HasSuffix(scriptName, ".sh") || strings.HasSuffix(scriptName, ".bash") {
		output, err = s.executor.ExecuteBash(ctx, skill.DirectoryPath, "scripts/"+scriptName, args)
	} else {
		return "", fmt.Errorf("unsupported script type: %s (must be .py, .sh, or .bash)", scriptName)
	}
	
	duration := time.Since(startTime)
	
	if err != nil {
		logging.Warn("Script execution failed after %v: %v", duration, err)
		return output, err
	}
	
	logging.Info("Script executed successfully in %v", duration)
	return output, nil
}

// ExecuteSkillScript is a convenience method that looks up the skill and executes the script
func (s *Service) ExecuteSkillScript(skillName string, scriptName string, args []string) (*skills.ExecutionResult, error) {
	// Get skill
	skill, exists := s.GetSkill(skillName)
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}
	
	// Execute script
	startTime := time.Now()
	output, err := s.ExecuteScript(skill, scriptName, args)
	duration := time.Since(startTime).Milliseconds()
	
	result := &skills.ExecutionResult{
		Output:   output,
		ExitCode: 0,
		Error:    err,
		Duration: duration,
	}
	
	if err != nil {
		result.ExitCode = 1
	}
	
	return result, nil
}

// validateCodePaths checks if code tries to save to invalid paths and provides helpful errors
func validateCodePaths(code string) error {
	var invalidPaths []string
	
	// Split code into lines
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Check for file operation patterns
		if strings.Contains(line, ".save(") || strings.Contains(line, "open(") || 
		   strings.Contains(line, "to_csv(") || strings.Contains(line, "to_excel(") ||
		   strings.Contains(line, "Path(") {
			// Look for quoted paths in both single and double quotes
			for _, quote := range []string{"'", "\""} {
				if strings.Contains(line, quote+"/") {
					// Extract the path
					start := strings.Index(line, quote+"/")
					if start == -1 {
						continue
					}
					start += len(quote)
					
					end := strings.Index(line[start:], quote)
					if end == -1 {
						continue
					}
					
					path := line[start : start+end]
					
					// Check if it's an absolute path not starting with /outputs or /workspace
					// Allow:
					// - /outputs/ or /outputs (for writing results)
					// - /workspace/ or /workspace (for reading inputs)
					if strings.HasPrefix(path, "/") && 
					   !strings.HasPrefix(path, "/outputs/") && 
					   path != "/outputs" &&
					   !strings.HasPrefix(path, "/workspace/") &&
					   path != "/workspace" {
						invalidPaths = append(invalidPaths, fmt.Sprintf("Line %d: %s", i+1, trimmed))
					}
				}
			}
		}
	}
	
	if len(invalidPaths) > 0 {
		errorMsg := "❌ Invalid file paths detected in code:\n\n"
		errorMsg += "The code tries to save files to paths outside the /outputs/ directory.\n"
		errorMsg += "Files can only be saved to /outputs/ which maps to the host filesystem.\n\n"
		errorMsg += "Invalid paths found:\n"
		for _, path := range invalidPaths {
			errorMsg += "  • " + path + "\n"
		}
		errorMsg += "\n"
		errorMsg += "✅ Correct usage:\n"
		errorMsg += "  Path('/outputs/mydir/file.docx')      ← Shared workflow directory\n"
		errorMsg += "  Path('/workspace/uploaded_file.pdf')  ← Only for files uploaded in chat\n"
		errorMsg += "  open('/outputs/data.txt', 'w')        ← Write results\n\n"
		errorMsg += "💡 Most workflows use /outputs/ for both reading and writing.\n"
		errorMsg += "   Only use /workspace/ for files uploaded during a conversation.\n\n"
		errorMsg += "❌ Invalid usage:\n"
		errorMsg += "  Path('/media/laurie/file.docx')       ← Not accessible in container\n"
		errorMsg += "  Path('/home/user/file.docx')          ← Not accessible in container\n"
		return fmt.Errorf(errorMsg)
	}
	
	return nil
}

// validatePythonSyntax checks if Python code has basic syntax errors
func validatePythonSyntax(code string) error {
	// Look for common syntax errors that AI models make
	lines := strings.Split(code, "\n")
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Check for unescaped quotes in strings
		// This is a simplified check - look for patterns like: "text "more text""
		if strings.Count(line, "\"")%2 != 0 {
			// Odd number of quotes - might be unterminated
			return fmt.Errorf("❌ Python syntax error detected:\n\n" +
				"Line %d has unbalanced quotes - likely unterminated string literal.\n\n" +
				"Line content:\n  %s\n\n" +
				"Common fixes:\n" +
				"1. Escape internal quotes:\n" +
				"   \"He said \\\"hello\\\" to me\"  ← Use backslash before quotes\n\n" +
				"2. Use single quotes for outer string:\n" +
				"   'He said \"hello\" to me'  ← Mix quote types\n\n" +
				"3. Use triple quotes for multi-line strings:\n" +
				"   \"\"\"He said \"hello\" to me\"\"\"  ← Triple quotes allow internal quotes\n\n" +
				"4. Use raw strings if many backslashes:\n" +
				"   r'C:\\path\\to\\file'  ← Raw strings don't interpret backslashes\n",
				i+1, trimmed)
		}
		
		// Check for multiple consecutive quotes that might indicate escaping issues
		if strings.Contains(line, "\"\"\"\"") && !strings.HasPrefix(trimmed, "\"\"\"") {
			return fmt.Errorf("❌ Python syntax error detected:\n\n" +
				"Line %d has suspicious quote pattern (\"\"\"\")\n\n" +
				"Line content:\n  %s\n\n" +
				"This usually indicates improperly escaped quotes in a string.\n\n" +
				"Fix: Escape internal quotes with backslash:\n" +
				"  \"He said \\\"hello\\\" to me\"\n" +
				"Or use single quotes for the outer string:\n" +
				"  'He said \"hello\" to me'\n",
				i+1, trimmed)
		}
	}
	
	return nil
}

// ExecuteCode executes arbitrary code with access to skill's helper libraries
// This is the correct implementation matching Anthropic's design:
// - LLM reads skill documentation
// - LLM writes custom code for the specific task
// - Code executes with access to helper libraries from scripts/
func (s *Service) ExecuteCode(request *skills.CodeExecutionRequest) (*skills.ExecutionResult, error) {
	// Validate request
	if request.SkillName == "" {
		return nil, fmt.Errorf("skill_name is required")
	}
	if request.Language == "" {
		return nil, fmt.Errorf("language is required")
	}
	if request.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	
	// Validate file paths in code
	if err := validateCodePaths(request.Code); err != nil {
		return nil, err
	}
	
	// Validate Python syntax (only for Python code)
	if request.Language == "python" {
		if err := validatePythonSyntax(request.Code); err != nil {
			return nil, err
		}
	}
	
	// Get skill
	skill, exists := s.GetSkill(request.SkillName)
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", request.SkillName)
	}
	
	// Check if executor available
	if s.executor == nil {
		return nil, fmt.Errorf("code execution not available (Docker/Podman not found)")
	}
	
	// Validate language
	if request.Language != "python" && request.Language != "bash" {
		return nil, fmt.Errorf("language '%s' not supported (supported: 'python', 'bash')", request.Language)
	}
	
	// Create temporary workspace
	workspaceDir, err := os.MkdirTemp("", "skill-workspace-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer os.RemoveAll(workspaceDir)
	
	logging.Info("Created workspace: %s", workspaceDir)
	
	// Write files to workspace
	for filename, content := range request.Files {
		filePath := filepath.Join(workspaceDir, filename)
		
		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}
		
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", filename, err)
		}
		
		logging.Debug("Wrote file: %s (%d bytes)", filename, len(content))
	}
	
	// Write code to workspace
	var scriptPath string
	if request.Language == "python" {
		scriptPath = "script.py"
	} else if request.Language == "bash" {
		scriptPath = "script.sh"
	} else {
		return nil, fmt.Errorf("unsupported language: %s", request.Language)
	}
	
	codeFilePath := filepath.Join(workspaceDir, scriptPath)
	if err := os.WriteFile(codeFilePath, []byte(request.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}
	
	logging.Info("Executing code for skill: %s", skill.Name)
	logging.Debug("Code length: %d bytes", len(request.Code))
	
	// Determine timeout
	timeout := 60 * time.Second
	if request.Timeout > 0 {
		timeout = time.Duration(request.Timeout) * time.Second
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Execute with dual mounts
	// - /workspace (read-write): temporary workspace with files and code
	// - /skill (read-only): skill directory for importing helper libraries
	startTime := time.Now()
	var output string
	
	if request.Language == "python" {
		output, err = s.executor.ExecutePythonCode(
			ctx,
			workspaceDir,          // workspace (read-write)
			skill.DirectoryPath,   // skill libs (read-only)
			scriptPath,            // script path relative to workspace
			nil,                   // no args
		)
	} else if request.Language == "bash" {
		output, err = s.executor.ExecuteBashCode(
			ctx,
			workspaceDir,          // workspace (read-write)
			skill.DirectoryPath,   // skill libs (read-only)
			scriptPath,            // script path relative to workspace
			nil,                   // no args
		)
	} else {
		return nil, fmt.Errorf("unsupported language: %s", request.Language)
	}
	
	duration := time.Since(startTime).Milliseconds()
	
	result := &skills.ExecutionResult{
		Output:   output,
		ExitCode: 0,
		Error:    err,
		Duration: duration,
	}
	
	if err != nil {
		result.ExitCode = 1
		logging.Warn("Code execution failed after %dms: %v", duration, err)
	} else {
		logging.Info("Code executed successfully in %dms", duration)
	}
	
	return result, nil
}

// LoadAsActive loads skill in active mode (executes workflow)
func (s *Service) LoadAsActive(skill *skills.Skill, request *skills.SkillLoadRequest) (*skills.SkillLoadResult, error) {
	logging.Info("Loading skill '%s' in active mode", skill.Name)
	
	// Check if skill has a workflow
	if skill.HasWorkflow {
		return s.ExecuteWorkflow(skill, request.InputData)
	}
	
	// If no workflow, return guidance for using execute_skill_code
	return &skills.SkillLoadResult{
		SkillName: skill.Name,
		Mode:      skills.SkillLoadModeActive,
		Result: fmt.Sprintf(`Skill '%s' does not have a workflow.yaml file.

To use this skill programmatically, use the 'execute_skill_code' tool instead:

Example:
{
  "skill_name": "%s",
  "code": "from pptx import Presentation\n\nprs = Presentation()\n# Add slides...\nprs.save('/outputs/output.pptx')",
  "language": "python"
}

Files created in /outputs/ will persist and be available on the host.`, skill.Name, skill.Name),
	}, nil
}

// LoadSkillByRequest loads a skill according to the request
func (s *Service) LoadSkillByRequest(request *skills.SkillLoadRequest) (*skills.SkillLoadResult, error) {
	// Get the skill
	skill, exists := s.GetSkill(request.SkillName)
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", request.SkillName)
	}
	
	// Load based on mode
	switch request.Mode {
	case skills.SkillLoadModePassive:
		return s.LoadAsPassive(skill, request)
	
	case skills.SkillLoadModeActive:
		return s.LoadAsActive(skill, request)
	
	default:
		return nil, fmt.Errorf("invalid load mode: %s", request.Mode)
	}
}

// SetEnabledSkills sets which skills should be enabled
// Pass nil or empty slice to enable all skills (default behavior)
// Pass a list of skill names to enable only those skills
func (s *Service) SetEnabledSkills(skillNames []string) {
	if len(skillNames) == 0 {
		s.enabledSkills = nil
		logging.Debug("All skills enabled (no filter)")
		return
	}
	
	s.enabledSkills = make(map[string]bool)
	for _, name := range skillNames {
		s.enabledSkills[name] = true
	}
	
	logging.Info("Enabled skills filter: %v", skillNames)
}

// IsSkillEnabled checks if a skill is enabled based on the current filter
func (s *Service) IsSkillEnabled(skillName string) bool {
	// If no filter is set, all skills are enabled
	if s.enabledSkills == nil {
		return true
	}
	
	// Check if skill is in the enabled list
	return s.enabledSkills[skillName]
}

// GetEnabledSkills returns a list of currently enabled skill names
func (s *Service) GetEnabledSkills() []string {
	// If no filter, return all skills
	if s.enabledSkills == nil {
		names := make([]string, 0, len(s.skills))
		for name := range s.skills {
			names = append(names, name)
		}
		return names
	}
	
	// Return only enabled skills that exist
	names := make([]string, 0, len(s.enabledSkills))
	for name := range s.enabledSkills {
		if _, exists := s.skills[name]; exists {
			names = append(names, name)
		}
	}
	return names
}

// GenerateRunAsTools generates MCP tool definitions for all skills
func (s *Service) GenerateRunAsTools() ([]map[string]interface{}, error) {
	tools := make([]map[string]interface{}, 0, len(s.skills)+1)
	
	// Add passive mode tools for each enabled skill
	for _, skill := range s.skills {
		// Skip if skill is not enabled
		if !s.IsSkillEnabled(skill.Name) {
			logging.Debug("Skipping disabled skill: %s", skill.Name)
			continue
		}
		
		tool := map[string]interface{}{
			"name":         skill.GetMCPToolName(),
			"description":  skill.GetToolDescription(),
			"template":     "load_skill", // References a template we'll create
			"input_schema": skill.GetMCPInputSchema(),
			// Store skill name in input mapping for template
			"input_mapping": map[string]string{
				"skill_name": skill.Name,
			},
		}
		
		tools = append(tools, tool)
	}
	
	logging.Info("Generated %d skill tools", len(tools))
	
	// Add execute_skill_code tool for dynamic code execution
	executeCodeTool := map[string]interface{}{
		"name": "execute_skill_code",
		"description": "[SKILL CODE EXECUTION] Execute code with access to a skill's helper libraries. " +
			"Code executes in a sandboxed container with the skill's scripts and dependencies. " +
			"\n\n**CRITICAL FILE PATHS - READ THIS CAREFULLY:**" +
			"\n• INPUT files: Read from /outputs/ directory (e.g., /outputs/document_parsing_test/policy.xml)" +
			"\n• OUTPUT files: Write to /outputs/ directory (e.g., /outputs/ism_assessment_full_v2/results.json)" +
			"\n• /outputs/ is the ONLY directory that persists between workflow steps" +
			"\n• NEVER use /workspace/ - it's temporary and gets deleted after execution" +
			"\n• Files in /workspace/ will NOT be accessible to other workflow steps" +
			"\n\n**Usage:** Load the skill first (passive mode) to see documentation and available libraries. " +
			"The skill documentation explains language, libraries, and helper scripts available.",
		"template": "execute_skill_code", // Special marker for this tool
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"skill_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of skill whose helper libraries to use (e.g., 'docx', 'pdf', 'xlsx')",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"python", "bash"},
					"description": "Programming language ('python' or 'bash')",
					"default":     "python",
				},
				"code": map[string]interface{}{
					"type":        "string",
					"description": "Code to execute (Python or Bash). IMPORTANT: Save all files to /outputs/ directory only. Example: doc.save('/outputs/file.docx')",
				},
				"files": map[string]interface{}{
					"type":        "object",
					"description": "Optional files to make available in workspace (filename -> base64 content)",
				},
			},
			"required": []string{"skill_name", "code"},
		},
		"input_mapping": map[string]string{}, // Empty mapping - tool handles its own parameters
	}
	
	tools = append(tools, executeCodeTool)
	
	logging.Info("Generated %d MCP tool definitions from skills", len(tools))
	
	return tools, nil
}
