# Workflow Documentation

**Transform AI from chat interface to production infrastructure.**

Workflows are YAML-based definitions that enable iterative, agentic AI workflows with LLM-evaluated exit conditions, automatic provider failover, consensus validation, and MCP tool integration.

---

## What's New in Workflow v2.0

✨ **Major Features:**

- **Iterative Loops:** Workflows that run until LLM-evaluated conditions are met
- **Property Inheritance:** Define once, override where needed
- **Provider Fallback:** Automatic failover across multiple providers
- **Workflow Composition:** Call workflows from workflows
- **Consensus Validation:** Multi-provider agreement on critical decisions
- **Directory Organization:** Organize workflows in subdirectories with intelligent resolution

🔄 **Breaking Changes:**

- Schema identifier: `$schema: "workflow/v2.0"` (was `$template: "v2"`)
- Template calls: `template: {name, with}` (was `template: name`)
- Loops: New top-level `loops` array with LLM-evaluated exit conditions

---

## Quick Start

### Basic Workflow

```yaml
$schema: "workflow/v2.0"
name: code_reviewer
version: 1.0.0
description: Automated code review

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: review
    run: "Review this code: {{input}}"

  - name: report
    needs: [review]
    run: "Format as markdown: {{review}}"
```

### With Provider Failover

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: llama3.2
```

### With Iterative Loop

```yaml
loops:
  - name: develop_until_pass
    workflow: dev_cycle
    with:
      requirements: "{{spec}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "The review says PASS"
    on_failure: continue
```

---

## Documentation

### Core Guides

- **[Schema Reference](SCHEMA.md)** - Complete schema documentation
- **[Loop System](LOOPS.md)** - Iterative execution guide
- **[Authoring Guide](AUTHORING_GUIDE.md)** - How to write workflows
- **[Migration Guide](MIGRATION.md)** - Upgrading from template v1

### Examples & Patterns

- **[Examples](examples/)** - Working workflow examples
- **[Patterns](patterns/)** - Common design patterns

---

## Key Concepts

### Property Inheritance

Properties flow from workflow → step → consensus execution:

```yaml
execution:
  provider: anthropic        # Default for all steps
  temperature: 0.7

steps:
  - name: step1
    run: "..."              # Uses: anthropic, temp=0.7

  - name: step2
    run: "..."
    temperature: 0.3        # Override: anthropic, temp=0.3

  - name: step3
    run: "..."
    provider: openai        # Override: openai, temp=0.7
    model: gpt-4o
```

### Iterative Loops

Execute workflows repeatedly until an LLM-evaluated condition is met:

```yaml
loops:
  - name: develop
    workflow: dev_cycle
    with:
      spec: "{{requirements}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "All tests pass"
```

**Loop variables:**

- `{{loop.iteration}}` - Current iteration number
- `{{loop.last.output}}` - Previous iteration result
- `{{loop.history}}` - All results concatenated

### Provider Fallback

Automatically try multiple providers in order:

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-opus-4      # Try first
    - provider: openai
      model: gpt-4o             # Fallback
    - provider: ollama
      model: llama3.2           # Local fallback
```

### Consensus Validation

Require agreement from multiple providers:

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Is this safe? YES or NO"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: unanimous
```

### Workflow Composition

Call workflows from workflows:

```yaml
steps:
  - name: analyze
    template:
      name: code_analyzer
      with:
        code: "{{input}}"

  - name: review
    needs: [analyze]
    template:
      name: security_checker
      with:
        analysis: "{{analyze}}"
```

---

## Complete Example: Iterative Development

This example demonstrates the power of workflow v2.0 with iterative loops:

```yaml
$schema: "workflow/v2.0"
name: iterative_developer
version: 1.0.0
description: Complete iterative development - plan, test, develop until passing

execution:
  provider: deepseek
  model: deepseek-chat
  temperature: 0.5
  logging: verbose

steps:
  - name: requirements
    template:
      name: planner
      with:
        input: "{{input}}"

  - name: test_criteria
    needs: [requirements]
    template:
      name: test_designer
      with:
        input: "{{requirements}}"

  - name: final_report
    needs: [develop_until_pass]
    run: |
      Development Summary:

      Original request: {{input}}

      Iterations completed: {{loop.iteration}}

      Final code:
      {{develop_until_pass}}

      Provide a brief 2-3 sentence summary of what was built and how many iterations it took.

loops:
  - name: develop_until_pass
    workflow: dev_cycle
    with:
      requirements: "{{requirements}}"
      tests: "{{test_criteria}}"
      previous_code: "{{loop.last.output}}"
      previous_feedback: "From last review"
    max_iterations: 5
    until: "The review says PASS"
    on_failure: continue
    accumulate: development_history
```

**What this does:**

1. **Planning:** Analyzes request and creates requirements
2. **Test Design:** Creates test criteria
3. **Iterative Development:** Loops up to 5 times:
   - Writes code based on requirements and previous attempts
   - Reviews code against test criteria
   - Continues if review says "FAIL", exits if "PASS"
4. **Final Report:** Summarizes development process

---

## Use Cases

### Development & Code Quality

- **Iterative Code Development:** Write, test, refine until passing
- **Code Review Automation:** Multi-step analysis with consensus
- **Test Generation:** Create comprehensive test suites
- **Refactoring:** Iteratively improve code structure

### Operations & DevOps

- **Deployment Validation:** Consensus approval before deploy
- **Incident Analysis:** Multi-provider root cause analysis
- **Configuration Validation:** Verify configs with fallback
- **Log Analysis:** Iterative pattern detection

### Content & Analysis

- **Document Refinement:** Iteratively improve until quality threshold
- **Research Synthesis:** Multi-source analysis with consensus
- **Translation Quality:** Iterative translation with validation
- **Content Generation:** Generate and refine until approved

### Business Processes

- **Decision Validation:** Consensus on critical decisions
- **Risk Assessment:** Multi-provider risk evaluation
- **Compliance Checking:** Iterative compliance verification
- **Contract Analysis:** Multi-step legal review

---

## Best Practices

### 1. Use Property Inheritance

Define common settings once:

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  temperature: 0.7

steps:
  - name: step1
    run: "..."  # Inherits all settings
```

### 2. Set Realistic Loop Limits

```yaml
max_iterations: 5   # Good for iterative development
max_iterations: 3   # Good for refinement
max_iterations: 10  # Good for exploration
```

### 3. Write Clear Exit Conditions

```yaml
# ✅ Good: Clear and specific
until: "The output says PASS"
until: "Error count is zero"

# ❌ Bad: Vague
until: "It looks good"
```

### 4. Use Provider Fallback for Reliability

```yaml
execution:
  providers:
    - provider: anthropic      # Primary
      model: claude-sonnet-4
    - provider: openai         # Backup
      model: gpt-4o
    - provider: ollama         # Local fallback
      model: llama3.2
```

### 5. Validate Critical Decisions with Consensus

```yaml
steps:
  - name: approve_deployment
    consensus:
      prompt: "Approve? YES or NO"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

---

## CLI Usage

### Execute Workflow

```bash
# Root-level workflow
./mcp-cli --workflow my_workflow --input-data "task description"

# Workflow in subdirectory
./mcp-cli --workflow iterative_dev/dev_cycle --input-data "task"
```

### List Available Workflows

```bash
# See all workflows (including nested ones)
./mcp-cli --list-workflows

# Filter by directory
./mcp-cli --list-workflows | jq -r '.workflows[] | select(startswith("iterative_dev/"))'
```

### With Specific Provider

```bash
./mcp-cli --workflow my_workflow \
  --provider anthropic \
  --model claude-sonnet-4 \
  --input-data "task"
```

### With MCP Servers

```bash
./mcp-cli --workflow my_workflow \
  --server filesystem \
  --server brave-search \
  --input-data "task"
```

### Verbose Logging

```bash
./mcp-cli --workflow my_workflow \
  --input-data "task" \
  --verbose
```

### List Available Workflows

```bash
./mcp-cli --list-workflows
```

---

## MCP Server Integration

Expose workflows as MCP tools that any LLM can discover and use.

### Workflow as MCP Tool

**1. Create workflow:**

```yaml
# config/workflows/code_reviewer.yaml
$schema: "workflow/v2.0"
name: code_reviewer
version: 1.0.0
description: Reviews code for security and quality issues
# ... steps ...
```

**2. Create MCP server config:**

```yaml
# config/servers/code_tools.yaml
server_name: code_tools
config:
  command: /path/to/mcp-cli
  args: ["serve", "config/runasMCP/code_tools.yaml"]
```

**3. Define tool exposure:**

```yaml
# config/runasMCP/code_tools.yaml
runas_type: mcp-server
tools:
  - name: review_code
    description: "Reviews code for issues"
    template: code_reviewer
    input_schema:
      type: object
      properties:
        code:
          type: string
      required: [code]
```

**4. Use in Claude Desktop:**

```json
{
  "mcpServers": {
    "code_tools": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/code_tools.yaml"]
    }
  }
}
```

Now any LLM using MCP can call your workflow!

---

## Architecture

### Workflow Execution Flow

```
User Input
    ↓
Workflow Loader (parse YAML)
    ↓
Execution Context (resolve properties)
    ↓
Step Orchestrator
    ↓
┌─────────────────────────────────┐
│ For each step:                  │
│   1. Resolve properties         │
│   2. Interpolate variables      │
│   3. Execute (LLM/template/loop)│
│   4. Store result               │
└─────────────────────────────────┘
    ↓
Result Aggregation
    ↓
Output
```

### Loop Execution Flow

```
Loop Start
    ↓
Iteration = 1
    ↓
┌──────────────────────────────┐
│ Loop Iteration:              │
│   1. Interpolate variables   │
│   2. Call workflow           │
│   3. Store result            │
│   4. Evaluate exit condition │
│   5. Check max_iterations    │
│   6. Decision:               │
│      - Exit or Continue      │
└──────────────────────────────┘
    ↓
Loop Complete
    ↓
Return final result
```

---

## Performance

### Typical Performance Metrics

- **Simple workflow (2 steps):** ~2-4 seconds
- **Loop iteration:** ~3-5 seconds (workflow + condition eval)
- **Consensus (3 providers):** ~8-12 seconds (parallel execution)
- **Workflow composition:** Adds ~1-2 seconds per call

### Optimization Tips

1. **Use provider fallback:** Fast models first, expensive models as backup
2. **Minimize loop iterations:** Set realistic max_iterations
3. **Cache results:** Reuse expensive computations
4. **Parallel consensus:** Executions run in parallel
5. **Keep workflows focused:** Single responsibility principle

---

## Troubleshooting

### Common Issues

**Problem:** Loop never exits early

**Solution:** Check exit condition clarity

```bash
./mcp-cli --workflow my_workflow --input-data "test" --verbose
# Look for: Condition evaluation: 'Your condition' -> YES/NO
```

**Problem:** Provider failover not working

**Solution:** Verify provider configuration

```yaml
execution:
  providers:  # Note: plural
    - provider: anthropic
      model: claude-sonnet-4
```

**Problem:** Variables not interpolating

**Solution:** Check variable names and scope

```yaml
steps:
  - name: step1
    run: "Use {{step1}}"  # ❌ Can't reference self

  - name: step2
    needs: [step1]
    run: "Use {{step1}}"  # ✅ Can reference previous step
```

---

## Migration from Template v1

Quick migration checklist:

- [ ] Update schema: `$schema: "workflow/v2.0"`
- [ ] Wrap execution properties in `execution:` block
- [ ] Update template calls: `template: {name, with}`
- [ ] Move consensus to step-level `consensus:` field
- [ ] Convert iterative patterns to `loops:` array
- [ ] Test thoroughly

See [MIGRATION.md](MIGRATION.md) for detailed guide.

---

## Support & Resources

### Documentation

- [Schema Reference](SCHEMA.md) - Complete schema
- [Loop Guide](LOOPS.md) - Iterative execution
- [Authoring Guide](AUTHORING_GUIDE.md) - Writing workflows
- [Migration Guide](MIGRATION.md) - Upgrade from v1

### Examples

- [examples/](examples/) - Working examples
- [patterns/](patterns/) - Design patterns
- [config/workflows/iterative_dev/](../../config/workflows/iterative_dev/) - Production example

### Community

- GitHub Issues: Bug reports and feature requests
- Discussions: Questions and community support

---

## What's Next

### Planned Features

- **Conditional loops:** Different workflows based on iteration
- **Parallel execution:** Run independent steps concurrently  
- **State persistence:** Resume workflows across sessions
- **Enhanced observability:** Better debugging and monitoring
- **Workflow marketplace:** Share and discover workflows

### Contributing

Contributions welcome! See CONTRIBUTING.md for guidelines.

---

**Last Updated:** January 7, 2026  
**Schema Version:** workflow/v2.0  
**Status:** Production Ready
