# Writing Effective Tool Descriptions

**The most important factor in Claude's ability to discover and use your tools correctly.**

---

## Why Tool Descriptions Are Critical

When you expose templates as MCP tools, **Claude only sees 3 things**:
1. Tool name
2. **Tool description** ← THIS GUIDE
3. Parameter schema

Since Claude doesn't see your template's internal workflow, **the description is your only opportunity** to guide how Claude uses your tool.

**Poor descriptions** = Claude picks wrong tools or uses them incorrectly  
**Great descriptions** = Claude picks right tools and uses them effectively

---

## The Proven Formula

### Structure That Works

```yaml
description: |
  [CAPABILITY NAME IN CAPS]
  
  [2-3 sentence summary: what it does and how it helps]
  
  ✓ [Key capability/feature 1]
  ✓ [Key capability/feature 2]
  ✓ [Output format or quality indicator]
  
  → Use when: [Specific trigger scenarios]
  
  → Best for: [Ideal use cases and constraints]
  
  → Time: [Execution time estimate]
  
  ⚠ [Limitations or when NOT to use]
```

### Why This Works

1. **CAPS TITLE** - Quickly scannable, sets expectations
2. **Summary** - Gives Claude context about purpose
3. **✓ Bullets** - Specific capabilities (not vague promises)
4. **→ Use when** - Explicit triggers help Claude match intent
5. **→ Best for** - Clarifies scope and constraints
6. **→ Time** - Sets expectations about speed
7. **⚠ Warning** - Prevents misuse, suggests alternatives

---

## Real Examples: Bad → Good → Great

### Example 1: Code Analysis

**❌ Bad (Vague):**
```yaml
description: "Analyzes code"
```

Problems: No specifics, no languages, no output format, no use cases

**⚠️ Okay (Informative but Generic):**
```yaml
description: "Analyzes code for bugs, security issues, and performance problems. Supports multiple programming languages."
```

Better but missing: Which languages? What specific issues? Output format? When to use?

**✅ Great (Rich, Actionable):**
```yaml
description: |
  COMPREHENSIVE CODE ANALYSIS
  
  Multi-language static analysis detecting bugs, security vulnerabilities,
  performance issues, and style violations with actionable recommendations.
  
  ✓ Detects: Logic errors, null checks, SQL injection, XSS, CSRF,
    race conditions, memory leaks, performance anti-patterns
  
  ✓ Supports: Python, JavaScript, TypeScript, Go, Java, C#, Ruby, PHP
  
  ✓ Returns: Severity-rated report (CRITICAL/HIGH/MEDIUM/LOW) with
    line numbers, root cause explanations, and fix suggestions
  
  → Use when: PR reviews, security audits, code quality checks,
     pre-commit validation, learning best practices, refactoring prep
  
  → Best for: Single files, functions, or classes (up to 1000 lines).
     For full repositories, use 'audit_repository' instead.
  
  → Time: 60-90 seconds (3-step analysis: detection, validation, reporting)
  
  ⚠ This is static analysis only. For runtime behavior analysis,
     use 'profile_execution' or 'trace_runtime' instead.
```

---

## Differentiating Similar Tools

When you have multiple similar tools, make distinctions crystal clear.

**Example: Quick vs. Deep Analysis**

**Quick Analysis:**
```yaml
description: |
  FAST SINGLE-STEP ANALYSIS
  
  Quick analytical insights without multi-step overhead.
  
  ✓ Output: Direct analysis in conversational format
  ✓ Speed: Single LLM call (~30 seconds)
  
  → Use when: Quick insights, rapid feedback, exploratory analysis,
     simple questions, initial assessment before deep dive
  
  → Best for: Quick checks, brainstorming, preliminary analysis
  
  ⚠ For comprehensive validated analysis, use 'deep_analysis'
```

**Deep Analysis:**
```yaml
description: |
  COMPREHENSIVE MULTI-STEP ANALYSIS WORKFLOW
  
  Four-stage analytical process with quality validation:
  1. Initial analysis, 2. Deep dive, 3. Quality check, 4. Final report
  
  ✓ Output: Validated comprehensive report with recommendations
  ✓ Quality: Built-in QA validation with retry logic
  
  → Use when: Business decisions, strategic planning, quality-critical
     evaluations requiring thorough validated analysis
  
  → Time: 2-4 minutes (4 LLM calls with validation)
  
  ⚠ For quick insights, use 'analyze' instead (30 seconds vs 2-4 minutes)
```

**Key Differences Highlighted:**
- Speed: 30s vs 2-4 min
- Process: Single-step vs 4-stage
- Output: Conversational vs Structured report
- Quality: None vs Built-in validation

---

## Element Breakdown

### 1. Capability Name (Title)

**Format:** ALL CAPS, descriptive, specific

**✅ Good:**
- `COMPREHENSIVE CODE ANALYSIS`
- `MULTI-SOURCE RESEARCH WITH VERIFICATION`
- `SECURITY VULNERABILITY SCANNER`

**❌ Bad:**
- `Analysis` (too vague)
- `Tool` (meaningless)
- `Helper` (no information)

---

### 2. Summary (2-3 sentences)

**What to include:**
- What the tool does
- How it does it (workflow overview)
- What value it provides

**✅ Example:**
```
Multi-language static analysis detecting bugs, security vulnerabilities,
performance issues, and style violations with actionable recommendations.
```

**❌ Not:**
```
This tool analyzes code and finds issues.
```

---

### 3. Capabilities (✓ Bullets)

**Be specific, not generic.**

**✅ Good:**
```
✓ Detects: SQL injection, XSS, CSRF, buffer overflows, race conditions
✓ Supports: Python, JavaScript, TypeScript, Go, Java, C#, Ruby
✓ Returns: JSON report with severity levels, line numbers, fixes
```

**❌ Bad:**
```
✓ Finds security issues
✓ Works with code
✓ Gives results
```

---

### 4. Triggers (→ Use when)

**Critical for Claude's tool selection.**

**✅ Good:**
```
→ Use when: PR reviews, security audits, pre-commit checks,
   code quality gates, learning best practices, refactoring prep
```

**❌ Bad:**
```
→ Use when: You want to analyze code
```

---

### 5. Time Estimate (→ Time)

**Set realistic expectations.**

**Examples:**
```
→ Time: ~5 seconds (single LLM call)
→ Time: 2-4 minutes (4 LLM calls with validation)
→ Time: 60-90 seconds (3-step analysis workflow)
→ Time: Quick (30s) / Standard (60s) / Deep (2-3 min) based on depth
```

---

### 6. Limitations (⚠)

**Prevent misuse and set boundaries.**

**✅ Good:**
```
⚠ This is static analysis only. For runtime behavior, use
   'profile_execution'. For full codebase analysis, use
   'audit_repository' instead.
```

**Examples:**
```
⚠ For simple definitions, use regular chat instead (faster)

⚠ Requires well-formed JSON input. For unstructured text,
   use 'analyze_text' instead.

⚠ Maximum 10,000 words. For longer documents, use 'process_large_document'
```

---

## Common Mistakes to Avoid

### ❌ Mistake 1: Too Generic

```yaml
description: "Analyzes data"
```

Problems: What kind of data? What analysis? What output?

### ❌ Mistake 2: Implementation Details

```yaml
description: |
  Uses GPT-4 to process input through a 3-step workflow calling
  template_analyzer which then calls entity_extractor...
```

Problems: Too technical, not user-focused, exposes internals Claude doesn't need

### ❌ Mistake 3: No Differentiation

```yaml
# Tool 1
description: "Analyzes code quality"

# Tool 2  
description: "Checks code quality"
```

Problems: Can't tell difference, Claude will pick randomly

---

## Testing Your Descriptions

### Test in Claude Desktop

```
You: I need to analyze this code for security issues

# Does Claude pick the right tool?
# If not, improve the description
```

### Edge Case Test

```
You: Quick code check

# Does Claude pick quick_scan, not comprehensive_review?
```

---

## Workflow for New Tools

```yaml
description: |
  [CAPABILITY IN CAPS - CLEAR AND SPECIFIC]
  
  [2-3 sentences explaining what it does, how it works, and value.
   Focus on user benefits, not implementation.]
  
  ✓ [Key capability 1 with specifics]
  ✓ [Key capability 2 with numbers/examples]
  ✓ [Output format and quality indicators]
  
  → Use when: [Scenario 1], [scenario 2], [scenario 3],
     [be generous here - list many triggers]
  
  → Best for: [Ideal data types, size constraints, use cases].
     [Include quantitative limits where relevant]
  
  → Time: [Execution time estimate with explanation]
  
  ⚠ [When NOT to use]. [Alternative tool]. [Important limitations]
```

---

## Checklist

Before finalizing any tool description:

- [ ] Title is specific and descriptive (not generic)
- [ ] Summary explains what, how, and why (2-3 sentences)
- [ ] Capabilities are specific with examples/numbers
- [ ] Output format is clearly specified
- [ ] Multiple use-case triggers are listed
- [ ] Ideal use cases and constraints are defined
- [ ] Time estimate is provided
- [ ] Limitations and alternatives are mentioned
- [ ] If similar tools exist, differences are clear
- [ ] No implementation details exposed
- [ ] Tested with Claude - it selects correctly

---

## Real Examples

See the enhanced descriptions in:
- `config/runas/research_agent.yaml`
- `config/runas/document_intelligence_agent.yaml`
- `config/runas/code_reviewer.yaml`

Each demonstrates effective description writing for different domains.

---

## Key Takeaway

**Your tool description is Claude's ONLY guide to using your tool correctly.**

The difference between:

```
"Analyzes code"
```

and the comprehensive example above is the difference between Claude rarely using your tool correctly and Claude consistently choosing and using it effectively.

**Write descriptions for humans first, Claude second. If a human would understand when and how to use the tool, so will Claude.**
