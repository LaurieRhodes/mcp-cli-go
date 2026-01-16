# Getting Started with MCP-CLI-Go

Welcome! This guide will get you from zero to productive with AI workflows in under 30 minutes.

**What is MCP-CLI-Go?** A command-line tool that lets you:

- Run AI queries from your terminal
- Build multi-step AI workflows (templates)
- Mix different AI providers (Claude + GPT-4 + local models)
- Give AI access to tools (filesystem, web search, databases)
- Automate AI tasks in scripts and pipelines

**Why use it?**

- 🚀 **50-87% token savings** through smart context isolation
- 🔧 **Composable workflows** - build complex AI pipelines from simple parts
- 🎯 **Multi-provider** - use best AI for each task
- 🔌 **Tool integration** - AI can use real tools via MCP
- 📝 **Version controlled** - workflows in YAML, commit to git
- 🆓 **Free local option** - works with Ollama (no API keys)

---

## 🎯 Quick Start Path (30 minutes)

### 1. Install (5 minutes)

**[→ Installation Guide](installation.md)**

**Quick install commands:**

**Linux:**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
chmod +x mcp-cli-linux-amd64
sudo mv mcp-cli-linux-amd64 mcp-cli
mcp-cli --version
```

**macOS:**

```bash
# Intel Mac
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-amd64
# Apple Silicon (M1/M2/M3)
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-arm64

chmod +x mcp-cli-darwin-*
xattr -d com.apple.quarantine mcp-cli-darwin-*
sudo mv mcp-cli-darwin-* mcp-cli
mcp-cli --version
```

**Windows PowerShell:**

```powershell
Invoke-WebRequest -Uri "https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-windows-amd64.exe" -OutFile "mcp-cli.exe"
New-Item -ItemType Directory -Force -Path $env:USERPROFILE\bin
Move-Item mcp-cli.exe $env:USERPROFILE\bin\
$env:Path += ";$env:USERPROFILE\bin"
mcp-cli --version
```

---

### 2. Initialize Configuration (2 minutes)

**Option A: Quick setup (free local AI, no API keys):**

```bash
mcp-cli init --quick
```

**Option B: Full setup (cloud AI providers):**

```bash
mcp-cli init
# Follow interactive prompts
# Add API keys to .env file
```

**What this creates:**

- `config/` directory with settings
- `.env` file for API keys
- Example templates to get started

---

### 3. Your First Query (1 minute)

**Simple question:**

```bash
mcp-cli query "What are the top 3 programming languages in 2024?"
```

**Expected output:**

```
Based on current trends and usage:

1. **Python** - Dominant in AI/ML, data science, and automation
2. **JavaScript/TypeScript** - Essential for web development
3. **Go** - Growing for cloud infrastructure and microservices

All three show strong job demand and active communities.
```

**What just happened:**

- MCP-CLI sent your question to default AI provider
- AI generated response
- Result printed to terminal

---

### 4. Your First Template (5 minutes)

**Create a template:**

```bash
# Create templates directory if needed
mkdir -p config/templates

# Create analyze.yaml
cat > config/workflows/analyze.yaml <<'EOF'
name: analyze
description: Analyze text and create summary
version: 1.0.0

steps:
  - name: analyze
    prompt: "Analyze this text: {{stdin}}"
    output: analysis

  - name: summarize
    prompt: "Create 3-bullet summary: {{analysis}}"
EOF
```

**Use your template:**

```bash
echo "Sales were up 15% in Q4..." | mcp-cli --template analyze
```

**Expected output:**

```
• Revenue growth strong at 15% in Q4
• Customer acquisition exceeded targets
• Churn rate needs monitoring at 5.2%
```

**What just happened:**

1. Template received input via stdin
2. Step 1 analyzed the text
3. Step 2 summarized analysis into bullets
4. Result printed to terminal

**This is template power:** 2 steps, automatic variable flow, reusable!

---

### 5. Understand the Concepts (10 minutes)

**[→ Core Concepts](concepts.md)**

**Key concepts to understand:**

**MCP (Model Context Protocol):**

- Standard for connecting AI to tools
- Like USB for AI integrations
- AI can read files, search web, query databases

**Templates:**

- Multi-step AI workflows in YAML
- Reusable and version-controlled
- 50-87% token savings vs manual

**Template Composition:**

- Templates can call other templates
- Build complex workflows from simple parts
- Each sub-template isolated (saves tokens)

**Multi-Provider:**

- Use Claude for research, GPT-4 for code, Ollama for formatting
- Best model for each task
- Mix free local + paid cloud

**Variables:**

- `{{stdin}}` - input from pipe or flag
- `{{step_name}}` - output from previous step
- Pass data between steps automatically

---

### 6. Try More Templates (5 minutes)

**List available templates:**

```bash
mcp-cli --list-templates
```

**Try example templates:**

```bash
# Sentiment analysis
echo "I love this product!" | mcp-cli --template sentiment

# Entity extraction
echo "John Smith works at Acme Corp in Boston" | mcp-cli --template extract_entities

# Multi-provider research
mcp-cli --template research --input-data "impact of AI on healthcare"
```

---

### 7. Next Steps (Your Choice)

**For scripters and automators:**
→ [Query Mode Guide](../guides/query-mode.md)

- Use in bash scripts
- CI/CD integration
- Automation patterns

**For template creators:**
→ [Workflow Authoring Guide](../workflows/authoring-guide.md)

- Advanced template features
- Conditions, loops, parallel execution
- Template composition

**For interactive users:**
→ [Chat Mode Guide](../guides/chat-mode.md)

- Conversational AI
- Tool use (MCP servers)
- Development and debugging

**For production deployment:**
→ [Deployment Guide](../deployment/)

- Docker containers
- Kubernetes deployment
- Systemd services

---

## 🚀 Essential Commands Reference

### Setup Commands

```bash
mcp-cli init                    # Full interactive setup
mcp-cli init --quick            # Quick setup (Ollama only, no API keys)
mcp-cli --version               # Check version
mcp-cli --help                  # Show all commands
```

### Query Mode (One-Shot Questions)

```bash
# Basic query
mcp-cli query "What is 2+2?"

# With specific provider
mcp-cli query --provider anthropic "Explain quantum computing"

# With specific model
mcp-cli query --provider openai --model gpt-4o "Write a haiku"

# JSON output (for parsing in scripts)
mcp-cli query --json "List top 5 cities" | jq '.response'

# Save to file
mcp-cli query "Summarize AI trends" > report.txt
```

### Chat Mode (Interactive Conversation)

```bash
# Start chat with default provider
mcp-cli chat

# With specific provider
mcp-cli chat --provider anthropic

# With MCP servers (tools for AI)
mcp-cli chat --server filesystem,brave-search

# Chat commands (use inside chat):
/help      # Show available commands
/clear     # Clear conversation history
/model     # Switch model
/exit      # Exit chat
```

### Workflow Mode (Multi-Step Workflows)

```bash
# Run template with piped input
echo "data" | mcp-cli --template analyze

# Run template with flag input
mcp-cli --template analyze --input-data "text to analyze"

# Run template with file input
cat document.txt | mcp-cli --template summarize

# List all available templates
mcp-cli --list-templates

# List templates with descriptions
mcp-cli --list-templates --verbose
```

### Server Mode (Expose as MCP Server)

```bash
# Start as MCP server
mcp-cli serve config/runas/agent.yaml

# Now other apps (like Claude Desktop) can use your templates!
```

### Debugging Commands

```bash
# Verbose output (see what's happening)
mcp-cli --verbose query "test"

# Extra verbose (even more detail)
mcp-cli --verbose --verbose query "test"

# Check configuration
mcp-cli --config config.yaml --verbose query "test"
```

### Directory Structure

After initialization:

```
your-directory/
├── mcp-cli                  # Executable
├── .env                     # API keys
├── config.yaml              # Main config
└── config/
    ├── settings.yaml        # Global settings
    ├── providers/           # AI provider configs
    ├── embeddings/          # Embedding configs
    ├── servers/             # MCP server configs
    ├── workflows/           # Workflows
    └── runas/               # MCP server mode configs
```

### Workflow Basics

Minimal template:

```yaml
name: my_template
description: What it does
version: 1.0.0

steps:
  - name: step1
    prompt: "Do something with {{stdin}}"
    output: result

  - name: step2
    prompt: "Do more with {{result}}"
```

---

## 📚 Next Steps

### Beginner Path

1. **[Installation](installation.md)** - Get it running
2. **[Quick Start](../quick_start.md)** - First queries and templates
3. **[Concepts](concepts.md)** - Understand the system
4. **[Chat Mode Guide](../guides/chat-mode.md)** - Interactive use
5. **[Query Mode Guide](../guides/query-mode.md)** - Scripting use

### Intermediate Path

1. **[Workflow Authoring](../workflows/authoring-guide.md)** - Write powerful templates
2. **[Workflow Examples](../workflows/examples/)** - Learn from examples
3. **[Multi-Provider Guide](../providers/)** - Mix different AIs
4. **[MCP Server Setup](../mcp-server/)** - Expose as tools
5. **[Automation Guide](../guides/automation.md)** - Production workflows

### Advanced Path

1. **[Workflow Patterns](../workflows/patterns/)** - Design patterns
2. **[Architecture](../architecture/)** - System design
3. **[Deployment](../deployment/)** - Production deployment
4. **[Performance](../performance/)** - Optimization
5. **[Contributing](../../CONTRIBUTING.md)** - Contribute back

---

## 🎓 Learn by Example

### Simple Query

```bash
# Ask a question
mcp-cli query "What are the top 5 programming languages?"

# With specific model
mcp-cli query --provider anthropic --model claude-sonnet-4 \
  "Explain quantum computing"

# JSON output for parsing
mcp-cli query --json "List cloud providers" | jq '.response'
```

### Simple Template

```yaml
# config/workflows/analyze.yaml
name: analyze
description: Analyze and summarize text
version: 1.0.0

steps:
  - name: analyze
    prompt: "Analyze this: {{stdin}}"
    output: analysis

  - name: summarize
    prompt: "Summarize: {{analysis}}"
```

Usage:

```bash
echo "Sales data Q4..." | mcp-cli --template analyze
```

### Multi-Provider Template

```yaml
# config/workflows/research.yaml
name: research
description: Multi-provider research workflow
version: 1.0.0

steps:
  # Claude researches
  - name: research
    provider: anthropic
    model: claude-sonnet-4
    prompt: "Research: {{stdin}}"
    output: findings

  # GPT-4 verifies
  - name: verify
    provider: openai
    model: gpt-4o
    prompt: "Verify: {{findings}}"
    output: verified

  # Ollama summarizes (free!)
  - name: summarize
    provider: ollama
    model: qwen2.5:32b
    prompt: "Summarize: {{verified}}"
```

---

## 💡 Tips for Success

### 1. Start Simple

Begin with:

- Single-step templates
- One provider (Ollama for free local)
- Basic prompts
- Simple variables

**Then gradually add:**

- Multi-step workflows
- Multiple providers
- Template composition
- Error handling

### 2. Use the Right Mode

- **Query mode**: Scripts, automation, CI/CD
- **Chat mode**: Exploration, development, debugging
- **Template mode**: Production workflows
- **Server mode**: Integration with other tools

### 3. Leverage Examples

Don't start from scratch:

- Copy [example templates](../workflows/examples/)
- Adapt [design patterns](../workflows/patterns/)
- Learn from [real-world use cases](../workflows/examples/real-world.md)

### 4. Test Incrementally

Build templates step-by-step:

```bash
# Test step 1 alone
mcp-cli query "step 1 prompt"

# Add step 2, test
mcp-cli --template my_template

# Add step 3, test
# etc.
```

### 5. Read Error Messages

MCP-CLI provides detailed errors:

```
Error: Template 'analyze' not found
Searched: config/workflows/analyze.yaml
Available templates: research, summarize, extract
```

Use `--verbose` for more details:

```bash
mcp-cli --verbose --template my_template
```

---

## 🆘 Getting Help

### Documentation

- **Guides**: [docs/guides/](../guides/)
- **Templates**: [docs/workflows/](../workflows/)
- **Providers**: [docs/providers/](../providers/)
- **Architecture**: [docs/architecture/](../architecture/)

### Community

- **Discussions**: [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)
- **Issues**: [GitHub Issues](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Show & Tell**: [Share your templates](https://github.com/LaurieRhodes/mcp-cli-go/discussions/categories/show-and-tell)

### Troubleshooting

- **[Troubleshooting Guide](../troubleshooting/)** - Common issues
- **[FAQ](faq.md)** - Frequently asked questions
- **[CLI Reference](../../CLI-REFERENCE.md)** - Complete command reference

---

## 📖 Complete Documentation

### By Topic

**Getting Started** (You are here)

- [Installation](installation.md)
- [Quick Start](../quick_start.md)
- [Concepts](concepts.md)
- [FAQ](faq.md)

**Usage Guides**

- [Chat Mode](../guides/chat-mode.md)
- [Query Mode](../guides/query-mode.md)
- [Interactive Mode](../guides/interactive-mode.md)
- [Automation](../guides/automation.md)
- [Embeddings](../guides/embeddings.md)

**Templates**

- [Authoring Guide](../workflows/authoring-guide.md)
- [Reference](../workflows/reference.md)
- [Examples](../workflows/examples/)
- [Patterns](../workflows/patterns/)
- [Best Practices](../workflows/best-practices.md)

**Providers**

- [Overview](../providers/)
- [OpenAI](../providers/openai.md)
- [Anthropic](../providers/anthropic.md)
- [Ollama](../providers/ollama.md)
- [Comparison](../providers/comparison.md)

**MCP Server**

- [Setup](../mcp-server/)
- [Claude Desktop](../mcp-server/claude-desktop.md)
- [Examples](../mcp-server/examples/)

**Deployment**

- [Docker](../deployment/docker.md)
- [Kubernetes](../deployment/kubernetes.md)
- [Production Checklist](../deployment/production-checklist.md)

---

## ✅ Success Checklist

After completing this section, you should be able to:

- [ ] Install MCP-CLI-Go on your system
- [ ] Run your first query
- [ ] Create a simple template
- [ ] Execute a template
- [ ] Understand template composition
- [ ] Use multiple providers
- [ ] Navigate the documentation

**Ready?** Start with **[Installation →](installation.md)**

---

**Need help?** Join the [discussion](https://github.com/LaurieRhodes/mcp-cli-go/discussions) or ask in [issues](https://github.com/LaurieRhodes/mcp-cli-go/issues).

Happy automating! 🚀
