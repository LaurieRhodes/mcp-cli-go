# Skills Auto-Loading

Skills in `config/skills/` are automatically discovered and exposed as MCP tools.

## How It Works

1. mcp-cli scans `config/skills/` directory
2. Reads each skill's `SKILL.md` frontmatter
3. Generates MCP tool for each skill
4. Exposes tools via MCP protocol

## Configuration

**Minimal MCP server config** (`config/runasMCP/mcp_skills_stdio.yaml`):
```yaml
runas_type: mcp-skills
skills_config:
  skills_directory: ../skills
  execution_mode: auto
```

**Skill capabilities config** (`config/skills/skill-images.yaml`):
```yaml
skills:
  docx:
    image: mcp-skills-docx
    language: python
    description: "Word document manipulation"
```

This file is **critical** - it:
- Maps skills to container images
- Declares language capabilities (advertised via MCP)
- Sets resource limits and security

See [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) for complete reference.

**Optional filtering:**
```yaml
skills_config:
  include_skills:
    - docx
    - pptx
  exclude_skills:
    - deprecated-skill
```

## Skill Requirements

Each skill directory must have `SKILL.md` with frontmatter:

```markdown
---
name: skill-name
description: "What it does"
---

# Documentation
...
```

## Directory Structure

```
config/
├── runasMCP/
│   └── mcp_skills_stdio.yaml
└── skills/
    ├── docx/
    │   └── SKILL.md
    ├── pptx/
    │   ├── SKILL.md
    │   └── scripts/
    │       └── helpers.py
    └── xlsx/
        └── SKILL.md
```

## Execution Modes

**auto** (recommended):
- Detects Docker/Podman
- Enables `execute_skill_code` tool
- Full script execution support

**passive**:
- Load documentation only
- No code execution
- No container required

## Generated Tools

For each skill, mcp-cli generates:

1. **Skill tool** (passive loading):
   - Name: skill name from frontmatter
   - Loads SKILL.md content

2. **execute_skill_code** (if execution_mode: auto):
   - Execute Python code with skill libraries
   - Access to helper functions in `scripts/`

## Custom Skills Directory

```yaml
skills_config:
  skills_directory: /absolute/path/to/skills
  # or relative to config file:
  skills_directory: ../my-skills
```

## Verification

```bash
# Start server
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml

# Check logs
# Should see: "Initialized skill service with N skills"
# Should see: "Auto-generated N MCP tool definitions"
```

## Adding New Skills

1. Create directory in `config/skills/`
2. Add `SKILL.md` with frontmatter
3. Optional: Add `scripts/` directory
4. Restart MCP server
5. Skill automatically available

No configuration changes needed!

## Troubleshooting

**Skill not detected:**
- Check `SKILL.md` exists
- Verify frontmatter has `name` and `description`
- Check server logs for parsing errors

**Tool not appearing:**
- Restart MCP server
- Check `exclude_skills` list
- Verify `include_skills` list (if used)

**Execution failing:**
- Check execution_mode is `auto`
- Verify Docker/Podman available
- Check container images built

---

Last updated: January 20, 2026
