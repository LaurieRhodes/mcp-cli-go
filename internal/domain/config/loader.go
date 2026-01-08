package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	Workflows  string `yaml:"workflows,omitempty"`  // e.g., "config/workflows/*.yaml"
	Settings   string `yaml:"settings,omitempty"`   // e.g., "config/settings.yaml"
}

// MainConfigFile represents the main config file with optional includes
type MainConfigFile struct {
	Includes   *IncludeDirectives `yaml:"includes,omitempty"`
	// Legacy fields for backward compatibility with old monolithic configs
	Servers    map[string]ServerConfig      `yaml:"servers,omitempty"`
	AI         *AIConfig                    `yaml:"ai,omitempty"`
	Embeddings *EmbeddingsConfig            `yaml:"embeddings,omitempty"`
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
	if err := yaml.Unmarshal(data, &mainConfig); err != nil {
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
		Servers:   mainConfig.Servers,
		AI:        mainConfig.AI,
		Embeddings: mainConfig.Embeddings,
		Workflows: make(map[string]*WorkflowV2),
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
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Copy to result
	result.AI = settings.AI
	result.Embeddings = settings.Embeddings
	result.Chat = settings.Chat
	result.Skills = settings.Skills

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

		if err := yaml.Unmarshal(data, &provider); err != nil {
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
			InterfaceType InterfaceType         `yaml:"interface_type"`
			ProviderName  string                `yaml:"provider_name"`
			Config        EmbeddingProviderConfig `yaml:"config"`
		}

		if err := yaml.Unmarshal(data, &embedding); err != nil {
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

		if err := yaml.Unmarshal(data, &server); err != nil {
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

	// Use workflow loader for validation
	workflowLoader := NewWorkflowLoader()

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read workflow file %s: %w", file, err)
		}

		// Check if this is a workflow v2.0 by looking for schema field
		var schemaCheck struct {
			Schema string `yaml:"$schema"`
		}
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
	if err := yaml.Unmarshal(data, &workflow); err != nil {
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
