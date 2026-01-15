# RAG in Workflows

RAG steps can be integrated into workflows for powerful AI-powered applications.

## Basic RAG Workflow

```yaml
name: basic_rag_search
version: 1.0.0
description: Simple RAG search workflow

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: retrieve
    rag:
      query: "multi-factor authentication requirements"
      server: pgvector
      strategies: [default]
      top_k: 5
      output_format: json

output:
  format: json
  template: |
    {{step.retrieve}}
```

**Usage:**
```bash
mcp-cli --workflow basic_rag_search
```

## RAG with Variables

```yaml
name: variable_rag_search
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: retrieve
    rag:
      query: "{{user_query}}"
      server: pgvector
      strategies: [default]
      top_k: "{{top_k}}"

output:
  format: json
```

**Usage:**
```bash
mcp-cli --workflow variable_rag_search \
  --var user_query="authentication requirements" \
  --var top_k=10
```

## RAG + LLM Generation

Retrieve context, then generate answer:

```yaml
name: rag_qa
version: 1.0.0
description: RAG-powered Q&A

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: retrieve
    rag:
      query: "{{question}}"
      server: pgvector
      strategies: [default, context]
      fusion: rrf
      top_k: 5
      output_format: compact
  
  - name: answer
    run: |
      Based on these retrieved documents:
      {{step.retrieve}}
      
      Answer this question: {{question}}
      
      Provide a clear, accurate answer with references to the specific documents.

output:
  format: text
  template: |
    {{step.answer}}
```

**Usage:**
```bash
mcp-cli --workflow rag_qa \
  --var question="What are the MFA requirements for privileged users?"
```

## Multi-Strategy RAG

```yaml
name: multi_strategy_search
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: semantic_search
    rag:
      query: "{{query}}"
      server: pgvector
      strategies: [default]
      top_k: 10
      output_format: compact
  
  - name: contextual_search
    rag:
      query: "{{query}}"
      server: pgvector
      strategies: [context]
      top_k: 10
      output_format: compact
  
  - name: combine
    run: |
      Semantic results: {{step.semantic_search}}
      Contextual results: {{step.contextual_search}}
      
      Analyze and synthesize the best matches from both searches.

output:
  format: text
```

## RAG in Loops

Process multiple queries:

```yaml
name: batch_rag_search
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: search_all
    loop:
      items: "{{queries}}"
      workflow: search_single
      max_workers: 5
    
output:
  format: json
```

**Child workflow (search_single.yaml):**
```yaml
name: search_single
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  - name: retrieve
    rag:
      query: "{{input}}"
      server: pgvector
      strategies: [default]
      top_k: 5

output:
  format: json
```

**Usage:**
```bash
mcp-cli --workflow batch_rag_search \
  --var queries='["authentication", "encryption", "access control"]'
```

## Real-World Example: Compliance Assessment

```yaml
name: assess_policy_statement
version: 1.0.0
description: Assess policy statement against compliance controls

execution:
  provider: anthropic
  model: claude-sonnet-4-20250514

steps:
  # Extract statement text from input
  - name: parse_text
    run: |
      Input: {{input}}
      Extract the "text" field from the JSON.
      Output only the text value.
  
  # Search for similar compliance controls
  - name: retrieve
    rag:
      query: "{{step.parse_text}}"
      server: pgvector
      strategies: [default]
      top_k: 5
      expand_query: false
      output_format: json
  
  # Assess compliance
  - name: assess
    run: |
      Policy Statement: {{step.parse_text}}
      Similar Controls: {{step.retrieve}}
      
      Analyze the policy statement against the retrieved controls.
      Output JSON: 
      {
        "statement_id": "...",
        "compliance_status": "compliant|partial|non-compliant",
        "reasoning": "...",
        "best_match": "..."
      }

output:
  format: json
  template: |
    {{step.assess}}
```

## Output Formats

### JSON (Structured)
```yaml
rag:
  output_format: json
```
Returns full structured data with all metadata.

### Compact (Minimal)
```yaml
rag:
  output_format: compact
```
Returns JSON without extra whitespace (good for LLM consumption).

### Text (Human-Readable)
```yaml
rag:
  output_format: text
```
Returns formatted text output.

## Advanced Patterns

### Conditional RAG

```yaml
steps:
  - name: check_cache
    run: |
      Check if {{query}} is in cache
      Output: true or false
  
  - name: retrieve
    if: "{{step.check_cache}} == false"
    rag:
      query: "{{query}}"
      server: pgvector
```

### Iterative Refinement

```yaml
steps:
  - name: initial_search
    rag:
      query: "{{query}}"
      top_k: 10
  
  - name: refine_query
    run: |
      Based on these results: {{step.initial_search}}
      Generate a refined search query.
  
  - name: refined_search
    rag:
      query: "{{step.refine_query}}"
      top_k: 5
```

### Multi-Database Search

```yaml
steps:
  - name: search_policies
    rag:
      query: "{{query}}"
      server: policy_db
      top_k: 5
  
  - name: search_controls
    rag:
      query: "{{query}}"
      server: control_db
      top_k: 5
  
  - name: merge_results
    run: |
      Combine and rank results from both searches
```

## Best Practices

### 1. Use Compact Format for LLM Steps

```yaml
- name: retrieve
  rag:
    output_format: compact  # Reduces token usage
```

### 2. Limit Results for Speed

```yaml
- name: quick_lookup
  rag:
    top_k: 3  # Faster than top_k: 20
```

### 3. Use Multiple Strategies for Quality

```yaml
- name: comprehensive_search
  rag:
    strategies: [default, context]
    fusion: rrf
```

### 4. Cache RAG Results

```yaml
- name: retrieve
  rag:
    query: "{{query}}"
    cache: true  # (if supported)
```

### 5. Validate Results

```yaml
- name: validate
  run: |
    RAG results: {{step.retrieve}}
    
    Verify the results are relevant to: {{query}}
    If not relevant, explain why.
```

## Testing RAG Workflows

```bash
# Test with debug logging
mcp-cli --workflow your_workflow --log-level debug

# Test with different parameters
mcp-cli --workflow your_workflow \
  --var query="test query" \
  --var top_k=10

# Test error handling
mcp-cli --workflow your_workflow \
  --var query=""  # Empty query
```

## Common Patterns

### Pattern 1: Simple Retrieval
```yaml
retrieve → output
```

### Pattern 2: RAG + Generation
```yaml
retrieve → generate_answer → output
```

### Pattern 3: Multi-Step RAG
```yaml
retrieve → analyze → refine_query → retrieve_again → output
```

### Pattern 4: Batch Processing
```yaml
loop(retrieve + process) → aggregate → output
```

## Next Steps

- [Usage Guide](usage.md) - Master RAG commands
- [Configuration](configuration.md) - Customize RAG
- [Troubleshooting](troubleshooting.md) - Fix issues

## Important: Server Exposure

### Do NOT Expose RAG Servers to LLM

**❌ Wrong:**
```yaml
execution:
  servers: [pgvector]  # Don't do this!

steps:
  - name: retrieve
    rag:
      server: pgvector
```

**✅ Correct:**
```yaml
execution:
  # No servers needed for RAG steps

steps:
  - name: retrieve
    rag:
      server: pgvector  # Handled internally by mcp-cli
```

### Why?

- **`rag:` steps are handled by mcp-cli**, not by LLM tool calls
- mcp-cli's RAG service directly uses the MCP server
- Exposing pgvector tools to the LLM would:
  - Confuse the LLM (it might try to call pgvector tools directly)
  - Be inefficient (LLM calling tools vs. internal service)
  - Create unnecessary complexity

### When to Expose Servers

Only expose servers that the LLM needs to call tools from:

```yaml
execution:
  servers: [skills]  # Only expose servers with tools LLM should call

steps:
  - name: retrieve
    rag:
      server: pgvector  # Internal - no exposure needed
  
  - name: process
    servers: [skills]  # This step needs skills tools
    skills: [docx]
    run: |
      Create a document with: {{step.retrieve}}
```

### Architecture

```
Workflow Step Type → Handler
──────────────────────────────
rag:               → mcp-cli RAG service (internal)
run: (with tools)  → LLM + exposed MCP servers
run: (no tools)    → LLM only
```

**Key Point:** RAG steps bypass the LLM's tool-calling mechanism entirely.
