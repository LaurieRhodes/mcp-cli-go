# MCP Server Mode

**Expose your AI workflows as discoverable, callable tools for any MCP client.**

---

## The Core Concept

You've built AI workflows (templates) that solve specific problems:

- Code review processes
- Research methodologies  
- Data analysis pipelines
- Document generation workflows

**Server mode makes these workflows available as tools** that any MCP-compatible application can discover and use.

```
Your Templates              Server Mode              Any MCP Client
─────────────────          ─────────────          ──────────────────
code_review.yaml    →                      →      Claude Desktop
research.yaml       →   Exposed as Tools   →      Cursor IDE
analyze_data.yaml   →                      →      Custom Applications
```

---

## Why This Matters

### Without Server Mode

- Templates run only via `mcp-cli` command line
- Each user must know template names and parameters
- No IDE integration
- Manual execution only

### With Server Mode

- Templates become **discoverable tools**
- AI assistants can **automatically use** your workflows
- **IDE integration** with zero configuration
- **Programmatic access** for automation

---

## The Value Proposition

**You're building AI infrastructure, not just scripts.**

1. **Reusability**: Write workflow once, use everywhere
2. **Discoverability**: Tools auto-discovered by MCP clients
3. **Consistency**: Same analysis/review process across team
4. **Composability**: Tools can call other tools
5. **Integration**: Works with any MCP-compatible client

---

## How It Works

### 1. You Define the Mapping

**File:** `config/runas/my-server.yaml`

```yaml
name: dev_tools
version: 1.0.0

tools:
  - name: review_code
    description: Automated code review
    template: code_reviewer    # Maps to your template
    parameters:
      code:
        type: string
        description: Code to review
        required: true
```

This configuration:

- Exposes your `code_reviewer` template as a tool called `review_code`
- Defines the MCP tool schema (name, description, parameters)
- Maps tool parameters to template input variables

### 2. Start the Server

```bash
mcp-cli serve config/runas/my-server.yaml
```

Server listens on stdio, implementing MCP protocol.

### 3. MCP Client Connects

Client (Claude Desktop, Cursor, etc.):

1. Connects to server via stdio
2. Discovers available tools via `tools/list`
3. Sees `review_code` with its parameter schema
4. Can call the tool with parameters

### 4. Tool Execution Flow

```
Client calls: review_code(code="def foo(): pass")
      ↓
Server receives JSON-RPC request
      ↓
Maps parameters to template variables:
  {{input_data.code}} = "def foo(): pass"
      ↓
Executes code_reviewer template
      ↓
Returns results to client
```

---

## The runas Configuration

**This is the bridge between MCP protocol and your templates.**

Key responsibilities:

1. **Tool Schema Definition**: What parameters does the tool accept?
2. **Template Mapping**: Which template implements this tool?
3. **Parameter Translation**: How do tool parameters map to template variables?

**Detailed specification:** [runas-config.md](runas-config.md)

---

## Integration Patterns

### Pattern 1: IDE Integration

Expose development workflows as tools for IDEs:

```yaml
tools:
  - name: explain_code
    template: code_explainer

  - name: generate_tests
    template: test_generator

  - name: refactor_suggest
    template: refactoring_advisor
```

**Result:** Developers get AI-powered code tools directly in their IDE.

### Pattern 2: Research Assistant

Expose research methodologies:

```yaml
tools:
  - name: deep_research
    template: multi_source_research

  - name: fact_check
    template: claim_verification

  - name: synthesize
    template: research_synthesizer
```

**Result:** Researchers get consistent, repeatable research workflows.

### Pattern 3: Data Pipeline

Expose analysis workflows:

```yaml
tools:
  - name: analyze_dataset
    template: statistical_analyzer

  - name: detect_anomalies
    template: anomaly_detector

  - name: generate_report
    template: report_generator
```

**Result:** Data analysts get standardized analysis tools.

---

## Multi-Template Workflows

**Templates can call other templates**, creating powerful compositions:

```yaml
# Parent template
steps:
  - name: analyze
    template: code_analyzer
    template_input: "{{input_data}}"
    output: analysis

  - name: suggest
    template: improvement_suggester
    template_input: "{{analysis}}"
    output: suggestions

  - name: generate_tests
    template: test_generator
    template_input: "{{input_data.code}}"
    output: tests
```

**Exposed as single tool:**

```yaml
tools:
  - name: comprehensive_review
    template: full_code_review  # Calls multiple templates internally
```

Client sees one tool, gets multi-step workflow.

---

## Provider Flexibility

**Different templates can use different AI providers:**

```yaml
# Template 1: Use fast, cheap model
name: quick_check
config:
  defaults:
    provider: ollama
    model: qwen2.5:32b

# Template 2: Use powerful model
name: deep_analysis
config:
  defaults:
    provider: anthropic
    model: claude-opus-4
```

**Optimize cost vs. quality per workflow.**

---

## MCP Client Configuration

**Claude Desktop example:**

```json
{
  "mcpServers": {
    "dev_tools": {
      "command": "mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/dev-tools.yaml"]
    }
  }
}
```

**Cursor IDE example:**

```json
{
  "mcp.servers": {
    "dev_tools": {
      "command": "mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/dev-tools.yaml"]
    }
  }
}
```

**Key requirement:** Absolute paths to configuration files.

---

## Real-World Examples

See [examples/](examples/) for production-ready server configurations:

- **research-agent** - Multi-source research workflows
- **code-reviewer** - Automated code review pipeline
- **data-analyst** - Statistical analysis tools

Each example includes:

- Complete runas configuration
- All required templates
- Usage patterns
- Cost estimates

---

## Quick Start

**Prerequisites:**

- MCP-CLI installed and configured
- Working templates in `config/templates/`
- MCP client (Claude Desktop, Cursor, etc.)

**Steps:**

1. **Create runas config:**
   
   ```yaml
   # config/runas/my-server.yaml
   name: my_server
   version: 1.0.0
   tools:
     - name: my_tool
       template: my_template
       parameters:
         input:
           type: string
           required: true
   ```

2. **Test locally:**
   
   ```bash
   mcp-cli serve config/runas/my-server.yaml
   # Should show: Server ready! Listening on stdio...
   ```

3. **Add to MCP client config** (use absolute path)

4. **Restart client**

5. **Verify tools appear** in client

---

## Technical Details

- **Protocol:** JSON-RPC 2.0 over stdio
- **Transport:** Standard input/output
- **Discovery:** MCP `tools/list` method
- **Execution:** MCP `tools/call` method
- **State:** Stateless (each call independent)

---

## Documentation

- **[runas Configuration](runas-config.md)** - Complete specification
- **[Integration Guide](integration.md)** - Client integration patterns
- **[Examples](examples/)** - Production-ready configurations

---

## Key Insight

**You're not just running AI queries—you're building reusable AI infrastructure.**

Server mode transforms ad-hoc AI workflows into:

- Discoverable tools
- Standardized interfaces
- Team-wide capabilities
- Composable building blocks

**This is the unique power of MCP-CLI server mode.**
