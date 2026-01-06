# Skills Container Architecture

## Overview

Skills execute in isolated Docker/Podman containers for security and portability.

## Quick Setup

```bash
# 1. Build images
cd docker/skills && ./build-skills-images.sh

# 2. Configure outputs
echo "skills:
  outputs_dir: \"$HOME/mcp-outputs\"" >> config/settings.yaml

# 3. Create directory
mkdir -p ~/mcp-outputs
```

## How It Works

```
Request → mcp-cli → Container → File in configured outputs directory
```

Container mounts:

- `/workspace` - Temporary (deleted)
- `/outputs` - Persistent (from host)
- `/skill` - Read-only code

## Configuration

**File**: `config/settings.yaml`

```yaml
skills:
  outputs_dir: "/home/user/mcp-outputs"
```

## Images

Built locally via `docker/skills/build-skills-images.sh`:

- `mcp-skills-docx`
- `mcp-skills-pptx`
- `mcp-skills-xlsx`
- `mcp-skills-pdf`
- `mcp-skills-office`

Default: `python:3.11-alpine`

## Security

- No network (`--network=none`)
- Read-only root
- Memory limit: 256MB
- CPU limit: 0.5 cores

---

Last updated: January 6, 2026
