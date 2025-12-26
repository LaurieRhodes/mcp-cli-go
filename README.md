# MCP-CLI-Go: Multi-Provider AI Workflow Engine

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/LaurieRhodes/mcp-cli-go)](https://github.com/LaurieRhodes/mcp-cli-go/releases)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-blue)](docs/)

A Go implementation of the Model Context Protocol (MCP) CLI that enables multi-step AI workflows across multiple providers from a single binary.

> **📚 New to MCP-CLI?** Check out the [comprehensive documentation](docs/) for guides, examples, and technical details.

---

## Table of Contents

- [What This Does](#what-this-does)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Real-World Examples](#real-world-examples)
- [MCP Server Mode](#mcp-server-mode)
- [Documentation](#documentation) 📚
- [Usage Modes](#usage-modes)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
- [Architecture](#architecture)
- [Resources](#resources)

---

## What This Does

**MCP-CLI-Go** is a command-line tool for building AI workflows that:

✅ **Chains multiple AI providers** - Mix Claude, GPT-4, Ollama, and others in one workflow  
✅ **Uses YAML templates** - Define reusable AI workflows without code  
✅ **Works as MCP server** - Expose workflows as tools for Claude Desktop or other MCP clients  
✅ **Runs locally** - Single Go binary, no dependencies  
✅ **Template composition** - Call templates from templates for modular workflows

### The Core Innovation

Traditional AI tools execute single requests. MCP-CLI enables **multi-step workflows** where:

- Each step can use a different AI provider
- Steps can call other templates (composition)
- Context is managed efficiently between steps
- Workflows are defined in YAML, not code

**Example:**

```yaml
name: research_workflow
steps:
  - name: research
    provider: anthropic
    prompt: "Research:  {{input_data}}"
    output: findings

  - name: verify
    provider: openai
    prompt: "Fact-check: {{findings}}"
    output: verified

  - name: summarize
    provider: ollama
    prompt: "Summarize: {{verified}}"
```

This workflow uses **three different AI providers** in sequence, each doing what they do best.

---

## Features

### ✅ Currently Available

- **Multiple AI Providers**: OpenAI, Anthropic, Ollama, DeepSeek, Gemini, OpenRouter, LM Studio
- **YAML Workflow Templates**: Define multi-step workflows without code
- **Template Composition**: Call templates from within templates
- **MCP Server Mode**: Expose workflows as tools for Claude Desktop
- **Variable Management**: Pass data between workflow steps
- **Error Handling**: Retries, validation, and graceful failures
- **Multiple Modes**: Chat, query, interactive, and server modes

### 🚧 Experimental

- **Parallel Execution**: Run multiple steps concurrently
- **Conditional Routing**: Branch workflows based on results
- **Recursion Control**: Prevent infinite template loops

---

## Installation

### Download Pre-Built Binaries

**Linux:**

```bash
# Download
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64

# Make executable
chmod +x mcp-cli-linux-amd64

# Move to PATH
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli

# Verify
mcp-cli --version
```

**macOS (Intel):**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-amd64
chmod +x mcp-cli-darwin-amd64
sudo mv mcp-cli-darwin-amd64 /usr/local/bin/mcp-cli
mcp-cli --version
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-arm64
chmod +x mcp-cli-darwin-arm64
sudo mv mcp-cli-darwin-arm64 /usr/local/bin/mcp-cli
mcp-cli --version
```

**Windows:**

```powershell
# Download from: https://github.com/LaurieRhodes/mcp-cli-go/releases/latest
# Extract mcp-cli-windows-amd64.exe and add to PATH
```

### Build from Source

```bash
git clone https://github.com/LaurieRhodes/mcp-cli-go.git
cd mcp-cli-go
go build -o mcp-cli
sudo mv mcp-cli /usr/local/bin/
```

**📚 Next Steps:** See the [Getting Started Guide](docs/getting-started/) for configuration and first-time setup.

---

## Quick Start

### 1. Initialize Configuration

```bash
# Interactive setup
mcp-cli init

# Quick setup (Ollama only, no API keys needed)
mcp-cli init --quick

# Add API keys if using cloud providers
echo "OPENAI_API_KEY=sk-..." >> .env
echo "ANTHROPIC_API_KEY=sk-ant-..." >> .env
```

This creates:

```
config/
├── providers/      # AI provider configs
├── embeddings/     # Embedding configs
├── servers/        # MCP server configs
└── templates/      # Workflow templates
```

### 2. Run a Simple Query

```bash
# Basic query
mcp-cli query "What is the Model Context Protocol?"

# With specific provider
mcp-cli query --provider anthropic "Explain MCP in detail"

# JSON output
mcp-cli query --json "List cloud providers" > result.json
```

### 3. Create Your First Template

Create `config/templates/analyze.yaml`:

```yaml
name: analyze
description: Simple analysis workflow
version: 1.0.0

steps:
  - name: analyze
    prompt: "Analyze this: {{stdin}}"
    output: analysis

  - name: summarize
    prompt: "Summarize in 3 bullets: {{analysis}}"
```

Run it:

```bash
echo "Sales data for Q4..." | mcp-cli --template analyze
```

### 4. Use Template Composition

Create `config/templates/research.yaml`:

```yaml
name: deep_research
description: Multi-step research with verification
version: 1.0.0

steps:
  - name: research
    template: web_search      # Calls another template
    output: findings

  - name: verify
    template: fact_check      # Calls another template
    template_input: "{{findings}}"
    output: verified

  - name: report
    prompt: "Create report: {{verified}}"
```

```bash
echo "Impact of AI on healthcare" | mcp-cli --template deep_research
```

**📚 Want to learn more?** See the [comprehensive guides](docs/guides/) and [template documentation](docs/templates/).

---

## Real-World Examples

### Document Analysis Pipeline

```yaml
name: document_intelligence
version: 1.0.0
steps:
  - name: sentiment
    template: sentiment_analysis
    output: sentiment

  - name: entities
    template: entity_extraction
    output: entities

  - name: summary
    prompt: |
      Create intelligence report:
      Sentiment: {{sentiment}}
      Entities: {{entities}}
```

### Multi-Provider Validation

```yaml
name: validated_answer
version: 1.0.0
steps:
  - name: answer
    provider: anthropic
    model: claude-sonnet-4
    prompt: "{{question}}"
    output: initial

  - name: verify
    provider: openai
    model: gpt-4o
    prompt: "Fact-check: {{input_data}}"
    output: verified
```

**📚 More examples:** See [Template Examples](docs/templates/examples/) for production-ready patterns.

---

## MCP Server Mode

Expose workflows as MCP tools for Claude Desktop or other clients.

### 1. Create Server Config

`config/runas/research_server.yaml`:

```yaml
server_info:
  name: research-agent
  version: 1.0.0
  description: Research assistant with web search

tools:
  - name: research_topic
    description: Research a topic
    template: deep_research
    input_schema:
      type: object
      properties:
        topic:
          type: string
      required: [topic]
```

### 2. Start Server

```bash
mcp-cli serve config/runas/research_server.yaml
```

### 3. Configure Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/usr/local/bin/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/research_server.yaml"]
    }
  }
}
```

Now Claude can use your workflow as a tool!

**📚 For complete MCP server documentation:** See [MCP Server Mode Guide](docs/mcp-server/)

---

## Usage Modes

### Chat Mode

```bash
# Interactive chat with AI
mcp-cli chat

# With specific provider
mcp-cli chat --provider anthropic
```

### Query Mode

```bash
# Single query
mcp-cli query "Your question here"

# With template
mcp-cli --template analyze --input-data "Some text"
```

### Interactive Mode

```bash
# Execute MCP server commands directly
mcp-cli interactive

# List available tools
mcp-cli tools
```

### Server Mode

```bash
# Run as MCP server
mcp-cli serve config/runas/your_server.yaml
```

---

## Configuration

### Provider Configuration

Example: `config/providers/openai.yaml`

```yaml
provider_name: openai
models:
  - name: gpt-4o
    max_tokens: 4096
  - name: gpt-4o-mini
    max_tokens: 16384
api_key: ${OPENAI_API_KEY}
```

### Template Structure

```yaml
name: template_name
description: What this does
version: 1.0.0

# Default settings
defaults:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

# Workflow steps
steps:
  - name: step1
    prompt: "Your prompt with  {{input_data}}"
    output: result1

  - name: step2
    template: another_template  # Template composition
    template_input: "{{result1}}"
    output: result2
```

**📚 For complete configuration reference:** See [Templates Documentation](docs/templates/authoring-guide.md)

---

## Command Reference

### Common Commands

```bash
# Initialize config
mcp-cli init
mcp-cli init --quick

# Query
mcp-cli query "question"
mcp-cli query --provider openai "question"
mcp-cli query --json "question"

# Templates
mcp-cli --template <name>
mcp-cli --list-templates

# Chat
mcp-cli chat
mcp-cli chat --provider anthropic

# Server
mcp-cli serve <config.yaml>
mcp-cli tools

# Help
mcp-cli --help
mcp-cli <command> --help
```

### Flags

```bash
--provider <name>       # AI provider (openai, anthropic, ollama, etc.)
--model <name>          # Model name
--template <name>       # Template to execute
--input-data <string>   # Input data for template
--json                  # JSON output
--verbose               # Verbose logging
--quiet                 # Minimal output
```

---

## Documentation

Comprehensive documentation is available in the [`docs/`](docs/) directory.

### 📚 Quick Links

| Documentation                                | Description                              |
| -------------------------------------------- | ---------------------------------------- |
| **[Documentation Index](docs/README.md)**    | Start here - complete navigation guide   |
| **[Getting Started](docs/getting-started/)** | Installation, configuration, first steps |
| **[Usage Guides](docs/guides/)**             | Mode-specific guides and best practices  |
| **[Templates](docs/templates/)**             | Template authoring and examples          |
| **[MCP Server Mode](docs/mcp-server/)**      | Expose workflows as MCP tools            |
| **[Architecture](docs/architecture/)**       | Technical design and internals           |

### 📖 By Topic

**New Users:**

- [Installation Guide](docs/getting-started/installation.md) - Install and configure MCP-CLI
- [Core Concepts](docs/getting-started/concepts.md) - Understand modes, providers, templates
- [FAQ](docs/getting-started/faq.md) - Common questions answered

**Usage Guides:**

- [Chat Mode](docs/guides/chat-mode.md) - Interactive conversations with AI
- [Query Mode](docs/guides/query-mode.md) - Single-shot queries for automation
- [Interactive Mode](docs/guides/interactive-mode.md) - Direct MCP tool testing
- [Automation & Scripting](docs/guides/automation.md) - CI/CD integration patterns
- [Debugging](docs/guides/debugging.md) - Troubleshooting and logging

**Template Development:**

- [Template Authoring Guide](docs/templates/authoring-guide.md) - Complete template reference
- [Template Examples](docs/templates/examples/) - Real-world template patterns

**Advanced Topics:**

- [MCP Server Documentation](docs/mcp-server/) - Expose templates as discoverable tools
- [Architecture Documentation](docs/architecture/) - System design for developers

---

## Architecture

### Template Composition

Templates can call other templates, creating modular, reusable workflows:

```
parent_template
  ├─> Calls child_template_1 (executes independently)
  │     └─> Returns result
  ├─> Calls child_template_2 (executes independently)
  │     └─> Returns result
  └─> Synthesizes results into final output
```

**Benefits:**

- **Modularity**: Reuse templates across workflows
- **Context Efficiency**: Each template has isolated context
- **Maintainability**: Update templates independently
- **Testability**: Test templates in isolation

### Multi-Provider Workflows

Each step in a workflow can use a different provider:

```yaml
steps:
  - provider: anthropic    # Use Claude for research
    prompt: "Research {{input_data}}"

  - provider: openai       # Use GPT-4 for analysis
    prompt: "Analyze {{research}}"

  - provider: ollama       # Use local model for synthesis
    prompt: "Synthesize {{analysis}}"
```

---

## Project Background

This project started in February 2025 as a fork of [chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli), which I needed for Go-based MCP server development. That project has continued to grow with talented contributors.

I built this as a tool for my own automation needs and shared it as example code. If you find it useful, great! If you want to contribute or see it maintained more actively, please reach out through [laurierhodes.info](https://laurierhodes.info/).

### Why Go?

- **Single binary**: No runtime dependencies
- **Cross-platform**: Linux, macOS, Windows
- **Fast startup**: Ideal for CLI tools
- **Easy deployment**: Just copy the binary

---

## Contributing

This project is shared as example code for your own development. Feel free to:

- Fork and modify for your needs
- Open issues for bugs
- Submit pull requests for fixes
- Share your templates and workflows

I'm happy to review contributions, though I can't promise active maintenance.

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

**Original Project:** This started as a go fork of [chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli) in its first few weeks of development.  That project is actively maintained by a team of talented developers who have incorporated many new features sinve February 2025.  Check it out and give it a deserved star!

**Model Context Protocol:** Created by Anthropic - [modelcontextprotocol.io](https://modelcontextprotocol.io)

**The Go Community:** For excellent tooling and libraries

---

## Resources

### Documentation

- **[Complete Documentation](docs/README.md)** - Comprehensive guides and references
- **[Getting Started](docs/getting-started/)** - Installation and configuration
- **[Usage Guides](docs/guides/)** - Mode-specific tutorials
- **[Template Authoring](docs/templates/)** - Creating workflows
- **[MCP Server Mode](docs/mcp-server/)** - Expose workflows as tools
- **[Architecture](docs/architecture/)** - Technical design documentation

### Project Links

- **Source Code**: [github.com/LaurieRhodes/mcp-cli-go](https://github.com/LaurieRhodes/mcp-cli-go)
- **Releases**: [Releases Page](https://github.com/LaurieRhodes/mcp-cli-go/releases)
- **Issues**: [Issue Tracker](https://github.com/LaurieRhodes/mcp-cli-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)

### External Resources

- **MCP Protocol**: [modelcontextprotocol.io](https://modelcontextprotocol.io)
- **Author**: [laurierhodes.info](https://laurierhodes.info)

---

<div align="center">

**Built with Go • Powered by MCP**

If this project helps you, please give it a ⭐!

</div>
