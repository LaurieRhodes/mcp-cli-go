# Workflow Organization Guide

**Organize your workflows in subdirectories with intelligent resolution.**

---

## Quick Start

### Create Subdirectories

```bash
mkdir -p config/workflows/{data_processing,deployment,monitoring}
```

### Move Workflows

```bash
mv config/workflows/cleaner.yaml config/workflows/data_processing/
mv config/workflows/deploy.yaml config/workflows/deployment/
```

### List All Workflows

```bash
mcp-cli --list-workflows
```

**Output:**

```json
{
  "workflows": [
    "analyzer",
    "data_processing/cleaner",
    "data_processing/validator",
    "deployment/deploy",
    "deployment/test"
  ]
}
```

### Run Nested Workflow

```bash
mcp-cli --workflow data_processing/cleaner --input-data "data"
```

---

## Directory-Aware Resolution

When workflows call other workflows, the system searches intelligently:

### Resolution Priority

1. ✅ **Exact Match** - If you specify `operations/deployer`, looks for exactly that
2. ✅ **Same Directory** - If calling from `iterative_dev/`, looks for `iterative_dev/workflow` first
3. ✅ **Root Fallback** - Falls back to root-level workflows

### Example: Same-Directory Resolution

```yaml
# File: config/workflows/iterative_dev/dev_cycle.yaml
steps:
  - name: plan
    template:
      name: planner          # ✅ Finds iterative_dev/planner

  - name: write
    template:
      name: code_writer      # ✅ Finds iterative_dev/code_writer

  - name: review
    template:
      name: code_reviewer    # ✅ Finds iterative_dev/code_reviewer
```

**Benefits:**

- ✅ Workflows in a project reference each other simply
- ✅ No need to specify full paths for local workflows
- ✅ Easy to move entire directories

---

## Cross-Directory References

### Explicit Paths

```yaml
# File: config/workflows/iterative_dev/dev_cycle.yaml
steps:
  # Local workflow (same directory)
  - name: review
    template:
      name: code_reviewer              # Finds iterative_dev/code_reviewer

  # Different directory (explicit path)
  - name: deploy
    template:
      name: deployment/deploy          # Finds deployment/deploy

  # Root-level workflow (no directory prefix)
  - name: notify
    template:
      name: simple_greeting            # Finds root simple_greeting
```

### Resolution Examples

**Directory Structure:**

```
config/workflows/
├── analyzer.yaml                      # Root: analyzer
├── data_processing/
│   ├── cleaner.yaml                   # data_processing/cleaner
│   └── validator.yaml                 # data_processing/validator
└── deployment/
    ├── test.yaml                      # deployment/test
    └── deploy.yaml                    # deployment/deploy
```

**From `data_processing/cleaner.yaml`:**

```yaml
steps:
  - name: validate
    template:
      name: validator                  # ✅ Finds data_processing/validator (same directory)

  - name: analyze
    template:
      name: analyzer                   # ✅ Finds root analyzer (fallback)

  - name: deploy
    template:
      name: deployment/deploy          # ✅ Finds deployment/deploy (explicit)
```

---

## Handling Name Conflicts

### The Problem

```
config/workflows/
├── code_reviewer.yaml                 # Root code_reviewer
└── iterative_dev/
    └── code_reviewer.yaml             # iterative_dev/code_reviewer
```

### Resolution Behavior

**From `iterative_dev/dev_cycle.yaml`:**

```yaml
steps:
  - name: review
    template:
      name: code_reviewer              # ✅ Finds iterative_dev/code_reviewer (same directory)
```

**From root workflow:**

```yaml
steps:
  - name: review
    template:
      name: code_reviewer              # ✅ Finds root code_reviewer (only option)
```

### Best Practices

**Option 1: Use unique names**

```
config/workflows/
├── basic_code_reviewer.yaml
└── iterative_dev/
    └── detailed_code_reviewer.yaml
```

**Option 2: Organize by purpose**

```
config/workflows/
├── quick_review/
│   └── code_reviewer.yaml             # quick_review/code_reviewer
└── thorough_review/
    └── code_reviewer.yaml             # thorough_review/code_reviewer
```

**Option 3: Move root to subdirectory**

```
config/workflows/
├── legacy/
│   └── code_reviewer.yaml             # legacy/code_reviewer
└── iterative_dev/
    └── code_reviewer.yaml             # iterative_dev/code_reviewer
```

---

## Organization Patterns

### Pattern 1: By Feature

```
config/workflows/
├── user_management/
│   ├── create_user.yaml
│   ├── update_user.yaml
│   └── delete_user.yaml
├── data_processing/
│   ├── import.yaml
│   ├── clean.yaml
│   └── export.yaml
└── reporting/
    ├── daily_report.yaml
    └── monthly_report.yaml
```

**Use case:** Microservices, feature-based projects

### Pattern 2: By Phase

```
config/workflows/
├── planning/
│   ├── requirements.yaml
│   └── design.yaml
├── development/
│   ├── code_writer.yaml
│   └── test_generator.yaml
├── review/
│   ├── code_reviewer.yaml
│   └── security_scanner.yaml
└── deployment/
    ├── deploy.yaml
    └── rollback.yaml
```

**Use case:** SDLC workflows, multi-phase processes

### Pattern 3: By Complexity

```
config/workflows/
├── simple/
│   ├── quick_check.yaml
│   └── format.yaml
├── standard/
│   ├── analyzer.yaml
│   └── reporter.yaml
└── advanced/
    ├── deep_analysis.yaml
    └── ai_optimization.yaml
```

**Use case:** Tools with varying complexity levels

### Pattern 4: By Team/Project

```
config/workflows/
├── team_a/
│   ├── project_x/
│   │   ├── workflow1.yaml
│   │   └── workflow2.yaml
│   └── project_y/
│       └── workflow1.yaml
└── team_b/
    └── project_z/
        └── workflow1.yaml
```

**Use case:** Multi-team organizations

---

## 

## Best Practices

### 1. ✅ Use README Files

Add a README.md in each directory:

```markdown
# Data Processing Workflows

Main workflow: `pipeline` - Complete data processing pipeline

Components:
- `cleaner` - Cleans input data
- `validator` - Validates data quality
- `transformer` - Transforms data format

Usage:
    mcp-cli --workflow data_processing/pipeline --input-data "data"
```

### 2. ✅ Single Entry Point

Create one main workflow that orchestrates others:

```yaml
# config/workflows/iterative_dev/dev_cycle.yaml (main entry point)
$schema: "workflow/v2.0"
name: dev_cycle
description: Complete development cycle

steps:
  - name: plan
    template: { name: planner }
  - name: implement
    template: { name: code_writer }
  - name: review
    template: { name: code_reviewer }
```

### 3. ✅ Descriptive Naming

```
✅ Good:
config/workflows/
└── user_authentication/
    ├── login.yaml
    ├── logout.yaml
    └── password_reset.yaml

❌ Bad:
config/workflows/
└── auth/
    ├── wf1.yaml
    ├── wf2.yaml
    └── wf3.yaml
```

### 4. ✅ Keep Directories Focused

```
✅ Good: 5-10 workflows per directory
❌ Bad: 50+ workflows in one directory
```

### 5. ✅ Use Consistent Naming

- **Directories:** `snake_case` or `kebab-case`
- **Files:** `snake_case.yaml`
- **Workflow names:** Match filename (without `.yaml`)

---

## Troubleshooting

### Workflow Not Found

**Error:**

```
Error: workflow 'code_reviewer' not found (searched in 'iterative_dev' directory and root)
```

**Solutions:**

1. **List all workflows:**
   
   ```bash
   mcp-cli --list-workflows
   ```

2. **Check file exists:**
   
   ```bash
   ls config/workflows/iterative_dev/code_reviewer.yaml
   ```

3. **Use explicit path:**
   
   ```yaml
   template:
     name: iterative_dev/code_reviewer    # Explicit path
   ```

4. **Verify workflow name in file:**
   
   ```bash
   grep "^name:" config/workflows/iterative_dev/code_reviewer.yaml
   ```

### Wrong Workflow Called

If local workflow isn't being found:

1. **Check both exist:**
   
   ```bash
   mcp-cli --list-workflows | grep code_reviewer
   ```

2. **Verify local is valid:**
   
   ```bash
   cat config/workflows/iterative_dev/code_reviewer.yaml
   ```

3. **Use explicit path:**
   
   ```yaml
   template:
     name: iterative_dev/code_reviewer    # Force local version
   ```

### Directory Not Scanned

If workflows in subdirectories aren't listed:

1. **Check config.yaml includes pattern:**
   
   ```yaml
   includes:
     workflows: config/workflows/**/*.yaml    # Recursive scan
   ```

2. **Verify file extension:**
   
   - ✅ `.yaml`
   - ✅ `.yml`
   - ❌ `.txt`, `.md`

3. **Verify schema header:**
   
   ```yaml
   $schema: "workflow/v2.0"    # Required!
   ```

---

## See Also

- **[Authoring Guide](AUTHORING_GUIDE.md)** - Complete workflow authoring guide
- **[Schema Reference](schema/QUICK_REFERENCE.md)** - YAML schema reference
- **[Examples](examples/)** - Working examples
- **[Patterns](patterns/)** - Design patterns

---

**Last Updated:** January 8, 2026
