# MCP Stdio Server Configurations

This directory contains configurations for running mcp-cli as MCP stdio servers (primarily for Claude Desktop integration).

## Directory Purpose

**runasMCP** (renamed from "runas") contains configs for:
- MCP stdio servers for Claude Desktop
- Workflow template exposition via MCP protocol
- Skills servers for Claude Desktop

**Not for HTTP proxies** - those are in `config/proxy/`

## Simplified Configuration

All configs now use `config_source` pattern (same as proxy configs):

```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_agent
  version: 1.0.0
  description: "Agent description"

templates:
  - config_source: config/templates/template_name.yaml
    name: optional_custom_name  # Optional
```

**Benefits:**
- ✅ No duplication - derives from source
- ✅ Consistent with proxy configs
- ✅ 95% less configuration
- ✅ Single source of truth

## Quick Start

### 1. Choose an Example

See `examples/` directory for pre-made configs:
- `research_agent.yaml` - Multi-step analysis workflows
- `document_intelligence_agent.yaml` - Document processing
- `simple_analysis.yaml` - Minimal single-template example
- `mcp_skills_stdio.yaml` - Anthropic skills auto-discovery

### 2. Add to Claude Desktop

Edit Claude Desktop config:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/absolute/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/examples/research_agent.yaml"]
    }
  }
}
```

### 3. Restart Claude Desktop

Tools from your MCP server will appear in Claude Desktop!

## Structure

```
config/
├── templates/          # Template definitions (source of truth)
│   └── *.yaml
│
├── proxy/              # HTTP proxy configs (for Open WebUI, APIs)
│   └── *.yaml
│
└── runasMCP/          # MCP stdio server configs (for Claude Desktop)
    ├── README.md      # This file
    └── examples/      # Example configurations
        ├── research_agent.yaml
        ├── document_intelligence_agent.yaml
        ├── simple_analysis.yaml
        └── mcp_skills_stdio.yaml
```

## Pattern Comparison

### Old Approach (Before)
```yaml
tools:
  - template: simple_analysis
    name: simple_analysis
    description: "Long description repeated from template..."
    input_schema:  # Entire schema duplicated
      type: object
      properties:
        input_data:
          type: string
      required: [input_data]
    input_mapping:
      input_data: "{{input_data}}"
```

### New Approach (After)
```yaml
templates:
  - config_source: config/templates/simple_analysis.yaml
```

**Reduction:** ~95% less configuration, zero duplication!

## Creating Custom Configs

### Single Template Server
```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_analyzer
  version: 1.0.0
  description: "Custom analyzer"

templates:
  - config_source: config/templates/my_template.yaml
```

### Multi-Template Server
```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_agent
  version: 1.0.0
  description: "Multi-purpose agent"

templates:
  - config_source: config/templates/template1.yaml
  - config_source: config/templates/template2.yaml
  - config_source: config/templates/template3.yaml
```

### Skills Server
```yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: skills_server
  version: 1.0.0
  description: "Anthropic skills"

# Skills auto-discovered from config/skills/
```

## Key Principles

1. **Derive, Don't Duplicate** - Get info from source configs
2. **Explicit Source** - Use config_source to point to templates
3. **Consistency** - Same pattern as proxy configs
4. **Simplicity** - Minimal configuration

## See Also

- `examples/README.md` - Detailed examples and usage
- `../proxy/README.md` - HTTP proxy configs (different purpose)
- `../templates/` - Template definitions (source of truth)
