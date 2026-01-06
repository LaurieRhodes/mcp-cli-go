# MCP Stdio Server Examples

Example configurations for running workflow templates and skills as MCP stdio servers (for Claude Desktop integration).

## Simplified Configuration Pattern

All configs now use the same pattern as proxy configs:

```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_agent
  version: 1.0.0
  description: "Agent description"

# Templates to expose - info derived from template configs
templates:
  - config_source: config/templates/template_name.yaml
    name: optional_custom_name          # Optional
    description: optional_description   # Optional
```

**Benefits:**
- ✅ No duplication - info derived from source
- ✅ Consistent with proxy configs
- ✅ Simple and clear
- ✅ Easy to maintain

## Available Examples

### research_agent.yaml
Research and analysis workflows for comprehensive data analysis.

**Templates:**
- multi_step_analysis (as "deep_analysis")
- simple_analysis
- parallel_analysis

**Use case:** Complex multi-step research and analysis tasks

### document_intelligence_agent.yaml
Document processing and NLP workflows.

**Templates:**
- document_intelligence
- entity_extraction
- sentiment_analysis
- summarization

**Use case:** Document analysis, NER, sentiment, summarization

### simple_analysis.yaml
Minimal single-template example.

**Templates:**
- simple_analysis

**Use case:** Testing, basic analysis

### mcp_skills_stdio.yaml
Auto-discovered Anthropic skills.

**Type:** mcp-skills (auto-discovery)

**Use case:** Document processing, code execution (docx, xlsx, pdf skills)

## Usage

### Add to Claude Desktop

Edit your Claude Desktop config file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "research-agent": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/examples/research_agent.yaml"]
    },
    "document-intelligence": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/examples/document_intelligence_agent.yaml"]
    },
    "skills": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/examples/mcp_skills_stdio.yaml"]
    }
  }
}
```

### Test Locally

```bash
# Start MCP server (waits for stdio input)
/path/to/mcp-cli serve config/runasMCP/examples/research_agent.yaml

# Server will respond to MCP protocol messages on stdin/stdout
```

## Creating Custom Configs

### Single Template

```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_analyzer
  version: 1.0.0
  description: "Custom analysis tool"

templates:
  - config_source: config/templates/my_template.yaml
```

### Multiple Templates with Custom Names

```yaml
runas_type: mcp
version: "1.0"

server_info:
  name: my_agent
  version: 1.0.0
  description: "Multi-purpose agent"

templates:
  - config_source: config/templates/template1.yaml
    name: custom_name_1
  
  - config_source: config/templates/template2.yaml
    name: custom_name_2
    description: "Custom description"
  
  - config_source: config/templates/template3.yaml
    # Uses template name and description from source
```

### Skills Server

```yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: my_skills
  version: 1.0.0
  description: "Anthropic skills"

# That's it! Skills auto-discovered from config/skills/
```

## Comparison: Old vs New

### Old Approach (Lots of Duplication)

```yaml
tools:
  - template: simple_analysis
    name: simple_analysis
    description: "Basic analysis template demonstrating..."  # 200+ chars duplicated
    input_schema:                                             # Entire schema duplicated
      type: object
      properties:
        input_data:
          type: string
          description: "Data to analyze"
      required: [input_data]
    input_mapping:
      input_data: "{{input_data}}"
```

### New Approach (Derived from Source)

```yaml
templates:
  - config_source: config/templates/simple_analysis.yaml
```

**Result:** 95% less configuration, no duplication!

## Benefits

✅ **No Duplication** - Info derived from template configs  
✅ **Consistency** - Same pattern as proxy configs  
✅ **Simplicity** - Minimal configuration  
✅ **Maintainability** - Single source of truth  
✅ **Clarity** - Explicit config_source  

## Directory Structure

```
config/
├── templates/              # Template definitions (source of truth)
│   ├── simple_analysis.yaml
│   ├── sentiment_analysis.yaml
│   └── ...
│
├── proxy/                  # HTTP proxy configs
│   └── *.yaml
│
└── runasMCP/              # MCP stdio server configs
    ├── examples/          # Example configurations
    │   ├── research_agent.yaml
    │   ├── document_intelligence_agent.yaml
    │   └── mcp_skills_stdio.yaml
    └── README.md
```

## Troubleshooting

**Error: "template source points to unknown template"**
- Check template exists in `config/templates/`
- Verify filename matches template name in config

**Claude Desktop not showing tools:**
- Check Claude Desktop config path is absolute
- Restart Claude Desktop after config changes
- Check server name doesn't conflict with existing servers

**Server not starting:**
- Test with `mcp-cli serve /path/to/config.yaml`
- Check for error messages in Claude Desktop logs
- Verify mcp-cli binary path is correct
