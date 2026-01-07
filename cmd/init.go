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
}

func printWelcome() {
	color.New(color.FgCyan, color.Bold).Println("\n🚀 MCP CLI Setup Wizard")
	fmt.Println("This will help you set up mcp-cli for the first time.")
}

func createQuickConfig() *InitConfig {
	fmt.Println("📦 Quick Setup Mode")
	fmt.Println("   Creating minimal configuration with ollama (no API keys needed)")
	fmt.Println()
	
	return &InitConfig{
		Providers:       []string{"ollama"},
		Servers:         []string{}, // Empty - no assumptions about MCP servers
		IncludeOllama:   true,
		DefaultProvider: "ollama",
		IncludeSkills:   true, // Include skills directory by default
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
	if cfg.IncludeOpenAI || cfg.IncludeOpenRouter || cfg.IncludeOllama {
		for _, provider := range cfg.Providers {
			if provider == "openai" || provider == "openrouter" || provider == "ollama" {
				fmt.Printf("       │   ├── %s.yaml\n", provider)
			}
		}
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

// createSkillsDirectory creates the skills directory with README pointing to Anthropic skills
func createSkillsDirectory(configDir string, cfg *InitConfig) error {
	skillsDir := filepath.Join(configDir, "skills")
	
	fmt.Println()
	color.New(color.FgCyan).Println("📦 Creating Skills Directory...")
	
	// Create skills directory
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}
	
	// Create README pointing to Anthropic skills archive
	readmePath := filepath.Join(skillsDir, "README.md")
	readmeContent := `# MCP Skills Directory

This directory is for Anthropic-compatible skills that extend mcp-cli capabilities.

## What are Skills?

Skills provide Claude with:
- **Documentation** (SKILL.md) - Instructions and examples
- **Helper Libraries** (scripts/) - Reusable Python code
- **Dynamic Code Execution** - Claude writes custom code for each task

Example skill structure:
` + "```" + `
config/skills/my-skill/
├── SKILL.md              # Skill documentation
└── scripts/              # Helper libraries
    ├── __init__.py
    └── helpers.py
` + "```" + `

## Download Skills from Anthropic

Anthropic maintains an official skills repository:

**🔗 https://github.com/anthropics/skills/tree/main/skills**

### Available Skills Include:

- **docx** - Create and edit Word documents
- **pdf** - Create and manipulate PDFs
- **pptx** - Create PowerPoint presentations
- **xlsx** - Create Excel spreadsheets
- **frontend-design** - Build production-grade web UIs
- **product-self-knowledge** - Product documentation
- **bash-preference** - Bash tool usage guidance
- **And more...**

### How to Add a Skill

1. **Download from GitHub:**
   ` + "```bash" + `
   cd config/skills
   git clone --depth 1 --filter=blob:none --sparse https://github.com/anthropics/skills
   cd skills
   git sparse-checkout set skills/docx skills/pdf skills/pptx
   mv skills/* ..
   cd .. && rm -rf skills
   ` + "```" + `

2. **Or manually:**
   - Visit https://github.com/anthropics/skills/tree/main/skills
   - Download the skill folder you want
   - Place it in ` + "`config/skills/`" + `

3. **Restart MCP server:**
   ` + "```bash" + `
   ./mcp-cli serve config/runasMCP/skills-auto.yaml
   ` + "```" + `

## Using Skills as MCP Server

Start the skills MCP server to expose all skills to Claude:

` + "```bash" + `
./mcp-cli serve config/runasMCP/skills-auto.yaml
` + "```" + `

This provides:
- All skills in this directory as MCP tools
- ` + "`execute_skill_code`" + ` tool for dynamic Python execution
- Helper library imports from skill ` + "`scripts/`" + ` directories

## Dynamic Code Execution

Claude can execute custom Python code with access to skill helpers:

` + "```python" + `
# Example: Using docx skill
from scripts.document import Document

doc = Document()
doc.add_heading("My Report", level=1)
doc.add_paragraph("Generated by Claude")
doc.save("report.docx")
` + "```" + `

## Creating Your Own Skills

1. Create directory: ` + "`config/skills/my-skill/`" + `
2. Add ` + "`SKILL.md`" + ` with YAML frontmatter:
   ` + "```yaml" + `
   ---
   name: my-skill
   description: What this skill does
   license: MIT
   ---
   ` + "```" + `
3. Add ` + "`scripts/`" + ` directory with Python helper code
4. Document usage examples in SKILL.md
5. Restart MCP server

## More Information

- **Official Skills**: https://github.com/anthropics/skills
- **Skills Documentation**: https://github.com/anthropics/skills/blob/main/README.md
- **Create Skills**: https://docs.anthropic.com/en/docs/build-with-claude/skills

## Notes

- Skills require Docker or Podman for code execution
- The ` + "`skills-auto.yaml`" + ` config auto-detects all skills in this directory
- Skills are sandboxed for security
- Helper libraries can be imported with ` + "`from scripts import ...`" + `
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
	skillImagesConfig := `# Skill Container Image Mapping
# Maps skill names to their required container images
#
# This file allows us to specify which container image each skill needs
# without modifying the upstream Anthropic skill files.
#
# Location: config/skills/skill-images.yaml

# Default image for skills without a specific mapping
# This should be a general-purpose image with common dependencies
default_image: python:3.11-alpine

# Per-skill image mappings
# Key: skill name (must match the skill directory name)
# Value: container image name
skills:
  # Office document skills
  docx: mcp-skills-docx
  pptx: mcp-skills-pptx
  xlsx: mcp-skills-xlsx
  
  # PDF skill
  pdf: mcp-skills-pdf
  
  # If you need Excel formula recalculation, use this instead:
  # xlsx: mcp-skills-xlsx-libreoffice
  
  # Other skills can be added here as needed
  # imagegen: mcp-skills-imagegen

# Container runtime configuration (optional)
# These settings apply to all container executions
container_config:
  # Maximum memory per container
  memory_limit: "256m"
  
  # CPU limit (fractional cores)
  cpu_limit: "0.5"
  
  # Execution timeout (seconds)
  timeout: 60
  
  # Network mode (should always be 'none' for security)
  network: "none"
  
  # Maximum number of processes
  pids_limit: 100

# Image build information (for reference only)
# These are not used by the executor
# Build images with: cd docker/skills && ./build-skills-images.sh
image_info:
  mcp-skills-docx:
    size: "~170 MB"
    packages:
      - defusedxml
      - lxml
    description: "Word document manipulation via OOXML/XML"
    
  mcp-skills-pptx:
    size: "~190 MB"
    packages:
      - python-pptx
      - Pillow
      - lxml
    description: "PowerPoint presentation creation/editing"
    
  mcp-skills-xlsx:
    size: "~175 MB"
    packages:
      - openpyxl
      - lxml
    description: "Excel spreadsheet manipulation (basic)"
    
  mcp-skills-xlsx-libreoffice:
    size: "~350 MB"
    packages:
      - openpyxl
      - lxml
      - libreoffice-calc
    description: "Excel with formula recalculation support"
    
  mcp-skills-office:
    size: "~195 MB"
    packages:
      - defusedxml
      - python-pptx
      - openpyxl
      - Pillow
      - lxml
    description: "Combined image for all Office formats"
    
  mcp-skills-pdf:
    size: "~220 MB"
    packages:
      - pypdf
      - pdf2image
      - Pillow
      - pdfplumber
      - pytesseract
      - poppler-utils
      - tesseract-ocr
    description: "PDF manipulation, forms, text extraction, OCR"
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
