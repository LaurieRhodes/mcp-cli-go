# Example: Code Analysis Server

Production-ready MCP server exposing code analysis workflows.

---

## Server Configuration

**File:** `config/runas/code-analysis.yaml`

```yaml
name: code_analysis
version: 1.0.0
description: Automated code analysis and improvement tools

tools:
  # Static analysis
  - name: analyze_quality
    description: Analyze code for bugs, style issues, and improvements
    template: quality_analyzer
    parameters:
      code:
        type: string
        description: Source code to analyze
        required: true
      language:
        type: string
        description: Programming language
        enum: ["python", "javascript", "go", "java", "rust"]
        required: true
      focus:
        type: string
        description: Analysis focus area
        enum: ["all", "bugs", "style", "performance", "security"]
        default: "all"
  
  # Security scanning
  - name: scan_security
    description: Scan code for security vulnerabilities
    template: security_scanner
    parameters:
      code:
        type: string
        required: true
      language:
        type: string
        required: true
      severity_threshold:
        type: string
        enum: ["low", "medium", "high", "critical"]
        default: "medium"
  
  # Improvement suggestions
  - name: suggest_improvements
    description: Generate actionable improvement recommendations
    template: improvement_advisor
    parameters:
      code:
        type: string
        required: true
      analysis:
        type: string
        description: Previous analysis results (optional)
        required: false
      max_suggestions:
        type: number
        minimum: 1
        maximum: 20
        default: 5
```

---

## Required Templates

### quality_analyzer.yaml

```yaml
name: quality_analyzer
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o
    temperature: 0.3

steps:
  - name: analyze
    prompt: |
      Analyze this {{input_data.language}} code for {{input_data.focus}}:
      
      ```{{input_data.language}}
      {{input_data.code}}
      ```
      
      {% if input_data.focus == 'all' or input_data.focus == 'bugs' %}
      Check for:
      - Logic errors
      - Edge cases not handled
      - Error handling gaps
      {% endif %}
      
      {% if input_data.focus == 'all' or input_data.focus == 'style' %}
      - Code style violations
      - Naming conventions
      - Code organization
      {% endif %}
      
      {% if input_data.focus == 'all' or input_data.focus == 'performance' %}
      - Performance bottlenecks
      - Algorithmic inefficiencies
      - Resource usage issues
      {% endif %}
      
      {% if input_data.focus == 'all' or input_data.focus == 'security' %}
      - Security vulnerabilities
      - Input validation
      - Unsafe practices
      {% endif %}
      
      For each issue found:
      - Severity: CRITICAL | HIGH | MEDIUM | LOW
      - Location: Line number(s)
      - Description: What's wrong
      - Impact: Why it matters
```

### security_scanner.yaml

```yaml
name: security_scanner
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o
    temperature: 0.2  # Lower for consistent security checks

steps:
  - name: scan
    prompt: |
      Security scan for {{input_data.language}} code:
      
      ```{{input_data.language}}
      {{input_data.code}}
      ```
      
      Check for:
      - SQL injection vulnerabilities
      - XSS (Cross-Site Scripting)
      - Command injection
      - Path traversal
      - Insecure cryptography
      - Hardcoded secrets/credentials
      - Unsafe deserialization
      - Authentication/authorization flaws
      
      Report only issues at {{input_data.severity_threshold}} severity or higher.
      
      For each vulnerability:
      - Severity: CRITICAL | HIGH | MEDIUM | LOW
      - Type: OWASP category
      - Location: Line number(s)
      - Exploit scenario: How it could be exploited
      - Fix: How to remediate
```

### improvement_advisor.yaml

```yaml
name: improvement_advisor
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o

steps:
  - name: identify_improvements
    prompt: |
      {% if input_data.analysis %}
      Based on this analysis:
      {{input_data.analysis}}
      
      And this code:
      {% else %}
      Review this code:
      {% endif %}
      
      ```
      {{input_data.code}}
      ```
      
      Suggest {{input_data.max_suggestions}} improvements prioritized by:
      1. Impact (high impact first)
      2. Effort (low effort preferred)
      3. Best practices alignment
      
      For each suggestion:
      - What: Specific change to make
      - Why: Benefit/rationale
      - How: Brief implementation approach
      - Priority: High | Medium | Low
```

---

## Usage Examples

### From Claude Desktop

**User:** "Analyze this Python function for bugs:"

```python
def calculate_average(numbers):
    total = 0
    for num in numbers:
        total += num
    return total / len(numbers)
```

**Claude:** [Calls `analyze_quality` tool]

```
Analysis Results:

HIGH Severity - Logic Error (Line 5)
- No check for empty list
- Impact: Division by zero error
- Fix: Add empty list check

MEDIUM Severity - Type Safety (Line 4)
- No validation that items are numeric
- Impact: TypeError if list contains non-numeric values
- Fix: Add type validation or error handling
```

---

### From Cursor IDE

**Developer selects code, asks:** "Check this for security issues"

**Cursor:** [Calls `scan_security` tool]

```
Security Scan:

No CRITICAL or HIGH severity issues found.

MEDIUM Severity - Input Validation (Line 2)
- Type: CWE-20 Improper Input Validation
- User input not validated before use
- Exploit: Malformed input could cause errors
- Fix: Add input type checking and sanitization
```

---

### From Custom Automation

```python
import subprocess
import json

def analyze_file(filepath, language):
    with open(filepath) as f:
        code = f.read()
    
    # Call MCP server
    request = {
        'jsonrpc': '2.0',
        'method': 'tools/call',
        'params': {
            'name': 'analyze_quality',
            'arguments': {
                'code': code,
                'language': language,
                'focus': 'all'
            }
        },
        'id': 1
    }
    
    proc = subprocess.Popen(
        ['mcp-cli', 'serve', 'config/runas/code-analysis.yaml'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE
    )
    
    proc.stdin.write((json.dumps(request) + '\n').encode())
    proc.stdin.flush()
    
    response = json.loads(proc.stdout.readline())
    return response['result']

# Use in CI/CD
results = analyze_file('src/main.py', 'python')
if has_critical_issues(results):
    sys.exit(1)
```

---

## Workflow Composition

**Chain tools for comprehensive analysis:**

### Step 1: Quality Analysis

```
User: "Review this code thoroughly"
  ↓
Claude calls: analyze_quality(code, language="python", focus="all")
  ↓
Returns: Analysis with issues found
```

### Step 2: Get Improvements

```
Claude calls: suggest_improvements(code, analysis=<previous results>)
  ↓
Returns: Prioritized improvement suggestions
```

### Step 3: Security Check

```
Claude calls: scan_security(code, language="python")
  ↓
Returns: Security vulnerability report
```

**Result:** Comprehensive code review from multiple angles, all automated.

---

## Multi-Provider Optimization

**Different templates use optimal providers:**

```yaml
# quality_analyzer.yaml - Use fast model
config:
  defaults:
    provider: openai
    model: gpt-4o-mini  # Fast, cheap for routine checks

# security_scanner.yaml - Use thorough model
config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4  # More thorough for security

# improvement_advisor.yaml - Use local model
config:
  defaults:
    provider: ollama
    model: qwen2.5:32b  # Free for suggestions
```

**Cost optimization per workflow step.**

---

## Integration Patterns

### Pattern 1: Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Get staged files
STAGED=$(git diff --cached --name-only --diff-filter=ACM | grep '\.py$')

for FILE in $STAGED; do
  # Analyze each file
  mcp-cli serve config/runas/code-analysis.yaml << EOF | jq -r '.result.content[0].text'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "scan_security",
    "arguments": {
      "code": "$(cat $FILE)",
      "language": "python",
      "severity_threshold": "high"
    }
  },
  "id": 1
}
EOF

  # Block commit if critical issues
  if [ $? -ne 0 ]; then
    echo "Security issues found in $FILE"
    exit 1
  fi
done
```

### Pattern 2: CI/CD Pipeline

```yaml
# .github/workflows/code-analysis.yml
name: Code Analysis

on: [pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install MCP-CLI
        run: |
          curl -L https://github.com/.../mcp-cli -o mcp-cli
          chmod +x mcp-cli
      
      - name: Analyze Code
        run: |
          ./mcp-cli serve config/runas/code-analysis.yaml << EOF > analysis.json
          {
            "jsonrpc": "2.0",
            "method": "tools/call",
            "params": {
              "name": "analyze_quality",
              "arguments": {
                "code": "$(cat src/*.py)",
                "language": "python",
                "focus": "all"
              }
            },
            "id": 1
          }
          EOF
      
      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const analysis = require('./analysis.json');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              body: analysis.result.content[0].text
            });
```

---

## Deployment

### Local Development

```json
{
  "mcpServers": {
    "code_analysis": {
      "command": "mcp-cli",
      "args": ["serve", "/Users/dev/mcp-cli/config/runas/code-analysis.yaml"]
    }
  }
}
```

### Team Deployment

```bash
# Install on shared infrastructure
sudo cp mcp-cli /usr/local/bin/
sudo mkdir -p /etc/mcp/runas
sudo cp code-analysis.yaml /etc/mcp/runas/

# Team members use same config
{
  "mcpServers": {
    "code_analysis": {
      "command": "mcp-cli",
      "args": ["serve", "/etc/mcp/runas/code-analysis.yaml"]
    }
  }
}
```

---

## Key Insights

**This example demonstrates:**

1. **Real workflows as tools** - Not toy examples, actual code analysis
2. **Parameter flexibility** - Focus areas, severity thresholds, limits
3. **Template composition** - Build on previous results
4. **Multi-provider strategy** - Optimize cost vs. quality
5. **Integration patterns** - Git hooks, CI/CD, IDE
6. **Production deployment** - Team-wide consistency

**The power:** Turn your domain expertise into reusable AI infrastructure.
