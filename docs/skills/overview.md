# Skills Overview

## What Are Skills?

Skills are modular packages that provide:
1. Documentation for LLMs
2. Helper libraries (python-pptx, openpyxl, etc.)
3. Container-based safe execution

## How Skills Work

```
User Request
    ↓
LLM reads skill documentation
    ↓
LLM writes custom code
    ↓
Code executes in isolated container
    ↓
Result returned to user
```

## Skill Components

### Required

**SKILL.md** - Documentation with frontmatter:
```markdown
---
name: skill-name
description: "What it does"
---

# Documentation

Instructions for using this skill...
```

### Optional

**scripts/** - Python helper libraries:
```
skill-name/
└── scripts/
    ├── __init__.py
    └── helper.py
```

**examples/** - Usage examples:
```
skill-name/
└── examples/
    └── example1.py
```

## Execution Modes

### Passive Mode

Loads skill documentation only:
```json
{
  "tool": "skill_name",
  "arguments": {"mode": "passive"}
}
```

### Active Mode

Executes code with skill libraries:
```json
{
  "tool": "execute_skill_code",
  "arguments": {
    "skill_name": "pptx",
    "code": "from pptx import Presentation\n...",
    "language": "python"
  }
}
```

## Container Execution

All code runs in Docker/Podman containers:

**Mounts:**
- `/workspace` - Temporary (deleted after)
- `/outputs` - Persistent (from host)
- `/skill` - Read-only libraries

**Security:**
- No network access
- Read-only root filesystem
- Memory and CPU limits

## Available Skills

Located in `config/skills/`:
- **docx** - Word documents
- **pptx** - PowerPoint
- **xlsx** - Excel
- **pdf** - PDF manipulation

## Configuration

`config/settings.yaml`:
```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"
```

`config/skills/skill-images.yaml`:
```yaml
skills:
  pptx: mcp-skills-pptx
  docx: mcp-skills-docx
```

## Creating Skills

See [creating-skills.md](creating-skills.md) for details.

---

Last updated: January 6, 2026
