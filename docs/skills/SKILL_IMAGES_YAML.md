# skill-images.yaml Reference Guide

**Complete reference for configuring skills via skill-images.yaml**

Last updated: January 20, 2026

---

## Overview

`skill-images.yaml` is the **central configuration file** for the mcp-cli skills system. It maps skill names to container images and declares their capabilities to the MCP protocol.

**Location:** `config/skills/skill-images.yaml`

**Critical role:**
1. Maps skills → container images
2. Declares language capabilities (advertised via MCP)
3. Sets resource limits and security policies
4. Enables cross-LLM compatibility

Without this file, all skills use default Python container and may fail.

---

## Quick Reference

### Minimal Configuration

```yaml
skills:
  my-skill:
    image: python:3.11-slim
    language: python
```

### Full Configuration Example

```yaml
defaults:
  image: python:3.11-slim
  language: python
  network_mode: none
  memory: 256MB
  cpu: "0.5"
  timeout: 60s
  outputs_dir: /tmp/mcp-outputs

skills:
  my-skill:
    image: mcp-skills-custom
    language: python
    description: "My custom skill"
    dockerfile: docker/skills/Dockerfile.custom
    network_mode: none
    memory: 512MB
    cpu: "1.0"
    timeout: 120s
    mounts:
      - "/host/data:/data:ro"
    environment:
      - "API_KEY=secret"
```

---

## File Structure

### Top Level

```yaml
defaults:       # Global defaults for all skills
  [fields]

skills:         # Per-skill configurations
  skill-name:
    [fields]
```

### Defaults Section

All fields in `defaults` are inherited by every skill unless explicitly overridden.

**Available fields:**

```yaml
defaults:
  image: python:3.11-slim      # Default container image
  language: python              # Default language
  network_mode: none            # Network isolation
  memory: 256MB                 # Memory limit
  cpu: "0.5"                    # CPU cores
  timeout: 60s                  # Execution timeout
  outputs_dir: /tmp/mcp-outputs # Output directory
```

### Skills Section

Maps each skill name to its configuration. Inherits from `defaults`.

**Minimal skill:**

```yaml
skills:
  my-skill:
    image: my-image  # Only required field
```

**Typical skill:**

```yaml
skills:
  docx:
    image: mcp-skills-docx
    language: python
    description: "Word document manipulation"
    dockerfile: docker/skills/Dockerfile.docx
```

---

## Field Reference

### Required Fields

#### `image` (required)

Container image name for the skill.

**Type:** string  
**Default:** Inherits from `defaults.image`  
**Required in:** Per-skill config (required) or defaults

**Examples:**

```yaml
# Use default Python image
skills:
  my-skill:
    image: python:3.11-slim

# Use custom built image
skills:
  docx:
    image: mcp-skills-docx

# Use specific version
skills:
  pandas-skill:
    image: pandas/pandas:2.0.0
```

**Notes:**
- Image must exist before starting MCP server
- Use `docker images` to list available images
- Build custom images with `./build-skills-images.sh`

---

### Language Configuration

#### `language` (recommended)

Declares the programming language required by the skill.

**Type:** string  
**Values:** `"python"` | `"bash"`  
**Default:** Inherits from `defaults.language` (usually `"python"`)

**Examples:**

```yaml
# Python skill
skills:
  docx:
    language: python

# Bash skill
skills:
  xml-parser:
    language: bash

# Inherit default (python)
skills:
  data-processor:
    image: python:3.11-slim
    # language: python (inherited)
```

**How it works:**

1. **MCP Advertisement:**
   - Language is advertised to LLM clients via MCP protocol
   - LLM knows which language to use when calling `execute_skill_code`

2. **Auto-population:**
   - If LLM doesn't specify `language` parameter, it's auto-filled from config
   - Reduces token usage and LLM decision overhead

3. **Validation:**
   - If LLM specifies language, it's validated against skill's declared language
   - Prevents mismatches (e.g., trying to run bash in Python-only container)

**Example error if mismatched:**

```
Error: skill 'docx' requires language to be one of [python], got 'bash'
```

#### `languages` (optional)

For skills that support multiple languages.

**Type:** array of strings  
**Values:** `["python", "bash"]`  
**Conflicts with:** `language` (use one or the other)

**Example:**

```yaml
skills:
  flexible-processor:
    image: mcp-skills-dual
    languages: [python, bash]
    description: "Supports both Python and bash"
```

**Behavior:**
- LLM **must** specify `language` parameter when calling
- Both `python` and `bash` are accepted
- No auto-population (LLM chooses)

**Use cases:**
- Skills with both Python and bash implementations
- Containers with both interpreters installed
- Allowing LLM to choose based on task

---

### Resource Limits

#### `memory`

Memory limit for skill container.

**Type:** string  
**Format:** `"<number><unit>"` where unit is `MB` or `GB`  
**Default:** `"256MB"`

**Examples:**

```yaml
# Standard skill
skills:
  docx:
    memory: 256MB

# High memory skill (ML, video)
skills:
  ml-processor:
    memory: 2GB

# Low memory skill (bash scripts)
skills:
  text-processor:
    memory: 64MB
```

**Guidelines:**
- **64MB:** Simple bash scripts
- **256MB:** Standard Python skills
- **512MB:** PDF processing, image manipulation
- **1GB+:** ML models, video processing, large datasets

#### `cpu`

CPU limit (number of cores).

**Type:** string (quoted number)  
**Format:** `"<decimal>"` (e.g., `"0.5"`, `"2.0"`)  
**Default:** `"0.5"`

**Examples:**

```yaml
# Light processing
skills:
  simple-task:
    cpu: "0.25"

# Standard
skills:
  docx:
    cpu: "0.5"

# Heavy processing
skills:
  video-encoder:
    cpu: "2.0"
```

**Guidelines:**
- **0.25:** Lightweight bash scripts
- **0.5:** Standard Python processing
- **1.0:** CPU-intensive tasks (parsing, encoding)
- **2.0+:** Parallel processing, ML inference

#### `timeout`

Maximum execution time.

**Type:** string  
**Format:** `"<number><unit>"` where unit is `s`, `m`, or `h`  
**Default:** `"60s"`

**Examples:**

```yaml
# Quick tasks
skills:
  formatter:
    timeout: 30s

# Standard
skills:
  docx:
    timeout: 60s

# Long-running
skills:
  video-processor:
    timeout: 10m

# Very long
skills:
  ml-trainer:
    timeout: 1h
```

**Guidelines:**
- **30s:** Simple formatting, quick lookups
- **60s:** Standard document processing
- **2-5m:** Large file processing, complex operations
- **10m+:** Video encoding, ML inference
- **1h+:** Training, batch processing

---

### Security Settings

#### `network_mode`

Controls network access for the container.

**Type:** string  
**Values:** `"none"` | `"bridge"` | `"host"`  
**Default:** `"none"` (recommended)

**Options:**

**`none` (recommended):**
- No network access
- Maximum security
- Skills are isolated from internet

```yaml
skills:
  docx:
    network_mode: none  # Default, most secure
```

**`bridge` (use sparingly):**
- Container can access network
- Use only if absolutely necessary
- Document why in `network_justification`

```yaml
skills:
  api-client:
    network_mode: bridge
    network_justification: "Downloads data from external API"
```

**`host` (rarely used):**
- Container uses host network stack
- Not recommended for skills
- May be needed for specific network debugging

**Security note:**

⚠️ **Always use `none` unless you have a specific, documented reason for network access.**

Skills should:
- ✅ Process local files
- ✅ Use pre-downloaded data
- ✅ Work offline
- ❌ Download from internet
- ❌ Call external APIs
- ❌ Send data externally

#### `network_justification`

Required when using `network_mode: bridge` or `network_mode: host`.

**Type:** string  
**Purpose:** Documents why network access is needed

**Example:**

```yaml
skills:
  weather-api:
    network_mode: bridge
    network_justification: "Fetches real-time weather data from external API. No sensitive data transmitted."
```

---

### Build Configuration

#### `dockerfile`

Path to Dockerfile for building custom container.

**Type:** string  
**Format:** Relative path from project root

**Example:**

```yaml
skills:
  my-skill:
    image: mcp-skills-my-skill
    dockerfile: docker/skills/Dockerfile.my-skill
```

**Usage:**

1. **Create Dockerfile:**

```dockerfile
# docker/skills/Dockerfile.my-skill
FROM python:3.11-slim

RUN pip install --no-cache-dir --break-system-packages \
    pandas \
    numpy

WORKDIR /workspace
```

2. **Add to build script:**

Edit `docker/skills/build-skills-images.sh`:

```bash
IMAGES[my-skill]="Dockerfile.my-skill:mcp-skills-my-skill:My custom skill"
```

3. **Build:**

```bash
cd docker/skills
./build-skills-images.sh my-skill
```

4. **Reference in config:**

```yaml
skills:
  my-skill:
    image: mcp-skills-my-skill
    dockerfile: docker/skills/Dockerfile.my-skill
```

#### `description`

Brief description of what the skill does.

**Type:** string  
**Length:** Recommended <100 characters  
**Purpose:** Shows in tool listings, helps LLMs choose skills

**Examples:**

```yaml
skills:
  docx:
    description: "Word document creation and manipulation via OOXML"
  
  pdf:
    description: "PDF manipulation, forms filling, text extraction, OCR"
  
  xml-parser:
    description: "Parse and analyze XML documents with xmlstarlet"
```

**Best practices:**
- Be specific about capabilities
- Mention key features
- Keep concise
- Use present tense

---

### Advanced Configuration

#### `mounts`

Additional volume mounts beyond the standard `/outputs`, `/workspace`, `/skill`.

**Type:** array of strings  
**Format:** `["<host-path>:<container-path>:<options>"]`  
**Options:** `ro` (read-only) | `rw` (read-write)

**Example:**

```yaml
skills:
  data-analyzer:
    image: mcp-skills-analyzer
    mounts:
      - "/data/shared:/data:ro"           # Read-only data
      - "/cache/models:/models:ro"         # Pre-loaded models
      - "/tmp/working:/working:rw"         # Additional workspace
```

**Standard mounts (always present):**
- `/outputs` ← Host `outputs_dir` (read-write)
- `/workspace` ← Temporary tmpfs (read-write)
- `/skill` ← Skill directory (read-only)

**Use cases:**
- Sharing large datasets across skills
- Pre-loaded ML models
- Cached resources
- Additional temporary space

#### `environment`

Environment variables for the container.

**Type:** array of strings  
**Format:** `["VAR=value", "VAR2=value2"]`

**Example:**

```yaml
skills:
  ml-processor:
    image: mcp-skills-ml
    environment:
      - "MODEL_PATH=/models/default"
      - "CACHE_DIR=/cache"
      - "LOG_LEVEL=info"
```

**Standard environment variables (always set):**
- `OUTPUTS_DIR=/outputs`
- `SKILL_NAME=<skill-name>`
- `WORKSPACE_DIR=/workspace`

**Use cases:**
- Configuring library behavior
- Setting paths
- Debug flags
- Feature toggles

---

## Complete Examples

### Example 1: Basic Python Skill

```yaml
defaults:
  image: python:3.11-slim
  language: python
  network_mode: none
  memory: 256MB
  cpu: "0.5"
  timeout: 60s

skills:
  my-python-skill:
    image: python:3.11-slim
    language: python
    description: "Process text files with Python"
```

**Result:**
- Uses default Python image
- Python-only (advertised via MCP)
- 256MB memory, 0.5 CPU, 60s timeout
- No network access
- Files in `/outputs/` persist

---

### Example 2: Bash Script Skill

```yaml
skills:
  xml-parser:
    image: mcp-skills-bash-tools
    language: bash
    description: "Parse XML with xmlstarlet and jq"
    memory: 64MB
    cpu: "0.25"
    timeout: 30s
```

**Result:**
- Uses bash tools container
- Bash-only (advertised via MCP)
- Low resources (bash is lightweight)
- Quick timeout (scripts are fast)

---

### Example 3: Custom Container with Special Packages

```yaml
skills:
  ml-processor:
    image: mcp-skills-ml
    language: python
    description: "Machine learning data processing with PyTorch"
    dockerfile: docker/skills/Dockerfile.ml
    memory: 2GB
    cpu: "2.0"
    timeout: 600s
```

**Dockerfile:**

```dockerfile
# docker/skills/Dockerfile.ml
FROM python:3.11-slim

RUN pip install --no-cache-dir --break-system-packages \
    torch \
    numpy \
    pandas

WORKDIR /workspace
```

**Result:**
- Custom container with PyTorch
- High resources (ML needs memory/CPU)
- Long timeout (ML is slow)
- Built via build script

---

### Example 4: Multi-Language Skill

```yaml
skills:
  flexible-processor:
    image: mcp-skills-dual
    languages: [python, bash]
    description: "Process data with Python or bash scripts"
    memory: 256MB
```

**Container must have both:**

```dockerfile
FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    bash \
    coreutils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
```

**Result:**
- LLM can choose Python or bash
- Same container supports both
- LLM must specify `language` parameter

---

### Example 5: Network-Enabled Skill (Use Sparingly!)

```yaml
skills:
  web-scraper:
    image: mcp-skills-scraper
    language: python
    description: "Download and parse web content"
    network_mode: bridge
    network_justification: "Downloads HTML from specified URLs for parsing. No sensitive data transmitted."
    memory: 512MB
    timeout: 120s
```

**Result:**
- ⚠️ Has network access
- Documented justification
- Increased timeout (network I/O)
- More memory (caching)

---

## Migration from Old Format

### V1 Format (Deprecated)

```yaml
# Old format (simple map)
skills:
  docx: mcp-skills-docx
  pptx: mcp-skills-pptx
```

### V2 Format (Current)

```yaml
# New format (structured config)
skills:
  docx:
    image: mcp-skills-docx
    language: python
    description: "Word documents"
  
  pptx:
    image: mcp-skills-pptx
    language: python
    description: "PowerPoint presentations"
```

**Migration steps:**

1. **Add `image` field:**
   ```yaml
   # Before
   docx: mcp-skills-docx
   
   # After
   docx:
     image: mcp-skills-docx
   ```

2. **Add `language` field:**
   ```yaml
   docx:
     image: mcp-skills-docx
     language: python  # ← Add this
   ```

3. **Add descriptions (optional but recommended):**
   ```yaml
   docx:
     image: mcp-skills-docx
     language: python
     description: "Word document manipulation"  # ← Add this
   ```

---

## Validation

### Checking Your Configuration

**1. Syntax validation:**

```bash
# Check YAML is valid
python3 -c "import yaml; yaml.safe_load(open('config/skills/skill-images.yaml'))"
```

**2. Test loading:**

```bash
# Start MCP server and check logs
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml

# Look for:
# ✅ Initialized skill service with N skills
# ✅ Loaded skill 'docx'
# ✅ Loaded skill 'pptx'
```

**3. Verify containers exist:**

```bash
# Check all images referenced in config exist
grep "image:" config/skills/skill-images.yaml | \
  awk '{print $2}' | \
  sort -u | \
  while read img; do
    docker images | grep -q "$img" && echo "✅ $img" || echo "❌ $img MISSING"
  done
```

### Common Validation Errors

**1. Missing image:**

```yaml
skills:
  my-skill:
    # Error: no image field
    language: python
```

**Fix:**

```yaml
skills:
  my-skill:
    image: python:3.11-slim  # Add image
    language: python
```

**2. Both language and languages:**

```yaml
skills:
  my-skill:
    image: python:3.11-slim
    language: python      # ← Error
    languages: [python, bash]  # ← Can't have both
```

**Fix (choose one):**

```yaml
# Single language
skills:
  my-skill:
    image: python:3.11-slim
    language: python

# Or multi-language
skills:
  my-skill:
    image: python:3.11-slim
    languages: [python, bash]
```

**3. Invalid resource format:**

```yaml
skills:
  my-skill:
    memory: 256  # Error: needs unit
    cpu: 0.5     # Error: must be string
    timeout: 60  # Error: needs unit
```

**Fix:**

```yaml
skills:
  my-skill:
    memory: "256MB"  # Add unit
    cpu: "0.5"       # Quote number
    timeout: "60s"   # Add unit
```

---

## Best Practices

### 1. Always Specify Language

```yaml
# ✅ GOOD
skills:
  my-skill:
    image: python:3.11-slim
    language: python  # Explicit

# ⚠️ ACCEPTABLE but less clear
skills:
  my-skill:
    image: python:3.11-slim
    # language: python (inherited from defaults)
```

**Why:** Makes capabilities explicit, helps with debugging, clearer for other developers.

### 2. Use Descriptive Names

```yaml
# ✅ GOOD
skills:
  xml-policy-parser:
    description: "Parse policy XML with xmlstarlet"

# ❌ BAD
skills:
  xml:
    description: "Does stuff"
```

### 3. Set Appropriate Resource Limits

```yaml
# ✅ GOOD - Matches actual needs
skills:
  bash-formatter:
    memory: 64MB    # Bash is lightweight
    cpu: "0.25"
    timeout: 30s

# ❌ BAD - Over-provisioned
skills:
  bash-formatter:
    memory: 2GB     # Wasteful
    cpu: "2.0"
    timeout: 600s
```

### 4. Document Network Access

```yaml
# ✅ GOOD
skills:
  api-client:
    network_mode: bridge
    network_justification: "Calls weather API. No sensitive data."

# ❌ BAD
skills:
  api-client:
    network_mode: bridge
    # No justification!
```

### 5. Use Defaults Wisely

```yaml
# ✅ GOOD - Override only what's different
defaults:
  language: python
  memory: 256MB

skills:
  lightweight-bash:
    language: bash  # Override
    memory: 64MB    # Override
  
  standard-python:
    # Inherits: language=python, memory=256MB

# ❌ BAD - Repeat everything
skills:
  lightweight-bash:
    image: bash
    language: bash
    network_mode: none
    memory: 64MB
    cpu: "0.5"
    timeout: 60s
  
  standard-python:
    image: python:3.11-slim
    language: python
    network_mode: none
    memory: 256MB
    cpu: "0.5"
    timeout: 60s
```

---

## Troubleshooting

### Language Not Advertised

**Problem:** LLM doesn't know skill language

**Cause:** `language` field missing or in defaults only

**Solution:**

```yaml
# Add to specific skill
skills:
  my-skill:
    image: python:3.11-slim
    language: python  # ← Add this
```

### Container Not Found

**Problem:** `Error: image 'mcp-skills-custom' not found`

**Solutions:**

```bash
# 1. Check image exists
docker images | grep mcp-skills-custom

# 2. Build if Dockerfile exists
cd docker/skills
./build-skills-images.sh custom

# 3. Update config to use existing image
# In skill-images.yaml:
skills:
  my-skill:
    image: python:3.11-slim  # Use default instead
```

### Language Mismatch Error

**Problem:** `skill 'my-skill' requires language to be one of [python], got 'bash'`

**Cause:** LLM used wrong language or config wrong

**Solutions:**

```yaml
# Option 1: Fix config to match actual container
skills:
  my-skill:
    language: bash  # Change to bash

# Option 2: Support both languages
skills:
  my-skill:
    languages: [python, bash]
```

### Out of Memory

**Problem:** Container killed, files incomplete

**Solution:**

```yaml
skills:
  my-skill:
    memory: 512MB  # Increase from 256MB
```

Or:
```yaml
defaults:
  memory: 512MB  # Increase global default
```

---

## Reference Tables

### Memory Guidelines by Task

| Task Type | Recommended Memory |
|-----------|-------------------|
| Simple bash scripts | 64MB |
| Text processing | 128MB |
| Standard documents (docx, pptx) | 256MB |
| PDF processing | 512MB |
| Image manipulation | 512MB - 1GB |
| ML inference | 1GB - 4GB |
| Video processing | 2GB - 8GB |

### CPU Guidelines by Task

| Task Type | Recommended CPU |
|-----------|----------------|
| Quick scripts | 0.25 |
| Standard processing | 0.5 |
| Heavy processing | 1.0 |
| Parallel/encoding | 2.0+ |

### Timeout Guidelines by Task

| Task Type | Recommended Timeout |
|-----------|-------------------|
| Quick tasks | 30s |
| Standard processing | 60s |
| Large files | 2-5m |
| Very large files | 10m |
| Batch/training | 30m - 1h |

### Common Image Bases

| Base Image | Size | When to Use |
|------------|------|-------------|
| `python:3.11-slim` | ~150MB | Most Python skills |
| `python:3.11-alpine` | ~50MB | Minimal Python |
| `bash:latest` | ~15MB | Pure bash scripts |
| `ubuntu:22.04` | ~80MB | Need system tools |

---

## Changelog

### Version 2.0 (Current)

- Added `language` field for MCP advertising
- Added `languages` array for multi-language skills
- Added `defaults` section for inheritance
- Added `description` field
- Added `dockerfile` field for build tracking
- Added `network_justification` requirement
- Deprecated simple string mapping format

### Version 1.0 (Deprecated)

- Simple map: `skill-name: image-name`
- No language declaration
- No resource configuration
- No inheritance

---

## See Also

- [Complete Skills Guide](COMPLETE_GUIDE.md) - Full skills documentation
- [Creating Skills](creating-skills.md) - Build your own skills
- [Quick Start](quick-start.md) - Get started quickly
- [Troubleshooting](../TROUBLESHOOTING.md) - Common issues

---

**Last updated:** January 20, 2026  
**File format version:** 2.0  
**Maintained by:** mcp-cli-go community
