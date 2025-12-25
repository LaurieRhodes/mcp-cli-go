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
providers, embeddings, servers, and templates.

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
	DefaultProvider   string
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
		Providers:      []string{"ollama"},
		Servers:        []string{}, // Empty - no assumptions about MCP servers
		IncludeOllama:  true,
		DefaultProvider: "ollama",
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
	fmt.Println("  • DeepSeek      - DeepSeek Chat (requires API key)")
	fmt.Println("  • Gemini        - Google Gemini (requires API key)")
	fmt.Println("  • OpenRouter    - Access many models (requires API key)")
	fmt.Println("  • LM Studio     - Local model server (no API key)")
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
	
	// Only create .env if there are API keys to configure
	if config.IncludeOpenAI || config.IncludeAnthropic || config.IncludeDeepSeek || 
	   config.IncludeGemini || config.IncludeOpenRouter {
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
		Providers:         initCfg.Providers,
		Servers:           initCfg.Servers,
		DefaultProvider:   initCfg.DefaultProvider,
		IncludeOllama:     initCfg.IncludeOllama,
		IncludeOpenAI:     initCfg.IncludeOpenAI,
		IncludeAnthropic:  initCfg.IncludeAnthropic,
		IncludeDeepSeek:   initCfg.IncludeDeepSeek,
		IncludeGemini:     initCfg.IncludeGemini,
		IncludeOpenRouter: initCfg.IncludeOpenRouter,
		IncludeLMStudio:   initCfg.IncludeLMStudio,
	}
	
	// Generate modular config
	generator := config.NewModularConfigGenerator(configDir)
	if err := generator.Generate(genConfig); err != nil {
		return fmt.Errorf("failed to generate modular config: %w", err)
	}
	
	// Create .env file at executable level (parent directory)
	parentDir := filepath.Dir(configDir)
	if initCfg.IncludeOpenAI || initCfg.IncludeAnthropic || initCfg.IncludeDeepSeek ||
		initCfg.IncludeGemini || initCfg.IncludeOpenRouter {
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
	fmt.Println("       ├── servers/")
	fmt.Println("       │   └── README.md")
	fmt.Println("       └── templates/")
	fmt.Println()
	
	if cfg.IncludeOpenAI || cfg.IncludeAnthropic || cfg.IncludeDeepSeek ||
		cfg.IncludeGemini || cfg.IncludeOpenRouter {
		color.New(color.FgYellow).Println("⚠️  Important: Add your API keys")
		fmt.Printf("   Edit: %s/.env\n", parentDir)
		fmt.Println()
	}
	
	info.Println("🎯 Next steps:")
	fmt.Printf("   1. Review: %s/README.md\n", configDir)
	if cfg.IncludeOpenAI || cfg.IncludeAnthropic || cfg.IncludeDeepSeek ||
		cfg.IncludeGemini || cfg.IncludeOpenRouter {
		fmt.Printf("   2. Edit .env: %s/.env\n", parentDir)
		fmt.Printf("   3. Run: ./mcp-cli query \"hello\"\n")
	} else {
		fmt.Printf("   2. Run: ./mcp-cli query \"hello\"\n")
	}
	
	fmt.Println()
	info.Println("💡 Tips:")
	fmt.Println("   • LLM and embedding providers are separated for clarity")
	fmt.Println("   • config.yaml is at executable level for easy discovery")
	fmt.Println("   • Each provider type has its own subdirectory")
	fmt.Println("   • Add MCP servers in servers/ directory")
	fmt.Println()
}
