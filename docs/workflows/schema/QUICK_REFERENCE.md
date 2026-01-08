# Workflow Schema Quick Reference

**Version:** workflow/v2.0  
**Core Concept:** Workflows sequence mcp-cli calls with shared configuration

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

**This executes as:**
```bash
# Step 1
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data "Analyze this: user input"

# Step 2 (waits for step 1)
mcp-cli --provider anthropic --model claude-sonnet-4 \
  --input-data "Summarize: step 1 output"
```

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

### MCPQuery (Base Object)

All steps and executions inherit from this base object.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `provider` | string | Yes | - | AI provider: `anthropic`, `openai`, `deepseek`, `ollama`, `gemini`, `openrouter` |
| `model` | string | Yes | - | Model identifier (e.g., `claude-sonnet-4`, `gpt-4o`, `qwen2.5:32b`) |
| `temperature` | float | No | 0.7 | Randomness: 0.0 (deterministic) to 2.0 (creative) |
| `max_tokens` | int | No | (auto) | Maximum tokens in response |
| `servers` | string[] | No | [] | MCP servers to enable (e.g., `[filesystem, brave-search]`) |
| `skills` | string[] | No | [] | Anthropic Skills to enable (e.g., `[docx, pdf, xlsx]`) |
| `timeout` | duration | No | 60s | Call timeout: `"30s"`, `"5m"`, `"1h"` |

---

### Step Object

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique step identifier |
| `needs` | string[] | No | [] | Step dependencies - waits for these steps to complete |
| `if` | string | No | - | Skip step if expression evaluates to false (e.g., `${{ step == 'value' }}`) |
| **Inherited from MCPQuery** | | | | **Can override any MCPQuery property** |
| `provider` | string | No | (inherited) | Override provider for this step |
| `model` | string | No | (inherited) | Override model for this step |
| `temperature` | float | No | (inherited) | Override temperature for this step |
| `max_tokens` | int | No | (inherited) | Override max_tokens for this step |
| `servers` | string[] | No | (inherited) | Override servers for this step |
| `skills` | string[] | No | (inherited) | Override skills for this step |
| `timeout` | duration | No | (inherited) | Override timeout for this step |
| **Execution Mode (choose ONE)** | | | | |
| `run` | string | No | - | LLM prompt with `{{variable}}` interpolation |
| `template` | TemplateCall | No | - | Call another workflow |
| `embeddings` | EmbeddingsConfig | No | - | Generate vector embeddings |
| `consensus` | ConsensusConfig | No | - | Multi-provider validation |

---

### EmbeddingsConfig Object

Generates vector embeddings from text with full control over chunking and output.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| **Input Source** (one required) | | | | |
| `input` | string or array | Yes* | - | Text to embed (string or array of strings) |
| `input_file` | string | Yes* | - | Input file path (alternative to `input`) |
| **Provider Override** | | | | |
| `provider` | string | No | (inherited) | AI provider: `openai`, `deepseek`, `openrouter` |
| `model` | string | No | (inherited) | Embedding model (e.g., `text-embedding-3-small`) |
| **Chunking Configuration** | | | | |
| `chunk_strategy` | string | No | `"sentence"` | Chunking strategy: `sentence`, `paragraph`, `fixed` |
| `max_chunk_size` | int | No | 512 | Maximum chunk size in tokens |
| `overlap` | int | No | 0 | Overlap between chunks in tokens |
| **Model Configuration** | | | | |
| `dimensions` | int | No | (auto) | Number of dimensions (for supported models) |
| **Output Configuration** | | | | |
| `encoding_format` | string | No | `"float"` | Encoding format: `float`, `base64` |
| `include_metadata` | bool | No | true | Include chunk and model metadata |
| `output_format` | string | No | `"json"` | Output format: `json`, `csv`, `compact` |
| `output_file` | string | No | (stdout) | Output file path |

\* One of `input` or `input_file` is required

**Example with inline text:**
```yaml
steps:
  - name: embed_docs
    embeddings:
      model: text-embedding-3-small
      input:
        - "First document to embed"
        - "Second document to embed"
      chunk_strategy: sentence
      max_chunk_size: 512
      overlap: 50
```

**Example with file input:**
```yaml
steps:
  - name: embed_file
    embeddings:
      model: text-embedding-3-large
      input_file: "documents.txt"
      chunk_strategy: paragraph
      max_chunk_size: 1024
      dimensions: 3072
      output_file: "embeddings.json"
```

**Example with provider override:**
```yaml
execution:
  provider: anthropic  # Default for all steps

steps:
  - name: embed_with_openai
    embeddings:
      provider: openai              # Override for embeddings
      model: text-embedding-3-small
      input: "{{text_to_embed}}"
```

---

### ConsensusConfig Object

Runs multiple AI providers in parallel and requires agreement.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `prompt` | string | Yes | - | Prompt sent to all providers (supports `{{variables}}`) |
| `executions` | MCPQuery[] | Yes | - | Array of provider configurations to run in parallel |
| `require` | string | Yes | - | Agreement threshold: `"unanimous"`, `"majority"`, `"2/3"` |
| `timeout` | duration | No | 60s | Timeout for entire consensus operation |

**Execution inheritance:**
Each execution in `executions[]` inherits from step-level MCPQuery properties but can override them.

**Example:**
```yaml
steps:
  - name: security_check
    temperature: 0.2            # Inherited by all executions
    consensus:
      prompt: "Is this code safe? Answer YES or NO: {{code}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          # Inherits temperature: 0.2
        
        - provider: openai
          model: gpt-4o
          temperature: 0.1      # Override for this execution
        
        - provider: deepseek
          model: deepseek-chat
          # Inherits temperature: 0.2
      
      require: 2/3              # Need 2 out of 3 to agree
```

---

### TemplateCall Object

Calls another workflow.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Name of workflow to call |
| `with` | object | No | {} | Input data passed to workflow (key-value pairs) |

**Example:**
```yaml
steps:
  - name: review_code
    template:
      name: code_reviewer
      with:
        code: "{{input}}"
        language: "go"
        strict: true
```

---

### Loop Object

Iteratively calls a workflow until a condition is met.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Loop identifier |
| `workflow` | string | Yes | - | Name of workflow to call repeatedly |
| `with` | object | Yes | - | Input data for workflow (can use `{{loop.*}}` variables) |
| `max_iterations` | int | Yes | - | Safety limit - stops after this many iterations |
| `until` | string | Yes | - | LLM-evaluated exit condition (e.g., `"Tests pass"`) |
| `on_failure` | string | No | `"fail"` | What to do if iteration fails: `"continue"` or `"fail"` |
| `accumulate` | string | No | - | Variable name to store all iteration results |

**Example:**
```yaml
loops:
  - name: fix_until_works
    workflow: test_and_fix
    with:
      code: "{{input}}"
      previous_error: "{{loop.last.output.error}}"
    max_iterations: 10
    until: "All tests pass"
    on_failure: continue
    accumulate: fix_history
```

---

### Workflow Object

Root workflow definition.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `$schema` | string | Yes | - | Must be `"workflow/v2.0"` |
| `name` | string | Yes | - | Unique workflow identifier |
| `version` | string | Yes | - | Semantic version (e.g., `"1.0.0"`) |
| `description` | string | Yes | - | Human-readable description |
| `execution` | object | Yes | - | Workflow-level defaults (MCPQuery properties) |
| `env` | object | No | {} | Environment variables (key-value pairs) |
| `steps` | Step[] | Yes | [] | Array of steps to execute sequentially |
| `loops` | Loop[] | No | [] | Array of iterative loops |

**Example:**
```yaml
$schema: "workflow/v2.0"
name: my_workflow
version: 1.0.0
description: What this workflow does

execution:
  provider: anthropic
  model: claude-sonnet-4

env:
  API_KEY: "value"

steps:
  - name: step1
    run: "Prompt"

loops:
  - name: loop1
    workflow: other_workflow
    max_iterations: 5
    until: "condition"
```

---

## Workflow Execution Defaults

The `execution:` section in a workflow defines default query properties that all steps inherit.

**Single provider:**
```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7
  max_tokens: 2000
  servers: [filesystem]
  timeout: 60s
```

**Failover chain:**
```yaml
execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o
    - provider: ollama
      model: qwen2.5:32b
  temperature: 0.7
  servers: [filesystem, brave-search]
```

These properties are inherited by all steps and can be overridden at the step level.

---

## Property Inheritance Behavior

This table shows which properties can be overridden at different levels:

| Property | Defined At Workflow | Can Override At Step | Can Override At Consensus Execution |
|----------|---------------------|----------------------|-------------------------------------|
| `provider` | execution: | ✅ Yes | ✅ Yes |
| `model` | execution: | ✅ Yes | ✅ Yes |
| `temperature` | execution: | ✅ Yes | ✅ Yes |
| `max_tokens` | execution: | ✅ Yes | ✅ Yes |
| `servers` | execution: | ✅ Yes | ✅ Yes |
| `timeout` | execution: | ✅ Yes | ✅ Yes |
| `logging` | execution: | ❌ No | ❌ No |
| `no_color` | execution: | ❌ No | ❌ No |

---

## Step Execution Modes

| Mode | Use For | Maps To | Properties |
|------|---------|---------|------------|
| `run:` | LLM query | `mcp-cli --input-data "prompt"` | All MCPQuery |
| `template:` | Call workflow | `mcp-cli --workflow name` | Passes through |
| `embeddings:` | Generate vectors | `mcp-cli --embeddings` | All MCPQuery |
| `consensus:` | Multi-provider | Parallel mcp-cli calls | Per-execution MCPQuery |

---

## Common Patterns

### Provider Failover
```yaml
execution:
  providers:
    - provider: anthropic       # Try first
      model: claude-sonnet-4
    - provider: openai          # Fallback
      model: gpt-4o
    - provider: ollama          # Local backup
      model: qwen2.5:32b
```

### Step Dependencies
```yaml
steps:
  - name: step1
    run: "First step"
  
  - name: step2
    needs: [step1]              # Waits for step1
    run: "Use {{step1}}"
  
  - name: step3
    needs: [step1, step2]       # Waits for both
    run: "Use {{step1}} and {{step2}}"
```

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
      require: 2/3              # Need 2 of 3 to agree
```

### Iterative Loops
```yaml
loops:
  - name: refine
    workflow: code_reviewer
    with:
      code: "{{input}}"
      feedback: "{{loop.last.output}}"
    max_iterations: 5
    until: "The review says PASS"
    on_failure: continue
```

### Property Override
```yaml
execution:
  provider: anthropic           # Default for all steps
  temperature: 0.7

steps:
  - name: creative_step
    temperature: 1.5            # Override for creativity
    run: "Generate ideas"
  
  - name: analytical_step
    temperature: 0.2            # Override for precision
    run: "Analyze data"
```

### MCP Server Integration
```yaml
execution:
  servers: [filesystem, brave-search]  # Available to all steps

steps:
  - name: search
    run: "Search for recent news about {{topic}}"
    # brave-search MCP server automatically available
  
  - name: read_file
    run: "Read the file at {{filepath}}"
    # filesystem MCP server automatically available
```

---

## Variable Interpolation

| Context | Variable | Example | Description |
|---------|----------|---------|-------------|
| Input | `{{input}}` | `{{input}}` | User-provided input data |
| Step output | `{{step_name}}` | `{{analyze}}` | Output from step named "analyze" |
| Loop | `{{loop.iteration}}` | `{{loop.iteration}}` | Current iteration number (1-based) |
| Loop | `{{loop.last.output}}` | `{{loop.last.output}}` | Previous iteration result |
| Loop | `{{loop.history}}` | `{{loop.history}}` | All iteration results |
| Execution | `{{execution.timestamp}}` | `{{execution.timestamp}}` | When workflow started |
| Workflow | `{{workflow.name}}` | `{{workflow.name}}` | Workflow identifier |

---

## Exit Conditions (Loops)

**How it works:**
1. Execute loop iteration (calls workflow)
2. Send result + exit condition to LLM
3. LLM returns YES or NO
4. If YES: exit loop. If NO: continue

**Examples:**
```yaml
until: "The code passes all tests"              # Check for success
until: "The review says APPROVED"               # Look for keyword
until: "Error count is zero"                    # Numeric check
until: "The output contains 'COMPLETE'"         # Pattern match
```

**Best practices:**
- Be specific and clear
- Mention the exact condition
- Avoid vague language like "it looks good"

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

### With MCP Servers
```yaml
execution:
  provider: anthropic
  servers: [filesystem, brave-search]
steps:
  - name: step1
    run: "Search for: {{query}}"
```
**Equals:**
```bash
mcp-cli --provider anthropic \
  --server filesystem \
  --server brave-search \
  --input-data "Search for: user query"
```

### With Overrides
```yaml
execution:
  provider: anthropic
  temperature: 0.7
steps:
  - name: step1
    temperature: 0.3
    run: "Prompt"
```
**Equals:**
```bash
mcp-cli --provider anthropic --temperature 0.3 \
  --input-data "Prompt"
```

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
  # ... other MCPQuery properties

env:                            # Optional: Environment vars
  KEY: value

steps:                          # Sequential execution
  - name: step1
    run: "prompt"
  - name: step2
    needs: [step1]
    run: "prompt using {{step1}}"

loops:                          # Optional: Iterative execution
  - name: loop1
    workflow: other_workflow
    max_iterations: 5
    until: "condition met"
```

---

## Tips

**1. Start Simple:**
```yaml
$schema: "workflow/v2.0"
name: hello
version: 1.0.0
description: Hello world

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: greet
    run: "Say hello to {{input}}"
```

**2. Add Dependencies:**
```yaml
steps:
  - name: step1
    run: "First"
  - name: step2
    needs: [step1]              # Add this
    run: "Use {{step1}}"
```

**3. Add Failover:**
```yaml
execution:
  providers:                    # Change to array
    - provider: anthropic
      model: claude-sonnet-4
    - provider: ollama          # Backup
      model: qwen2.5:32b
```

**4. Add Validation:**
```yaml
steps:
  - name: validate
    consensus:                  # Change to consensus
      prompt: "Is this safe?"
      executions:
        - provider: anthropic
        - provider: openai
      require: 2/2
```

---

## Common Mistakes

❌ **Forgetting schema version:**
```yaml
name: my_workflow              # Missing $schema
```
✅ **Correct:**
```yaml
$schema: "workflow/v2.0"       # Always first
name: my_workflow
```

❌ **Circular dependencies:**
```yaml
steps:
  - name: step1
    needs: [step2]
  - name: step2
    needs: [step1]             # Circular!
```

❌ **Referencing non-existent step:**
```yaml
steps:
  - name: step1
    run: "{{step2}}"           # step2 doesn't exist yet
```

❌ **Wrong provider names:**
```yaml
execution:
  provider: claude             # Wrong! Use "anthropic"
  provider: chatgpt            # Wrong! Use "openai"
```
✅ **Correct provider names:**
- `anthropic`
- `openai`
- `deepseek`
- `ollama`
- `gemini`

---

## See Also

- **[CLI Mapping](CLI_MAPPING.md)** - Complete property → CLI argument mapping
- **[Full Schema](SCHEMA.md)** - Detailed schema documentation
- **[Loop Guide](LOOPS.md)** - Deep dive on iterative execution
- **[Examples](../examples/)** - Working example workflows

---

**Remember:** Workflows are just sequences of mcp-cli calls with shared configuration. Every step property maps to a CLI argument. Use inheritance to avoid repetition.
