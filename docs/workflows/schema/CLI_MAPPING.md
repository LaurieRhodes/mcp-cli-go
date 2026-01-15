# CLI Property Mapping

**Core Concept:** Every workflow property maps directly to an mcp-cli command-line argument.

Understanding this mapping is key to understanding workflows: **workflows are sequences of mcp-cli calls with shared configuration**.

---

## Table of Contents

- [Base Query Object](#base-query-object)
- [Inheritance Chain](#inheritance-chain)
- [Complete Property Reference](#complete-property-reference)
- [Execution Mode Mapping](#execution-mode-mapping)
- [Examples](#examples)

---

## Base Query Object

All steps inherit from the `MCPQuery` interface, which maps directly to mcp-cli arguments:

```typescript
interface MCPQuery {
  provider: string;              // --provider
  model: string;                 // --model
  temperature?: number;          // --temperature
  max_tokens?: number;           // --max-tokens
  servers?: string[];            // --server (repeated for each)
  skills?: string[];             // --skills (repeated for each)
  timeout?: string;              // (internal timeout setting)
}
```

**This is the foundation:** Every workflow step is fundamentally an mcp-cli call with these properties.

---

## Inheritance Chain

### Visual Flow

```
workflow.execution (provides defaults)
         ↓
    steps[] (inherits, can override)
         ↓
consensus.executions[] (inherits from step, can override)
```

### How It Works

1. **Workflow level:** Define once
   
   ```yaml
   execution:
     provider: anthropic
     model: claude-sonnet-4
     temperature: 0.7
   ```

2. **Step level:** Inherits all, can override any
   
   ```yaml
   steps:
     - name: step1
       # Inherits: provider, model, temperature
       run: "Prompt"
   
     - name: step2
       temperature: 0.3        # Override just this property
       run: "Prompt"
   ```

3. **Consensus level:** Inherits from step, can override per-execution
   
   ```yaml
   steps:
     - name: validate
       consensus:
         prompt: "Validate"
         executions:
           - provider: anthropic  # Override for this execution
             # Inherits: model, temperature from step
           - provider: openai     # Override differently
             model: gpt-4o        # Override model too
   ```

---

## Complete Property Reference

### Core Properties (All Modes)

| YAML Property | CLI Argument            | Type     | Default    | Inheritable | Description                                              |
| ------------- | ----------------------- | -------- | ---------- | ----------- | -------------------------------------------------------- |
| `provider`    | `--provider <name>`     | string   | (required) | ✅ Yes       | AI provider: anthropic, openai, deepseek, ollama, gemini |
| `model`       | `--model <name>`        | string   | (required) | ✅ Yes       | Model identifier: claude-sonnet-4, gpt-4o, etc.          |
| `temperature` | `--temperature <float>` | float    | 0.7        | ✅ Yes       | Randomness: 0.0 (deterministic) to 2.0 (creative)        |
| `max_tokens`  | `--max-tokens <int>`    | int      | (auto)     | ✅ Yes       | Maximum tokens in response                               |
| `servers`     | `--server <name>`       | string[] | []         | ✅ Yes       | MCP servers: filesystem, brave-search, etc.              |
| `skills`      | `--skills <n>`       | string[] | []         | ✅ Yes       | Anthropic Skills: docx, pdf, xlsx, pptx, etc.            |
| `timeout`     | (internal)              | duration | 60s        | ✅ Yes       | How long to wait: "30s", "5m", "1h"                      |

### Workflow-Only Properties

| YAML Property | CLI Argument            | Type   | Default    | Inheritable | Description                  |
| ------------- | ----------------------- | ------ | ---------- | ----------- | ---------------------------- |
| `$schema`     | N/A                     | string | (required) | ❌ No        | Always "workflow/v2.0"       |
| `name`        | `--workflow <name>`     | string | (required) | ❌ No        | Workflow identifier          |
| `version`     | N/A                     | string | (required) | ❌ No        | Semantic version             |
| `description` | N/A                     | string | (required) | ❌ No        | What workflow does           |
| `logging`     | `--verbose` / `--noisy` | string | normal     | ⚠️ Partial  | "normal", "verbose", "noisy" |
| `no_color`    | `--no-color`            | bool   | false      | ⚠️ Partial  | Disable colored output       |

### Step-Only Properties

| YAML Property | CLI Argument          | Type     | Default      | Inheritable | Description          |
| ------------- | --------------------- | -------- | ------------ | ----------- | -------------------- |
| `name`        | N/A                   | string   | (required)   | ❌ No        | Step identifier      |
| `needs`       | N/A                   | string[] | []           | ❌ No        | Wait for these steps |
| `condition`   | N/A                   | string   | (none)       | ❌ No        | Skip step if false   |
| `run`         | `--input-data <text>` | string   | (choose one) | ❌ No        | LLM prompt           |
| `template`    | `--workflow <name>`   | object   | (choose one) | ❌ No        | Call workflow        |
| `embeddings`  | `--embeddings`        | object   | (choose one) | ❌ No        | Generate embeddings  |
| `consensus`   | (parallel calls)      | object   | (choose one) | ❌ No        | Multi-provider       |
| `rag`         | `mcp-cli rag search`  | object   | (choose one) | ❌ No        | RAG retrieval        |
| `loop`        | (iterative workflow)  | object   | (choose one) | ❌ No        | Iterate over items   |

**Additional Step Properties:**

| YAML Property | CLI Argument          | Type     | Default      | Inheritable | Description                 |
| ------------- | --------------------- | -------- | ------------ | ----------- | --------------------------- |
| `execution_order` | N/A               | int      | (none)       | ❌ No        | Manual execution order      |
| `input`       | (passed as JSON)      | any      | (none)       | ❌ No        | Direct input data           |
| `providers`   | (failover sequence)   | array    | (none)       | ✅ Yes       | Step-level provider failover |
| `logging`     | `--verbose`/`--noisy` | string   | (inherited)  | ✅ Yes       | Step-level logging override  |
| `no_color`    | `--no-color`          | bool     | (inherited)  | ✅ Yes       | Step-level color override    |

### Environment Variables

| YAML Property | CLI Argument | Type   | Default | Inheritable | Description                     |
| ------------- | ------------ | ------ | ------- | ----------- | ------------------------------- |
| `env`         | N/A          | object | {}      | ✅ Yes       | Key-value environment variables |

---


### Example 6: Execution Order Control

**YAML:**

```yaml
steps:
  - name: step_c
    execution_order: 3
    run: "Third step"
  
  - name: step_a
    execution_order: 1
    run: "First step"
  
  - name: step_b
    execution_order: 2
    run: "Second step"
```

**CLI Execution (respects execution_order):**

```bash
# Step 1: step_a executes first
mcp-cli --input-data "First step"

# Step 2: step_b executes second
mcp-cli --input-data "Second step"

# Step 3: step_c executes third
mcp-cli --input-data "Third step"
```

**Key:** Execution order overrides definition order in YAML.

---

### Example 7: Direct Input Data

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: configure
    input:
      database: postgres
      max_connections: 100
      timeout: 30
    run: |
      Configure database:
      - Type: {{input.database}}
      - Connections: {{input.max_connections}}
      - Timeout: {{input.timeout}}s
```

**CLI Execution:**

```bash
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data '{
    "database": "postgres",
    "max_connections": 100,
    "timeout": 30
  }' \
  --prompt "Configure database: ..."
```

**Key:** Complex input data is passed as JSON, accessible via {{input.key}}.

---

### Example 8: Step-Level Provider Failover

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: critical
    providers:
      - provider: anthropic
        model: claude-opus-4
      - provider: openai
        model: gpt-4o
    run: "Critical analysis"
  
  - name: normal
    run: "Regular task"
```

**CLI Execution:**

```bash
# Critical step: Try Opus first, fall back to OpenAI
mcp-cli --provider anthropic --model claude-opus-4 \
  --input-data "Critical analysis"
# If fails, retry with:
mcp-cli --provider openai --model gpt-4o \
  --input-data "Critical analysis"

# Normal step: Uses default provider
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data "Regular task"
```

**Key:** Step-level failover chains override workflow defaults.

---

## Execution Mode Mapping

### Mode 1: LLM Query (`run:`)

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: analyze
    run: "Analyze this: {{input}}"
```

**CLI Equivalent:**

```bash
mcp-cli \
  --provider anthropic \
  --model claude-sonnet-4 \
  --temperature 0.7 \
  --input-data "Analyze this: user input"
```

---

### Mode 2: Workflow Call (`template:`)

**YAML:**

```yaml
execution:
  provider: anthropic

steps:
  - name: review
    template:
      name: code_reviewer
      with:
        code: "{{input}}"
```

**CLI Equivalent:**

```bash
mcp-cli \
  --workflow code_reviewer \
  --provider anthropic \
  --input-data '{"code": "user input"}'
```

---

### Mode 3: Embeddings (`embeddings:`)

**YAML:**

```yaml
execution:
  provider: openai
  model: text-embedding-3-small

steps:
  - name: embed
    embeddings:
      input: ["text1", "text2"]
      chunk_strategy: sentence
      max_chunk_size: 512
      overlap: 50
```

**CLI Equivalent:**

```bash
mcp-cli embeddings \
  --provider openai \
  --model text-embedding-3-small \
  --chunk-strategy sentence \
  --max-chunk-size 512 \
  --overlap 50 \
  "text1" "text2"
```

**Note:** All embeddings CLI flags are now supported in workflow YAML.

---

### Mode 4: Consensus (`consensus:`)

**YAML:**

```yaml
execution:
  temperature: 0.2

steps:
  - name: validate
    consensus:
      prompt: "Is this safe? {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3
```

**CLI Equivalent (parallel execution):**

```bash
# Execution 1
mcp-cli \
  --provider anthropic \
  --model claude-sonnet-4 \
  --temperature 0.2 \
  --input-data "Is this safe? user input"

# Execution 2 (parallel)
mcp-cli \
  --provider openai \
  --model gpt-4o \
  --temperature 0.2 \
  --input-data "Is this safe? user input"

# Execution 3 (parallel)
mcp-cli \
  --provider deepseek \
  --model deepseek-chat \
  --temperature 0.2 \
  --input-data "Is this safe? user input"

# Then: Evaluate consensus (2 of 3 must agree)
```

---

## Examples

### Example 1: Basic Inheritance

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "Prompt 1"

  - name: step2
    run: "Prompt 2"
```

**CLI Execution:**

```bash
# Step 1 inherits all execution properties
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.7 \
  --input-data "Prompt 1"

# Step 2 inherits all execution properties
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.7 \
  --input-data "Prompt 2"
```

**Key:** Both steps use same configuration without repetition.

---

### Example 2: Property Override

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: creative
    temperature: 1.5          # Override
    run: "Generate creative ideas"

  - name: analytical
    temperature: 0.2          # Override
    run: "Analyze data precisely"
```

**CLI Execution:**

```bash
# Creative step: Override temperature
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 1.5 \
  --input-data "Generate creative ideas"

# Analytical step: Override temperature differently
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.2 \
  --input-data "Analyze data precisely"
```

**Key:** Each step overrides only what it needs to change.

---

### Example 3: MCP Server Access

**YAML:**

```yaml
execution:
  provider: anthropic
  servers: [filesystem, brave-search]

steps:
  - name: search
    run: "Search for: {{query}}"

  - name: read
    run: "Read file: {{filepath}}"
```

**CLI Execution:**

```bash
# Step 1: Has access to both MCP servers
mcp-cli --provider anthropic \
  --server filesystem \
  --server brave-search \
  --input-data "Search for: user query"

# Step 2: Has access to both MCP servers
mcp-cli --provider anthropic \
  --server filesystem \
  --server brave-search \
  --input-data "Read file: /path/to/file"
```

**Key:** MCP servers specified once, available to all steps.

---


---

### Example 4: Anthropic Skills

**YAML:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  skills: [docx, pdf, xlsx]

steps:
  - name: create_doc
    run: "Create a professional report as a Word document"

  - name: create_spreadsheet
    run: "Generate sales data in an Excel spreadsheet"
```

**CLI Execution:**

```bash
# Step 1: Has access to specified skills
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --server filesystem \
  --skills docx,pdf,xlsx \
  --input-data "Create a professional report as a Word document"

# Step 2: Has access to specified skills
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --server filesystem \
  --skills docx,pdf,xlsx \
  --input-data "Generate sales data in an Excel spreadsheet"
```

**Key:** Skills filter which Anthropic Skills are available to steps.

### Example 5: Provider Failover

**YAML:**

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
  - name: analyze
    run: "Analyze: {{input}}"
```

**CLI Execution (with automatic retry):**

```bash
# Try provider 1
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data "Analyze: user input"
# If fails or times out...

# Try provider 2
mcp-cli --provider openai --model gpt-4o \
  --input-data "Analyze: user input"
# If fails or times out...

# Try provider 3 (local fallback)
mcp-cli --provider ollama --model qwen2.5:32b \
  --input-data "Analyze: user input"
```

**Key:** Automatic failover without changing step definition.

---

### Example 5: Consensus with Overrides

**YAML:**

```yaml
execution:
  provider: anthropic        # Default
  temperature: 0.3           # Default

steps:
  - name: validate
    temperature: 0.2         # Override for all executions
    consensus:
      prompt: "Is this safe?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          # Inherits temperature: 0.2

        - provider: openai
          model: gpt-4o
          temperature: 0.1   # Override for just this execution

        - provider: deepseek
          model: deepseek-chat
          # Inherits temperature: 0.2
      require: unanimous
```

**CLI Execution:**

```bash
# Execution 1: Inherits step temperature
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.2 \
  --input-data "Is this safe?"

# Execution 2: Overrides temperature
mcp-cli --provider openai --model gpt-4o --temperature 0.1 \
  --input-data "Is this safe?"

# Execution 3: Inherits step temperature
mcp-cli --provider deepseek --model deepseek-chat --temperature 0.2 \
  --input-data "Is this safe?"

# Require: All 3 must agree (unanimous)
```

**Key:** Three-level inheritance: workflow → step → consensus execution.

---

## Property Override Rules

### Precedence (Highest to Lowest)

1. **Consensus execution level** (most specific)
   
   ```yaml
   consensus:
     executions:
       - temperature: 0.1    # Highest priority
   ```

2. **Step level**
   
   ```yaml
   steps:
     - name: step1
       temperature: 0.3      # Overrides workflow
   ```

3. **Workflow execution level** (default)
   
   ```yaml
   execution:
     temperature: 0.7        # Used if not overridden
   ```

### What Can Be Overridden

| Property      | Can Override at Step | Can Override at Consensus |
| ------------- | -------------------- | ------------------------- |
| `provider`    | ✅ Yes                | ✅ Yes                     |
| `model`       | ✅ Yes                | ✅ Yes                     |
| `temperature` | ✅ Yes                | ✅ Yes                     |
| `max_tokens`  | ✅ Yes                | ✅ Yes                     |
| `servers`     | ✅ Yes                | ✅ Yes                     |
| `timeout`     | ✅ Yes                | ✅ Yes                     |
| `logging`     | ❌ No                 | ❌ No                      |
| `no_color`    | ❌ No                 | ❌ No                      |

---

## Special Cases

### Case 1: Multiple MCP Servers

**YAML:**

```yaml
execution:
  servers: [filesystem, brave-search]
```

**CLI Equivalent:**

```bash
mcp-cli --server filesystem --server brave-search ...
```

**Note:** Each server is a separate `--server` flag.

---

### Case 2: Timeout Durations

**YAML accepts multiple formats:**

```yaml
timeout: "30s"      # 30 seconds
timeout: "5m"       # 5 minutes
timeout: "1h"       # 1 hour
timeout: "90s"      # 90 seconds
```

**Internal:** Converted to seconds for execution.

---

### Case 3: Temperature Bounds

**Valid range:** 0.0 to 2.0

```yaml
temperature: 0.0    # Most deterministic
temperature: 0.7    # Default balanced
temperature: 1.0    # Creative
temperature: 2.0    # Most random
```

**Out of bounds:** Will be rejected or clamped by provider.

---

### Case 4: Environment Variables

**YAML:**

```yaml
env:
  API_KEY: "secret_key"
  DEBUG: "true"
```

**Effect:** Available to all steps, but not directly passed as CLI args. Used internally by workflow executor.

---

## Debugging Tips

### See What Gets Executed

Run with `--verbose` to see actual mcp-cli commands:

```bash
mcp-cli --workflow my_workflow --verbose
```

Output will show:

```
[DEBUG] Step 'analyze' executing with:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  input: "Analyze this: user input"
```

### Verify Property Inheritance

Use `--noisy` for even more detail:

```bash
mcp-cli --workflow my_workflow --noisy
```

Shows property resolution:

```
[TRACE] Step 'analyze':
  provider: 'anthropic' (inherited from execution)
  model: 'claude-sonnet-4' (inherited from execution)
  temperature: 0.3 (overridden at step level)
```

---

## Common Patterns

### Pattern 1: Same Config, Different Prompts

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "Prompt 1"
  - name: step2
    run: "Prompt 2"
  - name: step3
    run: "Prompt 3"
```

All use identical provider/model without repetition.

---

### Pattern 2: Different Models, Same Provider

```yaml
execution:
  provider: anthropic

steps:
  - name: fast
    model: claude-haiku-4
    run: "Quick task"

  - name: smart
    model: claude-sonnet-4
    run: "Complex task"

  - name: best
    model: claude-opus-4
    run: "Critical task"
```

Provider shared, models optimized per task.

---

### Pattern 3: Different Providers per Step

```yaml
execution:
  temperature: 0.7          # Shared parameter

steps:
  - name: analysis
    provider: anthropic
    model: claude-sonnet-4
    run: "Analyze"

  - name: search
    provider: openai
    model: gpt-4o
    run: "Search"

  - name: summary
    provider: deepseek
    model: deepseek-chat
    run: "Summarize"
```

Each step uses different provider but shares temperature.

---

## Summary

**Key Concepts:**

1. **Every workflow property maps to a CLI argument**
   
   - `provider` → `--provider`
   - `model` → `--model`
   - `temperature` → `--temperature`

2. **Inheritance saves repetition**
   
   - Define once at workflow level
   - All steps inherit
   - Override only what differs

3. **Three-level hierarchy**
   
   - workflow.execution (defaults)
   - steps[] (inherits, can override)
   - consensus.executions[] (inherits from step, can override)

4. **Workflows are sequences of mcp-cli calls**
   
   - Each step → one mcp-cli call
   - Consensus → multiple parallel calls
   - Loops → repeated workflow calls

**Remember:** If you can run it with `mcp-cli`, you can put it in a workflow. The workflow just handles the sequencing and property inheritance.

---

## See Also

- **[Quick Reference](QUICK_REFERENCE.md)** - One-page overview
- **[Full Schema](SCHEMA.md)** - Complete schema documentation
- **[Examples](../examples/)** - Working workflows
