# Property Inheritance Guide

**Version:** workflow/v2.0  
**Purpose:** Visual diagrams showing how properties flow through workflow hierarchy

---

## Core Principle

> Properties are defined once at the workflow level and inherited by all steps. Steps can override any property. Consensus executions inherit from their step and can override per-execution.

---

## Three-Level Hierarchy

```
┌─────────────────────────────────────────┐
│  LEVEL 1: workflow.execution            │
│  ├─ provider: anthropic                 │
│  ├─ model: claude-sonnet-4              │
│  ├─ temperature: 0.7                    │
│  ├─ max_tokens: 2000                    │
│  ├─ servers: [filesystem]               │
│  └─ timeout: 60s                        │
└──────────────┬──────────────────────────┘
               │ ALL properties inherited
               ↓
┌─────────────────────────────────────────┐
│  LEVEL 2: steps[]                       │
│  ├─ name: "analyze"                     │
│  ├─ provider: ← inherited               │
│  ├─ model: ← inherited                  │
│  ├─ temperature: 0.3 ← OVERRIDE         │
│  ├─ max_tokens: ← inherited             │
│  ├─ servers: ← inherited                │
│  └─ timeout: ← inherited                │
└──────────────┬──────────────────────────┘
               │ (only for consensus mode)
               ↓
┌─────────────────────────────────────────┐
│  LEVEL 3: consensus.executions[]        │
│  ├─ provider: openai ← OVERRIDE         │
│  ├─ model: gpt-4o ← OVERRIDE            │
│  ├─ temperature: ← inherited from step  │
│  ├─ max_tokens: ← inherited from step   │
│  ├─ servers: ← inherited from step      │
│  └─ timeout: ← inherited from step      │
└─────────────────────────────────────────┘
```

---

## Inheritance Flow Diagram

### Simple Step (No Overrides)

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  servers: [filesystem]

steps:
  - name: analyze
    run: "Analyze {{input}}"
```

**Resolved configuration:**

```
┌──────────────────────────────┐
│ Step "analyze" executes as:  │
│                              │
│  provider: anthropic         │ ← from execution
│  model: claude-sonnet-4      │ ← from execution
│  temperature: 0.7            │ ← from execution
│  servers: [filesystem]       │ ← from execution
└──────────────────────────────┘

      ↓ maps to CLI

mcp-cli query \
  --provider anthropic \
  --model claude-sonnet-4 \
  --temperature 0.7 \
  --server filesystem \
  --input-data "Analyze user input"
```

---

### Step with Override

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  servers: [filesystem]

steps:
  - name: creative
    temperature: 1.5              # Override
    max_tokens: 4000              # Override
    run: "Generate ideas"
```

**Resolved configuration:**

```
┌─────────────────────────────────┐
│ Step "creative" executes as:    │
│                                 │
│  provider: anthropic            │ ← from execution
│  model: claude-sonnet-4         │ ← from execution
│  temperature: 1.5               │ ← OVERRIDDEN at step
│  max_tokens: 4000               │ ← OVERRIDDEN at step
│  servers: [filesystem]          │ ← from execution
└─────────────────────────────────┘

      ↓ maps to CLI

mcp-cli query \
  --provider anthropic \
  --model claude-sonnet-4 \
  --temperature 1.5 \
  --max-tokens 4000 \
  --server filesystem \
  --input-data "Generate ideas"
```

---

### Consensus Mode (Three-Level Inheritance)

```yaml
execution:
  provider: anthropic              # Level 1: workflow default
  temperature: 0.7
  servers: [filesystem]

steps:
  - name: validate
    temperature: 0.2               # Level 2: step override
    consensus:
      prompt: "Is this safe?"
      executions:
        - provider: anthropic      # Level 3a: execution-specific
          model: claude-sonnet-4
          # Inherits temperature: 0.2 from step

        - provider: openai         # Level 3b: execution-specific
          model: gpt-4o
          temperature: 0.1         # Level 3b: override

        - provider: deepseek       # Level 3c: execution-specific
          model: deepseek-chat
          # Inherits temperature: 0.2 from step
      require: 2/3
```

**Resolved configuration for each execution:**

```
┌─────────────────────────────────┐
│ Execution 1 (anthropic):        │
│  provider: anthropic            │ ← from execution
│  model: claude-sonnet-4         │ ← from execution
│  temperature: 0.2               │ ← from step (overrode workflow)
│  servers: [filesystem]          │ ← from workflow
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Execution 2 (openai):           │
│  provider: openai               │ ← from execution
│  model: gpt-4o                  │ ← from execution
│  temperature: 0.1               │ ← from execution (overrode step)
│  servers: [filesystem]          │ ← from workflow
└─────────────────────────────────┘

┌─────────────────────────────────┐
│ Execution 3 (deepseek):         │
│  provider: deepseek             │ ← from execution
│  model: deepseek-chat           │ ← from execution
│  temperature: 0.2               │ ← from step (overrode workflow)
│  servers: [filesystem]          │ ← from workflow
└─────────────────────────────────┘

All three execute in parallel:
  mcp-cli query --provider anthropic --temperature 0.2 "Is this safe?"
  mcp-cli query --provider openai --temperature 0.1 "Is this safe?"
  mcp-cli query --provider deepseek --temperature 0.2 "Is this safe?"

Then consensus evaluates: 2 of 3 must agree
```

---

## Property Override Rules

### Which Properties Can Be Overridden

| Property      | Override at Step | Override at Consensus Execution | Notes                              |
| ------------- | ---------------- | ------------------------------- | ---------------------------------- |
| `provider`    | ✅ Yes            | ✅ Yes                           | Complete provider switch           |
| `model`       | ✅ Yes            | ✅ Yes                           | Model selection per step/execution |
| `temperature` | ✅ Yes            | ✅ Yes                           | Most commonly overridden           |
| `max_tokens`  | ✅ Yes            | ✅ Yes                           | Response length control            |
| `servers`     | ✅ Yes            | ✅ Yes                           | MCP server availability            |
| `timeout`     | ✅ Yes            | ✅ Yes                           | Per-step timeout control           |
| `logging`     | ❌ No             | ❌ No                            | Workflow-level only                |
| `no_color`    | ❌ No             | ❌ No                            | Workflow-level only                |

---

## Inheritance Precedence

**Highest to Lowest Priority:**

```
1. consensus.executions[].property  (most specific)
         ↓
2. step.property                    (step override)
         ↓
3. workflow.execution.property      (default)
```

**Example:**

```yaml
execution:
  temperature: 0.7          # Priority 3: default

steps:
  - name: step1
    temperature: 0.5        # Priority 2: overrides workflow
    consensus:
      prompt: "Check"
      executions:
        - provider: anthropic
          # Uses temperature: 0.5 (from step)

        - provider: openai
          temperature: 0.3  # Priority 1: overrides step
```

**Resolution:**

- Execution 1 uses temperature `0.5` (from step)
- Execution 2 uses temperature `0.3` (from execution, highest priority)

---

## Provider Failover Inheritance

### Single Provider Mode

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    # Inherits provider and model
```

**Execution:**

```
Step "step1" → anthropic/claude-sonnet-4
```

---

### Failover Chain Mode

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: qwen2.5:32b

steps:
  - name: step1
    # Inherits entire failover chain
```

**Execution:**

```
Step "step1" tries in order:
  1. anthropic/claude-sonnet-4
  2. openai/gpt-4o (if 1 fails)
  3. ollama/qwen2.5:32b (if 2 fails)
```

---

### Override Failover at Step

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o

steps:
  - name: step1
    # Uses failover chain

  - name: step2
    provider: deepseek      # Override with single provider
    model: deepseek-chat
```

**Execution:**

```
Step "step1":
  1. anthropic/claude-sonnet-4
  2. openai/gpt-4o (failover)

Step "step2":
  1. deepseek/deepseek-chat (no failover)
```

---

## MCP Server Inheritance

### Workflow-Level Servers

```yaml
execution:
  servers: [filesystem, brave-search]

steps:
  - name: step1
    # Inherits both servers

  - name: step2
    # Inherits both servers
```

**Both steps have access to filesystem and brave-search.**

---

### Step-Level Server Override

```yaml
execution:
  servers: [filesystem, brave-search]

steps:
  - name: step1
    # Inherits: [filesystem, brave-search]

  - name: step2
    servers: [sqlite]       # Override completely
```

**Result:**

- Step 1: filesystem, brave-search
- Step 2: sqlite only (no inheritance)

**Note:** Server override is complete replacement, not additive.

---

## Embeddings Mode Inheritance

```yaml
execution:
  provider: openai
  model: gpt-4o

steps:
  - name: embed_docs
    embeddings:
      model: text-embedding-3-small  # Model override for embeddings
      input: ["text1", "text2"]
      # provider: openai (inherited from execution)
```

**Resolution:**

```
Embeddings step uses:
  provider: openai                  ← inherited from execution
  model: text-embedding-3-small     ← overridden for embeddings
```

**Note:** Embeddings mode can override `provider` and `model` separately from the step level.

---

## Workflow Mode Inheritance

```yaml
execution:
  provider: anthropic
  servers: [filesystem]

steps:
  - name: call_workflow
    template:
      name: code_reviewer
      with:
        code: "{{input}}"
```

**Inheritance behavior:**

```
┌────────────────────────────────┐
│ Parent Workflow                │
│  provider: anthropic           │
│  servers: [filesystem]         │
└────────────┬───────────────────┘
             │
             │ Template call
             ↓
┌────────────────────────────────┐
│ code_reviewer Workflow         │
│  execution:                    │
│    provider: openai  ← USES OWN│
│    servers: [...]    ← USES OWN│
└────────────────────────────────┘
```

**Important:** Template calls do NOT inherit properties. Each workflow uses its own `execution` configuration.

---

## Complete Inheritance Example

```yaml
$schema: "workflow/v2.0"
name: inheritance_demo
version: 1.0.0
description: Demonstrates all inheritance patterns

execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  max_tokens: 2000
  servers: [filesystem]
  timeout: 60s

steps:
  # Step 1: Uses all defaults
  - name: analyze
    run: "Analyze: {{input}}"
    # provider: anthropic (inherited)
    # model: claude-sonnet-4 (inherited)
    # temperature: 0.7 (inherited)
    # max_tokens: 2000 (inherited)
    # servers: [filesystem] (inherited)

  # Step 2: Override temperature only
  - name: creative
    temperature: 1.5                    # Override
    run: "Generate creative ideas"
    # provider: anthropic (inherited)
    # model: claude-sonnet-4 (inherited)
    # max_tokens: 2000 (inherited)
    # servers: [filesystem] (inherited)

  # Step 3: Override provider and model
  - name: search
    provider: openai                    # Override
    model: gpt-4o                       # Override
    run: "Search for: {{query}}"
    # temperature: 0.7 (inherited)
    # max_tokens: 2000 (inherited)
    # servers: [filesystem] (inherited)

  # Step 4: Three-level inheritance with consensus
  - name: validate
    temperature: 0.2                    # Step-level override
    consensus:
      prompt: "Is this valid?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          # temperature: 0.2 (from step)
          # servers: [filesystem] (from workflow)

        - provider: openai
          model: gpt-4o
          temperature: 0.1              # Execution-level override
          # servers: [filesystem] (from workflow)

        - provider: deepseek
          model: deepseek-chat
          # temperature: 0.2 (from step)
          # servers: [filesystem] (from workflow)
      require: 2/3
```

**Execution Summary:**

| Step       | Provider  | Model           | Temperature | Servers      | Source                    |
| ---------- | --------- | --------------- | ----------- | ------------ | ------------------------- |
| analyze    | anthropic | claude-sonnet-4 | 0.7         | [filesystem] | All inherited             |
| creative   | anthropic | claude-sonnet-4 | 1.5         | [filesystem] | Temperature overridden    |
| search     | openai    | gpt-4o          | 0.7         | [filesystem] | Provider/model overridden |
| validate-1 | anthropic | claude-sonnet-4 | 0.2         | [filesystem] | Step temp override        |
| validate-2 | openai    | gpt-4o          | 0.1         | [filesystem] | Exec temp override        |
| validate-3 | deepseek  | deepseek-chat   | 0.2         | [filesystem] | Step temp override        |

---

## Troubleshooting

### Issue: Property Not Inheriting

**Problem:**

```yaml
execution:
  provider: anthropic

steps:
  - name: step1
    run: "Query"
    # Why isn't provider being used?
```

**Check:**

1. Verify property name spelling (case-sensitive)
2. Ensure no typos in YAML
3. Check indentation (YAML is whitespace-sensitive)

---

### Issue: Unexpected Override

**Problem:**

```yaml
execution:
  temperature: 0.7

steps:
  - name: step1
    temperature: 0.5
    consensus:
      executions:
        - provider: anthropic
          # Expected 0.5, but getting 0.7?
```

**Solution:** Step-level properties only inherit to consensus executions if the execution doesn't override them. Check if execution has its own temperature set.

---

### Issue: Template Not Inheriting

**Problem:**

```yaml
execution:
  servers: [filesystem]

steps:
  - name: call_other
    template:
      name: other_workflow
    # Why doesn't other_workflow have filesystem access?
```

**Solution:** Template calls DON'T inherit properties. Each workflow uses its own `execution` configuration. Add servers to the called workflow's YAML.

---

## Best Practices

### 1. Define Common Defaults at Workflow Level

```yaml
# ✅ Good
execution:
  provider: anthropic
  temperature: 0.7

steps:
  - name: step1
    run: "Query"
  - name: step2
    run: "Query"
```

```yaml
# ❌ Bad - Repetition
steps:
  - name: step1
    provider: anthropic
    temperature: 0.7
    run: "Query"
  - name: step2
    provider: anthropic
    temperature: 0.7
    run: "Query"
```

---

### 2. Override Only When Necessary

```yaml
# ✅ Good - Clear intent
execution:
  temperature: 0.7

steps:
  - name: precise
    temperature: 0.2        # Override for specific need
    run: "Analyze data"
```

---

### 3. Use Descriptive Names for Override Steps

```yaml
# ✅ Good
steps:
  - name: creative_brainstorm
    temperature: 1.5
    run: "Generate ideas"

  - name: analytical_review
    temperature: 0.2
    run: "Review results"
```

---

## See Also

- **[Object Model](OBJECT_MODEL.md)** - TypeScript interface definitions
- **[Quick Reference](QUICK_REFERENCE.md)** - One-page overview
- **[CLI Mapping](CLI_MAPPING.md)** - Property → CLI argument mapping
- **[Steps Reference](STEPS_REFERENCE.md)** - Detailed step modes

---

**Key Takeaway:** Properties flow down through the hierarchy. Define once at the top, override only when needed at lower levels. This eliminates repetition and makes workflows maintainable.
