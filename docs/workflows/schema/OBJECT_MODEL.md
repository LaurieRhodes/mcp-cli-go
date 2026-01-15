# Workflow Object Model

**Version:** workflow/v2.0  
**Purpose:** Understanding how workflow objects work together

---

## What You'll Learn

This guide explains:

- 🧩 **The Mental Model:** How to think about workflows
- 🏗️ **Object Relationships:** How objects inherit and relate to each other
- 🎯 **When to Use What:** Decision guidance for each property
- ⚠️ **Common Mistakes:** What to avoid and why
- 📐 **Design Patterns:** Proven approaches that work well

---

## The Mental Model

### Think of Workflows as Scripts

A workflow is **a script that sequences multiple mcp-cli calls** with shared configuration. That's it.

```
Workflow = Configuration + Steps + Orchestration
```

**Configuration:** Provider, model, servers, skills (defined once, inherited everywhere)  
**Steps:** Individual mcp-cli calls that do actual work  
**Orchestration:** Dependencies, conditions, loops that control execution flow

### The Three-Layer Hierarchy

```
┌─────────────────────────────────────────┐
│ 1. WORKFLOW LEVEL (execution:)          │
│    - Default configuration for all steps│
│    - "Set it and forget it"             │
└─────────────┬───────────────────────────┘
              │ Inherits down ↓
┌─────────────────────────────────────────┐
│ 2. STEP LEVEL (steps[]:)                │
│    - Can override ANY workflow property │
│    - Does the actual work               │
└─────────────┬───────────────────────────┘
              │ Inherits down ↓ (consensus only)
┌─────────────────────────────────────────┐
│ 3. EXECUTION LEVEL (consensus only)     │
│    - Per-provider overrides             │
│    - For multi-provider validation      │
└─────────────────────────────────────────┘
```

**Key Insight:** Properties flow downward. Define once at the top, override only where needed.

---

## Core Objects Explained

### WorkflowV2: The Container

**What it is:** The top-level object that contains everything.

**Think of it as:** A package that bundles configuration, steps, and metadata.

**What you define here:**

- **Metadata:** What is this workflow? (`name`, `version`, `description`)
- **Defaults:** What provider/model/settings should all steps use? (`execution`)
- **Work:** What steps should execute? (`steps`)
- **Environment:** What variables are available? (`env`)

**When you need it:** Every workflow starts with this.

```yaml
$schema: "workflow/v2.0"      # Always this for v2 workflows
name: my_workflow             # Unique identifier
version: 1.0.0                # Semantic versioning
description: What it does     # Human-readable purpose

execution:                    # Default configuration
  provider: anthropic
  model: claude-sonnet-4

steps:                        # The actual work
  - name: step1
    run: "Do something"
```

**Common Mistake:** Forgetting that `execution` provides **defaults**, not requirements. Steps can override everything.

---

### ExecutionContext: The Defaults

**What it is:** The default configuration inherited by all steps.

**Think of it as:** The "template" for how steps should execute unless they say otherwise.

**What you define here:**

| Property      | Purpose          | When to Set                            |
| ------------- | ---------------- | -------------------------------------- |
| `provider`    | Which AI service | Always (unless using `providers`)      |
| `model`       | Which model      | Always (unless using `providers`)      |
| `providers`   | Failover chain   | When you need reliability              |
| `temperature` | Creativity level | When you want consistency across steps |
| `servers`     | MCP servers      | When most steps need same servers      |
| `skills`      | Anthropic skills | When most steps use same skills        |
| `timeout`     | How long to wait | When default 60s isn't right           |
| `logging`     | Verbosity        | For debugging entire workflow          |

**Decision: Single Provider vs Failover Chain?**

```yaml
# Single provider (simpler)
execution:
  provider: anthropic
  model: claude-sonnet-4

# Failover chain (more reliable)
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: qwen2.5:32b
```

**When to use single:** Most workflows. Simpler, easier to debug.  
**When to use failover:** Production workflows where uptime matters more than simplicity.

**Common Mistake:** Setting `servers` or `skills` at workflow level when only one step needs them. This makes all steps load unnecessary tools.

---

### Step: The Worker

**What it is:** A single unit of work - one mcp-cli call.

**Think of it as:** A function call with configuration.

**Anatomy of a Step:**

```yaml
- name: unique_identifier        # Required: How to reference this step

  # ORCHESTRATION (optional)
  needs: [other_steps]           # Wait for these steps to complete
  if: "{{condition}}"            # Skip if condition false
  execution_order: 10            # Manual ordering (rarely needed)

  # CONFIGURATION OVERRIDES (optional - inherits from execution if not set)
  provider: anthropic            # Override just for this step
  temperature: 0.9               # Override just for this step

  # WORK (exactly ONE required)
  run: "Prompt text"             # LLM query
  template: {...}                # Call another workflow
  embeddings: {...}              # Generate vectors
  consensus: {...}               # Multi-provider validation
  rag: {...}                     # Search vector database
  loop: {...}                    # Iterate over items
```

**Decision Tree: Which Execution Mode?**

```
Need to query an LLM?
├─ Yes → Use `run:`
└─ No
   ├─ Need to call another workflow? → Use `template:`
   ├─ Need to generate embeddings? → Use `embeddings:`
   ├─ Need to search vector DB? → Use `rag:`
   ├─ Need multi-provider agreement? → Use `consensus:`
   └─ Need to iterate over items? → Use `loop:`
```

**When to Override Configuration:**

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.3              # Conservative default

steps:
  - name: creative_writing
    temperature: 1.5            # Override for creativity
    run: "Write a story"

  - name: data_analysis
    # Uses default 0.3 - good for precision
    run: "Analyze data"
```

**Rule of Thumb:** Only override when a step needs different behavior. Most steps should inherit.

**Common Mistakes:**

1. **Overriding unnecessarily:** Setting `provider` on every step when workflow default works fine
2. **Forgetting dependencies:** Step uses `{{other_step}}` but forgot `needs: [other_step]`
3. **Using wrong mode:** Using `run:` for embeddings instead of `embeddings:` mode

---

### Step Properties Deep Dive

#### execution_order: Manual Ordering

**What it does:** Forces steps to execute in numerical order.

**When to use:**

- Steps don't have dependencies but order matters
- Debugging execution flow
- Ensuring cleanup happens last

**When NOT to use:**

- Steps have clear dependencies (use `needs:` instead)
- Parallel execution is desired (order prevents parallelization)

```yaml
steps:
  - name: initialize
    execution_order: 1
    run: "Set up environment"

  - name: work
    execution_order: 50
    run: "Do main work"

  - name: cleanup
    execution_order: 99
    run: "Clean up"
```

**Trade-off:** Order guarantees sequence but prevents parallel execution. Use `needs:` when possible for better performance.

---

#### input: Direct Data

**What it does:** Passes structured data to a step directly.

**When to use:**

- Step needs complex configuration
- You want to separate data from prompt
- Passing computed values between workflows

**When NOT to use:**

- Simple string input (just use `{{input}}`)
- Data already in another step (use `{{step_name}}`)

```yaml
steps:
  - name: configure
    input:
      database:
        host: localhost
        port: 5432
        name: mydb
      features:
        - authentication
        - logging
        - caching
    run: |
      Configure application:
      - Database: {{input.database.host}}:{{input.database.port}}
      - Features: {{input.features}}
```

**Pattern:** Use for configuration that's easier to write as YAML than in a prompt.

---

#### providers: Step-Level Failover

**What it does:** Defines provider failover **for this step only**, overriding workflow-level config.

**When to use:**

- One step is critical and needs higher reliability
- One step benefits from a specific model
- Cost optimization (try local first, fall back to cloud)

**When NOT to use:**

- All steps need same failover (set at workflow level instead)
- Debugging (adds complexity)

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4        # Default for most steps

steps:
  - name: critical_validation
    providers:                   # This step gets special treatment
      - provider: anthropic
        model: claude-opus-4     # Best model first
      - provider: openai
        model: gpt-4o            # Expensive fallback
    run: "Critical validation task"

  - name: simple_task
    run: "Simple task"           # Uses default sonnet-4
```

**Cost Impact:** First provider in chain is tried first. Put cheapest that works first for cost optimization.

---

### Execution Modes Explained

Each mode represents a different type of work:

#### Mode 1: run - LLM Query

**Purpose:** Send a prompt to an LLM and get a response.

**This is what you use:** 90% of the time.

```yaml
- name: analyze
  run: "Analyze this code: {{input}}"
```

**Variable Interpolation Available:**

- `{{input}}` - User's input
- `{{step_name}}` - Output from another step
- `{{env.VAR}}` - Environment variable
- `{{loop.item}}` - Current loop item (in loop context)

**Pattern:** Use multi-line strings for complex prompts:

```yaml
- name: review
  run: |
    Review the following code for:
    1. Security vulnerabilities
    2. Performance issues
    3. Code style violations

    Code:
    {{input}}

    Provide detailed feedback with specific line references.
```

---

#### Mode 2: template - Workflow Call

**Purpose:** Call another workflow as a subroutine.

**When to use:**

- Reusing logic across workflows
- Breaking complex workflows into manageable pieces
- Creating workflow libraries

```yaml
- name: process_item
  template:
    name: item_processor
    with:
      item_data: "{{input}}"
      strict_mode: true
```

**Pattern:** Child workflows should be focused and reusable. One clear responsibility.

---

#### Mode 3: embeddings - Vector Generation

**Purpose:** Convert text into vector embeddings for semantic search.

**When to use:**

- Building a RAG database
- Semantic similarity comparison
- Clustering or classification

**Not for:** Regular LLM queries (use `run:`)

```yaml
- name: embed_docs
  embeddings:
    model: text-embedding-3-small
    input: "{{input}}"
    chunk_strategy: sentence
    max_chunk_size: 512
```

**Key Decision: chunk_strategy**

- `sentence`: Best for Q&A systems
- `paragraph`: Best for document similarity
- `fixed`: Best for consistent-size chunks

---

#### Mode 4: consensus - Multi-Provider Validation

**Purpose:** Get multiple AI providers to agree on an answer.

**When to use:**

- Safety-critical decisions
- Validating outputs
- Reducing hallucinations

**Cost:** Runs multiple providers in parallel. 3x the cost for 3 providers.

```yaml
- name: safety_check
  consensus:
    prompt: "Is this safe? Answer YES or NO: {{input}}"
    executions:
      - provider: anthropic
        model: claude-sonnet-4
      - provider: openai
        model: gpt-4o
      - provider: deepseek
        model: deepseek-chat
    require: unanimous          # All must agree
```

**Decision: Agreement Threshold**

- `unanimous`: Highest safety, may never agree
- `2/3`: Balanced (recommended)
- `majority`: Lowest barrier

---

#### Mode 5: rag - Vector Search

**Purpose:** Retrieve relevant documents from a vector database.

**When to use:**

- You have a RAG database set up
- Need context from large document sets
- Building Q&A systems

**Pattern: RAG + LLM**

```yaml
steps:
  - name: retrieve_context
    rag:
      query: "{{input}}"
      server: pgvector
      top_k: 5

  - name: answer_with_context
    needs: [retrieve_context]
    run: |
      Context: {{retrieve_context}}

      Question: {{input}}

      Answer based on the context provided.
```

---

#### Mode 6: loop - Iteration

**Purpose:** Execute a child workflow multiple times.

**Two modes:**

1. **iterate:** Process each item in a list
2. **refine:** Repeat until condition met

**When to use iterate:**

```yaml
- name: process_all_items
  loop:
    workflow: process_single_item
    mode: iterate
    items: file:///outputs/items.json
    parallel: true
    max_workers: 10
    max_iterations: 1000
```

**When to use refine:**

```yaml
- name: fix_until_perfect
  loop:
    workflow: test_and_fix
    mode: refine
    with:
      code: "{{input}}"
    max_iterations: 10
    until: "All tests pass"
```

**Key Decisions:**

| Question                 | Answer            |
| ------------------------ | ----------------- |
| Process a list of items? | `mode: iterate`   |
| Repeat until condition?  | `mode: refine`    |
| Items independent?       | `parallel: true`  |
| Items sequential?        | `parallel: false` |

---

## Property Inheritance: How It Really Works

### The Inheritance Chain

```
workflow.execution
  ↓ [Step inherits ALL properties]
step.property
  ↓ [Consensus execution inherits from step]
consensus.execution.property
```

### What Gets Inherited

**Everything from MCPQuery:**

- `provider` / `model` / `providers`
- `temperature`
- `max_tokens`
- `servers`
- `skills`
- `timeout`
- `max_iterations`

### Practical Example: Temperature Inheritance

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.3           # Conservative default

steps:
  - name: step1
    run: "Analyze"           # Uses 0.3 ← inherited

  - name: step2
    temperature: 1.5         # Overrides to 1.5
    run: "Be creative"

  - name: step3
    consensus:
      prompt: "Validate"
      executions:
        - provider: anthropic
          # Uses 0.3 ← inherited from workflow
        - provider: openai
          temperature: 0.1   # Overrides to 0.1
      require: unanimous
```

**Execution:**

- step1: temperature = 0.3 (from workflow)
- step2: temperature = 1.5 (step override)
- step3, execution 1: temperature = 0.3 (from workflow)
- step3, execution 2: temperature = 0.1 (execution override)

---

## Common Patterns

### Pattern 1: Simple Sequential Processing

**Use case:** Steps depend on each other, execute in order.

```yaml
steps:
  - name: extract
    run: "Extract data from: {{input}}"

  - name: transform
    needs: [extract]
    run: "Transform: {{extract}}"

  - name: load
    needs: [transform]
    run: "Load: {{transform}}"
```

**Why it works:** Clear dependencies, easy to debug.

---

### Pattern 2: RAG-Augmented Generation

**Use case:** Retrieve context, then generate response.

```yaml
steps:
  - name: retrieve
    rag:
      query: "{{input}}"
      server: pgvector
      top_k: 5

  - name: generate
    needs: [retrieve]
    run: |
      Context: {{retrieve}}
      Question: {{input}}
      Generate answer based on context.
```

**Why it works:** Separates retrieval from generation for clarity.

---

### Pattern 3: Parallel Processing with Loops

**Use case:** Process many items independently.

```yaml
steps:
  - name: extract_items
    run: "Extract items as JSON array: {{input}}"

  - name: process_all
    needs: [extract_items]
    loop:
      workflow: process_single
      mode: iterate
      items: "{{extract_items}}"
      parallel: true
      max_workers: 15
      max_iterations: 1000
      on_failure: continue
```

**Why it works:** Fast, fault-tolerant, scalable.

---

### Pattern 4: Cost-Optimized Failover

**Use case:** Try cheap options first, fall back to expensive.

```yaml
execution:
  providers:
    - provider: ollama
      model: qwen2.5:32b
      timeout: "30s"         # Local, free, but might be slow
    - provider: anthropic
      model: claude-sonnet-4 # Cloud, paid, reliable
```

**Why it works:** Minimizes cost while maintaining reliability.

---

## Anti-Patterns (What NOT to Do)

### ❌ Anti-Pattern 1: Repeating Configuration

```yaml
# BAD
steps:
  - name: step1
    provider: anthropic
    model: claude-sonnet-4
    temperature: 0.7
    run: "Task 1"

  - name: step2
    provider: anthropic
    model: claude-sonnet-4
    temperature: 0.7
    run: "Task 2"
```

**Why it's bad:** Repetitive, error-prone, hard to change.

**Fix:** Use workflow-level defaults:

```yaml
# GOOD
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "Task 1"
  - name: step2
    run: "Task 2"
```

---

### ❌ Anti-Pattern 2: Unnecessary Execution Order

```yaml
# BAD
steps:
  - name: step1
    execution_order: 1
    run: "First"

  - name: step2
    execution_order: 2
    needs: [step1]           # Already has dependency!
    run: "Second"
```

**Why it's bad:** `needs:` already ensures order. `execution_order` is redundant and prevents optimization.

**Fix:** Use `needs:` only:

```yaml
# GOOD
steps:
  - name: step1
    run: "First"

  - name: step2
    needs: [step1]
    run: "Second"
```

---

### ❌ Anti-Pattern 3: Wrong Tool for Job

```yaml
# BAD - Using run: for embeddings
steps:
  - name: embed
    run: "Generate embeddings for: {{input}}"
```

**Why it's bad:** LLMs aren't embedding models. This will return text, not vectors.

**Fix:** Use embeddings mode:

```yaml
# GOOD
steps:
  - name: embed
    embeddings:
      model: text-embedding-3-small
      input: "{{input}}"
```

---

## Validation Rules Explained

### Required vs Optional

| Level     | Required                                                          | Optional       |
| --------- | ----------------------------------------------------------------- | -------------- |
| Workflow  | `$schema`, `name`, `version`, `description`, `execution`, `steps` | `env`, `loops` |
| Execution | `provider` + `model` OR `providers`                               | All others     |
| Step      | `name`, ONE execution mode                                        | All others     |

### Mutual Exclusivity

**At execution level:**

```yaml
# Can't have both
execution:
  provider: anthropic    # Option 1: Single provider
  providers: [...]       # Option 2: Failover chain
  # ❌ INVALID: Can't specify both
```

**At step level:**

```yaml
steps:
  - name: step1
    run: "Query"         # Option 1
    embeddings: {...}    # Option 2
    # ❌ INVALID: Can only have ONE execution mode
```

### Dependency Validation

```yaml
steps:
  - name: step1
    run: "First"

  - name: step2
    needs: [step3]       # ❌ INVALID: step3 doesn't exist
    run: "Second"

  - name: step3
    needs: [step2]       # ❌ INVALID: Circular dependency
    run: "Third"
```

**Rules:**

1. Steps in `needs:` must exist
2. No circular dependencies
3. Can only reference steps defined earlier (or use `execution_order`)

---

## Troubleshooting Guide

### "Step skipped unexpectedly"

**Cause:** Condition in `if:` evaluated to false.

**Debug:**

```yaml
- name: my_step
  if: "{{some_condition}}"
  logging: verbose         # Add this to see why
  run: "Task"
```

---

### "Property not being inherited"

**Check:**

1. Is property overridden at step level?
2. For pointers (temperature, max_tokens), is it set to nil?
3. Is spelling exact? (`max_tokens` not `maxTokens`)

---

### "Providers not failing over"

**Check:**

1. Did first provider succeed? (No failover on success)
2. Is timeout long enough? (Provider might be slow, not failing)
3. Are provider credentials configured?

---

## Decision Trees

### "Which execution mode should I use?"

```
Is this an LLM query?
├─ Yes → run:
└─ No
   ├─ Calling another workflow? → template:
   ├─ Generating embeddings? → embeddings:
   ├─ Searching vector DB? → rag:
   ├─ Need provider agreement? → consensus:
   └─ Processing multiple items? → loop:
```

### "Should I set this at workflow or step level?"

```
Will ALL steps use this value?
├─ Yes → Set at workflow level (execution:)
└─ No
   ├─ Will MOST steps use it? → Set at workflow, override in exceptions
   └─ Only one step needs it? → Set at step level only
```

### "Single provider or failover chain?"

```
Is this workflow production-critical?
├─ Yes
│  └─ Can you afford multiple providers? → Use failover chain
└─ No → Single provider (simpler)
```

---

## Complete Annotated Example

```yaml
# Schema version - always use workflow/v2.0 for current features
$schema: "workflow/v2.0"

# Unique identifier - used in logs, error messages, workflow calls
name: "document_processor"

# Semantic versioning - increment when behavior changes
version: "2.0.0"

# Human-readable description - shows in workflow listings
description: "Process uploaded documents through OCR, analysis, and storage"

# ═══════════════════════════════════════════════════════════
# WORKFLOW-LEVEL CONFIGURATION (applies to all steps by default)
# ═══════════════════════════════════════════════════════════
execution:
  # Primary provider - most steps will use this
  provider: anthropic
  model: claude-sonnet-4

  # Conservative temperature - we want consistent results
  temperature: 0.3

  # MCP servers available to all steps
  servers: [skills]

  # Skills most steps might need
  skills: [docx, pdf]

  # Generous timeout - document processing can be slow
  timeout: "5m"

  # Normal logging unless debugging
  logging: normal

# ═══════════════════════════════════════════════════════════
# ENVIRONMENT VARIABLES (available to all steps as {{env.name}})
# ═══════════════════════════════════════════════════════════
env:
  output_dir: "/outputs/processed"
  quality_threshold: "0.85"

# ═══════════════════════════════════════════════════════════
# STEPS (the actual work)
# ═══════════════════════════════════════════════════════════
steps:
  # ─────────────────────────────────────────────────────────
  # Step 1: Extract text from document
  # Uses OCR, so needs higher temperature for flexibility
  # ─────────────────────────────────────────────────────────
  - name: extract_text
    temperature: 0.7              # Override: OCR needs flexibility
    run: |
      Extract all text from the uploaded document.
      Handle OCR errors gracefully.
      Return clean text preserving structure.

  # ─────────────────────────────────────────────────────────
  # Step 2: Classify document type
  # Critical step - use failover for reliability
  # ─────────────────────────────────────────────────────────
  - name: classify_document
    needs: [extract_text]         # Wait for text extraction
    providers:                    # Override: special reliability needs
      - provider: anthropic
        model: claude-opus-4      # Best model first
      - provider: openai
        model: gpt-4o             # Fallback
    run: |
      Based on this text: {{extract_text}}

      Classify document type: invoice, contract, receipt, letter, other
      Return JSON: {"type": "...", "confidence": 0.0-1.0}

  # ─────────────────────────────────────────────────────────
  # Step 3: Retrieve relevant context from RAG database
  # ─────────────────────────────────────────────────────────
  - name: retrieve_context
    needs: [classify_document]
    rag:
      query: "{{classify_document}}"
      server: pgvector
      top_k: 5
      output_format: json

  # ─────────────────────────────────────────────────────────
  # Step 4: Analyze with context
  # Main analysis step - inherits default config
  # ─────────────────────────────────────────────────────────
  - name: analyze_content
    needs: [extract_text, retrieve_context]
    run: |
      Document text: {{extract_text}}
      Similar documents: {{retrieve_context}}

      Extract key information:
      - Dates
      - Amounts
      - Parties involved
      - Key terms

      Return structured JSON.

  # ─────────────────────────────────────────────────────────
  # Step 5: Quality validation with consensus
  # Safety check - requires agreement from multiple providers
  # ─────────────────────────────────────────────────────────
  - name: validate_quality
    needs: [analyze_content]
    consensus:
      prompt: |
        Analysis: {{analyze_content}}

        Is this analysis complete and accurate?
        Answer only: PASS or FAIL
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3                # At least 2 must agree

  # ─────────────────────────────────────────────────────────
  # Step 6: Store results
  # Only runs if validation passed
  # ─────────────────────────────────────────────────────────
  - name: store_results
    needs: [validate_quality]
    if: "{{validate_quality}} == 'PASS'"
    run: |
      Store analysis results to: {{env.output_dir}}
      Analysis: {{analyze_content}}

      Return storage location.
```

**What this example demonstrates:**

1. ✅ Workflow-level defaults with step-level overrides
2. ✅ Property inheritance (most steps use default config)
3. ✅ Targeted overrides (temperature, providers where needed)
4. ✅ Different execution modes (run, rag, consensus)
5. ✅ Dependencies (needs:) and conditions (if:)
6. ✅ Environment variables for configuration
7. ✅ Clear comments explaining **why** each choice was made

---

## Summary: Key Takeaways

1. **Think in layers:** Workflow → Step → Execution (for consensus)
2. **Define once, override where needed:** Use inheritance, don't repeat
3. **Choose the right mode:** run vs embeddings vs rag vs loop
4. **Use needs: for dependencies:** Not execution_order unless required
5. **Failover for critical steps:** Not everything needs it
6. **Comment your why:** Future you will thank you

---

## See Also

- **[Quick Reference](QUICK_REFERENCE.md)** - All properties in one page
- **[Steps Reference](STEPS_REFERENCE.md)** - Detailed step documentation  
- **[CLI Mapping](CLI_MAPPING.md)** - How YAML maps to CLI arguments
- **[Examples](../examples/)** - Real workflows to learn from

---

**Last Updated:** 2026-01-15

**Remember:** Workflows are just scripted sequences of mcp-cli calls. If you understand mcp-cli, you understand workflows. The object model is just about organizing and reusing configuration.
