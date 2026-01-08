# Skills Support in Workflow Schema

## Overview

The workflow schema now fully supports the `skills` property for filtering which Anthropic Skills are available to workflow steps.

This allows for a detailed skills library to be constructed with specific filtering targetted to workflow tasks to minimise context noise.

---

## What Changed

### Added `skills` Property

The `skills` property has been added to the workflow schema at the same level as `servers`, allowing you to filter which Anthropic Skills are available to your workflow.

**Property Details:**

- **Name:** `skills`
- **Type:** `string[]` (array of strings)
- **Default:** `[]` (empty = all available skills)
- **Inheritable:** ✅ Yes
- **CLI Argument:** `--skills <name>,<name>,...`

**Available Skills:**

- `docx` - Document creation/editing
- `pdf` - PDF manipulation
- `pptx` - Presentation creation
- `xlsx` - Spreadsheet editing
- `bash-preference` - Bash tool guidance

---

## Usage in Workflows

### Basic Example

```yaml
$schema: "workflow/v2.0"
name: document_workflow
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4
  servers: [filesystem]
  skills: [docx, pdf, xlsx]  # ← Filter available skills

steps:
  - name: create_report
    run: "Create a quarterly report as a Word document"

  - name: create_charts
    run: "Generate financial charts in Excel"
```

### Step-Level Override

```yaml
execution:
  provider: anthropic
  skills: [docx, pdf, xlsx]  # Default for all steps

steps:
  - name: word_doc
    skills: [docx]  # ← Override: only docx for this step
    run: "Create a Word document"

  - name: spreadsheet
    skills: [xlsx]  # ← Override: only xlsx for this step
    run: "Create an Excel spreadsheet"
```

### Without Skills Filter

```yaml
execution:
  provider: anthropic
  # No skills property = all available skills exposed

steps:
  - name: step1
    run: "Process document"
```

---

## Property Inheritance

The `skills` property follows the same inheritance pattern as `servers`:

```
workflow.execution
  ↓
  skills: [docx, pdf, xlsx]
  ↓
steps[0]
  ↓
  (inherits: [docx, pdf, xlsx])
  ↓
steps[1]  
  ↓
  skills: [docx]  ← Override
  ↓
  (uses: [docx])
```

---

## CLI Mapping

### Workflow YAML → CLI Execution

**YAML:**

```yaml
execution:
  provider: anthropic
  skills: [docx, xlsx]

steps:
  - name: create
    run: "Create document"
```

**Executes As:**

```bash
mcp-cli --provider anthropic \
  --skills docx,xlsx \
  --input-data "Create document"
```

---

## Documentation Files Updated

All schema documentation has been updated to include `skills`:

1. **OBJECT_MODEL.md**
   
   - Added `skills` to `ExecutionContext` interface
   - Updated examples to show skills usage
   - Added skills to inheritance examples

2. **QUICK_REFERENCE.md**
   
   - Added `skills` row to MCPQuery table
   - Added `skills` row to Step Object table  
   - Updated property inheritance diagram

3. **CLI_MAPPING.md**
   
   - Added `skills` to Core Properties table
   - Added `skills` to MCPQuery interface
   - Added Example 4: Anthropic Skills

4. **INHERITANCE_GUIDE.md** (if applicable)
   
   - Updated inheritance examples

5. **STEPS_REFERENCE.md** (if applicable)
   
   - Updated step examples

---

## Schema Definition

### TypeScript Interface

```typescript
interface ExecutionContext extends MCPQuery {
  provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
  servers?: string[];
  skills?: string[];      // ← Added
  timeout?: Duration;
  logging?: "normal" | "verbose" | "noisy";
  no_color?: boolean;
}
```

### YAML Schema

```yaml
execution:
  type: object
  properties:
    provider:
      type: string
    model:
      type: string
    temperature:
      type: number
    max_tokens:
      type: integer
    servers:
      type: array
      items:
        type: string
    skills:             # ← Added
      type: array
      items:
        type: string
      default: []
```

---

## Use Cases

### 1. Office Document Workflows

```yaml
skills: [docx, xlsx, pptx]
# Enable only office document skills
```

### 2. PDF Processing

```yaml
skills: [pdf]
# Only PDF manipulation
```

### 3. Mixed Document Types

```yaml
skills: [docx, pdf]
# Documents that might need conversion
```

### 4. Restricted Environment

```yaml
skills: [bash-preference]
# Only passive skills (no active processing)
```

**Date:** January 8, 2026

---

## Related Documentation

- [Skills Filtering](../../../SKILLS_FILTERING.md) - Using the --skills flag
- [Skills Display in Chat](../../../SKILLS_DISPLAY_IN_CHAT.md) - Skills in chat mode
- [CLI Mapping](CLI_MAPPING.md) - Complete CLI property mapping
- [Quick Reference](QUICK_REFERENCE.md) - One-page schema overview
