# Skills Complete Guide

**mcp-cli Skills System - From Zero to Production**

Last updated: January 20, 2026

---

## Table of Contents

1. [Quick Start](#quick-start) - Get working in 5 minutes
2. [Understanding Skills](#understanding-skills) - Core concepts
3. [Configuration](#configuration) - skill-images.yaml explained
4. [Creating Skills](#creating-skills) - Build your own
5. [Using Skills](#using-skills) - Integration patterns
6. [Troubleshooting](#troubleshooting) - Common issues
7. [Reference](#reference) - Complete API

---

## Quick Start

**Get skills running in 5 minutes:**

### 1. Build Container Images

```bash
cd docker/skills
./build-skills-images.sh
```

This builds the base skill containers (docx, pptx, xlsx, pdf).

### 2. Configure Outputs Directory

Edit `config/settings.yaml`:

```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"
```

Create the directory:

```bash
mkdir -p /tmp/mcp-outputs
chmod 755 /tmp/mcp-outputs
```

### 3. Test with mcp-cli

**Chat mode:**

```bash
./mcp-cli chat --servers skills
```

Ask: "Create a PowerPoint with 3 slides about Python"

**Workflow mode:**

```bash
./mcp-cli --workflow your_workflow
```

Workflows can use skills via `servers: [skills]` and `skills: [pptx]`.

### 4. Use as MCP Server

**For Claude Desktop, VS Code, or other MCP clients:**

```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

Add to your MCP client config (e.g., `claude_desktop_config.json`):

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

**That's it!** Skills are now available to MCP-compatible LLMs. [See compatibility table](COMPREHENSIVE_MODEL_TESTING_REPORT.md#complete-results-matrix) for which models work.

---

## Understanding Skills

### What Are Skills?

Skills are **containerized packages** that extend LLM capabilities through the Model Context Protocol (MCP). They provide:

1. **Specialized knowledge** - Documentation the LLM can read
2. **Helper libraries** - Pre-written, tested code
3. **Safe execution** - Isolated containers, no network access
4. **Selective LLM compatibility** - Works with capable models ([see test report](COMPREHENSIVE_MODEL_TESTING_REPORT.md))

### How Skills Work

```
┌─────────────────────────────────────────────────────────┐
│  LLM (GPT-4, Claude, DeepSeek, Gemini, etc.)          │
└─────────────────────┬───────────────────────────────────┘
                      │ MCP Protocol
                      ↓
┌─────────────────────────────────────────────────────────┐
│  mcp-cli Skills Server                                  │
│  • Advertises available skills & their languages       │
│  • Routes requests to containers                        │
│  • Manages file persistence                             │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ↓
┌─────────────────────────────────────────────────────────┐
│  Docker/Podman Container (isolated)                     │
│  ┌───────────────────────────────────────────────────┐ │
│  │  /workspace/  (temporary)                         │ │
│  │  /outputs/    (mounted from host - PERSISTS)      │ │
│  │  /skill/      (read-only skill libraries)         │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                      │
                      ↓
              Files saved to host
         /tmp/mcp-outputs/yourfile.docx
```

### Key Concepts

**1. Skills Directory** (`config/skills/`)

Each skill is a folder with:

- `SKILL.md` - Documentation the LLM reads
- `scripts/` - Helper libraries (optional)
- `examples/` - Usage patterns (optional)

**2. skill-images.yaml** ⭐ **CRITICAL**

Maps skills to container images and declares their capabilities:

```yaml
skills:
  docx:
    image: mcp-skills-docx
    language: python        # ← Advertised to LLM
    description: "Word document manipulation"
```

Without this file, skills use default Python container.

**3. File Persistence**

ONLY files written to `/outputs/` persist:

```python
# ✅ CORRECT - File persists to host
doc.save('/outputs/report.docx')

# ❌ WRONG - File deleted when container exits
doc.save('report.docx')  # Defaults to /workspace/
```

**4. Two Modes**

**Passive Mode** - LLM reads skill documentation:

```json
{
  "tool": "docx",
  "arguments": {"mode": "passive"}
}
```

**Active Mode** - LLM executes code with skill libraries:

```json
{
  "tool": "execute_skill_code",
  "arguments": {
    "skill_name": "docx",
    "code": "from docx import Document...",
    "language": "python"
  }
}
```

---

## Configuration

### skill-images.yaml - The Critical Role

**Location:** `config/skills/skill-images.yaml`

This file is the **central registry** for all skill configurations. It:

1. **Maps skill names to container images**
2. **Declares language capabilities** (advertised via MCP)
3. **Sets resource limits** (memory, CPU, timeout)
4. **Controls security** (network access, mounts)

**Without it:** Skills use default Python container and may fail if they need bash or special packages.

**With it:** Skills have correct environments and LLMs know their capabilities.

### File Structure

```yaml
# Global defaults (inherited by all skills)
defaults:
  image: python:3.11-slim
  language: python           # Default language
  network_mode: none         # Isolated (secure)
  memory: 256MB
  cpu: "0.5"
  timeout: 60s
  outputs_dir: /tmp/mcp-outputs

# Per-skill configuration (overrides defaults)
skills:

  # Python skill example
  docx:
    image: mcp-skills-docx
    language: python
    description: "Word document manipulation"
    dockerfile: docker/skills/Dockerfile.docx

  # Bash skill example  
  odt-parser:
    image: mcp-skills-document-parsing
    language: bash           # ← Advertised as bash skill
    description: "Parse ODT to XML"

  # Multi-language skill
  data-processor:
    image: mcp-skills-advanced
    languages: [python, bash]  # LLM chooses
    description: "Data processing"

  # High-resource skill
  pdf:
    image: mcp-skills-pdf
    language: python
    memory: 512MB            # More memory
    timeout: 120s            # Longer timeout
```

### Configuration Fields

#### Defaults Section

| Field          | Type   | Default            | Description                          |
| -------------- | ------ | ------------------ | ------------------------------------ |
| `language`     | string | `python`           | Default language for all skills      |
| `image`        | string | `python:3.11-slim` | Default container image              |
| `network_mode` | string | `none`             | Network isolation (none/bridge/host) |
| `memory`       | string | `256MB`            | Memory limit per container           |
| `cpu`          | string | `0.5`              | CPU cores limit                      |
| `timeout`      | string | `60s`              | Execution timeout                    |
| `outputs_dir`  | string | `/tmp/mcp-outputs` | Host directory for output files      |

#### Per-Skill Fields

| Field          | Type   | Required    | Description                          |
| -------------- | ------ | ----------- | ------------------------------------ |
| `image`        | string | ✅ Yes       | Container image name                 |
| `language`     | string | Recommended | Single language (python/bash)        |
| `languages`    | array  | Optional    | Multiple languages [python, bash]    |
| `description`  | string | Optional    | Brief description                    |
| `dockerfile`   | string | Optional    | Path to Dockerfile for auto-building |
| `network_mode` | string | Optional    | Override default network mode        |
| `memory`       | string | Optional    | Override memory limit                |
| `cpu`          | string | Optional    | Override CPU limit                   |
| `timeout`      | string | Optional    | Override timeout                     |
| `mounts`       | array  | Optional    | Additional volume mounts             |
| `environment`  | array  | Optional    | Environment variables                |

### Language Configuration - How It Works

The `language` field is **critical** for MCP advertising:

**1. Single Language Skill**

```yaml
docx:
  language: python  # Advertised as Python-only
```

When LLM calls `execute_skill_code`:

- If no `language` param → auto-populated to `python`
- If `language: bash` param → **ERROR** (skill requires python)

**2. Multi-Language Skill**

```yaml
data-processor:
  languages: [python, bash]  # Supports both
```

When LLM calls `execute_skill_code`:

- LLM **must** specify `language` parameter
- Both `python` and `bash` are valid

**3. No Language Specified**

```yaml
custom-skill:
  image: my-image
  # No language field
```

Inherits `defaults.language` (usually `python`).

**Why This Matters:**

Different LLMs have different language preferences:
**Tested working models:**

- **DeepSeek, Claude Haiku 4.5, GPT-5 mini** → Language-agnostic (work with both Python and bash)
- **Gemini 2.0 Flash Exp** → Python-only (forces Python even for bash skills)

**Incompatible models:** GPT-4o-mini, Kimi, Gemini Lite variants

By advertising language capabilities via MCP, capable LLMs know what's supported and can choose appropriately.

**[Full compatibility testing](COMPREHENSIVE_MODEL_TESTING_REPORT.md)**

### Example Configurations

**Basic Python Skill:**

```yaml
skills:
  my-python-skill:
    image: python:3.11-slim
    language: python
    description: "Process data with Python"
```

**Bash Script Skill:**

```yaml
skills:
  xml-parser:
    image: mcp-skills-bash-tools
    language: bash
    description: "Parse XML with xmlstarlet"
    memory: 64MB        # Bash scripts are lightweight
    cpu: "0.25"
```

**Custom Container with Special Packages:**

```yaml
skills:
  ml-processor:
    image: mcp-skills-ml
    language: python
    dockerfile: docker/skills/Dockerfile.ml
    memory: 1GB         # ML needs more memory
    timeout: 300s       # 5 minutes for model loading
    description: "Machine learning data processing"
```

**Network-Enabled Skill (Use Sparingly!):**

```yaml
skills:
  api-client:
    image: mcp-skills-api
    language: python
    network_mode: bridge  # ⚠️ Allows network access
    description: "Call external APIs"
```

⚠️ **Security Note:** Only enable `network_mode: bridge` if absolutely necessary. Skills should be isolated by default.

---

## Creating Skills

### Step-by-Step Guide

#### 1. Create Skill Directory

```bash
cd config/skills
mkdir my-skill
cd my-skill
```

#### 2. Write SKILL.md

Create `SKILL.md` with required frontmatter:

```markdown
---
name: my-skill
description: "Brief description of what the skill does"
---

# My Skill

Detailed documentation the LLM will read.

## Purpose

What this skill does and when to use it.

## Available Functions

\`\`\`python
from scripts.helpers import process_data

# Process some data
result = process_data("input.txt", "output.txt")
\`\`\`

## Parameters

- `input` (str): Input file path
- `output` (str): Output file path

## Returns

- `result` (dict): Processing results

## Examples

### Example 1: Basic Usage

\`\`\`python
from scripts.helpers import process_data

# Read from /outputs/, write to /outputs/
result = process_data(
    "/outputs/input.dat",
    "/outputs/output.dat"
)

print(f"Processed {result['count']} items")
\`\`\`
```

**Frontmatter fields:**

- `name` - Skill identifier (lowercase-with-hyphens)
- `description` - Brief description (shows in tool listings)

#### 3. Add Helper Libraries (Optional)

If your skill provides reusable code:

**scripts/helpers.py:**

```python
"""Helper functions for my-skill."""

def process_data(input_path, output_path):
    """
    Process data from input file and save to output.

    Args:
        input_path (str): Path to input file (should be in /outputs/)
        output_path (str): Path to output file (must be in /outputs/)

    Returns:
        dict: Processing results with 'count' and 'status' keys

    Example:
        >>> result = process_data("/outputs/in.txt", "/outputs/out.txt")
        >>> print(result['count'])
        42
    """
    # Implementation here
    with open(input_path, 'r') as f:
        data = f.read()

    # Process data...
    processed = data.upper()

    with open(output_path, 'w') as f:
        f.write(processed)

    return {
        'count': len(processed),
        'status': 'success'
    }
```

**scripts/__init__.py:**

```python
"""My Skill package."""
from .helpers import process_data

__all__ = ['process_data']
```

#### 4. Register in skill-images.yaml

Edit `config/skills/skill-images.yaml`:

```yaml
skills:
  my-skill:
    image: python:3.11-slim  # Use default Python
    language: python
    description: "My custom skill"
```

Or for a custom container:

```yaml
skills:
  my-skill:
    image: mcp-skills-my-skill
    language: python
    dockerfile: docker/skills/Dockerfile.my-skill
    description: "My custom skill with special packages"
```

#### 5. Build Custom Container (If Needed)

If your skill needs additional packages:

**docker/skills/Dockerfile.my-skill:**

```dockerfile
FROM python:3.11-slim

# Install required packages
RUN pip install --no-cache-dir --break-system-packages \
    pandas \
    numpy \
    requests

# Set working directory
WORKDIR /workspace

# Container will run with:
# /workspace (temporary)
# /outputs (mounted from host)
# /skill (mounted skill directory)
```

**Add to build script:**

Edit `docker/skills/build-skills-images.sh`:

```bash
# Add to IMAGES array
IMAGES[my-skill]="Dockerfile.my-skill:mcp-skills-my-skill:My custom skill"
```

**Build:**

```bash
cd docker/skills
./build-skills-images.sh my-skill
```

Or build all:

```bash
./build-skills-images.sh
```

#### 6. Test Your Skill

**Via chat mode:**

```bash
./mcp-cli chat --servers skills
```

Ask: "Use my-skill to process data"

**Via workflow:**

```yaml
# test-my-skill.yaml
name: test_my_skill
version: 1.0.0

execution:
  provider: deepseek
  servers: [skills]

steps:
  - name: test
    skills: [my-skill]
    max_iterations: 3
    run: |
      Use my-skill to process test data.

      Read from: /outputs/test_input.txt
      Write to: /outputs/test_output.txt
```

Run:

```bash
echo "test data" > /tmp/mcp-outputs/test_input.txt
./mcp-cli --workflow test-my-skill
cat /tmp/mcp-outputs/test_output.txt
```

---

## Using Skills

### In Chat Mode

```bash
./mcp-cli chat --servers skills
```

Ask natural language questions:

- "Create a PowerPoint about quantum computing"
- "Read the Excel file data.xlsx and summarize it"
- "Extract text from document.pdf"

The LLM will:

1. Read skill documentation (passive mode)
2. Write code using helper libraries
3. Execute code (active mode)
4. Save results to `/outputs/`

### In Workflow Mode

**Basic usage:**

```yaml
name: create_report
version: 1.0.0

execution:
  provider: deepseek
  servers: [skills]  # Enable skills server

steps:
  - name: generate_doc
    skills: [docx]  # Limit to specific skills
    max_iterations: 5
    run: |
      Use docx skill to create a report.

      Title: Monthly Status Report
      Save to: /outputs/{{env.report_date}}/report.docx
```

**Advanced - Multiple skills:**

```yaml
steps:
  - name: process_data
    skills: [xlsx, pdf]
    run: |
      1. Use xlsx to read /outputs/data.xlsx
      2. Process the data
      3. Use pdf to create /outputs/summary.pdf
```

**With file-based state:**

```yaml
steps:
  - name: extract
    skills: [pdf]
    run: |
      Extract text from /outputs/input.pdf
      Save extracted text to /outputs/extracted.txt

  - name: analyze
    needs: [extract]
    skills: [python-context-builder]
    run: |
      Read /outputs/extracted.txt
      Analyze the text
      Save results to /outputs/analysis.json
```

### As MCP Server

Skills can be used by **any MCP-compatible client**:

**1. Start MCP server:**

```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

**2. Configure MCP client:**

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "skills": {
      "command": "/Users/you/mcp-cli-go/mcp-cli",
      "args": ["serve", "/Users/you/mcp-cli-go/config/runasMCP/mcp_skills_stdio.yaml"]
    }
  }
}
```

**VS Code with Cline** (`.vscode/settings.json`):

```json
{
  "cline.mcpServers": {
    "skills": {
      "command": "/absolute/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runasMCP/mcp_skills_stdio.yaml"]
    }
  }
}
```

**3. Use in your LLM client:**

The LLM will see available skills as MCP tools and can use them directly.

---

## Troubleshooting

### Files Not Appearing in outputs/

**Problem:** LLM creates file but it's not in `/tmp/mcp-outputs/`

**Causes & Solutions:**

1. **Wrong path in code**
   
   ```python
   # ❌ WRONG
   doc.save('report.docx')  # Goes to /workspace/
   
   # ✅ CORRECT
   doc.save('/outputs/report.docx')
   ```

2. **Outputs directory not configured**
   
   Check `config/settings.yaml`:
   
   ```yaml
   skills:
     outputs_dir: "/tmp/mcp-outputs"
   ```

3. **Directory doesn't exist**
   
   ```bash
   mkdir -p /tmp/mcp-outputs
   chmod 755 /tmp/mcp-outputs
   ```

4. **Permission issues**
   
   ```bash
   ls -ld /tmp/mcp-outputs
   # Should show: drwxr-xr-x
   
   # Fix if needed:
   chmod 755 /tmp/mcp-outputs
   ```

### Skill Not Loading

**Problem:** Skill doesn't appear in available tools

**Debug steps:**

1. **Check server logs:**
   
   ```bash
   ./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
   # Look for: "✅ Initialized skill service with N skills"
   ```

2. **Verify SKILL.md exists and has frontmatter:**
   
   ```bash
   cat config/skills/my-skill/SKILL.md | head -5
   # Should show:
   # ---
   # name: my-skill
   # description: "..."
   # ---
   ```

3. **Check skill-images.yaml:**
   
   ```bash
   grep -A 3 "my-skill:" config/skills/skill-images.yaml
   ```

4. **Verify container image exists:**
   
   ```bash
   docker images | grep mcp-skills-my-skill
   # Should show image if using custom container
   ```

### Language/Container Mismatch

**Problem:** Skill requires bash but is configured for python

**Error:**

```
skill 'odt-parser' requires language to be one of [python], got 'bash'
```

**Solution:**

Update `config/skills/skill-images.yaml`:

```yaml
skills:
  odt-parser:
    language: bash  # ← Add this
```

Or for multi-language:

```yaml
skills:
  flexible-skill:
    languages: [python, bash]  # Accepts both
```

### Container Build Failures

**Problem:** `./build-skills-images.sh` fails

**Common causes:**

1. **Missing Dockerfile:**
   
   ```bash
   ls -l docker/skills/Dockerfile.my-skill
   ```

2. **Build script not updated:**
   
   Edit `docker/skills/build-skills-images.sh`:
   
   ```bash
   IMAGES[my-skill]="Dockerfile.my-skill:mcp-skills-my-skill:My skill"
   ```

3. **Docker/Podman not running:**
   
   ```bash
   docker ps  # Should not error
   ```

### Import Errors in Skills

**Problem:** `ModuleNotFoundError` when using skill libraries

**Causes:**

1. **Missing `scripts/__init__.py`:**
   
   ```bash
   # Check it exists:
   ls config/skills/my-skill/scripts/__init__.py
   
   # Create if missing:
   touch config/skills/my-skill/scripts/__init__.py
   ```

2. **Package not installed in container:**
   
   Update Dockerfile:
   
   ```dockerfile
   RUN pip install --break-system-packages my-package
   ```
   
   Rebuild:
   
   ```bash
   cd docker/skills
   ./build-skills-images.sh my-skill
   ```

3. **Wrong import path:**
   
   ```python
   # ✅ CORRECT
   from scripts.helpers import function
   
   # ❌ WRONG
   from my-skill.scripts.helpers import function
   ```

### Network Access Issues

**Problem:** Skill needs network but can't connect

**Solution:**

⚠️ **Only if absolutely necessary:**

Update `config/skills/skill-images.yaml`:

```yaml
skills:
  api-client:
    network_mode: bridge  # Allows network
    description: "Downloads from external API"
```

**Security Note:** Most skills should work without network access. Consider:

- Pre-downloading required data
- Using local files instead of APIs
- Separating network calls from processing

---

## Reference

### Available Skills (Default)

| Skill                    | Language | Description                    | Container        |
| ------------------------ | -------- | ------------------------------ | ---------------- |
| `docx`                   | python   | Word document creation/editing | mcp-skills-docx  |
| `pptx`                   | python   | PowerPoint presentations       | mcp-skills-pptx  |
| `xlsx`                   | python   | Excel spreadsheets             | mcp-skills-xlsx  |
| `pdf`                    | python   | PDF manipulation, forms, OCR   | mcp-skills-pdf   |
| `python-context-builder` | python   | Multi-hop context building     | python:3.11-slim |

### MCP Tools Exposed

#### Individual Skill Tools

Each skill exposes a passive mode tool:

```json
{
  "name": "docx",
  "description": "Word document manipulation...",
  "input_schema": {
    "type": "object",
    "properties": {
      "mode": {
        "type": "string",
        "enum": ["passive"],
        "default": "passive"
      }
    }
  }
}
```

#### execute_skill_code Tool

Universal execution tool for all skills:

```json
{
  "name": "execute_skill_code",
  "description": "Execute code with skill libraries",
  "input_schema": {
    "type": "object",
    "properties": {
      "skill_name": {
        "type": "string",
        "description": "Name of skill to use (e.g., 'docx', 'pptx')"
      },
      "code": {
        "type": "string",
        "description": "Code to execute. MUST save files to /outputs/"
      },
      "language": {
        "type": "string",
        "enum": ["python", "bash"],
        "description": "Auto-populated from skill config. Only specify if skill supports multiple languages."
      }
    },
    "required": ["skill_name", "code"]
  }
}
```

### Directory Structure

```
mcp-cli-go/
├── config/
│   ├── settings.yaml          # Global config
│   └── skills/
│       ├── skill-images.yaml  # ⭐ Skill configurations
│       ├── docx/
│       │   ├── SKILL.md
│       │   ├── scripts/
│       │   └── examples/
│       ├── pptx/
│       └── [other skills]/
├── docker/
│   └── skills/
│       ├── build-skills-images.sh
│       ├── Dockerfile.docx
│       ├── Dockerfile.pptx
│       └── [other Dockerfiles]
└── /tmp/mcp-outputs/          # Output files (configured location)
```

### Environment Variables

Skills can access these environment variables:

- `OUTPUTS_DIR` - Path to outputs directory (`/outputs`)
- `SKILL_NAME` - Name of the skill being executed
- `WORKSPACE_DIR` - Temporary workspace (`/workspace`)

### Container Mounts

Each skill container has:

| Mount Point  | Source               | Permissions | Purpose                        |
| ------------ | -------------------- | ----------- | ------------------------------ |
| `/outputs`   | Host `outputs_dir`   | Read/Write  | **Persistent** file storage    |
| `/workspace` | Container tmpfs      | Read/Write  | Temporary work (deleted after) |
| `/skill`     | Host skill directory | Read-Only   | Skill scripts and libraries    |

---

## Best Practices

### For Skill Creators

1. **Clear Documentation**
   
   - Write SKILL.md as if explaining to a junior developer
   - Include complete examples with expected outputs
   - Document all function parameters and return values

2. **Helper Libraries**
   
   - Keep functions focused and reusable
   - Use type hints and docstrings
   - Provide sensible defaults

3. **File Paths**
   
   - Always use `/outputs/` for persistent files
   - Document expected input/output paths
   - Validate paths in helper functions

4. **Testing**
   
   - Test with working models (DeepSeek, Haiku 4.5, or GPT-5 mini)
   - Verify files persist to host
   - Check memory and CPU usage

5. **Security**
   
   - Default to `network_mode: none`
   - Minimize container image size
   - Use read-only mounts where possible

### For Skill Users

1. **File Persistence**
   
   - Always save to `/outputs/`
   - Use absolute paths
   - Check outputs directory configuration

2. **Resource Management**
   
   - Set appropriate timeouts for long-running tasks
   - Monitor memory usage for large files
   - Use parallel workflows when possible

3. **Error Handling**
   
   - Check skill documentation first
   - Look at examples
   - Review container logs if issues persist

4. **Model Compatibility**
   
   - Only 3 models work with both Python and bash: DeepSeek, Haiku 4.5, GPT-5 mini
   - Gemini 2.0 Flash Exp: Python only
   - Many models are incompatible (see test report)
   - Configure `language` field to advertise skill capabilities

---

## Advanced Topics

### Multi-Language Skills

Some skills support both Python and bash:

```yaml
skills:
  data-processor:
    image: mcp-skills-dual
    languages: [python, bash]
    description: "Process data with Python or bash"
```

LLM must specify language:

```json
{
  "tool": "execute_skill_code",
  "arguments": {
    "skill_name": "data-processor",
    "language": "bash",
    "code": "#!/bin/bash\ncat /outputs/data.txt | sort > /outputs/sorted.txt"
  }
}
```

### Custom Resource Limits

For resource-intensive skills:

```yaml
skills:
  video-processor:
    image: mcp-skills-video
    language: python
    memory: 2GB
    cpu: "2.0"
    timeout: 600s  # 10 minutes
    description: "Process video files"
```

### Network-Enabled Skills

⚠️ Use sparingly:

```yaml
skills:
  web-scraper:
    image: mcp-skills-scraper
    language: python
    network_mode: bridge
    description: "Download web content"
```

### Skill Composition

Skills can call other skills via workflows:

```yaml
steps:
  - name: extract
    skills: [pdf]
    run: Extract text from PDF

  - name: analyze
    needs: [extract]
    skills: [python-context-builder]
    run: Analyze extracted text

  - name: report
    needs: [analyze]
    skills: [docx]
    run: Create Word report
```

---

## FAQ

**Q: Do I need Docker or Podman?**
A: Yes, skills run in containers for security and isolation.

**Q: Can I use skills without containers?**
A: No, containerization is core to the security model.

**Q: Do skills work with all LLMs?**
A: No. Based on testing, only **DeepSeek, Claude Haiku 4.5, and GPT-5 mini** work with both Python and bash skills. Gemini 2.0 Flash Exp works with Python only. Many models (GPT-4o-mini, Kimi, Gemini Lite) are incompatible. [See test report](COMPREHENSIVE_MODEL_TESTING_REPORT.md) for details.

**Q: Can skills access the internet?**
A: By default no (`network_mode: none`). You can enable it per-skill but it's discouraged for security.

**Q: Where do files go?**
A: Files saved to `/outputs/` go to the configured `outputs_dir` on the host (e.g., `/tmp/mcp-outputs/`).

**Q: Can I create skills in other languages?**
A: Currently Python and bash are supported. Other languages can be added via custom containers.

**Q: How do I update a skill?**
A: Edit the files in `config/skills/skill-name/` and restart the MCP server.

**Q: Can skills be private?**
A: Yes, skills in your `config/skills/` directory are only accessible to your mcp-cli instance.

**Q: Can I share skills?**
A: Yes! Package the skill directory and Dockerfile. Others can drop it in their `config/skills/` folder.

---

## Additional Resources

- [Workflow Guide](../workflows/README.md) - Using skills in workflows
- [MCP Specification](https://modelcontextprotocol.io) - Protocol details
- [Docker Documentation](https://docs.docker.com) - Container basics
- [Example Skills](../../config/skills/) - See working examples

---

## Contributing

Have a useful skill? Consider contributing it!

1. Create skill following best practices
2. Test with multiple LLMs
3. Document thoroughly
4. Submit PR with:
   - Skill directory
   - Dockerfile (if custom)
   - Tests
   - Documentation

---

**Last updated:** January 20, 2026  
**Version:** 1.0.0  
**Maintainers:** mcp-cli-go community
