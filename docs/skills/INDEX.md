# Skills Documentation Index

## Getting Started

**New to skills?** Start here:

1. [README.md](README.md) - Overview and quick start
2. [CONTAINER_SETUP.md](CONTAINER_SETUP.md) - Build and configure containers
3. [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md) - Configure file persistence
4. [quick-start.md](quick-start.md) - 5-minute setup guide

## Core Guides

- [overview.md](overview.md) - How skills work
- [WHY_SKILLS_MATTER.md](WHY_SKILLS_MATTER.md) - Benefits and use cases
- [auto-loading.md](auto-loading.md) - Automatic skill discovery
- [creating-skills.md](creating-skills.md) - Build custom skills

## Reference

- [quick-reference.md](quick-reference.md) - Quick lookup
- [docker-podman-execution.md](docker-podman-execution.md) - Container details
- [CONTAINERS_README.md](CONTAINERS_README.md) - Container architecture
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues

## Quick Setup

```bash
# 1. Build images
cd docker/skills && ./build-skills-images.sh

# 2. Configure outputs
echo "skills:
  outputs_dir: \"/tmp/mcp-outputs\"" >> config/settings.yaml

# 3. Create directory
mkdir -p /tmp/mcp-outputs

# 4. Start server
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

## Available Skills

- **docx** - Word documents
- **pptx** - PowerPoint presentations
- **xlsx** - Excel spreadsheets
- **pdf** - PDF manipulation

## Documentation by Topic

### Setup & Configuration
- [CONTAINER_SETUP.md](CONTAINER_SETUP.md)
- [OUTPUTS_DIRECTORY.md](OUTPUTS_DIRECTORY.md)
- [quick-start.md](quick-start.md)

### Understanding Skills
- [overview.md](overview.md)
- [WHY_SKILLS_MATTER.md](WHY_SKILLS_MATTER.md)
- [auto-loading.md](auto-loading.md)

### Development
- [creating-skills.md](creating-skills.md)
- [docker-podman-execution.md](docker-podman-execution.md)

### Reference & Help
- [quick-reference.md](quick-reference.md)
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

---

Last updated: January 6, 2026
