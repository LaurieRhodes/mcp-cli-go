# Skills Quick Reference

## Setup

**1. Build images:**
```bash
cd docker/skills && ./build-skills-images.sh
```

**2. Configure outputs** (`config/settings.yaml`):
```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"
```

**3. Create directory:**
```bash
mkdir -p /tmp/mcp-outputs
```

**4. Start server:**
```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

## Skill Structure

```
skill-name/
├── SKILL.md           # Required: name + description
├── scripts/           # Optional: helper libraries
└── examples/          # Optional: usage examples
```

### SKILL.md Format

```markdown
---
name: skill-name
description: "What it does"
---

# Skill Name

Documentation...
```

## Configuration

### MCP Server Config

`config/runasMCP/mcp_skills_stdio.yaml`:
```yaml
runas_type: mcp-skills
skills_config:
  skills_directory: ../skills
  execution_mode: auto
```

### Settings

`config/settings.yaml`:
```yaml
skills:
  outputs_dir: "/path/to/outputs"
```

### Image Mapping

`config/skills/skill-images.yaml`:
```yaml
skills:
  pptx: mcp-skills-pptx
  docx: mcp-skills-docx
```

## Usage

**Load skill documentation (passive):**
```json
{
  "tool": "pptx",
  "arguments": {"mode": "passive"}
}
```

**Execute code (active):**
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

## Troubleshooting

**Check images:**
```bash
docker images | grep mcp-skills
```

**Check outputs:**
```bash
ls -ld /tmp/mcp-outputs
grep outputs_dir config/settings.yaml
```

**Check skill loading:**
```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
# Look for: "Initialized skill service with N skills"
```

---

Last updated: January 6, 2026
