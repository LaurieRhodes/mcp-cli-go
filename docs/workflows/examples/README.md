# Workflow Examples

Working workflow examples demonstrating real-world use cases with workflow v2.0.

---

## Available Examples

### Basic Workflows

- **[summarize.yaml](summarize.yaml)** - Document summarization
- **[code-review.yaml](code-review.yaml)** - Code review workflow
- **[multi-provider.yaml](multi-provider.yaml)** - Multiple AI providers with consensus

### Advanced Workflows

- **[iterative-improvement.yaml](iterative-improvement.yaml)** - Iterative refinement with loops
- **[quality-gate.yaml](quality-gate.yaml)** - Quality validation with consensus
- **[document-pipeline.yaml](document-pipeline.yaml)** - Sequential document processing

### Composition

- **[composed-analysis.yaml](composed-analysis.yaml)** - Calling workflows from workflows

---

## Quick Start

### 1. Copy Example

```bash
# Copy to your workflows directory
cp docs/workflows/examples/summarize.yaml config/workflows/
```

### 2. Run Example

```bash
# Basic usage
./mcp-cli --workflow summarize --input-data "Your text here"

# With file input
./mcp-cli --workflow summarize --input-data "$(cat document.txt)"

# With servers
./mcp-cli --workflow summarize --server filesystem --input-data "content"
```

### 3. Customize

```bash
# Copy and modify
cp docs/workflows/examples/summarize.yaml config/workflows/my-summary.yaml

# Edit for your needs
vim config/workflows/my-summary.yaml

# Run custom version
./mcp-cli --workflow my-summary --input-data "..."
```

---

## Example Categories

### 🟢 Beginner

Simple workflows for learning:
- **summarize.yaml** - Basic multi-step workflow
- **code-review.yaml** - Step dependencies and analysis

### 🟡 Intermediate

Real-world applications:
- **multi-provider.yaml** - Consensus validation
- **document-pipeline.yaml** - Sequential processing

### 🔴 Advanced

Complex workflows with loops:
- **iterative-improvement.yaml** - LLM-evaluated loops
- **quality-gate.yaml** - Loops + consensus validation

---

## Using Examples

### With Input Data

```bash
./mcp-cli --workflow code-review --input-data '{
  "code": "def hello():\n    print(\"hello\")"
}'
```

### With MCP Servers

```bash
./mcp-cli --workflow document-pipeline \
  --server filesystem \
  --input-data "document.txt"
```

### With Different Providers

```bash
# Override execution provider
./mcp-cli --workflow summarize \
  --provider ollama \
  --model llama3.2 \
  --input-data "text"
```

---

## Example Structure

All examples follow workflow v2.0 schema:

```yaml
$schema: "workflow/v2.0"
name: example_name
version: 1.0.0
description: What this workflow does

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "Prompt with {{input}}"
  
  - name: step2
    needs: [step1]
    run: "Use result: {{step1}}"
```

---

## Creating Your Own

### Start Simple

```yaml
$schema: "workflow/v2.0"
name: my_workflow
version: 1.0.0
description: My custom workflow

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: process
    run: "Process: {{input}}"
```

### Add Features

1. **Add steps:** Build multi-step workflows
2. **Add dependencies:** Control execution order with `needs`
3. **Add loops:** Iterative improvement with `until` conditions
4. **Add consensus:** Multi-provider validation
5. **Add composition:** Call other workflows with `template`

### Test Frequently

```bash
# Test after each change
./mcp-cli --workflow my_workflow --input-data "test data"
```

---

## Workflow Features

### Property Inheritance

Set defaults in `execution`, override in steps:

```yaml
execution:
  provider: anthropic
  temperature: 0.7

steps:
  - name: creative
    run: "..."
    temperature: 0.9  # Override
```

### Step Dependencies

Control execution order:

```yaml
steps:
  - name: extract
    run: "..."
  
  - name: analyze
    needs: [extract]  # Waits for extract
    run: "..."
```

### Iterative Loops

Improve until criteria met:

```yaml
loops:
  - name: improve
    workflow: refine
    with:
      content: "{{input}}"
    max_iterations: 5
    until: "Quality exceeds threshold"
```

### Consensus Validation

Multi-provider agreement:

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Approve?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

---

## Best Practices

### 1. Clear Descriptions

```yaml
# ✅ Good
description: Summarize documents in 200 words with key themes

# ❌ Bad
description: Summarizer
```

### 2. Use Dependencies

```yaml
# ✅ Good
- name: analyze
  needs: [extract]  # Clear dependency

# ❌ Bad
- name: analyze  # Unclear order
```

### 3. Set Appropriate Temperature

```yaml
# ✅ Good
execution:
  temperature: 0.3  # Deterministic analysis

# ❌ Bad
execution:
  temperature: 1.5  # Too random for structured tasks
```

### 4. Provide Context

```yaml
# ✅ Good
run: |
  Analyze this code for security issues:
  {{code}}
  
  Check for:
  - SQL injection
  - XSS vulnerabilities

# ❌ Bad
run: "Check {{code}}"
```

---

## Contributing Examples

Have a useful workflow? Share it!

1. Test thoroughly with various inputs
2. Add clear descriptions and comments
3. Follow schema v2.0
4. Submit PR to `docs/workflows/examples/`

---

## Quick Reference

```bash
# Copy example
cp docs/workflows/examples/NAME.yaml config/workflows/

# Run example
./mcp-cli --workflow NAME --input-data "..."

# Customize
vim config/workflows/NAME.yaml

# Test
./mcp-cli --workflow NAME --input-data "test"
```

---

## See Also

- [Schema Reference](../SCHEMA.md) - Complete schema documentation
- [Authoring Guide](../AUTHORING_GUIDE.md) - How to write workflows
- [Patterns](../patterns/) - Design patterns
- Working example: `config/workflows/iterative_dev/`

---

**Learn by example, build your own workflows!** 📚
