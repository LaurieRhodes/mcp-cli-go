package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// unmarshalStrict unmarshals YAML with strict field validation
// Unknown fields will cause an error instead of being silently ignored
func unmarshalStrict(data []byte, v interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // Enable strict mode - reject unknown fields
	err := decoder.Decode(v)

	if err != nil {
		// Enhance error message with helpful suggestions
		return enhanceValidationError(err, v)
	}

	return nil
}

// enhanceValidationError adds helpful context to YAML validation errors
func enhanceValidationError(err error, v interface{}) error {
	errMsg := err.Error()

	// Check if it's an unknown field error
	if !strings.Contains(errMsg, "field") || !strings.Contains(errMsg, "not found") {
		return err // Not a field validation error, return as-is
	}

	// Common mistakes and suggestions
	suggestions := map[string]string{
		"pass_env":      "Use 'with:' to pass parameters to child workflows",
		"input":         "Use 'items:' for iterate mode or 'with:' for parameters. 'input' is only valid at step level",
		"inputs":        "Not a valid workflow field. Use 'env:' for environment variables or step-level 'input'",
		"output":        "Not a valid workflow field. Output is returned automatically from steps",
		"outputs":       "Not a valid workflow field. Use step names to reference outputs",
		"output_var":    "Not a valid field. Step outputs are accessible via step names",
		"execution":     "Cannot be nested inside loop. Set execution at workflow or step level",
		"workfow":       "Typo: should be 'workflow'",
		"max_iteration": "Typo: should be 'max_iterations' (plural)",
	}

	// Try to extract field name from error
	var fieldName string
	if strings.Contains(errMsg, "field ") {
		parts := strings.Split(errMsg, "field ")
		if len(parts) > 1 {
			fieldParts := strings.Split(parts[1], " ")
			if len(fieldParts) > 0 {
				fieldName = fieldParts[0]
			}
		}
	}

	// Build enhanced error message
	enhanced := fmt.Sprintf("%s\n\n", errMsg)

	if suggestion, ok := suggestions[fieldName]; ok {
		enhanced += fmt.Sprintf("💡 Suggestion: %s\n\n", suggestion)
	}

	// Add valid fields for common types
	if strings.Contains(errMsg, "type config.WorkflowV2") {
		enhanced += "Valid workflow-level fields:\n"
		enhanced += "  - name (required)\n"
		enhanced += "  - version (required)\n"
		enhanced += "  - description\n"
		enhanced += "  - execution (provider, model, etc.)\n"
		enhanced += "  - env (environment variables)\n"
		enhanced += "  - steps (required)\n"
		enhanced += "  - loops\n"
	} else if strings.Contains(errMsg, "type config.LoopMode") {
		enhanced += "Valid loop fields:\n"
		enhanced += "  - workflow (required)\n"
		enhanced += "  - mode (iterate | refine)\n"
		enhanced += "  - items (for iterate mode)\n"
		enhanced += "  - with (parameters to pass)\n"
		enhanced += "  - max_iterations (required)\n"
		enhanced += "  - parallel, max_workers\n"
		enhanced += "  - on_failure, max_retries, retry_delay\n"
		enhanced += "  - timeout_per_item, total_timeout\n"
		enhanced += "\nExample:\n"
		enhanced += "  loop:\n"
		enhanced += "    workflow: child_worker\n"
		enhanced += "    mode: iterate\n"
		enhanced += "    items: file:///path/to/items.json\n"
		enhanced += "    with:\n"
		enhanced += "      config_param: \"{{env.value}}\"\n"
		enhanced += "    max_iterations: 10\n"
	} else if strings.Contains(errMsg, "type config.StepV2") {
		enhanced += "Valid step fields:\n"
		enhanced += "  - name (required)\n"
		enhanced += "  - run (for prompts)\n"
		enhanced += "  - loop (for child workflows)\n"
		enhanced += "  - rag (for retrieval)\n"
		enhanced += "  - template, consensus, embeddings\n"
		enhanced += "  - provider, model (overrides)\n"
		enhanced += "  - servers, skills\n"
		enhanced += "  - needs (dependencies)\n"
		enhanced += "  - if (conditional)\n"
	}

	return fmt.Errorf("%s", enhanced)
}

// Loader handles loading configuration files in both monolithic and modular formats
type Loader struct {
	baseDir string
}

// NewLoader creates a new config loader
func NewLoader() *Loader {
	return &Loader{}
}

// IncludeDirectives specifies file patterns to include for modular config
type IncludeDirectives struct {
	Providers  string `yaml:"providers,omitempty"`  // e.g., "config/providers/*.yaml"
	Servers    string `yaml:"servers,omitempty"`    // e.g., "config/servers/*.yaml"
	RunAs      string `yaml:"runas,omitempty"`      // e.g., "config/runas/*.yaml"
	Embeddings string `yaml:"embeddings,omitempty"` // e.g., "config/embeddings/*.yaml"
	Templates  string `yaml:"templates,omitempty"`  // e.g., "config/templates/*.yaml" (legacy, backward compatibility)
	Workflows  string `yaml:"workflows,omitempty"`  // e.g., "config/workflows/*.yaml"
	Settings   string `yaml:"settings,omitempty"`   // e.g., "config/settings.yaml"
	RAG        string `yaml:"rag,omitempty"`        // e.g., "config/rag/*.yaml"
	Skills     string `yaml:"skills,omitempty"`     // e.g., "config/skills/*.yaml"
}

// MainConfigFile represents the main config file with optional includes
type MainConfigFile struct {
	Includes *IncludeDirectives `yaml:"includes,omitempty"`
	// Legacy fields for backward compatibility with old monolithic configs
	Servers    map[string]ServerConfig `yaml:"servers,omitempty"`
	AI         *AIConfig               `yaml:"ai,omitempty"`
	Embeddings *EmbeddingsConfig       `yaml:"embeddings,omitempty"`
}

// Load loads configuration from a single file or detects modular structure
func (l *Loader) Load(path string) (*ApplicationConfig, error) {
	// Set base directory for relative path resolution
	l.baseDir = filepath.Dir(path)

	// Read main config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse main config
	var mainConfig MainConfigFile
	if err := unmarshalStrict(data, &mainConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if this is a modular config (has includes)
	if mainConfig.Includes != nil {
		return l.loadModular(mainConfig.Includes)
	}

	// Handle monolithic config
	return l.loadMonolithic(&mainConfig)
}

// loadModular loads configuration from modular structure
func (l *Loader) loadModular(includes *IncludeDirectives) (*ApplicationConfig, error) {
	result := &ApplicationConfig{
		Servers:   make(map[string]ServerConfig),
		Workflows: make(map[string]*WorkflowV2),
	}

	// Load settings first (contains AI, embeddings, etc.)
	if includes.Settings != "" {
		if err := l.loadSettings(includes.Settings, result); err != nil {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}
	}

	// Initialize maps if nil
	if result.Servers == nil {
		result.Servers = make(map[string]ServerConfig)
	}
	if result.Workflows == nil {
		result.Workflows = make(map[string]*WorkflowV2)
	}

	// Load components in order
	if err := l.loadIncludes(includes, result); err != nil {
		return nil, err
	}

	return result, nil
}

// loadMonolithic loads configuration from a single file
func (l *Loader) loadMonolithic(mainConfig *MainConfigFile) (*ApplicationConfig, error) {
	result := &ApplicationConfig{
		Servers:    mainConfig.Servers,
		AI:         mainConfig.AI,
		Embeddings: mainConfig.Embeddings,
		Workflows:  make(map[string]*WorkflowV2),
	}

	// Initialize maps if nil
	if result.Servers == nil {
		result.Servers = make(map[string]ServerConfig)
	}
	if result.Workflows == nil {
		result.Workflows = make(map[string]*WorkflowV2)
	}

	return result, nil
}

// loadSettings loads settings from a YAML file
func (l *Loader) loadSettings(pattern string, result *ApplicationConfig) error {
	// Make pattern absolute if needed
	if !filepath.IsAbs(pattern) && l.baseDir != "" {
		pattern = filepath.Join(l.baseDir, pattern)
	}

	data, err := os.ReadFile(pattern)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse settings into a temporary struct
	var settings struct {
		AI         *AIConfig         `yaml:"ai,omitempty"`
		Embeddings *EmbeddingsConfig `yaml:"embeddings,omitempty"`
		Chat       *ChatConfig       `yaml:"chat,omitempty"`
		Skills     *SkillsConfig     `yaml:"skills,omitempty"`
		RAG        *RagConfig        `yaml:"rag,omitempty"`
	}

	if err := unmarshalStrict(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Copy to result
	result.AI = settings.AI
	result.Embeddings = settings.Embeddings
	result.Chat = settings.Chat
	result.Skills = settings.Skills
	if settings.RAG != nil {
		if result.RAG == nil {
			result.RAG = settings.RAG
		} else {
			// Merge RAG settings: settings.yaml provides defaults, config/rag/*.yaml provides servers
			if settings.RAG.DefaultServer != "" {
				result.RAG.DefaultServer = settings.RAG.DefaultServer
			}
			if settings.RAG.DefaultFusion != "" {
				result.RAG.DefaultFusion = settings.RAG.DefaultFusion
			}
			if settings.RAG.DefaultTopK > 0 {
				result.RAG.DefaultTopK = settings.RAG.DefaultTopK
			}
			result.RAG.QueryExpansion = settings.RAG.QueryExpansion
			result.RAG.Fusion = settings.RAG.Fusion
		}
	}

	return nil
}

// loadIncludes loads all included configuration files
func (l *Loader) loadIncludes(includes *IncludeDirectives, result *ApplicationConfig) error {
	// Load providers
	if includes.Providers != "" {
		if err := l.loadProviders(includes.Providers, result); err != nil {
			return fmt.Errorf("failed to load providers: %w", err)
		}
	}

	// Load embeddings
	if includes.Embeddings != "" {
		if err := l.loadEmbeddings(includes.Embeddings, result); err != nil {
			return fmt.Errorf("failed to load embeddings: %w", err)
		}
	}

	// Load servers
	if includes.Servers != "" {
		if err := l.loadServers(includes.Servers, result); err != nil {
			return fmt.Errorf("failed to load servers: %w", err)
		}
	}

	// Load RAG configurations
	if includes.RAG != "" {
		if err := l.loadRAG(includes.RAG, result); err != nil {
			return fmt.Errorf("failed to load RAG: %w", err)
		}
	}

	// Load workflows (new v2.0 system)
	if includes.Workflows != "" {
		if err := l.loadWorkflows(includes.Workflows, result); err != nil {
			return fmt.Errorf("failed to load workflows: %w", err)
		}
	}

	return nil
}

// loadProviders loads provider configurations from files
func (l *Loader) loadProviders(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read provider file %s: %w", file, err)
		}

		var provider struct {
			InterfaceType InterfaceType  `yaml:"interface_type"`
			ProviderName  string         `yaml:"provider_name"`
			Config        ProviderConfig `yaml:"config"`
		}

		if err := unmarshalStrict(data, &provider); err != nil {
			return fmt.Errorf("failed to parse provider file %s: %w", file, err)
		}

		// Initialize AI config if needed
		if result.AI == nil {
			result.AI = &AIConfig{
				Interfaces: make(map[InterfaceType]InterfaceConfig),
			}
		}
		if result.AI.Interfaces == nil {
			result.AI.Interfaces = make(map[InterfaceType]InterfaceConfig)
		}

		// Get or create interface config
		interfaceConfig, exists := result.AI.Interfaces[provider.InterfaceType]
		if !exists {
			interfaceConfig = InterfaceConfig{
				Providers: make(map[string]ProviderConfig),
			}
		}
		if interfaceConfig.Providers == nil {
			interfaceConfig.Providers = make(map[string]ProviderConfig)
		}

		// Add provider
		interfaceConfig.Providers[provider.ProviderName] = provider.Config
		result.AI.Interfaces[provider.InterfaceType] = interfaceConfig
	}

	return nil
}

// loadEmbeddings loads embedding provider configurations
// loadEmbeddings loads embedding provider configurations
func (l *Loader) loadEmbeddings(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read embedding file %s: %w", file, err)
		}

		var embedding struct {
			InterfaceType InterfaceType           `yaml:"interface_type"`
			ProviderName  string                  `yaml:"provider_name"`
			Config        EmbeddingProviderConfig `yaml:"config"`
		}

		if err := unmarshalStrict(data, &embedding); err != nil {
			return fmt.Errorf("failed to parse embedding file %s: %w", file, err)
		}

		// Initialize embeddings config if needed
		if result.Embeddings == nil {
			result.Embeddings = &EmbeddingsConfig{
				Interfaces: make(map[InterfaceType]EmbeddingInterfaceConfig),
			}
		}
		if result.Embeddings.Interfaces == nil {
			result.Embeddings.Interfaces = make(map[InterfaceType]EmbeddingInterfaceConfig)
		}

		// Get or create interface config
		interfaceConfig, exists := result.Embeddings.Interfaces[embedding.InterfaceType]
		if !exists {
			interfaceConfig = EmbeddingInterfaceConfig{
				Providers: make(map[string]EmbeddingProviderConfig),
			}
		}
		if interfaceConfig.Providers == nil {
			interfaceConfig.Providers = make(map[string]EmbeddingProviderConfig)
		}

		// Add embedding provider
		interfaceConfig.Providers[embedding.ProviderName] = embedding.Config
		result.Embeddings.Interfaces[embedding.InterfaceType] = interfaceConfig
	}

	return nil
}

// loadServers loads server configurations from files
func (l *Loader) loadServers(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read server file %s: %w", file, err)
		}

		var server struct {
			ServerName string       `yaml:"server_name"`
			Config     ServerConfig `yaml:"config"`
		}

		if err := unmarshalStrict(data, &server); err != nil {
			return fmt.Errorf("failed to parse server file %s: %w", file, err)
		}

		result.Servers[server.ServerName] = server.Config
	}

	return nil
}

// loadWorkflows loads workflow v2.0 files
func (l *Loader) loadWorkflows(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	// Get the base workflow directory for calculating relative paths
	basePattern := pattern
	if idx := strings.Index(pattern, "*"); idx != -1 {
		basePattern = pattern[:idx]
	}
	baseWorkflowDir := filepath.Clean(basePattern)

	// CRITICAL: If pattern is relative, join with loader's baseDir first
	// This ensures we resolve relative to the config file's location, not cwd
	if !filepath.IsAbs(baseWorkflowDir) && l.baseDir != "" {
		baseWorkflowDir = filepath.Join(l.baseDir, baseWorkflowDir)
	}

	// Now convert to absolute (should already be absolute after the join above)
	if !filepath.IsAbs(baseWorkflowDir) {
		var err error
		baseWorkflowDir, err = filepath.Abs(baseWorkflowDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for workflow directory: %w", err)
		}
	}

	// Use workflow loader for validation
	workflowLoader := NewWorkflowLoader()

	for _, file := range files {
		// CRITICAL: Convert file path to absolute if it's relative
		// This ensures consistency with baseWorkflowDir which is also absolute
		if !filepath.IsAbs(file) {
			var err error
			file, err = filepath.Abs(file)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for workflow file %s: %w", file, err)
			}
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read workflow file %s: %w", file, err)
		}

		// Check if this is a workflow v2.0 by looking for schema field
		var schemaCheck struct {
			Schema string `yaml:"$schema"`
		}
		// Use non-strict here since we're only checking one field
		if err := yaml.Unmarshal(data, &schemaCheck); err != nil {
			return fmt.Errorf("failed to parse workflow file %s: %w", file, err)
		}

		// Only load workflow v2.0 files
		if schemaCheck.Schema != "workflow/v2.0" {
			// Skip non-v2.0 files
			continue
		}

		// Parse and validate using workflow loader
		workflow, err := workflowLoader.LoadFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to load workflow from %s: %w", file, err)
		}

		// Calculate relative path from base workflow directory
		relPath, err := filepath.Rel(baseWorkflowDir, file)
		if err != nil {
			// If we can't get relative path, just use workflow name
			result.Workflows[workflow.Name] = workflow
		} else {
			// Remove .yaml extension
			relPath = strings.TrimSuffix(relPath, ".yaml")
			relPath = strings.TrimSuffix(relPath, ".yml")

			// If the file is in a subdirectory, use subdirectory/workflowname format
			dir := filepath.Dir(relPath)

			if dir != "." {
				// Use forward slashes for consistency across platforms
				workflowKey := filepath.ToSlash(filepath.Join(dir, workflow.Name))
				result.Workflows[workflowKey] = workflow
			} else {
				// File is in root workflow directory, use just the name
				result.Workflows[workflow.Name] = workflow
			}
		}
	}

	return nil
}

// glob expands a file pattern, supporting * and ** wildcards
func (l *Loader) glob(pattern string) ([]string, error) {
	// Make pattern absolute relative to base directory
	if !filepath.IsAbs(pattern) && l.baseDir != "" {
		pattern = filepath.Join(l.baseDir, pattern)
	}

	// Check if pattern contains **/ (recursive glob)
	if strings.Contains(pattern, "**/") {
		return l.recursiveGlob(pattern)
	}

	// Standard glob
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern %s: %w", pattern, err)
	}

	return files, nil
}

// recursiveGlob handles recursive ** patterns
func (l *Loader) recursiveGlob(pattern string) ([]string, error) {
	// Split on **/
	parts := strings.Split(pattern, "**/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid recursive glob pattern: %s", pattern)
	}

	baseDir := parts[0]
	filePattern := parts[1]

	var results []string

	// Walk directory tree
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches pattern
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return results, nil
}

// Save saves configuration to a file
func (l *Loader) Save(config *ApplicationConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// NewWorkflowLoader creates a workflow-specific loader
func NewWorkflowLoader() *WorkflowLoader {
	return &WorkflowLoader{}
}

// WorkflowLoader is a helper that delegates to the workflow service loader
type WorkflowLoader struct{}

// LoadFromBytes loads a workflow from bytes
func (wl *WorkflowLoader) LoadFromBytes(data []byte) (*WorkflowV2, error) {
	var workflow WorkflowV2
	if err := unmarshalStrict(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	// Basic validation
	if workflow.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if workflow.Version == "" {
		return nil, fmt.Errorf("workflow version is required")
	}
	if len(workflow.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	return &workflow, nil
}

// loadRAG loads RAG server configurations from pattern
func (l *Loader) loadRAG(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil // No RAG configs found, not an error
	}

	// Initialize RAG config if nil
	if result.RAG == nil {
		result.RAG = &RagConfig{
			Servers: make(map[string]RagServerConfig),
		}
	}
	if result.RAG.Servers == nil {
		result.RAG.Servers = make(map[string]RagServerConfig)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read RAG file %s: %w", file, err)
		}

		var ragServer struct {
			ServerName string          `yaml:"server_name"`
			Config     RagServerConfig `yaml:"config"`
		}

		if err := unmarshalStrict(data, &ragServer); err != nil {
			return fmt.Errorf("failed to parse RAG file %s: %w", file, err)
		}

		// Merge the config fields into the server config
		ragServer.Config.ServerName = ragServer.ServerName
		result.RAG.Servers[ragServer.ServerName] = ragServer.Config
	}

	return nil
}
