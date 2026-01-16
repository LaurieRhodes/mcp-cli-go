# Workflow Validation Guide

**Status:** ✅ Production Ready  
**Last Updated:** January 16, 2026

---

## Overview

mcp-cli performs comprehensive workflow validation before execution to catch configuration errors early with helpful error messages. All validation rules enforce the workflow schema and ensure correct parallel execution behavior.

---

## Validation Triggers

**Automatic Validation:**
- ✅ Before every workflow execution
- ✅ During workflow loading (basic structure)
- ✅ During parallel execution setup

**Manual Validation:**
- Not currently available (validation is automatic)

---

## Validation Categories

### 1. Execution Context Validation

Validates workflow-level execution settings.

#### max_workers Validation

**Rule:** Must be a positive integer between 1 and 100

**Invalid Examples:**
```yaml
execution:
  parallel: true
  max_workers: -5  # ❌ ERROR: Cannot be negative
```

```yaml
execution:
  parallel: true
  max_workers: 150  # ❌ ERROR: Too high (max: 100)
```

**Error Message:**
```
Step 'execution': max_workers cannot be negative
  Hint: Set max_workers to a positive integer (recommended: 3-10)
```

**Valid Example:**
```yaml
execution:
  parallel: true
  max_workers: 5  # ✅ Valid
```

#### on_error Validation

**Rule:** Must be one of: `cancel_all`, `complete_running`, `continue`

**Invalid Example:**
```yaml
execution:
  parallel: true
  on_error: stop_immediately  # ❌ ERROR: Invalid policy
```

**Error Message:**
```
Step 'execution': invalid error policy 'stop_immediately'
  Hint: Valid values: 'cancel_all', 'complete_running', 'continue'
```

**Valid Examples:**
```yaml
execution:
  parallel: true
  on_error: cancel_all  # ✅ Stop everything on first error
```

```yaml
execution:
  parallel: true
  on_error: complete_running  # ✅ Finish running steps, don't start new
```

```yaml
execution:
  parallel: true
  on_error: continue  # ✅ Keep going despite errors
```

#### Parallel Settings Consistency

**Rule:** If `parallel: false`, `max_workers` and `on_error` should not be set

**Warning Example:**
```yaml
execution:
  parallel: false
  max_workers: 5  # ⚠️ WARNING: parallel is disabled
  on_error: cancel_all  # ⚠️ WARNING: parallel is disabled
```

**Error Message:**
```
Step 'execution': max_workers is set but parallel execution is disabled
  Hint: Set 'parallel: true' to enable parallel execution
```

#### logging Level Validation

**Rule:** Must be one of: `error`, `warn`, `info`, `step`, `steps`, `debug`, `verbose`, `noisy`

**Invalid Example:**
```yaml
execution:
  logging: detailed  # ❌ ERROR: Invalid log level
```

**Error Message:**
```
Step 'execution': invalid log level 'detailed'
  Hint: Valid values: error, warn, info, step, steps, debug, verbose, noisy
```

---

### 2. Step Execution Mode Validation

Validates that each step has exactly one execution mode.

#### Missing Execution Mode

**Invalid Example:**
```yaml
steps:
  - name: broken_step
    # ❌ ERROR: No execution mode specified
```

**Error Message:**
```
Step 'broken_step': no execution mode specified
  Hint: Steps must have ONE of: run, template, rag, embeddings, consensus, or loop
```

#### Multiple Execution Modes

**Invalid Example:**
```yaml
steps:
  - name: broken_step
    run: "Query 1"  # ❌ ERROR: Two modes specified
    template:
      name: other_workflow
```

**Error Message:**
```
Step 'broken_step': multiple execution modes specified
  Hint: Steps can only have ONE execution mode (run, template, rag, embeddings, consensus, or loop)
```

---

### 3. Dependency Validation

Validates step dependencies and prevents circular dependencies.

#### Non-Existent Dependency

**Invalid Example:**
```yaml
steps:
  - name: step1
    run: "First"
  
  - name: step2
    needs: [nonexistent]  # ❌ ERROR: Doesn't exist
    run: "Second"
```

**Error Message:**
```
Step 'step2': dependency 'nonexistent' does not exist
  Hint: Available steps/loops: step1, step2
```

#### Self-Dependency

**Invalid Example:**
```yaml
steps:
  - name: step1
    needs: [step1]  # ❌ ERROR: Cannot depend on itself
    run: "Test"
```

**Error Message:**
```
Step 'step1': step cannot depend on itself
  Hint: Remove 'step1' from the needs array
```

#### Circular Dependency

**Invalid Example:**
```yaml
steps:
  - name: step1
    needs: [step2]  # ❌ ERROR: Circular!
    run: "First"
  
  - name: step2
    needs: [step1]  # ❌ ERROR: Circular!
    run: "Second"
```

**Error Message:**
```
Step 'workflow': circular dependency detected: step1 → step2 → step1
  Hint: Remove one of the dependencies to break the cycle
```

**Complex Circular Dependency:**
```yaml
steps:
  - name: A
    needs: [C]  # ❌ ERROR: A → C → B → A
  - name: B
    needs: [A]
  - name: C
    needs: [B]
```

**Error Message:**
```
Step 'workflow': circular dependency detected: A → C → B → A
  Hint: Remove one of the dependencies to break the cycle
```

---

### 4. Variable Reference Validation

Validates that variable references have corresponding dependencies.

**Rule:** Any `{{step_name}}` reference must have `step_name` in the `needs` array

**Invalid Example:**
```yaml
steps:
  - name: fetch_data
    run: "Fetch data"
  
  - name: process_data
    # ❌ ERROR: Missing needs: [fetch_data]
    run: "Process {{fetch_data}}"
```

**Error Message:**
```
step 'process_data' references '{{fetch_data}}' but 'fetch_data' is not in needs: array (add 'needs: [fetch_data]' to ensure correct execution order)
```

**Valid Example:**
```yaml
steps:
  - name: fetch_data
    run: "Fetch data"
  
  - name: process_data
    needs: [fetch_data]  # ✅ Correct
    run: "Process {{fetch_data}}"
```

#### Built-In Variables (Exempt)

These variables don't require `needs` declarations:

- `{{input}}` - User input
- `{{env.VAR}}` - Environment variables
- `{{loop}}` - Loop context
- `{{iteration}}` - Current iteration
- `{{item}}` - Current loop item
- `{{index}}` - Current loop index
- `{{consensus}}` - Consensus results

**Valid Example:**
```yaml
steps:
  - name: greet
    run: "Hello {{input}}"  # ✅ No needs required
```

---

### 5. Loop Mode Validation

Validates loop-specific configuration.

#### Missing Workflow Name

**Invalid Example:**
```yaml
steps:
  - name: process_loop
    loop:
      # ❌ ERROR: workflow is required
      max_iterations: 10
```

**Error Message:**
```
Step 'process_loop': loop workflow name is required
  Hint: Example: loop:
          workflow: child_workflow
          max_iterations: 5
```

#### Invalid max_iterations

**Invalid Example:**
```yaml
steps:
  - name: process_loop
    loop:
      workflow: child_workflow
      max_iterations: 0  # ❌ ERROR: Must be > 0
```

**Error Message:**
```
Step 'process_loop': max_iterations must be > 0
  Hint: Set a reasonable limit like max_iterations: 10
```

#### Parallel Loop Without max_workers

**Invalid Example:**
```yaml
steps:
  - name: process_loop
    loop:
      workflow: child_workflow
      parallel: true  # ❌ ERROR: max_workers required
      # Missing: max_workers
```

**Error Message:**
```
Step 'process_loop': max_workers must be > 0 when parallel is true
  Hint: Set max_workers to control concurrency (e.g., max_workers: 3)
```

---

### 6. Consensus Mode Validation

Validates consensus-specific configuration.

#### Missing Prompt

**Invalid Example:**
```yaml
steps:
  - name: validate
    consensus:
      # ❌ ERROR: prompt is required
      executions:
        - provider: anthropic
          model: claude-sonnet-4
```

**Error Message:**
```
Step 'validate': consensus prompt is required
  Hint: Example: consensus:
          prompt: "Is this valid?"
          executions: [...]
```

#### Insufficient Executions

**Invalid Example:**
```yaml
steps:
  - name: validate
    consensus:
      prompt: "Is this valid?"
      executions:  # ❌ ERROR: Need at least 2
        - provider: anthropic
          model: claude-sonnet-4
```

**Error Message:**
```
Step 'validate': at least 2 executions required for consensus
  Hint: Add multiple provider/model combinations to get consensus
```

---

### 7. RAG Mode Validation

Validates RAG-specific configuration.

#### Missing Server

**Invalid Example:**
```yaml
steps:
  - name: search
    rag:
      # ❌ ERROR: server is required
      query: "search terms"
```

**Error Message:**
```
Step 'search': RAG server name is required
  Hint: Example: rag:
          server: pgvector
          query: "search terms"
```

#### Missing Query

**Invalid Example:**
```yaml
steps:
  - name: search
    rag:
      server: pgvector
      # ❌ ERROR: query is required
```

**Error Message:**
```
Step 'search': RAG query is required
  Hint: Specify the search query for RAG retrieval
```

---

## Validation Error Format

### Single Error

```
Workflow validation failed with 1 error(s):

1. Step 'step1': dependency 'nonexistent' does not exist
  Hint: Available steps/loops: step1, step2, step3

═══════════════════════════════════════════════════════════
[Helpful reference documentation]
═══════════════════════════════════════════════════════════
```

### Multiple Errors

```
Workflow validation failed with 3 error(s):

1. Step 'execution': max_workers cannot be negative
  Hint: Set max_workers to a positive integer (recommended: 3-10)

2. Step 'step1': dependency 'nonexistent' does not exist
  Hint: Available steps/loops: step1, step2

3. Step 'workflow': circular dependency detected: step1 → step2 → step1
  Hint: Remove one of the dependencies to break the cycle

═══════════════════════════════════════════════════════════
[Helpful reference documentation]
═══════════════════════════════════════════════════════════
```

---

## Reference Documentation (Included in Errors)

Every validation error includes this helpful reference:

```
═══════════════════════════════════════════════════════════
Valid step execution modes:
  • run: "LLM query with {{variables}}"
  • template:
      name: workflow_name
      with:
        param: value
  • rag:
      server: pgvector
      query: "search query"
  • loop:
      workflow: child_workflow
      max_iterations: 10
  • embeddings: {...}
  • consensus: {...}
───────────────────────────────────────────────────────────
Parallel execution settings (execution block):
  parallel: true               # Enable parallel execution
  max_workers: 3               # Concurrent steps (1-100)
  on_error: cancel_all         # cancel_all|complete_running|continue
───────────────────────────────────────────────────────────
Step dependencies:
  needs: [step1, step2]        # List dependencies
  • Dependencies must exist
  • No circular dependencies
  • No self-dependencies
  • Variables used must be in needs array
═══════════════════════════════════════════════════════════
```

---

## Validation Timing

### Before Execution

Validation occurs **before** any workflow steps execute:

```
1. Load workflow YAML
2. Parse configuration
3. ✅ VALIDATE (catches errors here)
4. Execute workflow steps
```

**Benefit:** No wasted LLM API calls for invalid workflows

---

## Complete Validation Checklist

When authoring workflows, ensure:

**Execution Context:**
- [ ] `parallel` is boolean (true/false)
- [ ] `max_workers` is 1-100 (if parallel enabled)
- [ ] `on_error` is one of: cancel_all, complete_running, continue
- [ ] `logging` is valid log level (if set)
- [ ] parallel settings only used when `parallel: true`

**Steps:**
- [ ] Each step has exactly one execution mode
- [ ] All `needs` dependencies exist
- [ ] No circular dependencies
- [ ] No self-dependencies
- [ ] All `{{variable}}` references in `needs` array

**Loops:**
- [ ] `workflow` name specified
- [ ] `max_iterations` > 0
- [ ] `max_workers` set if `parallel: true`

**Consensus:**
- [ ] `prompt` specified
- [ ] At least 2 executions

**RAG:**
- [ ] `server` specified
- [ ] `query` specified

---

## Testing Validation

### Test Invalid Workflows

Create test workflows to verify validation:

```bash
# Test negative max_workers
./mcp-cli --workflow test_invalid_max_workers

# Test invalid on_error
./mcp-cli --workflow test_invalid_on_error

# Test circular dependency
./mcp-cli --workflow test_circular_dependency

# Test missing dependency
./mcp-cli --workflow test_missing_dependency

# Test self-dependency
./mcp-cli --workflow test_self_dependency
```

---

## Common Validation Errors

### Top 5 Errors

1. **Missing `needs` for variable reference**
   - Most common with parallel workflows
   - Fix: Add referenced step to `needs` array

2. **Invalid `on_error` value**
   - Typo in error policy
   - Fix: Use cancel_all, complete_running, or continue

3. **Non-existent dependency**
   - Typo in step name
   - Fix: Check step name spelling (case-sensitive)

4. **Negative or zero `max_workers`**
   - Invalid worker count
   - Fix: Use positive integer (recommended: 3-10)

5. **Circular dependency**
   - Steps depend on each other
   - Fix: Remove one dependency to break cycle

---

## Debugging Validation Errors

### Step 1: Read the Error Message

Validation errors include:
- Exact location (step name)
- Specific problem
- **Hint with solution**

### Step 2: Check the Reference

Every error includes reference documentation showing:
- Valid syntax examples
- Allowed values
- Common patterns

### Step 3: Fix and Retry

1. Fix the specific error mentioned
2. Run workflow again
3. Repeat until all errors resolved

### Example Debug Session

```
❌ Error: invalid error policy 'stop'
📋 Fix: Change to 'cancel_all'
✅ Fixed

❌ Error: dependency 'fetch-data' does not exist
📋 Fix: Change to 'fetch_data' (underscore, not dash)
✅ Fixed

✅ All validation passed!
```

---

## Best Practices

### 1. Test Incrementally

Add one step at a time and validate:

```yaml
# Start simple
steps:
  - name: step1
    run: "Test"

# Add dependency
steps:
  - name: step1
    run: "Test"
  - name: step2
    needs: [step1]  # Validate now!
    run: "Use {{step1}}"
```

### 2. Use Descriptive Names

Clear step names make errors easier to understand:

```yaml
# ❌ Poor names
steps:
  - name: s1
  - name: s2
    needs: [s1]  # Error: hard to understand

# ✅ Clear names
steps:
  - name: fetch_config
  - name: validate_config
    needs: [fetch_config]  # Error: easy to understand
```

### 3. Start Sequential, Then Parallelize

1. Build workflow sequentially first
2. Validate it works
3. Add `parallel: true` and `needs` arrays
4. Validate again

### 4. Document Dependencies

Add comments explaining why dependencies exist:

```yaml
steps:
  - name: fetch_data
    run: "Fetch data"
  
  - name: process_data
    needs: [fetch_data]  # Needs raw data before processing
    run: "Process {{fetch_data}}"
```

---

## Validation Coverage

### What's Validated ✅

- ✅ Execution context settings
- ✅ Step execution modes
- ✅ Dependencies (existence, cycles, self)
- ✅ Variable references
- ✅ Loop configuration
- ✅ Consensus configuration
- ✅ RAG configuration
- ✅ Enum values (on_error, logging)
- ✅ Numeric ranges (max_workers)

### What's Not Validated ❌

- ❌ Provider/model availability (runtime check)
- ❌ MCP server connectivity (runtime check)
- ❌ API authentication (runtime check)
- ❌ File paths existence (runtime check)
- ❌ Workflow name references (template mode)

**Note:** Runtime errors have separate error handling

---

## Future Enhancements

Potential future validation improvements:

1. **Schema Validation** - JSON Schema validation
2. **Provider Validation** - Check provider/model combinations
3. **Dry-Run Mode** - Validate without executing
4. **Workflow Linting** - Style and best practice checks
5. **Dependency Visualization** - Graph of dependencies

---

## Related Documentation

- [PARALLEL_EXECUTION.md](schema/PARALLEL_EXECUTION.md) - Parallel execution guide
- [QUICK_REFERENCE.md](schema/QUICK_REFERENCE.md) - Syntax reference
- [OBJECT_MODEL.md](schema/OBJECT_MODEL.md) - Type system

---

**Status:** ✅ **VALIDATION SYSTEM COMPLETE**  
**Coverage:** ✅ **ALL SCHEMA PROPERTIES**  
**Error Messages:** ✅ **HELPFUL WITH HINTS**  
**Testing:** ✅ **ALL CASES VALIDATED**

---

**Last Updated:** January 16, 2026
