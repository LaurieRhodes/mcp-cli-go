# Skills Overview

## What Are Skills?

Skills are **not pre-written scripts**. They are **documentation + helper libraries** that enable Claude to write and execute custom code for your specific needs.

### The Key Insight

Traditional approach:
```
User request → Run pre-written script → Limited results
```

Skills approach:
```
User request → Claude reads docs → Claude writes custom code → Executes with helpers → Exactly what you need
```

## Architecture

```
┌─────────────────────────────────────────────┐
│ User Request                                │
│ "Create a report with Q4 sales data"       │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ Claude Reads Documentation                  │
│ • Loads docx/SKILL.md                       │
│ • Learns about Document class               │
│ • Sees example code                         │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ Claude Writes Custom Code                   │
│                                              │
│ from scripts.document import Document       │
│                                              │
│ doc = Document()                            │
│ doc.add_heading('Q4 Sales Report', 0)       │
│ # ... custom code for THIS report           │
│ doc.save('/workspace/report.docx')          │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ Sandbox Execution                           │
│ • Creates workspace                         │
│ • Mounts skill directory (read-only)        │
│ • Sets PYTHONPATH=/skill                    │
│ • Runs code in Docker/Podman                │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│ Results Returned                            │
│ • Output captured                           │
│ • Files available in workspace              │
│ • Claude presents to user                   │
└─────────────────────────────────────────────┘
```

## What Makes Skills Different?

### vs. Pre-written Scripts

| Aspect | Pre-written Scripts | Skills |
|--------|-------------------|--------|
| Flexibility | ❌ Fixed options | ✅ Infinite - custom code |
| Customization | ❌ One-size-fits-all | ✅ Tailored to each request |
| Maintenance | ❌ Update every script | ✅ Update helper libraries once |
| User Experience | ❌ Learn command options | ✅ Natural language requests |

### vs. Function Calling

| Aspect | Function Calling | Skills |
|--------|-----------------|--------|
| Capabilities | ❌ Limited to defined functions | ✅ Unlimited - write any code |
| Complexity | ❌ Simple operations only | ✅ Complex workflows possible |
| Composability | ❌ Hard to combine | ✅ Natural code composition |
| Output | ❌ Structured data only | ✅ Any output type |

## Skill Structure

A skill is a directory containing:

```
my-skill/
├── SKILL.md              # Documentation for Claude
├── scripts/              # Helper libraries
│   ├── __init__.py       # Makes it a Python package
│   └── helpers.py        # Your helper functions
└── references/           # Optional: additional docs
    └── examples.md
```

### SKILL.md - Documentation

Tells Claude what the skill can do:

```markdown
---
name: my-skill
description: What this skill does
---

# My Skill

This skill provides X, Y, and Z capabilities.

## Helper Functions

\`\`\`python
from scripts.helpers import my_function

result = my_function("input")
\`\`\`
```

### scripts/ - Helper Libraries

Reusable Python code Claude can import:

```python
# scripts/helpers.py
"""Helper functions for my-skill."""

def my_function(data):
    """Process data and return result."""
    return f"Processed: {data}"
```

### scripts/__init__.py - Package File

Makes the directory importable:

```python
"""My skill helper libraries."""
from .helpers import my_function
__all__ = ['my_function']
```

## How Execution Works

### 1. Workspace Setup

```
Container:
├── /workspace/           # Read-write - for code execution
│   ├── script.py         # Claude's generated code
│   ├── input.txt         # Optional input files
│   └── output.txt        # Generated outputs
│
└── /skill/               # Read-only - skill directory
    ├── SKILL.md
    └── scripts/
        ├── __init__.py
        └── helpers.py
```

### 2. PYTHONPATH Configuration

```bash
PYTHONPATH=/skill
```

This allows Claude's code to import:
```python
from scripts.helpers import my_function
```

### 3. Security Model

All executions are sandboxed:

- ✅ **Read-only skill directory** - Can't modify helper libraries
- ✅ **No network access** - Container runs with `--network=none`
- ✅ **Resource limits** - 256MB RAM, 0.5 CPU, 100 PIDs
- ✅ **Timeout** - 60 seconds maximum
- ✅ **Isolated filesystem** - No access to host files
- ✅ **Process isolation** - Each execution is independent

### 4. Performance

Typical execution breakdown:
- Container startup: ~250ms
- Code execution: ~30-50ms
- **Total: ~300ms**

Fast enough for interactive use!

## Real-World Example

### User Request
> "Create a professional report with our Q4 sales data, include charts and executive summary"

### Claude's Process

**Step 1: Read Documentation**
```javascript
const docs = await skills.docx({mode: "passive"})
```

Claude learns:
- Document class available
- Can add headings, paragraphs, tables
- Can import from scripts.document

**Step 2: Write Custom Code**
```python
from scripts.document import Document

doc = Document()

# Title page
doc.add_heading('Q4 2024 Sales Report', 0)
doc.add_heading('Executive Summary', 1)

# Claude writes custom logic for THIS specific report
doc.add_paragraph('Sales exceeded targets by 23%...')

# Add data table
doc.add_heading('Quarterly Breakdown', 1)
# ... custom table with user's data ...

doc.save('/workspace/q4-report.docx')
```

**Step 3: Execute**
```javascript
const result = await execute_skill_code({
    skill_name: "docx",
    code: customCodeFromStep2,
    files: {"sales-data.csv": userData}
})
```

**Step 4: Result**
Perfect report, customized exactly to user's needs!

## Why This Matters

### Infinite Flexibility

Pre-written scripts can't handle:
- "Make the title red and centered"
- "Add a watermark on every page"
- "Format numbers as currency with Euro symbol"
- "Include only data from Q3 and Q4"

Skills can handle ALL of these because Claude writes custom code.

### Reusable Libraries

Helper libraries are building blocks:
```python
# Helper libraries provide primitives
from scripts.document import Document
from scripts.formatting import apply_style
from scripts.tables import create_table

# Claude combines them in custom ways
doc = Document()
styled_para = apply_style(doc.add_paragraph('text'), 'Heading1')
table = create_table(data, columns=['Q1', 'Q2', 'Q3', 'Q4'])
```

### Natural Interface

Users don't need to learn:
- Command-line flags
- API endpoints
- Configuration syntax

They just describe what they want in natural language.

## Limitations

### Execution Environment

- **No network** - Can't fetch external data
- **No persistent state** - Each execution is isolated
- **60-second timeout** - Long operations may fail
- **Python only** - No Node.js support yet
- **Standard library** - External dependencies need custom images

### Security Trade-offs

- Code runs in sandbox (secure)
- But arbitrary code execution (by design)
- Trust the LLM to write safe code
- Review generated code if concerned

## When to Use Skills

### ✅ Good Use Cases

- Document creation (Word, PDF, PowerPoint)
- Data processing and transformation
- Text analysis and formatting
- Code generation
- Custom report generation
- File format conversions

### ❌ Not Ideal For

- Long-running processes (>60s)
- Network requests (sandbox blocks)
- Real-time data streaming
- Database operations (no network)
- Operations needing persistent state

## Next Steps

- **Get Started:** [Quick Start Guide](quick-start.md)
- **Use Skills:** [Using Existing Skills](using-skills.md)
- **Create Skills:** [Creating Skills Guide](creating-skills.md)
- **Technical Details:** [Reference](reference.md)
