# Workflow Patterns

Common design patterns for building effective workflows using workflow v2.0 features.

---

## Available Patterns

### Core Patterns

- **[Iterative Refinement](iterative-refinement.md)** - Automatic improvement through LLM-evaluated loops
- **[Consensus Validation](consensus-validation.md)** - Multi-provider agreement for critical decisions  
- **[Document Pipeline](document-pipeline.md)** - Sequential document processing stages

---

## Pattern Selection Guide

| Pattern | Best For | Key Feature | Complexity |
|---------|----------|-------------|------------|
| **[Iterative Refinement](iterative-refinement.md)** | Code, content, data quality | Loops with LLM exit conditions | Medium |
| **[Consensus Validation](consensus-validation.md)** | Critical decisions | Multi-provider parallel execution | Medium |
| **[Document Pipeline](document-pipeline.md)** | Document processing | Sequential stages with dependencies | Low |

---

## Key Workflow v2.0 Features

### 1. LLM-Evaluated Loop Exit Conditions

**What it does:** LLM evaluates whether to continue or exit loop based on natural language criteria.

**Example:**
```yaml
loops:
  - name: improve
    workflow: refine
    until: "All tests pass"  # LLM evaluates this
    max_iterations: 5
```

**Use in patterns:**
- Iterative Refinement: Continue until quality met
- Any improvement process: Stop when criteria satisfied

### 2. Consensus Mode

**What it does:** Execute same prompt across multiple providers, evaluate agreement.

**Example:**
```yaml
steps:
  - name: validate
    consensus:
      prompt: "Approve deployment?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

**Use in patterns:**
- Consensus Validation: Critical decisions
- Quality gates: Production approval

### 3. Property Inheritance

**What it does:** Define defaults once, override where needed.

**Example:**
```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "..."  # Inherits all

  - name: step2
    run: "..."
    temperature: 0.3  # Override
```

**Use in patterns:**
- All patterns: Consistent configuration
- Document Pipeline: Same model for all stages

### 4. Step Dependencies

**What it does:** Control execution order with `needs`.

**Example:**
```yaml
steps:
  - name: extract
    run: "..."

  - name: analyze
    needs: [extract]  # Waits for extract
    run: "..."
```

**Use in patterns:**
- Document Pipeline: Sequential stages
- Any multi-step process: Ordered execution

### 5. Template Composition

**What it does:** Call workflows from workflows.

**Example:**
```yaml
steps:
  - name: process
    template:
      name: document_processor
      with:
        doc: "{{input}}"
```

**Use in patterns:**
- Document Pipeline: Reusable stages
- Any complex workflow: Break into modules

---

## When to Use Each Pattern

### Iterative Refinement

**Use when:**
- Output quality improves with feedback
- Can describe exit criteria naturally
- Multiple attempts beneficial

**Examples:**
- Code development until tests pass
- Content writing until quality threshold
- Data cleaning until no errors
- Bug fixing until resolved

**Key features used:**
- Loops with LLM-evaluated `until`
- Loop variables (`loop.last.output`)
- Provider selection (cheap for iterations)

→ [Full Pattern](iterative-refinement.md)

---

### Consensus Validation

**Use when:**
- High-stakes decisions required
- Single-model bias is concern
- Confidence through agreement needed
- Cost justified by risk

**Examples:**
- Security approval decisions
- Medical information validation
- Financial advice verification
- Legal document review

**Key features used:**
- Consensus mode
- Multiple provider execution
- Require levels (unanimous, 2/3, majority)

→ [Full Pattern](consensus-validation.md)

---

### Document Pipeline

**Use when:**
- Documents need systematic processing
- Multiple analysis stages required
- Reproducible workflow needed
- Sequential dependencies clear

**Examples:**
- Contract analysis (extract → analyze → report)
- Technical doc review (structure → completeness → accuracy)
- Financial report processing (extract → calculate → insights)

**Key features used:**
- Step dependencies (`needs`)
- Property inheritance
- Template composition
- Sequential execution

→ [Full Pattern](document-pipeline.md)

---

## Pattern Combinations

Patterns work well together.

### Refinement + Validation

Improve iteratively, then validate with consensus:

```yaml
loops:
  # Iteratively improve
  - name: develop
    workflow: code_cycle
    until: "All tests pass"
    max_iterations: 5

steps:
  # Validate with consensus
  - name: approve
    needs: [develop]
    consensus:
      prompt: "Approve for production?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
```

### Pipeline + Refinement

Process through pipeline, refine specific stages:

```yaml
steps:
  # Extract
  - name: extract
    run: "Extract from: {{input}}"

loops:
  # Iteratively improve transformation
  - name: transform
    workflow: improve_transform
    with:
      data: "{{extract}}"
    until: "Data quality is acceptable"
    max_iterations: 3

steps:
  # Final analysis
  - name: analyze
    needs: [transform]
    run: "Analyze: {{transform}}"
```

---

## Best Practices

### 1. Start Simple

Begin with basic structure, add complexity as needed:

```yaml
# Start here
steps:
  - name: process
    run: "{{input}}"

# Add dependencies
steps:
  - name: extract
    run: "..."
  - name: analyze
    needs: [extract]
    run: "..."

# Add loops if beneficial
loops:
  - name: improve
    workflow: refine
    until: "Quality met"
```

### 2. Use Clear Names

```yaml
# ✅ Good
steps:
  - name: extract_parties
  - name: analyze_obligations
  - name: assess_risks

# ❌ Bad
steps:
  - name: step1
  - name: step2
  - name: step3
```

### 3. Set Appropriate Limits

```yaml
# ✅ Good: Based on task
loops:
  - name: refine_content
    max_iterations: 3  # Content improves quickly

  - name: develop_code
    max_iterations: 5  # Code may need more attempts

# ❌ Bad: Too high
loops:
  - name: task
    max_iterations: 100  # Wasteful
```

### 4. Use Property Inheritance

```yaml
# ✅ Good: Set defaults once
execution:
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1
    run: "..."  # Inherits
  - name: step2
    run: "..."  # Inherits
```

### 5. Provide Clear Exit Conditions

```yaml
# ✅ Good: Specific
until: "All 5 unit tests pass"
until: "Error count is zero"
until: "Quality score exceeds 8/10"

# ❌ Bad: Vague
until: "Looks good"
until: "Done"
```

---

## Quick Reference

### Pattern Decision Tree

```
Need iterative improvement?
└─ Iterative Refinement

Need high-confidence validation?
└─ Consensus Validation

Need sequential document processing?
└─ Document Pipeline

Need combination?
└─ Use multiple patterns together
```

### Common Workflows

```yaml
# Iterative: improve until criteria met
loops:
  - workflow: improve
    until: "criteria met"
    max_iterations: 5

# Validation: multi-provider agreement
steps:
  - consensus:
      prompt: "approve?"
      require: unanimous

# Pipeline: sequential stages
steps:
  - name: stage1
  - name: stage2
    needs: [stage1]
  - name: stage3
    needs: [stage2]
```

---

## Verified Features

All patterns use verified, working features:

✅ **Loops with LLM-evaluated exit conditions**
- Code: `internal/services/workflow/loop_executor.go`
- Example: `config/workflows/iterative_dev/`

✅ **Consensus mode**
- Code: `internal/services/workflow/consensus.go`
- Struct: `ConsensusMode` in `workflow_v2.go`

✅ **Property inheritance**
- Code: `internal/services/workflow/orchestrator_v2.go`
- Execution → Step → Consensus

✅ **Step dependencies**
- Field: `needs` in `StepV2`
- Ensures execution order

✅ **Template composition**
- Field: `template` in `StepV2`
- Call workflows from workflows

---

## Next Steps

**New to patterns?**
1. Start with [Document Pipeline](document-pipeline.md) (simplest)
2. Try [Iterative Refinement](iterative-refinement.md) (powerful)
3. Use [Consensus Validation](consensus-validation.md) (when stakes are high)

**Building complex workflows?**
1. Combine patterns
2. Use property inheritance
3. Break into reusable templates

**Need help?**
1. See [Authoring Guide](../AUTHORING_GUIDE.md)
2. Read [Loop Guide](../LOOPS.md)
3. Check [Schema Reference](../SCHEMA.md)

---

**Build better workflows with proven patterns!** 🎯
