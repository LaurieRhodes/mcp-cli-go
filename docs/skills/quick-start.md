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

## Step 3: Test It

**Option A: Chat Mode (Quickest)**

```bash
./mcp-cli chat --servers skills
```

Ask: "Create a simple PowerPoint presentation about Python"

**Option B: As MCP Server (For Claude Desktop, VS Code, etc.)**

```bash
./mcp-cli serve config/runasMCP/mcp_skills_stdio.yaml
```

You should see:
```
✅ Initialized skill service with N skills
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

Restart your MCP client.

## Verify

Files created by skills should appear in `/tmp/mcp-outputs/`

```bash
ls -la /tmp/mcp-outputs/
```

## Troubleshooting

**Files not appearing?**
- Check `config/settings.yaml` has correct `outputs_dir`
- Verify directory exists: `ls -ld /tmp/mcp-outputs`
- Ensure code saves to `/outputs/` path

**Images not built?**
- Run: `docker images | grep mcp-skills`
- Rebuild: `cd docker/skills && ./build-skills-images.sh`

**Need more help?**
- See [COMPLETE_GUIDE.md](COMPLETE_GUIDE.md#troubleshooting)
- See [INDEX.md](INDEX.md#troubleshooting)

---

Last updated: January 20, 2026
