# Workflow Variables Reference

**Version:** 2.0  
**Last Updated:** January 13, 2026  
**Status:** Code-Verified Documentation

---

## Overview

This document provides the authoritative reference for variable interpolation in mcp-cli workflows. All information is verified against the actual codebase implementation in `/internal/services/workflow/interpolator.go`.

---

## Table of Contents

1. [How Variables Work](#how-variables-work)
2. [Supported Variable Patterns](#supported-variable-patterns)
3. [Unsupported Patterns](#unsupported-patterns)
4. [Loop Variables](#loop-variables)
5. [Common Patterns](#common-patterns)
6. [Troubleshooting](#troubleshooting)
7. [Examples](#examples)

---

## How Variables Work

### Implementation

Variables in mcp-cli workflows use **simple regex-based string replacement**:

```go
// From interpolator.go
re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
value, ok := i.variables[varName]
result = strings.Replace(result, placeholder, value, -1)
```

**Key characteristics:**
- Variables stored as flat `map[string]string`
- Simple string lookup and replacement
- No JSON parsing
- No nested field access
- No template language (Jinja2, Go templates, etc.)

### Variable Storage

```go
variables := map[string]string{
    "input":              `{"item":{"id":"STMT-001","text":"..."}}`,
    "env.work_dir":       "ism_assessment",
    "step.extract":       "Use authenticated access",
    "loop.iteration":     "5",
    "loop.current":       `{"id":"STMT-001","text":"..."}`,
}
```

**Important:** All values are strings, even complex objects (stored as JSON strings).

---

## Supported Variable Patterns

### Base Variables

These patterns work everywhere (in `run:`, `rag:`, `loop:`, etc.):

```yaml
{{input}}                  # Full input (often JSON string)
{{env.work_dir}}           # Environment variable
{{env.policy_url}}         # Environment variable
{{step.step_name}}         # Output from previous step
{{loop.iteration}}         # Current loop iteration
{{loop.current}}           # Current loop item (JSON string)
{{loop.count}}             # Total items in loop
```

**Syntax rules:**
- Must be wrapped in `{{` and `}}`
- Whitespace is trimmed: `{{ input }}` works
- Only alphanumeric, dots, and underscores: `[a-zA-Z0-9._]`

### Environment Variables

Set in workflow execution section:

```yaml
execution:
  provider: deepseek
  model: deepseek-chat

env:
  work_dir: "output"
  policy_url: "https://example.com/policy.html"
  max_items: "100"
```

Access with `env.` prefix:

```yaml
steps:
  - name: process
    run: |
      Output directory: {{env.work_dir}}
      Policy URL: {{env.policy_url}}
```

### Step Output Variables

Reference output from previous steps:

```yaml
steps:
  - name: fetch_data
    run: |
      Fetch data and return results
    output_var: fetched_data  # Optional: name the output
  
  - name: process_data
    needs: [fetch_data]
    run: |
      Data to process: {{step.fetch_data}}
      # Or if output_var is set: {{fetched_data}}
```

**Note:** Step outputs are also strings. If a step returns JSON, it's stored as a JSON string.

---

## Unsupported Patterns

### Nested Field Access ❌

These patterns **DO NOT WORK**:

```yaml
{{input.text}}             # ❌ No nested fields
{{input.item.text}}        # ❌ No deep nesting
{{data.items}}             # ❌ No object properties
{{step.name.field}}        # ❌ No step result fields
{{env.config.timeout}}     # ❌ No nested env vars
```

**Why:** The interpolator does literal string lookup. It looks for a variable named `"input.text"` (the entire string), not `input` with a `text` field.

**Error you'll see:**
```
failed to interpolate: undefined variables: [input.text]
```

### Template Language Features ❌

These do NOT work:

```yaml
{{input | upper}}          # ❌ No filters
{{items | join(', ')}}     # ❌ No Jinja2
{% if condition %}         # ❌ No conditionals
{% for item in list %}     # ❌ No loops
{{input ?: "default"}}     # ❌ No operators
```

**Why:** mcp-cli does not use Jinja2, Go templates, or any template language. It's pure string replacement.

### Array/Object Indexing ❌

```yaml
{{items[0]}}               # ❌ No array indexing
{{data["key"]}}            # ❌ No object indexing
{{list.first}}             # ❌ No array methods
```

---

## Loop Variables

### Iterate Mode Variables

When using `loop.mode: iterate`, these variables are available:

```yaml
loop:
  mode: iterate
  items: "file:///path/to/data.json"
  workflow: child_workflow
```

**In child workflow:**

```yaml
{{loop.index}}             # 0, 1, 2, ... (zero-based)
{{loop.count}}             # Total number of items
{{loop.current}}           # Current item (JSON string)
{{loop.stats.succeeded}}   # Number of succeeded items so far
{{loop.stats.failed}}      # Number of failed items so far
```

**Example:**
```yaml
steps:
  - name: process
    run: |
      Processing item {{loop.index}} of {{loop.count}}
      Current item data: {{loop.current}}
      Success rate: {{loop.stats.succeeded}}/{{loop.count}}
```

### Refine Mode Variables

When using `loop.mode: refine` (default):

```yaml
{{loop.iteration}}         # 1, 2, 3, ... (one-based)
{{loop.output}}            # Output from current iteration
{{loop.last.output}}       # Output from previous iteration
{{loop.history}}           # All outputs joined with separator
```

### Loop Input Format

**How loops pass data to child workflows:**

```go
// From loop_iterate.go - prepareIterateInput()
inputMap := map[string]interface{}{
    "item": currentItem,  // The current array element
    // Plus any 'with' parameters
}
jsonBytes := json.Marshal(inputMap)
childInput := string(jsonBytes)  // This becomes {{input}}
```

**Result:** Child workflow receives:

```yaml
{{input}} = '{"item":{"id":"STMT-001","text":"Use authenticated access"}}'
```

---

## Common Patterns

### Pattern 1: Simple Value Passing

**Parent workflow:**
```yaml
env:
  output_dir: "results"

steps:
  - name: generate
    run: |
      Generate report
      Save to: {{env.output_dir}}
```

### Pattern 2: Chain Step Outputs

**Pass data between steps:**
```yaml
steps:
  - name: extract
    run: |
      Extract data from source
      Output as JSON
    output_var: extracted_data
  
  - name: transform
    needs: [extract]
    run: |
      Transform this data: {{extracted_data}}
  
  - name: load
    needs: [transform]
    run: |
      Load transformed data: {{step.transform}}
```

### Pattern 3: Loop with LLM Parsing JSON

**Parent workflow:**
```yaml
loop:
  mode: iterate
  items: "file:///data/items.json"  # [{id:"1", text:"..."}, ...]
  workflow: process_item
```

**Child workflow:**
```yaml
steps:
  - name: process
    run: |
      Input JSON: {{input}}
      
      Parse the JSON to extract:
      - id field
      - text field
      
      Then process the text field.
```

**This is the ONLY way to access nested fields!**

### Pattern 4: Conditional Execution via Step Dependencies

Since there's no `if/else`, use step dependencies:

```yaml
steps:
  - name: check_condition
    run: |
      Check if condition is met
      Output: "true" or "false"
  
  - name: action_if_true
    needs: [check_condition]
    run: |
      Condition result: {{step.check_condition}}
      
      If the result is "true", perform action.
      Otherwise, skip.
```

---

## Troubleshooting

### Error: "undefined variables: [variable_name]"

**Cause:** Variable doesn't exist or using unsupported syntax.

**Common mistakes:**

1. **Nested field access:**
   ```yaml
   # ❌ Wrong
   query: "{{input.text}}"
   
   # ✅ Right - parse in run section
   run: |
     Input: {{input}}
     Parse JSON and extract text field
   ```

2. **Typo in variable name:**
   ```yaml
   # ❌ Wrong
   {{step.exract}}  # Typo: exract vs extract
   
   # ✅ Right
   {{step.extract}}
   ```

3. **Missing dependency:**
   ```yaml
   # ❌ Wrong - step doesn't exist yet
   - name: process
     run: "{{step.fetch}}"
   
   # ✅ Right - add dependency
   - name: process
     needs: [fetch]
     run: "{{step.fetch}}"
   ```

### Error: "step X failed: failed to interpolate query"

**Cause:** Using nested field access in RAG or loop fields.

**Solution:**

```yaml
# ❌ Wrong
- name: query_rag
  rag:
    query: "{{input.text}}"

# ✅ Right - LLM parses JSON
- name: search_and_process
  run: |
    Input: {{input}}
    Parse JSON, extract text, then use RAG tools
```

### Variables Not Updating

**Cause:** Variables are set once and don't update dynamically.

**Solution:** Use step chaining:

```yaml
steps:
  - name: get_value
    run: "Output current value"
  
  - name: use_value
    needs: [get_value]
    run: "Use: {{step.get_value}}"  # Gets latest value
```

---

## Examples

### Example 1: Environment Variables

```yaml
execution:
  provider: deepseek
  model: deepseek-chat

env:
  database_host: "192.168.1.100"
  database_port: "5432"
  output_path: "/outputs/results"

steps:
  - name: connect
    run: |
      Connect to database at {{env.database_host}}:{{env.database_port}}
      Save results to {{env.output_path}}
```

### Example 2: Loop Processing with JSON

**Parent:**
```yaml
steps:
  - name: process_all
    loop:
      mode: iterate
      items: "file:///data/items.json"
      workflow: process_item
      parallel: true
      max_workers: 3
```

**Child (process_item.yaml):**
```yaml
steps:
  - name: process
    run: |
      Item data: {{input}}
      
      Instructions:
      1. Parse the JSON {"item": {...}}
      2. Extract the "text" field from item
      3. Process the text
      4. Output result
      
      Loop stats: Processing {{loop.index}} of {{loop.count}}
```

### Example 3: Multi-Step Pipeline

```yaml
steps:
  - name: fetch
    run: |
      Fetch data from API
      Return as JSON
    output_var: raw_data
  
  - name: validate
    needs: [fetch]
    run: |
      Validate this data: {{raw_data}}
      Return validation result
    output_var: validation_result
  
  - name: process
    needs: [validate]
    run: |
      Data: {{raw_data}}
      Validation: {{validation_result}}
      
      If validation passed, process the data
  
  - name: save
    needs: [process]
    run: |
      Save result: {{step.process}}
      To directory: {{env.output_dir}}
```

### Example 4: RAG Query (Correct Pattern)

```yaml
# ❌ WRONG - This fails with "undefined variables"
steps:
  - name: search
    rag:
      query: "{{input.text}}"  # Nested field access
      server: pgvector

# ✅ RIGHT - LLM parses and uses tools
steps:
  - name: search_and_process
    servers: [pgvector]
    run: |
      Input data: {{input}}
      
      1. Parse JSON to extract text field
      2. Use pgvector:search_vectors tool with the text
      3. Process results
```

### Example 5: File Path from Loop

```yaml
# Parent
loop:
  mode: iterate
  items: "file:///data/files.json"  # [{"filename":"doc1.txt"}, ...]
  workflow: process_file
  with:
    base_path: "/data"

# Child
steps:
  - name: process
    run: |
      Input: {{input}}
      
      Parse JSON to get:
      - filename from item object
      - base_path from with parameters
      
      Construct full path: base_path + filename
      Process the file
```

---

## Best Practices

### ✅ DO

1. **Use base variables only**
   ```yaml
   run: "Process {{input}}"
   ```

2. **Let LLMs parse complex data**
   ```yaml
   run: |
     JSON data: {{input}}
     Parse to extract fields...
   ```

3. **Chain steps for data transformation**
   ```yaml
   - name: step1
     run: "..."
     output_var: data
   - name: step2
     needs: [step1]
     run: "Use {{data}}"
   ```

4. **Use descriptive variable names**
   ```yaml
   env:
     ism_policy_url: "https://..."
     assessment_output_dir: "/outputs/assessment"
   ```

### ❌ DON'T

1. **Don't use nested field access**
   ```yaml
   # ❌ Won't work
   run: "{{input.text}}"
   ```

2. **Don't try to use template features**
   ```yaml
   # ❌ Won't work
   run: "{{items | join(', ')}}"
   ```

3. **Don't expect variables to be objects**
   ```yaml
   # ❌ Variables are strings, not objects
   run: "{{data.items[0].name}}"
   ```

4. **Don't use undefined variables**
   ```yaml
   # ❌ Must exist or will fail
   run: "{{nonexistent}}"
   ```

---

## Quick Reference

### Variable Syntax

| Pattern | Supported | Example |
|---------|-----------|---------|
| `{{name}}` | ✅ Yes | `{{input}}` |
| `{{env.var}}` | ✅ Yes | `{{env.work_dir}}` |
| `{{step.name}}` | ✅ Yes | `{{step.extract}}` |
| `{{loop.index}}` | ✅ Yes | `{{loop.index}}` |
| `{{name.field}}` | ❌ No | `{{input.text}}` |
| `{{name[0]}}` | ❌ No | `{{items[0]}}` |
| `{{name \| filter}}` | ❌ No | `{{text \| upper}}` |

### When to Use What

| Task | Pattern |
|------|---------|
| Pass simple value | `{{variable}}` |
| Access env var | `{{env.var_name}}` |
| Use step output | `{{step.step_name}}` |
| Loop item data | `{{input}}` + LLM parsing |
| Extract JSON field | LLM in `run:` section |
| RAG query | Pass `{{input}}` + LLM uses tools |

---

## See Also

- [Authoring Guide](AUTHORING_GUIDE.md) - Complete workflow authoring guide
- [Loops Documentation](LOOPS.md) - Detailed loop patterns
- [Workflow Organization](WORKFLOW_ORGANIZATION.md) - Project structure

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2026-01-13 | Code-verified documentation based on interpolator.go |
| 1.0 | 2025-01-08 | Initial documentation |
