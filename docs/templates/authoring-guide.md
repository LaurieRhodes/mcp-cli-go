# Template Authoring Guide

Create reusable, composable workflows with YAML templates.

---

## Table of Contents

- [What are Templates?](#what-are-templates)
- [Basic Template Structure](#basic-template-structure)
- [Steps and Execution](#steps-and-execution)
- [Variables and Data Flow](#variables-and-data-flow)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Complete Examples](#complete-examples)

---

## What are Templates?

Templates are **YAML workflows** that define multi-step AI operations.

**Why use templates?**

- ✅ **Reusable** - Define once, use everywhere
- ✅ **Version controlled** - Track changes in git
- ✅ **Composable** - Build complex workflows from simple parts
- ✅ **Maintainable** - Update prompts without touching scripts
- ✅ **Shareable** - Share workflows with team

**Example use cases:**

- Code review workflows
- Document analysis pipelines
- Multi-step research
- Data processing chains
- Report generation

---

## Basic Template Structure

### Minimal Template

The simplest template has a name and one step:

```yaml
# config/templates/hello.yaml
name: hello
description: A simple greeting template
version: 1.0.0

steps:
  - name: greet
    prompt: "Say hello to {{input_data.name}}"
```

**Use it:**

```bash
mcp-cli --template hello --input-data '{"name": "Alice"}'
```

**What happens:**

1. Template receives input: `{"name": "Alice"}`
2. Variable `{{input_data.name}}` becomes "Alice"
3. AI gets prompt: "Say hello to Alice"
4. Response returned to user

### Complete Template Structure

A production template includes metadata, configuration, and workflow steps:

```yaml
# config/templates/example.yaml
name: example_template
description: Shows all main components
version: 1.0.0
author: Your Name
tags: [analysis, review]

# Default AI behavior
config:
  defaults:
    provider: anthropic        # Which AI service
    model: claude-sonnet-4     # Which model
    temperature: 0.7           # Creativity level (0.0-1.0)
    max_tokens: 4000           # Max response length

# Workflow steps (run in order)
steps:
  - name: step1
    prompt: "Process: {{input_data}}"
    output: result1            # Store response as "result1"

  - name: step2
    prompt: "Refine: {{result1}}"  # Use step1's output
    output: final_result
```

**How it works:**

1. Input → `{{input_data}}` variable
2. Step 1 processes → saves as `{{result1}}`
3. Step 2 uses `{{result1}}` → saves as `{{final_result}}`
4. Final result returned to user

### Template Metadata

```yaml
name: my_template           # Required: Template identifier
description: What it does   # Recommended: Purpose description
version: 1.0.0             # Recommended: Semantic versioning
author: Your Name          # Optional: Attribution
tags: [analysis, review]   # Optional: Categories
```

---

## Steps and Execution

### Basic Step

```yaml
steps:
  - name: analyze
    prompt: "Analyze this code: {{code}}"
    output: analysis
```

**Components:**

- `name` - Step identifier (required)
- `prompt` - AI instruction (required)
- `output` - Variable name for result (optional)

### Step Configuration

```yaml
steps:
  - name: detailed_analysis
    prompt: "Analyze: {{input}}"
    output: analysis

    # Provider override
    provider: openai
    model: gpt-4o

    # Behavior settings
    temperature: 0.3      # 0.0-1.0 (lower = more focused)
    max_tokens: 2000      # Response limit
    timeout: 30           # Seconds

    # System prompt
    system_prompt: "You are a senior code reviewer"
```

### System Prompts

```yaml
steps:
  - name: review
    system_prompt: |
      You are a security expert.
      Focus on finding vulnerabilities.
      Be concise but thorough.
    prompt: "Review: {{code}}"
```

### Multiple Steps

```yaml
steps:
  # Step 1: Extract
  - name: extract_info
    prompt: "Extract key points from: {{document}}"
    output: key_points

  # Step 2: Summarize  
  - name: summarize
    prompt: "Summarize these points: {{key_points}}"
    output: summary

  # Step 3: Format
  - name: format_report
    prompt: |
      Create a markdown report:
      {{summary}}
    output: final_report
```

**Execution:** Steps run sequentially by default.

---

## Variables and Data Flow

### Built-in Variables

Templates have access to several built-in variables:

| Variable                  | Description                                                           | When to Use                      |
| ------------------------- | --------------------------------------------------------------------- | -------------------------------- |
| `{{input_data}}`          | Input from --input-data flag or pipe<br>*(recommended, clear naming)* | Primary way to access user input |
| `{{stdin}}`               | Same as input_data<br>*(alternative name)*                            | If you prefer Unix terminology   |
| `{{template.name}}`       | Current template name                                                 | For logging or debugging         |
| `{{template.version}}`    | Template version                                                      | For tracking changes             |
| `{{execution.timestamp}}` | When execution started (RFC3339)                                      | For timestamping results         |

**Example usage:**

```yaml
steps:
  - name: info
    prompt: |
      Analyzing with template: {{template.name}} v{{template.version}}
      Started: {{execution.timestamp}}

      Input to process:
      {{input_data}}
```

**Important:** Variables from `--input-data` must be accessed via `{{input_data}}`:

```yaml
# ✅ Correct - explicitly access via input_data
- prompt: "Hello {{input_data.name}}"

# ❌ Wrong - will fail with "variable not found"
- prompt: "Hello {{name}}"
```

**Why this matters:**

- `--input-data '{"name": "Alice"}'` creates `input_data.name`, not just `name`
- This prevents variable name collisions
- Makes data source explicit

### Step Output Variables

Each step can save its output for use by later steps:

```yaml
steps:
  # Step 1: Analyze code
  - name: analyze_code
    prompt: "Find bugs in: {{input_data.code}}"
    output: bug_list
    # After execution, {{bug_list}} contains the AI's response

  # Step 2: Use step 1's output
  - name: prioritize_bugs
    prompt: "Prioritize by severity: {{bug_list}}"
    output: prioritized
    # {{prioritized}} now contains the prioritized list

  # Step 3: Use multiple previous outputs
  - name: create_report
    prompt: |
      Create report:

      All Bugs: {{bug_list}}
      Priority Bugs: {{prioritized}}
```

**The data flow:**

1. Input → `{{input_data.code}}` → analyze_code
2. analyze_code → `{{bug_list}}` → prioritize_bugs
3. prioritize_bugs → `{{prioritized}}` → create_report
4. create_report → final output to user

**Key rule:** Use `{{step_name}}` or `{{output_name}}` to reference any step's result.

### Template-Level Variables

Define constants that are available to all steps:

```yaml
# Define reusable values
config:
  variables:
    max_items: 10
    output_format: "markdown"
    analysis_style: "concise"
    company_name: "Acme Corp"

steps:
  - name: generate_report
    prompt: |
      Create {{output_format}} report for {{company_name}}.
      Style: {{analysis_style}}
      Include up to {{max_items}} items.

      Data: {{input_data}}
```

**Why use template variables:**

- Avoid repeating values across steps
- Easy to update (one place to change)
- Self-documenting configuration
- Can be overridden if needed

---

### Accessing Input Data

Input comes from `--input-data` flag or piped input. Access it via `{{input_data}}`:

```yaml
# template.yaml
steps:
  - name: greet_user
    prompt: |
      Hello {{input_data.user.name}}!
      Role: {{input_data.user.role}}
      Department: {{input_data.user.department}}
```

```bash
# Use it
mcp-cli --template greet_user --input-data '{
  "user": {
    "name": "Alice",
    "role": "Engineer",
    "department": "Backend"
  }
}'
```

**Nested data access:**

- `{{input_data.user.name}}` - Access nested fields
- `{{input_data.config.api.key}}` - Multiple levels deep
- `{{input_data.items[0]}}` - Array element access

**Common mistake:**

```yaml
# ❌ Wrong - variable not found
prompt: "Hello {{name}}"

# ❌ Wrong - missing input_data prefix
prompt: "Hello {{user.name}}"

# ✅ Correct
prompt: "Hello {{input_data.user.name}}"
```

**Why the `input_data` prefix is required:**

- Prevents collisions with step outputs
- Makes data source explicit
- Matches how the system stores user input

---

## Advanced Features

### 1. Template Composition

Call other templates as reusable building blocks.

**Why use template composition:**

- Reuse existing templates (DRY principle)
- Build complex workflows from simple parts
- Each template has isolated variable scope
- More maintainable and testable

**Example:**

```yaml
# config/templates/security-check.yaml
name: security_check
steps:
  - name: scan
    prompt: "Security scan: {{input_data}}"
    output: findings
```

```yaml
# config/templates/full-review.yaml
name: full_review
steps:
  # Step 1: Call security template
  - name: security
    template: security_check              # Template to call
    template_input: "{{input_data.code}}" # Pass our code to it
    output: security_report               # Store its result

  # Step 2: Call quality template
  - name: quality
    template: quality_check
    template_input: "{{input_data.code}}" # Same input, different template
    output: quality_report

  # Step 3: Combine results from both templates
  - name: final_report
    prompt: |
      Combine these reviews:

      Security Findings: {{security_report}}
      Quality Issues: {{quality_report}}

      Create unified markdown report.
```

**What happens:**

1. User runs `full_review` with code input
2. `security_check` template runs with that code → returns findings
3. `quality_check` template runs with same code → returns issues
4. Final step combines both → unified report

**Context isolation:** Each sub-template has its own variable space - they can't accidentally interfere with each other.

### 2. Parallel Execution

Run multiple independent steps simultaneously for faster execution.

**When to use parallel execution:**

- Steps don't depend on each other's outputs
- Each step uses the same input
- Speed matters (can save significant time)
- Processing independent analyses

**Example:**

```yaml
steps:
  - name: comprehensive_analysis
    parallel:
      # All three run at the same time
      - name: security
        prompt: "Security analysis: {{input_data.code}}"
        output: security_findings

      - name: performance
        prompt: "Performance analysis: {{input_data.code}}"
        output: performance_findings

      - name: style
        prompt: "Style analysis: {{input_data.code}}"
        output: style_findings

    # Optional: limit how many run at once
    max_concurrent: 2

    # How to combine results
    aggregate: merge

    # Store combined output
    output: all_findings

  # Next step uses combined results
  - name: create_summary
    prompt: "Summarize all findings: {{all_findings}}"
```

**What happens:**

1. All three analyses start at the same time
2. With `max_concurrent: 2`, only 2 run simultaneously (3rd waits)
3. Results are combined based on `aggregate` setting
4. Combined output stored as `{{all_findings}}`
5. Next step can use the combined results

**Aggregate modes:**

- `merge` - Merge into single object/text:
  
  ```
  Security: [findings]
  Performance: [findings]
  Style: [findings]
  ```
- `array` - Store as JSON array:
  
  ```json
  [
    "security findings...",
    "performance findings...",
    "style findings..."
  ]
  ```

**Performance benefit:**

- **Sequential:** 10s + 10s + 10s = **30 seconds total**
- **Parallel:** max(10s, 10s, 10s) = **10 seconds total**

**Cost consideration:** Parallel execution runs all steps even if one could provide the answer alone. Balance speed vs. API costs.

### 3. Loops (Processing Arrays)

Process each item in an array with the same logic.

**When to use loops:**

- Analyzing multiple files
- Processing list of items
- Batch operations
- Generating reports for each entry

**Basic loop example:**

```yaml
steps:
  - name: analyze_files
    for_each: "{{input_data.files}}"  # Variable containing array
    item_name: current_file           # Name for current item
    prompt: "Analyze file: {{current_file}}"
    output: all_analyses              # Array of all results
```

```bash
# Input
mcp-cli --template analyze_files --input-data '{
  "files": ["main.go", "utils.go", "config.go"]
}'
```

**What happens:**

1. Loop runs 3 times (once per file)
2. First iteration: `current_file` = "main.go"
3. Second iteration: `current_file` = "utils.go"  
4. Third iteration: `current_file` = "config.go"
5. All results stored in `all_analyses` array

**Loop variables available:**

- `{{current_file}}` - Current item (name set by `item_name`)
- `{{index}}` - Position: 0, 1, 2, ... (zero-based)
- `{{first}}` - `true` only on first iteration
- `{{last}}` - `true` only on last iteration
- `{{total}}` - Total number of items

**Example using loop variables:**

```yaml
steps:
  - name: process_with_context
    for_each: "{{input_data.items}}"
    item_name: data
    prompt: |
      Processing item {{index}} of {{total}}:
      {{data}}

      {% if first %}
      [This is the first item - include header]
      {% endif %}

      {% if last %}
      [This is the last item - include summary]
      {% endif %}
```

**Output:** `all_analyses` contains an array like:

```json
[
  "analysis of main.go...",
  "analysis of utils.go...",
  "analysis of config.go..."
]
```

**Using loop results:**

```yaml
steps:
  # Loop through files
  - name: analyze_each
    for_each: "{{input_data.files}}"
    item_name: file
    prompt: "Analyze: {{file}}"
    output: individual_analyses

  # Summarize all analyses
  - name: summarize_all
    prompt: |
      Summarize these analyses:
      {{individual_analyses}}

      Provide overview of common issues.
```

### 4. Conditional Execution

Skip steps based on conditions - useful for branching logic.

**When to use conditions:**

- Different paths based on analysis results
- Skip expensive steps when not needed
- Error handling (only process valid data)
- A/B testing different approaches

**Basic conditional:**

```yaml
steps:
  # Step 1: Analyze code quality
  - name: check_quality
    prompt: "Rate code quality (Good/Bad): {{input_data.code}}"
    output: quality_rating

  # Step 2: Only runs if quality is bad
  - name: detailed_review
    condition: "{{quality_rating}} contains 'Bad'"
    prompt: "Detailed analysis of issues: {{input_data.code}}"
    output: detailed_findings

  # Step 3: Only runs if quality is good
  - name: approval
    condition: "{{quality_rating}} contains 'Good'"
    prompt: "Generate approval message for: {{input_data.code}}"
```

**What happens:**

1. `check_quality` always runs → returns "Good" or "Bad"
2. If "Bad": `detailed_review` runs, `approval` skips
3. If "Good": `detailed_review` skips, `approval` runs

**Condition syntax:**

```yaml
# Contains text
condition: "{{variable}} contains 'text'"

# Does not contain
condition: "{{variable}} not contains 'text'"

# Boolean values
condition: "{{is_valid}}"                    # True if truthy
condition: "{{is_valid}} == 'true'"          # Exact match

# Equality
condition: "{{status}} == 'complete'"        # Equals
condition: "{{status}} != 'failed'"          # Not equals
```

**Real-world example - error handling:**

```yaml
steps:
  # Validate input
  - name: validate
    prompt: "Is this valid JSON? Yes/No: {{input_data}}"
    output: is_valid

  # Only process if valid
  - name: process
    condition: "{{is_valid}} contains 'Yes'"
    prompt: "Process this JSON: {{input_data}}"
    output: processed

  # Only run if invalid
  - name: error_handling
    condition: "{{is_valid}} contains 'No'"
    prompt: "Explain what's wrong with: {{input_data}}"
    output: error_message
```

**Benefits:**

- Avoid wasting API calls on unnecessary steps
- Implement branching logic
- Fail fast when inputs are invalid
- Create adaptive workflows

### 5. Transform Operations

Transform data between steps.

```yaml
steps:
  - name: get_data
    prompt: "Extract all issues from: {{text}}"
    output: raw_issues

  - name: transform
    transform:
      - operation: filter
        condition: "severity == 'high'"
      - operation: limit
        count: 5
      - operation: pluck
        field: "description"
    input: raw_issues
    output: top_issues

  - name: report
    prompt: "Summarize these issues: {{top_issues}}"
```

**Available operations:**

- `filter` - Filter by condition
- `limit` - Limit number of items
- `pluck` - Extract specific field
- `group` - Group by key
- `map` - Extract fields (partial support)

### 6. Reusable Step Definitions

Define steps once, use multiple times.

```yaml
# Define reusable steps
step_definitions:
  analyze_code:
    prompt: "Analyze code: {{code}}"
    inputs:
      - name: code
        type: string
    outputs:
      - analysis

steps:
  # Use definition multiple times
  - name: analyze_main
    use: analyze_code
    inputs:
      code: "{{main_file}}"
    output: main_analysis

  - name: analyze_test
    use: analyze_code
    inputs:
      code: "{{test_file}}"
    output: test_analysis
```

### 7. Error Handling

Configure how templates handle failures with retry logic and fallbacks.

**When to use error handling:**

- External API calls that might fail
- Network-dependent operations
- Want graceful degradation
- Production workflows that must be resilient

**Basic error handling:**

```yaml
steps:
  - name: fetch_data
    prompt: "Fetch data from API"
    error_handling:
      on_failure: retry          # What to do: retry, stop, continue
      max_retries: 3             # Try up to 3 times
      retry_backoff: exponential # Wait longer each time
      initial_delay: "1s"        # Start with 1 second delay
      default_output: "API unavailable"  # Use if all retries fail
```

**What happens on failure:**

1. Step fails on first attempt
2. Waits 1 second (initial_delay)
3. Retries (attempt 2)
4. If fails, waits 2 seconds (exponential backoff)
5. Retries (attempt 3)
6. If fails, waits 4 seconds
7. Retries (attempt 4 - last try)
8. If still fails, uses "API unavailable" as output
9. Template continues to next step

**Error handling strategies:**

```yaml
# Strategy 1: Stop on error (default, safest)
error_handling:
  on_failure: stop    # Stop entire template execution

# Strategy 2: Continue regardless
error_handling:
  on_failure: continue           # Keep going
  default_output: "Step failed"  # Use this value

# Strategy 3: Retry with backoff
error_handling:
  on_failure: retry
  max_retries: 5
  retry_backoff: exponential  # or "linear"
  initial_delay: "2s"
```

**Real-world example:**

```yaml
steps:
  # Try external API with retry
  - name: external_lookup
    prompt: "Call external API for: {{input_data.query}}"
    error_handling:
      on_failure: retry
      max_retries: 3
      retry_backoff: exponential
      default_output: "External service unavailable"

  # Process results (or fallback)
  - name: process
    condition: "{{external_lookup}} not contains 'unavailable'"
    prompt: "Process: {{external_lookup}}"

  # Fallback path if external service failed
  - name: use_cache
    condition: "{{external_lookup}} contains 'unavailable'"
    prompt: "Use cached data for: {{input_data.query}}"
```

**Backoff strategies:**

- `linear`: 1s, 2s, 3s, 4s, 5s
- `exponential`: 1s, 2s, 4s, 8s, 16s (doubles each time)

### 8. Step Dependencies

Explicit execution order.

```yaml
steps:
  - name: fetch_data
    prompt: "Get data"
    output: data

  - name: validate_data
    prompt: "Validate: {{data}}"
    output: valid_data

  - name: process
    depends_on: [fetch_data, validate_data]
    prompt: "Process: {{valid_data}}"
```

### 9. MCP Server Integration

Use MCP (Model Context Protocol) tools in your templates - gives AI access to external tools like file systems, databases, APIs.

**What are MCP servers:**

- External tools AI can call
- Examples: filesystem, database, web search, git
- Configured in your config file
- AI decides when to use them based on prompt

**Basic usage:**

```yaml
steps:
  - name: read_and_analyze
    servers: [filesystem]        # Make filesystem tools available
    prompt: |
      Read the file config.yaml and analyze it.
      What provider is configured?
      What's the default model?
    output: analysis
```

**What happens:**

1. AI sees prompt asking to read file
2. AI has access to filesystem tools (because of `servers: [filesystem]`)
3. AI calls filesystem tool to read config.yaml
4. AI analyzes the content
5. AI responds with analysis

**Multiple servers:**

```yaml
steps:
  - name: comprehensive_check
    servers: [filesystem, git, web-search]
    prompt: |
      1. Read README.md
      2. Check git commit history
      3. Search web for latest best practices

      Compare project against best practices.
```

**Available servers** (configure in your config file):

- `filesystem` - Read/write files
- `git` - Git operations
- `database` - Database queries
- `brave-search` - Web search
- Custom servers you've configured

**When to use MCP:**

- Need to read local files
- Query databases
- Search the web
- Access external APIs
- Perform system operations

**Important:** The AI decides when to use tools based on the prompt. Write prompts that make it clear when tool use is needed.

### 10. Multi-Provider Workflows

Use different AI models for different steps - optimize for cost, speed, or capability.

**Why use multiple providers:**

- **Cost optimization:** Expensive models for analysis, cheap for formatting
- **Specialized strengths:** GPT-4 for code, Claude for writing
- **Speed vs quality:** Fast model for drafts, best model for final
- **Fallback options:** If one provider fails, try another

**Example - optimized for cost:**

```yaml
config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # Claude for deep analysis (its strength)
  - name: deep_analysis
    provider: anthropic
    model: claude-sonnet-4
    prompt: "Deeply analyze this document: {{input_data.document}}"
    output: analysis
    # Cost: High quality, higher price

  # GPT-4 for code generation (its strength)
  - name: generate_code
    provider: openai
    model: gpt-4o
    prompt: "Generate code based on: {{analysis}}"
    output: code
    # Cost: High quality, higher price

  # Local model for formatting (cheap!)
  - name: format_output
    provider: ollama
    model: qwen2.5:32b
    prompt: "Format this as markdown: {{code}}"
    # Cost: FREE (runs locally)
```

**What this achieves:**

- Use expensive models only where their quality matters
- Use free local model for simple formatting
- Significant cost savings on high-volume workflows

**Example - consensus building:**

```yaml
steps:
  # Get Claude's opinion
  - name: claude_analysis
    provider: anthropic
    model: claude-sonnet-4
    prompt: "Analyze: {{input_data.question}}"
    output: claude_view

  # Get GPT-4's opinion
  - name: gpt_analysis
    provider: openai
    model: gpt-4o
    prompt: "Analyze: {{input_data.question}}"
    output: gpt_view

  # Get DeepSeek's opinion (good at reasoning)
  - name: deepseek_analysis
    provider: deepseek
    prompt: "Analyze: {{input_data.question}}"
    output: deepseek_view

  # Synthesize with cheap local model
  - name: build_consensus
    provider: ollama
    model: qwen2.5:32b
    prompt: |
      Compare these three analyses:

      Claude says: {{claude_view}}
      GPT-4 says: {{gpt_view}}
      DeepSeek says: {{deepseek_view}}

      Where do they agree? Where do they disagree?
      What's the most likely correct answer?
```

**Cost strategy guide:**

| Provider       | Cost | Best For                               |
| -------------- | ---- | -------------------------------------- |
| Claude Opus    | $$$$ | Critical analysis, long documents      |
| Claude Sonnet  | $$$  | Balanced quality/cost                  |
| GPT-4          | $$$$ | Code generation, structured output     |
| GPT-4o-mini    | $$   | Fast responses, simple tasks           |
| DeepSeek       | $$   | Reasoning, math                        |
| Ollama (local) | FREE | Formatting, summarization, high volume |

**Example - speed optimization:**

```yaml
steps:
  # Fast draft with cheap model
  - name: quick_draft
    provider: openai
    model: gpt-4o-mini    # Fast and cheap
    temperature: 0.8
    prompt: "Quick draft: {{input_data.topic}}"
    output: draft

  # Polish with quality model
  - name: final_version
    provider: anthropic
    model: claude-sonnet-4  # High quality
    temperature: 0.3
    prompt: "Polish this draft: {{draft}}"
```

---

## Best Practices

### 1. Descriptive Names

```yaml
# Good
name: code_security_review
steps:
  - name: extract_security_issues
  - name: classify_severity
  - name: generate_recommendations

# Bad
name: template1
steps:
  - name: step1
  - name: step2
```

### 2. Clear Prompts

```yaml
# Good
prompt: |
  Review this Go code for security vulnerabilities.
  Focus on:
  - SQL injection
  - XSS attacks
  - Authentication issues

  Code:
  {{code}}

# Bad
prompt: "check this {{code}}"
```

### 3. Appropriate Output Names

```yaml
# Good
steps:
  - name: extract
    output: extracted_data
  - name: transform
    output: transformed_data

# Bad
steps:
  - name: extract
    output: output1
  - name: transform
    output: output2
```

### 4. Use Template Composition

```yaml
# Good: Reusable components
- template: security_check
- template: quality_check
- template: format_report

# Bad: Everything in one giant template
- name: do_everything
  prompt: "..." # 1000 lines
```

### 5. Fail Fast with Conditions

```yaml
# Good: Early exit
- name: validate
  prompt: "Is this valid? Yes or No: {{input}}"
  output: is_valid

- name: process
  condition: "{{is_valid}} contains 'Yes'"
  prompt: "Process: {{input}}"

# Bad: Process invalid data
- name: process
  prompt: "Process: {{input}}"
```

### 6. Meaningful Defaults

```yaml
# Good
config:
  defaults:
    temperature: 0.3    # Focused for code
    max_tokens: 2000    # Reasonable for reviews

# Bad
config:
  defaults:
    temperature: 1.0    # Too random for technical work
```

### 7. Version Your Templates

```yaml
name: code_review
version: 2.1.0  # Semantic versioning
description: Code review with security focus (updated 2024-12)
```

### 8. Document Complex Logic

```yaml
steps:
  # Extract all function definitions
  # This is needed for the dependency analysis in step 2
  - name: extract_functions
    prompt: "List all functions in: {{code}}"
    output: functions
```

---

## Complete Examples

### Example 1: Code Review Template

A practical template that performs security and quality checks:

```yaml
# config/templates/code-review.yaml
name: code_review
description: Comprehensive code review workflow
version: 1.0.0

config:
  defaults:
    provider: openai
    model: gpt-4o
    temperature: 0.3        # Low temperature for focused analysis

steps:
  # Step 1: Security analysis
  - name: security_check
    system_prompt: "You are a security expert. Focus on vulnerabilities."
    prompt: |
      Review this code for security issues:
      {{input_data}}

      Check for:
      - SQL injection
      - XSS attacks
      - Authentication issues
      - Input validation
    output: security_issues

  # Step 2: Quality analysis
  - name: quality_check
    system_prompt: "You are a code quality expert. Focus on best practices."
    prompt: |
      Review this code for quality issues:
      {{input_data}}

      Check for:
      - Code smells
      - Naming conventions
      - Error handling
      - Documentation
    output: quality_issues

  # Step 3: Generate report
  - name: create_report
    prompt: |
      Create a markdown code review report:

      **Security Findings:**
      {{security_issues}}

      **Quality Issues:**
      {{quality_issues}}

      Format with:
      # Code Review Report
      ## Security
      ## Quality  
      ## Recommendations
      ## Priority Actions
```

**Usage:**

```bash
cat mycode.go | mcp-cli --template code_review
```

**What it does:**

1. Runs security analysis on your code
2. Runs quality analysis on same code
3. Combines both into formatted markdown report

### Example 2: Document Analysis Pipeline

Process documents through multiple analysis stages:

```yaml
# config/templates/document-analysis.yaml
name: document_analysis
description: Multi-stage document processing
version: 1.0.0

steps:
  # Stage 1: Extract key points
  - name: extract_points
    prompt: |
      Extract key points from this document:
      {{input_data.document}}

      Focus on:
      - Main arguments
      - Important facts
      - Conclusions
    output: key_points

  # Stage 2: Classify by topic
  - name: classify_by_topic
    prompt: |
      Classify these points by topic:
      {{key_points}}

      Return as JSON:
      {
        "topic_name": ["point1", "point2"],
        "another_topic": ["point3", "point4"]
      }
    output: classified_points

  # Stage 3: Summarize each topic (using loop)
  - name: summarize_topics
    for_each: "{{classified_points}}"
    item_name: topic
    prompt: |
      Summarize this topic's points:
      Topic: {{topic.name}}
      Points: {{topic.points}}
    output: topic_summaries

  # Stage 4: Create executive summary
  - name: executive_summary
    prompt: |
      Create executive summary from:
      {{topic_summaries}}

      Format:
      # Document Analysis
      ## Executive Summary
      [3-5 sentences]

      ## Key Topics
      [Summary per topic]
```

**Usage:**

```bash
mcp-cli --template document_analysis --input-data '{
  "document": "Your long document text here..."
}'
```

**What it does:**

1. Extracts key points from document
2. Groups points by topic
3. Summarizes each topic
4. Creates executive summary of all topics

### Example 3: Multi-Provider Consensus

Get multiple AI opinions and build consensus:

```yaml
# config/templates/consensus-analysis.yaml
name: consensus_analysis
description: Get consensus from multiple AI models
version: 1.0.0

steps:
  # Claude's analysis
  - name: claude_perspective
    provider: anthropic
    model: claude-sonnet-4
    prompt: |
      Analyze this question thoroughly:
      {{input_data.question}}

      Provide your analysis and reasoning.
    output: claude_analysis

  # GPT-4's analysis
  - name: gpt_perspective
    provider: openai
    model: gpt-4o
    prompt: |
      Analyze this question thoroughly:
      {{input_data.question}}

      Provide your analysis and reasoning.
    output: gpt_analysis

  # DeepSeek's analysis (good at reasoning)
  - name: deepseek_perspective
    provider: deepseek
    prompt: |
      Analyze this question thoroughly:
      {{input_data.question}}

      Provide your analysis and reasoning.
    output: deepseek_analysis

  # Synthesis with local model (cost-effective)
  - name: build_consensus
    provider: ollama
    model: qwen2.5:32b
    prompt: |
      Compare these three expert analyses:

      **Claude's View:**
      {{claude_analysis}}

      **GPT-4's View:**
      {{gpt_analysis}}

      **DeepSeek's View:**
      {{deepseek_analysis}}

      Answer these questions:
      1. Where do all three agree? (high confidence)
      2. Where do they disagree? (needs investigation)
      3. What's the most likely correct answer?
      4. What's the confidence level?
```

**Usage:**

```bash
mcp-cli --template consensus_analysis --input-data '{
  "question": "What is the best approach for microservices architecture?"
}'
```

**Why this works:**

- Different models have different training and biases
- Agreement = high confidence
- Disagreement = areas needing more research
- Cheap local model can synthesize expensive analyses

### Example 4: Parallel Processing for Speed (*experimental*)

Run multiple analyses simultaneously:

```yaml
# config/templates/parallel-review.yaml
name: parallel_review
description: Fast comprehensive analysis using parallel execution
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4
    temperature: 0.3

steps:
  # Run all analyses at the same time
  - name: comprehensive_analysis
    parallel:
      # Security analysis
      - name: security
        system_prompt: "You are a security expert"
        prompt: |
          Security review:
          {{input_data}}

          Find vulnerabilities.
        output: security_findings

      # Performance analysis
      - name: performance
        system_prompt: "You are a performance expert"
        prompt: |
          Performance review:
          {{input_data}}

          Find bottlenecks.
        output: performance_findings

      # Maintainability analysis
      - name: maintainability
        system_prompt: "You are a code quality expert"
        prompt: |
          Maintainability review:
          {{input_data}}

          Find code smells.
        output: maintainability_findings

    # Run all 3 at once
    max_concurrent: 3

    # Combine as merged object
    aggregate: merge

    # Store combined results
    output: all_findings

  # Synthesize all findings
  - name: create_comprehensive_report
    prompt: |
      Create comprehensive review report from:
      {{all_findings}}

      Include:
      - Executive summary
      - Critical issues (from all analyses)
      - Recommended actions
      - Priority ranking
```

**Usage:**

```bash
cat application.go | mcp-cli --template parallel_review
```

**Performance:**

- **Sequential:** 30 seconds (10s + 10s + 10s)
- **Parallel:** 10 seconds (all run simultaneously)
- **Speedup:** 3x faster!

**When to use parallel:**

- Independent analyses (don't depend on each other)
- Speed is important
- All steps need same input
- Willing to pay for all API calls upfront

---

## Common Issues and Solutions

### "Variable not found" Error

**Problem:**

```yaml
prompt: "Hello {{name}}"
# Error: variable not found: name
```

**Why it happens:**
When you pass `--input-data '{"name": "Alice"}'`, the system creates `input_data.name`, not just `name`.

**Solution:**

```yaml
# ✅ Correct - access via input_data
prompt: "Hello {{input_data.name}}"
```

---

### Accessing Input Data

**How input works:**

```bash
# Method 1: Command line flag
mcp-cli --template my_template --input-data '{"key": "value"}'

# Method 2: Piped input
echo "some text" | mcp-cli --template my_template

# Both methods put data in the same place!
```

**In your template:**

```yaml
# ✅ Correct - works for both methods
prompt: "Process: {{input_data}}"

# ✅ Also correct - alternative name
prompt: "Process: {{stdin}}"

# For nested data from --input-data flag
prompt: "Hello {{input_data.user.name}}"
```

**Key point:** Whether you use `--input-data` or pipe input, both go into `{{input_data}}` (or `{{stdin}}` - they're the same).

---

### Missing Step Output

**Problem:**

```yaml
- name: step1
  prompt: "Process: {{input_data}}"
  # No output specified!

- name: step2
  prompt: "Use: {{step1}}"  # Error: step1 is empty
```

**Solution:**

```yaml
- name: step1
  prompt: "Process: {{input_data}}"
  output: step1_result  # ✅ Name the output

- name: step2
  prompt: "Use: {{step1_result}}"  # ✅ Reference the output name
```

---

### Condition Not Working

**Problem:**

```yaml
condition: "{{var}} = 'value'"  # Single = doesn't work
```

**Solution:**

```yaml
# ✅ Correct condition syntax
condition: "{{var}} == 'value'"     # Equality
condition: "{{var}} contains 'text'" # Contains
condition: "{{var}} not contains 'x'" # Not contains
```

---

### Loop Not Iterating

**Problem:**

```yaml
for_each: "{{items}}"
# Error: for_each value is not an array
```

**Why it happens:**
The variable doesn't contain an array.

**Solution:**

```bash
# ✅ Correct - pass an array
mcp-cli --template my_template --input-data '{
  "items": ["a", "b", "c"]
}'

# ❌ Wrong - not an array
mcp-cli --template my_template --input-data '{
  "items": "not an array"
}'
```

---

### Template Composition Input

**Problem:**

```yaml
- template: other_template
  # Missing: how to pass data to it
```

**Solution:**

```yaml
- template: other_template
  template_input: "{{input_data}}"  # ✅ Pass your input
  # or
  template_input: "{{previous_step}}"  # ✅ Pass step output
```

---

## Testing Templates

### Test with Sample Data

Always test templates before using in production:

```bash
# Test with simple input
echo '{"code": "def hello(): print(\"hi\")"}' | \
  mcp-cli --template code_review

# Test with file
cat test-data.json | mcp-cli --template my_template

# Test with input-data flag
mcp-cli --template my_template --input-data '{
  "key": "value",
  "items": ["a", "b", "c"]
}'
```

### Create Test Fixtures

Keep test data in files:

```bash
# Create test input file
cat > test-input.json << EOF
{
  "document": "This is a test document for analysis...",
  "max_length": 100,
  "format": "markdown"
}
EOF

# Run template with test data
mcp-cli --template document_analysis \
  --input-data "$(cat test-input.json)"
```

### Debug with Verbose Mode

See what's happening at each step:

```bash
# Enable verbose output
mcp-cli --verbose --template my_template --input-data '{...}'
```

**What verbose shows:**

- Each step as it executes
- Variables and their values
- API calls being made
- Errors with full context

### Test Individual Steps

Create minimal test templates:

```yaml
# test-step.yaml
name: test_step
version: 1.0.0
steps:
  - name: test
    prompt: "{{input_data}}"
```

```bash
# Test just one prompt
mcp-cli --template test_step --input-data "Test this prompt"
```

### Validate Template Syntax

Check if template is valid before running:

```bash
# Validate template syntax
mcp-cli --template my_template --validate

# Will show syntax errors if any
```

---

## Quick Reference

### Basic Template

```yaml
name: template_name
version: 1.0.0
steps:
  - name: step1
    prompt: "..."
    output: result
```

### With Variables

```yaml
variables:
  key: value

steps:
  - prompt: "Use {{key}}"
```

### With Composition

```yaml
steps:
  - template: other_template
    template_input: "{{data}}"
```

### With Conditions

```yaml
- condition: "{{var}} contains 'text'"
  prompt: "..."
```

### With Loops

```yaml
- for_each: "{{items}}"
  item_name: item
  prompt: "Process {{item}}"
```

---

## Next Steps

- **[Template Examples](examples/)** - Working templates
- **[Pattern Library](patterns/)** - Common patterns
- **[Automation Guide](../guides/automation.md)** - Use templates in CI/CD

---

**Ready to create?** Start with a simple template and build from there! 📝
