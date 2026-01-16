# Workflow Schema Quick Reference

**Version:** workflow/v2.0  
**Core Concept:** Workflows sequence mcp-cli calls with shared configuration

---

## Table of Contents

### Core Objects

- [Workflow Object (Root)](#workflow-object-root)
- [ExecutionContext](#executioncontext)
- [Step Object](#step-object)

### Execution Modes

- [run: LLM Query](#mode-1-run---llm-query)
- [template: Workflow Call](#mode-2-template---workflow-call)
- [embeddings: Vector Generation](#mode-3-embeddings---vector-generation)
- [consensus: Multi-Provider Validation](#mode-4-consensus---multi-provider-validation)
- [rag: Vector Database Search](#mode-5-rag---vector-database-search)
- [loop: Iterative Execution](#mode-6-loop---iterative-execution)

### Configuration Details

- [LoopConfig Object](#loopconfig-object-step-level)
- [RagConfig Object](#ragconfig-object)
- [EmbeddingsConfig Object](#embeddingsconfig-object)
- [ConsensusConfig Object](#consensusconfig-object)

### Reference Tables

- [Allowed Values](#allowed-values-reference)
- [Validation Constraints](#validation-constraints)
- [Variable Interpolation](#variable-interpolation)

### Quick Lookups

- [Common Patterns](#common-patterns)
- [CLI Equivalents](#cli-equivalents)
- [Task Index](#task-index)

---

## Minimal Workflow

```yaml
$schema: "workflow/v2.0"
name: my_workflow
version: 1.0.0
description: What this workflow does

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: analyze
    run: "Analyze this: {{input}}"

  - name: summarize
    needs: [analyze]
    run: "Summarize: {{analyze}}"
```

---

## Allowed Values Reference

### Enum Properties

All properties with restricted values:

| Property                     | Allowed Values                                                                                      | Default      | Notes                                 |
| ---------------------------- | --------------------------------------------------------------------------------------------------- | ------------ | ------------------------------------- |
| `logging`                    | `"error"` \| `"warn"` \| `"info"` \| `"step"` \| `"steps"` \| `"debug"` \| `"verbose"` \| `"noisy"` | `"info"`     | "noisy" is legacy alias for "verbose" |
| `chunk_strategy`             | `"sentence"` \| `"paragraph"` \| `"fixed"` \| `"semantic"` \| `"sliding"`                           | `"sentence"` | For embeddings mode                   |
| `encoding_format`            | `"float"` \| `"base64"`                                                                             | `"float"`    | For embeddings mode                   |
| `output_format` (embeddings) | `"json"` \| `"csv"` \| `"compact"`                                                                  | `"json"`     | Embeddings output format              |
| `output_format` (RAG)        | `"json"` \| `"text"` \| `"compact"`                                                                 | `"json"`     | RAG output format                     |
| `fusion`                     | `"rrf"` \| `"weighted"` \| `"max"` \| `"avg"`                                                       | `"rrf"`      | RAG fusion method                     |
| `on_failure`                 | `"halt"` \| `"continue"` \| `"retry"`                                                               | `"halt"`     | Loop error handling                   |
| `on_error` (parallel)        | `"cancel_all"` \| `"complete_running"` \| `"continue"`                                              | `"cancel_all"` | Parallel execution error policy     |
| `require`                    | `"unanimous"` \| `"majority"` \| `"2/3"`                                                            | (required)   | Consensus agreement threshold         |
| `mode` (loop)                | `"iterate"` \| `"refine"`                                                                           | `"refine"`   | Loop execution mode                   |

### Logging Levels Explained

| Level                 | Description                | Use When                     |
| --------------------- | -------------------------- | ---------------------------- |
| `"error"`             | Errors only                | Production, quiet operation  |
| `"warn"`              | Errors + warnings          | Production, need alerts      |
| `"info"`              | Normal operation (default) | Standard use                 |
| `"step"` or `"steps"` | Step-level workflow events | Tracking workflow progress   |
| `"debug"`             | Detailed debugging info    | Development, troubleshooting |
| `"verbose"`           | All internal operations    | Deep debugging               |
| `"noisy"`             | Alias for "verbose"        | Legacy compatibility         |

### Chunk Strategy Explained

| Strategy      | Description                   | Best For                            |
| ------------- | ----------------------------- | ----------------------------------- |
| `"sentence"`  | Split on sentence boundaries  | Q&A systems, precise retrieval      |
| `"paragraph"` | Split on paragraph boundaries | Document similarity, longer context |
| `"fixed"`     | Fixed-size chunks             | Consistent chunk sizes              |
| `"semantic"`  | Semantically coherent chunks  | Complex documents                   |
| `"sliding"`   | Overlapping sliding window    | Dense information extraction        |

---

## Validation Constraints

### Numeric Constraints

| Property           | Type    | Range     | Default         | Required       |
| ------------------ | ------- | --------- | --------------- | -------------- |
| `temperature`      | float   | 0.0 - 2.0 | 0.7             | No             |
| `max_tokens`       | integer | > 0       | (auto)          | No             |
| `max_iterations`   | integer | > 0       | -               | For loops: Yes |
| `max_workers`      | integer | > 0       | 3               | No             |
| `top_k`            | integer | > 0       | 5               | No             |
| `min_success_rate` | float   | 0.0 - 1.0 | 0.0             | No             |
| `max_chunk_size`   | integer | > 0       | 512             | No             |
| `overlap`          | integer | ≥ 0       | 0               | No             |
| `dimensions`       | integer | > 0       | (model default) | No             |
| `max_retries`      | integer | ≥ 0       | 0               | No             |
| `execution_order`  | integer | any       | -               | No             |
| `query_variants`   | integer | > 0       | -               | No             |

### String Constraints

| Property           | Format          | Example                                     | Notes                  |
| ------------------ | --------------- | ------------------------------------------- | ---------------------- |
| `timeout`          | duration        | `"30s"`, `"5m"`, `"1h"`                     | Golang duration format |
| `retry_delay`      | duration        | `"5s"`, `"100ms"`                           | Golang duration format |
| `timeout_per_item` | duration        | `"30s"`                                     | Per-iteration timeout  |
| `total_timeout`    | duration        | `"1h"`                                      | Total loop timeout     |
| `items` (loop)     | URI or template | `file:///path/file.json`, `{{step_output}}` | Must be JSON array     |
| `$schema`          | literal         | `"workflow/v2.0"`                           | Must be exact          |
| `version`          | semver          | `"1.0.0"`, `"2.1.3-beta"`                   | Semantic versioning    |

### Array Constraints

| Property                 | Element Type     | Min Items | Notes                               |
| ------------------------ | ---------------- | --------- | ----------------------------------- |
| `servers`                | string           | 0         | MCP server names                    |
| `skills`                 | string           | 0         | Anthropic skill names               |
| `needs`                  | string           | 0         | Must reference existing steps       |
| `providers`              | ProviderFallback | 1         | At least one provider required      |
| `executions` (consensus) | ConsensusExec    | 2         | At least 2 for meaningful consensus |
| `strategies` (RAG)       | string           | 0         | RAG search strategies               |

---

## Property Inheritance Model

```
┌─────────────────────────────────┐
│ workflow.execution              │  ← Workflow defaults
│   provider: anthropic           │
│   model: claude-sonnet-4        │
│   temperature: 0.7              │
│   servers: [filesystem]         │
│   skills: [docx, xlsx]          │
│   logging: info                 │
└──────────────┬──────────────────┘
               │ (inherits all properties)
               ↓
┌─────────────────────────────────┐
│ steps[].{step properties}       │  ← Step can override
│   provider: ← inherited         │
│   model: ← inherited            │
│   temperature: 0.9 ← OVERRIDE   │
│   servers: ← inherited          │
│   skills: ← inherited           │
│   logging: verbose ← OVERRIDE   │
└──────────────┬──────────────────┘
               │ (for consensus mode only)
               ↓
┌─────────────────────────────────┐
│ consensus.executions[]          │  ← Per-execution override
│   provider: openai ← OVERRIDE   │
│   model: gpt-4o ← OVERRIDE      │
│   temperature: ← inherited      │
└─────────────────────────────────┘
```

**Key principle:** Define once at workflow level, override only where needed.

---

## Complete Object Schemas

### Workflow Object (Root)

| Property      | Type              | Required | Default | Description                        |
| ------------- | ----------------- | -------- | ------- | ---------------------------------- |
| `$schema`     | `"workflow/v2.0"` | Yes      | -       | Schema version identifier          |
| `name`        | string            | Yes      | -       | Unique workflow identifier         |
| `version`     | string (semver)   | Yes      | -       | Semantic version (e.g., `"1.0.0"`) |
| `description` | string            | Yes      | -       | Human-readable description         |
| `execution`   | ExecutionContext  | Yes      | -       | Workflow-level defaults            |
| `env`         | map[string]string | No       | `{}`    | Environment variables              |
| `steps`       | Step[]            | Yes      | -       | Array of steps to execute          |
| `loops`       | Loop[]            | No       | `[]`    | Array of top-level loops           |

---

### ExecutionContext

Provides default configuration inherited by all steps.

| Property                                        | Type                                                                                                | Required | Default  | Description                                                                      |
| ----------------------------------------------- | --------------------------------------------------------------------------------------------------- | -------- | -------- | -------------------------------------------------------------------------------- |
| **Provider Configuration (Option 1: Single)**   |                                                                                                     |          |          |                                                                                  |
| `provider`                                      | string                                                                                              | Yes*     | -        | AI provider: `anthropic`, `openai`, `deepseek`, `ollama`, `gemini`, `openrouter` |
| `model`                                         | string                                                                                              | Yes*     | -        | Model identifier (e.g., `claude-sonnet-4`, `gpt-4o`)                             |
| **Provider Configuration (Option 2: Failover)** |                                                                                                     |          |          |                                                                                  |
| `providers`                                     | ProviderFallback[]                                                                                  | Yes*     | -        | Provider failover chain (mutually exclusive with single provider)                |
| **Infrastructure**                              |                                                                                                     |          |          |                                                                                  |
| `servers`                                       | string[]                                                                                            | No       | `[]`     | MCP servers to enable (e.g., `[filesystem, brave-search]`)                       |
| `skills`                                        | string[]                                                                                            | No       | `[]`     | Anthropic skills to enable (e.g., `[docx, pdf, xlsx]`)                           |
| **Model Parameters**                            |                                                                                                     |          |          |                                                                                  |
| `temperature`                                   | float (0.0-2.0)                                                                                     | No       | 0.7      | Randomness: 0.0 (deterministic) to 2.0 (creative)                                |
| `max_tokens`                                    | integer (>0)                                                                                        | No       | (auto)   | Maximum tokens in response                                                       |
| **Execution Control**                           |                                                                                                     |          |          |                                                                                  |
| `timeout`                                       | duration                                                                                            | No       | `"60s"`  | Call timeout: `"30s"`, `"5m"`, `"1h"`                                            |
| `max_iterations`                                | integer (>0)                                                                                        | No       | -        | Global iteration safety limit                                                    |
| **Logging**                                     |                                                                                                     |          |          |                                                                                  |
| `logging`                                       | `"error"` \| `"warn"` \| `"info"` \| `"step"` \| `"steps"` \| `"debug"` \| `"verbose"` \| `"noisy"` | No       | `"info"` | Logging verbosity                                                                |
| `no_color`                                      | boolean                                                                                             | No       | false    | Disable colored output                                                           |
| **Parallel Execution (v2.1.0+)**                |                                                                                                     |          |          |                                                                                  |
| `parallel`                                      | boolean                                                                                             | No       | false    | Enable parallel step execution                                                   |
| `max_workers`                                   | integer (>0)                                                                                        | No       | 3        | Maximum concurrent steps                                                         |
| `on_error`                                      | `"cancel_all"` \| `"complete_running"` \| `"continue"`                                              | No       | `"cancel_all"` | Error handling policy for parallel execution                             |

\* Either (`provider` + `model`) OR `providers` is required.

---

### ProviderFallback

Used in provider failover chains.

| Property      | Type            | Required | Default     | Description                            |
| ------------- | --------------- | -------- | ----------- | -------------------------------------- |
| `provider`    | string          | Yes      | -           | AI provider name                       |
| `model`       | string          | Yes      | -           | Model identifier                       |
| `temperature` | float (0.0-2.0) | No       | (inherited) | Override temperature for this provider |
| `max_tokens`  | integer (>0)    | No       | (inherited) | Override max_tokens for this provider  |
| `timeout`     | duration        | No       | (inherited) | Override timeout for this provider     |

---

### Step Object

| Property                                               | Type               | Required | Default     | Description                                                  |
| ------------------------------------------------------ | ------------------ | -------- | ----------- | ------------------------------------------------------------ |
| **Identity**                                           |                    |          |             |                                                              |
| `name`                                                 | string             | Yes      | -           | Unique step identifier                                       |
| **Orchestration**                                      |                    |          |             |                                                              |
| `execution_order`                                      | integer            | No       | -           | Manual execution order (overrides dependency-based ordering) |
| `needs`                                                | string[]           | No       | `[]`        | Step dependencies - waits for these steps to complete        |
| `if`                                                   | string             | No       | -           | Skip step if expression evaluates to false                   |
| `input`                                                | any                | No       | -           | Direct input data for the step                               |
| **Inherited from ExecutionContext (can override any)** |                    |          |             |                                                              |
| `provider`                                             | string             | No       | (inherited) | Override provider for this step                              |
| `model`                                                | string             | No       | (inherited) | Override model for this step                                 |
| `providers`                                            | ProviderFallback[] | No       | (inherited) | Override provider failover chain                             |
| `temperature`                                          | float (0.0-2.0)    | No       | (inherited) | Override temperature for this step                           |
| `max_tokens`                                           | integer (>0)       | No       | (inherited) | Override max_tokens for this step                            |
| `servers`                                              | string[]           | No       | (inherited) | Override servers for this step                               |
| `skills`                                               | string[]           | No       | (inherited) | Override skills for this step                                |
| `timeout`                                              | duration           | No       | (inherited) | Override timeout for this step                               |
| `max_iterations`                                       | integer (>0)       | No       | (inherited) | Override max iterations for this step                        |
| `logging`                                              | enum               | No       | (inherited) | Override logging level (see Allowed Values table)            |
| `no_color`                                             | boolean            | No       | (inherited) | Override color output for this step                          |
| **Execution Mode (choose exactly ONE)**                |                    |          |             |                                                              |
| `run`                                                  | string             | No       | -           | LLM prompt with `{{variable}}` interpolation                 |
| `template`                                             | TemplateCall       | No       | -           | Call another workflow                                        |
| `embeddings`                                           | EmbeddingsConfig   | No       | -           | Generate vector embeddings                                   |
| `consensus`                                            | ConsensusConfig    | No       | -           | Multi-provider validation                                    |
| `rag`                                                  | RagConfig          | No       | -           | RAG retrieval from vector database                           |
| `loop`                                                 | LoopConfig         | No       | -           | Iterate over items calling a child workflow                  |

---

## Execution Modes

### Mode 1: run - LLM Query

**Purpose:** Execute a single LLM query with variable interpolation.

**Syntax:**

```yaml
- name: step_name
  run: string    # Prompt with {{variables}}
```

**Example:**

```yaml
- name: analyze
  run: "Analyze this code: {{input}}"
```

---

### Mode 2: template - Workflow Call

**Purpose:** Call another workflow as a subroutine.

**Syntax:**

```yaml
- name: step_name
  template:
    name: string               # Workflow name
    with: {key: value}         # Input data (optional)
```

**Example:**

```yaml
- name: review_code
  template:
    name: code_reviewer
    with:
      code: "{{input}}"
```

---

### Mode 3: embeddings - Vector Generation

**Purpose:** Generate vector embeddings from text.

**Syntax:** See [EmbeddingsConfig Object](#embeddingsconfig-object)

---

### Mode 4: consensus - Multi-Provider Validation

**Purpose:** Execute across multiple providers and require agreement.

**Syntax:** See [ConsensusConfig Object](#consensusconfig-object)

---

### Mode 5: rag - Vector Database Search

**Purpose:** Retrieve relevant documents from a vector database using semantic search.

**Syntax:** See [RagConfig Object](#ragconfig-object)

---

### Mode 6: loop - Iterative Execution

**Purpose:** Execute a child workflow repeatedly over a collection of items or until a condition is met.

**Syntax:** See [LoopConfig Object](#loopconfig-object-step-level)

---

## RagConfig Object

| Property                   | Type                                          | Required | Default       | Description                                          |
| -------------------------- | --------------------------------------------- | -------- | ------------- | ---------------------------------------------------- |
| **Query Configuration**    |                                               |          |               |                                                      |
| `query`                    | string                                        | Yes      | -             | Search query (supports `{{variables}}`)              |
| `query_vector`             | float[]                                       | No       | -             | Pre-computed vector (optional, alternative to query) |
| **Server Configuration**   |                                               |          |               |                                                      |
| `server`                   | string                                        | No       | (from config) | Single RAG server name (from `config/rag/*.yaml`)    |
| `servers`                  | string[]                                      | No       | -             | Multiple servers for fusion                          |
| **Strategy Configuration** |                                               |          |               |                                                      |
| `strategies`               | string[]                                      | No       | `["default"]` | Vector search strategies to use                      |
| `top_k`                    | integer (>0)                                  | No       | 5             | Number of results to return                          |
| **Fusion Configuration**   |                                               |          |               |                                                      |
| `fusion`                   | `"rrf"` \| `"weighted"` \| `"max"` \| `"avg"` | No       | `"rrf"`       | Result fusion method                                 |
| **Query Expansion**        |                                               |          |               |                                                      |
| `expand_query`             | boolean                                       | No       | false         | Enable query expansion (synonyms, acronyms)          |
| `query_variants`           | integer (>0)                                  | No       | -             | Number of query variants to generate                 |
| **Output Configuration**   |                                               |          |               |                                                      |
| `output_format`            | `"json"` \| `"text"` \| `"compact"`           | No       | `"json"`      | Output format                                        |

**Example:**

```yaml
steps:
  - name: retrieve_controls
    rag:
      query: "{{step.parse_text}}"
      server: pgvector
      strategies: [default]
      top_k: 5
      output_format: json

  - name: assess
    needs: [retrieve_controls]
    run: |
      Based on these controls: {{retrieve_controls}}
      Assess compliance of: {{input}}
```

---

## LoopConfig Object (Step-Level)

| Property               | Type                                  | Required | Default    | Description                                                    |
| ---------------------- | ------------------------------------- | -------- | ---------- | -------------------------------------------------------------- |
| **Core**               |                                       |          |            |                                                                |
| `workflow`             | string                                | Yes      | -          | Child workflow to execute                                      |
| `mode`                 | `"iterate"` \| `"refine"`             | No       | `"refine"` | Loop mode                                                      |
| `items`                | string                                | No*      | -          | Item source for iterate mode (e.g., `file:///path/items.json`) |
| `with`                 | map[string]any                        | No       | `{}`       | Input parameters passed to child workflow                      |
| **Iteration Control**  |                                       |          |            |                                                                |
| `max_iterations`       | integer (>0)                          | Yes      | -          | Maximum iterations (safety limit)                              |
| `until`                | string                                | No*      | -          | Exit condition for refine mode (LLM-evaluated)                 |
| **Error Handling**     |                                       |          |            |                                                                |
| `on_failure`           | `"halt"` \| `"continue"` \| `"retry"` | No       | `"halt"`   | Failure handling strategy                                      |
| `max_retries`          | integer (≥0)                          | No       | 0          | Retries per item (when `on_failure: retry`)                    |
| `retry_delay`          | duration                              | No       | -          | Delay between retries (e.g., `"5s"`)                           |
| **Success Criteria**   |                                       |          |            |                                                                |
| `min_success_rate`     | float (0.0-1.0)                       | No       | 0.0        | Minimum success rate to consider loop successful               |
| **Timeouts**           |                                       |          |            |                                                                |
| `timeout_per_item`     | duration                              | No       | -          | Timeout per iteration (e.g., `"30s"`)                          |
| `total_timeout`        | duration                              | No       | -          | Total loop timeout (e.g., `"1h"`)                              |
| **Parallel Execution** |                                       |          |            |                                                                |
| `parallel`             | boolean                               | No       | false      | Enable parallel processing                                     |
| `max_workers`          | integer (>0)                          | No       | 3          | Maximum concurrent workers                                     |
| **Output**             |                                       |          |            |                                                                |
| `accumulate`           | string                                | No       | -          | Variable name to store all iteration results                   |

\* `items` required for `mode: iterate`, `until` required for `mode: refine`

**Example - Iterate Mode (Parallel):**

```yaml
steps:
  - name: assess_statements
    loop:
      workflow: assess_single_statement
      mode: iterate
      items: file:///outputs/statements.json
      parallel: true
      max_workers: 15
      max_iterations: 500
      on_failure: continue
      min_success_rate: 0.8
```

**Example - Refine Mode:**

```yaml
steps:
  - name: fix_code
    loop:
      workflow: test_and_fix
      mode: refine
      with:
        code: "{{input}}"
        feedback: "{{loop.last.output}}"
      max_iterations: 10
      until: "All tests pass"
```

---

## EmbeddingsConfig Object

| Property                        | Type                                                                      | Required | Default      | Description                                      |
| ------------------------------- | ------------------------------------------------------------------------- | -------- | ------------ | ------------------------------------------------ |
| **Input Source (one required)** |                                                                           |          |              |                                                  |
| `input`                         | string \| string[]                                                        | Yes*     | -            | Text to embed (string or array of strings)       |
| `input_file`                    | string                                                                    | Yes*     | -            | Input file path (alternative to `input`)         |
| **Provider Override**           |                                                                           |          |              |                                                  |
| `provider`                      | string                                                                    | No       | (inherited)  | AI provider: `openai`, `deepseek`, `openrouter`  |
| `model`                         | string                                                                    | No       | (inherited)  | Embedding model (e.g., `text-embedding-3-small`) |
| **Chunking Configuration**      |                                                                           |          |              |                                                  |
| `chunk_strategy`                | `"sentence"` \| `"paragraph"` \| `"fixed"` \| `"semantic"` \| `"sliding"` | No       | `"sentence"` | Chunking strategy                                |
| `max_chunk_size`                | integer (>0)                                                              | No       | 512          | Maximum chunk size in tokens                     |
| `overlap`                       | integer (≥0)                                                              | No       | 0            | Overlap between chunks in tokens                 |
| **Model Configuration**         |                                                                           |          |              |                                                  |
| `dimensions`                    | integer (>0)                                                              | No       | (auto)       | Number of dimensions (for supported models)      |
| **Output Configuration**        |                                                                           |          |              |                                                  |
| `encoding_format`               | `"float"` \| `"base64"`                                                   | No       | `"float"`    | Encoding format                                  |
| `include_metadata`              | boolean                                                                   | No       | true         | Include chunk and model metadata                 |
| `output_format`                 | `"json"` \| `"csv"` \| `"compact"`                                        | No       | `"json"`     | Output format                                    |
| `output_file`                   | string                                                                    | No       | (stdout)     | Output file path                                 |

\* One of `input` or `input_file` is required

---

## ConsensusConfig Object

| Property     | Type                                     | Required | Default | Description                                             |
| ------------ | ---------------------------------------- | -------- | ------- | ------------------------------------------------------- |
| `prompt`     | string                                   | Yes      | -       | Prompt sent to all providers (supports `{{variables}}`) |
| `executions` | ConsensusExec[]                          | Yes      | -       | Array of provider configurations (≥2 recommended)       |
| `require`    | `"unanimous"` \| `"majority"` \| `"2/3"` | Yes      | -       | Agreement threshold                                     |
| `timeout`    | duration                                 | No       | `"60s"` | Timeout for entire consensus operation                  |

### ConsensusExec

| Property      | Type            | Required | Default     | Description                             |
| ------------- | --------------- | -------- | ----------- | --------------------------------------- |
| `provider`    | string          | Yes      | -           | AI provider                             |
| `model`       | string          | Yes      | -           | Model identifier                        |
| `temperature` | float (0.0-2.0) | No       | (inherited) | Override temperature for this execution |
| `max_tokens`  | integer (>0)    | No       | (inherited) | Override max_tokens for this execution  |
| `timeout`     | duration        | No       | (inherited) | Override timeout for this execution     |

---

## Variable Interpolation

| Context        | Variable               | Example                | Description                        |
| -------------- | ---------------------- | ---------------------- | ---------------------------------- |
| Input          | `{{input}}`            | `{{input}}`            | User-provided input data           |
| Step output    | `{{step_name}}`        | `{{analyze}}`          | Output from step named "analyze"   |
| Environment    | `{{env.VAR}}`          | `{{env.work_dir}}`     | Environment variable               |
| Loop (iterate) | `{{loop.index}}`       | `{{loop.index}}`       | Current iteration index (0-based)  |
| Loop (iterate) | `{{loop.item}}`        | `{{loop.item}}`        | Current item being processed       |
| Loop (iterate) | `{{loop.total}}`       | `{{loop.total}}`       | Total number of items              |
| Loop (iterate) | `{{loop.succeeded}}`   | `{{loop.succeeded}}`   | Count of successful iterations     |
| Loop (iterate) | `{{loop.failed}}`      | `{{loop.failed}}`      | Count of failed iterations         |
| Loop (refine)  | `{{loop.iteration}}`   | `{{loop.iteration}}`   | Current iteration number (1-based) |
| Loop (refine)  | `{{loop.last.output}}` | `{{loop.last.output}}` | Previous iteration result          |

---

## Common Patterns

### RAG-Augmented Assessment

```yaml
steps:
  - name: retrieve
    rag:
      query: "{{input}}"
      server: pgvector
      strategies: [default]
      top_k: 5

  - name: assess
    needs: [retrieve]
    run: |
      Context: {{retrieve}}
      Assess: {{input}}
```

### Parallel Loop Processing

```yaml
steps:
  - name: process_items
    loop:
      workflow: process_single_item
      mode: iterate
      items: file:///outputs/items.json
      parallel: true
      max_workers: 10
      max_iterations: 1000
      on_failure: continue
      min_success_rate: 0.9
```

### Provider Failover (Workflow-Level)

```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: qwen2.5:32b
```

### Provider Failover (Step-Level)

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: critical_step
    providers:  # Override with failover for this step only
      - provider: anthropic
        model: claude-opus-4  # Try Opus first
      - provider: openai
        model: gpt-4o  # Fallback to OpenAI
    run: "Critical analysis requiring high reliability"

  - name: normal_step
    run: "Regular processing"
    # Uses workflow-level provider (claude-sonnet-4)
```

### Step Dependencies

```yaml
steps:
  - name: step1
    run: "First step"

  - name: step2
    needs: [step1]
    run: "Use {{step1}}"

  - name: step3
    needs: [step1, step2]
    run: "Use {{step1}} and {{step2}}"
```

### Parallel Execution (v2.1.0+)

Enable parallel execution for workflows with independent steps:

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4-20250514
  parallel: true        # Enable parallel execution
  max_workers: 5        # Max concurrent steps
  on_error: cancel_all  # Error handling policy

steps:
  # These three steps have no dependencies - run in parallel
  - name: fetch_config
    run: "Fetch configuration"
  
  - name: fetch_schema
    run: "Fetch schema"
  
  - name: fetch_metadata
    run: "Fetch metadata"
  
  # This step waits for all three to complete
  - name: consolidate
    needs: [fetch_config, fetch_schema, fetch_metadata]
    run: "Consolidate {{fetch_config}}, {{fetch_schema}}, {{fetch_metadata}}"
```

**Important:** When using `parallel: true`, all variable references (like `{{step_name}}`) MUST have the referenced step in the `needs` array. This ensures correct execution order.

**Timeline visualization:**
```
fetch_config   |████|
fetch_schema   |████|
fetch_metadata |████|
consolidate    |    ████|
```

See [PARALLEL_EXECUTION.md](./PARALLEL_EXECUTION.md) for complete documentation.

### Consensus Validation

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Is this safe? Answer YES or NO: {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3
```

### Skills Integration

```yaml
execution:
  servers: [skills]
  skills: [docx, pdf, compliance_assessor]

steps:
  - name: generate_report
    servers: [skills]
    skills: [report_generator]
    run: |
      Use report-generator skill to create a Word document
      from the assessment results.
```

### Execution Order Control

```yaml
steps:
  - name: step_c
    execution_order: 3
    run: "Third"

  - name: step_a
    execution_order: 1
    run: "First"

  - name: step_b
    execution_order: 2
    run: "Second"
# Executes in order: step_a → step_b → step_c (ignoring definition order)
```

### Direct Input Data

```yaml
steps:
  - name: process_config
    input:
      database: postgres
      max_connections: 100
      timeout: 30
    run: |
      Configure database:
      - Type: {{input.database}}
      - Max connections: {{input.max_connections}}
      - Timeout: {{input.timeout}}s
```

---

## CLI Equivalents

### Basic Workflow

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
steps:
  - name: step1
    run: "Prompt here"
```

**Equals:**

```bash
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data "Prompt here"
```

### With MCP Servers & Skills

```yaml
execution:
  provider: anthropic
  servers: [skills]
  skills: [docx, pdf]
steps:
  - name: step1
    run: "Create a document"
```

**Equals:**

```bash
mcp-cli --provider anthropic \
  --server skills \
  --skills docx,pdf \
  --input-data "Create a document"
```

### RAG Search

```yaml
steps:
  - name: search
    rag:
      query: "authentication requirements"
      server: pgvector
      top_k: 5
```

**Equals:**

```bash
mcp-cli rag search "authentication requirements" \
  --server pgvector --top-k 5
```

---

## Task Index

**Common tasks mapped to workflow features:**

| Task                      | Solution                                       | See                                                            |
| ------------------------- | ---------------------------------------------- | -------------------------------------------------------------- |
| Process a list of items   | Loop with `mode: iterate`                      | [LoopConfig](#loopconfig-object-step-level)                    |
| Process items in parallel | Loop with `parallel: true`, `max_workers: N`   | [Parallel Loop Pattern](#parallel-loop-processing)             |
| Retry until condition met | Loop with `mode: refine`, `until: "condition"` | [LoopConfig](#loopconfig-object-step-level)                    |
| Search documentation      | RAG mode with `query`                          | [RagConfig](#ragconfig-object)                                 |
| Get multiple AI opinions  | Consensus mode with multiple `executions`      | [ConsensusConfig](#consensusconfig-object)                     |
| Call another workflow     | Template mode with `name`                      | [Mode 2: template](#mode-2-template---workflow-call)           |
| Generate embeddings       | Embeddings mode with `input`                   | [EmbeddingsConfig](#embeddingsconfig-object)                   |
| Ensure high reliability   | Provider failover with `providers` array       | [Provider Failover Pattern](#provider-failover-workflow-level) |
| Control execution order   | Use `execution_order` property                 | [Step Object](#step-object)                                    |
| Pass structured data      | Use `input` property                           | [Direct Input Pattern](#direct-input-data)                     |
| Wait for other steps      | Use `needs` array                              | [Step Dependencies Pattern](#step-dependencies)                |
| Conditional execution     | Use `if` expression                            | [Step Object](#step-object)                                    |
| Debug specific step       | Override `logging` to `"verbose"`              | [Step Object](#step-object)                                    |
| Share configuration       | Set in `execution`, inherit in steps           | [ExecutionContext](#executioncontext)                          |

---

## File Structure

```yaml
$schema: "workflow/v2.0"        # Always first line
name: workflow_name             # Unique identifier
version: 1.0.0                  # Semantic versioning
description: What it does       # Human-readable

execution:                      # Workflow-level config
  provider: string
  model: string
  servers: [string]
  skills: [string]
  logging: "info"               # or "error", "warn", "step", "debug", "verbose"

env:                            # Optional: Environment vars
  KEY: value

steps:                          # Sequential execution
  - name: step1
    run: "prompt"
  - name: step2
    needs: [step1]
    rag:
      query: "{{step1}}"
  - name: step3
    needs: [step2]
    loop:
      workflow: child_workflow
      mode: iterate
      items: "{{step2}}"
```

---

## See Also

- **[Object Model](OBJECT_MODEL.md)** - Conceptual understanding and detailed guidance
- **[Steps Reference](STEPS_REFERENCE.md)** - Detailed step documentation
- **[CLI Mapping](CLI_MAPPING.md)** - Complete property → CLI argument mapping
- **[RAG Documentation](../../rag/)** - RAG configuration and usage
- **[Loops Guide](../LOOPS.md)** - Deep dive on iterative execution

---

**Last Updated:** 2026-01-15  
**Remember:** Workflows are sequences of mcp-cli calls with shared configuration. Every step property maps to a CLI argument.
