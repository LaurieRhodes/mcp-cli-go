# Loop System Guide

**Status:** Production  
**Date:** January 7, 2026

Complete guide to iterative execution with LLM-evaluated exit conditions.

---

## Table of Contents

- [Overview](#overview)
- [Basic Concepts](#basic-concepts)
- [Loop Structure](#loop-structure)
- [Loop Variables](#loop-variables)
- [Exit Conditions](#exit-conditions)
- [Error Handling](#error-handling)
- [Complete Examples](#complete-examples)
- [Best Practices](#best-practices)

---

## Overview

Loops enable **iterative execution** where:
- A workflow is called repeatedly
- An LLM evaluates whether to continue or exit
- Loop variables track state across iterations
- Results accumulate for later use

**Use cases:**
- Iterative development (code until tests pass)
- Refinement (improve until quality threshold met)
- Exploration (search until answer found)
- Retry with feedback (fix errors until success)

---

## Basic Concepts

### Loop Lifecycle

```
1. Initialize (iteration = 1)
    ↓
2. Interpolate `with` parameters using loop variables
    ↓
3. Execute workflow
    ↓
4. Store result
    ↓
5. Evaluate `until` condition with LLM
    ↓
6. Decision:
   - Condition met → Exit (reason: condition_met)
   - Max iterations → Exit (reason: max_iterations)
   - Error + halt → Exit (reason: error)
   - Otherwise → iteration++, go to step 2
```

### Key Features

- **LLM-evaluated exit:** Natural language conditions
- **Context isolation:** Each workflow call is independent
- **State tracking:** Loop variables carry information between iterations
- **Safety limits:** `max_iterations` prevents runaway loops
- **Flexible error handling:** Continue, halt, or retry on failure

---

## Loop Structure

### Minimal Loop

```yaml
loops:
  - name: my_loop
    workflow: task_workflow
    max_iterations: 5
    until: "The output says DONE"
```

### Complete Loop

```yaml
loops:
  - name: my_loop                    # Required: Unique identifier
    workflow: task_workflow           # Required: Workflow to call
    with:                             # Optional: Input parameters
      spec: "{{requirements}}"
      previous: "{{loop.last.output}}"
      iteration: "{{loop.iteration}}"
    max_iterations: 10                # Required: Safety limit
    until: "The output says PASS"    # Required: Exit condition
    on_failure: continue              # Optional: halt|continue|retry
    accumulate: task_history          # Optional: Store all iterations
```

---

## Loop Variables

Loop variables are automatically available for interpolation in `with` parameters.

### Available Variables

| Variable | Type | Description | Example Value |
|----------|------|-------------|---------------|
| `{{loop.iteration}}` | int | Current iteration number | `1`, `2`, `3`, ... |
| `{{loop.output}}` | string | Current iteration output | `"Generated code"` |
| `{{loop.last.output}}` | string | Previous iteration output | `"Previous attempt"` |
| `{{loop.history}}` | string | All outputs separated by `---` | `"Iter1---Iter2---Iter3"` |

### Usage Example

```yaml
loops:
  - name: develop
    workflow: dev_cycle
    with:
      iteration_num: "{{loop.iteration}}"
      last_attempt: "{{loop.last.output}}"
      full_context: "{{loop.history}}"
    max_iterations: 5
    until: "All tests pass"
```

**Iteration 1:**
```yaml
with:
  iteration_num: "1"
  last_attempt: ""              # Empty on first iteration
  full_context: ""
```

**Iteration 2:**
```yaml
with:
  iteration_num: "2"
  last_attempt: "Output from iteration 1"
  full_context: "Output from iteration 1"
```

**Iteration 3:**
```yaml
with:
  iteration_num: "3"
  last_attempt: "Output from iteration 2"
  full_context: "Output from iteration 1---Output from iteration 2"
```

---

## Exit Conditions

The `until` field contains a natural language condition that an LLM evaluates after each iteration.

### How Condition Evaluation Works

After each iteration:
1. Loop executor extracts the workflow output
2. Constructs evaluation prompt:
   ```
   Evaluate this condition: "The output says PASS"
   
   Output to evaluate:
   [actual output from workflow]
   
   Answer with YES if condition is met, NO if not met.
   ```
3. LLM responds with YES or NO
4. If YES → exit loop early (before max_iterations)

### Writing Good Conditions

**✅ Good Conditions (Clear, Simple):**

```yaml
# Check for specific text
until: "The output says PASS"
until: "The review contains APPROVED"
until: "The response includes COMPLETE"

# Check for absence
until: "There are zero errors"
until: "No issues found"

# Check for threshold
until: "The score is above 90"
until: "Quality rating is excellent"

# Check for success
until: "All tests pass"
until: "Deployment succeeded"
```

**❌ Avoid (Confusing, Ambiguous):**

```yaml
# Don't interpolate output into condition
until: "The output {{loop.output}} is correct"  # Confusing!

# Don't use vague terms
until: "It looks good"  # What is "good"?
until: "The thing is done"  # What "thing"?

# Don't combine multiple checks
until: "Tests pass AND code is clean AND no errors"  # Too complex
```

### Why Avoid Interpolation

**Bad Example:**
```yaml
until: "The output contains {{loop.output}}"
```

After interpolation on iteration 3:
```
Condition: The output contains "Code with bug fix"
Output to evaluate: Code with bug fix
```

The LLM sees the answer IN the condition itself - very confusing!

**Good Example:**
```yaml
until: "The output contains a bug fix"
```

After evaluation:
```
Condition: The output contains a bug fix
Output to evaluate: Code with bug fix
```

Clear and unambiguous.

---

## Error Handling

Control what happens when a workflow execution fails.

### `on_failure` Options

| Value | Behavior | Use When |
|-------|----------|----------|
| `halt` | Stop immediately, fail workflow | Errors are critical |
| `continue` | Log error, continue to next iteration | Errors are expected |
| `retry` | Retry same iteration | Transient failures |

### Examples

**Halt on Error:**
```yaml
loops:
  - name: critical_task
    workflow: important_workflow
    max_iterations: 5
    until: "Task complete"
    on_failure: halt  # Any error stops everything
```

**Continue Despite Errors:**
```yaml
loops:
  - name: exploration
    workflow: try_approach
    max_iterations: 10
    until: "Found solution"
    on_failure: continue  # Keep trying different approaches
```

**Retry Transient Failures:**
```yaml
loops:
  - name: network_task
    workflow: api_call
    max_iterations: 3
    until: "Success"
    on_failure: retry  # Retry same call if network fails
```

---

## Complete Examples

### Example 1: Iterative Code Development

```yaml
$schema: "workflow/v2.0"
name: iterative_developer
version: 1.0.0
description: Develop code iteratively until tests pass

execution:
  provider: deepseek
  model: deepseek-chat
  temperature: 0.5

steps:
  - name: requirements
    run: "Analyze request: {{input}}"
  
  - name: tests
    needs: [requirements]
    run: "Create test criteria for: {{requirements}}"

loops:
  - name: develop_until_pass
    workflow: dev_cycle
    with:
      requirements: "{{requirements}}"
      tests: "{{tests}}"
      previous_code: "{{loop.last.output}}"
      iteration_number: "{{loop.iteration}}"
    max_iterations: 5
    until: "The review says PASS"
    on_failure: continue
    accumulate: development_history

steps:
  - name: report
    needs: [develop_until_pass]
    run: |
      Development complete!
      Iterations: {{loop.iteration}}
      Final code: {{develop_until_pass}}
```

**Workflow called each iteration (dev_cycle.yaml):**
```yaml
$schema: "workflow/v2.0"
name: dev_cycle
version: 1.0.0
description: Write code and review it

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: write
    run: |
      Requirements: {{requirements}}
      Tests: {{tests}}
      Previous attempt: {{previous_code}}
      
      Write improved Python code. Output ONLY the code.
  
  - name: review
    needs: [write]
    run: |
      Tests: {{tests}}
      Code: {{write}}
      
      Review the code. Respond EXACTLY:
      - "PASS: All tests met" if good
      - "FAIL: [issue]" if problems exist
```

### Example 2: Refinement Loop

```yaml
loops:
  - name: refine_document
    workflow: improve_text
    with:
      original: "{{input}}"
      previous: "{{loop.last.output}}"
      feedback: "Make it more concise and professional"
    max_iterations: 3
    until: "Word count is under 500 and tone is professional"
    on_failure: continue
```

### Example 3: Search Until Found

```yaml
loops:
  - name: search_solution
    workflow: try_approach
    with:
      problem: "{{issue}}"
      attempts_so_far: "{{loop.history}}"
      current_attempt: "{{loop.iteration}}"
    max_iterations: 10
    until: "Solution found and verified"
    on_failure: continue
    accumulate: all_attempts
```

### Example 4: Quality Gate

```yaml
loops:
  - name: generate_until_quality
    workflow: content_generator
    with:
      topic: "{{subject}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "Quality score is above 8/10"
    on_failure: continue
```

---

## Best Practices

### 1. Set Realistic max_iterations

```yaml
# Good: Reasonable limits
max_iterations: 5   # For iterative development
max_iterations: 3   # For refinement
max_iterations: 10  # For exploration

# Avoid: Too high
max_iterations: 100  # Rarely needed, wastes resources
```

### 2. Write Clear Exit Conditions

```yaml
# ✅ Good: Specific and measurable
until: "The output says PASS"
until: "Error count is zero"
until: "Score exceeds 90%"

# ❌ Bad: Vague or complex
until: "Everything looks good"
until: "Quality is acceptable and no major issues exist"
```

### 3. Use Loop Variables Effectively

```yaml
# ✅ Good: Pass context between iterations
with:
  previous_attempt: "{{loop.last.output}}"
  iteration: "{{loop.iteration}}"
  feedback: "Improve based on previous attempt"

# ❌ Bad: No context
with:
  task: "Do the thing"
  # No way to improve!
```

### 4. Choose Right on_failure Strategy

```yaml
# Critical operations
on_failure: halt

# Exploratory/iterative tasks
on_failure: continue

# Network/transient errors
on_failure: retry
```

### 5. Accumulate Results When Needed

```yaml
# Store iteration history
accumulate: dev_history

# Access later
steps:
  - name: analyze
    needs: [my_loop]
    run: "Analyze evolution: {{dev_history}}"
```

### 6. Test Exit Conditions

Before deploying, test that your condition actually works:

```bash
# Run with --verbose to see condition evaluation
./mcp-cli --workflow my_workflow --input-data "test" --verbose

# Look for:
# [INFO] Condition evaluation: 'Your condition' -> YES/NO
```

---

## Accessing Loop Results

### In Subsequent Steps

```yaml
steps:
  - name: use_loop_result
    needs: [my_loop]
    run: |
      Loop finished in {{loop.iteration}} iterations
      Final result: {{my_loop}}
      Accumulated history: {{my_loop_accumulator}}
```

### Loop Result Variables

| Variable | Description |
|----------|-------------|
| `{{loop_name}}` | Final output from loop |
| `{{loop.iteration}}` | Total iterations completed |
| `{{accumulator_name}}` | Full iteration history (if `accumulate` specified) |

---

## Performance Considerations

### Typical Performance

- ~3-5 seconds per iteration (workflow + condition evaluation)
- 2 API calls per iteration (workflow + condition check)
- Exits immediately when condition met (no wasted iterations)

### Optimization Tips

1. **Keep workflows fast:** Simple workflows iterate faster
2. **Use specific conditions:** LLM evaluates faster
3. **Set reasonable max_iterations:** Don't over-allocate
4. **Cache when possible:** Reuse previous work

---

## Troubleshooting

### Loop Never Exits Early

**Problem:** Loop runs all max_iterations every time

**Solutions:**
1. Check condition phrasing - is it clear?
2. Verify workflow output format matches condition
3. Enable verbose logging to see evaluation:
   ```bash
   ./mcp-cli --workflow my_workflow --input-data "test" --verbose
   ```
4. Test condition manually

### Loop Exits Too Early

**Problem:** Loop exits on iteration 1 when it shouldn't

**Cause:** Condition too broad or LLM misinterpreting

**Solution:** Make condition more specific:
```yaml
# Too broad
until: "Output is good"

# Better
until: "Output contains EXACTLY the word COMPLETE"
```

### Errors Not Handled

**Problem:** Workflow fails and stops loop

**Check:** `on_failure` setting
```yaml
on_failure: continue  # Don't stop on errors
```

---

## Advanced Patterns

### Nested Iteration

Call a loop from within a loop by using workflow composition:

```yaml
# Parent workflow with loop
loops:
  - name: outer_loop
    workflow: inner_workflow  # This workflow has its own loop
    max_iterations: 3
    until: "All phases complete"

# inner_workflow.yaml also has a loop
loops:
  - name: inner_loop
    workflow: task
    max_iterations: 5
    until: "Phase complete"
```

### Progressive Refinement

```yaml
loops:
  - name: refine
    workflow: improver
    with:
      current: "{{loop.last.output}}"
      iteration: "{{loop.iteration}}"
      all_attempts: "{{loop.history}}"
    max_iterations: 5
    until: "Quality metrics all above threshold"
```

### Multi-Stage Pipeline

```yaml
steps:
  - name: stage1
    run: "Initial analysis"

loops:
  - name: develop
    workflow: dev_cycle
    max_iterations: 5
    until: "Code complete"

loops:
  - name: test
    workflow: test_cycle
    max_iterations: 3
    until: "Tests pass"

steps:
  - name: deploy
    needs: [develop, test]
    run: "Deploy: {{develop}}"
```

---

## See Also

- [Schema Reference](SCHEMA.md) - Complete schema documentation
- [Authoring Guide](AUTHORING_GUIDE.md) - How to write workflows
- [Examples](examples/) - Working examples with loops

---

**Last Updated:** January 7, 2026
