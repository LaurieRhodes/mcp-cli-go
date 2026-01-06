# Skills Documentation

Skills extend LLM capabilities through the Model Context Protocol (MCP).

## Quick Start

1. **Build container images**:
   
   ```bash
   cd docker/skills
   ./build-skills-images.sh
   ```

2. **Configure outputs directory** in `config/settings.yaml`:
   
   ```yaml
   skills:
     outputs_dir: "/path/to/outputs"
   ```

3. **Create directory**:
   
   ```bash
   mkdir -p /path/to/outputs
   ```

## What Are Skills?

Skills are modular packages that provide:

- Specialized knowledge for specific tasks
- Helper libraries (python-pptx, openpyxl, etc.)
- Safe container-based execution

## Available Skills

Located in `config/skills/`:

- **docx** - Word document creation
- **pptx** - PowerPoint presentations
- **xlsx** - Excel spreadsheets
- **pdf** - PDF manipulation

## Skill Structure

```
skill-name/
├── SKILL.md           # Documentation
├── scripts/           # Helper libraries (optional)
└── examples/          # Usage examples (optional)
```

## How Skills Work

```
LLM Request
    ↓
mcp-cli (via MCP)
    ↓
Container Execution
    /workspace (temporary)
    /outputs (persistent from host)
    /skill (read-only libraries)
    ↓
File persists on host
```

## Usage

Skills are accessed via MCP tools:

**Passive mode** - Load skill documentation:

```json
{
  "tool": "pptx",
  "arguments": {
    "mode": "passive"
  }
}
```

**Active mode** - Execute code with skill libraries:

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

## Configuration

### Outputs Directory

`config/settings.yaml`:

```yaml
skills:
  outputs_dir: "/home/user/outputs"
```

### Container Images

`config/skills/skill-images.yaml`:

```yaml
skills:
  pptx: mcp-skills-pptx
  docx: mcp-skills-docx
  xlsx: mcp-skills-xlsx
  pdf: mcp-skills-pdf
```

## Security

All skill execution runs in isolated containers:

- No network access
- Read-only root filesystem
- Memory and CPU limits
- Automatic cleanup

## Documentation

See [INDEX.md](INDEX.md) for complete documentation navigation.

**Essential guides**:

- [CONTAINER_SETUP.md](CONTAINER_SETUP.md) - Container configuration
- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - File persistence
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues
- [creating-skills.md](creating-skills.md) - Build custom skills

## Cross-LLM Support

Skills work with any MCP-compatible LLM:

- GPT-4
- DeepSeek
- Gemini
- Claude
- Kimi

---

Last updated: January 6, 2026
