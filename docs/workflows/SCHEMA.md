# Workflow Schema Reference v2.0

**Status:** Production  
**Date:** January 7, 2026  
**Schema Version:** `workflow/v2.0`

This document defines the complete schema for mcp-cli workflows (formerly called templates).

---

## Table of Contents

- [Overview](#overview)
- [Complete Schema](#complete-schema)
- [Core Objects](#core-objects)
  - [WorkflowV2](#workflowv2)
  - [ExecutionContext](#executioncontext)
  - [StepV2](#stepv2)
  - [LoopV2](#loopv2)
- [Special Modes](#special-modes)
  - [Template Mode](#template-mode)
  - [Embeddings Mode](#embeddings-mode)
  - [Consensus Mode](#consensus-mode)
- [Property Inheritance](#property-inheritance)
- [Loop System](#loop-system)
- [Examples](#examples)

---

## Overview

Workflows define AI-powered automation using a YAML schema. Key features:

- **Property Inheritance:** Define defaults once, override where needed
- **Provider Fallback:** Automatic failover across multiple providers
- **Iterative Loops:** LLM-evaluated exit conditions for agentic workflows
- **Workflow Composition:** Call workflows from workflows
- **Consensus Validation:** Multi-provider agreement on critical decisions
- **MCP Integration:** Access to tools via Model Context Protocol

---

## Complete Schema

```yaml
$schema: "workflow/v2.0"
name: string                    # Required: Unique workflow identifier
version: string                 # Required: Semantic version (e.g., "1.0.0")
description: string             # Required: What this workflow does

# ============================================================================
# Execution Context: Workflow-level defaults
# ============================================================================
execution:
  # Provider configuration (use one approach):

  # Option 1: Single provider (no fallback)
  provider: string              # e.g., "anthropic", "openai", "deepseek"
  model: string                 # e.g., "claude-sonnet-4", "gpt-4o"

  # Option 2: Fallback chain (tries in order)
  providers:
    - provider: string
      model: string
    - provider: string
      model: string

  # MCP servers to make available
  servers: [string]             # e.g., ["filesystem", "brave-search"]

  # Model parameters
  temperature: float            # 0.0 to 2.0, default: 0.7
  max_tokens: int               # Maximum tokens in response

  # Execution control
  timeout: duration             # e.g., "30s", "5m", "1h"

  # Logging level
  logging: string               # "normal" | "verbose" | "noisy"
  no_color: boolean             # Disable colored output

# Environment variables accessible to all steps
env:
  KEY: value                    # String values only

# ============================================================================
# Steps: Sequential execution
# ============================================================================
steps:
  - name: string                # Required: Step identifier

    # ---- Primary execution mode (choose one) ----

    run: string                 # LLM prompt with {{variable}} interpolation

    template:                   # Call another workflow
      name: string
      with:
        key: value

    embeddings:                 # Generate embeddings
      model: string
      input: string | array

    consensus:                  # Multi-provider consensus
      prompt: string
      executions: [...]
      require: string

    # ---- Property overrides (inherit from execution if not specified) ----

    provider: string            # Override execution.provider
    model: string               # Override execution.model
    providers: [...]            # Override with different fallback chain
    servers: [string]           # Override execution.servers
    temperature: float          # Override execution.temperature
    max_tokens: int             # Override execution.max_tokens
    timeout: duration           # Override execution.timeout
    logging: string             # Override execution.logging
    no_color: boolean           # Override execution.no_color
    input: string | object      # Specific input for this step

    # ---- Control flow ----

    if: string                  # Conditional expression
    needs: [string]             # Dependencies on other steps
    for_each: string            # Loop over array
    item_name: string           # Variable name for loop item

    # ---- Error handling ----

    on_error:
      retry: int                # Number of retries
      backoff: string           # "exponential" | "linear"
      fallback: string          # Step name to execute on failure

    # ---- Output configuration ----

    outputs:
      name: string              # Output variable name
      transform: string         # Transformation expression

# ============================================================================
# Loops: Iterative execution until condition met
# ============================================================================
loops:
  - name: string                # Required: Loop identifier
    workflow: string            # Required: Workflow to call each iteration
    with:                       # Input parameters (interpolated each iteration)
      key: value
    max_iterations: int         # Required: Safety limit
    until: string               # Required: Exit condition (LLM evaluates)
    on_failure: string          # "halt" | "continue" | "retry"
    accumulate: string          # Variable name to store all iterations
```

---

## Core Objects

### WorkflowV2

The root object defining a workflow.

**Fields:**

| Field         | Type              | Required | Description                      |
| ------------- | ----------------- | -------- | -------------------------------- |
| `$schema`     | string            | Yes      | Must be `"workflow/v2.0"`        |
| `name`        | string            | Yes      | Unique workflow identifier       |
| `version`     | string            | Yes      | Semantic version (e.g., "1.0.0") |
| `description` | string            | Yes      | Human-readable description       |
| `execution`   | ExecutionContext  | Yes      | Workflow-level defaults          |
| `env`         | map[string]string | No       | Environment variables            |
| `steps`       | []StepV2          | No       | Sequential steps to execute      |
| `loops`       | []LoopV2          | No       | Iterative execution blocks       |

**Example:**

```yaml
$schema: "workflow/v2.0"
name: code_reviewer
version: 1.0.0
description: Automated code review workflow

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: review
    run: "Review this code: {{input}}"
```

---

### ExecutionContext

Defines workflow-level defaults that all steps inherit.

**Fields:**

| Field         | Type               | Required | Description                     |
| ------------- | ------------------ | -------- | ------------------------------- |
| `provider`    | string             | No*      | Single provider name            |
| `model`       | string             | No*      | Model name                      |
| `providers`   | []ProviderFallback | No*      | Fallback chain                  |
| `servers`     | []string           | No       | MCP servers                     |
| `temperature` | float64            | No       | Sampling temperature (0.0-2.0)  |
| `max_tokens`  | int                | No       | Maximum response tokens         |
| `timeout`     | duration           | No       | Execution timeout               |
| `logging`     | string             | No       | "normal", "verbose", or "noisy" |
| `no_color`    | bool               | No       | Disable colored output          |

\* Must specify either `provider`+`model` OR `providers`

**ProviderFallback:**

```yaml
providers:
  - provider: anthropic
    model: claude-sonnet-4
  - provider: openai
    model: gpt-4o
```

**Example:**

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: llama3.2
  servers: [filesystem, brave-search]
  temperature: 0.7
  timeout: 30s
  logging: verbose
```

---

### StepV2

Represents a single step in a workflow.

**Core Fields:**

| Field        | Type           | Required | Description                          |
| ------------ | -------------- | -------- | ------------------------------------ |
| `name`       | string         | Yes      | Step identifier (unique in workflow) |
| `run`        | string         | No*      | LLM prompt                           |
| `template`   | TemplateMode   | No*      | Call another workflow                |
| `embeddings` | EmbeddingsMode | No*      | Generate embeddings                  |
| `consensus`  | ConsensusMode  | No*      | Multi-provider consensus             |

\* Must specify one execution mode

**Property Overrides:**

All ExecutionContext fields can be overridden at the step level:

```yaml
- name: specialized_step
  run: "Analyze: {{input}}"
  provider: openai           # Override
  model: gpt-4o              # Override
  temperature: 0.3           # Override
  servers: [brave-search]    # Override
  logging: noisy             # Override
```

**Control Flow:**

| Field       | Type     | Description                 |
| ----------- | -------- | --------------------------- |
| `if`        | string   | Conditional expression      |
| `needs`     | []string | Dependencies on other steps |
| `for_each`  | string   | Loop over array             |
| `item_name` | string   | Variable name in loop       |

**Error Handling:**

```yaml
on_error:
  retry: 3
  backoff: exponential
  fallback: error_handler
```

---

### LoopV2

Defines iterative execution until an LLM-evaluated condition is met.

**Fields:**

| Field            | Type           | Required | Description                        |
| ---------------- | -------------- | -------- | ---------------------------------- |
| `name`           | string         | Yes      | Loop identifier                    |
| `workflow`       | string         | Yes      | Workflow to call each iteration    |
| `with`           | map[string]any | No       | Input parameters (interpolated)    |
| `max_iterations` | int            | Yes      | Safety limit                       |
| `until`          | string         | Yes      | Exit condition for LLM to evaluate |
| `on_failure`     | string         | No       | "halt", "continue", or "retry"     |
| `accumulate`     | string         | No       | Variable to store all iterations   |

**Loop Variables:**

Available in `with` parameters via interpolation:

- `{{loop.iteration}}` - Current iteration number (1, 2, 3...)
- `{{loop.output}}` - Output from current iteration
- `{{loop.last.output}}` - Output from previous iteration
- `{{loop.history}}` - All outputs separated by newlines and `---`

**Example:**

```yaml
loops:
  - name: develop_until_pass
    workflow: dev_cycle
    with:
      spec: "{{test_plan}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "The output says PASS"
    on_failure: continue
    accumulate: dev_history
```

---

## Special Modes

### Template Mode

Call another workflow from within a step.

**Schema:**

```yaml
template:
  name: string                 # Workflow name
  with:                        # Input parameters
    key: value
```

**Example:**

```yaml
steps:
  - name: analyze
    template:
      name: code_analyzer
      with:
        code: "{{input}}"
        language: "python"
```

**Context Isolation:**

Each template call runs in an isolated context. Parent variables are not automatically available.

---

### Embeddings Mode

Generate vector embeddings from text.

**Schema:**

```yaml
embeddings:
  model: string                # Embedding model
  input: string | array        # Text to embed
```

**Example:**

```yaml
steps:
  - name: embed_docs
    embeddings:
      model: text-embedding-3-large
      input: "{{documents}}"
```

---

### Consensus Mode

Execute the same prompt across multiple providers and require agreement.

**Schema:**

```yaml
consensus:
  prompt: string               # The question/prompt
  executions:                  # Provider configurations
    - provider: string
      model: string
      temperature: float       # Optional override
      max_tokens: int          # Optional override
      timeout: duration        # Optional override
  require: string              # "unanimous", "2/3", "majority"
  allow_partial: boolean       # Allow partial failures
  timeout: duration            # Overall timeout
```

**Example:**

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Is this code safe? YES or NO"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
      require: unanimous
      timeout: 60s
```

**Result Format:**

```json
{
  "success": true,
  "result": "YES",
  "agreement": 1.0,
  "votes": {
    "anthropic/claude-sonnet-4": "YES",
    "openai/gpt-4o": "YES",
    "deepseek/deepseek-chat": "YES"
  },
  "confidence": "high"
}
```

---

## Property Inheritance

Properties are inherited from workflow → step → consensus execution.

**Resolution Order:**

1. **Consensus Execution** (most specific)
2. **Step** (overrides execution)
3. **Execution Context** (workflow defaults)
4. **mcp-cli defaults** (fallback)

**Example:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  servers: [filesystem, brave-search]
  logging: verbose

steps:
  # Inherits all from execution
  - name: step1
    run: "Analyze: {{input}}"
    # Uses: anthropic/claude-sonnet-4, temp=0.7, servers=[...], logging=verbose

  # Overrides temperature
  - name: step2
    run: "Generate code"
    temperature: 0.3
    # Uses: anthropic/claude-sonnet-4, temp=0.3, servers=[...], logging=verbose

  # Overrides provider and servers
  - name: step3
    run: "Search web"
    provider: openai
    model: gpt-4o
    servers: [brave-search]
    # Uses: openai/gpt-4o, temp=0.7, servers=[brave-search], logging=verbose
```

---

## Loop System

Loops enable iterative execution with LLM-evaluated exit conditions.

### How Loops Work

1. **Initialization:** Loop starts with `iteration = 1`
2. **Execution:** Workflow called with interpolated `with` parameters
3. **Evaluation:** LLM evaluates the `until` condition
4. **Decision:**
   - If condition met → Exit loop
   - If max iterations → Exit loop
   - Otherwise → Continue to next iteration
5. **Storage:** Results stored in loop variable and accumulator

### Loop Variables

Available in `with` parameters:

```yaml
with:
  iteration_number: "{{loop.iteration}}"
  current_output: "{{loop.output}}"
  previous_output: "{{loop.last.output}}"
  full_history: "{{loop.history}}"
```

### Exit Conditions

The `until` condition is evaluated by an LLM after each iteration:

**Good conditions (clear):**

```yaml
until: "The output says PASS"
until: "The review contains no errors"
until: "The score is above 90"
```

**Avoid (ambiguous):**

```yaml
until: "Check if {{loop.output}} is correct"  # Don't interpolate into condition
```

### Accessing Loop Results

After loop completion:

```yaml
steps:
  - name: report
    needs: [my_loop]
    run: |
      Loop completed in {{loop.iteration}} iterations.
      Final result: {{my_loop}}
      Full history: {{loop_history}}
```

---

## Examples

### Basic Workflow

```yaml
$schema: "workflow/v2.0"
name: simple_analysis
version: 1.0.0
description: Analyze input and generate report

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: analyze
    run: "Analyze: {{input}}"

  - name: report
    needs: [analyze]
    run: "Create report from: {{analyze}}"
```

### With Failover

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: llama3.2
  temperature: 0.7
```

### Iterative Development

```yaml
$schema: "workflow/v2.0"
name: iterative_coder
version: 1.0.0
description: Iteratively develop code until tests pass

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: requirements
    run: "Analyze: {{input}}"

  - name: tests
    needs: [requirements]
    run: "Create tests for: {{requirements}}"

loops:
  - name: develop
    workflow: write_and_test
    with:
      spec: "{{requirements}}"
      tests: "{{tests}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "The output says PASS"
    on_failure: continue

steps:
  - name: report
    needs: [develop]
    run: "Completed in {{loop.iteration}} iterations"
```

### Workflow Composition

```yaml
steps:
  - name: analyze
    template:
      name: code_analyzer
      with:
        code: "{{input}}"

  - name: review
    template:
      name: security_checker
      with:
        analysis: "{{analyze}}"
```

### Consensus Validation

```yaml
steps:
  - name: critical_decision
    consensus:
      prompt: "Approve deployment? YES or NO"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
        - provider: deepseek
          model: deepseek-chat
          temperature: 0
      require: unanimous
```

---

## Schema Validation

Required fields by object:

**WorkflowV2:**

- `name` ✓
- `version` ✓
- `description` ✓
- `execution` ✓ (must have provider config)

**StepV2:**

- `name` ✓
- One of: `run`, `template`, `embeddings`, `consensus` ✓

**LoopV2:**

- `name` ✓
- `workflow` ✓
- `max_iterations` ✓
- `until` ✓

**ConsensusExec:**

- `provider` ✓
- `model` ✓

---

## Migration from Template v1

Key changes:

1. **Schema identifier:** `$schema: "workflow/v2.0"` (was `$template: "v2"`)
2. **Property inheritance:** Explicit execution context
3. **Loops:** New top-level `loops` array
4. **Consensus:** Moved to step-level `consensus` mode
5. **Template calls:** Use `template: {name, with}` not `template: name`

See [MIGRATION.md](MIGRATION.md) for detailed migration guide.

---

## See Also

- [Authoring Guide](AUTHORING_GUIDE.md) - How to write workflows
- [Loop Guide](LOOPS.md) - Complete loop documentation
- [Examples](examples/) - Working workflow examples
- [Patterns](patterns/) - Common design patterns

---

**Last Updated:** January 7, 2026  
**Schema Version:** workflow/v2.0
