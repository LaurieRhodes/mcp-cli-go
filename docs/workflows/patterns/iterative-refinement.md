# Iterative Refinement Pattern

**Automatically improve output through LLM-evaluated feedback loops.**

---

## Overview

The Iterative Refinement pattern uses loops with LLM-evaluated exit conditions to automatically improve output until quality standards are met.

**Verified feature:** Based on working example in `config/workflows/iterative_dev/`

**Key innovation:** LLM evaluates exit conditions semantically - no exact string matching required.

---

## When to Use

**Use when:**

- Quality improves with feedback and iteration
- Exit criteria can be described naturally
- Willing to invest in quality (multiple iterations = higher cost)

**Examples:**

- Code development until tests pass
- Content writing until quality threshold met
- Data cleaning until no errors
- Translation until accuracy high

---

## Basic Structure

```yaml
$schema: "workflow/v2.0"
name: iterative_refinement
version: 1.0.0

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  - name: requirements
    run: "Define quality criteria for: {{input}}"

loops:
  - name: improve
    workflow: refine_step
    with:
      target: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "Output meets all quality criteria"
    on_failure: continue
```

**How it works:**

1. Define requirements/criteria
2. Loop starts iteration 1
3. Workflow executes and produces output
4. LLM evaluates: "Does output meet criteria?"
5. If no → iteration 2 with feedback
6. If yes → exit loop

---

## Complete Example: Code Development

Based on verified working example from `config/workflows/iterative_dev/`.

**Main workflow:**

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
  # Define requirements
  - name: requirements
    run: |
      Analyze this coding request:
      {{input}}

      Provide:
      1. What needs to be built
      2. Key requirements
      3. Success criteria

  # Design tests
  - name: test_criteria
    needs: [requirements]
    run: |
      Create test criteria for:
      {{requirements}}

      List specific tests that must pass.

loops:
  # Iteratively develop until passing
  - name: develop_until_pass
    workflow: dev_cycle
    with:
      requirements: "{{requirements}}"
      tests: "{{test_criteria}}"
      previous_code: "{{loop.last.output}}"
    max_iterations: 5
    until: "The review says PASS"
    on_failure: continue
    accumulate: development_history

steps:
  # Final report
  - name: final_report
    needs: [develop_until_pass]
    run: |
      Development completed in {{loop.iteration}} iterations.

      Final code: {{develop_until_pass}}

      Provide brief summary.
```

**Loop workflow (dev_cycle.yaml):**

```yaml
$schema: "workflow/v2.0"
name: dev_cycle
version: 1.0.0
description: Single development cycle

execution:
  provider: deepseek
  model: deepseek-chat

steps:
  # Write code
  - name: write
    run: |
      Write code for: {{requirements}}

      Tests to satisfy: {{tests}}
      Previous attempt: {{previous_code}}

      Improve based on feedback. Output ONLY the code.

  # Review code
  - name: review
    needs: [write]
    run: |
      Review code against tests:

      Tests: {{tests}}
      Code: {{write}}

      Respond EXACTLY:
      - "PASS: All tests met" if passes
      - "FAIL: [specific issues]" if problems exist
```

**Usage:**

```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Write a function that calculates fibonacci numbers"
```

**What happens:**

```
Iteration 1:
  Write: def fib(n): return n
  Review: FAIL: Doesn't calculate correctly

Iteration 2:
  Write: def fib(n): return fib(n-1) + fib(n-2)
  Review: FAIL: Missing base case, infinite recursion

Iteration 3:
  Write: def fib(n):
           if n <= 1: return n
           return fib(n-1) + fib(n-2)
  Review: PASS: All tests met

Loop exits after 3 iterations
```

---

## Loop Variables

Available in `with` parameters:

```yaml
loops:
  - name: improve
    with:
      iteration_number: "{{loop.iteration}}"
      last_output: "{{loop.last.output}}"
      full_history: "{{loop.history}}"
```

**Variables:**

- `{{loop.iteration}}` - Current iteration (1, 2, 3...)
- `{{loop.last.output}}` - Output from previous iteration
- `{{loop.output}}` - Current iteration output
- `{{loop.history}}` - All outputs separated by newlines and `---`

**Why this matters:** LLM learns from previous attempts and mistakes.

---

## Exit Conditions

The `until` field is evaluated by an LLM after each iteration.

### Good Exit Conditions

```yaml
# ✅ Specific and measurable
until: "All tests pass"
until: "Error count is zero"
until: "Quality score exceeds 8 out of 10"
until: "The review says PASS"
until: "No syntax errors detected"
```

### Poor Exit Conditions

```yaml
# ❌ Vague
until: "It's good"
until: "Done"
until: "Better than before"
```

**Why clarity matters:** LLM needs to clearly understand when to exit.

---

## Pattern Variations

### Variation 1: Content Refinement

```yaml
$schema: "workflow/v2.0"
name: refine_content
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

loops:
  - name: refine
    workflow: improve_content
    with:
      content: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 3
    until: "Quality score is 8 or higher"
    on_failure: continue
```

**Loop workflow:**

```yaml
$schema: "workflow/v2.0"
name: improve_content
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: improve
    run: "Improve this content: {{previous}}"

  - name: score
    needs: [improve]
    run: |
      Score this content: {{improve}}

      Rate 1-10 for:
      - Clarity
      - Engagement
      - Accuracy

      Provide overall score.
```

### Variation 2: Data Cleaning

```yaml
loops:
  - name: clean
    workflow: clean_and_validate
    with:
      data: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "No errors or issues detected"
```

### Variation 3: Translation Quality

```yaml
loops:
  - name: translate
    workflow: translate_and_review
    with:
      source: "{{input}}"
      previous: "{{loop.last.output}}"
    max_iterations: 3
    until: "Translation accuracy exceeds 95%"
```

---

## Best Practices

### 1. Set Realistic max_iterations

```yaml
# ✅ Good: Based on task complexity
max_iterations: 3   # Content refinement
max_iterations: 5   # Code development
max_iterations: 10  # Complex optimization

# ❌ Bad: Too high or too low
max_iterations: 100  # Wasteful
max_iterations: 1    # Not iterative
```

### 2. Provide Previous Context

```yaml
# ✅ Good: Learn from mistakes
with:
  previous: "{{loop.last.output}}"
  iteration: "{{loop.iteration}}"

# ❌ Bad: No context
with:
  task: "Do the thing"  # No learning
```

### 3. Use Clear Exit Conditions

```yaml
# ✅ Good: Specific criteria
until: "All 5 unit tests pass"
until: "Readability score > 8 AND no grammar errors"

# ❌ Bad: Ambiguous
until: "Looks good"
until: "Quality acceptable"
```

### 4. Use Cheaper Models

```yaml
# ✅ Good: Cheap model for iterations
execution:
  provider: deepseek  # $0.14/$1.10 per M tokens
  model: deepseek-chat

# Final validation with better model
steps:
  - name: final_check
    provider: anthropic
    model: claude-sonnet-4
```

### 5. Accumulate History

```yaml
# ✅ Good: Track improvements
loops:
  - name: improve
    accumulate: improvement_history

# Use later
steps:
  - name: analyze_progress
    needs: [improve]
    run: "Show improvement: {{improvement_history}}"
```

---

## Error Handling

Control behavior when workflow execution fails:

```yaml
loops:
  - name: develop
    on_failure: continue  # Keep trying despite errors
    # OR
    on_failure: halt      # Stop immediately
    # OR
    on_failure: retry     # Retry same iteration
```

**Recommendations:**

- Use `continue` for exploratory tasks
- Use `halt` for critical operations
- Use `retry` for transient network failures

---

## Common Use Cases

| Use Case         | Iterations | Exit Condition       | Provider  |
| ---------------- | ---------- | -------------------- | --------- |
| Code development | 3-5        | "All tests pass"     | deepseek  |
| Content writing  | 2-3        | "Quality score > 8"  | anthropic |
| Data cleaning    | 3-7        | "No errors detected" | deepseek  |
| Translation      | 2-3        | "Accuracy > 95%"     | anthropic |
| Bug fixing       | 2-4        | "Bug is fixed"       | deepseek  |

---

## Cost Analysis

**Per iteration cost:**

- DeepSeek: ~$0.001
- GPT-4: ~$0.01
- Claude: ~$0.015

**Example workflow:**

- 5 iterations × $0.001 = $0.005 (DeepSeek)
- 5 iterations × $0.01 = $0.05 (GPT-4)

**Cost optimization:**

1. Use cheap model for iterations
2. Use premium model for final validation
3. Set aggressive max_iterations
4. Cache expensive operations

---

## Performance

**Typical timing:**

- Per iteration: 3-5 seconds
- 5 iterations: 15-25 seconds
- Plus condition evaluation: ~1-2 seconds per iteration

**Total for 5 iterations:** ~20-30 seconds

---

## Troubleshooting

### Loop Never Exits Early

**Problem:** Always runs max_iterations

**Solutions:**

1. Make exit condition more specific:
   
   ```yaml
   # Too vague
   until: "Better than before"
   
   # More specific
   until: "The output explicitly says COMPLETE"
   ```

2. Check output format matches condition

3. Enable verbose logging:
   
   ```bash
   ./mcp-cli --workflow my_workflow --verbose
   ```

### Quality Doesn't Improve

**Problem:** Iterations don't help

**Solutions:**

1. Provide better feedback:
   
   ```yaml
   with:
     previous: "{{loop.last.output}}"
     issues: "Specific problems to fix"
   ```

2. Check if task benefits from iteration

3. Try different model

### Too Expensive

**Problem:** Costs add up

**Solutions:**

1. Use cheaper provider (deepseek, ollama)
2. Reduce max_iterations
3. Add quality check before loop
4. Consider if iteration necessary

---

## Complete Example: Quality-Controlled Development

```yaml
$schema: "workflow/v2.0"
name: quality_development
version: 1.0.0
description: Develop with quality gates

execution:
  provider: deepseek
  model: deepseek-chat
  temperature: 0.5
  servers: [filesystem]

steps:
  # Requirements
  - name: requirements
    run: "Analyze request: {{input}}"

  # Test criteria
  - name: tests
    needs: [requirements]
    run: "Create test criteria: {{requirements}}"

loops:
  # Iterative development
  - name: develop
    workflow: dev_cycle
    with:
      requirements: "{{requirements}}"
      tests: "{{tests}}"
      previous: "{{loop.last.output}}"
    max_iterations: 5
    until: "All tests pass"
    on_failure: continue
    accumulate: dev_history

steps:
  # Final quality check with consensus
  - name: final_validation
    needs: [develop]
    consensus:
      prompt: |
        Approve this code for production?

        Code: {{develop}}
        Tests: {{tests}}
        Development history: {{dev_history}}

        Answer APPROVED or REJECTED.
      executions:
        - provider: anthropic
          model: claude-sonnet-4
          temperature: 0
        - provider: openai
          model: gpt-4o
          temperature: 0
      require: unanimous

  # Generate report
  - name: report
    needs: [final_validation]
    run: |
      Development complete!

      Iterations: {{loop.iteration}}
      Final approval: {{final_validation}}

      Summary report.
```

---

## Related Patterns

- **[Consensus Validation](consensus-validation.md)** - Validate iteration results
- **[Document Pipeline](document-pipeline.md)** - Sequential processing

---

## See Also

- [Loop Guide](../LOOPS.md) - Complete loop documentation
- [Schema Reference](../SCHEMA.md) - Loop schema details
- Working example: `config/workflows/iterative_dev/`

---

**Automatic improvement through iteration!** 🔄
