# MCP-CLI-Go Documentation

Complete documentation for MCP-CLI-Go: A powerful command-line interface for AI model interactions using the Model Context Protocol.

---

## Quick Navigation

### For New Users

- [Installation & Setup](#getting-started) - Install and configure MCP-CLI
- [Core Concepts](#core-concepts) - Understand modes, providers, and templates
- [Quick Start Examples](#quick-start-examples) - Get running in 5 minutes

### For Developers

- [Templates](#templates) - Create reusable AI workflows
- [Automation](#automation) - Script and integrate MCP-CLI
- [Architecture](#architecture) - Technical design and internals

### For Advanced Users

- [MCP Server Mode](#mcp-server-mode) - Expose workflows as discoverable tools
- [HTTP Proxy Server](#http-proxy-server) - Convert MCP servers to REST APIs
- [Skills System](#skills-system) - Anthropic-compatible skills with auto-loading
- [Debugging](#debugging) - Troubleshoot and optimize

---

## Documentation Structure

### Getting Started

> **Installation, configuration, and first steps**

*Documentation planned - see project README for basic installation*

**Topics:**

- Installation (binary, build from source, package managers)
- Initial configuration (config.yaml setup)
- Provider configuration (OpenAI, Anthropic, Ollama, etc.)
- First query and chat session

---

### [Guides](guides/)

> **Task-oriented guides for using MCP-CLI effectively**

#### Operational Modes

- **[Chat Mode](guides/chat-mode.md)** - Interactive conversations with AI
  
  - Conversation history and context
  - Tool integration
  - Commands and shortcuts
  - Use cases: Research, coding assistance, general Q&A

- **[Query Mode](guides/query-mode.md)** - Single-shot queries for automation
  
  - Scripting and piping
  - Output formatting (text, JSON)
  - Error handling
  - Use cases: CI/CD, data processing, batch operations

- **[Interactive Mode](guides/interactive-mode.md)** - Direct MCP tool execution
  
  - Tool discovery and inspection
  - Manual tool calling
  - Testing MCP servers
  - Use cases: Tool development, debugging, exploration

- **[Embeddings](guides/embeddings.md)** - Vector embeddings for semantic search
  
  - Generating embeddings
  - Similarity matching
  - Integration patterns
  - Use cases: Semantic search, content recommendation, clustering

#### Advanced Topics

- **[Automation & Scripting](guides/automation.md)** - Integrate MCP-CLI into workflows
  
  - Shell scripts and piping
  - CI/CD integration (GitHub Actions, GitLab CI)
  - Error handling and retries
  - Environment management

- **[Debugging](guides/debugging.md)** - Troubleshoot and optimize
  
  - Verbose logging
  - Common issues and solutions
  - Performance optimization
  - Provider-specific debugging

---

### [Workflows](workflows/)

> **Create reusable, multi-step AI workflows**

- **[Authoring Guide](workflows/AUTHORING_GUIDE.md)** - Complete template creation reference
  
  - Template structure and syntax
  - Variable substitution
  - Step types and execution flow
  - Conditional logic and loops
  - Provider selection per step
  - Template composition patterns

- **[Examples](workflows/examples/)** - Real, working templates
  
  - Code review workflows
  - Research and fact-checking
  - Data analysis pipelines
  - Document generation
  - Content transformation

**What are templates?**

Templates are YAML-defined, multi-step AI workflows that solve specific problems. They enable:

- **Reusability** - Define once, use everywhere
- **Consistency** - Same process every time
- **Complexity** - Multi-step reasoning and analysis
- **Flexibility** - Different providers for different steps

---

### MCP Server Mode

> **Expose templates as discoverable tools for any MCP client**

*Documentation planned*

**What is Server Mode?**

Server mode transforms your AI workflows (templates) into tools that can be discovered and used by any MCP-compatible application (Claude Desktop, Cursor IDE, custom applications).

**Planned Documentation:**

- Server setup and configuration
- runas configuration reference (mapping templates to tools)
- Tool parameter mapping
- Integration patterns (Claude Desktop, Cursor, custom clients)
- Production deployment strategies
- Complete working examples

---

### [HTTP Proxy Server](proxy/)

> **Convert MCP servers to REST APIs**

Expose any MCP server as a production-ready REST API with auto-generated OpenAPI documentation.

- **[Server Guide](proxy/proxy-server-guide.md)** - Complete HTTP proxy documentation
  - Auto-discovery from MCP servers
  - OpenAPI 3.0 specification generation
  - Swagger UI integration
  - API key authentication
  - CORS support
  - TLS/HTTPS configuration
  - Production deployment
  - OpenWebUI integration

- **[Quick Reference](proxy/proxy-quick-reference.md)** - Fast lookup and common commands
  - 2-minute setup
  - Configuration examples
  - Testing commands
  - Troubleshooting tips

**What is the HTTP Proxy?**

The HTTP proxy converts MCP servers (stdio/SSE protocol) into standard REST APIs, enabling:

- **Web integration** - Use MCP tools from web applications
- **Tool aggregation** - Single REST API for multiple MCP servers (e.g., OpenWebUI)
- **Legacy systems** - Integrate with systems that only support REST
- **Standard docs** - Auto-generated OpenAPI specs and Swagger UI

**Example:**
```yaml
# config/proxy/bash.yaml
runas_type: proxy
config_source: config/servers/bash.yaml
proxy_config:
  port: 4000
  api_key: "${API_KEY}"
```

Instantly creates REST API:
- `POST /bash` - Execute bash commands
- `GET /docs` - Swagger UI  
- `GET /openapi.json` - OpenAPI spec

✅ **Production Status:** Fully implemented and tested with OpenWebUI

---

### [Skills System](skills/)

> **Anthropic-compatible skills with auto-loading**

Auto-discover and expose Anthropic Skills as MCP tools with zero configuration.

- **[Auto-Loading Guide](skills/auto-loading.md)** - Complete skills auto-discovery guide
  - How auto-loading works
  - Configuration reference
  - Directory structure requirements
  - Troubleshooting guide

- **[Quick Reference](skills/quick-reference.md)** - Fast lookup
  - 2-minute setup
  - Common configurations
  - Debugging commands

- **[Quick Start](skills/quick-start.md)** - Get started in 5 minutes
  - Initial setup
  - Testing skills
  - Claude Desktop integration

- **[Docker/Podman Execution](skills/docker-podman-execution.md)** - Code execution guide
  - Containerized sandbox execution
  - Security boundaries
  - Installation and setup
  - Performance tuning

**What are Skills?**

Skills are modular packages that extend LLM capabilities with specialized knowledge and helper libraries:

- **Documentation** (`SKILL.md`) - Guidance for Claude
- **Scripts** (`scripts/`) - Python helper libraries
- **References** (`references/`) - Additional docs loaded on demand

**Auto-Loading:**
```yaml
# config/runasMCP/mcp_skills_stdio.yaml
runas_type: mcp-skills  # That's it!
```

All skills in `config/skills/` automatically become MCP tools.

✅ **Production Status:** Skills auto-discovery and `execute_skill_code` fully implemented

---

### [Architecture](ARCHITECTURE.md)

> **Technical design, internals, and extension points**

Comprehensive technical documentation for developers and contributors.

**Covers:**

- System architecture and layer responsibilities
- Configuration hierarchy and data flow patterns
- Security, performance, and testing architecture
- Deployment strategies and extension points
- Concurrency model and error handling
- Code examples and implementation patterns

**Read this to:**

- Understand the internal design
- Contribute to the codebase
- Extend with new providers or modes
- Integrate MCP-CLI into your systems

---

## Core Concepts

### Operational Modes

MCP-CLI operates in several modes, each optimized for different use cases:

| Mode            | Purpose                   | Use When                               |
| --------------- | ------------------------- | -------------------------------------- |
| **Chat**        | Interactive conversations | Research, exploration, assistance      |
| **Query**       | Single-shot queries       | Automation, scripting, pipelines       |
| **Interactive** | Direct tool testing       | Tool development, debugging            |
| **Template**    | Multi-step workflows      | Complex analysis, consistent processes |
| **Server**      | Expose tools to clients   | IDE integration, team sharing          |

### Providers

MCP-CLI supports multiple AI providers with a unified interface:

- **OpenAI** - GPT-4, GPT-4 Turbo, GPT-3.5
- **Anthropic** - Claude 3 and Claude 4 family (Opus, Sonnet, Haiku)
- **Google** - Gemini models
- **Ollama** - Local models (Llama, Mistral, Qwen, etc.)
- **DeepSeek** - DeepSeek models
- **OpenRouter** - Multi-provider gateway

**Provider Selection:**

- Per-command via `--provider` flag
- Per-template via config section
- Global default in config.yaml
- Automatic fallback to local models

### MCP (Model Context Protocol)

MCP enables AI models to interact with external tools and data sources:

- **Tool Discovery** - Automatic detection of available tools
- **Tool Execution** - AI decides when and how to use tools
- **Data Access** - Read files, query databases, search the web
- **Extensibility** - Any application can provide MCP tools

**MCP Servers:**

- Filesystem operations
- Web search (Brave, Google)
- Database access
- Custom tools (build your own)

---

## Quick Start Examples

### Simple Query

```bash
mcp-cli query "Explain quantum computing in simple terms"
```

### Chat Session

```bash
mcp-cli chat
> What are the best practices for API design?
> Can you review this OpenAPI spec? [paste spec]
```

### Workflow Execution

```bash
mcp-cli --template code_review \
  --input-data '{"code": "def hello(): print(\"hi\")", "language": "python"}'
```

### Automation

```bash
# In CI/CD pipeline
cat pull_request.diff | mcp-cli query "Review this code change" > review.txt
```

---

## Documentation Conventions

### Structure

All documentation follows consistent patterns:

- **Purpose statement** - What this document covers
- **Quick examples** - Get started immediately  
- **Detailed explanations** - Deep understanding
- **Real-world patterns** - Practical applications
- **Troubleshooting** - Common issues and solutions

### Code Examples

- All examples are **tested and working**
- **Copy-paste ready** - Run directly without modification
- **Real-world** - Not contrived or oversimplified
- **Explained** - Why, not just what

### Technical Level

Documentation is written for:

- **Getting Started** - No prior knowledge assumed
- **Guides** - Basic familiarity with MCP-CLI
- **Templates** - Understanding of AI workflows
- **Server Mode** - Technical competence assumed
- **Architecture** - Developer-level detail

---

## Contributing to Documentation

### Improvements Welcome

Found an issue? Have a suggestion?

- **Typos and errors** - Submit a PR
- **Missing examples** - Share your use cases
- **Unclear sections** - Open an issue
- **New guides** - Propose topics

### Documentation Standards

When contributing:

1. **Be specific** - Avoid vague descriptions
2. **Show, don't tell** - Include working examples
3. **Stay current** - Verify against latest release
4. **Be concise** - Respect reader's time
5. **Test examples** - Ensure code works

---

## Getting Help

### Documentation Issues

- **Can't find what you need?** - Open an issue requesting documentation
- **Example doesn't work?** - Report it with details
- **Concept unclear?** - Ask for clarification

### Technical Support

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Questions and community help
- **Examples** - Check [workflows/examples](workflows/examples/) first

---

## Version Information

This documentation is for **MCP-CLI-Go**.

**Last Updated:** December 2024

**Documentation Version:** See individual files for updates

**Note:** Some documentation sections are planned but not yet complete. Check back for updates or contribute!

---

## Quick Reference

### Essential Commands

```bash
# Query mode
mcp-cli query "your question"

# Chat mode  
mcp-cli chat

# With specific provider
mcp-cli query --provider anthropic "question"

# With MCP server
mcp-cli chat --server filesystem

# Workflow execution
mcp-cli --template template_name

# List templates
mcp-cli --list-templates

# Interactive mode
mcp-cli interactive

# Server mode
mcp-cli serve config/runas/server.yaml
```

### Configuration Files

- **Main config:** `config.yaml`
- **Providers:** `config/providers/*.yaml`
- **Templates:** `config/workflows/*.yaml`
- **MCP servers:** `config/mcp/*.yaml`
- **Server configs:** `config/runas/*.yaml`

### Environment Variables

```bash
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...
MCP_CLI_CONFIG=/path/to/config.yaml
```

---

**Ready to dive in?** Start with [Chat Mode](guides/chat-mode.md) or explore [Templates](workflows/AUTHORING_GUIDE.md).
