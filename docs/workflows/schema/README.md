# Workflow Schema Documentation

**Version:** workflow/v2.0  
**Status:** Production Ready

---

## Core Concept

> **Workflows are sequences of mcp-cli calls with shared configuration.**

Every workflow step is fundamentally an `mcp-cli` call. The workflow YAML provides:

1. **Configuration inheritance** - Define once, use everywhere
2. **Sequencing** - Control execution order with dependencies
3. **Composition** - Call workflows from workflows
4. **Validation** - Multi-provider consensus
5. **Iteration** - Loop until conditions met

---

## Documentation Structure

### 🚀 [Quick Reference](QUICK_REFERENCE.md)

**Start here.** One-page overview with:

- Minimal workflow example
- Property inheritance diagram
- Object schemas with complete property tables
- Common patterns
- CLI equivalents

**Best for:** Getting started, quick lookups, understanding the basics.

---

### 🔗 [CLI Property Mapping](CLI_MAPPING.md)

**Essential reading.** Complete mapping of YAML properties to CLI arguments:

- Base query object (what all steps inherit)
- Property inheritance chain
- Execution mode mapping
- Override rules and precedence

**Best for:** Understanding how workflows map to CLI calls, debugging property inheritance.

---

### 🏗️ [Object Model](OBJECT_MODEL.md)

**TypeScript-style interfaces.** Clean interface definitions showing:

- Base MCPQuery interface
- Workflow structure
- Step definitions with all modes
- Type hierarchy and validation rules
- Complete inheritance examples

**Best for:** Understanding the object-oriented design, seeing the complete type system.

---

### 📊 [Inheritance Guide](INHERITANCE_GUIDE.md)

**Visual diagrams.** Property flow through hierarchy:

- Three-level inheritance (workflow → step → consensus)
- Override precedence rules
- Provider failover inheritance
- MCP server inheritance
- Complete visual examples

**Best for:** Understanding property inheritance, troubleshooting configuration issues.

---

### ⚙️ [Steps Reference](STEPS_REFERENCE.md)

**Detailed step modes.** Complete reference for all execution modes:

- LLM Query (`run:`) with variable interpolation
- Template Call (`template:`) for workflow composition
- Embeddings (`embeddings:`) with full configuration
- Consensus (`consensus:`) basics
- Dependencies and conditions
- Common patterns

**Best for:** Learning step capabilities, building complex workflows.

---

### 📖 [Full Schema Reference](SCHEMA.md)

**Comprehensive documentation.** Complete schema with:

- All objects and properties
- Detailed examples
- Special modes
- Advanced features

**Best for:** Deep dives, learning advanced features, reference documentation.

---

## Quick Start

### 1. Understand the Foundation

Every workflow step inherits from the base `MCPQuery` object:

```typescript
interface MCPQuery {
  provider: string;         // --provider
  model: string;            // --model
  temperature?: number;     // --temperature
  max_tokens?: number;      // --max-tokens
  servers?: string[];       // --server (repeated)
  timeout?: string;         // (internal)
}
```

**This maps directly to `mcp-cli` arguments:**

```bash
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.7
```

---

### 2. Learn Property Inheritance

**Define once, use everywhere:**

```yaml
execution:              # ← Workflow defaults
  provider: anthropic
  model: claude-sonnet-4
  temperature: 0.7

steps:
  - name: step1        # ← Inherits all properties
    run: "Prompt"

  - name: step2        # ← Inherits, but overrides temperature
    temperature: 0.3
    run: "Prompt"
```

**This executes as:**

```bash
# Step 1
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.7 \
  --input-data "Prompt"

# Step 2
mcp-cli --provider anthropic --model claude-sonnet-4 --temperature 0.3 \
  --input-data "Prompt"
```

---

### 3. Master the Execution Modes

Each step is **one of these modes**:

| Mode          | Use For          | Example                                         |
| ------------- | ---------------- | ----------------------------------------------- |
| `run:`        | LLM query        | `run: "Analyze {{input}}"`                      |
| `template:`   | Call workflow    | `template: {name: code_review}`                 |
| `embeddings:` | Generate vectors | `embeddings: {texts: [...]}`                    |
| `consensus:`  | Multi-provider   | `consensus: {prompt: "...", executions: [...]}` |

---

### 4. Add Dependencies

Control execution order:

```yaml
steps:
  - name: step1
    run: "First"

  - name: step2
    needs: [step1]          # ← Waits for step1
    run: "Use {{step1}}"

  - name: step3
    needs: [step1, step2]   # ← Waits for both
    run: "Use {{step1}} and {{step2}}"
```

---

### 5. Add Failover

Automatic provider fallback:

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

If Anthropic fails → tries OpenAI → tries Ollama.

---

## Inheritance Flow

**Visual representation:**

```
┌────────────────────────────────┐
│ workflow.execution             │  ← Provides defaults
│   provider: anthropic          │
│   model: claude-sonnet-4       │
│   temperature: 0.7             │
└───────────────┬────────────────┘
                │ (all properties inherit)
                ↓
┌────────────────────────────────┐
│ steps[].{properties}           │  ← Can override
│   provider: ← inherited        │
│   model: ← inherited           │
│   temperature: 0.9 ← OVERRIDE  │
└───────────────┬────────────────┘
                │ (for consensus only)
                ↓
┌────────────────────────────────┐
│ consensus.executions[]         │  ← Can override per-execution
│   provider: openai ← OVERRIDE  │
│   model: ← inherited           │
│   temperature: ← inherited     │
└────────────────────────────────┘
```

---

## Common Patterns

### Pattern: Basic Workflow

```yaml
$schema: "workflow/v2.0"
name: analyzer
version: 1.0.0
description: Analyze and summarize

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: analyze
    run: "Analyze: {{input}}"

  - name: summarize
    needs: [analyze]
    run: "Summarize: {{analyze}}"
```

### Pattern: Different Temperatures

```yaml
execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: creative
    temperature: 1.5
    run: "Generate ideas"

  - name: precise
    temperature: 0.2
    run: "Analyze data"
```

### Pattern: Consensus Validation

```yaml
steps:
  - name: validate
    consensus:
      prompt: "Is this safe? {{input}}"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
        - provider: deepseek
          model: deepseek-chat
      require: 2/3
```

### Pattern: MCP Integration

```yaml
execution:
  provider: anthropic
  servers: [filesystem, brave-search]

steps:
  - name: search
    run: "Search for: {{query}}"

  - name: read
    run: "Read and analyze: {{filepath}}"
```

---

## Object Hierarchy

```
Workflow (root)
├── $schema: "workflow/v2.0"
├── name, version, description
├── execution: MCPQuery
│   ├── provider, model
│   ├── temperature, max_tokens
│   ├── servers, timeout
│   └── OR providers: [MCPQuery, ...]
├── env: {KEY: value}
├── steps: [Step, ...]
│   ├── name (required)
│   ├── needs: [step_names]
│   ├── if: "skip condition"
│   ├── Inherits: MCPQuery properties
│   └── Mode (choose one):
│       ├── run: "prompt"
│       ├── template: {name, with}
│       ├── embeddings: {...}
│       └── consensus: {prompt, executions, require}
└── loops: [Loop, ...]
    ├── name (required)
    ├── workflow: name
    ├── with: {inputs}
    ├── max_iterations: number
    └── until: "condition"
```

---

## Examples by Use Case

### DevOps

- **[Consensus Security Audit](../showcases/devops/workflows/consensus_security_audit.yaml)** - Multi-provider validation
- **[Resilient Health Monitor](../showcases/devops/workflows/resilient_health_monitor.yaml)** - Provider failover
- **[Incident Response](../showcases/devops/workflows/incident_response.yaml)** - Step dependencies

### Security

- **[SOAR Alert Enrichment](../showcases/security/workflows/soar_alert_enrichment.yaml)** - MCP integration
- **[Vulnerability Assessment](../showcases/security/workflows/vulnerability_assessment.yaml)** - Unanimous consensus
- **[Incident Playbook](../showcases/security/workflows/incident_playbook.yaml)** - Systematic response

### Development

- **[API Documentation](../showcases/development/workflows/api_documentation_generator.yaml)** - Multi-step analysis
- **[Code Review Assistant](../showcases/development/workflows/code_review_assistant.yaml)** - Consensus reduces false positives
- **[Database Optimizer](../showcases/development/workflows/database_query_optimizer.yaml)** - Systematic detection

### Data Engineering

- **[RAG Pipeline Builder](../showcases/data-engineering/workflows/rag_pipeline_builder.yaml)** - Validation before execution
- **[ML Data Quality](../showcases/data-engineering/workflows/ml_data_quality_validator.yaml)** - Consensus validation
- **[ETL Pipeline](../showcases/data-engineering/workflows/data_transformation_pipeline.yaml)** - Phase dependencies

### Business Intelligence

- **[Financial Reports](../showcases/business-intelligence/workflows/financial_report_generator.yaml)** - Systematic analysis
- **[Cohort Analysis](../showcases/business-intelligence/workflows/customer_cohort_analyzer.yaml)** - Multi-stage insights
- **[Metrics Dashboard](../showcases/business-intelligence/workflows/business_metrics_dashboard.yaml)** - Comprehensive metrics

---

## Learning Path

1. **Start:** Read [Quick Reference](QUICK_REFERENCE.md) (10 minutes)
2. **Understand:** Read [CLI Mapping](CLI_MAPPING.md) (20 minutes)
3. **Practice:** Try examples from [../examples/](../examples/)
4. **Deep Dive:** Read [Full Schema](SCHEMA.md) when needed
5. **Master:** Study showcase workflows for patterns

---

## Key Principles

### 1. Workflows Sequence mcp-cli Calls

Every workflow is just a series of `mcp-cli` invocations with shared config:

```yaml
# This workflow...
steps:
  - name: step1
    run: "Prompt"
  - name: step2
    run: "Prompt"

# ...is equivalent to:
# mcp-cli ... "Prompt"
# mcp-cli ... "Prompt"
```

### 2. Inheritance Eliminates Repetition

Define common configuration once:

```yaml
execution:
  provider: anthropic      # Used by all steps

steps:
  - name: step1            # Inherits provider
  - name: step2            # Inherits provider
  - name: step3            # Inherits provider
```

### 3. Override Only What Differs

```yaml
execution:
  temperature: 0.7         # Default

steps:
  - name: creative
    temperature: 1.5       # Override
  - name: precise
    temperature: 0.2       # Override
  - name: normal           # Uses default 0.7
```

### 4. Dependencies Control Order

```yaml
steps:
  - name: A
  - name: B
    needs: [A]             # Waits for A
  - name: C
    needs: [A, B]          # Waits for both
```

### 5. Everything is an Object

Workflows, steps, and consensus executions all inherit from the base `MCPQuery` object.

---

## Troubleshooting

### Issue: Property not inheriting

**Problem:**

```yaml
execution:
  provider: anthropic
steps:
  - name: step1
    run: "Prompt"         # Not using anthropic?
```

**Solution:** Check for typos in property names. Property inheritance is automatic for valid properties.

### Issue: Wrong CLI arguments

**Problem:** Not sure what CLI args a workflow uses?

**Solution:** Run with `--verbose`:

```bash
mcp-cli --workflow my_workflow --verbose
```

See actual `mcp-cli` commands being executed.

### Issue: Circular dependencies

**Problem:**

```yaml
steps:
  - name: A
    needs: [B]
  - name: B
    needs: [A]           # Circular!
```

**Solution:** Workflow will error. Fix dependency chain.

---

## Additional Resources

- **[Authoring Guide](../AUTHORING_GUIDE.md)** - How to write effective workflows
- **[Loop System](../LOOPS.md)** - Iterative execution deep dive
- **[Migration Guide](../MIGRATION.md)** - Upgrading from template v1
- **[Pattern Library](../patterns/)** - Common workflow patterns

---

## Need Help?

1. Check [Quick Reference](QUICK_REFERENCE.md) for common patterns
2. Review [CLI Mapping](CLI_MAPPING.md) for property issues
3. Study [Examples](../examples/) for working code
4. Read [Full Schema](SCHEMA.md) for detailed docs

---

**Remember:** If you understand `mcp-cli`, you understand workflows. Workflows just add sequencing, inheritance, and composition on top of the CLI.
