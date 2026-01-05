# Quick Start Guide

Get up and running with skills in 5 minutes.

## Prerequisites

- mcp-cli installed
- Docker or Podman installed
- Claude Desktop (optional, for MCP integration)

## Step 1: Initialize with Skills

```bash
./mcp-cli init
```

When prompted:

```
🎯 Anthropic Skills System:
Skills provide helper libraries for document creation, data processing, etc.
Set up example skills (docx, pdf, pptx, xlsx, test-execution)? [Y/n]: y
```

This creates:

```
config/
├── skills/
│   ├── test-execution/
│   ├── docx/
│   ├── pdf/
│   ├── pptx/
│   ├── xlsx/
│   └── frontend-design/
└── runas/
    └── skills-auto.yaml
```

## Step 2: Start Skills MCP Server

```bash
./mcp-cli serve config/runas/skills-auto.yaml
```

You should see:

```
✅ Executor initialized: Podman 4.9.3 (native)
✅ Script execution enabled for 6 skills
✅ Initialized skill service with 18 skills
Auto-generated 19 MCP tool definitions from skills
```

## Step 3: Configure Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or  
`%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "skills": {
      "command": "/absolute/path/to/mcp-cli",
      "args": ["serve", "/absolute/path/to/config/runas/skills-auto.yaml"]
    }
  }
}
```

**Important:** Use absolute paths!

## Step 4: Restart Claude Desktop

Close and reopen Claude Desktop to load the skills server.

## Step 5: Test It!

In Claude Desktop, try:

> "Can you use the test-execution skill to create a greeting using the helper library?"

Claude should:

1. Load test-execution skill documentation
2. See that `greet()` function is available
3. Write code that imports and uses it
4. Execute via `execute_skill_code`
5. Show you the result

**Expected output:**

```
Hello, [Name]!
```

## Verify It's Working

### Check Available Tools

In Claude Desktop, skills should appear as tools. You can verify by asking:

> "What skills/tools do you have access to?"

Claude should list:

- test_execution (and other skill tools)
- execute_skill_code

### Next Steps

## Next Steps

### Learn About Auto-Loading

- **[Auto-Loading Guide](auto-loading.md)** - Complete guide to automatic skill discovery
- **[Quick Reference](quick-reference.md)** - Fast lookup for common tasks

### Learn More About Skills

- **[Overview](overview.md)** - Understand how skills work
- **[Why Skills Matter](WHY_SKILLS_MATTER.md)** - The philosophy behind skills

### Create Your Own Skills

- **[Creating Skills Guide](creating-skills.md)** - Build custom skills

### Technical Details

- **[README](README.md)** - Full documentation index

## Troubleshooting

If skills aren't appearing, see the [Auto-Loading Guide troubleshooting section](auto-loading.md#troubleshooting) or the [Quick Reference](quick-reference.md#-troubleshooting-commands).

---

**Pro Tip:** The auto-loading feature means you can add new skills by just dropping directories into `config/skills/` - no configuration changes needed!
