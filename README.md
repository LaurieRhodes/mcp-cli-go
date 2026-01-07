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
- [Skills: Cross-LLM Document Creation](#skills-cross-llm-document-creation)
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
✅ **Uses YAML workflows** - Define reusable AI workflows without code  
✅ **Works as MCP server** - Expose workflows as tools for Claude Desktop or other MCP clients  
✅ **Runs locally** - Single Go binary, no dependencies  
✅ **Workflow composition** - Call workflows from workflows for modular designs  
✅ **Cross-LLM document creation** - GPT-4, DeepSeek, Gemini can create PowerPoints, Excel files, etc. via Skills  
✅ **Iterative refinement** - LLM-evaluated loops improve output until criteria met

### The Core Innovation

Traditional AI tools execute single requests. MCP-CLI enables **multi-step workflows** where:

- Each step can use a different AI provider
- Steps can call other workflows (composition)
- Loops can improve output iteratively based on LLM-evaluated exit conditions
- Consensus validation uses multiple providers for critical decisions
- Context is managed efficiently between steps
- Workflows are defined in YAML, not code

**Example:**

```yaml
$schema: "workflow/v2.0"
name: research_workflow
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: research
    run: "Research: {{input}}"

  - name: verify
    needs: [research]
    provider: openai
    model: gpt-4o
    run: "Fact-check: {{research}}"

  - name: summarize
    needs: [verify]
    provider: ollama
    model: llama3.2
    run: "Summarize: {{verify}}"
```

This workflow uses **three different AI providers** in sequence, each doing what they do best.

---

## Features

### ✅ Currently Available

- **Multiple AI Providers**: OpenAI, Anthropic, Ollama, DeepSeek, Gemini, Kimi K2, OpenRouter, LM Studio, AWS Bedrock, Azure Foundry, Google Vertex
- **YAML Workflow System (v2.0)**: Define multi-step workflows with advanced features
- **Iterative Loops**: LLM-evaluated exit conditions for automatic improvement
- **Consensus Validation**: Multi-provider agreement for critical decisions
- **Workflow Composition**: Call workflows from within workflows
- **Property Inheritance**: Set defaults once, override where needed
- **Step Dependencies**: Control execution order with `needs`
- **MCP Server Mode**: Expose workflows as tools for LLMs
- **Skills System**: Cross-LLM document creation (PowerPoint, Excel, Word, PDF) via containerized execution
- **Variable Management**: Pass data between workflow steps
- **Error Handling**: Retries, fallback chains, and graceful failures
- **Multiple Modes**: Chat, query, interactive, and server modes

### 🎯 Workflow v2.0 Highlights

- **LLM-Evaluated Loops**: Continue iterating until "all tests pass" or "quality exceeds 8"
- **Consensus Mode**: Get unanimous agreement from multiple providers
- **Semantic Exit Conditions**: No exact string matching - LLM understands intent
- **Loop Variables**: Access previous iterations with `{{loop.last.output}}`
- **Provider Failover**: Automatic failover across provider chains

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
└── workflows/      # Workflow definitions
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

### 3. Create Your First Workflow

Create `config/workflows/analyze.yaml`:

```yaml
$schema: "workflow/v2.0"
name: analyze
version: 1.0.0
description: Simple analysis workflow

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: analyze
    run: "Analyze this: {{input}}"

  - name: summarize
    needs: [analyze]
    run: "Summarize in 3 bullets: {{analyze}}"
```

Run it:

```bash
./mcp-cli --workflow analyze --input-data "Sales data for Q4..."
```

### 4. Use Workflow Composition

Create `config/workflows/research.yaml`:

```yaml
$schema: "workflow/v2.0"
name: deep_research
version: 1.0.0
description: Multi-step research with verification

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: research
    template:
      name: web_search      # Calls another workflow
      with:
        query: "{{input}}"

  - name: verify
    needs: [research]
    template:
      name: fact_check      # Calls another workflow
      with:
        content: "{{research}}"

  - name: report
    needs: [verify]
    run: "Create report: {{verify}}"
```

```bash
./mcp-cli --workflow deep_research --input-data "Impact of AI on healthcare"
```

### 5. Use Iterative Loops

Create `config/workflows/iterative_dev.yaml`:

```yaml
$schema: "workflow/v2.0"
name: iterative_developer
version: 1.0.0

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: requirements
    run: "Analyze request: {{input}}"

loops:
  - name: develop
    workflow: code_cycle
    with:
      requirements: "{{requirements}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "All tests pass"  # LLM evaluates this
    on_failure: continue
```

The loop continues until LLM determines "all tests pass"!

**📚 Want to learn more?** See [Workflow Documentation](docs/workflows/) and [examples](docs/workflows/examples/).

---

## Real-World Examples

### Quick Workflow Examples

### Document Analysis Pipeline

```yaml
$schema: "workflow/v2.0"
name: document_intelligence
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: extract
    run: "Extract key information from: {{input}}"

  - name: analyze
    needs: [extract]
    run: "Analyze content: {{extract}}"

  - name: summarize
    needs: [analyze]
    run: |
      Create intelligence report:
      Analysis: {{analyze}}
```

### Multi-Provider Consensus

```yaml
$schema: "workflow/v2.0"
name: validated_decision
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: validate
    consensus:
      prompt: "Is this safe to deploy? {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

### Iterative Code Development

```yaml
$schema: "workflow/v2.0"
name: code_developer
version: 1.0.0

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: requirements
    run: "Define requirements: {{input}}"

loops:
  - name: develop
    workflow: write_and_test
    with:
      requirements: "{{requirements}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "All tests pass"
```

**📚 More examples:** 

- [Workflow Examples](docs/workflows/examples/) - 13 working examples
- [Workflow Patterns](docs/workflows/patterns/) - Design patterns
- Working example: `config/workflows/iterative_dev/`

> **Note:** Industry showcases in `docs/templates/showcases/` use the older template format and are being migrated to workflow v2.0. Check back soon for updated versions!

---

## Skills: Cross-LLM Document Creation

**Skills** enable any LLM to create documents (PowerPoint, Excel, Word, PDFs) through secure container-based execution.

### What Makes This Special

Traditional approach: Only Claude can create documents (via Anthropic's computer use)  
**Skills approach**: GPT-4, DeepSeek, Gemini, Claude - all can create documents via MCP

### How It Works

```
LLM Request → Skills Documentation → LLM Writes Code → Secure Container → File on Host
```

Skills provide:

- **Documentation** - Instructions for LLMs on library usage
- **Helper Libraries** - Pre-installed packages (python-pptx, openpyxl, etc.)
- **Container Execution** - Isolated, secure Python environment
- **File Persistence** - Output files appear on your filesystem

### Quick Start

**1. Build container images:**

```bash
cd docker/skills
./build-skills-images.sh
```

**2. Configure outputs** (in `config/settings.yaml`):

```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"
```

**3. Start skills server:**

```bash
mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

**4. Use with any LLM via MCP**

### Available Skills

- **docx** - Word documents
- **pptx** - PowerPoint presentations  
- **xlsx** - Excel spreadsheets
- **pdf** - PDF manipulation and forms

### Example Usage

User to GPT-4 via MCP: *"Create a PowerPoint about Q4 sales"*

GPT-4:

1. Reads pptx skill documentation
2. Writes Python code using python-pptx
3. Code executes in container
4. File appears at `~/outputs/q4-sales.pptx`

### Security

All code runs in isolated containers with:

- No network access
- Read-only root filesystem
- Memory and CPU limits
- Automatic cleanup

**📚 Complete documentation:** [docs/skills/](docs/skills/)

---

## MCP Server Mode

Expose workflows as MCP tools for Claude Desktop or other clients.

### 1. Create Server Config

`config/runasMCP/research_server.yaml`:

```yaml
server_info:
  name: research-agent
  version: 1.0.0
  description: Research assistant with web search

tools:
  - name: research_topic
    description: Research a topic
    workflow: deep_research  # Points to workflow file
    input_schema:
      type: object
      properties:
        topic:
          type: string
      required: [topic]
```

### 2. Start Server

```bash
mcp-cli serve config/runasMCP/research_server.yaml
```

### 3. Configure Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/usr/local/bin/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/research_server.yaml"]
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

# With workflow
mcp-cli --workflow analyze --input-data "Some text"
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
mcp-cli serve config/runasMCP/your_server.yaml
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

### Workflow Structure (v2.0)

```yaml
$schema: "workflow/v2.0"
name: workflow_name
version: 1.0.0
description: What this workflow does

# Execution context (defaults for all steps)
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  servers: [filesystem, brave-search]

# Environment variables
env:
  API_KEY: ${CUSTOM_API_KEY}

# Sequential steps
steps:
  - name: step1
    run: "Your prompt with {{input}}"

  - name: step2
    needs: [step1]
    run: "Process result: {{step1}}"

# Iterative loops
loops:
  - name: improve
    workflow: refinement_cycle
    with:
      content: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "Quality score exceeds 8"
    on_failure: continue
```

**📚 For complete configuration reference:**

- [Schema Documentation](docs/workflows/SCHEMA.md) - Complete schema reference
- [Authoring Guide](docs/workflows/AUTHORING_GUIDE.md) - How to write workflows
- [Loop Guide](docs/workflows/LOOPS.md) - Iterative execution
- [Examples](docs/workflows/examples/) - Working examples

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

# Workflows
mcp-cli --workflow <name> --input-data "..."
mcp-cli --list-workflows

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
--workflow <name>       # Workflow to execute
--input-data <string>   # Input data for workflow
--server <name>         # MCP server to use
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
| **[Workflow Documentation](docs/workflows/)** | ⭐ Complete workflow v2.0 system         |
| **[Skills Documentation](docs/skills/)**     | Cross-LLM document creation guide        |
| **[Getting Started](docs/getting-started/)** | Installation, configuration, first steps |
| **[Usage Guides](docs/guides/)**             | Mode-specific guides and best practices  |
| **[MCP Server Mode](docs/mcp-server/)**      | Expose workflows as MCP tools            |
| **[Architecture](docs/architecture/)**       | Technical design and internals           |

### 📖 By Topic

**New Users:**

- [Installation Guide](docs/getting-started/installation.md) - Install and configure MCP-CLI
- [Core Concepts](docs/getting-started/concepts.md) - Understand modes, providers, workflows
- [FAQ](docs/getting-started/faq.md) - Common questions answered

**Usage Guides:**

- [Chat Mode](docs/guides/chat-mode.md) - Interactive conversations with AI
- [Query Mode](docs/guides/query-mode.md) - Single-shot queries for automation
- [Interactive Mode](docs/guides/interactive-mode.md) - Direct MCP tool testing
- [Automation & Scripting](docs/guides/automation.md) - CI/CD integration patterns
- [Debugging](docs/guides/debugging.md) - Troubleshooting and logging

**Workflow Development (v2.0):**

- **[Schema Reference](docs/workflows/SCHEMA.md)** - Complete schema documentation
- **[Authoring Guide](docs/workflows/AUTHORING_GUIDE.md)** - How to write workflows
- **[Loop Guide](docs/workflows/LOOPS.md)** - Iterative execution patterns
- **[Patterns](docs/workflows/patterns/)** - Design patterns (Iterative Refinement, Consensus Validation, Document Pipeline)
- **[Examples](docs/workflows/examples/)** - 13 working examples from simple to advanced
- Working example: `config/workflows/iterative_dev/` - Complete iterative development workflow

**Advanced Topics:**

- [MCP Server Documentation](docs/mcp-server/) - Expose workflows as discoverable tools
- [Architecture Documentation](docs/architecture/) - System design for developers

---

## Architecture

### Workflow Composition

Workflows can call other workflows, creating modular, reusable designs:

```
parent_workflow
  ├─> Calls child_workflow_1 (executes independently)
  │     └─> Returns result
  ├─> Calls child_workflow_2 (executes independently)
  │     └─> Returns result
  └─> Synthesizes results into final output
```

**Benefits:**

- **Modularity**: Reuse workflows across projects
- **Context Efficiency**: Each workflow has isolated context
- **Maintainability**: Update workflows independently
- **Testability**: Test workflows in isolation

### Multi-Provider Workflows

Each step in a workflow can use a different provider:

```yaml
execution:
  provider: anthropic  # Default

steps:
  - name: research
    run: "Research {{input}}"
    # Uses anthropic (inherited)

  - name: analyze
    provider: openai  # Override
    model: gpt-4o
    run: "Analyze {{research}}"

  - name: synthesize
    provider: ollama  # Override
    model: llama3.2
    run: "Synthesize {{analyze}}"
```

### Provider Failover

Automatic failover across provider chains:

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: llama3.2
```

If Anthropic fails → tries OpenAI → falls back to Ollama

### Iterative Loops

LLM-evaluated exit conditions enable agentic workflows:

```yaml
loops:
  - name: improve
    workflow: refinement_cycle
    with:
      content: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "Quality score exceeds 8"  # LLM evaluates
```

Loop continues until LLM determines condition is met!

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
- Share your workflows and patterns

I'm happy to review contributions, though I can't promise active maintenance.

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

**Original Project:** This started as a Go fork of [chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli) in its first few weeks of development. That project is actively maintained by a team of talented developers who have incorporated many new features since February 2025. Check it out and give it a deserved star!

**Model Context Protocol:** Created by Anthropic - [modelcontextprotocol.io](https://modelcontextprotocol.io)

**The Go Community:** For excellent tooling and libraries

---

## Resources

### Documentation

- **[Complete Documentation](docs/README.md)** - Comprehensive guides and references
- **[Workflow Documentation](docs/workflows/)** - ⭐ v2.0 workflows with loops, consensus, composition
- **[Workflow Examples](docs/workflows/examples/)** - 13 working examples from beginner to advanced
- **[Workflow Patterns](docs/workflows/patterns/)** - Design patterns for common use cases
- **[Getting Started](docs/getting-started/)** - Installation and configuration
- **[Usage Guides](docs/guides/)** - Mode-specific tutorials
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

**Built with Go • Powered by MCP • Workflow v2.0**

If this project helps you, please give it a ⭐!

</div>
