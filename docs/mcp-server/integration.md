# Integration Guide

Technical patterns for integrating MCP-CLI server mode with various clients and workflows.

---

## MCP Client Integration

### Claude Desktop

**Config Location:**
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

**Configuration:**

```json
{
  "mcpServers": {
    "server_name": {
      "command": "mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/server.yaml"]
    }
  }
}
```

**Multiple servers:**

```json
{
  "mcpServers": {
    "dev_tools": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/dev-tools.yaml"]
    },
    "research": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/research.yaml"]
    }
  }
}
```

**With environment variables:**

```json
{
  "mcpServers": {
    "my_server": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/server.yaml"],
      "env": {
        "OPENAI_API_KEY": "sk-...",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

**With Unix socket support (for nested MCP execution):**

```json
{
  "mcpServers": {
    "skills": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/skills.yaml"],
      "env": {
        "MCP_SOCKET_PATH": "/tmp/mcp-sockets/skills.sock"
      }
    }
  }
}
```

When `MCP_SOCKET_PATH` is set, the server operates in **dual-mode**:
- Listens on stdio for Claude Desktop
- Listens on Unix socket for nested workflow execution

This enables workflows executed via bash tools to connect without stdio conflicts. See [Unix Socket Support](../architecture/unix-sockets.md) for details.

---

### Cursor IDE

**Config Location:** `.cursor/config.json` in project root

**Configuration:**

```json
{
  "mcp": {
    "servers": {
      "code_tools": {
        "command": "mcp-cli",
        "args": ["serve", "/path/to/code-tools.yaml"]
      }
    }
  }
}
```

---

### Custom MCP Client

**JSON-RPC over stdio protocol:**

```javascript
const { spawn } = require('child_process');

// Start server
const server = spawn('mcp-cli', [
  'serve',
  '/path/to/server.yaml'
]);

// Initialize connection
const initRequest = {
  jsonrpc: '2.0',
  id: 1,
  method: 'initialize',
  params: {
    protocolVersion: '2024-11-05',
    clientInfo: {
      name: 'my-client',
      version: '1.0.0'
    }
  }
};

server.stdin.write(JSON.stringify(initRequest) + '\n');

// Read response
server.stdout.on('data', (data) => {
  const response = JSON.parse(data.toString());
  console.log('Server response:', response);
});

// List tools
const listToolsRequest = {
  jsonrpc: '2.0',
  id: 2,
  method: 'tools/list'
};

server.stdin.write(JSON.stringify(listToolsRequest) + '\n');

// Call tool
const callToolRequest = {
  jsonrpc: '2.0',
  id: 3,
  method: 'tools/call',
  params: {
    name: 'analyze_code',
    arguments: {
      code: 'def foo(): pass'
    }
  }
};

server.stdin.write(JSON.stringify(callToolRequest) + '\n');
```

---

## Integration Patterns

### Pattern 1: Development Workflow Integration

**Expose code-related tools for IDE use:**

```yaml
# config/runas/dev-workflow.yaml
name: dev_workflow
version: 1.0.0

tools:
  # Code understanding
  - name: explain_code
    template: code_explainer
    parameters:
      code: {type: string, required: true}
      language: {type: string, required: true}
  
  # Code generation
  - name: generate_function
    template: function_generator
    parameters:
      description: {type: string, required: true}
      language: {type: string, required: true}
  
  # Code quality
  - name: review_changes
    template: pr_reviewer
    parameters:
      diff: {type: string, required: true}
  
  # Testing
  - name: generate_tests
    template: test_generator
    parameters:
      code: {type: string, required: true}
      framework: {type: string, default: "pytest"}
```

**Integration:**
- Developer writes code
- IDE recognizes available tools
- Can ask "explain this function" → calls `explain_code`
- Can ask "generate tests" → calls `generate_tests`
- Can trigger on save → calls `review_changes`

---

### Pattern 2: Research Pipeline

**Multi-stage research workflow:**

```yaml
# config/runas/research-pipeline.yaml
name: research_pipeline
version: 1.0.0

tools:
  # Stage 1: Initial research
  - name: research_topic
    template: initial_research
    parameters:
      topic: {type: string, required: true}
      depth: {type: string, enum: ["quick", "deep"], default: "quick"}
  
  # Stage 2: Validation
  - name: verify_claims
    template: fact_checker
    parameters:
      claims: {type: array, items: {type: string}, required: true}
  
  # Stage 3: Synthesis
  - name: synthesize_findings
    template: research_synthesizer
    parameters:
      findings: {type: string, required: true}
      format: {type: string, enum: ["report", "summary"], default: "summary"}
```

**Workflow:**
```
User: "Research quantum computing applications"
  → research_topic called
  → Returns findings with claims
  
User: "Verify these claims: [...]"
  → verify_claims called
  → Returns verified facts
  
User: "Create a report from these findings"
  → synthesize_findings called
  → Returns formatted report
```

---

### Pattern 3: Data Analysis Service

**Stateless analysis endpoints:**

```yaml
# config/runas/data-analysis.yaml
name: data_analysis
version: 1.0.0

tools:
  - name: analyze_dataset
    template: statistical_analyzer
    parameters:
      data: {type: string, description: "CSV or JSON data", required: true}
      analysis_type: {type: string, enum: ["descriptive", "correlation", "trend"]}
  
  - name: detect_anomalies
    template: anomaly_detector
    parameters:
      data: {type: string, required: true}
      sensitivity: {type: number, minimum: 0, maximum: 1, default: 0.5}
  
  - name: forecast
    template: forecaster
    parameters:
      historical_data: {type: string, required: true}
      periods: {type: number, minimum: 1, maximum: 100, default: 10}
```

**Integration with data pipeline:**

```python
# Python client
import subprocess
import json

class MCPAnalyzer:
    def __init__(self, config_path):
        self.server = subprocess.Popen(
            ['mcp-cli', 'serve', config_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE
        )
    
    def analyze(self, data, analysis_type='descriptive'):
        request = {
            'jsonrpc': '2.0',
            'id': 1,
            'method': 'tools/call',
            'params': {
                'name': 'analyze_dataset',
                'arguments': {
                    'data': data,
                    'analysis_type': analysis_type
                }
            }
        }
        
        self.server.stdin.write((json.dumps(request) + '\n').encode())
        response = json.loads(self.server.stdout.readline())
        return response['result']

# Usage
analyzer = MCPAnalyzer('/path/to/data-analysis.yaml')
results = analyzer.analyze(csv_data, 'correlation')
```

---

### Pattern 4: Team Workflow Standardization

**Consistent processes across team:**

```yaml
# config/runas/team-workflows.yaml
name: team_workflows
version: 1.0.0

tools:
  # Design review
  - name: review_design
    template: design_reviewer
    parameters:
      design_doc: {type: string, required: true}
      focus_areas: {type: array, items: {type: string}}
  
  # Code review
  - name: review_pr
    template: pr_reviewer
    parameters:
      diff: {type: string, required: true}
      repository: {type: string}
  
  # Documentation
  - name: generate_docs
    template: doc_generator
    parameters:
      code: {type: string, required: true}
      style: {type: string, enum: ["api", "guide", "reference"]}
```

**Team benefit:**
- Same review process for everyone
- Consistent documentation style
- Automated quality gates
- Shared knowledge codified in templates

---

### Pattern 5: Multi-Provider Strategy

**Use different AI providers optimally:**

```yaml
# Quick checks use local model
# config/workflows/quick_check.yaml
name: quick_check
config:
  defaults:
    provider: ollama
    model: qwen2.5:32b

# Deep analysis uses powerful model
# config/workflows/deep_analysis.yaml
name: deep_analysis
config:
  defaults:
    provider: anthropic
    model: claude-opus-4
```

**Exposed as tools:**

```yaml
# config/runas/optimized-analysis.yaml
tools:
  - name: quick_check
    template: quick_check
    # Uses Ollama (free, fast)
  
  - name: thorough_analysis
    template: deep_analysis
    # Uses Claude (paid, comprehensive)
```

**Cost optimization:**
- Routine checks: Free local model
- Important analysis: Paid powerful model
- User chooses based on needs

---

## Environment-Specific Configuration

### Development Environment

```json
{
  "mcpServers": {
    "dev_tools": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/dev-tools.yaml"],
      "env": {
        "LOG_LEVEL": "debug",
        "ANTHROPIC_API_KEY": "${ANTHROPIC_DEV_KEY}"
      }
    }
  }
}
```

### Production Environment

```json
{
  "mcpServers": {
    "prod_tools": {
      "command": "mcp-cli",
      "args": ["serve", "/path/to/prod-tools.yaml"],
      "env": {
        "LOG_LEVEL": "error",
        "ANTHROPIC_API_KEY": "${ANTHROPIC_PROD_KEY}"
      }
    }
  }
}
```

---

## Advanced Integration

### Chained Tool Calls

**Client orchestrates multi-step workflow:**

```javascript
// Step 1: Research
const research = await callTool('research_topic', {
  topic: 'AI safety'
});

// Step 2: Verify key claims
const claims = extractClaims(research.content);
const verified = await callTool('verify_claims', {
  claims: claims
});

// Step 3: Synthesize
const report = await callTool('synthesize_findings', {
  findings: verified.content,
  format: 'report'
});
```

**Each tool is independent, client controls flow.**

---

### Parallel Tool Execution

**Call multiple tools concurrently:**

```javascript
const [codeReview, securityScan, testSuggestions] = await Promise.all([
  callTool('review_code', {code: sourceCode}),
  callTool('security_scan', {code: sourceCode}),
  callTool('suggest_tests', {code: sourceCode})
]);

const comprehensive = combineResults(codeReview, securityScan, testSuggestions);
```

---

### Streaming Results

**For long-running operations:**

```yaml
# Workflow configured for streaming
name: long_research
config:
  defaults:
    streaming: true
```

**Client handles incremental results:**

```javascript
server.stdout.on('data', (chunk) => {
  const lines = chunk.toString().split('\n');
  lines.forEach(line => {
    if (line.trim()) {
      const message = JSON.parse(line);
      if (message.method === 'notifications/progress') {
        updateProgress(message.params);
      }
    }
  });
});
```

---

## Production Considerations

### Process Management

**Use process manager for reliability:**

```bash
# systemd service
[Unit]
Description=MCP Server - Dev Tools

[Service]
ExecStart=/usr/local/bin/mcp-cli serve /etc/mcp/dev-tools.yaml
Restart=always
User=mcp-user

[Install]
WantedBy=multi-user.target
```

### Monitoring

**Log tool calls:**

```yaml
# config.yaml
logging:
  level: info
  file: /var/log/mcp-server.log
```

**Track usage:**

```bash
# Parse logs for metrics
grep "tools/call" /var/log/mcp-server.log | \
  jq '.tool_name' | \
  sort | uniq -c
```

### Rate Limiting

**Protect AI provider quotas:**

```yaml
# config/providers/anthropic.yaml
rate_limit:
  requests_per_minute: 50
  tokens_per_minute: 100000
```

---

## Testing Integration

### Unit Test Tool Definitions

```bash
# Verify server starts
mcp-cli serve config/runas/server.yaml &
SERVER_PID=$!
sleep 2

# Check process running
ps -p $SERVER_PID > /dev/null
if [ $? -eq 0 ]; then
  echo "✓ Server started"
else
  echo "✗ Server failed to start"
fi

kill $SERVER_PID
```

### Integration Test Tool Calls

```bash
# Test tool execution
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "analyze_code",
    "arguments": {"code": "def test(): pass"}
  },
  "id": 1
}' | mcp-cli serve config/runas/server.yaml > response.json

# Verify response
jq '.result' response.json
```

---

## See Also

- **[runas Configuration](runas-config.md)** - Server configuration reference
- **[Examples](examples/)** - Complete working examples
- **[MCP Protocol Spec](https://modelcontextprotocol.io)** - Protocol details
