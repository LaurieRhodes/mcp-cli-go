# Skills - Extend Any LLM with MCP

**Skills** extend LLM capabilities through the Model Context Protocol (MCP), making them work with **any MCP-compatible LLM** - GPT-4, Claude, DeepSeek, Gemini, Qwen, and more.

---

## 🚀 Quick Start (5 Minutes)

```bash
# 1. Build containers
cd docker/skills && ./build-skills-images.sh

# 2. Setup outputs directory
mkdir -p /tmp/mcp-outputs
echo "skills:
  outputs_dir: \"/tmp/mcp-outputs\"" >> config/settings.yaml

# 3. Test it!
./mcp-cli chat --servers skills
# Ask: "Create a PowerPoint about quantum computing"
```

**Works?** ✅ [Read complete guide](COMPLETE_GUIDE.md) | **Issues?** See [Troubleshooting](INDEX.md#troubleshooting)

---

## What Are Skills?

Skills are both a context store of information related to LLM tasks and **containerized packages** that extend LLM capabilities:

```
LLM (any MCP client)
    ↓ MCP Protocol
mcp-cli Skills Server
    ↓ Configured via skill-images.yaml
Docker Container (isolated, secure)
    ├── /workspace (temporary)
    ├── /outputs (persists to host) ✅
    └── /skill (helper libraries)
    ↓
Files saved to /tmp/mcp-outputs/
```

**Three ways to use:**

1. **Chat mode:** `./mcp-cli chat --servers skills`
2. **Workflows:** Add `servers: [skills]` to any workflow
3. **MCP server:** `./mcp-cli serve` → Use with Claude Desktop, VS Code, etc.

---

## 📦 Available Skills (Default)

| Skill    | Language | Use Case                                 |
| -------- | -------- | ---------------------------------------- |
| **docx** | Python   | Word documents, reports, letters         |
| **pptx** | Python   | PowerPoint presentations                 |
| **xlsx** | Python   | Excel spreadsheets, data analysis        |
| **pdf**  | Python   | PDF manipulation, forms, text extraction |

**[Build your own](creating-skills.md)** → Add custom capabilities

---

## 🔑 Critical Configuration: skill-images.yaml

**Location:** `config/skills/skill-images.yaml`

Default config directories are created from the first use of **mcp-cli init**

This file is **critical** - it maps skills to containers and advertises capabilities via MCP:

```yaml
skills:
  docx:
    image: mcp-skills-docx
    language: python          # ← Advertised to LLMs via MCP
    description: "Word document manipulation"
```

**Why it matters:**

- ✅ LLMs know which language to use (Python vs bash)
- ✅ Skills get the right container
- ✅ Resource limits enforced (memory, CPU, timeout)
- ✅ Security policies applied (no network by default)

**[Read full guide](SKILL_IMAGES_YAML.md)**

---

## 📂 File Persistence

Files ONLY persist when saved to `/outputs/`:

```python
# ✅ CORRECT - File persists to host
doc.save('/outputs/report.docx')

# ❌ WRONG - File deleted when container exits
doc.save('report.docx')
```

**Configured in:** `config/settings.yaml` → `skills.outputs_dir`

**[Read full guide](OUTPUTS_DIRECTORY.md)**

---

## 🌐 Model Compatibility

**MCP-capable models (tested):**

| Model | Python Skills | Bash Skills | Notes |
|-------|---------------|-------------|-------|
| **DeepSeek Chat** | ✅ Excellent | ✅ Excellent | Language-agnostic, budget-friendly |
| **Claude Haiku 4.5** | ✅ Excellent | ✅ Excellent | Language-agnostic, fastest |
| **GPT-5 mini** | ✅ Excellent | ✅ Excellent | Language-agnostic, balanced |
| **Gemini 2.0 Flash Exp** | ✅ Excellent | ❌ Python-only | Fast & free, but forces Python |

**Incompatible models (do not work):**
- ❌ GPT-4o-mini (use GPT-5 mini instead)
- ❌ Kimi (fundamentally confused)
- ❌ Gemini Lite variants (too weak)

**Key insight:** Only **3 models** (DeepSeek, Haiku 4.5, GPT-5 mini) work with any skill type. Others either fail completely or force Python.

**[Read full test report](COMPREHENSIVE_MODEL_TESTING_REPORT.md)**

---

## 📖 Documentation

### By Experience Level

**New users:**

1. [quick-start.md](quick-start.md) - 5-minute setup
2. [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md) - Comprehensive guide

**Experienced users:**

- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) - Configuration reference
- [creating-skills.md](creating-skills.md) - Build custom skills
- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - File persistence

**Reference:**

- [INDEX.md](INDEX.md) - Complete documentation index
- [overview.md](overview.md) - Architecture deep dive

---

## 🎯 Common Use Cases

### 1. Document Generation (Chat Mode)

```bash
./mcp-cli chat --servers skills
```

Ask: "Create a Word document with monthly sales data"

The LLM will:

1. Read docx skill documentation
2. Write code using python-docx library
3. Execute in container
4. Save to `/tmp/mcp-outputs/sales_report.docx` ✅

### 2. Automated Workflows

```yaml
# workflow.yaml
execution:
  provider: deepseek
  servers: [skills]

steps:
  - name: analyze_and_report
    skills: [xlsx, pptx]
    run: |
      1. Read sales data from /outputs/data.xlsx
      2. Analyze trends
      3. Create PowerPoint presentation
```

*Note that with workflows we need to load the generic skills MCP server and later filter individual skills to respective steps.  This minimises context bloat for small LLMs.*

### 3. MCP Server (Use with Claude Desktop, VS Code, etc.)

```bash
# Start MCP server
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

Add to Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "skills": {
      "command": "/absolute/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/mcp_skills_stdio.yaml"]
    }
  }
}
```

Now Claude Desktop can use all your skills!

---

## 🔧 Configuration

### settings.yaml (Global)

```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"  # Where shared files persist on the host
```

### skill-images.yaml (Skill Registry) ⭐

```yaml
defaults:
  language: python
  memory: 256MB
  network_mode: none

skills:
  docx:
    image: mcp-skills-docx
    language: python
    description: "Word document manipulation"

  my-custom-skill:
    image: my-container
    language: bash
    memory: 512MB
```

**[Read complete reference](SKILL_IMAGES_YAML.md)**

---

## 🔒 Security

All skill execution runs in **isolated containers**:

- ✅ No network access (by default)
- ✅ Read-only root filesystem
- ✅ Memory and CPU limits
- ✅ Automatic cleanup
- ✅ Isolated from host system

**Only** `/outputs/` directory is writable and persists to host.

---

## 🌍 Model Compatibility (Tested)

**Language-agnostic models (work with any skill):**

| Model | Python | Bash | Speed | Cost/1K runs |
|-------|--------|------|-------|--------------|
| **DeepSeek Chat** | ✅ | ✅ | 39-60s | $2.45 |
| **Claude Haiku 4.5** | ✅ | ✅ | 13-16s | $10.00 |
| **GPT-5 mini** | ✅ | ✅ | 24-28s | $3.25 |

**Python-only models:**
- Gemini 2.0 Flash Exp: ✅ Python (9s, FREE) | ❌ Bash (forces Python)

**Incompatible models (do not use):**
- ❌ GPT-4o-mini (wrong output structure)
- ❌ Kimi (confused by MCP patterns)
- ❌ Gemini Lite variants (too weak)

**Untested models:**
- GPT-4, GPT-4o (likely work but not tested)
- Claude Opus, Sonnet (likely work but not tested)
- Qwen (untested)

**[Read full test report](COMPREHENSIVE_MODEL_TESTING_REPORT.md)** for detailed benchmarks.

---

## 🐛 Common Issues & Solutions

### Files Not Appearing

**Problem:** LLM creates file but it's not in outputs directory

**Solutions:**

1. Check code uses `/outputs/` path:
   
   ```python
   doc.save('/outputs/file.docx')  # ✅ Correct
   ```
2. Verify `config/settings.yaml` has correct `outputs_dir`
3. Ensure directory exists: `mkdir -p /tmp/mcp-outputs`

**[Read full guide](OUTPUTS_DIRECTORY.md)**

### Skill Not Loading

**Problem:** Skill doesn't appear in tools list

**Solutions:**

1. Check `SKILL.md` has frontmatter (name + description)
2. Verify entry in `config/skills/skill-images.yaml`
3. Check server logs for errors

**[Read full guide](INDEX.md#troubleshooting)**

### Language Mismatch

**Problem:** "skill requires language [python], got 'bash'"

**Solution:** Update `config/skills/skill-images.yaml`:

```yaml
skills:
  my-skill:
    language: bash  # Match actual container
```

**[Read full guide](SKILL_IMAGES_YAML.md#language-configuration)**

---

## 📚 Learn More

**Start here:**

- [INDEX.md](INDEX.md) - Complete documentation index
- [quick-start.md](quick-start.md) - 5-minute setup
- [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md) - Everything you need

**Configuration:**

- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) ⭐ - Critical configuration
- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - File persistence

**Development:**

- [creating-skills.md](creating-skills.md) - Build custom skills
- [SKILL_DESIGN_PRINCIPLES.md](SKILL_DESIGN_PRINCIPLES.md) - Best practices

**Architecture:**

- [overview.md](overview.md) - How it works
- [auto-loading.md](auto-loading.md) - Skill discovery
- [CONTAINERS_README.md](CONTAINERS_README.md) - Container details

---

## 🤝 Community

**Questions?** Check [INDEX.md](INDEX.md) for comprehensive docs  
**Found a bug?** Submit an issue  
**Built something cool?** Share your skill!

Skills are portable and shareable - package your skill directory and Dockerfile, then share with the community.

---

Last updated: January 20, 2026
