# Skills Auto-Loading Quick Reference

Quick reference for setting up and troubleshooting skills auto-loading.

## ⚡ Quick Setup (2 Minutes)

### 1. Create Config

```yaml
# config/runasMCP/mcp_skills_stdio.yaml
runas_type: mcp-skills
version: "1.0"

server_info:
  name: skills-engine
  version: 1.0.0
  description: "Skills server"

skills_config:
  skills_directory: ../skills  # Relative to this file
  execution_mode: auto         # auto|passive|active
```

### 2. Configure Claude Desktop

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "skills": {
      "command": "/absolute/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/mcp_skills_stdio.yaml"]
    }
  }
}
```

### 3. Restart Claude Desktop

Cmd+Q (macOS) or fully close and reopen.

---

## 📁 Directory Structure

```
config/
├── runasMCP/
│   └── mcp_skills_stdio.yaml    ← Config file
└── skills/                       ← Skills directory (auto-discovered)
    ├── skill-one/
    │   ├── LICENSE.txt
    │   └── SKILL.md              ← Required: name + description
    └── skill-two/
        ├── LICENSE.txt
        ├── SKILL.md              ← Required
        └── scripts/              ← Optional: helper libraries
            └── helper.py
```

---

## 🔍 Skill Requirements

### Minimal Skill

```
my-skill/
├── LICENSE.txt
└── SKILL.md
```

### SKILL.md Format

```markdown
---
name: my-skill
description: "What it does and when to use it"
license: MIT License
---

# My Skill

Documentation for Claude...
```

**Required frontmatter fields:**
- `name` - Skill identifier (lowercase-with-dashes)
- `description` - What it does, when to use it

---

## 🛠️ Configuration Options

### Execution Modes

```yaml
execution_mode: auto      # Detect Docker/Podman (recommended) ✅ Production
execution_mode: passive   # Docs only, no scripts ✅ Production
execution_mode: active    # ⚠️ NOT IMPLEMENTED (stub only)
```

**Note:** Use `auto` for full functionality. `active` mode is not yet implemented.

### Filtering Skills

**Include only specific skills:**
```yaml
skills_config:
  include_skills:
    - docx
    - pdf
    - frontend-design
```

**Exclude specific skills:**
```yaml
skills_config:
  exclude_skills:
    - experimental
    - deprecated-api
```

### Custom Directory

**Relative path (to config file):**
```yaml
skills_directory: ../skills
```

**Absolute path:**
```yaml
skills_directory: /opt/mcp-cli/skills
```

---

## 🐛 Troubleshooting Commands

### Test with verbose logging

```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | head -50
```

**Expected output:**
```
[INFO] Auto-discovering skills for mcp-skills server type
[INFO] Skills directory: /path/to/config/skills
[INFO] Discovered X skills from /path/to/config/skills
[INFO] Created tool 'skill_name' for skill 'skill-name'
[INFO] Generated X MCP tools from skills
```

### Check Claude Desktop logs

**macOS:**
```bash
tail -f ~/Library/Logs/Claude/mcp*.log
```

**Linux:**
```bash
tail -f ~/.config/Claude/logs/mcp*.log
```

### Verify skills directory

```bash
ls -la config/skills/*/SKILL.md
```

### Check Docker/Podman

```bash
docker --version
# or
podman --version
```

---

## 🚨 Common Issues

### Issue: No skills discovered

**Log shows:**
```
[WARN] Skills directory does not exist: ../skills
[INFO] Discovered 0 skills
```

**Fix:**
```yaml
# Use absolute path
skills_config:
  skills_directory: /absolute/path/to/config/skills
```

Or verify relative path is correct:
```bash
# If config is at: config/runasMCP/mcp_skills_stdio.yaml
# Then ../skills resolves to: config/skills
```

### Issue: SKILL.md parse error

**Log shows:**
```
[WARN] Failed to parse frontmatter: SKILL.md must start with '---'
```

**Fix:**

Check SKILL.md format:
```markdown
---
name: skill-name
description: "Valid description"
---

Content starts here
```

Common mistakes:
- Missing opening `---`
- Missing closing `---`  
- Invalid YAML in frontmatter
- Missing required fields

### Issue: Scripts not executing

**Log shows:**
```
[WARN] Docker/Podman not available, falling back to passive mode
```

**Fix:**

Install Docker or Podman:

**macOS:**
```bash
brew install podman
```

**Linux:**
```bash
sudo apt-get install podman
```

**Verify:**
```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | grep Executor
```

Expected:
```
[INFO] ✅ Executor initialized: Podman 4.9.3 (native)
```

### Issue: Claude Desktop doesn't see tools

**Check:**
1. ✅ Used absolute paths in config?
2. ✅ Restarted Claude Desktop (fully quit)?
3. ✅ Server starts without errors?

**Test server manually:**
```bash
/absolute/path/to/mcp-cli serve /absolute/path/to/config.yaml
```

If it crashes, you'll see the error.

---

## 📊 What Gets Auto-Generated

### For each skill:

**Input:**
```
config/skills/docx/
└── SKILL.md (with name: docx, description: "...")
```

**Output:**
```json
{
  "name": "docx",
  "description": "...",
  "inputSchema": { /* standard schema */ }
}
```

### Plus special tool:

```json
{
  "name": "execute_skill_code",
  "description": "Execute Python code with skill helper libraries",
  "inputSchema": {
    "properties": {
      "skill_name": {"type": "string"},
      "code": {"type": "string"},
      "files": {"type": "object"}
    }
  }
}
```

---

## 🔄 Workflow

```
Start server
    ↓
Detect runas_type: mcp-skills
    ↓
Scan config/skills/
    ↓
Parse each SKILL.md
    ↓
Generate MCP tools
    ↓
Expose via MCP protocol
    ↓
Claude Desktop sees tools
```

---

## ✅ Verification Checklist

Before deploying:

- [ ] Skills directory exists and contains skills
- [ ] Each skill has SKILL.md with valid frontmatter
- [ ] Config uses absolute paths
- [ ] Tested with `--verbose` flag
- [ ] Saw "Discovered X skills" in logs
- [ ] Docker/Podman installed (if using scripts)
- [ ] Claude Desktop config updated
- [ ] Claude Desktop restarted
- [ ] Asked Claude "what skills do you have?"

---

## 📚 Full Documentation

For detailed information, see:
- **[Auto-Loading Guide](auto-loading.md)** - Complete guide
- **[Quick Start](quick-start.md)** - Step-by-step tutorial
- **[Overview](overview.md)** - How skills work
- **[Creating Skills](creating-skills.md)** - Build your own

---

**Last Updated:** January 4, 2026
