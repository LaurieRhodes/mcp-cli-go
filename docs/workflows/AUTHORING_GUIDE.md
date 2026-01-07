# Workflow Authoring Guide

**Complete guide to writing effective workflows.**

---

## Table of Contents

- [Getting Started](#getting-started)
- [Workflow Structure](#workflow-structure)
- [Writing Prompts](#writing-prompts)
- [Variable Interpolation](#variable-interpolation)
- [Property Inheritance](#property-inheritance)
- [Control Flow](#control-flow)
- [Error Handling](#error-handling)
- [Testing Workflows](#testing-workflows)
- [Best Practices](#best-practices)

---

## Getting Started

### Minimal Workflow

```yaml
$schema: "workflow/v2.0"
name: hello_world
version: 1.0.0
description: Simple hello world workflow

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: greet
    run: "Say hello to {{input}}"
```

**Run it:**

```bash
./mcp-cli --workflow hello_world --input-data "World"
```

### Workflow Template

Use this template as a starting point:

```yaml
$schema: "workflow/v2.0"
name: my_workflow
version: 1.0.0
description: What this workflow does

execution:
  # Provider configuration
  provider: anthropic
  model: claude-sonnet-4

  # Optional: Add MCP servers
  servers: [filesystem]

  # Optional: Model parameters
  temperature: 0.7

  # Optional: Logging
  logging: verbose

# Optional: Environment variables
env:
  KEY: value

steps:
  - name: step1
    run: "First step: {{input}}"

  - name: step2
    needs: [step1]
    run: "Second step using: {{step1}}"
```

---

## Workflow Structure

### Required Fields

Every workflow must have:

```yaml
$schema: "workflow/v2.0"     # Schema identifier
name: unique_name            # Unique identifier
version: 1.0.0              # Semantic version
description: "What it does" # Human-readable description
execution:                   # Execution context
  provider: anthropic
  model: claude-sonnet-4
```

### Optional Sections

```yaml
env:                        # Environment variables
  API_KEY: ${MY_API_KEY}

steps:                      # Sequential execution
  - name: step1
    run: "..."

loops:                      # Iterative execution
  - name: loop1
    workflow: task
    max_iterations: 5
    until: "condition"
```

---

## Writing Prompts

### Basic Prompts

```yaml
steps:
  - name: analyze
    run: "Analyze this code: {{input}}"
```

### Multi-line Prompts

Use `|` for multi-line strings:

```yaml
steps:
  - name: review
    run: |
      Review this code for:
      1. Security issues
      2. Performance problems
      3. Best practices

      Code:
      {{code}}

      Provide detailed feedback.
```

### Structured Prompts

```yaml
steps:
  - name: generate
    run: |
      Task: {{task}}
      Context: {{context}}
      Constraints: {{constraints}}

      Generate a solution that:
      - Meets all requirements
      - Follows best practices
      - Is well documented
```

---

## Variable Interpolation

### Available Variables

**1. Input data:**

```yaml
# Accessed via {{input}}
run: "Analyze: {{input}}"
```

**2. Previous step results:**

```yaml
- name: step2
  needs: [step1]
  run: "Use result from step1: {{step1}}"
```

**3. Environment variables:**

```yaml
env:
  PROJECT: my-project

steps:
  - name: deploy
    run: "Deploy to {{env.PROJECT}}"
```

**4. Loop variables:**

```yaml
loops:
  - name: process
    workflow: task
    with:
      iteration: "{{loop.iteration}}"
      previous: "{{loop.last.output}}"
```

### Interpolation Syntax

```yaml
# Simple variable
"{{variable}}"

# Nested access (if supported)
"{{step1.result}}"

# Environment variable
"{{env.VAR_NAME}}"

# Loop variable
"{{loop.iteration}}"
```

### Best Practices

**✅ Good:**

```yaml
run: "Analyze {{code}} for security issues"
run: "Previous attempt: {{loop.last.output}}"
```

**❌ Avoid:**

```yaml
run: "Analyze {{step1.output.code}}"  # Over-nesting
run: "{{loop.output}} is correct?"    # Interpolating in condition
```

---

## Property Inheritance

### Three Levels of Configuration

Properties can be set at three levels:

```yaml
# Level 1: Execution context (defaults)
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  servers: [filesystem]

steps:
  # Level 2: Step (overrides execution)
  - name: step1
    run: "..."
    temperature: 0.3        # Override

  # Level 3: Consensus execution (overrides step)
  - name: step2
    consensus:
      prompt: "..."
      executions:
        - provider: openai  # Override
          model: gpt-4o
          temperature: 0    # Override
```

### Common Patterns

**Pattern 1: Consistent defaults**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "..."  # Uses all defaults
  - name: step2
    run: "..."  # Uses all defaults
```

**Pattern 2: Per-step specialization**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: creative
    run: "Write a story"
    temperature: 0.9      # More creative

  - name: analytical
    run: "Analyze data"
    temperature: 0.1      # More deterministic
```

**Pattern 3: Provider fallback**

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: llama3.2

steps:
  - name: critical
    run: "..."
    # Uses full fallback chain

  - name: simple
    run: "..."
    provider: ollama      # Override with single provider
    model: llama3.2
```

---

## Control Flow

### Dependencies

Use `needs` to define step order:

```yaml
steps:
  - name: step1
    run: "First step"

  - name: step2
    needs: [step1]
    run: "Uses: {{step1}}"

  - name: step3
    needs: [step1, step2]
    run: "Uses: {{step1}} and {{step2}}"
```

### Conditional Execution

Use `if` to conditionally execute steps:

```yaml
steps:
  - name: check
    run: "Is this safe? YES or NO"

  - name: proceed
    if: ${{ check == "YES" }}
    run: "Proceeding with action"

  - name: abort
    if: ${{ check == "NO" }}
    run: "Aborting due to safety concerns"
```

### Loops

Use `for_each` for iteration:

```yaml
steps:
  - name: process_items
    for_each: "{{items}}"
    item_name: item
    run: "Process: {{item}}"
```

---

## Error Handling

### Retry Logic

```yaml
steps:
  - name: network_call
    run: "Make API request"
    on_error:
      retry: 3
      backoff: exponential
```

### Fallback Steps

```yaml
steps:
  - name: primary_method
    run: "Try primary approach"
    on_error:
      fallback: backup_method

  - name: backup_method
    run: "Try backup approach"
```

### Loop Error Handling

```yaml
loops:
  - name: develop
    workflow: dev_cycle
    max_iterations: 5
    until: "Tests pass"
    on_failure: continue  # Don't stop on errors
```

---

## Testing Workflows

### Local Testing

**1. Test with simple input:**

```bash
./mcp-cli --workflow my_workflow --input-data "test input"
```

**2. Test with verbose logging:**

```bash
./mcp-cli --workflow my_workflow \
  --input-data "test" \
  --verbose
```

**3. Test with specific provider:**

```bash
./mcp-cli --workflow my_workflow \
  --provider ollama \
  --model llama3.2 \
  --input-data "test"
```

### Testing Loops

**1. Test with low max_iterations:**

```yaml
loops:
  - name: test_loop
    workflow: task
    max_iterations: 2  # Low for testing
    until: "Complete"
```

**2. Watch condition evaluation:**

```bash
./mcp-cli --workflow loop_workflow --input-data "test" --verbose
# Look for: "Condition evaluation: 'Complete' -> YES/NO"
```

### Testing Consensus

**1. Test with single provider first:**

```yaml
consensus:
  prompt: "Approve?"
  executions:
    - provider: anthropic
      model: claude-sonnet-4
  require: unanimous
```

**2. Add providers incrementally:**

```yaml
consensus:
  executions:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai  # Add second
      model: gpt-4o
```

---

## Best Practices

### 1. Clear Naming

**✅ Good:**

```yaml
name: analyze_security_vulnerabilities
steps:
  - name: scan_dependencies
  - name: check_authentication
  - name: validate_inputs
```

**❌ Bad:**

```yaml
name: workflow1
steps:
  - name: step1
  - name: step2
  - name: step3
```

### 2. Descriptive Prompts

**✅ Good:**

```yaml
run: |
  Review this code for security vulnerabilities.

  Focus on:
  - SQL injection
  - XSS attacks
  - Authentication bypass

  Code: {{input}}
```

**❌ Bad:**

```yaml
run: "Review {{input}}"
```

### 3. Use Dependencies

**✅ Good:**

```yaml
- name: analyze
  run: "Analyze: {{input}}"

- name: report
  needs: [analyze]  # Explicit dependency
  run: "Report on: {{analyze}}"
```

**❌ Bad:**

```yaml
- name: analyze
  run: "Analyze: {{input}}"

- name: report
  run: "Report on: {{analyze}}"  # No explicit dependency
```

### 4. Set Appropriate Temperatures

```yaml
# Code generation: Low temperature
- name: generate_code
  run: "Generate function"
  temperature: 0.2

# Creative writing: High temperature
- name: write_story
  run: "Write a story"
  temperature: 0.9

# Analysis: Medium temperature
- name: analyze
  run: "Analyze data"
  temperature: 0.5
```

### 5. Use Provider Fallback

**✅ Good:**

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
```

**❌ Risky:**

```yaml
execution:
  provider: anthropic  # Single point of failure
  model: claude-sonnet-4
```

### 6. Keep Workflows Focused

**✅ Good:**

```yaml
# One workflow: code review
name: code_reviewer
steps:
  - name: analyze
  - name: check_security
  - name: report
```

**❌ Bad:**

```yaml
# One workflow: everything
name: do_everything
steps:
  - name: review_code
  - name: deploy_app
  - name: send_email
  - name: generate_docs
```

### 7. Document Complex Logic

```yaml
steps:
  # This step validates the deployment configuration
  # by checking:
  # 1. Environment variables are set
  # 2. Required files exist
  # 3. Network connectivity is available
  - name: validate_deployment
    run: |
      Validate deployment configuration...
```

---

## Common Patterns

### Pattern: Multi-Step Analysis

```yaml
steps:
  - name: gather_data
    run: "Extract relevant data from: {{input}}"

  - name: analyze
    needs: [gather_data]
    run: "Analyze data: {{gather_data}}"

  - name: recommend
    needs: [analyze]
    run: "Recommend actions based on: {{analyze}}"
```

### Pattern: Iterative Refinement

```yaml
loops:
  - name: refine
    workflow: improve_content
    with:
      content: "{{input}}"
      previous: "{{loop.last.output}}"
      iteration: "{{loop.iteration}}"
    max_iterations: 3
    until: "Quality score above 8/10"
```

### Pattern: Consensus Validation

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

### Pattern: Workflow Composition

```yaml
steps:
  - name: analyze
    template:
      name: code_analyzer
      with:
        code: "{{input}}"

  - name: security_check
    needs: [analyze]
    template:
      name: security_scanner
      with:
        analysis: "{{analyze}}"
```

---

## Debugging Tips

### 1. Enable Verbose Logging

```bash
./mcp-cli --workflow my_workflow \
  --input-data "test" \
  --verbose
```

### 2. Test Steps Individually

Break workflow into smaller pieces and test each:

```yaml
# Test just step 1
steps:
  - name: step1
    run: "..."
```

### 3. Check Variable Interpolation

Add debug step:

```yaml
steps:
  - name: debug
    run: |
      Debug info:
      - Input: {{input}}
      - Step1: {{step1}}
      - Env: {{env.VAR}}
```

### 4. Validate Exit Conditions

```yaml
loops:
  - name: test
    workflow: task
    max_iterations: 2  # Low for testing
    until: "The output says COMPLETE"
```

Run and check condition evaluation in logs.

---

## Advanced Topics

### Custom Environment Variables

```yaml
env:
  API_KEY: ${MY_SECRET_KEY}
  PROJECT: production
  REGION: us-west-2

steps:
  - name: deploy
    run: |
      Deploy to {{env.REGION}}
      Project: {{env.PROJECT}}
```

### Dynamic Input Mapping

```yaml
steps:
  - name: process
    template:
      name: processor
      with:
        data: "{{input}}"
        mode: "{{env.MODE}}"
        timestamp: "2024-01-01"
```

### Parallel-Style Execution

While true parallel execution isn't supported yet, use independent steps:

```yaml
steps:
  # These have no dependencies, can run in any order
  - name: task_a
    run: "Independent task A"

  - name: task_b
    run: "Independent task B"

  # This waits for both
  - name: combine
    needs: [task_a, task_b]
    run: "Combine: {{task_a}} and {{task_b}}"
```

---

## Workflow Organization

### Directory Structure

```
config/workflows/
├── README.md
├── basic/
│   ├── analyzer.yaml
│   └── reporter.yaml
├── development/
│   ├── code_reviewer.yaml
│   ├── test_generator.yaml
│   └── iterative_dev/
│       ├── planner.yaml
│       ├── dev_cycle.yaml
│       └── iterative_developer.yaml
└── operations/
    ├── deployer.yaml
    └── validator.yaml
```

### Naming Conventions

- **Files:** `snake_case.yaml`
- **Workflow names:** `snake_case`
- **Step names:** `snake_case`
- **Variables:** `snake_case`

---

## See Also

- [Schema Reference](SCHEMA.md) - Complete schema
- [Loop Guide](LOOPS.md) - Iterative execution
- [Examples](examples/) - Working examples
- [Patterns](patterns/) - Design patterns

---

**Last Updated:** January 7, 2026
