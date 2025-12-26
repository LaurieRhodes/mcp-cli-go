package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ModularConfigGenerator creates a modular config directory structure
type ModularConfigGenerator struct {
	baseDir string
}

// NewModularConfigGenerator creates a new modular config generator
func NewModularConfigGenerator(baseDir string) *ModularConfigGenerator {
	return &ModularConfigGenerator{
		baseDir: baseDir,
	}
}

// Generate creates the modular config directory structure
func (g *ModularConfigGenerator) Generate(config *GeneratorConfig) error {
	// Create base directory
	if err := os.MkdirAll(g.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"providers", "embeddings", "servers", "templates", "runas"}
	for _, dir := range dirs {
		path := filepath.Join(g.baseDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Create main config.yaml with includes at parent level (next to executable)
	if err := g.createMainConfig(config); err != nil {
		return fmt.Errorf("failed to create main config: %w", err)
	}

	// Create settings.yaml in config directory
	if err := g.createSettings(config); err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}

	// Create provider files (LLM)
	if err := g.createProviderFiles(config); err != nil {
		return fmt.Errorf("failed to create provider files: %w", err)
	}

	// Create embedding provider files
	if err := g.createEmbeddingFiles(config); err != nil {
		return fmt.Errorf("failed to create embedding files: %w", err)
	}

	// Create server files if requested
	if err := g.createServerFiles(config); err != nil {
		return fmt.Errorf("failed to create server files: %w", err)
	}

	// Create runas directory README
	if err := g.createRunasReadme(); err != nil {
		return fmt.Errorf("failed to create runas README: %w", err)
	}

	// Create README
	if err := g.createReadme(); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	return nil
}

// GeneratorConfig holds configuration for the generator
type GeneratorConfig struct {
	Providers         []string
	Servers           []string
	DefaultProvider   string
	IncludeOllama     bool
	IncludeOpenAI     bool
	IncludeAnthropic  bool
	IncludeDeepSeek   bool
	IncludeGemini     bool
	IncludeOpenRouter bool
	IncludeLMStudio   bool
}

// createMainConfig creates the main config.yaml file at parent level
func (g *ModularConfigGenerator) createMainConfig(config *GeneratorConfig) error {
	// Config.yaml goes at parent level (next to executable)
	parentDir := filepath.Dir(g.baseDir)
	configDirName := filepath.Base(g.baseDir)
	
	mainConfig := MainConfigFile{
		Includes: &IncludeDirectives{
			Providers:  filepath.Join(configDirName, "providers/*.yaml"),
			Servers:    filepath.Join(configDirName, "servers/*.yaml"),
			Embeddings: filepath.Join(configDirName, "embeddings/*.yaml"),
			Templates:  filepath.Join(configDirName, "templates/*.yaml"),
			Settings:   filepath.Join(configDirName, "settings.yaml"),
		},
	}

	data, err := yaml.Marshal(mainConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal main config: %w", err)
	}

	path := filepath.Join(parentDir, "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write main config: %w", err)
	}

	return nil
}

// createSettings creates the settings.yaml file
func (g *ModularConfigGenerator) createSettings(config *GeneratorConfig) error {
	settings := map[string]interface{}{
		"ai": map[string]interface{}{
			"default_provider":      config.DefaultProvider,
			"default_system_prompt": "You are a helpful assistant.",
		},
		"embeddings": map[string]interface{}{
			"default_chunk_strategy": "sentence",
			"default_max_chunk_size": 512,
			"default_overlap":        0,
			"output_precision":       6,
		},
		"logging": map[string]interface{}{
			"level":  "info",
			"format": "text",
		},
		"chat": map[string]interface{}{
			"default_temperature": 0.7,
			"max_history_size":    50,
		},
	}

	data, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	path := filepath.Join(g.baseDir, "settings.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// createProviderFiles creates individual provider YAML files
func (g *ModularConfigGenerator) createProviderFiles(config *GeneratorConfig) error {
	providersDir := filepath.Join(g.baseDir, "providers")

	if config.IncludeOllama {
		if err := g.createOllamaProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeOpenAI {
		if err := g.createOpenAIProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeAnthropic {
		if err := g.createAnthropicProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeDeepSeek {
		if err := g.createDeepSeekProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeGemini {
		if err := g.createGeminiProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeOpenRouter {
		if err := g.createOpenRouterProvider(providersDir); err != nil {
			return err
		}
	}

	if config.IncludeLMStudio {
		if err := g.createLMStudioProvider(providersDir); err != nil {
			return err
		}
	}

	return nil
}

// createOllamaProvider creates ollama.yaml
func (g *ModularConfigGenerator) createOllamaProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "ollama",
		"config": map[string]interface{}{
			"api_endpoint":    "http://localhost:11434",
			"default_model":   "qwen2.5:32b",
			"timeout_seconds": 300,
			"max_retries":     5,
		},
	}

	return g.writeProviderFile(dir, "ollama.yaml", provider)
}

// createOpenAIProvider creates openai.yaml
func (g *ModularConfigGenerator) createOpenAIProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "openai",
		"config": map[string]interface{}{
			"api_key":         "${OPENAI_API_KEY}",
			"api_endpoint":    "https://api.openai.com/v1",
			"default_model":   "gpt-4o-mini",
			"timeout_seconds": 300,
			"max_retries":     2,
			"context_window":  128000,
			"reserve_tokens":  4000,
		},
	}

	return g.writeProviderFile(dir, "openai.yaml", provider)
}

// createAnthropicProvider creates anthropic.yaml
func (g *ModularConfigGenerator) createAnthropicProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "anthropic_native",
		"provider_name":  "anthropic",
		"config": map[string]interface{}{
			"api_key":         "${ANTHROPIC_API_KEY}",
			"default_model":   "claude-3-5-sonnet-20241022",
			"timeout_seconds": 300,
			"max_retries":     5,
		},
	}

	return g.writeProviderFile(dir, "anthropic.yaml", provider)
}

// createDeepSeekProvider creates deepseek.yaml
func (g *ModularConfigGenerator) createDeepSeekProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "deepseek",
		"config": map[string]interface{}{
			"api_key":         "${DEEPSEEK_API_KEY}",
			"api_endpoint":    "https://api.deepseek.com/v1",
			"default_model":   "deepseek-chat",
			"timeout_seconds": 300,
			"max_retries":     2,
			"context_window":  32000,
			"reserve_tokens":  2000,
		},
	}

	return g.writeProviderFile(dir, "deepseek.yaml", provider)
}

// createGeminiProvider creates gemini.yaml
func (g *ModularConfigGenerator) createGeminiProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",  // Gemini uses OpenAI-compatible API
		"provider_name":  "gemini",
		"config": map[string]interface{}{
			"api_key":         "${GEMINI_API_KEY}",
			"api_endpoint":    "https://generativelanguage.googleapis.com",
			"default_model":   "gemini-2.0-flash-exp",
			"timeout_seconds": 300,
			"max_retries":     2,
			"context_window":  1000000,
			"reserve_tokens":  8000,
		},
	}

	return g.writeProviderFile(dir, "gemini.yaml", provider)
}

// createOpenRouterProvider creates openrouter.yaml
func (g *ModularConfigGenerator) createOpenRouterProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "openrouter",
		"config": map[string]interface{}{
			"api_key":         "${OPENROUTER_API_KEY}",
			"api_endpoint":    "https://openrouter.ai/api/v1",
			"default_model":   "anthropic/claude-3.5-sonnet",
			"timeout_seconds": 300,
			"max_retries":     2,
			"context_window":  200000,
			"reserve_tokens":  2000,
		},
	}

	return g.writeProviderFile(dir, "openrouter.yaml", provider)
}

// createLMStudioProvider creates lmstudio.yaml
func (g *ModularConfigGenerator) createLMStudioProvider(dir string) error {
	provider := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "lmstudio",
		"config": map[string]interface{}{
			"api_endpoint":    "http://localhost:1234/v1",
			"default_model":   "local-model",
			"timeout_seconds": 300,
			"max_retries":     2,
		},
	}

	return g.writeProviderFile(dir, "lmstudio.yaml", provider)
}

// writeProviderFile writes a provider YAML file with proper field ordering
func (g *ModularConfigGenerator) writeProviderFile(dir, filename string, data interface{}) error {
	// Convert to ordered YAML manually to ensure interface_type and provider_name come first
	var yamlContent strings.Builder
	
	// Extract the map
	providerMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid provider data format")
	}
	
	// Write fields in specific order for readability
	yamlContent.WriteString(fmt.Sprintf("interface_type: %s\n", providerMap["interface_type"]))
	yamlContent.WriteString(fmt.Sprintf("provider_name: %s\n", providerMap["provider_name"]))
	yamlContent.WriteString("config:\n")
	
	// Marshal just the config section
	configMap := providerMap["config"].(map[string]interface{})
	configYAML, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal provider config: %w", err)
	}
	
	// Indent the config section
	lines := strings.Split(string(configYAML), "\n")
	for _, line := range lines {
		if line != "" {
			yamlContent.WriteString("  " + line + "\n")
		}
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(yamlContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}

	return nil
}

// createServerFiles creates example server configuration files
func (g *ModularConfigGenerator) createServerFiles(config *GeneratorConfig) error {
	// Only create example if requested
	if len(config.Servers) == 0 {
		// Create example README
		serversDir := filepath.Join(g.baseDir, "servers")
		readmePath := filepath.Join(serversDir, "README.md")
		readme := `# MCP Servers Configuration

Place your MCP server configuration files here.

## Example Server Configuration

Create a file like ` + "`filesystem.yaml`" + `:

` + "```yaml" + `
server_name: filesystem
config:
  command: /path/to/filesystem-server
  args: []
  env: {}
` + "```" + `

Each server gets its own YAML file for easy management.
`
		return os.WriteFile(readmePath, []byte(readme), 0644)
	}

	return nil
}

// createEmbeddingFiles creates embedding provider configuration files
func (g *ModularConfigGenerator) createEmbeddingFiles(config *GeneratorConfig) error {
	embeddingsDir := filepath.Join(g.baseDir, "embeddings")

	// Create OpenAI embeddings if OpenAI is enabled
	if config.IncludeOpenAI {
		if err := g.createOpenAIEmbedding(embeddingsDir); err != nil {
			return err
		}
	}

	// Create OpenRouter embeddings if OpenRouter is enabled
	if config.IncludeOpenRouter {
		if err := g.createOpenRouterEmbedding(embeddingsDir); err != nil {
			return err
		}
	}

	// Create Ollama embeddings if Ollama is enabled
	if config.IncludeOllama {
		if err := g.createOllamaEmbedding(embeddingsDir); err != nil {
			return err
		}
	}

	// Create README if no embeddings providers
	if !config.IncludeOpenAI && !config.IncludeOpenRouter && !config.IncludeOllama {
		readmePath := filepath.Join(embeddingsDir, "README.md")
		readme := `# Embeddings Configuration

Place embedding provider configuration files here.

Embedding providers generate vector representations of text for semantic search,
RAG applications, and similarity matching.

## Example Configuration

**openai.yaml:**
` + "```yaml" + `
interface_type: openai_compatible
provider_name: openai
config:
  api_key: ${OPENAI_API_KEY}
  api_endpoint: https://api.openai.com/v1
  default_model: text-embedding-3-small
  embedding_models:
    text-embedding-3-small:
      max_tokens: 8191
      dimensions: 1536
      default: true
` + "```" + `
`
		return os.WriteFile(readmePath, []byte(readme), 0644)
	}

	return nil
}

// createOpenAIEmbedding creates OpenAI embedding configuration
func (g *ModularConfigGenerator) createOpenAIEmbedding(dir string) error {
	embedding := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "openai",
		"config": map[string]interface{}{
			"api_key":                 "${OPENAI_API_KEY}",
			"api_endpoint":            "https://api.openai.com/v1",
			"default_model":           "text-embedding-3-small",
			"default_embedding_model": "text-embedding-3-small",
			"embedding_models": map[string]interface{}{
				"text-embedding-3-small": map[string]interface{}{
					"max_tokens":        8191,
					"dimensions":        1536,
					"cost_per_1k_tokens": 0.00002,
					"default":           true,
					"description":       "Most capable embedding model for both english and non-english tasks",
				},
				"text-embedding-3-large": map[string]interface{}{
					"max_tokens":         8191,
					"dimensions":         3072,
					"cost_per_1k_tokens": 0.00013,
					"description":        "Larger embedding model with higher performance",
				},
				"text-embedding-ada-002": map[string]interface{}{
					"max_tokens":         8191,
					"dimensions":         1536,
					"cost_per_1k_tokens": 0.0001,
					"description":        "Previous generation embedding model",
				},
			},
		},
	}

	return g.writeEmbeddingFile(dir, "openai.yaml", embedding)
}

// createOpenRouterEmbedding creates OpenRouter embedding configuration
func (g *ModularConfigGenerator) createOpenRouterEmbedding(dir string) error {
	embedding := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "openrouter",
		"config": map[string]interface{}{
			"api_key":                 "${OPENROUTER_API_KEY}",
			"api_endpoint":            "https://openrouter.ai/api/v1",
			"default_model":           "text-embedding-3-small",
			"default_embedding_model": "text-embedding-3-small",
			"embedding_models": map[string]interface{}{
				"text-embedding-3-small": map[string]interface{}{
					"max_tokens":  8191,
					"dimensions":  1536,
					"default":     true,
					"description": "OpenAI embedding model via OpenRouter",
				},
				"text-embedding-3-large": map[string]interface{}{
					"max_tokens":  8191,
					"dimensions":  3072,
					"description": "Larger OpenAI embedding model via OpenRouter",
				},
			},
		},
	}

	return g.writeEmbeddingFile(dir, "openrouter.yaml", embedding)
}

// createOllamaEmbedding creates Ollama embedding configuration
func (g *ModularConfigGenerator) createOllamaEmbedding(dir string) error {
	embedding := map[string]interface{}{
		"interface_type": "openai_compatible",
		"provider_name":  "ollama",
		"config": map[string]interface{}{
			"api_endpoint":            "http://localhost:11434",
			"default_model":           "nomic-embed-text",
			"default_embedding_model": "nomic-embed-text",
			"embedding_models": map[string]interface{}{
				"nomic-embed-text": map[string]interface{}{
					"max_tokens":  8192,
					"dimensions":  768,
					"default":     true,
					"description": "High-performance open embedding model",
				},
				"mxbai-embed-large": map[string]interface{}{
					"max_tokens":  512,
					"dimensions":  1024,
					"description": "Large multilingual embedding model",
				},
			},
		},
	}

	return g.writeEmbeddingFile(dir, "ollama.yaml", embedding)
}

// writeEmbeddingFile writes an embedding provider YAML file with proper field ordering
func (g *ModularConfigGenerator) writeEmbeddingFile(dir, filename string, data interface{}) error {
	// Convert to ordered YAML manually to ensure interface_type and provider_name come first
	var yamlContent strings.Builder
	
	// Extract the map
	embeddingMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid embedding data format")
	}
	
	// Write fields in specific order for readability
	yamlContent.WriteString(fmt.Sprintf("interface_type: %s\n", embeddingMap["interface_type"]))
	yamlContent.WriteString(fmt.Sprintf("provider_name: %s\n", embeddingMap["provider_name"]))
	yamlContent.WriteString("config:\n")
	
	// Marshal just the config section
	configMap := embeddingMap["config"].(map[string]interface{})
	configYAML, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding config: %w", err)
	}
	
	// Indent the config section
	lines := strings.Split(string(configYAML), "\n")
	for _, line := range lines {
		if line != "" {
			yamlContent.WriteString("  " + line + "\n")
		}
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(yamlContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write embedding file: %w", err)
	}

	return nil
}

// createRunasReadme creates a README for the runas directory
func (g *ModularConfigGenerator) createRunasReadme() error {
	runasDir := filepath.Join(g.baseDir, "runas")
	readmePath := filepath.Join(runasDir, "README.md")
	
	readme := `# MCP Server Mode Configurations

This directory contains configurations for running mcp-cli as an MCP server.

When you run ` + "`mcp-cli serve config.yaml`" + `, it exposes workflows as tools
that can be used by Claude Desktop, IDEs, or other MCP clients.

## Example Server Configuration

Create a file like ` + "`research_agent.yaml`" + `:

` + "```yaml" + `
server_info:
  name: research-agent
  version: 1.0.0
  description: AI research assistant with web search

tools:
  - name: research_topic
    description: Research a topic comprehensively
    template: research_workflow
    input_schema:
      type: object
      properties:
        topic:
          type: string
          description: Topic to research
      required: [topic]
    input_mapping:
      topic: "{{input_data}}"
` + "```" + `

## Running as MCP Server

` + "```bash" + `
# Start server
mcp-cli serve config/runas/research_agent.yaml

# With verbose logging
mcp-cli serve --verbose config/runas/research_agent.yaml
` + "```" + `

## Configure Claude Desktop

Add to ` + "`claude_desktop_config.json`" + `:

` + "```json" + `
{
  "mcpServers": {
    "research-agent": {
      "command": "/usr/local/bin/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/research_agent.yaml"]
    }
  }
}
` + "```" + `

## Benefits

- **Expose workflows as tools** for AI assistants
- **Reuse templates** without rewriting code
- **Multi-provider workflows** available to Claude
- **Template composition** accessible via MCP protocol

Each YAML file in this directory defines a separate MCP server configuration.
`
	
	return os.WriteFile(readmePath, []byte(readme), 0644)
}

// createReadme creates a README.md file explaining the structure
func (g *ModularConfigGenerator) createReadme() error {
	readme := `# MCP CLI Modular Configuration

This directory contains your modular MCP CLI configuration files.

## Structure

` + "```" + `
mcp-cli                  # Executable
.env                     # API keys (gitignored)
config.yaml              # Main config with includes
config/                  # Modular config directory
├── README.md            # This file
├── settings.yaml        # Global settings
├── providers/           # LLM provider configs
│   ├── ollama.yaml
│   ├── openai.yaml
│   └── anthropic.yaml
├── embeddings/          # Embedding provider configs
│   ├── openai.yaml
│   ├── openrouter.yaml
│   └── ollama.yaml
├── servers/             # MCP server configs
│   ├── README.md
│   └── *.yaml
├── templates/           # Workflow templates
│   ├── README.md
│   └── *.yaml
└── runas/               # MCP server mode configs
    ├── README.md
    └── *.yaml
` + "```" + `

## Main Config (config.yaml)

The main config file uses includes to load all modular configurations:

` + "```yaml" + `
includes:
  providers: config/providers/*.yaml
  embeddings: config/embeddings/*.yaml
  servers: config/servers/*.yaml
  templates: config/templates/*.yaml
  settings: config/settings.yaml
` + "```" + `

All settings are in ` + "`config/settings.yaml`" + ` - the main config just declares includes.

## Settings File (config/settings.yaml)

Global settings are in ` + "`config/settings.yaml`" + `:

` + "```yaml" + `
ai:
  default_provider: ollama
  default_system_prompt: You are a helpful assistant.

embeddings:
  default_chunk_strategy: sentence
  default_max_chunk_size: 512
  output_precision: 6

logging:
  level: info
  format: text

chat:
  default_temperature: 0.7
  max_history_size: 50
` + "```" + `

## Provider Files (LLM)

Each LLM provider gets its own file in ` + "`providers/`" + `:

**providers/ollama.yaml:**
` + "```yaml" + `
interface_type: openai_compatible
provider_name: ollama
config:
  api_endpoint: http://localhost:11434
  default_model: qwen2.5:32b
  timeout_seconds: 300
` + "```" + `

## Embedding Files

Embedding providers are separate from LLM providers in ` + "`embeddings/`" + `:

**embeddings/openai.yaml:**
` + "```yaml" + `
interface_type: openai_compatible
provider_name: openai
config:
  api_key: ${OPENAI_API_KEY}
  default_embedding_model: text-embedding-3-small
  embedding_models:
    text-embedding-3-small:
      max_tokens: 8191
      dimensions: 1536
      default: true
` + "```" + `

## Server Files

MCP servers are configured in ` + "`servers/`" + `:

**servers/filesystem.yaml:**
` + "```yaml" + `
server_name: filesystem
config:
  command: /path/to/filesystem-server
  args: []
` + "```" + `

## Templates

Workflow templates go in ` + "`templates/`" + `:

**templates/analyze.yaml:**
` + "```yaml" + `
name: analyze
description: Analyze input data
steps:
  - name: step1
    prompt: "Analyze this: {{stdin}}"
    output: analysis
  - name: step2
    prompt: "Summarize: {{analysis}}"
` + "```" + `

## MCP Server Mode (runas/)

Server mode configurations in ` + "`runas/`" + ` expose workflows as MCP tools:

**runas/research_agent.yaml:**
` + "```yaml" + `
server_info:
  name: research-agent
  version: 1.0.0

tools:
  - name: research_topic
    template: research_workflow
    input_schema:
      type: object
      properties:
        topic: {type: string}
` + "```" + `

Run with: ` + "`mcp-cli serve config/runas/research_agent.yaml`" + `

## Environment Variables

API keys should be in ` + "`.env`" + ` (next to executable):

` + "```bash" + `
OPENAI_API_KEY=your-key-here
ANTHROPIC_API_KEY=your-key-here
DEEPSEEK_API_KEY=your-key-here
GEMINI_API_KEY=your-key-here
OPENROUTER_API_KEY=your-key-here
` + "```" + `

## Truly Modular Design

**config.yaml**: Just includes, no settings
**config/settings.yaml**: All global settings (AI, embeddings, logging, chat)
**config/providers/**: Individual LLM provider configs
**config/embeddings/**: Individual embedding provider configs
**config/servers/**: Individual MCP server configs
**config/templates/**: Reusable workflow templates
**config/runas/**: MCP server mode configurations

Each file is self-contained and can be edited independently.
Version control friendly - track changes to individual components.
`

	path := filepath.Join(g.baseDir, "README.md")
	return os.WriteFile(path, []byte(readme), 0644)
}
