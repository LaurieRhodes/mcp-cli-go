# MCP CLI Modular Configuration Examples

This directory contains example configuration files for MCP CLI.

## Structure

```
mcp-cli/                    # Project root
├── config.yaml             # Main config (at executable level)
└── config/                 # Configuration directory
    ├── README.md           # Documentation
    ├── settings.yaml       # Application settings
    ├── providers/          # LLM provider configurations
    ├── embeddings/         # Embedding provider configurations
    ├── servers/            # MCP server configurations
    ├── templates/          # Workflow templates
    ├── runasMCP/          # MCP stdio server configs (Claude Desktop)
    ├── proxy/             # HTTP proxy configurations
    └── skills/            # Skills configuration
```

## Configuration Files

### Main Config (config.yaml)

Located at project root, uses includes to load modular configs:

```yaml
includes:
  providers: config/providers/*.yaml
  embeddings: config/embeddings/*.yaml
  servers: config/servers/*.yaml
  runas: config/runasMCP/*.yaml
  templates: config/templates/*.yaml
  settings: config/settings.yaml
```

### Application Settings (settings.yaml)

Global application configuration:

```yaml
ai:
  default_provider: deepseek

chat:
  default_temperature: 0.7
  max_history_size: 50
  chat_logs_location: ""

skills:
  outputs_dir: "/tmp/mcp-outputs"

logging:
  format: text
  level: info
```

## Directory Contents

### providers/

LLM provider configurations. Each file defines connection details for an AI provider:

- `anthropic.yaml` - Anthropic Claude
- `openai.yaml` - OpenAI GPT models
- `ollama.yaml` - Local Ollama models
- `deepseek.yaml` - DeepSeek models
- `gemini.yaml` - Google Gemini
- `kimik2.yaml` - Moonshot Kimi
- `openrouter.yaml` - OpenRouter proxy
- `lmstudio.yaml` - Local LM Studio
- `aws-bedrock.yaml` - AWS Bedrock
- `azure-foundry.yaml` - Azure AI Foundry
- `gcp-vertex-ai.yaml` - Google Cloud Vertex AI

### embeddings/

Embedding provider configurations for vector operations:

- `openai.yaml` - OpenAI embeddings
- `ollama.yaml` - Ollama embeddings
- `openrouter.yaml` - OpenRouter embeddings
- `aws-bedrock.yaml` - AWS Bedrock embeddings
- `azure-foundry.yaml` - Azure embeddings

### servers/

MCP server configurations for external tools:

- `bash.yaml` - Bash command execution
- `filesystem.yaml` - File system operations
- `search.yaml` - Web search capabilities

### templates/

Workflow template definitions for multi-step AI operations.

### runasMCP/

Configurations for running mcp-cli as MCP stdio servers (primarily for Claude Desktop integration):

- `mcp_skills_stdio.yaml` - Skills server
- `document_intelligence_agent.yaml` - Document analysis agent
- `research_agent.yaml` - Research assistant
- `simple_analysis.yaml` - Simple analysis workflows

### proxy/

HTTP proxy configurations for exposing MCP servers as REST APIs:

- `bash.yaml` - Bash proxy
- `filesystem.yaml` - Filesystem proxy
- `skills.yaml` - Skills proxy
- Various workflow proxies

### skills/

Skills system configuration:

- `skill-images.yaml` - Container image mappings for skills
- `README.md` - Skills documentation

## Getting Started

1. **Copy examples to config directory:**
   ```bash
   cp -r examples/config/* config/
   ```

2. **Set API keys in `.env`:**
   ```bash
   echo "OPENAI_API_KEY=sk-..." >> .env
   echo "ANTHROPIC_API_KEY=sk-ant-..." >> .env
   ```

3. **Build skills container images:**
   ```bash
   cd docker/skills
   ./build-skills-images.sh
   ```

4. **Test configuration:**
   ```bash
   ./mcp-cli query "Hello, world!"
   ```

## Environment Variables

API keys should be set in `.env` file at project root:

```bash
# LLM Providers
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
DEEPSEEK_API_KEY=...
GEMINI_API_KEY=...
MOONSHOT_API_KEY=...

# MCP Servers (if needed)
BRAVE_API_KEY=...

# Proxy API Keys (if using proxies)
BASH_PROXY_API_KEY=...
FILESYSTEM_PROXY_API_KEY=...
```

## Customization

### Adding a New Provider

1. Create `config/providers/myprovider.yaml`
2. Follow existing examples for structure
3. Restart mcp-cli

### Adding a New Template

1. Create `config/templates/myworkflow.yaml`
2. Define steps and prompts
3. Use with: `./mcp-cli --template myworkflow`

### Exposing as MCP Server

1. Create `config/runasMCP/myagent.yaml`
2. Reference templates to expose
3. Add to Claude Desktop config
4. Restart Claude Desktop

## Documentation

- **Full Documentation:** [docs/](../../docs/)
- **Getting Started:** [docs/getting-started/](../../docs/getting-started/)
- **Skills Guide:** [docs/skills/](../../docs/skills/)
- **Templates:** [docs/templates/](../../docs/templates/)

## Notes

- All API keys use environment variable references: `${VAR_NAME}`
- Server paths should be absolute or relative to project root
- Skills require Docker/Podman for execution
- Proxy configurations expose MCP tools as REST APIs
