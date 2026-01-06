# Quick Start

Get skills running in 5 minutes.

## Prerequisites

- Docker or Podman installed
- mcp-cli built

## Step 1: Build Container Images

```bash
cd docker/skills
./build-skills-images.sh
```

## Step 2: Configure Outputs

Edit `config/settings.yaml`:

```yaml
skills:
  outputs_dir: "/tmp/mcp-outputs"
```

Create directory:

```bash
mkdir -p /tmp/mcp-outputs
```

## Step 3: Start MCP Server

```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

You should see:

```
✅ Initialized skill service with N skills
```

## Step 4: Test with LLM

If using Claude Desktop, add to `claude_desktop_config.json`:

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

Restart Claude Desktop.

## Step 5: Verify

Ask the LLM:

> "Create a simple PowerPoint presentation"

The LLM should use the skills to create a .pptx file in your outputs directory.

## Troubleshooting

**Files not appearing?**
- Check `config/settings.yaml` has correct `outputs_dir`
- Verify directory exists: `ls -ld /tmp/mcp-outputs`

**Images not built?**
- Run: `docker images | grep mcp-skills`
- Rebuild: `cd docker/skills && ./build-skills-images.sh`

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for more help.

---

Last updated: January 6, 2026
