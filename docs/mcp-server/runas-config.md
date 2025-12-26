# runas Configuration Specification

Complete technical reference for runas configuration files that map templates to MCP tools.

---

## Purpose

The runas configuration file defines:
1. Server metadata (name, version)
2. Tools exposed to MCP clients
3. Mapping from tools to templates
4. Parameter schemas for each tool

**Location:** `config/runas/<server-name>.yaml`

---

## Schema

### Top Level

```yaml
name: string              # Required: Server identifier
version: string           # Required: Semantic version
description: string       # Optional: Server description
tools: []Tool            # Required: List of exposed tools
```

### Tool Definition

```yaml
name: string              # Required: Tool name (must be unique)
description: string       # Required: What the tool does
template: string          # Required: Template to execute
parameters: {}           # Required: JSON Schema for parameters
```

---

## Complete Example

```yaml
name: code_tools
version: 1.0.0
description: Automated code analysis and review tools

tools:
  # Tool 1: Simple tool with required parameter
  - name: analyze_code
    description: Analyze code for bugs and improvements
    template: code_analyzer
    parameters:
      code:
        type: string
        description: Source code to analyze
        required: true
  
  # Tool 2: Tool with multiple parameters
  - name: review_pr
    description: Review pull request changes
    template: pr_reviewer
    parameters:
      diff:
        type: string
        description: Git diff output
        required: true
      context:
        type: string
        description: Additional context
        required: false
  
  # Tool 3: Tool with typed parameters
  - name: generate_tests
    description: Generate unit tests
    template: test_generator
    parameters:
      code:
        type: string
        required: true
      framework:
        type: string
        enum: ["pytest", "jest", "go"]
        default: "pytest"
      coverage:
        type: number
        minimum: 0
        maximum: 100
        default: 80
```

---

## Parameter Types

### String

```yaml
parameters:
  name:
    type: string
    description: User-facing description
    required: true | false
    default: "value"           # Optional
    enum: ["opt1", "opt2"]     # Optional: restrict to values
    minLength: 1               # Optional
    maxLength: 1000            # Optional
```

**Maps to template:**
```yaml
# Template receives: {{input_data.name}}
```

---

### Number

```yaml
parameters:
  count:
    type: number
    description: Number of items
    required: false
    default: 10
    minimum: 1
    maximum: 100
```

**Maps to template:**
```yaml
# Template receives: {{input_data.count}}
```

---

### Boolean

```yaml
parameters:
  verbose:
    type: boolean
    description: Include detailed output
    required: false
    default: false
```

**Maps to template:**
```yaml
# Template uses:
{% if input_data.verbose %}
  Detailed output...
{% endif %}
```

---

### Object

```yaml
parameters:
  config:
    type: object
    description: Configuration options
    properties:
      mode:
        type: string
        enum: ["quick", "thorough"]
      threshold:
        type: number
    required: false
```

**Maps to template:**
```yaml
# Template receives:
# {{input_data.config.mode}}
# {{input_data.config.threshold}}
```

---

### Array

```yaml
parameters:
  files:
    type: array
    description: List of files to process
    items:
      type: string
    required: true
```

**Maps to template:**
```yaml
# Template can iterate:
{% for file in input_data.files %}
  Process: {{file}}
{% endfor %}
```

---

### Array of Objects

```yaml
parameters:
  tasks:
    type: array
    items:
      type: object
      properties:
        name:
          type: string
        priority:
          type: number
```

**Maps to template:**
```yaml
# Template iterates over structured data:
{% for task in input_data.tasks %}
  Task: {{task.name}}, Priority: {{task.priority}}
{% endfor %}
```

---

## Parameter Mapping Rules

### Rule 1: Prefix with input_data

**All tool parameters are available under `input_data` namespace in templates.**

**Tool parameter:**
```yaml
parameters:
  code:
    type: string
```

**Template access:**
```yaml
prompt: "Analyze: {{input_data.code}}"
```

**Not:** `{{code}}` ❌

---

### Rule 2: Nested Access

**Object parameters use dot notation.**

**Tool parameter:**
```yaml
parameters:
  config:
    type: object
    properties:
      mode:
        type: string
```

**Template access:**
```yaml
prompt: "Mode: {{input_data.config.mode}}"
```

---

### Rule 3: Array Iteration

**Arrays require for_each or template loop.**

**Tool parameter:**
```yaml
parameters:
  files:
    type: array
    items:
      type: string
```

**Template access:**
```yaml
{% for file in input_data.files %}
  Process: {{file}}
{% endfor %}
```

---

## Advanced Patterns

### Pattern 1: Conditional Tool Behavior

**Use parameters to control workflow:**

```yaml
tools:
  - name: analyze_code
    template: code_analyzer
    parameters:
      code:
        type: string
        required: true
      depth:
        type: string
        enum: ["quick", "thorough", "deep"]
        default: "thorough"
```

**Template implements different paths:**

```yaml
steps:
  - name: quick_analysis
    condition: "{{input_data.depth}} == 'quick'"
    prompt: "Quick check: {{input_data.code}}"
  
  - name: thorough_analysis
    condition: "{{input_data.depth}} == 'thorough'"
    prompt: "Thorough analysis: {{input_data.code}}"
  
  - name: deep_analysis
    condition: "{{input_data.depth}} == 'deep'"
    template: deep_analyzer
```

---

### Pattern 2: Optional Enhancements

**Allow clients to request additional features:**

```yaml
tools:
  - name: review_code
    template: code_reviewer
    parameters:
      code:
        type: string
        required: true
      include_tests:
        type: boolean
        default: false
      include_docs:
        type: boolean
        default: false
```

**Template conditionally adds steps:**

```yaml
steps:
  - name: review
    prompt: "Review: {{input_data.code}}"
    output: review
  
  - name: generate_tests
    condition: "{{input_data.include_tests}} == true"
    template: test_generator
    output: tests
  
  - name: generate_docs
    condition: "{{input_data.include_docs}} == true"
    template: doc_generator
    output: docs
```

---

### Pattern 3: Multi-Step with Intermediate Results

**Tool combines multiple templates:**

```yaml
tools:
  - name: comprehensive_review
    template: full_review_pipeline
    parameters:
      code:
        type: string
        required: true
```

**Template orchestrates workflow:**

```yaml
name: full_review_pipeline
steps:
  - name: analyze
    template: code_analyzer
    template_input: "{{input_data}}"
    output: analysis
  
  - name: security_check
    template: security_scanner
    template_input: "{{input_data}}"
    output: security
  
  - name: synthesize
    prompt: |
      Analysis: {{analysis}}
      Security: {{security}}
      
      Create comprehensive review.
```

---

## Validation Rules

### Tool Names

- **Must be unique** within server
- **Snake_case recommended** (e.g., `analyze_code`)
- **No spaces** or special characters
- **Descriptive** names preferred

### Template References

- **Must exist** in template paths
- **Name only** (no .yaml extension)
- **Case sensitive**

**Example:**
- Template file: `config/templates/code_analyzer.yaml`
- Reference: `template: code_analyzer` ✓
- Not: `template: code_analyzer.yaml` ❌

### Parameter Schemas

- **Must be valid JSON Schema**
- **Required parameters must be enforced**
- **Default values must match type**
- **Enums must be same type**

---

## Common Patterns

### Pattern: Simple Pass-Through

**Minimal mapping:**

```yaml
tools:
  - name: translate_text
    template: translator
    parameters:
      text:
        type: string
        required: true
      target_language:
        type: string
        required: true
```

**Template receives parameters directly:**

```yaml
prompt: |
  Translate to {{input_data.target_language}}:
  {{input_data.text}}
```

---

### Pattern: Parameter Transformation

**Tool accepts structured input:**

```yaml
tools:
  - name: analyze_pr
    template: pr_analyzer
    parameters:
      repository:
        type: string
        required: true
      pr_number:
        type: number
        required: true
```

**Template builds context:**

```yaml
steps:
  - name: fetch_pr
    prompt: |
      Get PR #{{input_data.pr_number}} from {{input_data.repository}}
    output: pr_data
  
  - name: analyze
    template: code_analyzer
    template_input: "{{pr_data}}"
```

---

### Pattern: Multi-Tool Server

**Expose related tools:**

```yaml
name: dev_suite
tools:
  - name: review_code
    template: code_reviewer
    parameters: {...}
  
  - name: generate_tests
    template: test_generator
    parameters: {...}
  
  - name: explain_code
    template: code_explainer
    parameters: {...}
  
  - name: refactor_suggest
    template: refactor_advisor
    parameters: {...}
```

**Client sees suite of related tools.**

---

## Testing Your Configuration

### 1. Validate YAML Syntax

```bash
python3 -c "import yaml; yaml.safe_load(open('config/runas/server.yaml'))"
```

### 2. Test Server Startup

```bash
mcp-cli serve config/runas/server.yaml
# Should show: Server ready! Listening on stdio...
```

### 3. Verify Tool Schema

**Check that tools are properly defined:**

```bash
mcp-cli serve config/runas/server.yaml --verbose
# Shows tool registration details
```

### 4. Test Tool Execution

**Simulate tool call:**

```bash
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "analyze_code",
    "arguments": {"code": "def test(): pass"}
  },
  "id": 1
}' | mcp-cli serve config/runas/server.yaml
```

---

## Best Practices

### 1. Descriptive Names and Descriptions

**Good:**
```yaml
- name: analyze_code_quality
  description: Analyze code for bugs, style issues, and performance problems
```

**Poor:**
```yaml
- name: analyze
  description: Analyzes stuff
```

---

### 2. Sensible Defaults

**Provide defaults for optional parameters:**

```yaml
parameters:
  depth:
    type: string
    enum: ["quick", "standard", "deep"]
    default: "standard"  # Most users want this
```

---

### 3. Clear Parameter Descriptions

**Help clients understand parameters:**

```yaml
parameters:
  threshold:
    type: number
    description: Confidence threshold (0.0-1.0). Higher values are stricter.
    minimum: 0.0
    maximum: 1.0
    default: 0.7
```

---

### 4. Validate at Tool Boundary

**Use JSON Schema constraints:**

```yaml
parameters:
  code:
    type: string
    minLength: 1        # No empty strings
    maxLength: 100000   # Reasonable limit
```

---

### 5. Version Your Servers

**Use semantic versioning:**

```yaml
name: code_tools
version: 1.2.0  # Breaking.Feature.Fix
```

**Breaking change:** New major version (1.x.x → 2.0.0)
**New tool:** New minor version (1.2.x → 1.3.0)
**Bug fix:** New patch version (1.2.3 → 1.2.4)

---

## Complete Reference Example

**Production-ready server configuration:**

```yaml
name: research_assistant
version: 1.0.0
description: Multi-source research and fact-checking tools

tools:
  - name: research_topic
    description: Conduct comprehensive research on any topic using multiple sources
    template: deep_research
    parameters:
      topic:
        type: string
        description: Topic to research (e.g., "quantum computing applications")
        required: true
        minLength: 3
        maxLength: 500
      
      depth:
        type: string
        description: Research depth level
        required: false
        default: "standard"
        enum: ["quick", "standard", "deep"]
      
      include_sources:
        type: boolean
        description: Include source citations in response
        required: false
        default: true
  
  - name: fact_check
    description: Verify factual claims with evidence and sources
    template: fact_checker
    parameters:
      claim:
        type: string
        description: Claim to verify (e.g., "The Eiffel Tower is 330m tall")
        required: true
        minLength: 5
        maxLength: 1000
      
      require_sources:
        type: boolean
        description: Require verifiable sources for verification
        required: false
        default: true
  
  - name: compare_sources
    description: Compare what different sources say about a topic
    template: source_comparison
    parameters:
      topic:
        type: string
        description: Topic to compare sources about
        required: true
      
      num_sources:
        type: number
        description: Number of sources to compare (2-10)
        required: false
        default: 3
        minimum: 2
        maximum: 10
```

---

## See Also

- **[Integration Guide](integration.md)** - How to connect MCP clients
- **[Examples](examples/)** - Production-ready configurations
- **[Template Documentation](../templates/)** - Creating templates
