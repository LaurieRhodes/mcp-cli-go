# MCP-CLI Command Reference

Complete reference for all mcp-cli commands, flags, and usage patterns.

## Table of Contents

- [Global Flags](#global-flags)
- [Commands](#commands)
  - [Chat Mode](#chat-mode)
  - [Query Mode](#query-mode)
  - [Interactive Mode](#interactive-mode)
  - [Workflow Templates](#workflow-templates)
  - [Serve Mode](#serve-mode)
  - [Embeddings](#embeddings)
  - [Configuration](#configuration)
  - [Init](#init)

---

## Global Flags

These flags work with all commands:

| Flag                   | Short | Default        | Description                            |
| ---------------------- | ----- | -------------- | -------------------------------------- |
| `--config`             | -     | `config.yaml`  | Path to configuration file (YAML/JSON) |
| `--server`             | `-s`  | All configured | MCP server(s) to use (comma-separated) |
| `--provider`           | `-p`  | From config    | AI provider to use                     |
| `--model`              | `-m`  | From config    | Model to use                           |
| `--disable-filesystem` | -     | `false`        | Disable filesystem server              |
| `--verbose`            | `-v`  | `false`        | Enable verbose logging                 |
| `--no-color`           | -     | `false`        | Disable colored output                 |

### Provider Options

Supported providers (set with `--provider` or `-p`):

- `openai` - OpenAI API
- `anthropic` - Anthropic Claude API
- `ollama` - Local Ollama instance
- `deepseek` - DeepSeek API
- `gemini` - Google Gemini API
- `openrouter` - OpenRouter API
- `lmstudio` - LM Studio local server

---

## Commands

### Chat Mode

**Default command** - Interactive conversation with the AI.

```bash
# Start chat (default)
mcp-cli

# Explicit chat command
mcp-cli chat

# With specific provider and model
mcp-cli chat --provider openai --model gpt-4o

# With specific servers
mcp-cli chat --server filesystem,brave-search

# Without filesystem access
mcp-cli chat --disable-filesystem
```

**Features:**

- Multi-turn conversation
- Automatic tool execution
- Conversation history
- Markdown rendering
- Syntax highlighting for code

---

### Query Mode

Single question and response - ideal for scripting and automation.

```bash
mcp-cli query [flags] "question"
```

**Flags:**

- `--json`, `-j` - Output response in JSON format
- `--context`, `-c` - File containing additional context
- `--system-prompt` - Custom system prompt
- `--max-tokens` - Maximum tokens in response
- `--output`, `-o` - Output file path
- `--noisy`, `-n` - Show detailed logs
- `--raw-data` - Output raw tool data instead of AI summary
- `--error-code-only` - Only return error codes

**Examples:**

```bash
# Basic query
mcp-cli query "What is 2+2?"

# With context file
mcp-cli query --context data.txt "Analyze this data"

# JSON output for parsing
mcp-cli query --json "List cloud providers" > results.json

# Raw tool data (bypass AI summarization)
mcp-cli query --raw-data "Show latest security incidents"

# Verbose output
mcp-cli query --noisy "What files are here?"

# Save to file
mcp-cli query "Analyze code" --output analysis.txt

# With specific servers
mcp-cli query --server filesystem,brave-search \
  "Search for MCP information and save to file"
```

**Exit Codes:**

- `0` - Success
- `1` - General error
- `2` - Configuration not found
- `3` - Provider not found
- `4` - Context file not found
- `5` - Initialization error
- `6` - LLM request error
- `7` - Tool execution error
- `8` - Server connection error
- `9` - Output format error
- `10` - Output write error

---

### Interactive Mode

Interactive shell with slash commands.

```bash
mcp-cli interactive
```

**Slash Commands:**

- `/help` - Show help
- `/exit` or `/quit` - Exit interactive mode
- `/clear` - Clear conversation history
- `/servers` - List available servers
- `/tools` - List available tools
- `/history` - Show conversation history
- `/config` - Show current configuration

**Example:**

```bash
$ mcp-cli interactive
> /tools
> What files are in my current directory?
> /clear
> Analyze the project structure
> /exit
```

---

### Workflow Templates

Execute multi-step workflows with data passing between steps.

```bash
# List available templates
mcp-cli --list-templates

# Execute template
mcp-cli --template <name>

# With input data
mcp-cli --template <name> --input-data "data"

# From stdin
echo "data" | mcp-cli --template <name>
```

**Template Flags:**

- `--template` - Template name to execute
- `--input-data` - Input data (JSON or plain text)
- `--list-templates` - List all available templates

**Examples:**

```bash
# List templates
mcp-cli --list-templates

# Run analysis template
mcp-cli --template analyze_file

# With input data
mcp-cli --template data_processing --input-data '{"key":"value"}'

# From stdin
cat data.txt | mcp-cli --template summarize

# With specific provider
mcp-cli --template research --provider anthropic --model claude-sonnet-4
```

**Template Structure:**

Templates are defined in `config/templates/*.yaml`:

```yaml
name: analyze
description: Analyze input data
version: "2.0"
steps:
  - step: 1
    name: analyze
    base_prompt: "Analyze this data: {{input_data}}"
    providers: [openai]
    models: [gpt-4o]
```

---

### Serve Mode

Run mcp-cli as an MCP server, exposing templates as tools.

```bash
mcp-cli serve [runas-config]
```

**Flags:**

- `--serve` - Path to runas config file

**Examples:**

```bash
# Start MCP server
mcp-cli serve config/runas/research_agent.yaml

# With verbose logging
mcp-cli serve --verbose config/runas/agent.yaml

# Using --serve flag
mcp-cli --serve config/runas/agent.yaml
```

**Runas Config Example:**

`config/runas/research_agent.yaml`:

```yaml
server_info:
  name: research-agent
  version: "1.0.0"
  description: Research assistant with web search

tools:
  - name: research_topic
    description: Research a topic using web search
    template: research
    input_schema:
      type: object
      properties:
        topic:
          type: string
          description: Topic to research
      required: [topic]
    input_mapping:
      topic: "{{input_data}}"
```

**Claude Desktop Integration:**

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/config/runas/research_agent.yaml"]
    }
  }
}
```

---

### Embeddings

Generate vector embeddings for text.

```bash
mcp-cli embeddings [text]
```

**Flags:**

- `--provider` - Embedding provider
- `--model` - Embedding model
- `--input-file` - Input file path
- `--output-file` - Output file path
- `--output-format` - Output format (json, csv, compact)
- `--chunk-strategy` - Chunking strategy (sentence, paragraph, fixed)
- `--max-chunk-size` - Maximum chunk size in tokens
- `--overlap` - Overlap between chunks in tokens
- `--encoding-format` - Encoding format (float, base64)
- `--dimensions` - Number of dimensions
- `--include-metadata` - Include chunk metadata
- `--show-models` - Show available models
- `--show-strategies` - Show chunking strategies

**Examples:**

```bash
# Basic embedding
mcp-cli embeddings "Your text here"

# From file
mcp-cli embeddings --input-file document.txt

# With specific model
mcp-cli embeddings --model text-embedding-3-large --input-file doc.txt

# Advanced chunking
mcp-cli embeddings --chunk-strategy sentence \
  --max-chunk-size 512 --overlap 50 \
  --input-file large-doc.txt

# CSV output
mcp-cli embeddings --output-format csv \
  --input-file doc.txt --output-file embeddings.csv

# From stdin
echo "Text to embed" | mcp-cli embeddings

# Show available models
mcp-cli embeddings --show-models

# Show chunking strategies
mcp-cli embeddings --show-strategies
```

---

### Configuration

Manage configuration files.

```bash
mcp-cli config [command]
```

**Commands:**

- `validate` - Validate configuration file
- `check` - Check configuration (alias for validate)

**Examples:**

```bash
# Validate config
mcp-cli config validate

# Validate specific file
mcp-cli config validate --config custom.yaml

# Check config
mcp-cli config check
```

---

### Init

Initialize mcp-cli configuration.

```bash
mcp-cli init [flags]
```

**Flags:**

- `--quick` - Quick setup (30 seconds)
- `--full` - Complete setup with all options

**Examples:**

```bash
# Interactive setup
mcp-cli init

# Quick setup
mcp-cli init --quick

# Full setup
mcp-cli init --full
```

**What Gets Created:**

- `config.yaml` - Main configuration
- `config/providers/*.yaml` - Provider configurations
- `config/servers/*.yaml` - MCP server configurations
- `.env` - Environment variables (API keys)

---

## Common Patterns

### Scripting with Query Mode

```bash
#!/bin/bash

# Get analysis
RESULT=$(mcp-cli query --json "Analyze project" | jq -r '.response')

# Use in script
echo "Analysis: $RESULT"
```

### Piping Data

```bash
# Analyze file
cat data.txt | mcp-cli query "Analyze this data"

# Generate embeddings
cat document.txt | mcp-cli embeddings --output-file embeddings.json
```

### Multiple Providers

```bash
# Use OpenAI for chat
mcp-cli chat --provider openai --model gpt-4o

# Use Anthropic for query
mcp-cli query --provider anthropic --model claude-sonnet-4 "Question"

# Use Ollama for templates
mcp-cli --template analyze --provider ollama --model qwen2.5:32b
```

### Server Selection

```bash
# Use specific servers
mcp-cli chat --server filesystem,brave-search

# Disable filesystem
mcp-cli chat --disable-filesystem

# Query with specific servers
mcp-cli query --server brave-search "Search for latest news"
```

---

## Environment Variables

Set API keys in `.env` file:

```bash
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
DEEPSEEK_API_KEY=...
GEMINI_API_KEY=...
OPENROUTER_API_KEY=...
```

Or export directly:

```bash
export OPENAI_API_KEY=sk-...
mcp-cli query "Hello"
```

---

## Configuration Files

### Main Config: `config.yaml`

```yaml
includes:
  providers: config/providers/*.yaml
  servers: config/servers/*.yaml
  templates: config/templates/*.yaml

ai:
  default_provider: ollama
  default_system_prompt: You are a helpful assistant.
```

### Provider Config: `config/providers/openai.yaml`

```yaml
interface_type: openai_compatible
provider_name: openai
config:
  api_endpoint: https://api.openai.com/v1
  api_key: ${OPENAI_API_KEY}
  default_model: gpt-4o-mini
  timeout_seconds: 300
```

### Server Config: `config/servers/filesystem.yaml`

```yaml
server_name: filesystem
config:
  command: /path/to/filesystem-server
  args: []
```

---

## Tips & Tricks

### Debugging

```bash
# Verbose output
mcp-cli --verbose query "test"

# Show raw tool data
mcp-cli query --raw-data "get data"

# Noisy mode
mcp-cli query --noisy "test"
```

### Output Control

```bash
# No colors (for logging)
mcp-cli --no-color query "test"

# JSON output
mcp-cli query --json "test" > output.json

# Save to file
mcp-cli query "analyze" --output result.txt
```

### Performance

```bash
# Use local Ollama for speed
mcp-cli --provider ollama --model qwen2.5:32b

# Limit servers for faster startup
mcp-cli --server filesystem chat

# Disable filesystem if not needed
mcp-cli --disable-filesystem query "2+2"
```

---

## Getting Help

```bash
# Main help
mcp-cli --help

# Command help
mcp-cli query --help
mcp-cli chat --help
mcp-cli embeddings --help

# Configuration help
mcp-cli config --help
```

---

## See Also

- [SECURITY.md](SECURITY.md) - API key security
- [config/README.md](config/README.md) - Configuration guide
- [examples/](examples/) - Example configurations
