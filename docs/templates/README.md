# Templates Documentation

Build powerful, reusable AI workflows with templates.

---

## Quick Links

- **[Authoring Guide](authoring-guide.md)** - Complete template reference
- **[Examples](examples/)** - Working template examples
- **[Patterns](patterns/)** - Common design patterns

---

## What are Templates?

Templates are **YAML workflows** that define multi-step AI operations. Think of them as scripts for AI, but better:

```yaml
name: code_review
steps:
  - name: analyze
    prompt: "Review this code: {{input_data}}"
    output: analysis

  - name: report
    prompt: "Format as markdown: {{analysis}}"
```

**Benefits:**

- ✅ **Version Controlled** - Track in git
- ✅ **Reusable** - Use across projects
- ✅ **Composable** - Combine simple templates
- ✅ **Maintainable** - Update prompts without touching code
- ✅ **Testable** - Iterate and improve

---

## Getting Started

### 1. Create Your First Template

```bash
# Create template
cat > config/templates/hello.yaml << EOF
name: hello
version: 1.0.0

steps:
  - name: greet
    prompt: "Say hello to {{input_data.name}}"
EOF
```

### 2. Use It

```bash
mcp-cli --template hello --input-data '{"name": "Alice"}'
```

### 3. Build From Examples

```bash
# Download example template
curl -o config/templates/code-review.yaml \
  https://raw.githubusercontent.com/LaurieRhodes/mcp-cli-go/main/docs/templates/examples/code-review.yaml

# Use with piped input (template uses {{stdin}})
cat mycode.go | mcp-cli --template code_review
```

Note that AI code reviews need human oversight - they often lack architectural context and apply security patterns indiscriminately.

Depending on your latency, the multi-step templates can take time to complete.  Use the addition of the --verbose flag for troubleshooting if required. 

---

## Understanding Templates

### Template Structure

Every template has three main sections that work together:

```yaml
# 1. METADATA - Identifies the template
name: template_name              # What you type: --template template_name
description: What it does        # Human-readable explanation
version: 1.0.0                   # Track changes over time

# 2. CONFIGURATION - AI behavior defaults
config:
  defaults:
    provider: anthropic          # Which AI (anthropic, openai, ollama)
    model: claude-sonnet-4       # Which specific model
    temperature: 0.7             # Creativity (0.0 = focused, 1.0 = creative)
    max_tokens: 4000             # Maximum response length

# 3. WORKFLOW - Steps that execute in order
steps:
  # First step receives input
  - name: step1
    prompt: "Process this input: {{input_data}}"
    output: result1              # Save AI's response as "result1"

  # Second step uses first step's output
  - name: step2
    prompt: "Refine this: {{result1}}"  # Uses "result1" from above
    output: result2
```

**How the pieces work together:**

1. User runs: `mcp-cli --template template_name --input-data '...'` or pipes input
2. Input goes into `{{input_data}}` variable
3. Step 1 processes it, saves output as `result1`
4. Step 2 uses `result1`, saves output as `result2`
5. Final step's output is returned to user

---

## Core Concepts

### Variables Reference

Templates access data through variables. Here are all available variables:

| Variable Type         | Variable                  | Description                                                           | Example                        |
| --------------------- | ------------------------- | --------------------------------------------------------------------- | ------------------------------ |
| **Input Variables**   |                           |                                                                       |                                |
|                       | `{{input_data}}`          | Input from --input-data flag or pipe<br>*(recommended, clear naming)* | `{{input_data}}`               |
|                       | `{{stdin}}`               | Same as input_data<br>*(alternative name, both work identically)*     | `{{stdin}}`                    |
| **Template Metadata** |                           |                                                                       |                                |
|                       | `{{template.name}}`       | Current template name                                                 | `{{template.name}}`            |
|                       | `{{template.version}}`    | Template version number                                               | `{{template.version}}`         |
|                       | `{{execution.timestamp}}` | Execution start time (RFC3339 format)                                 | `{{execution.timestamp}}`      |
| **Step Outputs**      |                           |                                                                       |                                |
|                       | `{{step_name}}`           | Output from a named step                                              | `{{extract_data}}`             |
|                       | `{{custom_var}}`          | User-defined output variable                                          | `{{analysis_result}}`          |
| **Loop Variables**    |                           |                                                                       |                                |
|                       | `{{item}}`                | Current item in loop                                                  | `{{item}}`                     |
|                       | `{{index}}`               | Current iteration index (0-based)                                     | `{{index}}`                    |
|                       | `{{first}}`               | Boolean, true on first iteration                                      | `{% if first %}...{% endif %}` |
|                       | `{{last}}`                | Boolean, true on last iteration                                       | `{% if last %}...{% endif %}`  |
| **Nested Access**     |                           |                                                                       |                                |
|                       | `{{var.field}}`           | Access nested JSON/map fields                                         | `{{user.name}}`                |
|                       | `{{var.nested.deep}}`     | Multi-level nesting                                                   | `{{config.api.key}}`           |
|                       | `{{array[0]}}`            | Array element by index                                                | `{{items[0]}}`                 |
|                       | `{{array[0].field}}`      | Nested access in arrays                                               | `{{users[0].email}}`           |

**Key Facts:**

- `{{input_data}}` is the recommended variable name (clearer than `stdin`)
- Both `{{input_data}}` and `{{stdin}}` work identically - use either one
- Input comes from `--input-data` flag OR piped input (automatically merged)
- Access nested data with dot notation: `{{user.profile.email}}`
- Step outputs are referenced by step name or custom output variable name

---

### How Data Flows Between Steps

Each step can save its output and other steps can use it:

```yaml
steps:
  # Step 1: Process input
  - name: analyze_code
    prompt: |
      Analyze this code for bugs:
      {{input_data}}
    output: bug_report
    # After execution, "bug_report" contains the AI's analysis

  # Step 2: Use step 1's output
  - name: prioritize
    prompt: |
      Prioritize these bugs by severity:
      {{bug_report}}
    output: prioritized_bugs
    # Now "prioritized_bugs" contains the prioritized list

  # Step 3: Use both previous outputs
  - name: create_report
    prompt: |
      Create a markdown report:

      Original Analysis:
      {{bug_report}}

      Prioritized List:
      {{prioritized_bugs}}
    # Final output returned to user
```

**The flow:**

1. Input → `{{input_data}}` → used in `analyze_code`
2. `analyze_code` produces → `{{bug_report}}`
3. `prioritize` uses `{{bug_report}}` → produces `{{prioritized_bugs}}`
4. `create_report` uses both → final output

---

### Template Composition (Calling Other Templates)

Reuse existing templates as building blocks:

```yaml
name: comprehensive_review
steps:
  # Step 1: Call security template
  - name: security
    template: security_check           # Name of another template
    template_input: "{{input_data}}"   # Pass our input to it
    output: security_findings          # Store what it returns

  # Step 2: Call quality template  
  - name: quality
    template: quality_check
    template_input: "{{input_data}}"   # Same input, different template
    output: quality_findings

  # Step 3: Combine results
  - name: final_report
    prompt: |
      Combine these into one report:

      Security: {{security_findings}}
      Quality: {{quality_findings}}
```

**What happens:**

1. `security_check` template runs with our input
2. Its output is saved as `{{security_findings}}`
3. `quality_check` template runs with our input
4. Its output is saved as `{{quality_findings}}`
5. Final step combines both into a report

**Why use template composition:**

- Reuse existing work (don't duplicate)
- Keep templates focused and simple
- Build complex workflows from simple pieces

---

### Parallel Execution (Run Steps Simultaneously)

Speed up independent tasks by running them at the same time:

```yaml
steps:
  - name: all_checks
    parallel:
      # These three run at the same time
      - name: security
        prompt: "Security scan: {{input_data}}"

      - name: performance  
        prompt: "Performance check: {{input_data}}"

      - name: style
        prompt: "Style review: {{input_data}}"

    max_concurrent: 2      # Run 2 at a time (optional limit)
    aggregate: merge       # Combine results as text
    output: all_results    # Store combined output

  # Use the combined results
  - name: summary
    prompt: "Summarize: {{all_results}}"
```

**Aggregation options:**

- `merge` - Concatenate all outputs as text
- `array` - Store as JSON array `["result1", "result2", "result3"]`

**Performance:** If each check takes 10 seconds, parallel runs in ~10 seconds instead of 30.

---

### Loops (Process Multiple Items)

Run the same operation on each item in a list:

```yaml
steps:
  - name: analyze_files
    for_each: "{{file_list}}"     # Variable with array of items
    item_name: current_file       # Name for current item
    prompt: |
      Analyze file: {{current_file}}

      File {{index}} of {{total}}
      {% if first %}(First file){% endif %}
      {% if last %}(Last file){% endif %}
    output: all_analyses          # Array of all results
```

**Loop variables available:**

- `{{current_file}}` - The current item (name set by `item_name`)
- `{{index}}` - Position: 0, 1, 2, ...
- `{{first}}` - `true` only on first iteration
- `{{last}}` - `true` only on last iteration

**Example:** 

```bash
# Input: {"file_list": ["a.go", "b.go", "c.go"]}
# Result: Prompt runs 3 times with current_file = "a.go", "b.go", "c.go"
# Output: all_analyses = ["analysis of a.go", "analysis of b.go", "analysis of c.go"]
```

---

## Documentation

### For Beginners

1. **[Quick Start](../getting-started/README.md)** - Get up and running
2. **[Core Concepts](../getting-started/concepts.md)** - Understand templates
3. **[Basic Examples](examples/)** - Start with simple templates

### For Template Authors

1. **[Authoring Guide](authoring-guide.md)** - Complete reference
2. **[Patterns](patterns/)** - Design patterns
3. **[Advanced Examples](examples/)** - Complex workflows

### For Automation

1. **[Automation Guide](../guides/automation.md)** - Use templates in CI/CD
2. **[Query Mode](../guides/query-mode.md)** - Script with templates
3. **[Daily Report Example](examples/daily-report.yaml)** - Real automation

---

## Common Use Cases

### Code Review Workflow

**What it does:** Runs multiple review checks and combines them into a report.

```yaml
name: complete_code_review
steps:
  # Step 1: Security check (calls existing template)
  - name: security
    template: security_check          # Reuse security template
    template_input: "{{input_data}}"  # Pass code to it
    output: security_report

  # Step 2: Quality check (calls existing template)
  - name: quality
    template: quality_check
    template_input: "{{input_data}}"  # Same code, different analysis
    output: quality_report

  # Step 3: Combine into final report
  - name: final
    template: format_report
    template_input: |
      Security: {{security_report}}
      Quality: {{quality_report}}
```

**Usage:**

```bash
cat mycode.go | mcp-cli --template complete_code_review
```

**What happens:**

1. Your code goes to `security_check` template → returns security findings
2. Same code goes to `quality_check` template → returns quality findings  
3. Both reports are combined into formatted output

→ Full example: [code-review.yaml](examples/code-review.yaml)

---

### Document Processing Pipeline

**What it does:** Multi-step document transformation (extract → categorize → summarize).

```yaml
name: document_processor
steps:
  # Extract key information
  - name: extract
    prompt: |
      Extract key points from:
      {{input_data}}
    output: key_points

  # Categorize by topic
  - name: categorize
    prompt: |
      Group these points by topic:
      {{key_points}}
    output: categorized

  # Create summary
  - name: summarize
    prompt: |
      Summarize each category:
      {{categorized}}
```

**Usage:**

```bash
cat document.txt | mcp-cli --template document_processor
```

**Data flow:**

- Input document → `{{input_data}}` → extract step
- Extract produces → `{{key_points}}` → categorize step  
- Categorize produces → `{{categorized}}` → summarize step
- Final summary returned to user

→ Full example: [summarize.yaml](examples/summarize.yaml)

---

### Multi-Provider Analysis

**What it does:** Get opinions from different AI models and compare them.

```yaml
name: multi_ai_review
steps:
  # Claude's analysis
  - name: claude_analysis
    provider: anthropic
    model: claude-sonnet-4
    prompt: "Analyze this: {{input_data}}"
    output: claude_view

  # GPT-4's analysis
  - name: gpt_analysis
    provider: openai
    model: gpt-4o
    prompt: "Analyze this: {{input_data}}"
    output: gpt_view

  # Local model synthesis
  - name: synthesize
    provider: ollama
    model: llama3.2
    prompt: |
      Compare these analyses:

      Claude: {{claude_view}}
      GPT-4: {{gpt_view}}

      Where do they agree? Disagree?
```

**Why use multiple providers:**

- Different models have different strengths
- Consensus on important points increases confidence
- Catches biases or errors from single model

**Cost optimization:** Use expensive models (Claude, GPT-4) for analysis, cheap model (Ollama) for synthesis.

→ Full example: [multi-provider.yaml](examples/multi-provider.yaml)

---

### Parallel Processing

**What it does:** Run independent checks simultaneously for speed.

```yaml
name: parallel_review
steps:
  - name: all_checks
    parallel:
      # All three run at the same time
      - name: security
        prompt: "Security check: {{input_data}}"

      - name: performance
        prompt: "Performance check: {{input_data}}"

      - name: maintainability
        prompt: "Maintainability check: {{input_data}}"

    aggregate: merge    # Combine as text
    output: all_findings

  # Final step uses combined results
  - name: summary
    prompt: "Summarize: {{all_findings}}"
```

**Performance benefit:**

- **Sequential:** 10s + 10s + 10s = 30 seconds total
- **Parallel:** max(10s, 10s, 10s) = 10 seconds total

**When to use parallel:**

- Steps don't depend on each other
- Each step uses the same input
- Speed matters more than cost

→ Full example: [parallel-analysis.yaml](examples/parallel-analysis.yaml)

---

## Best Practices

### ✅ Do

- Use descriptive names
- Version your templates
- Start simple, add complexity gradually
- Test with sample data
- Compose from smaller templates
- Document your prompts
- Use appropriate providers for each step

### ❌ Don't

- Hardcode values (use variables)
- Create giant single-step templates
- Skip error handling
- Ignore token costs
- Forget to version
- Mix concerns in one template

---

## Example Library

### Beginner Templates

- [hello.yaml](examples/hello.yaml) - Simple greeting
- [summarize.yaml](examples/summarize.yaml) - Document summary
- [code-review.yaml](examples/code-review.yaml) - Basic review

### Intermediate Templates

- [daily-report.yaml](examples/daily-report.yaml) - Automation
- [parallel-analysis.yaml](examples/parallel-analysis.yaml) - Concurrency
- [loop-processing.yaml](examples/loop-processing.yaml) - Batch processing

### Advanced Templates

- [multi-provider.yaml](examples/multi-provider.yaml) - Multiple AIs
- [composed-workflow.yaml](examples/composed-workflow.yaml) - Composition
- [error-handling.yaml](examples/error-handling.yaml) - Robust workflows

---

## Template Patterns

Common workflow patterns you can reuse:

### Extract-Transform-Load (ETL)

**Pattern:** Process data through sequential stages.

```yaml
steps:
  - name: extract
    prompt: "Extract data from: {{input_data}}"
    output: raw_data

  - name: transform
    prompt: "Clean and structure: {{raw_data}}"
    output: clean_data

  - name: load
    prompt: "Format as JSON: {{clean_data}}"
```

**When to use:** Any data processing that needs cleaning or restructuring.

---

### Map-Reduce

**Pattern:** Process many items, then combine results.

```yaml
steps:
  # Map: Process each item
  - name: process_all
    for_each: "{{items}}"
    item_name: item
    prompt: "Process: {{item}}"
    output: processed_items

  # Reduce: Combine results
  - name: combine
    prompt: "Combine these: {{processed_items}}"
```

**When to use:** Batch processing where you need to analyze each item then summarize.

---

### Pipeline

**Pattern:** Refine output through multiple stages.

```yaml
steps:
  - name: draft
    prompt: "Create draft: {{input_data}}"
    output: draft

  - name: improve
    prompt: "Improve this: {{draft}}"
    output: improved

  - name: polish
    prompt: "Polish this: {{improved}}"
```

**When to use:** When quality improves with multiple refinement passes.

→ See [Patterns Documentation](patterns/) for more advanced patterns

---

## Quick Reference

```bash
# Create template
vim config/templates/my-template.yaml

# Use with input-data
mcp-cli --template my_template --input-data '{...}'

# Use with piped input
cat file.txt | mcp-cli --template my_template

# Use with file contents in input-data
mcp-cli --template my_template --input-data "{\"text\": \"$(cat file.txt)\"}"

# Debug
mcp-cli --verbose --template my_template
```

---

## Contributing

Found a useful template pattern? 

1. Test thoroughly
2. Add documentation
3. Submit PR to `docs/templates/examples/`
4. Share in [Discussions](https://github.com/LaurieRhodes/mcp-cli-go/discussions)

---

## Next Steps

**New to templates?**

1. Read [Core Concepts](../getting-started/concepts.md)
2. Try [hello.yaml](examples/hello.yaml)
3. Copy and customize an [example](examples/)

**Ready to build?**

1. Read [Authoring Guide](authoring-guide.md)
2. Study [Patterns](patterns/)
3. Create your first template

**Need automation?**

1. See [Automation Guide](../guides/automation.md)
2. Check [daily-report.yaml](examples/daily-report.yaml)
3. Build your CI/CD workflow

---

**Start building powerful AI workflows!** 🚀
