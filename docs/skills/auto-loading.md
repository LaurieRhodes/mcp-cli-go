# Skills Auto-Loading Guide

Complete guide to automatically exposing Anthropic Skills as MCP tools through the `mcp-skills` server type.

## Table of Contents

- [What is Skills Auto-Loading?](#what-is-skills-auto-loading)
- [How It Works](#how-it-works)
- [Quick Setup](#quick-setup)
- [Configuration Reference](#configuration-reference)
- [Directory Structure](#directory-structure)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)
- [Under the Hood](#under-the-hood)

---

## What is Skills Auto-Loading?

**Skills auto-loading** is a feature that automatically discovers skills in your `config/skills/` directory and exposes them as MCP tools—**without requiring manual configuration for each skill**.

### Production Status

**✅ Production-Ready Features:**
- Auto-discovery of skills from directory
- SKILL.md frontmatter parsing
- MCP tool generation
- Passive mode (load documentation)
- `execute_skill_code` tool with Docker/Podman
- Include/exclude skill filtering
- Custom skills directory paths

**⚠️ Experimental/Not Implemented:**
- Active mode with `workflow.yaml` execution (stub only)
- Workflow-based skill orchestration

**Recommended Usage:** Use `execution_mode: auto` which enables full `execute_skill_code` functionality for dynamic code execution with skill helper libraries.

### The Problem It Solves

**Manual approach** (tedious):
```yaml
# config/runas/my-server.yaml
runas_type: mcp
tools:
  - name: docx_skill
    template: load_skill
    description: "Word document processing..."
    # Repeat for EVERY skill!
  - name: pdf_skill
    template: load_skill
    description: "PDF manipulation..."
  - name: xlsx_skill
    template: load_skill
    description: "Excel operations..."
  # ... 20 more skills ...
```

**Auto-loading approach** (effortless):
```yaml
# config/runas/skills-auto.yaml
runas_type: mcp-skills  # ← That's it!
```

All skills in `config/skills/` are automatically discovered and exposed as MCP tools!

### Benefits

✅ **Zero Configuration** - Add a skill directory, it's immediately available  
✅ **Single Source of Truth** - Skill metadata lives in `SKILL.md`, not duplicated in config  
✅ **Hot Reload Ready** - Add/remove skills without editing config files  
✅ **Consistent** - All skills follow the same pattern  
✅ **Maintainable** - Update skill description? Just edit SKILL.md  

---

## How It Works

### Auto-Discovery Pipeline

```
Start mcp-cli serve
        ↓
Detect runas_type: mcp-skills
        ↓
Scan config/skills/ directory
        ↓
For each subdirectory:
  1. Find SKILL.md
  2. Parse YAML frontmatter
  3. Extract: name, description
  4. Detect: scripts/, references/, assets/
        ↓
Generate MCP tool for each skill
        ↓
Expose tools via MCP protocol
        ↓
MCP clients see skills as native tools
```

### Example: What Gets Generated

**Skill directory:**
```
config/skills/docx/
├── SKILL.md          # Contains frontmatter with name & description
├── scripts/
│   └── document.py   # Helper library
└── references/
    └── ooxml.md      # Detailed docs
```

**Auto-generated MCP tool:**
```json
{
  "name": "docx",
  "description": "Create and edit Word documents with formatting...",
  "inputSchema": {
    "type": "object",
    "properties": {
      "mode": {
        "type": "string",
        "enum": ["passive", "active"],
        "default": "passive"
      },
      "include_references": {
        "type": "boolean",
        "default": false
      },
      "reference_files": {
        "type": "array",
        "items": {"type": "string"}
      },
      "input_data": {
        "type": "string"
      }
    }
  }
}
```

Plus the special `execute_skill_code` tool for dynamic code execution!

---

## Quick Setup

### 1. Create Skills Directory Structure

```bash
mkdir -p config/skills
```

### 2. Add Skills

Place skill directories in `config/skills/`:

```
config/skills/
├── bash-preference/
│   ├── LICENSE.txt
│   └── SKILL.md
├── docx/
│   ├── LICENSE.txt
│   ├── SKILL.md
│   ├── scripts/
│   │   └── document.py
│   └── references/
│       └── ooxml.md
└── frontend-design/
    ├── LICENSE.txt
    └── SKILL.md
```

### 3. Create Auto-Loading Config

Create `config/runasMCP/mcp_skills_stdio.yaml`:

```yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: skills-engine
  version: 1.0.0
  description: "Auto-generated MCP server from Anthropic skills directory"

skills_config:
  # Optional: specify custom directory
  # skills_directory: ../skills  # Relative to this config file
  
  # Execution mode: passive, active, or auto (detect Docker/Podman)
  execution_mode: auto
  
  # Optional: only expose specific skills
  # include_skills:
  #   - docx
  #   - frontend-design
  
  # Optional: exclude specific skills
  # exclude_skills:
  #   - experimental-skill
```

### 4. Start the Server

**For development (with logging):**
```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml
```

**Expected output:**
```
[INFO] Auto-discovering skills for mcp-skills server type
[INFO] Skills directory: /path/to/config/skills
[INFO] Execution mode: auto
[INFO] ✅ Executor initialized: Podman 4.9.3 (native)
[INFO] Discovered 3 skills from /path/to/config/skills
[INFO] Exposing 3 skills as MCP tools
[INFO] Created tool 'bash_preference' for skill 'bash-preference'
[INFO] Created tool 'docx' for skill 'docx'
[INFO] Created tool 'frontend_design' for skill 'frontend-design'
[INFO] Generated 4 MCP tools from skills (including execute_skill_code)
[INFO] MCP server starting...
```

**For production (Claude Desktop):**

Add to Claude Desktop config:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Linux:** `~/.config/Claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "skills": {
      "command": "/absolute/path/to/mcp-cli",
      "args": [
        "serve",
        "/absolute/path/to/config/runasMCP/mcp_skills_stdio.yaml"
      ]
    }
  }
}
```

**Important:** 
- Use **absolute paths** 
- Claude Desktop doesn't use `--verbose` (logging is at ERROR level by default)

### 5. Restart Claude Desktop

Close and reopen Claude Desktop to load the skills server.

### 6. Verify

In Claude Desktop, ask:

> "What skills do you have access to?"

Claude should list all discovered skills + `execute_skill_code`.

---

## Configuration Reference

### Minimal Configuration

```yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: skills-engine
  version: 1.0.0
  description: "Skills server"
```

That's all you need! Skills are auto-discovered from `config/skills/` by default.

### Full Configuration

```yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: skills-engine
  version: 1.0.0
  description: "Auto-generated MCP server from Anthropic skills directory"

skills_config:
  # Directory containing skills
  # Default: ../skills (relative to this config file)
  # Or: /absolute/path/to/skills
  skills_directory: ../skills
  
  # Execution mode for skills with scripts/
  # Options: passive, active, auto
  # - passive: Documentation only, no script execution
  # - active: Requires Docker/Podman, enables scripts
  # - auto: Detects Docker/Podman, falls back to passive
  execution_mode: auto
  
  # Optional: whitelist specific skills
  # If specified, ONLY these skills are exposed
  include_skills:
    - docx
    - pdf
    - frontend-design
  
  # Optional: blacklist specific skills
  # Excludes these from being exposed
  exclude_skills:
    - experimental-skill
    - deprecated-skill
```

### Configuration Field Details

#### `runas_type`
**Required.** Must be `mcp-skills` for auto-discovery.

#### `server_info`
**Required.** Metadata about the MCP server.
- `name`: Server name shown to MCP clients
- `version`: Server version
- `description`: Human-readable description

#### `skills_config.skills_directory`
**Optional.** Path to skills directory.

**Default behavior:**
- Relative to config file's directory
- Goes up one level from `config/runasMCP/` to `config/`
- Then looks for `skills/` subdirectory
- Result: `config/skills/`

**Custom paths:**
```yaml
# Relative to config file
skills_directory: ../skills

# Absolute path
skills_directory: /opt/mcp-cli/skills

# Different location
skills_directory: ../../shared-skills
```

**Path resolution:**
- If path is relative → resolved from config file's directory
- If path is absolute → used as-is

#### `skills_config.execution_mode`
**Optional.** Default: `auto`

Controls how skills with `scripts/` directories are handled:

| Mode | Behavior | Status |
|------|----------|--------|
| `passive` | Documentation only, scripts never execute | ✅ Production |
| `active` | Attempts workflow execution (experimental) | ⚠️ Stub only |
| `auto` | Detects Docker/Podman, enables code execution | ✅ Production |

**Current Implementation:**

**Passive mode (Production-ready):**
```
User: "Create a Word document"
→ Loads SKILL.md documentation
→ Claude reads and uses guidance
→ Claude uses execute_skill_code to write custom code
→ Executes in sandbox with helper libraries
✅ Document created
```

**Active mode (Experimental - Not Implemented):**
```
⚠️ WARNING: Active mode with workflow.yaml is NOT yet implemented.
The code accepts the parameter but returns "not yet implemented" error.
```

**Auto mode (Recommended - Production-ready):**
```
If Docker/Podman available:
  → Enables execute_skill_code (sandboxed execution)
  → Skills can use helper libraries
Else:
  → Passive mode only (documentation)
  → Warn user about disabled script execution
```

**Recommendation:** Use `auto` mode. This enables the fully-functional `execute_skill_code` tool for dynamic code execution with skill helper libraries.

#### `skills_config.include_skills`
**Optional.** Whitelist of skills to expose.

If specified, **ONLY** these skills become MCP tools:

```yaml
include_skills:
  - docx
  - pdf
```

Result: Only `docx` and `pdf` tools are created, all others ignored.

**Use case:** Large skill library, only want subset exposed.

#### `skills_config.exclude_skills`
**Optional.** Blacklist of skills to hide.

```yaml
exclude_skills:
  - experimental
  - deprecated-api
```

Result: These skills are **not** exposed as tools.

**Use case:** Skip incomplete or deprecated skills.

**Note:** Cannot use both `include_skills` and `exclude_skills` together—`include_skills` takes precedence.

---

## Directory Structure

### Required Directory Layout

```
your-project/
└── config/
    ├── runasMCP/
    │   └── mcp_skills_stdio.yaml  ← Auto-loading config
    └── skills/                     ← Skills directory
        ├── skill-one/
        │   ├── LICENSE.txt
        │   └── SKILL.md            ← Required
        ├── skill-two/
        │   ├── LICENSE.txt
        │   ├── SKILL.md            ← Required
        │   └── scripts/
        │       └── helper.py
        └── skill-three/
            ├── LICENSE.txt
            ├── SKILL.md            ← Required
            ├── scripts/
            │   └── automation.py
            └── references/
                └── docs.md
```

### Skill Directory Requirements

Each skill directory **must** contain:

1. **`SKILL.md`** with YAML frontmatter:
   ```markdown
   ---
   name: skill-name
   description: "What this skill does and when to use it"
   license: MIT License
   ---
   
   # Skill Name
   
   Main documentation content...
   ```

2. **`LICENSE.txt`** (recommended for distribution)

### Optional Components

**`scripts/`** - Python helper libraries:
```
scripts/
├── __init__.py
├── document.py
└── formatting.py
```

**`references/`** - Additional documentation:
```
references/
├── api-docs.md
├── examples.md
└── schemas.json
```

**`assets/`** - Static files:
```
assets/
├── templates/
│   └── default.docx
└── fonts/
    └── custom.ttf
```

**`workflow.yaml`** - ⚠️ **NOT IMPLEMENTED** (Planned for active mode):
```yaml
# Note: workflow.yaml detection exists but execution is not implemented
# Active mode currently returns "not yet implemented" error
name: create_document
steps:
  - action: load_template
  - action: fill_content
  - action: save
```

---

## Troubleshooting

### Skills Not Appearing

**Symptom:** Claude Desktop shows only `execute_skill_code` tool, no individual skills.

**Diagnosis:**

Run with verbose logging:
```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | head -50
```

Look for:
```
[INFO] Auto-discovering skills for mcp-skills server type
[INFO] Skills directory: /path/to/config/skills
[INFO] Discovered X skills from /path/to/config/skills
```

**Common causes:**

#### 1. Wrong Path

**Symptom:**
```
[WARN] Skills directory does not exist: ../skills
[INFO] Discovered 0 skills
```

**Fix:**

Check path resolution:
```yaml
# Config at: /path/to/config/runasMCP/mcp_skills_stdio.yaml
skills_config:
  skills_directory: ../skills  # Resolves to /path/to/config/skills ✅
```

Or use absolute path:
```yaml
skills_config:
  skills_directory: /absolute/path/to/skills
```

#### 2. Missing SKILL.md

**Symptom:**
```
[WARN] Failed to load skill from /path/to/skills/my-skill: SKILL.md not found
```

**Fix:**

Ensure every skill has `SKILL.md`:
```bash
ls config/skills/*/SKILL.md
```

#### 3. Invalid Frontmatter

**Symptom:**
```
[WARN] Failed to parse frontmatter: SKILL.md must start with '---'
```

**Fix:**

Check `SKILL.md` format:
```markdown
---
name: skill-name
description: "Valid YAML string"
---

# Content starts here
```

**Common mistakes:**
- Missing opening `---`
- Missing closing `---`
- Invalid YAML syntax in frontmatter
- Missing `name` or `description` field

#### 4. Binary Not Rebuilt

**Symptom:** Code changes not reflected.

**Fix:**
```bash
cd /path/to/mcp-cli-go
go build -o mcp-cli
```

Then restart Claude Desktop.

### Scripts Not Executing

**Symptom:** Skills load, but `execute_skill_code` returns "executor not available."

**Diagnosis:**

Check for Docker/Podman:
```bash
docker --version
# or
podman --version
```

**Fix:**

Install Docker or Podman:

**macOS:**
```bash
brew install --cask docker
# or
brew install podman
```

**Linux:**
```bash
# Docker
curl -fsSL https://get.docker.com | sh

# or Podman (recommended for rootless)
sudo apt-get install podman
```

**Windows:**
- Install Docker Desktop
- Or install Podman Desktop

**Verify initialization:**
```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | grep Executor
```

Expected:
```
[INFO] ✅ Executor initialized: Podman 4.9.3 (native)
```

If you see:
```
[WARN] Docker/Podman not available, falling back to passive mode
```

Then scripts won't execute (passive mode only).

### Permission Errors

**Symptom:** 
```
[ERROR] Failed to start MCP server: permission denied
```

**Fix:**

Ensure binary is executable:
```bash
chmod +x /path/to/mcp-cli
```

For Docker (if not using rootless):
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

For Podman (recommended - rootless by default):
```bash
# No special permissions needed
podman --version
```

### Claude Desktop Not Seeing Tools

**Symptom:** Server connects, but no tools available in Claude.

**Diagnosis:**

Check Claude Desktop logs:

**macOS:**
```bash
tail -f ~/Library/Logs/Claude/mcp*.log
```

**Linux:**
```bash
tail -f ~/.config/Claude/logs/mcp*.log
```

**Look for:**
```
[skills] [info] Message from server: {"jsonrpc":"2.0","id":1,"result":{"tools":[...]}}
```

**Common causes:**

1. **Wrong path in config:**
   ```json
   {
     "command": "mcp-cli",  // ❌ Relative path
   }
   ```
   
   Fix:
   ```json
   {
     "command": "/absolute/path/to/mcp-cli",  // ✅ Absolute path
   }
   ```

2. **Config not reloaded:**
   - Always restart Claude Desktop after changing config
   - On macOS: Cmd+Q to fully quit, then reopen

3. **Server crashed on startup:**
   ```
   [skills] [error] Server process exited with code 1
   ```
   
   Fix: Run manually to see error:
   ```bash
   /absolute/path/to/mcp-cli serve /absolute/path/to/config.yaml
   ```

---

## Advanced Configuration

### Multiple Skill Directories

You can't directly specify multiple directories, but you can use symlinks:

```bash
mkdir -p config/skills
ln -s /path/to/shared-skills/* config/skills/
ln -s /path/to/custom-skills/* config/skills/
```

### Selective Exposure

**Scenario:** You have 50 skills, only want 5 exposed.

```yaml
skills_config:
  include_skills:
    - docx
    - pdf
    - xlsx
    - frontend-design
    - mcp-builder
```

**Scenario:** Expose all except experimental ones.

```yaml
skills_config:
  exclude_skills:
    - experimental-pdf
    - deprecated-api
    - test-skill
```

### Custom Execution Modes Per Skill

Currently not supported directly, but you can work around it:

**Option 1:** Separate servers

```yaml
# config/runasMCP/passive-skills.yaml
runas_type: mcp-skills
server_info:
  name: passive-skills
skills_config:
  execution_mode: passive
  include_skills:
    - documentation-only-skill

# config/runasMCP/active-skills.yaml
runas_type: mcp-skills
server_info:
  name: active-skills
skills_config:
  execution_mode: active
  include_skills:
    - script-heavy-skill
```

Configure both in Claude Desktop:
```json
{
  "mcpServers": {
    "passive-skills": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/passive-skills.yaml"]
    },
    "active-skills": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/active-skills.yaml"]
    }
  }
}
```

**Option 2:** Remove `scripts/` directory from passive skills

If a skill doesn't have a `scripts/` directory, execution mode doesn't matter.

---

## Under the Hood

### Discovery Process

When `mcp-cli serve` starts with `runas_type: mcp-skills`:

```go
// 1. Detect skills directory
skillsDir := resolveSkillsDirectory(config)

// 2. Initialize skill service
skillService.Initialize(skillsDir, executionMode)

// 3. Scan directory
entries := os.ReadDir(skillsDir)
for entry in entries {
    if entry.IsDir() {
        // 4. Load skill
        skill := loadSkillFromDirectory(entry)
        
        // 5. Validate
        if skill.Validate() {
            skills[skill.Name] = skill
        }
    }
}

// 6. Generate MCP tools
for skill in skills {
    tool := ToolExposure{
        Name: skill.GetMCPToolName(),
        Description: skill.GetToolDescription(),
        Template: "load_skill",
        InputSchema: skill.GetMCPInputSchema(),
    }
    tools = append(tools, tool)
}

// 7. Add execute_skill_code tool
tools = append(tools, executeSkillCodeTool)

// 8. Start MCP server
server.Start(tools)
```

### Tool Generation

Each skill becomes an MCP tool:

**Input:** Skill directory with `SKILL.md`:
```markdown
---
name: my-skill
description: "Does something useful"
---
```

**Output:** MCP tool definition:
```json
{
  "name": "my_skill",
  "description": "Does something useful",
  "inputSchema": {
    "type": "object",
    "properties": {
      "mode": {
        "type": "string",
        "enum": ["passive", "active"],
        "default": "passive",
        "description": "Load mode"
      },
      "include_references": {
        "type": "boolean",
        "default": false,
        "description": "Include reference files"
      },
      "reference_files": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Specific references to load"
      },
      "input_data": {
        "type": "string",
        "description": "Input data for active mode"
      }
    }
  }
}
```

### Execute Skill Code Tool

Special tool for dynamic code execution:

```json
{
  "name": "execute_skill_code",
  "description": "Execute code with access to skill helper libraries",
  "inputSchema": {
    "type": "object",
    "properties": {
      "skill_name": {
        "type": "string",
        "description": "Name of skill (e.g., 'docx', 'pdf')"
      },
      "language": {
        "type": "string",
        "enum": ["python"],
        "default": "python"
      },
      "code": {
        "type": "string",
        "description": "Code to execute"
      },
      "files": {
        "type": "object",
        "description": "Files to make available (filename -> base64)"
      }
    },
    "required": ["skill_name", "code"]
  }
}
```

### Execution Flow

When Claude calls a skill tool:

```
MCP client (Claude Desktop)
    ↓
MCP protocol (JSON-RPC)
    ↓
mcp-cli MCP server
    ↓
Check template type
    ├─ load_skill → Load SKILL.md, return as text
    └─ execute_skill_code → Execute in sandbox
           ↓
       Create workspace
           ↓
       Mount skill directory (read-only)
           ↓
       Set PYTHONPATH=/skill
           ↓
       Execute code in Podman/Docker
           ↓
       Capture output
           ↓
       Return to Claude
```

---

## Best Practices

### 1. Use Auto-Discovery

Don't manually configure skills unless you need fine-grained control.

**Bad:**
```yaml
runas_type: mcp
tools:
  - name: docx
    template: load_skill
    # ... manual config
```

**Good:**
```yaml
runas_type: mcp-skills
# Auto-discovers all skills!
```

### 2. Keep Skills in `config/skills/`

Follow the conventional structure:
```
config/
├── runasMCP/
│   └── mcp_skills_stdio.yaml
└── skills/
    └── my-skill/
```

### 3. Use Absolute Paths in Claude Desktop Config

**Bad:**
```json
{
  "command": "./mcp-cli"
}
```

**Good:**
```json
{
  "command": "/Users/alice/projects/mcp-cli-go/mcp-cli"
}
```

### 4. Start Simple

Begin with `execution_mode: auto` and no filters:
```yaml
runas_type: mcp-skills
version: "1.0"
server_info:
  name: skills-engine
  version: 1.0.0
  description: "Skills server"
skills_config:
  execution_mode: auto
```

Add filters only when you need them.

### 5. Test Before Production

Always test with `--verbose` first:
```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml
```

Check:
- ✅ Skills directory found
- ✅ Expected number of skills discovered
- ✅ No warnings about missing SKILL.md files
- ✅ Executor initialized (if you want script execution)

### 6. Use Descriptive Frontmatter

The `description` field is what LLMs see:

**Bad:**
```yaml
description: "Handles documents"
```

**Good:**
```yaml
description: "Create and edit Word documents with formatting, tables, images, and styles. Use when creating reports, proposals, or any .docx files."
```

### 7. Version Your Skills Directory

```bash
git init config/skills
git add config/skills
git commit -m "Initial skills setup"
```

Track changes to skills separately from application code.

---

## Summary

**Skills auto-loading** with `runas_type: mcp-skills` provides:

✅ **Zero configuration** - Add skills, they're immediately available  
✅ **Automatic discovery** - Scans directory, parses metadata  
✅ **Dynamic tool generation** - Creates MCP tools on startup  
✅ **Flexible execution** - Passive (docs) or active (scripts) modes  
✅ **Easy filtering** - Include/exclude specific skills  
✅ **Production ready** - Works with Claude Desktop and any MCP client  

**Next steps:**

- **Create skills:** See [Creating Skills Guide](creating-skills.md)
- **Quick reference:** See [Quick Reference](quick-reference.md)
- **Understand internals:** See [Architecture Overview](overview.md)

---

**Last Updated:** January 4, 2026  
**Tested With:** mcp-cli v0.1.0, Claude Desktop, Podman 4.9.3
