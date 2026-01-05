# Docker/Podman Code Execution Guide

Complete guide to understanding and configuring code execution in skills using Docker or Podman.

## Table of Contents

- [Overview](#overview)
- [Why Containerization?](#why-containerization)
- [Docker vs Podman](#docker-vs-podman)
- [Installation](#installation)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Performance](#performance)

---

## Overview

The `execute_skill_code` tool runs arbitrary Python code in an **isolated sandbox** using either Docker or Podman. This enables skills to:

✅ Execute custom code written by Claude  
✅ Import helper libraries from skill directories  
✅ Process files safely in isolation  
✅ Create documents (Word, PDF, Excel, etc.)  

### Production Status

**Current Implementation:** ✅ **Fully Functional**

- Docker support
- Podman support (recommended for rootless operation)
- Automatic detection (`execution_mode: auto`)
- Secure sandboxing with resource limits
- PYTHONPATH configuration for skill libraries

**Not Implemented:** ⚠️ **Experimental**

- `workflow.yaml` execution (stub only)
- Active mode workflow orchestration

---

## Why Containerization?

### Security Isolation

Code execution happens in **complete isolation**:

```
┌─────────────────────────────────────┐
│ Host System                         │
│                                     │
│  ┌──────────────────────────────┐  │
│  │ Container (Isolated)         │  │
│  │                              │  │
│  │  • No network access         │  │
│  │  • Limited CPU/RAM           │  │
│  │  │  • Read-only skill libs   │  │
│  │  • Temporary workspace       │  │
│  │  • Process isolation         │  │
│  └──────────────────────────────┘  │
│                                     │
└─────────────────────────────────────┘
```

### What Gets Protected

**Host system protected from:**

- Malicious code execution
- Resource exhaustion (CPU/RAM bombs)
- File system access outside workspace
- Network requests
- Privilege escalation

**Skill libraries protected from:**

- Accidental modification
- Malicious tampering
- State corruption



---

## Installation

### macOS

**Docker:**

```bash
# Via Homebrew
brew install --cask docker

# Start Docker Desktop
open -a Docker
```

**Podman:**

```bash
# Via Homebrew
brew install podman

# Initialize
podman machine init
podman machine start
```

### Linux

**Docker:**

```bash
# Official installation script
curl -fsSL https://get.docker.com | sh

# Add user to docker group (optional, for non-root)
sudo usermod -aG docker $USER
# Log out and back in

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

**Podman (Recommended):**

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install podman

# Fedora/RHEL
sudo dnf install podman

# Arch
sudo pacman -S podman

# No additional setup needed - rootless by default!
```

### Windows

**Docker:**

- Download Docker Desktop from https://www.docker.com/
- Install and restart
- Enable WSL 2 backend if prompted

**Podman:**

- Download Podman Desktop from https://podman.io/
- Install and restart

### Verification

Test your installation:

```bash
# Docker
docker --version
docker run hello-world

# Podman
podman --version
podman run hello-world
```

---

## How It Works

### Execution Flow

```
Claude writes code
       ↓
MCP client sends to execute_skill_code
       ↓
mcp-cli receives request
       ↓
SkillService prepares workspace
       ↓
┌────────────────────────────────┐
│ Create Temporary Workspace    │
│                                │
│ /tmp/skill-workspace-XXXXX/    │
│ ├── script.py  (Claude's code) │
│ ├── input.csv  (optional)      │
│ └── output/    (results)       │
└────────────────────────────────┘
       ↓
┌────────────────────────────────┐
│ Launch Container               │
│                                │
│ Mounts:                        │
│ • /workspace (read-write)      │
│   → temporary workspace        │
│ • /skill (read-only)           │
│   → skill directory            │
│                                │
│ Environment:                   │
│ • PYTHONPATH=/skill            │
│                                │
│ Resources:                     │
│ • 256MB RAM                    │
│ • 0.5 CPU                      │
│ • 100 PIDs max                 │
│ • No network                   │
└────────────────────────────────┘
       ↓
Execute: python3 /workspace/script.py
       ↓
Capture stdout/stderr
       ↓
Copy output files from /workspace/output
       ↓
Cleanup container and workspace
       ↓
Return results to Claude
```

### Directory Mounts

Two directories are mounted into the container:

**1. `/workspace` (read-write):**

```
/workspace/
├── script.py       ← Claude's generated code
├── input.csv       ← Optional input files (base64 decoded)
└── output/         ← Generated files (copied back)
```

**2. `/skill` (read-only):**

```
/skill/
├── SKILL.md
├── scripts/
│   ├── __init__.py
│   ├── document.py   ← Helper libraries
│   └── formatting.py
└── references/
    └── docs.md
```

### PYTHONPATH Configuration

The container sets `PYTHONPATH=/skill`, allowing imports:

```python
# In Claude's generated code (script.py)
from scripts.document import Document  # ✅ Works!
from scripts.formatting import apply_style

doc = Document()
doc.add_heading('My Document', 0)
doc.save('/workspace/output/doc.docx')
```

### Image Configuration

**Container image:** `python:3.11-slim`

**Why this image:**

- ✅ Small size (~120MB)
- ✅ Python 3.11 (modern, stable)
- ✅ Minimal attack surface
- ✅ Standard library included

**What's included:**

- Python 3.11
- pip
- Standard library (json, csv, re, etc.)

**What's NOT included:**

- NumPy, Pandas (not needed for document skills)
- Network tools (blocked anyway)
- Development tools

**For skills requiring additional packages:**
Currently not supported. Future enhancement would allow custom images per skill.

---

## Configuration

### Execution Modes

Set in `config/runasMCP/mcp_skills_stdio.yaml`:

```yaml
skills_config:
  execution_mode: auto  # Recommended
```

**Options:**

| Mode      | Behavior                                    |
| --------- | ------------------------------------------- |
| `auto`    | Detects Docker/Podman, enables if available |
| `active`  | **Not implemented** (stub only)             |
| `passive` | No execution, documentation only            |

### Automatic Detection

With `execution_mode: auto`, the system:

1. Checks for `podman` command
2. If not found, checks for `docker` command
3. If found, initializes executor
4. If not found, falls back to passive mode

**Detection logic:**

```go
// Try Podman first (preferred for rootless)
if commandExists("podman") {
    executor = NewPodmanExecutor()
}
// Fall back to Docker
else if commandExists("docker") {
    executor = NewDockerExecutor()
}
// No container runtime available
else {
    log.Warn("No container runtime found, passive mode only")
}
```

### Logs

Check initialization logs:

```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | grep Executor
```

**Success:**

```
[INFO] ✅ Executor initialized: Podman 4.9.3 (native)
```

**Fallback:**

```
[WARN] Docker/Podman not available, falling back to passive mode
```

---

## Troubleshooting

### Issue: "Executor not available"

**Error:**

```
Code execution failed: executor not available (Docker/Podman not found)
```

**Diagnosis:**

```bash
# Check for Docker
docker --version

# Check for Podman
podman --version

# Check PATH
echo $PATH
which docker
which podman
```

**Fix:**

Install Docker or Podman (see [Installation](#installation))

### Issue: "Permission denied" (Docker)

**Error:**

```
Got permission denied while trying to connect to the Docker daemon socket
```

**Fix:**

Add user to docker group:

```bash
sudo usermod -aG docker $USER
```

Then log out and back in.

**Or:** Use rootless Docker/Podman.

### Issue: Container startup slow

**Symptom:** First execution takes 5-10 seconds.

**Cause:** Image needs to be pulled from registry.

**Fix:** Pre-pull the image:

```bash
# Docker
docker pull python:3.11-slim

# Podman
podman pull python:3.11-slim
```

Subsequent executions will be ~300ms.

### Issue: "No space left on device"

**Symptom:** Container fails to start.

**Cause:** Too many old containers/images.

**Fix:**

Clean up:

```bash
# Docker
docker system prune -a

# Podman
podman system prune -a
```

### Issue: Code works locally but fails in container

**Common causes:**

1. **Missing imports:**
   
   ```python
   # ❌ numpy not in python:3.11-slim
   import numpy as np
   
   # ✅ standard library only
   import json
   import csv
   ```

2. **File paths:**
   
   ```python
   # ❌ Absolute host paths don't exist
   open('/home/user/file.txt')
   
   # ✅ Use /workspace
   open('/workspace/file.txt')
   ```

3. **Network access:**
   
   ```python
   # ❌ Network disabled
   import requests
   requests.get('https://api.example.com')
   
   # ✅ File-based operations only
   ```

---

## Security

### Sandbox Boundaries

**What the container CAN'T do:**

❌ Access host filesystem (except mounted directories)  
❌ Make network requests  
❌ See other processes  
❌ Escalate privileges  
❌ Persist data between executions  
❌ Access environment variables from host  
❌ Modify skill libraries (read-only mount)  

**What the container CAN do:**

✅ Execute Python code  
✅ Read skill helper libraries  
✅ Write to /workspace  
✅ Use Python standard library  
✅ Create files in output directory  

### Resource Limits

Enforced per execution:

```
Memory:  256 MB
CPU:     0.5 cores
PIDs:    100 maximum
Timeout: 60 seconds
Network: Disabled (--network=none)
```

**Why these limits:**

- **256MB RAM:** Enough for document processing, prevents memory bombs
- **0.5 CPU:** Fair share, prevents CPU exhaustion
- **100 PIDs:** Prevents fork bombs
- **60s timeout:** Prevents infinite loops
- **No network:** Prevents data exfiltration

### Security Best Practices

**For users:**

1. Review generated code before execution if concerned
2. Use `execution_mode: passive` in untrusted environments
3. Don't store sensitive data in skill directories

**For developers:**

1. Keep skill libraries minimal
2. Don't include secrets in helper libraries
3. Validate user inputs
4. Use read-only mounts for skill directories

---

## Performance

### Typical Execution Times

```
Container startup:    ~250ms (cached image)
Code execution:       ~30-50ms
File I/O:            ~10-20ms
──────────────────────────────────
Total:               ~300ms
```

**First run:** ~5-10 seconds (image pull)  
**Subsequent runs:** ~300ms (image cached)

### Optimization Tips

**1. Pre-pull the image:**

```bash
docker pull python:3.11-slim
```

**2. Keep code simple:**

```python
# ✅ Fast - direct file operations
doc = Document()
doc.save('/workspace/output/file.docx')

# ❌ Slow - complex nested loops
for i in range(10000):
    for j in range(10000):
        # ...
```

**3. Minimize file I/O:**

```python
# ✅ Write once
with open('/workspace/output/file.txt', 'w') as f:
    f.write(large_content)

# ❌ Write many times
for line in lines:
    with open('/workspace/output/file.txt', 'a') as f:
        f.write(line)
```

### Performance Monitoring

Check execution times in logs:

```bash
./mcp-cli serve --verbose config/runasMCP/mcp_skills_stdio.yaml 2>&1 | grep "Executed in"
```

Output:

```
[Executed in 287ms]
```

---

## Advanced Topics

### Custom Container Images (Not Yet Supported)

**Future enhancement:** Allow skills to specify custom images:

```yaml
# skill/SKILL.md frontmatter (planned)
---
name: data-science-skill
description: "Data analysis with pandas"
container_image: python:3.11-slim-pandas  # Custom image
---
```

**Current workaround:** Build custom image with all needed packages, tag as `python:3.11-slim`, replace default.

### Rootless Podman (Recommended)

Podman runs rootless by default on Linux:

```bash
# Check if rootless
podman info | grep rootless
# Output: rootless: true

# All containers run as your user
ps aux | grep podman
```

**Benefits:**

- ✅ No root daemon
- ✅ Better security
- ✅ User namespace isolation
- ✅ No sudo required

### Container Cleanup

Containers are automatically removed after execution (`--rm` flag), but images persist:

```bash
# Check disk usage
docker system df
podman system df

# Clean up old images
docker system prune -a
podman system prune -a
```

---

## Summary

**Docker/Podman execution provides:**

✅ **Secure isolation** - Code runs in sandbox  
✅ **Resource limits** - Prevents resource exhaustion  
✅ **Read-only skills** - Helper libraries protected  
✅ **Fast execution** - ~300ms total time  
✅ **Automatic detection** - Zero configuration needed  
✅ **Production ready** - Fully implemented and tested  

**Recommended setup:**

```yaml
# config/runasMCP/mcp_skills_stdio.yaml
skills_config:
  execution_mode: auto  # Detects Docker/Podman automatically
```

**Next steps:**

- Install Docker or Podman (see [Installation](#installation))
- Verify with `--verbose` logging
- Test with `execute_skill_code` tool
- Review [Auto-Loading Guide](auto-loading.md) for full skills setup

---

**Last Updated:** January 4, 2026  
**Tested With:** Docker 24.0.7, Podman 4.9.3, mcp-cli v0.1.0
