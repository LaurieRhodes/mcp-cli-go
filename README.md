# MCP-CLI-Go

> **AI workflows that iterate until perfect, validated by multiple providers, with generative skills for every LLM.**

🔄 **Self-improving workflows** - Loop until LLM evaluates "all tests pass"  
🎯 **Multi-AI consensus** - Require unanimous agreement from 3+ providers  
🎨 **Universal Skills** - GPT-4, DeepSeek, Gemini create PowerPoints, Excel, PDFs (not just Claude)  
🚀 **Provider mixing** - Chain Claude, GPT-4, Llama, and more in one workflow  
📦 **Zero dependencies** - Single Go binary, works offline  

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-blue)](docs/)

```yaml
# Workflow that self-improves until LLM says it's perfect
steps:
  - name: generate
    provider: deepseek
    run: "Write code for: {{input}}"

loops:
  - name: improve
    workflow: code_reviewer
    until: "all tests pass"  # LLM evaluates this, not regex!
    max_iterations: 5
```

**[Quick Start](#quick-start)** • **[Examples](#real-world-examples)** • **[Documentation](docs/)** • **[Download](#installation)**

---

## What becomes possible

### 🔄 Iterative Self-Improvement

Workflows that automatically refine until an LLM decides they're perfect:

```yaml
loops:
  - workflow: code_generator
    until: "code quality exceeds 8/10"  # LLM evaluates quality
    max_iterations: 5
```

**Not string matching.** The LLM actually evaluates whether "all tests pass" or "quality > 8" is true. If not, it loops and tries again.

**Result:** Workflows that improve themselves without human intervention.

### 🎯 Multi-Provider Consensus

Require unanimous agreement from multiple AI providers on critical decisions:

```yaml
steps:
  - name: validate_security
    consensus:
      prompt: "Is this code safe to deploy?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: google
          model: gemini-2.0-flash-exp
      require: unanimous
```

**Use case:** Deploy only when Claude, GPT-4, *and* Gemini all agree it's safe.

### 🎨 Universal Document Creation (Skills)

**The breakthrough:** Anthropic's Skills system works with *every* LLM, not just Claude.

```yaml
execution:
  provider: openai      # GPT-4 can now do this!
  model: gpt-4o
  skills: [pptx, xlsx, docx, pdf]

steps:
  - name: create_presentation
    run: "Create Q4 sales presentation with charts"
```

**Before:** Only Claude  could create PowerPoints, Excel files, Word docs  
**Now:** GPT-4, DeepSeek, Gemini, Llama - any LLM can create documents using the same Anthropic Skills libraries.  Create your own custom skills and processes for all LLMs to follow. 

**How it works:**

1. Skills provide documentation on how to address a task or use libraries (python-pptx, openpyxl, etc.)
2. Any LLM reads the docs and writes Python code
3. Code executes in isolated container
4. File appears on your filesystem

**Result:** Document creation democratized across all AI providers.

### 🔗 Workflow Composition

Build modular AI agent systems by calling workflows from workflows 9no code):

```yaml
steps:
  - name: research
    template:
      name: web_researcher  # Calls another workflow
      with:
        topic: "{{input}}"

  - name: verify
    template:
      name: fact_checker    # Calls another workflow
      with:
        claims: "{{research}}"

  - name: synthesize
    run: "Create report from {{verify}}"
```

**Result:** Reusable workflow components, just like functions in code.

### ⚡ Intelligent Provider Failover

Auto-switch providers when one fails:

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

**Result:** Claude rate-limited? Automatically tries GPT-4. GPT-4 down? Falls back to local Ollama.  Risk protection against model or platform halucination

---

## Example: Multi-Provider Pipeline

```
Input: "Analyze this codebase for security issues"
   ↓
Step 1: Claude analyzes architecture (best at reasoning)
   ↓
Step 2: DeepSeek reviews code (specialized for code)
   ↓
Step 3: GPT-4 fact-checks findings (validation)
   ↓ 
Loop: Iterate until GPT-4 says "analysis complete"
   ↓
Step 4: Create PowerPoint report (Skills)
   ↓
Output: Validated security report + presentation
```

Each provider does what it does best. Automatic iteration. Professional output.

**All from YAML. Zero code.**

---

## Core Capabilities

| Feature                    | What It Enables                                                                                                                                                                                                                                  |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Iterative Loops**        | Workflows that self-improve until LLM-evaluated conditions met                                                                                                                                                                                   |
| **Multi-Provider**         | Chain 12+ providers (Claude, GPT-4, Gemini, Llama, DeepSeek, etc.)                                                                                                                                                                               |
| **Consensus Validation**   | Require agreement from 2-5 providers before proceeding                                                                                                                                                                                           |
| **Universal Skills**       | Every LLM can use [Anthropic Skills](https://github.com/anthropics/skills/tree/main/skills) to create documents (PowerPoint, Excel, Word, PDF) or follow ITIL and ITSM processes.  Create organisational skills to standardise all LLM outcomes. |
| **Workflow Composition**   | Call workflows from workflows for modular design                                                                                                                                                                                                 |
| **Provider Failover**      | Automatic fallback when a provider fails or rate-limits                                                                                                                                                                                          |
| **MCP Server Mode**        | Expose workflows as tools for Claude Desktop or other MCP clients                                                                                                                                                                                |
| **Variable Interpolation** | Pass data between steps with `{{variable}}` syntax                                                                                                                                                                                               |
| **Step Dependencies**      | Control execution order with `needs: [step1, step2]`                                                                                                                                                                                             |
| **Multiple Modes**         | Chat, query, interactive, server - use however you want                                                                                                                                                                                          |

---

## Quick Start

### 1. Install

**Linux/macOS:**

```bash
# Download binary
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64

# Make executable
chmod +x mcp-cli-linux-amd64

# Move to PATH
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
```

**Windows:** Download from [Releases](https://github.com/LaurieRhodes/mcp-cli-go/releases/latest)

### 2. Initialize

```bash
# Quick setup (Ollama only - no API keys needed)
mcp-cli init --quick

# Or full setup with cloud providers
mcp-cli init
```

Add API keys if using cloud providers:

```bash
echo "OPENAI_API_KEY=sk-..." >> .env
echo "ANTHROPIC_API_KEY=sk-ant-..." >> .env
```

### 3. Run Your First Workflow

Create `my_workflow.yaml`:

```yaml
$schema: "workflow/v2.0"
name: analyzer
version: 1.0.0

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
mcp-cli --workflow analyzer --input-data "Your text here"
```

**That's it.** No code. No setup. Just YAML and go.

---

## Real-World Examples

### Document Intelligence Pipeline

Analyze documents across multiple providers with iterative refinement:

```yaml
execution:
  provider: anthropic
  skills: [docx, pdf]

steps:
  - name: extract
    run: "Extract key data from {{input}}"

  - name: analyze
    needs: [extract]
    provider: openai
    run: "Analyze: {{extract}}"

loops:
  - name: refine
    workflow: report_generator
    with:
      analysis: "{{analyze}}"
    until: "report quality exceeds 8/10"
    max_iterations: 3
```

### Multi-Provider Security Review

Get consensus from multiple AIs before deployment:

```yaml
execution:
  provider: anthropic

steps:
  - name: review
    consensus:
      prompt: "Review this code for security issues: {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: google
          model: gemini-2.0-flash-exp
      require: unanimous

  - name: report
    needs: [review]
    skills: [docx]
    run: "Create security report: {{review}}"
```

### Code Development Loop

Automatically improve code until tests pass:

```yaml
execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: requirements
    run: "Define requirements: {{input}}"

loops:
  - name: develop
    workflow: code_and_test
    with:
      requirements: "{{requirements}}"
      previous: "{{loop.last.output}}"
    until: "all tests pass"
    max_iterations: 5
```

**📚 More examples:** [docs/workflows/examples/](docs/workflows/examples/) - 13 working examples

---

## Universal Skills: Document Creation for Every LLM

### The Problem

Anthropic's Skills enable document creation (PowerPoint, Excel, Word), but they only work with Claude. If you use GPT-4, Gemini, or local models, you can't create documents.

### The Solution

MCP-CLI brings Skills to **every LLM** through containerized execution:

```yaml
execution:
  provider: openai       # ← Not Claude!
  model: gpt-4o
  servers: [filesystem]
  skills: [pptx, xlsx]   # ← Skills work anyway!

steps:
  - name: create
    run: "Create Q4 sales presentation with charts"
```

### How It Works

```
┌─────────────┐
│   Any LLM   │  "Create PowerPoint"
│ GPT-4/Gemini│
└──────┬──────┘
       │
       ↓
┌─────────────┐
│   Skills    │  Provides documentation:
│Documentation│  "Use python-pptx like this..."
└──────┬──────┘
       │
       ↓
┌─────────────┐
│ LLM Writes  │  Generates Python code
│   Python    │  using python-pptx
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  Container  │  Executes code in
│  Execution  │  isolated environment
└──────┬──────┘
       │
       ↓
┌─────────────┐
│    File     │  presentation.pptx
│   Output    │  appears on disk
└─────────────┘
```

### Security

All code runs in isolated Docker containers with:

- No network access
- Read-only root filesystem
- Memory and CPU limits
- Automatic cleanup

**📚 Complete Skills guide:** [docs/skills/](docs/skills/)

---

## MCP Server Mode

Expose workflows as discoverable tools for Claude Desktop or other MCP clients.

### 1. Create Server Config

`research_server.yaml`:

```yaml
server_info:
  name: research-agent
  version: 1.0.0

tools:
  - name: research_topic
    description: Research a topic with web search
    workflow: deep_research
    input_schema:
      type: object
      properties:
        topic:
          type: string
      required: [topic]
```

### 2. Start Server

```bash
mcp-cli serve research_server.yaml
```

### 3. Configure Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/usr/local/bin/mcp-cli",
      "args": ["serve", "/path/to/research_server.yaml"]
    }
  }
}
```

Now Claude can use your workflows as tools!

**📚 Full MCP Server guide:** [docs/mcp-server/](docs/mcp-server/)

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
mcp-cli query "Your question"

# With workflow
mcp-cli --workflow analyzer --input-data "Your data"
```

### Interactive Mode

```bash
# Test MCP server tools directly
mcp-cli interactive
mcp-cli tools
```

### Server Mode

```bash
# Run as MCP server
mcp-cli serve your_server.yaml
```

---

## Supported Providers

| Provider          | Models                           | Notes        |
| ----------------- | -------------------------------- | ------------ |
| **Anthropic**     | Claude Opus, Sonnet, Haiku       | Full support |
| **OpenAI**        | GPT-4o, GPT-4o-mini, o1, o3-mini | Full support |
| **Google**        | Gemini 2.0 Flash, Pro            | Full support |
| **DeepSeek**      | DeepSeek Chat, Coder             | Full support |
| **Ollama**        | Llama, Qwen, Mistral, etc.       | Local models |
| **OpenRouter**    | 100+ models                      | Aggregator   |
| **Kimi/Moonshot** | K2                               | Full support |
| **LM Studio**     | Any local model                  | Local server |
| **AWS Bedrock**   | Claude, Llama, Mistral           | Enterprise   |
| **Azure Foundry** | GPT-4, others                    | Enterprise   |
| **Google Vertex** | Gemini, Claude                   | Enterprise   |

**Total:** 10+ providers, 100+ models

---

## Documentation

### 📚 Complete Guides

| Documentation                                     | Description                                |
| ------------------------------------------------- | ------------------------------------------ |
| **[Getting Started](docs/getting-started/)**      | Installation, configuration, first steps   |
| **[Workflow Guide](docs/workflows/)**             | ⭐ Complete workflow v2.0 system            |
| **[Workflow Examples](docs/workflows/examples/)** | 13 working examples (beginner to advanced) |
| **[Skills Guide](docs/skills/)**                  | Cross-LLM document creation                |
| **[MCP Server Mode](docs/mcp-server/)**           | Expose workflows as tools                  |
| **[Usage Guides](docs/guides/)**                  | Mode-specific tutorials                    |
| **[Architecture](docs/architecture/)**            | Technical design                           |

### 🎯 Quick Links

- **[Workflow Schema](docs/workflows/schema/)** - YAML reference
- **[Loop Guide](docs/workflows/LOOPS.md)** - Iterative execution
- **[Patterns](docs/workflows/patterns/)** - Design patterns
- **[FAQ](docs/getting-started/faq.md)** - Common questions

---

## Configuration

### Provider Setup

`config/providers/openai.yaml`:

```yaml
provider_name: openai
models:
  - name: gpt-4o
    max_tokens: 4096
api_key: ${OPENAI_API_KEY}
```

### Workflow Structure

```yaml
$schema: "workflow/v2.0"
name: my_workflow
version: 1.0.0

# Defaults for all steps
execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  skills: [docx, xlsx]

# Sequential steps
steps:
  - name: analyze
    run: "Analyze: {{input}}"

  - name: report
    needs: [analyze]
    run: "Create report: {{analyze}}"

# Iterative loops
loops:
  - name: improve
    workflow: refinement
    until: "quality > 8"
    max_iterations: 5
```

---

## Why MCP-CLI?

### vs. Direct API Calls

| Direct API                   | MCP-CLI                  |
| ---------------------------- | ------------------------ |
| Single provider              | Mix 10+ providers        |
| One-shot response            | Iterative refinement     |
| No multi-provider validation | Multi-provider consensus |
| No Skills system             | Skills for any LLM       |
| Manual chaining              | Automatic workflows      |

*MCP-CLI also supports MCPO proxy format for exposing Skills as MCP tools for other AI systems!*

### vs. LangChain/LlamaIndex

| LangChain/LlamaIndex      | MCP-CLI               |
| ------------------------- | --------------------- |
| Requires Python coding    | YAML configuration    |
| Complex setup (pip, deps) | Single binary         |
| Python framework          | Standalone binary     |
| Code-based agents         | Declarative workflows |
| Requires Python runtime   | Any environment       |

### vs. Claude Desktop

| Claude Desktop     | MCP-CLI                  |
| ------------------ | ------------------------ |
| Interactive only   | Automation + interactive |
| Claude only        | 10+ providers            |
| No workflows       | Multi-step workflows     |
| Manual interaction | CLI automation           |

**When to use MCP-CLI:**

- Automate multi-step AI tasks
- Need provider redundancy
- Want iterative refinement
- Require multi-AI validation
- Need Skills-based document creation across LLMs
- Building AI agent systems

---

## Command Reference

```bash
# Initialize
mcp-cli init
mcp-cli init --quick

# Query
mcp-cli query "question"
mcp-cli query --provider openai --model gpt-4o "question"

# Workflows
mcp-cli --workflow <name> --input-data "..."
mcp-cli --list-workflows

# Chat
mcp-cli chat
mcp-cli chat --provider anthropic

# Server
mcp-cli serve config.yaml

# Tools
mcp-cli tools
mcp-cli interactive

# Help
mcp-cli --help
```

---

## Project Background

This project started as a fork of [chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli) in February 2025 for Go-based MCP server development. That project continues to grow with very talented contributors.  Go and check it out!

I built this for my own automation needs and shared it as example code. If you find it useful, great! 

### Why Go?

- **Single binary** - No runtime dependencies
- **Cross-platform** - Linux, macOS, Windows
- **Fast startup** - Ideal for CLI tools
- **Easy deployment** - Just copy the binary

---

## Installation

### Pre-Built Binaries

**Linux:**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-linux-amd64
chmod +x mcp-cli-linux-amd64
sudo mv mcp-cli-linux-amd64 /usr/local/bin/mcp-cli
```

**macOS (Intel):**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-amd64
chmod +x mcp-cli-darwin-amd64
sudo mv mcp-cli-darwin-amd64 /usr/local/bin/mcp-cli
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/LaurieRhodes/mcp-cli-go/releases/latest/download/mcp-cli-darwin-arm64
chmod +x mcp-cli-darwin-arm64
sudo mv mcp-cli-darwin-arm64 /usr/local/bin/mcp-cli
```

**Windows:** Download from [Releases](https://github.com/LaurieRhodes/mcp-cli-go/releases/latest)

### Build from Source

```bash
git clone https://github.com/LaurieRhodes/mcp-cli-go.git
cd mcp-cli-go
go build -o mcp-cli
sudo mv mcp-cli /usr/local/bin/
```

---

## Contributing

This project is shared as example code. Feel free to:

- Fork and modify
- Open issues for bugs
- Submit PRs for fixes
- Share your workflows

---

## License

MIT License - see [LICENSE](LICENSE)

---

## Acknowledgments

- **Original Project:** [chrishayuk/mcp-cli](https://github.com/chrishayuk/mcp-cli) - Actively maintained with talented developers
- **Model Context Protocol:** Created by Anthropic - [modelcontextprotocol.io](https://modelcontextprotocol.io)
- **The Go Community:** For excellent tooling

---

## Resources

- **[Documentation](docs/README.md)** - Complete guides and references
- **[Workflow Examples](docs/workflows/examples/)** - 13 working examples
- **[Source Code](https://github.com/LaurieRhodes/mcp-cli-go)** - GitHub repository
- **[Releases](https://github.com/LaurieRhodes/mcp-cli-go/releases)** - Download binaries
- **[MCP Protocol](https://modelcontextprotocol.io)** - Official specification

---

<div align="center">

**Self-Improving Workflows • Multi-Provider • Universal Skills**

If this helps you, please give it a ⭐!

</div>
