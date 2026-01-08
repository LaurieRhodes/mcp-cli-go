# Steps Reference

**Version:** workflow/v2.0  
**Purpose:** Detailed reference for all step execution modes

---

## Overview

Steps are the building blocks of workflows. Each step is one of four execution modes:

1. **run:** LLM query with variable interpolation
2. **template:** Call another workflow
3. **embeddings:** Generate vector embeddings
4. **consensus:** Multi-provider validation

All steps inherit properties from `workflow.execution` and can override them.

---

## Step Structure

### Base Step Properties

Every step has these properties:

```yaml
- name: string                  # Required: Unique identifier
  needs: [string]               # Optional: Step dependencies
  if: string                    # Optional: Skip condition
  
  # Inheritable properties (from workflow.execution)
  provider: string              # Optional: Override provider
  model: string                 # Optional: Override model
  temperature: number           # Optional: Override temperature
  max_tokens: number            # Optional: Override max_tokens
  servers: [string]             # Optional: Override servers
  timeout: duration             # Optional: Override timeout
  
  # Execution mode (choose ONE)
  run: string
  template: {...}
  embeddings: {...}
  consensus: {...}
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
| `{{execution.timestamp}}` | Workflow start time | `{{execution.timestamp}}` |

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

**With temperature override:**
```yaml
steps:
  - name: creative
    temperature: 1.5
    run: "Generate 10 creative names for: {{input}}"
```

**With model override:**
```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: fast_task
    model: claude-haiku-4      # Use faster model
    run: "Quick summary: {{input}}"
  
  - name: complex_task
    model: claude-opus-4       # Use more capable model
    run: "Deep analysis: {{input}}"
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

### Important Notes

1. **No property inheritance:** Template calls don't inherit properties from parent
2. **Each workflow independent:** Called workflow uses its own `execution` config
3. **Input passing:** Use `with:` to pass data to called workflow

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

**Passing multiple inputs:**
```yaml
steps:
  - name: analyze
    run: "Analyze: {{input}}"
  
  - name: generate_report
    needs: [analyze]
    template:
      name: report_generator
      with:
        analysis: "{{analyze}}"
        title: "Analysis Report"
        format: "markdown"
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

### Workflow Call Directory Structure

```
config/workflows/
├── main_workflow.yaml        # Parent workflow
├── code_reviewer.yaml        # Called by template
└── report_generator.yaml     # Called by template
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

**Multiple texts:**
```yaml
steps:
  - name: embed_batch
    embeddings:
      model: text-embedding-3-small
      input:
        - "First document"
        - "Second document"
        - "Third document"
```

**From file:**
```yaml
steps:
  - name: embed_file
    embeddings:
      model: text-embedding-3-large
      input_file: "documents.txt"
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

**Save to file:**
```yaml
steps:
  - name: embed_and_save
    embeddings:
      model: text-embedding-3-small
      input: ["doc1", "doc2"]
      output_file: "/tmp/embeddings.json"
      include_metadata: true
```

**Provider override:**
```yaml
execution:
  provider: anthropic         # Default for LLM steps

steps:
  - name: embed
    embeddings:
      provider: openai         # Override for embeddings
      model: text-embedding-3-small
      input: "{{input}}"
```

### Output Format

**With metadata (default):**
```json
{
  "model": "text-embedding-3-small",
  "chunks": [
    {
      "index": 0,
      "text": "First chunk text",
      "start_pos": 0,
      "end_pos": 100,
      "token_count": 50
    }
  ],
  "embeddings": [
    {
      "chunk": {...},
      "vector": [0.123, 0.456, ...]
    }
  ]
}
```

**Without metadata:**
```json
{
  "model": "text-embedding-3-small",
  "vectors": [
    [0.123, 0.456, ...],
    [0.789, 0.012, ...]
  ],
  "count": 2
}
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

**See:** [Consensus Reference](CONSENSUS_REFERENCE.md) for detailed documentation.

**Quick example:**
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

## Step Dependencies (`needs:`)

### Basic Dependencies

**Wait for one step:**
```yaml
steps:
  - name: step1
    run: "First"
  
  - name: step2
    needs: [step1]             # Waits for step1
    run: "Use {{step1}}"
```

**Wait for multiple steps:**
```yaml
steps:
  - name: step1
    run: "First"
  
  - name: step2
    run: "Second"
  
  - name: step3
    needs: [step1, step2]      # Waits for both
    run: "Combine {{step1}} and {{step2}}"
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

### Rules

- ❌ Cannot reference steps that come later in the file
- ❌ Cannot create circular dependencies
- ✅ Can reference multiple steps
- ✅ Enables parallel execution when no dependencies

**Invalid (forward reference):**
```yaml
steps:
  - name: step1
    needs: [step2]             # ERROR: step2 not defined yet
  - name: step2
```

**Invalid (circular):**
```yaml
steps:
  - name: step1
    needs: [step2]
  - name: step2
    needs: [step1]             # ERROR: circular dependency
```

---

## Conditional Execution (`if:`)

### Basic Condition

```yaml
steps:
  - name: check
    run: "Is this valid?"
  
  - name: process
    needs: [check]
    if: "${{ check == 'YES' }}"
    run: "Process the data"
```

**If condition is false, step is skipped.**

### Condition Syntax

```yaml
if: "${{ step_name == 'value' }}"
if: "${{ step_name.result == 'value' }}"
```

### Examples

**Skip if validation fails:**
```yaml
steps:
  - name: validate
    run: "Validate input: {{input}}"
  
  - name: process
    needs: [validate]
    if: "${{ validate == 'VALID' }}"
    run: "Process: {{input}}"
```

**Multiple conditions:**
```yaml
steps:
  - name: check1
    run: "Check 1"
  
  - name: check2
    run: "Check 2"
  
  - name: proceed
    needs: [check1, check2]
    if: "${{ check1 == 'PASS' }}"
    run: "Proceed with operation"
```

---

## Step Result Storage

### How Results Are Stored

Each step's output is stored and available for interpolation:

```yaml
steps:
  - name: step1
    run: "Analyze"
    # Output stored as: step1 = "analysis result"
  
  - name: step2
    run: "Use {{step1}}"
    # Accesses stored result
```

### Result Variables

| Context | Variable | Value |
|---------|----------|-------|
| Step output | `{{step_name}}` | Complete step output |
| Consensus | `{{step_name}}` | Consensus result |
| Embeddings | `{{step_name}}` | Summary or file path |
| Template | `{{step_name}}` | Called workflow output |

---

## Common Patterns

### Pattern 1: Sequential Processing

```yaml
steps:
  - name: parse
    run: "Parse input: {{input}}"
  
  - name: validate
    needs: [parse]
    run: "Validate: {{parse}}"
  
  - name: process
    needs: [validate]
    run: "Process: {{validate}}"
  
  - name: format
    needs: [process]
    run: "Format: {{process}}"
```

---

### Pattern 2: Parallel + Merge

```yaml
steps:
  - name: analyze_security
    run: "Security analysis: {{input}}"
  
  - name: analyze_performance
    run: "Performance analysis: {{input}}"
  
  - name: merge_results
    needs: [analyze_security, analyze_performance]
    run: |
      Merge these analyses:
      Security: {{analyze_security}}
      Performance: {{analyze_performance}}
```

---

### Pattern 3: Progressive Refinement

```yaml
steps:
  - name: draft
    temperature: 1.0
    run: "Write draft: {{input}}"
  
  - name: review
    needs: [draft]
    temperature: 0.3
    run: "Review and critique: {{draft}}"
  
  - name: revise
    needs: [draft, review]
    temperature: 0.7
    run: "Revise draft based on: {{review}}"
  
  - name: final
    needs: [revise]
    temperature: 0.2
    run: "Polish: {{revise}}"
```

---

### Pattern 4: Validation Gate

```yaml
steps:
  - name: process
    run: "Process: {{input}}"
  
  - name: validate
    needs: [process]
    consensus:
      prompt: "Is this acceptable? {{process}}"
      executions:
        - provider: anthropic
        - provider: openai
      require: unanimous
  
  - name: finalize
    needs: [validate]
    if: "${{ validate == 'YES' }}"
    run: "Finalize: {{process}}"
```

---

### Pattern 5: Multi-Model Strategy

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: quick_check
    model: claude-haiku-4      # Fast model
    run: "Quick check: {{input}}"
  
  - name: deep_analysis
    needs: [quick_check]
    model: claude-opus-4       # Capable model
    if: "${{ quick_check == 'NEEDS_ANALYSIS' }}"
    run: "Deep analysis: {{input}}"
  
  - name: final
    needs: [quick_check, deep_analysis]
    run: "Summary: {{quick_check}} {{deep_analysis}}"
```

---

## Best Practices

### 1. Name Steps Descriptively

```yaml
# ✅ Good
steps:
  - name: parse_user_input
  - name: validate_schema
  - name: transform_data
  - name: generate_report

# ❌ Bad
steps:
  - name: step1
  - name: step2
  - name: step3
```

---

### 2. Use Dependencies Effectively

```yaml
# ✅ Good - Explicit dependencies
steps:
  - name: fetch_data
  - name: process_data
    needs: [fetch_data]
  - name: save_results
    needs: [process_data]

# ❌ Bad - No dependencies (may execute in wrong order)
steps:
  - name: fetch_data
  - name: process_data
  - name: save_results
```

---

### 3. Override Properties Judiciously

```yaml
# ✅ Good - Override with clear reason
execution:
  temperature: 0.7

steps:
  - name: creative_brainstorm
    temperature: 1.5           # Higher for creativity
    run: "Generate ideas"
  
  - name: precise_calculation
    temperature: 0.2           # Lower for accuracy
    run: "Calculate results"

# ❌ Bad - Unnecessary overrides
steps:
  - name: step1
    provider: anthropic        # Already set in execution
    temperature: 0.7           # Already set in execution
```

---

### 4. Handle Long Prompts with YAML Multi-line

```yaml
# ✅ Good - Readable
steps:
  - name: review
    run: |
      Review this code for:
      1. Security vulnerabilities
      2. Performance issues
      3. Style compliance
      
      Code:
      {{input}}

# ❌ Bad - Hard to read
steps:
  - name: review
    run: "Review this code for: 1. Security vulnerabilities 2. Performance issues 3. Style compliance Code: {{input}}"
```

---

## Troubleshooting

### Step Not Executing

**Check:**
1. Does `needs:` reference exist?
2. Is `if:` false?
3. Did a dependency fail?
4. Is exactly ONE execution mode specified?

---

### Variable Not Interpolating

**Check:**
1. Is variable name spelled correctly?
2. Does referenced step exist?
3. Did referenced step run successfully?
4. Is syntax correct: `{{var}}` not `{var}` or `${{var}}`

---

### Provider/Model Not Working

**Check:**
1. Is provider spelled correctly?
2. Is model available for provider?
3. Are API keys configured?
4. Is network accessible?

---

## See Also

- **[Object Model](OBJECT_MODEL.md)** - TypeScript interfaces
- **[Inheritance Guide](INHERITANCE_GUIDE.md)** - Property inheritance
- **[Consensus Reference](CONSENSUS_REFERENCE.md)** - Multi-provider validation
- **[Loops Reference](LOOPS_REFERENCE.md)** - Iterative execution
- **[Quick Reference](QUICK_REFERENCE.md)** - One-page overview

---

**Remember:** Every step is fundamentally an mcp-cli call. The workflow system adds orchestration, dependencies, and composition on top of this foundation.
