# MCP CLI Modular Configuration

This directory contains your modular MCP CLI configuration files.

## Structure

```
mcp-cli                  # Executable
config.yaml              # Main config at executable level
config/                  # Config directory
├── README.md            # This file
├── providers/           # LLM provider configs
│   ├── ollama.yaml
│   ├── openai.yaml
│   └── anthropic.yaml
├── embeddings/          # Embedding provider configs
│   ├── openai.yaml
│   ├── openrouter.yaml
│   └── ollama.yaml
├── servers/             # MCP server configs
│   └── *.yaml
└── templates/           # Workflow templates
    └── *.yaml
```

## Main Config (config.yaml)

The main config file is at the executable level and uses includes to load modular configs:

```yaml
includes:
  providers: config/providers/*.yaml
  embeddings: config/embeddings/*.yaml
  servers: config/servers/*.yaml
  templates: config/templates/*.yaml

ai:
  default_provider: ollama
  default_system_prompt: You are a helpful assistant.

embeddings:
  default_chunk_strategy: sentence
  default_max_chunk_size: 512
```

## Provider Files (LLM)

Each LLM provider gets its own file in `providers/`:

**providers/ollama.yaml:**
```yaml
interface_type: openai_compatible
provider_name: ollama
config:
  api_endpoint: http://localhost:11434
  default_model: qwen2.5:32b
  timeout_seconds: 300
```

## Embedding Files

Embedding providers are separate from LLM providers in `embeddings/`:

**embeddings/openai.yaml:**
```yaml
interface_type: openai_compatible
provider_name: openai
config:
  api_key: ${OPENAI_API_KEY}
  default_embedding_model: text-embedding-3-small
  embedding_models:
    text-embedding-3-small:
      max_tokens: 8191
      dimensions: 1536
      default: true
```

## Server Files

MCP servers are configured in `servers/`:

**servers/filesystem.yaml:**
```yaml
server_name: filesystem
config:
  command: /path/to/filesystem-server
  args: []
```

## Templates

Workflow templates go in `templates/`:

**templates/analyze.yaml:**
```yaml
name: analyze
description: Analyze input data
steps:
  - step: 1
    name: analyze
    base_prompt: Analyze this: {{input_data}}
```

## Environment Variables

API keys should be set in `.env` (next to executable):

```bash
OPENAI_API_KEY=your-key-here
ANTHROPIC_API_KEY=your-key-here
DEEPSEEK_API_KEY=your-key-here
GEMINI_API_KEY=your-key-here
OPENROUTER_API_KEY=your-key-here
```

## Usage

The CLI will automatically find config.yaml next to the executable:

```bash
# Automatic detection
./mcp-cli query "hello"

# Explicit config file
./mcp-cli --config config.yaml query "hello"
```

## Separation of Concerns

**LLM Providers** (`providers/`):
- Chat completions
- Text generation
- Conversation models

**Embedding Providers** (`embeddings/`):
- Vector embeddings
- Semantic search
- RAG applications
- May use same API but different models

**Benefits**:
- Clear separation between LLM and embedding configs
- Easy to configure different embedding providers than LLM providers
- Independent model selection for each purpose
- Clean organization for version control

## Why Separate Embeddings?

While many providers support both LLM and embeddings through the same API,
they serve different purposes:

1. **Different Models**: Embedding models (text-embedding-3-small) vs LLM models (gpt-4o)
2. **Different Use Cases**: Vector search vs text generation
3. **Different Pricing**: Embedding tokens are much cheaper
4. **Independent Selection**: You might use OpenAI for embeddings but Ollama for LLM

This modular structure lets you mix and match providers for each purpose.
