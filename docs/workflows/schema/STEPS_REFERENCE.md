# Steps Reference

**Version:** workflow/v2.0  
**Purpose:** Detailed reference for all step execution modes

---

## Overview

Steps are the building blocks of workflows. Each step is one of six execution modes:

1. **run:** LLM query with variable interpolation
2. **template:** Call another workflow
3. **embeddings:** Generate vector embeddings
4. **consensus:** Multi-provider validation
5. **rag:** Retrieve from vector database (NEW)
6. **loop:** Iterate over items with child workflow (NEW)

All steps inherit properties from `workflow.execution` and can override them.

---

## Step Structure

### Base Step Properties

Every step has these properties:

```yaml
- name: string                  # Required: Unique identifier
  execution_order: number       # Optional: Manual execution order
  needs: [string]               # Optional: Step dependencies
  if: string                    # Optional: Skip condition
  input: any                    # Optional: Direct input data
  
  # Inheritable properties (from workflow.execution)
  provider: string              # Optional: Override provider
  model: string                 # Optional: Override model
  providers: [...]              # Optional: Override provider failover chain
  temperature: number           # Optional: Override temperature
  max_tokens: number            # Optional: Override max_tokens
  servers: [string]             # Optional: Override servers
  skills: [string]              # Optional: Override skills
  timeout: duration             # Optional: Override timeout
  max_iterations: number        # Optional: Override max_iterations
  logging: string               # Optional: Override logging level
  no_color: boolean             # Optional: Override color output
  
  # Execution mode (choose ONE)
  run: string
  template: {...}
  embeddings: {...}
  consensus: {...}
  rag: {...}
  loop: {...}
```

---

## Advanced Step Properties

### Execution Order Control (`execution_order:`)

**Purpose:** Manually control the order in which steps execute, overriding the default dependency-based ordering.

**Syntax:**
```yaml
- name: step_name
  execution_order: number      # Lower numbers execute first
```

### When to Use

- When you need precise control over execution order
- When steps don't have explicit dependencies but order matters
- For debugging or testing specific execution sequences

### Examples

**Basic ordering:**
```yaml
steps:
  - name: cleanup
    execution_order: 99        # Runs last
    run: "Clean up temporary files"
  
  - name: initialize
    execution_order: 1         # Runs first
    run: "Initialize system"
  
  - name: process
    execution_order: 50        # Runs in middle
    run: "Process data"
```

**Mixed with dependencies:**
```yaml
steps:
  - name: step_a
    execution_order: 1
    run: "First"
  
  - name: step_b
    execution_order: 2
    needs: [step_a]            # Both ordering and dependency
    run: "Second (after step_a)"
  
  - name: step_c
    execution_order: 3
    run: "Third"
```

**Note:** When both `execution_order` and `needs` are specified, the system ensures both constraints are satisfied.

---

### Direct Input Data (`input:`)

**Purpose:** Pass structured data directly to a step without relying solely on variable interpolation.

**Syntax:**
```yaml
- name: step_name
  input: any                   # Can be string, number, object, array
  run: string                  # Access via {{input.key}}
```

### When to Use

- When you have complex structured configuration data
- When you want to separate data from prompts
- When you need to pass computed values between workflows

### Examples

**Simple configuration:**
```yaml
steps:
  - name: configure_service
    input:
      database: postgres
      max_connections: 100
      timeout: 30
    run: |
      Configure the service:
      - Database: {{input.database}}
      - Max connections: {{input.max_connections}}
      - Timeout: {{input.timeout}}s
```

**Complex structured data:**
```yaml
steps:
  - name: process_users
    input:
      users:
        - name: Alice
          role: admin
          permissions: [read, write, delete]
        - name: Bob
          role: user
          permissions: [read]
      config:
        strict_mode: true
        audit_enabled: true
    run: |
      Process users with configuration:
      {{input}}
```

**Combining with variable interpolation:**
```yaml
steps:
  - name: analyze
    run: "Extract metadata from: {{input}}"
  
  - name: report
    needs: [analyze]
    input:
      source_data: "{{input}}"
      analysis: "{{analyze}}"
      report_type: detailed
    run: |
      Generate {{input.report_type}} report combining:
      - Source: {{input.source_data}}
      - Analysis: {{input.analysis}}
```

---

### Step-Level Provider Failover (`providers:`)

**Purpose:** Define a provider failover chain specific to a single step, independent of the workflow-level configuration.

**Syntax:**
```yaml
- name: step_name
  providers:
    - provider: string
      model: string
      temperature: number       # Optional per-provider override
      max_tokens: number        # Optional per-provider override
    - provider: string
      model: string
```

### When to Use

- When a specific step requires higher reliability than others
- When a step needs a specific provider that differs from the workflow default
- When you want to try expensive models first and fall back to cheaper ones for specific tasks

### Examples

**Critical step with failover:**
```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4       # Default for most steps

steps:
  - name: critical_analysis
    providers:
      - provider: anthropic
        model: claude-opus-4   # Try best model first
      - provider: openai
        model: gpt-4o          # Fall back to OpenAI
      - provider: anthropic
        model: claude-sonnet-4 # Final fallback
    run: "Perform critical analysis on: {{input}}"
  
  - name: simple_task
    run: "Simple processing"   # Uses default claude-sonnet-4
```

**Different providers for different tasks:**
```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: code_analysis
    providers:
      - provider: deepseek
        model: deepseek-coder  # Specialized for code
      - provider: anthropic
        model: claude-sonnet-4
    run: "Analyze this code: {{input}}"
  
  - name: creative_writing
    providers:
      - provider: anthropic
        model: claude-opus-4   # Best for creativity
      - provider: openai
        model: gpt-4o
    run: "Write a creative story about: {{input}}"
```

**Cost optimization with fallback:**
```yaml
steps:
  - name: expensive_task
    providers:
      - provider: ollama
        model: qwen2.5:32b     # Try local first (free)
        timeout: "30s"
      - provider: anthropic
        model: claude-sonnet-4 # Fall back to cloud if local fails/slow
    run: "Process: {{input}}"
```

---

### Step-Level Logging (`logging:`)

**Purpose:** Override the logging level for a specific step without affecting other steps.

**Syntax:**
```yaml
- name: step_name
  logging: "normal" | "verbose" | "noisy"
```

### Examples

**Debug specific step:**
```yaml
execution:
  logging: normal              # Quiet by default

steps:
  - name: simple_task
    run: "Simple processing"   # Uses normal logging
  
  - name: debug_step
    logging: verbose           # Detailed logging for this step only
    run: "Complex analysis requiring debugging"
  
  - name: very_verbose_step
    logging: noisy             # Maximum detail for troubleshooting
    run: "Problematic operation"
```

---

### Step-Level Color Control (`no_color:`)

**Purpose:** Override color output for a specific step (rarely needed).

**Syntax:**
```yaml
- name: step_name
  no_color: true | false
```

---

## Mode 1: LLM Query (`run:`)

**Purpose:** Execute a single LLM query with variable interpolation

**Syntax:**
```yaml
- name: step_name
  run: string                   # Prompt with {{variables}}
```

### Variable Interpolation

Available variables in `run:` prompts:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{input}}` | User input | `{{input}}` |
| `{{step_name}}` | Output from another step | `{{analyze}}` |
| `{{env.VAR}}` | Environment variable | `{{env.API_KEY}}` |
| `{{workflow.name}}` | Workflow identifier | `{{workflow.name}}` |

### Examples

**Basic query:**
```yaml
steps:
  - name: analyze
    run: "Analyze this code: {{input}}"
```

**Using previous step output:**
```yaml
steps:
  - name: analyze
    run: "Analyze: {{input}}"
  
  - name: improve
    needs: [analyze]
    run: "Based on this analysis: {{analyze}}, suggest improvements"
```

**Multi-line prompt:**
```yaml
steps:
  - name: review
    run: |
      Review the following code for:
      1. Security vulnerabilities
      2. Performance issues
      3. Code style
      
      Code:
      {{input}}
      
      Provide detailed feedback.
```

---

## Mode 2: Workflow Call (`template:`)

**Purpose:** Call another workflow as a subroutine

**Syntax:**
```yaml
- name: step_name
  template:
    name: string               # Workflow name
    with: {key: value}         # Input data (optional)
```

### Examples

**Basic template call:**
```yaml
steps:
  - name: review_code
    template:
      name: code_reviewer
      with:
        code: "{{input}}"
```

**Chaining workflows:**
```yaml
steps:
  - name: stage1
    template:
      name: data_processor
      with:
        data: "{{input}}"
  
  - name: stage2
    needs: [stage1]
    template:
      name: results_formatter
      with:
        processed_data: "{{stage1}}"
```

---

## Mode 3: Embeddings (`embeddings:`)

**Purpose:** Generate vector embeddings from text

**Syntax:**
```yaml
- name: step_name
  embeddings:
    # Input (one required)
    input: string | [string]   # Inline text(s)
    input_file: string         # File path
    
    # Provider override
    provider: string
    model: string
    
    # Chunking
    chunk_strategy: string     # sentence, paragraph, fixed
    max_chunk_size: number     # tokens
    overlap: number            # tokens
    
    # Model config
    dimensions: number
    
    # Output
    encoding_format: string    # float, base64
    include_metadata: boolean
    output_format: string      # json, csv, compact
    output_file: string
```

### Examples

**Basic embeddings:**
```yaml
steps:
  - name: embed_docs
    embeddings:
      model: text-embedding-3-small
      input: "{{input}}"
```

**With chunking configuration:**
```yaml
steps:
  - name: embed_large_doc
    embeddings:
      model: text-embedding-3-small
      input: "{{input}}"
      chunk_strategy: sentence
      max_chunk_size: 512
      overlap: 50
```

---

## Mode 4: Consensus (`consensus:`)

**Purpose:** Execute across multiple providers and require agreement

**Syntax:**
```yaml
- name: step_name
  consensus:
    prompt: string             # Sent to all providers
    executions: [...]          # Provider configurations
    require: string            # unanimous, majority, 2/3
    timeout: duration          # optional
```

### Example

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

---

## Mode 5: RAG Retrieval (`rag:`)

**Purpose:** Retrieve relevant documents from a vector database using semantic search

**Syntax:**
```yaml
- name: step_name
  rag:
    query: string              # Search query (supports {{variables}})
    server: string             # RAG server name (from config/rag/*.yaml)
    strategies: [string]       # Vector search strategies
    top_k: number              # Number of results (default: 5)
    fusion: string             # Fusion method: rrf, weighted, max, avg
    expand_query: boolean      # Enable query expansion
    output_format: string      # json, text, compact
```

### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `query` | string | Yes | - | Search query (supports `{{variables}}`) |
| `server` | string | No | (from config) | RAG server name |
| `strategies` | string[] | No | `[default]` | Vector search strategies |
| `top_k` | int | No | 5 | Number of results |
| `fusion` | string | No | `rrf` | Result fusion: `rrf`, `weighted`, `max`, `avg` |
| `expand_query` | bool | No | false | Enable query expansion |
| `output_format` | string | No | `json` | Output format: `json`, `text`, `compact` |

### Examples

**Basic RAG retrieval:**
```yaml
steps:
  - name: retrieve_controls
    rag:
      query: "authentication requirements"
      server: pgvector
      top_k: 5
```

**RAG with previous step output:**
```yaml
steps:
  - name: parse_statement
    run: "Extract the key topic from: {{input}}"
  
  - name: retrieve
    needs: [parse_statement]
    rag:
      query: "{{parse_statement}}"
      server: pgvector
      strategies: [default]
      top_k: 5
      output_format: json
  
  - name: assess
    needs: [retrieve]
    run: |
      Based on these controls: {{retrieve}}
      
      Assess compliance of: {{input}}
```

**RAG with query expansion:**
```yaml
steps:
  - name: search
    rag:
      query: "MFA requirements"
      server: pgvector
      expand_query: true       # Expands "MFA" to include "multi-factor authentication"
      top_k: 10
```

### RAG Output Format

The RAG step returns structured results that can be used in subsequent steps:

```json
{
  "query": "authentication requirements",
  "results": [
    {
      "id": "ISM-1546",
      "text": {
        "identifier": "ISM-1546",
        "description": "Users are authenticated before being granted access..."
      },
      "combined_score": 0.85,
      "source": "description_vector"
    }
  ],
  "total_results": 5
}
```

---

## Mode 6: Loop Execution (`loop:`)

**Purpose:** Iterate over a collection of items, executing a child workflow for each

**Syntax:**
```yaml
- name: step_name
  loop:
    # Core
    workflow: string           # Child workflow to execute
    mode: string               # iterate | refine
    items: string              # Item source (iterate mode)
    with: {key: value}         # Input parameters
    
    # Iteration control
    max_iterations: number     # Safety limit (required)
    until: string              # Exit condition (refine mode)
    
    # Error handling
    on_failure: string         # halt | continue | retry
    max_retries: number        # Retries per item
    retry_delay: string        # Backoff duration
    
    # Success criteria
    min_success_rate: number   # 0.0 to 1.0
    
    # Timeouts
    timeout_per_item: string   # Per-iteration timeout
    total_timeout: string      # Total loop timeout
    
    # Parallel execution
    parallel: boolean          # Enable parallel processing
    max_workers: number        # Concurrent workers (default: 3)
    
    # Output
    accumulate: string         # Store all results in variable
```

### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| **Core** | | | | |
| `workflow` | string | Yes | - | Child workflow to execute |
| `mode` | string | No | `refine` | `iterate` (over items) or `refine` (until condition) |
| `items` | string | No* | - | Item source for iterate mode |
| `with` | object | No | {} | Input parameters for child workflow |
| **Iteration Control** | | | | |
| `max_iterations` | int | Yes | - | Maximum iterations (safety limit) |
| `until` | string | No* | - | Exit condition for refine mode |
| **Error Handling** | | | | |
| `on_failure` | string | No | `halt` | `halt`, `continue`, or `retry` |
| `max_retries` | int | No | 0 | Retries per item (when `on_failure: retry`) |
| `retry_delay` | string | No | - | Delay between retries (e.g., `"5s"`) |
| **Success Criteria** | | | | |
| `min_success_rate` | float | No | 0.0 | Minimum success rate (0.0-1.0) |
| **Timeouts** | | | | |
| `timeout_per_item` | string | No | - | Per-iteration timeout (e.g., `"30s"`) |
| `total_timeout` | string | No | - | Total loop timeout (e.g., `"1h"`) |
| **Parallel Execution** | | | | |
| `parallel` | bool | No | false | Enable parallel processing |
| `max_workers` | int | No | 3 | Maximum concurrent workers |
| **Output** | | | | |
| `accumulate` | string | No | - | Variable to store all iteration results |

\* `items` required for `mode: iterate`, `until` required for `mode: refine`

### Loop Modes

**Iterate Mode:** Process each item in a collection
```yaml
steps:
  - name: process_all
    loop:
      workflow: process_single_item
      mode: iterate
      items: file:///outputs/items.json
      max_iterations: 500
      parallel: true
      max_workers: 10
```

**Refine Mode:** Repeat until a condition is met
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

### Loop Variables

| Variable | Mode | Description |
|----------|------|-------------|
| `{{loop.index}}` | iterate | Current item index (0-based) |
| `{{loop.item}}` | iterate | Current item being processed |
| `{{loop.total}}` | iterate | Total number of items |
| `{{loop.succeeded}}` | iterate | Count of successful iterations |
| `{{loop.failed}}` | iterate | Count of failed iterations |
| `{{loop.iteration}}` | refine | Current iteration number (1-based) |
| `{{loop.last.output}}` | refine | Previous iteration's output |

### Examples

**Parallel Iterate with Error Handling:**
```yaml
steps:
  - name: assess_statements
    loop:
      workflow: ism_assess_statement_v2
      mode: iterate
      items: file:///outputs/statements_for_assessment.json
      parallel: true
      max_workers: 15
      max_iterations: 500
      on_failure: continue
      min_success_rate: 0.8
```

**Sequential Refine with Retry:**
```yaml
steps:
  - name: iterative_improvement
    loop:
      workflow: improve_code
      mode: refine
      with:
        code: "{{input}}"
        previous_feedback: "{{loop.last.output}}"
      max_iterations: 5
      until: "Code review passes"
      on_failure: retry
      max_retries: 2
      retry_delay: "5s"
```

**Iterate with Accumulation:**
```yaml
steps:
  - name: batch_process
    loop:
      workflow: process_item
      mode: iterate
      items: "{{step.extract_items}}"
      max_iterations: 100
      on_failure: continue
      accumulate: all_results
  
  - name: summarize
    needs: [batch_process]
    run: |
      Summarize these results:
      {{all_results}}
```

### Item Sources for Iterate Mode

The `items` property supports multiple formats:

**File reference:**
```yaml
items: file:///outputs/items.json      # JSON array file
items: file:///outputs/items.jsonl     # JSONL file
```

**Step output:**
```yaml
items: "{{extract_step}}"              # Output from previous step (must be JSON array)
```

**Inline (via with):**
```yaml
loop:
  workflow: process_item
  mode: iterate
  items: "{{env.items_json}}"          # From environment variable
```

### Parallel Execution Considerations

When using `parallel: true`:

1. **Race conditions:** Each worker writes to separate files (use individual file names)
2. **Resource limits:** Monitor API rate limits with `max_workers`
3. **Order independence:** Results may complete out of order
4. **Error isolation:** One failure doesn't affect other workers (with `on_failure: continue`)

**Best practice for parallel file output:**
```yaml
# In child workflow, write to individual files:
- name: write_result
  run: |
    Write result to /outputs/results/{{loop.item.id}}.json
```

---

## Step Dependencies (`needs:`)

### Basic Dependencies

```yaml
steps:
  - name: step1
    run: "First"
  
  - name: step2
    needs: [step1]             # Waits for step1
    run: "Use {{step1}}"
```

### Execution Order

```yaml
steps:
  - name: A
  - name: B
    needs: [A]
  - name: C
    needs: [A]
  - name: D
    needs: [B, C]
```

**Execution flow:**
```
    A (runs first)
   / \
  B   C (run in parallel after A)
   \ /
    D (runs after both B and C)
```

---

## Conditional Execution (`if:`)

```yaml
steps:
  - name: check
    run: "Is this valid?"
  
  - name: process
    needs: [check]
    if: "${{ check == 'YES' }}"
    run: "Process the data"
```

---

## Common Patterns

### Pattern 1: RAG-Augmented Assessment

```yaml
steps:
  - name: parse
    run: "Extract key topic: {{input}}"
  
  - name: retrieve
    needs: [parse]
    rag:
      query: "{{parse}}"
      server: pgvector
      top_k: 5
  
  - name: assess
    needs: [retrieve]
    run: |
      Controls: {{retrieve}}
      Assess: {{input}}
```

### Pattern 2: Batch Processing with Parallel Loop

```yaml
steps:
  - name: extract_items
    run: "Parse input into JSON array: {{input}}"
  
  - name: process_batch
    needs: [extract_items]
    loop:
      workflow: process_single
      mode: iterate
      items: "{{extract_items}}"
      parallel: true
      max_workers: 10
      max_iterations: 1000
      on_failure: continue
      min_success_rate: 0.9
  
  - name: summarize
    needs: [process_batch]
    run: "Summarize processing results"
```

### Pattern 3: Iterative Refinement

```yaml
steps:
  - name: initial_draft
    run: "Create initial draft: {{input}}"
  
  - name: refine_loop
    needs: [initial_draft]
    loop:
      workflow: review_and_improve
      mode: refine
      with:
        content: "{{initial_draft}}"
        feedback: "{{loop.last.output}}"
      max_iterations: 5
      until: "Review score is 9 or higher"
```

### Pattern 4: Multi-Stage Pipeline with RAG

```yaml
steps:
  - name: classify
    run: "Classify this document: {{input}}"
  
  - name: retrieve_templates
    needs: [classify]
    rag:
      query: "{{classify}} document templates"
      server: pgvector
      top_k: 3
  
  - name: generate
    needs: [retrieve_templates]
    run: |
      Using templates: {{retrieve_templates}}
      Generate document for: {{input}}
  
  - name: validate
    needs: [generate]
    consensus:
      prompt: "Is this document complete? {{generate}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

---

## Best Practices

### 1. Name Steps Descriptively
```yaml
# ✅ Good
steps:
  - name: retrieve_ism_controls
  - name: assess_compliance
  - name: generate_report

# ❌ Bad
steps:
  - name: step1
  - name: step2
```

### 2. Use RAG for Context
```yaml
# ✅ Good - RAG provides relevant context
steps:
  - name: retrieve
    rag:
      query: "{{input}}"
      top_k: 5
  
  - name: answer
    needs: [retrieve]
    run: "Based on: {{retrieve}}, answer: {{input}}"

# ❌ Bad - No context, relies only on model knowledge
steps:
  - name: answer
    run: "Answer: {{input}}"
```

### 3. Handle Loop Failures Gracefully
```yaml
# ✅ Good - Continues on failure, checks success rate
loop:
  on_failure: continue
  min_success_rate: 0.8

# ❌ Bad - Halts entire loop on first failure
loop:
  on_failure: halt  # One bad item stops everything
```

### 4. Set Appropriate Parallelism
```yaml
# ✅ Good - Reasonable worker count
loop:
  parallel: true
  max_workers: 10   # Respects API rate limits

# ❌ Bad - Too many workers
loop:
  parallel: true
  max_workers: 100  # May hit rate limits
```

---

## Troubleshooting

### RAG Returns No Results
- Check RAG server configuration in `config/rag/*.yaml`
- Verify database has embeddings
- Try lowering similarity threshold
- Enable `expand_query: true`

### Loop Not Processing All Items
- Check `max_iterations` is high enough
- Verify items file format (JSON array or JSONL)
- Check `min_success_rate` if loop exits early

### Parallel Loop Race Conditions
- Write to individual files per item, not shared file
- Merge results in a separate step after loop completes

---

## See Also

- **[Quick Reference](QUICK_REFERENCE.md)** - One-page overview
- **[Object Model](OBJECT_MODEL.md)** - TypeScript interfaces
- **[RAG Documentation](../../rag/)** - RAG configuration
- **[Loops Guide](../LOOPS.md)** - Deep dive on loops

---

**Last Updated:** 2026-01-15
