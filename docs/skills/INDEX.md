# Skills Documentation Index

**Welcome!** Skills extend LLM capabilities through the Model Context Protocol (MCP). Based on comprehensive testing, **3 models work with any skill type** (DeepSeek, Haiku 4.5, GPT-5 mini), while others have limitations or are incompatible. [See test report](COMPREHENSIVE_MODEL_TESTING_REPORT.md).

---

## 🚀 Quick Start (5 Minutes)

**Just want to get it working?**

```bash
# 1. Build containers
cd docker/skills && ./build-skills-images.sh

# 2. Configure outputs directory
mkdir -p /tmp/mcp-outputs

# Add to config/settings.yaml:
echo "skills:
  outputs_dir: \"/tmp/mcp-outputs\"" >> config/settings.yaml

# 3. Test it
./mcp-cli chat --servers skills
# Ask: "Create a PowerPoint about quantum computing"
```

**Works?** ✅ You're done! [Read complete guide](COMPLETE_GUIDE.md) to learn more.

**Issues?** See [Troubleshooting](#troubleshooting) below.

---

## 📚 Documentation by Experience Level

### New Users (Never Used Skills)

Start here in order:

1. **[README.md](README.md)** - What are skills and why use them?
2. **[quick-start.md](quick-start.md)** - Get working in 5 minutes
3. **[COMPLETE_GUIDE.md](COMPLETE_GUIDE.md)** ⭐ - Everything from basics to advanced

### Experienced Users

Pick what you need:

- **[SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md)** ⭐ - Configure skills (language, resources, containers)
- **[creating-skills.md](creating-skills.md)** - Build custom skills
- **[OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md)** - File persistence explained

### Reference & Deep Dives

- **[auto-loading.md](auto-loading.md)** - Skill discovery mechanism
- **[SKILL_DESIGN_PRINCIPLES.md](SKILL_DESIGN_PRINCIPLES.md)** - Design guidelines
- **[CONTAINERS_README.md](CONTAINERS_README.md)** - Container architecture
- **[quick-reference.md](quick-reference.md)** - Quick lookup table

---

## 📖 Documentation by Task

### "I want to use skills with my LLM"

1. [quick-start.md](quick-start.md) - Setup in 5 minutes
2. [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#using-skills) - Usage patterns

**Using with specific tools:**

- **Claude Desktop / VS Code / Other MCP clients:**
  
  ```bash
  ./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
  ```
  
  Then add to your client's MCP config. [Details here](COMPLETE_GUIDE.md#as-mcp-server)

- **mcp-cli chat mode:**
  
  ```bash
  ./mcp-cli chat --servers skills
  ```

- **mcp-cli workflows:**
  
  ```yaml
  execution:
    servers: [skills]
  steps:
    - skills: [docx, pptx]
  ```
  
  [Full workflow guide](COMPLETE_GUIDE.md#in-workflow-mode)

### "I want to create my own skill"

1. [creating-skills.md](creating-skills.md) - Step-by-step guide
2. [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) - Register your skill
3. [SKILL_DESIGN_PRINCIPLES.md](SKILL_DESIGN_PRINCIPLES.md) - Best practices
4. [SKILL_WRITING_CHECKLIST.md](SKILL_WRITING_CHECKLIST.md) - Quality checklist

### "I want to configure skill behavior"

1. [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) ⭐ - **Critical file explained**
   
   - Map skills to containers
   - Declare language capabilities
   - Set resource limits
   - Control security

2. [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - Where files go

3. [CONTAINERS_README.md](CONTAINERS_README.md) - Container details

### "I need help with..."

**Files not appearing:**

- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - File persistence
- [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#troubleshooting)

**Skill not loading:**

- [auto-loading.md](auto-loading.md) - Discovery mechanism  
- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md#validation)

**Language/container mismatch:**

- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md#language-configuration)

**Container build failures:**

- Check `docker/skills/build-skills-images.sh`
- [CONTAINERS_README.md](CONTAINERS_README.md)

**Performance tuning:**

- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md#resource-limits)

---

## 🔑 Critical Concepts

### 1. skill-images.yaml is Critical

**Location:** `config/skills/skill-images.yaml`

This file maps skills to containers and advertises capabilities via MCP:

```yaml
skills:
  docx:
    image: mcp-skills-docx
    language: python          # ← Advertised to LLMs
    description: "Word docs"
```

**Without it:** Skills use default Python container (may fail)  
**With it:** LLMs know skill capabilities, proper containers used

📖 **[Read full guide](SKILL_IMAGES_YAML.md)**

### 2. File Persistence

Files ONLY persist when saved to `/outputs/`:

```python
# ✅ CORRECT - persists to host
doc.save('/outputs/report.docx')

# ❌ WRONG - deleted when container exits  
doc.save('report.docx')
```

📖 **[Read full guide](OUTPUTS_DIRECTORY.md)**

### 3. Model Compatibility (Tested)

**Language-agnostic models (work with any skill):**

- ✅ DeepSeek Chat
- ✅ Claude Haiku 4.5
- ✅ GPT-5 mini

**Python-only models:**

- ⚠️ Gemini 2.0 Flash Exp (forces Python).  

**Incompatible models:**

- ❌ GPT-4o-mini, Kimi, Gemini Lite variants

**[Full compatibility testing](COMPREHENSIVE_MODEL_TESTING_REPORT.md)**

---

## 📦 Available Skills (Default)

| Skill    | Language | Description                    | Use Case                             |
| -------- | -------- | ------------------------------ | ------------------------------------ |
| **docx** | Python   | Word document creation/editing | Reports, letters, documentation      |
| **pptx** | Python   | PowerPoint presentations       | Slides, presentations                |
| **xlsx** | Python   | Excel spreadsheets             | Data analysis, tables, charts        |
| **pdf**  | Python   | PDF manipulation, forms, OCR   | Extract text, fill forms, merge PDFs |

**Add your own:** [Creating Skills Guide](creating-skills.md)

---

## 🏗️ Architecture (Quick Overview)

```
LLM (any MCP client)
    ↓ MCP Protocol
mcp-cli Skills Server
    ↓ skill-images.yaml config
Docker/Podman Container (isolated)
    ├── /workspace (temporary)
    ├── /outputs (persists to host)
    └── /skill (read-only libraries)
    ↓
Files saved to host outputs_dir
```

**Deep dive:** [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#understanding-skills)

---

## 🎯 Common Use Cases

### Use Case 1: Document Generation

```yaml
# workflow.yaml
steps:
  - name: create_report
    skills: [docx]
    run: Create monthly status report
```

**Learn more:** [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#in-workflow-mode)

### Use Case 2: Data Analysis

```yaml
steps:
  - name: analyze
    skills: [xlsx, pptx]
    run: |
      1. Read data.xlsx
      2. Analyze trends
      3. Create presentation
```

### Use Case 3: PDF Processing

```yaml
steps:
  - name: extract_and_summarize
    skills: [pdf, docx]
    run: |
      1. Extract text from PDFs
      2. Summarize findings
      3. Create Word report
```

### Use Case 4: As MCP Server

```bash
# Serve to Claude Desktop, VS Code, etc.
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

Add to client config:

```json
{
  "mcpServers": {
    "skills": {
      "command": "/path/to/mcp-cli",
      "args": ["serve", "/path/to/config/runasMCP/mcp_skills_stdio.yaml"]
    }
  }
}
```

---

## 🔧 Configuration Files

### settings.yaml (Global Config)

```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"  # Where files persist
```

**Learn more:** [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md)

### skill-images.yaml (Skill Registry)

```yaml
defaults:
  language: python
  memory: 256MB
  network_mode: none

skills:
  my-skill:
    image: my-container
    language: python
    description: "My custom skill"
```

**Learn more:** [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) ⭐

### Skill Directory Structure

```
config/skills/my-skill/
├── SKILL.md           # Documentation (LLM reads this)
├── scripts/           # Helper libraries
│   ├── __init__.py
│   └── helpers.py
└── examples/          # Usage examples
```

**Learn more:** [creating-skills.md](creating-skills.md)

---

## 🐛 Troubleshooting

### Files not appearing in outputs directory

**Symptom:** LLM creates file but it's not on disk

**Solutions:**

1. Check code saves to `/outputs/`:
   
   ```python
   # ✅ Correct
   doc.save('/outputs/file.docx')
   ```

2. Verify config:
   
   ```yaml
   # config/settings.yaml
   skills:
     outputs_dir: "/tmp/mcp-outputs"
   ```

3. Create directory:
   
   ```bash
   mkdir -p /tmp/mcp-outputs
   chmod 755 /tmp/mcp-outputs
   ```

**Full guide:** [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md)

### Skill not loading

**Symptom:** Skill doesn't appear in tools list

**Solutions:**

1. Check SKILL.md has frontmatter:
   
   ```markdown
   ---
   name: my-skill
   description: "What it does"
   ---
   ```

2. Verify skill-images.yaml entry:
   
   ```yaml
   skills:
     my-skill:
       image: my-container
       language: python
   ```

3. Check logs:
   
   ```bash
   ./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
   # Look for: "✅ Loaded skill 'my-skill'"
   ```

**Full guide:** [auto-loading.md](auto-loading.md)

### Language/container mismatch

**Symptom:** Error like "skill requires language to be [python], got 'bash'"

**Solution:**

Update skill-images.yaml:

```yaml
skills:
  my-skill:
    language: bash  # Match actual container capability
```

**Full guide:** [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md#language-configuration)

### Container build failures

**Solutions:**

1. Check Dockerfile exists:
   
   ```bash
   ls docker/skills/Dockerfile.my-skill
   ```

2. Verify build script updated:
   
   ```bash
   grep my-skill docker/skills/build-skills-images.sh
   ```

3. Check Docker/Podman running:
   
   ```bash
   docker ps
   ```

**Full guide:** [CONTAINERS_README.md](CONTAINERS_README.md)

---

## 📋 Documentation Maintenance

### For Documentation Writers

**Structure:**

- One main guide: `COMPLETE_GUIDE.md`
- Topic-specific deep dives: `SKILL_IMAGES_YAML.md`, `OUTPUTS_DIRECTORY.md`
- Quick references: `quick-start.md`, `quick-reference.md`

**Standards:**

- Start with working examples
- Explain "why" not just "how"
- Include troubleshooting sections
- Keep examples copy-pasteable
- Update this INDEX when adding docs

**Last major update:** January 20, 2026

---

## 📚 All Documentation Files

### Getting Started

- [README.md](README.md) - Overview
- [quick-start.md](quick-start.md) - 5-minute setup
- [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md) ⭐ - Complete guide

### Configuration

- [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md) ⭐ - Skill configuration reference
- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - File persistence

### Development

- [creating-skills.md](creating-skills.md) - Build custom skills
- [SKILL_DESIGN_PRINCIPLES.md](SKILL_DESIGN_PRINCIPLES.md) - Design patterns
- [SKILL_WRITING_CHECKLIST.md](SKILL_WRITING_CHECKLIST.md) - Quality checklist

### Architecture & Concepts

- [auto-loading.md](auto-loading.md) - Skill discovery
- [CONTAINERS_README.md](CONTAINERS_README.md) - Container architecture

### Reference

- [quick-reference.md](quick-reference.md) - Quick lookup
- [COMPREHENSIVE_MODEL_TESTING_REPORT.md](COMPREHENSIVE_MODEL_TESTING_REPORT.md) - LLM compatibility testing

---

## 🎯 Next Steps

**New to skills?**
→ [quick-start.md](quick-start.md) (5 minutes)
→ [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md) (full understanding)

**Want to create skills?**
→ [creating-skills.md](creating-skills.md)
→ [SKILL_IMAGES_YAML.md](SKILL_IMAGES_YAML.md)

**Need help?**
→ Check [Troubleshooting](#troubleshooting) above
→ Read [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#troubleshooting)

---

Last updated: January 20, 2026
