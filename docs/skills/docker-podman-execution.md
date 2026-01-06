# Docker/Podman Execution

Skills execute in isolated containers for security and portability.

## Container Runtime Detection

mcp-cli automatically detects available runtime:
1. Checks for Podman
2. Falls back to Docker
3. Fails if neither available

## Container Images

Built via `docker/skills/build-skills-images.sh`:

- `mcp-skills-docx` - python-docx, lxml
- `mcp-skills-pptx` - python-pptx, Pillow
- `mcp-skills-xlsx` - openpyxl
- `mcp-skills-pdf` - pypdf, pdf2image
- `mcp-skills-office` - Combined (docx + pptx + xlsx)

Default fallback: `python:3.11-alpine`

## Image Selection

Configured in `config/skills/skill-images.yaml`:

```yaml
skills:
  pptx: mcp-skills-pptx
  docx: mcp-skills-docx
  xlsx: mcp-skills-xlsx
  pdf: mcp-skills-pdf
```

## Container Mounts

Every execution creates a fresh container with:

```bash
-v /workspace:/workspace:rw    # Temporary
-v /outputs:/outputs:rw        # Persistent (from host)
-v /skill:/skill:ro            # Read-only libraries
```

## Security Settings

```bash
--rm                          # Auto-remove
--read-only                   # Read-only root
--network=none               # No network
--memory=256m                # Memory limit
--cpus=0.5                   # CPU limit
--pids-limit=100             # Process limit
--security-opt=no-new-privileges
--cap-drop=ALL
```

## Execution Flow

1. Create temporary workspace
2. Write code to workspace
3. Start container with mounts
4. Execute code
5. Capture output
6. Clean up workspace
7. Container auto-removed

## Building Images

```bash
cd docker/skills
./build-skills-images.sh
```

Build specific image:
```bash
./build-skills-images.sh pptx
```

## Verification

```bash
# Check images exist
docker images | grep mcp-skills

# Test execution
docker run --rm mcp-skills-pptx python -c "from pptx import Presentation; print('OK')"
```

---

Last updated: January 6, 2026
