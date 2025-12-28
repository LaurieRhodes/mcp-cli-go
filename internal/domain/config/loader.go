package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and saving configuration files
type Loader struct {
	baseDir string // Base directory for resolving relative paths
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// IncludeDirectives defines file patterns to include for modular configs
type IncludeDirectives struct {
	Providers  string `yaml:"providers,omitempty"`  // e.g., "config/providers/*.yaml"
	Servers    string `yaml:"servers,omitempty"`    // e.g., "config/servers/*.yaml"
	Embeddings string `yaml:"embeddings,omitempty"` // e.g., "config/embeddings/*.yaml"
	Templates  string `yaml:"templates,omitempty"`  // e.g., "config/templates/**/*.yaml"
	Settings   string `yaml:"settings,omitempty"`   // e.g., "config/settings.yaml"
}

// MainConfigFile represents the main config file with optional includes
type MainConfigFile struct {
	Includes   *IncludeDirectives `yaml:"includes,omitempty"`
	// Legacy fields for backward compatibility - prefer using settings.yaml
	Servers    map[string]ServerConfig      `yaml:"servers,omitempty"`
	AI         *AIConfig                    `yaml:"ai,omitempty"`
	Embeddings *EmbeddingsConfig            `yaml:"embeddings,omitempty"`
	Templates  map[string]*WorkflowTemplate `yaml:"templates,omitempty"`
}

// Load loads configuration from a single file or detects modular structure
func (l *Loader) Load(path string) (*ApplicationConfig, error) {
	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path: %w", err)
	}

	if info.IsDir() {
		// Directory: look for config.yaml inside
		configFile := filepath.Join(path, "config.yaml")
		return l.LoadWithIncludes(configFile)
	}

	// Single file: load directly
	return l.loadSingleFile(path)
}

// LoadWithIncludes loads a config file and processes include directives
func (l *Loader) LoadWithIncludes(mainFile string) (*ApplicationConfig, error) {
	// Set base directory for resolving relative paths
	l.baseDir = filepath.Dir(mainFile)

	// Read main config file
	data, err := os.ReadFile(mainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read main config file: %w", err)
	}

	// Parse main config
	var mainConfig MainConfigFile
	if err := yaml.Unmarshal(data, &mainConfig); err != nil {
		return nil, fmt.Errorf("failed to parse main config file: %w", err)
	}

	// Start with config from main file
	result := &ApplicationConfig{
		Servers:    mainConfig.Servers,
		AI:         mainConfig.AI,
		Embeddings: mainConfig.Embeddings,
		Templates:  mainConfig.Templates,
		TemplatesV2: make(map[string]*TemplateV2),
	}

	// Initialize empty maps if nil
	if result.Servers == nil {
		result.Servers = make(map[string]ServerConfig)
	}
	if result.Templates == nil {
		result.Templates = make(map[string]*WorkflowTemplate)
	}

	// Process includes if present
	if mainConfig.Includes != nil {
		if err := l.processIncludes(mainConfig.Includes, result); err != nil {
			return nil, fmt.Errorf("failed to process includes: %w", err)
		}
	}

	return result, nil
}

// loadSingleFile loads a monolithic config file
func (l *Loader) loadSingleFile(path string) (*ApplicationConfig, error) {
	// Set base directory for resolving relative paths
	l.baseDir = filepath.Dir(path)
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Try to parse as MainConfigFile first to check for includes
	var mainConfig MainConfigFile
	if err := yaml.Unmarshal(data, &mainConfig); err == nil && mainConfig.Includes != nil {
		// File has includes, process them
		result := &ApplicationConfig{
			Servers:    mainConfig.Servers,
			AI:         mainConfig.AI,
			Embeddings: mainConfig.Embeddings,
			Templates:  mainConfig.Templates,
		}
		
		// Initialize empty maps if nil
		if result.Servers == nil {
			result.Servers = make(map[string]ServerConfig)
		}
		if result.Templates == nil {
			result.Templates = make(map[string]*WorkflowTemplate)
		}
		
		// Process includes
		if err := l.processIncludes(mainConfig.Includes, result); err != nil {
			return nil, fmt.Errorf("failed to process includes: %w", err)
		}
		
		return result, nil
	}
	
	// No includes, parse as regular ApplicationConfig
	var config ApplicationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize empty maps if nil
	if config.Servers == nil {
		config.Servers = make(map[string]ServerConfig)
	}
	if config.Templates == nil {
		config.Templates = make(map[string]*WorkflowTemplate)
	}

	return &config, nil
}

// processIncludes processes all include directives and merges into result
func (l *Loader) processIncludes(includes *IncludeDirectives, result *ApplicationConfig) error {
	// Process settings first (if present)
	if includes.Settings != "" {
		if err := l.loadSettings(includes.Settings, result); err != nil {
			return fmt.Errorf("failed to load settings: %w", err)
		}
	}

	// Process provider includes
	if includes.Providers != "" {
		if err := l.loadProviders(includes.Providers, result); err != nil {
			return fmt.Errorf("failed to load providers: %w", err)
		}
	}

	// Process server includes
	if includes.Servers != "" {
		if err := l.loadServers(includes.Servers, result); err != nil {
			return fmt.Errorf("failed to load servers: %w", err)
		}
	}

	// Process embedding includes
	if includes.Embeddings != "" {
		if err := l.loadEmbeddings(includes.Embeddings, result); err != nil {
			return fmt.Errorf("failed to load embeddings: %w", err)
		}
	}

	// Process template includes
	if includes.Templates != "" {
		if err := l.loadTemplates(includes.Templates, result); err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}
	}

	return nil
}

// loadSettings loads global settings from settings.yaml
func (l *Loader) loadSettings(path string, result *ApplicationConfig) error {
	// Make path absolute relative to base directory
	if !filepath.IsAbs(path) && l.baseDir != "" {
		path = filepath.Join(l.baseDir, path)
	}

	// Read settings file
	data, err := os.ReadFile(path)
	if err != nil {
		// Settings file is optional
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read settings file %s: %w", path, err)
	}

	// Parse settings file
	var settings struct {
		AI         *AIConfig         `yaml:"ai,omitempty"`
		Embeddings *EmbeddingsConfig `yaml:"embeddings,omitempty"`
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file %s: %w", path, err)
	}

	// Merge AI settings (settings.yaml takes precedence over main config)
	if settings.AI != nil {
		if result.AI == nil {
			result.AI = settings.AI
		} else {
			// Merge settings, preferring settings.yaml values
			if settings.AI.DefaultProvider != "" {
				result.AI.DefaultProvider = settings.AI.DefaultProvider
			}
			if settings.AI.DefaultSystemPrompt != "" {
				result.AI.DefaultSystemPrompt = settings.AI.DefaultSystemPrompt
			}
		}
	}

	// Merge embeddings settings
	if settings.Embeddings != nil {
		if result.Embeddings == nil {
			result.Embeddings = settings.Embeddings
		} else {
			// Merge settings, preferring settings.yaml values
			if settings.Embeddings.DefaultChunkStrategy != "" {
				result.Embeddings.DefaultChunkStrategy = settings.Embeddings.DefaultChunkStrategy
			}
			if settings.Embeddings.DefaultMaxChunkSize > 0 {
				result.Embeddings.DefaultMaxChunkSize = settings.Embeddings.DefaultMaxChunkSize
			}
			if settings.Embeddings.DefaultOverlap > 0 {
				result.Embeddings.DefaultOverlap = settings.Embeddings.DefaultOverlap
			}
			if settings.Embeddings.OutputPrecision > 0 {
				result.Embeddings.OutputPrecision = settings.Embeddings.OutputPrecision
			}
		}
	}

	return nil
}

// loadProviders loads provider configurations from files matching pattern
func (l *Loader) loadProviders(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	// Ensure AI config exists
	if result.AI == nil {
		result.AI = &AIConfig{
			Interfaces: make(map[InterfaceType]InterfaceConfig),
		}
	}
	if result.AI.Interfaces == nil {
		result.AI.Interfaces = make(map[InterfaceType]InterfaceConfig)
	}

	for _, file := range files {
		// Read provider file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read provider file %s: %w", file, err)
		}

		// Parse as a single provider config wrapped in interface structure
		var providerFile struct {
			InterfaceType InterfaceType `yaml:"interface_type"`
			ProviderName  string        `yaml:"provider_name"`
			Config        ProviderConfig `yaml:"config"`
		}

		if err := yaml.Unmarshal(data, &providerFile); err != nil {
			return fmt.Errorf("failed to parse provider file %s: %w", file, err)
		}

		// Add to appropriate interface
		interfaceType := providerFile.InterfaceType
		if interfaceType == "" {
			// Auto-detect interface type from provider name
			interfaceType = inferInterfaceType(providerFile.ProviderName)
		}

		// Get or create interface config
		interfaceConfig, exists := result.AI.Interfaces[interfaceType]
		if !exists {
			interfaceConfig = InterfaceConfig{
				Providers: make(map[string]ProviderConfig),
			}
		}
		if interfaceConfig.Providers == nil {
			interfaceConfig.Providers = make(map[string]ProviderConfig)
		}

		// Add provider to interface
		interfaceConfig.Providers[providerFile.ProviderName] = providerFile.Config
		result.AI.Interfaces[interfaceType] = interfaceConfig
	}

	return nil
}

// loadServers loads server configurations from files matching pattern
func (l *Loader) loadServers(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		// Read server file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read server file %s: %w", file, err)
		}

		// Parse as a single server config
		var serverFile struct {
			ServerName string       `yaml:"server_name"`
			Config     ServerConfig `yaml:"config"`
		}

		if err := yaml.Unmarshal(data, &serverFile); err != nil {
			return fmt.Errorf("failed to parse server file %s: %w", file, err)
		}

		// Add to servers map
		result.Servers[serverFile.ServerName] = serverFile.Config
	}

	return nil
}

// loadEmbeddings loads embedding configurations from files matching pattern
func (l *Loader) loadEmbeddings(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	// Ensure embeddings config exists
	if result.Embeddings == nil {
		result.Embeddings = &EmbeddingsConfig{
			Interfaces: make(map[InterfaceType]EmbeddingInterfaceConfig),
		}
	}
	if result.Embeddings.Interfaces == nil {
		result.Embeddings.Interfaces = make(map[InterfaceType]EmbeddingInterfaceConfig)
	}

	for _, file := range files {
		// Read embedding file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read embedding file %s: %w", file, err)
		}

		// Parse as a single embedding provider config
		var embeddingFile struct {
			InterfaceType InterfaceType           `yaml:"interface_type"`
			ProviderName  string                  `yaml:"provider_name"`
			Config        EmbeddingProviderConfig `yaml:"config"`
		}

		if err := yaml.Unmarshal(data, &embeddingFile); err != nil {
			return fmt.Errorf("failed to parse embedding file %s: %w", file, err)
		}

		// Add to appropriate interface
		interfaceType := embeddingFile.InterfaceType
		if interfaceType == "" {
			interfaceType = inferInterfaceType(embeddingFile.ProviderName)
		}

		// Get or create interface config
		interfaceConfig, exists := result.Embeddings.Interfaces[interfaceType]
		if !exists {
			interfaceConfig = EmbeddingInterfaceConfig{
				Providers: make(map[string]EmbeddingProviderConfig),
			}
		}
		if interfaceConfig.Providers == nil {
			interfaceConfig.Providers = make(map[string]EmbeddingProviderConfig)
		}

		// Add provider to interface
		interfaceConfig.Providers[embeddingFile.ProviderName] = embeddingFile.Config
		result.Embeddings.Interfaces[interfaceType] = interfaceConfig
	}

	return nil
}

// loadTemplates loads workflow templates from files matching pattern
func (l *Loader) loadTemplates(pattern string, result *ApplicationConfig) error {
	files, err := l.glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		// Read template file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", file, err)
		}

		// Try to detect template version by checking for "version" field
		var versionCheck struct {
			Version string `yaml:"version"`
		}
		yaml.Unmarshal(data, &versionCheck)

		if versionCheck.Version != "" {
			// This is a template v2
			var templateV2 TemplateV2
			if err := yaml.Unmarshal(data, &templateV2); err != nil {
				return fmt.Errorf("failed to parse template v2 file %s: %w", file, err)
			}

			// Validate template v2
			// Basic validation - full validation happens in template package
			if templateV2.Name == "" {
				templateV2.Name = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			}

			// Add to templates v2 map
			if result.TemplatesV2 == nil {
				result.TemplatesV2 = make(map[string]*TemplateV2)
			}
			result.TemplatesV2[templateV2.Name] = &templateV2
		} else {
			// This is an old workflow template
			var template WorkflowTemplate
			if err := yaml.Unmarshal(data, &template); err != nil {
				return fmt.Errorf("failed to parse template file %s: %w", file, err)
			}

			// Use template name from file or filename as key
			templateName := template.Name
			if templateName == "" {
				// Use filename without extension as template name
				templateName = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
				template.Name = templateName
			}

			// Add to old templates map
			result.Templates[templateName] = &template
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
		return nil, fmt.Errorf("failed to walk directory for pattern %s: %w", pattern, err)
	}

	return results, nil
}

// Save saves configuration to a file
func (l *Loader) Save(config *ApplicationConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// inferInterfaceType determines interface type from provider name (fallback for configs without interface_type)
func inferInterfaceType(providerName string) InterfaceType {
	// This is a fallback for old configs that don't specify interface_type
	// New configs should always specify interface_type in the provider file
	providerLower := strings.ToLower(providerName)
	
	switch {
	case providerLower == "anthropic":
		return AnthropicNative
	case providerLower == "ollama":
		return OllamaNative
	case providerLower == "gemini":
		return GeminiNative
	case strings.Contains(providerLower, "azure"):
		return AzureOpenAI
	case strings.Contains(providerLower, "bedrock"):
		return AWSBedrock
	case strings.Contains(providerLower, "vertex"):
		return GCPVertexAI
	default:
		// Safe default for OpenAI-compatible providers
		// This includes: openai, deepseek, openrouter, lmstudio, and any custom providers
		return OpenAICompatible
	}
}
