package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// InitCmd initializes a new mcp-cli configuration
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize mcp-cli configuration",
	Long: `Interactive setup wizard for mcp-cli configuration.

Creates a modular configuration structure with separate directories for
providers, embeddings, servers, and workflows.

Modes:
  --quick     Quick setup with minimal questions (uses ollama, no API keys)
  --full      Full setup with all configuration options
  (default)   Standard interactive setup

Examples:
  mcp-cli init              # Interactive setup
  mcp-cli init --quick      # Quick setup (ollama only)
  mcp-cli init --full       # Complete setup wizard`,
	RunE: runInit,
}

var (
	quickMode bool
	fullMode  bool
)

func init() {
	InitCmd.Flags().BoolVar(&quickMode, "quick", false, "Quick setup with defaults")
	InitCmd.Flags().BoolVar(&fullMode, "full", false, "Full setup with all options")
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	
	// Get executable directory
	execDir := getExecutableDir()
	
	// Welcome message
	printWelcome()
	
	// Check if config already exists
	configPath := filepath.Join(execDir, "config.yaml")
	
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("⚠️  Configuration file already exists at:", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		response, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			fmt.Println("Setup cancelled.")
			return nil
		}
		
		// Backup existing config
		backupPath := configPath + ".backup"
		if err := copyFile(configPath, backupPath); err != nil {
			fmt.Printf("⚠️  Warning: Could not create backup: %v\n", err)
		} else {
			fmt.Println("✓ Created backup at:", backupPath)
		}
	}
	
	var cfg *InitConfig
	
	// Check for "all services" mode first (unless using --quick or --full flags)
	if !quickMode && !fullMode {
		fmt.Println()
		fmt.Println("📦 Setup Options:")
		fmt.Println()
		fmt.Println("You can create a complete configuration with all services enabled,")
		fmt.Println("or proceed with interactive setup to customize your installation.")
		fmt.Println()
		
		if askYesNo(reader, "Create default config for all services (DeepSeek, RAG, Skills, all providers)", false) {
			cfg = createAllServicesConfig()
			return createModularConfig(execDir, cfg)
		}
	}
	
	if quickMode {
		cfg = createQuickConfig()
	} else if fullMode {
		cfg = createFullConfig(reader)
	} else {
		cfg = createStandardConfig(reader)
	}
	
	// Always create modular config
	return createModularConfig(execDir, cfg)
}

// InitConfig holds configuration choices
type InitConfig struct {
	Providers         []string
	Servers           []string
	IncludeOllama     bool
	IncludeOpenAI     bool
	IncludeAnthropic  bool
	IncludeDeepSeek   bool
	IncludeGemini     bool
	IncludeOpenRouter bool
	IncludeLMStudio   bool
	IncludeMoonshot   bool
	IncludeBedrock    bool
	IncludeAzureFoundry bool
	IncludeVertexAI   bool
	DefaultProvider   string
	IncludeSkills     bool
	IncludeRAG        bool
}

func printWelcome() {
	color.New(color.FgCyan, color.Bold).Println("\n🚀 MCP CLI Setup Wizard")
	fmt.Println("This will help you set up mcp-cli for the first time.")
}

func createQuickConfig() *InitConfig {
	fmt.Println("📦 Quick Setup Mode")
	fmt.Println("   Creating configuration with DeepSeek (requires API key)")
	fmt.Println("   Includes: RAG support, Skills system")
	fmt.Println()
	
	return &InitConfig{
		Providers:       []string{"deepseek"},
		Servers:         []string{},
		IncludeDeepSeek: true,
		DefaultProvider: "deepseek",
		IncludeSkills:   true,
		IncludeRAG:      true,
	}
}

func createAllServicesConfig() *InitConfig {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📦 All Services Configuration")
	fmt.Println("   Creating complete configuration with all providers and services")
	fmt.Println("   Default provider: DeepSeek")
	fmt.Println()
	
	return &InitConfig{
		Providers: []string{
			"ollama",
			"openai",
			"anthropic",
			"deepseek",
			"gemini",
			"openrouter",
			"lmstudio",
			"kimik2",
			"bedrock",
			"azure-foundry",
			"vertex-ai",
		},
		Servers:              []string{},
		IncludeOllama:        true,
		IncludeOpenAI:        true,
		IncludeAnthropic:     true,
		IncludeDeepSeek:      true,
		IncludeGemini:        true,
		IncludeOpenRouter:    true,
		IncludeLMStudio:      true,
		IncludeMoonshot:      true,
		IncludeBedrock:       true,
		IncludeAzureFoundry:  true,
		IncludeVertexAI:      true,
		DefaultProvider:      "deepseek",
		IncludeSkills:        true,
		IncludeRAG:           true,
	}
}

func createStandardConfig(reader *bufio.Reader) *InitConfig {
	config := &InitConfig{
		Servers: []string{}, // Empty - no assumptions about MCP servers
	}
	
	fmt.Println("📋 Configuration Setup")
	fmt.Println()
	
	// Ask about providers
	fmt.Println("Which AI providers would you like to configure?")
	fmt.Println()
	fmt.Println("Available providers:")
	fmt.Println("  • Ollama        - Local AI (no API key needed)")
	fmt.Println("  • OpenAI        - GPT-4, GPT-4o (requires API key)")
	fmt.Println("  • Anthropic     - Claude 3.5 (requires API key)")
	fmt.Println("  • AWS Bedrock   - Claude, Titan models (requires AWS credentials)")
	fmt.Println("  • Azure Foundry - GPT, embeddings on Azure (requires Azure credentials)")
	fmt.Println("  • GCP Vertex AI - Gemini on GCP (requires GCP credentials)")
	fmt.Println("  • DeepSeek      - DeepSeek Chat (requires API key)")
	fmt.Println("  • Gemini        - Google Gemini (requires API key)")
	fmt.Println("  • OpenRouter    - Access many models (requires API key)")
	fmt.Println("  • LM Studio     - Local model server (no API key)")
	fmt.Println("  • Moonshot      - Kimi K2 models (requires API key)")
	fmt.Println()
	fmt.Println("(You can add more providers later by editing config files)")
	fmt.Println()
	
	// Ollama (local, no API key)
	if askYesNo(reader, "Use Ollama (local, no API key needed)", true) {
		config.IncludeOllama = true
		config.Providers = append(config.Providers, "ollama")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "ollama"
		}
	}
	
	// OpenAI
	if askYesNo(reader, "Use OpenAI (requires API key)", false) {
		config.IncludeOpenAI = true
		config.Providers = append(config.Providers, "openai")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "openai"
		}
	}
	
	// Anthropic
	if askYesNo(reader, "Use Anthropic Claude (requires API key)", false) {
		config.IncludeAnthropic = true
		config.Providers = append(config.Providers, "anthropic")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "anthropic"
		}
	}
	
	// AWS Bedrock
	if askYesNo(reader, "Use AWS Bedrock (requires AWS credentials)", false) {
		config.IncludeBedrock = true
		config.Providers = append(config.Providers, "bedrock")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "bedrock"
		}
	}
	
	// Azure Foundry
	if askYesNo(reader, "Use Azure AI Foundry (requires Azure credentials)", false) {
		config.IncludeAzureFoundry = true
		config.Providers = append(config.Providers, "azure-foundry")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "azure-foundry"
		}
	}
	
	// GCP Vertex AI
	if askYesNo(reader, "Use GCP Vertex AI (requires GCP credentials)", false) {
		config.IncludeVertexAI = true
		config.Providers = append(config.Providers, "vertex-ai")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "vertex-ai"
		}
	}
	
	// DeepSeek
	if askYesNo(reader, "Use DeepSeek (requires API key)", false) {
		config.IncludeDeepSeek = true
		config.Providers = append(config.Providers, "deepseek")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "deepseek"
		}
	}
	
	// Gemini
	if askYesNo(reader, "Use Gemini (requires API key)", false) {
		config.IncludeGemini = true
		config.Providers = append(config.Providers, "gemini")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "gemini"
		}
	}
	
	// OpenRouter
	if askYesNo(reader, "Use OpenRouter (requires API key)", false) {
		config.IncludeOpenRouter = true
		config.Providers = append(config.Providers, "openrouter")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "openrouter"
		}
	}
	
	// LM Studio
	if askYesNo(reader, "Use LM Studio (local server, no API key)", false) {
		config.IncludeLMStudio = true
		config.Providers = append(config.Providers, "lmstudio")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "lmstudio"
		}
	}
	
	// Moonshot
	if askYesNo(reader, "Use Moonshot Kimi (requires API key)", false) {
		config.IncludeMoonshot = true
		config.Providers = append(config.Providers, "kimik2")
		if config.DefaultProvider == "" {
			config.DefaultProvider = "kimik2"
		}
	}
	
	// Default to ollama if no providers selected
	if len(config.Providers) == 0 {
		fmt.Println("\n💡 No providers selected. Defaulting to Ollama (local)")
		config.IncludeOllama = true
		config.Providers = append(config.Providers, "ollama")
		config.DefaultProvider = "ollama"
	}
	
	// Ask about default provider if multiple
	if len(config.Providers) > 1 {
		fmt.Println("\nMultiple providers configured.")
		fmt.Println("Available: " + strings.Join(config.Providers, ", "))
		fmt.Printf("Which provider should be the default? [%s]: ", config.DefaultProvider)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response)) // Normalize to lowercase
		if response != "" {
			config.DefaultProvider = response
		}
	}
	
	// Ask about skills
	fmt.Println()
	fmt.Println("🎯 Anthropic Skills System:")
	fmt.Println("Skills provide helper libraries for document creation, data processing, etc.")
	fmt.Println("These enable dynamic code execution with helper library imports.")
	fmt.Println()
	
	if askYesNo(reader, "Set up skills directory (download skills from GitHub)", true) {
		config.IncludeSkills = true
		fmt.Println("   ✓ Will create skills directory with README")
		fmt.Println("   💡 Download skills from: https://github.com/anthropics/skills")
	}
	
	// Ask about RAG
	fmt.Println()
	fmt.Println("🔍 RAG (Retrieval-Augmented Generation):")
	fmt.Println("RAG allows workflows to search vector databases and retrieve context.")
	fmt.Println("This is useful for document search, knowledge bases, and contextual responses.")
	fmt.Println()
	
	if askYesNo(reader, "Set up RAG configuration directory", false) {
		config.IncludeRAG = true
		fmt.Println("   ✓ Will create config/rag/ directory")
	}
	
	fmt.Println()
	return config
}

func createFullConfig(reader *bufio.Reader) *InitConfig {
	config := createStandardConfig(reader)
	
	// Ask about MCP servers
	fmt.Println("\nMCP Servers:")
	fmt.Println("MCP servers provide additional capabilities like file access, web search,")
	fmt.Println("database connections, etc. You'll need to install and configure them separately.")
	fmt.Println()
	
	if askYesNo(reader, "Add placeholder for MCP servers in config", false) {
		fmt.Println()
		fmt.Println("💡 To add MCP servers:")
		fmt.Println("   1. Install your MCP server binary")
		fmt.Println("   2. Create a YAML file in config/servers/")
		fmt.Println("   3. Example config/servers/myserver.yaml:")
		fmt.Println()
		fmt.Println(`      server_name: myserver`)
		fmt.Println(`      config:`)
		fmt.Println(`        command: /path/to/myserver-binary`)
		fmt.Println()
	}
	
	return config
}

func askYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	prompt := "[y/N]"
	if defaultYes {
		prompt = "[Y/n]"
	}
	
	fmt.Printf("%s %s: ", question, prompt)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response == "" {
		return defaultYes
	}
	
	return strings.HasPrefix(response, "y")
}

func createEnvFile(path string, config *InitConfig) error {
	var content strings.Builder
	
	content.WriteString("# MCP CLI Environment Variables\n")
	content.WriteString("# Generated by mcp-cli init\n\n")
	
	if config.IncludeOpenAI {
		content.WriteString("# OpenAI API Key\n")
		content.WriteString("# Get from: https://platform.openai.com/api-keys\n")
		content.WriteString("OPENAI_API_KEY=\n\n")
	}
	
	if config.IncludeAnthropic {
		content.WriteString("# Anthropic API Key\n")
		content.WriteString("# Get from: https://console.anthropic.com/\n")
		content.WriteString("ANTHROPIC_API_KEY=\n\n")
	}
	
	if config.IncludeDeepSeek {
		content.WriteString("# DeepSeek API Key\n")
		content.WriteString("# Get from: https://platform.deepseek.com/\n")
		content.WriteString("DEEPSEEK_API_KEY=\n\n")
	}
	
	if config.IncludeGemini {
		content.WriteString("# Gemini API Key\n")
		content.WriteString("# Get from: https://makersuite.google.com/app/apikey\n")
		content.WriteString("GEMINI_API_KEY=\n\n")
	}
	
	if config.IncludeOpenRouter {
		content.WriteString("# OpenRouter API Key\n")
		content.WriteString("# Get from: https://openrouter.ai/keys\n")
		content.WriteString("OPENROUTER_API_KEY=\n\n")
	}
	
	if config.IncludeBedrock {
		content.WriteString("# AWS Bedrock Credentials\n")
		content.WriteString("# Get from: AWS IAM Console\n")
		content.WriteString("AWS_ACCESS_KEY_ID=\n")
		content.WriteString("AWS_SECRET_ACCESS_KEY=\n")
		content.WriteString("AWS_REGION=us-east-1\n")
		content.WriteString("# AWS_SESSION_TOKEN=  # Optional, for temporary credentials\n\n")
	}
	
	if config.IncludeAzureFoundry {
		content.WriteString("# Azure AI Foundry Credentials\n")
		content.WriteString("# Get from: Azure Portal > AI Foundry Resource > Keys and Endpoint\n")
		content.WriteString("AZURE_FOUNDRY_API_KEY=\n")
		content.WriteString("# AZURE_FOUNDRY_ENDPOINT=https://your-resource.openai.azure.com/openai/v1/\n\n")
	}
	
	if config.IncludeVertexAI {
		content.WriteString("# GCP Vertex AI Credentials\n")
		content.WriteString("# Setup:\n")
		content.WriteString("#   1. Create GCP project: https://console.cloud.google.com/\n")
		content.WriteString("#   2. Enable Vertex AI API: https://console.cloud.google.com/apis/library/aiplatform.googleapis.com\n")
		content.WriteString("#   3. Create service account with 'Vertex AI User' role\n")
		content.WriteString("#   4. Download service account JSON key\n")
		content.WriteString("GCP_PROJECT_ID=\n")
		content.WriteString("GCP_LOCATION=us-central1\n")
		content.WriteString("GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json\n\n")
	}
	
	// Only create .env if there are API keys to configure
	if config.IncludeOpenAI || config.IncludeAnthropic || config.IncludeDeepSeek || 
	   config.IncludeGemini || config.IncludeOpenRouter || config.IncludeBedrock ||
	   config.IncludeAzureFoundry || config.IncludeVertexAI {
		return os.WriteFile(path, []byte(content.String()), 0644)
	}
	
	return nil
}

func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// createModularConfig creates a modular config directory structure
func createModularConfig(baseDir string, initCfg *InitConfig) error {
	// Create config directory next to executable
	configDir := filepath.Join(baseDir, "config")
	
	// Ask user if they want a different location
	fmt.Println()
	fmt.Printf("📁 Config directory will be created at: %s\n", configDir)
	fmt.Print("Use this location? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	if strings.ToLower(response) == "n" {
		fmt.Print("Enter custom path: ")
		customPath, _ := reader.ReadString('\n')
		customPath = strings.TrimSpace(customPath)
		if customPath != "" {
			configDir = customPath
			// Expand ~ to home directory
			if strings.HasPrefix(configDir, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					configDir = filepath.Join(home, configDir[2:])
				}
			}
		}
	}
	
	// Check if directory exists
	if _, err := os.Stat(configDir); err == nil {
		fmt.Printf("⚠️  Directory already exists: %s\n", configDir)
		fmt.Print("Overwrite? [y/N]: ")
		response, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}
	
	// Create generator config
	genConfig := &config.GeneratorConfig{
		Providers:          initCfg.Providers,
		Servers:            initCfg.Servers,
		DefaultProvider:    initCfg.DefaultProvider,
		IncludeOllama:      initCfg.IncludeOllama,
		IncludeOpenAI:      initCfg.IncludeOpenAI,
		IncludeAnthropic:   initCfg.IncludeAnthropic,
		IncludeDeepSeek:    initCfg.IncludeDeepSeek,
		IncludeGemini:      initCfg.IncludeGemini,
		IncludeOpenRouter:  initCfg.IncludeOpenRouter,
		IncludeLMStudio:    initCfg.IncludeLMStudio,
		IncludeMoonshot:    initCfg.IncludeMoonshot,
		IncludeBedrock:     initCfg.IncludeBedrock,
		IncludeAzureFoundry: initCfg.IncludeAzureFoundry,
		IncludeVertexAI:    initCfg.IncludeVertexAI,
	}
	
	// Generate modular config
	generator := config.NewModularConfigGenerator(configDir)
	if err := generator.Generate(genConfig); err != nil {
		return fmt.Errorf("failed to generate modular config: %w", err)
	}
	
	// Create RAG directory if requested
	if initCfg.IncludeRAG {
		if err := createRAGDirectory(configDir); err != nil {
			return fmt.Errorf("failed to create RAG directory: %w", err)
		}
	}
	
	// Create skills if requested
	if initCfg.IncludeSkills {
		if err := createSkillsDirectory(configDir, initCfg); err != nil {
			return fmt.Errorf("failed to create skills: %w", err)
		}
		
		// Create skills-auto.yaml runas config
		if err := createSkillsRunAsConfig(configDir); err != nil {
			return fmt.Errorf("failed to create skills runas config: %w", err)
		}
	}
	
	// Create .env file at executable level (parent directory)
	parentDir := filepath.Dir(configDir)
	if initCfg.IncludeOpenAI || initCfg.IncludeAnthropic || initCfg.IncludeDeepSeek ||
		initCfg.IncludeGemini || initCfg.IncludeOpenRouter || initCfg.IncludeBedrock ||
		initCfg.IncludeAzureFoundry || initCfg.IncludeVertexAI {
		envPath := filepath.Join(parentDir, ".env")
		if err := createEnvFile(envPath, initCfg); err != nil {
			return fmt.Errorf("failed to create .env file: %w", err)
		}
	}
	
	// Print success message
	printModularSuccess(configDir, initCfg)
	
	return nil
}

// printModularSuccess prints success message for modular config
func printModularSuccess(configDir string, cfg *InitConfig) {
	success := color.New(color.FgGreen, color.Bold)
	info := color.New(color.FgCyan)
	
	parentDir := filepath.Dir(configDir)
	
	fmt.Println()
	success.Println("✅ Modular Configuration Created!")
	fmt.Println()
	
	info.Println("📁 Created structure:")
	fmt.Printf("   %s/\n", parentDir)
	fmt.Println("   ├── config.yaml          # Main config")
	fmt.Printf("   └── %s/\n", filepath.Base(configDir))
	fmt.Println("       ├── README.md")
	fmt.Println("       ├── providers/        # LLM configs")
	for _, provider := range cfg.Providers {
		fmt.Printf("       │   ├── %s.yaml\n", provider)
	}
	fmt.Println("       ├── embeddings/       # Embedding configs")
	if cfg.IncludeOpenAI || cfg.IncludeOpenRouter || cfg.IncludeOllama || cfg.IncludeDeepSeek {
		for _, provider := range cfg.Providers {
			if provider == "openai" || provider == "openrouter" || provider == "ollama" || provider == "deepseek" {
				fmt.Printf("       │   ├── %s.yaml\n", provider)
			}
		}
	}
	if cfg.IncludeRAG {
		fmt.Println("       ├── rag/              # RAG configurations")
		fmt.Println("       │   ├── README.md")
		fmt.Println("       │   └── expansion/")
	}
	if cfg.IncludeSkills {
		fmt.Println("       ├── skills/           # Anthropic skills")
		fmt.Println("       │   └── README.md     # Download from github.com/anthropics/skills")
		fmt.Println("       ├── runasMCP/")
		fmt.Println("       │   └── skills-auto.yaml")
	}
	fmt.Println("       ├── servers/")
	fmt.Println("       │   └── README.md")
	fmt.Println("       └── workflows/")
	fmt.Println()
	
	if cfg.IncludeOpenAI || cfg.IncludeAnthropic || cfg.IncludeDeepSeek ||
		cfg.IncludeGemini || cfg.IncludeOpenRouter || cfg.IncludeBedrock ||
		cfg.IncludeAzureFoundry || cfg.IncludeVertexAI {
		color.New(color.FgYellow).Println("⚠️  Important: Add your API keys")
		fmt.Printf("   Edit: %s/.env\n", parentDir)
		fmt.Println()
	}
	
	info.Println("🎯 Next steps:")
	fmt.Printf("   1. Review: %s/README.md\n", configDir)
	if cfg.IncludeOpenAI || cfg.IncludeAnthropic || cfg.IncludeDeepSeek ||
		cfg.IncludeGemini || cfg.IncludeOpenRouter || cfg.IncludeBedrock ||
		cfg.IncludeAzureFoundry || cfg.IncludeVertexAI {
		fmt.Printf("   2. Edit .env: %s/.env\n", parentDir)
		if cfg.IncludeSkills {
			fmt.Printf("   3. Start skills server: ./mcp-cli serve %s/runasMCP/skills-auto.yaml\n", filepath.Base(configDir))
			fmt.Printf("   4. Run query: ./mcp-cli query \"hello\"\n")
		} else {
			fmt.Printf("   3. Run: ./mcp-cli query \"hello\"\n")
		}
	} else {
		if cfg.IncludeSkills {
			fmt.Printf("   2. Start skills server: ./mcp-cli serve %s/runasMCP/skills-auto.yaml\n", filepath.Base(configDir))
			fmt.Printf("   3. Run query: ./mcp-cli query \"hello\"\n")
		} else {
			fmt.Printf("   2. Run: ./mcp-cli query \"hello\"\n")
		}
	}
	
	fmt.Println()
	info.Println("💡 Tips:")
	fmt.Println("   • LLM and embedding providers are separated for clarity")
	fmt.Println("   • config.yaml is at executable level for easy discovery")
	fmt.Println("   • Each provider type has its own subdirectory")
	fmt.Println("   • Add MCP servers in servers/ directory")
	if cfg.IncludeSkills {
		fmt.Println("   • Use skills as MCP server: ./mcp-cli serve config/runasMCP/skills-auto.yaml")
		fmt.Println("   • Skills support dynamic code execution with helper libraries")
	}
	fmt.Println()
}

// createRAGDirectory creates the RAG configuration directory
func createRAGDirectory(configDir string) error {
	ragDir := filepath.Join(configDir, "rag")
	expansionDir := filepath.Join(ragDir, "expansion")
	
	color.New(color.FgCyan).Println("🔍 Creating RAG Configuration...")
	
	// Create directories
	if err := os.MkdirAll(expansionDir, 0755); err != nil {
		return fmt.Errorf("failed to create RAG directories: %w", err)
	}
	
	// Create README
	readmeContent := `# RAG Configuration

This directory contains RAG (Retrieval-Augmented Generation) server configurations.

## What is RAG?

RAG enhances LLM responses by retrieving relevant context from vector databases
before generating answers. This is useful for:

- Answering questions from your documents
- Searching knowledge bases
- Contextual code completion
- Document analysis

## Directory Structure

- **rag/*.yaml** - RAG server configurations
- **rag/expansion/*.yaml** - Query expansion strategies (optional)

## Example: pgvector.yaml

Create a file at ` + "`config/rag/pgvector.yaml`" + `:

` + "```yaml" + `
server_name: pgvector
rag_type: pgvector

connection:
  host: localhost
  port: 5432
  database: vector_db
  user: postgres
  password: ${POSTGRES_PASSWORD}

search:
  default_strategy: semantic
  top_k: 5
  
strategies:
  - name: semantic
    distance_function: cosine
  - name: hybrid
    distance_function: l2
` + "```" + `

## Usage in Workflows

` + "```yaml" + `
steps:
  - name: search_docs
    rag:
      server: pgvector
      query: "${user_question}"
      top_k: 10
      strategies: [semantic, hybrid]
      fusion: rrf
      
  - name: answer
    prompt: |
      Based on these search results:
      ${search_docs.result}
      
      Answer: ${user_question}
` + "```" + `

## RAG Commands

` + "```bash" + `
# Search vector database
mcp-cli rag search "query" --server pgvector --top-k 10

# Show RAG configuration
mcp-cli rag config

# List available strategies
mcp-cli rag config --strategies
` + "```" + `

## Setting Up pgvector

1. Install PostgreSQL with pgvector extension:
   ` + "```bash" + `
   # Ubuntu/Debian
   sudo apt install postgresql-15-pgvector
   
   # macOS
   brew install pgvector
   ` + "```" + `

2. Create database and enable extension:
   ` + "```sql" + `
   CREATE DATABASE vector_db;
   \c vector_db
   CREATE EXTENSION vector;
   ` + "```" + `

3. Create embeddings table:
   ` + "```sql" + `
   CREATE TABLE documents (
     id SERIAL PRIMARY KEY,
     content TEXT,
     embedding vector(1536)  -- OpenAI embedding size
   );
   
   CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops);
   ` + "```" + `

4. Add connection config to config/rag/pgvector.yaml

5. Use in workflows or via CLI commands

## More Information

- pgvector: https://github.com/pgvector/pgvector
- RAG patterns: See config/workflows/ examples
- Query expansion: See config/rag/expansion/ for strategies
`
	
	readmePath := filepath.Join(ragDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create RAG README: %w", err)
	}
	
	fmt.Println("   ✓ Created config/rag/")
	fmt.Println("   ✓ Created config/rag/expansion/")
	fmt.Println("   ✓ Created config/rag/README.md")
	
	return nil
}

// createSkillsDirectory creates the skills directory with README pointing to Anthropic skills
func createSkillsDirectory(configDir string, cfg *InitConfig) error {
	skillsDir := filepath.Join(configDir, "skills")
	
	fmt.Println()
	color.New(color.FgCyan).Println("📦 Creating Skills Directory...")
	
	// Create skills directory
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}
	
	// Create README with clear native vs MCP distinction
	readmePath := filepath.Join(skillsDir, "README.md")
	readmeContent := `# Skills Directory

This directory contains Anthropic-compatible Skills for use with mcp-cli.

## 🔥 Native Integration (Automatic)

Skills are **automatically available** when you use mcp-cli commands:

` + "```bash" + `
# Skills work DIRECTLY - no server needed!
mcp-cli chat --skills docx,pdf
mcp-cli query "Create report" --skills xlsx
mcp-cli --workflow analyze --skills pdf
` + "```" + `

**No MCP server required!** Skills are loaded directly into the LLM context.

---

## 🌐 MCP Server Mode (Share with Other Clients)

You can **also** expose skills to OTHER MCP clients:

` + "```bash" + `
# Run mcp-cli as an MCP server
mcp-cli serve config/runas/skills-auto.yaml
` + "```" + `

This makes skills available to:
- Claude Desktop
- Claude Code  
- Other MCP-compatible clients

---

## Three Ways to Use Skills

### 1. Native in mcp-cli (Most Common) ✅

` + "```bash" + `
# Use in chat
mcp-cli chat --skills docx,pdf,xlsx

# Use in query
mcp-cli query "Create a report" --skills docx

# Use in workflow
mcp-cli --workflow analyze --skills pdf
` + "```" + `

**When:** You're using mcp-cli directly  
**Setup:** Zero  
**Works:** Immediately

### 2. MCP Server for Other Clients

` + "```bash" + `
mcp-cli serve config/runas/skills-auto.yaml
` + "```" + `

**When:** You want Claude Desktop/Code to use your skills  
**Setup:** Configure server  
**Works:** Via MCP protocol

### 3. Workflow Templates

` + "```yaml" + `
execution:
  skills: [docx, xlsx]  # Loaded automatically
steps:
  - name: create_report
    run: Use skills to generate report
` + "```" + `

**When:** You want reusable workflows  
**Setup:** Create YAML file  
**Works:** Via template system

---

## Available Skills

Download from Anthropic's official repository:

**🔗 https://github.com/anthropics/skills**

Popular skills:
- **docx** - Word document creation/editing
- **pdf** - PDF creation and manipulation  
- **pptx** - PowerPoint presentations
- **xlsx** - Excel spreadsheets
- **frontend-design** - Production-grade web UIs
- **bash-preference** - Bash tool guidance

---

## Installing Skills

### From GitHub (Recommended)

` + "```bash" + `
cd config/skills

# Download specific skills
git clone --depth 1 --filter=blob:none --sparse https://github.com/anthropics/skills
cd skills
git sparse-checkout set skills/docx skills/pdf skills/pptx skills/xlsx
mv skills/* ..
cd .. && rm -rf skills
` + "```" + `

### Manual Download

1. Visit https://github.com/anthropics/skills/tree/main/skills
2. Download the skill folder you want
3. Place it in ` + "`config/skills/`" + `

---

## Skill Structure

` + "```" + `
config/skills/my-skill/
├── SKILL.md              # Main instructions (required)
├── scripts/              # Helper libraries (optional)
│   └── helpers.py
└── reference.md          # Additional docs (optional)
` + "```" + `

---

## Creating Your Own Skills

### 1. Create Directory

` + "```bash" + `
mkdir -p config/skills/my-skill-name
` + "```" + `

### 2. Create SKILL.md

` + "```markdown" + `
---
name: my-skill-name
description: Brief description of what this skill does
---

# My Skill Name

## Overview
What this skill does and when Claude should use it.

## Instructions
Step-by-step instructions for Claude.

## Examples
Concrete usage examples.
` + "```" + `

### 3. Add Scripts (Optional)

` + "```python" + `
# config/skills/my-skill/scripts/helper.py
def process_data(data):
    """Helper function for skill"""
    return result
` + "```" + `

---

## Quick Reference

### List Available Skills
` + "```bash" + `
mcp-cli skills
` + "```" + `

### Use Skills Natively
` + "```bash" + `
mcp-cli chat --skills docx,pdf
mcp-cli query "..." --skills xlsx
` + "```" + `

### Start MCP Server
` + "```bash" + `
mcp-cli serve config/runas/skills-auto.yaml
` + "```" + `

---

## Resources

- **Official Skills**: https://github.com/anthropics/skills
- **Skills Documentation**: https://docs.anthropic.com/en/docs/build-with-claude/skills  
- **mcp-cli Documentation**: See docs/ directory
- **Container Details**: docs/CONTAINER_MOUNTING_EXPLAINED.md

---

**🚀 Skills work natively in mcp-cli - no server required!**  
**🌐 But you can also share them via MCP server when needed.**
`
	
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create skills README: %w", err)
	}
	
	fmt.Println("   ✓ Created README.md")
	
	// Create skill-images.yaml
	if err := createSkillImagesConfig(skillsDir); err != nil {
		return fmt.Errorf("failed to create skill-images.yaml: %w", err)
	}
	fmt.Println("   ✓ Created skill-images.yaml")
	
	fmt.Println("   💡 Download skills from: https://github.com/anthropics/skills")
	fmt.Println()
	
	return nil
}

// createSkillsRunAsConfig creates the skills-auto.yaml runas configuration
func createSkillsRunAsConfig(configDir string) error {
	runasDir := filepath.Join(configDir, "runasMCP")
	if err := os.MkdirAll(runasDir, 0755); err != nil {
		return err
	}
	
	skillsConfig := `# Skills MCP Server Configuration
# This configuration exposes all skills in config/skills as MCP tools

runas_type: mcp-skills

server_info:
  name: skills-engine
  version: 1.0.0
  description: MCP server providing Anthropic-compatible skills with dynamic code execution

# Skills configuration
skills_config:
  # Auto-detect Docker/Podman for code execution
  # Modes: auto (detect), active (require), passive (documentation only)
  execution_mode: auto
  
  # Skills directory (relative to config.yaml location)
  skills_directory: config/skills

# This configuration automatically:
# 1. Discovers all skills in the skills directory
# 2. Generates MCP tool definitions for each skill
# 3. Exposes execute_skill_code tool for dynamic code execution
# 4. Enables helper library imports from skill scripts/

# Usage:
#   ./mcp-cli serve config/runasMCP/skills-auto.yaml

# Claude can then:
# - Load skill documentation (passive mode)
# - Execute custom Python code with skill helpers
# - Import from skill's scripts/ directory
# - Create documents, process data, etc.
`
	
	configPath := filepath.Join(runasDir, "skills-auto.yaml")
	if err := os.WriteFile(configPath, []byte(skillsConfig), 0644); err != nil {
		return err
	}
	
	fmt.Println("   ✓ Created runasMCP/skills-auto.yaml")
	return nil
}

// createSkillImagesConfig creates the skill-images.yaml configuration
func createSkillImagesConfig(skillsDir string) error {
	skillImagesConfig := `# Skills Configuration V2
# Hierarchical structure with defaults and per-skill overrides
#
# Location: config/skills/skill-images.yaml

# Global defaults inherited by all skills (unless overridden)
defaults:
  image: python:3.11-alpine
  network_mode: none
  memory: 256MB
  cpu: "0.5"
  timeout: 60s
  outputs_dir: /tmp/mcp-outputs

# Per-skill configuration
# All settings for each skill in one place
skills:
  
  # Document processing skills (secure, no network needed)
  docx:
    image: mcp-skills-docx
    description: "Word document manipulation via OOXML/XML"
    dockerfile: docker/skills/Dockerfile.docx
    # Inherits: network_mode, memory, cpu, timeout from defaults
  
  pptx:
    image: mcp-skills-pptx
    description: "PowerPoint presentation creation/editing"
    dockerfile: docker/skills/Dockerfile.pptx
  
  xlsx:
    image: mcp-skills-xlsx
    description: "Excel spreadsheet manipulation"
    dockerfile: docker/skills/Dockerfile.xlsx
  
  pdf:
    image: mcp-skills-pdf
    description: "PDF manipulation, forms, text extraction, OCR"
    dockerfile: docker/skills/Dockerfile.pdf
    memory: 512MB  # Override: OCR needs more memory
    timeout: 120s  # Override: OCR takes longer

# ⚠️  SECURITY NOTES
#
# Network Access:
# - Default: 'none' (isolated, secure)
# - Skills are sandboxed with no internet access
# - Only enable network if absolutely required
#
# To enable network for a skill (e.g., AI model downloads):
#
#   imagegen:
#     image: mcp-skills-imagegen
#     description: "AI image generation"
#     network_mode: bridge  # Override default
#     memory: 2GB           # Override default
#     timeout: 300s         # Override default
#     network_justification: "Downloads models from HuggingFace"
#
# Security checklist before enabling network:
#   □ Understand why the skill needs network
#   □ Trust the skill code
#   □ Review what data it sends/receives
#   □ Document the justification
#   □ Inform security team (if applicable)

# Configuration Guide:
#
# 1. Defaults Section
#    - Sets base configuration for all skills
#    - Individual skills inherit these values
#
# 2. Skills Section  
#    - Each skill can override any default
#    - Only specify what differs from defaults
#    - Keep related settings together
#
# 3. Per-Skill Overrides
#    Available fields:
#    - image: Container image name (required)
#    - description: Brief description (optional)
#    - network_mode: none|bridge|host (default: none)
#    - dockerfile: Path to Dockerfile for auto-building (optional)
#    - memory: Memory limit (e.g., "512MB", "2GB")
#    - cpu: CPU limit (e.g., "0.5", "2.0")
#    - timeout: Execution timeout (e.g., "60s", "300s")
#    - mounts: Custom volume mounts (list, optional)
#    - environment: Environment variables (list, optional)
#    - network_justification: Why network is needed (required if network_mode != none)
#
# 4. Building Images
#    Build all skill images:
#      cd docker/skills && ./build-skills-images.sh
#
#    Build specific skill:
#      docker build -t mcp-skills-docx -f Dockerfile.docx .
#
# For more information:
#   - Documentation: docs/SKILLS_CONFIG_V2_PROPOSAL.md
#   - Quick Reference: docs/SKILLS_NETWORK_QUICK_REFERENCE.md
#   - Security Guide: docs/SECURITY.md
`
	
	configPath := filepath.Join(skillsDir, "skill-images.yaml")
	if err := os.WriteFile(configPath, []byte(skillImagesConfig), 0644); err != nil {
		return err
	}
	
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		return copyFile(path, dstPath)
	})
}
