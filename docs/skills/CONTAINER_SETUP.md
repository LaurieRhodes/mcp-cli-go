# Skills Container Setup

## Requirements

1. Docker or Podman installed
2. Outputs directory configured

## Setup

### 1. Build Skill Images

```bash
cd docker/skills
./build-skills-images.sh
```

Built images:

- `mcp-skills-docx`
- `mcp-skills-pptx`
- `mcp-skills-xlsx`
- `mcp-skills-pdf`
- `mcp-skills-office`

### 2. Configure Outputs Directory

Edit `config/settings.yaml`:

```yaml
skills:
  outputs_dir: "/path/to/your/outputs"
```

Create the directory:

```bash
mkdir -p /path/to/your/outputs
```

### 3. Verify

```bash
# Images built
docker images | grep mcp-skills

# Directory exists
ls -ld /path/to/your/outputs

# Configuration set
grep -A 2 "^skills:" config/settings.yaml
```

## Default Image

Skills use `python:3.11-alpine` unless specified in `config/skills/skill-images.yaml`.

## Container Security

Containers run with:

- `--read-only`
- `--network=none`
- `--memory=256m`
- `--cpus=0.5`

---

Last updated: January 6, 2026
